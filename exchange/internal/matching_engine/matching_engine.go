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

// replay wal logs,including incremental log and full log
func (engine *MatchEngine) replay() {
	// 获取最新的全量快照，反序列化到orderBook
	// 获取增量日志，在全量的基础上进行回放
}
