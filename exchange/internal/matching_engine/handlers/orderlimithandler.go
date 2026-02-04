package handlers

import (
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"

	"github.com/shopspring/decimal"
)

/*
*

	limit限价单是需要判断是否存在对手订单，如果存在在判断selftrade策略
*/
func LimitHandler(taker *dto.Order, obFunc func(side dto.Side) *orderbook.OrderBook) ([]*dto.OrderResult, error) {
	makerSide := dto.Side_SIDE_BUY
	if taker.Side == dto.Side_SIDE_BUY {
		makerSide = dto.Side_SIDE_SELL
	}
	ob := obFunc(makerSide)
	takerPrice, _ := decimal.NewFromString(taker.Price)
	iterator := ob.Iterator()
	for iterator.Next() {
		pl := iterator.Value()
		priceLevel := pl.(*dto.PriceLevel)

		makerPrice := decimal.NewFromFloat(priceLevel.Price)
		subPrice := makerPrice.Sub(takerPrice)
		if (taker.Side == dto.Side_SIDE_SELL && subPrice.IsNegative()) ||
			(taker.Side == dto.Side_SIDE_BUY && subPrice.IsPositive()) {
			break
		}

		for {
			maker := priceLevel.Head
			if maker == nil {
				break
			}
		}
	}
}
