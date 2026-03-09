package orderbook

import (
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/utils"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	*treemap.Map
	ringBuffer *utils.RingBuffer[dto.OrderEvent]
}

// -1 降序 1升序
func NewOrderBook(sort int, rb *utils.RingBuffer[dto.OrderEvent]) *OrderBook {
	return &OrderBook{Map: treemap.NewWith(func(a, b interface{}) int {
		aLevel := a.(*dto.PriceLevel)
		bLevel := b.(*dto.PriceLevel)
		if aLevel.Price > bLevel.Price {
			return sort
		}
		return -sort
	}),
		ringBuffer: rb,
	}
}

func (ob *OrderBook) Add(order *dto.Order) {
	price, _ := strconv.ParseFloat(order.Price, 64)
	orderVol, _ := decimal.NewFromString(order.GetUnfilledAmount())
	if priceLevel, found := ob.Get(price); !found {
		pl := new(dto.PriceLevel)
		pl.Price = price
		pl.Head = order
		pl.Tail = order
		order.Parent = pl
		ob.Put(price, pl)
		pl.TotalVolume = orderVol.String()
	} else {
		pl := priceLevel.(*dto.PriceLevel)
		tailOrder := pl.Tail
		tailOrder.Next = order
		order.Pre = tailOrder
		pl.Tail = order
		totalVolume, _ := decimal.NewFromString(pl.TotalVolume)
		pl.TotalVolume = totalVolume.Add(orderVol).String()
	}
	event := dto.OrderEvent{
		Op:         dto.Operate_OpAdd,
		Price:      order.Price,
		OrderId:    order.OrderId,
		UserId:     order.UserId,
		Amt:        order.UnfilledAmount,
		OrderState: order.State,
		Side:       order.Side,
	}
	ob.ringBuffer.Put(event)
}

func (ob *OrderBook) Del(price string) {
	value, found := ob.Get(price)
	if !found {
		return
	}
	pl := value.(*dto.PriceLevel)
	o := pl.Head
	totalVolume, _ := decimal.NewFromString(pl.TotalVolume)
	orderVolume, _ := decimal.NewFromString(o.UnfilledAmount)
	pl.TotalVolume = totalVolume.Sub(orderVolume).String()
	next := o.Next
	if next == nil {
		ob.Remove(price)
		return
	}
	pl.Head = next
	next.Pre = nil
	event := dto.OrderEvent{
		Op:    dto.Operate_OpDel,
		Price: price,
	}
	ob.ringBuffer.Put(event)
}

func (ob *OrderBook) ModifyMaker(order *dto.Order) {
	event := dto.OrderEvent{
		Op:         dto.Operate_OpMod,
		Price:      order.Price,
		OrderId:    order.OrderId,
		Side:       order.Side,
		Amt:        order.UnfilledAmount,
		OrderState: order.State,
	}
	ob.ringBuffer.Put(event)
}

func (ob *OrderBook) ModifyMakerWithPriority(order *dto.Order, priority dto.Priority) {
	event := dto.OrderEvent{
		Op:         dto.Operate_OpMod,
		Price:      order.Price,
		OrderId:    order.OrderId,
		Side:       order.Side,
		Amt:        order.UnfilledAmount,
		OrderState: order.State,
		Priority:   priority,
	}
	ob.ringBuffer.Put(event)
}
