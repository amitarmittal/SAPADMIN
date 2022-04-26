package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type LoadTest struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	OperatorId string             `bson:"operator_id"` // Unique ID of an operation
	Message    string             `bson:"message"`     // Any test message
	CreatedAt  string             `bson:"created_at"`  // Date in readable string format in UTC (from UNIX timestamp)
	UpdatedAt  string             `bson:"updated_at"`  // Date in readable string format in UTC (from UNIX timestamp)
}
