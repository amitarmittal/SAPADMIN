package betfair

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	"Sp/dto/providers/betfair"
	betfairdto "Sp/dto/providers/betfair"
	dreamdto "Sp/dto/providers/dream"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"Sp/dto/sports"
	provider "Sp/providers"
	betfairmodule "Sp/providers/betfairModule"
	"Sp/providers/betfairModule/request"
	"Sp/providers/betfairModule/response"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	//BaseUrl             string        = "https://stage-feed.mysportsfeed.io/api/v1/betfair-api/"
	//BaseUrl             string        = "https://feed-dev.hypexone.com/api/v1/"
	// BaseUrl             string        = "https://api.indisports.live/api/v1/"
	// BaseUrl string = "https://feed.mysportsfeed.io/api/v1/"
	BaseUrl               string        = os.Getenv("L2_API_URL")
	GetSportsUrl          string        = BaseUrl + "list-sports/?providerId=BetFair"
	GetCompetitionsUrl    string        = BaseUrl + "list-competitions/?providerId=BetFair"
	GetEventsUrl          string        = BaseUrl + "list-events/?providerId=BetFair"
	GetLiveEventsUrl      string        = BaseUrl + "list-live-events/?providerId=BetFair"
	GetMarketsUrl         string        = BaseUrl + "list-all-markets/?providerId=BetFair"
	ValidateOddsUrl       string        = BaseUrl + "validate-odds/?providerId=BetFair"
	PlaceOrderUrl         string        = BaseUrl + "place-order/?providerId=BetFair"
	PlaceOrderUrlAsync    string        = BaseUrl + "place-order-async/"
	PlaceOrderUrlStatus   string        = BaseUrl + "place-order-status/"
	CancelOrderUrl        string        = BaseUrl + "cancel-order/?providerId=BetFair"
	GetClearedOrdersUrl   string        = BaseUrl + "list-cleared-order/?providerId=BetFair"
	GetCurrentOrdersUrl   string        = BaseUrl + "list-current-orders/?providerId=BetFair"
	GetMarketBookUrl      string        = BaseUrl + "list-market-books?providerId=BetFair"
	GetMarketResultUrl    string        = BaseUrl + "get-market-result?providerId=BetFair"
	GetMatchedSportUrl    string        = BaseUrl + "/matched-sport?sportid="
	GetUnmatchedSportUrl  string        = BaseUrl + "/unmatched-sport?sportid="
	GetAllSportUrl        string        = BaseUrl + "/all-sport?sportid="
	GetUpdateCardUrl      string        = BaseUrl + "/sport-card/update"
	GetSportCardsUrl      string        = BaseUrl + "/score-card/sport-radar?sportid="
	TimeOut               time.Duration = 15 // 5 seconds
	ProviderId            string        = "BetFair"
	ProviderName          string        = "Bet Fair"
	POStatusCheckInterval int           = 250  //ms
	ResponseTimeInterval  int           = 6000 //ms
	KF_BF_L2              string        = os.Getenv("KF_BF_L2")
)

func GetSports() {
	//log.Println("GetSports: BetFair Start.")
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
	//log.Println("GetCompetitions: BetFair Start.")
	sports, err := database.GetSports(ProviderId)
	if err != nil {
		log.Println("GetCompetitions: GetSports failed with error - ", err.Error())
		return
	}
	if len(sports) == 0 {
		log.Println("GetCompetitions: GetSports returned records - ", len(sports))
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
	//log.Println("GetUpcomingEvents: BetFair Start.")
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
			log.Println("GetUpcomingEvents: BetFair Sport is not ACTIVE - ", sport.SportId)
			continue
		}
		events, err := GetEvents(sport.SportId)
		if err != nil {
			log.Println("GetUpcomingEvents: Failed to fetch BetFair events for sport - ", sport.SportId)
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
	//log.Println("GetLiveEvents: BetFair sport count - ", len(sportIds))
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
	//log.Println("GetEvents: BetFair  for sport - ", sportId)
	//log.Println("GetEvents: BetFair Start.")
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
	//log.Println(string(resp))
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

func GetMarkets(sportId string, eventId string, operatorId string) dto.EventDto {
	//log.Println("GetMarkets: BetFair Start.")
	operator, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetMarkets: cache.GetOperatorDetails failed with error - ", err.Error())
		operator = operatordto.OperatorDTO{}
	}
	event := dto.EventDto{}
	getMarketsUrl := GetMarketsUrl
	if operatorId == "kia-sap" || operatorId == "kiaexch" {
		if KF_BF_L2 != "" {
			getMarketsUrl = KF_BF_L2 + "list-all-markets/?providerId=BetFair"
		}
	}
	if operator.BetFairFeed != "" {
		getMarketsUrl = operator.BetFairFeed + "list-all-markets/?providerId=BetFair"
	}
	resp, err := provider.SportsFeedCall([]byte{}, "GET", getMarketsUrl+"&sportId="+sportId+"&eventId="+eventId, TimeOut)
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
	//log.Println("ValidateOdds: BetFair Start.")
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
	if betReq.MarketType == constants.SAP.MarketType.LINE_ODDS() {
		reqDto.OddValue = betReq.SessionOutcome
	}
	reqDto.SessionOutcome = betReq.SessionOutcome

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("ValidateOdds: Failed to convert DTO to JSON")
		return respDto, err
	}
	log.Println("ValidateOdds: PlaceBet: request data is - ", string(jsonData))
	validateOddsUrl := ValidateOddsUrl
	if betReq.OperatorId == "kia-sap" || betReq.OperatorId == "kiaexch" {
		if KF_BF_L2 != "" {
			validateOddsUrl = KF_BF_L2 + "validate-odds/?providerId=BetFair"
		}
	}
	if operator.BetFairFeed != "" {
		validateOddsUrl = operator.BetFairFeed + "validate-odds/?providerId=BetFair"
	}
	resp, err := provider.SportsFeedCall(jsonData, "POST", validateOddsUrl, TimeOut)
	if err != nil {
		log.Println("ValidateOdds: SportsFeedCall failed with error - ", err.Error())
		log.Println("ValidateOdds: request data is - ", string(jsonData))
		return respDto, err
	}
	//log.Println(string(resp))
	log.Println("ValidateOdds: PlaceBet: response data is - ", string(resp))
	err = json.Unmarshal(resp, &respDto)
	if err != nil {
		log.Println("ValidateOdds: json.Unmarshal error", err.Error())
		log.Println("ValidateOdds: response data is - ", string(resp))
		return respDto, err
	}
	return respDto, nil
}

func PlaceOrder(reqDto requestdto.PlaceBetReqDto, betDto sports.BetDto) (response.PlaceInstructionReport, error) {
	//log.Println("PlaceOrder: BetFair Start.")
	// 0. Default response object
	placeOrderResp := response.PlaceInstructionReport{}
	// CustomerOrderRef length should be <= 32
	customerOrderRef := strings.ReplaceAll(betDto.BetId, "-", "") // Sending SAP betId (Trimmed version) for Async
	// 1. Request Object
	placeOrderReq := request.NewPlaceOrderReq(reqDto, customerOrderRef) //betfairdto.NewPlaceOrderReq(reqDto)
	// jsonData, err := json.Marshal(placeOrderReq)
	// if err != nil {
	// 	log.Println("PlaceOrder: Failed to convert DTO to JSON")
	// 	return placeOrderResp, err
	// }
	// log.Println("PlaceOrder: PlaceBet: APITiming: request data is - ", string(jsonData))
	// resp, err := provider.SportsFeedCall(jsonData, "POST", PlaceOrderUrl, TimeOut)
	// if err != nil {
	// 	log.Println("PlaceOrder: SportsFeedCall failed with error - ", err.Error())
	// 	log.Println("PlaceOrder: request data is - ", string(jsonData))
	// 	return placeOrderResp, err
	// }
	// log.Println("PlaceOrder: PlaceBet: APITiming: response data is - ", string(resp))
	// bfPlaceOrderResp := betfairdto.BFPlaceOrderResp{}
	// err = json.Unmarshal(resp, &bfPlaceOrderResp)
	// if err != nil {
	// 	log.Println("PlaceOrder: json.Unmarshal error", err.Error())
	// 	log.Println("PlaceOrder: response data is - ", string(resp))
	// 	return placeOrderResp, err
	// }
	// // 4. Return Response
	// if bfPlaceOrderResp.Status != "RS_OK" {
	// 	log.Println("PlaceOrder: Failed with error - ", bfPlaceOrderResp.ErrorDescription)
	// 	return placeOrderResp, fmt.Errorf(bfPlaceOrderResp.ErrorDescription)
	// }
	// return bfPlaceOrderResp.PlaceOrderResp, nil
	placeOrderResp2, err := betfairmodule.PlaceOrders(placeOrderReq)
	if err != nil { // returned error
		log.Println("PlaceBet: PlaceOrder: betfairmodule.PlaceOrders failed with error - ", err.Error())
		return placeOrderResp, err
	}
	if placeOrderResp2.Status != "SUCCESS" { // returned failure
		if len(placeOrderResp2.InstructionReports) == 0 { // top level failure
			log.Println("PlaceBet: PlaceOrder: betfairmodule.PlaceOrders Failed with Status & Error Code - ", placeOrderResp2.Status, placeOrderResp2.ErrorCode)
			return placeOrderResp, fmt.Errorf(placeOrderResp2.ErrorCode)
		}
		// inner level failure
		instructionReport := placeOrderResp2.InstructionReports[0]
		log.Println("PlaceBet: PlaceOrder: betfairmodule.PlaceOrders InstructionReport failed with Status & ErrorCode - ", instructionReport.Status, instructionReport.ErrorCode)
		return placeOrderResp, fmt.Errorf(instructionReport.ErrorCode)
	}
	// returned success
	return placeOrderResp2.InstructionReports[0], nil
}

// func PlaceOrder2(reqDto requestdto.PlaceBetReqDto, betDto sports.BetDto, ubs models.UserBetStatusDto) (betfairdto.PlaceInstructionReport, error) {
// 	//log.Println("PlaceBet: PlaceOrder: BetFair Start.")
// 	// 0. Default response object
// 	placeOrderResp := betfairdto.PlaceInstructionReport{}
// 	// CustomerOrderRef length should be <= 32
// 	customerOrderRef := strings.ReplaceAll(betDto.BetId, "-", "") // Sending SAP betId (Trimmed version) for Async
// 	customerRef := reqDto.OperatorId + "-" + betDto.UserId
// 	// 1. Request Object
// 	placeOrderReq := betfairdto.NewPlaceOrderReqAsync(reqDto, customerOrderRef, customerRef)
// 	jsonData, err := json.Marshal(placeOrderReq)
// 	if err != nil {
// 		log.Println("PlaceBet: PlaceOrder: Failed to convert DTO to JSON")
// 		return placeOrderResp, err
// 	}
// 	log.Println("PlaceBet: PlaceOrder: APITiming: request data is - ", string(jsonData))
// 	resp, err := provider.SportsFeedCall(jsonData, "POST", PlaceOrderUrlAsync, TimeOut)
// 	if err != nil {
// 		log.Println("PlaceBet: PlaceOrder: SportsFeedCall failed with error - ", err.Error())
// 		log.Println("PlaceBet: PlaceOrder: request data is - ", string(jsonData))
// 		return placeOrderResp, err
// 	}
// 	log.Println("PlaceBet: PlaceOrder: APITiming: response data is - ", string(resp))
// 	bfPlaceOrderAsyncResp := betfairdto.PlaceOrderAsyncResp{}
// 	err = json.Unmarshal(resp, &bfPlaceOrderAsyncResp)
// 	if err != nil {
// 		log.Println("PlaceBet: PlaceOrder: json.Unmarshal error", err.Error())
// 		log.Println("PlaceBet: PlaceOrder: response data is - ", string(resp))
// 		return placeOrderResp, err
// 	}
// 	// 4. Check Async Call Status
// 	if bfPlaceOrderAsyncResp.Status != "RS_OK" {
// 		log.Println("PlaceBet: PlaceOrder: Failed with error - ", bfPlaceOrderAsyncResp.ErrorDescription)
// 		return placeOrderResp, fmt.Errorf(bfPlaceOrderAsyncResp.ErrorDescription)
// 	}
// 	// 5. Call place-order-status
// 	log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus loop STARTED for betId - ", betDto.BetId)
// 	reponseTime := time.Now().Add(time.Duration(ResponseTimeInterval * int(time.Millisecond)))
// 	isResponseTime := false
// 	for i := 0; i < 200; i++ {
// 		// 5.1. Delay
// 		time.Sleep(time.Duration(POStatusCheckInterval * int(time.Millisecond)))
// 		// 5.2. Place Order Status call
// 		placeOrderStatusResp, err := PlaceOrderStatus(customerOrderRef)
// 		if err != nil {
// 			// 5.2.1. Call failed. Retry after some additional delay
// 			log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus failed with error - ", err.Error())
// 			log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus failed for betId - ", betDto.BetId)
// 			time.Sleep(time.Duration(POStatusCheckInterval * int(time.Millisecond))) // Additional delay for failure
// 			continue
// 		}
// 		// 5.3. PlaceOrder call completed
// 		if len(placeOrderStatusResp) == 0 {
// 			log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus returned EMPTY list for - ", betDto.BetId)
// 			time.Sleep(time.Duration(POStatusCheckInterval * int(time.Millisecond))) // Additional delay for failure
// 			continue
// 		}
// 		for _, posResp := range placeOrderStatusResp {
// 			if posResp.Instruction.CustomerOrderRef == customerOrderRef {
// 				if posResp.Status != "IN_PROGRESS" {
// 					log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus loop ENDED for betId - ", betDto.BetId)
// 					return posResp, nil
// 				}
// 				break
// 			}
// 		}
// 		// 5.4. PlaceOrder call not completed, Check for Response Time
// 		if isResponseTime {
// 			continue
// 		}
// 		curTime := time.Now()
// 		if curTime.After(reponseTime) {
// 			log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus Response Time completed for betId - ", betDto.BetId)
// 			// 5.4.1. Update Bet, so that it will start appearing in Unmatched Bets for user
// 			betDto.Status = constants.SAP.BetStatus.INPROCESS() // "INPROCESS"
// 			err = database.UpdateBet(betDto)
// 			if err != nil {
// 				log.Println("PlaceBet: PlaceOrder: database.UpdateBet failed with error - ", err.Error())
// 				continue
// 			}
// 			// 5.4.2. Update UserBetStatus, so that it will stop spinning in frontend
// 			ubs.ErrorMessage = ""
// 			ubs.Status = "COMPLETED"
// 			err = database.UpsertUserBetStatus(ubs) // Log error
// 			if err != nil {
// 				log.Println("PlaceBet: PlaceOrder: database.UpsertUserBetStatus failed with error - ", err.Error())
// 				continue
// 			}
// 			isResponseTime = true
// 		}
// 	}
// 	log.Println("PlaceBet: PlaceOrder: PlaceOrderStatus loop FAILED with TimeOut - ", betDto.BetId)
// 	return placeOrderResp, fmt.Errorf("PlaceOrder failed with TimeOut - %s", betDto.BetId)
// }

// func PlaceOrderAsync(placeOrderReq request.PlaceOrderReq) {
// 	placeOrderResp2, err := betfairmodule.PlaceOrders(placeOrderReq)
// 	if err != nil {
// 		log.Println("PlaceOrder: betfairmodule.PlaceOrders failed with error - ", err.Error())
// 		// TODO: Error Handling
// 		return
// 	}
// 	// Handle response
// 	return
// }

func PlaceOrderStatus(customerOrderRef string) ([]betfairdto.PlaceInstructionReport, error) {
	// 0. Default response object
	placeOrderResps := []betfairdto.PlaceInstructionReport{}
	// 1. Request Object
	placeOrderStatusReq := betfairdto.PlaceOrderStatusReq{}
	placeOrderStatusReq.BetIds = append(placeOrderStatusReq.BetIds, customerOrderRef)
	jsonData, err := json.Marshal(placeOrderStatusReq)
	if err != nil {
		log.Println("PlaceOrderStatus: Failed to convert DTO to JSON with error - ", err.Error())
		return placeOrderResps, err
	}
	log.Println("PlaceOrderStatus: PlaceBet: APITiming: request data is - ", string(jsonData))
	resp, err := provider.SportsFeedCall(jsonData, "POST", PlaceOrderUrlStatus, TimeOut)
	if err != nil {
		log.Println("PlaceOrderStatus: SportsFeedCall failed with error - ", err.Error())
		log.Println("PlaceOrderStatus: request data is - ", string(jsonData))
		return placeOrderResps, err
	}
	log.Println("PlaceOrderStatus: PlaceBet: APITiming: response data is - ", string(resp))
	bfPlaceOrderStatusResps := betfair.PlaceOrderStatusResp{}
	err = json.Unmarshal(resp, &bfPlaceOrderStatusResps)
	if err != nil {
		log.Println("PlaceOrderStatus: json.Unmarshal error", err.Error())
		log.Println("PlaceOrderStatus: response data is - ", string(resp))
		return placeOrderResps, err
	}
	if bfPlaceOrderStatusResps.Status != "RS_OK" {
		log.Println("PlaceBet: PlaceOrder: Failed with error - ", bfPlaceOrderStatusResps.ErrorDescription)
		return placeOrderResps, fmt.Errorf(bfPlaceOrderStatusResps.ErrorDescription)
	}
	return bfPlaceOrderStatusResps.PlaceInstructionReports, nil
}

func CancelOrder(reqDto requestdto.CancelBetReqDto, betDtos []sports.BetDto, betFairRate int) (response.CancelOrdersResp, error) {
	//log.Println("CancelOrder: BetFair Start.")
	// 0. Default response object
	cancelOrderResp := response.CancelOrdersResp{}
	// 1. Request Object
	cancelOrderReq := request.NewCancelOrderReq2(reqDto, betDtos, betFairRate)
	// jsonData, err := json.Marshal(cancelOrderReq)
	// if err != nil {
	// 	log.Println("CancelOrder: Failed to convert DTO to JSON")
	// 	return cancelOrderResp, err
	// }
	// // 2. Make Cancel Order request
	// log.Println("CancelOrder: request data is - ", string(jsonData))
	// resp, err := provider.SportsFeedCall(jsonData, "POST", CancelOrderUrl, TimeOut)
	// if err != nil {
	// 	log.Println("CancelOrder: SportsFeedCall failed with error - ", err.Error())
	// 	log.Println("CancelOrder: request data is - ", string(jsonData))
	// 	return cancelOrderResp, err
	// }
	// //log.Println(string(resp))
	// log.Println("CancelOrder: response data is - ", string(resp))
	// // 3. Unmarhsal to response object
	// bfCancelOrderResp := betfairdto.BFCancelOrderResp{}
	// err = json.Unmarshal(resp, &bfCancelOrderResp)
	// if err != nil {
	// 	log.Println("CancelOrder: json.Unmarshal error", err.Error())
	// 	log.Println("CancelOrder: response data is - ", string(resp))
	// 	return cancelOrderResp, err
	// }
	// // 4. Return Response
	// if bfCancelOrderResp.Status != "RS_OK" {
	// 	log.Println("CancelOrder: Failed with error - ", bfCancelOrderResp.ErrorDescription)
	// 	return cancelOrderResp, fmt.Errorf(bfCancelOrderResp.ErrorDescription)
	// }
	cancelOrderResp, err := betfairmodule.CancelOrders(cancelOrderReq)
	if err != nil {
		log.Println("CancelBet: CancelOrder: betfairmodule.CancelOrders failed with error - ", err.Error())
		return cancelOrderResp, err
	}
	return cancelOrderResp, nil
}

func CurrentOrders() (response.CurrentOrdersResp, error) {
	//log.Println("CurrentOrders: BetFair Start.")
	// 0. Default response object
	currentOrdersResp := response.CurrentOrdersResp{}
	// 1. Request Object
	currentOrdersReq := request.ListCurrentOrdersReq{} // commondto.ListCurrentOrdersDto{}
	// jsonData, err := json.Marshal(currentOrdersReq)
	// if err != nil {
	// 	log.Println("CurrentOrders: DTO to JSON failed with error - ", err.Error())
	// 	return currentOrdersResp, err
	// }
	// // 2. Make Cancel Order request
	// resp, err := provider.SportsFeedCall(jsonData, "POST", GetCurrentOrdersUrl, TimeOut)
	// if err != nil {
	// 	log.Println("CurrentOrders: SportsFeedCall failed with error - ", err.Error())
	// 	log.Println("CurrentOrders: request data is - ", string(jsonData))
	// 	return currentOrdersResp, err
	// }
	// //log.Println(string(resp))
	// // 3. Unmarhsal to response object
	// bfCurrentOrdersResp := commondto.BFCurrentOrdersResp{}
	// err = json.Unmarshal(resp, &bfCurrentOrdersResp)
	// if err != nil {
	// 	log.Println("CurrentOrders: json.Unmarshal error - ", err.Error())
	// 	log.Println("CurrentOrders: response data is - ", string(resp))
	// 	return currentOrdersResp, err
	// }
	// log.Println("CurrentOrders: response data is - ", string(resp))
	// // 4. Return Response
	// if bfCurrentOrdersResp.Status != "RS_OK" {
	// 	log.Println("CurrentOrders: Failed with error - ", bfCurrentOrdersResp.ErrorDescription)
	// 	return currentOrdersResp, fmt.Errorf(bfCurrentOrdersResp.ErrorDescription)
	// }
	// return bfCurrentOrdersResp.CurrentOrdersResp, nil
	currentOrdersResp, err := betfairmodule.CurrentOrders(currentOrdersReq)
	if err != nil {
		log.Println("CurrentOrders: betfairmodule.CurrentOrders failed with error - ", err.Error())
		return currentOrdersResp, err
	}
	return currentOrdersResp, nil
}

func ClearedOrders(betStatus string, betIds []string, bets []sports.BetDto) (response.ClearedOrdersResp, error) {
	//log.Println("ClearedOrders: BetFair Start.")
	// 0. Default response object
	clearedOrdersResp := response.ClearedOrdersResp{}
	// 1. Request Object
	clearedOrdersReq := request.ListClearedOrdersReq{}
	clearedOrdersReq.BetIds = []string{}
	clearedOrdersReq.BetIds = append(clearedOrdersReq.BetIds, betIds...)
	clearedOrdersReq.BetStatus = betStatus
	// jsonData, err := json.Marshal(clearedOrdersReq)
	// if err != nil {
	// 	log.Println("ClearedOrders: DTO to JSON failed with error - ", err.Error())
	// 	return clearedOrdersResp, err
	// }
	// // 2. Make Market Book request
	// resp, err := provider.SportsFeedCall(jsonData, "POST", GetClearedOrdersUrl, TimeOut)
	// if err != nil {
	// 	log.Println("ClearedOrders: Cleared Orders failed with error - ", err.Error())
	// 	log.Println("ClearedOrders: request data is - ", string(jsonData))
	// 	return clearedOrdersResp, err
	// }
	// //log.Println(string(resp))
	// // 3. Unmarhsal to response object
	// bfClearedOrdersResp := commondto.BFClearedOrdersResp{}
	// err = json.Unmarshal(resp, &bfClearedOrdersResp)
	// if err != nil {
	// 	log.Println("ClearedOrders: json.Unmarshal error - ", err.Error())
	// 	log.Println("ClearedOrders: response data is - ", string(resp))
	// 	return clearedOrdersResp, err
	// }
	// // 4. Return Response
	// if bfClearedOrdersResp.Status != "RS_OK" {
	// 	log.Println("ClearedOrders: Failed with error - ", bfClearedOrdersResp.ErrorDescription)
	// 	return clearedOrdersResp, fmt.Errorf(bfClearedOrdersResp.ErrorDescription)
	// }
	// return bfClearedOrdersResp.ClearedOrdersResp, nil
	clearedOrdersResp, err := betfairmodule.ClearedOrders(clearedOrdersReq)
	if err != nil {
		log.Println("ClearedOrders: betfairmodule.ClearedOrders failed with error - ", err.Error())
		return clearedOrdersResp, err
	}
	return clearedOrdersResp, nil
}

func GetMarketBook(marketIds []string) (commondto.ListMarketRespDto, error) {
	//log.Println("GetMarketBook: BetFair Start.")
	// 0. Default response object
	marketBookResp := commondto.ListMarketRespDto{}
	// 1. Create Request object
	marketBookReq := commondto.ListMarketReqDto{}
	marketBookReq.MarketIds = marketIds
	jsonData, err := json.Marshal(marketBookReq)
	if err != nil {
		log.Println("GetMarketBook: DTO to JSON failed with error - ", err.Error())
		return marketBookResp, err
	}
	// 2. Make Market Book request
	resp, err := provider.SportsFeedCall(jsonData, "POST", GetMarketBookUrl, TimeOut)
	if err != nil {
		log.Println("GetMarketBook: MarketBook API call failed with error - ", err.Error())
		log.Println("GetMarketBook: request data is - ", string(jsonData))
		return marketBookResp, err
	}
	//log.Println(string(resp))
	// 3. Unmarhsal to response object
	err = json.Unmarshal(resp, &marketBookResp)
	if err != nil {
		log.Println("GetMarketBook: json.Unmarshal error - ", err.Error())
		log.Println("GetMarketBook: response data is - ", string(resp))
		return marketBookResp, err
	}
	// 4. Return Response
	return marketBookResp, nil
}

func GetMarketResults(getMarketResultReqs []betfair.GetMarketResultReq) ([]betfair.GetMarketResultResp, error) {
	//log.Println("GetMarketResults: BetFair markets count - ", len(getMarketResultReqs))
	// 0. Default response object
	respObjs := []betfair.GetMarketResultResp{}
	// 1. Reqest object
	for _, getMarketResultReq := range getMarketResultReqs {
		respObj := betfair.GetMarketResultResp{}
		//respObj.Status = "RS_ERROR"
		//respObj.Message = "Default Error from SAP"
		//respObj.MarketResult = betfair.MarketResult{}
		jsonData, err := json.Marshal(getMarketResultReq)
		if err != nil {
			log.Println("GetMarketResults: Failed to convert DTO to JSON with error - ", err.Error())
			//respObjs = append(respObjs, respObj)
			continue
		}
		log.Println("GetMarketResults: request data is - ", string(jsonData))
		resp, err := provider.SportsFeedCall(jsonData, "POST", GetMarketResultUrl, TimeOut)
		if err != nil {
			log.Println("GetMarketResults: SportsFeedCall failed with error - ", err.Error())
			log.Println("GetMarketResults: request data is - ", string(jsonData))
			//respObjs = append(respObjs, respObj)
			continue
		}
		log.Println("GetMarketResults: response data is - ", string(resp))
		err = json.Unmarshal(resp, &respObj)
		if err != nil {
			log.Println("GetMarketResults: json.Unmarshal error - ", err.Error())
			log.Println("GetMarketResults: response data is - ", string(resp))
			//respObjs = append(respObjs, respObj)
			continue
		}
		if respObj.Status != "RS_OK" {
			respObj.MarketResult = betfairdto.MarketResult{}
			respObj.MarketResult.MarketId = getMarketResultReq.MarketId
		}
		respObjs = append(respObjs, respObj)
	}
	return respObjs, nil
}

func MatchedSportsbySportId(sportId string) ([]responsedto.MatchedSport, error) {
	var matchedSport []responsedto.MatchedSport
	var reqBody []byte
	// resp, err := ExternalCall(reqBody, "GET", sportId, 5)
	url := GetMatchedSportUrl + sportId
	resp, err := provider.SportsFeedCall(reqBody, "GET", url, TimeOut)
	if err != nil {
		log.Println("MatchedSportsbySportId: ExternalCall failed with error - ", err.Error())
		return matchedSport, err
	}
	err = json.Unmarshal(resp, &matchedSport)
	if err != nil {
		log.Println("MatchedSportsbySportId: json.Unmarshal error", err.Error())
		log.Println("MatchedSportsbySportId: response data is - ", string(resp))
	}
	return matchedSport, err
}

func UnMatchedSportsBySportId(sportId string) ([]responsedto.UnmatchedSport, error) {
	var unmatchedSport []responsedto.UnmatchedSport
	var reqBody []byte
	url := GetUnmatchedSportUrl + sportId
	resp, err := provider.SportsFeedCall(reqBody, "GET", url, TimeOut)
	if err != nil {
		log.Println("UnMatchedSportsBySportId: ExternalCall failed with error - ", err.Error())
		return unmatchedSport, err
	}
	err = json.Unmarshal(resp, &unmatchedSport)
	if err != nil {
		log.Println("UnMatchedSportsBySportId: json.Unmarshal error", err.Error())
		log.Println("UnMatchedSportsBySportId: response data is - ", string(resp))
	}
	return unmatchedSport, err
}

func AllSportsBySportId(sportId string) ([]responsedto.AllSport, error) {
	var sports []responsedto.AllSport
	var reqBody []byte
	url := GetAllSportUrl + sportId
	resp, err := provider.SportsFeedCall(reqBody, "GET", url, TimeOut)
	if err != nil {
		log.Println("AllSportsBySportId: ExternalCall failed with error - ", err.Error())
		return sports, err
	}
	err = json.Unmarshal(resp, &sports)
	if err != nil {
		log.Println("AllSportsBySportId: json.Unmarshal error", err.Error())
		log.Println("AllSportsBySportId: response data is - ", string(resp))
	}
	return sports, err
}

func UpdateCard(reqBody []byte) (responsedto.UpdateSportCardResp, error) {
	var respDto responsedto.UpdateSportCardResp
	resp, err := provider.SportsFeedCall(reqBody, "POST", GetUpdateCardUrl, TimeOut)
	if err != nil {
		log.Println("UpdateCard: ExternalCall failed with error - ", err.Error())
		return respDto, err
	}
	err = json.Unmarshal(resp, &respDto)
	if err != nil {
		log.Println("UpdateCard: json.Unmarshal error", err.Error())
		log.Println("UpdateCard: response data is - ", string(resp))
	}
	return respDto, err
}

func GetSportRadarCards(sportId string) ([]responsedto.AllSport, error) {
	var respDto []responsedto.AllSport
	var reqBody []byte
	url := GetSportCardsUrl + sportId
	resp, err := provider.SportsFeedCall(reqBody, "GET", url, TimeOut)
	if err != nil {
		log.Println("GetSportRadarCards: ExternalCall failed with error - ", err.Error())
		return respDto, err
	}
	err = json.Unmarshal(resp, &respDto)
	if err != nil {
		log.Println("GetSportRadarCards: json.Unmarshal error", err.Error())
		log.Println("GetSportRadarCards: response data is - ", string(resp))
	}
	return respDto, err
}
