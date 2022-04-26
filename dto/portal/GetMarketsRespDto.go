package dto

import (
	"Sp/dto/models"
)

type GetMarketsRespDto struct {
	Status           string          `json:"status"`
	ErrorDescription string          `json:"errorDescription"`
	Markets          []models.Market `json:"markets"`
}
