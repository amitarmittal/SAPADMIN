package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type OperationLock struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	OperationKey string             `bson:"operation_key"` // Unique ID of an operation
	Description  string             `bson:"description"`
	Status       string             `bson:"status"`     // FREE / BUSY / HOLD
	UpdatedAt    string             `bson:"updated_at"` // Date in readable string format in UTC (from UNIX timestamp)
}
