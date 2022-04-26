package responsedto

// GetEventsRespDto represents response body of this API
type GetCacheResp struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	RespJson         string `json:"respJson"`
}
