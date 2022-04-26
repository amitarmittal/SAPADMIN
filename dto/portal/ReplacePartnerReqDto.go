package dto

type ReplacePartnerReqDto struct {
	OperatorId   string `json:"operatorId"`
	OperatorName string `json:"operatorName"`
	Rate         int32  `json:"rate"`
	Status       string `json:"status"`
	WalletType   string `json:"walletType"`
	PartnerId    string `json:"partnerId"`
}
