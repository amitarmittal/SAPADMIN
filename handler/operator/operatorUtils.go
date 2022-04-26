package operatorsvc

import (
	"Sp/cache"
	"Sp/database"
	dto "Sp/dto/core"
	"Sp/dto/models"
	opdto "Sp/dto/operator"
	portaldto "Sp/dto/portal"
	sessDto "Sp/dto/session"
	"Sp/dto/sports"
	"Sp/operator"
	utils "Sp/utilities"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	OperatorHttpReqTimeout time.Duration = 5
	ErrorUserIdMissing     string        = "Bad request - UserId is missing!"
	ErrorOperatorIdMissing string        = "Bad request - OperatorId is missing!"
	ErrorProviderIdMissing string        = "Bad request - ProviderId is missing!"
	ErrorSportIdMissing    string        = "Bad request - SportId is missing!"
	ErrorStartDateMissing  string        = "Bad request - StartDate is missing!"
	ErrorEndDateMissing    string        = "Bad request - EndDate is missing!"
	ErrorInvalidRange      string        = "Bad request - Filter value out of range!"
)

func IssueB2BSessionToken(reqDto dto.AuthReqDto, opDto opdto.OperatorDTO) sessDto.B2BSessionDto {
	// 1. Create New Token
	newToken := uuid.New().String()
	sessionDto := sessDto.B2BSessionDto{}
	sessionDto.OperatorId = reqDto.OperatorId
	sessionDto.PartnerId = reqDto.PartnerId
	sessionDto.UserId = reqDto.UserId
	sessionDto.UserName = reqDto.Username
	sessionDto.UserIp = reqDto.ClientIp
	sessionDto.Token = newToken
	sessionDto.UserKey = reqDto.OperatorId + "-" + reqDto.UserId
	sessionDto.CreatedAt = time.Now().UnixMilli()
	sessionDto.ExpireAt = time.Now().Add(time.Minute * 60 * 8).Unix()
	sessionDto.BaseURL = opDto.BaseURL
	// 2. Save in DB
	err := database.InsertSessionDetails(sessionDto)
	if err != nil {
		// TODO: Handle error
		log.Println("IssueB2BSessionToken: DB Insertion FAILED with error - ", err.Error())
	}
	// 3. Save in Cache
	cache.SetSessionDetails(sessionDto)
	// 4. Save in User Cache
	return sessionDto
}

func OpBalanceCall(sessDto sessDto.B2BSessionDto, priKey string) {
	// Sleep for 2 seconds
	log.Println("OpBalanceCall: delayed start of 2 seconds")
	time.Sleep(2 * time.Second)
	opBalRespDto, err := operator.WalletBalance(sessDto, priKey)
	if err != nil {
		// TODO: Handle this error appropriately
		log.Println("OpBalanceCall: Failed in getting user balance using Wallet Balance call! - ", err.Error())
		return
	}
	if opBalRespDto.Status != "RS_OK" {
		// TODO: Handle this error appropriately
		log.Println("OpBalanceCall: Operator's Wallet Balance call returned error - !", opBalRespDto.Status)
		return
	}
	//log.Println("OpBalanceCall: User Id is - ", sessDto.UserId)
	//log.Println("OpBalanceCall: User Token is - ", sessDto.Token)
	//log.Println("OpBalanceCall: User Balance is - ", opBalRespDto.Balance)
	return
}

func CheckGetBets(req opdto.BetsHistoryReqDto) error {
	// 1. OperatorId must have
	if req.OperatorId == "" {
		return fmt.Errorf(ErrorOperatorIdMissing)
	}
	// 2. ProviderId need if SportId provided
	if req.SportId != "" && req.ProviderId == "" {
		return fmt.Errorf(ErrorProviderIdMissing)
	}
	// 3. SportId needed if EventId provided
	if req.EventId != "" && req.SportId == "" {
		return fmt.Errorf(ErrorSportIdMissing)
	}
	// 4. EndDate needed if StartDate provided
	if req.StartDate != 0 && req.EndDate == 0 {
		return fmt.Errorf(ErrorEndDateMissing)
	}
	// 5. StartDate needed if EndDate provided
	if req.EndDate != 0 && req.StartDate == 0 {
		return fmt.Errorf(ErrorStartDateMissing)
	}
	// 6. -ve values check
	if req.Page < 0 || req.PageSize < 0 || req.StartDate < 0 || req.EndDate < 0 {
		return fmt.Errorf(ErrorInvalidRange)
	}
	return nil
}

func CheckGetAllBets(req opdto.BetsHistoryReqDto) error {

	// 1. ProviderId need if SportId provided
	if req.SportId != "" && req.ProviderId == "" {
		return fmt.Errorf(ErrorProviderIdMissing)
	}
	// 2. SportId needed if EventId provided
	//if req.EventId != "" && req.SportId == "" {
	//	return fmt.Errorf(ErrorSportIdMissing)
	//}
	// 3. EndDate needed if StartDate provided
	if req.StartDate != 0 && req.EndDate == 0 {
		return fmt.Errorf(ErrorEndDateMissing)
	}
	// 4. StartDate needed if EndDate provided
	if req.EndDate != 0 && req.StartDate == 0 {
		return fmt.Errorf(ErrorStartDateMissing)
	}
	// 5. -ve values check
	if req.Page < 0 || req.PageSize < 0 || req.StartDate < 0 || req.EndDate < 0 {
		return fmt.Errorf(ErrorInvalidRange)
	}
	return nil
}

func GetBetHistory(bet sports.BetDto) opdto.BetHistory {
	betHistory := opdto.BetHistory{}
	betHistory.BetId = bet.BetId
	betHistory.UserId = bet.UserId
	betHistory.UserName = bet.UserName
	betHistory.ProviderName = utils.ProviderMapById[bet.ProviderId]
	betHistory.SportName = bet.BetDetails.SportName
	betHistory.CompetitionName = bet.BetDetails.CompetitionName
	betHistory.EventName = bet.BetDetails.EventName
	betHistory.MarketType = bet.BetDetails.MarketType
	betHistory.MarketName = bet.BetDetails.MarketName
	betHistory.BetAmount = utils.Truncate64(bet.BetReq.DebitAmount)
	betHistory.WonAmount = utils.Truncate64(bet.NetAmount) // GetWinAmount(bet)
	betHistory.BetTime = bet.BetReq.ReqTime
	betHistory.Status = bet.Status
	betHistory.UserIp = bet.UserIp
	betHistory.UpdatedAt = bet.UpdatedAt
	// Bet Req
	betHistory.BetReq.BetType = bet.BetDetails.BetType
	betHistory.BetReq.BetOdds = bet.BetDetails.OddValue
	betHistory.BetReq.BetStake = utils.Truncate64(bet.BetDetails.StakeAmount)
	betHistory.BetReq.RunnerName = bet.BetDetails.RunnerName
	betHistory.BetReq.SessionOutcome = bet.BetDetails.SessionOutcome
	betHistory.BetReq.BetTime = bet.BetReq.ReqTime
	// Bet Results if any
	betHistory.BetResults = []opdto.BetResult{}
	for _, result := range bet.ResultReqs {
		betResult := opdto.BetResult{}
		betResult.RunnerName = result.RunnerName
		betResult.SessionOutcome = result.SessionOutcome
		betResult.ResultTime = result.ReqTime
		betResult.WinAmount = utils.Truncate64(result.CreditAmount)
		betHistory.BetResults = append(betHistory.BetResults, betResult)
	}
	// Bet Rollbacks if any
	betHistory.BetRollbacks = []opdto.BetRollback{}
	for _, rollback := range bet.RollbackReqs {
		betRollback := opdto.BetRollback{}
		betRollback.RollbackReason = rollback.RollbackReason
		betRollback.RollbackTime = rollback.ReqTime
		betRollback.RollbackAmount = utils.Truncate64(rollback.RollbackAmount)
		betHistory.BetRollbacks = append(betHistory.BetRollbacks, betRollback)
	}
	return betHistory
}

func GetBetHistory2(bet sports.BetDto) opdto.BetHistory {
	if bet.BetReq.Rate == 0 {
		bet.BetReq.Rate = 1
	}
	betHistory := opdto.BetHistory{}
	betHistory.BetId = bet.BetId
	betHistory.UserId = bet.UserId
	betHistory.UserName = bet.UserName
	betHistory.ProviderName = utils.ProviderMapById[bet.ProviderId]
	betHistory.SportName = bet.BetDetails.SportName
	betHistory.CompetitionName = bet.BetDetails.CompetitionName
	betHistory.EventName = bet.BetDetails.EventName
	betHistory.EventId = bet.EventId
	betHistory.MarketType = bet.BetDetails.MarketType
	betHistory.MarketName = bet.BetDetails.MarketName
	betHistory.BetAmount = utils.Truncate64(bet.BetReq.DebitAmount / float64(bet.BetReq.Rate))
	betHistory.WonAmount = utils.Truncate64(bet.NetAmount / float64(bet.BetReq.Rate)) //bet.NetAmount // GetWinAmount(bet)
	betHistory.BetTime = bet.BetReq.ReqTime
	betHistory.Status = bet.Status
	betHistory.UserIp = bet.UserIp
	betHistory.UpdatedAt = bet.UpdatedAt
	// Bet Req
	betHistory.BetReq.BetType = bet.BetDetails.BetType
	betHistory.BetReq.BetOdds = bet.BetDetails.OddValue
	betHistory.BetReq.UnifiedOdds = utils.GetOddsFactor(bet.BetDetails.OddValue, bet.BetDetails.MarketType)
	betHistory.BetReq.BetStake = utils.Truncate64(bet.BetDetails.StakeAmount / float64(bet.BetReq.Rate)) // bet.BetDetails.StakeAmount
	betHistory.BetReq.RunnerName = bet.BetDetails.RunnerName
	betHistory.BetReq.SessionOutcome = bet.BetDetails.SessionOutcome
	betHistory.BetReq.BetTime = bet.BetReq.ReqTime
	// Bet Results if any
	betHistory.BetResults = []opdto.BetResult{}
	for _, result := range bet.ResultReqs {
		betResult := opdto.BetResult{}
		betResult.RunnerName = result.RunnerName
		betResult.SessionOutcome = result.SessionOutcome
		betResult.ResultTime = result.ReqTime
		betResult.WinAmount = utils.Truncate64(result.CreditAmount / float64(bet.BetReq.Rate)) // result.CreditAmount
		betHistory.BetResults = append(betHistory.BetResults, betResult)
	}
	// Bet Rollbacks if any
	betHistory.BetRollbacks = []opdto.BetRollback{}
	for _, rollback := range bet.RollbackReqs {
		betRollback := opdto.BetRollback{}
		betRollback.RollbackReason = rollback.RollbackReason
		betRollback.RollbackTime = rollback.ReqTime
		betRollback.RollbackAmount = utils.Truncate64(rollback.RollbackAmount / float64(bet.BetReq.Rate)) // rollback.RollbackAmount
		betHistory.BetRollbacks = append(betHistory.BetRollbacks, betRollback)
	}
	return betHistory
}

func GetWinAmount(bet sports.BetDto) float64 {
	var winAmount float64 = 0
	for _, result := range bet.ResultReqs {
		winAmount += result.CreditAmount
	}
	for _, rollback := range bet.RollbackReqs {
		winAmount += rollback.RollbackAmount
	}
	winAmount -= bet.BetReq.DebitAmount
	return utils.Truncate4Decfloat64(winAmount)
	//return winAmount
}

func GetProvidersList(operatorId string, partnerId string) []models.Provider {
	providers := []models.Provider{}
	// 1. Get Blocked Providers list from "providers" collection and say col-1
	blockedProviders, err := database.GetBlockedProviders()
	if err != nil {
		log.Println("GetProvidersList: GetBlockedProviders failed with error: ", err.Error())
		return providers
	}
	// 2. Get ACTIVE Providers by operator from "provider_status" collection and say col-2
	activeProviders, err := database.GetActiveProvidersPS(operatorId, partnerId)
	if err != nil {
		log.Println("GetProvidersList: GetBlockedProviders failed with error: ", err.Error())
		return providers
	}
	// 3. Loop throug col-2, loop throug col-1, if providerId matches in col-1, skip, else add to the new list
	for _, aProvider := range activeProviders {
		isBlocked := false
		for _, bProvider := range blockedProviders {
			if aProvider.ProviderId == bProvider.ProviderId {
				isBlocked = true
				break
			}
		}
		if isBlocked {
			continue
		}
		providers = append(providers, models.Provider{ProviderId: aProvider.ProviderId, ProviderName: aProvider.ProviderName, Status: aProvider.OperatorStatus})
	}
	// 4. Return
	return providers
}

func CheckUserStatement(req portaldto.UserStatementReqDto) error {
	// 1. OperatorId must have
	if req.OperatorId == "" {
		return fmt.Errorf(ErrorOperatorIdMissing)
	}
	// 2. ProviderId need if SportId provided
	if req.UserId == "" {
		return fmt.Errorf(ErrorUserIdMissing)
	}
	// 3. EndDate needed if StartDate provided
	if req.StartDate != 0 && req.EndDate == 0 {
		return fmt.Errorf(ErrorEndDateMissing)
	}
	// 4. StartDate needed if EndDate provided
	if req.EndDate != 0 && req.StartDate == 0 {
		return fmt.Errorf(ErrorStartDateMissing)
	}
	// 5. -ve values check
	if req.Page < 0 || req.PageSize < 0 || req.StartDate < 0 || req.EndDate < 0 {
		return fmt.Errorf(ErrorInvalidRange)
	}
	return nil
}

func GetTransaction(ul models.UserLedgerDto) portaldto.UserTransaction {
	ut := portaldto.UserTransaction{}
	ut.TxTime = ul.TransactionTime
	ut.TxType = ul.TransactionType
	ut.RefId = ul.ReferenceId
	ut.Amount = ul.Amount
	ut.Remark = ul.Remark
	return ut
}
