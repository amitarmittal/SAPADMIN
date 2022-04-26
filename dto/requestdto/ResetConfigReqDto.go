package requestdto

//GetSportsReqDto
type ResetConfigReqDto struct {
	OperatorId string `json:"operatorId"` // mandatory for feed service, optional for portal service
}
