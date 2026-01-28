package service

import (
	"Web3Study/scanchains/internal/models"
	"Web3Study/scanchains/internal/node"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
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
			//   考虑超时时间
			block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(latestBlockNumber)))
			if err != nil {
				//print out err logs
				return err
			}
			err = p.ProcessBlock(client, block)
			if err != nil {
				//print out err logs
				return err
			}
			return nil
		})
		if err != nil {
			//print out err logs, stop the progress gracefully
			return err
		}
		latestBlockNumber++
	}
}

func (p *ScanBlockProcessor) ProcessBlock(client *ethclient.Client, block *types.Block) error {
	txs := block.Transactions()
	/*	for i := range txs {
		receipt, err := client.TransactionReceipt(context.Background(), txs[i].Hash())
		if err != nil {
			return err
		}

	}*/
	return nil
}

func (p *ScanBlockProcessor) processTransactions(client *ethclient.Client, block *types.Block) error {
	var batchSize uint16 = 100
	batchRecords := make([]*models.TransactionRecord, batchSize)
	processTx := func(client *ethclient.Client, pgTx *gorm.DB, block *types.Block, tx *types.Transaction, index uint16) error {
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			return err
		}
		txRecord, err := models.GenerateTransactionRecord(block, tx, receipt)
		if err != nil {
			return err
		}
		batchRecords[index] = txRecord
		return nil
	}

	txs := block.Transactions()
	pgTx := p.pg.Begin()
	var j uint16 = 0
	for _, tx := range txs {
		err := processTx(client, pgTx, block, tx, j)
		if err != nil {
			pgTx.Rollback()
			return err
		}
		j++
		if j >= batchSize {
			pgTx.CreateInBatches(batchRecords, int(batchSize))
			j = 0
		}
	}
	pgTx.CreateInBatches(batchRecords[:j], int(j))
	_ = pgTx.Commit()
	return nil
}
