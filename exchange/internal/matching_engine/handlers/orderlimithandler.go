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
	var result []*dto.OrderResult
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
			tempResult, isSelfTrade, b1, err := SelfTradeHandler(ob, taker, maker)
			if err != nil {
				return nil, err
			}
			if isSelfTrade {
				result = append(result, tempResult...)
				if !b1 {
					return result, nil
				}
				continue
			}

			//正常交易,返回true，taker结束撮合，返回false继续下个maker撮合
			isComplete, matchResult := match(ob, taker, maker)
			result = append(result, matchResult...)
			if isComplete {
				return result, nil
			}
		}
	}
	//taker 剩余的加入order book，成为maker
	unfillAmt, _ := decimal.NewFromString(taker.UnfilledAmount)
	if !(taker.State == dto.OrderState_ORDER_STATE_CANCELED || taker.State == dto.OrderState_ORDER_STATE_PARTIAL_CANCELED) && unfillAmt.GreaterThan(decimal.Zero) {
		tempOb := obFunc(taker.Side)
		tempOb.Add(taker)
	}
	return result, nil
}
