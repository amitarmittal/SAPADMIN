package dto

type SportsBetReqDto struct {
	Token          string  `json:"token"`
	OperatorId     string  `json:"operatorId"`
	ProviderId     string  `json:"providerId"` // DREAM/BET_FAIR/SPORTS_RADAR
	SportId        string  `json:"sportId"`    // CRICKET/SOCCER/TENNIS
	CompetetionId  string  `json:"competetionId"`
	EventId        string  `json:"eventId"`
	BetType        string  `json:"betType"` // BACK/LAY
	OddValue       float64 `json:"oddValue"`
	StakeAmount    float64 `json:"stakeAmount"`
	MarketType     string  `json:"marketType"` // MATCH_ODDS/BOOK_MAKER/GOAL_MARKETS/FANCY_MARKET/
	MarketName     string  `json:"marketName"` // for soccer, match_odds or goal_markets, can be avoided
	MarketId       string  `json:"marketId"`
	RunnerId       string  `json:"runnerId"`       // ?? what is for fancy
	RunnerName     string  `json:"runnerName"`     // ?? what is for fancy
	SessionOutcome float64 `json:"sessionOutcome"` // fancy scrore ex: 45 NO, 46 YES
}
