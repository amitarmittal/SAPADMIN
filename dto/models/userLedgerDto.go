package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserLedgerDto struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserKey         string             `bson:"user_key"` // OperatorId+UserId
	OperatorId      string             `bson:"operator_id"`
	UserId          string             `bson:"user_id"`
	TransactionType string             `bson:"transaction_type"` // DEPOSIT / WITHDRAW / BETPLACEMENT / BETRESULT / BETROLLBACK
	TransactionTime int64              `bson:"transaction_time"`
	ReferenceId     string             `bson:"reference_id"` // only applicable for BET / RESULT / ROLLBACK
	Amount          float64            `bson:"amount,truncate"`
	Remark          string             `bson:"remark"`
	CompetitionName string             `bson:"competition_name"`
	EventName       string             `bson:"event_name"`
	MarketType      string             `bson:"market_type"`
	MarketName      string             `bson:"market_name"`
	//UserName        string  `bson:"user_name"`
	//Balance         float64 `bson:"balance"`
}
