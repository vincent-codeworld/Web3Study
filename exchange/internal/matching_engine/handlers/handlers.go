package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"

	"github.com/shopspring/decimal"
)

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
