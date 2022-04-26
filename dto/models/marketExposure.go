package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MarketExposure struct {
	ID                primitive.ObjectID   `bson:"_id,omitempty"`
	MarketKey         string               `bson:"market_key"`    // OperatorId+"-"+ProviderId+"-"+SportId+"-"+EventId+"-"+MarketId
	OperatorId        string               `bson:"operator_id"`   // Dream / BetFair / SportRadar
	OperatorName      string               `bson:"operator_name"` // Dream Sports / Bet Fair / Sport Radar
	ProviderId        string               `bson:"provider_id"`   // Dream / BetFair / SportRadar
	SportId           string               `bson:"sport_id"`      // 1 / 2 / 4
	EventId           string               `bson:"event_id"`      // Unique event id
	EventName         string               `bson:"event_name"`    // IPL / BBL / IC T20 WC
	MarketId          string               `bson:"market_id"`     // Unique Id of the market
	MarketName        string               `bson:"market_name"`   // Name of Market
	MarketType        string               `bson:"market_type"`   // Market Type
	Category          string               `bson:"category"`      // applicable for Fancy Markets
	RunnerId          string               `bson:"runner_id"`
	RunnerName        string               `bson:"runner_name"`
	MatchedExposure   primitive.Decimal128 `bson:"matched_exposure"`
	UnmatchedExposure primitive.Decimal128 `bson:"unmatched_exposure"`
	CreatedAt         int64                `bson:"created_at"` // DateTime in Unix seconds
	UpdatedAt         int64                `bson:"updated_at"` // DateTime in Unix seconds
}
