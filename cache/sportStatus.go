package cache

import (
	"Sp/database"
	"Sp/dto/models"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

// Key: OperartorId+"-"+ProviderId 					// OLD
// Key: OperartorId+"-"+PartnerId+"-"+ProviderId	// NEW
// Value: Map[SportId]SportStatus

var sportStatusCache *ristretto.Cache

func init() {
	sportStatusCache, _ = InitializeCache(10000, 1<<12, 1000)
}

// Set SportStatus by OperatorId & ProviderId
func SetOpSportStatus(operatorId string, partnerId string, providerId string, mapValue map[string]models.SportStatus) {
	key := operatorId + "-" + providerId
	if partnerId != "" {
		key = operatorId + "-" + partnerId + "-" + providerId
	}
	sportStatusCache.Set(key, mapValue, 0)
	sportStatusCache.Wait()
	return
}

// Set SportStatus
func SetSportStatus(sportStatus models.SportStatus) {
	// 0. Define MapValue
	mapValue := make(map[string]models.SportStatus)
	key := sportStatus.OperatorId + "-" + sportStatus.ProviderId
	if sportStatus.PartnerId != "" {
		key = sportStatus.OperatorId + "-" + sportStatus.PartnerId + "-" + sportStatus.ProviderId
	}
	// 1. Get MapValue By OperatorId
	value, found := sportStatusCache.Get(key)
	if found {
		// 1.1. FOUND, assign value
		mapValue = value.(map[string]models.SportStatus)
	} else {
		// 1.2. NOT FOUND in cache, make new mapvalue
		mapValue = make(map[string]models.SportStatus)
	}
	// 2. add/update value in mapValue
	mapValue[sportStatus.SportId] = sportStatus
	// 3. add/update it mapValue in Cache
	sportStatusCache.Set(key, mapValue, 0)
	sportStatusCache.Wait()
	return
}

// Get SportStatus by OpertorId & ProviderId
func GetOpSportStatus(operatorId string, partnerId string, providerId string) (map[string]models.SportStatus, error) {
	// 0. Create resp object
	mapValue := make(map[string]models.SportStatus)
	key := operatorId + "-" + providerId
	if partnerId != "" {
		key = operatorId + "-" + partnerId + "-" + providerId
	}
	// 1. Get from Cache
	value, found := sportStatusCache.Get(key)
	if found {
		// 1.1 FOUND in cache, retrun object
		mapValue = value.(map[string]models.SportStatus)
		return mapValue, nil
	}
	log.Println("sportStatusCache: SportStatus NOT FOUND in Cache - " + key)
	// 2. NOT FOUND in cache, get from DB and update to cache
	sportStatus, err := database.GetOpPrSports(operatorId, partnerId, providerId)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("sportStatusCache: SportStatus NOT FOUND in DB - ", err.Error())
		log.Println("sportStatusCache: SportStatus NOT FOUND in DB for key - ", key)
		return mapValue, fmt.Errorf("SportStatus NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	for _, ss := range sportStatus {
		mapValue[ss.SportId] = ss
	}
	SetOpSportStatus(operatorId, partnerId, providerId, mapValue)
	// 4. return object
	return mapValue, nil
}

// Get SportStatus by OpertorId & ProviderId & SportId
func GetSportStatus(operatorId string, partnerId string, providerId string, sportId string) (models.SportStatus, error) {
	// 0. Create resp object
	ss := models.SportStatus{}
	mapValue := make(map[string]models.SportStatus)
	key := operatorId + "-" + providerId
	// if partnerId != "" { // TODO: Uncomment this
	// 	key = operatorId + "-" + partnerId + "-" + providerId
	// }
	// 1. Get from Cache
	value, found := sportStatusCache.Get(key)
	if found {
		// 1.1. FOUND in cache, retrun object
		mapValue = value.(map[string]models.SportStatus)
		ss, ok := mapValue[sportId]
		if ok {
			// 1.1.1. FOUND in map
			return ss, nil
		}
		// 1.2. NOT FOUND in map, get from DB
		log.Println("sportStatusCache: SportStatus NOT FOUND in Cache for sport - " + sportId)
		sportKey := operatorId + "-" + providerId + "-" + sportId
		if partnerId != "" {
			sportKey = operatorId + "-" + partnerId + "-" + providerId + "-" + sportId
		}
		ss, err := database.GetSportStatus(sportKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("sportStatusCache: GetSportStatus failed with error - ", err.Error())
			log.Println("sportStatusCache: GetSportStatus failed for sportkey - ", sportKey)
			return ss, fmt.Errorf("SportStatus NOT FOUND!")
		}
		mapValue[sportId] = ss
		SetOpSportStatus(operatorId, partnerId, providerId, mapValue)
		return ss, nil
	}
	log.Println("sportStatusCache: SportStatus NOT FOUND in Cache for key - " + key)
	// 2. NOT FOUND in cache, get from DB and update to cache
	sportStatus, err := database.GetOpPrSports(operatorId, "", providerId)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("sportStatusCache: Operator SportStatus NOT FOUND in DB - ", err.Error())
		log.Println("sportStatusCache: Operator SportStatus NOT FOUND in DB for key- ", key)
		return ss, fmt.Errorf("SportStatus NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	for _, ss := range sportStatus {
		mapValue[ss.SportId] = ss
	}
	SetOpSportStatus(operatorId, partnerId, providerId, mapValue)
	// 4. return object
	ss, ok := mapValue[sportId]
	if !ok {
		// 4.1. NOT FOUND in DB map
		log.Println("sportStatusCache: SportStatus NOT FOUND in DB for sport - " + sportId)
		return ss, fmt.Errorf("SportStatus NOT FOUND!")
	}
	return ss, nil
}

func GetSportStatusByKey(sportKey string) (string, error) {
	// 0. Create resp object
	jsonResp := "NOT FOUND IN CACHE"
	//sportStatus := make(map[string]models.SportStatus)
	// 1. Get from Cache
	value, found := sportStatusCache.Get(sportKey)
	if found {
		// 1.1. FOUND in cache, retrun object
		sportStatus := value.(map[string]models.SportStatus)
		jsonBytes, err := json.Marshal(sportStatus)
		if err != nil {
			return jsonResp, err
		}
		return string(jsonBytes), nil
	}
	// // 2. NOT FOUND in cache, get from DB and update to cache
	// sportStatus, err := database.GetSportStatus(sportKey)
	// if err != nil {
	// 	// 2.1 NOT FOUND in DB, return error
	// 	log.Println("GetConfigForSport: Get sportStatus for DB failed with error - ", err.Error())
	// 	return sportStatus.Config, fmt.Errorf("SportStatus NOT FOUND!")
	// }
	// // 3. FOUND in DB, add to Cache
	// SetSportStatus(sportStatus)
	// 4. return object
	return jsonResp, nil
}
