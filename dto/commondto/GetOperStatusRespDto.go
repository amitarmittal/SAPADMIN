package commondto

type OperatorStatus struct {
	OperatorName string `json:"operatorName"`
	OperatorId   string `json:"OperatorId"`
	PartnerId    string `json:"partnerId"`
	IsActive     string `json:"isActive"`
}

type GetOperDetailRespDto struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	OperatorStatus   []OperatorStatus `json:"providers"`
}
