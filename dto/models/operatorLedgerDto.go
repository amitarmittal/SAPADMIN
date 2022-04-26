package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type OperatorLedgerDto struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	OperatorId      string             `bson:"operator_id"`
	OperatorName    string             `bson:"operator_name"`
	TransactionType string             `bson:"transaction_type"` // DEPOSIT / WITHDRAW / BETPLACEMENT / BETRESULT / BETROLLBACK
	TransactionTime int64              `bson:"transaction_time"`
	ReferenceId     string             `bson:"reference_id"` // only applicable for BET / RESULT / ROLLBACK
	Amount          float64            `bson:"amount,truncate"`
}
