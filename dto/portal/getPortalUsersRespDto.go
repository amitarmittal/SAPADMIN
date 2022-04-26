package dto

import operatordto "Sp/dto/operator"

type GetPortalUsersRespDto struct {
	Status           string                   `json:"status"`
	ErrorDescription string                   `json:"errorDescription"`
	PortalUsers      []operatordto.PortalUser `json:"portalUsers"`
}
