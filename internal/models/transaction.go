package models

import "time"

type TransactionRecord struct {
	Id          uint64
	ChainName   string
	BlockNumber uint64
	BlockHash   string
	TxHash      string
	TxIndex     uint
	FromAddress string
	ToAddress   string
	Value       string
	GasUsed     uint64
	GasPrice    string
	RawStatus   uint8
	Status      uint8
	Timestamp   uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}
