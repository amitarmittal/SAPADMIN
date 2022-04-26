package commondto

type ListClearedOrdersReqDto struct {
	BetIds    []string `json:"betIds"`
	BetStatus string   `json:"betStatus"`
}
