package cache

import (
	"Sp/database"
	dto "Sp/dto/operator"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

var operatorCache *ristretto.Cache

func init() {
	operatorCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save in Cache
func SetOperatorDetails(operatorDto dto.OperatorDTO) {
	//log.Println("SetOperatorDetails: Updating Cache - ", operatorDto.OperatorId)
	// 1. Update it in token Session Cache
	operatorCache.Set(operatorDto.OperatorId, operatorDto, 0)
	operatorCache.Wait()
	_, found := operatorCache.Get(operatorDto.OperatorId)
	if found {
		//log.Println("operatorCache: Operator SAVE is SUCCESS in Cache - " + operatorDto.OperatorId)
		return
	}
	log.Println("operatorCache: Operator NOT SAVED in Cache - " + operatorDto.OperatorId)
}

// Get Operator Details from Cache
func GetOperatorDetails(operatorId string) (dto.OperatorDTO, error) {
	//log.Println("GetOperatorDetails: looking for token - ", operatorId)
	// 0. Create resp object
	operatorDto := dto.OperatorDTO{}
	// 1. Get from Cache
	value, found := operatorCache.Get(operatorId)
	if found {
		// 1.1 Token FOUND in cache, retrun session object
		operatorDto = value.(dto.OperatorDTO)
		return operatorDto, nil
	}
	log.Println("operatorCache: Operator NOT FOUND in Cache - " + operatorId)
	// 2. Token NOT FOUND in cache, get from DB and update to cache
	operatorDto, err := database.GetOperatorDetails(operatorId)
	if err != nil {
		// 2.1 Token NOT FOUND in DB, return error
		log.Println("operatorCache: Operator NOT FOUND in DB - ", err.Error())
		return operatorDto, fmt.Errorf("Operator NOT FOUND!")
	}
	// 3. Token FOUND in DB, add to Cache
	SetOperatorDetails(operatorDto)
	// 4. return session object
	return operatorDto, nil
}

func GetOperatorCacheMetrics() {
	log.Println("operatorCache: Metrics - ", operatorCache.Metrics.String())
}
