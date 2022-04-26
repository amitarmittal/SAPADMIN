package request

type ListClearedOrdersReq struct {
	BetIds    []string `json:"betIds"`
	BetStatus string   `json:"betStatus"`
}
