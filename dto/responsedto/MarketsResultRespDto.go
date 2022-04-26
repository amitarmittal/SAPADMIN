package responsedto

import (
	"Sp/dto/models"
)

type MarketResult struct {
	ProviderId     string  `json:"providerId"` // Dream / BetFair / SportRadar
	SportId        string  `json:"sportId"`    // 1 / 2 / 4
	EventId        string  `json:"eventId"`    // Unique event id
	MarketId       string  `json:"marketId"`   // Unique Id of the market
	Status         string  `json:"status"`     // OPEN / MAPPED / INPROGRESS / SETTLED / VOIDED / CANCELLED / SUSPENDED
	RunnerId       string  `json:"runnerId"`
	RunnerName     string  `json:"runnerName"`
	SessionOutcome float64 `json:"sessionOutcome"` // fancy scrore ex: 45 NO, 46 YES
}

type MarketsResultRespDto struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	MarketResults    []MarketResult `json:"marketsResult"`
}

func GetMarketResult(market models.Market) MarketResult {
	marketResult := MarketResult{}
	marketResult.ProviderId = market.ProviderId
	marketResult.SportId = market.SportId
	marketResult.EventId = market.EventId
	marketResult.MarketId = market.MarketId
	marketResult.Status = market.MarketStatus
	if marketResult.Status == "SETTLED" {
		if len(market.Results) > 0 {
			marketResult.RunnerId = market.Results[len(market.Results)-1].RunnerId
			marketResult.RunnerName = market.Results[len(market.Results)-1].RunnerName
			marketResult.SessionOutcome = market.Results[len(market.Results)-1].SessionOutcome
		}
	}
	return marketResult
}
