package testsvc

import (
	"Sp/cache"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	opdto "Sp/dto/operator"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"Sp/dto/sports"
	testDto "Sp/dto/test"
	operatorsvc "Sp/handler/operator"
	keyutils "Sp/utilities"
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	bal float64 = 10000
)

func TestKeys(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := testDto.TestAuthRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("TestKeys: Req. Body is - ", bodyStr)
	reqDto := new(dto.AuthReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("TestKeys: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("TestKeys: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("TestKeys: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// Verify Signature
	//signature := c.Request().Header.Peek("Signature")
	priKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(operatorDto.Keys.PrivateKey)
	signature, err := keyutils.CreateSignature(string(c.Body()), *priKey)
	log.Println("TestKeys: PublicKey is: ", operatorDto.Keys.PublicKey)
	log.Println("TestKeys: Signature is: ", string(signature))
	operatorKeys := operatorDto.Keys
	pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorKeys.PublicKey)
	if err != nil {
		log.Println("TestKeys: Parsing public key failed: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
	if !signValid {
		log.Println("TestKeys: Signature verification failed - ", err.Error())
		respDto.ErrorDescription = "Bad Request.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	respDto.Signature = signature
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func TestAuth(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := testDto.TestAuthRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := new(dto.AuthReqDto)
	if err := c.BodyParser(reqDto); err != nil {
		log.Println("TestAuth: Body Parsing failed")
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
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("TestAuth: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("TestAuth: Failed to get Operator Details: ", err.Error())
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	clientDto, err := cache.GetOperatorDetails("TestClient")
	if err != nil {
		log.Println("TestAuth: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// Verify Signature
	//signature := c.Request().Header.Peek("Signature")
	clientKeys := clientDto.Keys
	priKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(clientKeys.PrivateKey)
	signature, err := keyutils.CreateSignature(string(c.Body()), *priKey)
	log.Println("Signature", string(signature))
	operatorKeys := operatorDto.Keys
	pubKey, err := keyutils.ParseRsaPublicKeyFromPemStr(operatorKeys.OperatorKey)
	if err != nil {
		log.Println("TestAuth: Parsing public key failed: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	signValid := keyutils.VerifySignature(string(signature), string(c.Body()), *pubKey)
	if !signValid {
		log.Println("TestAuth: Signature verification failed - ", err.Error())
		respDto.ErrorDescription = "Bad Request.!"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	respDto.Signature = signature
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func SimplePing(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	t := time.Now().UnixNano()
	log.Println("SimplePing: nano time is - ", t)
	log.Println("SimplePing: milli time is - ", t/time.Hour.Milliseconds())
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Success",
	})
}

func SleepPing(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	time.Sleep(2 * time.Second)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Success",
	})
}

/*
func DatabasePing(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	Operator := new(models.Operator)
	if err := c.BodyParser(Operator); err != nil {
		return err
	}
	Operator.OperatorKey = getToken(10)
	Operator.PrivateKey = getToken(10)
	Operator.PublicKey = getToken(10)
	result := database.DBWrite.Create(&Operator)
	fmt.Println(result.RowsAffected)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Success",
	})
}
*/
func HttpPing(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")

	resp, err := http.Get("http://webcode.me")
	time.Sleep(7 * time.Second)
	if err != nil {
		log.Println("Response went wrong")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": false,
			"message": resp,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Success",
		"desc":    resp.Body,
	})
}

func TestBalance(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	dt := time.Now()
	log.Println("TestBalance: Current date and time is: ", dt.String())
	balance := new(opdto.BalanceReqDto)
	fmt.Println(balance.OperatorId)
	if err := c.BodyParser(balance); err != nil {
		return err
	}

	if balance.OperatorId == "" {
		fmt.Println("OperatorId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if balance.Token == "" {
		fmt.Println("OperatorToken not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if balance.UserId == "" {
		fmt.Println("UserId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	// TODO: Verify Signature

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Status":  0,
		"Balance": bal,
	})
}

func TestBet(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	dt := time.Now()
	log.Println("Current date and time is: ", dt.String())
	bet := new(opdto.BetReqDto)
	fmt.Println(bet.OperatorId)
	if err := c.BodyParser(bet); err != nil {
		return err
	}

	if bet.OperatorId == "" {
		fmt.Println("OperatorId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.Token == "" {
		fmt.Println("OperatorToken not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.UserId == "" {
		fmt.Println("UserId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.ReqId == "" {
		fmt.Println("ReqId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.TransactionId == "" {
		fmt.Println("TransactionId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.EventId == "" {
		fmt.Println("TableId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.MarketId == "" {
		fmt.Println("RoundId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if bet.DebitAmount <= 0 {
		fmt.Println("Debit Amount not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	// TODO: Verify Signature

	bal -= bet.DebitAmount
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Status":  0,
		"Balance": bal,
	})
}

func TestRollback(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	dt := time.Now()
	log.Println("Current date and time is: ", dt.String())
	rollBack := new(opdto.RollbackReqDto)
	fmt.Println(rollBack.OperatorId)
	if err := c.BodyParser(rollBack); err != nil {
		return err
	}

	if rollBack.OperatorId == "" {
		fmt.Println("OperatorId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if rollBack.Token == "" {
		fmt.Println("OperatorToken not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if rollBack.UserId == "" {
		fmt.Println("UserId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if rollBack.ReqId == "" {
		fmt.Println("ReqId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if rollBack.TransactionId == "" {
		fmt.Println("TransactionId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if rollBack.EventId == "" {
		fmt.Println("TableId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if rollBack.MarketId == "" {
		fmt.Println("RoundId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	// TODO: Verify Signature

	// rollback Amount can be positive or negative
	bal -= rollBack.RollbackAmount
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Status":  0,
		"Balance": bal,
	})
}

func TestResult(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	dt := time.Now()
	log.Println("Current date and time is: ", dt.String())
	result := new(opdto.ResultReqDto)
	fmt.Println(result.OperatorId)
	if err := c.BodyParser(result); err != nil {
		return err
	}

	if result.OperatorId == "" {
		fmt.Println("OperatorId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.Token == "" {
		fmt.Println("OperatorToken not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.UserId == "" {
		fmt.Println("UserId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.ReqId == "" {
		fmt.Println("ReqId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.TransactionId == "" {
		fmt.Println("TransactionId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.EventId == "" {
		fmt.Println("TableId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.MarketId == "" {
		fmt.Println("RoundId not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	if result.CreditAmount <= 0 {
		fmt.Println("Credit Amount not found")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status":  400,
			"Balance": "0",
		})
	}

	// TODO: Verify Signature

	bal += result.CreditAmount
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Status":  0,
		"Balance": bal,
	})
}

// Automated Test cases
var (
	AutomationId      string = "TestClient"
	OperatorId        string = "TestOperator"
	UserId            string = "AutomationUser"
	PlatformIdStr     string = "desktop"
	PlatformIdInt     int    = 0
	Currency          string = "INR"
	ClientIp          string = "1.1.1.1"
	UserName          string = "Automation User"
	GameId            string = "test"
	LoginURL          string = "http://localhost:3000/auth/login"
	GetUserBalanceURL string = "http://localhost:3000/core/getuserbalance"
	GetSportsURL      string = "http://localhost:3000/core/getsports"
	GetLiveEventsURL  string = "http://localhost:3000/core/getliveevents"
	GetEventsURL      string = "http://localhost:3000/core/getevents"
	GetMarketsURL     string = "http://localhost:3000/core/getmarkets"
	GetOpenBetsURL    string = "http://localhost:3000/core/getopenbets"
	SportsBetURL      string = "http://localhost:3000/core/sportsbet"
)

type ResultMap map[string]string

func RunTestAutomation(c *fiber.Ctx) error {
	var results []ResultMap
	log.Println("RunTestAutomation: *** Test Automation Started. ***")
	res1 := TestDreamSuccess()
	results = append(results, res1)
	log.Println("RunTestAutomation: res is - ", results)
	//log.Println("RunTestAutomation: *** Test Automation Completed. ***")
	return c.Status(fiber.StatusOK).JSON(results)
}

// Test case 1: Ezugi - Happyday scenario
func TestDreamSuccess() map[string]string {
	result := make(map[string]string)
	//providerName := "Dream"
	result["0-DreamTestCase"] = "Started ..."
	// 1. Make Login Call and get Token
	log.Println("\n\nTestDreamSuccess: *** 1. Login/Authentication ***")
	opLoginResp, err := TLLogin()
	if !err {
		result["1-DreamLogin"] = "Failed"
		return result
	}
	result["1-DreamLogin"] = "Success"
	token := opLoginResp.Token
	log.Println("DreamTest: Operator Token is - ", token)
	log.Println("DreamTest: Game URL is - ", opLoginResp.Url)
	strLoginResp := ""
	bytesLoginResp, err1 := json.Marshal(opLoginResp)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strLoginResp = string(bytesLoginResp)
		log.Println("DreamTest: GetMarket response is - \n\n", strLoginResp)
	}

	// 2. Get User Balance
	time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 2. GetUserBalance ***")
	getUserBalanceResp, status := TLGetUserBalance(token, OperatorId)
	if !status {
		result["2-DreamGetUserBalance"] = "Failed"
		log.Println("DreamTest: GetUserBalance failed with error - ", getUserBalanceResp.ErrorDescription)
		return result
	}
	result["2-DreamGetUserBalance"] = "Success"
	strGetUserBalanceResp := ""
	bytesGetUserBalanceResp, err1 := json.Marshal(getUserBalanceResp)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strGetUserBalanceResp = string(bytesGetUserBalanceResp)
		log.Println("DreamTest: GetUserBalance response is - \n\n", strGetUserBalanceResp)
	}

	// 3. Get Sports
	//time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 3. GetSports ***")
	getSportsResp, status := TLGetSports(token, OperatorId)
	if !status {
		result["3-DreamGetSports"] = "Failed"
		log.Println("DreamTest: GetSports failed with error - ", getSportsResp.ErrorDescription)
		return result
	}
	result["3-DreamGetSports"] = "Success"
	strGetSportsResp := ""
	bytesGetSportsResp, err1 := json.Marshal(getSportsResp)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strGetSportsResp = string(bytesGetSportsResp)
		log.Println("DreamTest: GetMarket response is - \n\n", strGetSportsResp)
	}

	// 4. Get Live Events
	//time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 4. GetLiveEvents ***")
	getLiveEventsResp, status := TLGetLiveEvents(token, OperatorId, "Dream")
	if !status {
		result["4-DreamGetLiveEvents"] = "Failed"
		log.Println("DreamTest: GetLiveEvents failed with error - ", getLiveEventsResp.ErrorDescription)
		return result
	}
	result["4-DreamGetLiveEvents"] = "Success"
	/*
		strGetLiveEventsResp := ""
		bytesGetLiveEventsResp, err1 := json.Marshal(getLiveEventsResp)
		if err1 != nil {
			log.Println("DreamTest: json marsal has issues - ", err1.Error())
		} else {
			strGetLiveEventsResp = string(bytesGetLiveEventsResp)
			log.Println("DreamTest: GetMarket response is - \n\n", strGetLiveEventsResp)
		}
	*/
	// 5. Get Events
	//time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 5. GetEvents ***")
	getEventsResp, status := TLGetEvents(token, OperatorId, "Dream", "4")
	if !status {
		result["5-DreamGetEvents"] = "Failed"
		log.Println("DreamTest: GetGetEvents failed with error - ", getEventsResp.ErrorDescription)
		return result
	}
	result["5-DreamGetEvents"] = "Success"
	// 6. Get Markets
	eventdto := dto.EventDto{}
	for _, event := range getLiveEventsResp.Events {
		//if event.SportId == "4" {
		eventdto = event
		break
		//}
	}
	strCricLiveEventData := ""
	bytesCricLiveEventData, err1 := json.Marshal(eventdto)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strCricLiveEventData = string(bytesCricLiveEventData)
		log.Println("DreamTest: Cric Live Event Data is - \n\n", strCricLiveEventData)
	}
	//time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 6. GetMarkets ***")
	getMarketsResp, status := TLGetMarkets(token, OperatorId, eventdto)
	if !status {
		result["6-DreamGetMarkets"] = "Failed"
		log.Println("DreamTest: GetMarkets failed with error - ", getMarketsResp.ErrorDescription)
		return result
	}
	result["6-DreamGetMarkets"] = "Success"
	/**/
	strGetMarketResp := ""
	bytesGetMarketResp, err1 := json.Marshal(getMarketsResp)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strGetMarketResp = string(bytesGetMarketResp)
		log.Println("DreamTest: GetMarket response is - \n\n", strGetMarketResp)
	}
	/**/
	// 7. Make Bet Call
	eventDto := getMarketsResp.Event
	//time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 7. Bet/Debit ***")
	txId := uuid.New().String()
	betResp, betReq, err := TLDreamBet(token, txId, eventDto)
	if !err {
		result["7-DreamBet"] = "Failed"
		return result
	}
	result["7-DreamBet"] = "Success"
	log.Println("DreamTest: Bet Status is - ", betResp.Status)
	log.Println("DreamTest: Bet Type is - ", betReq.BetType)
	strBetResp := ""
	bytesBetResp, err1 := json.Marshal(betResp)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strBetResp = string(bytesBetResp)
		log.Println("DreamTest: DreamBet response is - \n\n", strBetResp)
	}

	// 8. Get Open Bets
	//time.Sleep(1 * time.Second)
	log.Println("\n\nTestDreamSuccess: *** 8. GetOpenBets ***")
	openBetsResp, err := TLGetOpenBets(token, OperatorId, eventDto)
	if !err {
		result["8-OpenBets"] = "Failed"
		return result
	}
	result["8-OpenBets"] = "Success"
	strOpenBetsResp := ""
	bytesOpenBetsResp, err1 := json.Marshal(openBetsResp)
	if err1 != nil {
		log.Println("DreamTest: json marsal has issues - ", err1.Error())
	} else {
		strOpenBetsResp = string(bytesOpenBetsResp)
		log.Println("DreamTest: GetOpenBets response is - \n\n", strOpenBetsResp)
	}

	// 9. Make Result Call

	// 10. Make Rollback Call

	return result
}

func TLLogin() (dto.AuthRespDto, bool) {
	data := dto.AuthRespDto{}
	// 1. Get Test Automation Keys from Operator table
	automationOpDTO, err := cache.GetOperatorDetails(AutomationId)
	//autoKeyPair := automationOpDTO.KeySet()
	// 2. Construct Request
	authReqDto := dto.AuthReqDto{}
	authReqDto.OperatorId = OperatorId
	authReqDto.UserId = UserId
	authReqDto.Username = UserName
	authReqDto.PlatformId = PlatformIdStr
	authReqDto.ClientIp = ClientIp
	authReqDto.Currency = Currency
	jsonData, err := json.Marshal(authReqDto)
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: Failed to convert DTO to JSON")
		return data, false
	}
	// 3. Make Request
	//payload := string(jsonData)
	//log.Println("TLLogin: Request Body is - ", payload)
	rsaPriKey, err := keyutils.ParseRsaPrivateKeyFromPemStr(automationOpDTO.Keys.PrivateKey)
	respBody, isSuccess := TLMakeRequest(jsonData, LoginURL, *rsaPriKey)
	if !isSuccess {
		log.Println("TLLogin: Test Automation Failure: *** Login Request Failed ***")
		return data, false
	}
	//bodyString := string(respBody)
	//log.Println("TLLogin: Response Body is - ", bodyString)
	// 4. Parse response
	err = json.Unmarshal([]byte(respBody), &data)
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: *** Login Response Convertion Failed ***")
		return data, false
	}
	if data.Status != "RS_OK" {
		log.Println("TLLogin: Test Automation Failure: *** Login Request Failed with Status Code - ***", data.Status)
		log.Println("TLLogin: Test Automation Failure: *** Login Request Failed with Error Description - ***", data.ErrorDescription)
		return data, false
	}
	return data, true
}

func TLGetUserBalance(token string, operatorId string) (dto.GetUserBalanceRespDto, bool) {
	// 0. Default response object
	respDto := dto.GetUserBalanceRespDto{}
	// 1. Create request object
	reqDto := dto.GetUserBalanceReqDto{}
	reqDto.Token = token
	reqDto.OperatorId = operatorId

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("TLGetUserBalance: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, GetUserBalanceURL)
	if !isSuccess {
		log.Println("TLGetUserBalance: Test Automation Failure: *** Request Failed ***")
		return respDto, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &respDto)
	if err != nil {
		log.Println("TLGetUserBalance: Test Automation Failure: *** Response Convertion Failed ***")
		return respDto, false
	}
	if respDto.Status != "RS_OK" {
		log.Println("TLGetUserBalance: Test Automation Failure: *** Request Failed with Status Code - ***", respDto.Status)
		return respDto, false
	}
	return respDto, true
}

func TLGetSports(token string, operatorId string) (commondto.GetSportsRespDto, bool) {
	// 0. Default response object
	respDto := commondto.GetSportsRespDto{}
	// 1. Create request object
	reqDto := commondto.GetSportsReqDto{}
	reqDto.Token = token
	reqDto.OperatorId = operatorId

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("TLGetSports: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, GetSportsURL)
	if !isSuccess {
		log.Println("TLGetSports: Test Automation Failure: *** Hub88 Bet Request Failed ***")
		return respDto, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &respDto)
	if err != nil {
		log.Println("TLGetSports: Test Automation Failure: *** Login Response Convertion Failed ***")
		return respDto, false
	}
	if respDto.Status != "RS_OK" {
		log.Println("TLGetSports: Test Automation Failure: *** Balance Request Failed with Status Code - ***", respDto.Status)
		return respDto, false
	}
	return respDto, true
}

func TLGetLiveEvents(token string, operatorId string, providerId string) (dto.GetLiveEventsRespDto, bool) {
	// 0. Default response object
	respDto := dto.GetLiveEventsRespDto{}
	// 1. Create request object
	reqDto := dto.GetLiveEventsReqDto{}
	reqDto.Token = token
	reqDto.OperatorId = operatorId
	reqDto.ProviderId = providerId

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("TLGetLiveEvents: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, GetLiveEventsURL)
	if !isSuccess {
		log.Println("TLGetLiveEvents: Test Automation Failure: *** Hub88 Bet Request Failed ***")
		return respDto, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &respDto)
	if err != nil {
		log.Println("TLGetLiveEvents: Test Automation Failure: *** Login Response Convertion Failed ***")
		return respDto, false
	}
	if respDto.Status != "RS_OK" {
		log.Println("TLGetLiveEvents: Test Automation Failure: *** Balance Request Failed with Status Code - ***", respDto.Status)
		return respDto, false
	}
	return respDto, true
}

func TLGetEvents(token string, operatorId string, providerId string, sportId string) (dto.GetEventsRespDto, bool) {
	// 0. Default response object
	respDto := dto.GetEventsRespDto{}
	// 1. Create request object
	reqDto := dto.GetEventsReqDto{}
	reqDto.Token = token
	reqDto.OperatorId = operatorId
	reqDto.ProviderId = providerId
	reqDto.SportId = sportId

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("TLGetEvents: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, GetEventsURL)
	if !isSuccess {
		log.Println("TLGetEvents: Test Automation Failure: *** Hub88 Bet Request Failed ***")
		return respDto, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &respDto)
	if err != nil {
		log.Println("TLGetEvents: Test Automation Failure: *** Login Response Convertion Failed ***")
		return respDto, false
	}
	if respDto.Status != "RS_OK" {
		log.Println("TLGetEvents: Test Automation Failure: *** Balance Request Failed with Status Code - ***", respDto.Status)
		return respDto, false
	}
	return respDto, true
}

func TLGetMarkets(token string, operatorId string, event dto.EventDto) (dto.GetMarketsRespDto, bool) {
	// 0. Default response object
	respDto := dto.GetMarketsRespDto{}
	// 1. Create request object
	reqDto := dto.GetMarketsReqDto{}
	reqDto.Token = token
	reqDto.OperatorId = operatorId
	reqDto.ProviderId = event.ProviderId
	reqDto.SportId = event.SportId
	reqDto.EventId = event.EventId

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("TLGetMarkets: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, GetMarketsURL)
	if !isSuccess {
		log.Println("TLGetMarkets: Test Automation Failure: *** Hub88 Bet Request Failed ***")
		return respDto, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &respDto)
	if err != nil {
		log.Println("TLGetMarkets: Test Automation Failure: *** Login Response Convertion Failed ***")
		return respDto, false
	}
	if respDto.Status != "RS_OK" {
		log.Println("TLGetMarkets: Test Automation Failure: *** Balance Request Failed with Status Code - ***", respDto.Status)
		return respDto, false
	}
	return respDto, true
}

func TLGetOpenBets(token string, operatorId string, event dto.EventDto) (dto.GetOpenBetsRespDto, bool) {
	// 0. Default response object
	respDto := dto.GetOpenBetsRespDto{}
	// 1. Create request object
	reqDto := dto.GetOpenBetsReqDto{}
	reqDto.Token = token
	reqDto.OperatorId = operatorId
	reqDto.ProviderId = event.ProviderId
	reqDto.SportId = event.SportId
	reqDto.EventId = event.EventId

	jsonData, err := json.Marshal(reqDto)
	if err != nil {
		log.Println("TLGetOpenBets: Test Automation Failure: Failed to convert DTO to JSON")
		return respDto, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, GetOpenBetsURL)
	if !isSuccess {
		log.Println("TLGetOpenBets: Test Automation Failure: *** Request Failed ***")
		return respDto, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &respDto)
	if err != nil {
		log.Println("TLGetOpenBets: Test Automation Failure: *** Response Convertion Failed ***")
		return respDto, false
	}
	if respDto.Status != "RS_OK" {
		log.Println("TLGetOpenBets: Test Automation Failure: *** Request Failed with Status Code - ***", respDto.Status)
		return respDto, false
	}
	return respDto, true
}

func TLDreamBet(token string, txId string, event dto.EventDto) (dto.SportsBetRespDto, dto.SportsBetReqDto, bool) {
	// 0. Default response object
	betResponse := dto.SportsBetRespDto{}

	// 1. Construct Request
	betRequest := dto.SportsBetReqDto{}
	betRequest.Token = token
	betRequest.OperatorId = OperatorId
	betRequest.ProviderId = event.ProviderId
	betRequest.SportId = event.SportId
	betRequest.CompetetionId = event.CompetitionId
	betRequest.EventId = event.EventId
	betRequest.MarketType = event.Markets.MatchOdds[0].MarketType
	betRequest.MarketName = event.Markets.MatchOdds[0].MarketName
	betRequest.MarketId = event.Markets.MatchOdds[0].MarketId
	betRequest.BetType = "BACK"
	betRequest.OddValue = event.Markets.MatchOdds[0].Runners[0].BackPrices[2].Price
	betRequest.StakeAmount = 1000
	betRequest.RunnerId = event.Markets.MatchOdds[0].Runners[0].RunnerId
	betRequest.RunnerName = event.Markets.MatchOdds[0].Runners[0].RunnerName
	betRequest.SessionOutcome = 0

	jsonData, err := json.Marshal(betRequest)
	if err != nil {
		log.Println("TLDreamBet: Test Automation Failure: Failed to convert DTO to JSON")
		return betResponse, betRequest, false
	}
	// 2. Make Request
	respBody, isSuccess := TLMakeRequest2(jsonData, SportsBetURL)
	if !isSuccess {
		log.Println("TLDreamBet: Test Automation Failure: *** Hub88 Bet Request Failed ***")
		return betResponse, betRequest, false
	}
	// 3. Parse response
	err = json.Unmarshal([]byte(respBody), &betResponse)
	if err != nil {
		log.Println("TLDreamBet: Test Automation Failure: *** Login Response Convertion Failed ***")
		return betResponse, betRequest, false
	}
	if betResponse.Status != "RS_OK" {
		log.Println("TLDreamBet: Test Automation Failure: *** Balance Request Failed with Status Code - ***", betResponse.Status)
		return betResponse, betRequest, false
	}
	return betResponse, betRequest, true
}

// Test Automation common methods
func TLMakeRequest(jsonData []byte, URL string, privKey rsa.PrivateKey) ([]byte, bool) {
	var emptyResp string
	//resp := []byte(respStr)
	// 1. Create Signature
	NewSignature, err := keyutils.CreateSignature(string(jsonData), privKey)
	if err != nil {
		log.Println("TLMakeRequest: Create signature failed with error - ", err.Error())
		return []byte(emptyResp), false
	}
	log.Println(URL)
	//log.Println("TLMakeRequest: Signature is - ", NewSignature)
	// 4. Create HTTP Request
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: Failed to create HTTP Req object")
		return []byte(emptyResp), false
	}
	req.Header.Add("Signature", NewSignature)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	// 5. Make Request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: *** Login Request Failed ***")
		return []byte(emptyResp), false
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: *** Login Response Failure ***")
		return []byte(emptyResp), false
	}
	return respbody, true
}

func TLMakeRequest2(jsonData []byte, URL string) ([]byte, bool) {
	var emptyResp string
	log.Println(URL)
	// 4. Create HTTP Request
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: Failed to create HTTP Req object")
		return []byte(emptyResp), false
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	// 5. Make Request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: *** Login Request Failed ***")
		return []byte(emptyResp), false
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("TLLogin: Test Automation Failure: *** Login Response Failure ***")
		return []byte(emptyResp), false
	}
	return respbody, true
}

//Internal Testings
//GO-Concurrency Testing
func Concurrency(c *fiber.Ctx) error {
	go FiveHundreadMilliSecDelay()
	i := 1
	for {
		TwoSecDelay()
		if i >= 5 {
			break
		}
		i += 1
	}
	data := dto.AuthRespDto{}
	return c.Status(fiber.StatusOK).JSON(data)
}

func TwoSecDelay() {
	time.Sleep(time.Millisecond * 2000)
	fmt.Println("TwoSecDelay")
}

func FiveHundreadMilliSecDelay() bool {
	i := 0
	for {
		time.Sleep(time.Millisecond * 500)
		fmt.Println(i, "FiveHundreadMilliSecDelay")
		if i >= 25 {
			break
		}
		i += 1
	}
	return false
}

//GO-Concurrency with Channel Testing
func ConcurrencyWithChannel(c *fiber.Ctx) error {
	flag := make(chan bool)
	go FiveHundreadMilliSecDelayChannel(flag)
	i := 1
	for {
		TwoSecDelay()
		if i >= 5 {
			break
		}
		i += 1
	}
	data := dto.AuthRespDto{}
	fmt.Println(<-flag)
	return c.Status(fiber.StatusOK).JSON(data)
}

func FiveHundreadMilliSecDelayChannel(nums chan bool) {
	i := 0
	for {
		time.Sleep(time.Millisecond * 500)
		fmt.Println(i, "FiveHundreadMilliSecDelay")
		if i >= 25 {
			break
		}
		i += 1
	}
	nums <- false
}

func InsertSport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.AddSportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("InsertSport: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check
	if reqDto.ProviderId == "" {
		log.Println("InsertSport: ProviderId is missing")
		respDto.ErrorDescription = "ProviderId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportId == "" {
		log.Println("InsertSport: SportId is missing")
		respDto.ErrorDescription = "SportId is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.SportName == "" {
		log.Println("InsertSport: SportName is missing")
		respDto.ErrorDescription = "SportName is missing!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if reqDto.Status == "" {
		reqDto.Status = "ACTIVE"
	}
	// 4. Get Provider Details
	provider, err := cache.GetProvider(reqDto.ProviderId)
	if err != nil {
		log.Println("InsertSport: GetProvider failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Provider"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Create Sport
	sport := models.Sport{}
	sport.SportKey = reqDto.ProviderId + "-" + reqDto.SportId
	sport.ProviderId = provider.ProviderId
	sport.ProviderName = provider.ProviderName
	sport.SportId = reqDto.SportId
	sport.SportName = reqDto.SportName
	sport.Status = reqDto.Status
	sport.CreatedAt = time.Now().Unix()
	// Insert Sport
	err = database.InsertSport(sport)
	if err != nil {
		log.Println("InsertSport: InsertSport failed with error - ", err.Error())
		respDto.ErrorDescription = "InsertSport Failed"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Get Operators
	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("InsertSport: GetAllOperators failed with error - ", err.Error())
		respDto.ErrorDescription = "GetAllOperators Failed"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	listSportStatus := []models.SportStatus{}
	for _, op := range operators {
		// create SportStatus
		sportStatus := models.SportStatus{}
		sportStatus.SportKey = op.OperatorId + "-" + sport.SportKey
		sportStatus.OperatorId = op.OperatorId
		sportStatus.OperatorName = op.OperatorName
		sportStatus.ProviderId = sport.ProviderId
		sportStatus.ProviderName = sport.ProviderName
		sportStatus.SportId = sport.SportId
		sportStatus.SportName = sport.SportName
		sportStatus.ProviderStatus = reqDto.Status
		sportStatus.OperatorStatus = reqDto.Status
		sportStatus.CreatedAt = sport.CreatedAt
		// add to the list SportStatus
		listSportStatus = append(listSportStatus, sportStatus)
	}
	if len(listSportStatus) > 0 {
		err = database.InsertManySportStatus(listSportStatus)
		if err != nil {
			log.Println("InsertSport: InsertManySportStatus failed with error - ", err.Error())
			respDto.ErrorDescription = "InsertManySportStatus Failed"
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	}
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func ResetConfig(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "DEPRICATED"
	return c.Status(fiber.StatusOK).JSON(respDto)

	// reqDto := requestdto.ResetConfigReqDto{}
	// if err := c.BodyParser(&reqDto); err != nil {
	// 	log.Println("ResetConfig: Body Parsing failed")
	// 	respDto.ErrorDescription = "Invalid Request"
	// 	return c.Status(fiber.StatusBadRequest).JSON(respDto)
	// }
	// // Events
	// events, err := database.GetOpEvents(reqDto.OperatorId)
	// if err != nil {
	// 	log.Println("ResetConfig: GetOpEvents failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "GetOpEvents Failed"
	// 	return c.Status(fiber.StatusBadRequest).JSON(respDto)
	// }
	// count := 0
	// for _, event := range events {
	// 	if event.Config.IsSet {
	// 		event.Config.IsSet = false
	// 		count += 1
	// 		err = database.ReplaceEventStatus(event)
	// 		if err != nil {
	// 			log.Println("ResetConfig: Update Event Details failed with error - ", err.Error())
	// 			return err
	// 		}
	// 		// Update Event Cache
	// 		mapValue := map[string]models.EventStatus{}
	// 		mapValue[event.EventId] = event
	// 		//cache.SetOpEventStatus(event.OperatorId, event.ProviderId, event.SportId, mapValue)
	// 	}
	// }
	// log.Println("ResetConfig: ResetConfig - ", count, " events")
	// count = 0
	// // Competitions
	// comps, err := database.GetOpCompetitions(reqDto.OperatorId)
	// if err != nil {
	// 	log.Println("ResetConfig: GetOpCompetitions failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "GetOpCompetitions Failed"
	// 	return c.Status(fiber.StatusBadRequest).JSON(respDto)
	// }
	// for _, comp := range comps {
	// 	if comp.Config.IsSet {
	// 		count += 1
	// 		comp.Config.IsSet = false
	// 		// update Competition DB
	// 		err = database.ReplaceCompetitionStatus(comp)
	// 		if err != nil {
	// 			log.Println("ResetConfig: Update Competition Details failed with error - ", err.Error())
	// 			return err
	// 		}
	// 		// Update Competition Cache
	// 		mapValue := map[string]models.CompetitionStatus{}
	// 		mapValue[comp.CompetitionId] = comp
	// 		//cache.SetOpCompetitionStatus(comp.OperatorId, comp.ProviderId, comp.SportId, mapValue)
	// 	}
	// }
	// log.Println("ResetConfig: ResetConfig - ", count, " competitions")
	// count = 0
	// // Sports
	// sports, err := database.GetOpSports(reqDto.OperatorId, "")
	// if err != nil {
	// 	log.Println("ResetConfig: GetOpSports failed with error - ", err.Error())
	// 	respDto.ErrorDescription = "GetOpSports Failed"
	// 	return c.Status(fiber.StatusBadRequest).JSON(respDto)
	// }
	// for _, sport := range sports {
	// 	if sport.Config.IsSet {
	// 		count += 1
	// 		sport.Config.IsSet = false
	// 		// update Sport DB
	// 		err = database.ReplaceSportStatus(sport)
	// 		if err != nil {
	// 			log.Println("ResetConfig: Update Sport Details failed with error - ", err.Error())
	// 			return err
	// 		}
	// 		// Update Sport Cache
	// 		mapValue := map[string]models.SportStatus{}
	// 		mapValue[sport.SportId] = sport
	// 		//cache.SetOpSportStatus(sport.OperatorId, sport.ProviderId, mapValue)
	// 	}
	// }
	// log.Println("ResetConfig: ResetConfig - ", count, " sports")
	// respDto.Status = "RS_OK"
	// respDto.ErrorDescription = ""
	// return c.Status(fiber.StatusBadRequest).JSON(respDto)
}

// func UpdateOperators(c *fiber.Ctx) error {
// 	c.Accepts("json", "text")
// 	c.Accepts("application/json")
// 	// 1. Create Default Response Object
// 	respDto := responsedto.DefaultRespDto{}
// 	respDto.ErrorDescription = "Generic Error"
// 	respDto.Status = "RS_ERROR"

// 	// Get Operators
// 	operators, err := database.GetAllOperators()
// 	if err != nil {
// 		log.Println("UpdateOperators: GetAllOperators failed with error - ", err.Error())
// 		respDto.ErrorDescription = "GetAllOperators Failed"
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	log.Println("UpdateOperators: operators count is - ", len(operators))
// 	// Get Providers
// 	providers, err := database.GetAllProviders()
// 	if err != nil {
// 		log.Println("UpdateOperators: GetAllProviders failed with error - ", err.Error())
// 		respDto.ErrorDescription = "GetAllProviders Failed"
// 		return c.Status(fiber.StatusOK).JSON(respDto)
// 	}
// 	log.Println("UpdateOperators: providers count is - ", len(providers))
// 	// Create update operators list
// 	updatedOperators := []opdto.OperatorDTO{}
// 	// Iterate through all operators
// 	for _, operator := range operators {
// 		isUpdated := false
// 		// Check for config map
// 		if operator.Config == (commondto.ConfigDto{}) || len(operator.Config) == 0 {
// 			log.Println("UpdateOperators: Config is missing for operator - ", operator.OperatorId)
// 			// Create config map
// 			operator.Config = make(map[string]opdto.ProviderConfig)
// 			for _, provider := range providers {
// 				config := opdto.ProviderConfig{}
// 				config.OperatorHold = 90
// 				config.PlatformHold = 2
// 				config.PlatformComm = 2
// 				config.BetDelay = make(map[string]int32)
// 				config.BetDelay["MATCH_ODDS"] = 5
// 				config.BetDelay["BOOKMAKER"] = 2
// 				config.BetDelay["FANCY"] = 2
// 				operator.Config[provider.ProviderId] = config
// 			}
// 			isUpdated = true
// 		}
// 		//
// 		if isUpdated {
// 			updatedOperators = append(updatedOperators, operator)
// 		}
// 	}
// 	log.Println("UpdateOperators: updatedOperators count is - ", len(updatedOperators))
// 	if len(updatedOperators) > 0 {
// 		/*
// 			count, msgs, err := database.UpdateOperators(updatedOperators)
// 			if err != nil {
// 				log.Println("UpdateOperators: UpdateOperators failed with error - ", err.Error())
// 				respDto.ErrorDescription = "UpdateOperators Failed"
// 				return c.Status(fiber.StatusOK).JSON(respDto)
// 			}
// 			log.Println("UpdateOperators: msgs count is - ", len(msgs))
// 			log.Println("UpdateOperators: count is - ", count)
// 		*/
// 	}
// 	respDto.Status = "RS_OK"
// 	respDto.ErrorDescription = ""
// 	return c.Status(fiber.StatusOK).JSON(respDto)
// }

// func DuplicateCheck(c *fiber.Ctx) error {
// 	c.Accepts("json", "text")
// 	c.Accepts("application/json")
// 	events, _ := database.GetLatestEvents()
// 	respDto := responsedto.DefaultRespDto{}
// 	var eveKey []string
// 	for _, event := range events {
// 		eveKey = append(eveKey, event.EventKey)
// 	}
// 	log.Println(eveKey)
// 	respDto.Status = "RS_OK"
// 	respDto.ErrorDescription = ""
// 	return c.Status(fiber.StatusOK).JSON(respDto)
// }

// func UpdateStatus(c *fiber.Ctx) error {
// 	c.Accepts("json", "text")
// 	c.Accepts("application/json")
// 	// 1. Create Default Response Object
// 	database.UpdateALLCompetitionStatus("ACTIVE")
// 	return c.Status(fiber.StatusOK).JSON("SUCCESS")
// }

func UpdateNetAmountToBet() {
	bets, err := database.GetAllBets(opdto.BetsHistoryReqDto{})
	if err != nil {
		log.Println("UpdateNetAmountToBet: GetAllBets failed with error - ", err.Error())
		return
	}
	log.Println("UpdateNetAmountToBet: bets count is - ", len(bets))
	for _, bet := range bets {
		if bet.Status != "OPEN" {
			bet.NetAmount = -bet.BetReq.DebitAmount
			if len(bet.ResultReqs) > 0 {
				for _, result := range bet.ResultReqs {
					bet.NetAmount += result.CreditAmount
				}
			}
			if len(bet.RollbackReqs) > 0 {
				for _, rollback := range bet.RollbackReqs {
					bet.NetAmount += rollback.RollbackAmount
				}
			}
			log.Println("UpdateNetAmountToBet: bet.NetAmount is - ", bet.NetAmount, " betId: ", bet.BetId)
		}
	}
	log.Println("UpdateNetAmountToBet: bets count is - ", len(bets))
	// count, msgs := database.UpdateBets(bets)
	// if len(msgs) > 0 {
	// 	log.Println("UpdateNetAmountToBet: msgs count is - ", len(msgs))
	// }
	// log.Println("UpdateNetAmountToBet: count is - ", count)
}

// @Summary      PaginationReport is a function to bring data from database
// @Description  bring bets data from database.
// @Tags         Test-Service
// @Accept       json
// @Produce      json
// @Param        PaginationReport  body      requestdto.BetsReportReq  true  "BetsReportReq model is used"
// @Success      200               {object}  responsedto.BetReportResp{}
// @Failure      503               {object}  responsedto.BetReportResp{}
// @Router       /test/pagination-report [post]
func PaginationReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	log.Println("PaginationReport: START!!! - ", time.Now())
	// 1. Create Default Response Object
	resp := responsedto.BetReportResp{}
	resp.Status = "RS_ERROR"
	resp.ErrorDescription = "GNERIC ERROR"
	//resp.BetsData = responsedto.BetData
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("PaginationReport: Req. Body is - ", bodyStr)
	reqDto := requestdto.BetsReportReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("PlaceBet: Body Parsing failed with error - ", err.Error())
		log.Println("PlaceBet: Req. Body is - ", bodyStr)
		resp.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	reqDto.Page = 1
	reqDto.PageSize = 100
	data, count, err := database.GetBetsReportA(reqDto.Page, reqDto.PageSize)
	if err != nil {
		log.Println("PaginationReport: database.GetBetsReport failed with error - ", err.Error())
		return c.Status(fiber.StatusOK).JSON(resp)
	}
	log.Println("PaginationReport: database.GetBetsReport count is - ", count)
	var totalCount int = 0
	updateBets := []sports.BetDto{}
	// var isFound bool = false
	log.Println("PaginationReport: Loop START - ", time.Now())
	for {
		data, count, err = database.GetBetsReportA(reqDto.Page, reqDto.PageSize)
		if err != nil {
			log.Println("PaginationReport: database.GetBetsReport failed with error - ", err.Error())
			return c.Status(fiber.StatusOK).JSON(resp)
		}
		//log.Println("PaginationReport: database.GetBetsReport reqDto.Page is - ", reqDto.Page, time.Now())
		totalCount += len(data[0].Data)
		for _, bet := range data[0].Data {
			if bet.NetAmount != 0 {
				continue
			}
			switch bet.Status {
			case "CANCELLED", "CANCELLED-failed", "LAPSED", "LAPSED-failed", "VOIDED", "VOIDED-failed", "EXPIRED", "TIMELY_VOIDED", "SETTLED_VOIDED":
				continue
			default:
				log.Println("PaginationReport: ZERO NetAmount for bet.Status - ", bet.Status, bet.BetId)
			}
			netAmount := operatorsvc.GetWinAmount(bet)
			if bet.NetAmount != netAmount {
				log.Println("PaginationReport: NetAmount mismatch for betId - ", bet.BetId)
				log.Println("PaginationReport: NetAmount mismatch bet.Status - ", bet.Status)
				log.Println("PaginationReport: NetAmount mismatch bet.NetAmout - ", bet.NetAmount)
				log.Println("PaginationReport: NetAmount mismatch NetAmount is - ", netAmount)
				bet.NetAmount = netAmount
				updateBets = append(updateBets, bet)
			}
			//bet.NetAmount = operatorsvc.GetWinAmount(bet)
			// if len(bet.RollbackReqs) > 0 && len(bet.ResultReqs) > 0 {
			// 	log.Println("PaginationReport: betid - NetAmount : ", bet.BetId, data[0].Data[i].NetAmount)
			// }
		}
		// if isFound == true {
		// 	break
		// }
		if len(data[0].Data) != int(reqDto.PageSize) {
			log.Println("PaginationReport: database.GetBetsReport len(data[0].Data) is - ", len(data[0].Data))
			break
		}
		reqDto.Page++
	}
	log.Println("PaginationReport: Loop END - ", time.Now())
	log.Println("PaginationReport: updateBets count is - ", len(updateBets))
	// sCount, eMsgs := database.UpdateBets(updateBets)
	// log.Println("PaginationReport: database.UpdateBets Success count is - ", sCount)
	// if sCount != len(updateBets) {
	// 	for _, eMsg := range eMsgs {
	// 		log.Println("PaginationReport: database.UpdateBets error message is - ", eMsg)
	// 	}
	// }
	// log.Println("PaginationReport: Update END - ", time.Now())
	//resp.BetsData = append(resp.BetsData, data[0].Data...)
	resp.Page = reqDto.Page
	resp.PageSize = int64(len(data[0].Data))
	resp.TotalRecords = data[0].Total[0].Count
	resp.Status = "RS_OK"
	resp.ErrorDescription = ""
	log.Println("PaginationReport: END!!! - ", time.Now())
	return c.Status(fiber.StatusOK).JSON(resp)
}

// func GetNetAmount(betDto sports.BetDto) int64 {
// 	var netAmount int64 = 0
// 	// Subtract BetReq.DebitAmount (user risk)
// 	// Add all result.CreditAmount s (user win)
// 	// Subtract all Rollback.RollbackAmount
// 	return netAmount
// }

func PaginationReport2(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	log.Println("PaginationReport: START!!! - ", time.Now())
	// 1. Create Default Response Object
	resp := responsedto.BetReportResp{}
	resp.Status = "RS_ERROR"
	resp.ErrorDescription = "GNERIC ERROR"
	//resp.BetsData = responsedto.BetData
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("PaginationReport: Req. Body is - ", bodyStr)
	reqDto := requestdto.BetsReportReq{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("PlaceBet: Body Parsing failed with error - ", err.Error())
		log.Println("PlaceBet: Req. Body is - ", bodyStr)
		resp.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	data, count, err := database.GetBetsReport(reqDto.Page, reqDto.PageSize, reqDto.OperatorId, "OPEN")
	if err != nil {
		log.Println("PaginationReport: database.GetBetsReport failed with error - ", err.Error())
		return c.Status(fiber.StatusOK).JSON(resp)
	}
	log.Println("PaginationReport: database.GetBetsReport count is - ", count)
	resp.BetsData = append(resp.BetsData, data[0].Data...)
	resp.Page = reqDto.Page
	resp.PageSize = reqDto.PageSize
	resp.TotalRecords = data[0].Total[0].Count
	resp.Status = "RS_OK"
	resp.ErrorDescription = ""
	log.Println("PaginationReport: END!!! - ", time.Now())
	return c.Status(fiber.StatusOK).JSON(resp)
}
