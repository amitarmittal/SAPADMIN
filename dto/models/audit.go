package models

import "time"

type Audit struct {
	UserId     string    `bson:"user_id"`     // UserId of the user who performed the action
	Operation  string    `bson:"operation"`   // Operation performed by the user
	IP         string    `bson:"ip"`          // IP address of the user who performed the action
	OperatorId string    `bson:"operator_id"` // OperatorId of the user who performed the action
	Time       time.Time `bson:"time"`        // Time when the action was performed
	ObjectId   string    `bson:"object_id"`   // ObjectId of the object on which the action was performed
	ObjectType string    `bson:"object_type"` // ObjectType of the object on which the action was performed
	Payload    string    `bson:"payload"`     // Payload of the action
	UserRole   string    `bson:"user_role"`   // UserRole of the user who performed the action
}
