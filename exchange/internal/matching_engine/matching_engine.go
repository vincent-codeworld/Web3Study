package matching_engine

import (
	"context"

	"github.com/emirpasic/gods/trees"
)

type MatchEngine struct {
	ctx           context.Context
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

}
