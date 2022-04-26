package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserBalanceDto struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserKey      string             `bson:"user_key"` // OperatorId+UserId
	OperatorId   string             `bson:"operator_id"`
	UserId       string             `bson:"user_id"`
	Balance      float64            `bson:"balance"`
	LastSyncTime int64              `bson:"last_sync_time"`
}
