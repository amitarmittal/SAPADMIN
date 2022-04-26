package dto

import (
	"Sp/dto/models"
)

type GetMarketStatusRespDto struct {
	Status           string                `json:"status"`
	ErrorDescription string                `json:"errorDescription"`
	Markets          []models.MarketStatus `json:"marketStatus"`
}
