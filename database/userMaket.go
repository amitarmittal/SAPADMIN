package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get UserMarket Document
// func GetUserMarket(userMarketKey string) (models.UserMarket, error) {
// 	userMarket := models.UserMarket{}
// 	err := UserMarketCollection.FindOne(Ctx, bson.M{"user_market_key": userMarketKey}).Decode(&userMarket)
// 	if err != nil {
// 		//log.Println("GetMarket: userMarketKey - "+userMarketKey+" failed with error - ", err.Error())
// 		return userMarket, err
// 	}
// 	return userMarket, nil
// }

// Get UserMarket Documents by userMarketKeys
func GetUserMarkets(userMarketKeys []string) ([]models.UserMarket, error) {
	userMarkets := []models.UserMarket{}
	// 1. Create Filter
	filter := bson.M{}
	filter["user_market_key"] = bson.M{"$in": userMarketKeys}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	// findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := UserMarketCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("MarketCommissionBL: GetUserMarkets: Failed with error - ", err.Error())
		return userMarkets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		opMarket := models.UserMarket{}
		err = cursor.Decode(&opMarket)
		if err != nil {
			log.Println("MarketCommissionBL: GetUserMarkets: Decode failed with error - ", err.Error())
			continue
		}
		userMarkets = append(userMarkets, opMarket)
	}
	return userMarkets, nil
}

// // Insert UserMarket Document
// func InsertUserMarket(userMarket models.UserMarket) error {
// 	userMarket.CreatedAt = time.Now().Format(time.RFC3339Nano)
// 	userMarket.UpdatedAt = userMarket.CreatedAt
// 	_, err := UserMarketCollection.InsertOne(Ctx, userMarket)
// 	if err != nil {
// 		log.Println("InsertMarket: FAILED to INSERT UserMarket details - ", err.Error())
// 		return err
// 	}
// 	return nil
// }

// Update UserMarket Document
func ReplaceUserMarket(userMarket models.UserMarket) (int64, int64, int64, error) {
	version := userMarket.Version
	userMarket.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	filter := bson.D{{"user_market_key", userMarket.UserMarketKey}, {"version", version}}
	isUpsert := true
	replaceOptions := options.ReplaceOptions{}
	replaceOptions.Upsert = &isUpsert
	userMarket.Version++
	result, err := UserMarketCollection.ReplaceOne(Ctx, filter, userMarket, &replaceOptions)
	if err != nil {
		log.Println("MarketCommissionBL: ReplaceUserMarket: FAILED to UPDATE UserMarket details - ", err.Error(), userMarket.UserMarketKey)
		return 0, 0, 0, err
	}
	log.Println("MarketCommissionBL: ReplaceUserMarket: Matched recoreds Count - ", result.MatchedCount, userMarket.UserMarketKey)
	log.Println("MarketCommissionBL: ReplaceUserMarket: Modified recoreds Count - ", result.ModifiedCount, userMarket.UserMarketKey)
	log.Println("MarketCommissionBL: ReplaceUserMarket: Upserted recoreds Count - ", result.UpsertedCount, userMarket.UserMarketKey)
	return result.MatchedCount, result.ModifiedCount, result.UpsertedCount, nil
}
