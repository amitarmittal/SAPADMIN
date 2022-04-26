package dto

type GetRoleRespDto struct {
	Status           string   `json:"status"`
	ErrorDescription string   `json:"errorDescription"`
	Roles            []string `json:"roles"`
}
