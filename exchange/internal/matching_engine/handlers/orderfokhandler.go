package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"
	"strconv"
)

/*
*
FOK:全部成交或全部取消，不允许部分成交
*/
func FillOrKillHandler(taker *dto.Order, obFunc func(side dto.Side) *orderbook.OrderBook) ([]*dto.OrderResult, error) {

	makerSide := dto.Side_SIDE_BUY
	if taker.Side == dto.Side_SIDE_BUY {
		makerSide = dto.Side_SIDE_SELL
	}
	ob := obFunc(makerSide)
	var result []*dto.OrderResult
	iterator := ob.Iterator()
	takerPrice, _ := strconv.ParseFloat(taker.Price, 64)
	for iterator.Next() {
		pl := iterator.Value().(*dto.PriceLevel)
		subPrice := takerPrice - pl.Price
		//先判断是否存在maker符合交易条件，不符合直接返回
		if (taker.Side == dto.Side_SIDE_SELL && subPrice < 0) ||
			(taker.Side == dto.Side_SIDE_BUY && subPrice > 0) {
			return []*dto.OrderResult{
				{
					OrderId:      taker.OrderId,
					UserId:       taker.UserId,
					Price:        taker.Price,
					CancelReason: constant.CancelReasonFillOrKill,
				},
			}, nil
		}
		// 通过selfTradeHandler 判断

		// 统计成交金额，能满足订单全部金额则成交，否则撤销订单
	}
	return result, nil
}

func PreCalDepth(ob *orderbook.OrderBook, taker *dto.Order) bool {

	return false
}
