package requestdto

type GetAllBetsReqDto struct {
	Token       string `json:"token"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	ReferenceId string `json:"referenceId"`
}
