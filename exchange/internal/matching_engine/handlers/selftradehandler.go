package handlers

import (
	"Web3Study/exchange/internal/dto"

	"github.com/shopspring/decimal"
)

/*
*

	        AST: Allow Self-Trade，允许自成交
		    DC:Decrement and Cancel，减量成交
			CO:cancel old，取消旧订单
			CN:cancel new，取消新订单
			CB:cancel both，maker跟taker都取消
*/
func SelfTradeHandler(taker *dto.Order, maker *dto.Order) {
	cancelOrder := func(o *dto.Order) {
		if o.State == dto.OrderState_ORDER_STATE_PARTIAL_FILLED {
			o.State = dto.OrderState_ORDER_STATE_PARTIAL_CANCELED
		} else {
			o.State = dto.OrderState_ORDER_STATE_CANCELED
		}
	}

	switch taker.Stp {
	case dto.SelfTradeWMType_STP_AST:

	case dto.SelfTradeWMType_STP_DC:
		takerAmt, _ := decimal.NewFromString(taker.GetUnfilledAmount())
		makerAmt, _ := decimal.NewFromString(maker.GetUnfilledAmount())
		if takerAmt.GreaterThan(makerAmt) {
			takerAmt = takerAmt.Sub(makerAmt)
			taker.UnfilledAmount = takerAmt.String()
			cancelOrder(maker)
		} else if takerAmt.Equal(makerAmt) {
			cancelOrder(maker)
			cancelOrder(taker)
		} else {
			makerAmt = makerAmt.Sub(takerAmt)
			maker.UnfilledAmount = makerAmt.String()
			cancelOrder(taker)
		}
	case dto.SelfTradeWMType_STP_CO:
		cancelOrder(maker)
	case dto.SelfTradeWMType_STP_CN:
		cancelOrder(taker)
	case dto.SelfTradeWMType_STP_CB:
		cancelOrder(maker)
		cancelOrder(taker)
	}
}
