package dto

type UserDto struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
}

type ListUsersRespDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	Users            []UserDto `json:"users"`
}
