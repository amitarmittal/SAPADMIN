package requestdto

// GetEventsReqDto represents request body of this API
type GetCacheReq struct {
	CacheType string `json:"cacheType"`
	Key       string `json:"key"`
}
