package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type Exposure struct {
// 	RunnerId   string  `bson:"runner_id"`
// 	RunnerName string  `bson:"runner_name"`
// 	Exposure   float64 `bson:"exposure"`
// }

type OperatorMarket struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	OperatorMarketKey string             `bson:"operator_market_key"` // OperatorId+"-"+ProviderId+"-"+SportId+"-"+EventId+"-"+MarketId
	Version           int64              `bson:"version"`             // increment for every update start from 0
	OperatorId        string             `bson:"operator_id"`         // Operator Id for that operator
	OperatorName      string             `bson:"operator_name"`       // User Id for that user
	ProviderId        string             `bson:"provider_id"`         // Dream / BetFair / SportRadar
	SportId           string             `bson:"sport_id"`            // 1 / 2 / 4
	CompetitionId     string             `bson:"competition_id"`      //
	EventId           string             `bson:"event_id"`            //
	MarketId          string             `bson:"market_id"`           // Unique Id of the market
	Exposures         []Exposure         `bson:"exposures"`           // Exposrues for all reunners
	Exposure          float64            `bson:"exposure"`            // Final Exposure
	WinnerId          string             `bson:"winner_id"`           // RunnerId
	WinnerName        string             `bson:"winner_name"`         // RunnerName
	WinningAmount     float64            `bson:"winning_amount"`      // Wining Amount
	Commission        float64            `bson:"commission"`          // Commission %
	CommissionAmount  float64            `bson:"commission_amount"`   // Commission Amount
	UsersCommission   float64            `bson:"users_commission"`    // Commission Amount
	CreatedAt         string             `bson:"created_at"`          // Date in readable string format in UTC (from UNIX timestamp)
	UpdatedAt         string             `bson:"updated_at"`          // Date in readable string format in UTC (from UNIX timestamp)
}
