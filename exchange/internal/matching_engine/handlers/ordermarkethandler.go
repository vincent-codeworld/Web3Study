package handlers

import (
	"Web3Study/exchange/internal/constant"
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine/orderbook"

	"github.com/shopspring/decimal"
)

func MarketHandler(taker *dto.Order, ob *orderbook.OrderBook) ([]*dto.OrderResult, error) {
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
			if taker.UserId == maker.UserId {
				selfTrade, err := selfTradeHandler(taker, maker)
				if err != nil {
					return nil, err
				}
				result = append(result, selfTrade...)
				if taker.Stp == dto.SelfTradeWMType_STP_CO {
					ob.Del(maker.Price)
					continue
				} else if taker.Stp == dto.SelfTradeWMType_STP_CN {
					return result, nil
				} else if taker.Stp == dto.SelfTradeWMType_STP_CB {
					ob.Del(maker.Price)
					return result, nil
				} else if taker.Stp == dto.SelfTradeWMType_STP_DC {
					isRemove := false
					if maker.State == dto.OrderState_ORDER_STATE_PARTIAL_CANCELED || maker.State == dto.OrderState_ORDER_STATE_CANCELED {
						nt := maker.Next
						if nt == nil {
							ob.Remove(priceLevel.Price)
							isRemove = true
						} else {
							nt.Pre = nil
							priceLevel.Head = nt
							maker = priceLevel.Head
						}
					}

					if taker.State == dto.OrderState_ORDER_STATE_PARTIAL_CANCELED || taker.State == dto.OrderState_ORDER_STATE_CANCELED {
						return result, nil
					}

					if isRemove {
						break
					}

					continue
				}

			}

			//正常交易,返回true，taker结束撮合，返回false继续下个maker撮合
			b, matchResult := match(ob, taker, maker)
			result = append(result, matchResult...)
			if b {
				return result, nil
			}
		}
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
