package responsedto

type BasicRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"message"`
}
