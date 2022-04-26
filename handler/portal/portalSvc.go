package portalsvc

import (
	"Sp/cache"
	operatorCache "Sp/cache/operator"
	"Sp/common/function"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	coredto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	dto "Sp/dto/portal"
	portaldto "Sp/dto/portal"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"Sp/handler"
	operatorsvc "Sp/handler/operator"
	"Sp/providers/betfair"
	"Sp/providers/dream"
	"Sp/providers/sportradar"
	keyutils "Sp/utilities"
	utils "Sp/utilities"
	"encoding/json"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

var (
	GENERAL_ERROR_DESC            = "Generic Error"
	INVALID_REQ_ERROR_DESC        = "Invalid Request"
	INVALID_USERID_ERROR_DESC     = "UserId not found"
	INVALID_OPERATORID_ERROR_DESC = "Invalid OperatorId"
	INVALID_PARTNERID_ERROR_DESC  = "Invalid PartnerId"
	INVALID_TOKEN_ERROR_DESC      = "Invalid Token!"
	UNAUTH_ACCESS                 = "Unauthorized Access!"
	ERROR_STATUS                  = "RS_ERROR"
	OK_STATUS                     = "RS_OK"
	SOMETHING_WENT_WRONG          = "Something Went Worng"

	ACTIVE  = "ACTIVE"
	BLOCKED = "BLOCKED"
)

// Create an Operator
func CreateOperator(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(portaldto.CreateOperatorRespDto)
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("CreateOperator: Body string is - ", bodyStr)
	reqDto := new(portaldto.CreateOperatorReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("CreateOperator: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	//2.1 Check if Operator_id has '-'
	if strings.Contains(reqDto.OperatorId, "-") {
		respDto.ErrorDescription = "Operator Id should not contain '-'"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	//2.2 Check if Operator_id is empty
	if reqDto.OperatorId == "" {
		respDto.ErrorDescription = "Operator Id should not be empty"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 3. Create RSA Keys
	privKey, pubKey, err := keyutils.GetNewKeys()
	if err != nil {
		log.Println("CreateOperator: Key pair creation failed")
		respDto.ErrorDescription = "Failed to create create pair"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 4. Create Operator object
	operatorDto := operatordto.OperatorDTO{}
	operatorDto.OperatorId = reqDto.OperatorId
	operatorDto.OperatorName = reqDto.OperatorName
	operatorDto.Keys.OperatorKey = reqDto.OperatorKey
	operatorDto.Status = reqDto.Status
	operatorDto.BaseURL = reqDto.BaseURL
	operatorDto.WalletType = reqDto.WalletType
	operatorDto.Partners = []operatordto.Partner{}
	operatorDto.Partners = append(operatorDto.Partners, reqDto.Partners...)
	operatorDto.Currencies = []operatordto.Currency{{Currency: reqDto.Currency, Commisssion: reqDto.Commisssion}}
	operatorDto.Keys.PrivateKey = privKey
	operatorDto.Keys.PublicKey = pubKey
	operatorDto.Config = reqDto.Config
	operatorDto.Ips = reqDto.Ips
	reqDto.AddSubFields = true
	if reqDto.AddSubFields {
		// 7.1 Create Provider_Status
		providers, err := database.GetAllProviders()
		if err != nil {
			log.Println("GetOperators: Get Operators failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		partnerStatus := []models.PartnerStatus{}
		for _, partner := range reqDto.Partners {
			for _, provider := range providers {
				partnerSts := models.PartnerStatus{}
				partnerSts.PartnerKey = reqDto.OperatorId + "-" + partner.PartnerId + "-" + provider.ProviderId
				partnerSts.OperatorId = reqDto.OperatorId
				partnerSts.OperatorName = reqDto.OperatorName
				partnerSts.PartnerId = partner.PartnerId
				partnerSts.ProviderId = provider.ProviderId
				partnerSts.ProviderName = provider.ProviderName
				partnerSts.ProviderStatus = BLOCKED
				partnerSts.OperatorStatus = BLOCKED
				partnerStatus = append(partnerStatus, partnerSts)
			}
		}
		// 7.2 Save in provider_status to DB
		if len(partnerStatus) > 0 {
			err = database.InsertManyPartnerStatus(partnerStatus)
			if err != nil {
				log.Println("CreateOperator: Insert ProviderStatus failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.JSON(respDto)
			}
			// 7.3 Save in provider_status to Cache
			for _, ps := range partnerStatus {
				cache.SetPartnerStatus(ps)
			}
		}

		// 8.1 Create Sport_Status
		sports, err := database.GetAllSports()
		if err != nil {
			log.Println("GetOperators: Get Operators failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.JSON(respDto)
		}
		sportStatsuDtos := []models.SportStatus{}
		for _, sport := range sports {
			sportKey := reqDto.OperatorId + "-" + sport.SportKey // OperatorId+"-"+ProviderId+"-"+SportId

			// Not found, create and add to the missing list
			sportStatus := models.SportStatus{}
			sportStatus.SportKey = sportKey
			sportStatus.OperatorId = reqDto.OperatorId
			sportStatus.OperatorName = reqDto.OperatorName
			sportStatus.ProviderId = sport.ProviderId
			sportStatus.ProviderName = sport.ProviderName
			sportStatus.SportId = sport.SportId
			sportStatus.SportName = sport.SportName
			sportStatus.ProviderStatus = ACTIVE
			sportStatus.OperatorStatus = ACTIVE
			sportStatus.Favourite = false
			sportStatus.CreatedAt = time.Now().Unix()
			sportStatsuDtos = append(sportStatsuDtos, sportStatus)
		}
		// 8.2 Save in sport_status to DB
		err = database.InsertManySportStatus(sportStatsuDtos)
		if err != nil {
			log.Println("CreateOperator: Insert SportStatus failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.JSON(respDto)
		}
		// 8.3 Save in sport_status to Cache
		for _, sportStatus := range sportStatsuDtos {
			cache.SetSportStatus(sportStatus)
		}

		//9.1 Create Competition_status
		// var competitions []models.Competition
		// for _, provider := range providers {
		// 	providerCompetitions, err := database.GetAllCompetitions(provider.ProviderId)
		// 	if err != nil {
		// 		log.Println("GetOperators: Get Operators failed with error - ", err.Error())
		// 		respDto.ErrorDescription = err.Error()
		// 		return c.JSON(respDto)
		// 	}
		// 	competitions = append(competitions, providerCompetitions...)
		// }
		// competitionStatusDtos := []models.CompetitionStatus{}
		// for _, competition := range competitions {
		// 	competitionKey := reqDto.OperatorId + "-" + competition.CompetitionKey // OperatorId+"-"+ProviderId+"-"+SportId+"-"+CompetitionId
		// 	// Not found, create and add to the missing list
		// 	competitionStatus := models.CompetitionStatus{}
		// 	competitionStatus.CompetitionKey = competitionKey
		// 	competitionStatus.OperatorId = reqDto.OperatorId
		// 	competitionStatus.OperatorName = reqDto.OperatorName
		// 	competitionStatus.ProviderId = competition.ProviderId
		// 	competitionStatus.ProviderName = competition.ProviderName
		// 	competitionStatus.SportId = competition.SportId
		// 	competitionStatus.SportName = competition.SportName
		// 	competitionStatus.CompetitionId = competition.CompetitionId
		// 	competitionStatus.CompetitionName = competition.CompetitionName
		// 	competitionStatus.ProviderStatus = ACTIVE
		// 	competitionStatus.OperatorStatus = ACTIVE
		// 	competitionStatus.Favourite = false
		// 	competitionStatus.CreatedAt = time.Now().Unix()
		// 	competitionStatusDtos = append(competitionStatusDtos, competitionStatus)
		// }
		// // 9.2 Save in competition_status to DB
		// err = database.InsertManyCompetitionStatus(competitionStatusDtos)
		// if err != nil {
		// 	log.Println("CreateOperator: Insert CompetitionStatus failed with error - ", err.Error())
		// 	respDto.ErrorDescription = err.Error()
		// 	return c.JSON(respDto)
		// }
		// // 9.3 Save in competition_status to Cache
		// for _, competitionStatus := range competitionStatusDtos {
		// 	cache.SetCompetitionStatus(competitionStatus)
		// }

		// 10.1 Create Event_status
		// var events []models.Event
		// for _, provider := range providers {
		// 	var providerEvents []models.Event
		// 	if provider.ProviderId == "BetFair" {
		// 		sports := []string{"1", "2", "4"}
		// 		var sportEvents []models.Event
		// 		for _, sport := range sports {
		// 			sportEvents, err = database.GetEventsByProviderIdAndSportId(provider.ProviderId, sport)
		// 			providerEvents = append(providerEvents, sportEvents...)
		// 		}
		// 	} else {
		// 		providerEvents, err = database.GetAllEvents(provider.ProviderId)
		// 	}
		// 	if err != nil {
		// 		log.Println("GetOperators: Get Operators failed with error - ", err.Error())
		// 		respDto.ErrorDescription = err.Error()
		// 		return c.JSON(respDto)
		// 	}
		// 	events = append(events, providerEvents...)
		// }
		// eventStatusDtos := []models.EventStatus{}
		// for _, event := range events {
		// 	eventKey := reqDto.OperatorId + "-" + event.EventKey // OperatorId+"-"+ProviderId+"-"+SportId+"-"+CompetitionId+"-"+EventId
		// 	// Not found, create and add to the missing list
		// 	eventStatus := models.EventStatus{}
		// 	eventStatus.EventKey = eventKey
		// 	eventStatus.OperatorId = reqDto.OperatorId
		// 	eventStatus.OperatorName = reqDto.OperatorName
		// 	eventStatus.ProviderId = event.ProviderId
		// 	eventStatus.ProviderName = event.ProviderName
		// 	eventStatus.SportId = event.SportId
		// 	eventStatus.SportName = event.SportName
		// 	eventStatus.CompetitionId = event.CompetitionId
		// 	eventStatus.CompetitionName = event.CompetitionName
		// 	eventStatus.EventId = event.EventId
		// 	eventStatus.EventName = event.EventName
		// 	eventStatus.ProviderStatus = ACTIVE
		// 	eventStatus.OperatorStatus = ACTIVE
		// 	eventStatus.Favourite = false
		// 	eventStatus.CreatedAt = time.Now().Unix()
		// 	eventStatus.UpdatedAt = eventStatus.CreatedAt
		// 	eventStatusDtos = append(eventStatusDtos, eventStatus)
		// }
		// // 10.2 Save in event_status to DB
		// err = database.InsertManyEventStatus(eventStatusDtos)
		// if err != nil {
		// 	log.Println("CreateOperator: Insert EventStatus failed with error - ", err.Error())
		// 	respDto.ErrorDescription = err.Error()
		// 	return c.JSON(respDto)
		// }
		// // 10.3 Save in event_status to Cache
		// for _, eventStatus := range eventStatusDtos {
		// 	cache.SetEventStatus(eventStatus)
		// }

		// 11.1 Create Market_status
		var markets []models.Market
		for _, provider := range providers {
			providerMarkets, err := database.GetAllMarketByProviderId(provider.ProviderId)
			if err != nil {
				log.Println("GetOperators: Get Operators failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.JSON(respDto)
			}
			markets = append(markets, providerMarkets...)
		}
		marketStatusDtos := []models.MarketStatus{}
		for _, market := range markets {
			marketKey := reqDto.OperatorId + "-" + market.MarketKey // OperatorId+"-"+ProviderId+"-"+SportId+"-"+CompetitionId+"-"+EventId+"-"+MarketId
			// Not found, create and add to the missing list
			marketStatus := models.MarketStatus{}
			marketStatus.MarketKey = marketKey
			marketStatus.EventKey = market.EventKey
			marketStatus.OperatorId = reqDto.OperatorId
			marketStatus.OperatorName = reqDto.OperatorName
			marketStatus.ProviderId = market.ProviderId
			marketStatus.ProviderName = market.ProviderName
			marketStatus.SportId = market.SportId
			marketStatus.SportName = market.SportName
			marketStatus.CompetitionId = market.CompetitionId
			marketStatus.CompetitionName = market.CompetitionName
			marketStatus.EventId = market.EventId
			marketStatus.EventName = market.EventName
			marketStatus.MarketId = market.MarketId
			marketStatus.MarketName = market.MarketName
			marketStatus.ProviderStatus = ACTIVE
			marketStatus.OperatorStatus = ACTIVE
			marketStatus.Favourite = false
			marketStatus.CreatedAt = time.Now().Unix()
			marketStatus.UpdatedAt = marketStatus.CreatedAt
			marketStatusDtos = append(marketStatusDtos, marketStatus)
		}
		// 10.2 Save in market_status to DB
		err = database.InsertManyMarketStatus(marketStatusDtos)
		if err != nil {
			log.Println("CreateOperator: Insert MarketStatus failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.JSON(respDto)
		}
		// 10.3 Save in event_status to Cache
		for _, marketStatus := range marketStatusDtos {
			cache.SetMarketStatus(marketStatus)
		}
	}
	// 5. Save in DB
	database.InsertOperatorDetails(operatorDto)
	// 6. Save in Cache
	cache.SetOperatorDetails(operatorDto)

	respDto.PublicKey = pubKey
	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS
	log.Println("CreateOperator: public key start - \n\n", pubKey)
	log.Println("\n\nCreateOperator: public key end")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func DeleteOperator(c *fiber.Ctx) error {
	log.Println("DeleteOperator: start")
	respDto := new(portaldto.CommonPortalRespDto)
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = "GENERIC Error"

	reqDto := new(portaldto.DeleteOperatorReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("DeleteOperator: Error while parsing body - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 2. Delete from DB
	err = database.DeleteOperatorDetails(reqDto.OperatorId)
	if err != nil {
		log.Println("DeleteOperator: Delete Operator failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 3.Delete events status by Operator
	err = database.DeleteEventsByOperator(reqDto.OperatorId)
	if err != nil {
		log.Println("DeleteOperator: Delete Events by Operator failed with error - ", err.Error())
	}

	// 4. Delete competitions status by Operator
	err = database.DeleteCompetitionsByOperator(reqDto.OperatorId)
	if err != nil {
		log.Println("DeleteOperator: Delete Competitions by Operator failed with error - ", err.Error())
	}

	// 5. Delete sports status by Operator
	err = database.DeleteSportsByOperator(reqDto.OperatorId)
	if err != nil {
		log.Println("DeleteOperator: Delete Sports by Operator failed with error - ", err.Error())
	}

	// 6. Delete Partner status by Operator
	err = database.DeletePartnersByOperator(reqDto.OperatorId)
	if err != nil {
		log.Println("DeleteOperator: Delete Partners by Operator failed with error - ", err.Error())
	}

	// 7. Delete Markets status by Operator
	err = database.DeleteMarketsByOperator(reqDto.OperatorId)
	if err != nil {
		log.Println("DeleteOperator: Delete Markets by Operator failed with error - ", err.Error())
	}

	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS
	log.Println("DeleteOperator: end")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func DeleteSportJunkData(c *fiber.Ctx) error {
	log.Println("DeleteSportJunkData: start")
	respDto := new(portaldto.CommonPortalRespDto)
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = "GENERIC Error"

	reqDto := new(portaldto.DeleteSportJunkDataReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("DeleteSportJunkData: Error while parsing body - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 1. Delete sports status by sportId
	if reqDto.SportId != "" {
		err = database.DeleteSportsBySportId(reqDto.SportId)
		if err != nil {
			log.Println("DeleteJunkData: Delete Sports by SportId failed with error - ", err.Error())
		}
		err = database.DeleteSportstatusBySportId(reqDto.SportId)
		if err != nil {
			log.Println("DeleteJunkData: Delete Sports by SportId failed with error - ", err.Error())
		}
	}

	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS
	log.Println("DeleteJunkData: end")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func DeleteCompetitionJunkData(c *fiber.Ctx) error {
	log.Println("DeleteCompetitionJunkData: start")
	respDto := new(portaldto.CommonPortalRespDto)
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = "GENERIC Error"

	reqDto := new(portaldto.DeleteCompetitionJunkDataReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("DeleteCompetitionJunkData: Error while parsing body - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 1. Delete competitions status by Comp_id
	if reqDto.CompetitionId != "" {
		err = database.DeleteCompetitionsBycompetitionId(reqDto.CompetitionId)
		if err != nil {
			log.Println("DeleteCompetitionJunkData: Delete Competitions by CompetitionId failed with error - ", err.Error())
		}
		err = database.DeleteCompetitionstatusBycompetitionId(reqDto.CompetitionId)
		if err != nil {
			log.Println("DeleteCompetitionJunkData: Delete Competitions by CompId failed with error - ", err.Error())
		}
	}

	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS
	log.Println("DeleteCompetitionJunkData: end")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func DeleteEventJunkData(c *fiber.Ctx) error {
	log.Println("DeleteEventJunkData: start")
	respDto := new(portaldto.CommonPortalRespDto)
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = "GENERIC Error"

	reqDto := new(portaldto.DeleteEventJunkDataReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("DeleteEventJunkData: Error while parsing body - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 1.Delete events status by Event_id
	if reqDto.EventId != "" {
		err = database.DeleteEventsByEventId(reqDto.EventId)
		if err != nil {
			log.Println("DeleteEventJunkData: Delete Events by EventId failed with error - ", err.Error())
		}
		err = database.DeleteEventstatusByEventId(reqDto.EventId)
		if err != nil {
			log.Println("DeleteEventJunkData: Delete Events by EventId failed with error - ", err.Error())
		}
	}

	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS
	log.Println("DeleteEventJunkData: end")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Login for Portal User - Role: SPAdmin/OperatorAdmin
// Login is a function to login into portal for administration
// @Summary      Portal Login
// @Description  Login into an administration portal
// @Tags         Portal-OperatorAdmin, Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        Login  body      portaldto.PortalLoginReqDto  true  "PortalLoginReqDto model is used"
// @Success      200    {object}  portaldto.PortalLoginRespDto
// @Failure      503    {object}  portaldto.PortalLoginRespDto
// @Router       /portal/login [post]
func Login(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	//Authenticate Token

	//Check Role

	respDto := new(portaldto.PortalLoginRespDto)
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(portaldto.PortalLoginReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("CreateOperator: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	userId := reqDto.UserId
	pass := reqDto.Password
	ip := reqDto.IP
	// Create UserInstance of the existing user
	userInstance, err := database.GetPortalUserDetailsByUserId(userId)
	if userInstance.UserKey == "" {
		userInstance, err = database.GetPortalUserDetailsByUserName(userId)
		if err != nil {
			log.Println("Portal Login: User Instance not found")
			respDto.ErrorDescription = INVALID_USERID_ERROR_DESC
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}

	// Check if user is active
	if userInstance.Status != ACTIVE {
		log.Println("Portal Login: User is not active")
		respDto.ErrorDescription = "User is Blocked: Please contact Support"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// Validate the password entered by user
	if !CheckPasswordHash(pass, userInstance.Password) {
		log.Println("Portal Login: Invalid Password or UserId")
		respDto.ErrorDescription = "Invalid Password or UserId"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	//Add Claims
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = userId
	claims["role"] = userInstance.Role
	claims["operatorId"] = userInstance.OperatorId
	claims["userId"] = userInstance.UserId
	claims["ip"] = ip
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	//Sign a new Token.
	tkn, err := token.SignedString([]byte("SECRET"))
	if err != nil {
		log.Println("Portal Login: Unable to sign a new token")
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	//Insert The token to session
	session := new(operatordto.PortalSession)
	session.ExpiresAt = claims["exp"].(int64)
	session.UserId = userId
	session.JWTToken = tkn
	session.OperatorId = userInstance.OperatorId
	database.InsertSessionTokenDetails(*session)

	//Upate the DB and Cache
	if !operatorCache.SetPortalSessionDetails(*session) {
		log.Println("Error: Updating Portal Session Cache. Key: UserId - ", userId, ": ", err)
	}
	respDto.Token = tkn
	respDto.ErrorDescription = "Success"
	respDto.Status = OK_STATUS
	log.Println("Auth Tkn: ", tkn)
	log.Println("Portal Auth Login: ended")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Create Portal User
func CreatePortalUser(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := new(portaldto.PortalCreateUserRespDto)
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	//Authenticate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		//Check for error and send 400 and error.
		log.Println("CreatePortalUser: Token Authentication Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	//Check Token User Role
	if !IsApplicable(Tknmeta, "createuser") {
		log.Println("CreatePortalUser: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// portalUserReq := operatordto.PortalUserReq{UserId: "amit", OperatorId: "hypexone", Password: "1234", UserName: "Amit", Status: ACTIVE, Role: "SPAdmin"}

	portalUserReq := new(operatordto.PortalUserReq)
	if err := c.BodyParser(portalUserReq); err != nil {
		log.Println("CreatePortalUser: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	//Hashing the password.
	hash, err := HashPassword(portalUserReq.Password)
	if err != nil {
		log.Println("CreatePortalUser: Couldn't Hash Password - ", err.Error())
		respDto.ErrorDescription = "Couldn't Hash Password"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//Converting JSON Request to BSON DTO
	var userInstance operatordto.PortalUser
	userInstance.UserKey = portalUserReq.UserId + "-" + portalUserReq.OperatorId
	userInstance.Password = hash
	userInstance.UserId = portalUserReq.UserId
	userInstance.OperatorId = portalUserReq.OperatorId
	userInstance.Role = portalUserReq.Role
	userInstance.Status = portalUserReq.Status
	userInstance.UserName = portalUserReq.UserName

	// Checking If user with same userId is already persent in DB.
	checkExisting, err := database.GetPortalUserDetailsByUserId(userInstance.UserId)
	if err != nil {
		if err.Error() != "mongo: no documents in result" {
			log.Println("CreatePortalUser: Error while checking for existing user - ", err.Error())
			respDto.ErrorDescription = "Error while checking for existing user"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	}
	// user exists
	if checkExisting.UserId == userInstance.UserId {
		log.Println("CreatePortalUser: User exists")
		respDto.ErrorDescription = "User Already Present in DB"
		respDto.UserDetail = checkExisting
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//Insert In DB
	err = database.InsertPortalUserDetails(userInstance)
	if err != nil {
		respDto.ErrorDescription = "Couldn't create user"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	//Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "Created user"
	respDto.UserDetail = userInstance
	return c.JSON(respDto)
}

//Get All users list - Only for OA
func GetUsersList(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetUsersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	//Authenticate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		//Check for error and send 400 and error.
		log.Println("GetUsersList: Token Authentication Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	//Check Token User Role
	if !IsApplicable(Tknmeta, "getusers") {
		log.Println("GetUsersList: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetUsersList: Request Body is - ", reqStr)
	reqDto := portaldto.GetUsersReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetUsersList: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	users, err := database.GetB2BUsers(Tknmeta.OperatorId, reqDto.PartialUserName)
	if err != nil {
		log.Println("GetUsersList: Get All users Failed for given Operator - ", err.Error())
		respDto.ErrorDescription = "Unable To get All Users"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(users)
	log.Println("GetUsersList: Users Count: ", totalRecords)
	respDto.Page = 1 // default is page 1
	if reqDto.Page > 1 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize != 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}
	count := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Users = append(respDto.Users, GetUser(users[itr]))
		count++
	}
	respDto.PageSize = count
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get Operarotors List - Only SA
// GetOperators is a function to get list of all operators
// @Summary      Portal Get All Operators
// @Description  Get All Operators
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer Token"
// @Success      200            {object}  portaldto.GetOperatorsRespDto
// @Failure      503            {object}  portaldto.GetOperatorsRespDto
// @Router       /portal/get-operators [post]
func GetOperators(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetOperatorsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Operators = []dto.Operator{}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetOperators: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getoperators") {
		log.Println("GetOperators: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operators
	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("GetOperators: Get Operators failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//log.Println("GetOperators: operators count is - ", len(operators))
	// 5. Iterate throug all operators
	for _, operator := range operators {
		_, err := json.Marshal(operator)
		if err != nil {
			log.Println("GetOperators: OperatorId is - ", operator.OperatorId)
			log.Println("GetOperators: json.Marshal failed with error - ", err.Error())
			continue
		}
		//log.Println("GetOperators: Operator JSON is - ", string(operBytes))
		respOper := GetOperator(operator)
		_, err = json.Marshal(respOper)
		if err != nil {
			log.Println("GetOperators: OperatorId is - ", operator.OperatorId)
			log.Println("GetOperators: json.Marshal failed with error - ", err.Error())
			continue
		}
		//log.Println("GetOperators: resp operator JSON is - ", string(respOpBytes))
		respDto.Operators = append(respDto.Operators, respOper)
	}
	//log.Println("GetOperators: respDto.Operators count is - ", len(respDto.Operators))
	_, err = json.Marshal(respDto)
	if err != nil {
		log.Println("GetOperators: json.Marshal failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//log.Println("GetOperators: resp JSON is - ", string(respBytes))
	// 6. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get Operarotors List - Only SA
func GetPartners(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetPartnersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetPartners: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getpartners") {
		log.Println("GetPartners: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operators
	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("GetPartners: Get Operators failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Iterate throug all operators
	for _, operator := range operators {
		for _, partner := range operator.Partners {
			respDto.Partners = append(respDto.Partners, GetPartner(operator, partner))
		}
	}
	// 6. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get Providers List - Only SA
func GetProviders(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetProvidersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetOperators: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getProviders") {
		log.Println("GetOperators: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operators
	providers, err := database.GetAllProviders()
	if err != nil {
		log.Println("GetOperators: Get Operators failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Iterate through all operators
	// for _, provider := range providers {
	// 	respDto.Providers = append(respDto.Providers, provider)
	// }
	// 6. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Providers = providers
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GetProviderStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetProviderStatusRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(portaldto.GetProviderStatusReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetProviderStatus: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetOperators: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getProviderStatus") {
		log.Println("GetOperators: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operators
	log.Println(reqDto)
	partnerStatus, err := database.GetPartnersByProviderId(reqDto.ProviderId)
	if err != nil {
		log.Println("GetOperators: GetPartnersByProviderId failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.PartnerStatus = partnerStatus
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Provider Status
func BlockProviderStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1.1. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	//1.2 Get Request
	reqDto := portaldto.BlockProviderReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockProviderStatus: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockProviderStatus: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "updateProviderStatus") {
		log.Println("BlockProviderStatus: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Update the Status in DB and Cache
	err = database.UpdatePartnerProviderStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId, BLOCKED)
	if err != nil {
		// 4.1. Failed to update operator in database
		log.Println("BlockProviderStatus: UpdatePartnerProviderStatus failed with error - ", err.Error())
	} else {
		// 4.2 Update Cache only if there is not error in saving data in db
		ps, err := cache.GetPartnerStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
		if err != nil {
			log.Println("BlockProviderStatus: Cache Update failed with error - ", err.Error())
		}
		ps.ProviderStatus = BLOCKED
		cache.SetPartnerStatus(ps)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PROVIDER " + BLOCKED
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Activate Provider Status
func UnblockProviderStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1.1. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	//1.2 Get Request
	reqDto := portaldto.BlockProviderReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockProviderStatus: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockProviderStatus: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "updateProviderStatus") {
		log.Println("UnblockProviderStatus: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Update the Status in DB and Cache
	err = database.UpdatePartnerProviderStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId, ACTIVE)
	if err != nil {
		// 4.1. Failed to update operator in database
		log.Println("UnblockProviderStatus: UpdatePartnerProviderStatus failed with error - ", err.Error())
	} else {
		// 4.2 Update Cache only if there is not error in saving data in db
		ps, err := cache.GetPartnerStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
		if err != nil {
			log.Println("UnblockProviderStatus: Cache Update failed with error - ", err.Error())
		}
		ps.ProviderStatus = ACTIVE
		cache.SetPartnerStatus(ps)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PROVIDER " + ACTIVE
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Operator Status
// func BlockOperatorStatus(c *fiber.Ctx) error {
// 	c.Accepts("json", "text")
// 	c.Accepts("application/json")
// 	// 1.1. Create Default Response Object
// 	respDto := portaldto.CommonPortalRespDto{}
// 	respDto.ErrorDescription = GENERAL_ERROR_DESC
// 	respDto.Status = ERROR_STATUS

// 	//1.2 Get Request
// 	reqDto := portaldto.BlockOperatorReqDto{}
// 	err := c.BodyParser(&reqDto)
// 	if err != nil {
// 		// 4.1. Parsing failed
// 		log.Println("BlockOperatorStatus: Body Parsing failed with error - ", err.Error())
// 		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}

// 	// 2. Validation Token
// 	Tknmeta, ok := Authenticate(c)
// 	if !ok {
// 		// 2.1. Token validaton failed.
// 		log.Println("BlockOperatorStatus: Token Validation Failed")
// 		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	// 3. Check Role Permissions
// 	if !IsApplicable(Tknmeta, "updateProviderStatus") {
// 		log.Println("BlockOperatorStatus: User not Permitted to access the API")
// 		respDto.ErrorDescription = UNAUTH_ACCESS
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	// 4. Update the Status in DB and Cache
// 	err = database.UpdateOAOperatorStatus(reqDto.OperatorId, reqDto.ProviderId, BLOCKED)
// 	if err != nil {
// 		// 4.1. Failed to update operator in database
// 		log.Println("BlockOperatorStatus: Database Update failed with error - ", err.Error())
// 	} else {
// 		// 4.2 Update Cache only if there is not error in saving data in db
// 		ps, err := cache.GetProviderStatus(reqDto.OperatorId, reqDto.ProviderId)
// 		if err != nil {
// 			log.Println("BlockOperatorStatus: Cache Update failed with error - ", err.Error())
// 		}
// 		ps.ProviderStatus = BLOCKED
// 		cache.SetProviderStatus(ps)
// 	}
// 	respDto.Status = OK_STATUS
// 	respDto.ErrorDescription = "OPERATOR " + BLOCKED
// 	return c.Status(fiber.StatusOK).JSON(respDto)
// }

// // Activate Operator Status
// func UnblockOperatorStatus(c *fiber.Ctx) error {
// 	c.Accepts("json", "text")
// 	c.Accepts("application/json")
// 	// 1.1. Create Default Response Object
// 	respDto := portaldto.CommonPortalRespDto{}
// 	respDto.ErrorDescription = GENERAL_ERROR_DESC
// 	respDto.Status = ERROR_STATUS

// 	//1.2 Get Request
// 	reqDto := portaldto.BlockProviderReqDto{}
// 	err := c.BodyParser(&reqDto)
// 	if err != nil {
// 		// 4.1. Parsing failed
// 		log.Println("UnblockOperatorStatus: Body Parsing failed with error - ", err.Error())
// 		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	// 2. Validation Token
// 	Tknmeta, ok := Authenticate(c)
// 	if !ok {
// 		// 2.1. Token validaton failed.
// 		log.Println("UnblockOperatorStatus: Token Validation Failed")
// 		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	// 3. Check Role Permissions
// 	if !IsApplicable(Tknmeta, "updateProviderStatus") {
// 		log.Println("UnblockOperatorStatus: User not Permitted to access the API")
// 		respDto.ErrorDescription = UNAUTH_ACCESS
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	// 4. Update the Status in DB and Cache
// 	err = database.UpdateOAOperatorStatus(reqDto.OperatorId, reqDto.ProviderId, ACTIVE)
// 	if err != nil {
// 		// 4.1. Failed to update operator in database
// 		log.Println("UnblockOperatorStatus: Database Update failed with error - ", err.Error())
// 	} else {
// 		// 4.2 Update Cache only if there is not error in saving data in db
// 		ps, err := cache.GetProviderStatus(reqDto.OperatorId, reqDto.ProviderId)
// 		if err != nil {
// 			log.Println("UnblockOperatorStatus: Cache Update failed with error - ", err.Error())
// 		}
// 		ps.ProviderStatus = ACTIVE
// 		cache.SetProviderStatus(ps)
// 	}
// 	respDto.Status = OK_STATUS
// 	respDto.ErrorDescription = "OPERATOR " + ACTIVE
// 	return c.Status(fiber.StatusOK).JSON(respDto)
// }

func ReplaceOperators(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	//1.2 Get Request
	reqDto := operatordto.OperatorDTO{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockOperatorStatus: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockOperator: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "updateOperators") {
		log.Println("BlockOperator: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//5. Get Operator - needs to be updated
	operatorDTO, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		log.Println("BlockOperator: Database Get failed with error - ", err.Error())
	}
	updateOperatorDTOFromRequest(&operatorDTO, reqDto)
	// 4. Update the Status in DB and Cache
	err = database.ReplaceOperator(operatorDTO)
	if err != nil {
		// 4.1. Failed to update operator in database
		log.Println("UnblockOperatorStatus: Database Update failed with error - ", err.Error())
	} else {
		// 4.2 Update Cache only if there is not error in saving data in db
		cache.SetOperatorDetails(operatorDTO)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "Operator Details Replaced"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Operator - Only SA
func BlockOperator(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.BlockOperatorRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockOperator: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockoperator") {
		log.Println("BlockOperator: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockOperatorReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockOperator: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.OperatorId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockOperator: OperatorId is empty")
		respDto.ErrorDescription = "OperatorId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdateOperatorStatus(reqDto.OperatorId, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockOperator: Operator Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Operator Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	operatorDto, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockOperator: Get operator failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetOperatorDetails(operatorDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "OPERATOR " + BLOCKED
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Operator - Only SA
func UnblockOperator(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.UnblockOperatorRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockOperator: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockoperator") {
		log.Println("UnblockOperator: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.UnblockOperatorReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockOperator: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.OperatorId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockOperator: OperatorId is empty")
		respDto.ErrorDescription = "OperatorId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdateOperatorStatus(reqDto.OperatorId, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockOperator: Operator Unblocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Operator Unblocking Failed"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	operatorDto, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockOperator: Get operator failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetOperatorDetails(operatorDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "OPERATOR " + ACTIVE
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Partner - Only SA
func BlockPartner(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.BlockOperatorRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockPartner: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockpartner") {
		log.Println("BlockPartner: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockPartnerReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockPartner: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.OperatorId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockPartner: OperatorId is empty")
		respDto.ErrorDescription = "OperatorId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.PartnerId == "" {
		// 4.2. PartnerId missing
		log.Println("BlockPartner: PartnerId is empty")
		respDto.ErrorDescription = "PartnerId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdatePartnerStatus(reqDto.OperatorId, reqDto.PartnerId, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockPartner: Partner Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Partner Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	operatorDto, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockPartner: Get operator failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetOperatorDetails(operatorDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "OPERATOR " + BLOCKED
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Partner - Only SA
func UnblockPartner(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.UnblockOperatorRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockPartner: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockpartner") {
		log.Println("UnblockPartner: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.UnblockPartnerReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockPartner: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.OperatorId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockPartner: OperatorId is empty")
		respDto.ErrorDescription = "OperatorId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.PartnerId == "" {
		// 4.2. PartnerId missing
		log.Println("UnblockPartner: PartnerId is empty")
		respDto.ErrorDescription = "PartnerId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdatePartnerStatus(reqDto.OperatorId, reqDto.PartnerId, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockPartner: Partner Unblocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Partner Unblocking Failed"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	operatorDto, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockPartner: Get operator failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetOperatorDetails(operatorDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "OPERATOR " + ACTIVE
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get Providers List - Only SA
func GetSAProviders(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetProvidersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetSAProviders: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getsaproviders") {
		log.Println("GetSAProviders: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Providers
	providers, err := database.GetAllProviders()
	if err != nil {
		log.Println("GetSAProviders: Get Providers failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Iterate through providers
	// for _, provider := range providers {
	// 	respDto.Providers = append(respDto.Providers, GetSAProvider(provider))
	// }
	// 6. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Providers = providers
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Provider
func BlockSAProvider(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.BlockProviderRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockProvider: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blocksaprovider") {
		log.Println("BlockProvider: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockProviderReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockProvider: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockProvider: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdateProviderStatus(reqDto.ProviderId, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockProvider: Provider Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Provider Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	provider, err := database.GetProvider(reqDto.ProviderId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockProvider: Get provider failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetProvider(provider)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PROVIDER " + BLOCKED
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Provider
func UnblockSAProvider(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.UnblockProviderRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockProvider: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblocksaprovider") {
		log.Println("UnblockProvider: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockProviderReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockProvider: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockProvider: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdateProviderStatus(reqDto.ProviderId, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockProvider: Provider Unblocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Provider Unblocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	provider, err := database.GetProvider(reqDto.ProviderId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockProvider: Get provider failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetProvider(provider)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PROVIDER " + ACTIVE
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get OA Providers List - Only OA
func GetOAProviders(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetOAProvidersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetOAProviders: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getoaproviders") {
		log.Println("GetOAProviders: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.0. Get Operator Details
	operatorId := Tknmeta.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetOAProviders: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.1. Get SA All Providers
	reqStr := string(c.Body())
	log.Println("GetOAProviders: Request Body is - ", reqStr)
	saProviders, err := database.GetAllProviders()
	if err != nil {
		log.Println("GetOAProviders: GetAllProviders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.2. List SA blocked provider
	blockedProviders := []models.Provider{}
	for _, provider := range saProviders {
		if provider.Status != ACTIVE {
			blockedProviders = append(blockedProviders, provider)
		}
	}
	// 4.3. Get OA All Providers
	oaProviders, err := database.GetProvidersPS(Tknmeta.OperatorId)
	if err != nil {
		log.Println("GetOAProviders: GetProviders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.4. Filter SA Blocked providers from OA Providers
	for _, oaprovider := range oaProviders {

		// filter, if provider status is not active
		if oaprovider.ProviderStatus != ACTIVE {
			continue
		}
		isBlocked := false
		// iterate through all sa blocked providers
		for _, saprovider := range blockedProviders {
			// is sa blocked provider = oa provider, then mark isBlocked true and break the loop
			if saprovider.ProviderId == oaprovider.ProviderId {
				isBlocked = true
				break
			}
		}
		if isBlocked {
			continue
		}
		respDto.Providers = append(respDto.Providers, GetProvider(oaprovider, operatorDto))
	}
	// 5. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get OA Providers List - Only OA
func GetProvidersForTabs(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetOAProvidersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetOAProviders: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getoaproviders") {
		log.Println("GetOAProviders: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.0. Get Operator Details
	operatorId := Tknmeta.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetOAProviders: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.1. Get SA All Providers
	reqStr := string(c.Body())
	log.Println("GetOAProviders: Request Body is - ", reqStr)
	saProviders, err := database.GetAllProviders()
	if err != nil {
		log.Println("GetOAProviders: GetAllProviders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.2. List SA blocked provider
	blockedProviders := []models.Provider{}
	for _, provider := range saProviders {
		if provider.Status != ACTIVE {
			blockedProviders = append(blockedProviders, provider)
		}
	}
	// 4.3. Get OA All Providers
	oaProviders, err := database.GetProvidersPS(Tknmeta.OperatorId)
	if err != nil {
		log.Println("GetOAProviders: GetProviders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// List of Providers Added to Response
	addedProvider := map[string]bool{}
	// 4.4. Filter SA Blocked providers from OA Providers
	for _, oaprovider := range oaProviders {

		// 4.4.1. Check if provider is present in response
		if addedProvider[oaprovider.ProviderId] {
			continue
		}

		// filter, if provider status is not active
		if oaprovider.ProviderStatus != ACTIVE {
			continue
		}
		isBlocked := false
		// iterate through all sa blocked providers
		for _, saprovider := range blockedProviders {
			// is sa blocked provider = oa provider, then mark isBlocked true and break the loop
			if saprovider.ProviderId == oaprovider.ProviderId {
				isBlocked = true
				break
			}
		}
		if isBlocked {
			continue
		}
		respDto.Providers = append(respDto.Providers, GetProvider(oaprovider, operatorDto))
		addedProvider[oaprovider.ProviderId] = true
	}
	// 5. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block OA Provider - only OA
func BlockOAProvider(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.BlockProviderRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockOAProvider: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockoaprovider") {
		log.Println("BlockOAProvider: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockProviderReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockOAProvider: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockOAProvider: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdatePartnerOperatorStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockOAProvider: UpdatePartnerOperatorStatus failed with error - ", err.Error())
		respDto.ErrorDescription = "Provider Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache

	provider, err := database.GetProvider(reqDto.ProviderId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockOAProvider: Get provider failed with error - ", err.Error())
	} else {
		// 6.1. Update Cache
		cache.SetProvider(provider)
	}

	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PROVIDER " + BLOCKED
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock OA Provider - only OA
func UnblockOAProvider(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.UnblockProviderRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockOAProvider: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockoaprovider") {
		log.Println("UnblockOAProvider: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockProviderReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockOAProvider: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockOAProvider: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	err = database.UpdatePartnerOperatorStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockOAProvider: UpdatePartnerOperatorStatus failed with error - ", err.Error())
		respDto.ErrorDescription = "Provider Unblocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	/*
		provider, err := database.GetProvider(reqDto.ProviderId)
		if err != nil {
			// 6.1. Failed to update operator in database
			log.Println("UnblockOAProvider: Get provider failed with error - ", err.Error())
		} else {
			// 6.1. Update Cache
			cache.SetProvider(provider)
		}
	*/
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PROVIDER " + ACTIVE
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block User - only OA
func BlockUser(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockUser: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockuser") {
		log.Println("BlockUser: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockUserReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockUser: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.UserId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockUser: UserId is empty")
		respDto.ErrorDescription = "UserId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	userKey := Tknmeta.OperatorId + "-" + reqDto.UserId
	err = database.UpdateB2BUserStatus(userKey, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockUser: User Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "User Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	/*
		provider, err := database.GetProvider(reqDto.ProviderId)
		if err != nil {
			// 6.1. Failed to update operator in database
			log.Println("BlockUser: Get provider failed with error - ", err.Error())
		} else {
			// 6.1. Update Cache
			cache.SetProvider(provider)
		}
	*/
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "USER " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.UserId, constants.USER)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// type Audit struct {
// 	UserId     string `bson:"user_id"`     // UserId of the user who performed the action
// 	Operation  string `bson:"operation"`   // Operation performed by the user
// 	IP         string `bson:"ip"`          // IP address of the user who performed the action
// 	OperatorId string `bson:"operator_id"` // OperatorId of the user who performed the action
// 	Time       string `bson:"time"`        // Time when the action was performed
// 	ObjectId   string `bson:"object_id"`   // ObjectId of the object on which the action was performed
// 	ObjectType string `bson:"object_type"` // ObjectType of the object on which the action was performed
// 	Payload    string `bson:"payload"`     // Payload of the action
// 	UserRole   string `bson:"user_role"`   // UserRole of the user who performed the action
// }

// Unblock User - only OA
func UnblockUser(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockUser: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockuser") {
		log.Println("UnblockUser: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockUserReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockUser: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.UserId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockUser: UserId is empty")
		respDto.ErrorDescription = "UserId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Update Database
	userKey := Tknmeta.OperatorId + "-" + reqDto.UserId
	err = database.UpdateB2BUserStatus(userKey, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockUser: User Unblocking failed with error - ", err.Error())
		respDto.ErrorDescription = "User Unblocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Cache
	/*
		provider, err := database.GetProvider(reqDto.ProviderId)
		if err != nil {
			// 6.1. Failed to update operator in database
			log.Println("UnblockUser: Get provider failed with error - ", err.Error())
		} else {
			// 6.1. Update Cache
			cache.SetProvider(provider)
		}
	*/
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "USER " + ACTIVE
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.UserId, constants.USER)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Sports for SAP - only SAP
// @Summary      Get Sports for SAP
// @Description  List All Sports of an SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization        header    string                  true  "Bearer Token"
// @Param        GetSportsListForSAP  body      portaldto.SportsReqDto  true  "SportsReqDto model is used"
// @Success      200                  {object}  portaldto.SportsRespDto
// @Failure      503                  {object}  portaldto.SportsRespDto
func GetSportsListForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.SportsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetSportsListForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getSportsListForSAP") {
		log.Println("GetSportsListForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.SportsReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetSportsListForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("GetSportsListForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	sports, err := database.GetSports(reqDto.ProviderId)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("GetSportsListForSAP: Sports Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Getting Sports List Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	// Convert Sports.model to sports.dto
	for _, sportsDTO := range sports {
		sport := dto.Sport{}
		sport.SportName = sportsDTO.SportName
		sport.SportId = sportsDTO.SportId
		sport.Status = sportsDTO.Status
		respDto.Sports = append(respDto.Sports, sport)
	}

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Sports For SAP - only SAP
// @Summary      Block Sports For SAP
// @Description  Block Sports For SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization     header    string                        true  "Bearer Token"
// @Param        BlockSportForSAP  body      portaldto.BlockedSportReqDto  true  "BlockedSportReqDto model is used"
// @Success      200               {object}  portaldto.BlockedSportRespDto
// @Failure      503               {object}  portaldto.BlockedSportRespDto
func BlockSportForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedSportRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockSportsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockSportsForSAP") {
		log.Println("BlockSportsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedSportReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockSportsForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockSportsForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockSportsForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	sportModel := models.Sport{}
	sportModel.SportKey = reqDto.ProviderId + "-" + reqDto.SportId
	sportModel.Status = BLOCKED
	err = database.UpdateSport(sportModel)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockSportsForSAP: Get Sport failed with error - ", err.Error())
		respDto.ErrorDescription = "Sport Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	err = database.UpdatePASportStatus(reqDto.ProviderId+"-"+reqDto.SportId, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockSportsForSAP: Get Sport failed with error - ", err.Error())
		respDto.ErrorDescription = "Sport Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "SPORT " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.SportId, constants.SPORT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Sport For SAP - only SAP
// @Summary      Unblock Sport For SAP
// @Description  Unblock Sport For SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                          true  "Bearer Token"
// @Param        UnblockSportForSAP  body      portaldto.UnblockedSportReqDto  true  "UnblockedSportReqDto model is used"
// @Success      200                 {object}  portaldto.UnblockedSportRespDto
// @Failure      503                 {object}  portaldto.UnblockedSportRespDto
func UnblockSportForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UnblockedSportRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockSportsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockSportsForSAP") {
		log.Println("UnblockSportsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.UnblockedSportReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockSportsForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockSportsForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockSportsForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	sportModel := models.Sport{}
	sportModel.SportKey = reqDto.ProviderId + "-" + reqDto.SportId
	sportModel.Status = ACTIVE
	err = database.UpdateSport(sportModel)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockSportsForSAP: Get Sport failed with error - ", err.Error())
		respDto.ErrorDescription = "Sport Activation Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	err = database.UpdatePASportStatus(reqDto.ProviderId+"-"+reqDto.SportId, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockSportsForSAP: Get Sport failed with error - ", err.Error())
		respDto.ErrorDescription = "Sport Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "SPORT " + ACTIVE
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.SportId, constants.SPORT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Sports for OA - only Operators
// @Summary      Get Sports
// @Description  List All Sports of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                  true  "Bearer Token"
// @Param        GetSportsListForOP  body      portaldto.SportsReqDto  true  "SportsReqDto model is used"
// @Success      200                 {object}  portaldto.SportsRespDto
// @Failure      503                 {object}  portaldto.SportsRespDto
// @Router       /portal/opadmin/sports [post]
func GetSportsListForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.SportsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetSportsListForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getSportsListForOP") {
		log.Println("GetSportsListForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4 Get Request
	reqDto := portaldto.SportsReqDto{}
	reqStr := string(c.Body())
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetSportsListForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5.1. Get SA All Sports
	log.Println("GetOASports: Request Body is - ", reqStr)
	saSports, err := database.GetSports(reqDto.ProviderId)
	if err != nil {
		log.Println("GetSportsListForOP: GetSports failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.2. List SA blocked Sports
	blockedSports := []models.Sport{}
	for _, sport := range saSports {
		if sport.Status != ACTIVE {
			blockedSports = append(blockedSports, sport)
		}
	}
	// 4.3. Get OA All Sports
	opPrSports, err := database.GetOpPrSports(Tknmeta.OperatorId, "", reqDto.ProviderId)
	if err != nil {
		log.Println("GetSportsListForOP: GetSports failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4.4. Filter SA Blocked Sports from OA Sports
	for _, oaSport := range opPrSports {
		// filter, if provider status is not active
		if oaSport.ProviderStatus != ACTIVE {
			continue
		}
		isBlocked := false
		// iterate through all sa blocked providers
		for _, sasport := range blockedSports {
			// is sa blocked provider = oa provider, then mark isBlocked true and break the loop
			if oaSport.SportId == sasport.SportId {
				isBlocked = true
				break
			}
		}
		if isBlocked {
			continue
		}
		respDto.Sports = append(respDto.Sports, GetSportStatus(oaSport))
	}
	// 5. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

//Get Operarotor details List - Only OP
func GetOperatorDetails(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetOperatorDetailsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetOperatorDetails: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetOperatorDetails: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator details
	operatorDetails, err := database.GetOperatorDetails(Tknmeta.OperatorId)
	if err != nil {
		log.Println("GetOperatorDetails: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// TODO: Satya modified this. Need to fix this based on the functional requirement
	respDto.Operator = append(respDto.Operator, GetOperator(operatorDetails))
	// for _, partners := range operatorDetails.Partners {
	// 	respDto.Operator = append(respDto.Operator, GetOperator(operatorDetails, partners))
	// }

	// 5. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Sports For OP - only OP
// @Summary      Block Sports For OP
// @Description  Block Sports For OP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization     header    string                        true  "Bearer Token"
// @Param        BlockSportsForOP  body      portaldto.BlockedSportReqDto  true  "BlockedSportReqDto model is used"
// @Success      200               {object}  portaldto.BlockedSportRespDto
// @Failure      503               {object}  portaldto.BlockedSportRespDto
// @Router       /portal/opadmin/block-sport [post]
func BlockSportsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedSportRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockSportsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockSportsForOP") {
		log.Println("BlockSportsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedSportReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockSportsForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockSportsForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockSportsForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	err = database.UpdateOASportStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId, BLOCKED)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockSportsForSAP: Get Sport failed with error - ", err.Error())
		respDto.ErrorDescription = "Sport Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "SPORT " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.SportId, constants.SPORT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Sport For OP - only OP
// @Summary      Unblock Sport For OP
// @Description  Unblock Sport For OP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                          true  "Bearer Token"
// @Param        UnblockSportsForOP  body      portaldto.UnblockedSportReqDto  true  "UnblockedSportReqDto model is used"
// @Success      200                 {object}  portaldto.UnblockedSportRespDto
// @Failure      503                 {object}  portaldto.UnblockedSportRespDto
// @Router       /portal/opadmin/unblock-sport [post]
func UnblockSportsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedSportRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockSportsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockSportsForOP") {
		log.Println("UnblockSportsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedSportReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockSportsForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockSportsForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockSportsForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	err = database.UpdateOASportStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId, ACTIVE)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockSportsForSAP: Get Sport failed with error - ", err.Error())
		respDto.ErrorDescription = "Sport Activation Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "SPORT " + ACTIVE
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.SportId, constants.SPORT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Competitions for SAP - only SAP
// @Summary      Get Competitions
// @Description  List All Competitions of an SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization              header    string                true  "Bearer Token"
// @Param        GetCompetitionsListForSAP  body      portaldto.CompReqDto  true  "CompReqDto model is used"
// @Success      200                        {object}  portaldto.CompRespDto
// @Failure      503                        {object}  portaldto.CompRespDto
// @Router       /portal/sapadmin/compititions [post]
func GetCompetitionsListForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.CompRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetCompetitionsListForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getCompetitionsListForSAP") {
		log.Println("GetCompetitionsListForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.CompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetCompetitionsListForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("GetCompetitionsListForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockCompetitionForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	competitions, err := database.GetCompetitionsbySport(reqDto.ProviderId, reqDto.SportId)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("GetCompetitionsListForSAP: Get Competition List failed with error - ", err.Error())
		respDto.ErrorDescription = "Get Competition List Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	// Convert Comprtitions.model to Comprtitions.dto
	for _, competitionsDTO := range competitions {
		competition := dto.Competition{}
		competition.CompetitionId = competitionsDTO.CompetitionId
		competition.CompetitionName = competitionsDTO.CompetitionName
		competition.Status = competitionsDTO.Status
		respDto.Competitions = append(respDto.Competitions, competition)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Recent Competitions for SAP - only SAP
// @Summary      Get Recent Competitions
// @Description  List All Recent Competitions of an SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization              header    string                true  "Bearer Token"
// @Param        GetCompetitionsListForSAP  body      portaldto.CompReqDto  true  "CompReqDto model is used"
// @Success      200                        {object}  portaldto.CompRespDto
// @Failure      503                        {object}  portaldto.CompRespDto
// @Router       /portal/sapadmin/recent-compititions [post]
func RecentCompetitionsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.CompRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("RecentCompetitions: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("RecentCompetitions: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.CompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("RecentCompetitions: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("RecentCompetitions: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("RecentCompetitions: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6 Get Events for Layer 2
	var events []coredto.EventDto
	switch reqDto.ProviderId {
	case "Dream":
		events, err = dream.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("RecentCompetitions: Get Events failed for Dream with error - ", err.Error())
			respDto.ErrorDescription = "Get Events Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case "BetFair":
		events, err = betfair.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("RecentCompetitions: Get Events failed for BetFair with error - ", err.Error())
			respDto.ErrorDescription = "Get Events Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case "SportRadar":
		events, err = sportradar.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("RecentCompetitions: Get Events failed for SportRadar with error - ", err.Error())
			respDto.ErrorDescription = "Get Events Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	default:
		log.Println("RecentCompetitions: ProviderId is invalid")
		respDto.ErrorDescription = "ProviderId is invalid!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	competitionkeys := []string{}
	for _, event := range events {
		// if event.CompetitionId == "" || event.CompetitionId == "-1" {
		// 	continue
		// }
		competitionKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + event.CompetitionId
		competitionkeys = append(competitionkeys, competitionKey)
	}
	// 7. Get Database
	competitions, err := database.GetCompetitionsByKeys(competitionkeys)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("GetCompetitionsListForSAP: Get Competition List failed with error - ", err.Error())
		respDto.ErrorDescription = "Get Competition List Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	// 8 Convert Comprtitions.model to Comprtitions.dto
	for _, competitionsDTO := range competitions {
		competition := dto.Competition{}
		competition.CompetitionId = competitionsDTO.CompetitionId
		competition.CompetitionName = competitionsDTO.CompetitionName
		competition.Status = competitionsDTO.Status
		respDto.Competitions = append(respDto.Competitions, competition)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Competitions for SAP - only SAP
// @Summary      Block Competitions
// @Description  Block Competitions of an SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization           header    string                       true  "Bearer Token"
// @Param        BlockCompetitionForSAP  body      portaldto.BlockedCompReqDto  true  "BlockedCompReqDto model is used"
// @Success      200                     {object}  portaldto.BlockedCompResqDto
// @Failure      503                     {object}  portaldto.BlockedCompResqDto
// @Router       /portal/sapadmin/block-competition [post]
func BlockCompetitionForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedCompResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockCompetitionForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockCompetitionForSAP") {
		log.Println("BlockCompetitionForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedCompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockCompetitionForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockCompetitionForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockCompetitionForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	competition := models.Competition{}
	competition.CompetitionKey = reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
	competition.Status = BLOCKED
	err = database.UpdateCompetition(competition)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("BlockCompetitionForSAP: Competition Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Competition Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err = database.UpdatePACompetitionStatus(competition.CompetitionKey, BLOCKED)
	// if err != nil {
	// 	// 5.1. Failed to update operator in database
	// 	log.Println("BlockCompetitionForSAP: Competition Blocking failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "Competition Blocking Failed!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "COMPETITION " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.CompetitionId, constants.COMPETITION)
	}
	// Convert Comprtitions.model to Comprtitions.dto
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Competitions for SAP - only SAP
// @Summary      Unblock Competitions
// @Description  Unblock Competitions of an SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization             header    string                         true  "Bearer Token"
// @Param        UnblockCompetitionForSAP  body      portaldto.UnblockedCompReqDto  true  "UnblockedCompReqDto model is used"
// @Success      200                       {object}  portaldto.UnblockedCompResqDto
// @Failure      503                       {object}  portaldto.UnblockedCompResqDto
// @Router       /portal/sapadmin/unblock-competition [post]
func UnblockCompetitionForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UnblockedCompResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockCompetitionForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockCompetitionForSAP") {
		log.Println("UnblockCompetitionForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.UnblockedCompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockCompetitionForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockCompetitionForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockCompetitionForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	competition := models.Competition{}
	competition.CompetitionKey = reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
	competition.Status = ACTIVE
	err = database.UpdateCompetition(competition)
	if err != nil {
		// 5.1. Failed to update operator in database
		log.Println("UnblockCompetitionForSAP: User Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Competition activation Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err = database.UpdatePACompetitionStatus(competition.CompetitionKey, ACTIVE)
	// if err != nil {
	// 	// 5.1. Failed to update operator in database
	// 	log.Println("UnblockCompetitionForSAP: User Blocking failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "Competition activation Failed!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "COMPETITION " + ACTIVE
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.CompetitionId, constants.COMPETITION)
	}
	// Convert Comprtitions.model to Comprtitions.dto
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Competitions for OP - only OP
// @Summary      Get Competitions
// @Description  List All Competitions of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization             header    string                true  "Bearer Token"
// @Param        GetCompetitionsListForOP  body      portaldto.CompReqDto  true  "CompReqDto model is used"
// @Success      200                       {object}  portaldto.CompRespDto
// @Failure      503                       {object}  portaldto.CompRespDto
// @Router       /portal/opadmin/compititions [post]
func GetCompetitionsListForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.CompRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetCompetitionsListForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getCompetitionsListForOP") {
		log.Println("GetCompetitionsListForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.CompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetCompetitionsListForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("GetCompetitionsListForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("GetCompetitionsListForOP: ProviderId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5.1. Get SA All competitions
	competitions, err := database.GetCompetitionsbySport(reqDto.ProviderId, reqDto.SportId)
	if err != nil {
		// 5.1. Failed to Get competitions in database
		log.Println("GetCompetitionsListForOP: Get Competitions failed with error - ", err.Error())
		respDto.ErrorDescription = "Get competitions list for OP Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// // 6.1. List SA blocked competitions
	// blockedCompetitions := []models.Competition{}
	// for _, competition := range competitions {
	// 	if competition.Status != ACTIVE {
	// 		blockedCompetitions = append(blockedCompetitions, competition)
	// 	}
	// }
	// 6.2. Get OA All competitions
	opPrCompetitions, err := database.GetOpPrCompetitions(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId)
	if err != nil {
		log.Println("GetCompetitionsListForOP: GetCompetitions failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	for _, competition := range competitions {
		if competition.Status != ACTIVE {
			// Platform Level Blocked, dont show in the list
			continue
		}
		isBlocked := false
		for _, opcompetition := range opPrCompetitions {
			if opcompetition.CompetitionId == competition.CompetitionId {
				if opcompetition.ProviderStatus == BLOCKED {
					// Platfor Admin Blocked for this operator, dont show in the list
					isBlocked = true
				}
				if opcompetition.OperatorStatus == BLOCKED {
					// Operator Admin blocked, show as disabled
					competition.Status = BLOCKED
				}
				break
			}
		}
		if isBlocked == true {
			// Platfor Admin Blocked for this operator, dont show in the list
			continue
		}
		respDto.Competitions = append(respDto.Competitions, GetCompetitionsStatus2(competition))
	}

	// 6.3. Filter SA Blocked competitions from OA competitions
	// for _, opPrCompetition := range opPrCompetitions {
	// 	// filter, if provider status is not active
	// 	if opPrCompetition.ProviderStatus != ACTIVE {
	// 		continue
	// 	}
	// 	isBlocked := false
	// 	// iterate through all sa blocked competitions
	// 	for _, saCompetition := range blockedCompetitions {
	// 		// is sa blocked competitions = oa competitions, then mark isBlocked true and break the loop
	// 		if opPrCompetition.CompetitionId == saCompetition.CompetitionId {
	// 			isBlocked = true
	// 			break
	// 		}
	// 	}
	// 	if isBlocked {
	// 		continue
	// 	}
	// 	respDto.Competitions = append(respDto.Competitions, GetCompetitionsStatus(opPrCompetition))
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Recent Competitions for OP - only OP
// @Summary      Get Recent Competitions
// @Description  List All Recent Competitions of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization             header    string                true  "Bearer Token"
// @Param        GetCompetitionsListForOP  body      portaldto.CompReqDto  true  "CompReqDto model is used"
// @Success      200                       {object}  portaldto.CompRespDto
// @Failure      503                       {object}  portaldto.CompRespDto
// @Router       /portal/opadmin/recent-compititions [post]
func RecentCompetitionsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.CompRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetCompetitionsListForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetCompetitionsListForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.CompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetCompetitionsListForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("GetCompetitionsListForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("GetCompetitionsListForOP: ProviderId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	var events []coredto.EventDto
	switch reqDto.ProviderId {
	case "Dream":
		events, err = dream.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("RecentCompetitions: Get Events failed for Dream with error - ", err.Error())
			respDto.ErrorDescription = "Get Events Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case "BetFair":
		events, err = betfair.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("RecentCompetitions: Get Events failed for BetFair with error - ", err.Error())
			respDto.ErrorDescription = "Get Events Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case "SportRadar":
		events, err = sportradar.GetEvents(reqDto.SportId)
		if err != nil {
			log.Println("RecentCompetitions: Get Events failed for SportRadar with error - ", err.Error())
			respDto.ErrorDescription = "Get Events Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	default:
		log.Println("RecentCompetitions: ProviderId is invalid")
		respDto.ErrorDescription = "ProviderId is invalid!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	competitionkeys := []string{}
	competitionStatuskeys := []string{}
	for _, event := range events {
		// if event.CompetitionId == "" || event.CompetitionId == "-1" {
		// 	continue
		// }
		competitionKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + event.CompetitionId
		competitionkeys = append(competitionkeys, competitionKey)
		competitionStatuskeys = append(competitionStatuskeys, Tknmeta.OperatorId+"-"+competitionKey)
	}

	// 5.1. Get SA All competitions
	competitions, err := database.GetCompetitionsByKeys(competitionkeys)
	if err != nil {
		// 5.1. Failed to Get competitions in database
		log.Println("GetCompetitionsListForOP: Get Competitions failed with error - ", err.Error())
		respDto.ErrorDescription = "Get competitions list for OP Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// // 6.1. List SA blocked competitions
	// blockedCompetitions := []models.Competition{}
	// for _, competition := range competitions {
	// 	if competition.Status != ACTIVE {
	// 		blockedCompetitions = append(blockedCompetitions, competition)
	// 	}
	// }

	// 6.2. Get OA All competitions
	opPrCompetitions, err := database.GetCompetitionStatusByKeys(competitionStatuskeys)
	if err != nil {
		log.Println("GetCompetitionsListForOP: GetCompetitions failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	for _, competition := range competitions {
		if competition.Status != ACTIVE {
			// Platform Level Blocked, dont show in the list
			continue
		}
		isBlocked := false
		for _, opcompetition := range opPrCompetitions {
			if opcompetition.CompetitionId == competition.CompetitionId {
				if opcompetition.ProviderStatus == BLOCKED {
					// Platfor Admin Blocked for this operator, dont show in the list
					isBlocked = true
				}
				if opcompetition.OperatorStatus == BLOCKED {
					// Operator Admin blocked, show as disabled
					competition.Status = BLOCKED
				}
				break
			}
		}
		if isBlocked == true {
			// Platfor Admin Blocked for this operator, dont show in the list
			continue
		}
		respDto.Competitions = append(respDto.Competitions, GetCompetitionsStatus2(competition))
	}

	// // 6.3. Filter SA Blocked competitions from OA competitions
	// for _, opPrCompetition := range opPrCompetitions {
	// 	// filter, if provider status is not active
	// 	if opPrCompetition.ProviderStatus != ACTIVE {
	// 		continue
	// 	}
	// 	isBlocked := false
	// 	// iterate through all sa blocked competitions
	// 	for _, saCompetition := range blockedCompetitions {
	// 		// is sa blocked competitions = oa competitions, then mark isBlocked true and break the loop
	// 		if opPrCompetition.CompetitionId == saCompetition.CompetitionId {
	// 			isBlocked = true
	// 			break
	// 		}
	// 	}
	// 	if isBlocked {
	// 		continue
	// 	}
	// 	respDto.Competitions = append(respDto.Competitions, GetCompetitionsStatus(opPrCompetition))
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Competitions for SAP - only SAP
// @Summary      Block Competitions
// @Description  Block Competitions of an SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization          header    string                       true  "Bearer Token"
// @Param        BlockCompetitionForOP  body      portaldto.BlockedCompReqDto  true  "BlockedCompReqDto model is used"
// @Success      200                    {object}  portaldto.BlockedCompResqDto
// @Failure      503                    {object}  portaldto.BlockedCompResqDto
// @Router       /portal/opadmin/block-compititions [post]
func BlockCompetitionForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedCompResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockCompetitionForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockCompetitionForOP") {
		log.Println("BlockCompetitionForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedCompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockCompetitionForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockCompetitionForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockCompetitionForOP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	competitionKey := Tknmeta.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
	competitionStats, err := database.GetCompetitionStatus(competitionKey)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockCompetitionForOP: database.GetCompetitionStatus failed with error for competitionKey - ", err.Error(), competitionKey)
		err = cache.AddCompetitionStatusOpStatus(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, BLOCKED)
		if err != nil {
			log.Println("BlockCompetitionForOP: cache.AddCompetitionStatusOpStatus failed with error for competitionKey - ", err.Error(), competitionKey)
			respDto.ErrorDescription = "Competition Blocking Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Status = OK_STATUS
		respDto.ErrorDescription = "EVENT " + BLOCKED
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	competitionStats.OperatorStatus = BLOCKED
	// err = database.UpdateOAEventStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId+"-"+reqDto.EventId, BLOCKED)
	err = database.ReplaceCompetitionStatus(competitionStats)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockEventForOP: database.ReplaceEventStatus failed with error for competitionKey - ", err.Error(), competitionKey)
		respDto.ErrorDescription = "Event Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err = database.UpdateOACompetitionStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId+"-"+reqDto.CompetitionId, BLOCKED)
	// if err != nil {
	// 	// 5.1. Failed to update operator in database
	// 	log.Println("BlockCompetitionForOP: Competition Blocking failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "Competition Blocking Failed!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "COMPETITION " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.CompetitionId, constants.COMPETITION)
	}
	// Convert Comprtitions.model to Comprtitions.dto
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Competitions for OP - only OP
// @Summary      Unblock Competitions
// @Description  Unblock Competitions of an OP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization            header    string                          true  "Bearer Token"
// @Param        UnblockCompetitionForOP  body      portaldto.UnblockedCompResqDto  true  "UnblockedCompResqDto model is used"
// @Success      200                      {object}  portaldto.UnblockedCompReqDto
// @Failure      503                      {object}  portaldto.UnblockedCompReqDto
// @Router       /portal/opadmin/unblock-compititions [post]
func UnblockCompetitionForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UnblockedCompResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockCompetitionForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockCompetitionForOP") {
		log.Println("UnblockCompetitionForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.UnblockedCompReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockCompetitionForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockCompetitionForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockCompetitionForOP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Database
	competitionKey := Tknmeta.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
	competitionStats, err := database.GetCompetitionStatus(competitionKey)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockCompetitionForOP: database.GetCompetitionStatus failed with error for competitionKey - ", err.Error(), competitionKey)
		err = cache.AddCompetitionStatusOpStatus(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, ACTIVE)
		if err != nil {
			log.Println("UnblockCompetitionForOP: cache.AddCompetitionStatusOpStatus failed with error for competitionKey - ", err.Error(), competitionKey)
			respDto.ErrorDescription = "Competition UnBlocking Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Status = OK_STATUS
		respDto.ErrorDescription = "EVENT " + ACTIVE
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	competitionStats.OperatorStatus = ACTIVE
	err = database.ReplaceCompetitionStatus(competitionStats)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockCompetitionForOP: database.ReplaceCompetitionStatus failed with error for competitionKey - ", err.Error(), competitionKey)
		respDto.ErrorDescription = "Competition UnBlocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err = database.UpdateOACompetitionStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId+"-"+reqDto.CompetitionId, ACTIVE)
	// if err != nil {
	// 	// 5.1. Failed to update operator in database
	// 	log.Println("UnblockCompetitionForOP: User Blocking failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "Competition activation Failed!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "COMPETITION " + ACTIVE

	// Convert Comprtitions.model to Comprtitions.dto
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Events for SAP - only SAP
// @Summary      Get Events
// @Description  List All Events of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization        header    string                     true  "Bearer Token"
// @Param        GetEventsListForSAP  body      portaldto.GetEventsReqDto  true  "GetEventsReqDto model is used"
// @Success      200                  {object}  portaldto.GetEventsRespDto
// @Failure      503                  {object}  portaldto.GetEventsRespDto
// @Router       /portal/sapadmin/events [post]
func GetEventsListForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.GetEventsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetEventsListForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getEventsListForSAP") {
		log.Println("GetEventsListForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.GetEventsReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetEventsListForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("GetEventsListForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("GetEventsListForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// if reqDto.CompetitionId == "" {
	// 	// 4.2. CompetitionId missing
	// 	log.Println("GetEventsListForSAP: CompetitionId is empty")
	// 	respDto.ErrorDescription = "CompetitionId is missing!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }

	// 6. Get Database
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	events, err := database.GetEvents(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, true) // TODO: Added Get All Events for Function once it is created
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("GetEventsListForSAP: Get Events List failed with error - ", err.Error())
		respDto.ErrorDescription = "Get Events List Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	//TODO: Uncomment and complete the Functions
	// Convert Event.model to Event.dto
	for _, eventsDTO := range events {
		event := dto.Event{}
		event.EventId = eventsDTO.EventId
		event.EventName = eventsDTO.EventName
		event.Status = eventsDTO.Status
		event.Favourite = eventsDTO.Favourite
		event.ProviderId = eventsDTO.ProviderId
		event.SportsId = eventsDTO.SportId
		event.SportName = eventsDTO.SportName
		event.CompetitionId = eventsDTO.CompetitionId
		event.CompetitionName = eventsDTO.CompetitionName
		event.OpenDate = eventsDTO.OpenDate
		respDto.Events = append(respDto.Events, event)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GetEventSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.GetEventRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetEventSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetEventSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.GetEventReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetEventSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 5.1. ProviderId missing
		log.Println("GetEventSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 5.2. SportId missing
		log.Println("GetEventSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.EventId == "" {
		// 5.4. EventId missing
		log.Println("GetEventSAP: EventId is empty")
		respDto.ErrorDescription = "EventId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	eventsDTO, err := database.GetEventDetails(reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("GetEventSAP: Get Event Details failed with error - ", err.Error())
		respDto.ErrorDescription = "Get Event Details Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	event := dto.Event{}
	event.EventId = eventsDTO.EventId
	event.EventName = eventsDTO.EventName
	event.Status = eventsDTO.Status
	event.Favourite = eventsDTO.Favourite
	event.ProviderId = eventsDTO.ProviderId
	event.SportsId = eventsDTO.SportId
	event.SportName = eventsDTO.SportName
	event.OpenDate = eventsDTO.OpenDate
	respDto.Event = event
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Events for SAP - only SAP
// @Summary      Block Events
// @Description  Block Events of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization     header    string                        true  "Bearer Token"
// @Param        BlockEventForSAP  body      portaldto.BlockedEventReqDto  true  "BlockedEventReqDto model is used"
// @Success      200               {object}  portaldto.BlockedEventResqDto
// @Failure      503               {object}  portaldto.BlockedEventResqDto
// @Router       /portal/sapadmin/block-event [post]
func BlockEventForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedEventResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockEventForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockEventForSAP") {
		log.Println("BlockEventForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedEventReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockEventForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockEventForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("BlockEventForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.CompetitionId == "" {
		// 4.2. CompetitionId missing
		log.Println("BlockEventForSAP: CompetitionId is empty")
		respDto.ErrorDescription = "CompetitionId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6. Get Database
	event := models.Event{}
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	event.EventKey = reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	event.Status = BLOCKED
	err = database.UpdateEventDetails(event)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockEventForSAP: Event Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Event Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err = database.UpdatePAEventStatus(event.EventKey, BLOCKED)
	// if err != nil {
	// 	// 6.1. Failed to update operator in database
	// 	log.Println("BlockEventForSAP: Event Blocking failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "Event Blocking Failed!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "EVENT " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.EventId, constants.EVENT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Events for SAP - only SAP
// @Summary      Unblock Events
// @Description  Unblock Events of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                          true  "Bearer Token"
// @Param        UnblockEventForSAP  body      portaldto.UnblockedEventReqDto  true  "UnblockedEventReqDto model is used"
// @Success      200                 {object}  portaldto.UnblockedEventResqDto
// @Failure      503                 {object}  portaldto.UnblockedEventResqDto
// @Router       /portal/sapadmin/unblock-event [post]
func UnblockEventForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UnblockedEventResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockEventForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockEventForSAP") {
		log.Println("UnblockEventForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.UnblockedEventReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockEventForSAP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockEventForSAP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("UnblockEventForSAP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.CompetitionId == "" {
		// 4.2. CompetitionId missing
		log.Println("UnblockEventForSAP: CompetitionId is empty")
		respDto.ErrorDescription = "CompetitionId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6. Get Database
	event := models.Event{}
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	event.EventKey = reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	event.Status = ACTIVE
	err = database.UpdateEventDetails(event)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockEventForSAP: Event Blocking failed with error - ", err.Error())
		respDto.ErrorDescription = "Event Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err = database.UpdatePAEventStatus(event.EventKey, ACTIVE)
	// if err != nil {
	// 	// 6.1. Failed to update operator in database
	// 	log.Println("UnblockEventForSAP: Event Blocking failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "Event Blocking Failed!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "EVENT " + ACTIVE
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.EventId, constants.EVENT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Events for OP - only OP
// @Summary      Block Events
// @Description  Block Events of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization    header    string                        true  "Bearer Token"
// @Param        BlockEventForOP  body      portaldto.BlockedEventReqDto  true  "BlockedEventReqDto model is used"
// @Success      200              {object}  portaldto.BlockedEventReqDto
// @Failure      503              {object}  portaldto.BlockedEventReqDto
// @Router       /portal/opadmin/block-event [post]
func BlockEventForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.BlockedEventResqDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockEventForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "blockEventForOP") {
		log.Println("BlockEventForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.BlockedEventReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("BlockEventForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("BlockEventForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("BlockEventForOP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.CompetitionId == "" {
		// 4.2. CompetitionId missing
		log.Println("BlockEventForOP: CompetitionId is empty")
		respDto.ErrorDescription = "CompetitionId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	// 6. Get Database
	eventKey := Tknmeta.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	eventStats, err := database.GetEventStatus(eventKey)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockEventForOP: database.GetEventStatus failed with error for eventKey - ", err.Error(), eventKey)
		err = cache.AddEventStatusOpStatus(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, BLOCKED)
		if err != nil {
			log.Println("BlockEventForOP: cache.AddEventStatusOpStatus failed with error for eventKey - ", err.Error(), eventKey)
			respDto.ErrorDescription = "Event Blocking Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Status = OK_STATUS
		respDto.ErrorDescription = "EVENT " + BLOCKED
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	eventStats.OperatorStatus = BLOCKED
	// err = database.UpdateOAEventStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId+"-"+reqDto.EventId, BLOCKED)
	err = database.ReplaceEventStatus(eventStats)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("BlockEventForOP: database.ReplaceEventStatus failed with error for eventKey - ", err.Error(), eventKey)
		respDto.ErrorDescription = "Event Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "EVENT " + BLOCKED
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.EventId, constants.EVENT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Events for OP - only OP
// @Summary      Block Events
// @Description  Block Events of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization    header    string                     true  "Bearer Token"
// @Param        BlockEventForOP  body      portaldto.GetEventsReqDto  true  "GetEventsReqDto model is used"
// @Success      200              {object}  portaldto.GetEventsRespDto
// @Failure      503              {object}  portaldto.GetEventsRespDto
// @Router       /portal/opadmin/unblock-event [post]
func UnblockEventForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.GetEventsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockEventForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "unblockEventForOP") {
		log.Println("UnblockEventForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.GetEventsReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("UnblockEventForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("UnblockEventForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("UnblockEventForOP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.CompetitionId == "" {
		// 4.2. CompetitionId missing
		log.Println("UnblockEventForOP: CompetitionId is empty")
		respDto.ErrorDescription = "CompetitionId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	// 6. Get Database
	eventKey := Tknmeta.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	eventStats, err := database.GetEventStatus(eventKey)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockEventForOP: database.GetEventStatus failed with error for eventKey - ", err.Error(), eventKey)
		err = cache.AddEventStatusOpStatus(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, ACTIVE)
		if err != nil {
			log.Println("UnblockEventForOP: cache.AddEventStatusOpStatus failed with error for eventKey - ", err.Error(), eventKey)
			respDto.ErrorDescription = "Event UnBlocking Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Status = OK_STATUS
		respDto.ErrorDescription = "EVENT " + ACTIVE
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	eventStats.OperatorStatus = ACTIVE
	// err = database.UpdateOAEventStatus(Tknmeta.OperatorId+"-"+reqDto.ProviderId+"-"+reqDto.SportId+"-"+reqDto.EventId, ACTIVE)
	err = database.ReplaceEventStatus(eventStats)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("UnblockEventForOP: database.ReplaceEventStatus failed with error for eventKey - ", err.Error(), eventKey)
		respDto.ErrorDescription = "Event UnBlocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "EVENT " + ACTIVE
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.EventId, constants.EVENT)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Competitions for OP - only OP
// @Summary      Get Competitions
// @Description  List All Competitions of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                     true  "Bearer Token"
// @Param        GetEventsListForOP  body      portaldto.GetEventsReqDto  true  "GetEventsReqDto model is used"
// @Success      200                 {object}  portaldto.GetEventsRespDto
// @Failure      503                 {object}  portaldto.GetEventsRespDto
// @Router       /portal/opadmin/events [post]
func GetEventsListForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.GetEventsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetEventsListForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getEventsListForOP") {
		log.Println("GetEventsListForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqDto := portaldto.GetEventsReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		// 4.1. Parsing failed
		log.Println("GetEventsListForOP: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check request for not null element
	if reqDto.ProviderId == "" {
		// 4.2. OperatorId missing
		log.Println("GetEventsListForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if reqDto.ProviderId == "" {
		// 4.2. CompetitionId missing
		log.Println("GetEventsListForOP: ProviderId is empty")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		// 4.2. SportId missing
		log.Println("GetEventsListForOP: SportId is empty")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// if reqDto.CompetitionId == "" {
	// 	// 4.2. CompetitionId missing
	// 	log.Println("GetEventsListForOP: CompetitionId is empty")
	// 	respDto.ErrorDescription = "CompetitionId is missing!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }

	// 5.1. Get SA All Events
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	events, err := database.GetEvents(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, true)
	if err != nil {
		// 5.1. Failed to Get competitions in database
		log.Println("GetEventsListForOP: GetEvents failed with error - ", err.Error())
		respDto.ErrorDescription = "Get Events list for OP Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6.1. List SA blocked competitions
	// blockedEvents := []models.Event{}
	// for _, event := range events {
	// 	if event.Status != ACTIVE {
	// 		blockedEvents = append(blockedEvents, event)
	// 	}
	// }
	// 6.2. Get OA All competitions
	opPrEvents, err := database.GetOpPrEvents(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
	if err != nil {
		log.Println("GetEventsListForOP: GetEvents failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	for _, event := range events {
		if event.Status != ACTIVE {
			// Platform Level Blocked, dont show in the list
			continue
		}
		isBlocked := false
		for _, opevent := range opPrEvents {
			if opevent.EventId == event.EventId {
				if opevent.ProviderStatus == BLOCKED {
					// Platfor Admin Blocked for this operator, dont show in the list
					isBlocked = true
				}
				if opevent.OperatorStatus == BLOCKED {
					// Operator Admin blocked, show as disabled
					event.Status = BLOCKED
				}
				break
			}
		}
		if isBlocked == true {
			// Platfor Admin Blocked for this operator, dont show in the list
			continue
		}
		respDto.Events = append(respDto.Events, GetEventsStatus2(event))
	}

	// 6.3. Filter SA Blocked competitions from OA competitions
	// for _, opPrEvent := range opPrEvents {
	// 	// filter, if provider status is not active
	// 	if opPrEvent.ProviderStatus != ACTIVE {
	// 		continue
	// 	}
	// 	isBlocked := false
	// 	// iterate through all sa blocked competitions
	// 	for _, saEvent := range blockedEvents {
	// 		// is sa blocked competitions = oa competitions, then mark isBlocked true and break the loop
	// 		if opPrEvent.EventId == saEvent.EventId {
	// 			isBlocked = true
	// 			break
	// 		}
	// 	}
	// 	if isBlocked {
	// 		continue
	// 	}
	// 	respDto.Events = append(respDto.Events, GetEventsStatus(opPrEvent))
	// }
	log.Println("GetEventsListForOP:", respDto.Events)
	sort.SliceStable(respDto.Events, func(i, j int) bool {
		return respDto.Events[i].OpenDate > respDto.Events[j].OpenDate
	})
	log.Println("GetEventsListForOP:", respDto.Events)
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Bets by OperatorId
// @Summary      Get Bets
// @Description  List All Bets of an operator. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                         true  "Bearer Token"
// @Param        GetBets        body      operatordto.BetsHistoryReqDto  true  "BetsHistoryReqDto model is used"
// @Success      200            {object}  operatordto.BetsHistoryRespDto
// @Failure      503            {object}  operatordto.BetsHistoryRespDto
// @Router       /portal/opadmin/get-bets [post]
func GetBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetBets: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getbets") {
		log.Println("GetBets: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetBets: Request Body is - ", reqStr)
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetBets: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := operatorsvc.CheckGetBets(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetBets: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check Operator Status
	if operatorDto.Status != ACTIVE {
		log.Println("GetBets: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Bets By Req
	bets, err := database.GetBetsByOperator(reqDto)
	if err != nil {
		log.Println("GetBets: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	log.Println("GetBets: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, operatorsvc.GetBetHistory(bets[itr]))
		betCount++
	}
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 7. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Bets by SAP
// @Summary      Get Bets for SAP
// @Description  List All Bets of SAPAdmin. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                         true  "Bearer Token"
// @Param        GetBetsForSAP  body      operatordto.BetsHistoryReqDto  true  "BetsHistoryReqDto model is used"
// @Success      200            {object}  operatordto.BetsHistoryRespDto
// @Failure      503            {object}  operatordto.BetsHistoryRespDto
// @Router       /portal/sapadmin/get-bets [post]
func GetBetsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetBets: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getBetsForSAP") {
		log.Println("GetBets: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetBets: Request Body is - ", reqStr)
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetBets: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := operatorsvc.CheckGetAllBets(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Bets By Req
	bets, err := database.GetAllBets(reqDto)
	if err != nil {
		log.Println("GetBets: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	log.Println("GetBets: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, operatorsvc.GetBetHistory(bets[itr]))
		betCount++
	}
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 5. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Bets by OP
// @Summary      Get Bets for OP
// @Description  List All Bets of OPAdmin. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                         true  "Bearer Token"
// @Param        GetBetsForOP   body      operatordto.BetsHistoryReqDto  true  "BetsHistoryReqDto model is used"
// @Success      200            {object}  operatordto.BetsHistoryRespDto
// @Failure      503            {object}  operatordto.BetsHistoryRespDto
// @Router       /portal/opadmin/get-bets [post]
func GetBetsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetBets: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getBetsForOP") {
		log.Println("GetBets: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetBets: Request Body is - ", reqStr)
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetBets: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := operatorsvc.CheckGetAllBets(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Bets By Req
	bets, err := database.GetAllOperatorBets(reqDto)
	if err != nil {
		log.Println("GetBets: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	log.Println("GetBets: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, operatorsvc.GetBetHistory(bets[itr]))
		betCount++
	}
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 5. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Settled Bets by SAP
// @Summary      Get Settled Bets for SAP
// @Description  List Settled Bets of SAPAdmin. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization         header    string                         true  "Bearer Token"
// @Param        GetSettledBetsForSAP  body      operatordto.BetsHistoryReqDto  true  "BetsHistoryReqDto model is used"
// @Success      200                   {object}  operatordto.BetsHistoryRespDto
// @Failure      503                   {object}  operatordto.BetsHistoryRespDto
// @Router       /portal/sapadmin/get-settled-bets [post]
func GetSettledBetsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetSettledBetsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetSettledBetsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetSettledBetsForSAP: Request Body is - ", reqStr)
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetSettledBetsForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := operatorsvc.CheckGetAllBets(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Bets By Req
	bets, err := database.GetAllSettledBets()
	if err != nil {
		log.Println("GetSettledBetsForSAP: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	log.Println("GetSettledBetsForSAP: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, operatorsvc.GetBetHistory(bets[itr]))
		betCount++
	}
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 5. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Lapsed Bets by SAP
// @Summary      Get Lapsed Bets for SAP
// @Description  List Lapsed Bets of SAPAdmin. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization        header    string                         true  "Bearer Token"
// @Param        GetLapsedBetsForSAP  body      operatordto.BetsHistoryReqDto  true  "BetsHistoryReqDto model is used"
// @Success      200                  {object}  operatordto.BetsHistoryRespDto
// @Failure      503                  {object}  operatordto.BetsHistoryRespDto
// @Router       /portal/sapadmin/get-lapsed-bets [post]
func GetLapsedBetsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetLapsedBetsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetLapsedBetsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetLapsedBetsForSAP: Request Body is - ", reqStr)
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetLapsedBetsForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := operatorsvc.CheckGetAllBets(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Bets By Req
	bets, err := database.GetAllLapsedBets()
	if err != nil {
		log.Println("GetLapsedBetsForSAP: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	log.Println("GetLapsedBetsForSAP: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, operatorsvc.GetBetHistory(bets[itr]))
		betCount++
	}
	log.Println("GetLapsedBetsForSAP: Bet Count: ", betCount)
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 5. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Lapsed Bets by SAP
// @Summary      Get Lapsed Bets for SAP
// @Description  List Lapsed Bets of SAPAdmin. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization           header    string                         true  "Bearer Token"
// @Param        GetCancelledBetsForSAP  body      operatordto.BetsHistoryReqDto  true  "BetsHistoryReqDto model is used"
// @Success      200                     {object}  operatordto.BetsHistoryRespDto
// @Failure      503                     {object}  operatordto.BetsHistoryRespDto
// @Router       /portal/sapadmin/get-cancelled-bets [post]
func GetCancelledBetsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetCancelledBetsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetCancelledBetsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetCancelledBetsForSAP: Request Body is - ", reqStr)
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetCancelledBetsForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := operatorsvc.CheckGetAllBets(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Bets By Req
	bets, err := database.GetAllCancelledBets()
	if err != nil {
		log.Println("GetCancelledBetsForSAP: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	log.Println("GetCancelledBetsForSAP: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, operatorsvc.GetBetHistory(bets[itr]))
		betCount++
	}
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 5. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetUsersForSAP is a function to get all users of SAP
// @Summary      List All Users
// @Description  List All Users of an SAPAdmin. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization   header    string                    true  "Bearer Token"
// @Param        GetUsersForSAP  body      portaldto.GetUsersReqDto  true  "GetUsersReqDto model is used"
// @Success      200             {object}  portaldto.ListUsersRespDto
// @Failure      503             {object}  portaldto.ListUsersRespDto
// @Router       /portal/sapadmin/list-users [post]
func GetUsersForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.ListUsersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Authenticate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		//Check for error and send 400 and error.
		log.Println("ListUsers: Token Authentication Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 3. Check Token User Role
	if !IsApplicable(Tknmeta, "getUsersForSAP") {
		log.Println("ListUsers: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("ListUsers: Request Body is - ", reqStr)
	reqDto := portaldto.GetUsersReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ListUsers: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Get Users from DB
	users, err := database.GetAllB2BUsers()
	if err != nil {
		log.Println("ListUsers: Get All users Failed for given Operator - ", err.Error())
		respDto.ErrorDescription = "Unable To get All Users"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6. Prepare response
	totalRecords := len(users)
	log.Println("ListUsers: Users Count: ", totalRecords)
	for itr := 0; itr < len(users); itr++ {
		respDto.Users = append(respDto.Users, GetUserDto(users[itr]))
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Bet by TxId
// @Summary      Get Bet
// @Description  Get Bet Details.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                  true  "Bearer Token"
// @Param        GetBet         body      portaldto.GetBetReqDto  true  "GetBetReqDto model is used"
// @Success      200            {object}  operatordto.GetBetRespDto
// @Failure      503            {object}  operatordto.GetBetRespDto
// @Router       /portal/get-bet [post]
func GetBet(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.GetBetRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.BetDetails = operatordto.BetHistory{}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetBet: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "getbet") {
		log.Println("GetBet: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("GetBet: Request Body is - ", reqStr)
	reqDto := portaldto.GetBetReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetBet: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 5. Request Check
	err := CheckGetBet(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetBet: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Check Operator Status
	if operatorDto.Status != ACTIVE {
		log.Println("GetBet: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Get Bets By Req
	bet, err := database.GetBetDetails(reqDto.TxId)
	if err != nil {
		log.Println("GetBet: Failed to get Bet Details: ", err.Error())
		respDto.ErrorDescription = "Bet not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.BetDetails = operatorsvc.GetBetHistory(bet)
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 9. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// ListUsers is a function to get all users of an operator
// @Summary      List All Users
// @Description  List All Users of an operator. Pagination is present. A maximum of 50 Users in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                    true  "Bearer Token"
// @Param        ListUsers      body      portaldto.GetUsersReqDto  true  "GetUsersReqDto model is used"
// @Success      200            {object}  portaldto.ListUsersRespDto
// @Failure      503            {object}  portaldto.ListUsersRespDto
// @Router       /portal/list-users [post]
func ListUsers(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.ListUsersRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Authenticate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		//Check for error and send 400 and error.
		log.Println("ListUsers: Token Authentication Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 3. Check Token User Role
	if !IsApplicable(Tknmeta, "listusers") {
		log.Println("ListUsers: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("ListUsers: Request Body is - ", reqStr)
	reqDto := portaldto.GetUsersReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ListUsers: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 5. Get Users from DB
	users, err := database.GetB2BUsers(Tknmeta.OperatorId, reqDto.PartialUserName)
	if err != nil {
		log.Println("ListUsers: Get All users Failed for given Operator - ", err.Error())
		respDto.ErrorDescription = "Unable To get All Users"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6. Prepare response
	totalRecords := len(users)
	log.Println("ListUsers: Users Count: ", totalRecords)
	for itr := 0; itr < len(users); itr++ {
		respDto.Users = append(respDto.Users, GetUserDto(users[itr]))
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// ListSports by Operator by Provider
// @Summary      List All Sports
// @Description  List All Sports by provider of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                     true  "Bearer Token"
// @Param        ListSports     body      commondto.GetSportsReqDto  true  "GetSportsReqDto model is used"
// @Success      200            {object}  commondto.GetSportsRespDto
// @Failure      503            {object}  commondto.GetSportsRespDto
// @Router       /portal/list-sports [post]
func ListSports(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := responsedto.SportsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Sports = []responsedto.SportDto{}

	// 2. Authenticate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		//Check for error and send 400 and error.
		log.Println("ListSports: Token Authentication Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 3. Check Token User Role
	if !IsApplicable(Tknmeta, "listsports") {
		log.Println("ListSports: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := commondto.GetSportsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ListSports: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Sports list from cache
	sportDtos, err := function.GetOpSports(reqDto.OperatorId, "", reqDto.ProviderId)
	if err != nil {
		// 8.1. Return Error
		log.Println("ListSports: GetOpSports failed with error - ", err.Error())
		respDto.ErrorDescription = "Failed to get Sports list."
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if len(sportDtos) > 0 {
		respDto.Sports = append(respDto.Sports, sportDtos...)
	}
	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// ListEvents by Provider by Sport
// @Summary      List All Events
// @Description  List All Events by Sport by provider of an operator.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                     true  "Bearer Token"
// @Param        ListSports     body      portaldto.GetEventsReqDto  true  "GetEventsReqDto model is used"
// @Success      200            {object}  portaldto.GetEventsRespDto
// @Failure      503            {object}  portaldto.GetEventsRespDto
// @Router       /portal/list-events [post]
func ListEvents(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := portaldto.GetEventsRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Events = []portaldto.Event{}

	// 2. Authenticate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		//Check for error and send 400 and error.
		log.Println("ListEvents: Token Authentication Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 3. Check Token User Role
	if !IsApplicable(Tknmeta, "listevents") {
		log.Println("ListEvents: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Parse request body to Request Object
	reqDto := portaldto.GetEventsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ListEvents: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	//check providerId -- Betfair
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	//events := []portaldto.Event{}
	// 5. Get Events from Database
	events, err := database.GetEvents(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, true)
	if err != nil {
		log.Println("ListEvents: Failed to get events from db")
		respDto.ErrorDescription = "Events not found!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	for _, event := range events {
		respDto.Events = append(respDto.Events, GetEvent(event))
	}
	respDto.ErrorDescription = ""
	respDto.Status = OK_STATUS

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get User Statement
// @Summary      Get User Statement
// @Description  List Transactions of a user. Pagination is present. A maximum of 50 Transactions in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                         true  "Bearer Token"
// @Param        UserStatement  body      portaldto.UserStatementReqDto  true  "UserStatementReqDto model is used"
// @Success      200            {object}  portaldto.UserStatementRespDto
// @Failure      503            {object}  portaldto.UserStatementRespDto
// @Router       /portal/user-statement [post]
func UserStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UserStatementRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Transactions = []portaldto.UserTransaction{}
	respDto.Balance = 0
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UserStatement: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, "userstatement") {
		log.Println("UserStatement: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("UserStatement: Request Body is - ", reqStr)
	reqDto := portaldto.UserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UserStatement: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 5. Request Check
	err := CheckUserStatement(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("UserStatement: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Check Operator Status
	if operatorDto.Status != ACTIVE {
		log.Println("UserStatement: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Check if Operator wallet type is seamless or transfer?
	if strings.ToLower(operatorDto.WalletType) == "seamless" {
		// 8.1. Seamless wallet, do balance call to operator
		log.Println("UserStatement: Operator wallet type is not transfer: ", operatorDto.WalletType)
		respDto.ErrorDescription = "This operation is not supported for seamless wallet operator!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8.2. Transfer wallet, get user balance from db
	userKey := reqDto.OperatorId + "-" + reqDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println("UserStatement: User NOT FOUND: ", err.Error())
		respDto.ErrorDescription = "Invalid User!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Balance = b2bUser.Balance
	// 9. Get User Transactions from DB
	records, err := database.GetStatement(reqDto)
	if err != nil {
		log.Println("UserStatement: Database operation failed with error: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 10. Construct response
	totalRecords := len(records)
	log.Println("UserStatement: Records Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}
	count := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Transactions = append(respDto.Transactions, GetTransaction(records[itr]))
		count++
	}
	respDto.PageSize = count
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 10. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get All User Statement for SAP
// @Summary      Get All User Statement for SAP
// @Description  List Transactions of All user for SAP. Pagination is present. A maximum of 50 Transactions in a single request.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization        header    string                         true  "Bearer Token"
// @Param        UserStatementForSAP  body      portaldto.UserStatementReqDto  true  "UserStatementReqDto model is used"
// @Success      200                  {object}  portaldto.UserStatementRespDto
// @Failure      503                  {object}  portaldto.UserStatementRespDto
// @Router       /portal/sapadmin/user-statement [post]
func UserStatementForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UserStatementRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	respDto.Transactions = []portaldto.UserTransaction{}
	respDto.Balance = 0
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UserStatementForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UserStatementForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("UserStatementForSAP: Request Body is - ", reqStr)
	reqDto := portaldto.UserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UserStatementForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 5. Request Check
	err := CheckAllUserStatement(reqDto)
	if err != nil {
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get User Transactions from DB
	records, err := database.GetAllStatement(reqDto)
	if err != nil {
		log.Println("UserStatementForSAP: Database operation failed with error: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Construct response
	totalRecords := len(records)
	log.Println("UserStatementForSAP: Records Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page > 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize > 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}
	count := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Transactions = append(respDto.Transactions, GetTransaction(records[itr]))
		count++
	}
	respDto.PageSize = count
	respDto.TotalRecords = totalRecords
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	// 8. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Operator Details form the status tables for SAP
// @Summary      Get Operator Details form the status tables for SAP
// @Description  List Operator of specific Tab for SAP. Pagination is present.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization          header    string                         true  "Bearer Token"
// @Param        OperatorDetailsForSAP  body      commondto.GetOperDetailReqDto  true  "GetOperDetailReqDto model is used"
// @Success      200                    {object}  commondto.GetOperDetailRespDto
// @Failure      503                    {object}  commondto.GetOperDetailRespDto
// @Router       /portal/sapadmin/operator-details [post]
func OperatorDetailsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.GetOperDetailRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(commondto.GetOperDetailReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("OperatorDetailsForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorDetailsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorDetailsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if reqDto.SportId != "" && reqDto.ProviderId != "" && reqDto.EventId == "" && reqDto.CompetitionId == "" {
		// for Sport sports
		sports, err := database.GetOperatorsFromSpPr(reqDto.SportId, reqDto.ProviderId)
		if err != nil {
			log.Println("OperatorDetailsForSAP: Failed operators from sports status with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		for _, sport := range sports {
			status := commondto.OperatorStatus{}
			status.IsActive = sport.ProviderStatus
			status.OperatorId = sport.OperatorId
			status.OperatorName = sport.OperatorName
			status.PartnerId = sport.PartnerId
			respDto.OperatorStatus = append(respDto.OperatorStatus, status)
		}
	} else if reqDto.CompetitionId != "" && reqDto.SportId != "" && reqDto.ProviderId != "" && reqDto.EventId == "" {
		competitionKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
		operators, err := database.GetAllOperators()
		if err != nil {
			// 5.1. Failed to Get competitions in database
			log.Println("OperatorDetailsForSAP: database.GetAllOperators failed with error for competitionKey - ", err.Error(), competitionKey)
			respDto.ErrorDescription = "Get Competitions list Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		competitions, err := database.GetOperatorsFromCompPr(reqDto.CompetitionId, reqDto.ProviderId)
		if err != nil {
			log.Println("OperatorDetailsForSAP: Failed operators from Competition status with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		for _, operator := range operators {
			status := commondto.OperatorStatus{}
			status.OperatorId = operator.OperatorId
			status.OperatorName = operator.OperatorName
			status.IsActive = ACTIVE
			for _, es := range competitions {
				if operator.OperatorId == es.OperatorId {
					status.IsActive = es.ProviderStatus
					break
				}
			}
			respDto.OperatorStatus = append(respDto.OperatorStatus, status)
		}
		// for _, competition := range competitions {
		// 	status := commondto.OperatorStatus{}
		// 	status.IsActive = competition.ProviderStatus
		// 	status.OperatorId = competition.OperatorId
		// 	status.OperatorName = competition.OperatorName
		// 	respDto.OperatorStatus = append(respDto.OperatorStatus, status)
		// }
	} else if reqDto.EventId != "" && reqDto.SportId != "" && reqDto.ProviderId != "" && reqDto.CompetitionId == "" {
		//if reqDto.ProviderId == "BetFair" {
		//	reqDto.ProviderId = "Betfair"
		//}
		eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
		operators, err := database.GetAllOperators()
		if err != nil {
			// 5.1. Failed to Get competitions in database
			log.Println("OperatorDetailsForSAP: database.GetAllOperators failed with error for eventKey - ", err.Error(), eventKey)
			respDto.ErrorDescription = "Get Events list Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		events, err := database.GetOperatorsFromEvPr(reqDto.EventId, reqDto.ProviderId)
		if err != nil {
			log.Println("OperatorDetailsForSAP: Failed operators from Competition status with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		for _, operator := range operators {
			status := commondto.OperatorStatus{}
			status.OperatorId = operator.OperatorId
			status.OperatorName = operator.OperatorName
			status.IsActive = ACTIVE
			for _, es := range events {
				if operator.OperatorId == es.OperatorId {
					status.IsActive = es.ProviderStatus
					break
				}
			}
			respDto.OperatorStatus = append(respDto.OperatorStatus, status)
		}
		// for _, event := range events {
		// 	status := commondto.OperatorStatus{}
		// 	status.IsActive = event.ProviderStatus
		// 	status.OperatorId = event.OperatorId
		// 	status.OperatorName = event.OperatorName
		// 	respDto.OperatorStatus = append(respDto.OperatorStatus, status)
		// }
	} else if reqDto.ProviderId != "" && reqDto.CompetitionId == "" && reqDto.EventId == "" && reqDto.SportId == "" {
		partners, err := database.GetPartnersByProviderId(reqDto.ProviderId)
		if err != nil {
			log.Println("OperatorDetailsForSAP: GetPartnersByProviderId failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		for _, partner := range partners {
			status := commondto.OperatorStatus{}
			status.IsActive = partner.ProviderStatus
			status.OperatorId = partner.OperatorId
			status.OperatorName = partner.OperatorName
			status.PartnerId = partner.PartnerId
			respDto.OperatorStatus = append(respDto.OperatorStatus, status)
		}
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "Success"
	return c.JSON(respDto)
}

// Get Operator Status  Block form the status tables for SAP
// @Summary      Get Operator Status  Block form the status tables for SAP
// @Description  Get Operator Status  Block form the status tables for SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization                   header    string                                true  "Bearer Token"
// @Param OperatorStatusBlockForSAP body commondto.OperStatusBlockReqDto true "OperStatusBlockReqDto model is used"
// @Success 200 {object} commondto.OperStatusBlockRespDto
// @Failure 503 {object} commondto.OperStatusBlockRespDto
// @Router /portal/sapadmin/operator-status-block [post]
// func OperatorStatusBlockForSAP(c *fiber.Ctx) error {
// 	c.Accepts("json", "text")
// 	c.Accepts("application/json")

// 	// 1. Create Default Response Object
// 	respDto := commondto.OperStatusBlockRespDto{}
// 	respDto.ErrorDescription = GENERAL_ERROR_DESC
// 	respDto.Status = ERROR_STATUS

// 	reqDto := new(commondto.OperStatusBlockReqDto)
// 	if err := c.BodyParser(reqDto); err != nil {
// 		log.Println("OperatorStatusBlockForSAP: Body Parsing failed")
// 		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}

// 	// 2. Validation Token
// 	Tknmeta, ok := Authenticate(c)
// 	if !ok {
// 		// 2.1. Token validaton failed.
// 		log.Println("OperatorStatusBlockForSAP: Token Validation Failed")
// 		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	// 3. Check Role Permissions
// 	if !IsApplicable(Tknmeta, c.OriginalURL()) {
// 		log.Println("OperatorStatusBlockForSAP: User not Permitted to access the API")
// 		respDto.ErrorDescription = UNAUTH_ACCESS
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}

// 	if reqDto.ProviderId != "" && reqDto.OperatorId != "" && reqDto.SportId != "" && reqDto.CompetitionId == "" && reqDto.EventId == "" {
// 		sportKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId
// 		err := database.UpdateOASportStatus(sportKey, BLOCKED)
// 		if err != nil {
// 			log.Println("OperatorStatusBlockForSAP: Failed update operators from sport status with error - ", err.Error())
// 			respDto.ErrorDescription = err.Error()
// 			return c.Status(fiber.StatusOK).JSON(respDto)
// 		}
// 		respDto.ErrorDescription = "OPERATOR " + BLOCKED
// 	} else if reqDto.ProviderId != "" && reqDto.OperatorId != "" && reqDto.SportId != "" && reqDto.CompetitionId != "" && reqDto.EventId == "" {
// 		competitionKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
// 		err := database.UpdateOACompetitionStatus(competitionKey, BLOCKED)
// 		if err != nil {
// 			log.Println("OperatorStatusBlockForSAP: Failed update operators from completition status with error - ", err.Error())
// 			respDto.ErrorDescription = err.Error()
// 			return c.Status(fiber.StatusOK).JSON(respDto)
// 		}
// 		respDto.ErrorDescription = "OPERATOR " + BLOCKED
// 	} else if reqDto.ProviderId != "" && reqDto.OperatorId != "" && reqDto.SportId != "" && reqDto.CompetitionId == "" && reqDto.EventId != "" {
// 		eventKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
// 		err := database.UpdateOAEventStatus(eventKey, BLOCKED)
// 		if err != nil {
// 			log.Println("OperatorStatusBlockForSAP: Failed update operators from event status with error - ", err.Error())
// 			respDto.ErrorDescription = err.Error()
// 			return c.Status(fiber.StatusOK).JSON(respDto)
// 		}
// 		respDto.ErrorDescription = "OPERATOR " + BLOCKED
// 	} else if reqDto.ProviderId != "" && reqDto.OperatorId != "" && reqDto.SportId == "" && reqDto.CompetitionId == "" && reqDto.EventId == "" {
// 		err := database.UpdateOAProviderStatus(reqDto.OperatorId, reqDto.ProviderId, BLOCKED)
// 		if err != nil {
// 			log.Println("OperatorStatusBlockForSAP: Failed update operators status from providers status with error - ", err.Error())
// 			respDto.ErrorDescription = err.Error()
// 			return c.Status(fiber.StatusOK).JSON(respDto)
// 		}
// 		respDto.ErrorDescription = "OPERATOR " + BLOCKED
// 	}
// 	respDto.Status = OK_STATUS

// 	return c.JSON(respDto)
// }

// Update Status of Operator with respect to providers in SAP
// @Summary      Update Status of Operator with respect to providers in SAP
// @Description  Update Status of Operator with respect to providers in SAP
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization                   header    string                                true  "Bearer Token"
// @Param        OperatorStatusInProviderForSAP  body      commondto.OperStatusInProviderReqDto  true  "OperStatusInProviderReqDto model is used"
// @Success      200                             {object}  commondto.OperStatusRespDto
// @Failure      503                             {object}  commondto.OperStatusRespDto
// @Router       /portal/sapadmin/provider-operator-status [post]
func OperatorStatusInProviderForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.OperStatusRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(commondto.OperStatusInProviderReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("OperatorStatusInProviderForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorStatusInProviderForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorStatusInProviderForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	err := database.UpdatePartnerProviderStatus(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId, reqDto.Status)
	if err != nil {
		log.Println("OperatorStatusInProviderForSAP: Failed update operators status from providers status with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.ErrorDescription = "OPERATOR " + reqDto.Status
	respDto.Status = OK_STATUS

	return c.JSON(respDto)
}

// Update Status of Operator with respect to Sport in SAP
// @Summary      Update Status of Operator with respect to Sport in SAP
// @Description  Update Status of Operator with respect to Sport in SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization                      header    string                                   true  "Bearer Token"
// @Param        OperatorStatusInSportForSAP  body      commondto.OperStatusInSportReqDto  true  "OperStatusInSportReqDto model is used"
// @Success      200                          {object}  commondto.OperStatusRespDto
// @Failure      503                          {object}  commondto.OperStatusRespDto
// @Router       /portal/sapadmin/sport-operator-status [post]
func OperatorStatusInSportForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.OperStatusRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(commondto.OperStatusInSportReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("OperatorStatusInSportForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorStatusInSportForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorStatusInSportForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	sportKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId
	err := database.UpdatePASportStatus(sportKey, reqDto.Status)
	if err != nil {
		log.Println("OperatorStatusInSportForSAP: Failed update operators from sport status with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.ErrorDescription = "OPERATOR " + reqDto.Status
	respDto.Status = OK_STATUS

	return c.JSON(respDto)
}

// Update Status of Operator with respect to Competition in SAP
// @Summary      Update Status of Operator with respect to Competition in SAP
// @Description  Update Status of Operator with respect to Competition in SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization                header    string                             true  "Bearer Token"
// @Param        OperatorStatusInCompetitionForSAP  body      commondto.OperStatusInCompetitionReqDto  true  "OperStatusInCompetitionReqDto model is used"
// @Success      200                                {object}  commondto.OperStatusRespDto
// @Failure      503                                {object}  commondto.OperStatusRespDto
// @Router       /portal/sapadmin/competition-operator-status [post]
func OperatorStatusInCompetitionForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.OperStatusRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(commondto.OperStatusInCompetitionReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("OperatorStatusInCompetitionForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorStatusInCompetitionForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorStatusInCompetitionForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	competitionKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
	competitionStats, err := database.GetCompetitionStatus(competitionKey)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("OperatorStatusInCompetitionForSAP: database.GetCompetitionStatus failed with error for competitionKey - ", err.Error(), competitionKey)
		err = cache.AddCompetitionStatusPaStatus(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, reqDto.Status)
		if err != nil {
			log.Println("OperatorStatusInCompetitionForSAP: cache.AddCompetitionStatusOpStatus failed with error for competitionKey - ", err.Error(), competitionKey)
			respDto.ErrorDescription = "Competition Blocking Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Status = OK_STATUS
		respDto.ErrorDescription = "OPERATOR " + reqDto.Status
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	competitionStats.ProviderStatus = reqDto.Status
	err = database.ReplaceCompetitionStatus(competitionStats)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("OperatorStatusInCompetitionForSAP: database.ReplaceCompetitionStatus failed with error for competitionKey - ", err.Error(), competitionKey)
		respDto.ErrorDescription = "Competition Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err := database.UpdatePACompetitionStatus(competitionKey, reqDto.Status)
	// if err != nil {
	// 	log.Println("OperatorStatusInCompetitionForSAP: Failed update operators from completition status with error - ", err.Error())
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.ErrorDescription = "OPERATOR " + reqDto.Status
	respDto.Status = OK_STATUS

	return c.JSON(respDto)
}

// Update Status of Operator with respect to Event in SAP
// @Summary      Update Status of Operator with respect to Event in SAP
// @Description  Update Status of Operator with respect to Event in SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization                header    string                             true  "Bearer Token"
// @Param        OperatorStatusInEventForSAP  body      commondto.OperStatusInEventReqDto  true  "OperStatusInEventReqDto model is used"
// @Success      200                          {object}  commondto.OperStatusRespDto
// @Failure      503                          {object}  commondto.OperStatusRespDto
// @Router       /portal/sapadmin/event-operator-status [post]
func OperatorStatusInEventForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.OperStatusRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(commondto.OperStatusInEventReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("OperatorStatusInEventForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorStatusInEventForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorStatusInEventForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//if reqDto.ProviderId == "BetFair" {
	//	reqDto.ProviderId = "Betfair"
	//}
	eventKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	eventStats, err := database.GetEventStatus(eventKey)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("OperatorStatusInEventForSAP: database.GetEventStatus failed with error for eventKey - ", err.Error(), eventKey)
		err = cache.AddEventStatusPaStatus(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.Status)
		if err != nil {
			log.Println("OperatorStatusInEventForSAP: cache.AddEventStatusOpStatus failed with error for eventKey - ", err.Error(), eventKey)
			respDto.ErrorDescription = "Event Blocking Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.Status = OK_STATUS
		respDto.ErrorDescription = "OPERATOR " + reqDto.Status
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	eventStats.ProviderStatus = reqDto.Status
	err = database.ReplaceEventStatus(eventStats)
	if err != nil {
		// 6.1. Failed to update operator in database
		log.Println("OperatorStatusInEventForSAP: database.ReplaceEventStatus failed with error for eventKey - ", err.Error(), eventKey)
		respDto.ErrorDescription = "Event Blocking Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// err := database.UpdatePAEventStatus(eventKey, reqDto.Status)
	// if err != nil {
	// 	log.Println("OperatorStatusInEventForSAP: Failed update operators from event status with error - ", err.Error())
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	respDto.ErrorDescription = "OPERATOR " + reqDto.Status
	respDto.Status = OK_STATUS

	return c.JSON(respDto)
}

// Get Operator Status  Block form the status tables for SAP
// @Summary      Get Operator Status  Block form the status tables for SAP
// @Description  Get Operator Status  Block form the status tables for SAP.
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        Authorization                header    string                             true  "Bearer Token"
// @Param        OperatorStatusUnblockForSAP  body      commondto.OperStatusUnblockReqDto  true  "OperStatusUnblockReqDto model is used"
// @Success      200                          {object}  commondto.OperStatusUnblockRespDto
// @Failure      503                          {object}  commondto.OperStatusUnblockRespDto
// @Router       /portal/sapadmin/operator-status-unblock [post]
func OperatorStatusUnblockForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.OperStatusUnblockRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	reqDto := new(commondto.OperStatusUnblockReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("OperatorStatusBlockForSAP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorStatusBlockForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorStatusBlockForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if reqDto.SportId != "" && reqDto.ProviderId != "" && reqDto.OperatorId != "" {
		sportKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId
		err := database.UpdateOASportStatus(sportKey, ACTIVE)
		if err != nil {
			log.Println("OperatorStatusBlockForSAP: Failed update operators from sport status with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.ErrorDescription = "OPERATOR " + ACTIVE
	} else if reqDto.SportId != "" && reqDto.ProviderId != "" && reqDto.OperatorId != "" && reqDto.CompetitionId != "" {
		respDto.ErrorDescription = "NOT IMPLEMENTED!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
		// competitionKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.CompetitionId
		// err := database.UpdateOACompetitionStatus(competitionKey, ACTIVE)
		// if err != nil {
		// 	log.Println("OperatorStatusBlockForSAP: Failed update operators from completition status with error - ", err.Error())
		// 	respDto.ErrorDescription = err.Error()
		// 	return c.Status(fiber.StatusOK).JSON(respDto)
		// }
		// respDto.ErrorDescription = "OPERATOR " + ACTIVE
	} else if reqDto.EventId != "" && reqDto.SportId != "" && reqDto.ProviderId != "" && reqDto.OperatorId != "" && reqDto.CompetitionId != "" {
		respDto.ErrorDescription = "NOT IMPLEMENTED!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
		// eventKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
		// err := database.UpdateOAEventStatus(eventKey, ACTIVE)
		// if err != nil {
		// 	log.Println("OperatorStatusBlockForSAP: Failed update operators from event status with error - ", err.Error())
		// 	respDto.ErrorDescription = err.Error()
		// 	return c.Status(fiber.StatusOK).JSON(respDto)
		// }
		// respDto.ErrorDescription = "OPERATOR " + ACTIVE
	} else {
		err := database.UpdatePartnerProviderStatus(reqDto.OperatorId, reqDto.PartnerId, reqDto.ProviderId, ACTIVE)
		if err != nil {
			log.Println("OperatorStatusBlockForSAP: UpdatePartnerProviderStatus failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.ErrorDescription = "OPERATOR " + ACTIVE
	}
	respDto.Status = OK_STATUS

	return c.JSON(respDto)
}

func GetConfigForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.GetConfigRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetConfigForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetConfigForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	reqDto := new(commondto.GetConfigReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetConfigForOP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}
	var sapConfig commondto.ConfigDto
	var err error
	var lvl string
	switch reqDto.ConfigCall {
	case "partner":
		sapConfig, lvl, err = getProviderConfigSAP(reqDto.ProviderId)
	case "sport":
		sapConfig, lvl, err = getSportConfigSAP(reqDto.ProviderId, reqDto.SportId)
	case "competition":
		sapConfig, lvl, err = getCompetitionConfigSAP(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
	case "event":
		sapConfig, lvl, err = getEventConfigSAP(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, reqDto.EventId)
	default:
		log.Println("GetConfigForSAP: Wrong Config Call - ", reqDto.ConfigCall)
		respDto.ErrorDescription = "Wrong Config Call - " + reqDto.ConfigCall
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	if err != nil {
		log.Println("GetConfigForSAP: Failed to get config with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}
	respDto.Config = sapConfig
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Level = lvl
	return c.JSON(respDto)

}

func GetConfigForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.GetConfigRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetConfigForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetConfigForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	reqDto := new(commondto.GetConfigReqDto)
	log.Println("GetConfigForOP: Request Body is - ", string(c.Body()))
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetConfigForOP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}
	jsonReq, err := json.Marshal(reqDto)
	if err == nil {
		log.Println("GetConfigForOP: Request JSON is - ", string(jsonReq))
	}
	// 4. Check input values
	if reqDto.ConfigCall == "" {
		log.Println("GetConfigForOP: ConfigCall value is empty!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// if reqDto.PartnerId != "" && reqDto.ConfigCall != "Partner" {
	// 	reqDto.PartnerId = ""
	// }
	// 4. Get Operator Details
	operatorId := Tknmeta.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetConfigForOP: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Failed to get Operator Details"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" && reqDto.ConfigCall == "partner" {
		log.Println("GetConfigForOP: PartnerId is missing!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if len(operatorDto.Partners) == 0 {
		log.Println("GetConfigForOP: Operator doesnt have partners!!!")
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Partner Details
	var found bool = false
	partner := operatorDto.Partners[0] // Default Partner
	for _, partnerDto := range operatorDto.Partners {
		if partnerDto.PartnerId == partnerId {
			partner = partnerDto
			found = true
			break
		}
	}
	if found == false && reqDto.ConfigCall == "partner" {
		log.Println("GetConfigForOP: Invalid PartnerId - ", partnerId)
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	var opConfig commondto.ConfigDto
	var lvl string
	switch reqDto.ConfigCall {
	case "partner":
		opConfig, lvl, err = getPartnerConfigOP(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId)
	case "sport":
		opConfig, lvl, err = getSportConfigOP(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId)
	case "competition":
		opConfig, lvl, err = getCompetitionConfigOP(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId)
	case "event":
		opConfig, lvl, err = getEventConfigOP(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, reqDto.EventId)
	default:
		log.Println("GetConfigForOP: Wrong Config Call - ", reqDto.ConfigCall)
		respDto.ErrorDescription = "Wrong Config Call - " + reqDto.ConfigCall
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	if err != nil && lvl != "event" {
		log.Println("GetConfigForOP: Failed to get "+reqDto.ConfigCall+" config with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	// update values based on rate in partnerid
	newConfig := opConfig
	if partner.Rate == 0 {
		partner.Rate = 1
	}
	newConfig.MatchOdds.Min = opConfig.MatchOdds.Min / partner.Rate
	newConfig.MatchOdds.Max = opConfig.MatchOdds.Max / partner.Rate
	newConfig.Bookmaker.Min = opConfig.Bookmaker.Min / partner.Rate
	newConfig.Bookmaker.Max = opConfig.Bookmaker.Max / partner.Rate
	newConfig.Fancy.Min = opConfig.Fancy.Min / partner.Rate
	newConfig.Fancy.Max = opConfig.Fancy.Max / partner.Rate
	configJson, err := json.Marshal(newConfig)
	log.Println("GetConfigForOP: newConfig is - ", string(configJson))
	log.Println("GetConfigForOP: Level is - ", lvl)
	// send response
	respDto.Config = newConfig
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Level = lvl
	return c.JSON(respDto)
}

func SetConfigForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.SetConfigRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("SetConfigForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("SetConfigForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	reqDto := new(commondto.SetConfigReqDto)
	log.Println(string(c.Body()))
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("SetConfigForOP: Body Parsing failed")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Check if hold is not in between 0 and 25
	if reqDto.Config.Hold < 0 || reqDto.Config.Hold > 25 {
		log.Println("SetConfigForOP: Hold is not in between 0 and 25")
		respDto.ErrorDescription = "Hold is out of range"
		return c.JSON(respDto)
	}

	var err error
	var lvl string
	switch reqDto.ConfigCall {
	case "partner":
		err = setProviderConfigSAP(reqDto.ProviderId, reqDto.Config)
	case "sport":
		err = setSportConfigSAP(reqDto.ProviderId, reqDto.SportId, reqDto.Config)
	case "competition":
		err = setCompetitionConfigSAP(reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, reqDto.Config)
	case "event":
		err = setEventConfigSAP(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.Config)
	default:
		log.Println("SetConfigForOP: Wrong Config Call - ", reqDto.ConfigCall)
		respDto.ErrorDescription = "Wrong Config Call - " + reqDto.ConfigCall
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	if err != nil {
		log.Println("SetConfigForOP: Failed to set "+reqDto.ConfigCall+" config with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Level = lvl
	// Audit Portal
	if respDto.Status == OK_STATUS {
		switch reqDto.ConfigCall {
		case "partner":
			PortalAudit(*Tknmeta, c, reqDto.PartnerId, constants.PORTAL_USER)
		case "sport":
			PortalAudit(*Tknmeta, c, reqDto.SportId, constants.PORTAL_USER)
		case "competition":
			PortalAudit(*Tknmeta, c, reqDto.CompetitionId, constants.PORTAL_USER)
		case "event":
			PortalAudit(*Tknmeta, c, reqDto.EventId, constants.PORTAL_USER)
		default:
			log.Println("SetConfigForOP: Wrong Config Call - ", reqDto.ConfigCall)
		}
	}
	return c.JSON(respDto)
}

func SetConfigForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := commondto.SetConfigRespDto{}
	respDto.ErrorDescription = GENERAL_ERROR_DESC
	respDto.Status = ERROR_STATUS
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("SetConfigForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("SetConfigForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	reqDto := new(commondto.SetConfigReqDto)
	log.Println("SetConfigForOP: Request Body is - ", string(c.Body()))
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("SetConfigForOP: Body Parsing failed", err.Error())
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check input values
	if reqDto.ConfigCall == "" {
		log.Println("SetConfigForOP: ConfigCall value is empty!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// Check if hold is not in between 0 and 100
	if reqDto.Config.Hold < 0 || reqDto.Config.Hold > 100 {
		log.Println("SetConfigForOP: Hold is not in between 0 and 100")
		respDto.ErrorDescription = "Hold is out of range"
		return c.JSON(respDto)
	}
	// 4. Get Operator Details
	operatorId := Tknmeta.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("SetConfigForOP: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Failed to get Operator Details"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" && reqDto.ConfigCall == "partner" {
		log.Println("SetConfigForOP: PartnerId is missing!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if len(operatorDto.Partners) == 0 {
		log.Println("SetConfigForOP: Operator doesnt have partners!!!")
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Partner Details
	var found bool = false
	partner := operatorDto.Partners[0] // Default Partner
	for _, partnerDto := range operatorDto.Partners {
		if partnerDto.PartnerId == partnerId {
			partner = partnerDto
			found = true
			break
		}
	}
	if found == false && reqDto.ConfigCall == "partner" {
		log.Println("SetConfigForOP: Invalid PartnerId - ", partnerId)
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if partner.Rate == 0 {
		partner.Rate = 1
	}
	// update values based on rate in partnerid
	newConfig := reqDto.Config
	newConfig.MatchOdds.Min = reqDto.Config.MatchOdds.Min * partner.Rate
	newConfig.MatchOdds.Max = reqDto.Config.MatchOdds.Max * partner.Rate
	newConfig.Bookmaker.Min = reqDto.Config.Bookmaker.Min * partner.Rate
	newConfig.Bookmaker.Max = reqDto.Config.Bookmaker.Max * partner.Rate
	newConfig.Fancy.Min = reqDto.Config.Fancy.Min * partner.Rate
	newConfig.Fancy.Max = reqDto.Config.Fancy.Max * partner.Rate
	configJson, err := json.Marshal(newConfig)
	log.Println("SetConfigForOP: newConfig is - ", string(configJson))
	var lvl string = reqDto.ConfigCall
	switch reqDto.ConfigCall {
	case "partner":
		err = setPartnerStatus(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId, newConfig)
	case "sport":
		err = setSportConfigOP(Tknmeta.OperatorId, reqDto.PartnerId, reqDto.ProviderId, reqDto.SportId, newConfig)
	case "competition":
		err = setCompetitionConfigOP(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.CompetitionId, newConfig)
	case "event":
		err = setEventConfigOP(Tknmeta.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, newConfig)
	default:
		log.Println("SetConfigForOP: Wrong Config Call - ", reqDto.ConfigCall)
		respDto.ErrorDescription = "Wrong Config Call - " + reqDto.ConfigCall
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	if err != nil {
		log.Println("SetConfigForOP: Failed to set "+reqDto.ConfigCall+" config with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		respDto.Level = lvl
		return c.JSON(respDto)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Config = reqDto.Config
	respDto.Level = lvl

	// Audit Portal
	if respDto.Status == OK_STATUS {
		switch reqDto.ConfigCall {
		case "partner":
			PortalAudit(*Tknmeta, c, reqDto.PartnerId, constants.PORTAL_USER)
		case "sport":
			PortalAudit(*Tknmeta, c, reqDto.SportId, constants.PORTAL_USER)
		case "competition":
			PortalAudit(*Tknmeta, c, reqDto.CompetitionId, constants.PORTAL_USER)
		case "event":
			PortalAudit(*Tknmeta, c, reqDto.EventId, constants.PORTAL_USER)
		default:
			log.Println("SetConfigForOP: Wrong Config Call - ", reqDto.ConfigCall)
		}
	}
	return c.JSON(respDto)
}

// Operator's Balance Endpoints
// Transfer Wallet APIs
func OperatorBalance(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 0. Create Default Response Object
	respDto := dto.OperatorBalanceRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	// 1. Parse request body to Request Object
	reqDto := dto.OperatorBalanceReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("OperatorBalance: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorBalance: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorBalance: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := database.GetOperatorDetails(operatorId)
	//operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("OperatorBalance: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get User Balance from db
	respDto.Balance = operatorDto.Balance
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func OperatorDeposit(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 0. Create Default Response Object
	respDto := dto.OperatorBalanceRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	// 1. Parse request body to Request Object
	reqDto := dto.OperatorDepositReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("OperatorDeposit: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorDeposit: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorDeposit: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Request Check
	operatorId := reqDto.OperatorId
	if operatorId == "" {
		log.Println("OperatorDeposit: OperatorId is missing!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("OperatorDeposit: PartnerId is missing!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Operator Details from DB
	operatorDto, err := database.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("OperatorDeposit: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = INVALID_OPERATORID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if len(operatorDto.Partners) == 0 {
		log.Println("OperatorDeposit: Operator doesnt have partners!!!")
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Partner Details
	var found bool = false
	partner := operatordto.Partner{}
	for _, partnerDto := range operatorDto.Partners {
		if partnerDto.PartnerId == partnerId {
			partner = partnerDto
			found = true
			break
		}
	}
	if found == false {
		log.Println("OperatorDeposit: Invalid PartnerId - ", partnerId)
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	creditAmount := reqDto.CreditAmount * float64(partner.Rate)
	// 7. Save deposit transaction in User Ledger
	operatorLedger := models.OperatorLedgerDto{}
	operatorLedger.OperatorId = reqDto.OperatorId
	operatorLedger.OperatorName = operatorDto.OperatorName
	operatorLedger.TransactionType = "Deposit-Funds"
	operatorLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	operatorLedger.ReferenceId = ""
	operatorLedger.Amount = creditAmount
	err = database.InsertOperatorLedger(operatorLedger)
	if err != nil {
		// 7.1. inserting ledger document failed
		log.Println("OperatorDeposit: insert ledger failed with error - ", err.Error())
		respDto.ErrorDescription = "Deposit Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Update operator balance
	err = database.UpdateOperatorBalance(operatorId, creditAmount)
	if err != nil {
		// 8.1. update failed, send failure response
		log.Println("OperatorDeposit: Failed to deposit funds: ", err.Error())
		respDto.ErrorDescription = "Deposit Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 9. Update Operator Details in Cache
	operatorDTO, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("OperatorDeposit: Failed to get Operator Details: ", err.Error())
	}
	operatorDTO.Balance = operatorDTO.Balance + creditAmount
	cache.SetOperatorDetails(operatorDTO)
	operatorDto.Balance += creditAmount
	// 10. Send success response
	respDto.Balance = operatorDto.Balance
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func OperatorWithdraw(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 0. Create Default Response Object
	respDto := dto.OperatorBalanceRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	// 1. Parse request body to Request Object
	reqDto := dto.OperatorWithdrawReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("OperatorWithdraw: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorWithdraw: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorWithdraw: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	if operatorId == "" {
		log.Println("OperatorWithdraw: OperatorId is missing!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("OperatorWithdraw: PartnerId is missing!!!")
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	operatorDto, err := database.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("OperatorWithdraw: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = INVALID_OPERATORID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if len(operatorDto.Partners) == 0 {
		log.Println("OperatorWithdraw: Operator doesnt have partners!!!")
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	var found bool = false
	partner := operatordto.Partner{}
	for _, partnerDto := range operatorDto.Partners {
		if partnerDto.PartnerId == partnerId {
			partner = partnerDto
			found = true
			break
		}
	}
	if found == false {
		log.Println("OperatorWithdraw: Invalid PartnerId - ", partnerId)
		respDto.ErrorDescription = INVALID_PARTNERID_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Check for sufficient operator balance
	debitAmount := reqDto.DebitAmount * float64(partner.Rate)
	if debitAmount > operatorDto.Balance {
		// send failure response
		log.Println("OperatorWithdraw: debit amount is greater than the balance!")
		log.Println("OperatorWithdraw: Operator balance is - ", operatorDto.Balance)
		log.Println("OperatorWithdraw: Debit amount is - ", debitAmount)
		respDto.ErrorDescription = "Insufficient Funds!"
		respDto.Balance = operatorDto.Balance / float64(partner.Rate)
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Add to Funds Transaction collection
	operatorLedger := models.OperatorLedgerDto{}
	operatorLedger.OperatorId = reqDto.OperatorId
	operatorLedger.OperatorName = operatorDto.OperatorName
	operatorLedger.TransactionType = "Withdraw-Funds"
	operatorLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	operatorLedger.ReferenceId = ""
	operatorLedger.Amount = debitAmount * -1
	err = database.InsertOperatorLedger(operatorLedger)
	if err != nil {
		// 8.1. inserting ledger document failed
		log.Println("OperatorWithdraw: insert ledger failed with error - ", err.Error())
		respDto.ErrorDescription = "Withdraw Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 9. Update Operator Balance
	err = database.UpdateOperatorBalance(operatorId, debitAmount*-1)
	if err != nil {
		// 7.1. update failed, send failure response
		log.Println("OperatorWithdraw: Failed to deposit funds: ", err.Error())
		respDto.ErrorDescription = "Withdraw Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 10. Update Operator Details in cache
	operatorDTO, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("OperatorWithdraw: Failed to get Operator Details from cache: ", err.Error())
	}
	operatorDTO.Balance = operatorDTO.Balance - debitAmount
	cache.SetOperatorDetails(operatorDTO)
	operatorDto.Balance -= debitAmount
	// 11. Send success response
	respDto.Balance = operatorDto.Balance
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func ViewFundStatement(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 0. Create Default Response Object
	respDto := dto.OperatorStatementRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 1. Parse request body to Request Object
	reqDto := dto.OperatorStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("OperatorWithdraw: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("ViewFundStatement: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ViewFundStatement: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// Get Operator Ledger for Operator
	operatorLedger, err := database.GetOperatorLedgersForFundStatement(reqDto.OperatorId)
	if err != nil {
		log.Println("ViewFundStatement: Failed to get Operator Ledger: ", err.Error())
		respDto.ErrorDescription = "Failed to get Operator Ledger"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Send success response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.OperatorStatement = operatorLedger
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Function to sync sport
func SyncSports(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	reqDto := commondto.SyncSportsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SyncSports: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("OperatorWithdraw: Token Validation Failed")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("OperatorWithdraw: User not Permitted to access the API")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//Get All Operators
	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("SyncSports: Failed to get operators: ", err.Error())
		respDto.Status = "RS_ERROR"
		return c.JSON(err.Error())
	}
	// Get All Sports for a Provider
	sports, err := database.GetSports(reqDto.ProviderId)
	if err != nil {
		log.Println("SyncSports: Failed to get sports: ", err.Error())
		respDto.Status = "RS_ERROR"
		return c.JSON(err.Error())
	}
	sportStatsuDtos := []models.SportStatus{}
	for _, sport := range sports {
		// Get All Providers for a Sport
		// sport.SportId
		for _, operator := range operators {
			sportKey := operator.OperatorId + "-" + reqDto.ProviderId + "-" + sport.SportId
			sportStatus, err := database.GetSportStatus(sportKey)
			if err != nil && err.Error() != "mongo: no documents in result" {
				log.Println("SyncSports: Failed to get sport status: ", err.Error())
				respDto.Status = "RS_ERROR"
				return c.JSON(err.Error())
			}
			if sportStatus.SportKey == "" {
				sportStatus := models.SportStatus{}
				sportStatus.SportKey = sportKey
				sportStatus.OperatorId = operator.OperatorId
				sportStatus.OperatorName = operator.OperatorName
				sportStatus.ProviderId = sport.ProviderId
				sportStatus.ProviderName = sport.ProviderName
				sportStatus.SportId = sport.SportId
				sportStatus.SportName = sport.SportName
				sportStatus.ProviderStatus = "BLOCKED"
				sportStatus.OperatorStatus = "BLOCKED"
				sportStatus.Favourite = false
				sportStatus.CreatedAt = time.Now().Unix()
				sportStatsuDtos = append(sportStatsuDtos, sportStatus)
			}
		}
	}
	if len(sportStatsuDtos) > 0 {
		err := database.InsertManySportStatus(sportStatsuDtos)
		if err != nil {
			log.Println("SyncSports: Failed to insert sport status: ", err.Error())
			respDto.Status = "RS_ERROR"
			return c.JSON(err.Error())
		}
	}
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.JSON(respDto)
}

//Function to Sync Competitions
func SyncCompetitions(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "DEPRICATED!!!"
	return c.Status(fiber.StatusOK).JSON(respDto)

	// //Get All Operators
	// operators, err := database.GetAllOperators()
	// if err != nil {
	// 	log.Println("SyncCompetitions: Failed to get operators: ", err.Error())
	// 	respDto.Status = "RS_ERROR"
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.JSON(respDto)
	// }
	// log.Println("SyncCompetitions: GetAllOperators: ", len(operators))
	// // Get All Provider
	// providers, err := database.GetAllProviders()
	// if err != nil {
	// 	log.Println("SyncCompetitions: Failed to get providers: ", err.Error())
	// 	respDto.Status = "RS_ERROR"
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.JSON(respDto)
	// }
	// log.Println("SyncCompetitions: GetAllProviders: ", len(providers))
	// // Get All Sports
	// competitionStatsuDtos := []models.CompetitionStatus{}
	// for _, provider := range providers {
	// 	sports, err := database.GetSports(provider.ProviderId)
	// 	if err != nil {
	// 		log.Println("SyncCompetitions: Failed to get sports: ", err.Error())
	// 		respDto.Status = "RS_ERROR"
	// 		respDto.ErrorDescription = err.Error()
	// 		return c.JSON(respDto)
	// 	}
	// 	log.Println("SyncCompetitions: GetSports: ", len(sports))
	// 	for _, sport := range sports {
	// 		competitions, err := database.GetCompetitionsbySport(provider.ProviderId, sport.SportId)
	// 		// Get All Competitions by Provider and Sport
	// 		if err != nil {
	// 			log.Println("SyncCompetitions: Failed to get competitions: ", err.Error())
	// 			respDto.Status = "RS_ERROR"
	// 			respDto.ErrorDescription = err.Error()
	// 			return c.JSON(respDto)
	// 		}
	// 		log.Println("SyncCompetitions: GetCompetitionsbyProviderAndSport: for Provider: " + provider.ProviderId + " Sport " + sport.SportId + " " + strconv.Itoa(len(competitions)))
	// 		competitionsStatus, err := database.GetCompetitionStatusByPrSport(provider.ProviderId, sport.SportId)
	// 		if err != nil {
	// 			log.Println("SyncCompetitions: Failed to get competitions status: ", err.Error())
	// 			respDto.Status = "RS_ERROR"
	// 			respDto.ErrorDescription = err.Error()
	// 			return c.JSON(respDto)
	// 		}
	// 		log.Println("SyncCompetitions: GetCompetitionStatusByPrSport: for Provider: " + provider.ProviderId + " Sport " + sport.SportId + " " + strconv.Itoa(len(competitionsStatus)))
	// 		log.Println("SyncCompetitions: Expected CompetitionStatus Count is ", len(operators)*len(competitions))
	// 		log.Println("SyncCompetitions: Actual   CompetitionStatus Count is ", len(competitionsStatus))
	// 		log.Println("SyncCompetitions: Subs     CompetitionStatus Count is ", (len(operators)*len(competitions))-len(competitionsStatus))
	// 		for _, operator := range operators {
	// 			for _, competition := range competitions {
	// 				competitionKey := operator.OperatorId + "-" + competition.CompetitionKey
	// 				flag := true
	// 				for _, competitionStatus := range competitionsStatus {
	// 					if competitionStatus.CompetitionKey == competitionKey {
	// 						flag = false
	// 						break
	// 					}
	// 				}
	// 				if flag {
	// 					competitionStatus := models.CompetitionStatus{}
	// 					competitionStatus.CompetitionKey = competitionKey
	// 					competitionStatus.OperatorId = operator.OperatorId
	// 					competitionStatus.OperatorName = operator.OperatorName
	// 					competitionStatus.ProviderId = competition.ProviderId
	// 					competitionStatus.ProviderName = competition.ProviderName
	// 					competitionStatus.SportId = competition.SportId
	// 					competitionStatus.SportName = competition.SportName
	// 					competitionStatus.CompetitionId = competition.CompetitionId
	// 					competitionStatus.CompetitionName = competition.CompetitionName
	// 					competitionStatus.ProviderStatus = "ACTIVE"
	// 					competitionStatus.OperatorStatus = "ACTIVE"
	// 					competitionStatus.Favourite = false
	// 					competitionStatsuDtos = append(competitionStatsuDtos, competitionStatus)
	// 				}
	// 			}
	// 		}
	// 		log.Println("SyncCompetitions: Inserting competition status: ", len(competitionStatsuDtos))
	// 		log.Println("SyncCompetitions: ========================================== ")
	// 	}
	// }
	// log.Println("SyncCompetitions: Inserting competition status: ", len(competitionStatsuDtos))
	// if len(competitionStatsuDtos) > 0 {
	// 	var batches [][]models.CompetitionStatus
	// 	for i := 0; i < len(competitionStatsuDtos); i += 100 {
	// 		batches = append(batches, competitionStatsuDtos[i:min(i+100, len(competitionStatsuDtos))])
	// 	}

	// 	for _, batch := range batches {
	// 		err := database.InsertManyCompetitionStatus(batch)
	// 		if err != nil {
	// 			log.Println("SyncCompetitions: Failed to insert competition status: ", err.Error())
	// 		} else {
	// 			log.Println("SyncCompetitions: Inserted competition status: ", len(batch))
	// 		}
	// 		//time.Sleep(2000 * time.Millisecond) // Sleep for 2 seconds
	// 	}
	// }
	// respDto.Status = "RS_OK"
	// respDto.ErrorDescription = ""
	// return c.JSON(respDto)
}

//Function to Sync Events
func SyncEvents(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "DEPRICATED!!!"
	return c.JSON(respDto)

	// // 1. Parse request body to Request Object
	// reqDto := commondto.SyncEventsReqDto{}
	// if err := c.BodyParser(&reqDto); err != nil {
	// 	log.Println("SyncEvents: Body Parsing failed")
	// 	respDto.ErrorDescription = "Invalid Request"
	// 	return c.JSON(respDto)
	// }

	// // Tknmeta, ok := Authenticate(c)
	// // if !ok {
	// // 	// 2.1. Token validaton failed.
	// // 	log.Println("SyncEvents: Token Validation Failed")
	// // 	respDto.Status = "RS_ERROR"
	// // 	respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
	// // 	return c.JSON(respDto)
	// // }
	// // // 3. Check Role Permissions
	// // if !IsApplicable(Tknmeta, c.OriginalURL()) {
	// // 	log.Println("SyncEvents: User not Permitted to access the API")
	// // 	respDto.Status = "RS_ERROR"
	// // 	respDto.ErrorDescription = UNAUTH_ACCESS
	// // 	return c.JSON(respDto)
	// // }

	// //Get All Operators
	// operators, err := database.GetAllOperators()
	// if err != nil {
	// 	log.Println("SyncEvents: Failed to get operators: ", err.Error())
	// 	respDto.Status = "RS_ERROR"
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.JSON(respDto)
	// }
	// // 6 Get Events for Layer 2
	// var l2Events []coredto.EventDto
	// switch reqDto.ProviderId {
	// case "Dream":
	// 	l2Events, err = dream.GetEvents(reqDto.SportId)
	// 	if err != nil {
	// 		log.Println("RecentCompetitions: Get l2Events failed for Dream with error - ", err.Error())
	// 		respDto.ErrorDescription = "Get l2Events Failed!"
	// 		return c.Status(fiber.StatusOK).JSON(respDto)
	// 	}
	// case "BetFair":
	// 	l2Events, err = betfair.GetEvents(reqDto.SportId)
	// 	if err != nil {
	// 		log.Println("RecentCompetitions: Get l2Events failed for BetFair with error - ", err.Error())
	// 		respDto.ErrorDescription = "Get l2Events Failed!"
	// 		return c.Status(fiber.StatusOK).JSON(respDto)
	// 	}
	// case "SportRadar":
	// 	l2Events, err = sportradar.GetEvents(reqDto.SportId)
	// 	if err != nil {
	// 		log.Println("RecentCompetitions: Get l2Events failed for SportRadar with error - ", err.Error())
	// 		respDto.ErrorDescription = "Get l2Events Failed!"
	// 		return c.Status(fiber.StatusOK).JSON(respDto)
	// 	}
	// default:
	// 	log.Println("RecentCompetitions: ProviderId is invalid")
	// 	respDto.ErrorDescription = "ProviderId is invalid!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }

	// competitionIds := []string{}
	// compIdMap := map[string]bool{}
	// for _, l2Event := range l2Events {
	// 	if l2Event.CompetitionId == "" || l2Event.CompetitionId == "-1" {
	// 		continue
	// 	}
	// 	if _, ok := compIdMap[l2Event.CompetitionId]; !ok {
	// 		competitionIds = append(competitionIds, l2Event.CompetitionId)
	// 		compIdMap[l2Event.CompetitionId] = true
	// 	}
	// }

	// // Get All Events for a Sport and provider
	// events, err := database.GetEventsLast25(reqDto.ProviderId, reqDto.SportId, competitionIds)
	// if err != nil {
	// 	log.Println("SyncEvents: Failed to get events: ", err.Error())
	// 	respDto.Status = "RS_ERROR"
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.JSON(respDto)
	// }
	// eventKeys := []string{}
	// for _, event := range events {
	// 	for _, operator := range operators {
	// 		eventKey := operator.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + event.EventId
	// 		eventKeys = append(eventKeys, eventKey)
	// 	}
	// }
	// eventStatus, err := database.GetUpdatedEventStatus(eventKeys)
	// if err != nil && err.Error() != "mongo: no documents in result" {
	// 	log.Println("SyncEvents: Failed to get event status: ", err.Error())
	// 	respDto.Status = "RS_ERROR"
	// 	respDto.ErrorDescription = err.Error()
	// 	return c.JSON(respDto)
	// }
	// log.Println("SyncEvents: Event Status: ", len(eventStatus))
	// eventStatsuDtos := []models.EventStatus{}
	// for _, event := range events {
	// 	// Get All Providers for a Sport
	// 	// sport.SportId
	// 	for _, operator := range operators {
	// 		eventKey := operator.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + event.EventId
	// 		eventStatus, err := database.GetEventStatus(eventKey)
	// 		if err != nil && err.Error() != "mongo: no documents in result" {
	// 			log.Println("SyncEvents: Failed to get event status: ", err.Error())
	// 			respDto.Status = "RS_ERROR"
	// 			respDto.ErrorDescription = err.Error()
	// 			return c.JSON(respDto)
	// 		}
	// 		if eventStatus.EventKey == "" {
	// 			eventStatus := models.EventStatus{}
	// 			eventStatus.EventKey = eventKey
	// 			eventStatus.OperatorId = operator.OperatorId
	// 			eventStatus.OperatorName = operator.OperatorName
	// 			eventStatus.ProviderId = event.ProviderId
	// 			eventStatus.ProviderName = event.ProviderName
	// 			eventStatus.SportId = event.SportId
	// 			eventStatus.SportName = event.SportName
	// 			eventStatus.CompetitionId = event.CompetitionId
	// 			eventStatus.CompetitionName = event.CompetitionName
	// 			eventStatus.EventId = event.EventId
	// 			eventStatus.EventName = event.EventName
	// 			eventStatus.ProviderStatus = "ACTIVE"
	// 			eventStatus.OperatorStatus = "ACTIVE"
	// 			eventStatus.Favourite = false
	// 			eventStatsuDtos = append(eventStatsuDtos, eventStatus)
	// 		}
	// 	}
	// }
	// if len(eventStatsuDtos) > 0 {
	// 	err := database.InsertManyEventStatus(eventStatsuDtos)
	// 	if err != nil {
	// 		log.Println("SyncEvents: Failed to insert event status: ", err.Error())
	// 		respDto.Status = "RS_ERROR"
	// 		return c.JSON(err.Error())
	// 	}
	// }
	// respDto.Status = "RS_OK"
	// respDto.ErrorDescription = ""
	// return c.JSON(respDto)
}

func GetPortalUsers(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.GetPortalUsersRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	reqDto := portaldto.GetPortalUsersReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetPortalUsers: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetPortalUsers: Token Validation Failed")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetPortalUsers: User not Permitted to access the API")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	//Get All Portal Users
	portalUsers, err := database.GetPortalUsers()
	if err != nil {
		log.Println("GetPortalUsers: Failed to get portal users: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.PortalUsers = portalUsers
	return c.JSON(respDto)
}

func ResetPasswordForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	reqDto := portaldto.ResetPasswordReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ResetPasswordForOP: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("ResetPasswordForOP: Token Validation Failed")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ResetPasswordForOP: User not Permitted to access the API")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}
	if Tknmeta.UserId != reqDto.UserId {
		log.Println("ResetPasswordForOP: UserId donot match Token Claims")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = "Invalid Password or UserId"
		return c.JSON(respDto)
	}
	//Get All Portal Users
	portalUser, err := database.GetPortalUserDetailsByUserId(reqDto.UserId)
	if err != nil {
		log.Println("ResetPasswordForOP: Failed to get portal user: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	//Check Password and Confirm Password
	if !CheckPasswordHash(reqDto.OldPassword, portalUser.Password) {
		log.Println("Portal Login: Invalid Password or UserId")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = "Invalid Password or UserId"
		return c.JSON(respDto)
	}

	portalUser.Password, err = HashPassword(reqDto.NewPassword)
	if err != nil {
		log.Println("ResetPasswordForOP: Failed to hash password: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	//Update Portal User
	err = database.UpdatePortalUserDetails(portalUser)
	if err != nil {
		log.Println("ResetPasswordForOP: Failed to update portal user: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.UserId, constants.PORTAL_USER)
	}
	return c.JSON(respDto)
}

func ResetPasswordForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	reqDto := portaldto.ResetPasswordReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ResetPasswordForSAP: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("ResetPasswordForSAP: Token Validation Failed")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ResetPasswordForSAP: User not Permitted to access the API")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}
	if Tknmeta.UserId != reqDto.UserId {
		log.Println("ResetPasswordForSAP: UserId donot match Token Claims")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = "Invalid Password or UserId"
		return c.JSON(respDto)
	}
	//Get All Portal Users
	portalUser, err := database.GetPortalUserDetailsByUserId(reqDto.UserId)
	if err != nil {
		log.Println("ResetPasswordForSAP: Failed to get portal user: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	//Check Password and Confirm Password
	if !CheckPasswordHash(reqDto.OldPassword, portalUser.Password) {
		log.Println("Portal Login: Invalid Password or UserId")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = "Invalid Password or UserId"
		return c.JSON(respDto)
	}

	portalUser.Password, err = HashPassword(reqDto.NewPassword)
	if err != nil {
		log.Println("ResetPasswordForSAP: Failed to hash password: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	//Update Portal User
	err = database.UpdatePortalUserDetails(portalUser)
	if err != nil {
		log.Println("ResetPasswordForSAP: Failed to update portal user: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.UserId, constants.PORTAL_USER)
	}
	return c.JSON(respDto)
}

func ResetOPPasswordForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	reqDto := portaldto.ResetPasswordForOPReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ResetOPPasswordForSAP: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("ResetOPPasswordForSAP: Token Validation Failed")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ResetOPPasswordForSAP: User not Permitted to access the API")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	//Get All Portal Users
	portalUser, err := database.GetPortalUserDetailsByUserId(reqDto.UserId)
	if err != nil {
		log.Println("ResetOPPasswordForSAP: Failed to get portal user: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	if portalUser.OperatorId != reqDto.OperatorId {
		log.Println("ResetOPPasswordForSAP: OperatorId donot match Portal User OperatorId")
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = "Invalid Operator Id"
		return c.JSON(respDto)
	}

	portalUser.Password, err = HashPassword(reqDto.Password)
	if err != nil {
		log.Println("ResetOPPasswordForSAP: Failed to hash password: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	//Update Portal User
	err = database.UpdatePortalUserDetails(portalUser)
	if err != nil {
		log.Println("ResetOPPasswordForSAP: Failed to update portal user: ", err.Error())
		respDto.Status = "RS_ERROR"
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	respDto.Status = "RS_OK"
	respDto.ErrorDescription = "Password Changed Successfully"
	// Audit Portal
	if respDto.Status == OK_STATUS {
		PortalAudit(*Tknmeta, c, reqDto.UserId, constants.PORTAL_USER)
	}
	return c.JSON(respDto)
}

// Add Partner API
// @Summary      Operator Partner
// @Description  Adding a partner to an operator
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        add-partner  body      dto.AddPartnerReqDto  true  "AddPartnerReqDto model is used"
// @Success      200          {object}  dto.CommonPortalRespDto
// @Failure      503          {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/add-partner [post]
func AddPartner(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 0. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 1. Parse request body to Request Object
	reqDto := dto.AddPartnerReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("AddPartner: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("AddPartner: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("AddPartner: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Request Check

	// 4. Get Operator Details from DB
	operatorId := reqDto.OperatorId
	operatorDto, err := database.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("AddPartner: GetOperatorDetails failed with error: ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Add Partner to Operator
	operatorDto.Partners = append(operatorDto.Partners, reqDto.Partner)
	// 6. Update Operator
	err = database.ReplaceOperator(operatorDto)
	if err != nil {
		log.Println("AddPartner: ReplaceOperator failed with error : ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Create Partner Status
	providers, err := database.GetAllProviders()
	if err != nil {
		log.Println("AddPartner: Get Operators failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerStatus := []models.PartnerStatus{}
	for _, provider := range providers {
		partnerSts := models.PartnerStatus{}
		partnerSts.PartnerKey = reqDto.OperatorId + "-" + reqDto.Partner.PartnerId + "-" + provider.ProviderId
		partnerSts.OperatorId = reqDto.OperatorId
		partnerSts.OperatorName = operatorDto.OperatorName
		partnerSts.PartnerId = reqDto.Partner.PartnerId
		partnerSts.ProviderId = provider.ProviderId
		partnerSts.ProviderName = provider.ProviderName
		partnerSts.ProviderStatus = BLOCKED
		partnerSts.OperatorStatus = BLOCKED
		partnerStatus = append(partnerStatus, partnerSts)
	}
	// 7.2 Save in provider_status to DB
	if len(partnerStatus) > 0 {
		err = database.InsertManyPartnerStatus(partnerStatus)
		if err != nil {
			log.Println("AddPartner: Insert ProviderStatus failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.JSON(respDto)
		}
		// 7.3 Save in provider_status to Cache
		for _, ps := range partnerStatus {
			cache.SetPartnerStatus(ps)
		}
	}
	// 9. Send success response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GetRole(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 0. Create Default Response Object
	respDto := dto.GetRoleRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 1. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 1.1. Token validaton failed.
		log.Println("AddPartner: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 2. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("AddPartner: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 3. Send success response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.Roles = []string{"OperatorAdmin", "SAPAdmin", "OperatorUser", "SAPUser"}

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Market APIs

// Get Markets API
// @Summary      GetMarkets by Platform Admin
// @Description  Get all Markets for an event by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        GetMarketsForSAP  body      dto.GetMarketsReqDto  true  "GetMarketsReqDto model is used"
// @Success      200               {object}  dto.GetMarketsRespDto
// @Failure      503               {object}  dto.GetMarketsRespDto
// @Router       /portal/sapadmin/get-markets [post]
func GetMarketsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.GetMarketsRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Markets = []models.Market{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(dto.GetMarketsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetMarkets: Body Parsing failed with error - ", err.Error())
		log.Println("GetMarkets: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("GetMarkets: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetMarkets: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	markets, err := handler.GetMarkets(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
	if err != nil {
		log.Println("GetMarkets: handler.GetMarkets failed with error - ", err.Error())
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Markets = append(respDto.Markets, markets...)
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Market API
// @Summary      Blocking a Market by Platform Admin
// @Description  Blocking a Market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        BlockMarketForSAP  body      dto.UpdateMarketsReqDto  true  "UpdateMarketsReqDto model is used"
// @Success      200                {object}  dto.CommonPortalRespDto
// @Failure      503                {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/block-market [post]
func BlockMarketForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("BlockMarketForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("BlockMarketForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("BlockMarketForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("BlockMarketForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("BlockMarketForSAP: database.GetMarket failed with error - ", err.Error())
		log.Println("BlockMarketForSAP: database.GetMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	market.Status = constants.SAP.ObjectStatus.BLOCKED()
	err = database.ReplaceMarket(market)
	if err != nil {
		log.Println("BlockMarketForSAP: database.ReplaceMarket failed with error - ", err.Error())
		log.Println("BlockMarketForSAP: database.ReplaceMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Market API
// @Summary      Unblocking a Market by Platform Admin
// @Description  Unblocking a Market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        UnblockMarketForSAP  body      dto.UpdateMarketsReqDto  true  "UpdateMarketsReqDto model is used"
// @Success      200                  {object}  dto.CommonPortalRespDto
// @Failure      503                  {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/unblock-market [post]
func UnblockMarketForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UnblockMarketForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("UnblockMarketForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("UnblockMarketForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UnblockMarketForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("UnblockMarketForSAP: database.GetMarket failed with error - ", err.Error())
		log.Println("UnblockMarketForSAP: database.GetMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	market.Status = constants.SAP.ObjectStatus.ACTIVE()
	err = database.ReplaceMarket(market)
	if err != nil {
		log.Println("UnblockMarketForSAP: database.ReplaceMarket failed with error - ", err.Error())
		log.Println("UnblockMarketForSAP: database.ReplaceMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Suspend Market API
// @Summary      Suspending a Market by Platform Admin
// @Description  Suspending a Market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        SuspendMarketForSAP  body      dto.UpdateMarketsReqDto  true  "UpdateMarketsReqDto model is used"
// @Success      200                  {object}  dto.CommonPortalRespDto
// @Failure      503                  {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/suspend-market [post]
func SuspendMarketForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SuspendMarketForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("SuspendMarketForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("SuspendMarketForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("SuspendMarketForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("SuspendMarketForSAP: database.GetMarket failed with error - ", err.Error())
		log.Println("SuspendMarketForSAP: database.GetMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update market in DB
	market.IsSuspended = true
	err = database.ReplaceMarket(market)
	if err != nil {
		log.Println("SuspendMarketForSAP: database.ReplaceMarket failed with error - ", err.Error())
		log.Println("SuspendMarketForSAP: database.ReplaceMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Resume Market API
// @Summary      Resuming a Market by Platform Admin
// @Description  Resuming a Market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        ResumeMarketForSAP  body      dto.UpdateMarketsReqDto  true  "UpdateMarketsReqDto model is used"
// @Success      200                 {object}  dto.CommonPortalRespDto
// @Failure      503                 {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/resume-market [post]
func ResumeMarketForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ResumeMarketForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("ResumeMarketForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("ResumeMarketForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ResumeMarketForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("ResumeMarketForSAP: database.GetMarket failed with error - ", err.Error())
		log.Println("ResumeMarketForSAP: database.GetMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	market.IsSuspended = false
	err = database.ReplaceMarket(market)
	if err != nil {
		log.Println("ResumeMarketForSAP: database.ReplaceMarket failed with error - ", err.Error())
		log.Println("ResumeMarketForSAP: database.ReplaceMarket failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Op Markets API
// @Summary      GetMarkets by Platform Admin
// @Description  Get all Markets for a market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        GetOpMarketsForSAP  body      dto.GetOpMarketsReqDto  true  "GetOpMarketsReqDto model is used"
// @Success      200                 {object}  dto.GetMarketStatusRespDto
// @Failure      503                 {object}  dto.GetMarketStatusRespDto
// @Router       /portal/sapadmin/get-op-markets [post]
func GetOpMarketsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetMarketStatusRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Markets = []models.MarketStatus{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(dto.GetOpMarketsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetOpMarketsForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("GetOpMarketsForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("GetOpMarketsForSAP: Token Validation Failed")
		log.Println("GetOpMarketsForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetOpMarketsForSAP: User not Permitted to access the API")
		log.Println("GetOpMarketsForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Markets from MarketStatus table
	// 7. Get MarketStatus by marketKeys
	marketStatus, err := database.GetMarketStatusesByMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	if err != nil {
		log.Println("GetOpMarketsForSAP: database.GetMarketStatuses failed with error - ", err.Error())
		log.Println("GetOpMarketsForSAP: Req. Body is - ", bodyStr)
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Send response
	respDto.Markets = append(respDto.Markets, marketStatus...)
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Market API
// @Summary      Blocking an Operator Market by Platform Admin
// @Description  Blocking an Operator Market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        BlockOPMarketForSAP  body      dto.UpdateOpMarketsReqDto  true  "UpdateOpMarketsReqDto model is used"
// @Success      200                  {object}  dto.CommonPortalRespDto
// @Failure      503                  {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/block-op-market [post]
func BlockOPMarketForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateOpMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("BlockOPMarketForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("BlockOPMarketForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("BlockOPMarketForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("BlockOPMarketForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	marketStatus, err := database.GetMarketStatus(marketKey)
	if err != nil {
		log.Println("BlockOPMarketForSAP: database.GetMarketStatus failed with error - ", err.Error())
		log.Println("BlockOPMarketForSAP: database.GetMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	marketStatus.ProviderStatus = constants.SAP.ObjectStatus.BLOCKED()
	err = database.ReplaceMarketStatus(marketStatus)
	if err != nil {
		log.Println("BlockOPMarketForSAP: database.ReplaceMarketStatus failed with error - ", err.Error())
		log.Println("BlockOPMarketForSAP: database.ReplaceMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Market API
// @Summary      Unblocking an Operator Market by Platform Admin
// @Description  Unblocking an Operator Market by Platform Admin
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        UnblockOPMarketForSAP  body      dto.UpdateOpMarketsReqDto  true  "UpdateOpMarketsReqDto model is used"
// @Success      200                    {object}  dto.CommonPortalRespDto
// @Failure      503                    {object}  dto.CommonPortalRespDto
// @Router       /portal/sapadmin/unblock-op-market [post]
func UnblockOPMarketForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateOpMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UnblockOPMarketForSAP: Body Parsing failed with error - ", err.Error())
		log.Println("UnblockOPMarketForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("UnblockOPMarketForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UnblockOPMarketForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := reqDto.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	marketStatus, err := database.GetMarketStatus(marketKey)
	if err != nil {
		log.Println("UnblockOPMarketForSAP: database.GetMarketStatus failed with error - ", err.Error())
		log.Println("UnblockOPMarketForSAP: database.GetMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	marketStatus.ProviderStatus = constants.SAP.ObjectStatus.ACTIVE()
	err = database.ReplaceMarketStatus(marketStatus)
	if err != nil {
		log.Println("UnblockOPMarketForSAP: database.ReplaceMarketStatus failed with error - ", err.Error())
		log.Println("UnblockOPMarketForSAP: database.ReplaceMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Markets API
// @Summary      GetMarkets by Operator Admin
// @Description  Get all Markets for an event by Operator Admin
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        GetMarketsForOP  body      dto.GetMarketsReqDto  true  "GetMarketsReqDto model is used"
// @Success      200              {object}  dto.GetMarketStatusRespDto
// @Failure      503              {object}  dto.GetMarketStatusRespDto
// @Router       /portal/opadmin/get-markets [post]
func GetMarketsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.GetMarketStatusRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Markets = []models.MarketStatus{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := new(dto.GetMarketsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetMarketsForOP: Body Parsing failed with error - ", err.Error())
		log.Println("GetMarketsForOP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("GetMarketsForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetMarketsForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(Tknmeta.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetMarketsForOP:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("GetMarketsForOP: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Markets from Markets table
	markets, err := handler.GetMarkets(reqDto.ProviderId, reqDto.SportId, reqDto.EventId)
	if err != nil {
		log.Println("GetMarketsForOP: handler.GetMarkets failed with error - ", err.Error())
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get list of active markets
	marketKeys := []string{}
	for _, market := range markets {
		if market.Status != constants.SAP.ObjectStatus.ACTIVE() {
			continue
		}
		marketKey := Tknmeta.OperatorId + "-" + market.MarketKey
		marketKeys = append(marketKeys, marketKey)
	}
	// 7. Get MarketStatus by marketKeys
	marketStatus, err := database.GetMarketStatuses(marketKeys)
	if err != nil {
		log.Println("GetMarketsForOP: database.GetMarketStatuses failed with error - ", err.Error())
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Send response
	respDto.Markets = append(respDto.Markets, marketStatus...)
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Block Market API
// @Summary      Blocking a Market by Operator Admin
// @Description  Blocking a Market by Operator Admin
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        BlockMarketForOP  body      dto.UpdateMarketsReqDto  true  "UpdateMarketsReqDto model is used"
// @Success      200               {object}  dto.CommonPortalRespDto
// @Failure      503               {object}  dto.CommonPortalRespDto
// @Router       /portal/opadmin/block-market [post]
func BlockMarketForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("BlockMarketForOP: Body Parsing failed with error - ", err.Error())
		log.Println("BlockMarketForOP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("BlockMarketForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("BlockMarketForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := Tknmeta.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	marketStatus, err := database.GetMarketStatus(marketKey)
	if err != nil {
		log.Println("BlockMarketForOP: database.GetMarketStatus failed with error - ", err.Error())
		log.Println("BlockMarketForOP: database.GetMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	marketStatus.OperatorStatus = constants.SAP.ObjectStatus.BLOCKED()
	err = database.ReplaceMarketStatus(marketStatus)
	if err != nil {
		log.Println("BlockMarketForOP: database.ReplaceMarketStatus failed with error - ", err.Error())
		log.Println("BlockMarketForOP: database.ReplaceMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Unblock Market API
// @Summary      Unblocking a Market by Operator Admin
// @Description  Unblocking a Market by Operator Admin
// @Tags         Portal-OperatorAdmin
// @Accept       json
// @Produce      json
// @Param        UnblockMarketForOP  body      dto.UpdateMarketsReqDto  true  "UpdateMarketsReqDto model is used"
// @Success      200                 {object}  dto.CommonPortalRespDto
// @Failure      503                 {object}  dto.CommonPortalRespDto
// @Router       /portal/opadmin/unblock-market [post]
func UnblockMarketForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := dto.CommonPortalRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := dto.UpdateMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UnblockMarketForOP: Body Parsing failed with error - ", err.Error())
		log.Println("UnblockMarketForOP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 3.1. Token validaton failed.
		log.Println("UnblockMarketForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UnblockMarketForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Get Market from DB
	eventKey := Tknmeta.OperatorId + "-" + reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	marketKey := eventKey + "-" + reqDto.MarketId
	marketStatus, err := database.GetMarketStatus(marketKey)
	if err != nil {
		log.Println("UnblockMarketForOP: database.GetMarketStatus failed with error - ", err.Error())
		log.Println("UnblockMarketForOP: database.GetMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Update Market in DB
	marketStatus.OperatorStatus = constants.SAP.ObjectStatus.ACTIVE()
	err = database.ReplaceMarketStatus(marketStatus)
	if err != nil {
		log.Println("UnblockMarketForOP: database.ReplaceMarketStatus failed with error - ", err.Error())
		log.Println("UnblockMarketForOP: database.ReplaceMarketStatus failed for marketkey - ", marketKey)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Send Success Response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GetBalanceForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.GetBalanceRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0.0
	respDto.Currency = "PTS"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetBalance: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetBalance: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(Tknmeta.OperatorId)
	if err != nil {
		// 4.1. Operator not found, Return Error
		log.Println("GetBalance:  Get Operator Details failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}
	var rate int32 = 1
	if len(operatorDto.Partners) > 0 {
		rate = operatorDto.Partners[0].Rate
		respDto.Currency = operatorDto.Partners[0].Currency
		for _, partner := range operatorDto.Partners {
			if partner.Status != constants.SAP.ObjectStatus.ACTIVE() {
				continue
			}
			rate = partner.Rate
			respDto.Currency = partner.Currency
			break
		}
	}

	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.Balance = utils.Truncate64(operatorDto.Balance / float64(rate))

	return c.JSON(respDto)
}

func GetPartnerIdsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.GetPartnerIdsRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.PartnerIds = []operatordto.Partner{}

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetPartnerIds: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetPartnerIds: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 4. Get Operator DTO
	operatorDetails, err := database.GetOperatorDetails(Tknmeta.OperatorId)
	if err != nil {
		log.Println("GetOperatorDetails: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	for _, partner := range operatorDetails.Partners {
		respDto.PartnerIds = append(respDto.PartnerIds, partner)
	}

	// 5. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.JSON(respDto)
}

func GetPartnerIds(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.GetPartnerIdsRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.PartnerIds = []operatordto.Partner{}

	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("GetPartnerIds: Req. Body is - ", bodyStr)
	reqDto := new(dto.GetPartnerIdsReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("GetPartnerIds: Body Parsing failed with error - ", err.Error())
		log.Println("GetPartnerIds: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetPartnerIds: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetPartnerIds: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 4. Get Operator DTO
	operatorDetails, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		log.Println("GetOperatorDetails: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	for _, partner := range operatorDetails.Partners {
		respDto.PartnerIds = append(respDto.PartnerIds, partner)
	}

	// 5. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.JSON(respDto)
}

func BlockPortalUserForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.CommonPortalRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BlockPortalUser: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("BlockPortalUser: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 4. Get Request DTO
	reqDto := new(dto.BlockPortalUserReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("BlockPortalUser: Failed to parse request body with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 5. Get Portal User DTO
	portalUserDto, err := database.GetPortalUserDetailsByUserKey(reqDto.UserKey)
	if err != nil {
		log.Println("BlockPortalUser: GetPortalUserDetailsByUserKey failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 6. Check role of the user
	if portalUserDto.Role == "SAPAdmin" {
		log.Println("BlockPortalUser: SAP Admin cannot be blocked")
		respDto.ErrorDescription = "SAP Admin cannot be Blocked"
		return c.JSON(respDto)
	}

	// 7. Block Portal User
	portalUserDto.Status = constants.SAP.ObjectStatus.BLOCKED()
	err = database.UpdatePortalUserDetails(portalUserDto)
	if err != nil {
		log.Println("BlockPortalUser: UpdatePortalUserDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 8. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PORTAL USER " + BLOCKED
	return c.JSON(respDto)
}

func UnblockPortalUserForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.CommonPortalRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("UnblockPortalUser: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UnblockPortalUser: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 4. Get Request DTO
	reqDto := new(dto.UnblockPortalUserReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("UnblockPortalUser: Failed to parse request body with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 5. Get Portal User DTO
	portalUserDto, err := database.GetPortalUserDetailsByUserKey(reqDto.UserKey)
	if err != nil {
		log.Println("UnblockPortalUser: GetPortalUserDetailsByUserKey failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 6. Unblock Portal User
	portalUserDto.Status = constants.SAP.ObjectStatus.ACTIVE()
	err = database.UpdatePortalUserDetails(portalUserDto)
	if err != nil {
		log.Println("UnblockPortalUser: UpdatePortalUserDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 7. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PORTAL USER " + ACTIVE
	return c.JSON(respDto)
}

func ReplacePartnerForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 1. Create Default Response Object
	respDto := new(dto.CommonPortalRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("ReplacePartner: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ReplacePartner: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 4. Get Request DTO
	reqDto := new(dto.ReplacePartnerReqDto)
	err := c.BodyParser(reqDto)
	if err != nil {
		log.Println("ReplacePartner: Failed to parse request body with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 5. Get Portal User DTO
	operatorDTO, err := database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		log.Println("BlockOperator: Database Get failed with error - ", err.Error())
	}

	for num, partner := range operatorDTO.Partners {
		if partner.PartnerId == reqDto.PartnerId {
			partner.Rate = reqDto.Rate
		}
		// replace partner in partner list
		operatorDTO.Partners[num] = partner
	}
	// 6. Update operatorDTO
	err = database.ReplaceOperator(operatorDTO)
	if err != nil {
		log.Println("ReplacePartner: UpdateOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.JSON(respDto)
	}

	// 7. Send Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = "PARTNER UPDATED"

	return c.JSON(respDto)
}

func MatchedSports(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := responsedto.MatchSportResp{}
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = GENERAL_ERROR_DESC

	// Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.MatchSportReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("MatchedSports: Body Parsing failed: ", err.Error())
		log.Println("MatchedSports: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("MatchedSports: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("MatchedSports: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// Validate Request
	if reqDto.SportId == "" {
		log.Println("MatchedSports: Invalid Request - ", reqDto)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Call Service
	matchSports, err := betfair.MatchedSportsbySportId(reqDto.SportId)
	if err != nil {
		log.Println("MatchedSports: Service call failed: ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.JSON(respDto)
	}

	// Return Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.MatchedSports = matchSports

	return c.JSON(respDto)
}

func UnmatchedSports(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := responsedto.UnmatchSportResp{}
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = GENERAL_ERROR_DESC

	// Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.UnmatchSportReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UnmatchedSports: Body Parsing failed: ", err.Error())
		log.Println("UnmatchedSports: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("UnmatchedSports: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UnmatchedSports: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// Validate Request
	if reqDto.SportId == "" {
		log.Println("UnmatchedSports: Invalid Request - ", reqDto)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Call Service
	unMatchSports, err := betfair.UnMatchedSportsBySportId(reqDto.SportId)
	if err != nil {
		log.Println("UnmatchedSports: Service call failed: ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.JSON(respDto)
	}

	// Return Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.UnmatchedSports = unMatchSports

	return c.JSON(respDto)
}

func AllSports(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := responsedto.AllSportResp{}
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = GENERAL_ERROR_DESC

	// Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.AllSportReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("AllSports: Body Parsing failed: ", err.Error())
		log.Println("AllSports: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("AllSports: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("AllSports: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// Validate Request
	if reqDto.SportId == "" {
		log.Println("AllSports: Invalid Request - ", reqDto)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Call Service
	sports, err := betfair.AllSportsBySportId(reqDto.SportId)
	if err != nil {
		log.Println("AllSports: Service call failed: ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.JSON(respDto)
	}

	// Return Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Sports = sports

	return c.JSON(respDto)
}

func UpdateSportCard(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := dto.CommonPortalRespDto{}
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = GENERAL_ERROR_DESC

	// Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.UpdateSportCardReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UpdateSportCard: Body Parsing failed: ", err.Error())
		log.Println("UpdateSportCard: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("UpdateSportCard: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("UpdateSportCard: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// Validate Request
	if reqDto.EventID == "" {
		log.Println("UpdateSportCard: Invalid Request - ", reqDto)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Call Service
	updateCardResp, err := betfair.UpdateCard(c.Body())
	if err != nil {
		log.Println("UpdateSportCard: Service call failed: ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.JSON(respDto)
	}
	if updateCardResp.Status != 200 {
		respDto.Status = ERROR_STATUS
		respDto.ErrorDescription = GENERAL_ERROR_DESC
		return c.JSON(respDto)
	}

	// Return Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = updateCardResp.Message

	return c.JSON(respDto)
}

func GetSportsRadatCards(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := responsedto.AllSportResp{}
	respDto.Status = ERROR_STATUS
	respDto.ErrorDescription = GENERAL_ERROR_DESC

	// Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.AllSportReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetSportsRadatCards: Body Parsing failed: ", err.Error())
		log.Println("GetSportsRadatCards: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("GetSportsRadatCards: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetSportsRadatCards: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// Validate Request
	if reqDto.SportId == "" {
		log.Println("GetSportsRadatCards: Invalid Request - ", reqDto)
		respDto.ErrorDescription = INVALID_REQ_ERROR_DESC
		return c.JSON(respDto)
	}

	// Call Service
	sportRadarCards, err := betfair.GetSportRadarCards(reqDto.SportId)
	if err != nil {
		log.Println("GetSportsRadatCards: Service call failed: ", err.Error())
		respDto.ErrorDescription = SOMETHING_WENT_WRONG
		return c.JSON(respDto)
	}

	// Return Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.Sports = sportRadarCards
	return c.JSON(respDto)
}

// Swagger hits
// @Summary      To Get New Sports from Provider
// @Description  To add new sports to SAP Platform from a given provider
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        GetNewSports  body      commondto.SyncSportsReqDto  true  "SyncSportsReqDto model is used"
// @Success      200          {object}  portaldto.CommonPortalRespDto
// @Failure      503          {object}  portaldto.CommonPortalRespDto
// @Router       /sapadmin/get-new-sports [post]
func GetNewSports(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("GetNewSports: Req. Body is - ", bodyStr)
	reqDto := commondto.SyncSportsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetNewSports: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}
	switch reqDto.ProviderId {
	case constants.SAP.ProviderType.Dream():
		go dream.GetSports()
	case constants.SAP.ProviderType.BetFair():
		go betfair.GetSports()
	case constants.SAP.ProviderType.SportRadar():
		go sportradar.GetSports()
	default:
		log.Println("GetNewSports: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid ProviderId!!!"
		return c.JSON(respDto)
	}
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.JSON(respDto)
}

// Close Events by EventIds
// @Summary      Close Events by EventIds
// @Description  Close Events by EventIds
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        CloseEvents  body      portaldto.CloseEventsReqDto  true  "CloseEventsReqDto model is used"
// @Success      200           {object}  portaldto.CommonPortalRespDto
// @Failure      503           {object}  portaldto.CommonPortalRespDto
// @Router       /portal/close-events [post]
func CloseEvents(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("CloseEvent: Req. Body is - ", bodyStr)
	reqDto := portaldto.CloseEventsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("CloseEvent: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	// Update Events in DB
	err := database.UpdateEventStatus(reqDto.EventIds, "CLOSED")
	if err != nil {
		log.Println("CloseEvent: Error in getting events from DB")
		respDto.ErrorDescription = "Error in getting events from DB"
		return c.JSON(respDto)
	}

	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.JSON(respDto)
}

func DeleteBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Create Default Response Object
	respDto := portaldto.CommonPortalRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"

	// 1. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("DeleteBets: Req. Body is - ", bodyStr)
	reqDto := portaldto.DeleteBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("DeleteBets: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.JSON(respDto)
	}

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("DeleteBets: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("DeleteBets: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// Update Events in DB
	count, err := database.UpdateBetsStatus(reqDto.BetIds, "DELETED")
	if err != nil {
		log.Println("DeleteBets: Error in getting events from DB")
		respDto.ErrorDescription = "Error in getting events from DB"
		return c.JSON(respDto)
	}
	log.Println("DeleteBets: Set Status to DELETED for ", count, " Bets")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.JSON(respDto)
}

// Get Open Bets For Operators
// @Summary      Get Open Bets For Operators
// @Description  Get Open Bets For Operators
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        GetOpenBetsForOP  body      portaldto.OpenBetsReqDto  true  "OpenBetsReqDto model is used"
// @Success      200               {object}  portaldto.OpenBetsRespDto
// @Failure      503               {object}  portaldto.OpenBetsRespDto
// @Router       /opadmin/get-open-bets [post]
func GetOpenBetsForOP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.OpenBetsRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("GetOpenBetsForOP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetOpenBetsForOP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := portaldto.OpenBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetOpenBetsForOP: Body Parsing failed")
		log.Println("GetOpenBetsForOP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetOpenBets(Tknmeta.OperatorId, Tknmeta.Role, reqDto)
	if err != nil {
		log.Println("GetOpenBetsForOP: Error in getting events from DB")
		respDto.ErrorDescription = "Error in getting events from DB"
		return c.JSON(respDto)
	}
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	for _, bet := range bets {
		openBet := portaldto.OpenBetDto{}
		openBet.BetId = bet.BetId
		openBet.BetType = bet.BetDetails.BetType
		if bet.BetReq.OddsMatched == 0 {
			openBet.OddValue = bet.BetDetails.OddValue
		} else {
			openBet.OddValue = bet.BetReq.OddsMatched
		}
		openBet.StakeAmount = bet.BetDetails.StakeAmount
		openBet.RunnerName = bet.BetDetails.RunnerName
		openBet.RunnerId = bet.BetDetails.RunnerId
		openBet.MarketType = bet.BetDetails.MarketType
		openBet.MarketName = bet.BetDetails.MarketName
		openBet.MarketId = bet.MarketId
		openBet.EventId = bet.EventId
		openBet.SportId = bet.SportId
		openBet.SessionOutcome = bet.BetDetails.SessionOutcome
		openBet.IsUnmatched = bet.BetDetails.IsUnmatched
		openBet.UserId = bet.UserId
		openBet.BetTime = bet.BetReq.ReqTime

		respDto.OpenBets = append(respDto.OpenBets, openBet)
	}

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Open Bets For SAP
// @Summary      Get Open Bets For SAP
// @Description  Get Open Bets For SAP
// @Tags         Portal-PlatformAdmin
// @Accept       json
// @Produce      json
// @Param        GetOpenBetsForSAP  body      portaldto.OpenBetsReqDto  true  "OpenBetsReqDto model is used"
// @Success      200                {object}  portaldto.OpenBetsRespDto
// @Failure      503                {object}  portaldto.OpenBetsRespDto
// @Router       /sapadmin/get-open-bets [post]
func GetOpenBetsForSAP(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.OpenBetsRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// Validate Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		log.Println("GetOpenBetsForSAP: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetOpenBetsForSAP: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.JSON(respDto)
	}

	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := portaldto.OpenBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetOpenBetsForSAP: Body Parsing failed")
		log.Println("GetOpenBetsForSAP: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetOpenBets(Tknmeta.OperatorId, Tknmeta.Role, reqDto)
	if err != nil {
		log.Println("GetOpenBetsForSAP: Error in getting events from DB")
		respDto.ErrorDescription = "Error in getting events from DB"
		return c.JSON(respDto)
	}
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	for _, bet := range bets {
		openBet := portaldto.OpenBetDto{}
		openBet.BetId = bet.BetId
		openBet.BetType = bet.BetDetails.BetType
		if bet.BetReq.OddsMatched == 0 {
			openBet.OddValue = bet.BetDetails.OddValue
		} else {
			openBet.OddValue = bet.BetReq.OddsMatched
		}
		openBet.StakeAmount = bet.BetDetails.StakeAmount
		openBet.RunnerName = bet.BetDetails.RunnerName
		openBet.RunnerId = bet.BetDetails.RunnerId
		openBet.MarketType = bet.BetDetails.MarketType
		openBet.MarketName = bet.BetDetails.MarketName
		openBet.MarketId = bet.MarketId
		openBet.EventId = bet.EventId
		openBet.SportId = bet.SportId
		openBet.SessionOutcome = bet.BetDetails.SessionOutcome
		openBet.IsUnmatched = bet.BetDetails.IsUnmatched
		openBet.UserId = bet.UserId
		openBet.BetTime = bet.BetReq.ReqTime
		openBet.OperatorId = bet.OperatorId
		respDto.OpenBets = append(respDto.OpenBets, openBet)
	}

	return c.Status(fiber.StatusOK).JSON(respDto)
}
