package dto

import operatordto "Sp/dto/operator"

type AddPartnerReqDto struct {
	OperatorId string              `json:"operatorId"`
	Partner    operatordto.Partner `json:"partner"`
}
