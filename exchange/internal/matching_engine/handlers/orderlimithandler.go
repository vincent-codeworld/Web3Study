package handlers

import (
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"
)

func LimitHandler(taker *dto.Order, ob *orderbook.OrderBook) ([]*dto.OrderResult, error) {
	iterator := ob.Iterator()

	for iterator.Next() {
		pl := iterator.Value()
		priceLevel := pl.(*dto.PriceLevel)
		for {

			break
		}
	}
}
