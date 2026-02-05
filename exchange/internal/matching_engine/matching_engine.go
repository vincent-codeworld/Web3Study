package matching_engine

import (
	"Web3Study/exchange/config"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/handlers"
	"Web3Study/exchange/internal/matching_engine/orderbook"
	"Web3Study/exchange/middleware"
	"Web3Study/exchange/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/maps/treemap"
)

/**
撮合流程：
1、判断订单一些信息，是否不符合风控要求，例如价格短时间波动，如果是需要取消订单
2、判断是否跟Taker是同个user，即self trade类型，是的话需要根据用户对应的配置进行处理

3、判断订单类型，market、limit、fok、ioc、iceberg.....
*/

/*
*

	                    ┌─────────────┐

		                │   New/Open  │
		                └──────┬──────┘
		                       │
		       ┌───────────────┼───────────────┐
		       │               │               │
		       ▼               ▼               ▼
		┌──────────┐   ┌───────────────┐   ┌──────────┐
		│ Cancelled│   │Partially Filled│   │  Filled  │
		└──────────┘   └───────┬───────┘   └──────────┘
		  (终结态)             │              (终结态)
		                       │
		           ┌───────────┴───────────┐
		           │                       │
		           ▼                       ▼
		┌────────────────────┐      ┌──────────┐
		│ Partially Cancelled │      │  Filled  │
		└────────────────────┘      └──────────┘
		      (终结态)                 (终结态)

状态分类
订单状态可以分为两类：中间态和终结态。

中间态
中间态表示订单生命周期尚未结束，后续还可能发生变化。

New/Open 订单已创建，尚未成交，挂在订单簿中等待撮合。

Partially Filled 订单部分成交，仍有剩余数量挂在订单簿中，等待后续撮合或用户操作。

终结态
终结态表示订单生命周期已结束，不会再发生任何变化。

Filled 订单完全成交，没有剩余数量。

Cancelled 订单未成交就被取消，可能是用户主动取消，也可能是系统取消如IOC未成交部分。

Partially Cancelled 订单部分成交后被取消了剩余部分。

Rejected 订单未能进入订单簿，在校验阶段就被拒绝。
*/
type MatchEngine struct {
	ctx           context.Context
	coinPairGroup uint8
	buyOrderBook  *orderbook.OrderBook
	sellOrderBook *orderbook.OrderBook
	orderMap      map[string]*dto.Order
}

func NewMatchEngine(ctx context.Context) *MatchEngine {
	return &MatchEngine{ctx: ctx}
}

func (engine *MatchEngine) Start() {
	//初始化订单簿，replay

	go func() {
		engine.replay()
		consumer := middleware.Consumer
		//todo 需要获取wal最新的offset
		err := utils.Retry(3, func() error {
			return consumer.SetOffset(123)
		})

		if err != nil {
			return
		}
		for {
			//1.消费数据，反序列化
			msg, err := consumer.FetchMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}
			newOrder := &dto.Order{}
			err = newOrder.UnmarshalVT(msg.Value)
			if err != nil {
				continue
			}
			//2.撮合
			engine.match(newOrder)
			//3.写入WAL缓存
			//4.批量写入WAL持久化后，提交offset(将对应的offset写入文件？)
			//5 将批量撮合结果写入下游kafka，提交offset，下游必须保证幂等性
		}
	}()
}

/*
* 1. Post Only订单要求只做Maker，不能立即成交。只要能遇到成交的对手单，立即拒绝，该模式是为了成为maker，提供订单流动性
 */
func (engine *MatchEngine) match(order *dto.Order) ([]*dto.OrderResult, error) {
	var result []*dto.OrderResult
	obFunc := func(side dto.Side) *orderbook.OrderBook {
		return engine.getOrderBook(side)
	}
	switch order.Type {
	case dto.OrderType_ORDER_TYPE_MARKET:
		{
			r, err := handlers.MarketHandler(order, obFunc)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		}
	case dto.OrderType_ORDER_TYPE_LIMIT:
		{
			r, err := handlers.LimitHandler(order, obFunc)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		}
	case dto.OrderType_ORDER_TYPE_POST_ONLY:
		{
			r, err := handlers.PostOnlyHandler(order, obFunc)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		}
	case dto.OrderType_ORDER_TYPE_FOK:
		{

		}
	default:
		return nil, fmt.Errorf("invalid order type: %s", order.Type)
	}

	return result, nil
}

func (engine *MatchEngine) Stop() {

}

func (engine *MatchEngine) getOrderBook(side dto.Side) *orderbook.OrderBook {
	if side == dto.Side_SIDE_BUY {
		return engine.buyOrderBook
	}
	return engine.sellOrderBook
}

/**
  1、日志写入，包括kafka消费的offset，不依赖kafka自身自带的
  2、全量日志为快照，是订单簿某个时间的状态，增量是input操作，需要重新执行撮合
  3、发送下游可以使用批量发送：条数+时间 判断，下游必须保证幂等性，有可能重复发送
  4、redis可以记录最大发送的seqId
*/
// replay wal logs,including incremental log and full log
// 路径: base_url/{coin_pair_group}/{trade_side}/{sequeue_id}
func (engine *MatchEngine) replay() {
	// 获取最新的全量快照，反序列化到orderBook
	// 获取增量日志，在全量的基础上进行回放

	fileSeparator := utils.GetFileSeparator()
	getFullLogPath := func(side dto.Side) string {
		sidePath := engine.getFullLogPath(side)
		paths, err := middleware.RkDb.FindPathsByPrefix(sidePath)
		//todo  这个错误处理后期优化，理论来说这里不能出现错误，如果有错误，后续的逻辑时无法执行
		if err != nil {
			panic(err)
		}
		maxSeqId := 0
		fullPath := ""
		for _, path := range paths {
			tempaths := strings.Split(path, fileSeparator)
			seqIdStr := tempaths[len(tempaths)-1]
			seqId, _ := strconv.Atoi(seqIdStr)
			if seqId > maxSeqId {
				maxSeqId = seqId
				fullPath = path
			}
		}
		return fullPath
	}
	sellFullLogPath := getFullLogPath(dto.Side_SIDE_SELL)
	read, err := middleware.RkDb.Read(sellFullLogPath)
	if !errors.Is(err, middleware.RockFileNotFound) {
		panic(read)
	}
}
func (engine *MatchEngine) getFullLogPath(side dto.Side) string {
	fileSeparator := utils.GetFileSeparator()
	findPath := fmt.Sprintf("%s%s%d%s%d", config.GlobalConf.Wal.FullLogsPrePath, fileSeparator, engine.coinPairGroup, fileSeparator, side)
	return findPath
}

func (engine *MatchEngine) MarshalSnapShort() []byte {
	buyBook := make([]*dto.Order, engine.buyOrderBook.Size())
	sellBook := make([]*dto.Order, engine.sellOrderBook.Size())
	var sequenceId int64
	marshalFunc := func(book *treemap.Map, orders []*dto.Order) {
		iter := book.Iterator()
		for i := 0; iter.Next(); i++ {
			o := iter.Value().(*dto.Order)
			if o.SeqId > sequenceId {
				sequenceId = o.SeqId
			}
			orders[i] = o
		}
	}

	marshalFunc(engine.buyOrderBook.Map, buyBook)
	marshalFunc(engine.sellOrderBook.Map, sellBook)

	snapshot := dto.OrderBookSnapshot{
		SequenceId: uint64(sequenceId),
		CoinGroup:  uint32(engine.coinPairGroup),
		Asks:       sellBook,
		Bids:       buyBook,
	}
	data, err := snapshot.MarshalVT()
	if err != nil {
		return nil
	}
	return data
}

// todo
func (engine *MatchEngine) UnMarshalSnapShort() {}
