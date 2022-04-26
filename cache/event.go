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
//			[EventId] -> Event					key 3, value 			=> Level 3 Data

// Level 1:
// mapKey1: ProviderId
// mapValue1: Map[SportId]Level2Data

// Level 2:
// mapKey2: SportId
// mapValue2: Map[CompetitionId]Level3Data

// Level 3:
// Key3: EventId
// Value3: models.Event

var eventCache *ristretto.Cache

func init() {
	eventCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save Event in Cache
func SetEvent(event models.Event) {
	eventKey := event.ProviderId + "-" + event.SportId + "-" + event.EventId
	// Level 1 - ProviderId
	mapValue1, isFound := eventCache.Get(event.ProviderId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, add to the cache
		log.Println("eventCache: (SET) Provider NOT FOUND in Cache - " + eventKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Event) // initializing map[EventId]event
		level3Data[event.EventId] = event

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Event) // initializing map[SportId]level3Data
		level2Data[event.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		eventCache.Set(event.ProviderId, level2Data, 0)
		eventCache.Wait()

		return
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Event)
	level3Data, isFound := level2Data[event.SportId]
	if false == isFound {
		// Leve 2 Key not found
		log.Println("eventCache: (SET) Sport NOT FOUND in Cache - " + eventKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Event) // initializing map[EventId]event
		level3Data[event.EventId] = event

		// Level 2 Data - add
		level2Data[event.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		eventCache.Set(event.ProviderId, level2Data, 0)
		eventCache.Wait()

		return
	}
	// Level 3 Data -  add or update
	level3Data[event.EventId] = event

	// Level 2 Data - update
	level2Data[event.SportId] = level3Data

	// Level 1 Data - Update it in Cache
	eventCache.Set(event.ProviderId, level2Data, 0)
	eventCache.Wait()

	return
}

// Get Event from Cache
func GetEvent(providerId string, sportId string, eventId string) (models.Event, error) {
	// 0. Create resp object
	event := models.Event{}
	eventKey := providerId + "-" + sportId + "-" + eventId

	// Level 1 - ProviderId
	mapValue1, isFound := eventCache.Get(providerId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		log.Println("eventCache: (GET) Provider NOT FOUND in Cache - " + eventKey)

		// Get from DB
		event, err := database.GetEventDetails(eventKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("eventCache: Event NOT FOUND in DB - ", err.Error())
			log.Println("eventCache: EventKey is - ", eventKey)
			return event, fmt.Errorf("Event NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Event) // initializing map[EventId]event
		level3Data[event.EventId] = event

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Event) // initializing map[SportId]level3Data
		level2Data[event.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		eventCache.Set(event.ProviderId, level2Data, 0)
		eventCache.Wait()

		return event, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Event)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		log.Println("eventCache: (GET) Sport NOT FOUND in Cache - " + eventKey)

		// Get from DB
		event, err := database.GetEventDetails(eventKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("eventCache: Event NOT FOUND in DB - ", err.Error())
			log.Println("eventCache: EventKey is - ", eventKey)
			return event, fmt.Errorf("Event NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Event) // initializing map[EventId]event
		level3Data[event.EventId] = event

		// Level 2 Data - add
		level2Data[event.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		eventCache.Set(event.ProviderId, level2Data, 0)
		eventCache.Wait()

		return event, nil
	}
	// Level 3 - EventId
	event, isFound = level3Data[eventId]
	if false == isFound {
		// Level 3 Key not found, Get from DB and add to the cache
		log.Println("eventCache: (GET) Event NOT FOUND in Cache - " + eventKey)

		// Get from DB
		event, err := database.GetEventDetails(eventKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("eventCache: Event NOT FOUND in DB - ", err.Error())
			log.Println("eventCache: EventKey is - ", eventKey)
			return event, fmt.Errorf("Event NOT FOUND!")
		}

		// Level 2 Data - add
		level3Data[event.EventId] = event

		// Level 2 Data - add
		level2Data[event.SportId] = level3Data

		// Level 1 Data - update it in Cache
		eventCache.Set(event.ProviderId, level2Data, 0)
		eventCache.Wait()

		return event, nil
	}

	return event, nil
}
