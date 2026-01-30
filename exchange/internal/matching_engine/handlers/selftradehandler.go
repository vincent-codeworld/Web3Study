package handlers

import (
	"Web3Study/exchange/internal/dto"
)

func SelfTradeHandler(taker *dto.Order, maker *dto.Order) {
	switch taker.Stp {
	case dto.SelfTradeWMType_STP_AST:

	case dto.SelfTradeWMType_STP_DC:

	case dto.SelfTradeWMType_STP_CO:
		maker.State = dto.OrderState_ORDER_STATE_CANCELED
		if maker.State == dto.OrderState_ORDER_STATE_PARTIAL_FILLED {
			maker.State = dto.OrderState_ORDER_STATE_PARTIAL_CANCELED
		}
	case dto.SelfTradeWMType_STP_CN:
		taker.State = dto.OrderState_ORDER_STATE_CANCELED
	case dto.SelfTradeWMType_STP_CB:
		taker.State = dto.OrderState_ORDER_STATE_CANCELED
		maker.State = dto.OrderState_ORDER_STATE_CANCELED
		if maker.State == dto.OrderState_ORDER_STATE_PARTIAL_FILLED {
			maker.State = dto.OrderState_ORDER_STATE_PARTIAL_CANCELED
		}
	}
}
