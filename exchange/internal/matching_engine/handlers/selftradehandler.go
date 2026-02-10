package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"
	"fmt"

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
func selfTradeMatch(taker *dto.Order, maker *dto.Order) ([]*dto.OrderResult, error) {
	cancelOrder := func(o *dto.Order) {
		if o.State == dto.OrderState_ORDER_STATE_PARTIAL_FILLED {
			o.State = dto.OrderState_ORDER_STATE_PARTIAL_CANCELED
		} else {
			o.State = dto.OrderState_ORDER_STATE_CANCELED
		}
	}
	//设置取消原因，stp_deducted_quantity 字段
	switch taker.Stp {
	// 正常的撮合逻辑
	case dto.SelfTradeWMType_STP_AST:
		return nil, nil
	//DC类型是不产生标准成交记录
	case dto.SelfTradeWMType_STP_DC:
		takerAmt, _ := decimal.NewFromString(taker.GetUnfilledAmount())
		makerAmt, _ := decimal.NewFromString(maker.GetUnfilledAmount())
		if takerAmt.GreaterThan(makerAmt) {
			takerAmt = takerAmt.Sub(makerAmt)
			taker.UnfilledAmount = takerAmt.String()
			cancelOrder(maker)
			return []*dto.OrderResult{
				{
					OrderId:             maker.GetOrderId(),
					UserId:              maker.GetUserId(),
					Role:                constant.RoleMaker,
					Stp:                 taker.Stp,
					StpDeductedQuantity: maker.GetUnfilledAmount(),
					CancelReason:        constant.CancelReasonDeducted,
				},
				{
					OrderId:             taker.GetOrderId(),
					UserId:              taker.GetUserId(),
					Role:                constant.RoleTaker,
					Stp:                 taker.Stp,
					StpDeductedQuantity: maker.GetUnfilledAmount(),
				},
			}, nil
		} else if takerAmt.Equal(makerAmt) {
			cancelOrder(maker)
			cancelOrder(taker)
			return []*dto.OrderResult{
				{
					OrderId:             maker.GetOrderId(),
					UserId:              maker.GetUserId(),
					Role:                constant.RoleMaker,
					Stp:                 taker.Stp,
					StpDeductedQuantity: maker.GetUnfilledAmount(),
					CancelReason:        constant.CancelReasonDeducted,
				},
				{
					OrderId:             taker.GetOrderId(),
					UserId:              taker.GetUserId(),
					Role:                constant.RoleTaker,
					Stp:                 taker.Stp,
					StpDeductedQuantity: taker.GetUnfilledAmount(),
					CancelReason:        constant.CancelReasonDeducted,
				},
			}, nil
		}
		makerAmt = makerAmt.Sub(takerAmt)
		maker.UnfilledAmount = makerAmt.String()
		cancelOrder(taker)
		return []*dto.OrderResult{
			{
				OrderId:             maker.GetOrderId(),
				UserId:              maker.GetUserId(),
				Role:                constant.RoleMaker,
				Stp:                 taker.Stp,
				StpDeductedQuantity: taker.GetUnfilledAmount(),
			},
			{
				OrderId:             taker.GetOrderId(),
				UserId:              taker.GetUserId(),
				Role:                constant.RoleTaker,
				Stp:                 taker.Stp,
				StpDeductedQuantity: taker.GetUnfilledAmount(),
				CancelReason:        constant.CancelReasonCancelNew,
			},
		}, nil
	case dto.SelfTradeWMType_STP_CO:
		cancelOrder(maker)
		return []*dto.OrderResult{
			{
				OrderId:             maker.GetOrderId(),
				UserId:              maker.GetUserId(),
				Role:                constant.RoleMaker,
				Stp:                 taker.Stp,
				StpDeductedQuantity: maker.GetUnfilledAmount(),
				CancelReason:        constant.CancelReasonCancelOld,
			},
		}, nil
	case dto.SelfTradeWMType_STP_CN:
		cancelOrder(taker)
		return []*dto.OrderResult{
			{
				OrderId:             taker.GetOrderId(),
				UserId:              taker.GetUserId(),
				Role:                constant.RoleTaker,
				Stp:                 taker.Stp,
				StpDeductedQuantity: taker.GetUnfilledAmount(),
				CancelReason:        constant.CancelReasonCancelNew,
			},
		}, nil
	case dto.SelfTradeWMType_STP_CB:
		cancelOrder(maker)
		cancelOrder(taker)

		return []*dto.OrderResult{
			{
				OrderId:             maker.GetOrderId(),
				UserId:              maker.GetUserId(),
				Role:                constant.RoleMaker,
				Stp:                 taker.Stp,
				StpDeductedQuantity: maker.GetUnfilledAmount(),
				CancelReason:        constant.CancelReasonCancelBoth,
			},
			{
				OrderId:             taker.GetOrderId(),
				UserId:              taker.GetUserId(),
				Role:                constant.RoleTaker,
				Stp:                 taker.Stp,
				StpDeductedQuantity: taker.GetUnfilledAmount(),
				CancelReason:        constant.CancelReasonCancelBoth,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Trade Type %d", taker.Stp)
	}

}

func SelfTradeHandler(ob *orderbook.OrderBook, taker *dto.Order, maker *dto.Order) ([]*dto.OrderResult, bool, bool, error) {
	var result []*dto.OrderResult
	if taker.UserId == maker.UserId && taker.Stp != dto.SelfTradeWMType_STP_AST {
		if taker.Type == dto.OrderType_ORDER_TYPE_ICEBERG && taker.Stp == dto.SelfTradeWMType_STP_DC {
			return nil, false, false, nil
		}
		b := true
		selfTrade, err := selfTradeMatch(taker, maker)
		if err != nil {
			return nil, true, false, err
		}
		result = append(result, selfTrade...)
		if maker.State == dto.OrderState_ORDER_STATE_CANCELED || maker.State == dto.OrderState_ORDER_STATE_PARTIAL_CANCELED {
			ob.Del(maker.Price)
		}
		if taker.State == dto.OrderState_ORDER_STATE_CANCELED || taker.State == dto.OrderState_ORDER_STATE_PARTIAL_CANCELED {
			b = false
		}
		return result, true, b, nil
	}
	return result, false, false, nil
}
