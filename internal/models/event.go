package models

import (
	"time"

	"gorm.io/gorm"
)

type EventRecord struct {
	// 主键，自增
	Id uint64 `gorm:"primaryKey;autoIncrement;column:id"`

	// 交易ID，使用 bigint 存储，加索引方便反查
	TxId uint64 `gorm:"type:bigint;not null;index;column:tx_id"`

	// 链名称，如 "ethereum", "bsc"，通常长度有限，加索引
	ChainName string `gorm:"type:varchar(50);not null;index;column:chain_name"`

	// 区块高度，核心查询字段，必须加索引
	BlockNumber uint64 `gorm:"type:bigint;not null;index;column:block_number"`

	// 日志在区块内的索引
	LogIndex uint `gorm:"type:integer;not null;column:log_index"`

	// 合约地址，EVM地址通常42位，但为了兼容其他链预留66位，加索引
	Address string `gorm:"type:varchar(100);not null;index;column:address"`

	Topic0 string `gorm:"type:varchar(66);index;column:topic0"`
	Topic1 string `gorm:"type:varchar(66);index;column:topic1"`
	Topic2 string `gorm:"type:varchar(66);index;column:topic2"`
	Topic3 string `gorm:"type:varchar(66);index;column:topic3"`

	Data string `gorm:"type:text;column:data"`

	Status uint8 `gorm:"type:smallint;default:0;not null;column:status;comment:0-pending,1-processed,2-failed"`

	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`

	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func (EventRecord) TableName() string {
	return "event_records"
}
