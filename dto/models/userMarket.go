package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Exposure struct {
	RunnerId   string  `bson:"runner_id"`
	RunnerName string  `bson:"runner_name"`
	Exposure   float64 `bson:"exposure"`
}

type UserMarket struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UserMarketKey    string             `bson:"user_market_key"`   // OperatorId+"-"+UserId+"-"+ProviderId+"-"+SportId+"-"+EventId+"-"+MarketId
	Version          int64              `bson:"version"`           // increment for every update start from 0
	Token            string             `bson:"token"`             // Users latest token
	OperatorId       string             `bson:"operator_id"`       // Operator Id for that operator
	UserId           string             `bson:"user_id"`           // User Id for that user
	UserName         string             `bson:"user_name"`         // User Id for that user
	ProviderId       string             `bson:"provider_id"`       // Dream / BetFair / SportRadar
	SportId          string             `bson:"sport_id"`          // 1 / 2 / 4
	CompetitionId    string             `bson:"competition_id"`    //
	EventId          string             `bson:"event_id"`          //
	MarketId         string             `bson:"market_id"`         // Unique Id of the market
	Exposures        []Exposure         `bson:"exposures"`         // Exposrues for all reunners
	Exposure         float64            `bson:"exposure"`          // Final Exposure
	Rate             int32              `bson:"rate"`              // Currency Rate to communicate in currency
	WinnerId         string             `bson:"winner_id"`         // RunnerId
	WinnerName       string             `bson:"winner_name"`       // RunnerName
	WinningAmount    float64            `bson:"winning_amount"`    // Wining Amount
	Commission       float64            `bson:"commission"`        // Commission %
	CommLevel        string             `bson:"commission_level"`  // commission configuration level
	CommissionAmount float64            `bson:"commission_amount"` // Commission Amount
	UserCommission   float64            `bson:"user_commission"`   // Commission Amount
	CreatedAt        string             `bson:"created_at"`        // Date in readable string format in UTC (from UNIX timestamp)
	UpdatedAt        string             `bson:"updated_at"`        // Date in readable string format in UTC (from UNIX timestamp)
}
