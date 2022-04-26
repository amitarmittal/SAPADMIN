package dto

import (
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
)

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type CreateOperatorRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	PublicKey        string `json:"publicKey"`
}

type PortalLoginRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	Token            string `json:"token"`
}

type PortalCreateUserRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	UserDetail       operatordto.PortalUser
}

type PortalGetUserRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	UserDetail       []models.B2BUserDto
}

type PortalGetOperatorRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	UserDetail       []models.B2BUserDto
}
