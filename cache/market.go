package cache

import (
	"Sp/database"
	"Sp/dto/models"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

// 	[ProviderId]								mapKey 1, mapValue1
//		[SportId]								mapKey 2, mapValue2		=> Level 2 Data
//			[EventId-MarketId] -> Market		mapKey 3, mapValue3		=> Level 3 Data

// Level 1:
// mapKey1: ProviderId
// mapValue1: Map[SportId]mapValue2

// Level 2:
// mapKey2: SportId
// mapValue2: Map[EventId-MarketId]mapValue3

// Level 3:
// mapKey3: EventId-MarketId
// mapValue3: models.Market

var marketCache *ristretto.Cache

func init() {
	marketCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save Market in Cache
func SetMarket(market models.Market) {
	marketKey := market.ProviderId + "-" + market.SportId + "-" + market.EventId + "-" + market.MarketId
	// Level 1 - ProviderId
	mapValue1, isFound := marketCache.Get(market.ProviderId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, add to the cache
		log.Println("marketCache: (SET) Provider NOT FOUND in Cache - " + marketKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Market) // initializing map[EventId-MarketId]market
		level3Key := market.EventId + "-" + market.MarketId
		level3Data[level3Key] = market

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Market) // initializing map[SportId]level3Data
		level2Data[market.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		marketCache.Set(market.ProviderId, level2Data, 0)
		marketCache.Wait()

		return
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Market)
	level3Data, isFound := level2Data[market.SportId]
	if false == isFound {
		// Leve 2 Key not found
		log.Println("marketCache: (SET) Sport NOT FOUND in Cache - " + marketKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Market) // initializing map[EventId-MarketId]market
		level3Key := market.EventId + "-" + market.MarketId
		level3Data[level3Key] = market

		// Level 2 Data - add
		level2Data[market.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		marketCache.Set(market.ProviderId, level2Data, 0)
		marketCache.Wait()

		return
	}
	// Level 3 Data -  add or update
	level3Key := market.EventId + "-" + market.MarketId
	level3Data[level3Key] = market

	// Level 2 Data - update
	level2Data[market.SportId] = level3Data

	// Level 1 Data - Update it in Cache
	marketCache.Set(market.ProviderId, level2Data, 0)
	marketCache.Wait()

	return
}

// Get Market from Cache
func GetMarket(providerId string, sportId string, eventId string, marketId string) (models.Market, error) {
	// 0. Create resp object
	market := models.Market{}
	marketKey := providerId + "-" + sportId + "-" + eventId + "-" + marketId

	// Level 1 - ProviderId
	mapValue1, isFound := marketCache.Get(providerId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		log.Println("marketCache: (GET) Provider NOT FOUND in Cache - " + marketKey)

		// Get from DB
		market, err := database.GetMarket(marketKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("marketCache: Market NOT FOUND in DB - ", err.Error())
			log.Println("marketCache: MarketKey is - ", marketKey)
			return market, fmt.Errorf("Market NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Market) // initializing map[MarketId]market
		level3Key := market.EventId + "-" + market.MarketId
		level3Data[level3Key] = market

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Market) // initializing map[SportId]level3Data
		level2Data[market.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		marketCache.Set(market.ProviderId, level2Data, 0)
		marketCache.Wait()

		return market, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Market)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		log.Println("marketCache: (GET) Sport NOT FOUND in Cache - " + marketKey)

		// Get from DB
		market, err := database.GetMarket(marketKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("marketCache: Market NOT FOUND in DB - ", err.Error())
			log.Println("marketCache: MarketKey is - ", marketKey)
			return market, fmt.Errorf("Market NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Market) // initializing map[MarketId]market
		level3Key := market.EventId + "-" + market.MarketId
		level3Data[level3Key] = market

		// Level 2 Data - add
		level2Data[market.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		marketCache.Set(market.ProviderId, level2Data, 0)
		marketCache.Wait()

		return market, nil
	}
	// Level 3 - MarketId
	level3Key := market.EventId + "-" + market.MarketId
	market, isFound = level3Data[level3Key]
	if false == isFound {
		// Level 3 Key not found, Get from DB and add to the cache
		log.Println("marketCache: (GET) Market NOT FOUND in Cache - " + marketKey)

		// Get from DB
		market, err := database.GetMarket(marketKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("marketCache: Market NOT FOUND in DB - ", err.Error())
			log.Println("marketCache: MarketKey is - ", marketKey)
			return market, fmt.Errorf("Market NOT FOUND!")
		}

		// Level 2 Data - add
		level3Key := market.EventId + "-" + market.MarketId
		level3Data[level3Key] = market

		// Level 2 Data - add
		level2Data[market.SportId] = level3Data

		// Level 1 Data - update it in Cache
		marketCache.Set(market.ProviderId, level2Data, 0)
		marketCache.Wait()
		return market, nil
	}
	return market, nil
}

// Get Market from Cache
func GetMarkets(providerId string, sportId string, eventId string) ([]models.Market, error) {
	// 0. Create resp object
	markets := []models.Market{}
	eventKey := providerId + "-" + sportId + "-" + eventId

	// Level 1 - ProviderId
	mapValue1, isFound := marketCache.Get(providerId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		log.Println("marketCache: (GET) Provider NOT FOUND in Cache - " + eventKey)

		// Get from DB
		markets, err := database.GetMarketsByEventKey(eventKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("marketCache: Markets NOT FOUND in DB for eventKey - ", eventKey, err.Error())
			return markets, fmt.Errorf("Markets NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Market) // initializing map[MarketId]market
		for _, market := range markets {
			level3Key := market.EventId + "-" + market.MarketId
			level3Data[level3Key] = market
		}

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Market) // initializing map[SportId]level3Data
		level2Data[sportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		marketCache.Set(providerId, level2Data, 0)
		marketCache.Wait()

		return markets, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Market)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		log.Println("marketCache: (GET) Sport NOT FOUND in Cache - " + eventKey)

		// Get from DB
		markets, err := database.GetMarketsByEventKey(eventKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("marketCache: Markets NOT FOUND in DB for eventKey - ", eventKey, err.Error())
			return markets, fmt.Errorf("Markets NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Market) // initializing map[MarketId]market
		for _, market := range markets {
			level3Key := market.EventId + "-" + market.MarketId
			level3Data[level3Key] = market
		}

		// Level 2 Data - add
		level2Data[sportId] = level3Data

		// Level 1 Data - Update it in Cache
		marketCache.Set(providerId, level2Data, 0)
		marketCache.Wait()

		return markets, nil
	}
	for _, market := range level3Data {
		if market.EventId == eventId {
			markets = append(markets, market)
		}
	}
	return markets, nil
}

/*
var marketCache *ristretto.Cache

func init() {
	marketCache, _ = InitializeCache(100000, 1<<12, 5000)
}

// Save Market in Cache
func SetMarket(market models.Market) {
	// 1. Update it in Cache
	marketCache.Set(market.MarketKey, market, 0)
	marketCache.Wait()
}

// Get Market from Cache
func GetMarket(marketKey string) (models.Market, error) {
	// 0. Create resp object
	market := models.Market{}
	// 1. Get from Cache
	value, found := marketCache.Get(marketKey)
	if found {
		// 1.1 FOUND in cache, retrun object
		market = value.(models.Market)
		return market, nil
	}
	// 2. NOT FOUND in cache, get from DB and update to cache
	market, err := database.GetMarket(marketKey)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		return market, fmt.Errorf("Market NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	SetMarket(market)
	// 4. return object
	return market, nil
}
*/
