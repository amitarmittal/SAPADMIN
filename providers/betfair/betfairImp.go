package betfair

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	"Sp/dto/providers/betfair"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	sessdto "Sp/dto/session"
	"Sp/dto/sports"
	"Sp/operator"
	"Sp/providers"
	utils "Sp/utilities"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	// time in milli seconds
	CricketMatchOdds int = 3500
	CricketBookmaker int = 1500
	CricetFancy      int = 1500
	SoccerMatchOdds  int = 4500
	TennisMatchOdds  int = 4500
	//MinBetValue      int = 2 // Minimum bet value to call betfair API, if less than this, we will make it success
	BetFairRate int = 10 // HKD
)

/*
func GetLiveEvents(sportsList []string) ([]dto.EventDto, error) {
	//log.Println("GetLiveEvents: BetFair Start")
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

/*
func PlaceBet(reqDto requestdto.PlaceBetReqDto, betDto sports.BetDto, operatorDto operatordto.OperatorDTO, opConfig providers.Config) (sports.BetDto, error) {
	// I. Get Operator Hold - MarketId, EventId, CompetitionId, SportId, ProviderId, OperatorId
	// II. Get Operator Balance - OperatorId
	// III. Get Platform Hold - MarketId, EventId, CompetitionId, SportId, ProviderId
	// IV. Get Min Bet Value - ProviderId

	// implementaion
	// 1. Apply Operator Hold (#III) on Stake & Debit Amounts
	log.Println("PlaceBet: User Bet Value - ", reqDto.StakeAmount)
	log.Println("PlaceBet: User Risk - ", betDto.BetReq.DebitAmount)
	betDto.BetReq.OperatorHold = opConfig.Hold
	betDto.BetReq.OperatorAmount = utils.Truncate64(betDto.BetReq.DebitAmount * (100 - opConfig.Hold) / 100)
	log.Println("PlaceBet: Operator Risk - ", betDto.BetReq.OperatorAmount)
	//reqDto.StakeAmount = utils.Truncate64(reqDto.StakeAmount * (100 - opConfig.Hold) / 100)
	// 2. Check Operator Balance is greater than the betvalue (#IV) - Error
	operator := providers.GetOperator(operatorDto.OperatorId)
	if betDto.BetReq.OperatorAmount > operator.Balance {
		log.Println("PlaceBet: Operator Balance is low - ", operator.Balance)
		return betDto, fmt.Errorf("Operator - Insufficient Funds!")
	}
	// 3. Apply Platform Hold (#V) on Stake & Debit Amounts
	sportKey := reqDto.ProviderId + "-" + reqDto.SportId
	//compKey := sportKey + "-" + reqDto.CompetetionId
	//eventKey := sportKey + "-" + reqDto.EventId
	provider, err := cache.GetProvider(reqDto.ProviderId)
	sport, err := cache.GetSport(sportKey)
	competition, err := cache.GetCompetition(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
	event, err := cache.GetEvent(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
	sapConfig := providers.GetSapConfig(reqDto.MarketType, event, competition, sport, operatorDto, provider)
	configJson, err := json.Marshal(sapConfig)
	log.Println("PlaceBet: sapConfig JSON is - ", string(configJson))
	betDto.BetReq.PlatformHold = sapConfig.Hold
	betDto.BetReq.PlatformAmount = utils.Truncate64(betDto.BetReq.OperatorAmount * (100 - sapConfig.Hold) / 100)
	log.Println("PlaceBet: Platform Risk - ", betDto.BetReq.PlatformAmount)
	//reqDto.StakeAmount = utils.Truncate64(reqDto.StakeAmount * (100 - sapConfig.Hold) / 100)
	reqDto.StakeAmount = utils.Truncate64(betDto.BetDetails.StakeAmount * (100 - opConfig.Hold) / 100)
	reqDto.StakeAmount = utils.Truncate64(reqDto.StakeAmount * (100 - sapConfig.Hold) / 100)
	reqDto.StakeAmount = utils.Truncate64(reqDto.StakeAmount / float64(BetFairRate))
	log.Println("PlaceBet: reqDto.StakeAmount - ", reqDto.StakeAmount)
	// 6. Check betvalue is greater than minbet value (#VI)
	log.Println("PlaceBet: MinBetValue is - ", MinBetValue)
	if reqDto.StakeAmount < float64(MinBetValue) {
		// TODO: Accept the bet and dont send to BetFair
		respObj, err := ValidateOdds(reqDto)
		if err != nil {
			// 1.1. Failed to validate odds
			log.Println("PlaceBet: Odds Validation call failed with - ", err.Error())
			return betDto, fmt.Errorf("Odds Changed!")
		}
		respJson, _ := json.Marshal(respObj)
		//log.Println("PlaceBet: Event Open Date is - ", time.Unix(respObj.OpenDate/1000, 0))
		if !respObj.IsValid {
			// 1.2. Invalid Odds
			log.Println("PlaceBet: Invalid Odds - ", reqDto.OddValue)
			log.Println("PlaceBet: ValidateOdds response json is - ", string(respJson))
			return betDto, fmt.Errorf("Odds Changed!")
		}
		// 2. Check for acceptance criteria
		if respObj.Status != "IN_PLAY" {
			// 2.1. Match Odds bet on upcoming match
			// TODO: add to collection
			log.Println("PlaceBet: MATCH_ODDS bet only allowed on IN_PLAY events - ", respObj.Status)
			log.Println("PlaceBet: ValidateOdds response json is - ", string(respJson))
			return betDto, fmt.Errorf("Betting NOT ALLOWED!")
		}
		betDto.Status = constants.SAP.BetStatus.OPEN()
		log.Println("PlaceBet: PlaceOrder, BetFair oddvalue - Requsted at OddValue - ", betDto.BetDetails.OddValue)
		log.Println("PlaceBet: PlaceOrder, BetFair oddvalue - Matched  at OddValue - ", respObj.MatchedOddValue)
		if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respObj.MatchedOddValue > betDto.BetDetails.OddValue {
			betDto.BetDetails.OddValue = respObj.MatchedOddValue
		}
		if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respObj.MatchedOddValue < betDto.BetDetails.OddValue {
			betDto.BetDetails.OddValue = respObj.MatchedOddValue
		}
		// bet Sizes
		betDto.BetReq.SizePlaced = reqDto.StakeAmount
		betDto.BetReq.SizeMatched = reqDto.StakeAmount
		betDto.BetReq.SizeRemaining = 0
		betDto.BetReq.SizeLapsed = 0
		betDto.BetReq.SizeCancelled = 0
		betDto.BetReq.SizeVoided = 0
		betDto.BetReq.OddsMatched = respObj.MatchedOddValue
	} else {
		// 3. BetFair - Place Order
		log.Println("PlaceBet: PlaceOrder Before")
		respDto, err := PlaceOrder(reqDto)
		if err != nil {
			log.Println("PlaceBet: PlaceOrder failed with error - ", err.Error())
			return betDto, err
		}
		log.Println("PlaceBet: PlaceOrder Status is - ", respDto.Status)
		if respDto.Status != "SUCCESS" {
			// API Call failed. Mark bet status as failed. and return
			log.Println("PlaceBet: PlaceOrder API Status is - ", respDto.Status)
			log.Println("PlaceBet: PlaceOrder API ErrorCode is - ", respDto.ErrorCode)
			for _, inst := range respDto.InstructionReports {
				log.Println("PlaceBet: PlaceOrder InstructionReport Status - ", inst.Status)
				log.Println("PlaceBet: PlaceOrder InstructionReport ErrorCode is - ", inst.ErrorCode)
				log.Println("PlaceBet: PlaceOrder InstructionReport OrderStatus is - ", inst.OrderStatus)
			}
			return betDto, fmt.Errorf("BetFair returned failure!")
		}
		if len(respDto.InstructionReports) < 1 {
			log.Println("PlaceBet: InstructionReports count is 0 - ", len(respDto.InstructionReports))
			return betDto, fmt.Errorf("BetFair returned failure!")
		}
		if respDto.InstructionReports[0].Status != "SUCCESS" {
			// API Call failed. Mark bet status as failed. and return
			log.Println("PlaceBet: PlaceOrder, InstructionReports[0] Status is - ", respDto.InstructionReports[0].Status)
			log.Println("PlaceBet: PlaceOrder, InstructionReports[0] ErrorCode is - ", respDto.InstructionReports[0].ErrorCode)
			log.Println("PlaceBet: PlaceOrder InstructionReport OrderStatus is - ", respDto.InstructionReports[0].OrderStatus)
			return betDto, fmt.Errorf("BetFair returned failure!")
		}
		// 4. Check operator wallet type
		switch respDto.InstructionReports[0].OrderStatus {
		case "EXECUTION_COMPLETE":
			betDto.Status = "OPEN"
			betDto.BetReq.BetId = respDto.InstructionReports[0].BetId
			log.Println("PlaceBet: PlaceOrder, BetFair oddvalue - Requsted at OddValue - ", reqDto.OddValue)
			log.Println("PlaceBet: PlaceOrder, BetFair oddvalue - Matched  at OddValue - ", respDto.InstructionReports[0].AveragePriceMatched)
			if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respDto.InstructionReports[0].AveragePriceMatched > betDto.BetDetails.OddValue {
				betDto.BetDetails.OddValue = respDto.InstructionReports[0].AveragePriceMatched
			}
			if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respDto.InstructionReports[0].AveragePriceMatched < betDto.BetDetails.OddValue {
				betDto.BetDetails.OddValue = respDto.InstructionReports[0].AveragePriceMatched
			}
			// bet Sizes
			betDto.BetReq.SizePlaced = reqDto.StakeAmount
			betDto.BetReq.SizeMatched = reqDto.StakeAmount
			betDto.BetReq.SizeRemaining = 0
			betDto.BetReq.SizeLapsed = 0
			betDto.BetReq.SizeCancelled = 0
			betDto.BetReq.SizeVoided = 0
			betDto.BetReq.OddsMatched = respDto.InstructionReports[0].AveragePriceMatched
		case "EXECUTABLE":
			betDto.Status = "UNMATCHED"
			betDto.BetReq.BetId = respDto.InstructionReports[0].BetId
			// bet Sizes
			betDto.BetReq.SizePlaced = reqDto.StakeAmount
			betDto.BetReq.SizeMatched = respDto.InstructionReports[0].SizeMatched
			betDto.BetReq.SizeRemaining = betDto.BetReq.SizePlaced - betDto.BetReq.SizeMatched
			betDto.BetReq.SizeLapsed = 0
			betDto.BetReq.SizeCancelled = 0
			betDto.BetReq.SizeVoided = 0
			betDto.BetReq.OddsMatched = respDto.InstructionReports[0].AveragePriceMatched
			if betDto.BetReq.SizeRemaining != 0 && betDto.BetReq.SizeMatched != 0 {
				log.Println("PlaceBet: PartialMatched Bet Found for BetFair betId - ", betDto.BetReq.BetId)
				log.Println("PlaceBet: PartialMatched Bet SizeMatched - ", betDto.BetReq.SizeMatched)
				log.Println("PlaceBet: PartialMatched Bet SizeRemaining - ", betDto.BetReq.SizeRemaining)
			}
		case "EXPIRED":
			log.Println("PlaceBet: Unexpected PlaceOrder InstructionReport BetId is - ", respDto.InstructionReports[0].BetId)
			return betDto, fmt.Errorf("FAILED!")
		default:
			log.Println("PlaceBet: Unexpected PlaceOrder InstructionReport OrderStatus is - ", respDto.InstructionReports[0].OrderStatus)
			log.Println("PlaceBet: Unexpected PlaceOrder InstructionReport BetId is - ", respDto.InstructionReports[0].BetId)
			return betDto, fmt.Errorf(respDto.InstructionReports[0].OrderStatus)
		}
	}
	log.Println("PlaceBet: BetId - ", betDto.BetReq.BetId)
	log.Println("PlaceBet: Bet Status - ", betDto.Status)
	// 7. Bet Success, Add to operator ledger
	err = providers.OperatorLedgerTx(operator, constants.SAP.LedgerTxType.BETPLACEMENT(), betDto.BetReq.OperatorAmount*-1, betDto.BetId)
	if err != nil {
		log.Println("PlaceBet: OperatorLedgerTx failed with error - ", err.Error())
	}
	return betDto, nil
}
*/

func BetFairCancelBet(reqDto requestdto.CancelBetReqDto, betDtos []sports.BetDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto) ([]sports.BetDto, []responsedto.CancelBetResp, error) {
	log.Println("BetFairCancelBet: Bets Count is - ", len(betDtos))
	updateBets := []sports.BetDto{}
	CancelBetsResp := []responsedto.CancelBetResp{}
	// 1. BetFair - Cancel Order
	respDto, err := CancelOrder(reqDto, betDtos, BetFairRate)
	if err != nil {
		log.Println("BetFairCancelBet: CancelOrder failed with error - ", err.Error())
		return updateBets, CancelBetsResp, err
	}
	// TODO: Handle BetFair Response
	if respDto.Status != "SUCCESS" {
		log.Println("BetFairCancelBet: CancelOrder Status is - ", respDto.Status)
		log.Println("BetFairCancelBet: CancelOrder ErrorCode is - ", respDto.ErrorCode)
		return updateBets, CancelBetsResp, fmt.Errorf(respDto.ErrorCode)
	}
	log.Println("BetFairCancelBet: CancelOrder Status is - ", respDto.Status)
	// 4. Update
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	var operatorDelta float64 = 0
	for _, insReport := range respDto.InstructionReports {
		cancelBetResp := responsedto.CancelBetResp{}
		cancelBetResp.Status = insReport.Status
		cancelBetResp.ErrorCode = insReport.ErrorCode
		cancelBetResp.BetId = insReport.Instruction.BetId
		//cancelBetResp.SizeCancelled = insReport.SizeCancelled
		if insReport.Status == "SUCCESS" {
			for _, betDto := range betDtos {
				if insReport.Instruction.BetId == betDto.BetReq.BetId {
					// update bet sizes
					betDto.BetReq.SizeCancelled += insReport.SizeCancelled
					betDto.BetReq.SizeRemaining -= insReport.SizeCancelled
					if betDto.BetReq.SizeRemaining == 0 && betDto.BetReq.SizeMatched == 0 {
						betDto.Status = constants.SAP.BetStatus.CANCELLED()
					}
					if betDto.BetReq.SizeRemaining == 0 && betDto.BetReq.SizeMatched != 0 {
						betDto.Status = constants.SAP.BetStatus.OPEN()
					}
					if betDto.BetReq.SizeRemaining != 0 {
						betDto.Status = constants.SAP.BetStatus.UNMATCHED()
					}
					// matched amount
					sizeMatched := betDto.BetReq.SizeMatched * float64(BetFairRate)
					sizeMatched = (sizeMatched * 100) / (100 - betDto.BetReq.PlatformHold)
					sizeMatched = (sizeMatched * 100) / (100 - betDto.BetReq.OperatorHold)
					if betDto.BetReq.Rate != 0 {
						sizeMatched = utils.Truncate4Decfloat64(sizeMatched / float64(betDto.BetReq.Rate))
					}
					// remaining amount
					sizeRemaining := betDto.BetReq.SizeRemaining * float64(BetFairRate)
					sizeRemaining = (sizeRemaining * 100) / (100 - betDto.BetReq.PlatformHold)
					sizeRemaining = (sizeRemaining * 100) / (100 - betDto.BetReq.OperatorHold)
					if betDto.BetReq.Rate != 0 {
						sizeRemaining = utils.Truncate4Decfloat64(sizeRemaining / float64(betDto.BetReq.Rate))
					}
					// cancelling amount
					sizeCancelled := betDto.BetReq.SizeCancelled * float64(BetFairRate)
					sizeCancelled = (sizeCancelled * 100) / (100 - betDto.BetReq.PlatformHold)
					sizeCancelled = (sizeCancelled * 100) / (100 - betDto.BetReq.OperatorHold)
					if betDto.BetReq.Rate != 0 {
						sizeCancelled = utils.Truncate4Decfloat64(sizeCancelled / float64(betDto.BetReq.Rate))
					}
					cancelBetResp.SizeMatched = sizeMatched
					cancelBetResp.SizeRemaining = sizeRemaining
					cancelBetResp.SizeCancelled = sizeCancelled
					cancelBetResp.BetId = betDto.BetId
					// add rollback request
					rollbackReq := sports.RollbackReqDto{}
					rollbackReq.ReqId = uuid.New().String()
					rollbackReq.ReqTime = time.Now().UnixMilli()
					rollbackReq.RollbackReason = constants.BetFair.BetStatus.CANCELLED()
					if betDto.BetDetails.BetType == constants.SAP.BetType.BACK() {
						rollbackReq.PlatformAmount = insReport.SizeCancelled * float64(BetFairRate)
					} else {
						rollbackReq.PlatformAmount = utils.Truncate4Decfloat64(insReport.SizeCancelled * (betDto.BetDetails.OddValue - 1) * float64(BetFairRate))
					}
					rollbackReq.OperatorAmount = utils.Truncate4Decfloat64((rollbackReq.PlatformAmount * 100) / (100 - betDto.BetReq.PlatformHold))
					rollbackReq.RollbackAmount = utils.Truncate4Decfloat64((rollbackReq.OperatorAmount * 100) / (100 - betDto.BetReq.OperatorHold))
					betDto.RollbackReqs = append(betDto.RollbackReqs, rollbackReq)
					betDto.NetAmount = utils.Truncate4Decfloat64(betDto.NetAmount + rollbackReq.RollbackAmount)
					// Add Bet to updated bets list
					updateBets = append(updateBets, betDto)
					// Add OperatorLedgerTx
					operatorLedgerTx := providers.GetOperatorLedgerTx(operatorDto, constants.SAP.LedgerTxType.BETCANCEL(), rollbackReq.OperatorAmount, betDto.BetId)
					operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
					// Add to Operator Balance Delta
					operatorDelta += rollbackReq.OperatorAmount
					break
				}
			}
		} else {
			log.Println("BetFairCancelBet: Cancel Bet Failed with Status - ", insReport.Status)
			log.Println("BetFairCancelBet: Cancel Bet failed for BetId - ", insReport.Instruction.BetId)
		}
		CancelBetsResp = append(CancelBetsResp, cancelBetResp)
	}
	if len(updateBets) == 0 {
		log.Println("BetFairCancelBet: updateBets count is ZERO - ", len(updateBets))
		return updateBets, CancelBetsResp, fmt.Errorf("ZERO Bets Cancelled!")
	}
	// 5. Update Bet in db
	/*
		status := constants.BetFair.BetStatus.CANCELLED()
		count, err := database.UpdateBetsStatus(updateBets, status)
		if err != nil {
			log.Println("BetFairCancelBet: UpdateBetsStatus failed with error - ", err.Error())
			return updateBets, err
		}
		if len(updateBets) != count {
			log.Println("BetFairCancelBet: UpdateBetsStatus failed for few bets - ", len(updateBets)-count)
		}
	*/
	// 6. Insert Operator Ledger Txs
	if len(operatorLedgerTxs) > 0 {
		err = database.InsertOperatorLedgers(operatorLedgerTxs)
		if err != nil {
			// inserting ledger documents failed
			log.Println("BetFairCancelBet: database.InsertOperatorLedgers failed with error - ", err.Error())
		}
	}
	// Update Operator Balance balance and save
	if operatorDelta != 0 {
		err = database.UpdateOperatorBalance(operatorDto.OperatorId, operatorDelta)
		if err != nil {
			// updating operator balance failed
			log.Println("BetFairCancelBet: database.UpdateOperatorBalance failed with error - ", err.Error())
		}
	}
	jsonResp, err := json.Marshal(CancelBetsResp)
	if err != nil {
		// updating operator balance failed
		log.Println("BetFairCancelBet: json.Marshal failed with error - ", err.Error())
	}
	log.Println("BetFairCancelBet: CancelBetsResp is - ", string(jsonResp))
	return updateBets, CancelBetsResp, nil
}

// func BetFairOrders() {
// 	// log.Println("BetFairOrders: *** START *** ")
// 	// 1. Current Orders to find unmatched bets status
// 	//log.Println("BetFairOrders: 1. Current Orders ")
// 	BetFairCurretnOrders()
// 	// 2. Cleared Orders to find open(matched) bets status
// 	//log.Println("BetFairOrders: 2. Cleared Orders ")
// 	BetFairClearedOrders()
// 	// log.Println("BetFairOrders: *** END *** ")
// }

func BetFairCurretnOrders() {
	log.Println("BetFairCurretnOrders: START!!!")
	// 1. Get Unmatched Bets
	unmatchedBets, err := database.GetBetsByStatus("BetFair", "UNMATCHED")
	if err != nil {
		log.Println("BetFairCurretnOrders: GetBetsByStatus failed with error - ", err.Error())
		log.Println("BetFairCurretnOrders: END FAILED 1!!!")
		return
	}
	if len(unmatchedBets) == 0 {
		log.Println("BetFairCurretnOrders: END SUCCESS 2!!!")
		return
	}
	//log.Println("BetFairCurretnOrders: UNMATCHED bets count is - ", len(unmatchedBets))
	// 2. Make BetFair Current Orders call
	currentOrdersResp, err := CurrentOrders()
	if err != nil {
		log.Println("BetFairCurretnOrders: CurrentOrders failed with error - ", err.Error())
		log.Println("BetFairCurretnOrders: END FAILED 3!!!")
		return
	}
	//log.Println("BetFairCurretnOrders: CurrentOrders count is - ", len(currentOrdersResp.CurrentOrders))
	if len(currentOrdersResp.CurrentOrders) == 0 {
		log.Println("BetFairCurretnOrders: END SUCCESS 4!!!")
		return
	}
	// 3. Create two empty bets lists for open & expired
	openBets := []sports.BetDto{}
	expiredBets := []sports.BetDto{}
	partialBets := []sports.BetDto{}
	operatorsMap := make(map[string]operatordto.OperatorDTO)
	operatorsDelta := make(map[string]float64)
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	//execution_complete := 0
	//expired := 0
	executable := 0
	misc := 0
	for _, currentOrder := range currentOrdersResp.CurrentOrders {
		// log.Printf("currentOrder: %+v\n", currentOrder)
		switch currentOrder.Status {
		case constants.BetFair.OrderStatus.EXECUTION_COMPLETE():
			for _, bet := range unmatchedBets {
				if bet.BetReq.BetId == currentOrder.BetID {
					bet.Status = "OPEN"
					// if bet.BetDetails.MarketType == constants.SAP.MarketType.LINE_ODDS() {
					// 	log.Println("BetFairCurretnOrders: Unmatched, BetFair oddvalue - Requsted at SessionOutcome - ", bet.BetDetails.SessionOutcome)
					// 	log.Println("BetFairCurretnOrders: Unmatched, BetFair oddvalue - Matched  at SessionOutcome - ", currentOrder.AveragePriceMatched)
					// 	if bet.BetDetails.BetType == constants.BetFair.Side.BACK() && currentOrder.AveragePriceMatched > bet.BetDetails.SessionOutcome {
					// 		bet.BetReq.OddsMatched = currentOrder.AveragePriceMatched
					// 	}
					// 	if bet.BetDetails.BetType == constants.BetFair.Side.LAY() && currentOrder.AveragePriceMatched < bet.BetDetails.SessionOutcome {
					// 		bet.BetReq.OddsMatched = currentOrder.AveragePriceMatched
					// 	}
					// } else {
					// 	log.Println("BetFairCurretnOrders: Unmatched, BetFair oddvalue - Requsted at OddValue - ", bet.BetDetails.OddValue)
					// 	log.Println("BetFairCurretnOrders: Unmatched, BetFair oddvalue - Matched  at OddValue - ", currentOrder.AveragePriceMatched)
					// 	if bet.BetDetails.BetType == constants.BetFair.Side.BACK() && currentOrder.AveragePriceMatched > bet.BetDetails.OddValue {
					// 		bet.BetReq.OddsMatched = currentOrder.AveragePriceMatched
					// 	}
					// 	if bet.BetDetails.BetType == constants.BetFair.Side.LAY() && currentOrder.AveragePriceMatched < bet.BetDetails.OddValue {
					// 		bet.BetReq.OddsMatched = currentOrder.AveragePriceMatched
					// 	}
					// }
					bet.BetReq.SizeMatched = currentOrder.SizeMatched
					bet.BetReq.SizeRemaining = currentOrder.SizeRemaining
					bet.BetReq.SizeLapsed = currentOrder.SizeLapsed
					bet.BetReq.SizeVoided = currentOrder.SizeVoided
					bet.BetReq.SizeCancelled = currentOrder.SizeCancelled
					bet.BetReq.OddsMatched = currentOrder.AveragePriceMatched
					openBets = append(openBets, bet)
					break
				}
			}
			//execution_complete++
		case constants.BetFair.OrderStatus.EXPIRED():
			for _, openBet := range unmatchedBets {
				if openBet.BetReq.BetId == currentOrder.BetID {
					value, ok := operatorsMap[openBet.OperatorId]
					if !ok {
						value, err = cache.GetOperatorDetails(openBet.OperatorId)
						operatorsMap[openBet.OperatorId] = value
					}
					delta, ok := operatorsDelta[openBet.OperatorId]
					if !ok {
						delta = 0
						operatorsDelta[openBet.OperatorId] = delta
					}
					// Add Rollback Request
					rollbackReq := providers.ComputeRollback(openBet, constants.BetFair.OrderStatus.EXPIRED())
					// Add OperatorLedgerTx
					operatorLedgerTx := providers.GetOperatorLedgerTx(value, constants.SAP.LedgerTxType.BETROLLBACK(), openBet.BetReq.OperatorAmount, openBet.BetId)
					operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
					// Add to Operator Balance Delta
					delta += rollbackReq.OperatorAmount
					operatorsDelta[openBet.OperatorId] = delta
					log.Println("BetFairCurretnOrders: BetId is : ", openBet.BetId)
					log.Println("BetFairCurretnOrders: rollbackReq.RollbackAmount is : ", rollbackReq.RollbackAmount)
					openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
					openBet.NetAmount += rollbackReq.RollbackAmount
					openBet.Status = constants.BetFair.OrderStatus.EXPIRED()
					// setting updatedAt
					openBet.UpdatedAt = rollbackReq.ReqTime
					expiredBets = append(expiredBets, openBet)
					break
				}
			}
			//expired++
		case constants.BetFair.OrderStatus.EXECUTABLE():
			executable++
			for _, bet := range unmatchedBets {
				if bet.BetReq.BetId == currentOrder.BetID {
					//bet.Status = "OPEN"
					isModified := false
					if bet.BetReq.SizeMatched != currentOrder.SizeMatched {
						isModified = true
						bet.BetReq.SizeMatched = currentOrder.SizeMatched
					}
					if bet.BetReq.SizeRemaining != currentOrder.SizeRemaining {
						isModified = true
						bet.BetReq.SizeRemaining = currentOrder.SizeRemaining
					}
					if bet.BetReq.SizeLapsed != currentOrder.SizeLapsed {
						isModified = true
						bet.BetReq.SizeLapsed = currentOrder.SizeLapsed
					}
					if bet.BetReq.SizeVoided != currentOrder.SizeVoided {
						isModified = true
						bet.BetReq.SizeVoided = currentOrder.SizeVoided
					}
					if bet.BetReq.SizeCancelled != currentOrder.SizeCancelled {
						isModified = true
						bet.BetReq.SizeCancelled = currentOrder.SizeCancelled
					}
					if currentOrder.AveragePriceMatched > 0 && bet.BetReq.OddsMatched != currentOrder.AveragePriceMatched {
						bet.BetReq.OddsMatched = currentOrder.AveragePriceMatched
						isModified = true
					}
					if bet.BetReq.OddsMatched == 0 {
						bet.BetReq.OddsMatched = bet.BetDetails.OddValue
					}
					if isModified == true {
						log.Println("BetFairCurretnOrders: currentOrder.SizeMatched - ", currentOrder.SizeMatched)
						log.Println("BetFairCurretnOrders: currentOrder.SizeLapsed - ", currentOrder.SizeLapsed)
						log.Println("BetFairCurretnOrders: currentOrder.SizeVoided - ", currentOrder.SizeVoided)
						log.Println("BetFairCurretnOrders: currentOrder.SizeCancelled - ", currentOrder.SizeCancelled)
						log.Println("BetFairCurretnOrders: currentOrder.SizeRemaining - ", currentOrder.SizeRemaining)
						log.Println("BetFairCurretnOrders: currentOrder.AveragePriceMatched - ", currentOrder.AveragePriceMatched)
						partialBets = append(partialBets, bet)
					}
					break
				}
			}

		default:
			log.Println("BetFairCurretnOrders: Order.Status is - ", currentOrder.Status)
			misc++
		}
	}
	//log.Println("BetFairCurretnOrders: execution_complete count is: ", execution_complete)
	//log.Println("BetFairCurretnOrders: expired count is: ", len(expiredBets))
	//log.Println("BetFairCurretnOrders: executable count is: ", executable)
	//log.Println("BetFairCurretnOrders: misc count is: ", misc)
	// 4. Update bet in databse for OPEN bets
	if len(openBets) > 0 {
		log.Println("BetFairCurretnOrders: openBets count is : ", len(openBets))
		//database.UpdateBetsStatus(openBets, "OPEN")
		count, msgs := database.UpdateBets(openBets)
		if len(msgs) > 0 {
			log.Println("BetFairCurretnOrders: updatedBets count is : ", count)
			for _, msg := range msgs {
				log.Println("BetFairCurretnOrders: Error msg is : ", msg)
			}
		}
	}
	// 4. Update bet in databse for OPEN bets
	if len(partialBets) > 0 {
		log.Println("BetFairCurretnOrders: partialBets count is : ", len(partialBets))
		//database.UpdateBetsStatus(openBets, "OPEN")
		count, msgs := database.UpdateBets(partialBets)
		if len(msgs) > 0 {
			log.Println("BetFairCurretnOrders: updatedBets count is : ", count)
			for _, msg := range msgs {
				log.Println("BetFairCurretnOrders: Error msg is : ", msg)
			}
		}
	}
	if len(expiredBets) == 0 {
		//log.Println("BetFairRollbackOrders: ZERO bets to void.")
		log.Println("BetFairCurretnOrders: END SUCCESS 5!!!")
		return
	}
	// 5. Insert Operator Ledger Txs
	if len(operatorLedgerTxs) > 0 {
		err = database.InsertOperatorLedgers(operatorLedgerTxs)
		if err != nil {
			// inserting ledger documents failed
			log.Println("BetFairCurretnOrders: database.InsertOperatorLedgers failed with error - ", err.Error())
			log.Println("BetFairCurretnOrders: END FAILED 6!!!")
		}
	}
	// 6. Update Operator Balance balance and save
	for operatorId, delta := range operatorsDelta {
		if delta != 0 {
			err = database.UpdateOperatorBalance(operatorId, delta)
			if err != nil {
				// updating operator balance failed
				log.Println("BetFairCurretnOrders: database.UpdateOperatorBalance failed with error - ", err.Error())
				log.Println("BetFairCurretnOrders: END FAILED 7!!!")
			}
		}
	}
	// 5. Process Voided Bets
	log.Println("BetFairCurretnOrders: rollbackBets count is - ", len(expiredBets))
	//operator.BetFairRollbackRoutine(rollbackBets, "Bet Voided", RollbackType)
	operator.CommonRollbackRoutine(constants.BetFair.OrderStatus.EXPIRED(), expiredBets)
	log.Println("BetFairCurretnOrders: Ended for RollbackType is - ", constants.BetFair.OrderStatus.EXPIRED())
	log.Println("BetFairCurretnOrders: END SUCCESS 8!!!")
	return
}

func BetFairClearedOrders() {
	// 1.1. Get Open Bets list
	openBets2, err := database.GetBetsByStatus("BetFair", "OPEN")
	if err != nil {
		log.Println("BetFairOrders: GetBetsByStatus failed with error - ", err.Error())
		return
	}
	if len(openBets2) == 0 {
		return
	}
	//log.Println("BetFairOrders: openBets2 count is : ", len(openBets2))
	openBets := []sports.BetDto{}
	for _, bet := range openBets2 {
		if bet.BetId == bet.BetReq.BetId {
			continue
		}
		openBets = append(openBets, bet)
	}
	if len(openBets) == 0 {
		return
	}
	//log.Println("BetFairOrders: openBets count is : ", len(openBets))
	// 1.2. Create betIds and marketIds lists
	betIds := []string{}
	//marketIds := []string{}
	for _, bet := range openBets {
		if bet.BetId != bet.BetReq.BetId {
			betIds = append(betIds, bet.BetReq.BetId)
			//marketIds = append(marketIds, bet.MarketId)
		}
	}
	if len(betIds) == 0 {
		return
	}
	//log.Println("BetFairOrders: OPEN bets count is : ", len(betIds))
	// 1.3. Process settled bets
	//ClearedOrders("SETTLED", betIds, marketBookResp, bets)
	BetFairSettledOrders(betIds, openBets)
	// 1.4. Process voided bets
	//ClearedOrders("VOIDED", betIds, marketBookResp, bets)
	//BetFairVoidedOrders(betIds, openBets)
	BetFairRollbackOrders(constants.BetFair.BetStatus.VOIDED(), betIds, openBets)
	// 2.1. Get Unmatched Bets
	unmatchedBets, err := database.GetBetsByStatus("BetFair", "UNMATCHED")
	if err != nil {
		log.Println("BetFairOrders: GetBetsByStatus failed with error - ", err.Error())
		return
	}
	if len(unmatchedBets) == 0 {
		return
	}
	//log.Println("BetFairOrders: Unmatched bets count is - ", len(unmatchedBets))
	// 2.2. Create betIds and marketIds lists
	betIds = []string{}
	//marketIds = []string{}
	for _, bet := range unmatchedBets {
		if bet.BetId != bet.BetReq.BetId {
			betIds = append(betIds, bet.BetReq.BetId)
			//marketIds = append(marketIds, bet.MarketId)
		}
	}
	if len(betIds) == 0 {
		return
	}
	log.Println("BetFairOrders: UNMATCHED bets count is : ", len(betIds))
	// 2.3. Process lapsed bets
	//ClearedOrders("LAPSED", betIds, marketBookResp, bets)
	//BetFairLapsedOrders(betIds, unmatchedBets)
	BetFairRollbackOrders(constants.BetFair.BetStatus.LAPSED(), betIds, unmatchedBets)
	// 2.4. process cancelled bets
	//ClearedOrders("CANCELLED", betIds, marketBookResp, bets)
	//BetFairCancelledOrders(betIds, unmatchedBets)
	BetFairRollbackOrders(constants.BetFair.BetStatus.CANCELLED(), betIds, unmatchedBets)
}

func BetFairSettledOrders(betIds []string, bets []sports.BetDto) {
	logKey := "BetFairSettledOrders: ClearedOrders "
	log.Println(logKey+"START - ", time.Now())
	// 1. Make BetFair Cleared Orders call with SETTLED betStatus
	clearedOrdersResp, err := ClearedOrders(constants.BetFair.BetStatus.SETTLED(), betIds, bets)
	if err != nil {
		log.Println(logKey+"END FAILED 1 with error - ", time.Now(), err.Error())
		return
	}
	//log.Println("BetFairSettledOrders: Open Bets count is - ", len(betIds))
	//log.Println("BetFairSettledOrders: Cleared Orders count is - ", len(clearedOrdersResp.ClearedOrders))
	// 2. Create an empty settled bet list
	// 2. Create a new map
	settledBets := []sports.BetDto{}
	operatorsMap := make(map[string]operatordto.OperatorDTO)
	operatorsDelta := make(map[string]float64)
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	// 3. Iterate through Cleared Orders
	for _, clearedOrder := range clearedOrdersResp.ClearedOrders {
		//log.Printf("BetStatus:  "+betStatus+" Single Order: %+v\n", clearedOrder)
		// 3.1. Iterate through bets
		for _, openBet := range bets {
			// 3.1.1. find matching bet
			if openBet.BetReq.BetId == clearedOrder.BetID {
				value, ok := operatorsMap[openBet.OperatorId]
				if !ok {
					value, err = cache.GetOperatorDetails(openBet.OperatorId)
					operatorsMap[openBet.OperatorId] = value
				}
				delta, ok := operatorsDelta[openBet.OperatorId]
				if !ok {
					delta = 0
					operatorsDelta[openBet.OperatorId] = delta
				}
				// 3.1.1.1. Create result request
				resultReq := sports.ResultReqDto{}
				resultReq.ReqId = uuid.New().String()
				resultReq.ReqTime = time.Now().UnixMilli()
				resultReq.CreditAmount = 0
				resultReq.OperatorAmount = 0
				resultReq.PlatformAmount = 0
				resultReq.RunnerName = ""
				resultReq.SessionOutcome = 0
				// 3.1.1.2. Calculate winning amount and set result runner name
				log.Println(logKey+"clearedOrder.BetID, BetOutcome, Profit: ", clearedOrder.BetID, clearedOrder.BetOutcome, clearedOrder.Profit)
				switch clearedOrder.BetOutcome {
				case constants.BetFair.BetOutcome.WON():
					// // #1 BACK Bet same Odds & Better Odds and no stake change
					// if constants.BetFair.Side.BACK() == openBet.BetDetails.BetType && openBet.BetReq.SizePlaced == clearedOrder.SizeSettled {
					// 	resultReq.RunnerName = openBet.BetDetails.RunnerName
					// 	// #1 BACK and Same Odds and Odds Changed
					// 	resultReq.PlatformAmount = utils.Truncate4Decfloat64(openBet.BetReq.PlatformAmount * clearedOrder.PriceMatched)
					// 	resultReq.OperatorAmount = utils.Truncate4Decfloat64(openBet.BetReq.OperatorAmount * clearedOrder.PriceMatched)
					// 	resultReq.CreditAmount = utils.Truncate4Decfloat64(openBet.BetReq.DebitAmount * clearedOrder.PriceMatched)
					// }
					// // #2 LAY Bet same Odds & Better Odds and no stake change
					// if constants.BetFair.Side.LAY() == openBet.BetDetails.BetType && openBet.BetReq.SizePlaced == clearedOrder.SizeSettled {
					// 	resultReq.PlatformAmount = utils.Truncate4Decfloat64(openBet.BetReq.PlatformAmount * openBet.BetDetails.OddValue / (openBet.BetDetails.OddValue - 1))
					// 	resultReq.OperatorAmount = utils.Truncate4Decfloat64(openBet.BetReq.OperatorAmount * openBet.BetDetails.OddValue / (openBet.BetDetails.OddValue - 1))
					// 	resultReq.CreditAmount = utils.Truncate4Decfloat64(openBet.BetReq.DebitAmount * openBet.BetDetails.OddValue / (openBet.BetDetails.OddValue - 1))
					// }
					// // #3 BACK Bet same Odds & Better Odds and stake changed
					// if constants.BetFair.Side.BACK() == openBet.BetDetails.BetType && openBet.BetReq.SizePlaced != clearedOrder.SizeSettled {
					// 	resultReq.RunnerName = openBet.BetDetails.RunnerName
					// 	// #1 BACK and Same Odds and Odds Changed
					// 	resultReq.PlatformAmount = utils.Truncate4Decfloat64(openBet.BetReq.PlatformAmount * clearedOrder.PriceMatched)
					// 	resultReq.OperatorAmount = utils.Truncate4Decfloat64(openBet.BetReq.OperatorAmount * clearedOrder.PriceMatched)
					// 	resultReq.CreditAmount = utils.Truncate4Decfloat64(openBet.BetReq.DebitAmount * clearedOrder.PriceMatched)
					// }
					// // #4 LAY Bet same Odds & Better Odds and stake changed

					// // #1 BACK and Same Odds
					// // #2 BACK and Odds Changed
					// // #3 LAY and Same Odds
					// // #4 LAY and Odds Changed

					// if constants.BetFair.Side.BACK() == openBet.BetDetails.BetType {
					// 	resultReq.RunnerName = openBet.BetDetails.RunnerName
					// 	// #1 BACK and Same Odds and Odds Changed
					// 	winningFactor := clearedOrder.SizeSettled / openBet.BetReq.SizePlaced
					// 	cancelFactor := 1 - winningFactor
					// 	// result = Cancel Amount + Winning Amount
					// 	resultReq.PlatformAmount = utils.Truncate4Decfloat64((openBet.BetReq.PlatformAmount * cancelFactor) + (openBet.BetReq.PlatformAmount * winningFactor * clearedOrder.PriceMatched))
					// 	//odds := clearedOrder.PriceMatched - 1
					// 	resultReq.PlatformAmount = utils.Truncate4Decfloat64(openBet.BetReq.PlatformAmount * clearedOrder.PriceMatched)
					// 	resultReq.OperatorAmount = utils.Truncate4Decfloat64(openBet.BetReq.OperatorAmount * clearedOrder.PriceMatched)
					// 	resultReq.CreditAmount = utils.Truncate4Decfloat64(openBet.BetReq.DebitAmount * clearedOrder.PriceMatched)
					// } else {
					// 	resultReq.PlatformAmount = utils.Truncate4Decfloat64(openBet.BetReq.PlatformAmount * openBet.BetDetails.OddValue / (openBet.BetDetails.OddValue - 1))
					// 	resultReq.OperatorAmount = utils.Truncate4Decfloat64(openBet.BetReq.OperatorAmount * openBet.BetDetails.OddValue / (openBet.BetDetails.OddValue - 1))
					// 	resultReq.CreditAmount = utils.Truncate4Decfloat64(openBet.BetReq.DebitAmount * openBet.BetDetails.OddValue / (openBet.BetDetails.OddValue - 1))
					// }
					if clearedOrder.Profit > 0 {
						if openBet.BetReq.PlatformAmount == 0 {
							resultReq.CreditAmount = openBet.BetReq.DebitAmount + clearedOrder.Profit
						} else {
							// var profit float64 = 0
							// if constants.BetFair.Side.BACK() == openBet.BetDetails.BetType {
							// 	resultReq.RunnerName = openBet.BetDetails.RunnerName
							// 	profit = (clearedOrder.PriceMatched - 1) * clearedOrder.SizeSettled
							// } else {
							// 	// TODO: Find right result ruuner name
							// 	profit = clearedOrder.SizeSettled
							// }
							// platformAmount := openBet.BetReq.PlatformAmount
							// for _, rollbackReq := range openBet.RollbackReqs {
							// 	platformAmount = platformAmount - rollbackReq.PlatformAmount
							// }
							// for _, resultReq := range openBet.ResultReqs {
							// 	platformAmount = platformAmount - resultReq.PlatformAmount
							// }
							openBet.BetReq.OddsMatched = clearedOrder.PriceMatched
							openBet.BetReq.SizeMatched = clearedOrder.SizeSettled
							betFairAmount := clearedOrder.PriceMatched * clearedOrder.SizeSettled // liability + profit
							//betFairAmount2 :=
							resultReq.PlatformAmount = utils.Truncate4Decfloat64(betFairAmount * float64(BetFairRate))
							resultReq.OperatorAmount = utils.Truncate4Decfloat64(resultReq.PlatformAmount * 100 / (100 - openBet.BetReq.PlatformHold))
							resultReq.CreditAmount = utils.Truncate4Decfloat64(resultReq.OperatorAmount * 100 / (100 - openBet.BetReq.OperatorHold))
							if constants.BetFair.Side.BACK() == openBet.BetDetails.BetType {
								resultReq.RunnerName = openBet.BetDetails.RunnerName
							} else {
								// TODO: Find right result ruuner name
								if openBet.BetDetails.OddValue != openBet.BetReq.OddsMatched {
									// calculate diff amount
									factor := openBet.BetReq.SizeMatched / openBet.BetReq.SizePlaced
									stakeAmount := utils.Truncate4Decfloat64(openBet.BetDetails.StakeAmount * factor)
									diffAmount := utils.Truncate4Decfloat64(stakeAmount * (openBet.BetDetails.OddValue - openBet.BetReq.OddsMatched))
									resultReq.CreditAmount += diffAmount
									log.Println(logKey+"Better Odds Result Adjusted - ", openBet.BetId, diffAmount, resultReq.CreditAmount, factor)
								}
							}
							// Add OperatorLedgerTx
							operatorLedgerTx := providers.GetOperatorLedgerTx(value, constants.SAP.LedgerTxType.BETRESULT(), resultReq.OperatorAmount, openBet.BetId)
							operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
							// Add to Operator Balance Delta
							delta += resultReq.OperatorAmount
							operatorsDelta[openBet.OperatorId] = delta
						}
					}
				case constants.BetFair.BetOutcome.LOSE(), "LOST":
					if constants.BetFair.Side.LAY() == openBet.BetDetails.BetType {
						resultReq.RunnerName = openBet.BetDetails.RunnerName
					} else {
						// TODO: Find right result ruuner name
					}
				case constants.BetFair.BetOutcome.PLACE():
					log.Println(logKey+"clearedOrder PLACE !!!", clearedOrder)
				default:
					log.Println(logKey+"clearedOrder ", clearedOrder)
				}
				log.Println(logKey+"resultReq.CreditAmount for BetId is : ", openBet.BetId, resultReq.CreditAmount)
				openBet.ResultReqs = append(openBet.ResultReqs, resultReq)
				openBet.NetAmount += resultReq.CreditAmount
				if openBet.NetAmount > 0 {
					commConfig := providers.GetCommissionConfig(openBet.UserId, openBet.MarketId, openBet.EventId, openBet.CompetitionId, openBet.SportId, openBet.PartnerId, openBet.OperatorId, openBet.ProviderId)
					if openBet.OperatorId != "kiaexch" && openBet.OperatorId != "kia-sap" {
						openBet.Commission = utils.Truncate4Decfloat64(openBet.NetAmount * commConfig.CommPercentage / 100)
					}
					openBet.CommLevel = commConfig.CommLevel
				}
				openBet.Status = constants.BetFair.BetStatus.SETTLED()
				// setting updatedAt
				openBet.UpdatedAt = resultReq.ReqTime
				settledBets = append(settledBets, openBet)
				break
			}
		}
	}
	// 4. Check for length, return if 0
	if len(settledBets) == 0 {
		log.Println(logKey+"END SUCCESS 2 ZERO bets to settle - ", time.Now())
		return
	}
	// 5. Insert Operator Ledger Txs
	if len(operatorLedgerTxs) > 0 {
		err = database.InsertOperatorLedgers(operatorLedgerTxs)
		if err != nil {
			// inserting ledger documents failed
			log.Println(logKey+"database.InsertOperatorLedgers failed with error - ", err.Error())
		}
	}
	// 6. Update Operator Balance balance and save
	for operatorId, delta := range operatorsDelta {
		if delta != 0 {
			err = database.UpdateOperatorBalance(operatorId, delta)
			if err != nil {
				// updating operator balance failed
				log.Println(logKey+"database.UpdateOperatorBalance failed with error - ", err.Error())
			}
		}
	}
	// 5. Process Settled Bets
	log.Println(logKey+"settledBets count is - ", len(settledBets))
	operator.CommonResultRoutine(settledBets)
	log.Println(logKey+"END SUCCESS 3 - ", time.Now())
	return
}

func BetFairRollbackOrders(RollbackType string, betIds []string, bets []sports.BetDto) {
	//log.Println("BetFairRollbackOrders: Strated for RollbackType is - ", RollbackType)
	logKey := "BetFairRollbackOrders: ClearedOrders " + RollbackType
	// 1. Make BetFair Cleared Orders call with SETTLED betStatus
	clearedOrdersResp, err := ClearedOrders(RollbackType, betIds, bets)
	if err != nil {
		log.Println(logKey+" failed with error - ", err.Error())
		return
	}
	//log.Println("BetFairRollbackOrders: Open Bets count is - ", len(betIds))
	//log.Println("BetFairRollbackOrders: Cleared Orders count is - ", len(clearedOrdersResp.ClearedOrders))
	// 2. Create an empty voided bet list
	rollbackBets := []sports.BetDto{}
	operatorsMap := make(map[string]operatordto.OperatorDTO)
	operatorsDelta := make(map[string]float64)
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	// 3. Iterate through Cleared Orders
	for _, clearedOrder := range clearedOrdersResp.ClearedOrders {
		//log.Printf("BetStatus:  "+betStatus+" Single Order: %+v\n", clearedOrder)
		// 3.1. Iterate through bets
		for _, openBet := range bets {
			// 3.1.1. find matching bet
			if openBet.BetReq.BetId == clearedOrder.BetID {
				value, ok := operatorsMap[openBet.OperatorId]
				if !ok {
					value, err = cache.GetOperatorDetails(openBet.OperatorId)
					operatorsMap[openBet.OperatorId] = value
				}
				delta, ok := operatorsDelta[openBet.OperatorId]
				if !ok {
					delta = 0
					operatorsDelta[openBet.OperatorId] = delta
				}
				// update bet sizes
				switch RollbackType {
				case constants.BetFair.BetStatus.CANCELLED():
					openBet.BetReq.SizeCancelled += clearedOrder.SizeCancelled
				case constants.BetFair.BetStatus.LAPSED():
					openBet.BetReq.SizeLapsed += clearedOrder.SizeCancelled
				case constants.BetFair.BetStatus.VOIDED():
					openBet.BetReq.SizeVoided += clearedOrder.SizeCancelled
				default:
					openBet.BetReq.SizeCancelled += clearedOrder.SizeCancelled
				}
				sizeCancelled := openBet.BetReq.SizeCancelled + openBet.BetReq.SizeLapsed + openBet.BetReq.SizeVoided
				openBet.BetReq.SizeRemaining -= clearedOrder.SizeCancelled
				if openBet.BetReq.SizePlaced == sizeCancelled {
					openBet.Status = RollbackType
					log.Println(logKey+" COMPLETELY - ", openBet.BetId)
				} else {
					openBet.Status = constants.SAP.BetStatus.OPEN()
					log.Println(logKey+" PARTIALLY (Sizematched & SizeRemaining & SizeCancelled) is - ", openBet.BetId, openBet.BetReq.SizeMatched, openBet.BetReq.SizeRemaining, clearedOrder.SizeCancelled)
				}
				// Add Rollback Request
				rollbackReq := sports.RollbackReqDto{}
				rollbackReq.ReqId = uuid.New().String()
				rollbackReq.ReqTime = time.Now().UnixMilli()
				rollbackReq.RollbackReason = constants.BetFair.BetStatus.LAPSED()
				if openBet.BetDetails.BetType == constants.SAP.BetType.BACK() {
					rollbackReq.PlatformAmount = clearedOrder.SizeCancelled * float64(BetFairRate)
				} else {
					rollbackReq.PlatformAmount = utils.Truncate4Decfloat64(clearedOrder.SizeCancelled * (openBet.BetDetails.OddValue - 1) * float64(BetFairRate))
				}
				rollbackReq.OperatorAmount = utils.Truncate4Decfloat64((rollbackReq.PlatformAmount * 100) / (100 - openBet.BetReq.PlatformHold))
				rollbackReq.RollbackAmount = utils.Truncate4Decfloat64((rollbackReq.OperatorAmount * 100) / (100 - openBet.BetReq.OperatorHold))
				// Add OperatorLedgerTx
				operatorLedgerTx := providers.GetOperatorLedgerTx(value, constants.SAP.LedgerTxType.BETROLLBACK(), rollbackReq.OperatorAmount, openBet.BetId)
				operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
				// Add to Operator Balance Delta
				delta += rollbackReq.OperatorAmount
				operatorsDelta[openBet.OperatorId] = delta
				log.Println(logKey+" rollbackReq.RollbackAmount is : ", openBet.BetId, rollbackReq.RollbackAmount)
				openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
				openBet.NetAmount = utils.Truncate4Decfloat64(openBet.NetAmount + rollbackReq.RollbackAmount)
				// setting updatedAt
				openBet.UpdatedAt = rollbackReq.ReqTime
				rollbackBets = append(rollbackBets, openBet)
				break
			}
		}
	}
	// 4. Check for length, return if 0
	if len(rollbackBets) == 0 {
		log.Println(logKey + " ZERO bets to Rollback.")
		return
	}
	// 5. Insert Operator Ledger Txs
	if len(operatorLedgerTxs) > 0 {
		err = database.InsertOperatorLedgers(operatorLedgerTxs)
		if err != nil {
			// inserting ledger documents failed
			log.Println(logKey+" database.InsertOperatorLedgers failed with error - ", err.Error())
		}
	}
	// 6. Update Operator Balance balance and save
	for operatorId, delta := range operatorsDelta {
		if delta != 0 {
			err = database.UpdateOperatorBalance(operatorId, delta)
			if err != nil {
				// updating operator balance failed
				log.Println(logKey+" database.UpdateOperatorBalance failed with error - ", err.Error())
			}
		}
	}
	// 5. Process Voided Bets
	log.Println(logKey+" rollbackBets count is - ", len(rollbackBets))
	//operator.BetFairRollbackRoutine(rollbackBets, "Bet Voided", RollbackType)
	operator.CommonRollbackRoutine(RollbackType, rollbackBets)
	//log.Println(logKey + " Ended for RollbackType is - ", RollbackType)
	return
}

/*
func BetFairVoidedOrders(betIds []string, bets []sports.BetDto) {
	// 1. Make BetFair Cleared Orders call with SETTLED betStatus
	clearedOrdersResp, err := ClearedOrders(constants.BetFair.BetStatus.VOIDED(), betIds, bets)
	if err != nil {
		log.Println("BetFairVoidedOrders: ClearedOrders failed with error - ", err.Error())
		return
	}
	log.Println("BetFairVoidedOrders: Open Bets count is - ", len(betIds))
	log.Println("BetFairVoidedOrders: Cleared Orders count is - ", len(clearedOrdersResp.ClearedOrders))
	// 2. Create an empty voided bet list
	voidedBets := []sports.BetDto{}
	operatorsMap := make(map[string]operatordto.OperatorDTO)
	operatorsDelta := make(map[string]float64)
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	// 3. Iterate through Cleared Orders
	for _, clearedOrder := range clearedOrdersResp.ClearedOrders {
		//log.Printf("BetStatus:  "+betStatus+" Single Order: %+v\n", clearedOrder)
		// 3.1. Iterate through bets
		for _, openBet := range bets {
			// 3.1.1. find matching bet
			if openBet.BetReq.BetId == clearedOrder.BetID {
				value, ok := operatorsMap[openBet.OperatorId]
				if !ok {
					value, err = cache.GetOperatorDetails(openBet.OperatorId)
					operatorsMap[openBet.OperatorId] = value
				}
				delta, ok := operatorsDelta[openBet.OperatorId]
				if !ok {
					delta = 0
					operatorsDelta[openBet.OperatorId] = delta
				}
				// Add Rollback Request
				rollbackReq := providers.ComputeRollback(openBet, constants.BetFair.BetStatus.VOIDED())
				// Add OperatorLedgerTx
				operatorLedgerTx := providers.GetOperatorLedgerTx(value, constants.SAP.LedgerTxType.BETROLLBACK(), openBet.BetReq.OperatorAmount, openBet.BetId)
				operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
				// Add to Operator Balance Delta
				delta += rollbackReq.OperatorAmount
				operatorsDelta[openBet.OperatorId] = delta
				log.Println("BetFairVoidedOrders: BetId is : ", openBet.BetId)
				log.Println("BetFairVoidedOrders: rollbackReq.RollbackAmount is : ", rollbackReq.RollbackAmount)
				openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
				openBet.Status = constants.BetFair.BetStatus.VOIDED()
				// setting updatedAt
				openBet.UpdatedAt = rollbackReq.ReqTime
				voidedBets = append(voidedBets, openBet)
				break
			}
		}
	}
	// 4. Check for length, return if 0
	if len(voidedBets) == 0 {
		//log.Println("BetFairVoidedOrders: ZERO bets to void.")
		return
	}
	// 5. Insert Operator Ledger Txs
	if len(operatorLedgerTxs) > 0 {
		err = database.InsertOperatorLedgers(operatorLedgerTxs)
		if err != nil {
			// inserting ledger documents failed
			log.Println("BetFairVoidedOrders: database.InsertOperatorLedgers failed with error - ", err.Error())
		}
	}
	// 6. Update Operator Balance balance and save
	for operatorId, delta := range operatorsDelta {
		if delta != 0 {
			err = database.UpdateOperatorBalance(operatorId, delta)
			if err != nil {
				// updating operator balance failed
				log.Println("BetFairVoidedOrders: database.UpdateOperatorBalance failed with error - ", err.Error())
			}
		}
	}
	// 5. Process Voided Bets
	log.Println("BetFairVoidedOrders: voidedBets count is - ", len(voidedBets))
	//operator.BetFairRollbackRoutine(voidedBets, "Bet Voided", constants.BetFair.BetStatus.VOIDED())
	operator.CommonRollbackRoutine(constants.BetFair.BetStatus.VOIDED(), voidedBets)
	return
}

func BetFairLapsedOrders(betIds []string, bets []sports.BetDto) {
	// 1. Make BetFair Cleared Orders call with SETTLED betStatus
	clearedOrdersResp, err := ClearedOrders(constants.BetFair.BetStatus.LAPSED(), betIds, bets)
	if err != nil {
		log.Println("BetFairLapsedOrders: ClearedOrders failed with error - ", err.Error())
		return
	}
	log.Println("BetFairLapsedOrders: Unmatched Bets count is - ", len(betIds))
	log.Println("BetFairLapsedOrders: Cleared Orders count is - ", len(clearedOrdersResp.ClearedOrders))
	// 2. Create an empty voided bet list
	lapsedBets := []sports.BetDto{}
	operatorsMap := make(map[string]operatordto.OperatorDTO)
	operatorsDelta := make(map[string]float64)
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	// 3. Iterate through Cleared Orders
	for _, clearedOrder := range clearedOrdersResp.ClearedOrders {
		//log.Printf("BetStatus:  "+betStatus+" Single Order: %+v\n", clearedOrder)
		// 3.1. Iterate through bets
		for _, openBet := range bets {
			// 3.1.1. find matching bet
			if openBet.BetReq.BetId == clearedOrder.BetID {
				value, ok := operatorsMap[openBet.OperatorId]
				if !ok {
					value, err = cache.GetOperatorDetails(openBet.OperatorId)
					operatorsMap[openBet.OperatorId] = value
				}
				delta, ok := operatorsDelta[openBet.OperatorId]
				if !ok {
					delta = 0
					operatorsDelta[openBet.OperatorId] = delta
				}
				// Add Rollback Request
				rollbackReq := providers.ComputeRollback(openBet, constants.BetFair.BetStatus.LAPSED())
				// Add OperatorLedgerTx
				operatorLedgerTx := providers.GetOperatorLedgerTx(value, constants.SAP.LedgerTxType.BETROLLBACK(), openBet.BetReq.OperatorAmount, openBet.BetId)
				operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
				// Add to Operator Balance Delta
				delta += rollbackReq.OperatorAmount
				operatorsDelta[openBet.OperatorId] = delta
				log.Println("BetFairLapsedOrders: BetId is : ", openBet.BetId)
				log.Println("BetFairLapsedOrders: rollbackReq.RollbackAmount is : ", rollbackReq.RollbackAmount)
				openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
				openBet.Status = constants.BetFair.BetStatus.LAPSED()
				// setting updatedAt
				openBet.UpdatedAt = rollbackReq.ReqTime
				lapsedBets = append(lapsedBets, openBet)
				break
			}
		}
	}
	// 4. Check for length, return if 0
	if len(lapsedBets) == 0 {
		//log.Println("BetFairLapsedOrders: ZERO bets to lapse.")
		return
	}
	// 5. Insert Operator Ledger Txs
	if len(operatorLedgerTxs) > 0 {
		err = database.InsertOperatorLedgers(operatorLedgerTxs)
		if err != nil {
			// inserting ledger documents failed
			log.Println("BetFairLapsedOrders: database.InsertOperatorLedgers failed with error - ", err.Error())
		}
	}
	// 6. Update Operator Balance balance and save
	for operatorId, delta := range operatorsDelta {
		if delta != 0 {
			err = database.UpdateOperatorBalance(operatorId, delta)
			if err != nil {
				// updating operator balance failed
				log.Println("BetFairLapsedOrders: database.UpdateOperatorBalance failed with error - ", err.Error())
			}
		}
	}
	// 5. Process Voided Bets
	log.Println("BetFairLapsedOrders: voidedBets count is - ", len(lapsedBets))
	//operator.BetFairRollbackRoutine(voidedBets, "Bet Voided", constants.BetFair.BetStatus.VOIDED())
	operator.CommonRollbackRoutine(constants.BetFair.BetStatus.LAPSED(), lapsedBets)
	return
}

func BetFairCancelledOrders(betIds []string, bets []sports.BetDto) {
	// 1. Make BetFair Cleared Orders call with CANCELLED betStatus
	clearedOrdersResp, err := ClearedOrders(constants.BetFair.BetStatus.CANCELLED(), betIds, bets)
	if err != nil {
		log.Println("BetFairCancelledOrders: ClearedOrders failed with error - ", err.Error())
		return
	}
	log.Println("BetFairCancelledOrders: Unmatched Bets count is - ", len(betIds))
	log.Println("BetFairCancelledOrders: Cleared Orders count is - ", len(clearedOrdersResp.ClearedOrders))
	// 2. Create an empty voided bet list
	cancelledBets := []sports.BetDto{}
	operatorsMap := make(map[string]operatordto.OperatorDTO)
	operatorsDelta := make(map[string]float64)
	operatorLedgerTxs := []models.OperatorLedgerDto{}
	// 3. Iterate through Cleared Orders
	for _, clearedOrder := range clearedOrdersResp.ClearedOrders {
		//log.Printf("BetStatus:  "+betStatus+" Single Order: %+v\n", clearedOrder)
		// 3.1. Iterate through bets
		for _, openBet := range bets {
			// 3.1.1. find matching bet
			if openBet.BetReq.BetId == clearedOrder.BetID {
				value, ok := operatorsMap[openBet.OperatorId]
				if !ok {
					value, err = cache.GetOperatorDetails(openBet.OperatorId)
					operatorsMap[openBet.OperatorId] = value
				}
				delta, ok := operatorsDelta[openBet.OperatorId]
				if !ok {
					delta = 0
					operatorsDelta[openBet.OperatorId] = delta
				}
				// Add Rollback Request
				rollbackReq := providers.ComputeRollback(openBet, constants.BetFair.BetStatus.CANCELLED())
				// Add OperatorLedgerTx
				operatorLedgerTx := providers.GetOperatorLedgerTx(value, constants.SAP.LedgerTxType.BETCANCEL(), openBet.BetReq.OperatorAmount, openBet.BetId)
				operatorLedgerTxs = append(operatorLedgerTxs, operatorLedgerTx)
				// Add to Operator Balance Delta
				delta += rollbackReq.OperatorAmount
				operatorsDelta[openBet.OperatorId] = delta
				log.Println("BetFairVoidedOrders: BetId is : ", openBet.BetId)
				log.Println("BetFairVoidedOrders: rollbackReq.RollbackAmount is : ", rollbackReq.RollbackAmount)
				openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
				openBet.Status = constants.BetFair.BetStatus.CANCELLED()
				// setting updatedAt
				openBet.UpdatedAt = rollbackReq.ReqTime
				cancelledBets = append(cancelledBets, openBet)
				break
			}
		}
	}
	// 4. Check for length, return if 0
	if len(cancelledBets) == 0 {
		//log.Println("BetFairCancelledOrders: ZERO bets to cancel.")
		return
	}
	// 5. Process Cancelled Bets
	log.Println("BetFairCancelledOrders: cancelledBets count is - ", len(cancelledBets))
	operator.BetFairRollbackRoutine(cancelledBets, "Bet Cancelled by User", constants.BetFair.BetStatus.CANCELLED())
	return
}
*/

// func GetBetFairMarketResults() error {
// 	log.Println("GetBetFairMarketResults: Start Time is: ", time.Now())
// 	// 1. Get BetFair Open Bets List - Only local bets
// 	// 2. Create Unique MarketIds List & Create BetFair Market Result Requests
// 	// 3. Get Markets from Database
// 	// 4. Call BetFair Market Results endpoint
// 	// 5. Create Market Map, to handle missing markets
// 	// 6. Add missing markets
// 	// 7. Create Markets map with MarketIds as Key, Update Markets with Results
// 	// 8. Save Updated Markets
// 	// 9. Save New Markets
// 	// 10. Create Settled bets
// 	// 11. Process Settled Bets

// 	// 1. Get BetFair Open Bets List
// 	bets1, err := database.GetBetsByStatus(constants.SAP.ProviderType.BetFair(), constants.SAP.BetStatus.OPEN())
// 	if err != nil {
// 		log.Println("GetBetFairMarketResults: database.GetBetsByStatus for OPEN bets failed with error - ", err.Error())
// 		return err
// 	}
// 	if len(bets1) == 0 {
// 		return nil
// 	}
// 	log.Println("GetBetFairMarketResults: Open bets1 Count is - ", len(bets1))
// 	// Look for only local bets and skip betfair bets
// 	bets2 := []sports.BetDto{}
// 	for _, bet := range bets1 {
// 		if bet.BetId != bet.BetReq.BetId {
// 			continue
// 		}
// 		bets2 = append(bets2, bet)
// 	}
// 	if len(bets2) == 0 {
// 		return nil
// 	}
// 	log.Println("GetBetFairMarketResults: Open bets2 Count is - ", len(bets2))
// 	// 2. Create Unique MarketIds List & Create BetFair Market Result Requests
// 	marketIdMap := make(map[string]bool)
// 	betMapByMarketId := make(map[string]sports.BetDto) // This helps to create a market
// 	marketIds := []string{}
// 	getMarketResultReqs := []betfair.GetMarketResultReq{}
// 	for _, bet := range bets2 {
// 		_, ok := marketIdMap[bet.MarketId]
// 		if ok {
// 			continue
// 		}
// 		marketIdMap[bet.MarketId] = true
// 		betMapByMarketId[bet.MarketId] = bet
// 		marketIds = append(marketIds, bet.MarketId)
// 		getMarketResultReq := betfair.GetMarketResultReq{
// 			EventId:  bet.EventId,
// 			MarketId: bet.MarketId,
// 			SportId:  strings.Split(bet.EventKey, "-")[1],
// 		}
// 		getMarketResultReqs = append(getMarketResultReqs, getMarketResultReq)
// 	}
// 	if len(getMarketResultReqs) == 0 {
// 		return nil
// 	}
// 	log.Println("GetBetFairMarketResults: Markets Count is - ", len(getMarketResultReqs))
// 	// 3. Get Markets from Database
// 	markets, err := database.GetMarketsByMarketIds(marketIds)
// 	if err != nil {
// 		log.Println("GetBetFairMarketResults: database.GetMarketsByMarketIds failed with error - ", err.Error())
// 		return err
// 	}
// 	// 4. Call BetFair Market Results endpoint
// 	resp, err := GetMarketResults(getMarketResultReqs)
// 	if err != nil {
// 		log.Println("GetBetFairMarketResults: BetFairMarketResults failed with error - ", err.Error())
// 		return err
// 	}
// 	if len(resp) == 0 {
// 		log.Println("GetBetFairMarketResults: GetMarketResults returned ZERO results")
// 		return fmt.Errorf("FAILED to FETCH Market Results!!!")
// 	}
// 	log.Println("GetBetFairMarketResults: Markets Results Count is - ", len(resp))
// 	// 5. Create Market Map, to handle missing markets
// 	oldMarketsMap := make(map[string]models.Market)
// 	for _, market := range markets {
// 		oldMarketsMap[market.MarketId] = market
// 	}
// 	// 6. Add missing markets
// 	newMarketsMap := make(map[string]models.Market)
// 	for _, marketId := range marketIds {
// 		_, ok := oldMarketsMap[marketId]
// 		if ok {
// 			continue
// 		}
// 		log.Println("GetBetFairMarketResults: NOT FOUND (Market) for MarketId - ", marketId)
// 		// Add missing market
// 		bet := betMapByMarketId[marketId]
// 		market := models.Market{}
// 		market.MarketKey = bet.EventKey + "-" + marketId
// 		market.EventKey = bet.EventKey
// 		market.ProviderId = constants.SAP.ProviderType.BetFair()
// 		market.ProviderName = constants.SAP.ProviderName.BetFairName()
// 		market.SportId = strings.Split(bet.EventKey, "-")[1]
// 		market.SportName = bet.BetDetails.SportName
// 		market.CompetitionId = "" //???
// 		market.CompetitionName = bet.BetDetails.CompetitionName
// 		market.EventId = bet.EventId
// 		market.EventName = bet.BetDetails.EventName
// 		market.MarketId = marketId
// 		market.MarketName = bet.BetDetails.MarketName
// 		market.MarketType = bet.BetDetails.MarketType
// 		market.Category = "" //??
// 		market.Runners = []models.Runner{}
// 		market.Status = constants.SAP.ObjectStatus.ACTIVE()
// 		market.CreatedAt = time.Now().Unix()
// 		market.UpdatedAt = market.CreatedAt
// 		market.Config = commondto.ConfigDto{}
// 		market.MarketStatus = "SETTLED"
// 		market.Results = []models.Result{}
// 		market.Rollbacks = []models.Rollback{}
// 		newMarketsMap[marketId] = market
// 	}
// 	log.Println("GetBetFairMarketResults: Markets Count found in database is - ", len(oldMarketsMap))
// 	log.Println("GetBetFairMarketResults: Markets Count not found in database is - ", len(newMarketsMap))
// 	// 7. Create Markets map with MarketIds as Key, Update Markets with Results
// 	marketsMap := make(map[string]models.Result)
// 	updatedMarkets := []models.Market{}
// 	newMarkets := []models.Market{}
// 	errorMarketsMap := make(map[string]bool)
// 	openMarketsMap := make(map[string]bool)
// 	nowinnerMarketMap := make(map[string]bool)
// 	for _, getMarketResultResp := range resp {
// 		if getMarketResultResp.Status != "RS_OK" {
// 			log.Println("GetBetFairMarketResults: BetFairMarketResults failed with error - ", getMarketResultResp.Message)
// 			errorMarketsMap[getMarketResultResp.MarketResult.MarketId] = true
// 			continue
// 		}
// 		if getMarketResultResp.MarketResult.MarketStatus != constants.BetFair.MarketStatus.CLOSED() {
// 			log.Println("GetBetFairMarketResults: MarketStaus for "+getMarketResultResp.MarketResult.MarketId+" is not CLOSED Yet. The current status is - ", getMarketResultResp.MarketResult.MarketStatus)
// 			openMarketsMap[getMarketResultResp.MarketResult.MarketId] = true
// 			continue
// 		}
// 		// Update existing market or add new market with result
// 		var winnerFound bool = false
// 		var isOld bool = true
// 		market, isFound := oldMarketsMap[getMarketResultResp.MarketResult.MarketId]
// 		if isFound == false {
// 			isOld = false // new Market
// 			market, isFound = newMarketsMap[getMarketResultResp.MarketResult.MarketId]
// 			if isFound == false {
// 				log.Println("GetBetFairMarketResults: UNEXPECTED STATE - Makret not found - ", getMarketResultResp.MarketResult.MarketId)
// 				continue
// 			}
// 			// add runners to new market
// 			for _, runner := range getMarketResultResp.MarketResult.Runners {
// 				newRunner := models.Runner{}
// 				newRunner.RunnerId = runner.RunnerId
// 				newRunner.RunnerName = "" //???
// 				newRunner.RunnerStatus = constants.SAP.ObjectStatus.ACTIVE()
// 				market.Runners = append(market.Runners, newRunner)
// 			}
// 		}
// 		// Create Result, add result to market & results map, add market to the respective list (Old or New)
// 		for _, runner := range getMarketResultResp.MarketResult.Runners {
// 			if runner.RunnerStatus == "WINNER" {
// 				winnerFound = true // market has winner, add result to the market
// 				market.MarketStatus = "SETTLED"
// 				// create result
// 				result := models.Result{}
// 				result.RunnerId = runner.RunnerId
// 				for _, runr := range market.Runners {
// 					if runr.RunnerId == runner.RunnerId {
// 						result.RunnerName = runr.RunnerName
// 					}
// 				}
// 				result.SessionOutcome = 0
// 				result.ResultTime = time.Now().Unix()
// 				// add result to the market
// 				market.Results = append(market.Results, result)
// 				if isOld == true {
// 					updatedMarkets = append(updatedMarkets, market)
// 				} else {
// 					newMarkets = append(newMarkets, market)
// 				}
// 				// add result to the results map
// 				marketsMap[getMarketResultResp.MarketResult.MarketId] = result
// 				break
// 			}
// 		}
// 		if winnerFound == false {
// 			log.Println("GetBetFairMarketResults: NOT FOUND (Winner) for MarketId - ", getMarketResultResp.MarketResult.MarketId)
// 			nowinnerMarketMap[getMarketResultResp.MarketResult.MarketId] = true
// 		}
// 	}
// 	if len(updatedMarkets) == 0 && len(newMarkets) == 0 {
// 		return nil
// 	}
// 	// 8. Save Updated Markets
// 	if len(updatedMarkets) > 0 {
// 		log.Println("GetBetFairMarketResults: updatedMarkets Count is - ", len(updatedMarkets))
// 		count, msgs := database.UpdateManyMarkets(updatedMarkets)
// 		if count != len(updatedMarkets) {
// 			log.Println("GetBetFairMarketResults: database.UpdateManyMarkets able update only makerts count is - ", count)
// 			log.Println("GetBetFairMarketResults: database.UpdateManyMarkets failed makerts count is - ", len(msgs))
// 			for _, msg := range msgs {
// 				log.Println("GetBetFairMarketResults: database.UpdateManyMarkets failuer msg is  - ", msg)
// 			}
// 		}
// 	}
// 	// 9. Save New Markets
// 	if len(newMarkets) > 0 {
// 		log.Println("GetBetFairMarketResults: newMarkets Count is - ", len(newMarkets))
// 		err = database.InsertManyMarkets(newMarkets)
// 		if err != nil {
// 			log.Println("GetBetFairMarketResults: database.InsertManyMarkets failed with error - ", err.Error())
// 		}
// 	}
// 	log.Println("GetBetFairMarketResults: marketsMap Count is - ", len(marketsMap))
// 	log.Println("GetBetFairMarketResults: errorMarketsMap Count is - ", len(errorMarketsMap))
// 	log.Println("GetBetFairMarketResults: openMarketsMap Count is - ", len(openMarketsMap))
// 	log.Println("GetBetFairMarketResults: nowinnerMarketMap Count is - ", len(nowinnerMarketMap))
// 	// 10. Create Settled bets
// 	settledBets := []sports.BetDto{}
// 	for _, bet := range bets2 {
// 		// Create result request
// 		resultReq := sports.ResultReqDto{}
// 		resultReq.ReqId = uuid.New().String()
// 		resultReq.ReqTime = time.Now().UnixNano() / int64(time.Millisecond)
// 		resultReq.CreditAmount = 0
// 		resultReq.OperatorAmount = 0
// 		resultReq.PlatformAmount = 0
// 		resultReq.SessionOutcome = 0
// 		// Check marketId is present in marketsMap
// 		result, ok := marketsMap[bet.MarketId]
// 		if ok == false {
// 			// missing result, settle as LOST
// 			_, ok = errorMarketsMap[bet.MarketId]
// 			if ok == true {
// 				log.Println("GetBetFairMarketResults: ErrorMarket for market & bet: " + bet.MarketId + "-" + bet.BetId)
// 				continue
// 			}
// 			_, ok = openMarketsMap[bet.MarketId]
// 			if ok == true {
// 				log.Println("GetBetFairMarketResults: OpenMarket for market & bet: " + bet.MarketId + "-" + bet.BetId)
// 				continue
// 			}
// 			_, ok = nowinnerMarketMap[bet.MarketId]
// 			if ok == true {
// 				log.Println("GetBetFairMarketResults: NoWinner for market & bet: " + bet.MarketId + "-" + bet.BetId)
// 				continue // settling bets as LOST
// 			} else {
// 				log.Println("GetBetFairMarketResults: unexpected state for market & bet: " + bet.MarketId + "-" + bet.BetId)
// 				continue // settling bets as LOST
// 			}
// 		} else {
// 			resultReq.RunnerName = result.RunnerName
// 			if bet.BetDetails.BetType == "BACK" && bet.BetDetails.RunnerId == result.RunnerId {
// 				resultReq.CreditAmount = utils.Truncate4Decfloat64(bet.BetReq.DebitAmount * bet.BetDetails.OddValue)
// 				resultReq.OperatorAmount = utils.Truncate4Decfloat64(bet.BetReq.OperatorAmount * bet.BetDetails.OddValue)
// 				resultReq.PlatformAmount = utils.Truncate4Decfloat64(bet.BetReq.PlatformAmount * bet.BetDetails.OddValue)
// 				resultReq.RunnerName = bet.BetDetails.RunnerName // for new markets, result doesnt have runner name
// 			}
// 			if bet.BetDetails.BetType == "LAY" && bet.BetDetails.RunnerId != result.RunnerId {
// 				resultReq.CreditAmount = utils.Truncate4Decfloat64(bet.BetReq.DebitAmount * bet.BetDetails.OddValue / (bet.BetDetails.OddValue - 1))
// 				resultReq.OperatorAmount = utils.Truncate4Decfloat64(bet.BetReq.OperatorAmount * bet.BetDetails.OddValue / (bet.BetDetails.OddValue - 1))
// 				resultReq.PlatformAmount = utils.Truncate4Decfloat64(bet.BetReq.PlatformAmount * bet.BetDetails.OddValue / (bet.BetDetails.OddValue - 1))
// 			}
// 		}
// 		bet.ResultReqs = append(bet.ResultReqs, resultReq)
// 		bet.NetAmount += resultReq.CreditAmount
// 		if bet.NetAmount > 0 {
// 			commConfig := providers.GetCommissionConfig(bet.UserId, bet.MarketId, bet.EventId, bet.CompetitionId, bet.SportId, bet.PartnerId, bet.OperatorId, bet.ProviderId)
// 			if bet.OperatorId == "kiaexch" || bet.OperatorId == "kia-sap" {
// 				bet.Commission = utils.Truncate4Decfloat64(bet.NetAmount * commConfig.CommPercentage / 100)
// 			}
// 			bet.CommLevel = commConfig.CommLevel
// 		}
// 		bet.Status = constants.BetFair.BetStatus.SETTLED()
// 		// setting updatedAt
// 		bet.UpdatedAt = resultReq.ReqTime
// 		settledBets = append(settledBets, bet)
// 	}
// 	if len(settledBets) == 0 {
// 		return nil
// 	}
// 	// 11. Process Settled Bets
// 	log.Println("GetBetFairMarketResults: settledBets count is - ", len(settledBets))
// 	operator.CommonResultRoutine(settledBets)
// 	log.Println("GetBetFairMarketResults: End Time is: ", time.Now())
// 	return nil
// }

func GetBetFairMarketResults() error {
	log.Println("GetBetFairMarketResults: START - ", time.Now())
	olderMarkets := time.Now().Add(-1 * 7 * 24 * time.Hour).UnixMilli()
	// 1. Get BetFair Open Markets List from DB & Create BetFair Market Result Requests
	// 2. Create MarketMap & Create BetFair Market Result Requests
	// 3. Call BetFair Market Results endpoint sort by created date ascending (oldest first)
	// 4. Create Markets map with MarketIds as Key, Update Markets with Results
	// 5. Save Updated Markets

	// 6. Get BetFair Open Bets List - Only local bets
	// 7. Create Settled bets
	// 8. Process Settled Bets

	// 1. Get BetFair Open Markets List from DB & Create BetFair Market Result Requests
	markets, err := database.GetOpenMarketsByProvider(constants.SAP.ProviderType.BetFair())
	if err != nil {
		log.Println("GetBetFairMarketResults: database.GetOpenMarketsByProvider for failed with error - ", err.Error())
		log.Println("GetBetFairMarketResults: END FAILED 1 - ", time.Now())
		return err
	}
	if len(markets) == 0 {
		log.Println("GetBetFairMarketResults: END SUCCESS 2 - ", time.Now())
		return nil
	}
	log.Println("GetBetFairMarketResults: open markets Count is - ", len(markets))
	// 2. Create MarketMap & Create BetFair Market Result Requests
	marketMap := make(map[string]models.Market)
	getMarketResultReqs := []betfair.GetMarketResultReq{}
	for _, market := range markets {
		_, ok := marketMap[market.MarketId]
		if ok {
			continue
		}
		marketMap[market.MarketId] = market
		getMarketResultReq := betfair.GetMarketResultReq{
			EventId:  market.EventId,
			MarketId: market.MarketId,
			SportId:  market.SportId,
		}
		getMarketResultReqs = append(getMarketResultReqs, getMarketResultReq)
	}
	if len(getMarketResultReqs) == 0 {
		log.Println("GetBetFairMarketResults: END SUCCESS 3 - ", time.Now())
		return nil
	}
	log.Println("GetBetFairMarketResults: getMarketResultReqs Count is - ", len(getMarketResultReqs))
	// 3. Call BetFair Market Results endpoint sort by created date ascending (oldest first)
	resp, err := GetMarketResults(getMarketResultReqs)
	if err != nil {
		log.Println("GetBetFairMarketResults: BetFairMarketResults failed with error - ", err.Error())
		log.Println("GetBetFairMarketResults: END FAILED 4 - ", time.Now())
		return err
	}
	if len(resp) == 0 {
		log.Println("GetBetFairMarketResults: GetMarketResults returned ZERO results")
		log.Println("GetBetFairMarketResults: END FAILED 5 - ", time.Now())
		return fmt.Errorf("FAILED to FETCH Market Results!!!")
	}
	log.Println("GetBetFairMarketResults: Markets Results Count is - ", len(resp))
	// 4. Update Markets with Results
	eventKeyMap := make(map[string]bool)
	eventKeys := []string{}
	marketIds := []string{}
	voidEventKeyMap := make(map[string]bool)
	voidEventKeys := []string{}
	voidMarketIds := []string{}
	voidMarkets := []models.Market{}
	notfoundEventKeyMap := make(map[string]bool)
	notfoundEventKeys := []string{}
	notfoundMarketIds := []string{}
	notfoundMarkets := []models.Market{}

	marketsMap := make(map[string]models.Result)
	settledMarkets := []models.Market{}
	// newMarkets := []models.Market{}
	errorMarketsMap := make(map[string]bool)
	openMarketsMap := make(map[string]bool)
	nowinnerMarketMap := make(map[string]bool)
	for _, getMarketResultResp := range resp {
		// Update existing market or add new market with result
		var winnerFound bool = false
		market, isFound := marketMap[getMarketResultResp.MarketResult.MarketId]
		if isFound == false {
			log.Println("GetBetFairMarketResults: UNEXPECTED Market NOT FOUND in marketMap for marketId - ", getMarketResultResp.MarketResult.MarketId)
			continue
		}
		if getMarketResultResp.Status != "RS_OK" {
			log.Println("GetBetFairMarketResults: BetFairMarketResults failed with error - ", getMarketResultResp.Message)
			errorMarketsMap[getMarketResultResp.MarketResult.MarketId] = true
			if market.CreatedAt < olderMarkets {
				_, ok := notfoundEventKeyMap[market.EventKey]
				if ok == false {
					notfoundEventKeyMap[market.EventKey] = true
					notfoundEventKeys = append(notfoundEventKeys, market.EventKey)
				}
				notfoundMarketIds = append(notfoundMarketIds, market.MarketId)
				market.MarketStatus = "NOTFOUND"
				notfoundMarkets = append(notfoundMarkets, market)
			}
			continue
		}
		if getMarketResultResp.MarketResult.MarketStatus != constants.BetFair.MarketStatus.CLOSED() {
			log.Println("GetBetFairMarketResults: MarketStaus for "+getMarketResultResp.MarketResult.MarketId+" is not CLOSED Yet. The current status is - ", getMarketResultResp.MarketResult.MarketStatus)
			openMarketsMap[getMarketResultResp.MarketResult.MarketId] = true
			continue
		}
		// Create Result, add result to market & results map, add market to the respective list (Old or New)
		var removedCount int = 0
		for _, runner := range getMarketResultResp.MarketResult.Runners {
			if runner.RunnerStatus == "WINNER" {
				winnerFound = true // market has winner, add result to the market
				market.MarketStatus = "SETTLED"
				marketIds = append(marketIds, market.MarketId)
				// create result
				result := models.Result{}
				result.RunnerId = runner.RunnerId
				for _, runr := range market.Runners {
					if runr.RunnerId == runner.RunnerId {
						result.RunnerName = runr.RunnerName
						break
					}
				}
				result.SessionOutcome = 0
				result.ResultTime = time.Now().Unix()
				// add result to the market
				market.Results = append(market.Results, result)
				settledMarkets = append(settledMarkets, market)
				// add result to the results map
				marketsMap[getMarketResultResp.MarketResult.MarketId] = result
				_, ok := eventKeyMap[market.EventKey]
				if ok == false {
					eventKeyMap[market.EventKey] = true
					eventKeys = append(eventKeys, market.EventKey)
				}
				break
			} else if runner.RunnerStatus == "REMOVED" {
				removedCount++
			}
		}
		if winnerFound == false {
			// mrrJson, err := json.Marshal(getMarketResultResp)
			// if err != nil {
			// 	log.Println("GetBetFairMarketResults: NOT FOUND (Winner) for MarketId - ", getMarketResultResp.MarketResult.MarketId, err.Error())
			// } else {
			// 	log.Println("GetBetFairMarketResults: NOT FOUND (Winner) for MarketId - ", getMarketResultResp.MarketResult.MarketId, string(mrrJson))
			// }
			if removedCount == len(getMarketResultResp.MarketResult.Runners) {
				log.Println("GetBetFairMarketResults: NOT FOUND (Winner) for MarketId ALL Runner were REMOVED - ", getMarketResultResp.MarketResult.MarketId, removedCount)
				_, ok := voidEventKeyMap[market.EventKey]
				if ok == false {
					voidEventKeyMap[market.EventKey] = true
					voidEventKeys = append(voidEventKeys, market.EventKey)
				}
				voidMarketIds = append(voidMarketIds, market.MarketId)
				market.MarketStatus = "VOIDED"
				voidMarkets = append(voidMarkets, market)
			}
			nowinnerMarketMap[getMarketResultResp.MarketResult.MarketId] = true
		}
	}
	log.Println("GetBetFairMarketResults: errorMarketsMap Count is - ", len(errorMarketsMap))
	log.Println("GetBetFairMarketResults: openMarketsMap Count is - ", len(openMarketsMap))
	log.Println("GetBetFairMarketResults: nowinnerMarketMap Count is - ", len(nowinnerMarketMap))
	log.Println("GetBetFairMarketResults: settledMarkets Count is - ", len(settledMarkets))
	if len(settledMarkets) > 0 {
		// 5. Save Updated Markets
		count, msgs := database.UpdateManyMarkets(settledMarkets)
		if count != len(settledMarkets) {
			log.Println("GetBetFairMarketResults: database.UpdateManyMarkets success and failure counts are - ", count, len(msgs))
			for _, msg := range msgs {
				log.Println("GetBetFairMarketResults: database.UpdateManyMarkets failuer msg is  - ", msg)
			}
		}
		// 5.x Send Market Results to Operators
		go SendMarketResults(settledMarkets)
		// 6. Get BetFair Open Bets List - Only local bets
		bets1, err := database.GetBetsByMarkets(eventKeys, marketIds, "OPEN")
		if err != nil {
			log.Println("GetBetFairMarketResults: database.GetBetsByMarkets failed with error - ", err.Error())
			log.Println("GetBetFairMarketResults: END FAILED 6 - ", time.Now())
			return err
		}
		if len(bets1) > 0 {
			log.Println("GetBetFairMarketResults: Open bets1 Count is - ", len(bets1))
			// Look for only local bets and skip betfair bets
			bets2 := []sports.BetDto{}
			for _, bet := range bets1 {
				if bet.BetId != bet.BetReq.BetId {
					continue
				}
				bets2 = append(bets2, bet)
			}
			if len(bets2) > 0 {
				log.Println("GetBetFairMarketResults: Open bets2 Count is - ", len(bets2))
				// 7. Create Settled bets
				settledBets := []sports.BetDto{}
				for _, bet := range bets2 {
					// Create result request
					resultReq := sports.ResultReqDto{}
					resultReq.ReqId = uuid.New().String()
					resultReq.ReqTime = time.Now().UnixMilli()
					resultReq.CreditAmount = 0
					resultReq.OperatorAmount = 0
					resultReq.PlatformAmount = 0
					resultReq.SessionOutcome = 0
					// Check marketId is present in marketsMap
					result, ok := marketsMap[bet.MarketId]
					if ok == false {
						// missing result, SKIP
						log.Println("GetBetFairMarketResults: MISSING RESULT for market & bet: " + bet.MarketId + "-" + bet.BetId)
						continue
					}
					resultReq.RunnerName = result.RunnerName
					oddValue := bet.BetDetails.OddValue
					if bet.BetReq.OddsMatched > 0 {
						oddValue = bet.BetReq.OddsMatched
					}
					if bet.BetDetails.BetType == "BACK" && bet.BetDetails.RunnerId == result.RunnerId { // BACK & WON case
						resultReq.CreditAmount = utils.Truncate4Decfloat64(bet.BetReq.DebitAmount * oddValue)
						resultReq.OperatorAmount = utils.Truncate4Decfloat64(bet.BetReq.OperatorAmount * oddValue)
						resultReq.PlatformAmount = utils.Truncate4Decfloat64(bet.BetReq.PlatformAmount * oddValue)
						resultReq.RunnerName = bet.BetDetails.RunnerName // for new markets, result doesnt have runner name
					}
					if bet.BetDetails.BetType == "LAY" && bet.BetDetails.RunnerId != result.RunnerId { // LAY & WON case
						resultReq.CreditAmount = utils.Truncate4Decfloat64(bet.BetReq.DebitAmount * oddValue / (oddValue - 1))
						resultReq.OperatorAmount = utils.Truncate4Decfloat64(bet.BetReq.OperatorAmount * oddValue / (oddValue - 1))
						resultReq.PlatformAmount = utils.Truncate4Decfloat64(bet.BetReq.PlatformAmount * oddValue / (oddValue - 1))
					}
					bet.ResultReqs = append(bet.ResultReqs, resultReq)
					bet.NetAmount += resultReq.CreditAmount
					if bet.NetAmount > 0 {
						commConfig := providers.GetCommissionConfig(bet.UserId, bet.MarketId, bet.EventId, bet.CompetitionId, bet.SportId, bet.PartnerId, bet.OperatorId, bet.ProviderId)
						if bet.OperatorId != "kiaexch" && bet.OperatorId != "kia-sap" {
							bet.Commission = utils.Truncate4Decfloat64(bet.NetAmount * commConfig.CommPercentage / 100)
						}
						bet.CommLevel = commConfig.CommLevel
					}
					bet.Status = constants.BetFair.BetStatus.SETTLED()
					// setting updatedAt
					bet.UpdatedAt = resultReq.ReqTime
					settledBets = append(settledBets, bet)
				}
				if len(settledBets) > 0 {
					log.Println("GetBetFairMarketResults: settledBets count is - ", len(settledBets))
					// 8. Process Settled Bets
					operator.CommonResultRoutine(settledBets)
				}
			}
		}
	}
	// log.Println("GetBetFairMarketResults: operator.CommonResultRoutine ended at: ", time.Now())
	log.Println("GetBetFairMarketResults: voidEventKeys Count is - ", len(voidEventKeys))
	log.Println("GetBetFairMarketResults: voidMarketIds Count is - ", len(voidMarketIds))
	log.Println("GetBetFairMarketResults: voidMarkets Count is - ", len(voidMarkets))
	if len(voidMarkets) > 0 {
		// 9. Save Updated Markets
		count21, msgs21 := database.UpdateManyMarkets(voidMarkets)
		if count21 != len(voidMarkets) {
			log.Println("GetBetFairMarketResults: database.UpdateManyMarkets success and failure counts are - ", count21, len(msgs21))
			for _, msg21 := range msgs21 {
				log.Println("GetBetFairMarketResults: database.UpdateManyMarkets failuer msg21 is  - ", msg21)
			}
		}
		// 9.x Send Market Results to Operators
		go SendMarketResults(voidMarkets)
		// 10. Get BetFair Open Bets List - Only local bets
		bets21, err := database.GetBetsByMarkets(voidEventKeys, voidMarketIds, "OPEN")
		if err != nil {
			log.Println("GetBetFairMarketResults: database.GetBetsByMarkets failed with error - ", err.Error())
			log.Println("GetBetFairMarketResults: END FAILED 7 - ", time.Now())
			return err
		}
		if len(bets21) > 0 {
			log.Println("GetBetFairMarketResults: Open bets21 Count is - ", len(bets21))
			// Look for only local bets and skip betfair bets
			bets22 := []sports.BetDto{}
			for _, bet := range bets21 {
				if bet.BetId != bet.BetReq.BetId {
					continue
				}
				bets22 = append(bets22, bet)
			}
			if len(bets22) > 0 {
				log.Println("GetBetFairMarketResults: Open bets22 Count is - ", len(bets22))
				// 11. Create Settled bets
				voidBets := []sports.BetDto{}
				for _, bet := range bets22 {
					// Create rollback request
					rollbackReq := sports.RollbackReqDto{}
					rollbackReq.ReqId = uuid.New().String()
					rollbackReq.ReqTime = time.Now().UnixMilli()
					rollbackReq.RollbackAmount = bet.BetReq.DebitAmount
					rollbackReq.OperatorAmount = bet.BetReq.OperatorAmount
					rollbackReq.PlatformAmount = bet.BetReq.PlatformAmount
					rollbackReq.RollbackReason = "VOIDED"
					bet.RollbackReqs = append(bet.RollbackReqs, rollbackReq)
					bet.NetAmount = 0
					bet.Status = constants.BetFair.BetStatus.VOIDED()
					// setting updatedAt
					bet.UpdatedAt = rollbackReq.ReqTime
					voidBets = append(voidBets, bet)
				}
				if len(voidBets) > 0 {
					log.Println("GetBetFairMarketResults: voidBets count is - ", len(voidBets))
					// 12. Process Voided Bets
					operator.CommonRollbackRoutine("VOIDED", voidBets)
				}
			}
		}
	}
	log.Println("GetBetFairMarketResults: notfoundEventKeys Count is - ", len(notfoundEventKeys))
	log.Println("GetBetFairMarketResults: notfoundMarketIds Count is - ", len(notfoundMarketIds))
	log.Println("GetBetFairMarketResults: notfoundMarkets Count is - ", len(notfoundMarkets))
	if len(notfoundMarkets) > 0 {
		// 9. Save Updated Markets
		count31, msgs31 := database.UpdateManyMarkets(notfoundMarkets)
		if count31 != len(notfoundMarkets) {
			log.Println("GetBetFairMarketResults: database.UpdateManyMarkets success and failure counts are - ", count31, len(msgs31))
			for _, msg31 := range msgs31 {
				log.Println("GetBetFairMarketResults: database.UpdateManyMarkets failuer msg31 is  - ", msg31)
			}
		}
		// 9.x Send Market Results to Operators
		go SendMarketResults(notfoundMarkets)
		// 10. Get BetFair Open Bets List - Only local bets
		bets31, err := database.GetBetsByMarkets(notfoundEventKeys, notfoundMarketIds, "OPEN")
		if err != nil {
			log.Println("GetBetFairMarketResults: database.GetBetsByMarkets failed with error - ", err.Error())
			log.Println("GetBetFairMarketResults: END FAILED 8 - ", time.Now())
			return err
		}
		if len(bets31) > 0 {
			log.Println("GetBetFairMarketResults: Open bets31 Count is - ", len(bets31))
			// Look for only local bets and skip betfair bets
			bets32 := []sports.BetDto{}
			for _, bet := range bets31 {
				if bet.BetId != bet.BetReq.BetId {
					continue
				}
				bets32 = append(bets32, bet)
			}
			if len(bets32) > 0 {
				log.Println("GetBetFairMarketResults: Open bets32 Count is - ", len(bets32))
				// 11. Create Settled bets
				notfoundBets := []sports.BetDto{}
				for _, bet := range bets32 {
					// Create rollback request
					rollbackReq := sports.RollbackReqDto{}
					rollbackReq.ReqId = uuid.New().String()
					rollbackReq.ReqTime = time.Now().UnixMilli()
					rollbackReq.RollbackAmount = bet.BetReq.DebitAmount
					rollbackReq.OperatorAmount = bet.BetReq.OperatorAmount
					rollbackReq.PlatformAmount = bet.BetReq.PlatformAmount
					rollbackReq.RollbackReason = "VOIDED"
					bet.RollbackReqs = append(bet.RollbackReqs, rollbackReq)
					bet.NetAmount = 0
					bet.Status = constants.BetFair.BetStatus.VOIDED()
					// setting updatedAt
					bet.UpdatedAt = rollbackReq.ReqTime
					notfoundBets = append(notfoundBets, bet)
				}
				if len(notfoundBets) > 0 {
					log.Println("GetBetFairMarketResults: notfoundBets count is - ", len(notfoundBets))
					// 12. Process notfoundBets Bets
					operator.CommonRollbackRoutine("VOIDED", notfoundBets)
				}
			}
		}
	}
	log.Println("GetBetFairMarketResults: END SUCCESS 9 - ", time.Now())
	return nil
}

func SendMarketResults(markets []models.Market) {
	log.Println("SendMarketResults: START for markets count - ", time.Now(), len(markets))
	for _, market := range markets {
		err := fmt.Errorf("")
		switch market.MarketStatus {
		case "SETTLED":
			result := market.Results[len(market.Results)-1]
			err = operator.MarketResult(constants.SAP.ProviderType.BetFair(), market.SportId, market.EventId, market.MarketId, market.MarketType, result.RunnerId, result.RunnerName, 0)
		case "VOIDED":
			err = operator.MarketRollback(constants.SAP.ProviderType.BetFair(), market.SportId, market.EventId, market.MarketId, market.MarketType, market.MarketName, "VOIDED", "Market Voided by BetFair")
		case "NOTFOUND":
			err = operator.MarketRollback(constants.SAP.ProviderType.BetFair(), market.SportId, market.EventId, market.MarketId, market.MarketType, market.MarketName, "VOIDED", "Market NOT FOUND in BetFair")
		default:
			log.Println("SendMarketResults: unexpected marketstatus for market - ", market.MarketKey, market.MarketStatus)
			continue
		}
		if err != nil {
			log.Println("SendMarketResults: operator.MarketResult failed with error for marketKey - ", market.MarketKey, market.MarketStatus, err.Error())
		}
	}
	log.Println("SendMarketResults: END SUCCESS - ", time.Now())
}

// func UpdateEventsforComp() {

// 	sports := []string{"1", "2", "4"}
// 	allEvents := []dto.EventDto{}
// 	EventIds := []string{}
// 	for _, sport := range sports {
// 		event, err := GetEvents(sport)
// 		if err != nil {
// 			log.Println("UpdateEventsforComp: GetEvents failed with error - ", err.Error())
// 		}
// 		allEvents = append(allEvents, event...)
// 	}
// 	for _, event := range allEvents {
// 		EventIds = append(EventIds, event.EventId)
// 	}

// 	// Get Events from DB
// 	dbEvents, err := database.GetEventsByEventIds(EventIds)
// 	if err != nil {
// 		log.Println("UpdateEventsforComp: GetEventsByEventIds failed with error - ", err.Error())
// 	}
// 	emptyCompEvents := []models.Event{}
// 	for _, event := range dbEvents {
// 		if event.CompetitionId == "" || event.CompetitionId == "-1" {
// 			emptyCompEvents = append(emptyCompEvents, event)
// 		}
// 	}

// 	// Update competitionId for all events
// 	updateEvents := []models.Event{}
// 	eventKeys := []string{}
// 	eventIdsMap := make(map[string]string)
// 	eventNameMap := make(map[string]string)
// 	OperatorMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
// 	if err != nil {
// 		// log
// 	}
// 	for _, event := range allEvents {
// 		if event.CompetitionId == "" || event.CompetitionId == "-1" {
// 			continue
// 		}
// 		for _, emptyCompEvent := range emptyCompEvents {
// 			if event.EventId == emptyCompEvent.EventId {
// 				emptyCompEvent.CompetitionId = event.CompetitionId
// 				emptyCompEvent.CompetitionName = event.CompetitionName
// 				updateEvents = append(updateEvents, emptyCompEvent)
// 				eventIdsMap[event.EventId] = event.CompetitionId
// 				eventNameMap[event.EventId] = event.CompetitionName
// 				for _, operatorObj := range OperatorMap {
// 					operator := operatorObj.(operatordto.OperatorDTO)
// 					eventKey := operator.OperatorId + "-" + event.ProviderId + "-" + event.SportId + "-" + event.EventId
// 					eventKeys = append(eventKeys, eventKey)
// 				}
// 				break
// 			}
// 		}

// 	}
// 	log.Println("UpdateEventsforComp: updateEvents count is - ", len(updateEvents))
// 	if len(updateEvents) > 0 {
// 		err = database.UpdateEventDtos(updateEvents)
// 		if err != nil {
// 			log.Println("UpdateEventsforComp: database.UpdateEvents failed with error - ", err.Error())
// 		}
// 	}
// 	// // Update CompetitionIds for all EventStatus
// 	// eventStatus, err := database.GetUpdatedEventStatus(eventKeys)
// 	// if err != nil {
// 	// 	log.Println("UpdateEventsforComp: database.GetEventStatusByCompetitionIds failed with error - ", err.Error())
// 	// 	return
// 	// }
// 	// log.Println("UpdateEventsforComp: eventStatus count is - ", len(eventStatus))
// 	// updateEventStatus := []models.EventStatus{}
// 	// for _, event := range eventStatus {
// 	// 	event.CompetitionId = eventIdsMap[event.EventId]
// 	// 	event.CompetitionName = eventNameMap[event.EventId]
// 	// 	updateEventStatus = append(updateEventStatus, event)
// 	// }
// 	// log.Println("UpdateEventsforComp: updateEventStatus count is - ", len(updateEventStatus))
// 	// if len(updateEventStatus) > 0 {
// 	// 	err = database.UpdateEventStatusDtos(updateEventStatus)
// 	// 	if err != nil {
// 	// 		log.Println("UpdateEventsforComp: database.UpdateEventStatus failed with error - ", err.Error())
// 	// 	}
// 	// }
// 	log.Println("UpdateEventsforComp: End Time is: ", time.Now())
// }

func UpdateEventByCompetition() {
	log.Println("UpdateEventByCompetition: START - ", time.Now())
	// 1. Get BetFair Sports from Database
	sports, err := database.GetSports("BetFair")
	if err != nil {
		log.Println("UpdateEventByCompetition: GetSports failed with error - ", err.Error())
		log.Println("UpdateEventByCompetition: END FAILED 1 - ", time.Now())
		return
	}
	if len(sports) == 0 {
		log.Println("UpdateEventByCompetition: END FAILED 2 - ", time.Now())
		return
	}
	log.Println("UpdateEventByCompetition: sports Count is - ", len(sports))
	allEvents := []dto.EventDto{}
	// 2. Get BetFair Events from L2/L1
	for _, sport := range sports {
		events, err := GetEvents(sport.SportId)
		if err != nil {
			log.Println("UpdateEventByCompetition: GetEvents for BetFair failed with error - ", err.Error())
			continue
		}
		switch sport.SportId {
		case "1", "2", "4":
			log.Println("UpdateEventByCompetition: events Count for sport is - ", sport.SportId, len(events))
		default:
		}
		allEvents = append(allEvents, events...)
	}
	if len(allEvents) == 0 {
		log.Println("UpdateEventByCompetition: END FAILED 3 ZERO Events from L2/L1 - ", time.Now())
		return
	}
	log.Println("UpdateEventByCompetition: allEvents Count is - ", len(allEvents))
	// 3. Preapre EventKeys List which has valid competitionId
	eventKeys := []string{}
	events := []dto.EventDto{}
	for _, event := range allEvents {
		if event.CompetitionId != "-1" && event.CompetitionId != "" {
			eventKey := constants.SAP.ProviderType.BetFair() + "-" + event.SportId + "-" + event.EventId
			eventKeys = append(eventKeys, eventKey)
			events = append(events, event)
		}
	}
	if len(eventKeys) == 0 {
		log.Println("UpdateEventByCompetition: END FAILED 4 ZERO Events with valid competitionId from L2/L1 - ", time.Now())
		return
	}
	log.Println("UpdateEventByCompetition: eventKeys Count is - ", len(eventKeys))
	// 4. Get Events from DB for eventKeys and competitionId is -1
	dbEvents, err := database.GetEventsByEventKeys(eventKeys, "-1")
	if err != nil {
		log.Println("UpdateEventByCompetition: GetEventsByCompetitionId failed with error - ", err.Error())
		log.Println("UpdateEventByCompetition: END FAILED 5 - ", time.Now())
		return
	}
	if len(dbEvents) == 0 {
		log.Println("UpdateEventByCompetition: END SUCCESS 6 ZERO Events with -1 from DB - ", time.Now())
		return
	}
	log.Println("UpdateEventByCompetition: dbEvents Count is - ", len(dbEvents))
	// 5. Extract DbEvents with CompetitionId -1
	updatedEvents := []models.Event{}
	for _, dbEvent := range dbEvents {
		for _, event := range events {
			if event.EventId == dbEvent.EventId {
				dbEvent.CompetitionId = event.CompetitionId
				dbEvent.CompetitionName = event.CompetitionName
				updatedEvents = append(updatedEvents, dbEvent)
				log.Println("UpdateEventByCompetition: Corrected CompetitionId for event - ", event.SportId, event.EventId, event.CompetitionId)
				break
			}
		}
	}
	if len(updatedEvents) == 0 {
		log.Println("UpdateEventByCompetition: END FAILED 7 - ", time.Now())
		return
	}
	log.Println("UpdateEventByCompetition: updatedEvents Count is - ", len(updatedEvents))
	err = database.UpdateEventDtos(updatedEvents)
	if err != nil {
		log.Println("UpdateEventByCompetition: END FAILED 8 database.UpdateEvents failed with error - ", time.Now(), err.Error())
		return
	}
	log.Println("UpdateEventByCompetition: END SUCCESS 9 - ", time.Now())
	return
}
