package dto

import (
	"Sp/dto/models"
)

type GetProviderStatusRespDto struct {
	Status           string                 `json:"status"`
	ErrorDescription string                 `json:"errorDescription"`
	PartnerStatus    []models.PartnerStatus `json:"providers"`
}
