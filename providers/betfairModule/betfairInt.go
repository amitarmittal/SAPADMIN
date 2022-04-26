package betfairmodule

import (
	"Sp/database"
	"Sp/dto/models"
	"Sp/providers/betfairModule/request"
	"Sp/providers/betfairModule/response"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	// Prod
	// USERNAME       string = "HEERAEXCH3"
	// PASSWORD       string = "Dubai2022!"
	// APP_KEY        string = "kY1TMiN3l859uxwN"
	// Dev
	// USERNAME            string = "HEERAEXCH1"
	// PASSWORD            string = "HeeraX2022!!"
	// APP_KEY             string = "tUgKyOisayQLIn4f"
	USERNAME            string = os.Getenv("USERNAME")
	PASSWORD            string = os.Getenv("PASSWORD")
	APP_KEY             string = os.Getenv("APP_KEY")
	LOGIN_URL           string = "https://identitysso.betfair.com/api/login"
	KEEP_ALIVE_URL      string = "https://identitysso.betfair.com/api/keepAlive"
	BASE_URL            string = "https://api.betfair.com/exchange/betting/rest/v1.0"
	PLACE_ORDER         string = "/placeOrders/"
	LIST_CURRENT_ORDERS string = "/listCurrentOrders/"
	LIST_CLEARED_ORDER  string = "/listClearedOrders/"
	CANCEL_ORDERS       string = "/cancelOrders/"

	ReqTimeOut time.Duration = 30 // 5 seconds
	Session    models.BetFairSession
	// Metrics
	PlOMetrics BetFairMetrics = BetFairMetrics{Request: PLACE_ORDER}
	CuOMetrics BetFairMetrics = BetFairMetrics{Request: LIST_CURRENT_ORDERS}
	ClOMetrics BetFairMetrics = BetFairMetrics{Request: LIST_CLEARED_ORDER}
	CaOMetrics BetFairMetrics = BetFairMetrics{Request: CANCEL_ORDERS}
)

// KeepAlive method to call at init & scheduler
func BetFairKeepAlive() {
	log.Println("BetFairModule: BetFairKeepAlive: START!!!")
	// 1. Read Session Details from DB
	// 	1.1. Error - Call Login
	// 	1.2. Save Session in memory
	// 	1.3. Save Session Details in DB
	//	1.4. Return
	// 2. Save Session in memory
	// 3. Call Keep Alive
	// 	3.1. Error - Call Login
	// 	3.2. Save Session in memory
	// 	3.3. Save Session Details in DB
	//	3.4. Return
	// 4. Save Session in memory
	// 5. Update Session Details in DB
	// 6. Return

	Session = models.BetFairSession{}
	var err error
	// 1. Read Session Details from DB
	session1, err := database.GetSessionByProduct(APP_KEY)
	if err != nil {
		// 1.1. Error - Call Login
		session1, err = Login(USERNAME, PASSWORD, APP_KEY)
		if err != nil {
			log.Println("BetFairModule: BetFairKeepAlive: Login failed with error - ", err.Error())
			log.Println("BetFairModule: BetFairKeepAlive: FAILED END 1 !!!")
			return
		}
		// 1.2. Save Session in memory
		Session = session1
		// 1.3. Save Session Details in DB
		err = database.InsertSession(Session)
		if err != nil {
			log.Println("BetFairModule: BetFairKeepAlive: database.InsertSession failed with error - ", err.Error())
			log.Println("BetFairModule: BetFairKeepAlive: FAILED END 2 !!!")
			return
		}
		if Session.Status != "SUCCESS" {
			log.Println("BetFairModule: BetFairKeepAlive: Login failed with status & error - ", Session.Status, Session.Error)
			log.Println("BetFairModule: BetFairKeepAlive: FAILED END 3 !!!")
			return
		}
		log.Println("BetFairModule: BetFairKeepAlive: SUCCESS END 4 !!!")
		return
	}
	// 2. Save Session in memory
	Session = session1
	// 3. Call Keep Alive
	session2, err := KeepAlive(Session.Token, Session.Product)
	if err != nil {
		log.Println("BetFairModule: BetFairKeepAlive: KeepAlive failed with error - ", err.Error())
		// 2.1. Error - Call Login
		session2, err = Login(USERNAME, PASSWORD, APP_KEY)
		if err != nil {
			log.Println("BetFairModule: BetFairKeepAlive: Login failed with error - ", err.Error())
			log.Println("BetFairModule: BetFairKeepAlive: FAILED END 5 !!!")
			return
		}
		// 2.2. Save Session in memory
		// Session = session2
		Session.Status = session2.Status
		Session.Error = session2.Error
		// 2.3. Save Session Details in DB
		err = database.InsertSession(Session)
		if err != nil {
			log.Println("BetFairModule: BetFairKeepAlive: database.InsertSession failed with error - ", err.Error())
			log.Println("BetFairModule: BetFairKeepAlive: FAILED END 6 !!!")
			return
		}
		if Session.Status != "SUCCESS" {
			log.Println("BetFairModule: BetFairKeepAlive: Login failed with status & error - ", Session.Status, Session.Error)
			log.Println("BetFairModule: BetFairKeepAlive: FAILED END 7 !!!")
			return
		}
		log.Println("BetFairModule: BetFairKeepAlive: SUCCESS END 8 !!!")
		return
	}
	// 4. Save Session in memory
	// Session = session2
	Session.Status = session2.Status
	Session.Error = session2.Error
	// 5. Update Session Details in DB
	err = database.UpdateSession(Session)
	if err != nil {
		log.Println("BetFairModule: BetFairKeepAlive: database.UpdateSession failed with error - ", err.Error())
		log.Println("BetFairModule: BetFairKeepAlive: FAILED END 9 !!!")
		return
	}
	if Session.Status != "SUCCESS" {
		log.Println("BetFairModule: BetFairKeepAlive: KeepAlive failed with status & error - ", Session.Status, Session.Error)
		log.Println("BetFairModule: BetFairKeepAlive: FAILED END 10 !!!")
		return
	}
	// 6. Return
	log.Println("BetFairModule: BetFairKeepAlive: SUCCESS END 11 !!!")
	return
}

// Login method to call from swagger
func Login2() {
	Login(USERNAME, PASSWORD, APP_KEY)
}

// BetFair Login & KeepAlive Methods
func Login(userName string, password string, product string) (models.BetFairSession, error) {
	session := models.BetFairSession{}
	data := url.Values{}
	data.Set("username", userName)
	data.Set("password", password)
	dataStr := data.Encode()
	log.Println("BetFairModule: Login: request data is - ", dataStr, LOGIN_URL)

	bReqBody := strings.NewReader(dataStr)
	req, err := http.NewRequest("POST", LOGIN_URL, bReqBody)
	if err != nil {
		log.Println("BetFairModule: Login: http.NewRequest failed with error - ", err.Error())
		return session, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Application", product)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// 4. Create HTTP Client with Timeout
	client := &http.Client{
		Timeout: ReqTimeOut * time.Second,
	}
	// 5. Make Request
	// log.Println("BetFairModule: Login: Time before request is :  ", time.Now().String())
	resp, err := client.Do(req)
	// log.Println("BetFairModule: Login: Time after request is :  ", time.Now().String())
	if err != nil {
		log.Println("BetFairModule: Login: Request Failed with error - ", err.Error())
		return session, err
	}
	defer resp.Body.Close()
	// 6. Read response body
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("BetFairModule: Login: ioutil.ReadAll failed with error - ", err.Error())
		return session, err
	}
	log.Println("BetFairModule: Login: response data is - ", string(respbody))

	err = json.Unmarshal(respbody, &session)
	if err != nil {
		log.Println("GetEvents: Login: json.Unmarshal failed with error - ", err.Error())
		return session, err
	}
	return session, nil
}

func KeepAlive(token string, product string) (models.BetFairSession, error) {
	session := models.BetFairSession{}
	data := url.Values{}
	dataStr := data.Encode()
	log.Println("BetFairModule: KeepAlive: request data is - ", dataStr, KEEP_ALIVE_URL)

	bReqBody := strings.NewReader(dataStr)
	req, err := http.NewRequest("POST", KEEP_ALIVE_URL, bReqBody)
	if err != nil {
		log.Println("BetFairModule: KeepAlive: http.NewRequest failed with error - ", err.Error())
		return session, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Authentication", token)
	req.Header.Add("X-Application", product)
	// 4. Create HTTP Client with Timeout
	client := &http.Client{
		Timeout: ReqTimeOut * time.Second,
	}
	// 5. Make Request
	// log.Println("BetFairModule: KeepAlive: Time before request is :  ", time.Now().String())
	resp, err := client.Do(req)
	// log.Println("BetFairModule: KeepAlive: Time after request is :  ", time.Now().String())
	if err != nil {
		log.Println("BetFairModule: KeepAlive: Request Failed with error - ", err.Error())
		return session, err
	}
	defer resp.Body.Close()
	// 6. Read response body
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("BetFairModule: KeepAlive: ioutil.ReadAll failed with error - ", err.Error())
		return session, err
	}
	log.Println("BetFairModule: KeepAlive: response data is - ", string(respbody))

	err = json.Unmarshal(respbody, &session)
	if err != nil {
		log.Println("GetEvents: KeepAlive: json.Unmarshal failed with error - ", err.Error())
		return session, err
	}
	return session, nil
}

// BetFair Order Methods
func PlaceOrders(listPO request.PlaceOrderReq) (response.PlaceOrderResp, error) {
	logkey := "BetFairModule: PlaceBet: PlaceOrders: "
	respObj := response.PlaceOrderResp{}
	var url string = BASE_URL + PLACE_ORDER
	log.Println(logkey+"url is - ", url)
	payload := make(map[string]interface{})
	if listPO.MarketId == "" {
		return respObj, fmt.Errorf("MarketId cannot be empty!!!")
	}
	payload["marketId"] = listPO.MarketId
	instructions := []interface{}{}
	for _, pi := range listPO.Instructions {
		instruction := make(map[string]interface{})
		instruction["orderType"] = pi.OrderType
		instruction["selectionId"] = pi.SelectionId
		//instruction["handicap"] = pi.Handicap
		instruction["side"] = pi.Side
		limitOrder := make(map[string]interface{})
		limitOrder["size"] = pi.LimitOrder.Size
		limitOrder["price"] = pi.LimitOrder.Price
		limitOrder["persistenceType"] = pi.LimitOrder.PersistencaType
		if pi.LimitOrder.PersistencaType != "PERSIST" {
			limitOrder["timeInForce"] = pi.LimitOrder.TimeInForce
		}
		instruction["limitOrder"] = limitOrder
		if pi.CustomerOrderRef != "" {
			instruction["customerOrderRef"] = pi.CustomerOrderRef
		}
		instructions = append(instructions, instruction)
	}
	payload["instructions"] = instructions
	if listPO.CustomerRef != "" {
		payload["customerRef"] = listPO.CustomerRef
	}
	//payload["sync"] = true
	respbody, respDetails, err := HTTPCall(payload, url, logkey)
	if err != nil {
		log.Println(logkey+"HTTPCall failed with error - ", err.Error())
		return respObj, err
	}
	// Update Metric
	PlOMetrics.Details = append(PlOMetrics.Details, respDetails)
	err = json.Unmarshal(respbody, &respObj)
	if err != nil {
		log.Println(logkey+"json.Unmarshal failed with error - ", err.Error())
		return respObj, err
	}
	return respObj, nil
}

func CurrentOrders(listCO request.ListCurrentOrdersReq) (response.CurrentOrdersResp, error) {
	logkey := "BetFairModule: CurrentOrders: "
	respObj := response.CurrentOrdersResp{}
	var url string = BASE_URL + LIST_CURRENT_ORDERS
	log.Println(logkey+"url is - ", url)
	payload := make(map[string]interface{})
	if len(listCO.BetIds) > 0 {
		payload["betIds"] = listCO.BetIds
	}
	if len(listCO.MarketIds) > 0 {
		payload["marketIds"] = listCO.MarketIds
	}
	if len(listCO.CustomerOrderRefs) > 0 {
		payload["customerOrderRefs"] = listCO.CustomerOrderRefs
	}
	if listCO.RecordCount != 0 {
		payload["fromRecord"] = listCO.FromRecord
		payload["recordCount"] = listCO.RecordCount
	}
	respbody, respDetails, err := HTTPCall(payload, url, logkey)
	if err != nil {
		log.Println(logkey+"HTTPCall failed with error - ", err.Error())
		return respObj, err
	}
	// Update Metric
	CuOMetrics.Details = append(CuOMetrics.Details, respDetails)
	err = json.Unmarshal(respbody, &respObj)
	if err != nil {
		log.Println(logkey+"json.Unmarshal failed with error - ", err.Error())
		return respObj, err
	}
	return respObj, nil
}

func ClearedOrders(listCO request.ListClearedOrdersReq) (response.ClearedOrdersResp, error) {
	logkey := "BetFairModule: ClearedOrders: "
	respObj := response.ClearedOrdersResp{}
	var url string = BASE_URL + LIST_CLEARED_ORDER
	log.Println(logkey+"url is - ", url)
	payload := make(map[string]interface{})
	if len(listCO.BetIds) > 0 {
		payload["betIds"] = listCO.BetIds
	}
	if listCO.BetStatus != "" {
		payload["betStatus"] = listCO.BetStatus
	}
	respbody, respDetails, err := HTTPCall(payload, url, logkey)
	if err != nil {
		log.Println(logkey+"HTTPCall failed with error - ", err.Error())
		return respObj, err
	}
	// Update Metric
	ClOMetrics.Details = append(ClOMetrics.Details, respDetails)
	err = json.Unmarshal(respbody, &respObj)
	if err != nil {
		log.Println(logkey+"json.Unmarshal failed with error - ", err.Error())
		return respObj, err
	}
	return respObj, nil
}

func CancelOrders(listCO request.ListCancelOrdersReq) (response.CancelOrdersResp, error) {
	logkey := "BetFairModule: CancelOrders: "
	respObj := response.CancelOrdersResp{}
	var url string = BASE_URL + CANCEL_ORDERS
	log.Println(logkey+"url is - ", url)
	payload := make(map[string]interface{})
	if listCO.MarketId != "" {
		payload["marketId"] = listCO.MarketId
	}
	instructions := []interface{}{}
	for _, ci := range listCO.Instructions {
		instruction := make(map[string]interface{})
		instruction["betId"] = ci.BetId
		if ci.SizeReduction > 0 {
			instruction["sizeReduction"] = ci.SizeReduction
		}
		instructions = append(instructions, instruction)
	}
	payload["instructions"] = instructions
	if listCO.CustomerRef != "" {
		payload["customerRef"] = listCO.CustomerRef
	}
	respbody, respDetails, err := HTTPCall(payload, url, logkey)
	if err != nil {
		log.Println(logkey+"HTTPCall failed with error - ", err.Error())
		return respObj, err
	}
	// Update Metric
	CaOMetrics.Details = append(CaOMetrics.Details, respDetails)
	err = json.Unmarshal(respbody, &respObj)
	if err != nil {
		log.Println(logkey+"json.Unmarshal failed with error - ", err.Error())
		return respObj, err
	}
	return respObj, nil
}

// BetFair Order HTTP Call
func HTTPCall(payload map[string]interface{}, url string, logkey string) ([]byte, ResponseDetails, error) {
	respDetails := ResponseDetails{}
	respbody := []byte{}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		log.Println(logkey+"json.Marshal failed with error - ", err.Error())
		return respbody, respDetails, err
	}
	bReqBody := bytes.NewBuffer(reqBody)
	req, err := http.NewRequest("POST", url, bReqBody)
	if err != nil {
		log.Println(logkey+"http.NewRequest failed with error - ", err.Error())
		return respbody, respDetails, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Application", Session.Product)
	req.Header.Add("X-Authentication", Session.Token)
	// 4. Create HTTP Client with Timeout
	client := &http.Client{
		Timeout: ReqTimeOut * time.Second,
	}
	// 5. Make Request
	log.Println(logkey+"request data is - ", string(reqBody))
	startTime := time.Now()
	resp, err := client.Do(req)
	//log.Println("SportsFeedCall: Time after request is :  ", time.Now().String())
	if err != nil {
		log.Println(logkey+"Request Failed with error - ", err.Error())
		return respbody, respDetails, err
	}
	diff := time.Now().Sub(startTime)
	defer resp.Body.Close()
	// 6. Read response body
	respbody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(logkey+"Response ReadAll failed with error - ", err.Error())
		return respbody, respDetails, err
	}
	log.Println(logkey+"response data is - ", string(respbody))
	// Metrics
	respDetails.StartTime = startTime.Format(time.RFC3339Nano)
	respDetails.ExecutionTime = diff.Milliseconds()
	respDetails.Description = string(reqBody)
	return respbody, respDetails, nil
}
