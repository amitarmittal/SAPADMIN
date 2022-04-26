package operatorsvc

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	sessDto "Sp/dto/session"
	"Sp/handler"
	keyutils "Sp/utilities"
	utils "Sp/utilities"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
