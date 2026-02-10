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
		State:        dto.OrderState_name[int32(dto.OrderState_ORDER_STATE_PARTIAL_FILLED)],
	}
	makerResult = &dto.OrderResult{
		OrderId:      maker.GetOrderId(),
		UserId:       maker.GetUserId(),
		Role:         constant.RoleMaker,
		Price:        closingPrice.String(),
		FilledAmount: fillAmt.String(),
		State:        dto.OrderState_name[int32(dto.OrderState_ORDER_STATE_PARTIAL_FILLED)],
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

	p, _ := ob.Get(maker.GetPrice())
	pl, _ := p.(dto.PriceLevel)
	totalVol, _ := decimal.NewFromString(pl.TotalVolume)
	pl.TotalVolume = totalVol.Sub(fillAmt).String()

	maker.UnfilledAmount = leftMuf.String()
	taker.UnfilledAmount = leftTuf.String()
	//如果maker 是ice berg订单，可见数量用完，需要随机抽取部分隐藏数量填充到可见数量，然后重新将maker放到队列最后
	if leftMuf.LessThanOrEqual(decimal.Zero) {
		orderHandle(ob, maker, constant.RoleMaker)
	}
	//taker 订单
	// 如果是taker是ice berg订单，可见数量用完，需要随机抽取部分隐藏数量填充到可见数量，然后继续撮合直到结束
	if leftTuf.LessThanOrEqual(decimal.Zero) {
		return orderHandle(ob, maker, constant.RoleTaker), results
	}
	return false, results
}

func orderHandle(ob *orderbook.OrderBook, order *dto.Order, orderRole string) bool {
	var b bool
	switch orderRole {
	case constant.RoleTaker:
		b = true
		if order.Type == dto.OrderType_ORDER_TYPE_ICEBERG {
			// 先固定在隐藏数量取amout 到可见数量，default：500
			hiddernQty, _ := decimal.NewFromString(order.GetHiddenQuantity())
			s := decimal.Min(hiddernQty, decimal.NewFromFloat(500))
			if !s.IsPositive() {
				return b
			}
			b = false
			order.HiddenQuantity = hiddernQty.Sub(s).String()
			order.UnfilledAmount = s.String()
		}
	case constant.RoleMaker:
		if order.Type == dto.OrderType_ORDER_TYPE_ICEBERG {
			// 先固定在隐藏数量取amout 到可见数量，default：500
			hiddernQty, _ := decimal.NewFromString(order.GetHiddenQuantity())
			if hiddernQty.LessThanOrEqual(decimal.Zero) {
				ob.Del(order.Price)
				return b
			}
			s := decimal.Min(hiddernQty, decimal.NewFromFloat(500))
			order.HiddenQuantity = hiddernQty.Sub(s).String()
			order.UnfilledAmount = s.String()
			//需要重新排到队列最后面
			p, _ := ob.Get(order.GetPrice())
			pl := p.(dto.PriceLevel)
			head := pl.Head
			tail := pl.Tail
			if head != tail {
				pl.Head = head.Next
				head.Next = nil
				pl.Head.Pre = nil

				pl.Tail = head
				head.Pre = tail
				tail.Next = head
			}
			totalVol, _ := decimal.NewFromString(pl.TotalVolume)
			pl.TotalVolume = totalVol.Add(s).String()
		} else {
			ob.Del(order.Price)
		}
	}
	return b
}
