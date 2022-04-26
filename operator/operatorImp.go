package operator

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	sportsdto "Sp/dto/sports"
	utils "Sp/utilities"
	"fmt"
	"log"
	"strconv"
	"time"
)

var (
	ProviderId string = "Dream"
)

type OperatorResult struct {
	OperatorId  string
	TotalBets   int
	SettledBets int
	FailedBets  int
	ErrorsList  []string
}

/*
func VoidBetsRoutine(result sportsdto.ResultQueueDto, voidBets []sportsdto.BetDto) {
	log.Println("VoidBetsRoutine: New Rollback Routine Started")
	// 2. Create Result Req for each open bet, append to the results and create operator's open bets map
	// 2.1. Create a new map
	operatorBetsMap := make(map[string][]sportsdto.BetDto)
	// 2.2. Iterate through open bets list
	for _, openBet := range voidBets {
		// 2.2.1. update bet with result
		rollbackReq := ComputeRollback(openBet, result)
		openBet.RollbackReqs = append(openBet.RollbackReqs, rollbackReq)
		// setting updatedAt
		openBet.UpdatedAt = rollbackReq.ReqTime
		// 2.2.2. update operator's map
		operatorBets, ok := operatorBetsMap[openBet.OperatorId]
		if !ok {
			// 2.2.2.1. if not present, create empty bets list
			operatorBets = []sportsdto.BetDto{}
		}
		// 2.2.3. append bet to list
		operatorBets = append(operatorBets, openBet)
		// 2.2.4. update map
		operatorBetsMap[openBet.OperatorId] = operatorBets
	}
	// 3. Update Result & ResultQueue Status to in-progress
	result.Status = "in-progress"
	updareRQ, err := database.UpdateResultsQueueStatus(result.ID, result.Status)
	if err != nil {
		log.Println("VoidBetsRoutine: Update RQ failed with error - : ", err.Error())
	} else {
		log.Println("VoidBetsRoutine: Update RQ result - modified count is - ", updareRQ.ModifiedCount)
	}

	// 3. create one thread per operatorId
	voidChannel := make(chan OperatorResult)
	for operatorId := range operatorBetsMap {
		// 3.1. create thread and pass list of rollbackBets to it
		go DoWalletRollbackRequests(result.ResultType, operatorId, operatorBetsMap[operatorId], voidChannel)
	}
	// TODO: Below itesm. Probably need to use channel to post the status
	// 4. wait for all threads to complete
	log.Println("VoidBetsRoutine: Waiting for all Settlements to complete")
	//wgWalletResults.Wait()
	failedOperators := 0
	opResults := make([]OperatorResult, len(operatorBetsMap))
	for i, _ := range opResults {
		log.Println("VoidBetsRoutine: Waiting on channel for operator - ", i+1)
		opResults[i] = <-voidChannel
		if opResults[i].FailedBets != 0 {
			msg := "Total failed bet settlements for Operator (" + opResults[i].OperatorId + ") are - " + string(opResults[1].FailedBets)
			log.Println("VoidBetsRoutine: Operators count on Market - ", msg)
			failedOperators++
		}
	}
	result.Status = "completed"
	if failedOperators != 0 {
		// TODO: Overall operation is failed
		log.Println("VoidBetsRoutine: Failed Operators count is - ", failedOperators)
		result.Status = "failed"
	}
	updareRQ, err = database.UpdateResultsQueueStatus(result.ID, result.Status)
	if err != nil {
		log.Println("VoidBetsRoutine: Update RQ failed with error - : ", err.Error())
	} else {
		log.Println("VoidBetsRoutine: Update RQ result - modified count is - ", updareRQ.ModifiedCount)
	}
	// 5. Make wallet Result call
	// 6. Handle Error / retry mechanism
	// 7. Update each Bet Document
	log.Println("VoidBetsRoutine: New Settlement Routine Ended")
}
*/
/*
func VoidBetRoutine(result sportsdto.ResultQueueDto, voidBet sportsdto.BetDto) {
	log.Println("VoidBetRoutine: New Rollback Routine Started")
	// 2. Create Result Req for each open bet, append to the results and create operator's open bets map
	// 2.2.1. update bet with result
	rollbackReq := ComputeRollback(voidBet, result)
	voidBet.RollbackReqs = append(voidBet.RollbackReqs, rollbackReq)
	// setting updatedAt
	voidBet.UpdatedAt = rollbackReq.ReqTime
	// 2.1. Create a new map
	operatorBetsMap := make(map[string][]sportsdto.BetDto)
	operatorBets := []sportsdto.BetDto{}
	// 2.2.3. append bet to list
	operatorBets = append(operatorBets, voidBet)
	// 2.2.4. update map
	operatorBetsMap[voidBet.OperatorId] = operatorBets
	// 3. create one thread per operatorId
	voidBetChannel := make(chan OperatorResult)
	go DoWalletRollbackRequests(result.ResultType, voidBet.OperatorId, operatorBetsMap[voidBet.OperatorId], voidBetChannel)
	// 4. wait for all threads to complete
	log.Println("VoidBetRoutine: Waiting for all Settlements to complete")
	opResult := <-voidBetChannel
	if opResult.FailedBets != 0 {
		log.Println("VoidBetRoutine: Failed to void bet - ", voidBet.BetId)
		result.Status = "failed"
	}
	// 5. Make wallet Result call
	// 6. Handle Error / retry mechanism
	// 7. Update each Bet Document
	log.Println("VoidBetRoutine: New Settlement Routine Ended")
}
*/
/*
func ComputeResult(openBet sportsdto.BetDto, result sportsdto.ResultQueueDto) (sportsdto.ResultReqDto, error) {
	resultReq := sportsdto.ResultReqDto{}
	resultReq.ReqId = uuid.New().String()
	resultReq.ReqTime = time.Now().UnixNano() / int64(time.Millisecond)
	resultReq.CreditAmount = 0
	resultReq.RunnerName = result.RunnerName
	resultReq.SessionOutcome = result.SessionOutcome
	switch strings.ToUpper(openBet.BetDetails.MarketType) {
	case "MATCH_ODDS", "BOOKMAKER":
		if openBet.BetDetails.BetType == "BACK" && openBet.BetDetails.RunnerId == result.RunnerId {
			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
		}
		if openBet.BetDetails.BetType == "LAY" && openBet.BetDetails.RunnerId != result.RunnerId {
			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
		}
	case "FANCY":
		if openBet.BetDetails.BetType == "BACK" && result.SessionOutcome >= openBet.BetDetails.SessionOutcome {
			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
		}
		if openBet.BetDetails.BetType == "LAY" && result.SessionOutcome < openBet.BetDetails.SessionOutcome {
			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
		}
	default:
		log.Println("ComputeResult: Unexpected MarketType - ", openBet.BetDetails.MarketType)
		return resultReq, fmt.Errorf("Invalid Market Type - " + openBet.BetDetails.MarketType)
	}
	resultReq.CreditAmount = utils.Truncate64(resultReq.CreditAmount)
	return resultReq, nil
}
*/
/*
func ComputeRollback(openBet sportsdto.BetDto, result sportsdto.ResultQueueDto) sportsdto.RollbackReqDto {
	rollbackReq := sportsdto.RollbackReqDto{}
	rollbackReq.ReqId = uuid.New().String()
	rollbackReq.ReqTime = time.Now().UnixNano() / int64(time.Millisecond)
	rollbackReq.RollbackReason = result.Reason
	rollbackReq.RollbackAmount = 0 // positive means, deposit to user, negative means, deduct from user
	for _, result := range openBet.ResultReqs {
		rollbackReq.RollbackAmount -= result.CreditAmount
	}
	for _, rollback := range openBet.RollbackReqs {
		rollbackReq.RollbackAmount -= rollback.RollbackAmount
	}
	switch strings.ToLower(result.ResultType) {
	case "void", "voided", "cancelled", "deleted", "lapsed", "expired":
		rollbackReq.RollbackAmount += openBet.BetReq.DebitAmount
	case "rollback":
		// nothing to do here
	default:
		log.Println("ComputeRollback: Unexpected ResultType - ", result.ResultType)
	}
	rollbackReq.RollbackAmount = utils.Truncate64(rollbackReq.RollbackAmount)
	return rollbackReq
}
*/
// Common Result Routine called by provider specific logic
func CommonResultRoutine(openBets []sportsdto.BetDto) {
	log.Println("CommonResultRoutine: New Settlement Routine Started")
	// 1. Create a new map
	operatorBetsMap := make(map[string][]sportsdto.BetDto)
	// 2. Iterate through open bets list
	for _, openBet := range openBets {
		// 2.1. Get operator's bets list
		operatorBets, ok := operatorBetsMap[openBet.OperatorId]
		if !ok {
			// 2.1.1. if not present, create empty bets list
			operatorBets = []sportsdto.BetDto{}
		}
		// 2.2. append bet to list
		openBet.Status = constants.BetFair.BetStatus.SETTLED()
		operatorBets = append(operatorBets, openBet)
		// 2.3. update map
		operatorBetsMap[openBet.OperatorId] = operatorBets
	}
	if len(operatorBetsMap) == 0 {
		log.Println("CommonResultRoutine: ZERO operators to settle.")
		return
	}
	log.Println("CommonResultRoutine: Operators count is - ", len(operatorBetsMap))
	// 3. create one thread per operatorId
	resultChannel := make(chan OperatorResult)
	var channelCount int = 0
	for operatorId := range operatorBetsMap {
		// 3.1. create thread and pass list of openBets to it
		// 1. Get operator details by operator Id
		// TODO: Rety mechanism
		operatorDto, err := cache.GetOperatorDetails(operatorId)
		if err != nil {
			log.Println("CommonResultRoutine: cache.GetOperatorDetails failed for operatorId (" + operatorId + ") with error - " + err.Error())
			continue
		}
		log.Println("CommonResultRoutine: Bet settlement for operatorId-walletType: ", operatorId+"-"+operatorDto.WalletType)
		// 2. Check operator's wallet type
		switch operatorDto.WalletType {
		case constants.SAP.WalletType.Transfer():
			// 2.1. Transfer wallet, call transfer wallet functionality
			channelCount++
			go TransferWalletBetSettlement(operatorDto, operatorBetsMap[operatorId], resultChannel)
		case constants.SAP.WalletType.Seamless():
			// 2.2. Seamless wallet, call seamless wallet functionality
			channelCount++
			go SeamlessWalletBetSettlement(operatorDto, operatorBetsMap[operatorId], resultChannel)
		case constants.SAP.WalletType.Feed():
			// 2.3. Feed wallet, call seamless wallet functionality
			channelCount++
			go FeedWalletBetSettlement(operatorDto, operatorBetsMap[operatorId], resultChannel)
		default:
			log.Println("DoWalletRollbackRequests: Invalid wallet type - ", operatorDto.WalletType)
		}
		//rc <- opResult
		//return
	}
	// 4. wait for all threads to complete
	log.Println("CommonResultRoutine: Waiting for all Settlements to complete")
	failedOperators := 0
	opResults := make([]OperatorResult, channelCount)
	for i, _ := range opResults {
		//log.Println("CommonResultRoutine: Waiting on channel for operator - ", i+1)
		opResults[i] = <-resultChannel
		if opResults[i].FailedBets != 0 {
			msg := "Total failed bet settlements for Operator (" + opResults[i].OperatorId + ") are - " + strconv.Itoa(opResults[i].FailedBets)
			log.Println("CommonResultRoutine: Operators count on Market - ", msg)
			failedOperators++
		}
	}
	if failedOperators != 0 {
		// TODO: Overall operation is failed
		log.Println("CommonResultRoutine: Failed Operators count is - ", failedOperators)
	}
	// 6. TODO: Handle Error / retry mechanism
	log.Println("CommonResultRoutine: New Settlement Routine Ended")
}

func FeedWalletBetSettlement(operatorDto operatordto.OperatorDTO, openBets []sportsdto.BetDto, rc chan OperatorResult) {
	log.Println("FeedWalletBetSettlement: Bet Result Started!")
	opResult := OperatorResult{}
	opResult.OperatorId = operatorDto.OperatorId
	opResult.TotalBets = len(openBets)
	opResult.SettledBets = 0
	opResult.FailedBets = 0
	opResult.ErrorsList = []string{}
	count := 0
	for i, bet := range openBets {
		// Handling better odds case for LAY and for BetFair bet
		if constants.SAP.BetType.LAY() == bet.BetDetails.BetType && bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount > 0 {
			if bet.BetDetails.OddValue != bet.BetReq.OddsMatched {
				// calculate diff amount
				// diffAmount := utils.Truncate4Decfloat64(bet.BetDetails.StakeAmount * (bet.BetDetails.OddValue - bet.BetReq.OddsMatched))
				// bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount -= diffAmount
				// BUG in Feed wallet. Corrected creditamount but not debit amount.
				// It will create issues in rollbacks and reporting
				factor := bet.BetReq.SizeMatched / bet.BetReq.SizePlaced
				stakeAmount := utils.Truncate4Decfloat64(bet.BetDetails.StakeAmount * factor)
				diffAmount := utils.Truncate4Decfloat64(stakeAmount * (bet.BetDetails.OddValue - bet.BetReq.OddsMatched))
				bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount -= diffAmount // dont get updated in DB, but will communicate this to feed operator
				openBets[i].BetReq.DebitAmount -= diffAmount                     // Correction in DB for feed
				log.Println("FeedWalletBetSettlement: Better Odds Result Adjusted - ", bet.BetId, diffAmount, bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount, factor)
			}

		}
		opResp, resultReq, err := WalletResult(bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("FeedWalletBetSettlement: WalletResult failed with error - ", err.Error())
			log.Println("FeedWalletBetSettlement: WalletResult failed for betId - ", bet.BetId)
			openBets[i].Status = bet.Status + "-failed"
			log.Println("FeedWalletBetSettlement: Result Request is - ", resultReq)
			continue
		}
		if opResp.Status != "RS_OK" {
			log.Println("FeedWalletBetSettlement: Result Failed. Status is - ", opResp.Status)
			openBets[i].Status = bet.Status + "-failed"
			continue
		}
		log.Println("FeedWalletBetSettlement: Result Successfully completed for betId - ", bet.BetId)
	}
	count2, msgs := database.UpdateBets(openBets)
	if len(msgs) > 0 {
		log.Println("FeedWalletBetSettlement: Total bets     are - ", len(openBets))
		log.Println("FeedWalletBetSettlement: Total success  are - ", count2)
		log.Println("FeedWalletBetSettlement: Total failures are - ", len(msgs))
		log.Println("FeedWalletBetSettlement: Error messages are:")
		for _, msg := range msgs {
			log.Println("FeedWalletBetSettlement: *** ERROR *** - ", msg)
		}
	}
	opResult.SettledBets = count
	opResult.FailedBets = opResult.TotalBets - count
	opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	log.Println("FeedWalletBetSettlement: Bet Result Ended!")
	rc <- opResult
	return
}

func SeamlessWalletBetSettlement(operatorDto operatordto.OperatorDTO, openBets []sportsdto.BetDto, rc chan OperatorResult) {
	log.Println("SeamlessWalletBetSettlement: Bet Settlement Started for operator - ", operatorDto.OperatorId)
	log.Println("SeamlessWalletBetSettlement: Operator Base URL is - ", operatorDto.BaseURL)
	opResult := OperatorResult{}
	opResult.OperatorId = operatorDto.OperatorId
	opResult.TotalBets = len(openBets)
	opResult.SettledBets = 0
	opResult.FailedBets = opResult.TotalBets
	opResult.ErrorsList = []string{}
	// TODO: WalletResult call to operators
	count := 0
	for i, bet := range openBets {
		opResp, resultReq, err := WalletResult(bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("SeamlessWalletBetSettlement: WalletResult failed with error - ", err.Error())
			log.Println("SeamlessWalletBetSettlement: WalletResult failed for betId - ", bet.BetId)
			openBets[i].Status = bet.Status + "-failed"
			log.Println("SeamlessWalletBetSettlement: Result Request is - ", resultReq)
			continue
		}
		if opResp.Status != "RS_OK" {
			log.Println("SeamlessWalletBetSettlement: Result Failed. Status is - ", opResp.Status)
			openBets[i].Status = bet.Status + "-failed"
			continue
		}
		count++
		log.Println("SeamlessWalletBetSettlement: Result Successfully completed for betId - ", bet.BetId)
	}
	count2, msgs := database.UpdateBets(openBets)
	if len(msgs) > 0 {
		log.Println("SeamlessWalletBetSettlement: Total bets     are - ", len(openBets))
		log.Println("SeamlessWalletBetSettlement: Total success  are - ", count2)
		log.Println("SeamlessWalletBetSettlement: Total failures are - ", len(msgs))
		log.Println("SeamlessWalletBetSettlement: Error messages are:")
		for _, msg := range msgs {
			log.Println("SeamlessWalletBetSettlement: *** ERROR *** - ", msg)
		}
	}
	opResult.SettledBets = count
	opResult.FailedBets = opResult.TotalBets - count
	opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	//for _, bet := range openBets {
	//}
	// count, msgs := database.UpdateBets(openBets)
	// if len(msgs) > 0 {
	// 	log.Println("SeamlessWalletBetSettlement: Total bets     are - ", len(openBets))
	// 	log.Println("SeamlessWalletBetSettlement: Total success  are - ", count)
	// 	log.Println("SeamlessWalletBetSettlement: Total failures are - ", len(msgs))
	// 	log.Println("SeamlessWalletBetSettlement: Error messages are:")
	// 	for _, msg := range msgs {
	// 		log.Println("SeamlessWalletBetSettlement: *** ERROR *** - ", msg)
	// 	}
	// }
	// opResult.SettledBets = count
	// opResult.FailedBets = opResult.TotalBets - count
	// opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	log.Println("SeamlessWalletBetSettlement: Bet Settlement Ended!")
	rc <- opResult
	return
}

func TransferWalletBetSettlement(operatorDto operatordto.OperatorDTO, openBets []sportsdto.BetDto, rc chan OperatorResult) {
	log.Println("TransferWalletBetSettlement: Bet Settlement Started!")
	opResult := OperatorResult{}
	opResult.OperatorId = operatorDto.OperatorId
	opResult.TotalBets = len(openBets)
	opResult.SettledBets = 0
	opResult.FailedBets = opResult.TotalBets
	opResult.ErrorsList = []string{}
	// 1. get b2b users by operator id
	users := []models.B2BUserDto{}
	users, err := database.GetB2BUsers(operatorDto.OperatorId, "")
	if err != nil {
		log.Println("TransferWalletBetSettlement: Failed to get users: ", err.Error())
	}
	log.Println("TransferWalletBetSettlement: User count is: ", len(users))
	// 2. loop through openbets and prepare userMap (Id, delta) and userLedgers
	userMap := make(map[string]float64)
	userKeyRateMap := make(map[string]int32)
	userLedgers := []models.UserLedgerDto{}
	for _, bet := range openBets {
		creditAmount := bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount
		// 2.2.1. Create a user ledger object
		userLedger := models.UserLedgerDto{}
		userLedger.UserKey = bet.OperatorId + "-" + bet.UserId
		userLedger.OperatorId = bet.OperatorId
		userLedger.UserId = bet.UserId
		userLedger.TransactionType = constants.SAP.LedgerTxType.BETRESULT()
		userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
		userLedger.ReferenceId = bet.BetId
		userLedger.Amount = creditAmount - bet.Commission
		userLedger.Remark = utils.GetRemark(bet)
		userLedger.CompetitionName = bet.BetDetails.CompetitionName
		userLedger.EventName = bet.BetDetails.EventName
		userLedger.MarketType = bet.BetDetails.MarketType
		userLedger.MarketName = bet.BetDetails.MarketName
		// 2.2.2. add to user ledger list
		userLedgers = append(userLedgers, userLedger)
		// 2.2.3. add to user's delta
		delta, ok := userMap[bet.UserId]
		if !ok {
			// 2.2.3.1. if not present, initialize delta to zero
			delta = 0
		}
		delta += userLedger.Amount
		// 2.2.3. update map with user delta
		userMap[bet.UserId] = delta
		userKeyRateMap[userLedger.UserKey] = bet.BetReq.Rate
	}
	// 3. update user balances
	userDeltas := []database.UserDelta{}
	for _, user := range users {
		// 3.1. get delta from userMap
		delta, ok := userMap[user.UserId]
		if !ok {
			// 2.2.2.1. if not present, continue
			continue
		}
		userDelta := database.UserDelta{}
		userDelta.UserKey = user.UserKey
		userDelta.Delta = delta
		userDeltas = append(userDeltas, userDelta)
	}
	// 4. Save ledgers
	err = database.InsertLedgers(userLedgers)
	if err != nil {
		log.Println("TransferWalletBetSettlement: InsertLedgers failed with error: ", err.Error())
		// TODO: Mark bets as failed
		//rc <- opResult
		//return
	}
	// 5. Save balances
	count, err := database.UpdateB2BUserBalances(userDeltas)
	if err != nil {
		log.Println("TransferWalletBetSettlement: UpdateB2BUserBalances failed with error: ", err.Error())
		// TODO: Mark bets as failed
		//rc <- opResult
		//return
	}
	if count != len(userDeltas) {
		log.Println("TransferWalletBetSettlement: Failed to update user balances for user count: ", len(userDeltas)-count)
		// TODO: Mark bets as failed
		//return
	}
	// Sync Wallet
	go SyncWallets(userKeyRateMap)
	// 6. Update bets
	count, msgs := database.UpdateBets(openBets)
	if len(msgs) > 0 {
		log.Println("TransferWalletBetSettlement: Total bets     are - ", len(openBets))
		log.Println("TransferWalletBetSettlement: Total success  are - ", count)
		log.Println("TransferWalletBetSettlement: Total failures are - ", len(msgs))
		log.Println("TransferWalletBetSettlement: Error messages are:")
		for _, msg := range msgs {
			log.Println("TransferWalletBetSettlement: *** ERROR *** - ", msg)
		}
	}
	opResult.SettledBets = count
	opResult.FailedBets = opResult.TotalBets - count
	opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	// mark failure and add to retry table if there are any failures
	log.Println("TransferWalletBetSettlement: Bet Settlement Ended!")
	rc <- opResult
	return
}

// Common Rollback Routine called by provider specific logic
func CommonRollbackRoutine(rollbackType string, openBets []sportsdto.BetDto) {
	log.Println("CommonRollbackRoutine: New Rollback Routine Started for type - ", rollbackType)
	// 1. Create a new map
	operatorBetsMap := make(map[string][]sportsdto.BetDto)
	// 2. Iterate through open bets list
	for _, openBet := range openBets {
		// 2.1. Get operator's bets list
		operatorBets, ok := operatorBetsMap[openBet.OperatorId]
		if !ok {
			// 2.1.1. if not present, create empty bets list
			operatorBets = []sportsdto.BetDto{}
		}
		// 2.2. append bet to list
		operatorBets = append(operatorBets, openBet)
		// 2.3. update map
		operatorBetsMap[openBet.OperatorId] = operatorBets
	}
	if len(operatorBetsMap) == 0 {
		log.Println("CommonRollbackRoutine: ZERO operators to settle.")
		return
	}
	log.Println("CommonRollbackRoutine: Operators count is - ", len(operatorBetsMap))
	// 3. create one thread per operatorId
	resultChannel := make(chan OperatorResult)
	var channelCount int = 0
	for operatorId := range operatorBetsMap {
		// 3.1. create thread and pass list of openBets to it
		// 1. Get operator details by operator Id
		// TODO: Rety mechanism
		operatorDto, err := cache.GetOperatorDetails(operatorId)
		if err != nil {
			log.Println("CommonRollbackRoutine: cache.GetOperatorDetails failed for operatorId (" + operatorId + ") with error - " + err.Error())
			continue
		}
		log.Println("CommonRollbackRoutine: Bet settlement for operatorId-walletType: ", operatorId+"-"+operatorDto.WalletType)
		// 2. Check operator's wallet type
		switch operatorDto.WalletType {
		case constants.SAP.WalletType.Transfer():
			// 2.1. Transfer wallet, call transfer wallet functionality
			channelCount++
			go TransferWalletBetRollback(rollbackType, operatorDto, operatorBetsMap[operatorId], resultChannel)
		case constants.SAP.WalletType.Seamless():
			// 2.2. Seamless wallet, call seamless wallet functionality
			channelCount++
			go SeamlessWalletBetRollback(rollbackType, operatorDto, operatorBetsMap[operatorId], resultChannel)
		case constants.SAP.WalletType.Feed():
			// 2.3. Feed wallet, call seamless wallet functionality
			channelCount++
			go FeedWalletBetRollback(rollbackType, operatorDto, operatorBetsMap[operatorId], resultChannel)
		default:
			log.Println("DoWalletRollbackRequests: Invalid wallet type - ", operatorDto.WalletType)
		}
		//rc <- opResult
		//return
	}
	// 4. wait for all threads to complete
	log.Println("CommonRollbackRoutine: Waiting for all Rollbacks to complete")
	failedOperators := 0
	opResults := make([]OperatorResult, channelCount)
	for i, _ := range opResults {
		//log.Println("CommonRollbackRoutine: Waiting on channel for operator - ", i+1)
		opResults[i] = <-resultChannel
		if opResults[i].FailedBets != 0 {
			msg := "Total failed bet rollbacks for Operator (" + opResults[i].OperatorId + ") are - " + strconv.Itoa(opResults[i].FailedBets)
			log.Println("CommonRollbackRoutine: Operators count on Market - ", msg)
			failedOperators++
		}
	}
	if failedOperators != 0 {
		// TODO: Overall operation is failed
		log.Println("CommonRollbackRoutine: Failed Operators count is - ", failedOperators)
	}
	// 6. TODO: Handle Error / retry mechanism
	log.Println("CommonRollbackRoutine: New Rollback Routine Ended for type - ", rollbackType)
}

func FeedWalletBetRollback(rollbackType string, operatorDto operatordto.OperatorDTO, openBets []sportsdto.BetDto, rc chan OperatorResult) {
	log.Println("FeedWalletBetRollback: Bet Rollback Started for type - ", rollbackType)
	opResult := OperatorResult{}
	opResult.OperatorId = operatorDto.OperatorId
	opResult.TotalBets = len(openBets)
	opResult.SettledBets = 0
	opResult.FailedBets = 0
	opResult.ErrorsList = []string{}
	count := 0
	for i, bet := range openBets {
		//opResp, resultReq, err := WalletResult(bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		opResp, rollBackReq, err := WalletRollback(rollbackType, bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("FeedWalletBetRollback: WalletRollback failed with error - ", err.Error())
			log.Println("FeedWalletBetRollback: WalletRollback failed for betId - ", bet.BetId)
			openBets[i].Status = bet.Status + "-failed"
			log.Println("FeedWalletBetRollback: Rollback Request is - ", rollBackReq)
			continue
		}
		if opResp.Status != "RS_OK" {
			log.Println("FeedWalletBetRollback: Rollback Failed. Status is - ", opResp.Status)
			openBets[i].Status = bet.Status + "-failed"
			continue
		}
		openBets[i].Commission = 0
		log.Println("FeedWalletBetRollback: Rollback Successfully completed for betId - ", bet.BetId)
	}
	count2, msgs := database.UpdateBets(openBets)
	if len(msgs) > 0 {
		log.Println("FeedWalletBetRollback: Total bets     are - ", len(openBets))
		log.Println("FeedWalletBetRollback: Total success  are - ", count2)
		log.Println("FeedWalletBetRollback: Total failures are - ", len(msgs))
		log.Println("FeedWalletBetRollback: Error messages are:")
		for _, msg := range msgs {
			log.Println("FeedWalletBetRollback: *** ERROR *** - ", msg)
		}
	}
	opResult.SettledBets = count
	opResult.FailedBets = opResult.TotalBets - count
	opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	log.Println("FeedWalletBetRollback: Bet Rollback Ended for type - ", rollbackType)
	rc <- opResult
	return
}

func SeamlessWalletBetRollback(rollbackType string, operatorDto operatordto.OperatorDTO, openBets []sportsdto.BetDto, rc chan OperatorResult) {
	log.Println("SeamlessWalletBetRollback: Bet Settlement Started for operator - ", operatorDto.OperatorId)
	log.Println("SeamlessWalletBetRollback: Operator Base URL is - ", operatorDto.BaseURL)
	opResult := OperatorResult{}
	opResult.OperatorId = operatorDto.OperatorId
	opResult.TotalBets = len(openBets)
	opResult.SettledBets = 0
	opResult.FailedBets = opResult.TotalBets
	opResult.ErrorsList = []string{}
	// TODO: WalletResult call to operators
	count := 0
	for i, bet := range openBets {
		//opResp, resultReq, err := WalletResult(bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		opResp, rollBackReq, err := WalletRollback(rollbackType, bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("SeamlessWalletBetRollback: WalletRollback failed with error - ", err.Error())
			log.Println("SeamlessWalletBetRollback: WalletRollback failed for betId - ", bet.BetId)
			openBets[i].Status = bet.Status + "-failed"
			log.Println("SeamlessWalletBetRollback: Rollback Request is - ", rollBackReq)
			continue
		}
		if opResp.Status != "RS_OK" {
			log.Println("SeamlessWalletBetRollback: Rollback Failed. Status is - ", opResp.Status)
			openBets[i].Status = bet.Status + "-failed"
			continue
		}
		openBets[i].Commission = 0
		log.Println("SeamlessWalletBetRollback: Rollback Successfully completed for betId - ", bet.BetId)
	}
	count2, msgs := database.UpdateBets(openBets)
	if len(msgs) > 0 {
		log.Println("SeamlessWalletBetRollback: Total bets     are - ", len(openBets))
		log.Println("SeamlessWalletBetRollback: Total success  are - ", count2)
		log.Println("SeamlessWalletBetRollback: Total failures are - ", len(msgs))
		log.Println("SeamlessWalletBetRollback: Error messages are:")
		for _, msg := range msgs {
			log.Println("SeamlessWalletBetRollback: *** ERROR *** - ", msg)
		}
	}
	opResult.SettledBets = count
	opResult.FailedBets = opResult.TotalBets - count
	opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	//for _, bet := range openBets {
	//}
	// count, msgs := database.UpdateBets(openBets)
	// if len(msgs) > 0 {
	// 	log.Println("SeamlessWalletBetRollback: Total bets     are - ", len(openBets))
	// 	log.Println("SeamlessWalletBetRollback: Total success  are - ", count)
	// 	log.Println("SeamlessWalletBetRollback: Total failures are - ", len(msgs))
	// 	log.Println("SeamlessWalletBetRollback: Error messages are:")
	// 	for _, msg := range msgs {
	// 		log.Println("SeamlessWalletBetRollback: *** ERROR *** - ", msg)
	// 	}
	// }
	// opResult.SettledBets = count
	// opResult.FailedBets = opResult.TotalBets - count
	// opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	log.Println("SeamlessWalletBetRollback: Bet Rollback Ended for type - ", rollbackType)
	rc <- opResult
	return
}

func TransferWalletBetRollback(rollbackType string, operatorDto operatordto.OperatorDTO, openBets []sportsdto.BetDto, rc chan OperatorResult) {
	log.Println("TransferWalletBetRollback: Bet Rollback Started for type - ", rollbackType)
	opResult := OperatorResult{}
	opResult.OperatorId = operatorDto.OperatorId
	opResult.TotalBets = len(openBets)
	opResult.SettledBets = 0
	opResult.FailedBets = opResult.TotalBets
	opResult.ErrorsList = []string{}
	// 1. get b2b users by operator id
	users := []models.B2BUserDto{}
	users, err := database.GetB2BUsers(operatorDto.OperatorId, "")
	if err != nil {
		log.Println("TransferWalletBetRollback: Failed to get users: ", err.Error())
	}
	log.Println("TransferWalletBetRollback: User count is: ", len(users))
	// 2. loop through openbets and prepare userMap (Id, delta) and userLedgers
	userMap := make(map[string]float64)
	userKeyRateMap := make(map[string]int32)
	userLedgers := []models.UserLedgerDto{}
	for i, bet := range openBets {
		RollbackAmount := bet.RollbackReqs[len(bet.RollbackReqs)-1].RollbackAmount
		// 2.2.1. Create a user ledger object
		userLedger := models.UserLedgerDto{}
		userLedger.UserKey = bet.OperatorId + "-" + bet.UserId
		userLedger.OperatorId = bet.OperatorId
		userLedger.UserId = bet.UserId
		userLedger.TransactionType = constants.SAP.LedgerTxType.BETROLLBACK()
		userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
		userLedger.ReferenceId = bet.BetId
		userLedger.Amount = RollbackAmount - bet.Commission
		userLedger.Remark = utils.GetRemark(bet)
		userLedger.CompetitionName = bet.BetDetails.CompetitionName
		userLedger.EventName = bet.BetDetails.EventName
		userLedger.MarketType = bet.BetDetails.MarketType
		userLedger.MarketName = bet.BetDetails.MarketName
		// 2.2.2. add to user ledger list
		userLedgers = append(userLedgers, userLedger)
		openBets[i].Commission = 0
		// 2.2.3. add to user's delta
		delta, ok := userMap[bet.UserId]
		if !ok {
			// 2.2.3.1. if not present, initialize delta to zero
			delta = 0
		}
		delta += userLedger.Amount
		// 2.2.3. update map with user delta
		userMap[bet.UserId] = delta
		// 2.2.6. update bet status
		// bet status is either void, rollback, cancel or deleted
		// openBets[i].Status = rollbackType
		// switch rollbackType {
		// case "ROLLBACK":
		// 	openBets[i].Status = "OPEN"
		// default:
		// 	// TODO: Handle invalid result type
		// 	openBets[i].Status = "VOIDED"
		// 	log.Println("TransferWalletBetRollback: Invalid result type - ", rollbackType)
		// }
		// openBets[i].UpdatedAt = time.Now().UnixNano() / int64(time.Millisecond)
		userKeyRateMap[userLedger.UserKey] = bet.BetReq.Rate
	}
	// 3. update user balances
	userDeltas := []database.UserDelta{}
	for _, user := range users {
		// 3.1. get delta from userMap
		delta, ok := userMap[user.UserId]
		if !ok {
			// 2.2.2.1. if not present, continue
			continue
		}
		userDelta := database.UserDelta{}
		userDelta.UserKey = user.UserKey
		userDelta.Delta = delta
		userDeltas = append(userDeltas, userDelta)
	}
	// 4. Save ledgers
	err = database.InsertLedgers(userLedgers)
	if err != nil {
		log.Println("TransferWalletBetRollback: InsertLedgers failed with error: ", err.Error())
		// TODO: Mark bets as failed
		//rc <- opResult
		//return
	}
	// 5. Save balances
	count, err := database.UpdateB2BUserBalances(userDeltas)
	if err != nil {
		log.Println("TransferWalletBetRollback: UpdateB2BUserBalances failed with error: ", err.Error())
		// TODO: Mark bets as failed
		//rc <- opResult
		//return
	}
	if count != len(userDeltas) {
		log.Println("TransferWalletBetRollback: Failed to update user balances for user count: ", len(userDeltas)-count)
		// TODO: Mark bets as failed
		//return
	}
	// Sync Wallet
	go SyncWallets(userKeyRateMap)
	// 6. Update bets
	count, msgs := database.UpdateBets(openBets)
	if len(msgs) > 0 {
		log.Println("TransferWalletBetRollback: Total bets     are - ", len(openBets))
		log.Println("TransferWalletBetRollback: Total success  are - ", count)
		log.Println("TransferWalletBetRollback: Total failures are - ", len(msgs))
		log.Println("TransferWalletBetRollback: Error messages are:")
		for _, msg := range msgs {
			log.Println("TransferWalletBetRollback: *** ERROR *** - ", msg)
		}
	}
	opResult.SettledBets = count
	opResult.FailedBets = opResult.TotalBets - count
	opResult.ErrorsList = append(opResult.ErrorsList, msgs...)
	// mark failure and add to retry table if there are any failures
	log.Println("TransferWalletBetRollback: Bet Settlement Ended for type - ", rollbackType)
	rc <- opResult
	return
}

// Common Result Routine called by provider specific logic
func CommonMarketCommissionRoutine(userMarket models.UserMarket) {
	log.Println("CommonMarketCommissionRoutine: New Commission Routine Started")
	userSession, err := cache.GetSessionDetailsByUserKey(userMarket.OperatorId, userMarket.UserId)
	if err == nil {
		userMarket.Token = userSession.Token
	}
	// 1. Get Operator from Cache
	operator, err := cache.GetOperatorDetails(userMarket.OperatorId)
	if err != nil {
		log.Println("CommonMarketCommissionRoutine: END FAILED 1 cache.GetOperatorDetails failed with error for UserMarketKey - ", err.Error(), userMarket.UserMarketKey)
		return
	}
	// 2. Wallet Type Switch Case
	switch operator.WalletType {
	case constants.SAP.WalletType.Transfer():
		// 2.1. Transfer wallet, call transfer wallet functionality
		err = TransferWalletMarketCommission(userMarket, operator)
		if err != nil {
			log.Println("CommonMarketCommissionRoutine: END FAILED 2 TransferWalletMarketCommission failed with error for UserMarketKey - ", err.Error(), userMarket.UserMarketKey)
		}
	case constants.SAP.WalletType.Seamless():
		// 2.2. Seamless wallet, call seamless wallet functionality
		err = SeamlessWalletMarketCommission(userMarket, operator)
		if err != nil {
			log.Println("CommonMarketCommissionRoutine: END FAILED 3 SeamlessWalletMarketCommission failed with error for UserMarketKey - ", err.Error(), userMarket.UserMarketKey)
		}
	case constants.SAP.WalletType.Feed():
		// 2.3. Feed wallet, call seamless wallet functionality
		err = FeedWalletMarketCommission(userMarket, operator)
		if err != nil {
			log.Println("CommonMarketCommissionRoutine: END FAILED 4 FeedWalletMarketCommission failed with error for UserMarketKey - ", err.Error(), userMarket.UserMarketKey)
		}
	default:
		log.Println("CommonMarketCommissionRoutine: END FAILED 5 Invalid wallet type for UserMarketKey - ", operator.WalletType, userMarket.UserMarketKey)
	}
	// 6. TODO: Handle Error / retry mechanism
	log.Println("CommonMarketCommissionRoutine: END SUCCESS 2")
}

func FeedWalletMarketCommission(userMarket models.UserMarket, operator operatordto.OperatorDTO) error {
	log.Println("FeedWalletMarketCommission: Market Commission Started for - ", userMarket.UserMarketKey)
	opResp, _, err := WalletCommission(userMarket, operator)
	if err != nil {
		log.Println("FeedWalletMarketCommission: WalletCommission failed with error for usermarketkey- ", err.Error(), userMarket.UserMarketKey)
		return err
	}
	if opResp.Status != "RS_OK" {
		log.Println("FeedWalletMarketCommission: WalletCommission Failed. Status is - ", opResp.Status, userMarket.UserMarketKey)
		return fmt.Errorf(opResp.Status)
	}
	log.Println("FeedWalletMarketCommission: WalletCommission Successfully completed for UserMarketKey - ", userMarket.UserMarketKey)
	return nil
}

func SeamlessWalletMarketCommission(userMarket models.UserMarket, operator operatordto.OperatorDTO) error {
	log.Println("SeamlessWalletMarketCommission: Market Commission Started for - ", userMarket.UserMarketKey)
	opResp, _, err := WalletCommission(userMarket, operator)
	if err != nil {
		log.Println("SeamlessWalletMarketCommission: WalletCommission failed with error for usermarketkey- ", err.Error(), userMarket.UserMarketKey)
		return err
	}
	if opResp.Status != "RS_OK" {
		log.Println("SeamlessWalletMarketCommission: WalletCommission Failed. Status is - ", opResp.Status, userMarket.UserMarketKey)
		return fmt.Errorf(opResp.Status)
	}
	log.Println("SeamlessWalletMarketCommission: WalletCommission Successfully completed for UserMarketKey - ", userMarket.UserMarketKey)
	return nil
}

func TransferWalletMarketCommission(userMarket models.UserMarket, operator operatordto.OperatorDTO) error {
	log.Println("TransferWalletMarketCommission: Market Commission Started for - ", userMarket.UserMarketKey)
	userKey := userMarket.OperatorId + "-" + userMarket.UserId
	market, err := cache.GetMarketStatus(userMarket.OperatorId, userMarket.ProviderId, userMarket.SportId, userMarket.EventId, userMarket.MarketId)
	if err != nil {
		log.Println("TransferWalletMarketCommission: cache.GetMarketStatus failed with error for usermarketkey- ", err.Error(), userMarket.UserMarketKey)
		return err
	}
	// 2. loop through openbets and prepare userMap (Id, delta) and userLedgers
	creditAmount := userMarket.UserCommission - userMarket.CommissionAmount
	if userMarket.Rate != 0 {
		creditAmount = creditAmount / float64(userMarket.Rate)
	}
	creditAmount = utils.Truncate4Decfloat64(creditAmount) // round to to 4 decimal places
	// 2.2.1. Create a user ledger object
	userLedger := models.UserLedgerDto{}
	userLedger.UserKey = userKey
	userLedger.OperatorId = userMarket.OperatorId
	userLedger.UserId = userMarket.UserId
	userLedger.TransactionType = constants.SAP.LedgerTxType.MARKETCOMMISSION()
	userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	userLedger.ReferenceId = userMarket.UserMarketKey
	userLedger.Amount = creditAmount
	userLedger.Remark = ""
	userLedger.CompetitionName = market.CompetitionName
	userLedger.EventName = market.EventName
	userLedger.MarketType = market.MarketType
	userLedger.MarketName = market.MarketName
	// 4. Save ledgers
	err = database.InsertLedger(userLedger)
	if err != nil {
		log.Println("TransferWalletMarketCommission: InsertLedgers failed with error: ", err.Error(), userMarket.UserMarketKey)
		return err
	}
	// 5. Save balances
	err = database.UpdateB2BUserBalance(userKey, userLedger.Amount)
	if err != nil {
		log.Println("TransferWalletMarketCommission: UpdateB2BUserBalance failed with error: ", err.Error(), userMarket.UserMarketKey)
		return err
	}
	// Sync Wallet
	//go SyncWallets(userKeyRateMap)
	log.Println("TransferWalletMarketCommission: WalletCommission Successfully completed for UserMarketKey - ", userMarket.UserMarketKey)
	return nil
}

// Retry Failures
func RetryFailedBets() {
	// 0. SETTLED-Failed
	log.Println("RetryFailedBets: SETTLED - Start!!!")
	bets, err := database.GetBetsByStatus("", "SETTLED-failed")
	if err != nil {
		log.Println("RetryFailedBets: SETTLED database.GetBetsByStatus failed with error - ", err.Error())
	}
	if len(bets) > 0 {
		log.Println("RetryFailedBets: SETTLED Failed bets count is - ", len(bets))
		failedBets := []sportsdto.BetDto{}
		operatorBetsMap := make(map[string][]sportsdto.BetDto)
		for _, openBet := range bets {
			if openBet.OperatorId == "HypexOne" {
				continue
			}
			// 2.1. Get operator's bets list
			operatorBets, ok := operatorBetsMap[openBet.OperatorId]
			if !ok {
				// 2.1.1. if not present, create empty bets list
				operatorBets = []sportsdto.BetDto{}
			}
			// 2.2. append bet to list
			openBet.Status = constants.SAP.BetStatus.SETTLED()
			failedBets = append(failedBets, openBet)
			operatorBets = append(operatorBets, openBet)
			// 2.3. update map
			operatorBetsMap[openBet.OperatorId] = operatorBets
		}
		if len(failedBets) > 0 {
			log.Println("RetryFailedBets: SETTLED Operators count is - ", len(operatorBetsMap))
			for key, value := range operatorBetsMap {
				log.Println("RetryFailedBets: SETTLED OperatorId is - ", key)
				log.Println("RetryFailedBets: SETTLED Bets Count is - ", len(value))
			}
			log.Println("RetryFailedBets: SETTLED failedBets count is - ", len(failedBets))
			CommonResultRoutine(failedBets)
		}
	}
	log.Println("RetryFailedBets: SETTLED - End!!!")

	// 1. VOIDED-Failed
	log.Println("RetryFailedBets: VOIDED - Start!!!")
	bets, err = database.GetBetsByStatus("", "VOIDED-failed")
	if err != nil {
		log.Println("RetryFailedBets: VOIDED database.GetBetsByStatus failed with error - ", err.Error())
	}
	if len(bets) > 0 {
		log.Println("RetryFailedBets: VOIDED Failed bets count is - ", len(bets))
		failedBets := []sportsdto.BetDto{}
		operatorBetsMap := make(map[string][]sportsdto.BetDto)
		for _, openBet := range bets {
			if openBet.OperatorId == "HypexOne" {
				continue
			}
			// 2.1. Get operator's bets list
			operatorBets, ok := operatorBetsMap[openBet.OperatorId]
			if !ok {
				// 2.1.1. if not present, create empty bets list
				operatorBets = []sportsdto.BetDto{}
			}
			// 2.2. append bet to list
			openBet.Status = constants.SAP.BetStatus.VOIDED()
			failedBets = append(failedBets, openBet)
			operatorBets = append(operatorBets, openBet)
			// 2.3. update map
			operatorBetsMap[openBet.OperatorId] = operatorBets
		}
		if len(failedBets) > 0 {
			log.Println("RetryFailedBets: VOIDED Operators count is - ", len(operatorBetsMap))
			for key, value := range operatorBetsMap {
				log.Println("RetryFailedBets: VOIDED OperatorId is - ", key)
				log.Println("RetryFailedBets: VOIDED Bets Count is - ", len(value))
			}
			log.Println("RetryFailedBets: VOIDED failedBets count is - ", len(failedBets))
			CommonRollbackRoutine(constants.SAP.BetStatus.VOIDED(), failedBets)
		}
	}
	log.Println("RetryFailedBets: VOIDED - End!!!")

	// 2. LAPSED-Failed
	log.Println("RetryFailedBets: LAPSED - Start!!!")
	bets, err = database.GetBetsByStatus("", "LAPSED-failed")
	if err != nil {
		log.Println("RetryFailedBets: LAPSED database.GetBetsByStatus failed with error - ", err.Error())
	}
	if len(bets) > 0 {
		log.Println("RetryFailedBets: LAPSED Failed bets count is - ", len(bets))
		failedBets := []sportsdto.BetDto{}
		operatorBetsMap := make(map[string][]sportsdto.BetDto)
		for _, openBet := range bets {
			if openBet.OperatorId == "HypexOne" {
				continue
			}
			// 2.1. Get operator's bets list
			operatorBets, ok := operatorBetsMap[openBet.OperatorId]
			if !ok {
				// 2.1.1. if not present, create empty bets list
				operatorBets = []sportsdto.BetDto{}
			}
			// 2.2. append bet to list
			openBet.Status = constants.SAP.BetStatus.LAPSED()
			failedBets = append(failedBets, openBet)
			operatorBets = append(operatorBets, openBet)
			// 2.3. update map
			operatorBetsMap[openBet.OperatorId] = operatorBets
		}
		if len(failedBets) > 0 {
			log.Println("RetryFailedBets: Operators count is - ", len(operatorBetsMap))
			for key, value := range operatorBetsMap {
				log.Println("RetryFailedBets: OperatorId is - ", key)
				log.Println("RetryFailedBets: Bets Count is - ", len(value))
			}
			log.Println("RetryFailedBets: LAPSED failedBets count is - ", len(failedBets))
			CommonRollbackRoutine(constants.SAP.BetStatus.LAPSED(), failedBets)
		}
	}
	log.Println("RetryFailedBets: LAPSED - End!!!")

	// 3. CANCELLED-Failed
	log.Println("RetryFailedBets: CANCELLED - Start!!!")
	bets, err = database.GetBetsByStatus("", "CANCELLED-failed")
	if err != nil {
		log.Println("RetryFailedBets: database.GetBetsByStatus failed with error - ", err.Error())
	}
	if len(bets) > 0 {
		log.Println("RetryFailedBets: CANCELLED Failed bets count is - ", len(bets))
		failedBets := []sportsdto.BetDto{}
		operatorBetsMap := make(map[string][]sportsdto.BetDto)
		for _, openBet := range bets {
			if openBet.OperatorId == "HypexOne" {
				continue
			}
			// 2.1. Get operator's bets list
			operatorBets, ok := operatorBetsMap[openBet.OperatorId]
			if !ok {
				// 2.1.1. if not present, create empty bets list
				operatorBets = []sportsdto.BetDto{}
			}
			// 2.2. append bet to list
			openBet.Status = constants.SAP.BetStatus.CANCELLED()
			failedBets = append(failedBets, openBet)
			operatorBets = append(operatorBets, openBet)
			// 2.3. update map
			operatorBetsMap[openBet.OperatorId] = operatorBets
		}
		if len(failedBets) > 0 {
			log.Println("RetryFailedBets: CANCELLED Operators count is - ", len(operatorBetsMap))
			for key, value := range operatorBetsMap {
				log.Println("RetryFailedBets: CANCELLED OperatorId is - ", key)
				log.Println("RetryFailedBets: CANCELLED Bets Count is - ", len(value))
			}
			log.Println("RetryFailedBets: CANCELLED failedBets count is - ", len(failedBets))
			CommonRollbackRoutine(constants.SAP.BetStatus.CANCELLED(), failedBets)
		}
	}
	log.Println("RetryFailedBets: CANCELLED - End!!!")

	// 4. ROLLBACK-Failed
	log.Println("RetryFailedBets: ROLLBACK - Start!!!")
	bets, err = database.GetBetsByStatus("", "OPEN-failed")
	if err != nil {
		log.Println("RetryFailedBets: ROLLBACK database.GetBetsByStatus failed with error - ", err.Error())
	}
	if len(bets) > 0 {
		log.Println("RetryFailedBets: ROLLBACK Failed bets count is - ", len(bets))
		failedBets := []sportsdto.BetDto{}
		operatorBetsMap := make(map[string][]sportsdto.BetDto)
		for _, openBet := range bets {
			if openBet.OperatorId == "HypexOne" {
				continue
			}
			// 2.1. Get operator's bets list
			operatorBets, ok := operatorBetsMap[openBet.OperatorId]
			if !ok {
				// 2.1.1. if not present, create empty bets list
				operatorBets = []sportsdto.BetDto{}
			}
			// 2.2. append bet to list
			openBet.Status = constants.SAP.BetStatus.OPEN()
			failedBets = append(failedBets, openBet)
			operatorBets = append(operatorBets, openBet)
			// 2.3. update map
			operatorBetsMap[openBet.OperatorId] = operatorBets
		}
		if len(failedBets) > 0 {
			log.Println("RetryFailedBets: ROLLBACK Operators count is - ", len(operatorBetsMap))
			for key, value := range operatorBetsMap {
				log.Println("RetryFailedBets: ROLLBACK OperatorId is - ", key)
				log.Println("RetryFailedBets: ROLLBACK Bets Count is - ", len(value))
			}
			log.Println("RetryFailedBets: ROLLBACK failedBets count is - ", len(failedBets))
			CommonRollbackRoutine(constants.SAP.BetStatus.ROLLBACK(), failedBets)
		}
	}
	log.Println("RetryFailedBets: ROLLBACK - End!!!")

	// 5. EXPIRED-Failed
	log.Println("RetryFailedBets: EXPIRED - Start!!!")
	bets, err = database.GetBetsByStatus("", "EXPIRED-failed")
	if err != nil {
		log.Println("RetryFailedBets: EXPIRED database.GetBetsByStatus failed with error - ", err.Error())
	}
	if len(bets) > 0 {
		log.Println("RetryFailedBets: EXPIRED Failed bets count is - ", len(bets))
		failedBets := []sportsdto.BetDto{}
		operatorBetsMap := make(map[string][]sportsdto.BetDto)
		for _, openBet := range bets {
			if openBet.OperatorId == "HypexOne" {
				continue
			}
			// 2.1. Get operator's bets list
			operatorBets, ok := operatorBetsMap[openBet.OperatorId]
			if !ok {
				// 2.1.1. if not present, create empty bets list
				operatorBets = []sportsdto.BetDto{}
			}
			// 2.2. append bet to list
			openBet.Status = constants.SAP.BetStatus.EXPIRED()
			failedBets = append(failedBets, openBet)
			operatorBets = append(operatorBets, openBet)
			// 2.3. update map
			operatorBetsMap[openBet.OperatorId] = operatorBets
		}
		if len(failedBets) > 0 {
			log.Println("RetryFailedBets: EXPIRED Operators count is - ", len(operatorBetsMap))
			for key, value := range operatorBetsMap {
				log.Println("RetryFailedBets: EXPIRED OperatorId is - ", key)
				log.Println("RetryFailedBets: EXPIRED Bets Count is - ", len(value))
			}
			log.Println("RetryFailedBets: EXPIRED failedBets count is - ", len(failedBets))
			CommonRollbackRoutine(constants.SAP.BetStatus.EXPIRED(), failedBets)
		}
	}
	log.Println("RetryFailedBets: EXPIRED - End!!!")
}

func SyncWallets(userKeysRateMap map[string]int32) {
	log.Println("SyncWallets: START!!!")
	// create userKeys list
	userKeys := make([]string, 0, len(userKeysRateMap))
	for k := range userKeysRateMap {
		userKeys = append(userKeys, k)
	}
	// Read UserBalances from database
	operatorsMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println("SyncWallets: cache.GetObjectMap failed with error: ", err.Error())
		return
	}
	users, err := database.GetB2BUsersByKeys(userKeys)
	if err != nil {
		log.Println("SyncWallets: database.GetB2BUsersByKeys failed with error: ", err.Error())
		return
	}
	for _, user := range users {
		opObject := operatorsMap[user.OperatorId]
		operator := opObject.(operatordto.OperatorDTO)
		balance := user.Balance / float64(userKeysRateMap[user.UserKey])
		respObj, err := WalletSync(user.UserId, balance, operator.BaseURL, operator.Keys.PrivateKey)
		if err != nil {
			log.Println("SyncWallets: WalletSync failed with error: ", err.Error())
			continue
		}
		if respObj.Status != "RS_OK" {
			log.Println("SyncWallets: WalletSync failed with Status: ", respObj.Status)
		}
	}
	log.Println("SyncWallets: END!!!")
	return
}
