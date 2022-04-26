package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sport struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"`
	SportKey     string              `bson:"sport_key"`     // ProviderId+"-"+SportId
	ProviderId   string              `bson:"provider_id"`   // Dream / BetFair / SportRadar
	ProviderName string              `bson:"provider_name"` // Dream Sports / Bet Fair / Sport Radar
	SportId      string              `bson:"sport_id"`      // 1 / 2 / 4
	SportName    string              `bson:"sport_name"`    // Soccer / Tennis / Cricket
	Status       string              `bson:"status"`        // ACTIVE / BLOCKED
	Favourite    bool                `bson:"favourite"`     // False - non-favourite. Default is False
	CreatedAt    int64               `bson:"created_at"`    // DateTime in Unix seconds
	UpdatedAt    int64               `bson:"updated_at"`    // DateTime in Unix seconds
	Config       commondto.ConfigDto `bson:"config"`        // Sport Config
}
