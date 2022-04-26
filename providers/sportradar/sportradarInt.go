package sportradar

import (
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	dreamdto "Sp/dto/providers/dream"
	"Sp/dto/providers/sportradar"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"Sp/dto/sports"
	provider "Sp/providers"
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	//BaseUrl             string        = "https://stage-feed.mysportsfeed.io/api/v1/betfair-api/"
	//BaseUrl             string        = "https://feed-dev.hypexone.com/api/v1/"
	//BaseUrl2            string        = "http://sportradarodss.eu-west-2.elasticbeanstalk.com/api/v1/"
	//BaseUrl2            string        = "http://sportradarodds.eu-west-2.elasticbeanstalk.com/api/v1/"
	// BaseUrl string = "https://api.indisports.live/api/v1/"
	BaseUrl             string = os.Getenv("L2_API_URL")
	GetSportsUrl        string = BaseUrl + "list-sports/?providerId=SportRadar"
	GetCompetitionsUrl  string = BaseUrl + "list-competitions/?providerId=SportRadar"
	GetEventsUrl        string = BaseUrl + "list-events/?providerId=SportRadar"
	GetLiveEventsUrl    string = BaseUrl + "list-live-events/?providerId=SportRadar"
	GetMarketsUrl       string = BaseUrl + "list-all-markets/?providerId=SportRadar"
	ValidateOddsUrl     string = os.Getenv("L1_SR_VO_URL") + "sport-radar-api/validate-odds/?providerId=SportRadar"
	PlaceOrderUrl       string = os.Getenv("L1_SR_VO_URL") + "sport-radar-api/place-single-bet/" // ?providerId=SportRadar"
	GetMarketResultsUrl string = BaseUrl + "list-market-settlements/?providerId=SportRadar"
	CancelOrderUrl      string = os.Getenv("L1_SR_VO_URL") + "sport-radar-api/cancel-single-bet/"
	BetFairMaketUrl     string = BaseUrl + "sr-premium-markets/list-all-markets?betFairEventId="
	// https://wsfeed.indisports.live/api/v1/sport-radar-api/cancel-single-bet/
	TimeOut      time.Duration = 15 // 5 seconds
	ProviderId   string        = "SportRadar"
	ProviderName string        = "Sport Radar"
	BOOKMAKER_ID string        = "34948"
)

func GetSports() {
	//log.Println("GetSports: SportRadar Start.")
	sports := []commondto.SportDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "GET", GetSportsUrl, TimeOut)
	if err != nil {
		log.Println("GetSports: SportsFeedCall failed with error - ", err.Error())
		return
	}
	//log.Println(string(resp))
	err = json.Unmarshal(resp, &sports)
	if err != nil {
		log.Println("GetSports: json.Unmarshal error", err.Error())
		log.Println("GetSports: response data is - ", string(resp))
	}
	// Update cache and db
	//log.Println("GetSports: Adding Sports to cache & db")
	provider.SyncSports(sports, ProviderId, ProviderName)
	provider.SyncSportStatus(ProviderId)
	return
}

func GetCompetitions() {
	//log.Println("GetCompetitions: SportRadar Start.")
	sports, err := database.GetSports(ProviderId)
	if err != nil {
		log.Println("GetCompetitions: GetSports failed with error - ", err.Error())
		return
	}
	if len(sports) == 0 {
		log.Println("GetCompetitions: GetSports returned ZERO records - ", len(sports))
		return
	}
	competitions := []commondto.CompetitionDto{}
	sportMap := make(map[string]models.Sport)
	// 1. Reqest object
	sportIds := []string{}
	for _, sport := range sports {
		sportIds = append(sportIds, sport.SportId)
		sportMap[sport.SportId] = sport
	}
	jsonData, err := json.Marshal(sportIds)
	if err != nil {
		log.Println("GetCompetitions: Failed to convert DTO to JSON")
		return
	}
	resp, err := provider.SportsFeedCall(jsonData, "POST", GetCompetitionsUrl, TimeOut)
	if err != nil {
		log.Println("GetCompetitions: SportsFeedCall failed with error - ", err.Error())
		log.Println("GetCompetitions: request data is - ", string(jsonData))
		return
	}
	//log.Println(string(resp))
	err = json.Unmarshal(resp, &competitions)
	if err != nil {
		log.Println("GetCompetitions: json.Unmarshal error", err.Error())
		log.Println("GetCompetitions: response data is - ", string(resp))
		return
	}
	//log.Println("GetCompetitions: competitions count is - ", len(competitions))
	if len(competitions) == 0 {
		return
	}
	// Update cache and db
	//log.Println("GetCompetitions: Adding Competitions to cache & db")
	provider.SyncCompetitions(competitions, sportMap)
	//provider.SyncCompetitionStatus(ProviderId)
	return
}

func GetUpcomingEvents() {
	//log.Println("GetUpcomingEvents: SportRadar Start.")
	sports, err := database.GetSports(ProviderId)
	if err != nil {
		log.Println("GetUpcomingEvents: GetSports failed with error - ", err.Error())
		return
	}
	//log.Println("GetUpcomingEvents: GetSports returned ZERO records - ", len(sports))
	if len(sports) == 0 {
		return
	}
	for _, sport := range sports {
		if constants.SAP.ObjectStatus.ACTIVE() != sport.Status {
			log.Println("GetUpcomingEvents: SportRadar Sport is not ACTIVE - ", sport.SportId)
			continue
		}
		events, err := GetEvents(sport.SportId)
		if err != nil {
			log.Println("GetUpcomingEvents: Failed to fetch SportRadar events for sport - ", sport.SportId)
			log.Println("GetUpcomingEvents: Error is - ", err.Error())
			continue
		}
		//log.Println("GetUpcomingEvents: events count is - ", len(events))
		if len(events) == 0 {
			continue
		}
		provider.SyncEvents(events, ProviderId, sport.SportId)
	}
	return
}

func GetLiveEvents(sportIds []string) ([]dto.EventDto, error) {
	//log.Println("GetLiveEvents: SportRadar sport count - ", len(sportIds))
	// 0. Default response object
	events := []dto.EventDto{}
	// 1. Reqest object
	jsonData, err := json.Marshal(sportIds)
	if err != nil {
		log.Println("GetLiveEvents: Failed to convert DTO to JSON - ", err.Error())
		return events, err
	}

	resp, err := provider.SportsFeedCall(jsonData, "POST", GetLiveEventsUrl, TimeOut)
	if err != nil {
		log.Println("GetLiveEvents: SportsFeedCall failed with error - ", err.Error())
		log.Println("GetLiveEvents: request data is - ", string(jsonData))
		return events, err
	}
	//log.Println("GetEvents: 1. response data is - ", string(resp))
	err = json.Unmarshal(resp, &events)
	if err != nil {
		log.Println("GetLiveEvents: json.Unmarshal error - ", err.Error())
		log.Println("GetLiveEvents: response data is - ", string(resp))
		return events, err
	}
	return events, nil
}

func GetEvents(sportId string) ([]dto.EventDto, error) {
	//log.Println("GetEvents: SportRadar  for sport - ", sportId)
	//log.Println("GetEvents: SportRadar Start.")
	// 0. Default response object
	events := []dto.EventDto{}
	// 1. Reqest object
	reqDto := dreamdto.EventsReqDto{}
	reqDto.SportId = sportId
	reqDto.CompetitionsIds = []string{}

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("GetEvents: Failed to convert DTO to JSON")
		return events, err
	}

	resp, err := provider.SportsFeedCall(jsonData, "POST", GetEventsUrl, TimeOut)
	if err != nil {
		log.Println("GetEvents: SportsFeedCall failed with error - ", err.Error())
		log.Println("GetEvents: request data is - ", string(jsonData))
		return events, err
	}
	//log.Println("GetEvents: 1. response data is - ", string(resp))
	err = json.Unmarshal(resp, &events)
	if err != nil {
		log.Println("GetEvents: json.Unmarshal error - ", err.Error())
		log.Println("GetEvents: response data is - ", string(resp))
		return events, err
	}
	// Update cache and db asynchronously
	//go provider.AddEvents(events)
	//log.Println("GetEvents: Adding Events to cache & db asynchronously!!!")
	return events, nil
}

func GetMarkets(sportId string, eventId string) dto.EventDto {
	//log.Println("GetMarkets: SportRadar Start.")
	event := dto.EventDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "GET", GetMarketsUrl+"&sportId="+sportId+"&eventId="+eventId, TimeOut)
	if err != nil {
		log.Println("GetMarkets: SportsFeedCall failed with error - ", err.Error())
		return event
	}
	//log.Println(string(resp))
	err = json.Unmarshal(resp, &event)
	if err != nil {
		log.Println("GetMarkets: json.Unmarshal error", err.Error())
		log.Println("GetMarkets: response data is - ", string(resp))
	}
	return event
}

func ValidateOdds(betReq requestdto.PlaceBetReqDto) (dreamdto.ValidateOddsRespDto, error) {
	//log.Println("ValidateOdds: SportRadar Start.")
	// 0. Default response object
	reqJson, err := json.Marshal(betReq)
	log.Println("ValidateOdds: betReq json is - ", string(reqJson))
	respDto := dreamdto.ValidateOddsRespDto{}
	// 1. Reqest object
	reqDto := dreamdto.ValidateOddsReqDto{}
	reqDto.SportId = betReq.SportId
	reqDto.EventId = betReq.EventId
	reqDto.MarketType = betReq.MarketType
	reqDto.MarketId = betReq.MarketId
	reqDto.RunnerId = betReq.RunnerId
	reqDto.BetType = betReq.BetType
	reqDto.OddValue = betReq.OddValue
	reqDto.SessionOutcome = betReq.SessionOutcome

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("ValidateOdds: Failed to convert DTO to JSON")
		return respDto, err
	}
	log.Println("ValidateOdds: request data is - ", string(jsonData))
	resp, err := provider.SportsFeedCall(jsonData, "POST", ValidateOddsUrl, TimeOut)
	if err != nil {
		log.Println("ValidateOdds: SportsFeedCall failed with error - ", err.Error())
		log.Println("ValidateOdds: request data is - ", string(jsonData))
		return respDto, err
	}
	log.Println("ValidateOdds: response data is - ", string(resp))
	err = json.Unmarshal(resp, &respDto)
	if err != nil {
		log.Println("ValidateOdds: json.Unmarshal error", err.Error())
		log.Println("ValidateOdds: response data is - ", string(resp))
		return respDto, err
	}
	return respDto, nil
}

func PlaceOrder(reqDto requestdto.PlaceBetReqDto, betDto sports.BetDto) (sportradar.PlaceOrderResp, error) {
	//log.Println("PlaceOrder: SportRadar Start.")
	// 0. Default response object
	placeOrderResp := sportradar.PlaceOrderResp{}
	// 1. Request Object
	betDto.BetReq.BetId = BOOKMAKER_ID + "-" + betDto.BetId
	placeOrderReq := sportradar.NewPlaceOrderReq(reqDto, betDto)
	jsonData, err := json.Marshal(placeOrderReq)
	if err != nil {
		log.Println("PlaceOrder: Failed to convert DTO to JSON")
		return placeOrderResp, err
	}
	log.Println("PlaceOrder: SportRadar request data is - ", string(jsonData))
	resp, err := provider.SportsFeedCall(jsonData, "POST", PlaceOrderUrl, TimeOut)
	if err != nil {
		log.Println("PlaceOrder: SportsFeedCall failed with error - ", err.Error())
		log.Println("PlaceOrder: request data is - ", string(jsonData))
		return placeOrderResp, err
	}
	//log.Println(string(resp))
	log.Println("PlaceOrder: SportRadar response data is - ", string(resp))
	err = json.Unmarshal(resp, &placeOrderResp)
	if err != nil {
		log.Println("PlaceOrder: json.Unmarshal error", err.Error())
		log.Println("PlaceOrder: response data is - ", string(resp))
		return placeOrderResp, err
	}
	return placeOrderResp, nil
}

func GetMarketResults(marketIds []string) (commondto.GetMarketResultsResp, error) {
	//log.Println("GetMarketResults: SportRadar markets count - ", len(marketIds))
	// 0. Default response object
	respObj := commondto.GetMarketResultsResp{}
	respObj.Status = "RS_ERROR"
	respObj.ErrorDescription = "Default Error from SAP"
	respObj.MarketResults = []commondto.MarketResult{}
	// 1. Reqest object
	jsonData, err := json.Marshal(marketIds)
	if err != nil {
		log.Println("GetMarketResults: Failed to convert DTO to JSON with error - ", err.Error())
		return respObj, err
	}
	//log.Println("GetMarketResults: request data is - ", string(jsonData))
	resp, err := provider.SportsFeedCall(jsonData, "POST", GetMarketResultsUrl, TimeOut)
	if err != nil {
		log.Println("GetMarketResults: SportsFeedCall failed with error - ", err.Error())
		log.Println("GetMarketResults: request data is - ", string(jsonData))
		return respObj, err
	}
	//log.Println("GetMarketResults: 1. response data is - ", string(resp))
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		log.Println("GetMarketResults: json.Unmarshal error - ", err.Error())
		log.Println("GetMarketResults: response data is - ", string(resp))
		return respObj, err
	}
	return respObj, nil
}

func CancelOrder(ticketId string) (responsedto.BasicRespDto, error) {
	//log.Println("GetSports: SportRadar Start.")
	respDto := responsedto.BasicRespDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "POST", CancelOrderUrl+ticketId, TimeOut)
	if err != nil {
		log.Println("CancelOrder: SportsFeedCall failed with error - ", err.Error())
		return respDto, err
	}
	//log.Println(string(resp))
	log.Println("CancelOrder: response data is - ", ticketId, string(resp))
	err = json.Unmarshal(resp, &respDto)
	if err != nil {
		log.Println("CancelOrder: json.Unmarshal error", err.Error())
		log.Println("CancelOrder: response data is - ", string(resp))
		return respDto, err
	}
	return respDto, nil
}

func GetBetFairMarket(eventId string) (operatordto.SrPremiumMarketsRespDto, error) {
	var respDto operatordto.SrPremiumMarketsRespDto
	resp, err := provider.SportsFeedCall([]byte{}, "POST", BetFairMaketUrl+eventId, TimeOut)
	if err != nil {
		log.Println("GetBetFairMarket: SportsFeedCall failed with error - ", err.Error())
		return respDto, err
	}
	log.Println("GetBetFairMarket: response data is - ", eventId, string(resp))
	err = json.Unmarshal(resp, &respDto)
	
	if err != nil {
		log.Println("GetBetFairMarket: json.Unmarshal error", err.Error())
		log.Println("GetBetFairMarket: response data is - ", string(resp))
		return respDto, err
	}
	return respDto, nil
}
