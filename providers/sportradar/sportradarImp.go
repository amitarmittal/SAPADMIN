package sportradar

import (
	"Sp/constants"
	"Sp/database"
	"Sp/dto/models"
	"Sp/dto/providers/sportradar"
	"Sp/dto/sports"
	"Sp/operator"
	"Sp/providers"
	utils "Sp/utilities"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// time in milli seconds
	CricketMatchOdds int     = 3500
	CricketBookmaker int     = 1500
	CricetFancy      int     = 1500
	SoccerMatchOdds  int     = 4500
	TennisMatchOdds  int     = 4500
	MinBetValue      float64 = 0.01 // Minimum bet value to call betfair API, if less than this, we will make it success
	SportRadarRate   int     = 100  // EUR
)

/*
func GetLiveEvents(sportsList []string) ([]dto.EventDto, error) {
	//log.Println("GetLiveEvents: SportRadar Start")
	eventsList := []dto.EventDto{}
	var err error = nil
	for _, sport := range sportsList {
		events, err := GetEvents(sport)
		if err != nil {
			log.Println("GetLiveEvents: Failed to fetch betfair events for sport - ", sport)
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
// func PlaceBet(reqDto requestdto.PlaceBetReqDto, betDto sports.BetDto, operatorDto operatordto.OperatorDTO, opConfig providers.Config) (sports.BetDto, error) {
// 	// I. Get Operator Hold - MarketId, EventId, CompetitionId, SportId, ProviderId, OperatorId
// 	// II. Get Operator Balance - OperatorId
// 	// III. Get Platform Hold - MarketId, EventId, CompetitionId, SportId, ProviderId
// 	// IV. Get Min Bet Value - ProviderId

// 	// implementaion
// 	// 1. Apply Operator Hold (#III) on Stake & Debit Amounts
// 	betDto.BetReq.OperatorHold = 0 //opConfig.Hold
// 	betDto.BetReq.OperatorAmount = utils.Truncate4Decfloat64(betDto.BetReq.DebitAmount * (100 - betDto.BetReq.OperatorHold) / 100)
// 	log.Println("PlaceBet: betDto.BetReq.OperatorAmount - ", betDto.BetReq.OperatorAmount)
// 	// 2. Check Operator Balance is greater than the betvalue (#IV) - Error
// 	// operator := providers.GetOperator(operatorDto.OperatorId)
// 	// if betDto.BetReq.OperatorAmount > operator.Balance {
// 	// 	log.Println("PlaceBet: Operator Balance is low - ", operator.Balance)
// 	// 	return betDto, fmt.Errorf("Operator - Insufficient Funds!")
// 	// }
// 	// 3. Apply Platform Hold (#V) on Stake & Debit Amounts
// 	sportKey := reqDto.ProviderId + "-" + reqDto.SportId
// 	//compKey := sportKey + "-" + reqDto.CompetetionId
// 	//eventKey := sportKey + "-" + reqDto.EventId
// 	provider, err := cache.GetProvider(reqDto.ProviderId)
// 	if err != nil {
// 		log.Println("PlaceBet: cache.GetProvider failed with error - ", err.Error())
// 	}
// 	sport, err := cache.GetSport(sportKey)
// 	competition, err := cache.GetCompetition(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
// 	event, err := cache.GetEvent(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
// 	market, err := cache.GetMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
// 	sapConfig, level := providers.GetSapConfig(reqDto.MarketType, market, event, competition, sport, operatorDto, provider)
// 	configJson, err := json.Marshal(sapConfig)
// 	log.Println("PlaceBet: sapConfig JSON is - ", level, string(configJson))
// 	betDto.BetReq.PlatformHold = 0 //sapConfig.Hold
// 	betDto.BetReq.PlatformAmount = utils.Truncate4Decfloat64(betDto.BetReq.OperatorAmount * (100 - betDto.BetReq.PlatformHold) / 100)
// 	log.Println("PlaceBet: betDto.BetReq.PlatformAmount - ", betDto.BetReq.PlatformAmount)
// 	reqDto.StakeAmount = utils.Truncate4Decfloat64(betDto.BetDetails.StakeAmount * (100 - betDto.BetReq.OperatorHold) / 100)
// 	reqDto.StakeAmount = utils.Truncate4Decfloat64(reqDto.StakeAmount * (100 - betDto.BetReq.PlatformHold) / 100)
// 	reqDto.StakeAmount = utils.Truncate4Decfloat64(reqDto.StakeAmount / float64(SportRadarRate))
// 	log.Println("PlaceBet: reqDto.StakeAmount - ", reqDto.StakeAmount)
// 	// 6. Check betvalue is greater than minbet value (#VI)
// 	log.Println("PlaceBet: betDto.BetReq.PlatformAmount is - ", betDto.BetReq.PlatformAmount)
// 	log.Println("PlaceBet: MinBetValue is - ", MinBetValue)
// 	//if betDto.BetReq.PlatformAmount < float64(MinBetValue) {
// 	// TODO: Accept the bet and dont send to BetFair
// 	//time.Sleep(time.Duration(500 * int(time.Millisecond))) // TODO: Remove this delay. Added for sportradar testing
// 	respObj, err := ValidateOdds(reqDto)
// 	if err != nil {
// 		// 1.1. Failed to validate odds
// 		log.Println("PlaceBet: Odds Validation call failed with - ", err.Error())
// 		return betDto, fmt.Errorf("Odds Changed!")
// 	}
// 	respJson, _ := json.Marshal(respObj)
// 	//log.Println("PlaceBet: Event Open Date is - ", time.Unix(respObj.OpenDate/1000, 0))
// 	if !respObj.IsValid {
// 		// 1.2. Invalid Odds
// 		log.Println("PlaceBet: Invalid Odds - ", reqDto.OddValue)
// 		log.Println("PlaceBet: ValidateOdds response json is - ", string(respJson))
// 		return betDto, fmt.Errorf("Odds Changed!")
// 	}
// 	/* */
// 	// 2. Check for acceptance criteria
// 	if respObj.Status != "IN_PLAY" {
// 		// 2.1. Match Odds bet on upcoming match
// 		// TODO: add to collection
// 		log.Println("PlaceBet: MATCH_ODDS bet only allowed on IN_PLAY events - ", respObj.Status)
// 		log.Println("PlaceBet: ValidateOdds response json is - ", string(respJson))
// 		//return betDto, fmt.Errorf("Betting NOT ALLOWED!")
// 	}
// 	betDto.Status = constants.SAP.BetStatus.OPEN()
// 	/*
// 		} else {
// 			// 3. BetFair - Place Order
// 			log.Println("PlaceBet: PlaceOrder Before")
// 			respDto, err := PlaceOrder(reqDto, betDto)
// 			if err != nil {
// 				log.Println("PlaceBet: PlaceOrder failed with error - ", err.Error())
// 				return betDto, err
// 			}
// 			log.Println("PlaceBet: PlaceOrder Status is - ", respDto.Status)
// 			if respDto.Status != "RS_OK" {
// 				// API Call failed. Mark bet status as failed. and return
// 				log.Println("PlaceBet: PlaceOrder API Status is - ", respDto.Status)
// 				log.Println("PlaceBet: PlaceOrder API ErrorCode is - ", respDto.ErrorDescription)
// 				log.Println("PlaceBet: PlaceOrder API AltStake is - ", respDto.AltStake)
// 				return betDto, fmt.Errorf("SportRadar returned failure!")
// 			}
// 		}
// 	*/
// 	// 7. Bet Success, Add to operator ledger
// 	// err = providers.OperatorLedgerTx(operator, constants.SAP.LedgerTxType.BETPLACEMENT(), betDto.BetReq.OperatorAmount*-1, betDto.BetId)
// 	// if err != nil {
// 	// 	log.Println("PlaceBet: OperatorLedgerTx failed with error - ", err.Error())
// 	// }
// 	return betDto, nil
// }

func SettledOrders(reqDto sportradar.MarketResultReq) {
	if constants.SportRadar.ResultType.RESULT_SETTLEMENT() == reqDto.MarketStatus {
		result_settlement(reqDto)
	} else if constants.SportRadar.ResultType.VOID_SETTLEMENT() == reqDto.MarketStatus {
		void_settlement(reqDto)
	} else {
		// Inavlid Result Type
		log.Println("SettledOrders: Invalid ResultType - ", reqDto.MarketStatus)
		return
	}
	return
}

func result_settlement(reqDto sportradar.MarketResultReq) {
	// 1. Get Market from DB
	eventKey := constants.SAP.ProviderType.SportRadar() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	//log.Println("SettledOrders: result_settlement START for MarketKey is - ", marketKey)
	// 1. Get Market from DB
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		//log.Println("SettledOrders: database.GetMarket failed with error - ", err.Error())
		isMarketFound = false
		//return
	}
	// 2. Add a result to the Market
	result := models.Result{}
	isWinnerFound := false
	for _, runnerResult := range reqDto.Result.Runners {
		if runnerResult.Result == "Won" {
			isWinnerFound = true
			result.RunnerId = runnerResult.RunnerId
			result.RunnerName = runnerResult.RunnerName
			result.SessionOutcome = 0
			result.ResultTime = time.Now().Unix()
			break
		}
	}
	if isWinnerFound == false {
		// NO Winner
		log.Println("SettledOrders: None of the resultrunners are WON - ")
		return
	}
	if isMarketFound == true {
		market.Results = append(market.Results, result)
		market.MarketStatus = constants.SportRadar.MarketStatus.SETTLED()
		// 3. Update Market
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("SettledOrders: database.ReplaceMarket failed with error - ", err.Error())
			//return
		}
	}
	// 4. Get Open Bets
	openBets, _, err := database.GetOpenBetsByMarket(eventKey, reqDto.MarketId)
	if err != nil {
		log.Println("SettledOrders: database.GetOpenBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(openBets) == 0 {
		//log.Println("SettledOrders: ZERO Open bets to settle for market Id - ", reqDto.MarketId)
		return
	}
	log.Println("SettledOrders: openBets count is - ", len(openBets))
	reqBodyStr, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("SettledOrders: json.Marshal failed with error - ", err.Error())
	} else {
		log.Println("SettledOrders: Request body string is - ", string(reqBodyStr))
	}
	// 5. Add Result to the Bet
	settledBets := []sports.BetDto{}
	for _, openBet := range openBets {
		// 5.1. update bet with result
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
		// 5.2. append bet to list
		settledBets = append(settledBets, openBet)
	}
	log.Println("SettledOrders: settledBets count is - ", len(settledBets))
	if len(settledBets) == 0 {
		log.Println("SettledOrders: ZERO SettledBets to settle for market Id - ", reqDto.MarketId)
		return
	}
	// 6. Process Settlement
	operator.CommonResultRoutine(settledBets)
	log.Println("SettledOrders: result_settlement END for MarketKey is - ", marketKey)
	return
}

func void_settlement(reqDto sportradar.MarketResultReq) {
	// 1. Get Market from DB
	eventKey := constants.SAP.ProviderType.SportRadar() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	log.Println("SettledOrders: void_settlement START for MarketKey is - ", marketKey)
	// 1. Get Market from DB
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("SettledOrders: database.GetMarket failed with error - ", err.Error())
		isMarketFound = false
		//return
	}
	// 2. Add a rollback to the Market
	rollbackReq := models.Rollback{}
	rollbackReq.RollbackType = constants.SportRadar.MarketStatus.VOIDED()
	rollbackReq.RollbackReason = constants.SportRadar.MarketStatus.VOIDED()
	rollbackReq.RollbackTime = time.Now().UnixNano() / int64(time.Millisecond)
	if isMarketFound == true {
		market.Rollbacks = append(market.Rollbacks, rollbackReq)
		market.MarketStatus = constants.SportRadar.MarketStatus.VOIDED()
		// 3. Update Market
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("SettledOrders: database.ReplaceMarket failed with error - ", err.Error())
			//return
		}
	}
	// 4. Get Open Bets
	openBets, _, err := database.GetOpenBetsByMarket(eventKey, reqDto.MarketId)
	if err != nil {
		log.Println("SettledOrders: database.GetOpenBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(openBets) == 0 {
		//log.Println("SettledOrders: ZERO Open bets to settle for market Id - ", reqDto.MarketId)
		return
	}
	log.Println("SettledOrders: openBets count is - ", len(openBets))
	reqBodyStr, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("SettledOrders: json.Marshal failed with error - ", err.Error())
	} else {
		log.Println("SettledOrders: Request body string is - ", string(reqBodyStr))
	}
	// 5. Add Result to the Bet
	settledBets := []sports.BetDto{}
	for _, openBet := range openBets {
		// 5.1. update bet with rollback
		rollbackReq := ComputeRollback(openBet, constants.SportRadar.MarketStatus.VOIDED(), reqDto)
		if err != nil {
			log.Println("SettledOrders: ComputeRollback failed with error - ", err.Error())
			continue
		}
		openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
		openBet.NetAmount += rollbackReq.RollbackAmount
		openBet.Status = constants.SAP.BetStatus.SETTLED_VOIDED()
		// setting updatedAt
		openBet.UpdatedAt = rollbackReq.ReqTime
		// 5.2. append bet to list
		settledBets = append(settledBets, openBet)
	}
	log.Println("SettledOrders: settledBets count is - ", len(settledBets))
	if len(settledBets) == 0 {
		log.Println("SettledOrders: ZERO SettledBets to settle for market Id - ", reqDto.MarketId)
		return
	}
	// 6. Process Settlement
	operator.CommonRollbackRoutine(constants.SportRadar.MarketStatus.VOIDED(), settledBets)
	log.Println("SettledOrders: void_settlement END for MarketKey is - ", marketKey)
	return
}

func RollbackOrders(reqDto sportradar.MarketRollbackReq) {
	if constants.SportRadar.RollbackType.Rollback() == reqDto.RollbackType {
		rollback_rollback(reqDto)
	} else if constants.SportRadar.RollbackType.TimelyVoid() == reqDto.RollbackType {
		timelyvoid_rollback(reqDto)
	} else if constants.SportRadar.RollbackType.TimelyVoidRollback() == reqDto.RollbackType {
		timelyvoidrollback_rollback(reqDto)
	} else {
		// Inavlid Result Type
		log.Println("RollbackOrders: Invalid RollbackType - ", reqDto.RollbackType)
		return
	}
	return
}

func rollback_rollback(reqDto sportradar.MarketRollbackReq) {
	// 1. Get Market from DB
	// 2. Add Rollback to Market
	// 3. Update Market
	// 4. Get Open Bets
	// 5. Add RollbackRequest to the Bets
	// 6. Process Rollback
	// 1. Get Market from DB

	// 1. Get Market from DB
	eventKey := constants.SAP.ProviderType.SportRadar() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	log.Println("RollbackOrders: rollback_rollback START for MarketKey is - ", marketKey)
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("RollbackOrders: database.GetMarket failed with error - ", err.Error())
		isMarketFound = false
		//return
	}
	// 2. Add a Rollback to Market
	rollbackReq := models.Rollback{}
	rollbackReq.RollbackType = reqDto.RollbackType
	rollbackReq.RollbackReason = reqDto.Reason
	rollbackReq.RollbackTime = time.Now().UnixNano() / int64(time.Millisecond)
	if isMarketFound == true {
		market.Rollbacks = append(market.Rollbacks, rollbackReq)
		market.MarketStatus = constants.SportRadar.MarketStatus.OPEN()
		// 3. Update Market Documetn in db
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("RollbackOrders: database.ReplaceMarket failed with error - ", err.Error())
			return
		}
	}
	// 4. Get bets by Market & Status
	// if Rollback, then bring all settled bets
	// if TimelyVoid, then bring open bets between start and end time
	// if TimelyVoidRollback, then bring void bets between start and end time
	bets, err := database.GetBetsByMarket(eventKey, reqDto.MarketId, "rollback", 0, 0)
	if err != nil {
		log.Println("RollbackOrders: database.GetBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(bets) == 0 {
		log.Println("RollbackOrders: ZERO Open bets to rollback for marketId - ", reqDto.MarketId)
		return
	}
	log.Println("RollbackOrders: openBets count is - ", len(bets))
	// 5. Add RollbackReq to bets
	RollbackBets := []sports.BetDto{}
	for _, bet := range bets {
		// 2.2.1. update bet with result
		rollbackReq := providers.ComputeRollback(bet, reqDto.RollbackType)
		if err != nil {
			log.Println("RollbackOrders: ComputeResult failed with error - ", err.Error())
			continue
		}
		bet.RollbackReqs = append(bet.RollbackReqs, rollbackReq)
		bet.NetAmount += rollbackReq.RollbackAmount
		bet.Status = constants.SAP.BetStatus.OPEN()
		// setting updatedAt
		bet.UpdatedAt = rollbackReq.ReqTime
		// 2.2.3. append bet to list
		RollbackBets = append(RollbackBets, bet)
	}
	log.Println("RollbackOrders: RollbackBets count is - ", len(RollbackBets))
	operator.CommonRollbackRoutine(constants.SAP.BetStatus.ROLLBACK(), RollbackBets)
	log.Println("RollbackOrders: rollback_rollback END for MarketKey is - ", marketKey)
	return
}

func timelyvoid_rollback(reqDto sportradar.MarketRollbackReq) {
	// 1. Get Market from DB
	// 2. Add Rollback to Market
	// 3. Update Market
	// 4. Get Open Bets
	// 5. Add RollbackRequest to the Bets
	// 6. Process Rollback

	// 1. Get Market from DB
	eventKey := constants.SAP.ProviderType.SportRadar() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	log.Println("RollbackOrders: timelyvoid_rollback START for MarketKey is - ", marketKey)
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("RollbackOrders: database.GetMarket failed with error - ", err.Error())
		isMarketFound = false
		//return
	}
	// 2. Add a Rollback to Market
	rollbackReq := models.Rollback{}
	rollbackReq.RollbackType = reqDto.RollbackType
	rollbackReq.RollbackReason = reqDto.Reason // TODO: Add start and end times to reason
	rollbackReq.RollbackTime = time.Now().UnixNano() / int64(time.Millisecond)
	if isMarketFound == true {
		market.Rollbacks = append(market.Rollbacks, rollbackReq)
		//market.MarketStatus = constants.SportRadar.MarketStatus.OPEN() // NA
		// 3. Update Market Documetn in db
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("RollbackOrders: database.ReplaceMarket failed with error - ", err.Error())
			return
		}
	}
	// 4. Get bets by Market & Status
	// if Rollback, then bring all settled bets
	// if TimelyVoid, then bring open bets between start and end time
	// if TimelyVoidRollback, then bring void bets between start and end time
	bets, err := database.GetBetsByMarket(eventKey, reqDto.MarketId, "void", reqDto.StartTime, reqDto.EndTime)
	if err != nil {
		log.Println("RollbackOrders: database.GetBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(bets) == 0 {
		log.Println("RollbackOrders: ZERO Open bets to rollback for marketId - ", reqDto.MarketId)
		return
	}
	log.Println("RollbackOrders: openBets count is - ", len(bets))
	// 5. Add RollbackReq to bets
	RollbackBets := []sports.BetDto{}
	for _, bet := range bets {
		// 2.2.1. update bet with result
		rollbackReq := providers.ComputeRollback(bet, reqDto.RollbackType)
		if err != nil {
			log.Println("RollbackOrders: ComputeResult failed with error - ", err.Error())
			continue
		}
		bet.RollbackReqs = append(bet.RollbackReqs, rollbackReq)
		bet.NetAmount += rollbackReq.RollbackAmount
		bet.Status = constants.SAP.BetStatus.TIMELY_VOIDED()
		// setting updatedAt
		bet.UpdatedAt = rollbackReq.ReqTime
		// 2.2.3. append bet to list
		RollbackBets = append(RollbackBets, bet)
	}
	log.Println("RollbackOrders: RollbackBets count is - ", len(RollbackBets))
	operator.CommonRollbackRoutine(constants.SAP.BetStatus.VOIDED(), RollbackBets)
	log.Println("RollbackOrders: timelyvoid_rollback END for MarketKey is - ", marketKey)
	return
}

func timelyvoidrollback_rollback(reqDto sportradar.MarketRollbackReq) {
	// 1. Get Market from DB
	// 2. Add Rollback to Market
	// 3. Update Market
	// 4. Get Open Bets
	// 5. Add RollbackRequest to the Bets
	// 6. Process Rollback

	// 1. Get Market from DB
	eventKey := constants.SAP.ProviderType.SportRadar() + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	log.Println("RollbackOrders: timelyvoidrollback_rollback START for MarketKey is - ", marketKey)
	isMarketFound := true
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("RollbackOrders: database.GetMarket failed with error - ", err.Error())
		isMarketFound = false
		//return
	}
	// 2. Add a Rollback to Market
	rollbackReq := models.Rollback{}
	rollbackReq.RollbackType = reqDto.RollbackType
	rollbackReq.RollbackReason = reqDto.Reason // TODO: Add start and end times to reason
	rollbackReq.RollbackTime = time.Now().UnixNano() / int64(time.Millisecond)
	if isMarketFound == true {
		market.Rollbacks = append(market.Rollbacks, rollbackReq)
		//market.MarketStatus = constants.SportRadar.MarketStatus.OPEN()
		// 3. Update Market Documetn in db
		err = database.ReplaceMarket(market)
		if err != nil {
			log.Println("RollbackOrders: database.ReplaceMarket failed with error - ", err.Error())
			return
		}
	}
	// 4. Get bets by Market & Status
	// if Rollback, then bring all settled bets
	// if TimelyVoid, then bring open bets between start and end time
	// if TimelyVoidRollback, then bring void bets between start and end time
	bets, err := database.GetBetsByMarket(eventKey, reqDto.MarketId, "voidrollback", reqDto.StartTime, reqDto.EndTime)
	if err != nil {
		log.Println("RollbackOrders: database.GetBetsByMarket Failed with error - ", err.Error())
		return
	}
	if len(bets) == 0 {
		log.Println("RollbackOrders: ZERO Open bets to rollback for marketId - ", reqDto.MarketId)
		return
	}
	log.Println("RollbackOrders: openBets count is - ", len(bets))
	// 5. Add RollbackReq to bets
	RollbackBets := []sports.BetDto{}
	for _, bet := range bets {
		// 2.2.1. update bet with result
		rollbackReq := providers.ComputeRollback(bet, reqDto.RollbackType)
		if err != nil {
			log.Println("RollbackOrders: ComputeResult failed with error - ", err.Error())
			continue
		}
		bet.RollbackReqs = append(bet.RollbackReqs, rollbackReq)
		bet.NetAmount += rollbackReq.RollbackAmount
		bet.Status = constants.SAP.BetStatus.OPEN()
		// setting updatedAt
		bet.UpdatedAt = rollbackReq.ReqTime
		// 2.2.3. append bet to list
		RollbackBets = append(RollbackBets, bet)
	}
	log.Println("RollbackOrders: RollbackBets count is - ", len(RollbackBets))
	operator.CommonRollbackRoutine(constants.SAP.BetStatus.ROLLBACK(), RollbackBets) // Need to check for void rollback
	log.Println("RollbackOrders: timelyvoidrollback_rollback END for MarketKey is - ", marketKey)
	return
}

func GetSportRadarMarketResults() error {
	log.Println("GetSportRadarMarketResults: Start Time is: ", time.Now())
	// 1. Get SportRadar Open Bets List
	// 2. Create Unique MarketIds List
	// 3. Call SportRadar Market Results endpoint
	// 4. Create Markets map with MarketIds as Key
	// 5. Get Markets from Database
	// 6. Update Market Results in Market Collection and save
	// 7. Create Settled bets
	// 8. Process Settled Bets

	// 1. Get SportRadar Open Bets List
	bets, err := database.GetBetsByStatus(constants.SAP.ProviderType.SportRadar(), constants.SAP.BetStatus.OPEN())
	if err != nil {
		log.Println("GetSportRadarMarketResults: database.GetBetsByStatus for SportRadar OPEN bets failed with error - ", err.Error())
		return err
	}
	log.Println("GetSportRadarMarketResults: SportRadar Open bets Count is - ", len(bets))
	if len(bets) == 0 {
		return nil
	}
	// 2. Create Unique MarketIds List
	marketIdMap := make(map[string]bool)
	marketIds := []string{}
	for _, bet := range bets {
		_, ok := marketIdMap[bet.MarketId]
		if ok {
			continue
		}
		marketIdMap[bet.MarketId] = true
		marketIds = append(marketIds, bet.MarketId)
	}
	log.Println("GetSportRadarMarketResults: Markets Count is - ", len(marketIds))
	if len(marketIds) == 0 {
		return nil
	}
	// 3. Call SportRadar Market Results endpoint
	resp, err := GetMarketResults(marketIds)
	if err != nil {
		log.Println("GetSportRadarMarketResults: sportradar.GetMarketResults failed with error - ", err.Error())
		return err
	}
	if resp.Status != "RS_OK" {
		log.Println("GetSportRadarMarketResults: sportradar.GetMarketResults returned status is - ", resp.Status)
		log.Println("GetSportRadarMarketResults: sportradar.GetMarketResults ErrorDescription is - ", resp.ErrorDescription)
		return fmt.Errorf(resp.ErrorDescription)
	}
	log.Println("GetSportRadarMarketResults: Markets Count is - ", len(resp.MarketResults))
	if len(resp.MarketResults) == 0 {
		return nil
	}
	// 4. Create Markets map with MarketIds as Key
	resultMarketIds := []string{}
	marketsResultMap := make(map[string]models.Result)
	//marketsRollbackMap := make(map[string]models.Rollback)
	for _, marketResult := range resp.MarketResults {
		// Below are possibile outcomes and this function only take care of 1 & 2
		// 1. RESULT_SETTLEMENT       ("RESULT_SETTLEMENT")
		// 2. VOID_SETTLEMENT         ("VOID_SETTLEMENT")
		// 3. RESULT_ROLLBACK         ("RESULT_ROLLBACK")
		// 4. TIMELY_VOID             ("TIMELY_VOID")
		// 5. TIMELY_VOID_ROLLBACK    ("TIMELY_VOID_ROLLBACK")
		switch strings.ToUpper(marketResult.MarketStatus) {
		case "RESULT_SETTLEMENT":
			if len(marketResult.Results) == 0 {
				// No results to settle bets, LOG ERROR
				continue
			}
			log.Println("GetSportRadarMarketResults: MarketId Count is - ", marketResult.MarketId)
			log.Println("GetSportRadarMarketResults: Results Count is - ", len(marketResult.Results))
			for _, runnerResult := range marketResult.Results[len(marketResult.Results)-1].RunnerResults {
				if runnerResult.Result == "Won" {
					result := models.Result{}
					result.RunnerId = runnerResult.RunnerId
					result.RunnerName = runnerResult.RunnerName
					result.SessionOutcome = 0
					result.ResultTime = time.Now().Unix()
					marketsResultMap[marketResult.MarketId] = result
					resultMarketIds = append(resultMarketIds, marketResult.MarketId)
					break
				}
			}
		case "VOID_SETTLEMENT":
			// TODO: Handle Void Markets
			// 2. Add a Rollback to Market
			//rollbackReq := models.Rollback{}
			//rollbackReq.RollbackType = constants.SAP.BetStatus.VOIDED()
			//rollbackReq.RollbackReason = reqDto.RollbackType
			//rollbackReq.RollbackTime = time.Now().UnixNano() / int64(time.Millisecond)
			//marketResult.Rollbacks = append(marketResult.Rollbacks, rollbackReq)
			//marketResult.MarketStatus = constants.SAP.BetStatus.VOIDED()
		case "RESULT_ROLLBACK", "TIMELY_VOID", "TIMELY_VOID_ROLLBACK":
			// No action required
			continue
		default:
			// TODO: LOG ERROR
			continue
		}
		// if len(marketResult.Rollbacks) > 0 {
		// 	// TODO: Handle timely Rollbacks
		// 	log.Println("GetSportRadarMarketResults: MarketId Count is - ", marketResult.MarketId)
		// 	log.Println("GetSportRadarMarketResults: Rollbacks Count is - ", len(marketResult.Rollbacks))
		// }
		// if len(marketResult.Results) == 0 {
		// 	// No results to settle bets
		// 	continue
		// }
		// log.Println("GetSportRadarMarketResults: MarketId Count is - ", marketResult.MarketId)
		// log.Println("GetSportRadarMarketResults: Results Count is - ", len(marketResult.Results))
		// for _, runnerResult := range marketResult.Results[len(marketResult.Results)-1].RunnerResults {
		// 	if runnerResult.Result == "Won" {
		// 		result := models.Result{}
		// 		result.RunnerId = runnerResult.RunnerId
		// 		result.RunnerName = runnerResult.RunnerName
		// 		result.SessionOutcome = 0
		// 		result.ResultTime = time.Now().Unix()
		// 		marketsMap[marketResult.MarketId] = result
		// 		resultMarketIds = append(resultMarketIds, marketResult.MarketId)
		// 		break
		// 	}
		// }
	}
	log.Println("GetSportRadarMarketResults: marketsResultMap Count is - ", len(marketsResultMap))
	if len(marketsResultMap) == 0 {
		return nil
	}
	// 5. Get Markets from Database
	markets, err := database.GetMarketsByMarketIds(resultMarketIds)
	if err != nil {
		log.Println("GetSportRadarMarketResults: database.GetMarketsByMarketIds failed with error - ", err.Error())
		return err
	}
	if len(markets) == 0 {
		// TODO: Handle missing markets in database
		return nil
	}
	// 6. Update Market Results in Market Collection and save
	for i, market := range markets {
		markets[i].Results = append(markets[i].Results, marketsResultMap[market.MarketId])
	}
	count, msgs := database.UpdateManyMarkets(markets)
	if count != len(markets) {
		log.Println("GetSportRadarMarketResults: database.UpdateManyMarkets failed markets count is  - ", len(msgs))
		for _, msg := range msgs {
			log.Println("GetSportRadarMarketResults: Failure message is - ", msg)
		}
	}
	// 7. Create Settled bets
	settledBets := []sports.BetDto{}
	for _, bet := range bets {
		// 7.1. Check marketId is present in marketsMap
		result, ok := marketsResultMap[bet.MarketId]
		if ok == false {
			continue
		}
		// 7.2. Create result request
		resultReq := sports.ResultReqDto{}
		resultReq.ReqId = uuid.New().String()
		resultReq.ReqTime = time.Now().UnixMilli()
		resultReq.CreditAmount = 0
		resultReq.OperatorAmount = 0
		resultReq.PlatformAmount = 0
		resultReq.RunnerName = result.RunnerName
		resultReq.SessionOutcome = 0
		if bet.BetDetails.RunnerId == result.RunnerId { // All sportradar bets are BACK bets
			resultReq.CreditAmount = utils.Truncate4Decfloat64(bet.BetDetails.StakeAmount * bet.BetDetails.OddValue)
			resultReq.OperatorAmount = utils.Truncate4Decfloat64(bet.BetReq.OperatorAmount * bet.BetDetails.OddValue)
			resultReq.PlatformAmount = utils.Truncate4Decfloat64(bet.BetReq.PlatformAmount * bet.BetDetails.OddValue)
		}
		bet.ResultReqs = append(bet.ResultReqs, resultReq)
		bet.NetAmount += resultReq.CreditAmount
		bet.Status = constants.BetFair.BetStatus.SETTLED()
		// setting updatedAt
		bet.UpdatedAt = resultReq.ReqTime
		settledBets = append(settledBets, bet)
	}
	log.Println("GetSportRadarMarketResults: settledBets Count is - ", len(settledBets))
	if len(settledBets) == 0 {
		return nil
	}
	// 8. Process Settled Bets
	log.Println("GetSportRadarMarketResults: settledBets count is - ", len(settledBets))
	operator.CommonResultRoutine(settledBets)
	log.Println("GetSportRadarMarketResults: End Time is: ", time.Now())
	return nil
}

func ComputeRollback(openBet sports.BetDto, reason string, resultReq sportradar.MarketResultReq) sports.RollbackReqDto {
	rollbackReq := sports.RollbackReqDto{}
	rollbackReq.ReqId = uuid.New().String()
	rollbackReq.ReqTime = time.Now().UnixMilli()
	rollbackReq.RollbackReason = reason
	rollbackReq.RollbackAmount = 0 // positive means, deposit to user, negative means, deduct from user
	for _, result := range openBet.ResultReqs {
		rollbackReq.RollbackAmount -= result.CreditAmount
		rollbackReq.OperatorAmount -= result.OperatorAmount
		rollbackReq.PlatformAmount -= result.PlatformAmount
	}
	for _, rollback := range openBet.RollbackReqs {
		rollbackReq.RollbackAmount -= rollback.RollbackAmount
		rollbackReq.OperatorAmount -= rollback.OperatorAmount
		rollbackReq.PlatformAmount -= rollback.PlatformAmount
	}
	switch strings.ToLower(reason) {
	case "void", "voided", "cancelled", "deleted", "lapsed", "expired":
		// TODO: Handle VoidFactor by each runner
		isFound := false
		for _, runner := range resultReq.Result.Runners {
			if openBet.BetDetails.RunnerId == runner.RunnerId {
				log.Println("RollbackOrders: ComputeRollback runner.VoidFactor is - ", runner.VoidFactor)
				rollbackReq.RollbackAmount += (openBet.BetReq.DebitAmount * runner.VoidFactor)
				rollbackReq.OperatorAmount += (openBet.BetReq.OperatorAmount * runner.VoidFactor)
				rollbackReq.PlatformAmount += (openBet.BetReq.PlatformAmount * runner.VoidFactor)
				if runner.Result == "Won" {
					rollbackReq.RollbackAmount += (openBet.BetReq.DebitAmount * runner.VoidFactor * openBet.BetDetails.OddValue)
					rollbackReq.OperatorAmount += (openBet.BetReq.OperatorAmount * runner.VoidFactor * openBet.BetDetails.OddValue)
					rollbackReq.PlatformAmount += (openBet.BetReq.PlatformAmount * runner.VoidFactor * openBet.BetDetails.OddValue)
				}
				isFound = true
				break
			}
		}
		if isFound == false {
			log.Println("RollbackOrders: ComputeRollback unexpexted state - Runner not matched with any of the result runners - ", openBet.BetDetails.RunnerId)
			rollbackReq.RollbackAmount += openBet.BetReq.DebitAmount
			rollbackReq.OperatorAmount += openBet.BetReq.OperatorAmount
			rollbackReq.PlatformAmount += openBet.BetReq.PlatformAmount
		}
	case "rollback", "voidrollback":
		// nothing to do here
	default:
		log.Println("ComputeRollback: Unexpected ResultType - ", reason)
	}
	rollbackReq.RollbackAmount = utils.Truncate4Decfloat64(rollbackReq.RollbackAmount)
	rollbackReq.OperatorAmount = utils.Truncate4Decfloat64(rollbackReq.OperatorAmount)
	rollbackReq.PlatformAmount = utils.Truncate4Decfloat64(rollbackReq.PlatformAmount)
	return rollbackReq
}

func ComputeResult(openBet sports.BetDto, result sportradar.MarketResultReq) (sports.ResultReqDto, error) {
	resultReq := sports.ResultReqDto{}
	resultReq.ReqId = uuid.New().String()
	resultReq.ReqTime = time.Now().UnixMilli()
	resultReq.SessionOutcome = 0
	resultReq.CreditAmount = 0

	isFound := false
	for _, runner := range result.Result.Runners {
		if openBet.BetDetails.RunnerId == runner.RunnerId {
			log.Println("SettledOrders: ComputeResult runner.DeadHeatFactor is - ", runner.DeadHeatFactor)
			log.Println("SettledOrders: ComputeResult runner.VoidFactor is - ", runner.VoidFactor)
			resultReq.RunnerName = runner.RunnerName
			// x * vf + x * (1-vf) * dhf * oddvalue
			resultReq.CreditAmount = openBet.BetReq.DebitAmount * runner.VoidFactor
			resultReq.OperatorAmount = openBet.BetReq.OperatorAmount * runner.VoidFactor
			resultReq.PlatformAmount = openBet.BetReq.PlatformAmount * runner.VoidFactor
			if runner.Result == "Won" {
				resultReq.CreditAmount += openBet.BetReq.DebitAmount * (1 - runner.VoidFactor) * runner.DeadHeatFactor * openBet.BetDetails.OddValue
				resultReq.OperatorAmount += openBet.BetReq.OperatorAmount * (1 - runner.VoidFactor) * runner.DeadHeatFactor * openBet.BetDetails.OddValue
				resultReq.PlatformAmount += openBet.BetReq.PlatformAmount * (1 - runner.VoidFactor) * runner.DeadHeatFactor * openBet.BetDetails.OddValue
			}
			isFound = true
			break
		}
	}
	if isFound == false { // default is user LOST
		log.Println("SettledOrders: ComputeResult unexpexted state - Runner not matched with any of the result runners - ", openBet.BetDetails.RunnerId)
		resultReq.CreditAmount = 0
		resultReq.OperatorAmount = 0
		resultReq.PlatformAmount = 0
	}
	resultReq.CreditAmount = utils.Truncate4Decfloat64(resultReq.CreditAmount)
	return resultReq, nil
}
