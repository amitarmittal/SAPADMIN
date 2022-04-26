package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert EventStatus Document
func InsertMarketStatus(marketStatus models.MarketStatus) error {
	//log.Println("InsertEventStatus: Adding Documnet for eventId - ", eventDto.EventId)
	marketStatus.CreatedAt = time.Now().Unix()
	marketStatus.UpdatedAt = marketStatus.CreatedAt
	result, err := MarketStatusCollection.InsertOne(Ctx, marketStatus)
	if err != nil {
		log.Println("MarketStatusCollection: InsertMarketStatus: FAILED to INSERT eventStatus - ", err.Error())
		return err
	}
	log.Println("MarketStatusCollection: InsertMarketStatus: Event Document _id is - ", result.InsertedID)
	return nil
}

// Get MarketStatus By MarketKey
func GetMarketStatus(marketKey string) (models.MarketStatus, error) {
	//log.Println("GetMarketStatus: Looking for MarketKey - ", marketKey)
	marketStatus := models.MarketStatus{}
	err := MarketStatusCollection.FindOne(Ctx, bson.M{"market_key": marketKey}).Decode(&marketStatus)
	if err != nil {
		log.Println("GetMarketStatus: Failed with error - ", err.Error())
		return marketStatus, err
	}
	return marketStatus, nil
}

func GetMarketStatuses(marketKeys []string) ([]models.MarketStatus, error) {
	//log.Println("GetMarketStatus: Looking for MarketKey - ", marketKey)
	marketStatus := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["market_key"] = bson.M{"$in": marketKeys}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetMarketStatuses: Failed with error - ", err.Error())
		return marketStatus, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		ms := models.MarketStatus{}
		err = cursor.Decode(&ms)
		if err != nil {
			log.Println("GetMarketStatuses: Decode failed with error - ", err.Error())
			continue
		}
		marketStatus = append(marketStatus, ms)
	}
	return marketStatus, nil
}

func GetMarketStatusesByMarket(providerId string, sportId string, eventId string, marketId string) ([]models.MarketStatus, error) {
	//log.Println("GetMarketStatus: Looking for MarketKey - ", marketKey)
	marketStatus := []models.MarketStatus{}
	inputKey := providerId + "-" + sportId + "-" + eventId + "-" + marketId
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	filter["event_id"] = eventId
	filter["market_id"] = marketId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetMarketStatusesByMarket: Failed with error - ", err.Error())
		log.Println("GetMarketStatusesByMarket: Failed for inputKey - ", inputKey)
		return marketStatus, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		ms := models.MarketStatus{}
		err = cursor.Decode(&ms)
		if err != nil {
			log.Println("GetMarketStatusesByMarket: Decode failed with error - ", err.Error())
			continue
		}
		marketStatus = append(marketStatus, ms)
	}
	return marketStatus, nil
}

/*
// Get Markets By OperatorId
func GetLatestMarketStatus() ([]models.MarketStatus, error) {
	//log.Println("GetLatestMarketStatus: Trying to fetch 500 documents sort by updatedAt DESC")
	// 0. Response object
	markets := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	findOptions.SetLimit(500)
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetLatestMarketStatus: Failed with error - ", err.Error())
		return markets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	// err = cursor.All(Ctx, &markets)
	for cursor.Next(Ctx) {
		market := models.MarketStatus{}
		err = cursor.Decode(&market)
		if err != nil {
			log.Println("GetLatestMarketStatus: MarketStatus Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

// Get Markets By OperatorId
func GetOpMarkets(operatorId string) ([]models.MarketStatus, error) {
	//log.Println("GetOpMarkets: Looking for OperatorId - ", operatorId)
	// 0. Response object
	markets := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_id"] = operatorId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOpMarkets: Failed with error - ", err.Error())
		return markets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.MarketStatus{}
		err = cursor.Decode(&market)
		if err != nil {
			log.Println("GetOpMarkets: MarketStatus Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

// Get Markets By ProviderId
func GetPrMarkets(providerId string) ([]models.MarketStatus, error) {
	//log.Println("GetPrMarkets: Looking for ProviderId - ", providerId)
	// 0. Response object
	sports := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetPrMarkets: Failed with error - ", err.Error())
		return sports, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.MarketStatus{}
		err = cursor.Decode(&sport)
		if err != nil {
			log.Println("GetPrMarkets: MarketStatus Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// Get Markets By OperatorId & ProviderId & SportId
func GetOpPrMarkets(operatorId string, providerId string, sportId string, competitionId string, eventId string) ([]models.MarketStatus, error) {
	//log.Println("GetOpPrMarkets: Looking for OperatorId - ", operatorId)
	// 0. Response object
	markets := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_id"] = operatorId
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	if competitionId != "" {
		filter["competition_id"] = competitionId
	}
	if eventId != "" {
		filter["event_id"] = eventId
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOpPrMarkets: Failed with error - ", err.Error())
		return markets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.MarketStatus{}
		err = cursor.Decode(&market)
		if err != nil {
			log.Println("GetOpPrMarkets: MarketStatus Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

func GetOperatorsFromMrPr(marketId string, providerId string) ([]models.MarketStatus, error) {
	// 0. Response object
	markets := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["market_id"] = marketId
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOperatorsFromMrPr: Failed with error - ", err.Error())
		return markets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.MarketStatus{}
		err = cursor.Decode(&market)
		if err != nil {
			log.Println("GetOperatorsFromMrPr: MarketStatus Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}
*/

// Bulk insert Market Staus objects
// Use Cases:
// 1. When added new Operator
// 2. When added new Market
func InsertManyMarketStatus(marketStatus []models.MarketStatus) error {
	//log.Println("InsertManyMarketStatus: Adding MarketStatus, count is - ", len(marketStatus))
	proStatus := []interface{}{}
	for _, proSts := range marketStatus {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	result, err := MarketStatusCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManyMarketStatus: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertManyMarketStatus: Inserted count is - ", len(result.InsertedIDs))
	return nil
}

// Update Operator Status
func UpdateOAMarketStatus(marketKey string, status string) error {
	//if status == "ACTIVE" {
	//	log.Println("UpdateOAMarketStatus: Unblocking the market - ", marketKey)
	//} else {
	//	log.Println("UpdateOAMarketStatus: Status changing for the marketKey: ", marketKey+"-"+status)
	//}
	filter := bson.D{{"market_key", marketKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"operator_status", status}, {"updated_at", updatedAt}}}}
	opts := options.Update()
	result, err := MarketStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateOAMarketStatus: FAILED to UPDATE operator status - ", err.Error())
		return err
	}
	log.Println("UpdateOAMarketStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateOAMarketStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update Provider Status
func UpdatePAMarketStatus(marketKey string, status string) error {
	//if status == "ACTIVE" {
	//	log.Println("UpdatePAMarketStatus: Unblocking the market - ", marketKey)
	//} else {
	//	log.Println("UpdatePAMarketStatus: Status changing for the marketKey: ", marketKey+"-"+status)
	//}
	filter := bson.D{{"market_key", marketKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"provider_status", status}, {"updated_at", updatedAt}}}}
	opts := options.Update()
	result, err := MarketStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdatePAMarketStatus: FAILED to UPDATE provider status - ", err.Error())
		return err
	}
	log.Println("UpdatePAMarketStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdatePAMarketStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func GetUpdatedMarketStatus(marketKeys []string) ([]models.MarketStatus, error) {
	//log.Println("GetUpdatedMarketStatus: Looking for Keys Count - ", len(marketKeys))
	// 0. Response object
	marketStatuses := []models.MarketStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	if len(marketKeys) != 0 {
		filter["market_key"] = bson.M{"$in": marketKeys}
	} else {
		findOptions.SetLimit(1000)
	}
	// 3. Execute Query
	cursor, err := MarketStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUpdatedMarketStatus: Failed with error - ", err.Error())
		return marketStatuses, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		marketStatus := models.MarketStatus{}
		err = cursor.Decode(&marketStatus)
		if err != nil {
			log.Println("GetUpdatedMarketStatus: Decode failed with error - ", err.Error())
			continue
		}
		marketStatuses = append(marketStatuses, marketStatus)
	}
	return marketStatuses, nil
}

// Update Market Document
func ReplaceMarketStatus(marketDto models.MarketStatus) error {
	marketDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", marketDto.ID}}
	// update := bson.D{{"$set", operatorDto}}
	result, err := MarketStatusCollection.ReplaceOne(Ctx, filter, marketDto)
	if err != nil {
		log.Println("ReplaceMarketStatus: FAILED to UPDATE Market details - ", err.Error())
		return err
	}
	log.Println("ReplaceMarketStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceMarketStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func DeleteMarketsByOperator(operatorId string) error {
	filter := bson.D{{"operator_id", operatorId}}
	result, err := MarketStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteMarketsByOperator: FAILED to DELETE Market details - ", err.Error())
		return err
	}
	log.Println("DeleteMarketsByOperator: Matched recoreds Count - ", result.DeletedCount)
	return nil
}

// Get MarketStatuss updated in last 5 minutes
func GetUpdatedMarketStatuss() ([]models.MarketStatus, error) {
	retObjs := []models.MarketStatus{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := MarketStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedMarketStatuss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.MarketStatus{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedMarketStatuss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}
