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

func OperatorLedger(operator operatordto.OperatorDTO, betDto sportsdto.BetDto) error {
	// 7. Bet Success, Add to operator ledger
	err := providers.OperatorLedgerTx(operator, constants.SAP.LedgerTxType.BETPLACEMENT(), betDto.BetReq.OperatorAmount*-1, betDto.BetId)
	if err != nil {
		log.Println("PlaceBet: OperatorLedger: providers.OperatorLedgerTx failed with error - ", err.Error())
		return err
	}
	return nil
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
					// openBet.StakeAmount = bet.BetReq.SizeRemaining * float64(betfair.BetFairRate)
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
					// openBet.StakeAmount = bet.BetReq.SizeMatched * float64(betfair.BetFairRate)
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
