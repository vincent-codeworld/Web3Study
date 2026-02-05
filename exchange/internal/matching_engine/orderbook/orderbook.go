package orderbook

import (
	"Web3Study/exchange/internal/dto"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	*treemap.Map
}

// -1 降序 1升序
func NewOrderBook(sort int) *OrderBook {
	return &OrderBook{treemap.NewWith(func(a, b interface{}) int {
		aLevel := a.(*dto.PriceLevel)
		bLevel := b.(*dto.PriceLevel)
		if aLevel.Price > bLevel.Price {
			return sort
		}
		return -sort
	})}
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
}
