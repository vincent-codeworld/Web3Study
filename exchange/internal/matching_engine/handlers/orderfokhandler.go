package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"
	"strconv"

	"github.com/shopspring/decimal"
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
	iterator := ob.Iterator()
	takerPrice, _ := strconv.ParseFloat(taker.Price, 64)
	totalVolume := decimal.Zero
	for iterator.Next() {
		pl := iterator.Value().(*dto.PriceLevel)
		subPrice := takerPrice - pl.Price
		//先判断是否存在maker符合交易条件，不符合直接返回
		if (taker.Side == dto.Side_SIDE_SELL && subPrice < 0) ||
			(taker.Side == dto.Side_SIDE_BUY && subPrice > 0) {
			return false
		}
		// 通过selfTradeHandler 判断

		// 统计成交金额，能满足订单全部金额则成交，否则撤销订单
	}
	return false
}

func fokSelfTrade(taker, maker *dto.Order) (bool, decimal.Decimal) {
	switch taker.Stp {
	// 正常的撮合逻辑
	case dto.SelfTradeWMType_STP_AST:
		return true, decimal.Decimal{}
	//DC类型是不产生标准成交记录
	case dto.SelfTradeWMType_STP_DC:
		takerAmt, _ := decimal.NewFromString(taker.GetUnfilledAmount())
		makerAmt, _ := decimal.NewFromString(maker.GetUnfilledAmount())
		if takerAmt.GreaterThan(makerAmt) {
			takerAmt = takerAmt.Sub(makerAmt)
			taker.UnfilledAmount = takerAmt.String()

			return true, decimal.Decimal{}
		} else if takerAmt.Equal(makerAmt) {
			return true, decimal.Decimal{}
		}
		makerAmt = makerAmt.Sub(takerAmt)
		maker.UnfilledAmount = makerAmt.String()
		return true, decimal.Decimal{}
	case dto.SelfTradeWMType_STP_CO:

		return true, decimal.Decimal{}
	case dto.SelfTradeWMType_STP_CN:

		return false, decimal.Decimal{}
	case dto.SelfTradeWMType_STP_CB:

		return false, decimal.Decimal{}
	default:
		return false, decimal.Decimal{}
	}
}
