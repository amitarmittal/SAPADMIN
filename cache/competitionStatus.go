package cache

import (
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	"Sp/dto/models"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
)

// Key: OperartorId+"-"+ProviderId+"-"+SportId
// Value: Map[CompetitionId]CompetitionStatus

// 	[OperatorId-ProviderId]							mapKey 1, mapValue1
//		[SportId]									mapKey 2, mapValue2		=> Level 2 Data
//			[CompititionId] -> CompititionStatus	key 3, value 			=> Level 3 Data

// Level 1:
// mapKey1: OperatorId-ProviderId
// mapValue1: Map[SportId]Level2Data

// Level 2:
// mapKey2: SportId
// mapValue2: Map[CompetitionId]Level3Data

// Level 3:
// Key3: CompititionId
// Value3: models.Compitition

var competitionStatusCache *ristretto.Cache

func init() {
	competitionStatusCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save CompetitionStatus in Cache
func SetCompetitionStatus(competitionStatus models.CompetitionStatus) {
	competitionStatusKey := competitionStatus.OperatorId + "-" + competitionStatus.ProviderId + "-" + competitionStatus.SportId + "-" + competitionStatus.CompetitionId
	// Level 1 - OperatorId-ProviderId
	level1Key := competitionStatus.OperatorId + "-" + competitionStatus.ProviderId
	mapValue1, isFound := competitionStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, add to the cache
		log.Println("competitionStatusCache: (SET) Operator-Provider NOT FOUND in Cache - " + competitionStatusKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.CompetitionStatus) // initializing map[CompititionId]compitition
		level3Data[competitionStatus.CompetitionId] = competitionStatus

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.CompetitionStatus) // initializing map[SportId]level3Data
		level2Data[competitionStatus.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		competitionStatusCache.Set(level1Key, level2Data, 0)
		competitionStatusCache.Wait()

		return
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.CompetitionStatus)
	level3Data, isFound := level2Data[competitionStatus.SportId]
	if false == isFound {
		// Leve 2 Key not found
		log.Println("competitionStatusCache: (SET) Sport NOT FOUND in Cache - " + competitionStatusKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.CompetitionStatus) // initializing map[CompititionId]compitition
		level3Data[competitionStatus.CompetitionId] = competitionStatus

		// Level 2 Data - add
		level2Data[competitionStatus.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		competitionStatusCache.Set(level1Key, level2Data, 0)
		competitionStatusCache.Wait()

		return
	}
	// Level 3 - CompititionId

	// Level 3 Data -  add or update
	level3Data[competitionStatus.CompetitionId] = competitionStatus

	// Level 2 Data - update
	level2Data[competitionStatus.SportId] = level3Data

	// Level 1 Data - Update it in Cache
	competitionStatusCache.Set(level1Key, level2Data, 0)
	competitionStatusCache.Wait()

	return
}

// Get CompetitionStatus from Cache
func GetCompetitionStatus(operatorId string, providerId string, sportId string, competitionId string) (models.CompetitionStatus, error) {
	// 0. Create resp object
	competitionStatus := models.CompetitionStatus{}
	// competitionStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId

	// Level 1 - ProviderId
	level1Key := operatorId + "-" + providerId
	mapValue1, isFound := competitionStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// log.Println("competitionStatusCache: (GET) Provider NOT FOUND in Cache - " + competitionStatusKey)

		// // Get from DB
		// competitionStatus, err := getCompetitionStatusFromDB(operatorId, providerId, sportId, competitionId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("competitionStatusCache: CompetitionStatus NOT FOUND in DB is - ", competitionStatusKey)
		// 	return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data := make(map[string]models.CompetitionStatus) // initializing map[CompititionId]compitition
		// level3Data[competitionStatus.CompetitionId] = competitionStatus

		// // Level 2 Data - init & add
		// level2Data := make(map[string]map[string]models.CompetitionStatus) // initializing map[SportId]level3Data
		// level2Data[competitionStatus.SportId] = level3Data

		// // Level 1 Data - Save it in Cache (insert)
		// competitionStatusCache.Set(level1Key, level2Data, 0)
		// competitionStatusCache.Wait()

		// return competitionStatus, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.CompetitionStatus)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// log.Println("competitionStatusCache: (GET) Sport NOT FOUND in Cache - " + competitionStatusKey)

		// // Get from DB
		// competitionStatus, err := getCompetitionStatusFromDB(operatorId, providerId, sportId, competitionId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("competitionStatusCache: CompetitionStatus NOT FOUND in DB is - ", competitionStatusKey)
		// 	return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data = make(map[string]models.CompetitionStatus) // initializing map[CompititionId]compitition
		// level3Data[competitionStatus.CompetitionId] = competitionStatus

		// // Level 2 Data - add
		// level2Data[competitionStatus.SportId] = level3Data

		// // Level 1 Data - Update it in Cache
		// competitionStatusCache.Set(level1Key, level2Data, 0)
		// competitionStatusCache.Wait()

		// return competitionStatus, nil
	}
	// Level 3 - CompititionId
	competitionStatus, isFound = level3Data[competitionId]
	if false == isFound {
		// Level 3 Key not found, Get from DB and add to the cache
		return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// log.Println("competitionStatusCache: (GET) CompetitionStatus NOT FOUND in Cache - " + competitionStatusKey)

		// // Get from DB
		// competitionStatus, err := getCompetitionStatusFromDB(operatorId, providerId, sportId, competitionId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("competitionStatusCache: CompetitionStatus NOT FOUND in DB is - ", competitionStatusKey)
		// 	return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// }

		// // Level 3 Data - add
		// level3Data[competitionStatus.CompetitionId] = competitionStatus

		// // Level 2 Data - update
		// level2Data[competitionStatus.SportId] = level3Data

		// // Level 1 Data - update it in Cache
		// competitionStatusCache.Set(level1Key, level2Data, 0)
		// competitionStatusCache.Wait()

		// return competitionStatus, nil
	}

	return competitionStatus, nil
}

// func getCompetitionStatusFromDB(operatorId string, providerId string, sportId string, competitionId string) (models.CompetitionStatus, error) {
// 	// 0. Create resp object
// 	competitionStatus := models.CompetitionStatus{}
// 	competitionStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId
// 	competitionStatus, err := database.GetCompetitionStatus(competitionStatusKey)
// 	if err != nil {
// 		// 2.1 NOT FOUND in DB, return error
// 		log.Println("competitionStatusCache: CompetitionStatus NOT FOUND in DB is - ", competitionStatusKey)
// 		// Add CompetitionStatus if Competition in present in cache/db
// 		competition, err := GetCompetition(providerId, sportId, competitionId)
// 		if err != nil {
// 			return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
// 		}
// 		operator, err := GetOperatorDetails(operatorId)
// 		if err != nil {
// 			return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
// 		}
// 		// Not found, create and add to the missing list
// 		competitionStatus := models.CompetitionStatus{}
// 		competitionStatus.CompetitionKey = competitionStatusKey
// 		competitionStatus.OperatorId = operatorId
// 		competitionStatus.OperatorName = operator.OperatorName
// 		competitionStatus.ProviderId = providerId
// 		competitionStatus.ProviderName = competition.ProviderName
// 		competitionStatus.SportId = sportId
// 		competitionStatus.SportName = competition.SportName
// 		competitionStatus.CompetitionId = competitionId
// 		competitionStatus.CompetitionName = competition.CompetitionName
// 		competitionStatus.ProviderStatus = "ACTIVE"
// 		competitionStatus.OperatorStatus = "ACTIVE"
// 		competitionStatus.Favourite = false
// 		competitionStatus.CreatedAt = time.Now().Unix()
// 		err = database.InsertCompetitionStatus(competitionStatus)
// 		if err != nil {
// 			return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
// 		}
// 	}
// 	return competitionStatus, nil
// }

// Get CompetitionStatus from Cache
func GetCompetitionStatusBySport(operatorId string, providerId string, sportId string) ([]models.CompetitionStatus, error) {
	// 0. Create resp object
	csS := []models.CompetitionStatus{}

	// Level 1 - ProviderId
	level1Key := operatorId + "-" + providerId
	mapValue1, isFound := competitionStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		log.Println("competitionStatusCache: (GET) Provider NOT FOUND in Cache - " + level1Key)

		// Get from DB
		// competitionStatus, err := getCompetitionStatusFromDB(operatorId, providerId, sportId, competitionId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("competitionStatusCache: CompetitionStatus NOT FOUND in DB is - ", competitionStatusKey)
		// 	return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data := make(map[string]models.CompetitionStatus) // initializing map[CompititionId]compitition
		// level3Data[competitionStatus.CompetitionId] = competitionStatus

		// // Level 2 Data - init & add
		// level2Data := make(map[string]map[string]models.CompetitionStatus) // initializing map[SportId]level3Data
		// level2Data[competitionStatus.SportId] = level3Data

		// // Level 1 Data - Save it in Cache (insert)
		// competitionStatusCache.Set(level1Key, level2Data, 0)
		// competitionStatusCache.Wait()

		return csS, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.CompetitionStatus)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		competitionStatusKey := level1Key + "-" + sportId
		// Level 2 Key not found, Get from DB and add to the cache
		log.Println("competitionStatusCache: (GET) Sport NOT FOUND in Cache - " + competitionStatusKey)

		// Get from DB
		// competitionStatus, err := getCompetitionStatusFromDB(operatorId, providerId, sportId, competitionId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	log.Println("competitionStatusCache: CompetitionStatus NOT FOUND in DB is - ", competitionStatusKey)
		// 	return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data = make(map[string]models.CompetitionStatus) // initializing map[CompititionId]compitition
		// level3Data[competitionStatus.CompetitionId] = competitionStatus

		// // Level 2 Data - add
		// level2Data[competitionStatus.SportId] = level3Data

		// // Level 1 Data - Update it in Cache
		// competitionStatusCache.Set(level1Key, level2Data, 0)
		// competitionStatusCache.Wait()

		return csS, nil
	}
	for _, cs := range level3Data {
		if cs.ProviderStatus == constants.SAP.ObjectStatus.ACTIVE() {
			csS = append(csS, cs)
		}
	}
	return csS, nil
}

func AddCompetitionStatusConfig(operatorId string, providerId string, sportId string, competitionId string, config commondto.ConfigDto) error {
	competition, err := GetCompetition(providerId, sportId, competitionId)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	competitionStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId
	competitionStatus := models.CompetitionStatus{}
	competitionStatus.CompetitionKey = competitionStatusKey
	competitionStatus.OperatorId = operatorId
	competitionStatus.OperatorName = operator.OperatorName
	competitionStatus.ProviderId = providerId
	competitionStatus.ProviderName = competition.ProviderName
	competitionStatus.SportId = sportId
	competitionStatus.SportName = competition.SportName
	competitionStatus.CompetitionId = competition.CompetitionId
	competitionStatus.CompetitionName = competition.CompetitionName
	competitionStatus.ProviderStatus = "ACTIVE"
	competitionStatus.OperatorStatus = "ACTIVE"
	competitionStatus.Favourite = false
	competitionStatus.CreatedAt = time.Now().Unix()
	competitionStatus.UpdatedAt = competitionStatus.CreatedAt
	competitionStatus.Config = commondto.ConfigDto{}
	competitionStatus.Config = config
	err = database.InsertCompetitionStatus(competitionStatus)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	// Update Competition Dto.
	SetCompetitionStatus(competitionStatus)
	return nil
}

func AddCompetitionStatusOpStatus(operatorId string, providerId string, sportId string, competitionId string, status string) error {
	competition, err := GetCompetition(providerId, sportId, competitionId)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	competitionStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId
	competitionStatus := models.CompetitionStatus{}
	competitionStatus.CompetitionKey = competitionStatusKey
	competitionStatus.OperatorId = operatorId
	competitionStatus.OperatorName = operator.OperatorName
	competitionStatus.ProviderId = providerId
	competitionStatus.ProviderName = competition.ProviderName
	competitionStatus.SportId = sportId
	competitionStatus.SportName = competition.SportName
	competitionStatus.CompetitionId = competition.CompetitionId
	competitionStatus.CompetitionName = competition.CompetitionName
	competitionStatus.ProviderStatus = "ACTIVE"
	competitionStatus.OperatorStatus = status
	competitionStatus.Favourite = false
	competitionStatus.CreatedAt = time.Now().Unix()
	competitionStatus.UpdatedAt = competitionStatus.CreatedAt
	competitionStatus.Config = commondto.ConfigDto{}
	err = database.InsertCompetitionStatus(competitionStatus)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	// Update Competition Dto.
	SetCompetitionStatus(competitionStatus)
	return nil
}

func AddCompetitionStatusPaStatus(operatorId string, providerId string, sportId string, competitionId string, status string) error {
	competition, err := GetCompetition(providerId, sportId, competitionId)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	competitionStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId
	competitionStatus := models.CompetitionStatus{}
	competitionStatus.CompetitionKey = competitionStatusKey
	competitionStatus.OperatorId = operatorId
	competitionStatus.OperatorName = operator.OperatorName
	competitionStatus.ProviderId = providerId
	competitionStatus.ProviderName = competition.ProviderName
	competitionStatus.SportId = sportId
	competitionStatus.SportName = competition.SportName
	competitionStatus.CompetitionId = competition.CompetitionId
	competitionStatus.CompetitionName = competition.CompetitionName
	competitionStatus.ProviderStatus = status
	competitionStatus.OperatorStatus = "ACTIVE"
	competitionStatus.Favourite = false
	competitionStatus.CreatedAt = time.Now().Unix()
	competitionStatus.UpdatedAt = competitionStatus.CreatedAt
	competitionStatus.Config = commondto.ConfigDto{}
	err = database.InsertCompetitionStatus(competitionStatus)
	if err != nil {
		return err
		// return competitionStatus, fmt.Errorf("CompetitionStatus NOT FOUND!")
	}
	// Update Competition Dto.
	SetCompetitionStatus(competitionStatus)
	return nil
}
