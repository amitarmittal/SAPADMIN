package handler

import (
	"Sp/cache"
	"Sp/common/function"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	dreamdto "Sp/dto/providers/dream"
	"Sp/dto/reports"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	sessdto "Sp/dto/session"
	sportsdto "Sp/dto/sports"
	"Sp/operator"
	"encoding/json"
	"strings"
	"time"

	//coresvc "Sp/handler/core"
	"Sp/providers"
	"Sp/providers/betfair"
	"Sp/providers/dream"
	"Sp/providers/sportradar"
	utils "Sp/utilities"
	"fmt"
	"log"

	"github.com/google/uuid"
)

var (
	FutureOddsAfter int = 5000 // ms
)

func GetActiveProviders(operatorId string, partnerId string) ([]dto.ProviderInfo, error) {
	providerInfos := []dto.ProviderInfo{}
	partnerStatus, err := cache.GetOpPartnerStatus(operatorId, partnerId)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetActiveProviders: Failed with error - ", err.Error())
		return providerInfos, err
	}
	for _, ps := range partnerStatus {
		if ps.OperatorStatus != "ACTIVE" {
			continue
		}
		if ps.ProviderStatus != "ACTIVE" {
			continue
		}
		providerInfo := dto.ProviderInfo{}
		providerInfo.ProviderId = ps.ProviderId
		providerInfo.ProviderName = ps.ProviderName
		providerInfos = append(providerInfos, providerInfo)
	}
	return providerInfos, nil
}

func PlaceBet(reqDto requestdto.PlaceBetReqDto) (string, error) {
	// I. Request Check
	// II. User Bet Status
	// III. Odds Limit
	// IV. Bet Limits
	reqJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("PlaceBet: reqBody json.Marshal failed with error - ", err.Error())
	}
	betId := ""
	// 1. Request Validation
	sessionDto, operatorDto, partnerDto, err := PlaceBetRequestCheck(reqDto)
	if err != nil {
		// 1.1. Return Error
		log.Println("PlaceBet: PlaceBetRequestCheck failed with error - ", err.Error())
		log.Println("PlaceBet: reqDto json is - ", string(reqJson))
		return betId, err
	}
	// 2. Get UserBetStatus
	ubs, err := database.GetUserBetStatus(reqDto.OperatorId, sessionDto.UserId)
	if err != nil {
		// 3.2. Ignore error and continue
		log.Println("PlaceBet: database.GetUserBetStatus failed with error - ", err.Error())
		// 3.3. update ubs document with user details
		ubs.UserKey = sessionDto.UserKey
		ubs.OperatorId = reqDto.OperatorId
		ubs.UserId = sessionDto.UserId
	}
	if ubs.Status == "PENDING" {
		// 3.1. Previous bet is in-progress, return Error
		log.Println("PlaceBet: Previous bet is in-progress!")
		return betId, fmt.Errorf("Previous bet is in-progress!")
	}
	// if ubs.Status == "FAILED" && ubs.ErrorMessage != "Odds Changed!" {
	// 	// 3.1. Previous bet is in-progress, return Error
	// 	log.Println("PlaceBet: Failed with ubs.ErrorMessage - ", ubs.ErrorMessage)
	// 	return betId, fmt.Errorf(ubs.ErrorMessage)
	// }
	// Mixed Provider - PlaceBet Handling
	if operatorDto.BetFairPlus == true {
		switch reqDto.MarketType {
		case constants.SAP.MarketType.BOOKMAKER(), constants.SAP.MarketType.FANCY():
			reqDto.ProviderId = constants.SAP.ProviderType.Dream()
		default:
		}
	}
	oddValue := utils.GetOddsFactor(reqDto.OddValue, reqDto.MarketType)
	if oddValue <= 1 {
		log.Println("PlaceBet: Invalid Odd Value - ", oddValue, reqDto.OddValue)
		return betId, fmt.Errorf("Invalid odd value: %.2f ", reqDto.OddValue)
	}
	// Get Config
	config := GetConfig(operatorDto, reqDto)
	configJson, err := json.Marshal(config)
	log.Println("PlaceBet: opConfig JSON is - ", operatorDto.OperatorId, reqDto.PartnerId, reqDto.SportId, reqDto.CompetitionId, reqDto.EventId, reqDto.MarketType, string(configJson))
	// 3. Odds Limit
	if config.OddsLimit > 0 {
		if oddValue > float64(config.OddsLimit) {
			// 5.1. Return Error
			log.Println("PlaceBet: Bet Odds are too high - ", reqDto.OddValue, config.OddsLimit)
			return betId, fmt.Errorf("Bet not accepted. Rate is too high: %.2f ", reqDto.OddValue)
		}
	}
	// 4. Bet Limits (#VI)
	stakeAmount := reqDto.StakeAmount * float64(partnerDto.Rate)
	if config.MinBetValue > 0 && float64(config.MinBetValue) > stakeAmount {
		log.Println("PlaceBet: Bet Value is too low - ", stakeAmount)
		log.Println("PlaceBet: Minimum Value is - ", config.MinBetValue)
		return betId, fmt.Errorf("Stake is too LOW!")
	}
	if config.MaxBetValue > 0 && float64(config.MaxBetValue) < stakeAmount {
		log.Println("PlaceBet: Bet Value is too high - ", stakeAmount)
		log.Println("PlaceBet: Maximum Value is - ", config.MaxBetValue)
		return betId, fmt.Errorf("Stake is too HIGH!")
	}
	// 7. Insert result document if not present, async
	//go InsertMarket(reqDto)
	// Get OddsData - Function to call L2 Validate Odds based on Provider
	oddsValue, err := GetOdds(reqDto)
	if err != nil {
		log.Println("PlaceBet: GetOdds failed with error - ", err.Error())
	} else {
		// 3. Odds Limit check with BEFORE odds
		if config.OddsLimit > 0 {
			oddValue := utils.GetOddsFactor(oddsValue, reqDto.MarketType)
			if oddValue > float64(config.OddsLimit) {
				// 5.1. Return Error
				log.Println("PlaceBet: Current Odds are beyond allowed limit - ", oddValue, config.OddsLimit)
				return betId, fmt.Errorf("Current Odds are beyond allowed limit: %.2f ", oddValue)
			}
		}
	}
	oddsData := sportsdto.OddsData{}
	oddsData.OddsKey = "BEFORE"
	oddsData.OddsValue = oddsValue
	oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
	// Create a betDto
	betDto := GetBetDto(reqDto, sessionDto, partnerDto)
	betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
	log.Println("PlaceBet: OddsData BEFORE for betId - ", betDto.BetId)
	// Update UBS
	// ubs.Status = "PENDING"
	ubs.ReferenceId = betDto.BetId
	ubs.ReqTime = time.Now().UnixNano() / 1000000
	// database.UpsertUserBetStatus(ubs)
	UpdateUserBetStatus(ubs, "PENDING", "", operatorDto, sessionDto, betDto)
	// Process Bet Asynchromously
	// go ProcessBet(reqDto, operatorDto, sessionDto, betDto, ubs, config)
	// 9. Provider specific logic
	switch reqDto.ProviderId {
	case providers.DREAM_SPORT:
		// 9.1. Dream - Bet Placement
		go DreamProcessBet(reqDto, operatorDto, sessionDto, betDto, ubs, config)
	case providers.BETFAIR:
		// 9.1. BetFair - Bet Placement
		go BetFairProcessBet(reqDto, operatorDto, sessionDto, betDto, ubs, config)
	case providers.SPORT_RADAR:
		// 9.1. SportRadar - Bet Placement
		go SportRadarProcessBet(reqDto, operatorDto, sessionDto, betDto, ubs, config)
	default:
		return betId, fmt.Errorf("Internal Error - Invalid ProviderId - " + reqDto.ProviderId)
	}
	// Send success response
	return betDto.BetId, nil
}

func PlaceBetRequestCheck(reqDto requestdto.PlaceBetReqDto) (sessdto.B2BSessionDto, operatordto.OperatorDTO, operatordto.Partner, error) {
	// 0. Session Status
	// 1. Operator Status
	// 2. Partner Status
	// 3. Provider Status
	// 4. PartnerStatus ProviderStatus & OperatorStatus
	// 5. Sport Status
	// 6. SportStatus ProviderStatus & OperatorStatus
	// 7. Competition Status
	// 8. CompetitionStatus ProviderStatus & OperatorStatus
	// 9. Event Status
	// 10. EventStatus ProviderStatus & OperatorStatus
	// 11. Market Status
	// 12. MarketStatus ProviderStatus & OperatorStatus
	sessionDto := sessdto.B2BSessionDto{}
	operatorDto := operatordto.OperatorDTO{}
	partnerDto := operatordto.Partner{}
	// 0. Session Status
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 1.1. Return Error
		log.Println("PlaceBet: function.GetSession failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Session Expired")
	}
	// 1. Operator Status
	if reqDto.OperatorId == "" {
		log.Println("PlaceBet: OperatorId is empty in request!")
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Operator!")
	}
	operatorDto, err = cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		log.Println("PlaceBet: cache.GetOperatorDetails failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("PlaceBet: Operator is not active - ", operatorDto.Status)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Operator is not active!")
	}
	if operatorDto.BetLock == true {
		log.Println("PlaceBet: BetLock is TRUE for operator - ", operatorDto.OperatorId)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting not allowed. Please contact upline!")
	}
	// 2. Partner Status
	if reqDto.PartnerId == "" {
		log.Println("PlaceBet: PartnerId is empty in request!")
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Partner!")
	}
	isFound := false
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == reqDto.PartnerId {
			if partner.Status != "ACTIVE" {
				log.Println("PlaceBet: Partner is not active - ", partner.Status)
				return sessionDto, operatorDto, partnerDto, fmt.Errorf("Partner is not active!")
			}
			isFound = true
			partnerDto = partner
			break
		}
	}
	if isFound == false {
		log.Println("PlaceBet: PartnerId is invalid - ", reqDto.PartnerId)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Partner!")
	}
	// 3. Provider Status
	if reqDto.ProviderId == "" {
		log.Println("PlaceBet: ProviderId is empty in request!")
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Provider!")
	}
	// Mixed Provider - PlaceBet Handling
	if operatorDto.BetFairPlus == true {
		switch reqDto.MarketType {
		case constants.SAP.MarketType.BOOKMAKER(), constants.SAP.MarketType.FANCY():
			reqDto.ProviderId = constants.SAP.ProviderType.Dream()
		default:
		}
	}
	provider, err := cache.GetProvider(reqDto.ProviderId)
	if err != nil {
		log.Println("PlaceBet: cache.GetProvider failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
	}
	if provider.Status != "ACTIVE" {
		log.Println("PlaceBet: provider.Status is not ACTIVE - ", provider.Status)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this provider!!!")
	}
	// 4. PartnerStatus ProviderStatus & OperatorStatus
	partnerStatus, err := cache.GetPartnerStatus(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
	if err != nil {
		log.Println("PlaceBet: cache.GetPartnerStatus failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
	}
	if partnerStatus.ProviderStatus != "ACTIVE" {
		log.Println("PlaceBet: partnerStatus.ProviderStatus is not ACTIVE - ", partnerStatus.ProviderStatus)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this provider!!!")
	}
	if partnerStatus.OperatorStatus != "ACTIVE" {
		log.Println("PlaceBet: partnerStatus.OperatorStatus is not ACTIVE - ", partnerStatus.OperatorStatus)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this provider!!!")
	}
	// 5. Sport Status
	if reqDto.SportId == "" {
		log.Println("PlaceBet: SportId is empty in request!")
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Sport!")
	}
	sportKey := reqDto.ProviderId + "-" + reqDto.SportId
	sport, err := cache.GetSport(sportKey)
	if err != nil {
		log.Println("PlaceBet: cache.GetSport failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
	}
	if sport.Status != "ACTIVE" {
		log.Println("PlaceBet: sport.Status is not ACTIVE - ", sport.Status)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this sport!!!")
	}
	// 6. SportStatus ProviderStatus & OperatorStatus
	sportStatus, err := cache.GetSportStatus(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId, reqDto.SportId)
	if err != nil {
		log.Println("PlaceBet: cache.GetSportStatus failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
	}
	if sportStatus.ProviderStatus != "ACTIVE" {
		log.Println("PlaceBet: sportStatus.ProviderStatus is not ACTIVE - ", sportStatus.ProviderStatus)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this sport!!!")
	}
	if sportStatus.OperatorStatus != "ACTIVE" {
		log.Println("PlaceBet: sportStatus.OperatorStatus is not ACTIVE - ", sportStatus.OperatorStatus)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this sport!!!")
	}
	// 7. Competition Status
	if reqDto.CompetitionId == "" {
		log.Println("PlaceBet: CompetitionId is empty in request!")
		//return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Competition!")
	} else {
		competition, err := cache.GetCompetition(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
		if err != nil {
			log.Println("PlaceBet: cache.GetCompetition failed with error - ", err.Error())
			//return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
		} else {
			if competition.Status != "ACTIVE" {
				log.Println("PlaceBet: competition.Status is not ACTIVE - ", competition.Status)
				return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this competition!!!")
			}
		}
		// 8. CompetitionStatus ProviderStatus & OperatorStatus
		competitionStatus, err := cache.GetCompetitionStatus(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
		if err == nil {
			if competitionStatus.ProviderStatus != "ACTIVE" {
				log.Println("PlaceBet: competitionStatus.ProviderStatus is not ACTIVE - ", competitionStatus.ProviderStatus)
				return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this competition!!!")
			}
			if competitionStatus.OperatorStatus != "ACTIVE" {
				log.Println("PlaceBet: competitionStatus.OperatorStatus is not ACTIVE - ", competitionStatus.OperatorStatus)
				return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this competition!!!")
			}
			// log.Println("PlaceBet: cache.GetCompetitionStatus failed with error - ", err.Error())
			// return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
		}
	}
	// 9. Event Status
	if reqDto.EventId == "" {
		log.Println("PlaceBet: EventId is empty in request!")
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Event!")
	}
	event, err := cache.GetEvent(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
	if err != nil {
		log.Println("PlaceBet: cache.GetEvent failed with error - ", err.Error())
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
	}
	if event.Status != "ACTIVE" {
		log.Println("PlaceBet: event.Status is not ACTIVE - ", event.Status)
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this event!!!")
	}
	// 10. EventStatus => ProviderStatus & OperatorStatus
	eventStatus, err := cache.GetEventStatus(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
	if err == nil {
		if eventStatus.ProviderStatus != "ACTIVE" {
			log.Println("PlaceBet: eventStatus.ProviderStatus is not ACTIVE - ", eventStatus.ProviderStatus)
			return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this event!!!")
		}
		if eventStatus.OperatorStatus != "ACTIVE" {
			log.Println("PlaceBet: eventStatus.OperatorStatus is not ACTIVE - ", eventStatus.OperatorStatus)
			return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this event!!!")
		}
	}
	// 11. Market Status
	if reqDto.MarketId == "" {
		log.Println("PlaceBet: MarketId is empty in request!")
		return sessionDto, operatorDto, partnerDto, fmt.Errorf("Invalid Market!")
	}
	market, err := cache.GetMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	if err == nil {
		if market.Status != "ACTIVE" {
			log.Println("PlaceBet: market.Status is not ACTIVE - ", market.Status)
			return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this market!!!")
		}
		// 12. MarketStatus ProviderStatus & OperatorStatus
		marketStatus, err := cache.GetMarketStatus(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
		if err == nil {
			if marketStatus.ProviderStatus != "ACTIVE" {
				log.Println("PlaceBet: marketStatus.ProviderStatus is not ACTIVE - ", marketStatus.ProviderStatus)
				return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this market!!!")
			}
			if marketStatus.OperatorStatus != "ACTIVE" {
				log.Println("PlaceBet: marketStatus.OperatorStatus is not ACTIVE - ", marketStatus.OperatorStatus)
				return sessionDto, operatorDto, partnerDto, fmt.Errorf("Betting is not allowed on this market!!!")
			}
		}
	} else {
		if err.Error() != "Market NOT FOUND!" {
			log.Println("PlaceBet: cache.GetMarket failed with error - ", err.Error())
			return sessionDto, operatorDto, partnerDto, fmt.Errorf("Internal Error!")
		}
		// insert market
		err = InsertMarket(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId, reqDto.MarketType)
		if err != nil {
			log.Println("PlaceBet: InsertMarket failed with error - ", err.Error())
		}
	}
	return sessionDto, operatorDto, partnerDto, nil
}

func DreamProcessBet(reqDto requestdto.PlaceBetReqDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto, betDto sportsdto.BetDto, ubs models.UserBetStatusDto, config providers.Config) {
	// 1. Check Balance - Only for Transfer Wallet
	// 2. Bet Delay
	// 3. Validate Odds
	// 4. Wallet Bet - Only for Seamless
	// 5. Accept Bet
	var err error
	// 1. Check Balance - Only for Transfer Wallet
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	// 2. Bet Delay
	BetDelay(config)
	// 3. Validate Odds
	betDto, err = DreamValidateOdds(reqDto, betDto)
	if err != nil {
		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		return
	}
	// 4. Wallet Bet - Only for Seamless
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		err = WalletBet(operatorDto, sessionDto, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	// 4. Check Balance - Only for Transfer Wallet
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		// 2.1. Transfer wallet
		_, err := providers.PlaceBet_Transfer(betDto)
		if err != nil {
			log.Println("PlaceBet: DreamProcessBet: PlaceBet_Transfer failed with error - ", err.Error())
		}
	}
	// 5. Accept Bet
	AcceptBet(betDto, ubs, operatorDto, sessionDto)
	AddFutureOddsData(reqDto, betDto)
	return
}

func BetFairProcessBet(reqDto requestdto.PlaceBetReqDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto, betDto sportsdto.BetDto, ubs models.UserBetStatusDto, config providers.Config) {
	// 1. Check Balance - Only for Transfer Wallet
	// 2. Bet Delay
	// 3. Apply Config
	// 4. Min Value Check
	// 		4.1. ValidateOdds
	// 		4.2. WalletBet - Only for Seamless
	//		4.3. OperatorLedger
	// 		4.4. AcceptBet
	// 5. WalletBet - Only for Seamless
	// 6. OperatorLedger
	// 7. InitiateBet
	// 8. PlaceOrder
	// 9. Error
	//		9.1. RollbackBet
	//			Compute Rollback
	//			Operator Ledger
	//			Operator Delta
	//			WalletRollback - Only for Seamless - Handle Retry mechanism - TODO
	//			UpdateBet
	// 10. WalletUpdate - Only for Seamless - TODO
	// 11. UpdateBet

	var err error
	var minBetValue float64 = 0.0
	prObj, err := cache.GetObject(constants.SAP.ObjectTypes.PROVIDER(), constants.SAP.ProviderType.BetFair())
	if err == nil {
		provider := prObj.(models.Provider)
		minBetValue = provider.MinBetValue
		log.Println("PlaceBet: BetFairProcessBet: provider.MinBetValue - ", provider.MinBetValue, provider.ProviderId)
	} else {
		log.Println("PlaceBet: BetFairProcessBet: cache.GetObject failed with error - ", err.Error(), constants.SAP.ObjectTypes.PROVIDER(), constants.SAP.ProviderType.BetFair())
	}
	// 1. Check Balance - Only for Transfer Wallet
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	// 2. Bet Delay
	BetDelay(config)
	// 3. Apply Config
	betDto, reqDto, err = BetFairApplyConfig(reqDto, betDto, operatorDto, config)
	if err != nil {
		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		return
	}
	// 4. Min Value Check
	reqDto.StakeAmount = utils.Truncate4Decfloat64(reqDto.StakeAmount / float64(betfair.BetFairRate))
	log.Println("PlaceBet: BetFairProcessBet: reqDto.StakeAmount - ", reqDto.StakeAmount)
	log.Println("PlaceBet: BetFairProcessBet: MinBetValue is - ", minBetValue)
	if reqDto.StakeAmount < float64(minBetValue) {
		// 4.1. Validate Odds
		betDto, err = BetFairValidateOdds(reqDto, betDto)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
		// 4.2. Wallet Bet - Only for Seamless
		if strings.ToLower(operatorDto.WalletType) == "seamless" {
			err = WalletBet(operatorDto, sessionDto, betDto, ubs)
			if err != nil {
				UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
				return
			}
		}
		// 4.2. Check Balance - Only for Transfer Wallet
		if strings.ToLower(operatorDto.WalletType) == "transfer" {
			err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
			if err != nil {
				UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
				return
			}
		}
		if strings.ToLower(operatorDto.WalletType) == "transfer" {
			// 2.1. Transfer wallet
			_, err := providers.PlaceBet_Transfer(betDto)
			if err != nil {
				log.Println("PlaceBet: BetFairProcessBet: PlaceBet_Transfer failed with error - ", err.Error())
			}
		}
		// 4.3. Operator Ledger
		OperatorLedger(operatorDto, betDto)
		// 4.4. Accept Bet
		AcceptBet(betDto, ubs, operatorDto, sessionDto)
		// 4.5: Wallet Update Bet call to operator
		// if operatorDto.BetUpdates == true {
		// 	WalletUpdateBet(operatorDto, sessionDto, betDto)
		// }
		AddFutureOddsData(reqDto, betDto)
		return
	}
	// 5. Wallet Bet - Only for Seamless
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		err = WalletBet(operatorDto, sessionDto, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	// 5. Check Balance - Only for Transfer Wallet
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		// 2.1. Transfer wallet
		_, err := providers.PlaceBet_Transfer(betDto)
		if err != nil {
			log.Println("PlaceBet: BetFairProcessBet: PlaceBet_Transfer failed with error - ", err.Error())
		}
	}
	// 6. Operator Ledger
	OperatorLedger(operatorDto, betDto)
	if err != nil {
		log.Println("PlaceBet: BetFairProcessBet: OperatorLedger failed with error - ", err.Error())
		//return
	}
	// 7. Initate Bet
	betDto.Status = "INITIATED"
	err = InitiateBet(betDto, ubs)
	if err != nil {
		log.Println("PlaceBet: BetFairProcessBet: InitiateBet failed with error - ", err.Error())
		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		return
	}
	// 8. PlaceOrder
	betDto, err = BetFairPlaceOrder(reqDto, betDto, ubs)
	if err != nil {
		// 8.1. RollbackBet
		RollbackBet(operatorDto, betDto, ubs)
		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		return
	}
	// 9. WalletUpdate - conditional
	// if strings.ToLower(operatorDto.WalletType) == "seamless" {
	// 	// NOTES: Not needed. OpenBets endpoint will return updated bets
	// }
	// 10. Update Bet
	UpdateBet(betDto, ubs, betDto.Status)
	UpdateUserBetStatus(ubs, "COMPLETED", "", operatorDto, sessionDto, betDto)
	// 4.5: Wallet Update Bet call to operator
	// if operatorDto.BetUpdates == true {
	// 	WalletUpdateBet(operatorDto, sessionDto, betDto)
	// }
	AddFutureOddsData(reqDto, betDto)
	return
}

func SportRadarProcessBet(reqDto requestdto.PlaceBetReqDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto, betDto sportsdto.BetDto, ubs models.UserBetStatusDto, config providers.Config) {
	// 1. Check Balance - Only for Transfer Wallet
	// 2. Bet Delay
	// 3. Apply Config
	// 4. Min Value Check
	// 		4.1. ValidateOdds
	// 		4.2. WalletBet - Only for Seamless
	//		4.3. OperatorLedger
	// 		4.4. AcceptBet
	// 5. WalletBet - Only for Seamless
	// 6. OperatorLedger
	// 7. InitiateBet
	// 8. PlaceOrder
	// 9. Error
	//		9.1. RollbackBet
	//			Compute Rollback
	//			Operator Ledger
	//			Operator Delta
	//			WalletRollback - Only for Seamless - Handle Retry mechanism - TODO
	//			UpdateBet
	// 10. WalletUpdate - Only for Seamless - TODO
	// 11. UpdateBet

	var err error
	// 1. Check Balance - Only for Transfer Wallet
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	// 2. Bet Delay
	BetDelay(config)
	// 3. Apply Config
	betDto, reqDto, err = SportRadarApplyConfig(reqDto, betDto, operatorDto, config)
	if err != nil {
		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		return
	}
	// 4. Min Value Check
	reqDto.StakeAmount = utils.Truncate4Decfloat64(reqDto.StakeAmount / float64(sportradar.SportRadarRate))
	log.Println("PlaceBet: SportRadarProcessBet: reqDto.StakeAmount 2 - ", reqDto.StakeAmount)
	log.Println("PlaceBet: SportRadarProcessBet: MinBetValue is 2 - ", sportradar.MinBetValue)
	if reqDto.StakeAmount < sportradar.MinBetValue {
		// Return Error
		log.Println("PlaceBet: SportRadarProcessBet: Bet amount is too LOW!!! Stake, MinBetValue - ", reqDto.StakeAmount, sportradar.MinBetValue)
		UpdateUserBetStatus(ubs, "FAILED", "Bet amount is too LOW!!!", operatorDto, sessionDto, betDto)
		return
		// 4.1. Validate Odds
		// log.Println("PlaceBet: SportRadarProcessBet: reqDto.OddValue - ", reqDto.OddValue)
		// betDto, err = SportRadarValidateOdds(reqDto, betDto)
		// if err != nil {
		// 	UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		// 	return
		// }
		// // 4.2. Wallet Bet - Only for Seamless
		// if strings.ToLower(operatorDto.WalletType) == "seamless" {
		// 	err = WalletBet(operatorDto, sessionDto, betDto, ubs)
		// 	if err != nil {
		// 		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		// 		return
		// 	}
		// }
		// // 4.2. Check Balance - Only for Transfer Wallet
		// if strings.ToLower(operatorDto.WalletType) == "transfer" {
		// 	err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		// 	if err != nil {
		// 		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		// 		return
		// 	}
		// }
		// if strings.ToLower(operatorDto.WalletType) == "transfer" {
		// 	// 2.1. Transfer wallet
		// 	_, err := providers.PlaceBet_Transfer(betDto)
		// 	if err != nil {
		// 		log.Println("PlaceBet: SportRadarProcessBet: PlaceBet_Transfer failed with error - ", err.Error())
		// 	}
		// }
		// // 4.3. Operator Ledger
		// // OperatorLedger(operatorDto, betDto)
		// // 4.4. Accept Bet
		// AcceptBet(betDto, ubs, operatorDto, sessionDto)
		// AddFutureOddsData(reqDto, betDto)
		// return
	}
	// 5. Wallet Bet - Only for Seamless
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		err = WalletBet(operatorDto, sessionDto, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
		// // 6. Operator Ledger
		// OperatorLedger(operatorDto, betDto)
		// if err != nil {
		// 	log.Println("PlaceBet: BetFairProcessBet: OperatorLedger failed with error - ", err.Error())
		// 	//return
		// }
		// 7. Initate Bet
		betDto.Status = "INITIATED"
		err = InitiateBet(betDto, ubs)
		if err != nil {
			log.Println("PlaceBet: SportRadarProcessBet: InitiateBet failed with error - ", err.Error())
			RollbackBet(operatorDto, betDto, ubs)
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			// Duplicate call: UpdateUserBetStatus(ubs, "FAILED", err.Error())
			return
		}
	}
	// 5. Check Balance - Only for Transfer Wallet
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		err = CheckBalance(operatorDto.OperatorId, sessionDto.UserId, betDto, ubs)
		if err != nil {
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
	}
	// 8. PlaceOrder
	betDto, err = SportRadarPlaceOrder(reqDto, betDto)
	if err != nil {
		// 8.1. RollbackBet
		RollbackBet(operatorDto, betDto, ubs)
		UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
		// Duplicate call: UpdateUserBetStatus(ubs, "FAILED", err.Error())
		return
	}
	if strings.ToLower(operatorDto.WalletType) == "transfer" {
		// 2.1. Transfer wallet
		_, err := providers.PlaceBet_Transfer(betDto)
		if err != nil {
			log.Println("PlaceBet: SportRadarProcessBet: PlaceBet_Transfer failed with error - ", err.Error())
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			return
		}
		// 7. Initate Bet
		betDto.Status = "OPEN"
		err = InitiateBet(betDto, ubs)
		if err != nil {
			log.Println("PlaceBet: SportRadarProcessBet: InitiateBet failed with error - ", err.Error())
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			//return
		}
	}
	if strings.ToLower(operatorDto.WalletType) == "feed" {
		// 7. Initate Bet
		betDto.Status = "OPEN"
		err = InitiateBet(betDto, ubs)
		if err != nil {
			log.Println("PlaceBet: SportRadarProcessBet: InitiateBet failed with error - ", err.Error())
			UpdateUserBetStatus(ubs, "FAILED", err.Error(), operatorDto, sessionDto, betDto)
			//return
		}
	}
	// 9. WalletUpdate - conditional
	// if strings.ToLower(operatorDto.WalletType) == "seamless" {
	// 	// NOTES: Not needed. OpenBets endpoint will return updated bets
	// }
	// 10. Update Bet
	UpdateBet(betDto, ubs, betDto.Status)
	UpdateUserBetStatus(ubs, "COMPLETED", "", operatorDto, sessionDto, betDto)
	AddFutureOddsData(reqDto, betDto)
	return
}

func GetOdds(reqDto requestdto.PlaceBetReqDto) (float64, error) {
	var oddsValue float64
	oddsResp := dreamdto.ValidateOddsRespDto{}
	var err error = fmt.Errorf("Default Error")
	switch reqDto.ProviderId {
	case providers.DREAM_SPORT:
		// 9.1. Dream - Bet Placement
		oddsResp, err = dream.ValidateOdds(reqDto)
		if err != nil {
			return oddsValue, err
		}
	case providers.BETFAIR:
		// 9.1. BetFair - Bet Placement
		oddsResp, err = betfair.ValidateOdds(reqDto)
		if err != nil {
			return oddsValue, err
		}
	case providers.SPORT_RADAR:
		// 9.1. SportRadar - Bet Placement
		oddsResp, err = sportradar.ValidateOdds(reqDto)
		if err != nil {
			return oddsValue, err
		}
	default:
		return oddsValue, fmt.Errorf("Internal Error - Invalid ProviderId - " + reqDto.ProviderId)
	}
	if oddsResp.IsValid {
		oddsValue = oddsResp.MatchedOddValue
	} else {
		if len(oddsResp.OddValues) > 0 {
			oddsValue = oddsResp.OddValues[0]
		} else {
			oddsValue = reqDto.OddValue
		}
	}
	return oddsValue, nil
}

func AddFutureOddsData(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto) {
	// Delay for FUTURE OddsData
	time.Sleep(time.Duration(FutureOddsAfter * int(time.Millisecond)))
	// Get OddsData - Function to call L2 Validate Odds based on Provider
	oddsValue, err := GetOdds(reqDto)
	if err != nil {
		log.Println("PlaceBet: AddFutureOddsData: GetOdds failed with error - ", err.Error())
	}
	oddsData := sportsdto.OddsData{}
	oddsData.OddsKey = "AFTER"
	oddsData.OddsValue = oddsValue
	oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
	// Create a betDto
	betDto, err = database.GetBetDetails(betDto.BetId)
	if err != nil {
		log.Println("PlaceBet: AddFutureOddsData: database.GetBetDetails failed with error - ", err.Error())
		return
	}
	betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
	log.Println("PlaceBet: OddsData FUTURE for betId - ", betDto.BetId)
	database.UpdateBet(betDto)
	if err != nil {
		log.Println("PlaceBet: AddFutureOddsData: database.UpdateBet failed with error - ", err.Error())
	}
}

func CheckBalance(operatorId string, userId string, betDto sportsdto.BetDto, ubs models.UserBetStatusDto) error {
	user := providers.GetB2BUser(operatorId, userId)
	// Check User Balance is greater than the betvalue (#O) - Error
	if betDto.BetReq.DebitAmount > user.Balance {
		// Return Error
		log.Println("PlaceBet: CheckBalance: User Balance is too low - ", user.Balance)
		log.Println("PlaceBet: CheckBalance: Bet Value is - ", betDto.BetReq.DebitAmount)
		// ubs.ErrorMessage = "Low user balance!!!"
		// database.UpsertUserBetStatus(ubs)
		// UpdateUserBetStatus(ubs, ubs.Status, "Low user balance!!!")
		betDto.Status = "User - Insufficient Funds!"
		database.InsertFailedBet(betDto)
		return fmt.Errorf("Low user balance!!!")
	}
	return nil
}

func BetDelay(config providers.Config) {
	// Bet Delay (#VII)
	if config.BetDelay > 0 {
		log.Println("PlaceBet: BetDelay: 1. config.BetDelay is - ", config.BetDelay)
		time.Sleep(time.Duration(int(config.BetDelay) * int(time.Millisecond)))
		log.Println("PlaceBet: BetDelay: 2. Delay ended")
	} else {
		log.Println("PlaceBet: BetDelay: 0. ZERO config.BetDelay - ", config.BetDelay)
	}
}

func BetFairApplyConfig(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto, operatorDto operatordto.OperatorDTO, opConfig providers.Config) (sportsdto.BetDto, requestdto.PlaceBetReqDto, error) {
	// 1. Apply Operator Hold (#III) on Stake & Debit Amounts
	log.Println("PlaceBet: BetFairApplyConfig: User Bet Value - ", reqDto.StakeAmount)
	log.Println("PlaceBet: BetFairApplyConfig: User Risk - ", betDto.BetReq.DebitAmount)
	betDto.BetReq.OperatorHold = opConfig.Hold
	betDto.BetReq.OperatorAmount = utils.Truncate4Decfloat64(betDto.BetReq.DebitAmount * (100 - opConfig.Hold) / 100)
	log.Println("PlaceBet: BetFairApplyConfig: Operator Risk - ", betDto.BetReq.OperatorAmount)
	// 2. Check Operator Balance is greater than the betvalue (#IV) - Error
	operator := providers.GetOperator(operatorDto.OperatorId)
	if betDto.BetReq.OperatorAmount > operator.Balance {
		log.Println("PlaceBet: BetFairApplyConfig: Operator Balance is low - ", operator.Balance)
		return betDto, reqDto, fmt.Errorf("Operator - Insufficient Funds!")
	}
	// 3. Apply Platform Hold (#V) on Stake & Debit Amounts
	sportKey := reqDto.ProviderId + "-" + reqDto.SportId
	provider, _ := cache.GetProvider(reqDto.ProviderId)                                             // log error
	sport, _ := cache.GetSport(sportKey)                                                            // log error
	competition, _ := cache.GetCompetition(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId) // log error
	event, _ := cache.GetEvent(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)                   // log error
	market, _ := cache.GetMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	sapConfig, level := providers.GetSapConfig(reqDto.MarketType, market, event, competition, sport, operatorDto, provider) // log error
	configJson, _ := json.Marshal(sapConfig)                                                                                // log error
	log.Println("PlaceBet: BetFairApplyConfig: sapConfig JSON is - ", level, string(configJson))
	betDto.BetReq.PlatformHold = sapConfig.Hold
	betDto.BetReq.PlatformAmount = utils.Truncate1Decfloat64(betDto.BetReq.OperatorAmount * (100 - sapConfig.Hold) / 100)
	log.Println("PlaceBet: BetFairApplyConfig: Platform Risk - ", betDto.BetReq.PlatformAmount)
	reqDto.StakeAmount = utils.Truncate4Decfloat64(betDto.BetDetails.StakeAmount * (100 - opConfig.Hold) / 100)
	reqDto.StakeAmount = utils.Truncate1Decfloat64(reqDto.StakeAmount * (100 - sapConfig.Hold) / 100)
	return betDto, reqDto, nil
}

func SportRadarApplyConfig(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto, operatorDto operatordto.OperatorDTO, opConfig providers.Config) (sportsdto.BetDto, requestdto.PlaceBetReqDto, error) {
	// 1. Apply Operator Hold (#III) on Stake & Debit Amounts
	log.Println("PlaceBet: SportRadarApplyConfig: User Bet Value - ", reqDto.StakeAmount)
	log.Println("PlaceBet: SportRadarApplyConfig: User Risk - ", betDto.BetReq.DebitAmount)
	betDto.BetReq.OperatorHold = 0 //opConfig.Hold
	betDto.BetReq.OperatorAmount = utils.Truncate4Decfloat64(betDto.BetReq.DebitAmount * (100 - betDto.BetReq.OperatorHold) / 100)
	log.Println("PlaceBet: SportRadarApplyConfig: Operator Risk - ", betDto.BetReq.OperatorAmount)
	// 3. Apply Platform Hold (#V) on Stake & Debit Amounts
	sportKey := reqDto.ProviderId + "-" + reqDto.SportId
	provider, _ := cache.GetProvider(reqDto.ProviderId)                                             // log error
	sport, _ := cache.GetSport(sportKey)                                                            // log error
	competition, _ := cache.GetCompetition(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId) // log error
	event, _ := cache.GetEvent(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)                   // log error
	market, _ := cache.GetMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	sapConfig, level := providers.GetSapConfig(reqDto.MarketType, market, event, competition, sport, operatorDto, provider) // log error
	configJson, _ := json.Marshal(sapConfig)                                                                                // log error
	log.Println("PlaceBet: SportRadarApplyConfig: sapConfig JSON is - ", level, string(configJson))
	betDto.BetReq.PlatformHold = 0 //sapConfig.Hold
	betDto.BetReq.PlatformAmount = utils.Truncate4Decfloat64(betDto.BetReq.OperatorAmount * (100 - betDto.BetReq.PlatformHold) / 100)
	log.Println("PlaceBet: SportRadarApplyConfig: Platform Risk - ", betDto.BetReq.PlatformAmount)
	reqDto.StakeAmount = utils.Truncate4Decfloat64(betDto.BetDetails.StakeAmount * (100 - betDto.BetReq.OperatorHold) / 100)
	reqDto.StakeAmount = utils.Truncate4Decfloat64(reqDto.StakeAmount * (100 - betDto.BetReq.PlatformHold) / 100)
	return betDto, reqDto, nil
}

func DreamValidateOdds(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto) (sportsdto.BetDto, error) {
	// Validate Odds
	respObj, err := dream.ValidateOdds(reqDto)
	if err != nil {
		// 1.1. Failed to validate odds
		log.Println("PlaceBet: Dream Odds Validation call failed with - ", err.Error())
		return betDto, fmt.Errorf(respObj.ErrorDescription)
	}
	//log.Println("PlaceBet: Event Open Date is - ", time.Unix(respObj.OpenDate/1000, 0))
	if !respObj.IsValid {
		// 1.2. Invalid Odds
		log.Println("PlaceBet: Invalid Odds - ", reqDto.OddValue)
		return betDto, fmt.Errorf("Odds Changed!")
	}
	// 2. Is match live or two days from current date (#II)
	if reqDto.MarketType == "MATCH_ODDS" && respObj.Status != "IN_PLAY" {
		// 2.1. Match Odds bet on upcoming match
		// TODO: Advanced Betting for Dream
		log.Println("PlaceBet: MATCH_ODDS bet only allowed on IN_PLAY events - ", respObj.Status)
		return betDto, fmt.Errorf("Betting NOT ALLOWED NOT IN_PLAY!")
	}
	if respObj.Status != "IN_PLAY" && time.Now().Add(time.Hour*24*2).Unix() < (respObj.OpenDate/1000) {
		// 2.2. Upcoming match & more than 48 hours to start
		log.Println("PlaceBet: TOO EARLY TO BET - ", time.Unix(respObj.OpenDate/1000, 0))
		return betDto, fmt.Errorf("Betting NOT ALLOWED NOT IN_PLAY!")
	}
	betDto.Status = constants.SAP.BetStatus.OPEN()
	log.Println("PlaceBet: DreamValidateOdds: SportRadar oddvalue - Requsted at OddValue - ", betDto.BetDetails.OddValue)
	log.Println("PlaceBet: DreamValidateOdds: SportRadar oddvalue - Matched  at OddValue - ", respObj.MatchedOddValue)
	if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respObj.MatchedOddValue > betDto.BetDetails.OddValue {
		betDto.BetDetails.OddValue = respObj.MatchedOddValue
	}
	if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respObj.MatchedOddValue < betDto.BetDetails.OddValue {
		betDto.BetDetails.OddValue = respObj.MatchedOddValue
		oddValue := utils.GetOddsFactor(betDto.BetDetails.OddValue, betDto.BetDetails.MarketType)
		betDto.BetReq.DebitAmount = utils.Truncate4Decfloat64((oddValue - 1) * betDto.BetDetails.StakeAmount)
		betDto.NetAmount = betDto.BetReq.DebitAmount * -1
	}
	// Add OddsData for CURRENT
	oddsData := sportsdto.OddsData{}
	oddsData.OddsKey = "CURRENT"
	oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
	oddsData.OddsValue = respObj.MatchedOddValue
	betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
	log.Println("PlaceBet: OddsData CURRENT for betId - ", betDto.BetId)
	return betDto, nil
}

func BetFairValidateOdds(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto) (sportsdto.BetDto, error) {
	// Validate Odds
	respObj, err := betfair.ValidateOdds(reqDto)
	if err != nil {
		// 1.1. Failed to validate odds
		log.Println("PlaceBet: BetFairValidateOdds: Odds Validation call failed with - ", err.Error())
		return betDto, fmt.Errorf("Odds Changed!")
	}
	respJson, _ := json.Marshal(respObj)
	//log.Println("PlaceBet: Event Open Date is - ", time.Unix(respObj.OpenDate/1000, 0))
	if !respObj.IsValid {
		// 1.2. Invalid Odds
		log.Println("PlaceBet: BetFairValidateOdds: Invalid Odds - ", reqDto.OddValue)
		log.Println("PlaceBet: BetFairValidateOdds: ValidateOdds response json is - ", string(respJson))
		return betDto, fmt.Errorf("Odds Changed!")
	}
	/* */
	// 2. Check for acceptance criteria
	if respObj.Status != "IN_PLAY" {
		// 2.1. Match Odds bet on upcoming match
		// TODO: add to collection
		// TODO: Allow advanced betting
		log.Println("PlaceBet: BetFairValidateOdds:MATCH_ODDS bet only allowed on IN_PLAY events - ", respObj.Status)
		log.Println("PlaceBet: BetFairValidateOdds: ValidateOdds response json is - ", string(respJson))
		return betDto, fmt.Errorf("Betting NOT ALLOWED!")
	}
	betDto.Status = constants.SAP.BetStatus.OPEN()
	// if betDto.BetDetails.MarketType == constants.SAP.MarketType.LINE_ODDS() {
	// 	log.Println("PlaceBet: BetFairValidateOdds: BetFair oddvalue - Requsted at SessionOutcome - ", betDto.BetDetails.SessionOutcome)
	// 	log.Println("PlaceBet: BetFairValidateOdds: BetFair oddvalue - Matched  at SessionOutcome - ", respObj.MatchedOddValue)
	// 	if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respObj.MatchedOddValue > betDto.BetDetails.SessionOutcome {
	// 		betDto.BetReq.OddsMatched = respObj.MatchedOddValue
	// 	}
	// 	if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respObj.MatchedOddValue < betDto.BetDetails.SessionOutcome {
	// 		betDto.BetReq.OddsMatched = respObj.MatchedOddValue
	// 	}
	// } else {
	// 	log.Println("PlaceBet: BetFairValidateOdds: BetFair oddvalue - Requsted at OddValue - ", betDto.BetDetails.OddValue)
	// 	log.Println("PlaceBet: BetFairValidateOdds: BetFair oddvalue - Matched  at OddValue - ", respObj.MatchedOddValue)
	// 	if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respObj.MatchedOddValue > betDto.BetDetails.OddValue {
	// 		betDto.BetReq.OddsMatched = respObj.MatchedOddValue
	// 	}
	// 	if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respObj.MatchedOddValue < betDto.BetDetails.OddValue {
	// 		betDto.BetReq.OddsMatched = respObj.MatchedOddValue
	// 	}
	// }

	// Add OddsData for CURRENT
	oddsData := sportsdto.OddsData{}
	oddsData.OddsKey = "CURRENT"
	oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
	oddsData.OddsValue = respObj.MatchedOddValue
	betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
	log.Println("PlaceBet: OddsData CURRENT for betId - ", betDto.BetId)
	// bet Sizes
	betDto.BetReq.SizePlaced = reqDto.StakeAmount
	betDto.BetReq.SizeMatched = reqDto.StakeAmount
	betDto.BetReq.SizeRemaining = 0
	betDto.BetReq.SizeLapsed = 0
	betDto.BetReq.SizeCancelled = 0
	betDto.BetReq.SizeVoided = 0
	betDto.BetReq.OddsMatched = respObj.MatchedOddValue
	return betDto, nil
}

func SportRadarValidateOdds(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto) (sportsdto.BetDto, error) {
	respObj, err := sportradar.ValidateOdds(reqDto)
	if err != nil {
		// 1.1. Failed to validate odds
		log.Println("PlaceBet: SportRadarValidateOdds: Odds Validation call failed with - ", err.Error())
		return betDto, fmt.Errorf("Odds Changed!")
	}
	respJson, _ := json.Marshal(respObj)
	//log.Println("PlaceBet: Event Open Date is - ", time.Unix(respObj.OpenDate/1000, 0))
	if !respObj.IsValid {
		// 1.2. Invalid Odds
		log.Println("PlaceBet: SportRadarValidateOdds: Invalid Odds - ", reqDto.OddValue)
		log.Println("PlaceBet: SportRadarValidateOdds: ValidateOdds response json is - ", string(respJson))
		return betDto, fmt.Errorf("Odds Changed!")
	}
	/* */
	// 2. Check for acceptance criteria
	if respObj.Status != "IN_PLAY" {
		// 2.1. Match Odds bet on upcoming match
		// TODO: add to collection
		log.Println("PlaceBet: SportRadarValidateOdds: MATCH_ODDS bet only allowed on IN_PLAY events - ", respObj.Status)
		log.Println("PlaceBet: SportRadarValidateOdds: ValidateOdds response json is - ", string(respJson))
		//return betDto, fmt.Errorf("Betting NOT ALLOWED!")
	}
	betDto.Status = constants.SAP.BetStatus.OPEN()
	log.Println("PlaceBet: SportRadarValidateOdds: SportRadar oddvalue - Requsted at OddValue - ", betDto.BetDetails.OddValue)
	log.Println("PlaceBet: SportRadarValidateOdds: SportRadar oddvalue - Matched  at OddValue - ", respObj.MatchedOddValue)
	if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respObj.MatchedOddValue > betDto.BetDetails.OddValue {
		betDto.BetDetails.OddValue = respObj.MatchedOddValue
	}
	if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respObj.MatchedOddValue < betDto.BetDetails.OddValue {
		betDto.BetDetails.OddValue = respObj.MatchedOddValue
	}
	// Add OddsData for CURRENT
	oddsData := sportsdto.OddsData{}
	oddsData.OddsKey = "CURRENT"
	oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
	oddsData.OddsValue = respObj.MatchedOddValue
	betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
	log.Println("PlaceBet: OddsData CURRENT for betId - ", betDto.BetId)
	return betDto, nil
}

func WalletBet(operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto, betDto sportsdto.BetDto, ubs models.UserBetStatusDto) error {
	// Wallet Bet
	_, err := providers.PlaceBet_Seamless(betDto, operatorDto, sessionDto)
	if err != nil {
		log.Println("PlaceBet: WalletBet: PlaceBet_Seamless failed with error - ", err.Error())
		// ubs.ErrorMessage = err.Error()
		// database.UpsertUserBetStatus(ubs)
		UpdateUserBetStatus(ubs, ubs.Status, err.Error(), operatorDto, sessionDto, betDto)
		betDto.Status = err.Error()
		database.InsertFailedBet(betDto)
		return err
	}
	return nil
}

func WalletUpdateBet(ubs models.UserBetStatusDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto, betDto sportsdto.BetDto) error {
	// Wallet Bet
	_, err := operator.WalletUpdateBet(ubs, betDto, sessionDto, operatorDto.Keys.PrivateKey)
	if err != nil {
		log.Println("PlaceBet: WalletUpdateBet: operator.WalletUpdateBet failed with error - ", err.Error())
		//ubs.ErrorMessage = err.Error()
		//database.UpsertUserBetStatus(ubs)
		//betDto.Status = err.Error()
		//database.InsertFailedBet(betDto)
		//return err
	}
	return nil
}

func AcceptBet(betDto sportsdto.BetDto, ubs models.UserBetStatusDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto) error {
	// 11. Save Bet in db (#IV)
	err := database.InsertBetDetails(betDto)
	if err != nil {
		log.Println("PlaceBet: AcceptBet: Insert bet into db failed with error - ", err.Error())
		// ubs.ErrorMessage = "Internal Error!!!"
		// database.UpsertUserBetStatus(ubs) // Log error
		UpdateUserBetStatus(ubs, ubs.Status, "Internal Error!!!", operatorDto, sessionDto, betDto)
		betDto.Status = "INSERT Bet FAILED!"
		database.InsertFailedBet(betDto) // Log error
		//return err
	}
	// ubs.Status = "COMPLETED"
	// ubs.ErrorMessage = ""
	// time.Sleep(10 * time.Millisecond) // added 10ms sleep to make sure previous user bet status update will be completed
	// database.UpsertUserBetStatus(ubs) // Log error
	UpdateUserBetStatus(ubs, "COMPLETED", "", operatorDto, sessionDto, betDto)
	log.Println("PlaceBet: AcceptBet: Completed Successfully")
	return nil
}

func UpdateUserBetStatus(ubs models.UserBetStatusDto, status string, errMsg string, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto, betDto sportsdto.BetDto) {
	ubs.Status = status
	ubs.ErrorMessage = errMsg
	time.Sleep(10 * time.Millisecond) // added 10ms sleep to make sure previous user bet status update will be completed
	database.UpsertUserBetStatus(ubs) // Log error
	// TODO: Call WalletUpdateBet if status is FAILED or COMPLETED
	if operatorDto.BetUpdates == true {
		if status == "FAILED" {
			betDto.Status = status
			WalletUpdateBet(ubs, operatorDto, sessionDto, betDto)
		}
		if status == "COMPLETED" {
			//betDto.Status = "OPEN"
			WalletUpdateBet(ubs, operatorDto, sessionDto, betDto)
		}
	}
}

func InitiateBet(betDto sportsdto.BetDto, ubs models.UserBetStatusDto) error {
	// 11. Save Bet in db (#IV)
	err := database.InsertBetDetails(betDto)
	if err != nil { // TODO: add retry mechanism
		log.Println("PlaceBet: InitiateBet: Insert bet into db failed with error - ", err.Error())
		// ubs.ErrorMessage = "Internal Error!!!"
		// database.UpsertUserBetStatus(ubs) // Log error
		// UpdateUserBetStatus(ubs, ubs.Status, "Internal Error!!!")
		betDto.Status = "INSERT Bet FAILED!"
		database.InsertFailedBet(betDto) // Log error
		return err
	}
	return nil
}

func UpdateBet(betDto sportsdto.BetDto, ubs models.UserBetStatusDto, status string) error {
	// 11. Save Bet in db (#IV)
	betDto.Status = status
	err := database.UpdateBet(betDto)
	if err != nil {
		log.Println("PlaceBet: UpdateBet: database.UpdateBet failed with error - ", err.Error())
	}
	// ubs.Status = "COMPLETED"
	// ubs.ErrorMessage = ""
	// time.Sleep(10 * time.Millisecond) // added 10ms sleep to make sure previous user bet status update will be completed
	// database.UpsertUserBetStatus(ubs) // Log error
	log.Println("PlaceBet: UpdateBet: Completed Successfully")
	return nil
}

func BetFairPlaceOrder(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto, ubs models.UserBetStatusDto) (sportsdto.BetDto, error) {
	// 3. BetFair - Place Order
	log.Println("PlaceBet: BetFairPlaceOrder: Before")
	respDto, err := betfair.PlaceOrder(reqDto, betDto)
	if err != nil {
		log.Println("PlaceBet: BetFairPlaceOrder: PlaceOrder failed with error - ", err.Error())
		return betDto, err
	}
	log.Println("PlaceBet: BetFairPlaceOrder: PlaceOrder Status is - ", respDto.Status)
	if respDto.Status != "SUCCESS" {
		// API Call failed. Mark bet status as failed. and return
		log.Println("PlaceBet: BetFairPlaceOrder: PlaceOrder InstructionReport Status - ", respDto.Status)
		log.Println("PlaceBet: BetFairPlaceOrder: PlaceOrder InstructionReport ErrorCode is - ", respDto.ErrorCode)
		log.Println("PlaceBet: BetFairPlaceOrder: PlaceOrder InstructionReport OrderStatus is - ", respDto.OrderStatus)
		return betDto, fmt.Errorf(respDto.ErrorCode)
	}
	// 4. Check operator wallet type
	switch respDto.OrderStatus {
	case "EXECUTION_COMPLETE":
		betDto.Status = "OPEN"
		betDto.BetReq.BetId = respDto.BetId
		// if betDto.BetDetails.MarketType == constants.SAP.MarketType.LINE_ODDS() {
		// 	log.Println("PlaceBet: BetFairPlaceOrder: BetFair oddvalue - Requsted at SessionOutcome - ", reqDto.SessionOutcome)
		// 	log.Println("PlaceBet: BetFairPlaceOrder: BetFair oddvalue - Matched  at SessionOutcome - ", respDto.AveragePriceMatched)
		// 	if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respDto.AveragePriceMatched > betDto.BetDetails.SessionOutcome {
		// 		betDto.BetReq.OddsMatched = respDto.AveragePriceMatched
		// 	}
		// 	if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respDto.AveragePriceMatched < betDto.BetDetails.SessionOutcome {
		// 		betDto.BetReq.OddsMatched = respDto.AveragePriceMatched
		// 	}
		// } else {
		// 	log.Println("PlaceBet: BetFairPlaceOrder: BetFair oddvalue - Requsted at OddValue - ", reqDto.OddValue)
		// 	log.Println("PlaceBet: BetFairPlaceOrder: BetFair oddvalue - Matched  at OddValue - ", respDto.AveragePriceMatched)
		// 	if betDto.BetDetails.BetType == constants.BetFair.Side.BACK() && respDto.AveragePriceMatched > betDto.BetDetails.OddValue {
		// 		betDto.BetReq.OddsMatched = respDto.AveragePriceMatched
		// 	}
		// 	if betDto.BetDetails.BetType == constants.BetFair.Side.LAY() && respDto.AveragePriceMatched < betDto.BetDetails.OddValue {
		// 		betDto.BetReq.OddsMatched = respDto.AveragePriceMatched
		// 	}
		// }
		// Add OddsData for CURRENT
		oddsData := sportsdto.OddsData{}
		oddsData.OddsKey = "CURRENT"
		oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
		oddsData.OddsValue = betDto.BetDetails.OddValue
		betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
		log.Println("PlaceBet: OddsData CURRENT for betId - ", betDto.BetId)
		// bet Sizes
		betDto.BetReq.SizePlaced = reqDto.StakeAmount
		betDto.BetReq.SizeMatched = reqDto.StakeAmount
		betDto.BetReq.SizeRemaining = 0
		betDto.BetReq.SizeLapsed = 0
		betDto.BetReq.SizeCancelled = 0
		betDto.BetReq.SizeVoided = 0
		betDto.BetReq.OddsMatched = respDto.AveragePriceMatched
	case "EXECUTABLE":
		betDto.Status = "UNMATCHED"
		betDto.BetReq.BetId = respDto.BetId
		// Add OddsData for CURRENT
		oddsData := sportsdto.OddsData{}
		oddsData.OddsKey = "CURRENT"
		oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
		if respDto.SizeMatched > 0 {
			oddsData.OddsValue = respDto.AveragePriceMatched
		} else {
			oddsValue, _ := GetOdds(reqDto)
			oddsData.OddsValue = oddsValue
		}
		betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
		log.Println("PlaceBet: OddsData CURRENT for betId - ", betDto.BetId)
		// bet Sizes
		betDto.BetReq.SizePlaced = reqDto.StakeAmount
		betDto.BetReq.SizeMatched = respDto.SizeMatched
		betDto.BetReq.SizeRemaining = betDto.BetReq.SizePlaced - betDto.BetReq.SizeMatched
		betDto.BetReq.SizeLapsed = 0
		betDto.BetReq.SizeCancelled = 0
		betDto.BetReq.SizeVoided = 0
		if respDto.AveragePriceMatched > 0 {
			betDto.BetReq.OddsMatched = respDto.AveragePriceMatched
		} else {
			betDto.BetReq.OddsMatched = betDto.BetDetails.OddValue
		}
		if betDto.BetReq.SizeRemaining != 0 && betDto.BetReq.SizeMatched != 0 {
			log.Println("PlaceBet: BetFairPlaceOrder: PartialMatched Bet Found for BetFair betId - ", betDto.BetReq.BetId)
			log.Println("PlaceBet: BetFairPlaceOrder: PartialMatched Bet SizeMatched - ", betDto.BetReq.SizeMatched)
			log.Println("PlaceBet: BetFairPlaceOrder: PartialMatched Bet SizeRemaining - ", betDto.BetReq.SizeRemaining)
		}
	case "EXPIRED":
		log.Println("PlaceBet: BetFairPlaceOrder: Unexpected PlaceOrder InstructionReport BetId is - ", respDto.BetId)
		return betDto, fmt.Errorf("FAILED!")
	default:
		log.Println("PlaceBet: BetFairPlaceOrder: Unexpected PlaceOrder InstructionReport OrderStatus is - ", respDto.OrderStatus)
		log.Println("PlaceBet: BetFairPlaceOrder: Unexpected PlaceOrder InstructionReport BetId is - ", respDto.BetId)
		return betDto, fmt.Errorf(respDto.OrderStatus)
	}
	return betDto, nil
}

func SportRadarPlaceOrder(reqDto requestdto.PlaceBetReqDto, betDto sportsdto.BetDto) (sportsdto.BetDto, error) {
	log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder Before")
	respDto, err := sportradar.PlaceOrder(reqDto, betDto)
	if err != nil {
		log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder failed with error - ", err.Error())
		return betDto, err
	}
	log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder Status is - ", respDto.Status)
	if respDto.Status != "RS_OK" {
		// API Call failed. Mark bet status as failed. and return
		log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder API Status is - ", respDto.Status)
		log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder API ErrorCode is - ", respDto.ErrorDescription)
		log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder API AltStake is - ", respDto.AltStake)
		respDto.AltStake = respDto.AltStake * float64(sportradar.SportRadarRate)
		log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder API AltStake is - ", respDto.AltStake)
		respDto.AltStake = respDto.AltStake / float64(betDto.BetReq.Rate)
		log.Println("PlaceBet: SportRadarPlaceOrder: PlaceOrder API AltStake is - ", respDto.AltStake)
		if respDto.AltStake > 0 {
			return betDto, fmt.Errorf("Alternate Stake is - %f", respDto.AltStake)
		}
		return betDto, fmt.Errorf("SportRadar returned failure - " + respDto.Status)
	}
	// Add OddsData for CURRENT
	oddsData := sportsdto.OddsData{}
	oddsData.OddsKey = "CURRENT"
	oddsData.OddsAt = time.Now().Format(time.RFC3339Nano)
	oddsData.OddsValue = betDto.BetDetails.OddValue
	betDto.OddsHistory = append(betDto.OddsHistory, oddsData)
	log.Println("PlaceBet: OddsData CURRENT for betId - ", betDto.BetId)
	betDto.Status = "OPEN"
	return betDto, nil
}

func OperatorLedger(operator operatordto.OperatorDTO, betDto sportsdto.BetDto) error {
	// 7. Bet Success, Add to operator ledger
	err := providers.OperatorLedgerTx(operator, constants.SAP.LedgerTxType.BETPLACEMENT(), betDto.BetReq.OperatorAmount*-1, betDto.BetId)
	if err != nil {
		log.Println("PlaceBet: OperatorLedger: providers.OperatorLedgerTx failed with error - ", err.Error())
		return err
	}
	return nil
}

func RollbackBet(operatorDto operatordto.OperatorDTO, betDto sportsdto.BetDto, ubs models.UserBetStatusDto) error {
	// 1. Add Rollback Request
	betStatus := betDto.Status
	rollbackReq := providers.ComputeRollback(betDto, constants.BetFair.OrderStatus.EXPIRED())
	log.Println("PlaceBet: RollbackBet: BetId is : ", betDto.BetId)
	log.Println("PlaceBet: RollbackBet: rollbackReq.RollbackAmount is : ", rollbackReq.RollbackAmount)
	betDto.RollbackReqs = append(betDto.RollbackReqs, rollbackReq)
	betDto.NetAmount += rollbackReq.RollbackAmount
	betDto.Status = constants.BetFair.OrderStatus.EXPIRED()
	betDto.UpdatedAt = rollbackReq.ReqTime
	// 2. & 3. Add OperatorLedgerTx & Update Operator Balance
	err := providers.OperatorLedgerTx(operatorDto, constants.SAP.LedgerTxType.BETCANCEL(), rollbackReq.OperatorAmount, betDto.BetId)
	if err != nil {
		log.Println("PlaceBet: RollbackBet: providers.OperatorLedgerTx failed with error - ", err.Error())
		//return err
	}
	// 4. Wallet Rollback - Only for Seamless
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		opResp, rollBackReq, err := operator.WalletRollback(constants.BetFair.OrderStatus.EXPIRED(), betDto, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("PlaceBet: RollbackBet: operator.WalletRollback failed with error - ", err.Error())
			log.Println("PlaceBet: RollbackBet: operator.WalletRollback failed for betId - ", betDto.BetId)
			betDto.Status = betDto.Status + "-failed"
			log.Println("PlaceBet: RollbackBet: operator.WalletRollback Rollback Request is - ", rollBackReq)
		}
		if opResp.Status != "RS_OK" {
			log.Println("PlaceBet: RollbackBet: operator.WalletRollback failed. Status is - ", opResp.Status)
			betDto.Status = betDto.Status + "-failed"
		}
		log.Println("PlaceBet: RollbackBet: Rollback Successfully completed for betId - ", betDto.BetId)
	}
	if operatorDto.WalletType == constants.SAP.WalletType.Feed() {
		log.Println("PlaceBet: RollbackBet: Feed wallet Rollback " + betDto.BetId + " betStatus - " + betStatus)
		// if betStatus == constants.SAP.BetStatus.INPROCESS() {
		if betStatus != "INITIATED" {
			opResp, rollBackReq, err := operator.WalletRollback(constants.BetFair.OrderStatus.EXPIRED(), betDto, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
			if err != nil {
				log.Println("PlaceBet: RollbackBet: operator.WalletRollback failed with error - ", err.Error())
				log.Println("PlaceBet: RollbackBet: operator.WalletRollback failed for betId - ", betDto.BetId)
				betDto.Status = betDto.Status + "-failed"
				log.Println("PlaceBet: RollbackBet: operator.WalletRollback Rollback Request is - ", rollBackReq)
			}
			if opResp.Status != "RS_OK" {
				log.Println("PlaceBet: RollbackBet: operator.WalletRollback failed. Status is - ", opResp.Status)
				betDto.Status = betDto.Status + "-failed"
			}
			log.Println("PlaceBet: RollbackBet: Rollback Successfully completed for betId - ", betDto.BetId)
		}
	}
	if operatorDto.WalletType == constants.SAP.WalletType.Transfer() {
		log.Println("PlaceBet: RollbackBet: Transfer wallet Rollback " + betDto.BetId + " betStatus - " + betStatus)
		updatedBets := []sportsdto.BetDto{}
		updatedBets = append(updatedBets, betDto)
		updatedBets, err = providers.CancelBet_Transfer(updatedBets)
		if err != nil {
			log.Println("PlaceBet: RollbackBet: providers.CancelBet_Transfer failed with error for betId - ", err.Error(), betDto.BetId)
			betDto.Status = betDto.Status + "-failed"
		}
		log.Println("PlaceBet: RollbackBet: Rollback Successfully completed for betId - ", betDto.BetId)
	}
	// 5. Update Bet
	err = database.UpdateBet(betDto)
	if err != nil {
		log.Println("PlaceBet: UpdateBet: database.UpdateBet failed with error - ", err.Error())
	}
	// ubs.Status = "FAILED"
	// ubs.ErrorMessage = "LAPSED due to TIMEOUT!"
	// time.Sleep(10 * time.Millisecond) // added 10ms sleep to make sure previous user bet status update will be completed
	// database.UpsertUserBetStatus(ubs) // Log error
	// UpdateUserBetStatus(ubs, "FAILED", "LAPSED due to TIMEOUT!", operatorDto, sessionDto, betDto)
	return nil
}

func CancelBet(reqDto requestdto.CancelBetReqDto) (float64, []responsedto.CancelBetResp, error) {
	// I. Validate Session Token
	// II. Check Operator & Partner Status
	var balance float64 = 0.0
	//openBets := []responsedto.OpenBetDto{}
	CancelBetsResp := []responsedto.CancelBetResp{}
	// 1. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 1.1. Return Error
		log.Println("CancelBet: function.GetSession failed with error - ", err.Error())
		return balance, CancelBetsResp, fmt.Errorf("Session Expired")
	}
	// 2. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 2.1. Return Error
		log.Println("CancelBet: cache.GetOperatorDetails failed with error - ", err.Error())
		log.Println("CancelBet: User is - ", sessionDto.UserId)
		return balance, CancelBetsResp, fmt.Errorf("Unauthorized access, please contact support!")
	}
	if operatorDto.Status != "ACTIVE" {
		// 2.2. Return Error
		log.Println("CancelBet: Operator is not active - ", operatorDto.Status)
		return balance, CancelBetsResp, fmt.Errorf("Unauthorized access, please contact support!")
	}
	// 3. Get a betDto
	betIds := []string{}
	for _, cancelBet := range reqDto.CancelBets {
		betIds = append(betIds, cancelBet.BetId)
	}
	unmatchedBets := []sportsdto.BetDto{}
	var uBetsCount int = 0
	for {
		betDtos, err := database.GetBets(betIds) // Sort by createdAt
		if err != nil {
			// 6.1.1. Failed to cancel a bet, return error
			log.Println("CancelBet: database.GetBets failed with error - ", err.Error())
			return balance, CancelBetsResp, fmt.Errorf("Database Error!")
		}
		if len(betDtos) != len(betIds) {
			// 6.1.1. Failed to cancel a bet, return error
			log.Println("CancelBet: database.GetBets returned count is - ", len(betDtos))
			log.Println("CancelBet: NOT ALL BETS FOUND in DATABASE ", betIds)
			return balance, CancelBetsResp, fmt.Errorf("BetIds NOT FOUND IN DATABASE!")
		}
		unmatchedBets = []sportsdto.BetDto{}
		uBetsCount = 0
		var betsCount int = 0
		// Add Rollback Request to the database
		for _, betDto := range betDtos {
			//if betDto.Status == constants.SAP.BetStatus.INPROCESS() || betDto.Status == "INITIATED" {
			if betDto.Status == "INITIATED" {
				log.Println("CancelBet: Bet is in still in-process - ", betDto.BetId)
				break
			} else if betDto.Status == constants.SAP.BetStatus.CANCELLED() {
				betsCount++
				uBetsCount++
			} else if betDto.Status == constants.SAP.BetStatus.UNMATCHED() {
				// Add a rollback req to the bet
				//rollbackReq := providers.ComputeRollback(betDto, constants.BetFair.BetStatus.CANCELLED())
				//betDto.RollbackReqs = append(betDtos[i].RollbackReqs, rollbackReq)
				betDto.Status = constants.BetFair.BetStatus.CANCELLED()
				unmatchedBets = append(unmatchedBets, betDto)
				betsCount++
				uBetsCount++
			} else {
				log.Println("CancelBet: Bet is not in expexted state for betId - ", betDto.BetId)
				log.Println("CancelBet: Bet is not in expexted state - ", betDto.Status)
				betsCount++
			}
		}
		if betsCount == len(betIds) {
			break
		}
		log.Println("CancelBet: Count MISMATCH - betsCount - ", betsCount)
		time.Sleep(50 * time.Millisecond)
	}
	if len(betIds) != uBetsCount {
		log.Println("CancelBet: unmatched bets count is - ", uBetsCount)
		log.Println("CancelBet: ALL BETS MUST BE either UNMATCHED or CANCELLED state!!!")
		return balance, CancelBetsResp, fmt.Errorf("Bet(s) state changed!")
	}
	if len(unmatchedBets) == 0 {
		// return success
		log.Println("CancelBet: ZERO Bets to Cancel!!!")
		return balance, CancelBetsResp, nil
	}
	log.Println("CancelBet: unmatchedBets Bets count is - ", len(unmatchedBets))
	updatedBets := []sportsdto.BetDto{}
	// 4. Provider specific logic
	switch reqDto.ProviderId {
	case providers.BETFAIR:
		// 6.1. BetFair - Bet Cancel
		updatedBets, CancelBetsResp, err = betfair.BetFairCancelBet(reqDto, unmatchedBets, operatorDto, sessionDto)
		if err != nil {
			// 6.1.1. Failed to cancel a bet, return error
			log.Println("CancelBet: BetFair CancelBet failed with error - ", err.Error())
			return balance, CancelBetsResp, fmt.Errorf("Bet Cancellation Failed!")
		}
	default:
		log.Println("CancelBet: Invalid ProviderId - ", reqDto.ProviderId)
		return balance, CancelBetsResp, fmt.Errorf("Invalid Request!")
	}
	log.Println("CancelBet: updated Bets count is - ", len(updatedBets))
	// 5. Operator Wallet
	if len(updatedBets) > 0 {
		switch operatorDto.WalletType {
		case constants.SAP.WalletType.Seamless():
			// TODO: Communicate to operator about bets cancellation
			updatedBets, err = providers.CancelBet_Seamless(operatorDto, updatedBets)
		case constants.SAP.WalletType.Transfer():
			updatedBets, err = providers.CancelBet_Transfer(updatedBets)
		case constants.SAP.WalletType.Feed():
			log.Println("CancelBet: Feed Wallet Type - Nothing to perform here!")
		default:
			log.Println("CancelBet: Invalid Wallet Type - ", operatorDto.WalletType)
		}
	}
	if err != nil {
		// return error
		return balance, CancelBetsResp, err
	}
	log.Println("CancelBet: updated Bets count is - ", len(updatedBets))
	// 6. Update Bets
	if len(updatedBets) > 0 {
		failedCount, errMsgs := database.UpdateBets(updatedBets)
		if failedCount > 0 {
			log.Println("CancelBet: Failures count is - ", len(errMsgs))
			for _, msg := range errMsgs {
				log.Println("CancelBet: Error Message is - ", msg)
			}
		}
	} else {
		log.Println("CancelBet: UpdatedBets count is ZERO, SOMETHING WENT WRONG!")
	}
	// 7. Get Open Bets
	// time.Sleep(time.Duration(50 * time.Millisecond))
	// eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	// openBetDtos, _, err := database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
	// if err != nil {
	// 	log.Println("CancelBet: Failed with error - ", err.Error())
	// 	return balance, CancelBetsResp, nil
	// }
	// openBets = GetOpenBetsDto(openBetDtos, reqDto.SportId)
	// 8. Get User Balance from db
	userKey := reqDto.OperatorId + "-" + sessionDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println("CancelBet: User NOT FOUND: ", err.Error())
	} else {
		balance = b2bUser.Balance
	}
	// 9. Send success response
	return balance, CancelBetsResp, nil
}

func GetConfig(operatorDto operatordto.OperatorDTO, reqDto requestdto.PlaceBetReqDto) providers.Config {
	// Get all objects - marketStatus, eventStatus, competitionStatus, sportStatus, providerStatus
	ps, _ := cache.GetPartnerStatus(operatorDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
	ss, _ := cache.GetSportStatus(operatorDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId, reqDto.SportId)
	cs, _ := cache.GetCompetitionStatus(operatorDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
	es, _ := cache.GetEventStatus(operatorDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
	ms, _ := cache.GetMarketStatus(operatorDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	opConfig, _ := providers.GetOperatorConfig(reqDto.MarketType, ms, es, cs, ss, ps, operatorDto)
	return opConfig
}

func GetMarketConfig(operatorDto operatordto.OperatorDTO, partnerId string, providerId string, sportId string, competitionId string, eventId string, marketType string) providers.Config {
	// Get all objects - marketStatus, eventStatus, competitionStatus, sportStatus, providerStatus
	ps, _ := cache.GetPartnerStatus(operatorDto.OperatorId, partnerId, providerId)
	ss, _ := cache.GetSportStatus(operatorDto.OperatorId, partnerId, providerId, sportId)
	cs, _ := cache.GetCompetitionStatus(operatorDto.OperatorId, providerId, sportId, competitionId)
	es, _ := cache.GetEventStatus(operatorDto.OperatorId, providerId, sportId, eventId)
	ms := models.MarketStatus{} // dummy market status
	opConfig, _ := providers.GetOperatorConfig(marketType, ms, es, cs, ss, ps, operatorDto)
	return opConfig
}

// func GetOddsFactor(oddValue float64, marketType string) float64 {
// 	log.Println("GetOddsFactor: Data is - ", marketType, oddValue)
// 	if strings.ToUpper(marketType) == constants.SAP.MarketType.BOOKMAKER() {
// 		log.Println("GetOddsFactor: BOOKMAKER bet!!!")
// 		oddValue = 1.0 + oddValue*0.01
// 	}
// 	if strings.ToUpper(marketType) == constants.SAP.MarketType.FANCY() {
// 		log.Println("GetOddsFactor: FANCY bet!!!")
// 		oddValue = 1.0 + oddValue*100
// 	}
// 	log.Println("GetOddsFactor: oddValue is - ", oddValue)
// 	return oddValue
// }

// Can be called async
func GetMarket(operatorId, providerId, sportId, eventId, marketId, marketType string) (models.Market, error) {
	market := models.Market{}
	// Check is MarketKey exist in the Market Collection
	event := dto.EventDto{}
	switch providerId {
	case constants.SAP.ProviderType.Dream():
		event = dream.GetMarkets(sportId, eventId, operatorId)
	case constants.SAP.ProviderType.BetFair():
		event = betfair.GetMarkets(sportId, eventId, operatorId)
	case constants.SAP.ProviderType.SportRadar():
		event = sportradar.GetMarkets(sportId, eventId)
	default:
		log.Println("InsertMarket: INVALID providerId - ", providerId)
		return market, fmt.Errorf("Invalid ProviderId - " + providerId)
	}
	var err error
	eventKey := providerId + "-" + sportId + "-" + eventId
	// loop through markets
	switch marketType {
	case constants.SAP.MarketType.MATCH_ODDS():
		for _, moMarket := range event.Markets.MatchOdds {
			if moMarket.MarketId == marketId {
				marketKey := eventKey + "-" + moMarket.MarketId
				market, err = GetMOMarket(moMarket, event, marketKey)
				if err != nil {
					log.Println("InsertMarket: GetMOMarket failed with error - ", err.Error())
					return market, err
				}
				return market, nil
			}
		}
	case constants.SAP.MarketType.BOOKMAKER():
		for _, moMarket := range event.Markets.Bookmakers {
			if moMarket.MarketId == marketId {
				marketKey := eventKey + "-" + moMarket.MarketId
				market, err = GetMOMarket(moMarket, event, marketKey)
				if err != nil {
					log.Println("InsertMarket: GetMOMarket failed with error - ", err.Error())
					return market, err
				}
				return market, nil
			}
		}
	case constants.SAP.MarketType.FANCY():
		for _, fancyMarket := range event.Markets.FancyMarkets {
			if fancyMarket.MarketId == marketId {
				marketKey := eventKey + "-" + fancyMarket.MarketId
				market, err = GetFancyMarket(fancyMarket, event, marketKey)
				if err != nil {
					log.Println("InsertMarket: GetFancyMarket failed with error - ", err.Error())
					return market, err
				}
				return market, nil
			}
		}
	case constants.SAP.MarketType.LINE_ODDS():
		for _, loMarket := range event.Markets.MatchOdds {
			if loMarket.MarketId == marketId {
				marketKey := eventKey + "-" + loMarket.MarketId
				market, err = GetMOMarket(loMarket, event, marketKey)
				if err != nil {
					log.Println("InsertMarket: GetLOMarket failed with error - ", err.Error())
					return market, err
				}
				return market, nil
			}
		}
	default:
		log.Println("InsertMarket: Invalid betReq.MarketType - ", marketType)
		return market, fmt.Errorf("Invalid betReq.MarketType - " + marketType)
	}
	log.Println("InsertMarket: Invalid betReq.MarketId - ", marketId)
	return market, fmt.Errorf("Invalid betReq.MarketId - " + marketId)
}

// Can be called async
func GetMarkets(providerId, sportId, eventId string) ([]models.Market, error) {
	markets := []models.Market{}
	// Check is MarketKey exist in the Market Collection
	event := dto.EventDto{}
	switch providerId {
	case constants.SAP.ProviderType.Dream():
		event = dream.GetMarkets(sportId, eventId, "")
	case constants.SAP.ProviderType.BetFair():
		event = betfair.GetMarkets(sportId, eventId, "")
	case constants.SAP.ProviderType.SportRadar():
		event = sportradar.GetMarkets(sportId, eventId)
	default:
		log.Println("GetMarkets: INVALID providerId - ", providerId)
		return markets, fmt.Errorf("Invalid ProviderId!!!")
	}
	newMarkets := []models.Market{}
	eventKey := providerId + "-" + sportId + "-" + eventId // + "-" + event.MarketId
	dbMarkets, err := database.GetMarketsByEventKey(eventKey)
	if err != nil || len(dbMarkets) == 0 {
		// Iterate through match_odds markets in eventDto
		for _, moMarket := range event.Markets.MatchOdds {
			marketKey := eventKey + "-" + moMarket.MarketId
			market, err := GetMOMarket(moMarket, event, marketKey)
			if err != nil {
				log.Println("GetMarkets: GetMOMarket failed with error - ", err.Error())
				continue
			}
			newMarkets = append(newMarkets, market)
		}
		// Iterate through bookmaker markets in eventDto
		for _, moMarket := range event.Markets.Bookmakers {
			marketKey := eventKey + "-" + moMarket.MarketId
			market, err := GetMOMarket(moMarket, event, marketKey)
			if err != nil {
				log.Println("GetMarkets: GetMOMarket failed with error - ", err.Error())
				continue
			}
			newMarkets = append(newMarkets, market)
		}
		// Iterate through fancy markets in eventDto
		for _, fancyMarket := range event.Markets.FancyMarkets {
			marketKey := eventKey + "-" + fancyMarket.MarketId
			market, err := GetFancyMarket(fancyMarket, event, marketKey)
			if err != nil {
				log.Println("GetMarkets: GetFancyMarket failed with error - ", err.Error())
				continue
			}
			newMarkets = append(newMarkets, market)
		}
	} else {
		// Add all dbMarkets to return object
		for _, dbMarket := range dbMarkets {
			markets = append(markets, dbMarket)
		}
		// Check each market existance in DB
		// Iterate through match_odds markets in eventDto
		for _, moMarket := range event.Markets.MatchOdds {
			found := false
			for _, dbMarket := range dbMarkets { // Iterate through dbMarkets
				if dbMarket.MarketId == moMarket.MarketId {
					found = true
					break
				}
			}
			if found == true {
				continue
			}
			marketKey := eventKey + "-" + moMarket.MarketId
			market, err := GetMOMarket(moMarket, event, marketKey)
			if err != nil {
				log.Println("GetMarkets: GetMOMarket failed with error - ", err.Error())
				continue
			}
			newMarkets = append(newMarkets, market)
		}
		// Iterate through bookmaker markets in eventDto
		for _, moMarket := range event.Markets.Bookmakers {
			found := false
			for _, dbMarket := range dbMarkets { // Iterate through dbMarkets
				if dbMarket.MarketId == moMarket.MarketId {
					found = true
					break
				}
			}
			if found == true {
				continue
			}
			marketKey := eventKey + "-" + moMarket.MarketId
			market, err := GetMOMarket(moMarket, event, marketKey)
			if err != nil {
				log.Println("GetMarkets: GetMOMarket failed with error - ", err.Error())
				continue
			}
			newMarkets = append(newMarkets, market)
		}
		// Iterate through fancy markets in eventDto
		for _, fancyMarket := range event.Markets.FancyMarkets {
			found := false
			for _, dbMarket := range dbMarkets { // Iterate through dbMarkets
				if dbMarket.MarketId == fancyMarket.MarketId {
					found = true
					break
				}
			}
			if found == true {
				continue
			}
			marketKey := eventKey + "-" + fancyMarket.MarketId
			market, err := GetFancyMarket(fancyMarket, event, marketKey)
			if err != nil {
				log.Println("GetMarkets: GetFancyMarket failed with error - ", err.Error())
				continue
			}
			newMarkets = append(newMarkets, market)
		}
	}
	if len(newMarkets) > 0 {
		log.Println("GetMarkets: Adding new markets to database - ", len(newMarkets))
		if providerId != constants.SAP.ProviderType.SportRadar() {
			err = database.InsertManyMarkets(newMarkets)
			if err != nil {
				log.Println("GetMarkets: database.InsertManyMarkets failed with error - ", err.Error())
				return markets, nil
			}
		}
		markets = append(markets, newMarkets...)
		// insert MarketStatus
		marketStatus := []models.MarketStatus{}
		// Get Operators
		operators, err := database.GetAllOperators()
		if err != nil {
			log.Println("GetMarkets: database.GetAllOperators failed with error - ", err.Error())
			return markets, nil
		}
		for _, operator := range operators {
			for _, market := range newMarkets {
				ms := GetMarketStatus(operator, market)
				marketStatus = append(marketStatus, ms)
			}
		}
		log.Println("GetMarkets: Adding new marketstatus to database - ", len(marketStatus))
		if providerId != constants.SAP.ProviderType.SportRadar() {
			err = database.InsertManyMarketStatus(marketStatus)
			if err != nil {
				log.Println("GetMarkets: database.InsertManyMarketStatus failed with error - ", err.Error())
			}
		}
	}
	return markets, nil
}

// Can be called async
func InsertMarket(operatorId, providerId, sportId, eventId, marketId, marketType string) error {
	// Check is MarketKey exist in the Market Collection
	eventKey := providerId + "-" + sportId + "-" + eventId
	marketKey := eventKey + "-" + marketId
	log.Println("InsertMarket: market not found in Cache or DB for marketKey - ", marketKey)
	// not present in cache, Get from GetMarkets endpoint
	market, err := GetMarket(operatorId, providerId, sportId, eventId, marketId, marketType)
	if err != nil {
		// failed to get, log error and return
		log.Println("InsertMarket: GetMarkets failed with error - ", err.Error())
		return err
	}
	// Save market in Database
	err = database.InsertMarket(market)
	if err != nil {
		// failed to save in database, log error and return
		log.Println("InsertMarket: database.InsertMarket failed with error - ", err.Error())
		return err
	}
	// Add market to cahce
	cache.SetMarket(market)
	// Create MarketStatus list
	marketStatus := []models.MarketStatus{}
	// Get All Operators
	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("InsertMarket: database.GetAllOperators failed with error - ", err.Error())
		return err
	}
	for _, operator := range operators {
		ms := GetMarketStatus(operator, market)
		marketStatus = append(marketStatus, ms)
	}
	log.Println("InsertMarket: Adding new marketstatus to database - ", len(marketStatus))
	// Save marketstatus list in database
	err = database.InsertManyMarketStatus(marketStatus)
	if err != nil {
		log.Println("InsertMarket: database.InsertManyMarketStatus failed with error - ", err.Error())
		return err
	}
	// Add marketstatus list to cache
	for _, ms := range marketStatus {
		cache.SetMarketStatus(ms)
	}
	// Successfully added new market to the system, return
	marketJson, err := json.Marshal(market)
	if err != nil {
		log.Println("InsertMarket: json.Marshal failed with error - ", err.Error())
	} else {
		log.Println("InsertMarket: marketJson is - ", string(marketJson))
	}
	return nil
}

// Construct MarketStatus object from Operator & Market
func GetMarketStatus(operator operatordto.OperatorDTO, market models.Market) models.MarketStatus {
	ms := models.MarketStatus{}
	ms.MarketKey = operator.OperatorId + "-" + market.MarketKey
	ms.EventKey = operator.OperatorId + "-" + market.EventKey
	ms.OperatorId = operator.OperatorId
	ms.OperatorName = operator.OperatorName
	ms.ProviderId = market.ProviderId
	ms.ProviderName = market.ProviderName
	ms.SportId = market.SportId
	ms.SportName = market.SportName
	ms.CompetitionId = market.CompetitionId
	ms.CompetitionName = market.CompetitionName
	ms.EventId = market.EventId
	ms.EventName = market.EventName
	ms.MarketId = market.MarketId
	ms.MarketName = market.MarketName
	ms.MarketType = market.MarketType
	ms.ProviderStatus = constants.SAP.ObjectStatus.ACTIVE()
	ms.OperatorStatus = constants.SAP.ObjectStatus.ACTIVE()
	ms.Favourite = false
	ms.CreatedAt = time.Now().Unix()
	ms.UpdatedAt = ms.CreatedAt
	ms.Config = commondto.ConfigDto{}
	return ms
}

// Construct Market object from MatchOds Market
func GetMOMarket(moMarket dto.MatchOddsDto, eventDto dto.EventDto, marketKey string) (models.Market, error) {
	market := models.Market{}
	event, err := cache.GetEvent(eventDto.ProviderId, eventDto.SportId, eventDto.EventId)
	if err != nil {
		return market, err
	}
	market.MarketKey = marketKey
	market.EventKey = event.EventKey
	market.ProviderId = event.ProviderId
	market.ProviderName = event.ProviderName
	market.SportId = event.SportId
	market.SportName = event.SportName
	market.EventId = event.EventId
	market.EventName = event.EventName
	market.MarketId = moMarket.MarketId
	market.MarketName = moMarket.MarketName
	market.MarketType = moMarket.MarketType
	market.Category = ""
	market.Runners = []models.Runner{}
	for _, runDto := range moMarket.Runners {
		runner := models.Runner{}
		runner.RunnerId = runDto.RunnerId
		runner.RunnerName = runDto.RunnerName
		runner.RunnerStatus = runDto.Status
		market.Runners = append(market.Runners, runner)
	}
	market.Status = "ACTIVE"
	market.Favourite = false
	market.CreatedAt = time.Now().Unix()
	market.UpdatedAt = market.CreatedAt
	market.Config = commondto.ConfigDto{}
	market.MarketStatus = "OPEN" // TODO: Add MarketStatus constants (OPEN / MAPPED / INPROGRESS / SETTLED / VOIDED / CANCELLED / SUSPENDED)
	market.IsSuspended = false
	market.Results = []models.Result{}
	market.Rollbacks = []models.Rollback{}
	return market, nil
}

// Construct Market object from MatchOds Market
func GetFancyMarket(fancyMarket dto.FancyMarketDto, eventDto dto.EventDto, marketKey string) (models.Market, error) {
	market := models.Market{}
	event, err := cache.GetEvent(eventDto.ProviderId, eventDto.SportId, eventDto.EventId)
	if err != nil {
		return market, err
	}
	market.MarketKey = marketKey
	market.EventKey = event.EventKey
	market.ProviderId = event.ProviderId
	market.ProviderName = event.ProviderName
	market.SportId = event.SportId
	market.SportName = event.SportName
	market.EventId = event.EventId
	market.EventName = event.EventName
	market.MarketId = fancyMarket.MarketId
	market.MarketName = fancyMarket.MarketName
	market.MarketType = fancyMarket.MarketType
	market.Category = fancyMarket.Category
	market.Runners = []models.Runner{}
	market.Status = "ACTIVE"
	market.Favourite = false
	market.CreatedAt = time.Now().Unix()
	market.UpdatedAt = market.CreatedAt
	market.Config = commondto.ConfigDto{}
	market.MarketStatus = "OPEN" // TODO: Add MarketStatus constants (OPEN / MAPPED / INPROGRESS / SETTLED / VOIDED / CANCELLED / SUSPENDED)
	market.IsSuspended = false
	market.Results = []models.Result{}
	market.Rollbacks = []models.Rollback{}
	return market, nil
}

// Construct Bets object from Bet Reqest
func GetBetDto(betReq requestdto.PlaceBetReqDto, sessDto sessdto.B2BSessionDto, partner operatordto.Partner) sportsdto.BetDto {
	eventKey := betReq.ProviderId + "-" + betReq.SportId + "-" + betReq.EventId
	bet := sportsdto.BetDto{}
	bet.EventKey = eventKey
	bet.OperatorId = betReq.OperatorId
	bet.PartnerId = betReq.PartnerId
	bet.Token = betReq.Token
	bet.ProviderId = betReq.ProviderId
	bet.SportId = betReq.SportId
	bet.CompetitionId = betReq.CompetitionId
	bet.UserId = sessDto.UserId
	bet.UserName = sessDto.UserName
	bet.EventId = betReq.EventId
	bet.MarketId = betReq.MarketId
	bet.BetId = uuid.NewString() // Transaction Id
	bet.Status = "OPEN"
	if betReq.IsUnmatched {
		bet.Status = "UNMATCHED"
	}
	bet.UserIp = sessDto.UserIp
	bet.CreatedAt = time.Now().UnixMilli()
	bet.UpdatedAt = bet.CreatedAt
	// results & rollbacks
	bet.ResultReqs = []sportsdto.ResultReqDto{}
	bet.RollbackReqs = []sportsdto.RollbackReqDto{}
	// Bet Details
	bet.BetDetails = sportsdto.BetDetailsDto{}
	bet.BetDetails.BetType = betReq.BetType
	bet.BetDetails.OddValue = betReq.OddValue
	oddValue := betReq.OddValue
	if strings.ToUpper(betReq.MarketType) == "BOOKMAKER" {
		oddValue = (100 + betReq.OddValue) / 100
	} else if strings.ToUpper(betReq.MarketType) == "FANCY" {
		oddValue = (100 + betReq.OddValue) / 100
	} else if strings.ToUpper(betReq.MarketType) == "LINE_ODDS" {
		bet.BetDetails.OddValue = 2
		oddValue = 2
	}
	bet.BetDetails.StakeAmount = betReq.StakeAmount * float64(partner.Rate)
	bet.BetDetails.MarketType = betReq.MarketType
	bet.BetDetails.MarketName = betReq.MarketName
	bet.BetDetails.RunnerId = betReq.RunnerId
	bet.BetDetails.RunnerName = betReq.RunnerName
	bet.BetDetails.SessionOutcome = betReq.SessionOutcome
	bet.BetDetails.IsUnmatched = betReq.IsUnmatched
	event, err := cache.GetEvent(betReq.ProviderId, betReq.SportId, betReq.EventId)
	if err == nil {
		bet.BetDetails.SportName = event.SportName
		bet.BetDetails.CompetitionName = event.CompetitionName
		bet.BetDetails.EventName = event.EventName
	}
	// Bet Req (Operator Reference)
	bet.BetReq = sportsdto.BetReqDto{}
	bet.BetReq.BetId = bet.BetId        // for betfair, we will update with betfair betid
	bet.BetReq.ReqId = uuid.NewString() // Request Id
	bet.BetReq.ReqTime = bet.UpdatedAt
	if betReq.BetType == "BACK" {
		bet.BetReq.DebitAmount = bet.BetDetails.StakeAmount
	} else {
		bet.BetReq.DebitAmount = utils.Truncate4Decfloat64((oddValue - 1) * bet.BetDetails.StakeAmount)
	}
	bet.NetAmount = bet.BetReq.DebitAmount * -1
	bet.BetReq.Rate = partner.Rate
	return bet
}

// Construct openBet object from Bet Object
func GetOpenBetsDto(bets []sportsdto.BetDto, sportId string) []responsedto.OpenBetDto {
	openBets := []responsedto.OpenBetDto{}
	for _, bet := range bets {
		count := 0
		openBet := responsedto.OpenBetDto{}
		openBet.BetId = bet.BetId
		openBet.BetType = bet.BetDetails.BetType
		openBet.OddValue = bet.BetDetails.OddValue
		if bet.BetDetails.MarketType == constants.SAP.MarketType.LINE_ODDS() {
			openBet.OddValue = 2
		}
		openBet.RunnerName = bet.BetDetails.RunnerName
		openBet.RunnerId = bet.BetDetails.RunnerId
		openBet.MarketType = bet.BetDetails.MarketType
		openBet.MarketName = bet.BetDetails.MarketName
		openBet.MarketId = bet.MarketId
		openBet.EventId = bet.EventId
		openBet.SportId = sportId
		openBet.SessionOutcome = bet.BetDetails.SessionOutcome
		openBet.OddsHistory = []responsedto.OddsData{}
		for _, oddsData := range bet.OddsHistory {
			od := responsedto.OddsData{}
			od.OddsKey = oddsData.OddsKey
			od.OddsValue = oddsData.OddsValue
			od.OddsAt = oddsData.OddsAt
			openBet.OddsHistory = append(openBet.OddsHistory, od)
		}
		log.Println("GetOpenBetsDto: BetId, SP, SM, SR, PA - ", bet.BetId, bet.BetReq.SizePlaced, bet.BetReq.SizeMatched, bet.BetReq.SizeRemaining, bet.BetReq.PlatformAmount)

		if bet.BetReq.SizePlaced > 0 && bet.BetReq.SizePlaced == bet.BetReq.SizeMatched {
			// #1 Complete Matched
			log.Println("GetOpenBetsDto: #1 Complete Matched - ", bet.BetId)
			if bet.BetReq.OddsMatched > 0 {
				log.Println("GetOpenBetsDto: #1 bet.BetReq.OddsMatched - ", bet.BetId, bet.BetReq.OddsMatched)
				openBet.OddValue = bet.BetReq.OddsMatched
			}
			openBet.StakeAmount = bet.BetDetails.StakeAmount
			if bet.BetReq.Rate != 0 {
				openBet.StakeAmount = openBet.StakeAmount / float64(bet.BetReq.Rate)
			}
			openBet.IsUnmatched = false
			openBets = append(openBets, openBet)
		} else if bet.BetReq.SizePlaced > 0 && bet.BetReq.SizePlaced == bet.BetReq.SizeRemaining {
			// #2 Complete Unmatched
			log.Println("GetOpenBetsDto: #2 Complete Unmatched - ", bet.BetId)
			openBet.StakeAmount = bet.BetDetails.StakeAmount
			if bet.BetReq.Rate != 0 {
				openBet.StakeAmount = openBet.StakeAmount / float64(bet.BetReq.Rate)
			}
			openBet.IsUnmatched = true
			openBets = append(openBets, openBet)
		} else if bet.BetReq.SizePlaced > 0 {
			// #3 Partial Matched
			count = 0
			log.Println("GetOpenBetsDto: #3 Partial Matched - ", bet.BetId)
			if bet.BetReq.SizeRemaining > 0 {
				// #3.1. Unmatched Bet
				log.Println("GetOpenBetsDto: #3.1 Partial Unmatched bet.BetReq.SizeRemaining - ", bet.BetId, bet.BetReq.SizeRemaining)
				// unmatched amount
				openBet.StakeAmount = bet.BetReq.SizeRemaining
				if bet.ProviderId == constants.SAP.ProviderType.BetFair() {
					openBet.StakeAmount = bet.BetReq.SizeRemaining * float64(betfair.BetFairRate)
				}
				openBet.StakeAmount = (openBet.StakeAmount * 100) / (100 - bet.BetReq.PlatformHold)
				openBet.StakeAmount = (openBet.StakeAmount * 100) / (100 - bet.BetReq.OperatorHold)
				if bet.BetReq.Rate != 0 {
					openBet.StakeAmount = openBet.StakeAmount / float64(bet.BetReq.Rate)
				}
				openBet.StakeAmount = utils.Truncate64(openBet.StakeAmount)
				openBet.IsUnmatched = true
				openBets = append(openBets, openBet)
				count++
			}
			if bet.BetReq.SizeMatched > 0 {
				// #3.2. Matched Bet
				log.Println("GetOpenBetsDto: #3.2 Partial Matched bet.BetReq.SizeMatched - ", bet.BetId, bet.BetReq.SizeMatched)
				if bet.BetReq.OddsMatched != 0 {
					log.Println("GetOpenBetsDto: #3.2. bet.BetReq.OddsMatched - ", bet.BetId, bet.BetReq.OddsMatched)
					openBet.OddValue = bet.BetReq.OddsMatched
				}
				// matched amount
				openBet.StakeAmount = bet.BetReq.SizeMatched
				if bet.ProviderId == constants.SAP.ProviderType.BetFair() {
					openBet.StakeAmount = bet.BetReq.SizeMatched * float64(betfair.BetFairRate)
				}
				openBet.StakeAmount = (openBet.StakeAmount * 100) / (100 - bet.BetReq.PlatformHold)
				openBet.StakeAmount = (openBet.StakeAmount * 100) / (100 - bet.BetReq.OperatorHold)
				if bet.BetReq.Rate != 0 {
					openBet.StakeAmount = openBet.StakeAmount / float64(bet.BetReq.Rate)
				}
				openBet.StakeAmount = utils.Truncate64(openBet.StakeAmount)
				openBet.IsUnmatched = false
				openBets = append(openBets, openBet)
				count++
			}
			if count == 2 {
				log.Println("GetOpenBetsDto: #3 PartialMatched BetId, SP, SM, SR, SC - ", bet.BetId, bet.BetReq.SizePlaced, bet.BetReq.SizeMatched, bet.BetReq.SizeRemaining, bet.BetReq.SizeCancelled)
			}
		} else {
			// #4 Loacl Bet
			log.Println("GetOpenBetsDto: #4 Local Bet - ", bet.BetId)
			if bet.BetReq.OddsMatched > 0 {
				log.Println("GetOpenBetsDto: #4 bet.BetReq.OddsMatched - ", bet.BetId, bet.BetReq.OddsMatched)
				openBet.OddValue = bet.BetReq.OddsMatched
			}
			openBet.StakeAmount = bet.BetDetails.StakeAmount
			if bet.BetReq.Rate != 0 {
				openBet.StakeAmount = openBet.StakeAmount / float64(bet.BetReq.Rate)
			}
			openBet.IsUnmatched = false
			openBets = append(openBets, openBet)
		}
	}
	return openBets
}

// Construct allBets object from Bet Object
func GetAllBetsDto(bets []sportsdto.BetDto, events []models.Event, operatorDto operatordto.OperatorDTO) []responsedto.AllBetDto {
	allBets := []responsedto.AllBetDto{}
	partnerIdCurrencyMap := make(map[string]string)
	for _, partnerId := range operatorDto.Partners {
		partnerIdCurrencyMap[partnerId.PartnerId] = partnerId.Currency
	}
	for _, bet := range bets {
		allBet := responsedto.AllBetDto{}
		allBet.BetId = bet.BetId
		allBet.BetType = bet.BetDetails.BetType
		allBet.BetStatus = bet.Status
		allBet.RequestTime = bet.BetReq.ReqTime
		allBet.OddValue = bet.BetDetails.OddValue
		if bet.BetReq.OddsMatched > 0 {
			allBet.OddValue = bet.BetReq.OddsMatched
		}
		// allBet.StakeAmount = float64(bet.BetDetails.StakeAmount / 10)
		allBet.StakeAmount = utils.Truncate64(bet.BetDetails.StakeAmount / float64(bet.BetReq.Rate))
		allBet.RunnerName = bet.BetDetails.RunnerName
		allBet.RunnerId = bet.BetDetails.RunnerId
		allBet.MarketType = bet.BetDetails.MarketType
		allBet.MarketName = bet.BetDetails.MarketName
		allBet.MarketId = bet.MarketId
		allBet.EventId = bet.EventId
		allBet.EventName = bet.BetDetails.EventName
		allBet.SportId = strings.Split(bet.EventKey, "-")[1]
		allBet.SportName = bet.BetDetails.SportName
		allBet.SessionOutcome = bet.BetDetails.SessionOutcome
		allBet.Currency = partnerIdCurrencyMap[bet.PartnerId]
		allBet.IsUnmatched = false
		for _, event := range events {
			if event.EventId == allBet.EventId {
				allBet.OpenEventDate = event.OpenDate
			}
		}
		// if constants.SAP.BetStatus.UNMATCHED() == bet.Status || constants.SAP.BetStatus.INPROCESS() == bet.Status {
		if constants.SAP.BetStatus.UNMATCHED() == bet.Status {
			allBet.IsUnmatched = true
		}
		if bet.Status == constants.SAP.BetStatus.SETTLED() {
			var length int = len(bet.ResultReqs)
			if length > 0 {
				if bet.ResultReqs[length-1].CreditAmount > 0 {
					allBet.BetResult = "WON"
					allBet.BetReturns = utils.Truncate64((bet.ResultReqs[length-1].CreditAmount - bet.BetReq.DebitAmount) / float64(bet.BetReq.Rate))
				} else {
					allBet.BetResult = "LOST"
					allBet.BetReturns = utils.Truncate64((bet.BetReq.DebitAmount * -1) / float64(bet.BetReq.Rate))
				}
			}
		}
		allBets = append(allBets, allBet)
	}
	return allBets
}

func GetTransferUserStatement(operatorId, userId, referenceId string, startTime, endTime int64) ([]reports.Statement, models.B2BUserDto, error) {
	// Get User Balance from B2Busers
	userKey := operatorId + "-" + userId
	user, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println("GetTransferUserStatement: Error in getting user balance - ", err.Error())
		return nil, models.B2BUserDto{}, err
	}
	// Get User Bets from Bets
	ledger, err := database.GetLedgers(userKey, referenceId, startTime, endTime)
	if err != nil {
		log.Println("GetTransferUserStatement: Error in getting user bets - ", err.Error())
		return nil, user, err
	}
	log.Println("GetTransferUserStatement: len of bets - ", len(ledger))

	var lastTransAmt float64 = 0
	var currentBalance float64 = 0
	userStatement := []reports.Statement{}
	for _, ledgerItem := range ledger {
		statement := reports.Statement{}
		statement.TransactionTime = ledgerItem.TransactionTime
		statement.TransactionId = ledgerItem.ID.Hex()
		statement.TransactionType = ledgerItem.TransactionType
		statement.Balance = currentBalance
		lastTransAmt = 0
		if ledgerItem.TransactionType == constants.SAP.LedgerTxType.DEPOSIT() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETRESULT() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETROLLBACK() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETCANCEL() {
			statement.CreditAmount = ledgerItem.Amount
			lastTransAmt = ledgerItem.Amount * -1
		} else if ledgerItem.TransactionType == constants.SAP.LedgerTxType.WITHDRAW() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETPLACEMENT() {
			statement.DebitAmount = ledgerItem.Amount * -1
			lastTransAmt = ledgerItem.Amount * -1
		}
		currentBalance = currentBalance + lastTransAmt
		statement.ReferenceId = ledgerItem.ReferenceId
		statement.Remark = ledgerItem.Remark
		userStatement = append(userStatement, statement)
	}
	return userStatement, user, nil
}

func GetSeemlessUserStatement(operatorId, userId, referenceId, token string, startTime, endTime int64) ([]reports.SeemlessStatement, operatordto.OperatorRespDto, error) {
	resp := operatordto.OperatorRespDto{}
	if token == "" {
		sessionDto, err := function.GetSession(token)
		if err != nil {
			log.Println("GetSeemlessUserStatement: Error in getting session - ", err.Error())
			return nil, operatordto.OperatorRespDto{}, err
		}
		operatorDto, err := cache.GetOperatorDetails(operatorId)
		if err != nil {
			log.Println("GetSeemlessUserStatement: Error in getting operator details - ", err.Error())
			return nil, operatordto.OperatorRespDto{}, err
		}
		// Get User Balance from Operators
		resp, err = operator.WalletBalance(sessionDto, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("GetSeemlessUserStatement: Error in getting operator balance - ", err.Error())
			return nil, operatordto.OperatorRespDto{}, err
		}
	}
	// Get User Ledgers from OperatorLedgers
	userStatement := []reports.SeemlessStatement{}

	ledgers, err := database.GetOperatorLedgersByUserId(operatorId, userId, referenceId, startTime, endTime)
	if err != nil {
		log.Println("GetSeemlessUserStatement: Error in getting user bets - ", err.Error())
		return userStatement, operatordto.OperatorRespDto{}, err
	}
	var lastTransAmt float64 = 0
	var currentBalance float64 = resp.Balance
	for _, ledgerItem := range ledgers {
		statement := reports.SeemlessStatement{}
		statement.TransactionTime = ledgerItem.TransactionTime
		statement.TransactionId = ledgerItem.ID.Hex()
		statement.TransactionType = ledgerItem.TransactionType
		statement.Balance = currentBalance
		lastTransAmt = 0
		if ledgerItem.TransactionType == constants.SAP.LedgerTxType.DEPOSIT() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETRESULT() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETROLLBACK() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETCANCEL() {
			statement.CreditAmount = ledgerItem.Amount
			lastTransAmt = ledgerItem.Amount * -1
		} else if ledgerItem.TransactionType == constants.SAP.LedgerTxType.WITHDRAW() ||
			ledgerItem.TransactionType == constants.SAP.LedgerTxType.BETPLACEMENT() {
			statement.DebitAmount = ledgerItem.Amount * -1
			lastTransAmt = ledgerItem.Amount * -1
		}
		currentBalance = currentBalance + lastTransAmt
		statement.ReferenceId = ledgerItem.ReferenceId
		statement.Remark = "Seemless Transfer"
		userStatement = append(userStatement, statement)
	}
	return userStatement, resp, nil
}
