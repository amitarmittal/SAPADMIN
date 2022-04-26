package operatorsvc

import (
	"Sp/cache"
	"Sp/common/function"
	"Sp/constants"
	"Sp/database"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	portaldto "Sp/dto/portal"
	dreamdto "Sp/dto/providers/dream"
	"Sp/dto/reports"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	sessDto "Sp/dto/session"
	"Sp/dto/sports"
	"Sp/handler"
	coresvc "Sp/handler/core"
	"Sp/operator"
	"Sp/providers"
	"Sp/providers/betfair"
	"Sp/providers/dream"
	"Sp/providers/sportradar"
	keyutils "Sp/utilities"
	utils "Sp/utilities"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	TestOperator string = "TestOperator"
	TestProvider string = "Dream"
	TestToken    string = "f23004ee-2bf4-49b9-a98b-f0189520b795"
	// BaseUrl      string = "https://stage.mysportsfeed.io"
	BaseUrl  string = os.Getenv("IFRAME_URL")
	LobbyUrl string = "%s/auth?token=%s&operatorId=%s&partnerId=%s&providerId=%s"

	ERROR_STATUS = "RS_ERROR"
	OK_STATUS    = "RS_OK"
)

// Login API
// @Summary      User Login
// @Description  User Login to get Session Token
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature  header    string          true  "Hash Signature"
// @Param        Login      body      dto.AuthReqDto  true  "AuthReqDto model is used"
// @Success      200        {object}  dto.AuthRespDto
// @Failure      503        {object}  dto.AuthRespDto
// @Router       /feed/user-login [post]
func Authentication(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := new(dto.AuthRespDto)
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("Authentication: Auth Req. Body is - ", bodyStr)
	reqDto := dto.AuthReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("Authentication: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	// 3.1 OperatorId - Not Found, Not Active
	// 3.2 UserId - Empty
	// 3.3 ClientIp - Empty

	//userId := olDTO.UserId
	//providerName := olDTO.ProviderName
	//clientIp := olDTO.ClientIp
	//platformId := olDTO.PlatformId
	// Get Operator Details
	operatorId := reqDto.OperatorId
	logkey := "Authentication: " + operatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println(logkey+" Failed to get Operator Details : ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println(logkey+" Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println(logkey + " PartnerId is missing!")
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
				log.Println(logkey+" Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println(logkey+" Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Verify Signature
	if operatorDto.Signature {
		signature := c.Request().Header.Peek("Signature")
		log.Println(logkey+" Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println(logkey+" Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println(logkey + " Signature is - " + string(signature))
			log.Println(logkey + " Signature verification failed!!!")
			log.Println(logkey + " pubkey is - " + operatorDto.Keys.OperatorKey)
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}
	// 4. Get B2B User object & Check Staus. If new User, create B2B User
	userKey := reqDto.OperatorId + "-" + reqDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println(logkey+" User NOT FOUND - ", err.Error())
		// 4.1. Create New User
		b2bUser = models.B2BUserDto{}
		b2bUser.UserKey = userKey
		b2bUser.OperatorId = reqDto.OperatorId
		b2bUser.UserId = reqDto.UserId
		b2bUser.UserName = reqDto.Username
		b2bUser.Balance = 0
		b2bUser.Status = "ACTIVE"
		err = database.InsertB2BUser(b2bUser)
		if err != nil {
			// TODO: Retry mechanism
			log.Println(logkey+" User creation failed with error - ", err.Error())
			// Not sending error as it is fine for users from seamless wallet operator
			// if user is from transfer wallet operator, handle user creation at credit method
		}
	}
	// 5. Get Token Details by operatorId+userId - Get From Cache->DB
	sessionDto := sessDto.B2BSessionDto{}
	if operatorDto.NewSession == true || operatorDto.OperatorId == "ST8" {
		sessionDto = IssueB2BSessionToken(reqDto, operatorDto)
	} else {
		sessionDto, err = cache.GetSessionDetailsByUserKey(reqDto.OperatorId, reqDto.UserId)
		if err != nil {
			log.Println(logkey+" Session NOT FOUND - ", err.Error())
			// 4.1. Issue New Token
			sessionDto = IssueB2BSessionToken(reqDto, operatorDto)
		} else {
			// 4.2. If Token exists, Check for validity
			if cache.IsSessionValid(sessionDto) {
				// 4.2.1. Token Valid, check for Grace period and extend the validity
				sessionDto = cache.ExtendValidity(sessionDto)
			} else {
				// 4.2.2. Token Expired, Issue New Token
				sessionDto = IssueB2BSessionToken(reqDto, operatorDto)
			}
		}
	}

	// Get Active Provider list by Operator
	providers := GetProvidersList(reqDto.OperatorId, reqDto.PartnerId)
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.UserId = reqDto.UserId
	respDto.Token = sessionDto.Token
	if len(providers) > 0 {
		respDto.Url = fmt.Sprintf(LobbyUrl, BaseUrl, respDto.Token, reqDto.OperatorId, reqDto.PartnerId, providers[0].ProviderId)
		for _, provider := range providers {
			providerDto := dto.ProviderDto{}
			providerDto.ProviderId = provider.ProviderId
			providerDto.ProviderName = provider.ProviderName
			providerDto.Status = provider.Status
			providerDto.Url = fmt.Sprintf(LobbyUrl, BaseUrl, respDto.Token, reqDto.OperatorId, reqDto.PartnerId, providerDto.ProviderId)
			respDto.Providers = append(respDto.Providers, providerDto)
		}
	}
	respJSON, err := json.Marshal(respDto)
	if err != nil {
		log.Println(logkey+" json.Marshal fialed with error - ", err.Error())
	} else {
		log.Println(logkey+" respJSON is - ", string(respJSON))
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Transfer Wallet APIs

// Get User Balance API
// @Summary      User Balance API
// @Description  To get current user balance
// @Tags         Wallet-Service
// @Accept       json
// @Produce      json
// @Param        Login  body      operatordto.UserBalanceReqDto  true  "UserBalanceReqDto model is used"
// @Success      200    {object}  operatordto.UserBalanceRespDto
// @Failure      503    {object}  operatordto.UserBalanceRespDto
// @Router       /wallet/user-balance [post]
func GetUserBalance(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.UserBalanceRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	// 2. Parse request body to Request Object
	reqDto := operatordto.UserBalanceReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetUserBalance: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check

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
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("GetUserBalance: PartnerId is missing!")
		respDto.ErrorDescription = "PartnerId cannot be NULL!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	found := false
	var rate int32 = 1 // default multiplier is 1
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == partnerId {
			found = true
			rate = partner.Rate
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
	// 5. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("GetUserBalance: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("GetUserBalance: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 6. Get User Balance from db
	userKey := reqDto.OperatorId + "-" + reqDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println("GetUserBalance: User NOT FOUND: ", err.Error())
		respDto.ErrorDescription = "Invalid User!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.Balance = utils.Truncate64(b2bUser.Balance / float64(rate)) // devided by rate (in CUR)
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// User Deposit Funds API
// @Summary      User Deposit Funds API
// @Description  To add funds to user balance
// @Tags         Wallet-Service
// @Accept       json
// @Produce      json
// @Param        Login  body      operatordto.DepositReqDto  true  "DepositReqDto model is used"
// @Success      200    {object}  operatordto.DepositRespDto
// @Failure      503    {object}  operatordto.DepositRespDto
// @Router       /wallet/deposit-funds [post]
func Deposit(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.DepositRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	// 2. Parse request body to Request Object
	reqDto := operatordto.DepositReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("Deposit: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check

	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("Deposit: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("Deposit: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("Deposit: PartnerId is missing!")
		respDto.ErrorDescription = "PartnerId cannot be NULL!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	found := false
	var rate int32 = 1 // default multiplier is 1
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == partnerId {
			found = true
			rate = partner.Rate
			if partner.Status != "ACTIVE" {
				log.Println("Deposit: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("Deposit: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	var creditAmount float64 = reqDto.CreditAmount * float64(rate) // multiply by rate (in PTS)
	// 5. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("Deposit: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("Deposit: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 6. Get User Balance from db
	userKey := reqDto.OperatorId + "-" + reqDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		// 6.1. User not found, create user
		log.Println("Deposit: User NOT FOUND: ", err.Error())
		b2bUser = models.B2BUserDto{}
		b2bUser.UserKey = userKey
		b2bUser.OperatorId = reqDto.OperatorId
		b2bUser.UserId = reqDto.UserId
		b2bUser.UserName = ""
		b2bUser.Balance = creditAmount // in PTS
		b2bUser.Status = "ACTIVE"
		err = database.InsertB2BUser(b2bUser)
		if err != nil {
			// 6.1.1. User creation failed. Send failure response
			log.Println("Deposit: Failed to create user: ", err.Error())
			respDto.ErrorDescription = "Deposit Failed!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		// 6.2 User creation is success, return success message
		respDto.ErrorDescription = ""
		respDto.Status = "RS_OK"
		respDto.Balance = reqDto.CreditAmount // in CUR
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Update user balance
	b2bUser.Balance += creditAmount                            // in PTS
	err = database.UpdateB2BUserBalance(userKey, creditAmount) // in PTS
	if err != nil {
		// 7.1. update failed, send failure response
		log.Println("Deposit: Failed to deposit funds: ", err.Error())
		respDto.ErrorDescription = "Deposit Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Save deposit transaction in User Ledger
	userLedger := models.UserLedgerDto{}
	userLedger.UserKey = reqDto.OperatorId + "-" + reqDto.UserId
	userLedger.OperatorId = reqDto.OperatorId
	userLedger.UserId = reqDto.UserId
	userLedger.Remark = reqDto.Remark
	userLedger.TransactionType = constants.SAP.LedgerTxType.DEPOSIT() // "Deposit-Funds"
	userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	userLedger.ReferenceId = ""
	userLedger.Amount = creditAmount // in PTS
	err = database.InsertLedger(userLedger)
	if err != nil {
		// 8.1. inserting ledger document failed
		log.Println("Deposit: insert ledger failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 9. Send success response
	respDto.Balance = utils.Truncate64(b2bUser.Balance / float64(rate)) // Devide by rate (in CUR)
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// User Withdraw Funds API
// @Summary      User Withdraw Funds API
// @Description  To withdraw funds from user balance
// @Tags         Wallet-Service
// @Accept       json
// @Produce      json
// @Param        Login  body      operatordto.WithdrawReqDto  true  "WithdrawReqDto model is used"
// @Success      200    {object}  operatordto.WithdrawRespDto
// @Failure      503    {object}  operatordto.WithdrawRespDto
// @Router       /wallet/withdraw-funds [post]
func Withdraw(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.WithdrawRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Balance = 0
	// 2. Parse request body to Request Object
	reqDto := operatordto.WithdrawReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("Withdraw: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check

	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("Withdraw: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("Withdraw: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("Withdraw: PartnerId is missing!")
		respDto.ErrorDescription = "PartnerId cannot be NULL!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	found := false
	var rate int32 = 1 // default multiplier is 1
	for _, partner := range operatorDto.Partners {
		if partner.PartnerId == partnerId {
			found = true
			rate = partner.Rate
			if partner.Status != "ACTIVE" {
				log.Println("Withdraw: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("Withdraw: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	var debitAmount float64 = reqDto.DebitAmount * float64(rate) // multiply by rate (in PTS)
	// 5. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("Withdraw: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("Withdraw: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 6. Get User Balance from db
	userKey := reqDto.OperatorId + "-" + reqDto.UserId
	b2bUser, err := database.GetB2BUser(userKey)
	if err != nil {
		// 6.1. User not found, create user
		log.Println("Withdraw: User NOT FOUND: ", err.Error())
		respDto.ErrorDescription = "Invalid User!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Debit amount should not be greater than credit amount
	if debitAmount > b2bUser.Balance {
		// send failure response
		log.Println("Withdraw: debit amount is greater than the balance!")
		log.Println("Withdraw: reqDto.DebitAmount: ", debitAmount)
		log.Println("Withdraw: b2bUser.Balance: ", b2bUser.Balance)
		respDto.ErrorDescription = "Insufficient Funds!"
		respDto.Balance = utils.Truncate64(b2bUser.Balance / float64(rate)) // devide by rate (in CUR)
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Update user balance
	b2bUser.Balance -= debitAmount                               // in PTS
	err = database.UpdateB2BUserBalance(userKey, debitAmount*-1) // in PTS
	if err != nil {
		// 8.1. update failed, send failure response
		log.Println("Withdraw: Failed to update funds: ", err.Error())
		respDto.ErrorDescription = "Withdraw Failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 9. TODO: Add to Funds Transaction collection
	userLedger := models.UserLedgerDto{}
	userLedger.UserKey = reqDto.OperatorId + "-" + reqDto.UserId
	userLedger.OperatorId = reqDto.OperatorId
	userLedger.UserId = reqDto.UserId
	userLedger.Remark = reqDto.Remark
	userLedger.TransactionType = constants.SAP.LedgerTxType.WITHDRAW() // "Withdraw-Funds"
	userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	userLedger.ReferenceId = ""
	userLedger.Amount = debitAmount * -1 // in PTS
	err = database.InsertLedger(userLedger)
	if err != nil {
		// 9.1. inserting ledger document failed
		log.Println("Withdraw: insert ledger failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 10. Send success response
	respDto.Balance = utils.Truncate64(b2bUser.Balance / float64(rate)) // devide by rate (in CUR)
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Bet History APIs
func GetProviders(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.GetProvidersRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Providers = []dto.ProviderInfo{}
	// 2. Parse request body to Request Object
	reqDto := operatordto.GetProvidersReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetProviders: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check

	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("Withdraw: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("Withdraw: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("Withdraw: PartnerId is missing!")
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
				log.Println("ListOpProviders: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("ListOpProviders: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("Withdraw: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("Withdraw: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/

	// 6. Get Active Providers
	// TODO: Read from Cache/DB
	providerInfos, err := handler.GetActiveProviders(operatorId, partnerId)
	if err != nil {
		// 6.1. Return Error
		log.Println("ListOpProviders: Failed to get Active Provider with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// BetFairPlus Mode Check
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

	respDto.Providers = append(respDto.Providers, providerInfos...)
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Bets by OperatorId
func GetBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := operatordto.BetsHistoryRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.Bets = []operatordto.BetHistory{}
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 2. Parse request body to Request Object
	reqStr := string(c.Body())
	reqDto := operatordto.BetsHistoryReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetBets: Body Parsing failed")
		log.Println("GetBets: Request Body is - ", reqStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	err := CheckGetBets(reqDto)
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
	if operatorDto.Status != "ACTIVE" {
		log.Println("GetBets: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("GetUserBalance: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("GetUserBalance: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 6. Get Bets By Req
	bets, err := database.GetBetsByOperator(reqDto)
	if err != nil {
		log.Println("GetBets: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Bets not found, please try again!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	totalRecords := len(bets)
	//log.Println("GetBets: Bets Count: ", totalRecords)
	respDto.Page = 1 // Default is page 1
	if reqDto.Page != 0 {
		respDto.Page = reqDto.Page
	}
	respDto.PageSize = 50 // default page size
	if reqDto.PageSize != 0 && reqDto.PageSize < 50 {
		respDto.PageSize = reqDto.PageSize
	}

	betCount := 0
	startIndex := (respDto.Page - 1) * respDto.PageSize
	endIndex := startIndex + respDto.PageSize
	if endIndex > totalRecords {
		endIndex = totalRecords
	}
	for itr := startIndex; itr < endIndex; itr++ {
		respDto.Bets = append(respDto.Bets, GetBetHistory2(bets[itr]))
		betCount++
	}
	respDto.PageSize = betCount
	respDto.TotalRecords = totalRecords
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	// 7. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// ValidateOdds is a function to Validate Odds
// @Summary      Validate Odds
// @Description  Validate Odds endpoint
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature     header    string                     true  "Hash Signature"
// @Param        ValidateOdds  body      requestdto.PlaceBetReqDto  true  "PlaceBetReqDto model is used"
// @Success      200           {object}  responsedto.ValidateOddsRespDto{}
// @Failure      503           {object}  responsedto.ValidateOddsRespDto{}
// @Router       /feed/validate-odds [post]
func ValidateOdds(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.ValidateOddsRespDto{}
	respDto.ErrorDescription = "Generic Error!"
	respDto.Status = "RS_ERROR"
	respDto.IsValid = false
	respDto.EventStatus = ""
	respDto.EventDate = 0
	respDto.OddValues = []float64{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("ValidateOdds: Req. Body is - ", bodyStr)
	reqDto := requestdto.PlaceBetReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ValidateOdds: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Validate Token
	/*
		sessionDto, err := function.GetSession(reqDto.Token)
		if err != nil {
			// 3.1. Return Error
			log.Println("ValidateOdds: GetSession failed with error - ", err.Error())
			respDto.ErrorDescription = "Session Expired"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	*/
	//log.Println("ValidateOdds: User is - ", sessionDto.UserId)
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("ValidateOdds: GetOperatorDetails failed with error - ", err.Error())
		//log.Println("ValidateOdds: User is - ", sessionDto.UserId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("ValidateOdds: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if strings.ToLower(operatorDto.WalletType) != "feed" {
		log.Println("ValidateOdds: Invalid wallet type - ", operatorDto.WalletType)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("ValidateOdds: PartnerId is missing!")
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
				log.Println("ListOpProviders: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("ListOpProviders: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
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
	// 5.0. IsEventActive?
	if false == providers.IsProviderActive(reqDto.OperatorId, partnerId, reqDto.ProviderId) {
		// 4.2. Return Error
		log.Println("ValidateOdds: IsProviderActive returned false - ", reqDto.SportId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if false == providers.IsSportActive(reqDto.OperatorId, partnerId, reqDto.ProviderId, reqDto.SportId) {
		// 4.2. Return Error
		log.Println("ValidateOdds: IsSportActive returned false - ", reqDto.SportId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// if false == providers.IsCompetitionActive(operatorId, providerId, sportId, competitionId) {
	// 	// 4.2. Return Error
	// 	log.Println("ListOpMarkets: IsCompetitionActive returned false for eventId - ", eventId)
	// 	respDto.ErrorDescription = "Unauthorized access, please contact support!"
	// 	return c.Status(fiber.StatusOK).JSON(respDto)
	// }
	if false == providers.IsEventActive(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId) {
		// 4.2. Return Error
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 11. Market Status
	if reqDto.MarketId == "" {
		log.Println("ValidateOdds: MarketId is empty in request!")
		respDto.ErrorDescription = "Invalid Market!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	market, err := cache.GetMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	if err == nil {
		if market.Status != "ACTIVE" {
			log.Println("PlaceBet: market.Status is not ACTIVE - ", market.Status)
			respDto.ErrorDescription = "Betting is not allowed on this market!!!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		// 12. MarketStatus ProviderStatus & OperatorStatus
		marketStatus, err := cache.GetMarketStatus(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
		if err == nil {
			if marketStatus.ProviderStatus != "ACTIVE" {
				log.Println("PlaceBet: marketStatus.ProviderStatus is not ACTIVE - ", marketStatus.ProviderStatus)
				respDto.ErrorDescription = "Betting is not allowed on this market!!!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			if marketStatus.OperatorStatus != "ACTIVE" {
				log.Println("PlaceBet: marketStatus.OperatorStatus is not ACTIVE - ", marketStatus.OperatorStatus)
				respDto.ErrorDescription = "Betting is not allowed on this market!!!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
		}
	} else {
		if err.Error() == "Market NOT FOUND!" {
			// insert market
			err = handler.InsertMarket(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId, reqDto.MarketType)
			if err != nil {
				log.Println("PlaceBet: InsertMarket failed with error - ", err.Error())
			}
		}
	}
	// 9. Provider specific logic
	if operatorDto.BetFairPlus == true {
		switch reqDto.MarketType {
		case constants.SAP.MarketType.BOOKMAKER(), constants.SAP.MarketType.FANCY():
			reqDto.ProviderId = constants.SAP.ProviderType.Dream()
		default:
		}
	}
	resp := dreamdto.ValidateOddsRespDto{}
	switch reqDto.ProviderId {
	case providers.DREAM_SPORT:
		// 9.1. Dream - Bet Placement
		resp, err = dream.ValidateOdds(reqDto)
		if err != nil {
			// 4.1. Return Error
			log.Println("ValidateOdds: Dream ValidateOdds failed with error - ", err.Error())
			respDto.ErrorDescription = "Failed to validate odds!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case providers.BETFAIR:
		// 9.1. BetFair - Bet Placement
		resp, err = betfair.ValidateOdds(reqDto)
		if err != nil {
			// 4.1. Return Error
			log.Println("ValidateOdds: BetFair ValidateOdds failed with error - ", err.Error())
			respDto.ErrorDescription = "Failed to validate odds!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	case providers.SPORT_RADAR:
		// 9.1. SportRadar - Bet Placement
		resp, err = sportradar.ValidateOdds(reqDto)
		if err != nil {
			// 4.1. Return Error
			log.Println("ValidateOdds: BetFair ValidateOdds failed with error - ", err.Error())
			respDto.ErrorDescription = "Failed to validate odds!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	default:
		log.Println("ValidateOdds: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 10. Send SUCCESS response
	respDto.IsValid = resp.IsValid
	respDto.EventStatus = resp.Status
	respDto.EventDate = resp.OpenDate
	respDto.OddValues = []float64{}
	if len(resp.OddValues) > 0 {
		respDto.OddValues = append(respDto.OddValues, resp.OddValues...)
	}
	respDto.MatchedOddValue = resp.MatchedOddValue
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	respJson, err := json.Marshal(respDto)
	if err != nil {
		log.Println("ValidateOdds: json.Marshal failed with error - ", err.Error())
	} else {
		log.Println("ValidateOdds: Resp. Body is - ", string(respJson))
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetIsMatchedStatus is a function to get matched bets status
// @Summary      Get isMatched Bet status
// @Description  GetIsMatchedStatus takes a array of BetIds and returns weather the bet is matched or not
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature           header    string                           true  "Hash Signature"
// @Param        GetIsMatchedStatus  body      requestdto.GetMatchedBetsReqDto  true  "GetMatchedBetsReqDto model is used"
// @Success      200                 {object}  responsedto.GetMatchedBetsRespDto{}
// @Failure      503                 {object}  responsedto.GetMatchedBetsRespDto{}
// @Router       /feed/get-is-matched-status [post]
func GetIsMatchedStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// Get Default Response
	respDto := responsedto.GetMatchedBetsRespDto{}
	respDto.ErrorDescription = "Generic Error!"
	respDto.Status = "RS_ERROR"

	// Get Request
	reqDto := requestdto.GetMatchedBetsReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("GetMatchedBets: Error in body parser - ", err.Error())
		respDto.ErrorDescription = "Invalid Request!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Get is Bet details for BetIds
	bets, err := database.GetBets(reqDto.BetIds)
	if err != nil {
		log.Println("GetMatchedBets: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	isMatchedStatus := []responsedto.IsMatchedStatus{}
	for _, bet := range bets {
		status := bet.Status
		oddValue := bet.BetDetails.OddValue
		// if status == constants.SAP.BetStatus.INPROCESS() {
		// 	status = constants.SAP.BetStatus.UNMATCHED()
		// }
		if status == constants.SAP.BetStatus.UNMATCHED() && bet.BetReq.SizeRemaining != 0 && bet.BetReq.SizeMatched != 0 {
			status = "PARTIAL_MATCHED"
			oddValue = bet.BetReq.OddsMatched
			log.Println("GetMatchedBets: PartialMatched Bet Found for BetFair betId - ", bet.BetReq.BetId)
			log.Println("GetMatchedBets: PartialMatched Bet SizeMatched - ", bet.BetReq.SizeMatched)
			log.Println("GetMatchedBets: PartialMatched Bet SizeRemaining - ", bet.BetReq.SizeRemaining)
		}
		if status == constants.SAP.BetStatus.OPEN() {
			oddValue = bet.BetReq.OddsMatched
		}
		// matched amount
		sizeMatched := bet.BetReq.SizeMatched * float64(betfair.BetFairRate)
		sizeMatched = (sizeMatched * 100) / (100 - bet.BetReq.PlatformHold)
		sizeMatched = (sizeMatched * 100) / (100 - bet.BetReq.OperatorHold)
		if bet.BetReq.Rate != 0 {
			sizeMatched = utils.Truncate64(sizeMatched / float64(bet.BetReq.Rate))
		}
		// remaining amount
		sizeRemaining := bet.BetReq.SizeRemaining * float64(betfair.BetFairRate)
		sizeRemaining = (sizeRemaining * 100) / (100 - bet.BetReq.PlatformHold)
		sizeRemaining = (sizeRemaining * 100) / (100 - bet.BetReq.OperatorHold)
		if bet.BetReq.Rate != 0 {
			sizeRemaining = utils.Truncate64(sizeRemaining / float64(bet.BetReq.Rate))
		}
		isMatchedStatus = append(isMatchedStatus, responsedto.IsMatchedStatus{
			BetId:         bet.BetId,
			IsMatched:     bet.BetDetails.IsUnmatched,
			BetStatus:     status,
			OddValue:      oddValue,
			SizeMatched:   sizeMatched,
			SizeRemaining: sizeRemaining,
		})
	}
	// 4. Send SUCCESS response
	respDto.IsMatchedStatus = isMatchedStatus
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// PlaceBet is a function to Palce a Bet
// @Summary      Place Bet
// @Description  Bet Placement Async Endpoint
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature  header    string                     true  "Hash Signature"
// @Param        PlaceBet   body      requestdto.PlaceBetReqDto  true  "PlaceBetReqDto model is used"
// @Success      200        {object}  responsedto.PlaceBetRespDto{}
// @Failure      503        {object}  responsedto.PlaceBetRespDto{}
// @Router       /feed/place-bet [post]
func PlaceBet(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.PlaceBetRespDto{}
	respDto.ErrorDescription = "Generic Error!"
	respDto.Status = "RS_ERROR"
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
	log.Println("PlaceBet:reqDto json is - ", string(reqJson))
	// 3.0 Place Bet
	betId, err := handler.PlaceBet(reqDto)
	if err != nil {
		log.Println("PlaceBet: handler.PlaceBet failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Send SUCCESS response
	respDto.BetId = betId
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// CancelBets is a function to cancel an unmathed Bets
// @Summary      Cancel Bets
// @Description  Bets Cancel for multiple bets
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature  header    string                                true  "Hash Signature"
// @Param        PlaceBet   body      requestdto.CancelBetReqDto  true  "CancelBetReqDto model is used"
// @Success      200        {object}  responsedto.CancelBetRespDto{}
// @Failure      503        {object}  responsedto.CancelBetRespDto{}
// @Router       /feed/cancel-bets [post]
func CancelBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.CancelBetRespDto{}
	respDto.ErrorDescription = "Genric Error"
	respDto.CancelBetsResp = []responsedto.CancelBetResp{}
	respDto.Balance = 0
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.CancelBetReqDto{}
	log.Println("CancelBet: Req. Body is - ", bodyStr)
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("CancelBet: Body Parsing failed")
		log.Println("CancelBet: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Process Cancel Bets.
	respDto.Balance, respDto.CancelBetsResp, err = handler.CancelBet(reqDto)
	if err != nil {
		log.Println("CancelBet: handler.CancelBet failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 8. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	jsonResp, err := json.Marshal(respDto)
	if err != nil {
		log.Println("CancelBet: json.Marshal failed with error - ", err.Error())
	}
	log.Println("CancelBet: respDto - ", string(jsonResp))
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// SportRadarCancelBet is a function to cancel a sportradar bet
// @Summary      Cancel SportRadar Bet
// @Description  Bets Cancel for single sportradar bets
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature  header    string                      true  "Hash Signature"
// @Param        PlaceBet   body      requestdto.SportRadarCancelBetReqDto  true  "SportRadarCancelBetReqDto model is used"
// @Success      200        {object}  responsedto.SportRadarCancelBetRespDto{}
// @Failure      503        {object}  responsedto.SportRadarCancelBetRespDto{}
// @Router       /feed/sportradar-cancel-bet [post]
func SportRadarCancelBet(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.SportRadarCancelBetRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Genric Error"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.SportRadarCancelBetReqDto{}
	log.Println("SportRadarCancelBet: Req. Body is - ", bodyStr)
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("SportRadarCancelBet: Body Parsing failed")
		log.Println("SportRadarCancelBet: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// Get bet from DB
	betId := strings.Split(reqDto.BetId, "34948-")[1]
	bet, err := database.GetBetDetails(betId)
	if err != nil {
		log.Println("SportRadarCancelBet: database.GetBetDetails failed with error for betid - ", reqDto.BetId, err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if bet.Status == constants.SAP.BetStatus.CANCELLED() {
		respDto.ErrorDescription = "Bet is already in cancelled state!!!"
		respDto.Status = "RS_OK"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if bet.Status != constants.SAP.BetStatus.OPEN() {
		log.Println("SportRadarCancelBet: bet is not in OPEN state for betid - ", reqDto.BetId, bet.Status)
		respDto.ErrorDescription = "Bet is not in OPEN state!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// cancel order
	respObj, err := sportradar.CancelOrder(reqDto.BetId)
	if err != nil {
		log.Println("SportRadarCancelBet: sportradar.CancelOrder failed with error for betid - ", reqDto.BetId, err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if respObj.Status != "RS_OK" {
		log.Println("SportRadarCancelBet: sportradar.CancelOrder failed with status for betid - ", reqDto.BetId, respObj.Status, respObj.ErrorDescription)
		respDto.ErrorDescription = respObj.ErrorDescription
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Do update bet
	// 5. Add RollbackReq to bets
	RollbackBets := []sports.BetDto{}
	// 2.2.1. update bet with result
	rollbackReq := providers.ComputeRollback(bet, constants.SAP.BetStatus.CANCELLED())
	bet.RollbackReqs = append(bet.RollbackReqs, rollbackReq)
	bet.NetAmount += rollbackReq.RollbackAmount
	bet.Status = constants.SAP.BetStatus.CANCELLED()
	// setting updatedAt
	bet.UpdatedAt = rollbackReq.ReqTime
	// 2.2.3. append bet to list
	RollbackBets = append(RollbackBets, bet)
	log.Println("SportRadarCancelBet: RollbackBets count is - ", len(RollbackBets))
	operator.CommonRollbackRoutine(constants.SAP.BetStatus.CANCELLED(), RollbackBets)
	log.Println("SportRadarCancelBet: operator.CommonRollbackRoutine END for bet - ", reqDto.BetId)
	// 8. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// OpenBets is a function to get list of Opend Bets per event
// @Summary      Open Bets
// @Description  To get list of Open Bets in an event
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature  header    string                     true  "Hash Signature"
// @Param        OpenBets   body      requestdto.OpenBetsReqDto  true  "OpenBetsReqDto model is used"
// @Success      200        {object}  responsedto.OpenBetsRespDto{}
// @Failure      503        {object}  responsedto.OpenBetsRespDto{}
// @Router       /feed/Open-bets [post]
func OpenBets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.OpenBetsRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"
	respDto.OpenBets = []responsedto.OpenBetDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.OpenBetsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("OpenBets: Body Parsing failed with error - ", err.Error())
		log.Println("OpenBets: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	if reqDto.Token == "" {
		log.Println("OpenBets: Token is missing for req - ", bodyStr)
		respDto.ErrorDescription = "Invalid Token!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.ProviderId == "" {
		log.Println("OpenBets: ProviderId is missing for req - ", bodyStr)
		respDto.ErrorDescription = "Invalid Provider!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		log.Println("OpenBets: SportId is missing for req - ", bodyStr)
		respDto.ErrorDescription = "Invalid Sport!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.EventId == "" {
		log.Println("OpenBets: EventId is missing for req - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Validate Token
	sessionDto, err := function.GetSession(reqDto.Token)
	if err != nil {
		// 3.1. Return Error
		log.Println("OpenBets: Session Validation failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	//log.Println("OpenBets: User is - ", sessionDto.UserId)
	// 4. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("OpenBets: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 4.2. Return Error
		log.Println("OpenBets: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if strings.ToLower(operatorDto.WalletType) != "feed" {
		log.Println("OpenBets: Invalid wallet type - ", operatorDto.WalletType)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("OpenBets: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("OpenBets: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 7. Get Open Bets
	eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
	openBets, _, err := database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
	if err != nil {
		log.Println("OpenBets: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.OpenBets = handler.GetOpenBetsDto(openBets, reqDto.SportId)
	if operatorDto.BetFairPlus == true && reqDto.ProviderId == constants.SAP.ProviderType.BetFair() {
		log.Println("OpenBets: BetFairPlus Mode (MPC1) & BetFair!!!")
		eventKey = constants.SAP.ProviderType.Dream() + "-" + reqDto.SportId + "-" + reqDto.EventId
		openBets, _, err = database.GetOpenBetsByUser(eventKey, reqDto.OperatorId, sessionDto.UserId)
		if err != nil {
			log.Println("OpenBets: Failed with error - ", err.Error())
			//respDto.ErrorDescription = err.Error()
			//return c.Status(fiber.StatusOK).JSON(respDto)
		} else {
			log.Println("OpenBets: OperatorId - EventKey - betsCount - ", reqDto.OperatorId, eventKey, len(openBets))
			respDto.OpenBets = append(respDto.OpenBets, handler.GetOpenBetsDto(openBets, reqDto.SportId)...)
		}
	}
	// 8. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      Test Signature
// @Description  To test signature functionality
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature      header    string                      true  "Hash Signature"
// @Param        TestSignature  body      requestdto.ProvidersReqDto  true  "ProvidersReqDto model is used"
// @Success      200            {object}  responsedto.DefaultRespDto
// @Failure      503            {object}  responsedto.DefaultRespDto
// @Router       /feed/test-signature [post]
func TestSignarute(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.ProvidersReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("TestSignarute: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Request Check
	operatorId := reqDto.OperatorId
	if operatorId == "" {
		log.Println("TestSignarute: OperatorId is missing!")
		respDto.ErrorDescription = "OperatorId cannot be NULL!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Get Operator Details
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("TestSignarute: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Invalid Operator"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Check Operator Status
	if operatorDto.Status != "ACTIVE" {
		log.Println("TestSignarute: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Verify Signature
	signature := c.Request().Header.Peek("Signature")
	//log.Println("Signature", string(signature))
	pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
	if err != nil {
		log.Println("TestSignarute: Parsing public key failed: ", err.Error())
		respDto.ErrorDescription = "Parsing public key failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
	if !signValid {
		log.Println("TestSignarute: Signature verification failed : ")
		//respDto.ErrorDescription = "Signature verification failed!"
		//return c.Status(fiber.StatusOK).JSON(respDto)
	}
	signValid = keyutils.VerifySignature2(string(signature), string(c.Body()), *pubKey)
	if !signValid {
		log.Println("TestSignarute: Signature verification failed : ")
		respDto.ErrorDescription = "Signature verification failed!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// BetsStatus is a function to get Bets Status
// @Summary      Bets Status
// @Description  Bets Status for multiple bets
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature      header    string                          true  "Hash Signature"
// @Param        BetsStatus  body      requestdto.BetsStatusReqDto  true  "BetsStatusReqDto model is used"
// @Success      200         {object}  responsedto.BetsStatusRespDto{}
// @Failure      503         {object}  responsedto.BetsStatusRespDto{}
// @Router       /feed/bets-status [post]
func BetsStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.BetsStatusRespDto{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.BetsStatusReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("BetsStatus: Body Parsing failed")
		log.Println("BetsStatus: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 3.1. Return Error
		log.Println("BetsStatus: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 3.2. Return Error
		log.Println("BetsStatus: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if strings.ToLower(operatorDto.WalletType) != "feed" {
		log.Println("BetsStatus: Invalid wallet type - ", operatorDto.WalletType)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("CancelBets: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("CancelBets: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 5. Get a betDto
	betDtos, err := database.GetBets(reqDto.BetIds)
	// 6. Prepare respDto.BetsStaus
	respDto.BetsStatus = coresvc.GetBetsStatusDto(betDtos)
	// 7. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// BetsResult is a function to get Bets Status
// @Summary      Bets Result
// @Description  Bets Result for multiple bets
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature   header    string                       true  "Hash Signature"
// @Param        BetsResult  body      requestdto.BetsResultReqDto  true  "BetsResultReqDto model is used"
// @Success      200         {object}  responsedto.BetsResultRespDto{}
// @Failure      503         {object}  responsedto.BetsResultRespDto{}
// @Router       /feed/bets-result [post]
func BetsResult(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.BetsResultRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"
	respDto.BetsResult = []responsedto.BetResult{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.BetsResultReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("BetsResult: Body Parsing failed")
		log.Println("BetsResult: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 3.1. Return Error
		log.Println("BetsResult: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 3.2. Return Error
		log.Println("BetsResult: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("CancelBets: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("CancelBets: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 5. Get a betDto
	betDtos, err := database.GetBets(reqDto.BetIds)
	if err != nil {
		log.Println("BetsResult: database.GetBets failed with error - ", err.Error())
	}
	log.Println("BetsResult: database.GetBets returned bets count - ", len(betDtos))
	// 6. Prepare respDto.BetsResult
	for _, bet := range betDtos {
		respDto.BetsResult = append(respDto.BetsResult, responsedto.GetBetResult(bet))
	}
	// 7. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// MarketsResult is a function to get Bets Status
// @Summary      Markets Result
// @Description  Markets Result for multiple markets
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature   header    string                       true  "Hash Signature"
// @Param        MarketsResult  body      requestdto.MarketsResultReqDto  true  "MarketsResultReqDto model is used"
// @Success      200            {object}  responsedto.MarketsResultRespDto{}
// @Failure      503            {object}  responsedto.MarketsResultRespDto{}
// @Router       /feed/markets-result [post]
func MarketsResult(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.MarketsResultRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error"
	respDto.MarketResults = []responsedto.MarketResult{}
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	reqDto := requestdto.MarketsResultReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("MarketsResult: Body Parsing failed")
		log.Println("MarketsResult: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Get Operator DTO
	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 3.1. Return Error
		log.Println("MarketsResult: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// 3.2. Return Error
		log.Println("MarketsResult: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 4. Verify Signature
	/*
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("CancelBets: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("CancelBets: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	*/
	// 5. Get Markets
	marketKeys := []string{}
	for _, mrkt := range reqDto.Markets {
		marketKey := mrkt.ProviderId + "-" + mrkt.SportId + "-" + mrkt.EventId + "-" + mrkt.MarketId
		marketKeys = append(marketKeys, marketKey)
	}
	log.Println("MarketsResult: marketKeys count - ", len(marketKeys))
	markets, err := database.GetMarketsByMarketKeys(marketKeys)
	if err != nil {
		log.Println("MarketsResult: database.GetMarketsByMarketKeys failed with error - ", err.Error())
	}
	log.Println("MarketsResult: database.GetMarketsByMarketKeys returned markets count - ", len(markets))
	// 6. Prepare respDto.MarketResults
	for _, market := range markets {
		respDto.MarketResults = append(respDto.MarketResults, responsedto.GetMarketResult(market))
	}
	// 7. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// GetLicenseStatus is a function to get operator's license status
// @Summary      Get License status
// @Description  GetLicenseStatus checks operators current license status
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        Signature         header    string                             true  "Hash Signature"
// @Param        GetLicenseStatus  body      requestdto.GetLicenseStatusReqDto  true  "GetLicenseStatusReqDto model is used"
// @Success      200               {object}  responsedto.GetLicenseStatusRespDto{}
// @Failure      503               {object}  responsedto.GetLicenseStatusRespDto{}
// @Router       /feed/license-status [post]
func GetLicenseStatus(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	// 0. Get Default Response
	respDto := responsedto.GetLicenseStatusRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"
	respDto.LicenseStatus = false // Default to false

	// 1. Get Request
	reqDto := requestdto.GetMatchedBetsReqDto{}
	bodyStr := string(c.Body())
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("GetLicenseStatus: Error in body parser - ", err.Error())
		log.Println("GetLicenseStatus: Req. Body is - ", bodyStr)
		respDto.ErrorDescription = "Invalid Request!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Get License Status
	respDto.LicenseStatus = true
	// 4. Send SUCCESS response
	respDto.ErrorDescription = ""
	respDto.Status = "RS_OK"
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// User Statement
// Get User Statement
// @Summary      Get User Statement
// @Description  List Transactions of a user. Pagination is present. A maximum of 50 Transactions in a single request.
// @Tags         Feed
// @Accept       json
// @Produce      json
// @Param        Signature              header    string                           true  "Hash Signature"
// @Param        UserStatement  body      portaldto.UserStatementReqDto  true  "UserStatementReqDto model is used"
// @Success      200            {object}  portaldto.UserStatementRespDto
// @Failure      503            {object}  portaldto.UserStatementRespDto
// @Router       /feed/user-statement [post]
func UserStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := portaldto.UserStatementRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "GENERIC ERROR"
	respDto.Transactions = []portaldto.UserTransaction{}
	respDto.Balance = 0
	respDto.Page = 1 // default page number
	respDto.PageSize = 0
	respDto.TotalRecords = 0
	// 4. Parse request body to Request Object
	reqStr := string(c.Body())
	log.Println("UserStatement: Request Body is - ", reqStr)
	reqDto := portaldto.UserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("UserStatement: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request!!!"
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
	if operatorDto.Status != constants.SAP.ObjectStatus.ACTIVE() {
		log.Println("UserStatement: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Verify Signature
	if operatorDto.Signature {
		signature := c.Request().Header.Peek("Signature")
		log.Println("UserStatement: Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("UserStatement: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("UserStatement: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}
	// 8. Check if Operator wallet type is seamless or transfer?
	if operatorDto.WalletType != constants.SAP.WalletType.Transfer() {
		// 8.1. Seamless wallet, do balance call to operator
		log.Println("UserStatement: Operator wallet type is not transfer: ", operatorDto.WalletType)
		respDto.ErrorDescription = "This operation is not supported!!!"
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
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	// 10. Resturn data
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Transfer User Statement
// @Summary      Get Transfer User Statement
// @Description  Get Transfer User Statement
// @Tags         Feed
// @Accept       json
// @Produce      json
// @Param        Signature              header    string                           true  "Hash Signature"
// @Param        TransferUserStatement  body      operatordto.UserStatementReqDto  true  "UserStatementReqDto model is used"
// @Success      200                    {object}  reports.TransferUserStatementRespDto
// @Failure      503                    {object}  reports.TransferUserStatementRespDto
// @Router       /feed/transfer-user-statement [post]
func TransferUserStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := reports.TransferUserStatementRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"

	// 2. Create Request Object
	reqStr := string(c.Body())
	log.Println("TransferUserStatement: Request Body is - ", reqStr)
	reqDto := operatordto.UserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("TransferUserStatement: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request!!!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("TransferUserStatement: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Check Operator Status
	if operatorDto.Status != constants.SAP.ObjectStatus.ACTIVE() {
		log.Println("TransferUserStatement: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Verify Signature
	if operatorDto.Signature {
		signature := c.Request().Header.Peek("Signature")
		log.Println("TransferUserStatement: Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("TransferUserStatement: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("TransferUserStatement: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}
	statement, user, err := handler.GetTransferUserStatement(reqDto.OperatorId, reqDto.UserId, reqDto.ReferenceId, reqDto.StartTime, reqDto.EndTime)
	if err != nil {
		// 3.1. Return Error
		log.Println("TransferUserStatement: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Return Response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.UserId = reqDto.UserId
	respDto.UserName = user.UserName
	respDto.UserBalance = user.Balance
	respDto.Statement = statement
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Seemless User Statement
// @Summary      Get Seemless User Statement
// @Description  Get Seemless User Statement
// @Tags         Feed
// @Accept       json
// @Produce      json
// @Param        Signature      header    string                         true  "Hash Signature"
// @Param        SeemlessUserStatement  body      operatordto.UserStatementReqDto  true  "UserStatementReqDto model is used"
// @Success      200                    {object}  reports.SeemlessUserStatementRespDto
// @Failure      503                    {object}  reports.SeemlessUserStatementRespDto
// @Router       /feed/seemless-user-statement [post]
func SeemlessUserStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := reports.SeemlessUserStatementRespDto{}
	respDto.Status = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"

	// 2. Create Request Object
	reqStr := string(c.Body())
	log.Println("SeemlessUserStatement: Request Body is - ", reqStr)
	reqDto := operatordto.UserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SeemlessUserStatement: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request!!!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("SeemlessUserStatement: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, pleaes contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 7. Check Operator Status
	if operatorDto.Status != constants.SAP.ObjectStatus.ACTIVE() {
		log.Println("SeemlessUserStatement: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Access denied, Please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Verify Signature
	if operatorDto.Signature {
		signature := c.Request().Header.Peek("Signature")
		log.Println("SeemlessUserStatement: Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("SeemlessUserStatement: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("SeemlessUserStatement: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}

	statement, user, err := handler.GetSeemlessUserStatement(reqDto.OperatorId, reqDto.UserId, reqDto.ReferenceId, reqDto.Token, reqDto.StartTime, reqDto.EndTime)
	if err != nil {
		// 3.1. Return Error
		log.Println("SeemlessUserStatement: Failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Return Response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	respDto.UserId = reqDto.UserId
	respDto.UserBalance = user.Balance
	respDto.Statement = statement
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func SrPremiumMarkets(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// Create Default Response Object
	respDto := operatordto.SrPremiumMarketsRespDto{}
	respDto.EventStatus = "RS_ERROR"
	respDto.ErrorDescription = "Generic Error!"

	// Create Request Object
	reqStr := string(c.Body())
	log.Println("SrPremiumMarkets: Request Body is - ", reqStr)
	reqDto := operatordto.SrPremiumMarketsReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SrPremiumMarkets: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request!!!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	operatorDto, err := cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// Return Error
		log.Println("SrPremiumMarkets: GetOperatorDetails failed with error - ", err.Error())
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		// Return Error
		log.Println("SrPremiumMarkets: Operator is not active - ", operatorDto.Status)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if strings.ToLower(operatorDto.WalletType) != "feed" {
		log.Println("SrPremiumMarkets: Invalid wallet type - ", operatorDto.WalletType)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerId := reqDto.PartnerId
	if partnerId == "" {
		log.Println("SrPremiumMarkets: PartnerId is missing!")
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
				log.Println("SrPremiumMarkets: Partner is not Active: ", partner.Status)
				respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			break
		}
	}
	if false == found {
		log.Println("SrPremiumMarkets: Partner Id not found: ", partnerId)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Verify Signature
	if operatorDto.Signature {
		signature := c.Request().Header.Peek("Signature")
		log.Println("SrPremiumMarkets: Signature", string(signature))
		pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("SrPremiumMarkets: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
		signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("SrPremiumMarkets: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}
	// IsEventActive?
	if false == providers.IsProviderActive(reqDto.OperatorId, partnerId, reqDto.ProviderId) {
		// Return Error
		log.Println("SrPremiumMarkets: IsProviderActive returned false - ", reqDto.SportId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if false == providers.IsSportActive(reqDto.OperatorId, partnerId, reqDto.ProviderId, reqDto.SportId) {
		// Return Error
		log.Println("SrPremiumMarkets: IsSportActive returned false - ", reqDto.SportId)
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if false == providers.IsEventActive(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId) {
		// Return Error
		respDto.ErrorDescription = "Unauthorized access, please contact support!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	switch reqDto.ProviderId {
	case providers.DREAM_SPORT:
		// Dream - SrPremiumMarkets
		log.Println("SrPremiumMarkets: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	case providers.BETFAIR:
		// BetFair - SrPremiumMarkets
		log.Println("SrPremiumMarkets: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	case providers.SPORT_RADAR:
		// SportRadar - Bet Placement
		respDto, err = sportradar.GetBetFairMarket(reqDto.EventId)
		if err != nil {
			// 4.1. Return Error
			log.Println("SrPremiumMarkets: BetFair ValidateOdds failed with error - ", err.Error())
			respDto.ErrorDescription = "Failed to validate odds!"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	default:
		log.Println("SrPremiumMarkets: Invalid ProviderId - ", reqDto.ProviderId)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// Changing the status key to eventStatus
	respDto.EventStatus = respDto.Status

	// Return Response
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func ProviderPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.ProviderPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = ERROR_STATUS

	// 4. Get Request Body
	reqDto := reports.ProviderPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ProviderPnLReport: Body Parsing failed")
		log.Println("ProviderPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForProviderPnLReport(reqDto.OperatorId, reqDto)
	if err != nil {
		log.Println("ProviderPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	providerPlayed := make(map[string]reports.ProviderPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := providerPlayed[bet.SportId]; !ok {
				providerPlayed[bet.SportId] = reports.ProviderPnLReport{
					ProviderId: bet.ProviderId,
					SportName:  bet.BetDetails.SportName,
					SportId:    bet.SportId,
					ProfitLoss: bet.NetAmount,
					BetCount:   1,
				}
			} else {
				gReport := providerPlayed[bet.SportId]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				providerPlayed[bet.SportId] = gReport
			}
		}
	}

	for _, provider := range providerPlayed {
		pl := reports.ProviderPnLReport{}
		pl.ProviderId = provider.ProviderId
		pl.SportName = provider.SportName
		pl.SportId = provider.SportId
		pl.ProfitLoss = provider.ProfitLoss
		pl.BetCount = provider.BetCount
		respDto.ProviderPnLReports = append(respDto.ProviderPnLReports, pl)
	}
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.ProviderId = reqDto.ProviderId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func SportPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.SportPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = ERROR_STATUS

	// 4. Get Request Body
	reqDto := reports.SportPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SportPnLReport: Body Parsing failed")
		log.Println("SportPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForSportPnLReport(reqDto.OperatorId, reqDto)
	if err != nil {
		log.Println("SportPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	competitionPlayed := make(map[string]reports.SportPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := competitionPlayed[bet.CompetitionId]; !ok {
				competitionPlayed[bet.CompetitionId] = reports.SportPnLReport{
					SportName:       bet.BetDetails.SportName,
					SportId:         bet.SportId,
					CompetitionName: bet.BetDetails.CompetitionName,
					CompetitionId:   bet.CompetitionId,
					ProfitLoss:      bet.NetAmount,
					BetCount:        1,
				}
			} else {
				gReport := competitionPlayed[bet.CompetitionId]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				competitionPlayed[bet.CompetitionId] = gReport
			}
		}
	}

	for _, competition := range competitionPlayed {
		pl := reports.SportPnLReport{}
		pl.SportName = competition.SportName
		pl.SportId = competition.SportId
		pl.CompetitionName = competition.CompetitionName
		pl.CompetitionId = competition.CompetitionId
		pl.ProfitLoss = competition.ProfitLoss
		pl.BetCount = competition.BetCount
		respDto.SportPnLReports = append(respDto.SportPnLReports, pl)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.SportId = reqDto.SportId
	return c.Status(fiber.StatusOK).JSON(respDto)
}
func CompetitionPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.CompetitionPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = ERROR_STATUS

	// 4. Get Request Body
	reqDto := reports.CompetitionPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("CompetitionPnLReport: Body Parsing failed")
		log.Println("CompetitionPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForCompetitionPnLReport(reqDto.OperatorId, reqDto)
	if err != nil {
		log.Println("CompetitionPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	EventPlayed := make(map[string]reports.CompetitionPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := EventPlayed[bet.EventId]; !ok {
				EventPlayed[bet.EventId] = reports.CompetitionPnLReport{
					CompetitionName: bet.BetDetails.CompetitionName,
					CompetitionId:   bet.CompetitionId,
					EventName:       bet.BetDetails.EventName,
					EventId:         bet.EventId,
					ProfitLoss:      bet.NetAmount,
					BetCount:        1,
				}
			} else {
				gReport := EventPlayed[bet.EventId]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				EventPlayed[bet.EventId] = gReport
			}
		}
	}

	for _, event := range EventPlayed {
		pl := reports.CompetitionPnLReport{}
		pl.CompetitionName = event.CompetitionName
		pl.CompetitionId = event.CompetitionId
		pl.EventName = event.EventName
		pl.EventId = event.EventId
		pl.ProfitLoss = event.ProfitLoss
		pl.BetCount = event.BetCount
		respDto.CompetitionPnLReports = append(respDto.CompetitionPnLReports, pl)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.CompetitionId = reqDto.CompetitionId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func EventPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.EventPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = ERROR_STATUS

	// 4. Get Request Body
	reqDto := reports.EventPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("EventPnLReport: Body Parsing failed")
		log.Println("EventPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForEventPnLReport(reqDto.OperatorId, reqDto)
	if err != nil {
		log.Println("EventPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	marketTypePlayed := make(map[string]reports.EventPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := marketTypePlayed[bet.BetDetails.MarketName]; !ok {
				marketTypePlayed[bet.BetDetails.MarketName] = reports.EventPnLReport{
					EventId:    bet.EventId,
					EventName:  bet.BetDetails.EventName,
					MarketType: bet.BetDetails.MarketType,
					MarketName: bet.BetDetails.MarketName,
					MarketId:   bet.MarketId,
					ProfitLoss: bet.NetAmount,
					BetCount:   1,
				}
			} else {
				gReport := marketTypePlayed[bet.BetDetails.MarketName]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				marketTypePlayed[bet.BetDetails.MarketName] = gReport
			}
		}
	}

	for _, marketType := range marketTypePlayed {
		pl := reports.EventPnLReport{}
		pl.EventId = marketType.EventId
		pl.EventName = marketType.EventName
		pl.MarketType = marketType.MarketType
		pl.MarketName = marketType.MarketName
		pl.MarketId = marketType.MarketId
		pl.ProfitLoss = marketType.ProfitLoss
		pl.BetCount = marketType.BetCount
		respDto.EventPnLReports = append(respDto.EventPnLReports, pl)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.EventId = reqDto.EventId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func SyncOperatorBalance() {
	// Get all the operators
	operators, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println("SyncOperatorBalance: Error in getting operators - ", err.Error())
		return
	}
	for _, operMap := range operators {
		operator := operMap.(operatordto.OperatorDTO)
		// Get User Ledger
		userLedger, err := database.GetLedgersByOperatorId(operator.OperatorId)
		if err != nil {
			log.Println("SyncOperatorBalance: Error in getting user ledger - ", err.Error())
			return
		}
		// Get Individual User Data,
		users := make(map[string][]models.UserLedgerDto)
		for _, ledger := range userLedger {
			if _, ok := users[ledger.UserId]; !ok {
				users[ledger.UserId] = []models.UserLedgerDto{}
			}
			users[ledger.UserId] = append(users[ledger.UserId], ledger)
		}
		updateLedgers := []models.UserLedger2Dto{}
		for user, ledgers := range users {
			// Get User Balance
			userBalance, err := database.GetUserBalance(user)
			if err != nil {
				log.Println("SyncOperatorBalance: Error in getting user balance - ", err.Error())
				return
			}
			var beforeBalance float64
			var lastSyncTime int64
			if userBalance.Balance != 0 {
				beforeBalance = userBalance.Balance
			}
			for _, ledger := range ledgers {
				if ledger.TransactionTime > userBalance.LastSyncTime {
					updateLedger := models.UserLedger2Dto{}
					updateLedger.ID = ledger.ID
					updateLedger.UserKey = ledger.UserKey
					updateLedger.OperatorId = ledger.OperatorId
					updateLedger.UserId = ledger.UserId
					updateLedger.TransactionType = ledger.TransactionType
					updateLedger.TransactionTime = ledger.TransactionTime
					updateLedger.ReferenceId = ledger.ReferenceId
					updateLedger.Amount = ledger.Amount
					updateLedger.Remark = ledger.Remark
					updateLedger.CompetitionName = ledger.CompetitionName
					updateLedger.EventName = ledger.EventName
					updateLedger.MarketType = ledger.MarketType
					updateLedger.MarketName = ledger.MarketName
					updateLedger.BeforeBalance = beforeBalance
					updateLedger.AfterBalance = beforeBalance + ledger.Amount
					beforeBalance = updateLedger.AfterBalance
					lastSyncTime = ledger.TransactionTime
					updateLedgers = append(updateLedgers, updateLedger)
				}
			}
			// Update User Balance
			if len(updateLedgers) > 0 {
				err = database.UpdateUserBalance(user, beforeBalance, lastSyncTime)
				if err != nil {
					log.Println("SyncOperatorBalance: Error in updating user balance - ", err.Error())
				}
			}
		}
		// Insert UserLedger2
		if len(updateLedgers) > 0 {
			err = database.InsertManyLedgers2(updateLedgers)
			if err != nil {
				log.Println("SyncOperatorBalance: Error in inserting user ledger - ", err.Error())
			}
		}
	}
}
