package cache

import (
	"Sp/database"
	"Sp/dto/models"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

// Key: PartnerKey = OperartorId+"-"+PartnerId
// Value: Map[ProviderId]PartnerStatus
var partnerStatusCache *ristretto.Cache

func init() {
	partnerStatusCache, _ = InitializeCache(5000, 1<<12, 500)
}

// Set PartnerStatus by OperatorId & PartnerId
func SetOpPartnerStatus(operatorId string, partnerId string, mapValue map[string]models.PartnerStatus) {
	partnerKey := operatorId + "-" + partnerId
	partnerStatusCache.Set(partnerKey, mapValue, 0)
	partnerStatusCache.Wait()
	_, found := partnerStatusCache.Get(partnerKey)
	if found {
		//log.Println("partnerStatusCache: PartnerStatus SAVE is SUCCESS in Cache - " + partnerKey)
		return
	}
	log.Println("partnerStatusCache: PartnerStatus NOT SAVED in Cache - " + partnerKey)
	return
}

// Set PartnerStatus
func SetPartnerStatus(partnerStatus models.PartnerStatus) {
	// 0. Define MapValue
	mapValue := make(map[string]models.PartnerStatus)
	// 1. Get MapValue By OperatorId
	partnerKey := partnerStatus.OperatorId + "-" + partnerStatus.PartnerId
	value, found := partnerStatusCache.Get(partnerKey)
	if found {
		// 1.1. FOUND, assign value
		mapValue = value.(map[string]models.PartnerStatus)
	} else {
		// 1.2. NOT FOUND in cache, make new mapvalue
		mapValue = make(map[string]models.PartnerStatus)
	}
	// 2. add/update value in mapValue
	mapValue[partnerStatus.ProviderId] = partnerStatus
	// 3. add/update it mapValue in Cache
	partnerStatusCache.Set(partnerKey, mapValue, 0)
	partnerStatusCache.Wait()
	_, found = partnerStatusCache.Get(partnerKey)
	if found {
		//log.Println("partnerStatusCache: PartnerStatus SAVE is SUCCESS in Cache - " + partnerKey)
		return
	}
	log.Println("partnerStatusCache: PartnerStatus NOT SAVED in Cache - " + partnerKey)
	return
}

// Get PartnerStatus by OpertorId & PartnerId
func GetOpPartnerStatus(operatorId string, partnerId string) (map[string]models.PartnerStatus, error) {
	// 0. Create resp object
	mapValue := make(map[string]models.PartnerStatus)
	// 1. Get from Cache
	partnerKey := operatorId + "-" + partnerId
	value, found := partnerStatusCache.Get(partnerKey)
	if found {
		// 1.1 FOUND in cache, retrun object
		mapValue = value.(map[string]models.PartnerStatus)
		return mapValue, nil
	}
	log.Println("partnerStatusCache: PartnerStatus NOT FOUND in Cache - " + partnerKey)
	// 2. NOT FOUND in cache, get from DB and update to cache
	partnerStatus, err := database.GetPartnerProviders(operatorId, partnerId)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("partnerStatusCache: PartnerStatus NOT FOUND in DB - ", err.Error())
		return mapValue, fmt.Errorf("PartnerStatus NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	for _, ps := range partnerStatus {
		mapValue[ps.ProviderId] = ps
	}
	SetOpPartnerStatus(operatorId, partnerId, mapValue)
	// 4. return object
	return mapValue, nil
}

// Get PartnerStatus by OpertorId, PartnerId & ProviderId
func GetPartnerStatus(operatorId string, partnerId string, providerId string) (models.PartnerStatus, error) {
	// 0. Create resp object
	ps := models.PartnerStatus{}
	mapValue := make(map[string]models.PartnerStatus)
	// 1. Get from Cache
	partnerKey := operatorId + "-" + partnerId
	value, found := partnerStatusCache.Get(partnerKey)
	if found {
		// 1.1. FOUND in cache, look for partner
		mapValue = value.(map[string]models.PartnerStatus)
		val, ok := mapValue[providerId]
		if ok {
			// 1.1.1. FOUND in cache, return object
			return val, nil
		}
		// 1.2. NOT FOUND in map, get from DB and update cache
		log.Println("partnerStatusCache: PartnerStatus NOT FOUND in Cache Map!!!")
		partnerStatusKey := partnerKey + "-" + providerId
		ps, err := database.GetPartnerStatus(partnerStatusKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("partnerStatusCache: GetPartnerStatus failed with error - ", err.Error())
			return ps, fmt.Errorf("PartnerStatus NOT FOUND!")
		}
		mapValue[providerId] = ps
		SetOpPartnerStatus(operatorId, partnerId, mapValue)
		return ps, nil
	}
	log.Println("partnerStatusCache: PartnerStatus NOT FOUND in Cache - " + partnerKey)
	// 2. NOT FOUND in cache, get from DB and update to cache
	partnerStatus, err := database.GetPartnerProviders(operatorId, partnerId)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("partnerStatusCache: PartnerStatus NOT FOUND in DB - ", err.Error())
		return ps, fmt.Errorf("PartnerStatus NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	for _, ps := range partnerStatus {
		mapValue[ps.ProviderId] = ps
	}
	SetOpPartnerStatus(operatorId, partnerId, mapValue)
	// 4. return object
	val, ok := mapValue[providerId]
	if ok {
		// 1.1.1. FOUND in cache, return object
		return val, nil
	}
	return ps, nil
}

// Get PartnerStatus by OpertorId, PartnerId & ProviderId
func GetPartnerStatusFromCache(operatorId string, partnerId string, providerId string) (models.PartnerStatus, error) {
	// 0. Create resp object
	ps := models.PartnerStatus{}
	mapValue := make(map[string]models.PartnerStatus)
	// 1. Get from Cache
	partnerKey := operatorId + "-" + partnerId
	value, found := partnerStatusCache.Get(partnerKey)
	if found {
		// 1.1. FOUND in cache, look for partner
		mapValue = value.(map[string]models.PartnerStatus)
		val, ok := mapValue[providerId]
		if ok {
			// 1.1.1. FOUND in cache, return object
			return val, nil
		}
	}
	return ps, fmt.Errorf("PartnerStatus NOT FOUND!")
}
