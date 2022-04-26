package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Provider struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"`
	ProviderId   string              `bson:"provider_id"`   // Dream / BetFair / SportRadar
	ProviderName string              `bson:"provider_name"` // Dream Sports / Bet Fair / Sport Radar
	Status       string              `bson:"status"`        // ACTIVE / BLOCKED
	ProviderComm float64             `bson:"provider_comm"` // 2% for BetFair
	MinBetValue  float64             `bson:"minbet_value"`
	BetDelay     map[string]int32    `bson:"bet_delay"`
	Balance      float64             `bson:"balance"`
	MinBetSize   int32               `bson:"minimum_betsize"`
	CreatedAt    int64               `bson:"created_at"` // DateTime in Unix seconds
	UpdatedAt    int64               `bson:"updated_at"` // DateTime in Unix seconds
	Config       commondto.ConfigDto `bson:"config"`
}
