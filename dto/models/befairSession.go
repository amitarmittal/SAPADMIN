package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BetFairSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Token     string             `bson:"token"`      // Session Token
	Product   string             `bson:"product"`    // Product
	Status    string             `bson:"status"`     // Request Status
	Error     string             `bson:"error"`      // Error message in case of failure
	CreatedAt string             `bson:"created_at"` // Date in readable string format in UTC (from UNIX timestamp)
	UpdatedAt string             `bson:"updated_at"` // Date in readable string format in UTC (from UNIX timestamp)
}
