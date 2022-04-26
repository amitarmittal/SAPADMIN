package requestdto

type PlaceBetReqDto struct {
	Token          string  `json:"token"`
	OperatorId     string  `json:"operatorId"`
	PartnerId      string  `json:"partnerId"`
	ProviderId     string  `json:"providerId"` // DREAM/BET_FAIR/SPORTS_RADAR
	SportId        string  `json:"sportId"`    // CRICKET/SOCCER/TENNIS
	CompetitionId  string  `json:"competitionId"`
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
	IsUnmatched    bool    `json:"isUnmatched"`    // For BetFair. Default is false, means FILL_OR_KILL.
	UserAgent      string  `json:"userAgent"`      // Mobile or Internet
}
