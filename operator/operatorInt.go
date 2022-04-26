package operator

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	sessiondto "Sp/dto/session"
	sportsdto "Sp/dto/sports"
	keyutils "Sp/utilities"
	utils "Sp/utilities"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	ReqTimeout time.Duration = 15
	//jerambroOpId               = "jerambro"
	//jerambroURL  string        = "https://admin-api.jerambro.com/api/v1/sap/"
	//jerambroURL string = "https://admin-api.jerambro.com/api/v1/sap/market-result"
)

func WalletBalance(sessDto sessiondto.B2BSessionDto, priKey string) (operatordto.OperatorRespDto, error) {
	// 0. Create default response object
	respDto := operatordto.OperatorRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct balance request object
	reqDto := operatordto.BalanceReqDto{}
	reqDto.OperatorId = sessDto.OperatorId
	reqDto.Token = sessDto.Token
	reqDto.UserId = sessDto.UserId
	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletBalance: json Marshal failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(priKey)
	if err != nil {
		log.Println("WalletBalance: parsing private key failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	// 4. Make Balance wallet call
	log.Println("WalletBalance: request data is - ", string(reqDtoJson))
	respDto, err = WalletCall(reqDtoJson, *rsaPriKey, sessDto.BaseURL+"balance", ReqTimeout)
	if err != nil {
		log.Println("WalletBalance: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, err
	}
	respBytes, err := json.Marshal(respDto)
	log.Println("WalletBalance: response data for "+reqDto.OperatorId+"-"+reqDto.UserId+" is - ", string(respBytes))
	// 5. Return response
	return respDto, nil
}

func WalletBet(betDto sportsdto.BetDto, sessDto sessiondto.B2BSessionDto, priKey string) (operatordto.OperatorRespDto, error) {
	// 0. Create default response object
	respDto := operatordto.OperatorRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct bet request object
	reqDto := operatordto.BetReqDto{}
	reqDto.OperatorId = sessDto.OperatorId
	reqDto.Token = sessDto.Token
	reqDto.UserId = sessDto.UserId
	reqDto.OddValue = utils.GetOddsFactor(betDto.BetDetails.OddValue, betDto.BetDetails.MarketType) // betDto.BetDetails.OddValue
	reqDto.ReqId = betDto.BetReq.ReqId
	reqDto.TransactionId = betDto.BetId
	reqDto.BetType = betDto.BetDetails.BetType
	reqDto.EventId = betDto.EventId
	reqDto.MarketId = betDto.MarketId
	reqDto.DebitAmount = betDto.BetReq.DebitAmount
	if betDto.BetReq.Rate != 0 {
		reqDto.DebitAmount = betDto.BetReq.DebitAmount / float64(betDto.BetReq.Rate)
	} else { // Old bets
		reqDto.DebitAmount = betDto.BetReq.DebitAmount
	}

	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletBet: json Marshal failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(priKey)
	if err != nil {
		log.Println("WalletBet: parsing private key failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	log.Println("WalletBet: request data is - ", string(reqDtoJson))
	// 4. Make Balance wallet call
	respDto, err = WalletCall(reqDtoJson, *rsaPriKey, sessDto.BaseURL+"betrequest", ReqTimeout)
	if err != nil {
		log.Println("WalletBet: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, err
	}
	respBytes, err := json.Marshal(respDto)
	log.Println("WalletBet: response data for reqId "+reqDto.ReqId+" is - ", string(respBytes))
	// 5. Return response
	return respDto, nil
}

func WalletResult(betDto sportsdto.BetDto, baseUrl string, priKey string) (operatordto.OperatorResultRespDto, operatordto.ResultReqDto, error) {
	log.Println("WalletResult: for betId - ", betDto.BetId)
	// 0. Create default response object
	respDto := operatordto.OperatorResultRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct result request object
	reqDto := operatordto.ResultReqDto{}
	reqDto.OperatorId = betDto.OperatorId
	reqDto.Token = betDto.Token
	reqDto.UserId = betDto.UserId
	reqDto.ReqId = betDto.ResultReqs[len(betDto.ResultReqs)-1].ReqId
	reqDto.TransactionId = betDto.BetId
	reqDto.EventId = betDto.EventId
	reqDto.MarketId = betDto.MarketId
	creditAmount := betDto.ResultReqs[len(betDto.ResultReqs)-1].CreditAmount - betDto.Commission
	if betDto.BetReq.Rate != 0 {
		reqDto.CreditAmount = creditAmount / float64(betDto.BetReq.Rate)
	} else { // Old bets
		reqDto.CreditAmount = creditAmount
	}
	reqDto.CreditAmount = utils.Truncate4Decfloat64(reqDto.CreditAmount) // round to to 4 decimal places
	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletResult: json Marshal failed with error - ", err.Error())
		return respDto, reqDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(priKey)
	if err != nil {
		log.Println("WalletResult: parsing private key failed with error - ", err.Error())
		return respDto, reqDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	// 4. Make Result wallet call
	log.Println("WalletResult: request data is - ", string(reqDtoJson))
	respBytes, err := WalletCall2(reqDtoJson, *rsaPriKey, baseUrl+"resultrequest", ReqTimeout)
	if err != nil {
		log.Println("WalletResult: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, reqDto, err
	}
	// 5. Convert response body to Operator Response object
	log.Println("WalletResult: response data for reqId "+reqDto.ReqId+" is - ", string(respBytes))
	err = json.Unmarshal([]byte(respBytes), &respDto)
	if err != nil {
		log.Println("WalletResult: json.Unmarshal failed with error - ", err.Error())
		log.Println(string(respBytes))
	}
	// 6. Return response
	return respDto, reqDto, nil
}

func WalletRollback(rollbackType string, betDto sportsdto.BetDto, baseUrl string, priKey string) (operatordto.OperatorRollbackRespDto, operatordto.RollbackReqDto, error) {
	// 0. Create default response object
	respDto := operatordto.OperatorRollbackRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct result request object
	reqDto := operatordto.RollbackReqDto{}
	reqDto.OperatorId = betDto.OperatorId
	reqDto.Token = betDto.Token
	reqDto.UserId = betDto.UserId
	reqDto.ReqId = betDto.RollbackReqs[len(betDto.RollbackReqs)-1].ReqId
	reqDto.TransactionId = betDto.BetId
	reqDto.EventId = betDto.EventId
	reqDto.MarketId = betDto.MarketId
	rollbackAmount := betDto.RollbackReqs[len(betDto.RollbackReqs)-1].RollbackAmount - betDto.Commission
	if betDto.BetReq.Rate != 0 {
		reqDto.RollbackAmount = rollbackAmount / float64(betDto.BetReq.Rate)
	} else { // Old bets
		reqDto.RollbackAmount = rollbackAmount
	}
	reqDto.RollbackType = rollbackType
	reqDto.RollbackReason = betDto.RollbackReqs[len(betDto.RollbackReqs)-1].RollbackReason

	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletRollback: json Marshal failed with error - ", err.Error())
		return respDto, reqDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(priKey)
	if err != nil {
		log.Println("WalletRollback: parsing private key failed with error - ", err.Error())
		return respDto, reqDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	// 4. Make Result wallet call
	log.Println("WalletRollback: request data is - ", string(reqDtoJson))
	respBytes, err := WalletCall2(reqDtoJson, *rsaPriKey, baseUrl+"rollbackrequest", ReqTimeout)
	if err != nil {
		log.Println("WalletRollback: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, reqDto, err
	}
	// 5. Convert response body to Operator Response object
	log.Println("WalletRollback: response data for reqId "+reqDto.ReqId+" is - ", string(respBytes))
	err = json.Unmarshal([]byte(respBytes), &respDto)
	if err != nil {
		log.Println("WalletRollback: json.Unmarshal failed with error - ", err.Error())
		log.Println(string(respBytes))
	}
	// 6. Return response
	return respDto, reqDto, nil
}

func WalletCommission(userMarket models.UserMarket, operator operatordto.OperatorDTO) (operatordto.OperatorResultRespDto, operatordto.CommissionReq, error) {
	log.Println("WalletCommission: for userMarketKey - ", userMarket.UserMarketKey)
	// 0. Create default response object
	respDto := operatordto.OperatorResultRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct result request object
	reqDto := operatordto.CommissionReq{}
	reqDto.ReqId = uuid.New().String()
	reqDto.TransactionId = userMarket.UserMarketKey
	reqDto.Token = userMarket.Token
	reqDto.OperatorId = userMarket.OperatorId
	reqDto.UserId = userMarket.UserId
	reqDto.ProviderId = userMarket.ProviderId
	reqDto.SportId = userMarket.SportId
	reqDto.CompetitionId = userMarket.CompetitionId
	reqDto.EventId = userMarket.EventId
	reqDto.MarketId = userMarket.MarketId
	reqDto.WinningAmount = userMarket.WinningAmount
	reqDto.Commission = userMarket.Commission
	reqDto.CommissionAmount = userMarket.CommissionAmount
	reqDto.UserCommission = userMarket.UserCommission
	if userMarket.Rate > 0 {
		reqDto.WinningAmount = userMarket.WinningAmount / float64(userMarket.Rate)
		reqDto.CommissionAmount = userMarket.CommissionAmount / float64(userMarket.Rate)
		reqDto.UserCommission = userMarket.UserCommission / float64(userMarket.Rate)
	}
	reqDto.CommissionCredit = utils.Truncate4Decfloat64(reqDto.UserCommission - reqDto.CommissionAmount)
	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletCommission: json Marshal failed with error - ", err.Error())
		return respDto, reqDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(operator.Keys.PrivateKey)
	if err != nil {
		log.Println("WalletCommission: parsing private key failed with error - ", err.Error())
		return respDto, reqDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	// 4. Make Result wallet call
	log.Println("WalletCommission: request data is - ", string(reqDtoJson))
	respBytes, err := WalletCall2(reqDtoJson, *rsaPriKey, operator.BaseURL+"commissionrequest", ReqTimeout)
	if err != nil {
		log.Println("WalletCommission: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, reqDto, err
	}
	// 5. Convert response body to Operator Response object
	log.Println("WalletCommission: response data for reqId "+reqDto.ReqId+" is - ", string(respBytes))
	err = json.Unmarshal([]byte(respBytes), &respDto)
	if err != nil {
		log.Println("WalletCommission: json.Unmarshal failed with error - ", err.Error())
		log.Println(string(respBytes))
	}
	// 6. Return response
	return respDto, reqDto, nil
}

func WalletSync(userId string, balance float64, baseUrl string, priKey string) (operatordto.OperatorRespDto, error) {
	// 0. Create default response object
	respDto := operatordto.OperatorRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct bet request object
	reqDto := operatordto.SyncWalletReqDto{}
	reqDto.UserId = userId
	reqDto.Balance = balance

	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletSync: json Marshal failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(priKey)
	if err != nil {
		log.Println("WalletSync: parsing private key failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	log.Println("WalletSync: request data is - ", string(reqDtoJson))
	// 4. Make Balance wallet call
	respDto, err = WalletCall(reqDtoJson, *rsaPriKey, baseUrl+"betrequest", ReqTimeout)
	if err != nil {
		log.Println("WalletSync: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, err
	}
	respBytes, err := json.Marshal(respDto)
	log.Println("WalletSync: response data for userId "+userId+" is - ", string(respBytes))
	// 5. Return response
	return respDto, nil
}

func WalletUpdateBet(ubs models.UserBetStatusDto, betDto sportsdto.BetDto, sessDto sessiondto.B2BSessionDto, priKey string) (operatordto.OperatorRespDto, error) {
	// 0. Create default response object
	respDto := operatordto.OperatorRespDto{}
	respDto.Balance = 0
	respDto.Status = "" // generic error
	// 1. Construct bet request object
	reqDto := operatordto.UpdateBetReqDto{}
	reqDto.Status = "RS_OK"
	reqDto.ErrorDescription = ""
	if ubs.Status == "FAILED" {
		reqDto.Status = "RS_ERROR"
		reqDto.ErrorDescription = ubs.ErrorMessage
	}
	reqDto.BetUpdate = operatordto.BetUpdate{}
	reqDto.BetUpdate.Token = sessDto.Token
	reqDto.BetUpdate.BetId = betDto.BetId
	timeT := time.Unix(betDto.CreatedAt/1000, 0)
	reqDto.BetUpdate.BetPlacedTime = timeT.Format(time.RFC3339Nano)
	reqDto.BetUpdate.BetStatus = betDto.Status
	reqDto.BetUpdate.BetType = betDto.BetDetails.BetType
	reqDto.BetUpdate.SportId = betDto.BetDetails.SportName
	reqDto.BetUpdate.EventId = betDto.EventId
	reqDto.BetUpdate.MarketId = betDto.MarketId
	reqDto.BetUpdate.MarketName = betDto.BetDetails.MarketName
	reqDto.BetUpdate.MarketType = betDto.BetDetails.MarketType
	reqDto.BetUpdate.RunnerId = betDto.BetDetails.RunnerId
	reqDto.BetUpdate.RunnerName = betDto.BetDetails.RunnerName
	reqDto.BetUpdate.OddValue = betDto.BetDetails.OddValue
	reqDto.BetUpdate.StakeAmount = betDto.BetDetails.StakeAmount
	if betDto.BetReq.Rate != 0 {
		reqDto.BetUpdate.StakeAmount = betDto.BetDetails.StakeAmount / float64(betDto.BetReq.Rate)
	}
	reqDto.BetUpdate.SessionOutcome = betDto.BetDetails.SessionOutcome
	reqDto.BetUpdate.OddsHistory = []operatordto.OddsData{}
	for _, odds := range betDto.OddsHistory {
		oddsData := operatordto.OddsData{}
		oddsData.OddsKey = odds.OddsKey
		oddsData.OddsAt = odds.OddsAt
		oddsData.OddsValue = odds.OddsValue
		reqDto.BetUpdate.OddsHistory = append(reqDto.BetUpdate.OddsHistory, oddsData)
	}
	sizeMatched := betDto.BetDetails.StakeAmount
	var sizeRemaining float64 = 0
	if betDto.BetReq.SizePlaced != 0 {
		sizeMatched = betDto.BetReq.SizeMatched * float64(10)
		sizeMatched = (sizeMatched * 100) / (100 - betDto.BetReq.PlatformHold)
		sizeMatched = (sizeMatched * 100) / (100 - betDto.BetReq.OperatorHold)
		sizeRemaining = betDto.BetReq.SizeRemaining * float64(10)
		sizeRemaining = (sizeRemaining * 100) / (100 - betDto.BetReq.PlatformHold)
		sizeRemaining = (sizeRemaining * 100) / (100 - betDto.BetReq.OperatorHold)
	}
	reqDto.BetUpdate.SizeMatched = sizeMatched
	reqDto.BetUpdate.SizeRemaining = sizeRemaining
	if betDto.BetReq.Rate != 0 {
		reqDto.BetUpdate.SizeMatched = sizeMatched / float64(betDto.BetReq.Rate)
		reqDto.BetUpdate.SizeRemaining = sizeRemaining / float64(betDto.BetReq.Rate)
	}
	reqDto.BetUpdate.MatchedOddValue = betDto.BetReq.OddsMatched
	reqDto.BetUpdate.IsUnmatched = false
	if reqDto.BetUpdate.SizeRemaining != 0 {
		reqDto.BetUpdate.IsUnmatched = true
	}
	// 2. Create JSON payload
	reqDtoJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("WalletUpdateBet: json Marshal failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: json marshal failed!")
	}
	// 3. Create RSA Private key from PEM String
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(priKey)
	if err != nil {
		log.Println("WalletUpdateBet: parsing private key failed with error - ", err.Error())
		return respDto, fmt.Errorf("Internal Error: parsing private key failed!")
	}
	log.Println("WalletUpdateBet: request data is - ", string(reqDtoJson))
	// 4. Make Balance wallet call
	respDto, err = WalletCall(reqDtoJson, *rsaPriKey, sessDto.BaseURL+"update-user-bet", ReqTimeout)
	if err != nil {
		log.Println("WalletUpdateBet: WalletCall failed with error - ", err.Error())
		// TODO: Handle error code
		return respDto, err
	}
	respBytes, err := json.Marshal(respDto)
	log.Println("WalletUpdateBet: response data for BetId "+reqDto.BetUpdate.BetId+" is - ", string(respBytes))
	// 5. Return response
	return respDto, nil
}

func MarketResult(providerId, sportId, eventId, marketId, marketType, runnerId, runnerName string, sessionOutcome float64) error {
	// TODO: Call Operators Market Result
	logKey := "MarketResult: " + providerId + ": "
	eventKey := providerId + "-" + sportId + "-" + eventId
	marketKey := eventKey + "-" + marketId
	log.Println(logKey+"for marketKey - ", marketKey)
	// 0. Create default response object
	respDto := operatordto.MarketResultRespDto{}
	respDto.Status = "ERROR" // generic error
	respDto.ErrorCode = "INTERNAL_ERROR"
	// 1. Construct result request object
	req := operatordto.MarketResultReqDto{}
	req.ProviderId = providerId
	req.SportId = sportId
	req.EventId = eventId
	req.MarketId = marketId
	req.MarketType = marketType
	req.Result = runnerId
	req.ResultName = runnerName
	req.SessionPrice = sessionOutcome
	if marketType == constants.SAP.MarketType.FANCY() {
		req.MarketId = runnerId
	}

	objectMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println(logKey+"cache.GetObjectMap failed with error - ", err.Error())
		return fmt.Errorf("Internal Error - Failed to get Operators from Object Cache!")
	}
	opmrCount := 0
	for _, object := range objectMap {
		operator := object.(operatordto.OperatorDTO)
		if operator.MarketResult {
			if operator.BetFairPlus == true && req.ProviderId == constants.SAP.ProviderType.Dream() {
				if req.MarketType != constants.SAP.MarketType.MATCH_ODDS() {
					log.Println(logKey+"Changing ProviderId for Operator - ", operator.OperatorId, marketKey)
					req.ProviderId = constants.SAP.ProviderType.BetFair()
					// 	log.Println(logKey+"Skiping MATCH_ODDS Market Result to Operator - ", operator.OperatorId, marketKey)
					// continue
				}
			}
			log.Println(logKey+"Sending Market Result to Operator - ", operator.OperatorId, operator.BaseURL)
			opmrCount++
			// 2. Create JSON payload
			reqDtoJson, err := json.Marshal(req)
			if err != nil {
				log.Println(logKey+"json Marshal failed with error - ", err.Error())
				return fmt.Errorf("Internal Error: json marshal failed!")
			}
			// 3. Create RSA Private key from PEM String
			rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(operator.Keys.PrivateKey)
			if err != nil {
				log.Println(logKey+"parsing private key failed with error - ", err.Error())
				return fmt.Errorf("Internal Error: parsing private key failed!")
			}
			// 4. Make Result wallet call
			log.Println(logKey+"request data is - ", string(reqDtoJson))
			respBytes, err := WalletCall2(reqDtoJson, *rsaPriKey, operator.BaseURL+"market-result", ReqTimeout)
			if err != nil {
				log.Println(logKey+"WalletCall failed with error - ", err.Error())
				// TODO: Handle error code
				return err
			}
			// 5. Convert response body to Operator Response object
			log.Println(logKey+"response data is - ", string(respBytes))
			err = json.Unmarshal([]byte(respBytes), &respDto)
			if err != nil {
				log.Println(logKey+"json.Unmarshal failed with error - ", err.Error())
				log.Println(string(respBytes))
			}
		}
	}
	if opmrCount == 0 {
		log.Println(logKey + "MarketResult ZERO operators for receving result!!!")
		return nil
	}
	log.Println(logKey+"MarketResult operators count is - ", opmrCount)
	// 6. Return response
	if respDto.Status != "RS_OK" {
		log.Println(logKey+"MarketResult failed with Status - ", respDto.Status)
		log.Println(logKey+"MarketResult failed with error - ", respDto.ErrorCode)
	}
	return nil
}

func MarketRollback(providerId, sportId, eventId, marketId, marketType, marketName, rollbackType, reason string) error {
	// TODO: Call Operators Market Result
	logKey := "MarketRollback: " + providerId + ": "
	eventKey := providerId + "-" + sportId + "-" + eventId
	marketKey := eventKey + "-" + marketId
	log.Println(logKey+"for marketKey - ", marketKey)
	// 0. Create default response object
	respDto := operatordto.MarketRollbackRespDto{}
	respDto.Status = "ERROR" // generic error
	respDto.ErrorCode = "INTERNAL_ERROR"
	// 1. Construct result request object
	req := operatordto.MarketRollbackReqDto{}
	req.ProviderId = providerId
	req.SportId = sportId
	req.EventId = eventId
	req.MarketId = marketId
	req.MarketName = marketName
	req.MarketType = marketType
	req.Reason = reason
	req.RollbackType = rollbackType

	objectMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println(logKey+"cache.GetObjectMap failed with error - ", err.Error())
		return fmt.Errorf("Internal Error - Failed to get Operators from Object Cache!")
	}
	for _, object := range objectMap {
		operator := object.(operatordto.OperatorDTO)
		if operator.MarketResult {
			if operator.BetFairPlus == true {
				if req.MarketType == constants.SAP.MarketType.MATCH_ODDS() {
					log.Println(logKey+"Skiping MATCH_ODDS Market Result to Operator - ", operator.OperatorId, marketKey)
					continue
				}
				log.Println(logKey+"Changing ProviderId for Operator - ", operator.OperatorId, marketKey)
				req.ProviderId = constants.SAP.ProviderType.BetFair()
			}
			log.Println(logKey+"Sending Market Rollback to Operator - ", operator.OperatorId)
			log.Println(logKey+"Operator Base URL is - ", operator.BaseURL)
			// 2. Create JSON payload
			reqDtoJson, err := json.Marshal(req)
			if err != nil {
				log.Println(logKey+"json Marshal failed with error - ", err.Error())
				return fmt.Errorf("Internal Error: json marshal failed!")
			}
			// 3. Create RSA Private key from PEM String
			rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(operator.Keys.PrivateKey)
			if err != nil {
				log.Println(logKey+"parsing private key failed with error - ", err.Error())
				return fmt.Errorf("Internal Error: parsing private key failed!")
			}
			// 4. Make Result wallet call
			log.Println(logKey+"request data is - ", string(reqDtoJson))
			respBytes, err := WalletCall2(reqDtoJson, *rsaPriKey, operator.BaseURL+"market-rollback", ReqTimeout)
			if err != nil {
				log.Println(logKey+"WalletCall failed with error - ", err.Error())
				// TODO: Handle error code
				return err
			}
			// 5. Convert response body to Operator Response object
			log.Println(logKey+"response data is - ", string(respBytes))
			err = json.Unmarshal([]byte(respBytes), &respDto)
			if err != nil {
				log.Println(logKey+"json.Unmarshal failed with error - ", err.Error())
				log.Println(string(respBytes))
			}
		}
	}
	// 6. Return response
	if respDto.Status != "RS_OK" {
		log.Println(logKey+"MarketRollback failed with Status - ", respDto.Status)
		log.Println(logKey+"MarketRollback failed with error - ", respDto.ErrorCode)
	}
	return nil
}
