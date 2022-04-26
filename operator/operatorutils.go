package operator

import (
	operatordto "Sp/dto/operator"
	keyutils "Sp/utilities"
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func WalletCall(jsonData []byte, priKey rsa.PrivateKey, url string, reqtimeout time.Duration) (operatordto.OperatorRespDto, error) {
	opResponse := operatordto.OperatorRespDto{}
	log.Println("WalletCall: URL is - ", url)
	// 1. Create Signature
	NewSignature, err := keyutils.CreateSignature(string(jsonData), priKey)
	if err != nil {
		log.Println("WalletCall: Create signature failed with error - ", err.Error())
		return opResponse, fmt.Errorf("Internal Error: Create signature failed!")
	}
	log.Println(url)
	//log.Println("WalletCall: Signature is - ", NewSignature)
	// 2. Create HTTP Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("WalletCall: Failed to create HTTP Req object")
		return opResponse, fmt.Errorf("Internal Error: Create http request failed!")
	}
	// 3. Add Headers
	req.Header.Add("Signature", NewSignature)
	req.Header.Add("Content-Type", "application/json")
	// 4. Create HTTP Client with Timeout
	client := &http.Client{
		Timeout: reqtimeout * time.Second,
	}
	// 5. Make Request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("WalletCall: Operator request failed with error - ", err.Error())
		return opResponse, err
	}
	defer resp.Body.Close()
	// 6. Read response body
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("WalletCall: Response read failed with error - ", err.Error())
		return opResponse, err
	}
	log.Println(string(respbody))
	// 7. Convert response body to Operator Response object
	err = json.Unmarshal([]byte(respbody), &opResponse)
	if err != nil {
		log.Println("WalletCall: json.Unmarshal failed with error - ", err.Error())
	}
	return opResponse, err
}

func WalletCall2(jsonData []byte, priKey rsa.PrivateKey, url string, reqtimeout time.Duration) ([]byte, error) {
	// 0. Default response
	respbody := []byte{}
	log.Println("WalletCall: URL is - ", url)
	// 1. Create Signature
	NewSignature, err := keyutils.CreateSignature(string(jsonData), priKey)
	if err != nil {
		log.Println("WalletCall: Create signature failed with error - ", err.Error())
		return respbody, fmt.Errorf("Internal Error: Create signature failed!")
	}
	//log.Println("WalletCall: Signature is - ", NewSignature)
	// 2. Create HTTP Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("WalletCall: Failed to create HTTP Req object")
		log.Println(url)
		log.Println("WalletCall2: request data is - ", string(jsonData))
		return respbody, fmt.Errorf("Internal Error: Create http request failed!")
	}
	// 3. Add Headers
	req.Header.Add("Signature", NewSignature)
	req.Header.Add("Content-Type", "application/json")
	// 4. Create HTTP Client with Timeout
	client := &http.Client{
		Timeout: reqtimeout * time.Second,
	}
	// 5. Make Request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("WalletCall: Operator request failed with error - ", err.Error())
		return respbody, err
	}
	defer resp.Body.Close()
	// 6. Read response body
	respbody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("WalletCall: Response read failed with error - ", err.Error())
		return respbody, err
	}
	//log.Println(string(respbody))
	return respbody, nil
}
