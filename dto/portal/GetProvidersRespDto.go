package dto

import (
	"Sp/dto/models"
)

type GetProvidersRespDto struct {
	Status           string            `json:"status"`
	ErrorDescription string            `json:"errorDescription"`
	Providers        []models.Provider `json:"providers"`
}
