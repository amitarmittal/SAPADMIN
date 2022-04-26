package betfair

type CancelInstructionReport struct {
	Status        string            `json:"status"`
	ErrorCode     string            `json:"errorCode"`
	Instruction   CancelInstruction `json:"instruction"`
	SizeCancelled float64           `json:"sizeCancelled"`
	CancelledDate int64             `json:"cancelledDate"`
}

type CancelOrderResp struct {
	CustomerRef        string                    `json:"customerRef"`
	Status             string                    `json:"status"`
	ErrorCode          string                    `json:"errorCode"`
	MarketId           string                    `json:"marketId"`
	InstructionReports []CancelInstructionReport `json:"instructionReports"`
}

type BFCancelOrderResp struct {
	Status           string          `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription string          `json:"message"` // Error message
	CancelOrderResp  CancelOrderResp `json:"data"`
}
