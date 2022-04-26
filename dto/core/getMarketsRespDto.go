package dto

import (
	"encoding/json"
	"log"
)

type GetMarketsRespDto struct {
	Status           string   `json:"status"`
	ErrorDescription string   `json:"errorDescription"`
	Event            EventDto `json:"event"`
}

var (
	CricketMarkets string = `{"openDate":1629885600000,"sportId":"cricket","competitionId":"11365612","competitionName":"Test Matches","eventId":"30802553","eventName":"England v India","marketId":"1.186498304","market":{"matchOdds":[{"marketId":"1.186498304","marketName":"Match Odds","status":"OPEN","matchOddsRunners":[{"runnerId":"","runnerName":"England","status":"OPEN","layPrices":[{"price":1.86,"size":12293},{"price":1.87,"size":51809},{"price":1.88,"size":14287}],"backPrices":[{"price":1.85,"size":22916},{"price":1.84,"size":53327},{"price":1.83,"size":244813}]},{"runnerId":"","runnerName":"India","status":"OPEN","layPrices":[{"price":3.05,"size":15503},{"price":3.1,"size":36480},{"price":3.15,"size":115166}],"backPrices":[{"price":3,"size":813},{"price":2.96,"size":7725},{"price":2.94,"size":17102}]},{"runnerId":"","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":8,"size":54406},{"price":8.2,"size":2616},{"price":8.4,"size":5217}],"backPrices":[{"price":7.6,"size":38867},{"price":7.4,"size":28683},{"price":7.2,"size":4185}]}]}],"bookmakers":[{"marketId":"4.1629541855-BM","marketName":"Bookmaker 0%Comm","status":"","bookmakerRunners":[{"runnerId":"","runnerName":"England","backPrice":83,"backSize":100000,"layPrice":88,"laySize":100000},{"runnerId":"","runnerName":"India","backPrice":0,"backSize":100000,"layPrice":0,"laySize":100000},{"runnerId":"","runnerName":"The Draw","backPrice":0,"backSize":100000,"layPrice":0,"laySize":100000}]}],"fancyMarkets":[{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0},{"marketId":"","marketName":"","marketType":"fancy2","status":"OPEN","noRate":0,"noValue":0,"yesRate":0,"yesValue":0}],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}`
	SoccerMarkets  string = `{"openDate":1629885600000,"sportId":"soccer","competitionId":"12209550","competitionName":"South Korean K League Classic","eventId":"30822269","eventName":"Jeonbuk Motors v Pohang Steelers","marketId":"1.186703900","market":{"matchOdds":[{"marketId":"1.186703900","marketName":"Match Odds","status":"OPEN","inplay":true,"matchOddsRunners":[{"selectionId":"2224159","runnerName":"Jeonbuk Motors","status":"OPEN","layPrices":[{"price":2.48,"size":2348.0},{"price":2.5,"size":178.0},{"price":2.52,"size":135.0}],"backPrices":[{"price":2.46,"size":1678.0},{"price":2.44,"size":2021.0},{"price":2.4,"size":635.0}]},{"selectionId":"1363823","runnerName":"Pohang Steelers","status":"OPEN","layPrices":[{"price":4.8,"size":955.0},{"price":5.0,"size":226.0},{"price":5.1,"size":79.0}],"backPrices":[{"price":4.7,"size":564.0},{"price":4.6,"size":1237.0},{"price":4.5,"size":747.0}]},{"selectionId":"58805","runnerName":"The Draw","status":"OPEN","layPrices":[{"price":2.62,"size":1220.0},{"price":2.64,"size":802.0},{"price":2.66,"size":1211.0}],"backPrices":[{"price":2.58,"size":1377.0},{"price":2.56,"size":222.0},{"price":2.54,"size":848.0}]}]},{"marketId":"1.186703904","marketName":"Over/Under 0.5 Goals","status":"OPEN","inplay":true,"matchOddsRunners":[{"selectionId":"5851482","runnerName":"Under 0.5 Goals","status":"OPEN","layPrices":[{"price":4.1,"size":748.0},{"price":4.2,"size":2184.0},{"price":4.3,"size":3819.0}],"backPrices":[{"price":4.0,"size":587.0},{"price":3.95,"size":214.0},{"price":3.9,"size":1206.0}]},{"selectionId":"5851483","runnerName":"Over 0.5 Goals","status":"OPEN","layPrices":[{"price":1.34,"size":5271.0},{"price":1.35,"size":7071.0},{"price":1.36,"size":982.0}],"backPrices":[{"price":1.32,"size":6746.0},{"price":1.31,"size":11352.0},{"price":1.3,"size":6049.0}]}]},{"marketId":"1.186703910","marketName":"Over/Under 2.5 Goals","status":"OPEN","inplay":true,"matchOddsRunners":[{"selectionId":"47972","runnerName":"Under 2.5 Goals","status":"OPEN","layPrices":[{"price":1.2,"size":1565.0},{"price":1.21,"size":7767.0},{"price":1.22,"size":5911.0}],"backPrices":[{"price":1.19,"size":4057.0},{"price":1.18,"size":10303.0},{"price":1.17,"size":414.0}]},{"selectionId":"47973","runnerName":"Over 2.5 Goals","status":"OPEN","layPrices":[{"price":6.4,"size":2249.0},{"price":6.6,"size":393.0},{"price":7.0,"size":69.0}],"backPrices":[{"price":6.2,"size":53.0},{"price":6.0,"size":258.0},{"price":5.9,"size":422.0}]}]},{"marketId":"1.186703914","marketName":"Over/Under 1.5 Goals","status":"OPEN","inplay":true,"matchOddsRunners":[{"selectionId":"1221385","runnerName":"Under 1.5 Goals","status":"OPEN","layPrices":[{"price":1.72,"size":340.0},{"price":1.73,"size":2392.0},{"price":1.74,"size":1799.0}],"backPrices":[{"price":1.7,"size":717.0},{"price":1.69,"size":890.0},{"price":1.68,"size":1139.0}]},{"selectionId":"1221386","runnerName":"Over 1.5 Goals","status":"OPEN","layPrices":[{"price":2.44,"size":500.0},{"price":2.46,"size":1054.0},{"price":2.48,"size":6727.0}],"backPrices":[{"price":2.4,"size":97.0},{"price":2.38,"size":470.0},{"price":2.36,"size":1706.0}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}`
	TennisMarkets  string = `{"openDate":1629842400000,"sportId":"tennis","competitionId":"12372028","competitionName":"WTA Chicago  2021","eventId":"30828238","eventName":"Svitolina v Ferro","marketId":"1.186785868","market":{"matchOdds":[{"marketId":"1.186785868","marketName":"Match Odds","status":"OPEN","inplay":true,"matchOddsRunners":[{"selectionId":"6324464","runnerName":"Elina Svitolina","status":"OPEN","layPrices":[{"price":1.47,"size":4835.0},{"price":1.48,"size":1465.0},{"price":1.49,"size":2783.0}],"backPrices":[{"price":1.46,"size":5092.0},{"price":1.45,"size":6581.0},{"price":1.44,"size":5446.0}]},{"selectionId":"8535268","runnerName":"Fiona Ferro","status":"OPEN","layPrices":[{"price":3.2,"size":2750.0},{"price":3.25,"size":2516.0},{"price":3.3,"size":2376.0}],"backPrices":[{"price":3.1,"size":2970.0},{"price":3.0,"size":1405.0},{"price":2.96,"size":1060.0}]}]}],"bookmakers":[],"fancyMarkets":[],"enableMatchOdds":true,"enableBookmaker":true,"enableFancy":true},"inPlay":true,"providerName":"DREAM_API"}`
)

func GetMarketsTestData(sportId string) GetMarketsRespDto {
	getMarketsRespDto := GetMarketsRespDto{}
	getMarketsRespDto.Status = "RS_OK"
	getMarketsRespDto.ErrorDescription = ""
	getMarketsRespDto.Event = EventDto{}
	if sportId == "cricket" {
		err := json.Unmarshal([]byte(CricketMarkets), &getMarketsRespDto.Event)
		if err != nil {
			log.Println("GetMarketsTestData: Unmarshal failed to convert test data into event - ", err.Error())
			getMarketsRespDto.Status = "RS_ERROR"
			getMarketsRespDto.ErrorDescription = "Failed to fetche Market Data"
		}
	} else if sportId == "soccer" {
		err := json.Unmarshal([]byte(SoccerMarkets), &getMarketsRespDto.Event)
		if err != nil {
			log.Println("GetMarketsTestData: Unmarshal failed to convert test data into event - ", err.Error())
			getMarketsRespDto.Status = "RS_ERROR"
			getMarketsRespDto.ErrorDescription = "Failed to fetche Market Data"
		}
	} else if sportId == "tennis" {
		err := json.Unmarshal([]byte(TennisMarkets), &getMarketsRespDto.Event)
		if err != nil {
			log.Println("GetMarketsTestData: Unmarshal failed to convert test data into event - ", err.Error())
			getMarketsRespDto.Status = "RS_ERROR"
			getMarketsRespDto.ErrorDescription = "Failed to fetche Market Data"
		}
	} else {
		getMarketsRespDto.Status = "RS_ERROR"
		getMarketsRespDto.ErrorDescription = "Invalid sportsId"
	}
	/*
		resData, err := json.Marshal(getMarketsRespDto)
		if err != nil {
			log.Println("GetMarketsTestData: Marshal failed to convert test data into json string - ", err.Error())
		}
		resStr := string(resData)
		log.Println("GetMarketsTestData: response String is - ", resStr)
	*/
	return getMarketsRespDto
}
