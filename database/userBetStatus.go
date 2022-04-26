package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Upsert One UserBetStatus Document
func UpsertUserBetStatus(userBetStatus models.UserBetStatusDto) error {
	// 1. Create Find filter
	filter := bson.M{}
	filter["user_key"] = userBetStatus.UserKey
	//update := bson.D{{"$set", bson.D{{"status", userBetStatus.Status}}}, {"$set", bson.D{{"request_time", userBetStatus.ReqTime}}}}
	update := bson.M{"$set": bson.M{"status": userBetStatus.Status, "request_time": userBetStatus.ReqTime, "reference_id": userBetStatus.ReferenceId, "error_message": userBetStatus.ErrorMessage}}
	opts := options.Update().SetUpsert(true)
	// 2. Execute Query
	result, err := UserBetStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpsertUserBetStatus: UserBetStatus failed with error : ", err.Error())
		log.Println("UpsertUserBetStatus: UserBetStatus failed for user_key : ", userBetStatus.UserKey)
		return err
	}
	if result.MatchedCount == 0 {
		log.Println("UpsertUserBetStatus: Inserted new document for user_key : ", userBetStatus.UserKey)
		log.Println("UpsertUserBetStatus: Inserted new document with _id : ", result.UpsertedID)
	}
	return nil
}

func GetUserBetStatus(operatorId string, userId string) (models.UserBetStatusDto, error) {
	// 0. Prepare required objects
	userBetStatus := models.UserBetStatusDto{}
	// 1. Create Find filter
	userKey := operatorId + "-" + userId
	filter := bson.M{}
	filter["user_key"] = userKey
	// 2. Execute Query
	err := UserBetStatusCollection.FindOne(Ctx, filter).Decode(&userBetStatus)
	if err != nil {
		log.Println("GetUserBetStatus: UserBetStatus failed with error : ", err.Error())
		log.Println("GetUserBetStatus: UserBetStatus failed for user_key : ", userKey)
		return userBetStatus, err
	}
	if userBetStatus.Status == "PENDING" || userBetStatus.Status == "FAILED" {
		betTime := time.Unix(userBetStatus.ReqTime/1000, 0)
		betTime = betTime.Add(time.Second * 60)
		curTime := time.Now()
		//log.Println("GetUserBetStatus: betTime is : ", betTime.String())
		//log.Println("GetUserBetStatus: curTime is : ", curTime.String())
		if curTime.After(betTime) {
			userBetStatus.Status = "EXPIRED"
			userBetStatus.ErrorMessage = "Timeout"
		}
	}
	return userBetStatus, nil
}
