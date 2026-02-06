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
	if !PreCalDepth(ob, taker) {
		return []*dto.OrderResult{
			{
				OrderId:      taker.OrderId,
				UserId:       taker.UserId,
				Role:         constant.RoleTaker,
				Stp:          taker.Stp,
				CancelReason: constant.CancelReasonFillOrKill,
			},
		}, nil
	}

	// 真正执行

	return result, nil
}

func PreCalDepth(ob *orderbook.OrderBook, taker *dto.Order) bool {
	iterator := ob.Iterator()
	takerPrice, _ := strconv.ParseFloat(taker.Price, 64)
	takerAmt, _ := decimal.NewFromString(taker.UnfilledAmount)
	totalVolume := decimal.Zero
	for iterator.Next() {
		pl := iterator.Value().(*dto.PriceLevel)
		subPrice := takerPrice - pl.Price
		//先判断是否存在maker符合交易条件，不符合直接返回
		if (taker.Side == dto.Side_SIDE_SELL && subPrice < 0) ||
			(taker.Side == dto.Side_SIDE_BUY && subPrice > 0) {
			return false
		}
		maker := pl.Head
		for {
			if maker == nil {
				break
			}
			// 通过selfTradeHandler 判断
			isPass, preFillAmt := fokSelfTradeCheck(taker, maker)
			if !isPass {
				return false
			}
			totalVolume = totalVolume.Add(preFillAmt)
			if totalVolume.GreaterThanOrEqual(takerAmt) {
				return true
			}
			maker = maker.Next
		}
		// 统计成交金额，能满足订单全部金额则成交，否则撤销订单
		if totalVolume.GreaterThanOrEqual(takerAmt) {
			return true
		}
	}
	return false
}

func fokSelfTradeCheck(taker, maker *dto.Order) (bool, decimal.Decimal) {
	switch taker.Stp {
	case dto.SelfTradeWMType_STP_AST:
		takerAmt, _ := decimal.NewFromString(taker.GetUnfilledAmount())
		makerAmt, _ := decimal.NewFromString(maker.GetUnfilledAmount())
		fillAmt := decimal.Min(takerAmt, makerAmt)
		return true, fillAmt
	case dto.SelfTradeWMType_STP_DC, dto.SelfTradeWMType_STP_CO:
		return true, decimal.Zero
	case dto.SelfTradeWMType_STP_CN, dto.SelfTradeWMType_STP_CB:
		return false, decimal.Zero
	default:
		return false, decimal.Zero
	}
}

func fokSelfTradeHandler(ob *orderbook.OrderBook, taker, maker *dto.Order) ([]*dto.OrderResult, decimal.Decimal) {
	switch taker.Stp {
	case dto.SelfTradeWMType_STP_AST:
		takerAmt, _ := decimal.NewFromString(taker.GetUnfilledAmount())
		makerAmt, _ := decimal.NewFromString(maker.GetUnfilledAmount())
		fillAmt := decimal.Min(takerAmt, makerAmt)
		return nil, fillAmt
	case dto.SelfTradeWMType_STP_DC, dto.SelfTradeWMType_STP_CO:
		return nil, decimal.Zero
	case dto.SelfTradeWMType_STP_CN, dto.SelfTradeWMType_STP_CB:
		return nil, decimal.Zero
	default:
		return nil, decimal.Zero
	}
}
