package commondto

type CurrentOrdersRespDto struct {
	CurrentOrders []struct {
		BetID       string  `json:"betId"`
		MarketID    string  `json:"marketId"`
		SelectionID int     `json:"selectionId"`
		Handicap    float64 `json:"handicap"`
		PriceSize   struct {
			Price float64 `json:"price"`
			Size  float64 `json:"size"`
		} `json:"priceSize"`
		BspLiability        float64 `json:"bspLiability"`
		Side                string  `json:"side"`
		Status              string  `json:"status"`
		PersistenceType     string  `json:"persistenceType"`
		OrderType           string  `json:"orderType"`
		PlacedDate          string  `json:"placedDate"`
		MatchedDate         string  `json:"matchedDate"`
		AveragePriceMatched float64 `json:"averagePriceMatched"`
		SizeMatched         float64 `json:"sizeMatched"`
		SizeRemaining       float64 `json:"sizeRemaining"`
		SizeLapsed          float64 `json:"sizeLapsed"`
		SizeCancelled       float64 `json:"sizeCancelled"`
		SizeVoided          float64 `json:"sizeVoided"`
		RegulatorAuthCode   string  `json:"regulatorAuthCode"`
		RegulatorCode       string  `json:"regulatorCode"`
		CustomerOrderRef    string  `json:"customerOrderRef"`
		CustomerStrategyRef string  `json:"customerStrategyRef"`
		//CurrentItemDescription interface{} `json:"currentItemDescription"`
	} `json:"currentOrders"`
	MoreAvailable bool `json:"moreAvailable"`
}

type BFCurrentOrdersResp struct {
	Status            string               `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription  string               `json:"message"` // Error message
	CurrentOrdersResp CurrentOrdersRespDto `json:"data"`
}
