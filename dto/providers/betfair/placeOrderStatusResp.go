package betfair

type PlaceOrderStatusResp struct {
	Status                  string                   `json:"status"`
	ErrorDescription        string                   `json:"message"`
	PlaceInstructionReports []PlaceInstructionReport `json:"data"`
}
