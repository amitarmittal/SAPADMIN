package commondto

type ListCurrentOrdersDto struct {
	BetIds []string `json:"betIds"`
	FromRecord int `json:"fromRecord"`
	RecordCount int `json:"recordCount"`
}
