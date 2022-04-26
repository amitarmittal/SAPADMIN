package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get OperationLock document
func GetOperationLock(operationKey string) (models.OperationLock, error) {
	//log.Println("GetOperationLock: Looking for UserKey - ", operationKey)
	result := models.OperationLock{}
	err := OperationLockCollection.FindOne(Ctx, bson.M{"operation_key": operationKey}).Decode(&result)
	if err != nil {
		log.Println("GetOperationLock: Failed with error - ", err.Error())
		return result, err
	}
	return result, nil
}

// Insert user Document
func InsertOperationLock(operationLock models.OperationLock) error {
	//log.Println("InsertOperationLock: Adding Documnet for UserKey - ", operationLock.UserKey)
	operationLock.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	result, err := OperationLockCollection.InsertOne(Ctx, operationLock)
	if err != nil {
		log.Println("InsertOperationLock: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertOperationLock: Document _id is - ", result.InsertedID)
	return nil
}

// Update user Document
func UpdateOperationLockStatus(operationKey string, status string) error {
	//log.Println("UpdateOperationLockStatus: Updating Documnet for UserKey - ", operationKey)
	updatedAt := time.Now().Format(time.RFC3339Nano)
	opts := options.Update()
	filter := bson.D{{"operation_key", operationKey}}
	update := bson.D{{"$set", bson.D{{"status", status}, {"updated_at", updatedAt}}}}
	_, err := OperationLockCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateOperationLockStatus: FAILED to UPDATE - ", err.Error())
		return err
	}
	//log.Println("UpdateOperationLockStatus: Matched recoreds Count - ", result.MatchedCount)
	//log.Println("UpdateOperationLockStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}
