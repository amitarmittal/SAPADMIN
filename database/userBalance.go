package database

import (
	"Sp/dto/models"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

// Get UserBalance
func GetUserBalance(userKey string) (models.UserBalanceDto, error) {
	result := models.UserBalanceDto{}
	err := UserBalance.FindOne(Ctx, bson.M{"user_key": userKey}).Decode(&result)
	if err != nil {
		log.Println("GetUserBalance: Failed with error - ", err.Error())
		return result, err
	}
	return result, nil
}

func GetUserBalancesForOperator(operatorId string) ([]models.UserBalanceDto, error) {
	results := []models.UserBalanceDto{}
	cursor, err := UserBalance.Find(Ctx, bson.M{"operator_id": operatorId})
	if err != nil {
		log.Println("GetUserBalancesForOperator: Failed with error - ", err.Error())
		return results, err
	}
	defer cursor.Close(Ctx)
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	return results, nil
}

func UpdateUserBalance(userKey string, balance float64, lastSyncTime int64) error {
	_, err := UserBalance.UpdateOne(Ctx, bson.M{"user_key": userKey}, bson.M{"$set": bson.M{"balance": balance, "last_sync_time": lastSyncTime}})
	if err != nil {
		log.Println("UpdateUserBalance: Failed with error - ", err.Error())
		return err
	}
	return nil
}

func InsertUserBalance(userKey string, balance float64) error {
	_, err := UserBalance.InsertOne(Ctx, bson.M{"user_key": userKey, "balance": balance})
	if err != nil {
		log.Println("InsertUserBalance: Failed with error - ", err.Error())
		return err
	}
	return nil
}

func DeleteUserBalance(userKey string) error {
	_, err := UserBalance.DeleteOne(Ctx, bson.M{"user_key": userKey})
	if err != nil {
		log.Println("DeleteUserBalance: Failed with error - ", err.Error())
		return err
	}
	return nil
}

func DeleteUserBalancesForOperator(operatorId string) error {
	_, err := UserBalance.DeleteMany(Ctx, bson.M{"operator_id": operatorId})
	if err != nil {
		log.Println("DeleteUserBalancesForOperator: Failed with error - ", err.Error())
		return err
	}
	return nil
}
