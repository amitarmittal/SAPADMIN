package dto

type ResetPasswordForOPReq struct {
	UserId     string `json:"userId"`
	OperatorId string `json:"operatorId"`
	Password   string `json:"newPassword"`
}
