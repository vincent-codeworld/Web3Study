package matching_engine

import (
	"Web3Study/exchange/config"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/middleware"
	"Web3Study/exchange/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/trees"
)

type MatchEngine struct {
	ctx           context.Context
	coinPairGroup uint8
	buyOrderBook  *trees.Tree
	sellOrderBook *trees.Tree
}

func NewMatchEngine(ctx context.Context) *MatchEngine {
	return &MatchEngine{ctx: ctx}
}

func (engine *MatchEngine) Start() {

}

func (engine *MatchEngine) Stop() {

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
	sellFullLogPath := getFullLogPath(dto.SELL)
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
