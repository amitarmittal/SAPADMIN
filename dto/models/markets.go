package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rollback struct {
	RollbackType   string `bson:"rollback_type"`
	RollbackReason string `bson:"rollback_reason"`
	RollbackTime   int64  `bson:"rollback_time"`
}

type Result struct {
	RunnerId       string  `bson:"runner_id"`
	RunnerName     string  `bson:"runner_name"`
	SessionOutcome float64 `bson:"session_outcome,truncate"` // fancy scrore ex: 45 NO, 46 YES
	ResultTime     int64   `bson:"result_time"`
}

type Runner struct {
	RunnerId     string `bson:"runner_id"`
	RunnerName   string `bson:"runner_name"`
	RunnerStatus string `bson:"runner_status"`
}
type Market struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty"`
	MarketKey       string              `bson:"market_key"`       // ProviderId+"-"+SportId+"-"+EventId+"-"+MarketId
	EventKey        string              `bson:"event_key"`        // ProviderId+"-"+SportId+"-"+EventId+"-"
	ProviderId      string              `bson:"provider_id"`      // Dream / BetFair / SportRadar
	ProviderName    string              `bson:"provider_name"`    // Dream Sports / Bet Fair / Sport Radar
	SportId         string              `bson:"sport_id"`         // 1 / 2 / 4
	SportName       string              `bson:"sport_name"`       // Soccer / Tennis / Cricket
	CompetitionId   string              `bson:"competition_id"`   // IPL, BBL
	CompetitionName string              `bson:"competition_name"` // IPL, BBL
	EventId         string              `bson:"event_id"`         // Unique event id
	EventName       string              `bson:"event_name"`       // IPL / BBL / IC T20 WC
	MarketId        string              `bson:"market_id"`        // Unique Id of the market
	MarketName      string              `bson:"market_name"`      // Name of Market
	MarketType      string              `bson:"market_type"`      // Market Type
	Category        string              `bson:"category"`         // applicable for Fancy Markets
	Runners         []Runner            `bson:"runners"`
	Status          string              `bson:"status"`     // ACTIVE / BLOCKED
	Favourite       bool                `bson:"favourite"`  // False - non-favourite. Default is False
	CreatedAt       int64               `bson:"created_at"` // DateTime in Unix seconds
	UpdatedAt       int64               `bson:"updated_at"` // DateTime in Unix seconds
	Config          commondto.ConfigDto `bson:"config"`
	MarketStatus    string              `bson:"market_status"` // OPEN / MAPPED / INPROGRESS / SETTLED / VOIDED / CANCELLED / SUSPENDED
	IsSuspended     bool                `bson:"is_suspended"`  // to SUSPEND / RESUME markets.
	Results         []Result            `bson:"results"`
	Rollbacks       []Rollback          `bson:"rollbacks"`
}
