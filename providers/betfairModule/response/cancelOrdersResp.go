package response

type CancelInstruction struct {
	BetId         string  `json:"betId"`
	SizeReduction float64 `json:"sizeReduction"`
}

type CancelInstructionReport struct {
	Status        string            `json:"status"`
	ErrorCode     string            `json:"errorCode"`
	Instruction   CancelInstruction `json:"instruction"`
	SizeCancelled float64           `json:"sizeCancelled"`
	CancelledDate string            `json:"cancelledDate"`
}

type CancelOrdersResp struct {
	CustomerRef        string                    `json:"customerRef"`
	Status             string                    `json:"status"`
	ErrorCode          string                    `json:"errorCode"`
	MarketId           string                    `json:"marketId"`
	InstructionReports []CancelInstructionReport `json:"instructionReports"`
}

type BFCancelOrdersResp struct {
	Status           string           `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription string           `json:"message"` // Error message
	CancelOrdersResp CancelOrdersResp `json:"cancelOrdersResp"`
}
