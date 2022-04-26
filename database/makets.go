package database

import (
	"Sp/constants"
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get Market Document
func GetMarket(marketKey string) (models.Market, error) {
	market := models.Market{}
	err := MarketCollection.FindOne(Ctx, bson.M{"market_key": marketKey}).Decode(&market)
	if err != nil {
		//log.Println("GetMarket: marketKey - "+marketKey+" failed with error - ", err.Error())
		return market, err
	}
	return market, nil
}

// Get Markets by MarketIds
func GetMarketsByMarketKeys(marketKeys []string) ([]models.Market, error) {
	// 0. Response object
	makertsDto := []models.Market{}
	// 1. Create Query Fileter
	filter := bson.M{"market_key": bson.M{"$in": marketKeys}}
	// 2. Create FindOptions
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetMarketsByMarketKeys: MarketCollection.Find failed with error : ", err.Error())
		return makertsDto, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetMarketsByMarketKeys: Decode failed with error - ", err.Error())
			continue
		}
		makertsDto = append(makertsDto, market)
	}
	return makertsDto, nil
}

// Get Markets by MarketIds
func GetMarketsByMarketIds(marketIds []string) ([]models.Market, error) {
	// 0. Response object
	makertsDto := []models.Market{}
	// 1. Create Query Fileter
	filter := bson.M{"market_id": bson.M{"$in": marketIds}}
	// 2. Create FindOptions
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetMarketsByEventKey: MarketCollection.Find failed with error : ", err.Error())
		return makertsDto, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetMarketsByEventKey: Decode failed with error - ", err.Error())
			continue
		}
		makertsDto = append(makertsDto, market)
	}
	return makertsDto, nil
}

func GetMarketsByEventId(eventId, providerId string) ([]models.Market, error) {
	// 0. Response object
	makertsDto := []models.Market{}
	// 1. Create Query Fileter
	filter := bson.M{"event_id": eventId}
	filter["provider_id"] = providerId
	// 2. Create FindOptions
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetMarketsByEventId: MarketCollection.Find failed with error : ", err.Error())
		return makertsDto, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetMarketsByEventId: Decode failed with error - ", err.Error())
			continue
		}
		makertsDto = append(makertsDto, market)
	}
	return makertsDto, nil
}

// Insert Market Document
func InsertMarket(market models.Market) error {
	market.CreatedAt = time.Now().Unix()
	market.UpdatedAt = market.CreatedAt
	_, err := MarketCollection.InsertOne(Ctx, market)
	if err != nil {
		log.Println("InsertMarket: FAILED to INSERT Market details - ", err.Error())
		return err
	}
	return nil
}

// Bulk Insert Market Staus objects
func InsertManyMarkets(markets []models.Market) error {
	proStatus := []interface{}{}
	for _, proSts := range markets {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	_, err := MarketCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManyMarkets: FAILED to INSERT - ", err.Error())
		return err
	}
	return nil
}

// Update Market MarketStatus
func UpdateMarketMarketStatus(market models.Market) error {
	opts := options.Update()
	filter := bson.D{{"market_key", market.MarketKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"market_status", market.MarketStatus}, {"updated_at", updatedAt}}}}
	_, err := MarketCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateMarketMarketStatus: FAILED to UPDATE market - ", err.Error())
		return err
	}
	return nil
}

// Get Market by EventKey
func GetMarketsByEventKey(eventKey string) ([]models.Market, error) {
	// 0. Response object
	makertsDto := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
	filter["event_key"] = eventKey
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetMarketsByEventKey: Markets NOT FOUND for eventKey : ", eventKey)
		return makertsDto, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetMarketsByEventKey: Decode failed with error - ", err.Error())
			continue
		}
		makertsDto = append(makertsDto, market)
	}
	return makertsDto, nil
}

// Get Market by ProviderId, SportId, CompetitionId and EventId
func GetMarkets(providerId, sportId, competitionId, eventId string) ([]models.Market, error) {
	// 0. Response object
	makertsDto := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
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
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetMarkets: Markets NOT FOUND for ProviderId-SportId-competitionId-eventId : ", providerId+"-"+sportId+"-"+competitionId+"-"+eventId)
		return makertsDto, err
	}

	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetMarkets: Decode failed with error - ", err.Error())
			continue
		}
		makertsDto = append(makertsDto, market)
	}
	return makertsDto, nil
}

func GetLatestMarkets() ([]models.Market, error) {
	// 0. Response object
	markets := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	findOptions.SetLimit(500)
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetLatestMarkets: Markets NOT FOUND!")
		return markets, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetLatestMarkets: Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

func GetUpdatedMarkets() ([]models.Market, error) {
	// 0. Response object
	markets := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	findOptions.SetLimit(1000)
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetUpdatedMarkets: documents NOT FOUND!")
		return markets, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetUpdatedMarkets: Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

func GetAllMarkets() ([]models.Market, error) {
	// 0. Response object
	markets := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetLimit(5000)
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllMarkets: documents NOT FOUND!")
		return markets, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetAllMarkets: Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

func GetAllMarketByProviderId(providerId string) ([]models.Market, error) {
	// 0. Response object
	markets := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetLimit(3000)
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllMarketByProviderId: documents NOT FOUND!")
		return markets, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetAllMarketByProviderId: Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}

// Update Market Document
func ReplaceMarket(marketDto models.Market) error {
	marketDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", marketDto.ID}}
	result, err := MarketCollection.ReplaceOne(Ctx, filter, marketDto)
	if err != nil {
		log.Println("ReplaceMarket: FAILED to UPDATE Market details - ", err.Error())
		return err
	}
	log.Println("ReplaceMarket: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceMarket: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update Many Markets
func UpdateManyMarkets(markets []models.Market) (int, []string) {
	//log.Println("UpdateManyMarkets: Updating user bets for count - ", len(bets))
	msgs := []string{}
	successCount := 0
	// TODO: Find a way to update documents in one DB call
	for _, market := range markets {
		err := ReplaceMarket(market)
		if err != nil {
			log.Println("UpdateManyMarkets: ReplaceMarket failed with erro - ", err.Error())
			// TODO: Handle failures
			msgs = append(msgs, market.MarketId+": "+err.Error())
			continue
		}
		successCount++
	}
	return successCount, msgs
}

// Get Markets updated in last 5 minutes
func GetUpdatedMarketss() ([]models.Market, error) {
	retObjs := []models.Market{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := MarketCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedMarketss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.Market{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedMarketss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

// Get Markets settled in last 20 minutes
func GetLatestSettledMarkets() ([]models.Market, error) {
	retObjs := []models.Market{}
	filter := bson.M{}
	filter["provider_id"] = constants.SAP.ProviderType.BetFair()
	filter["market_status"] = "SETTLED"
	beforeUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$lt": beforeUpdateAt}
	// 2. Create FindOptions
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	findOptions.SetLimit(100)
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUpdatedMarketss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.Market{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedMarketss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

// Get Open Markets by ProviderId
func GetOpenMarketsByProvider(providerId string) ([]models.Market, error) {
	// 0. Response object
	markets := []models.Market{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["market_status"] = "OPEN"
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetLimit(30)
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := MarketCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetOpenMarketsByProvider: documents NOT FOUND!")
		return markets, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		market := models.Market{}
		err := cursor.Decode(&market)
		if err != nil {
			log.Println("GetOpenMarketsByProvider: Decode failed with error - ", err.Error())
			continue
		}
		markets = append(markets, market)
	}
	return markets, nil
}
