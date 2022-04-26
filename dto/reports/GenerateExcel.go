package reports

type GenerateExcelRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	Excel            []byte `json:"excel"`
}

type GenerateReqDto struct {
	Excel interface{} `json:"excel"`
}
