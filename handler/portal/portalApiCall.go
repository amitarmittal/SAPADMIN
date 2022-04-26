package portalsvc

// func ExternalCall(reqBody []byte, reqType string, url string, reqtimeout time.Duration) ([]byte, error) {
// 	bReqBody := bytes.NewBuffer(reqBody)
// 	req, err := http.NewRequest(reqType, url, bReqBody)
// 	if err != nil {
// 		log.Println("ExternalCall: Failed to create HTTP Req object")
// 		return []byte{}, err
// 	}
// 	req.Header.Add("Content-Type", "application/json")
// 	// 1. Create HTTP Client with Timeout
// 	client := &http.Client{
// 		Timeout: reqtimeout * time.Second,
// 	}
// 	// 2. Make Request
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Println("ExternalCall: Request Failed with error - ", err.Error())
// 		return []byte{}, err
// 	}
// 	defer resp.Body.Close()
// 	// 3. Read response body
// 	respbody, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println("ExternalCall: Response ReadAll failed with error - ", err.Error())
// 		return []byte{}, err
// 	}
// 	return respbody, nil
// }
