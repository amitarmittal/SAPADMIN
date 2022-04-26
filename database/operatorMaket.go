package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get OperatorMarket Document
// func GetOperatorMarket(operatorMarketKey string) (models.OperatorMarket, error) {
// 	operatorMarket := models.OperatorMarket{}
// 	err := OperatorMarketCollection.FindOne(Ctx, bson.M{"operator_market_key": operatorMarketKey}).Decode(&operatorMarket)
// 	if err != nil {
// 		//log.Println("GetMarket: operatorMarketKey - "+operatorMarketKey+" failed with error - ", err.Error())
// 		return operatorMarket, err
// 	}
// 	return operatorMarket, nil
// }

// Get OperatorMarket Documents by operatorMarketKeys
func GetOperatorMarkets(operatorMarketKeys []string) ([]models.OperatorMarket, error) {
	operatorMarkets := []models.OperatorMarket{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_market_key"] = bson.M{"$in": operatorMarketKeys}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	// findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := OperatorMarketCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("MarketCommissionBL: GetOperatorMarkets: Failed with error - ", err.Error())
		return operatorMarkets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		opMarket := models.OperatorMarket{}
		err = cursor.Decode(&opMarket)
		if err != nil {
			log.Println("MarketCommissionBL: GetOperatorMarkets: Decode failed with error - ", err.Error())
			continue
		}
		operatorMarkets = append(operatorMarkets, opMarket)
	}
	return operatorMarkets, nil
}

// // Insert OperatorMarket Document
// func InsertOperatorMarket(operatorMarket models.OperatorMarket) error {
// 	operatorMarket.CreatedAt = time.Now().Format(time.RFC3339Nano)
// 	operatorMarket.UpdatedAt = operatorMarket.CreatedAt
// 	_, err := OperatorMarketCollection.InsertOne(Ctx, operatorMarket)
// 	if err != nil {
// 		log.Println("InsertMarket: FAILED to INSERT OperatorMarket details - ", err.Error())
// 		return err
// 	}
// 	return nil
// }

// Update OperatorMarket Document
func ReplaceOperatorMarket(operatorMarket models.OperatorMarket) (int64, int64, int64, error) {
	version := operatorMarket.Version
	operatorMarket.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	filter := bson.D{{"operator_market_key", operatorMarket.OperatorMarketKey}, {"version", version}}
	isUpsert := true
	replaceOptions := options.ReplaceOptions{}
	replaceOptions.Upsert = &isUpsert
	operatorMarket.Version++
	result, err := OperatorMarketCollection.ReplaceOne(Ctx, filter, operatorMarket, &replaceOptions)
	if err != nil {
		log.Println("MarketCommissionBL: ReplaceOperatorMarket: FAILED to UPDATE OperatorMarket details - ", err.Error())
		return 0, 0, 0, err
	}
	log.Println("MarketCommissionBL: ReplaceOperatorMarket: Matched recoreds Count - ", result.MatchedCount)
	log.Println("MarketCommissionBL: ReplaceOperatorMarket: Modified recoreds Count - ", result.ModifiedCount)
	log.Println("MarketCommissionBL: ReplaceOperatorMarket: Upserted recoreds Count - ", result.UpsertedCount)
	return result.MatchedCount, result.MatchedCount, result.UpsertedCount, nil
}
