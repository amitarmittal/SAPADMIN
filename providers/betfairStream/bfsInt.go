package betfairstream

import (
	"Sp/database"
	"Sp/providers/betfairStream/request"
	"Sp/providers/betfairStream/response"
	"crypto/tls"
	"encoding/json"
	"log"
	"time"
)

var (
	// Prod
	// USERNAME       string = "HEERAEXCH3"
	// PASSWORD       string = "Dubai2022!"
	// APP_KEY        string = "kY1TMiN3l859uxwN"
	// Dev
	// USERNAME string = "HEERAEXCH1"
	// PASSWORD string = "HeeraX2022!!"
	APP_KEY string = "tUgKyOisayQLIn4f"
	// USERNAME            string = os.Getenv("USERNAME")
	// PASSWORD            string = os.Getenv("PASSWORD")
	// APP_KEY             string = os.Getenv("APP_KEY")
	BSA_URL    string = "stream-api.betfair.com:443"
	BSA_REQ_ID int32  = 10000
	OP_AUTh    string = "authentication"
)

func TestBFSConnections() {
	logKey := "betfairstream: TestBFSConnections: "
	log.Println(logKey+"START - ", time.Now())
	// 0. Get Session Details from DB
	session1, err := database.GetSessionByProduct(APP_KEY)
	if err != nil {
		log.Println(logKey+"database.GetSessionByProduct failed with error - ", err.Error())
		return
	}
	log.Println(logKey+"Session Details - ", session1.Product, session1.Token)
	// 1. Request - Establish Connection to BetFair Stream API
	config := tls.Config{}
	conn, err := tls.Dial("tcp", BSA_URL, &config)
	if err != nil {
		log.Println(logKey+"tls.Dial failed with error - ", err.Error())
		return
	}
	defer conn.Close()
	log.Println(logKey+"tls.Dial conn.RemoteAddr - ", conn.RemoteAddr())
	// 2. Response - Read Connection Message from BetFair Stream API
	dataBytes := make([]byte, 2048)
	dataLength, err := conn.Read(dataBytes)
	if err != nil {
		log.Println(logKey+"conn.Read failed with error - ", err.Error())
		return
	}
	log.Println(logKey+"conn.Read data - ", dataLength, string(dataBytes))
	if dataLength == 0 {
		log.Println(logKey + "dataLength is ZERO!!!")
		return
	}
	// supress CRLF
	dataBytes = dataBytes[:dataLength-1]
	// 3. Map to Connection Message Object
	connMessage := response.ConnectionMessage{}
	err = json.Unmarshal(dataBytes, &connMessage)
	if err != nil {
		log.Println(logKey+"json.Unmarshal failed with error - ", err.Error())
		return
	}
	log.Println(logKey+"ConnectionMessage op & connectionId are - ", connMessage.Op, connMessage.ConnectionId)
	// 4. TODO: Save to DB
	// 5. Request - Authentication Message
	authReq := request.AuthenticationMessage{}
	authReq.Op = "authentication" // make it constant
	BSA_REQ_ID++                  // increment ID
	authReq.Id = BSA_REQ_ID
	authReq.AppKey = APP_KEY
	authReq.Session = session1.Token
	authMessage1, err := json.Marshal(authReq)
	if err != nil {
		log.Println(logKey+"json.Marshal failed with error - ", err.Error())
		return
	}
	amLength := len(authMessage1)
	log.Println(logKey+"request authMessage1 Length - ", amLength)
	authMessage2 := make([]byte, amLength+2)
	for i, data := range authMessage1 {
		authMessage2[i] = data
	}
	authMessage2[amLength] = '\r'
	authMessage2[amLength+1] = '\n'
	log.Println(logKey+"request authMessage2 Length - ", len(authMessage2))
	authMsgString := string(authMessage2)
	log.Println(logKey+"authMessage - ", authMsgString)
	dataLength, err = conn.Write(authMessage2)
	if err != nil {
		log.Println(logKey+"conn.Write failed with error - ", err.Error())
		return
	}
	log.Println(logKey+"conn.Write dataLength - ", dataLength)
	// 6. Response - Read Status Message from BetFair Stream API
	dataBytes = make([]byte, 2048)
	dataLength, err = conn.Read(dataBytes)
	if err != nil {
		log.Println(logKey+"conn.Read failed with error - ", err.Error())
		return
	}
	log.Println(logKey+"conn.Read data - ", dataLength, string(dataBytes))
	if dataLength == 0 {
		log.Println(logKey + "dataLength is ZERO!!!")
		return
	}
	// supress CRLF
	dataBytes = dataBytes[:dataLength-1]
	// 7. Map to Connection Message Object
	statusMessage := response.StatusMessage{}
	err = json.Unmarshal(dataBytes, &statusMessage)
	if err != nil {
		log.Println(logKey+"json.Unmarshal failed with error - ", err.Error())
		return
	}
	log.Println(logKey+"statusMessage op, Id & StatusCode are - ", statusMessage.Op, statusMessage.Id, statusMessage.StatusCode)
}

// type APIClient struct {
// }

// func (c *APIClient) CallAPI(path string, method string, postBody interface{}, headerParams map[string]string,
// 	queryParams url.Values, formParams map[string]string, fileName string, fileBytes []byte) (*resty.Response, error) {

// 	request := prepareRequest(postBody, headerParams, queryParams, formParams, fileName, fileBytes)

// 	switch strings.ToUpper(method) {
// 	case "GET":
// 		response, err := request.Get(path)
// 		return response, err
// 	case "POST":
// 		response, err := request.Post(path)
// 		return response, err
// 	case "PUT":
// 		response, err := request.Put(path)
// 		return response, err
// 	case "PATCH":
// 		response, err := request.Patch(path)
// 		return response, err
// 	case "DELETE":
// 		response, err := request.Delete(path)
// 		return response, err
// 	}

// 	return nil, fmt.Errorf("invalid method %v", method)
// }

// func prepareRequest(postBody interface{},
// 	headerParams map[string]string,
// 	queryParams url.Values,
// 	formParams map[string]string,
// 	fileName string,
// 	fileBytes []byte) *resty.Request {

// 	request := resty.R()
// 	request.SetBody(postBody)

// 	// add header parameter, if any
// 	if len(headerParams) > 0 {
// 		request.SetHeaders(headerParams)
// 	}

// 	// add query parameter, if any
// 	if len(queryParams) > 0 {
// 		request.SetMultiValueQueryParams(queryParams)
// 	}

// 	// add form parameter, if any
// 	if len(formParams) > 0 {
// 		request.SetFormData(formParams)
// 	}

// 	if len(fileBytes) > 0 && fileName != "" {
// 		_, fileNm := filepath.Split(fileName)
// 		request.SetFileReader("file", fileNm, bytes.NewReader(fileBytes))
// 	}
// 	return request
// }
