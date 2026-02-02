package orderbook

import (
	"Web3Study/exchange/internal/dto"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"
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
	if priceLevel, found := ob.Get(price); !found {
		d := new(dto.PriceLevel)
		d.Price = price
		d.Head = order
		d.Tail = order
		order.Parent = d
		ob.Put(price, d)
	} else {
		pl := priceLevel.(*dto.PriceLevel)
		tailOrder := pl.Tail
		tailOrder.Next = order
		order.Pre = tailOrder
		pl.Tail = order
		orderVol, _ := strconv.ParseFloat(order.GetUnfilledAmount(), 64)
		pl.TotalVolume = pl.TotalVolume + orderVol
	}
}
