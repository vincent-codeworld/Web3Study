package dto

import "github.com/shopspring/decimal"

type Side uint8

const (
	SideUnknown Side = iota
	BUY
	SELL
)

type OrderType int

// 0撤单申请， 1市价买入，2市价卖出，3限价买入，4限价卖出 ，5立即成交否则取消未成交的剩余部分（IOC）买入，6立即成交否则取消未成交的剩余部分（IOC）卖出，7全部成交，否则全部取消（FOK）买入，8全部成交，否则全部取消（FOK）卖出，9只做Maker买入，10只做Maker卖出，11取消前有效（GTC）买入，12取消前有效（GTC）卖出 13 批量撤单
const (
	OrderTypeUnknown OrderType = iota
	Limit
	Market
	PostOnly
	Fok
	Ioc
	Aon
	Iceberg
	Cancel
	SystemCancel
)

type SelfTradeWMType int8

const (
	AST = iota
	DC
	CO
	CN
	CB
)

type OrderState int

const (
	OrderStateUnknown OrderState = iota
	Pending
	Accepted
	Filled
	PartialFilled
	Canceled
	PartialCanceled
	Failed
	Error
)

type Order struct {
	Id             int64
	UserId         int64
	OrderId        int64
	Side           Side
	Type           OrderType
	State          OrderState //  初始化的时候都是Submitted
	Price          decimal.Decimal
	UnfilledAmount decimal.Decimal
	CircuitRate    decimal.Decimal // 保护比例 ，市价单的会触发
	CreateAt       int64
	Stp            SelfTradeWMType
	PullTime       int64
	Extra          string // 批量撤单的订单ID集合
	Taker          string
	Maker          string
}
