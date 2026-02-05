package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"

	"github.com/shopspring/decimal"
)

func PostOnlyHandler(taker *dto.Order, obFunc func(side dto.Side) *orderbook.OrderBook) ([]*dto.OrderResult, error) {
	makerSide := dto.Side_SIDE_BUY
	if taker.Side == dto.Side_SIDE_BUY {
		makerSide = dto.Side_SIDE_SELL
	}
	ob := obFunc(makerSide)
	iterator := ob.Iterator()
	if iterator.Next() {
		pl := iterator.Value()
		priceLevel := pl.(*dto.PriceLevel)
		head := priceLevel.Head
		makerPrice, _ := decimal.NewFromString(head.Price)
		takerPrice, _ := decimal.NewFromString(taker.Price)
		sub := makerPrice.Sub(takerPrice)
		if (makerSide == dto.Side_SIDE_BUY && !sub.IsNegative()) || (makerSide == dto.Side_SIDE_SELL && !sub.IsPositive()) {
			return []*dto.OrderResult{
				{
					OrderId:      taker.OrderId,
					UserId:       taker.UserId,
					Price:        taker.Price,
					Role:         taker.Taker,
					CancelReason: constant.CancelReasonPostOnly,
				},
			}, nil
		}
	}

	// 成为 maker
	ob.Add(taker)
	return nil, nil
}
