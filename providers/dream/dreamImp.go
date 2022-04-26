package dream

import (
	"Sp/constants"
	"Sp/database"
	"Sp/dto/models"
	dreamdto "Sp/dto/providers/dream"
	"Sp/dto/requestdto"
	"Sp/dto/sports"
	"Sp/operator"
	utils "Sp/utilities"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	uuid2 "github.com/google/uuid"
)

var (
	// time in milli seconds
	CricketMatchOdds int = 3500
	CricketBookmaker int = 1500
	CricetFancy      int = 1500
	SoccerMatchOdds  int = 4500
	TennisMatchOdds  int = 4500
)

/*
func GetLiveEvents(sportsList []string) ([]dto.EventDto, error) {
	eventsList := []dto.EventDto{}
	var err error = nil
	for _, sport := range sportsList {
		events, err := GetEvents(sport)
		if err != nil {
			log.Println("GetLiveEvents: Failed to fetch events for sport - ", sport)
			log.Println("GetLiveEvents: Error is - ", err.Error())
			continue
		}
		if len(events) > 0 {
			for _, event := range events {
				if event.Status == "IN_PLAY" {
					eventsList = append(eventsList, event)
				}
			}
		}
	}
	return eventsList, err
}
*/
func PlaceBet(reqDto requestdto.PlaceBetReqDto) error {
	// I. Validate Odds
	// II. Is match live or two days from current date

	// implementaion
	// 1. Validate Odds (#I)
	respObj, err := ValidateOdds(reqDto)
	if err != nil {
		// 1.1. Failed to validate odds
		log.Println("PlaceBet: Dream Odds Validation call failed with - ", err.Error())
		return fmt.Errorf(respObj.ErrorDescription)
	}
	//log.Println("PlaceBet: Event Open Date is - ", time.Unix(respObj.OpenDate/1000, 0))
	if !respObj.IsValid {
		// 1.2. Invalid Odds
		log.Println("PlaceBet: Invalid Odds - ", reqDto.OddValue)
		return fmt.Errorf("Odds Changed!")
	}
	// 2. Is match live or two days from current date (#II)
	if reqDto.MarketType == "MATCH_ODDS" && respObj.Status != "IN_PLAY" {
		// 2.1. Match Odds bet on upcoming match
		log.Println("PlaceBet: MATCH_ODDS bet only allowed on IN_PLAY events - ", respObj.Status)
		return fmt.Errorf("Betting NOT ALLOWED NOT IN_PLAY!")
	}
	if respObj.Status != "IN_PLAY" && time.Now().Add(time.Hour*24*2).Unix() < (respObj.OpenDate/1000) {
		// 2.2. Upcoming match & more than 48 hours to start
		log.Println("PlaceBet: TOO EARLY TO BET - ", time.Unix(respObj.OpenDate/1000, 0))
		return fmt.Errorf("Betting NOT ALLOWED NOT IN_PLAY!")
	}
	return nil
}

func DreamSettledOrders(reqDto dreamdto.ResultReqDto) {
	// 0. Get Market from DB
	eventKey := constants.SAP.ProviderType.Dream() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	log.Println("DreamResult: DreamSettledOrders START for MarketKey is - ", marketKey)
	// 1. Send Market Result to Operator for MarketLevel Settlement for feed operators
	go operator.MarketResult(constants.SAP.ProviderType.Dream(), reqDto.SportId, reqDto.EventId, reqDto.MarketId, reqDto.MarketType, reqDto.RunnerId, reqDto.RunnerName, reqDto.SessionOutcome)
	// 1. Get Market from DB
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("DreamResult: database.GetMarket failed with error - ", err.Error())
		log.Println("DreamResult: database.GetMarket failed for Market - ", marketKey)
		isMarketFound = false
		//return
	}
	// 2. Add a result to the Market and save in DB
	if isMarketFound == true {
		result := models.Result{}
		result.RunnerId = reqDto.RunnerId
		for _, runner := range market.Runners {
			if reqDto.RunnerId == runner.RunnerId {
				result.RunnerName = runner.RunnerName
				break
			}
		}
		result.SessionOutcome = reqDto.SessionOutcome
		result.ResultTime = time.Now().Unix()
		market.Results = append(market.Results, result)
		market.MarketStatus = constants.SportRadar.MarketStatus.SETTLED()
		// 2.1. Update Market
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("DreamResult: database.ReplaceMarket failed with error - ", err.Error())
			log.Println("DreamResult: database.ReplaceMarket failed for Market - ", marketKey)
			//return
		}
	}
	// 3. Get Open Bets
	openBets, _, err := database.GetOpenBetsByMarket(eventKey, reqDto.MarketId)
	if err != nil {
		log.Println("DreamResult: database.GetOpenBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(openBets) == 0 {
		log.Println("DreamResult: ZERO Open bets to settle for marketkey - ", marketKey)
		return
	}
	log.Println("DreamResult: openBets count is - ", len(openBets))
	// 4. Add Result to the Bet
	settledBets := []sports.BetDto{}
	for _, openBet := range openBets {
		// 4.1. update bet with result
		resultReq, err := ComputeResult(openBet, reqDto)
		if err != nil {
			log.Println("SettledOrders: providers.ComputeResult failed with error - ", err.Error())
			continue
		}
		openBet.ResultReqs = append(openBet.ResultReqs, resultReq)
		openBet.NetAmount += resultReq.CreditAmount
		openBet.Status = constants.BetFair.BetStatus.SETTLED()
		// setting updatedAt
		openBet.UpdatedAt = resultReq.ReqTime
		// 4.2. append bet to list
		settledBets = append(settledBets, openBet)
	}
	if len(settledBets) == 0 {
		log.Println("DreamResult: ZERO SettledBets to settle for market Id - ", reqDto.MarketId)
		return
	}
	log.Println("DreamResult: settledBets count is - ", len(settledBets))
	// 5. Process Settlement
	operator.CommonResultRoutine(settledBets)
	log.Println("DreamResult: result_settlement END for MarketKey is - ", marketKey)
	return
}

func DreamRollbackOrders(reqDto dreamdto.RollbackReqDto) {
	// 0. Get Market from DB
	eventKey := constants.SAP.ProviderType.Dream() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	log.Println("DreamRollback: DreamRollbackOrders START for MarketKey is - ", marketKey)
	// 1. Send Market Rollback to Operator for MarketLevel Rollback for feed operators
	go operator.MarketRollback(constants.SAP.ProviderType.Dream(), reqDto.SportId, reqDto.EventId, reqDto.MarketId, reqDto.MarketType, reqDto.MarketName, reqDto.RollbackType, reqDto.Reason)
	// 1. Get Market from DB
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("DreamRollback: database.GetMarket failed with error - ", err.Error())
		log.Println("DreamRollback: database.GetMarket failed for Market - ", marketKey)
		isMarketFound = false
		//return
	}
	// 2. Add a result to the Market and save in DB
	if isMarketFound == true {
		rollback := models.Rollback{}
		rollback.RollbackType = reqDto.RollbackType
		rollback.RollbackReason = reqDto.Reason
		rollback.RollbackTime = time.Now().UnixNano() / int64(time.Millisecond)
		market.Rollbacks = append(market.Rollbacks, rollback)
		market.MarketStatus = constants.SportRadar.MarketStatus.VOIDED()
		if reqDto.RollbackType == "Rollback" {
			market.MarketStatus = constants.SportRadar.MarketStatus.OPEN()
		}
		// 2.1. Update Market
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("DreamRollback: database.ReplaceMarket failed with error - ", err.Error())
			log.Println("DreamRollback: database.ReplaceMarket failed for Market - ", marketKey)
			//return
		}
	}
	// 3. Get Open Bets by EventKey and MarketId
	openBets, _, err := database.GetOpenBetsByMarket(eventKey, reqDto.MarketId)
	if err != nil {
		log.Println("DreamRollback: database.GetOpenBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(openBets) == 0 {
		log.Println("DreamRollback: ZERO Open bets to rollback for market Id - ", reqDto.MarketId)
		return
	}
	log.Println("DreamRollback: openBets count is - ", len(openBets))
	// 4. Settlement Logic can be async
	RollbackBets := []sports.BetDto{}
	for _, openBet := range openBets {
		// 4.1. update bet with result
		rollbackReq := ComputeRollback(openBet, reqDto)
		if err != nil {
			log.Println("DreamRollback: ComputeResult failed with error - ", err.Error())
			continue
		}
		openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
		openBet.NetAmount += rollbackReq.RollbackAmount
		if strings.ToUpper(reqDto.RollbackType) == constants.SAP.BetStatus.ROLLBACK() {
			openBet.Status = constants.SAP.BetStatus.OPEN()
		} else {
			openBet.Status = constants.SAP.BetStatus.VOIDED()
		}
		// 4.2. setting updatedAt
		openBet.UpdatedAt = rollbackReq.ReqTime
		// 4.3. append bet to list
		RollbackBets = append(RollbackBets, openBet)
	}
	log.Println("DreamRollback: RollbackBets count is - ", len(RollbackBets))
	operator.CommonRollbackRoutine(reqDto.RollbackType, RollbackBets)
	return
}

func ComputeResult(openBet sports.BetDto, reqDto dreamdto.ResultReqDto) (sports.ResultReqDto, error) {
	resultReq := sports.ResultReqDto{}
	resultReq.ReqId = uuid2.New().String()
	resultReq.ReqTime = time.Now().UnixMilli()
	resultReq.CreditAmount = 0
	resultReq.RunnerName = reqDto.RunnerName
	resultReq.SessionOutcome = reqDto.SessionOutcome
	oddValue := openBet.BetDetails.OddValue
	if strings.ToUpper(reqDto.MarketType) == constants.SAP.MarketType.BOOKMAKER() {
		log.Println("GetOddsFactor: BOOKMAKER bet!!!")
		oddValue = 1.0 + (oddValue * 0.01)
	}
	if strings.ToUpper(reqDto.MarketType) == constants.SAP.MarketType.FANCY() {
		log.Println("GetOddsFactor: FANCY bet!!!")
		oddValue = 1.0 + (oddValue * 0.01)
	}
	log.Println("ComputeResult: oddValue is - ", oddValue)

	switch strings.ToUpper(openBet.BetDetails.MarketType) {
	case "MATCH_ODDS", "BOOKMAKER":
		if openBet.BetDetails.BetType == "BACK" && openBet.BetDetails.RunnerId == reqDto.RunnerId {
			resultReq.CreditAmount = oddValue * float64(openBet.BetDetails.StakeAmount)
		}
		if openBet.BetDetails.BetType == "LAY" && openBet.BetDetails.RunnerId != reqDto.RunnerId {
			resultReq.CreditAmount = oddValue * float64(openBet.BetDetails.StakeAmount)
		}
	case "FANCY":
		if openBet.BetDetails.BetType == "BACK" && reqDto.SessionOutcome >= openBet.BetDetails.SessionOutcome {
			resultReq.CreditAmount = oddValue * float64(openBet.BetDetails.StakeAmount)
		}
		if openBet.BetDetails.BetType == "LAY" && reqDto.SessionOutcome < openBet.BetDetails.SessionOutcome {
			resultReq.CreditAmount = oddValue * float64(openBet.BetDetails.StakeAmount)
		}
	default:
		log.Println("ComputeResult: Unexpected MarketType - ", openBet.BetDetails.MarketType)
		return resultReq, fmt.Errorf("Invalid Market Type - " + openBet.BetDetails.MarketType)
	}
	resultReq.CreditAmount = utils.Truncate4Decfloat64(resultReq.CreditAmount)
	return resultReq, nil
}

func ComputeRollback(openBet sports.BetDto, reqDto dreamdto.RollbackReqDto) sports.RollbackReqDto {
	rollbackReq := sports.RollbackReqDto{}
	rollbackReq.ReqId = uuid2.New().String()
	rollbackReq.ReqTime = time.Now().UnixMilli()
	rollbackReq.RollbackReason = reqDto.Reason
	rollbackReq.RollbackAmount = 0 // positive means, deposit to user, negative means, deduct from user
	for _, result := range openBet.ResultReqs {
		rollbackReq.RollbackAmount -= result.CreditAmount
	}
	for _, rollback := range openBet.RollbackReqs {
		rollbackReq.RollbackAmount -= rollback.RollbackAmount
	}
	switch strings.ToLower(reqDto.RollbackType) {
	case "void", "voided", "cancelled", "deleted", "lapsed", "expired":
		rollbackReq.RollbackAmount += openBet.BetReq.DebitAmount
	case "rollback":
		// nothing to do here
	default:
		log.Println("ComputeRollback: Unexpected ResultType - ", reqDto.RollbackType)
	}
	rollbackReq.RollbackAmount = utils.Truncate4Decfloat64(rollbackReq.RollbackAmount)
	return rollbackReq
}

func DreamMarketSuspend(reqDto dreamdto.MarketControlReq) {
	reqJson, err := json.Marshal(reqDto)
	// 1. Get Market from DB
	markets, err := database.GetMarketsByEventId(reqDto.EventId, constants.SAP.ProviderType.Dream())
	if err != nil {
		log.Println("DreamMarketSuspend: ERROR database.GetMarketsByEventId failed with error for reqJson - ", string(reqJson), err.Error())
		return
	}
	// 2. Prepare list of suspended markets
	suspendedMarkets := []models.Market{}
	for _, market := range markets {
		// iterate through all markets
		if reqDto.MarketType == constants.SAP.MarketType.FANCY() {
			// Requested Fancy Market suspension
			if reqDto.SessionId == "-1" {
				// Requested all Fancy Markets
				if market.MarketType == constants.SAP.MarketType.FANCY() {
					// Fancy market
					if market.MarketStatus == "OPEN" {
						// market is open, add to the suspended list
						market.MarketStatus = "SUSPENDED"
						suspendedMarkets = append(suspendedMarkets, market)
					}
				}
			} else {
				// Requested only one fancy market suspension
				if market.MarketId == reqDto.SessionId {
					// specified fancy market
					if market.MarketStatus == "OPEN" {
						// market is open, add to the suspended list
						market.MarketStatus = "SUSPENDED"
						suspendedMarkets = append(suspendedMarkets, market)
						break
					}
				}
			}
		} else {
			// Requsted matchodds or bookmaker market
			if market.MarketId == reqDto.MarketId {
				// specified matchodds/bookmaker market
				if market.MarketStatus == "OPEN" {
					// market is open, add to the suspended list
					market.MarketStatus = "SUSPENDED"
					suspendedMarkets = append(suspendedMarkets, market)
					break
				}
			}
		}
	}
	if len(suspendedMarkets) == 0 {
		// TODO: Log, no markets for request
		log.Println("DreamMarketSuspend: ZERO markets to suspend for reqJson - ", string(reqJson))
		return
	}
	// 3. Call Dream edpoint to suspend markets
	for _, market := range suspendedMarkets {
		log.Println("DreamMarketSuspend: Suspending market for MarketKey is - ", market.MarketKey)
		// TODO: Call L1 Dream - SuspendMarket endpoint
	}
	// 4. Save Markets
	count, msgs := database.UpdateManyMarkets(suspendedMarkets)
	if count != len(suspendedMarkets) {
		// Some markets failed to update
		for _, msg := range msgs {
			// Log msg
			log.Println("DreamMarketSuspend: ERROR database.UpdateManyMarkets failed with error for marketKey - ", msg)
		}
	}
	// 5. Return
	return
}

func DreamMarketResume(reqDto dreamdto.MarketControlReq) {
	reqJson, err := json.Marshal(reqDto)
	// 1. Get Markets from DB
	markets, err := database.GetMarketsByEventId(reqDto.EventId, constants.SAP.ProviderType.Dream())
	if err != nil {
		log.Println("DreamMarketResume: ERROR database.GetMarketsByEventId failed with error for reqJson - ", string(reqJson), err.Error())
		// TODO: Call L1 Dream endpoint irrespective of missing markets in SAP
		return
	}
	// 2. Prepare list of resumed markets
	suspendedMarkets := []models.Market{}
	for _, market := range markets {
		// iterate through all markets
		if reqDto.MarketType == constants.SAP.MarketType.FANCY() {
			// Requested Fancy Market suspension
			if reqDto.SessionId == "-1" {
				// Requested all Fancy Markets
				if market.MarketType == constants.SAP.MarketType.FANCY() {
					// Fancy market
					if market.MarketStatus == "SUSPENDED" {
						// market is open, add to the suspended list
						market.MarketStatus = "OPEN"
						suspendedMarkets = append(suspendedMarkets, market)
					}
				}
			} else {
				// Requested only one fancy market suspension
				if market.MarketId == reqDto.SessionId {
					// specified fancy market
					if market.MarketStatus == "SUSPENDED" {
						// market is open, add to the suspended list
						market.MarketStatus = "OPEN"
						suspendedMarkets = append(suspendedMarkets, market)
						break
					}
				}
			}
		} else {
			// Requsted matchodds or bookmaker market
			if market.MarketId == reqDto.MarketId {
				// specified matchodds/bookmaker market
				if market.MarketStatus == "SUSPENDED" {
					// market is open, add to the suspended list
					market.MarketStatus = "OPEN"
					suspendedMarkets = append(suspendedMarkets, market)
					break
				}
			}
		}
	}
	if len(suspendedMarkets) == 0 {
		log.Println("DreamMarketResume: ZERO markets to suspend for reqJson - ", string(reqJson))
		// TODO: Call L1 Dream endpoint irrespective of missing markets in SAP
		return
	}
	// 3. Call Dream edpoint to resume markets
	for _, market := range suspendedMarkets {
		log.Println("DreamMarketResume: Suspending market for MarketKey is - ", market.MarketKey)
		// TODO: Call L1 Dream - ResumeMarket endpoint
	}
	// 4. Save Markets
	count, msgs := database.UpdateManyMarkets(suspendedMarkets)
	if count != len(suspendedMarkets) {
		// Some markets failed to update
		for _, msg := range msgs {
			// Log msg
			log.Println("DreamMarketResume: ERROR database.UpdateManyMarkets failed with error for marketKey - ", msg)
		}
	}
	// 5. Return
	return
}
