package dto

import "Sp/dto/models"

type OperatorStatementRespDto struct {
	Status            string                     `json:"status"`
	ErrorDescription  string                     `json:"errorDescription"`
	OperatorStatement []models.OperatorLedgerDto `json:"operatorStatement"`
}
