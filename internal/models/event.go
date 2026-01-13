package models

import "time"

type EventRecord struct {
	Id          uint64
	TxId        uint64
	ChainName   string
	BlockNumber uint64
	LogIndex    uint
	Address     string
	Topic0      string
	Topic1      string
	Topic2      string
	Topic3      string
	Data        string
	Status      uint8
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}
