package dto

type SportsCancelReqDto struct {
	Token         string `json:"token"`
	OperatorId    string `json:"operatorId"`
	ProviderId    string `json:"providerId"` // DREAM/BET_FAIR/SPORTS_RADAR
	SportId       string `json:"sportId"`    // CRICKET/SOCCER/TENNIS
	CompetetionId string `json:"competetionId"`
	EventId       string `json:"eventId"`
	BetId         string `json:"marketId"`
}
