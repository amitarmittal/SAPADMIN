package response

import "Sp/providers/betfairModule/request"

type PlaceInstructionReport struct {
	Status              string                   `json:"status"`
	ErrorCode           string                   `json:"errorCode"`
	OrderStatus         string                   `json:"orderStatus"`
	Instruction         request.PlaceInstruction `json:"instruction"`
	BetId               string                   `json:"betId"`
	PlacedDate          string                   `json:"placedDate"`
	AveragePriceMatched float64                  `json:"averagePriceMatched"`
	SizeMatched         float64                  `json:"sizeMatched"`
}

type PlaceOrderResp struct {
	CustomerRef        string                   `json:"customerRef"`
	Status             string                   `json:"status"`
	ErrorCode          string                   `json:"errorCode"`
	MarketId           string                   `json:"marketId"`
	InstructionReports []PlaceInstructionReport `json:"instructionReports"`
}

type PlaceOrderAsyncResp struct {
	Status           string `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription string `json:"message"` // Error message
}

type BFPlaceOrderResp struct {
	Status           string         `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription string         `json:"message"` // Error message
	PlaceOrderResp   PlaceOrderResp `json:"data"`
}
