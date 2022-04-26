package request

import (
	"Sp/dto/requestdto"
	"Sp/dto/sports"
	utils "Sp/utilities"
)

type CancelInstruction struct {
	BetId         string  `json:"betId"`
	SizeReduction float64 `json:"sizeReduction"`
}

type ListCancelOrdersReq struct {
	MarketId     string              `json:"marketId"`
	Instructions []CancelInstruction `json:"instructions"`
	CustomerRef  string              `json:"customerRef"`
}

func NewCancelOrderReq2(reqDto requestdto.CancelBetReqDto, betDtos []sports.BetDto, betFairRate int) ListCancelOrdersReq {
	// cancelOrderReq.CustomerRef = "" // ???
	cancelOrderReq := ListCancelOrdersReq{}
	cancelOrderReq.MarketId = reqDto.MarketId
	cancelOrderReq.Instructions = []CancelInstruction{}
	// Cancel Instructions
	for _, req := range reqDto.CancelBets {
		cancelInstruction := CancelInstruction{}
		for _, bet := range betDtos {
			if req.BetId == bet.BetId {
				cancelInstruction.BetId = bet.BetReq.BetId
				cancelInstruction.SizeReduction = getSizeReduction(req.CancelSize, bet, betFairRate) // zero for full cancillation
			}
		}
		cancelOrderReq.Instructions = append(cancelOrderReq.Instructions, cancelInstruction)
	}
	return cancelOrderReq
}

func getSizeReduction(inputSize float64, betDto sports.BetDto, betFairRate int) float64 {
	returnSize := inputSize * float64(betDto.BetReq.Rate)
	returnSize = returnSize * (100 - betDto.BetReq.OperatorHold) / 100
	returnSize = returnSize * (100 - betDto.BetReq.PlatformHold) / 100
	returnSize = utils.Truncate64(returnSize / float64(betFairRate))
	return returnSize
}
