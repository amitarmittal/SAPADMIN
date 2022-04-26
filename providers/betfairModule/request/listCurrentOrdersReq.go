package request

type ListCurrentOrdersReq struct {
	BetIds            []string `json:"betIds"`
	MarketIds         []string `json:"marketIds"`
	CustomerOrderRefs []string `json:"customerOrderRefs"`
	FromRecord        int      `json:"fromRecord"`
	RecordCount       int      `json:"recordCount"`
}
