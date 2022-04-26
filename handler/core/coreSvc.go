package coresvc

import (
	"Sp/cache"
	"Sp/common/function"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	"Sp/dto/reports"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"Sp/handler"
	"Sp/providers"
	"Sp/providers/betfair"
	"Sp/providers/dream"
	"Sp/providers/sportradar"
	utils "Sp/utilities"
	"encoding/json"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	TestToken              string        = "f23004ee-2bf4-49b9-a98b-f0189520b795"
	OperatorHttpReqTimeout time.Duration = 5
)

func GetUserBalance(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetUserBalanceRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	respDto.WalletType = "seamless"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(dto.GetUserBalanceReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetUserBalance: Body Parsing failed")
		log.Println("GetUserBalance: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetUserBalance: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//log.Println("GetUserBalance: User is - ", sessionDto.UserId)
	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetUserBalance: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("GetUserBalance: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	sessionDto.BaseURL = operatorDto.BaseURL
	partnerId := reqDto.PartnerId
	var rate int32 = 1 // default rate
	if partnerId == "" {
		log.Println("GetUserBalance: PartnerId is missing!")
		if len(operatorDto.Partners) == 0 {
			respDto.ErrorDescription = "PartnerId cannot be NULL!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		partnerId = operatorDto.Partners[0].PartnerId
	}
	found := false
	currency := ""
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == partnerId {
			found = true
			rate = partner.Rate
			currency = partner.Currency
			if partner.Status != "ACTIVE" {
				log.Println("GetUserBalance: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("GetUserBalance: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check if Operator wallet type is seamless or transfer?
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		// 5.1. Seamless wallet, do balance call to operator
		opResp, err := OpBalanceCall(sessionDto, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("GetUserBalance: Failed to get Operator Details: ", err.Error())
			// Get User balance from Cache
			respDto.ErrorDescription = "Failed to get user balance!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		// Balance will get in currency, will show in currency. No conversion is needed
		respDto.Balance = opResp.Balance // / float64(rate)
	} else if strings.ToLower(operatorDto.WalletType) == "transfer" {
		// 5.2. Transfer wallet, get user balance from db
		userKey := reqDto.OperatorId + "-" + sessionDto.UserId
		b2bUser, err := database.GetB2BUser(userKey)
		if err != nil {
			log.Println("GetUserBalance: User NOT FOUND: ", err.Error())
			respDto.ErrorDescription = "Invalid User!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Balance = b2bUser.Balance / float64(rate)
		respDto.WalletType = "seamless" // TODO: change to transfer
	} else if strings.ToLower(operatorDto.WalletType) == "feed" {
		log.Println("GetUserBalance: Operator wallet is not valid: ", operatorDto.WalletType)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	respDto.Balance = utils.Truncate64(respDto.Balance)
	respDto.Currency = currency
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetProviders to get *only* Active Providers for an operator
func GetProviders(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetProvidersRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Providers = []dto.ProviderInfo{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(dto.GetProvidersReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetProviders: Body Parsing failed")
		log.Println("GetProviders: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	/*
		if reqDto.Token != "" {
			sessionDto, err := ValidateToken(reqDto.Token)
			if err != nil {
				// 3.1. Return Error
				log.Println("GetProviders: Session Validation failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			log.Println("GetProviders: User is - ", sessionDto.UserId)
		} else {
			log.Println("GetProviders: Without Token, OperatorId is - ", reqDto.OperatorId)
		}
	*/
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetProviders:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetProviders: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("GetProviders: PartnerId is missing!")
		if len(operatorDto.Partners) == 0 {
			respDto.ErrorDescription = "PartnerId cannot be NULL!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		partnerId = operatorDto.Partners[0].PartnerId
	}
	found := false
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == partnerId {
			found = true
			if partner.Status != "ACTIVE" {
				log.Println("GetProviders: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("GetProviders: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Active Providers
	providerInfos, err := handler.GetActiveProviders(reqDto.OperatorId, partnerId)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetProviders: Failed to get Active Provider with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if operatorDto.BetFairPlus == true {
		for _, providerInfo := range providerInfos {
			if providerInfo.ProviderId == constants.SAP.ProviderType.Dream() {
				// SKIP Dream for BetFairPlus Mode (MPC1)
				continue
			}
			respDto.Providers = append(respDto.Providers, providerInfo)
		}
	} else {
		respDto.Providers = append(respDto.Providers, providerInfos...)
	}
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetSports is a function to Get Sports
// @Summary      Get Sports
// @Description  Get Sports
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        GetSports  body      commondto.GetSportsReqDto  true  "GetSportsReqDto model is used"
// @Success      200        {object}  commondto.GetSportsRespDto
// @Failure      503        {object}  commondto.GetSportsRespDto
// @Router       /core/getsports [post]
func GetSports(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.SportsRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Sports = []responsedto.SportDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(commondto.GetSportsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetSports: Body Parsing failed")
		log.Println("GetSports: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	/*
		if reqDto.Token != "" {
			sessionDto, err := ValidateToken(reqDto.Token)
			if err != nil {
				// 3.1. Return Error
				log.Println("GetSports: Session Validation failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			log.Println("GetSports: User is - ", sessionDto.UserId)
		} else {
			log.Println("GetSports: Without Token, OperatorId is - ", reqDto.OperatorId)
		}
	*/
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetSports:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetSports: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5.0. IsProviderActive?
	/*
		if false == providers.IsProviderActive(reqDto.OperatorId, reqDto.ProviderId) {
			// 4.2. Return Error
			log.Println("GetSports: IsProviderActive returned false - ", reqDto.ProviderId)
			respDto.ErrorDescription = "Unauthorized access, please contact support!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	*/
	// 5. Read from Cache
	sportDtos, err := function.GetOpActiveSports(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetSports: GetOpActiveSports failed with error - ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5.2. Prepare SportIds List
	mapSports := make(map[string]bool)
	for _, si := range sportDtos {
		_, isFound := mapSports[si.SportId]
		if isFound == false {
			mapSports[si.SportId] = true
			respDto.Sports = append(respDto.Sports, si)
		}
	}
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetEvents is a function to Get Events
// @Summary      Get Events
// @Description  Get Events
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        GetEvents  body      dto.GetEventsReqDto  true  "GetEventsReqDto model is used"
// @Success      200        {object}  dto.GetEventsRespDto{}
// @Failure      503        {object}  dto.GetEventsReqDto{}
// @Router       /core/getevents [post]
func GetEvents(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetEventsRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Events = []dto.EventDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(dto.GetEventsReqDto)
	log.Println("GetEvents: Req. Body is - ", bodyStr)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetEvents: Body Parsing failed")
		log.Println("GetEvents: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	/*
		if reqDto.Token != "" {
			sessionDto, err := ValidateToken(reqDto.Token)
			if err != nil {
				// 3.1. Return Error
				log.Println("GetEvents: Session Validation failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			log.Println("GetEvents: User is - ", sessionDto.UserId)
		} else {
			log.Println("GetEvents: Without Token, OperatorId is - ", reqDto.OperatorId)
		}
	*/
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetEvents:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetEvents: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Set PartnerId
	if len(operatorDto.Partners) == 0 {
		// 4.2. Return Error
		log.Println("GetEvents: Operator has no partnerIDs!!!", operatorDto.OperatorId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerDto := operatorDto.Partners[0]
	if reqDto.PartnerId == "" && reqDto.Token != "" { // NO & YES case
		sessionDto, err := function.GetSession(reqDto.Token)
		if err != nil {
			// 3.1. Return Error
			log.Println("GetEvents: GetSession failed with error - ", err.Error())
		} else {
			reqDto.PartnerId = sessionDto.PartnerId
		}
	}
	if reqDto.PartnerId != "" { // YES
		for _, partner := range operatorDto.Partners {
			if reqDto.PartnerId == partner.PartnerId {
				partnerDto = partner
				break
			}
		}
	}
	// // 5. Check PartnerId
	// partnerId := reqDto.PartnerId
	// if partnerId == "" {
	// 	log.Println("GetEvents: PartnerId is missing!")
	// 	if len(operatorDto.Partners) == 0 {
	// 		respDto.ErrorDescription = "PartnerId cannot be NULL!"
	// 		return c.Status(fiber.StatusOK).JSON(respDto)
	// 	}
	// 	partnerId = operatorDto.Partners[0].PartnerId
	// }
	// found := false
	// for _, partner := range operatorDto.Partners {
	// 	if partner.PartnerId == partnerId {
	// 		found = true
	// 		if partner.Status != "ACTIVE" {
	// 			log.Println("GetEvents: Partner is not Active: ", partner.Status)
	// 			respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
	// 			return c.Status(fiber.StatusOK).JSON(respDto)
	// 		}
	// 		break
	// 	}
	// }
	// if false == found {
	// 	log.Println("GetEvents: Partner Id not found: ", partnerId)
	// 	respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	// 6. IsSportActive?
	if false == providers.IsSportActive(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId, reqDto.SportId) {
		// 6.1. Return Error
		log.Println("GetEvents: IsSportActive returned false - ", reqDto.SportId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Get Events
	events := []dto.EventDto{}
	switch reqDto.ProviderId {
	case providers.DREAM_SPORT:
		events, err = dream.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("GetEvents: GetEvents for Dream failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case providers.BETFAIR:
		events, err = betfair.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("GetEvents: GetEvents for BetFair failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case providers.SPORT_RADAR:
		events, err = sportradar.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("GetEvents: GetEvents for SportRadar failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		log.Println("GetEvents: GetEvents for SportRadar Count is - ", len(events))
	default:
		log.Println("GetEvents: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	mapValue := make(map[string]bool)
	otherEvents := []dto.EventDto{}
	i := 0
	for _, event := range events {
		if reqDto.ProviderId == providers.SPORT_RADAR {
			// 0.5 Event is active, append to the list
			respDto.Events = append(respDto.Events, event)
			continue
		}
		compititionId := event.CompetitionId
		// 0.1. is CompetitionId missing???
		// if event.CompetitionId == "" || event.CompetitionId == "-1" {
		// 	log.Println("GetEvents: CompetitionId is missing for event - ", event.EventName)
		// 	continue
		// }
		// 0.2. Do map contains competitionId???
		isActive, isFound := mapValue[compititionId]
		if isFound == false {
			// not found in map, Get competition status, save it in map
			isActive = providers.IsCompetitionActive(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, compititionId)
			mapValue[compititionId] = isActive
		}
		// 0.3 Is comptition not active???
		if isActive == false {
			continue
		}
		// 0.4 Is event not active???
		if false == providers.IsEventActive(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, event.EventId) {
			continue
		}
		// 0.5 Event is active, append to the list
		// GetConfigs
		if len(event.Markets.MatchOdds) > 0 {
			configDto := handler.GetMarketConfig(operatorDto, reqDto.ProviderId, reqDto.ProviderId, event.SportId, event.CompetitionId, event.EventId, constants.SAP.MarketType.MATCH_ODDS())
			limitDto := dto.LimitsDto{}
			limitDto.MinBetValue = utils.Truncate64(float64(configDto.MinBetValue) / float64(partnerDto.Rate))
			limitDto.MaxBetValue = utils.Truncate64(float64(configDto.MaxBetValue) / float64(partnerDto.Rate))
			limitDto.OddsLimit = float64(configDto.OddsLimit)
			limitDto.Currency = partnerDto.Currency
			for i, _ := range event.Markets.MatchOdds {
				event.Markets.MatchOdds[i].Limits = limitDto
			}
		}
		// 0.5 Event is active, append to the list
		if event.CompetitionId == "101480" { // IPL CompetitionId
			respDto.Events = append(respDto.Events, event)
		} else {
			otherEvents = append(otherEvents, event)
		}
		i++
		if i == 50 { // only 50 events to frontend
			break
		}
	}
	respDto.Events = append(respDto.Events, otherEvents...)
	//log.Println("list-events: GetEvents for 2 "+reqDto.ProviderId+" Count is - ", len(respDto.Events))
	if len(events) > 0 && len(respDto.Events) == 0 {
		// 9.3. Default
		log.Println("GetEvents: Either Events or Competitions are DISABLED!!!")
		respDto.ErrorDescription = "Either Events or Competitions are DISABLED!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetEvents: GetEvents for "+reqDto.ProviderId+" Count is - ", len(respDto.Events), len(events))
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetLiveEvents is a function to Get Live Events
// @Summary      Get Live Events
// @Description  Get Live Events
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        GetLiveEvents  body      dto.GetLiveEventsReqDto  true  "GetLiveEventsReqDto model is used"
// @Success      200            {object}  dto.GetLiveEventsRespDto{}
// @Failure      503            {object}  dto.GetLiveEventsReqDto{}
// @Router       /core/getliveevents [post]
func GetLiveEvents(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetLiveEventsRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("GetLiveEvents: Req. Body is - ", bodyStr)
	reqDto := new(dto.GetLiveEventsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetLiveEvents: Body Parsing failed")
		log.Println("GetLiveEvents: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	/*
		if reqDto.Token != "" {
			sessionDto, err := ValidateToken(reqDto.Token)
			if err != nil {
				// 3.1. Return Error
				log.Println("GetLiveEvents: Session Validation failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			log.Println("GetLiveEvents: User is - ", sessionDto.UserId)
		} else {
			log.Println("GetLiveEvents: Without Token, OperatorId is - ", reqDto.OperatorId)
		}
	*/
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetLiveEvents:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetLiveEvents: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Set PartnerId
	if len(operatorDto.Partners) == 0 {
		// 4.2. Return Error
		log.Println("GetLiveEvents: Operator has no partnerIDs!!!", operatorDto.OperatorId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerDto := operatorDto.Partners[0]
	if reqDto.PartnerId == "" && reqDto.Token != "" { // NO & YES case
		sessionDto, err := function.GetSession(reqDto.Token)
		if err != nil {
			// 3.1. Return Error
			log.Println("GetMarkets: GetSession failed with error - ", err.Error())
		} else {
			reqDto.PartnerId = sessionDto.PartnerId
		}
	}
	if reqDto.PartnerId != "" { // YES
		for _, partner := range operatorDto.Partners {
			if reqDto.PartnerId == partner.PartnerId {
				partnerDto = partner
				break
			}
		}
	}
	// 5.0. IsProviderActive?
	/*
		if false == providers.IsProviderActive(reqDto.OperatorId, reqDto.ProviderId) {
			// 4.2. Return Error
			log.Println("GetLiveEvents: IsProviderActive returned false - ", reqDto.ProviderId)
			respDto.ErrorDescription = "Unauthorized access, please contact support!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	*/
	// 5.1. Get Active Sports by Operator & Provider
	sportInfos, err := function.GetOpActiveSports(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
	if err != nil {
		// 5.1.1. Return Error
		log.Println("GetLiveEvents: GetOpActiveSports failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5.2. Prepare SportIds List
	mapSports := make(map[string]bool)
	sportsList := []string{}
	for _, si := range sportInfos {
		_, isFound := mapSports[si.SportId]
		if isFound == false {
			mapSports[si.SportId] = true
			sportsList = append(sportsList, si.SportId)
		}
	}
	// 4. Get Live Events
	events := []dto.EventDto{}
	switch reqDto.ProviderId {
	case providers.DREAM_SPORT:
		//sportsList := []string{"4", "1", "2"}
		events, err = dream.GetLiveEvents(sportsList)
		if err != nil {
			log.Println("GetLiveEvents: GetLiveEvents for Dream failed with error - ", err.Error())
			// 3.1.1. Return Session NOT FOUND Error
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case providers.BETFAIR:
		//sportsList := []string{"4", "1", "2"}
		events, err = betfair.GetLiveEvents(sportsList)
		if err != nil {
			log.Println("GetLiveEvents: GetLiveEvents for BetFair failed with error - ", err.Error())
			// 3.1.1. Return Session NOT FOUND Error
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case providers.SPORT_RADAR:
		//sportsList := []string{"4", "1", "2"}
		events, err = sportradar.GetLiveEvents(sportsList)
		if err != nil {
			log.Println("GetLiveEvents: GetLiveEvents for SportRadar failed with error - ", err.Error())
			// 3.1.1. Return Session NOT FOUND Error
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	default:
		log.Println("GetLiveEvents: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	mapValue := make(map[string]bool)
	otherEvents := []dto.EventDto{}
	cricCount := 0
	socerCount := 0
	tennisCount := 0
	otherCount := 0
	soccerEvents := []dto.EventDto{}
	tennisEvents := []dto.EventDto{}
	for _, event := range events {
		if reqDto.ProviderId == providers.SPORT_RADAR {
			// 0.5 Event is active, append to the list
			respDto.Events = append(respDto.Events, event)
			continue
		}
		compititionId := event.CompetitionId
		// 0.1. is CompetitionId missing???
		// if event.CompetitionId == "" || event.CompetitionId == "-1" {
		// 	log.Println("GetLiveEvents: CompetitionId is missing for event - ", event.EventName)
		// 	continue
		// }
		// 0.2. Do map contains competitionId???
		isActive, isFound := mapValue[compititionId]
		if isFound == false {
			// not found in map, Get competition status, save it in map
			isActive = providers.IsCompetitionActive(reqDto.OperatorId, reqDto.ProviderId, event.SportId, compititionId)
			mapValue[compititionId] = isActive
		}
		// 0.3 Is comptition not active???
		if isActive == false {
			continue
		}
		// 0.4 Is event not active???
		if false == providers.IsEventActive(reqDto.OperatorId, reqDto.ProviderId, event.SportId, event.EventId) {
			continue
		}
		// GetConfigs
		if len(event.Markets.MatchOdds) > 0 {
			configDto := handler.GetMarketConfig(operatorDto, reqDto.ProviderId, reqDto.ProviderId, event.SportId, event.CompetitionId, event.EventId, constants.SAP.MarketType.MATCH_ODDS())
			limitDto := dto.LimitsDto{}
			limitDto.MinBetValue = utils.Truncate64(float64(configDto.MinBetValue) / float64(partnerDto.Rate))
			limitDto.MaxBetValue = utils.Truncate64(float64(configDto.MaxBetValue) / float64(partnerDto.Rate))
			limitDto.OddsLimit = float64(configDto.OddsLimit)
			limitDto.Currency = partnerDto.Currency
			for i, _ := range event.Markets.MatchOdds {
				event.Markets.MatchOdds[i].Limits = limitDto
			}
		}
		// 0.5 Event is active, append to the list
		if event.CompetitionId == "101480" { // IPL CompetitionId
			respDto.Events = append(respDto.Events, event)
			cricCount++
		} else {
			switch event.SportId {
			case "1":
				if socerCount < 20 {
					otherEvents = append(otherEvents, event)
					socerCount++
				} else {
					soccerEvents = append(soccerEvents, event)
				}
			case "2":
				if tennisCount < 20 {
					otherEvents = append(otherEvents, event)
					tennisCount++
				} else {
					tennisEvents = append(tennisEvents, event)
				}
			case "4":
				if cricCount < 10 {
					otherEvents = append(otherEvents, event)
					cricCount++
				}
			default:
				if otherCount < 10 {
					otherEvents = append(otherEvents, event)
					otherCount++
				}
			}
		}
	}
	respDto.Events = append(respDto.Events, otherEvents...)
	//log.Println("list-events: GetEvents for 2 "+reqDto.ProviderId+" Count is - ", len(respDto.Events))
	if len(events) > 0 && len(respDto.Events) == 0 {
		// 9.3. Default
		log.Println("GetLiveEvents: Either Events or Competitions are DISABLED!!!")
		respDto.ErrorDescription = "Either Events or Competitions are DISABLED!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if len(respDto.Events) < 50 && len(soccerEvents) > 0 {
		i := len(respDto.Events)
		for _, event := range soccerEvents {
			respDto.Events = append(respDto.Events, event)
			socerCount++
			i++
			if i == 50 {
				break
			}
		}
	}
	if len(respDto.Events) < 50 && len(tennisEvents) > 0 {
		i := len(respDto.Events)
		for _, event := range tennisEvents {
			respDto.Events = append(respDto.Events, event)
			tennisCount++
			i++
			if i == 50 {
				break
			}
		}
	}
	log.Println("GetLiveEvents:Total, cricCount, socerCount, tennisCount, otherCount - ", len(respDto.Events), cricCount, socerCount, tennisCount, otherCount)
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetMarkets is a function to Get Markets
// @Summary      Get Markets
// @Description  Get Markets
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        GetMarkets  body      dto.GetMarketsReqDto  true  "GetMarketsReqDto model is used"
// @Success      200         {object}  dto.GetMarketsRespDto{}
// @Failure      503         {object}  dto.GetMarketsReqDto{}
// @Router       /core/getmarkets [post]
func GetMarkets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetMarketsRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("GetMarkets: Req. Body is - ", bodyStr)
	reqDto := new(dto.GetMarketsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetMarkets: Body Parsing failed")
		log.Println("GetMarkets: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	/*
		if reqDto.Token != "" {
			sessionDto, err := ValidateToken(reqDto.Token)
			if err != nil {
				// 3.1. Return Error
				log.Println("GetEvents: Session Validation failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			log.Println("GetEvents: User is - ", sessionDto.UserId)
		} else {
			log.Println("GetEvents: Without Token, OperatorId is - ", reqDto.OperatorId)
		}
	*/
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetMarkets:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetMarkets: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Set PartnerId
	partnerDto := operatordto.Partner{}
	partnerId := ""
	if reqDto.Token != "" {
		sessionDto, err := function.GetSession(reqDto.Token)
		if err != nil {
			// 3.1. Return Error
			log.Println("GetMarkets: GetSession failed with error - ", err.Error())
		} else {
			partnerId = sessionDto.PartnerId
		}
	}
	if partnerId == "" {
		log.Println("GetMarkets: Without Token/PartnerId, OperatorId is - ", reqDto.OperatorId)
		for _, partner := range operatorDto.Partners {
			partnerDto = partner
			partnerId = partner.PartnerId
			break
		}
	} else {
		for _, partner := range operatorDto.Partners {
			if partnerId == partner.PartnerId {
				partnerDto = partner
				partnerId = partner.PartnerId
				break
			}
		}
	}
	// 5.0. IsEventActive?
	if reqDto.ProviderId != constants.SAP.ProviderType.SportRadar() { // TODO: Need to remove this check
		if false == providers.IsEventActive(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId) {
			// 4.2. Return Error
			respDto.ErrorDescription = "Unauthorized access, please contact support!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	}
	eventDto := dto.EventDto{}
	if reqDto.ProviderId == constants.SAP.ProviderType.BetFair() && operatorDto.BetFairPlus == true {
		log.Println("GetMarkets: BetFairPlus Mode (MPC1) MixMarket Data - ", reqDto.OperatorId)
		// BetFair & Dream mix
		// Get BetFair
		eventDto = betfair.GetMarkets(reqDto.SportId, reqDto.EventId, reqDto.OperatorId)
		// Get Dream
		drEvent := dream.GetMarkets(reqDto.SportId, reqDto.EventId, reqDto.OperatorId)
		eventDto.Markets.Bookmakers = []dto.MatchOddsDto{}
		eventDto.Markets.Bookmakers = append(eventDto.Markets.Bookmakers, drEvent.Markets.Bookmakers...)
		eventDto.Markets.FancyMarkets = []dto.FancyMarketDto{}
		eventDto.Markets.FancyMarkets = append(eventDto.Markets.FancyMarkets, drEvent.Markets.FancyMarkets...)
		respDto.ErrorDescription = ""
		respDto.Status = "RS_OK"
	} else {
		switch reqDto.ProviderId {
		case providers.DREAM_SPORT:
			eventDto = dream.GetMarkets(reqDto.SportId, reqDto.EventId, reqDto.OperatorId)
			respDto.ErrorDescription = ""
			respDto.Status = "RS_OK"
		case providers.BETFAIR:
			eventDto = betfair.GetMarkets(reqDto.SportId, reqDto.EventId, reqDto.OperatorId)
			respDto.ErrorDescription = ""
			respDto.Status = "RS_OK"
		case providers.SPORT_RADAR:
			eventDto = sportradar.GetMarkets(reqDto.SportId, reqDto.EventId)
			respDto.ErrorDescription = ""
			respDto.Status = "RS_OK"
		default:
			log.Println("GetEvents: Invalid ProviderId - ", reqDto.ProviderId)
			respDto.ErrorDescription = "Invalid Request"
		}
	}
	// GetConfigs
	if len(eventDto.Markets.MatchOdds) > 0 {
		configDto := handler.GetMarketConfig(operatorDto, partnerId, reqDto.ProviderId, reqDto.SportId, eventDto.CompetitionId, reqDto.EventId, constants.SAP.MarketType.MATCH_ODDS())
		limitDto := dto.LimitsDto{}
		limitDto.MinBetValue = utils.Truncate64(float64(configDto.MinBetValue) / float64(partnerDto.Rate))
		limitDto.MaxBetValue = utils.Truncate64(float64(configDto.MaxBetValue) / float64(partnerDto.Rate))
		limitDto.OddsLimit = float64(configDto.OddsLimit)
		limitDto.Currency = partnerDto.Currency
		for i, _ := range eventDto.Markets.MatchOdds {
			eventDto.Markets.MatchOdds[i].Limits = limitDto
		}
	}
	if len(eventDto.Markets.Bookmakers) > 0 {
		configDto := handler.GetMarketConfig(operatorDto, partnerId, reqDto.ProviderId, reqDto.SportId, eventDto.CompetitionId, reqDto.EventId, constants.SAP.MarketType.BOOKMAKER())
		limitDto := dto.LimitsDto{}
		limitDto.MinBetValue = utils.Truncate64(float64(configDto.MinBetValue) / float64(partnerDto.Rate))
		limitDto.MaxBetValue = utils.Truncate64(float64(configDto.MaxBetValue) / float64(partnerDto.Rate))
		limitDto.OddsLimit = float64(configDto.OddsLimit)
		limitDto.Currency = partnerDto.Currency
		for i, _ := range eventDto.Markets.Bookmakers {
			eventDto.Markets.Bookmakers[i].Limits = limitDto
		}
	}
	if len(eventDto.Markets.FancyMarkets) > 0 {
		configDto := handler.GetMarketConfig(operatorDto, partnerId, reqDto.ProviderId, reqDto.SportId, eventDto.CompetitionId, reqDto.EventId, constants.SAP.MarketType.FANCY())
		limitDto := dto.LimitsDto{}
		limitDto.MinBetValue = utils.Truncate64(float64(configDto.MinBetValue) / float64(partnerDto.Rate))
		limitDto.MaxBetValue = utils.Truncate64(float64(configDto.MaxBetValue) / float64(partnerDto.Rate))
		limitDto.OddsLimit = float64(configDto.OddsLimit)
		limitDto.Currency = partnerDto.Currency
		for i, _ := range eventDto.Markets.FancyMarkets {
			eventDto.Markets.FancyMarkets[i].Limits = limitDto
		}
	}
	respDto.Event = eventDto
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GetOpenBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.OpenBetsRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.OpenBets = []responsedto.OpenBetDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.OpenBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetOpenBets: Body Parsing failed")
		log.Println("GetOpenBets: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetOpenBets: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetOpenBets: User is - ", sessionDto.UserId)
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	openBets, _, err := database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetOpenBets: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetOpenBets: OperatorId - EventKey - betsCount - ", reqDto.OperatorId, eventKey, len(openBets))
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	respDto.OpenBets = handler.GetOpenBetsDto(openBets, reqDto.SportId)
	if operatorDto.BetFairPlus == true && reqDto.ProviderId == constants.SAP.ProviderType.BetFair() {
		log.Println("GetOpenBets: BetFairPlus Mode (MPC1) & BetFair!!!")
		eventKey = constants.SAP.ProviderType.Dream() + "-" + reqDto.SportId + "-" + reqDto.EventId
		openBets, _, err = database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
		if err != nil {
			log.Println("GetOpenBets: Failed with error - ", err.Error())
			//respDto.ErrorDescription = err.Error()
			//return c.Status(fiber.StatusOK).JSON(respDto)
		} else {
			log.Println("GetOpenBets: OperatorId - EventKey - betsCount - ", reqDto.OperatorId, eventKey, len(openBets))
			respDto.OpenBets = append(respDto.OpenBets, handler.GetOpenBetsDto(openBets, reqDto.SportId)...)
		}
	}
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// SportsBet is a function to Palce a Bet
// @Summary      Place Bet
// @Description  Bet Placement Async Endpoint
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        SportsBet  body      requestdto.PlaceBetReqDto  true  "PlaceBetReqDto model is used"
// @Success      200        {object}  responsedto.PlaceBetRespDto{}
// @Failure      503        {object}  responsedto.PlaceBetRespDto{}
// @Router       /core/sportsbet [post]
func SportsBet(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.PlaceBetRespDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("PlaceBet: APITiming: Req. Body is - ", bodyStr)
	reqDto := requestdto.PlaceBetReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("PlaceBet: Body Parsing failed with error - ", err.Error())
		log.Println("PlaceBet: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	reqJson, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("PlaceBet: reqBody json.Marshal failed with error - ", err.Error())
	}
	log.Println("PlaceBet: reqDto json is - ", string(reqJson))
	// 3. Validate Token
	_, err = function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("PlaceBet: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.OperatorId == "MPC1" {
		reqDto.StakeAmount = math.Floor(reqDto.StakeAmount)
	}
	userAgent := string(c.Request().Header.Peek("User-Agent"))
	log.Println("PlaceBet: userAgent", userAgent)
	if strings.Contains(userAgent, "Mobi") == true {
		reqDto.UserAgent = "mobile"
	} else {
		reqDto.UserAgent = "internet"
	}
	// 4.0 Place Bet
	betId, err := handler.PlaceBet(reqDto)
	if err != nil {
		log.Println("PlaceBet: handler.PlaceBet failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Send SUCCESS response
	respDto.BetId = betId
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func CancelBet(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.CancelBetRespDto{}
	// 2. Parse request body to Request Object
	reqDto := requestdto.CancelBetReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("CancelBet: Body Parsing failed")
		log.Println("CancelBet: Req. Body is - ", string(c.Body()))
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	_, err = function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("CancelBet: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Process Cancel Bets
	respDto.Balance, respDto.CancelBetsResp, err = handler.CancelBet(reqDto)
	if err != nil {
		log.Println("CancelBets: handler.CancelBet failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func AddUserCredits(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.AddUserCreditsRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	respDto.WalletType = "seamless"
	// 2. Parse request body to Request Object
	reqDto := dto.AddUserCreditsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("AddUserCredits: Body Parsing failed")
		log.Println("AddUserCredits: Req. Body is - ", string(c.Body()))
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("AddUserCredits: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("AddUserCredits: User is - ", sessionDto.UserId)
	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("AddUserCredits: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("AddUserCredits: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	var rate int32 = 1 // default rate
	if partnerId == "" {
		log.Println("AddUserCredits: PartnerId is missing!")
		if len(operatorDto.Partners) == 0 {
			respDto.ErrorDescription = "PartnerId cannot be NULL!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		partnerId = operatorDto.Partners[0].PartnerId
	}
	found := false
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == partnerId {
			found = true
			rate = partner.Rate
			if partner.Status != "ACTIVE" {
				log.Println("AddUserCredits: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("AddUserCredits: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	sessionDto.BaseURL = operatorDto.BaseURL
	// 5. Check if Operator wallet type is seamless or transfer?
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		// 5.1. Seamless wallet, do balance call to operator
		respDto.ErrorDescription = "Invalid Operation!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5.2. TODO: Transfer wallet, initiate Add Funds Operator API endpoint request
	/*
		opResp, err := OpBalanceCall(sessionDto, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("GetUserBalance: Failed to get Operator Details: ", err.Error())
			// Get User balance from Cache
			respDto.ErrorDescription = "Invalid Operation!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		respDto.Balance = opResp.Balance
	*/
	userKey := reqDto.OperatorId + "-" + sessionDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println("AddUserCredits: User NOT FOUND: ", err.Error())
		respDto.ErrorDescription = "Invalid User!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6.2.3. Save in User Ledger
	userLedger := models.UserLedgerDto{}
	userLedger.UserKey = userKey
	userLedger.OperatorId = reqDto.OperatorId
	userLedger.UserId = sessionDto.UserId
	userLedger.TransactionType = constants.SAP.LedgerTxType.DEPOSIT() // "Deposit-Funds"
	userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	//userLedger.ReferenceId = betDto.BetId
	userLedger.Amount = reqDto.CreditAmount * float64(rate) // -ve value means, debited from user account
	err = database.InsertLedger(userLedger)
	if err != nil {
		// 6.2.3.1. inserting ledger document failed
		log.Println("AddUserCredits: insert ledger failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		//return userBalance, err
		respDto.ErrorDescription = err.Error()
		respDto.Balance = utils.Truncate64(b2bUser.Balance / float64(rate))
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6.2.4. Debit amount from user balance and save
	err = database.UpdateB2BUserBalance(userLedger.UserKey, userLedger.Amount)
	if err != nil {
		// 6.2.4.1. updating user balance failed
		log.Println("AddUserCredits: update user balance failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		respDto.ErrorDescription = err.Error()
		respDto.Balance = utils.Truncate64(b2bUser.Balance / float64(rate))
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6.2.5. set balance to resp object
	respDto.Balance = utils.Truncate64((b2bUser.Balance + userLedger.Amount) / float64(rate))
	respDto.WalletType = "transfer"
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// UserBetStatus is a function to get a status of bet placement
// @Summary      User Bet Status
// @Description  User's last bet status - to support Async Bet Placement
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        UserBetStatus  body      requestdto.UserBetStatusReqDto  true  "UserBetStatusReqDto model is used"
// @Success      200            {object}  responsedto.UserBetStatusRespDto{}
// @Failure      503            {object}  responsedto.UserBetStatusRespDto{}
// @Router       /core/userbet-status [post]
func UserBetStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.UserBetStatusRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"
	respDto.BetStatus = ""
	respDto.BetErrorMsg = ""
	respDto.BetReqTime = 0
	respDto.BetId = ""
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.UserBetStatusReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UserBetStatus: Body Parsing failed")
		log.Println("UserBetStatus: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("UserBetStatus: GetSession failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//log.Println("UserBetStatus: User is - ", sessionDto.UserId)
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("UserBetStatus: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("UserBetStatus: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// if strings.ToLower(operatorDto.WalletType) != "feed" {
	// 	log.Println("UserBetStatus: Invalid wallet type - ", operatorDto.WalletType)
	// 	respDto.ErrorDescription = "Unauthorized access, please contact support!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	// 6. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("ListOpMarkets: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("ListOpMarkets: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 5. Get UserBetStatus
	ubs, err := database.GetUserBetStatus(reqDto.OperatorId, sessionDto.UserId)
	if err != nil {
		// 5.1. return error
		log.Println("UserBetStatus: GetUserBetStatus failed with error - ", err.Error())
		respDto.ErrorDescription = "Failed to get the status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 10. Send SUCCESS response
	respDto.BetStatus = ubs.Status
	respDto.BetErrorMsg = ubs.ErrorMessage
	respDto.BetReqTime = ubs.ReqTime
	respDto.BetId = ubs.ReferenceId
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	if ubs.Status == "PENDING" {
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 11. Get Open Bets
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	openBets, _, err := database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
	if err != nil {
		log.Println("UserBetStatus: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("UserBetStatus: OperatorId - EventKey - betsCount - ", reqDto.OperatorId, eventKey, len(openBets))
	respDto.OpenBets = handler.GetOpenBetsDto(openBets, reqDto.SportId)
	if operatorDto.BetFairPlus == true && reqDto.ProviderId == constants.SAP.ProviderType.BetFair() {
		log.Println("UserBetStatus: BetFairPlus Mode (MPC1) & BetFair!!!")
		eventKey = constants.SAP.ProviderType.Dream() + "-" + reqDto.SportId + "-" + reqDto.EventId
		openBets, _, err = database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
		if err != nil {
			log.Println("UserBetStatus: Failed with error - ", err.Error())
			//respDto.ErrorDescription = err.Error()
			//return c.Status(fiber.StatusOK).JSON(respDto)
		} else {
			log.Println("UserBetStatus: OperatorId - EventKey - betsCount - ", reqDto.OperatorId, eventKey, len(openBets))
			respDto.OpenBets = append(respDto.OpenBets, handler.GetOpenBetsDto(openBets, reqDto.SportId)...)
		}
	}
	// 8. Send SUCCESS response
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetAllBets is a function to get all bets w.r.t provider and Operator
// @Summary      Get All Bet
// @Description  Get all the bets placed by that operator for that particular provider.
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        GetAllBets  body      requestdto.GetAllBetsReqDto  true  "GetAllBetsReqDto model is used"
// @Success      200         {object}  responsedto.GetAllBetsRespDto{}
// @Failure      503         {object}  responsedto.GetAllBetsRespDto{}
func GetAllBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.GetAllBetsRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"
	respDto.AllBets = []responsedto.AllBetDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.GetAllBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetAllBets: Body Parsing failed")
		log.Println("GetAllBets: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetAllBets: GetSession failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//log.Println("GetAllBets: User is - ", sessionDto.UserId)
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(sessionDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("GetAllBets: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetAllBets: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetAllBets: User is - ", sessionDto.UserId)
	allBets, err := database.GetAllBetsByOperatorIdUserId(sessionDto.OperatorId, sessionDto.UserId)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetAllBets: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("UserBetStatus: OperatorId - betsCount - ", operatorDto.OperatorId, len(allBets))

	// Get Event Ids from allBets
	eventIds := make([]string, 0)
	for _, bet := range allBets {
		eventIds = append(eventIds, bet.EventId)
	}
	events, err := database.GetEventsByEventIds(eventIds)
	if err != nil {
		log.Println("GetAllBets: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	respDto.AllBets = handler.GetAllBetsDto(allBets, events, operatorDto)
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Transfer User Statement API
// @Summary      Get Transfer User Statement API
// @Description  Get Transfer User Statement API
// @Tags         Core
// @Accept       json
// @Produce      json
// @Param        GetUserStatement  body      requestdto.GetAllBetsReqDto  true  "GetAllBetsReqDto model is used"
// @Success      200               {object}  reports.TransferUserStatementRespDto
// @Failure      503               {object}  reports.TransferUserStatementRespDto
// @Router       /core/get-transfer-user-statement [post]
func GetUserStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := reports.TransferUserStatementRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"

	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.GetAllBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetUserStatement: Body Parsing failed")
		log.Println("GetUserStatement: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetUserStatement: GetSession failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(sessionDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("GetUserStatement: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetUserStatement: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetUserStatement: User is - ", sessionDto.UserId)
	// 5. Get Transfer User Statement
	statement, user, err := handler.GetTransferUserStatement(sessionDto.OperatorId, sessionDto.UserId, reqDto.ReferenceId, reqDto.StartTime, reqDto.EndTime)
	if err != nil {
		// 3.1. Return Error
		log.Println("GetUserStatement: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Return Response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.UserId = sessionDto.UserId
	respDto.UserName = user.UserName
	respDto.UserBalance = user.Balance
	respDto.Statement = statement
	return c.Status(fiber.StatusOK).JSON(respDto)
}
