package handlers

import (
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"
	"strconv"
)

func MarketHandler(taker *dto.Order, ob *orderbook.OrderBook) ([]*dto.OrderResult, error) {
	takerPrice, _ := strconv.ParseFloat(taker.Price, 64)
	iterator := ob.Iterator()
	_, pl := ob.Min()
	priceLevel := pl.(*dto.PriceLevel)
	maker := priceLevel.Head
	if taker.UserId == maker.UserId {
		selfTrade, err := SelfTradeHandler(taker, maker)
		if err != nil {
			return nil, err
		}
		if taker.Stp != dto.SelfTradeWMType_STP_AST {

		} else {

		}

	}
}
