package commondto

// SetConfigRespDto represents response body of this API
type SetConfigRespDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	Level            string    `json:"level"`
	Config           ConfigDto `json:"config"`
}
