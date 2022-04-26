package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty"`
	EventKey        string              `bson:"event_key"`        // ProviderId+"-"+SportId+"-"+EventId
	ProviderId      string              `bson:"provider_id"`      // Dream / BetFair / SportRadar
	ProviderName    string              `bson:"provider_name"`    // Dream Sports / Bet Fair / Sport Radar
	SportId         string              `bson:"sport_id"`         // 1 / 2 / 4
	SportName       string              `bson:"sport_name"`       // Soccer / Tennis / Cricket
	CompetitionId   string              `bson:"competition_id"`   //
	CompetitionName string              `bson:"competition_name"` // IPL / BBL / IC T20 WC
	EventId         string              `bson:"event_id"`         //
	EventName       string              `bson:"event_name"`       // IPL / BBL / IC T20 WC
	OpenDate        int64               `bson:"open_date"`        // Open date for the Event
	Status          string              `bson:"status"`           // ACTIVE / BLOCKED
	EventStatus     string              `bson:"event_status"`     // CLOSED / OPEN / MAPPED / INPROGRESS / SETTLED / VOIDED / CANCELLED / SUSPENDED
	Favourite       bool                `bson:"favourite"`        // False - non-favourite. Default is False
	CreatedAt       int64               `bson:"created_at"`       // DateTime in Unix seconds
	UpdatedAt       int64               `bson:"updated_at"`       // DateTime in Unix seconds
	Config          commondto.ConfigDto `bson:"config"`
}
