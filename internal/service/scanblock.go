package service

import (
	"Web3Study/internal/models"
	"Web3Study/internal/node"
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

type ScanBlockProcessor struct {
	pg          *gorm.DB
	nodeManager *node.NodesManager
	node        *node.Node
}

func (p *ScanBlockProcessor) Start(ctx context.Context) error {
	// 获取数据库最新的区块
	latestBlockNumber, err := models.GetLastestBlockNumer(p.pg)
	if err != nil {
		panic(err)
	}
	//
	for {
		err = p.nodeManager.ExecuteWithRetry(func(client *ethclient.Client) error {

			return nil
		})
		if err != nil {
			//print out err logs, stop the progress gracefully
			return err
		}
		latestBlockNumber++
	}
}

func (p *ScanBlockProcessor) ProcessBlock() error {

	return nil
}
