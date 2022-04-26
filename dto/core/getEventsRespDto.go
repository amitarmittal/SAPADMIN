package dto

import (
	"encoding/json"
	"log"
)

type LimitsDto struct {
	MinBetValue float64 `json:"minBetValue"`
	MaxBetValue float64 `json:"maxBetValue"`
	OddsLimit   float64 `json:"oddsLimit"`
	Currency    string  `json:"currency"`
}

type PriceDto struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

type RunnerDto struct {
	RunnerId   string     `json:"runnerId"`
	RunnerName string     `json:"runnerName"`
	Status     string     `json:"status"`
	LayPrices  []PriceDto `json:"layPrices"`
	BackPrices []PriceDto `json:"backPrices"`
}

type MatchOddsDto struct {
	MarketId   string      `json:"marketId"`
	MarketName string      `json:"marketName"`
	MarketType string      `json:"marketType"`
	Status     string      `json:"status"`
	Runners    []RunnerDto `json:"runners"`
	Limits     LimitsDto   `json:"limits"`
}

type FancyMarketDto struct {
	MarketId   string    `json:"marketId"`
	MarketName string    `json:"marketName"`
	MarketType string    `json:"marketType"`
	Category   string    `json:"category"`
	Status     string    `json:"status"`
	NoRate     float64   `json:"noRate"`
	NoValue    float64   `json:"noValue"`
	YesRate    float64   `json:"yesRate"`
	YesValue   float64   `json:"yesValue"`
	Limits     LimitsDto `json:"limits"`
}
type AllMarketsDto struct {
	MatchOdds    []MatchOddsDto   `json:"matchOdds"`
	Bookmakers   []MatchOddsDto   `json:"bookmakers"`
	FancyMarkets []FancyMarketDto `json:"fancyMarkets"`
}

type EventDto struct {
	AwayScore       float64       `json:"awayScore"`
	HomeScore       float64       `json:"homeScore"`
	OpenDate        int64         `json:"openDate"`
	ProviderId      string        `json:"providerName"`
	SportId         string        `json:"sportId"`
	SportName       string        `json:"sportName"`
	CompetitionId   string        `json:"competitionId"`
	CompetitionName string        `json:"competitionName"`
	EventId         string        `json:"eventId"`
	EventName       string        `json:"eventName"`
	MarketId        string        `json:"marketId"`
	Status          string        `json:"status"`
	Markets         AllMarketsDto `json:"markets"`
}

// GetEventsRespDto represents response body of this API
type GetEventsRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Events           []EventDto `json:"sports"`
}

var (
	CricketEvent     string = `{"openDate":1629885600000,"sportId":"cricket","competitionId":"11365612","competitionName":"Test Matches","eventId":"30802553","eventName":"England v India","marketId":"1.186498304","market":{"matchOdds":[{"marketId":"1.186498304","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"10301","runnerName":"England","status":"OPEN","layPrices":[{"price":2.72,"size":730},{"price":2.74,"size":3906},{"price":2.76,"size":3612}],"backPrices":[{"price":2.7,"size":4500},{"price":2.68,"size":5497},{"price":2.66,"size":2992}]},{"runnerId":"414464","runnerName":"India","status":"OPEN","layPrices":[{"price":2.3,"size":1833},{"price":2.32,"size":11108},{"price":2.34,"size":3736}],"backPrices":[{"price":2.28,"size":707},{"price":2.26,"size":33837},{"price":2.24,"size":3811}]},{"runnerId":"60443","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":5.2,"size":887},{"price":5.3,"size":4033},{"price":5.4,"size":4011}],"backPrices":[{"price":5.1,"size":3613},{"price":5,"size":1494},{"price":4.9,"size":1097}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"}`
	CricketEventLive string = `{"openDate":1629826200000,"sportId":"cricket","competitionId":"11893330","competitionName":"T20 Blast","eventId":"30712805","eventName":"Yorkshire v Sussex","marketId":"1.185506027","market":{"matchOdds":[{"marketId":"1.185506027","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"655","runnerName":"Yorkshire","status":"OPEN","layPrices":[{"price":2.26,"size":101},{"price":2.28,"size":1324},{"price":2.3,"size":75488}],"backPrices":[{"price":2.24,"size":3265},{"price":2.2,"size":8162},{"price":2.16,"size":154}]},{"runnerId":"31379","runnerName":"Sussex","status":"OPEN","layPrices":[{"price":1.8,"size":4063},{"price":1.82,"size":9866},{"price":1.85,"size":180}],"backPrices":[{"price":1.79,"size":1354},{"price":1.78,"size":23435},{"price":1.77,"size":74986}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}`
	SoccerEvent      string = `{"openDate":1629849600000,"sportId":"soccer","competitionId":"67387","competitionName":"Argentinian Primera Division","eventId":"30824751","eventName":"Racing Club v Central Cordoba (SdE)","marketId":"1.186731059","market":{"matchOdds":[{"marketId":"1.186731059","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"198558","runnerName":"Racing Club","status":"OPEN","layPrices":[{"price":2,"size":425},{"price":2.02,"size":3319},{"price":2.04,"size":5202}],"backPrices":[{"price":1.95,"size":1171},{"price":1.94,"size":5058},{"price":1.93,"size":3096}]},{"runnerId":"19641049","runnerName":"Central Cordoba (SdE)","status":"OPEN","layPrices":[{"price":5.1,"size":377},{"price":5.2,"size":481},{"price":5.3,"size":353}],"backPrices":[{"price":5,"size":316},{"price":4.8,"size":80},{"price":4.7,"size":1139}]},{"runnerId":"58805","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":3.45,"size":98},{"price":3.5,"size":490},{"price":3.55,"size":462}],"backPrices":[{"price":3.3,"size":1577},{"price":3.25,"size":1110},{"price":3.2,"size":336}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"}`
	SoccerEventLive  string = `{"openDate":1629841500000,"sportId":"soccer","competitionId":"67387","competitionName":"Argentinian Primera Division","eventId":"30824607","eventName":"Atl Tucuman v CA Independiente","marketId":"1.186724887","market":{"matchOdds":[{"marketId":"1.186724887","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"5181445","runnerName":"Atl Tucuman","status":"OPEN","layPrices":[{"price":3.25,"size":1211},{"price":3.3,"size":1278},{"price":3.35,"size":3766}],"backPrices":[{"price":3.15,"size":590},{"price":3.1,"size":1940},{"price":3.05,"size":2344}]},{"runnerId":"8015153","runnerName":"CA Independiente","status":"OPEN","layPrices":[{"price":2.66,"size":207},{"price":2.68,"size":442},{"price":2.7,"size":669}],"backPrices":[{"price":2.64,"size":578},{"price":2.6,"size":567},{"price":2.58,"size":604}]},{"runnerId":"58805","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":3.3,"size":1092},{"price":3.35,"size":476},{"price":3.4,"size":530}],"backPrices":[{"price":3.2,"size":7028},{"price":3.15,"size":3574},{"price":3.1,"size":3406}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}`
	TennisEvent      string = `{"openDate":1629853200000,"sportId":"tennis","competitionId":"12371984","competitionName":"ATP Winston-Salem 2021","eventId":"30828517","eventName":"Carlos Alcaraz v Popyrin","marketId":"1.186787552","market":{"matchOdds":[{"marketId":"1.186787552","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"25215583","runnerName":"Carlos Alcaraz","status":"OPEN","layPrices":[{"price":1.82,"size":96},{"price":1.83,"size":30973},{"price":1.86,"size":384}],"backPrices":[{"price":1.81,"size":165},{"price":1.8,"size":7413},{"price":1.79,"size":1641}]},{"runnerId":"10706479","runnerName":"Alexei Popyrin","status":"OPEN","layPrices":[{"price":2.24,"size":1122},{"price":2.26,"size":6223},{"price":2.3,"size":3770}],"backPrices":[{"price":2.2,"size":25844},{"price":2.16,"size":876},{"price":2.14,"size":2833}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"}`
	TennisEventLive  string = `{"openDate":1629847800000,"sportId":"tennis","competitionId":"12372090","competitionName":"WTA Cleveland 2021","eventId":"30827252","eventName":"Begu v Hercog","marketId":"1.186777516","market":{"matchOdds":[{"marketId":"1.186777516","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"7642289","runnerName":"Irina-Camelia Begu","status":"OPEN","layPrices":[{"price":1.76,"size":7291},{"price":1.77,"size":832},{"price":1.79,"size":32}],"backPrices":[{"price":1.73,"size":422},{"price":1.72,"size":4280},{"price":1.71,"size":539}]},{"runnerId":"2548781","runnerName":"Polona Hercog","status":"OPEN","layPrices":[{"price":2.38,"size":398},{"price":2.4,"size":3361},{"price":2.44,"size":548}],"backPrices":[{"price":2.3,"size":5853},{"price":2.28,"size":369},{"price":2.26,"size":319}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}`
)

func GetEventsTestData(sportId string) GetEventsRespDto {
	//EventsJSON := `[{"openDate":1629885600000,"sportId":"cricket","competitionId":"11365612","competitionName":"Test Matches","eventId":"30802553","eventName":"England v India","marketId":"1.186498304","market":{"matchOdds":[{"marketId":"1.186498304","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"10301","runnerName":"England","status":"OPEN","layPrices":[{"price":2.72,"size":730},{"price":2.74,"size":3906},{"price":2.76,"size":3612}],"backPrices":[{"price":2.7,"size":4500},{"price":2.68,"size":5497},{"price":2.66,"size":2992}]},{"runnerId":"414464","runnerName":"India","status":"OPEN","layPrices":[{"price":2.3,"size":1833},{"price":2.32,"size":11108},{"price":2.34,"size":3736}],"backPrices":[{"price":2.28,"size":707},{"price":2.26,"size":33837},{"price":2.24,"size":3811}]},{"runnerId":"60443","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":5.2,"size":887},{"price":5.3,"size":4033},{"price":5.4,"size":4011}],"backPrices":[{"price":5.1,"size":3613},{"price":5,"size":1494},{"price":4.9,"size":1097}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"},{"openDate":1629826200000,"sportId":"cricket","competitionId":"11893330","competitionName":"T20 Blast","eventId":"30712805","eventName":"Yorkshire v Sussex","marketId":"1.185506027","market":{"matchOdds":[{"marketId":"1.185506027","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"655","runnerName":"Yorkshire","status":"OPEN","layPrices":[{"price":2.26,"size":101},{"price":2.28,"size":1324},{"price":2.3,"size":75488}],"backPrices":[{"price":2.24,"size":3265},{"price":2.2,"size":8162},{"price":2.16,"size":154}]},{"runnerId":"31379","runnerName":"Sussex","status":"OPEN","layPrices":[{"price":1.8,"size":4063},{"price":1.82,"size":9866},{"price":1.85,"size":180}],"backPrices":[{"price":1.79,"size":1354},{"price":1.78,"size":23435},{"price":1.77,"size":74986}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"},{"openDate":1629558000000,"sportId":"cricket","competitionId":"11365612","competitionName":"Test Matches","eventId":"30799340","eventName":"West Indies v Pakistan","marketId":"1.186441873","market":{"matchOdds":[{"marketId":"1.186441873","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"235","runnerName":"West Indies","status":"OPEN","layPrices":[{"price":510,"size":27},{"price":540,"size":27},{"price":580,"size":35}],"backPrices":[{"price":350,"size":86},{"price":340,"size":29},{"price":330,"size":29}]},{"runnerId":"7461","runnerName":"Pakistan","status":"OPEN","layPrices":[{"price":1.91,"size":42756},{"price":1.94,"size":6843},{"price":1.95,"size":1069}],"backPrices":[{"price":1.9,"size":7320},{"price":1.88,"size":104},{"price":1.87,"size":978}]},{"runnerId":"60443","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":2.14,"size":6591},{"price":2.16,"size":153},{"price":2.18,"size":961}],"backPrices":[{"price":2.12,"size":38521},{"price":2.06,"size":7456},{"price":2.04,"size":9420}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}]`
	//EventsJSON := `[{"openDate":1629849600000,"sportId":"soccer","competitionId":"67387","competitionName":"Argentinian Primera Division","eventId":"30824751","eventName":"Racing Club v Central Cordoba (SdE)","marketId":"1.186731059","market":{"matchOdds":[{"marketId":"1.186731059","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"198558","runnerName":"Racing Club","status":"OPEN","layPrices":[{"price":2,"size":425},{"price":2.02,"size":3319},{"price":2.04,"size":5202}],"backPrices":[{"price":1.95,"size":1171},{"price":1.94,"size":5058},{"price":1.93,"size":3096}]},{"runnerId":"19641049","runnerName":"Central Cordoba (SdE)","status":"OPEN","layPrices":[{"price":5.1,"size":377},{"price":5.2,"size":481},{"price":5.3,"size":353}],"backPrices":[{"price":5,"size":316},{"price":4.8,"size":80},{"price":4.7,"size":1139}]},{"runnerId":"58805","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":3.45,"size":98},{"price":3.5,"size":490},{"price":3.55,"size":462}],"backPrices":[{"price":3.3,"size":1577},{"price":3.25,"size":1110},{"price":3.2,"size":336}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"},{"openDate":1629841500000,"sportId":"soccer","competitionId":"67387","competitionName":"Argentinian Primera Division","eventId":"30824607","eventName":"Atl Tucuman v CA Independiente","marketId":"1.186724887","market":{"matchOdds":[{"marketId":"1.186724887","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"5181445","runnerName":"Atl Tucuman","status":"OPEN","layPrices":[{"price":3.25,"size":1211},{"price":3.3,"size":1278},{"price":3.35,"size":3766}],"backPrices":[{"price":3.15,"size":590},{"price":3.1,"size":1940},{"price":3.05,"size":2344}]},{"runnerId":"8015153","runnerName":"CA Independiente","status":"OPEN","layPrices":[{"price":2.66,"size":207},{"price":2.68,"size":442},{"price":2.7,"size":669}],"backPrices":[{"price":2.64,"size":578},{"price":2.6,"size":567},{"price":2.58,"size":604}]},{"runnerId":"58805","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":3.3,"size":1092},{"price":3.35,"size":476},{"price":3.4,"size":530}],"backPrices":[{"price":3.2,"size":7028},{"price":3.15,"size":3574},{"price":3.1,"size":3406}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"}]`
	//EventsJSON := `{"status":"RS_OK","errorDescription":"","sports":[{"openDate":1629853200000,"sportId":"tennis","competitionId":"12371984","competitionName":"ATP Winston-Salem 2021","eventId":"30828517","eventName":"Carlos Alcaraz v Popyrin","marketId":"1.186787552","market":{"matchOdds":[{"marketId":"1.186787552","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"25215583","runnerName":"Carlos Alcaraz","status":"OPEN","layPrices":[{"price":1.82,"size":96},{"price":1.83,"size":30973},{"price":1.86,"size":384}],"backPrices":[{"price":1.81,"size":165},{"price":1.8,"size":7413},{"price":1.79,"size":1641}]},{"runnerId":"10706479","runnerName":"Alexei Popyrin","status":"OPEN","layPrices":[{"price":2.24,"size":1122},{"price":2.26,"size":6223},{"price":2.3,"size":3770}],"backPrices":[{"price":2.2,"size":25844},{"price":2.16,"size":876},{"price":2.14,"size":2833}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"},{"openDate":1629853200000,"sportId":"tennis","competitionId":"12371984","competitionName":"ATP Winston-Salem 2021","eventId":"30823723","eventName":"T Monteiro v Goffin","marketId":"1.186709665","market":{"matchOdds":[{"marketId":"1.186709665","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"8326656","runnerName":"Thiago Monteiro","status":"OPEN","layPrices":[{"price":2.6,"size":1906},{"price":2.62,"size":2901},{"price":2.64,"size":3013}],"backPrices":[{"price":2.52,"size":768},{"price":2.5,"size":94},{"price":2.48,"size":208}]},{"runnerId":"19924835","runnerName":"David Goffin","status":"OPEN","layPrices":[{"price":1.65,"size":601},{"price":1.66,"size":569},{"price":1.67,"size":141}],"backPrices":[{"price":1.63,"size":3040},{"price":1.62,"size":4691},{"price":1.61,"size":4940}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"},{"openDate":1629847800000,"sportId":"tennis","competitionId":"12372090","competitionName":"WTA Cleveland 2021","eventId":"30827252","eventName":"Begu v Hercog","marketId":"1.186777516","market":{"matchOdds":[{"marketId":"1.186777516","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"7642289","runnerName":"Irina-Camelia Begu","status":"OPEN","layPrices":[{"price":1.76,"size":7291},{"price":1.77,"size":832},{"price":1.79,"size":32}],"backPrices":[{"price":1.73,"size":422},{"price":1.72,"size":4280},{"price":1.71,"size":539}]},{"runnerId":"2548781","runnerName":"Polona Hercog","status":"OPEN","layPrices":[{"price":2.38,"size":398},{"price":2.4,"size":3361},{"price":2.44,"size":548}],"backPrices":[{"price":2.3,"size":5853},{"price":2.28,"size":369},{"price":2.26,"size":319}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":false,"providerName":"DREAM_API"}]}`
	getEventsRespDto := GetEventsRespDto{}
	getEventsRespDto.Status = "RS_OK"
	getEventsRespDto.ErrorDescription = ""
	getEventsRespDto.Events = []EventDto{}

	if sportId == "cricket" {
		liveEventDto := EventDto{}
		err := json.Unmarshal([]byte(CricketEventLive), &liveEventDto)
		if err != nil {
			log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
		} else {
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
		}
		eventDto := EventDto{}
		err = json.Unmarshal([]byte(CricketEvent), &eventDto)
		if err != nil {
			log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
		} else {
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
		}
	} else if sportId == "soccer" {
		liveEventDto := EventDto{}
		err := json.Unmarshal([]byte(SoccerEventLive), &liveEventDto)
		if err != nil {
			log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
		} else {
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
		}
		eventDto := EventDto{}
		err = json.Unmarshal([]byte(SoccerEvent), &eventDto)
		if err != nil {
			log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
		} else {
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
		}
	} else if sportId == "tennis" {
		liveEventDto := EventDto{}
		err := json.Unmarshal([]byte(TennisEventLive), &liveEventDto)
		if err != nil {
			log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
		} else {
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, liveEventDto)
		}
		eventDto := EventDto{}
		err = json.Unmarshal([]byte(TennisEvent), &eventDto)
		if err != nil {
			log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
		} else {
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
			getEventsRespDto.Events = append(getEventsRespDto.Events, eventDto)
		}
	} else {
		getEventsRespDto.Status = "RS_ERROR"
		getEventsRespDto.ErrorDescription = "Invalid sportsId"
	}
	/*
		resData, err := json.Marshal(getEventsRespDto)
		if err != nil {
			log.Println("GetTestData: Marshal failed to convert test data into json string - ", err.Error())
		}
		resStr := string(resData)
		log.Println("GetTestData: response String is - ", resStr)
	*/
	return getEventsRespDto
}
