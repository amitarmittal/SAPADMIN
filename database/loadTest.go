package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get LoadTest document
func GetLoadTest(operatorId string) (models.LoadTest, error) {
	//log.Println("GetLoadTest: Looking for UserKey - ", operationKey)
	result := models.LoadTest{}
	err := LoadTestCollection.FindOne(Ctx, bson.M{"operator_id": operatorId}).Decode(&result)
	if err != nil {
		log.Println("GetLoadTest: Failed with error - ", err.Error())
		return result, err
	}
	return result, nil
}

// Insert user Document
func InsertLoadTest(loadTest models.LoadTest) error {
	//log.Println("InsertLoadTest: Adding Documnet for UserKey - ", loadTest.UserKey)
	loadTest.CreatedAt = time.Now().Format(time.RFC3339Nano)
	loadTest.UpdatedAt = loadTest.CreatedAt
	result, err := LoadTestCollection.InsertOne(Ctx, loadTest)
	if err != nil {
		log.Println("InsertLoadTest: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertLoadTest: Document _id is - ", result.InsertedID)
	return nil
}

// Update user Document
func UpdateLoadTest(operatorId string, message string) error {
	//log.Println("UpdateLoadTestStatus: Updating Documnet for UserKey - ", operationKey)
	updatedAt := time.Now().Format(time.RFC3339Nano)
	opts := options.Update()
	filter := bson.D{{"operator_id", operatorId}}
	update := bson.D{{"$set", bson.D{{"message", message}, {"updated_at", updatedAt}}}}
	_, err := LoadTestCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateLoadTestStatus: FAILED to UPDATE - ", err.Error())
		return err
	}
	//log.Println("UpdateLoadTestStatus: Matched recoreds Count - ", result.MatchedCount)
	//log.Println("UpdateLoadTestStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}
