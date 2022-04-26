package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Competition struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty"`
	CompetitionKey  string              `bson:"competition_key"`  // ProviderId+"-"+SportId+"-"+CompetitionId
	ProviderId      string              `bson:"provider_id"`      // Dream / BetFair / SportRadar
	ProviderName    string              `bson:"provider_name"`    // Dream Sports / Bet Fair / Sport Radar
	SportId         string              `bson:"sport_id"`         // 1 / 2 / 4
	SportName       string              `bson:"sport_name"`       // Soccer / Tennis / Cricket
	CompetitionId   string              `bson:"competition_id"`   //
	CompetitionName string              `bson:"competition_name"` // IPL / BBL / IC T20 WC
	Favourite       bool                `bson:"favourite"`        // False - non-favourite. Default is False
	Status          string              `bson:"status"`           // ACTIVE / BLOCKED
	CreatedAt       int64               `bson:"created_at"`       // DateTime in Unix seconds
	UpdatedAt       int64               `bson:"updated_at"`       // DateTime in Unix seconds
	Config          commondto.ConfigDto `bson:"config"`           //
}
