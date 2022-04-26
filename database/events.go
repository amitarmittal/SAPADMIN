package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get Event Document
func GetEventDetails(eventKey string) (models.Event, error) {
	//log.Println("GetEventDetails: Looking for event_id - ", eventId)
	event := models.Event{}
	err := EventCollection.FindOne(Ctx, bson.M{"event_key": eventKey}).Decode(&event)
	if err != nil {
		//log.Println("GetEventDetails: Get Event Details FAILED with error - ", err.Error())
		log.Println("GetEventDetails: eventKey - "+eventKey+" failed with error - ", err.Error())
		return event, err
	}
	return event, nil
}

func GetEventsByEventIds(eventIds []string) ([]models.Event, error) {
	eventsDto := []models.Event{}
	filter := bson.M{"event_id": bson.M{"$in": eventIds}}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetEventsByEventIds: Find failed with error - ", err.Error())
		log.Println("GetEventsByEventIds: Looking for betsCount - ", len(eventIds))
		return eventsDto, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		event := models.Event{}
		err = cursor.Decode(&event)
		if err != nil {
			log.Println("GetBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		eventsDto = append(eventsDto, event)
	}
	return eventsDto, nil
}

// Insert Event Document
func InsertEventDetails(event models.Event) error {
	//log.Println("InsertEventDetails: Adding Documnet for eventId - ", event.EventId)
	event.CreatedAt = time.Now().Unix()
	event.UpdatedAt = event.CreatedAt
	_, err := EventCollection.InsertOne(Ctx, event)
	if err != nil {
		log.Println("InsertEventDetails: FAILED to INSERT event details - ", err.Error())
		return err
	}
	//log.Println("InsertEventDetails: Event Document _id is - ", result.InsertedID)
	return nil
}

// Bulk insert Event Staus objects
func InsertManyEvents(events []models.Event) error {
	//log.Println("InsertManyEvents: Adding Events, count is - ", len(events))
	proStatus := []interface{}{}
	for _, proSts := range events {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	_, err := EventCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManyEvents: FAILED to INSERT - ", err.Error())
		return err
	}
	//log.Println("InsertManyEvents: Inserted count is - ", len(result.InsertedIDs))
	return nil
}

// Update Event Document
func UpdateEventDetails(event models.Event) error {
	//log.Println("UpdateEventDetails: Updating Documnet for event_key - ", event.EventKey)
	opts := options.Update()
	filter := bson.D{{"event_key", event.EventKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"status", event.Status}, {"updated_at", updatedAt}}}}
	_, err := EventCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateEventDetails: FAILED to UPDATE event - ", err.Error())
		return err
	}
	//log.Println("UpdateEventDetails: Matched recoreds Count - ", result.MatchedCount)
	//log.Println("UpdateEventDetails: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Events by ProviderId and SportId
func GetEvents(providerId string, sportId string, competitionId string, timecheck bool) ([]models.Event, error) {
	//log.Println("GetEvents: Looking Events for ProviderId-SportId : ", providerId+"-"+sportId)

	// 0. Response object
	eventsDto := []models.Event{}

	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	if competitionId != "" {
		filter["competition_id"] = competitionId
	}
	todayDate := time.Now()
	if timecheck {
		startDate := todayDate.Add(-1 * 5 * 24 * time.Hour) // five days older
		endDate := todayDate.Add(2 * 24 * time.Hour)        // two days prior
		// filter open_date > = 2 days and open_date < = 5 days
		filter["$and"] = []bson.M{{"open_date": bson.M{"$gte": startDate.UnixMilli(), "$lte": endDate.UnixMilli()}}}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"open_date", -1}})

	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetEvents: Events NOT FOUND for ProviderId-SportId : ", providerId+"-"+sportId)
		log.Println("GetEvents: EventCollection.Find failed with error : ", err.Error())
		return eventsDto, err
	}

	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetEvents: Decode failed with error - ", err.Error())
			continue
		}
		eventsDto = append(eventsDto, event)
	}
	return eventsDto, nil
}

// Get Events by ProviderId and SportId
func GetEventsLast25(providerId string, sportId string, competitionIds []string) ([]models.Event, error) {
	//log.Println("GetEvents: Looking Events for ProviderId-SportId : ", providerId+"-"+sportId)

	// 0. Response object
	eventsDto := []models.Event{}

	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	filter["competition_id"] = bson.M{"$in": competitionIds}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"open_date", -1}})
	findOptions.SetLimit(int64(25 * len(competitionIds)))
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetEvents: Events NOT FOUND for ProviderId-SportId : ", providerId+"-"+sportId)
		log.Println("GetEvents: EventCollection.Find failed with error : ", err.Error())
		return eventsDto, err
	}

	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetEvents: Decode failed with error - ", err.Error())
			continue
		}
		eventsDto = append(eventsDto, event)
	}
	return eventsDto, nil
}

func GetEventsByProviderIdAndSportId(providerId string, sportId string) ([]models.Event, error) {
	//log.Println("GetEvents: Looking Events for ProviderId-SportId : ", providerId+"-"+sportId)

	// 0. Response object
	eventsDto := []models.Event{}

	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"open_date", -1}})
	findOptions.SetLimit(2000)
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetEvents: Events NOT FOUND for ProviderId-SportId : ", providerId+"-"+sportId)
		log.Println("GetEvents: EventCollection.Find failed with error : ", err.Error())
		return eventsDto, err
	}

	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetEvents: Decode failed with error - ", err.Error())
			continue
		}
		eventsDto = append(eventsDto, event)
	}
	return eventsDto, nil
}

func GetLatestEvents() ([]models.Event, error) {
	//log.Println("GetLatestEvents: Return last 100 events")
	// 0. Response object
	events := []models.Event{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	findOptions.SetLimit(500)
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetLatestEvents: Events NOT FOUND!")
		return events, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetLatestEvents: Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func GetUpdatedEvents() ([]models.Event, error) {
	//log.Println("GetUpdatedEvents: Return last 100 documents")
	// 0. Response object
	events := []models.Event{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	findOptions.SetLimit(1000)
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetUpdatedEvents: documents NOT FOUND!")
		return events, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetUpdatedEvents: Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func GetAllEvents(providerId string) ([]models.Event, error) {
	//log.Println("GetAllEvents: Return all events")
	// 0. Response object
	events := []models.Event{}
	// 1. Create Filter
	filter := bson.M{}
	if providerId != "" {
		filter["provider_id"] = providerId
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetLimit(2000)
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllEvents: documents NOT FOUND!")
		return events, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetAllEvents: Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

// Update Event Document
func ReplaceEvent(eventDto models.Event) error {
	eventDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", eventDto.ID}}
	// update := bson.D{{"$set", operatorDto}}
	result, err := EventCollection.ReplaceOne(Ctx, filter, eventDto)
	if err != nil {
		log.Println("ReplaceEvent: FAILED to UPDATE Event details - ", err.Error())
		return err
	}
	log.Println("ReplaceEvent: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceEvent: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Events updated in last 5 minutes
func GetUpdatedEventss() ([]models.Event, error) {
	retObjs := []models.Event{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := EventCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedEventss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.Event{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedEventss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

// Update event dtos
func UpdateEventDtos(eventDtos []models.Event) error {
	for _, eventDto := range eventDtos {
		eventDto.UpdatedAt = time.Now().Unix()
		filter := bson.D{{"_id", eventDto.ID}}
		// update := bson.D{{"$set", operatorDto}}
		result, err := EventCollection.ReplaceOne(Ctx, filter, eventDto)
		if err != nil {
			log.Println("UpdateEventDtos: FAILED to UPDATE Event details - ", err.Error())
			return err
		}
		log.Println("UpdateEventDtos: Matched recoreds Count - ", result.MatchedCount)
		log.Println("UpdateEventDtos: Modified recoreds Count - ", result.ModifiedCount)
	}
	return nil
}

func DeleteEventsByEventId(event string) error {
	filter := bson.D{{"event_id", event}}
	result, err := EventCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteEventsByEventId: FAILED to delete Event details - ", err.Error())
		return err
	}
	log.Println("DeleteEventsByEventId: Matched recoreds Count - ", result.DeletedCount)
	return nil
}

// Update Event Status for EventIds
func UpdateEventStatus(eventIds []string, status string) error {
	filter := bson.D{{"event_id", bson.M{"$in": eventIds}}}
	update := bson.D{{"$set", bson.M{"event_status": status}}}
	result, err := EventCollection.UpdateMany(Ctx, filter, update)
	if err != nil {
		log.Println("UpdateEventStatus: FAILED to update Event details - ", err.Error())
		return err
	}
	log.Println("UpdateEventStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateEventStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func GetEventsByCompetitionId(competitionId string) ([]models.Event, error) {
	//log.Println("GetEventsByCompetitionId: Return all events")
	// 0. Response object
	events := []models.Event{}
	// 1. Create Filter
	filter := bson.M{}
	if competitionId != "" {
		filter["competition_id"] = competitionId
	}
	// todayDate := time.Now()
	// endDate := todayDate.Add(1 * 24 * time.Hour) // 24 hour later
	// filter["$and"] = []bson.M{{"open_date": bson.M{"$gte": todayDate.UnixMilli(), "$lte": endDate.UnixMilli()}}}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetLimit(500)
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetEventsByCompetitionId: documents NOT FOUND!")
		return events, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetEventsByCompetitionId: Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func GetEventsByEventKeys(eventKeys []string, competitionId string) ([]models.Event, error) {
	//log.Println("GetEventsByCompetitionId: Return all events")
	// 0. Response object
	events := []models.Event{}
	// 1. Create Filter
	filter := bson.M{}
	filter["event_key"] = bson.M{"$in": eventKeys}
	if competitionId != "" {
		filter["competition_id"] = competitionId
	}
	// todayDate := time.Now()
	// endDate := todayDate.Add(1 * 24 * time.Hour) // 24 hour later
	// filter["$and"] = []bson.M{{"open_date": bson.M{"$gte": todayDate.UnixMilli(), "$lte": endDate.UnixMilli()}}}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := EventCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetEventsByCompetitionId: documents NOT FOUND!")
		return events, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.Event{}
		err := cursor.Decode(&event)
		if err != nil {
			log.Println("GetEventsByCompetitionId: Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}
