package dto

import operatordto "Sp/dto/operator"

type GetPartnerIdsRespDto struct {
	Status           string                `json:"status"`
	ErrorDescription string                `json:"errorDescription"`
	PartnerIds       []operatordto.Partner `json:"partnerIds"`
}
