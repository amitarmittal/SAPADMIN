package sportradar

type PlaceOrderResp struct {
	Status           string  `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription string  `json:"message"` // Error message
	AltStake         float64 `json:"altStake"`
}
