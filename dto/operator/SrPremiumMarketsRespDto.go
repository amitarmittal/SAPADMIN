package operatordto

type SrPremiumMarketsRespDto struct {
	AwayScore        int64   `json:"awayScore"`
	CompetitionID    string  `json:"competitionId"`
	CompetitionName  string  `json:"competitionName"`
	EventID          string  `json:"eventId"`
	EventName        string  `json:"eventName"`
	HomeScore        int64   `json:"homeScore"`
	MarketID         string  `json:"marketId"`
	Markets          Markets `json:"markets"`
	OpenDate         string  `json:"openDate"`
	ProducerID       int64   `json:"producerId"`
	ProviderName     string  `json:"providerName"`
	SportID          string  `json:"sportId"`
	SportName        string  `json:"sportName"`
	EventStatus      string  `json:"eventStatus"`
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
}

type Markets struct {
	Bookmakers   []Bookmaker   `json:"bookmakers"`
	FancyMarkets []FancyMarket `json:"fancyMarkets"`
	MatchOdds    []Bookmaker   `json:"matchOdds"`
}

type Bookmaker struct {
	Groups     []string `json:"groups"`
	MarketID   string   `json:"marketId"`
	MarketName string   `json:"marketName"`
	MarketType string   `json:"marketType"`
	Runners    []Runner `json:"runners"`
	Status     string   `json:"status"`
}

type Runner struct {
	BackPrices []Price `json:"backPrices"`
	LayPrices  []Price `json:"layPrices"`
	RunnerID   string  `json:"runnerId"`
	RunnerName string  `json:"runnerName"`
	Status     string  `json:"status"`
}

type Price struct {
	Price int64 `json:"price"`
}

type FancyMarket struct {
	Category   string `json:"category"`
	MarketID   string `json:"marketId"`
	MarketName string `json:"marketName"`
	MarketType string `json:"marketType"`
	NoRate     int64  `json:"noRate"`
	NoValue    int64  `json:"noValue"`
	Status     string `json:"status"`
	YesRate    int64  `json:"yesRate"`
	YesValue   int64  `json:"yesValue"`
}
