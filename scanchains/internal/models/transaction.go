package models

import (
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type TransactionRecord struct {
	// 主键
	Id uint64 `gorm:"primaryKey;autoIncrement;column:id"`

	// 链名称
	// 建立复合唯一索引的一部分，防止同一条链出现重复的 TxHash
	ChainName string `gorm:"type:varchar(50);not null;uniqueIndex:idx_chain_tx_hash,priority:1;column:chain_name"`

	// 区块高度
	BlockNumber uint64 `gorm:"type:bigint;not null;index;column:block_number"`

	// 区块哈希
	BlockHash string `gorm:"type:varchar(66);not null;column:block_hash"`

	// 交易哈希
	// 建立复合唯一索引，确保 (ChainName, TxHash) 唯一
	TxHash string `gorm:"type:varchar(66);not null;uniqueIndex:idx_chain_tx_hash,priority:2;column:tx_hash"`

	// 交易在区块中的索引位置
	TxIndex uint `gorm:"type:integer;not null;column:tx_index"`

	// 发送方地址，加索引用于查询用户交易历史
	FromAddress string `gorm:"type:varchar(100);not null;index;column:from_address"`

	// 接收方地址，加索引。注意：合约创建交易 To 可能是空，所以不加 not null
	ToAddress string `gorm:"type:varchar(100);index;column:to_address"`

	// 交易金额 (Wei)
	// 区块链金额通常超大 (uint256)，Go中使用 string 承载
	// 数据库建议使用 varchar 存储以保证精度不丢失，或者使用 numeric(78, 0)
	Value string `gorm:"type:varchar(100);default:'0';column:value"`

	// Gas 消耗量
	GasUsed uint64 `gorm:"type:bigint;column:gas_used"`

	// Gas 价格
	// 同样可能很大，使用 varchar 存储
	GasPrice string `gorm:"type:varchar(100);default:'0';column:gas_price"`

	// 链上原始状态 (如以太坊: 1-成功, 0-失败)
	RawStatus uint8 `gorm:"type:smallint;column:raw_status"`

	// 本地系统处理状态
	Status uint8 `gorm:"type:smallint;default:0;not null;column:status;comment:0-pending,1-processed"`

	// 区块时间戳，用于排序和展示
	Timestamp uint64 `gorm:"type:bigint;index;column:timestamp"`

	// 时间字段
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`

	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func (_ TransactionRecord) TableName() string {
	return "transaction_records"
}

// 需要对block number 建索引
func GetLastestBlockNumer(db *gorm.DB) (uint64, error) {
	var model *TransactionRecord
	var blockNumber uint64
	err := db.Table(model.TableName()).
		Select("max(block_number)").
		First(&blockNumber).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

func GenerateTransactionRecord(block *types.Block, tx *types.Transaction, receipt *types.Receipt) (*TransactionRecord, error) {
	fromAdress, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return nil, err
	}
	toAdress := ""
	if tx.To() != nil {
		toAdress = tx.To().Hex()
	}
	return &TransactionRecord{
		BlockNumber: block.NumberU64(),
		TxHash:      tx.Hash().String(),
		TxIndex:     receipt.TransactionIndex,
		FromAddress: fromAdress.String(),
		ToAddress:   toAdress,
		Value:       tx.Value().String(),
		GasUsed:     receipt.GasUsed,
		GasPrice:    tx.GasPrice().String(),
		RawStatus:   uint8(receipt.Status),
		Timestamp:   block.Time(),
	}, nil

}
