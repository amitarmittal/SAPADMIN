package cache

import (
	"Sp/database"
	"Sp/dto/commondto"
	"Sp/dto/models"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
)

// Key: OperartorId+"-"+ProviderId+"-"+SportId
// Value: Map[EventId-MarketId]MarketStatus

// 	[OperatorId-ProviderId]							mapKey 1, mapValue1
//		[SportId]									mapKey 2, mapValue2		=> Level 2 Data
//			[EventId-MarketId] -> MarketStatus		key 3, value 			=> Level 3 Data

// Level 1:
// mapKey1: OperatorId-ProviderId
// mapValue1: Map[SportId]Level2Data

// Level 2:
// mapKey2: SportId
// mapValue2: Map[EventId-MarketId]Level3Data

// Level 3:
// Key3: EventId-MarketId
// Value3: models.MarketStatus

var marketStatusCache *ristretto.Cache

func init() {
	marketStatusCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save MarketStatus in Cache
func SetMarketStatus(marketStatus models.MarketStatus) {
	marketStatusKey := marketStatus.OperatorId + "-" + marketStatus.ProviderId + "-" + marketStatus.SportId + "-" + marketStatus.EventId + "-" + marketStatus.MarketId
	// Level 1 - OperatorId-ProviderId
	level1Key := marketStatus.OperatorId + "-" + marketStatus.ProviderId
	mapValue1, isFound := marketStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, add to the cache
		log.Println("marketStatusCache: (SET) Operator-Provider NOT FOUND in Cache - " + marketStatusKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.MarketStatus) // initializing map[CompititionId]compitition
		level3Key := marketStatus.EventId + "-" + marketStatus.MarketId
		level3Data[level3Key] = marketStatus

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.MarketStatus) // initializing map[SportId]level3Data
		level2Data[marketStatus.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		marketStatusCache.Set(level1Key, level2Data, 0)
		marketStatusCache.Wait()

		return
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.MarketStatus)
	level3Data, isFound := level2Data[marketStatus.SportId]
	if false == isFound {
		// Leve 2 Key not found
		log.Println("marketStatusCache: (SET) Sport NOT FOUND in Cache - " + marketStatusKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.MarketStatus) // initializing map[CompititionId]compitition
		level3Key := marketStatus.EventId + "-" + marketStatus.MarketId
		level3Data[level3Key] = marketStatus

		// Level 2 Data - add
		level2Data[marketStatus.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		marketStatusCache.Set(level1Key, level2Data, 0)
		marketStatusCache.Wait()

		return
	}
	// Level 3 - CompititionId

	// Level 3 Data -  add or update
	level3Key := marketStatus.EventId + "-" + marketStatus.MarketId
	level3Data[level3Key] = marketStatus

	// Level 2 Data - update
	level2Data[marketStatus.SportId] = level3Data

	// Level 1 Data - Update it in Cache
	marketStatusCache.Set(level1Key, level2Data, 0)
	marketStatusCache.Wait()

	return
}

// Get MarketStatus from Cache
func GetMarketStatus(operatorId string, providerId string, sportId string, eventId string, marketId string) (models.MarketStatus, error) {
	// 0. Create resp object
	marketStatus := models.MarketStatus{}
	// marketStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId + "-" + marketId
	level1Key := operatorId + "-" + providerId
	level3Key := eventId + "-" + marketId

	// Level 1 - ProviderId
	mapValue1, isFound := marketStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
		// log.Println("marketStatusCache: (GET) Provider NOT FOUND in Cache - " + marketStatusKey)

		// // Get from DB
		// marketStatus, err := database.GetMarketStatus(marketStatusKey)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("marketStatusCache: MarketStatus NOT FOUND in DB is - ", marketStatusKey)
		// 	return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data := make(map[string]models.MarketStatus) // initializing map[CompititionId]compitition
		// level3Key = marketStatus.EventId + "-" + marketStatus.MarketId
		// level3Data[level3Key] = marketStatus

		// // Level 2 Data - init & add
		// level2Data := make(map[string]map[string]models.MarketStatus) // initializing map[SportId]level3Data
		// level2Data[marketStatus.SportId] = level3Data

		// // Level 1 Data - Save it in Cache (insert)
		// marketStatusCache.Set(level1Key, level2Data, 0)
		// marketStatusCache.Wait()

		// return marketStatus, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.MarketStatus)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
		// log.Println("marketStatusCache: (GET) Sport NOT FOUND in Cache - " + marketStatusKey)

		// // Get from DB
		// marketStatus, err := database.GetMarketStatus(marketStatusKey)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("marketStatusCache: MarketStatus NOT FOUND in DB is - ", marketStatusKey)
		// 	return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data = make(map[string]models.MarketStatus) // initializing map[CompititionId]compitition
		// level3Key = marketStatus.EventId + "-" + marketStatus.MarketId
		// level3Data[level3Key] = marketStatus

		// // Level 2 Data - add
		// level2Data[marketStatus.SportId] = level3Data

		// // Level 1 Data - Update it in Cache
		// marketStatusCache.Set(level1Key, level2Data, 0)
		// marketStatusCache.Wait()

		// return marketStatus, nil
	}
	// Level 3 - CompititionId
	marketStatus, isFound = level3Data[level3Key]
	if false == isFound {
		// Level 3 Key not found, Get from DB and add to the cache
		return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
		// log.Println("marketStatusCache: (GET) MarketStatus NOT FOUND in Cache - " + marketStatusKey)

		// // Get from DB
		// marketStatus, err := database.GetMarketStatus(marketStatusKey)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("marketStatusCache: MarketStatus NOT FOUND in DB is - ", marketStatusKey)
		// 	return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
		// }

		// // Level 3 Data - add
		// level3Key := marketStatus.EventId + "-" + marketStatus.MarketId
		// level3Data[level3Key] = marketStatus

		// // Level 2 Data - update
		// level2Data[marketStatus.SportId] = level3Data

		// // Level 1 Data - update it in Cache
		// marketStatusCache.Set(level1Key, level2Data, 0)
		// marketStatusCache.Wait()

		// return marketStatus, nil
	}

	return marketStatus, nil
}

// Get Market from Cache
func GetMarketStatuss(operatorId string, providerId string, sportId string, eventId string) ([]models.MarketStatus, error) {
	// 0. Create resp object
	listMSs := []models.MarketStatus{}
	level1Key := operatorId + "-" + providerId

	// Level 1 - ProviderId
	mapValue1, isFound := marketStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		return listMSs, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.MarketStatus)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		return listMSs, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	for _, ms := range level3Data {
		if ms.EventId == eventId {
			listMSs = append(listMSs, ms)
		}
	}
	return listMSs, nil
}

/*
// Key: OperartorId+"-"+ProviderId+"-"+SportId+"-"+EventId
// Value: Map[MarketId]MarketStatus

var marketStatusCache *ristretto.Cache

func init() {
	marketStatusCache, _ = InitializeCache(100000, 1<<12, 5000)
}

// Set MarketStatus by OperatorId, ProviderId, SportId & EventId
func SetOpMarketStatus(operatorId, providerId, sportId, eventId string, mapValue map[string]models.MarketStatus) {
	key := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
	marketStatusCache.Set(key, mapValue, 0)
	marketStatusCache.Wait()
}

// Set MarketStatus
func SetMarketStatus(marketStatus models.MarketStatus) {
	// 0. Define MapValue
	mapValue := make(map[string]models.MarketStatus)
	key := marketStatus.OperatorId + "-" + marketStatus.ProviderId + "-" + marketStatus.SportId + "-" + marketStatus.EventId
	// 1. Get MapValue By OperatorId
	value, found := marketStatusCache.Get(key)
	if found {
		// 1.1. FOUND, assign value
		mapValue = value.(map[string]models.MarketStatus)
	} else {
		// 1.2. NOT FOUND in cache, make new mapvalue
		mapValue = make(map[string]models.MarketStatus)
	}
	// 2. add/update value in mapValue
	mapValue[marketStatus.MarketId] = marketStatus
	// 3. add/update it mapValue in Cache
	marketStatusCache.Set(key, mapValue, 64)
	marketStatusCache.Wait()
}
*/

/*
// Get MarketStatus by OpertorId, ProviderId, SportId, EventId & MarketId
func GetMarketStatus(operatorId string, providerId string, sportId string, eventId string, marketId string) (models.MarketStatus, error) {
	// 0. Create resp object
	ms := models.MarketStatus{}
	mapValue := make(map[string]models.MarketStatus)
	marketKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
	// 1. Get from Cache
	value, found := marketStatusCache.Get(marketKey)
	if found {
		// 1.1. FOUND in cache, retrun object
		mapValue = value.(map[string]models.MarketStatus)
		ms, ok := mapValue[marketId]
		if ok {
			// 1.1.1. FOUND in map
			return ms, nil
		}
		// 1.2. NOT FOUND in map, get from DB
		//log.Println("GetMarketStatus: MarketStatus NOT FOUND in Cache Map - " + operatorId)
		ms, err := database.GetMarketStatus(marketKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("GetMarketStatus: database.GetMarketStatus failed with error - ", err.Error())
			return ms, fmt.Errorf("MarketStatus NOT FOUND!")
		}
		mapValue[marketId] = ms
		SetOpMarketStatus(operatorId, providerId, sportId, eventId, mapValue)
		return ms, nil
	}
	//log.Println("GetMarketStatus: Operator MarketStatus NOT FOUND in Cache - " + operatorId)
	// 2. NOT FOUND in cache, get from DB and update to cache
	marketStatus, err := database.GetOpPrMarkets(operatorId, providerId, sportId, "", eventId)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("GetMarketStatus: Operator MarketStatus NOT FOUND in DB - ", err.Error())
		return ms, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	for _, ms := range marketStatus {
		mapValue[ms.MarketId] = ms
	}
	SetOpMarketStatus(operatorId, providerId, sportId, eventId, mapValue)
	// 4. return object
	ms, ok := mapValue[marketId]
	if !ok {
		// 4.1. NOT FOUND in DB map
		log.Println("GetMarketStatus: MarketStatus NOT FOUND in DB - " + operatorId)
		return ms, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	return ms, nil
}
*/
func AddMarketStatusConfig(operatorId string, providerId string, sportId string, eventId string, marketId string, config commondto.ConfigDto) error {
	market, err := GetMarket(providerId, sportId, eventId, marketId)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	marketStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId + "-" + marketId
	marketStatus := models.MarketStatus{}
	marketStatus.MarketKey = marketStatusKey
	marketStatus.EventKey = market.EventKey
	marketStatus.OperatorId = operatorId
	marketStatus.OperatorName = operator.OperatorName
	marketStatus.ProviderId = providerId
	marketStatus.ProviderName = market.ProviderName
	marketStatus.SportId = sportId
	marketStatus.SportName = market.SportName
	marketStatus.CompetitionId = market.CompetitionId
	marketStatus.CompetitionName = market.CompetitionName
	marketStatus.EventId = eventId
	marketStatus.EventName = market.EventName
	marketStatus.MarketId = market.MarketId
	marketStatus.MarketName = market.MarketName
	marketStatus.MarketType = market.MarketType
	marketStatus.ProviderStatus = "ACTIVE"
	marketStatus.OperatorStatus = "ACTIVE"
	marketStatus.Favourite = false
	marketStatus.CreatedAt = time.Now().Unix()
	marketStatus.UpdatedAt = marketStatus.CreatedAt
	marketStatus.Config = commondto.ConfigDto{}
	marketStatus.Config = config
	err = database.InsertMarketStatus(marketStatus)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Update Market Dto.
	SetMarketStatus(marketStatus)
	return nil
}

func AddMarketStatusOpStatus(operatorId string, providerId string, sportId string, eventId string, marketId string, status string) error {
	market, err := GetMarket(providerId, sportId, eventId, marketId)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	marketStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId + "-" + marketId
	marketStatus := models.MarketStatus{}
	marketStatus.MarketKey = marketStatusKey
	marketStatus.EventKey = market.EventKey
	marketStatus.OperatorId = operatorId
	marketStatus.OperatorName = operator.OperatorName
	marketStatus.ProviderId = providerId
	marketStatus.ProviderName = market.ProviderName
	marketStatus.SportId = sportId
	marketStatus.SportName = market.SportName
	marketStatus.CompetitionId = market.CompetitionId
	marketStatus.CompetitionName = market.CompetitionName
	marketStatus.EventId = eventId
	marketStatus.EventName = market.EventName
	marketStatus.MarketId = market.MarketId
	marketStatus.MarketName = market.MarketName
	marketStatus.MarketType = market.MarketType
	marketStatus.ProviderStatus = "ACTIVE"
	marketStatus.OperatorStatus = status
	marketStatus.Favourite = false
	marketStatus.CreatedAt = time.Now().Unix()
	marketStatus.UpdatedAt = marketStatus.CreatedAt
	marketStatus.Config = commondto.ConfigDto{}
	err = database.InsertMarketStatus(marketStatus)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Update Market Dto.
	SetMarketStatus(marketStatus)
	return nil
}

func AddMarketStatusPaStatus(operatorId string, providerId string, sportId string, eventId string, marketId string, status string) error {
	market, err := GetMarket(providerId, sportId, eventId, marketId)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	marketStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId + "-" + marketId
	marketStatus := models.MarketStatus{}
	marketStatus.MarketKey = marketStatusKey
	marketStatus.EventKey = market.EventKey
	marketStatus.OperatorId = operatorId
	marketStatus.OperatorName = operator.OperatorName
	marketStatus.ProviderId = providerId
	marketStatus.ProviderName = market.ProviderName
	marketStatus.SportId = sportId
	marketStatus.SportName = market.SportName
	marketStatus.CompetitionId = market.CompetitionId
	marketStatus.CompetitionName = market.CompetitionName
	marketStatus.EventId = eventId
	marketStatus.EventName = market.EventName
	marketStatus.MarketId = market.MarketId
	marketStatus.MarketName = market.MarketName
	marketStatus.MarketType = market.MarketType
	marketStatus.ProviderStatus = status
	marketStatus.OperatorStatus = "ACTIVE"
	marketStatus.Favourite = false
	marketStatus.CreatedAt = time.Now().Unix()
	marketStatus.UpdatedAt = marketStatus.CreatedAt
	marketStatus.Config = commondto.ConfigDto{}
	err = database.InsertMarketStatus(marketStatus)
	if err != nil {
		return err
		// return marketStatus, fmt.Errorf("MarketStatus NOT FOUND!")
	}
	// Update Market Dto.
	SetMarketStatus(marketStatus)
	return nil
}
