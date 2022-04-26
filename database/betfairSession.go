package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get document
func GetSessionByProduct(product string) (models.BetFairSession, error) {
	result := models.BetFairSession{}
	// 1. Create Find Filter
	filter := bson.M{}
	filter["product"] = product
	// 2. Create Find Options
	findOptions := options.FindOne()
	findOptions.SetSort(bson.D{{"_id", -1}})
	// 3. Execute Query
	err := BFSessionCollection.FindOne(Ctx, filter, findOptions).Decode(&result)
	if err != nil {
		log.Println("BetFairModule: GetSessionByProduct: Failed with error - ", err.Error())
		return result, err
	}
	// 4. Return result
	return result, nil
}

// Insert Document
func InsertSession(session models.BetFairSession) error {
	session.CreatedAt = time.Now().Format(time.RFC3339Nano)
	session.UpdatedAt = session.CreatedAt
	result, err := BFSessionCollection.InsertOne(Ctx, session)
	if err != nil {
		log.Println("BetFairModule: InsertSession: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("BetFairModule: InsertSession: Document _id is - ", result.InsertedID)
	return nil
}

// Update Document
func UpdateSession(session models.BetFairSession) error {
	session.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	log.Println("BetFairModule: UpdateSession: Document _id is - ", session.ID, session.Token, session.UpdatedAt)
	opts := options.Replace()
	filter := bson.D{{"token", session.Token}}
	result, err := BFSessionCollection.ReplaceOne(Ctx, filter, session, opts)
	if err != nil {
		log.Println("BetFairModule: UpdateSession: FAILED to UPDATE - ", err.Error())
		return err
	}
	log.Println("BetFairModule: UpdateSession: Matched recoreds Count - ", result.MatchedCount)
	log.Println("BetFairModule: UpdateSession: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}
