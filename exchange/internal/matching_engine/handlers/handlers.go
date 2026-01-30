package handlers

import (
	"Web3Study/exchange/internal/dto"
	"Web3Study/exchange/internal/matching_engine"
)

type PreMatchHandler = func(order *dto.Order)

type MatchHandler = func(engine *matching_engine.MatchEngine, order *dto.Order) *dto.OrderResult
