package dream

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	dreamdto "Sp/dto/providers/dream"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	provider "Sp/providers"
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	//BaseUrl         string        = "https://feed.mysportsfeed.io/api/v1/dream-api/"
	//BaseUrl            string        = "https://feed-dev.hypexone.com/api/v1/"
	// BaseUrl string = "https://api.indisports.live/api/v1/"
	BaseUrl            string        = os.Getenv("L2_API_URL")
	GetSportsUrl       string        = BaseUrl + "list-sports/?providerId=Dream"
	GetCompetitionsUrl string        = BaseUrl + "list-competitions/?providerId=Dream"
	GetEventsUrl       string        = BaseUrl + "list-events/?providerId=Dream"
	GetLiveEventsUrl   string        = BaseUrl + "list-live-events/?providerId=Dream"
	GetMarketsUrl      string        = BaseUrl + "list-all-markets/?providerId=Dream"
	ValidateOddsUrl    string        = BaseUrl + "validate-odds/?providerId=Dream"
	LoadTestUrl        string        = BaseUrl + "load-testing-api/"
	LoadTestL2Url      string        = BaseUrl + "loadtest/simple-ping"
	LoadTestBetFairUrl string        = os.Getenv("BETFAIR_L1") + "loadtest/simple-ping"
	LoadTestDreamUrl   string        = os.Getenv("DREAM_L1") + "dream-api/loadtest/simple-ping"
	LoadTestSPUrl      string        = os.Getenv("SPORTRADAR_L1") + "loadtest/simple-ping"
	TimeOut            time.Duration = 15 // 5 seconds
	ProviderId         string        = "Dream"
	ProviderName       string        = "Dream Feed"
	BB_DREAM_L2        string        = os.Getenv("BB_DREAM_L2")
)

func GetSports() {
	log.Println("GetSports: Dream.GetSports STARTED ")
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
	log.Println("GetSports: Dream.GetSports ENDED ")
	return
}

func GetCompetitions() {
	//log.Println("GetCompetitions: Dream Start.")
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
	//log.Println("GetCompetitions: request data is - ", string(jsonData))
	resp, err := provider.SportsFeedCall(jsonData, "POST", GetCompetitionsUrl, TimeOut)
	if err != nil {
		log.Println("GetCompetitions: SportsFeedCall failed with error - ", err.Error())
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
	//log.Println("GetUpcomingEvents: Dream Start.")
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
			log.Println("GetUpcomingEvents: Dream Sport is not ACTIVE - ", sport.SportId)
			continue
		}
		events, err := GetEvents(sport.SportId)
		if err != nil {
			log.Println("GetUpcomingEvents: Failed to fetch Dream events for sport - ", sport.SportId)
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
	//log.Println("GetLiveEvents: Dream sport count - ", len(sportIds))
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
	//log.Println("GetEvents: Dream Start for sport - ", sportId)
	// 0. Default response object
	events := []dto.EventDto{}
	// 1. Reqest object
	reqDto := dreamdto.EventsReqDto{}
	reqDto.SportId = sportId
	reqDto.CompetitionsIds = []string{}

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("ValidateOdds: Failed to convert DTO to JSON")
		return events, err
	}

	resp, err := provider.SportsFeedCall(jsonData, "POST", GetEventsUrl, TimeOut)
	//resp, err := provider.SportsFeedCall([]byte{}, "GET", GetEventsUrl+"&sportId="+sportId, TimeOut)
	//resp, err := provider.SportsFeedCall([]byte{}, "GET", GetEventsUrl+sportId, TimeOut)
	if err != nil {
		log.Println("GetEvents: SportsFeedCall failed with error - ", err.Error())
		return events, err
	}
	//log.Println(string(resp))
	err = json.Unmarshal(resp, &events)
	if err != nil {
		log.Println("GetEvents: json.Unmarshal error - ", err.Error())
		log.Println("GetEvents: response data is - ", string(resp))
		return events, err
	}
	// Update cache and db asynchronously
	//go provider.AddEvents(events)
	//log.Println("GetMarkets: Adding Events to cache & db asynchronously!!!")
	//log.Println("GetEvents: Events Count is - ", len(events))
	return events, nil
}

func GetMarkets(sportId string, eventId string, operatorId string) dto.EventDto {
	operator, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetMarkets: cache.GetOperatorDetails failed with error - ", err.Error())
		operator = operatordto.OperatorDTO{}
	}
	event := dto.EventDto{}
	getMarketsUrl := GetMarketsUrl
	// if operatorId == "betbhai" {
	// 	if BB_DREAM_L2 != "" {
	// 		getMarketsUrl = BB_DREAM_L2 + "list-all-markets/?providerId=Dream"
	// 	}
	// }
	if operator.DreamFeed != "" {
		getMarketsUrl = operator.DreamFeed + "list-all-markets/?providerId=Dream"
	}
	resp, err := provider.SportsFeedCall([]byte{}, "GET", getMarketsUrl+"&sportId="+sportId+"&eventId="+eventId, TimeOut)
	//resp, err := provider.SportsFeedCall([]byte{}, "GET", GetMarketsUrl+sportId+"/"+eventId, TimeOut)
	if err != nil {
		log.Println("GetMarkets: SportsFeedCall failed with error - ", err.Error())
		return event
	}
	//log.Println(string(resp))
	err = json.Unmarshal(resp, &event)
	if err != nil {
		log.Println("GetMarkets: json.Unmarshal error - ", err.Error())
		log.Println("GetMarkets: response data is - ", string(resp))
	}
	return event
}

func ValidateOdds(betReq requestdto.PlaceBetReqDto) (dreamdto.ValidateOddsRespDto, error) {
	operator, err := cache.GetOperatorDetails(betReq.OperatorId)
	if err != nil {
		log.Println("GetMarkets: cache.GetOperatorDetails failed with error - ", err.Error())
		operator = operatordto.OperatorDTO{}
	}
	// 0. Default response object
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
		log.Println("ValidateOdds: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, err
	}
	log.Println("ValidateOdds: Test Automation: request data is - ", string(jsonData))
	validateOddsUrl := ValidateOddsUrl
	// if betReq.OperatorId == "betbhai" {
	// 	if BB_DREAM_L2 != "" {
	// 		validateOddsUrl = BB_DREAM_L2 + "validate-odds/?providerId=Dream"
	// 	}
	// }
	if operator.DreamFeed != "" {
		validateOddsUrl = operator.DreamFeed + "validate-odds/?providerId=Dream"
	}
	resp, err := provider.SportsFeedCall(jsonData, "POST", validateOddsUrl, TimeOut)
	if err != nil {
		log.Println("ValidateOdds: SportsFeedCall failed with error - ", err.Error())
		return respDto, err
	}
	// log.Println(string(resp))
	log.Println("ValidateOdds: response data is - ", string(resp))
	err = json.Unmarshal(resp, &respDto)
	if err != nil {
		log.Println("ValidateOdds: json.Unmarshal error - ", err.Error())
		return respDto, err
	}
	return respDto, nil
}

func EndToEnd() (responsedto.BasicRespDto, error) {
	log.Println("EndToEnd: L2 endpoint URL is - ", LoadTestUrl)
	respObj := responsedto.BasicRespDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "POST", LoadTestUrl, TimeOut)
	if err != nil {
		log.Println("EndToEnd: SportsFeedCall failed with error - ", err.Error())
		return respObj, err
	}
	//log.Println(string(resp))
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		log.Println("EndToEnd: json.Unmarshal error", err.Error())
		log.Println("EndToEnd: response data is - ", string(resp))
		return respObj, err
	}
	return respObj, nil
}

func LayerTwo() (responsedto.BasicRespDto, error) {
	respObj := responsedto.BasicRespDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "POST", LoadTestL2Url, TimeOut)
	if err != nil {
		log.Println("LayerTwo: SportsFeedCall failed with error - ", err.Error())
		return respObj, err
	}
	log.Println(string(resp))
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		log.Println("LayerTwo: json.Unmarshal error", err.Error())
		log.Println("LayerTwo: response data is - ", string(resp))
		return respObj, err
	}
	return respObj, nil
}

func LayerOneBetFair() (responsedto.BasicRespDto, error) {
	respObj := responsedto.BasicRespDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "POST", LoadTestBetFairUrl, TimeOut)
	if err != nil {
		log.Println("LayerOneBetFair: SportsFeedCall failed with error - ", err.Error())
		return respObj, err
	}
	log.Println(string(resp))
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		log.Println("LayerOneBetFair: json.Unmarshal error", err.Error())
		log.Println("LayerOneBetFair: response data is - ", string(resp))
		return respObj, err
	}
	return respObj, nil
}

func LayerOneDream() (responsedto.BasicRespDto, error) {
	respObj := responsedto.BasicRespDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "POST", LoadTestDreamUrl, TimeOut)
	if err != nil {
		log.Println("LayerOneDream: SportsFeedCall failed with error - ", err.Error())
		return respObj, err
	}
	log.Println(string(resp))
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		log.Println("LayerOneDream: json.Unmarshal error", err.Error())
		log.Println("LayerOneDream: response data is - ", string(resp))
		return respObj, err
	}
	return respObj, nil
}

func LayerOneSportRadar() (responsedto.BasicRespDto, error) {
	respObj := responsedto.BasicRespDto{}
	resp, err := provider.SportsFeedCall([]byte{}, "POST", LoadTestSPUrl, TimeOut)
	if err != nil {
		log.Println("LayerOneSP: SportsFeedCall failed with error - ", err.Error())
		return respObj, err
	}
	log.Println(string(resp))
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		log.Println("LayerOneSP: json.Unmarshal error", err.Error())
		log.Println("LayerOneSP: response data is - ", string(resp))
		return respObj, err
	}
	return respObj, nil
}
