package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"

	"github.com/shopspring/decimal"
)

func MarketHandler(taker *dto.Order, obFunc func(side dto.Side) *orderbook.OrderBook) ([]*dto.OrderResult, error) {
	makerSide := dto.Side_SIDE_BUY
	if taker.Side == dto.Side_SIDE_BUY {
		makerSide = dto.Side_SIDE_SELL
	}
	ob := obFunc(makerSide)
	iterator := ob.Iterator()
	var result []*dto.OrderResult
	for iterator.Next() {
		pl := iterator.Value()
		priceLevel := pl.(*dto.PriceLevel)
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
			b2, matchResult := match(ob, taker, maker)
			result = append(result, matchResult...)
			if b2 {
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

/*
*

	market撮合马上成交
*/
func match(ob *orderbook.OrderBook, taker, maker *dto.Order) (bool, []*dto.OrderResult) {
	takerUnfillAmt, _ := decimal.NewFromString(taker.UnfilledAmount)
	makerUnfillAmt, _ := decimal.NewFromString(maker.UnfilledAmount)
	takerPrice, _ := decimal.NewFromString(taker.Price)
	makerPrice, _ := decimal.NewFromString(maker.Price)
	closingPrice := decimal.Min(takerPrice, makerPrice)
	fillAmt := decimal.Min(takerUnfillAmt, makerUnfillAmt)

	var takerResult, makerResult *dto.OrderResult
	takerResult = &dto.OrderResult{
		OrderId:      taker.GetOrderId(),
		UserId:       taker.GetUserId(),
		Role:         constant.RoleTaker,
		Price:        closingPrice.String(),
		FilledAmount: fillAmt.String(),
		State:        dto.OrderState_name[int32(dto.OrderState_ORDER_STATE_FILLED)],
	}
	makerResult = &dto.OrderResult{
		OrderId:      maker.GetOrderId(),
		UserId:       maker.GetUserId(),
		Role:         constant.RoleMaker,
		Price:        closingPrice.String(),
		FilledAmount: fillAmt.String(),
		State:        dto.OrderState_name[int32(dto.OrderState_ORDER_STATE_FILLED)],
	}
	if !takerUnfillAmt.Equal(fillAmt) {
		takerResult.State = dto.OrderState_name[int32(dto.OrderState_ORDER_STATE_PARTIAL_FILLED)]
	}

	if !makerUnfillAmt.Equal(fillAmt) {
		makerResult.State = dto.OrderState_name[int32(dto.OrderState_ORDER_STATE_PARTIAL_FILLED)]
	}

	var results []*dto.OrderResult
	results = append(results, takerResult, makerResult)
	leftMuf := makerUnfillAmt.Sub(fillAmt)
	leftTuf := takerUnfillAmt.Sub(fillAmt)

	maker.UnfilledAmount = leftMuf.String()
	taker.UnfilledAmount = leftTuf.String()

	if leftMuf.LessThanOrEqual(decimal.Zero) {
		//remove maker from order book
		ob.Del(maker.Price)
	}
	//taker 订单
	if leftTuf.LessThanOrEqual(decimal.Zero) {
		return true, results
	}
	return false, results
}
