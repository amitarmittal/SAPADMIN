package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MarketStatus struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty"`
	MarketKey       string              `bson:"market_key"`       // OperatorId+"-"+ProviderId+"-"+SportId+"-"+EventId+"-"+MarketId
	EventKey        string              `bson:"event_key"`        // OperatorId+"-"+ProviderId+"-"+SportId+"-"+EventId
	OperatorId      string              `bson:"operator_id"`      // Operator Id for that operator
	OperatorName    string              `bson:"operator_name"`    // Operator Name for that operator
	ProviderId      string              `bson:"provider_id"`      // Dream / BetFair / SportRadar
	ProviderName    string              `bson:"provider_name"`    // Dream Sports / Bet Fair / Sport Radar
	SportId         string              `bson:"sport_id"`         // 1 / 2 / 4
	SportName       string              `bson:"sport_name"`       // Soccer / Tennis / Cricket
	CompetitionId   string              `bson:"competition_id"`   //
	CompetitionName string              `bson:"competition_name"` // IPL / BBL / ICC T20 WC
	EventId         string              `bson:"event_id"`         //
	EventName       string              `bson:"event_name"`       // IPL / BBL / ICC T20 WC
	MarketId        string              `bson:"market_id"`        // Unique Id of the market
	MarketName      string              `bson:"market_name"`      // Name of Market
	MarketType      string              `bson:"market_type"`      // Market Type
	ProviderStatus  string              `bson:"provider_status"`  // Provider swtich to block to a particular operator, Default BLOCKED
	OperatorStatus  string              `bson:"operator_status"`  // Operator swatich to block for themselve, Default BLOCKED
	Favourite       bool                `bson:"favourite"`        // False - non-favourite. Default is False
	CreatedAt       int64               `bson:"created_at"`       // DateTime in Unix seconds
	UpdatedAt       int64               `bson:"updated_at"`       // DateTime in Unix Seconds
	Config          commondto.ConfigDto `bson:"config"`           // Configuration for the event
	IsCommission    bool                `bson:"is_commission"`    // Default false, NO commission
	WinCommission   float64             `bson:"win_commission"`   // if enabled, default (minimum) is 2%
}
