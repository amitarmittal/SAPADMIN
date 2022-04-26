package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type B2BUserDto struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	UserKey       string             `bson:"user_key"` // OperatorId+UserId
	Version       int64              `bson:"version"`  // increment for every update start from 0
	OperatorId    string             `bson:"operator_id"`
	UserId        string             `bson:"user_id"`
	UserName      string             `bson:"user_name"`
	Balance       float64            `bson:"balance,truncate"`
	Exposure      float64            `bson:"exposure,truncate"`
	Status        string             `bson:"status"`         // active / blocked / deleted
	IsCommission  bool               `bson:"is_commission"`  // Default false, NO commission
	WinCommission float64            `bson:"win_commission"` // if enabled, default (minimum) is 2%
}
