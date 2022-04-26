package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserBetStatusDto struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserKey      string             `bson:"user_key"` // OperatorId+UserId
	OperatorId   string             `bson:"operator_id"`
	UserId       string             `bson:"user_id"`
	ReqTime      int64              `bson:"request_time"`  // milli seconds
	ReferenceId  string             `bson:"reference_id"`  // BetId if applicable
	Status       string             `bson:"status"`        // PENDING / COMPLETED / CANCELLED / EXPIRED / FAILED
	ErrorMessage string             `bson:"error_message"` // Error message for failed bets
}
