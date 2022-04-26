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
// Value: Map[EventId]EventStatus

// 	[OperatorId-ProviderId]							mapKey 1, mapValue1
//		[SportId]									mapKey 2, mapValue2		=> Level 2 Data
//			[EventId] -> EventStatus				key 3, value 			=> Level 3 Data

// Level 1:
// mapKey1: OperatorId-ProviderId
// mapValue1: Map[SportId]Level2Data

// Level 2:
// mapKey2: SportId
// mapValue2: Map[EventId]Level3Data

// Level 3:
// Key3: CompititionId
// Value3: models.Compitition

var eventStatusCache *ristretto.Cache

func init() {
	eventStatusCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save EventStatus in Cache
func SetEventStatus(eventStatus models.EventStatus) {
	eventStatusKey := eventStatus.OperatorId + "-" + eventStatus.ProviderId + "-" + eventStatus.SportId + "-" + eventStatus.EventId
	// Level 1 - OperatorId-ProviderId
	level1Key := eventStatus.OperatorId + "-" + eventStatus.ProviderId
	mapValue1, isFound := eventStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, add to the cache
		log.Println("eventStatusCache: (SET) Operator-Provider NOT FOUND in Cache - " + eventStatusKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.EventStatus) // initializing map[CompititionId]compitition
		level3Data[eventStatus.EventId] = eventStatus

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.EventStatus) // initializing map[SportId]level3Data
		level2Data[eventStatus.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		eventStatusCache.Set(level1Key, level2Data, 0)
		eventStatusCache.Wait()

		return
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.EventStatus)
	level3Data, isFound := level2Data[eventStatus.SportId]
	if false == isFound {
		// Leve 2 Key not found
		log.Println("eventStatusCache: (SET) Sport NOT FOUND in Cache - " + eventStatusKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.EventStatus) // initializing map[CompititionId]compitition
		level3Data[eventStatus.EventId] = eventStatus

		// Level 2 Data - add
		level2Data[eventStatus.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		eventStatusCache.Set(level1Key, level2Data, 0)
		eventStatusCache.Wait()

		return
	}
	// Level 3 - CompititionId

	// Level 3 Data -  add or update
	level3Data[eventStatus.EventId] = eventStatus

	// Level 2 Data - update
	level2Data[eventStatus.SportId] = level3Data

	// Level 1 Data - Update it in Cache
	eventStatusCache.Set(level1Key, level2Data, 0)
	eventStatusCache.Wait()

	return
}

// Get EventStatus from Cache
func GetEventStatus(operatorId string, providerId string, sportId string, eventId string) (models.EventStatus, error) {
	// 0. Create resp object
	eventStatus := models.EventStatus{}
	// eventStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId

	// Level 1 - ProviderId
	level1Key := operatorId + "-" + providerId
	mapValue1, isFound := eventStatusCache.Get(level1Key) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
		// // log.Println("eventStatusCache: (GET) Provider NOT FOUND in Cache - " + eventStatusKey)

		// // Get from DB
		// eventStatus, err := getEventStatusFromDB(operatorId, providerId, sportId, eventId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	// log.Println("eventStatusCache: EventStatus NOT FOUND in DB is - ", eventStatusKey)
		// 	return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data := make(map[string]models.EventStatus) // initializing map[CompititionId]compitition
		// level3Data[eventStatus.EventId] = eventStatus

		// // Level 2 Data - init & add
		// level2Data := make(map[string]map[string]models.EventStatus) // initializing map[SportId]level3Data
		// level2Data[eventStatus.SportId] = level3Data

		// // Level 1 Data - Save it in Cache (insert)
		// eventStatusCache.Set(level1Key, level2Data, 0)
		// eventStatusCache.Wait()

		// return eventStatus, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.EventStatus)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
		// // log.Println("eventStatusCache: (GET) Sport NOT FOUND in Cache - " + eventStatusKey)

		// // Get from DB
		// eventStatus, err := getEventStatusFromDB(operatorId, providerId, sportId, eventId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	// log.Println("eventStatusCache: EventStatus NOT FOUND in DB is - ", eventStatusKey)
		// 	return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
		// }

		// // Level 3 Data - init & add
		// level3Data = make(map[string]models.EventStatus) // initializing map[CompititionId]compitition
		// level3Data[eventStatus.EventId] = eventStatus

		// // Level 2 Data - add
		// level2Data[eventStatus.SportId] = level3Data

		// // Level 1 Data - Update it in Cache
		// eventStatusCache.Set(level1Key, level2Data, 0)
		// eventStatusCache.Wait()

		// return eventStatus, nil
	}
	// Level 3 - CompititionId
	eventStatus, isFound = level3Data[eventId]
	if false == isFound {
		// Level 3 Key not found, Get from DB and add to the cache
		return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
		// log.Println("eventStatusCache: (GET) EventStatus NOT FOUND in Cache - " + eventStatusKey)

		// // Get from DB
		// eventStatus, err := getEventStatusFromDB(operatorId, providerId, sportId, eventId)
		// if err != nil {
		// 	// 2.1 NOT FOUND in DB, return error
		// 	// log.Println("eventStatusCache: EventStatus NOT FOUND in DB is - ", eventStatusKey)
		// 	return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
		// }

		// // Level 3 Data - add
		// level3Data[eventStatus.EventId] = eventStatus

		// // Level 2 Data - update
		// level2Data[eventStatus.SportId] = level3Data

		// // Level 1 Data - update it in Cache
		// eventStatusCache.Set(level1Key, level2Data, 0)
		// eventStatusCache.Wait()

		// return eventStatus, nil
	}

	return eventStatus, nil
}

// func getEventStatusFromDB(operatorId string, providerId string, sportId string, eventId string) (models.EventStatus, error) {
// 	// 0. Create resp object
// 	eventStatus := models.EventStatus{}
// 	eventStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
// 	eventStatus, err := database.GetEventStatus(eventStatusKey)
// 	if err != nil {
// 		// 2.1 NOT FOUND in DB, return error
// 		log.Println("eventStatusCache: EventStatus NOT FOUND in DB is - ", eventStatusKey)
// 		return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
// 		// // Add EventStatus if Event in present in cache/db
// 		// event, err := GetEvent(providerId, sportId, eventId)
// 		// if err != nil {
// 		// 	return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
// 		// }
// 		// operator, err := GetOperatorDetails(operatorId)
// 		// if err != nil {
// 		// 	return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
// 		// }
// 		// // Not found, create and add to the missing list
// 		// eventStatus := models.EventStatus{}
// 		// eventStatus.EventKey = eventStatusKey
// 		// eventStatus.OperatorId = operatorId
// 		// eventStatus.OperatorName = operator.OperatorName
// 		// eventStatus.ProviderId = providerId
// 		// eventStatus.ProviderName = event.ProviderName
// 		// eventStatus.SportId = sportId
// 		// eventStatus.SportName = event.SportName
// 		// eventStatus.CompetitionId = event.CompetitionId
// 		// eventStatus.CompetitionName = event.CompetitionName
// 		// eventStatus.EventId = eventId
// 		// eventStatus.EventName = event.EventName
// 		// eventStatus.ProviderStatus = "ACTIVE"
// 		// eventStatus.OperatorStatus = "ACTIVE"
// 		// eventStatus.Favourite = false
// 		// eventStatus.CreatedAt = time.Now().Unix()
// 		// eventStatus.UpdatedAt = eventStatus.CreatedAt
// 		// err = database.InsertEventStatus(eventStatus)
// 		// if err != nil {
// 		// 	return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
// 		// }
// 	}
// 	return eventStatus, nil
// }

func AddEventStatusConfig(operatorId string, providerId string, sportId string, eventId string, config commondto.ConfigDto) error {
	event, err := GetEvent(providerId, sportId, eventId)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	eventStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
	eventStatus := models.EventStatus{}
	eventStatus.EventKey = eventStatusKey
	eventStatus.OperatorId = operatorId
	eventStatus.OperatorName = operator.OperatorName
	eventStatus.ProviderId = providerId
	eventStatus.ProviderName = event.ProviderName
	eventStatus.SportId = sportId
	eventStatus.SportName = event.SportName
	eventStatus.CompetitionId = event.CompetitionId
	eventStatus.CompetitionName = event.CompetitionName
	eventStatus.EventId = eventId
	eventStatus.EventName = event.EventName
	eventStatus.ProviderStatus = "ACTIVE"
	eventStatus.OperatorStatus = "ACTIVE"
	eventStatus.Favourite = false
	eventStatus.CreatedAt = time.Now().Unix()
	eventStatus.UpdatedAt = eventStatus.CreatedAt
	eventStatus.Config = commondto.ConfigDto{}
	eventStatus.Config = config
	err = database.InsertEventStatus(eventStatus)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	// Update Event Dto.
	SetEventStatus(eventStatus)
	return nil
}

func AddEventStatusOpStatus(operatorId string, providerId string, sportId string, eventId string, status string) error {
	event, err := GetEvent(providerId, sportId, eventId)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	eventStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
	eventStatus := models.EventStatus{}
	eventStatus.EventKey = eventStatusKey
	eventStatus.OperatorId = operatorId
	eventStatus.OperatorName = operator.OperatorName
	eventStatus.ProviderId = providerId
	eventStatus.ProviderName = event.ProviderName
	eventStatus.SportId = sportId
	eventStatus.SportName = event.SportName
	eventStatus.CompetitionId = event.CompetitionId
	eventStatus.CompetitionName = event.CompetitionName
	eventStatus.EventId = eventId
	eventStatus.EventName = event.EventName
	eventStatus.ProviderStatus = "ACTIVE"
	eventStatus.OperatorStatus = status
	eventStatus.Favourite = false
	eventStatus.CreatedAt = time.Now().Unix()
	eventStatus.UpdatedAt = eventStatus.CreatedAt
	eventStatus.Config = commondto.ConfigDto{}
	err = database.InsertEventStatus(eventStatus)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	// Update Event Dto.
	SetEventStatus(eventStatus)
	return nil
}

func AddEventStatusPaStatus(operatorId string, providerId string, sportId string, eventId string, status string) error {
	event, err := GetEvent(providerId, sportId, eventId)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	// Not found, create and add to the missing list
	eventStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
	eventStatus := models.EventStatus{}
	eventStatus.EventKey = eventStatusKey
	eventStatus.OperatorId = operatorId
	eventStatus.OperatorName = operator.OperatorName
	eventStatus.ProviderId = providerId
	eventStatus.ProviderName = event.ProviderName
	eventStatus.SportId = sportId
	eventStatus.SportName = event.SportName
	eventStatus.CompetitionId = event.CompetitionId
	eventStatus.CompetitionName = event.CompetitionName
	eventStatus.EventId = eventId
	eventStatus.EventName = event.EventName
	eventStatus.ProviderStatus = status
	eventStatus.OperatorStatus = "ACTIVE"
	eventStatus.Favourite = false
	eventStatus.CreatedAt = time.Now().Unix()
	eventStatus.UpdatedAt = eventStatus.CreatedAt
	eventStatus.Config = commondto.ConfigDto{}
	err = database.InsertEventStatus(eventStatus)
	if err != nil {
		return err
		// return eventStatus, fmt.Errorf("EventStatus NOT FOUND!")
	}
	// Update Event Dto.
	SetEventStatus(eventStatus)
	return nil
}
