package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert EventStatus Document
func InsertEventStatus(eventStatus models.EventStatus) error {
	//log.Println("InsertEventStatus: Adding Documnet for eventId - ", eventDto.EventId)
	eventStatus.CreatedAt = time.Now().Unix()
	eventStatus.UpdatedAt = eventStatus.CreatedAt
	result, err := EventStatusCollection.InsertOne(Ctx, eventStatus)
	if err != nil {
		log.Println("EventStatusCollection: InsertEventStatus: FAILED to INSERT eventStatus - ", err.Error())
		return err
	}
	log.Println("EventStatusCollection: InsertEventStatus: Event Document _id is - ", result.InsertedID)
	return nil
}

// Get EventStatus By EventKey
func GetEventStatus(eventKey string) (models.EventStatus, error) {
	//log.Println("GetEventStatus: Looking for EventKey - ", eventKey)
	eventStatus := models.EventStatus{}
	err := EventStatusCollection.FindOne(Ctx, bson.M{"event_key": eventKey}).Decode(&eventStatus)
	if err != nil {
		// log.Println("GetEventStatus: Failed with error - ", err.Error())
		return eventStatus, err
	}
	return eventStatus, nil
}

// Get Events By OperatorId
// func GetLatestEventStatus() ([]models.EventStatus, error) {
// 	//log.Println("GetLatestEventStatus: Trying to fetch 500 documents sort by updatedAt DESC")
// 	// 0. Response object
// 	events := []models.EventStatus{}
// 	// 1. Create Filter
// 	filter := bson.M{}
// 	// 2. Create Find options - add sort
// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"updated_at", -1}})
// 	findOptions.SetLimit(500)
// 	// 3. Execute Query
// 	cursor, err := EventStatusCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetLatestEventStatus: Failed with error - ", err.Error())
// 		return events, err
// 	}
// 	defer cursor.Close(Ctx)
// 	// 4. Iterate through cursor
// 	// err = cursor.All(Ctx, &events)
// 	for cursor.Next(Ctx) {
// 		event := models.EventStatus{}
// 		err = cursor.Decode(&event)
// 		if err != nil {
// 			log.Println("GetLatestEventStatus: EventStatus Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		events = append(events, event)
// 	}
// 	return events, nil
// }

// Get Events By OperatorId
// func GetOpEvents(operatorId string) ([]models.EventStatus, error) {
// 	//log.Println("GetOpEvents: Looking for OperatorId - ", operatorId)
// 	// 0. Response object
// 	events := []models.EventStatus{}
// 	// 1. Create Filter
// 	filter := bson.M{}
// 	filter["operator_id"] = operatorId
// 	// 2. Create Find options - add sort
// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"created_at", 1}})
// 	// 3. Execute Query
// 	cursor, err := EventStatusCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetOpEvents: Failed with error - ", err.Error())
// 		return events, err
// 	}
// 	defer cursor.Close(Ctx)
// 	// 4. Iterate through cursor
// 	for cursor.Next(Ctx) {
// 		event := models.EventStatus{}
// 		err = cursor.Decode(&event)
// 		if err != nil {
// 			log.Println("GetOpEvents: EventStatus Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		events = append(events, event)
// 	}
// 	return events, nil
// }

// Get Events By ProviderId
// func GetPrEvents(providerId string) ([]models.EventStatus, error) {
// 	//log.Println("GetPrEvents: Looking for ProviderId - ", providerId)
// 	// 0. Response object
// 	sports := []models.EventStatus{}
// 	// 1. Create Filter
// 	filter := bson.M{}
// 	filter["provider_id"] = providerId
// 	// 2. Create Find options - add sort
// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"created_at", 1}})
// 	// 3. Execute Query
// 	cursor, err := EventStatusCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetPrEvents: Failed with error - ", err.Error())
// 		return sports, err
// 	}
// 	defer cursor.Close(Ctx)
// 	// 4. Iterate through cursor
// 	for cursor.Next(Ctx) {
// 		sport := models.EventStatus{}
// 		err = cursor.Decode(&sport)
// 		if err != nil {
// 			log.Println("GetPrEvents: EventStatus Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		sports = append(sports, sport)
// 	}
// 	return sports, nil
// }

// Get Events By OperatorId & ProviderId & SportId
func GetOpPrEvents(operatorId string, providerId string, sportId string, competitionId string) ([]models.EventStatus, error) {
	//log.Println("GetOpPrEvents: Looking for OperatorId - ", operatorId)
	// 0. Response object
	events := []models.EventStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_id"] = operatorId
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	if competitionId != "" {
		filter["competition_id"] = competitionId
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := EventStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("EventStatusCollection: GetOpPrEvents: Failed with error - ", err.Error())
		return events, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.EventStatus{}
		err = cursor.Decode(&event)
		if err != nil {
			log.Println("EventStatusCollection: GetOpPrEvents: EventStatus Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func GetOperatorsFromEvPr(eventId string, providerId string) ([]models.EventStatus, error) {
	// 0. Response object
	events := []models.EventStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["event_id"] = eventId
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := EventStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("EventStatusCollection: GetOperatorsFromEvPr: Failed with error - ", err.Error())
		return events, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		event := models.EventStatus{}
		err = cursor.Decode(&event)
		if err != nil {
			log.Println("EventStatusCollection: GetOperatorsFromEvPr: EventStatus Decode failed with error - ", err.Error())
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

// Bulk insert Event Staus objects
// Use Cases:
// 1. When added new Operator
// 2. When added new Event
// func InsertManyEventStatus(eventStatus []models.EventStatus) error {
// 	//log.Println("InsertManyEventStatus: Adding EventStatus, count is - ", len(eventStatus))
// 	proStatus := []interface{}{}
// 	for _, proSts := range eventStatus {
// 		proSts.CreatedAt = time.Now().Unix()
// 		proSts.UpdatedAt = proSts.CreatedAt
// 		proStatus = append(proStatus, proSts)
// 	}
// 	result, err := EventStatusCollection.InsertMany(Ctx, proStatus)
// 	if err != nil {
// 		log.Println("InsertManyEventStatus: FAILED to INSERT - ", err.Error())
// 		return err
// 	}
// 	log.Println("InsertManyEventStatus: Inserted count is - ", len(result.InsertedIDs))
// 	return nil
// }

// Update Operator Status
// func UpdateOAEventStatus(eventKey string, status string) error {
// 	//if status == "ACTIVE" {
// 	//	log.Println("UpdateOAEventStatus: Unblocking the event - ", eventKey)
// 	//} else {
// 	//	log.Println("UpdateOAEventStatus: Status changing for the eventKey: ", eventKey+"-"+status)
// 	//}
// 	filter := bson.D{{"event_key", eventKey}}
// 	updatedAt := time.Now().Unix()
// 	update := bson.D{{"$set", bson.D{{"operator_status", status}, {"updated_at", updatedAt}}}}
// 	opts := options.Update()
// 	result, err := EventStatusCollection.UpdateOne(Ctx, filter, update, opts)
// 	if err != nil {
// 		log.Println("EventStatusCollection: UpdateOAEventStatus: FAILED to UPDATE operator status - ", err.Error())
// 		return err
// 	}
// 	log.Println("EventStatusCollection: UpdateOAEventStatus: Matched recoreds Count - ", result.MatchedCount)
// 	log.Println("EventStatusCollection: UpdateOAEventStatus: Modified recoreds Count - ", result.ModifiedCount)
// 	return nil
// }

// Update Provider Status
// func UpdatePAEventStatus(eventKey string, status string) error {
// 	//if status == "ACTIVE" {
// 	//	log.Println("UpdatePAEventStatus: Unblocking the event - ", eventKey)
// 	//} else {
// 	//	log.Println("UpdatePAEventStatus: Status changing for the eventKey: ", eventKey+"-"+status)
// 	//}
// 	filter := bson.D{{"event_key", eventKey}}
// 	updatedAt := time.Now().Unix()
// 	update := bson.D{{"$set", bson.D{{"provider_status", status}, {"updated_at", updatedAt}}}}
// 	opts := options.Update()
// 	result, err := EventStatusCollection.UpdateOne(Ctx, filter, update, opts)
// 	if err != nil {
// 		log.Println("EventStatusCollection: UpdatePAEventStatus: FAILED to UPDATE provider status - ", err.Error())
// 		return err
// 	}
// 	log.Println("EventStatusCollection: UpdatePAEventStatus: Matched recoreds Count - ", result.MatchedCount)
// 	log.Println("EventStatusCollection: UpdatePAEventStatus: Modified recoreds Count - ", result.ModifiedCount)
// 	return nil
// }

func GetUpdatedEventStatus(eventKeys []string) ([]models.EventStatus, error) {
	//log.Println("GetUpdatedEventStatus: Looking for Keys Count - ", len(eventKeys))
	// 0. Response object
	eventStatuses := []models.EventStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	if len(eventKeys) != 0 {
		filter["event_key"] = bson.M{"$in": eventKeys}
	} else {
		findOptions.SetLimit(300)
	}
	// 3. Execute Query
	cursor, err := EventStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("EventStatusCollection: GetUpdatedEventStatus: Failed with error - ", err.Error())
		return eventStatuses, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		eventStatus := models.EventStatus{}
		err = cursor.Decode(&eventStatus)
		if err != nil {
			log.Println("EventStatusCollection: GetUpdatedEventStatus: Decode failed with error - ", err.Error())
			continue
		}
		eventStatuses = append(eventStatuses, eventStatus)
	}
	return eventStatuses, nil
}

// Update Event Document
func ReplaceEventStatus(eventDto models.EventStatus) error {
	eventDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", eventDto.ID}}
	// update := bson.D{{"$set", operatorDto}}
	result, err := EventStatusCollection.ReplaceOne(Ctx, filter, eventDto)
	if err != nil {
		log.Println("EventStatusCollection: ReplaceEventStatus: FAILED to UPDATE Event details - ", err.Error())
		return err
	}
	log.Println("EventStatusCollection: ReplaceEventStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("EventStatusCollection: ReplaceEventStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func DeleteEventsByOperator(operatorId string) error {
	filter := bson.D{{"operator_id", operatorId}}
	result, err := EventStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("EventStatusCollection: DeleteEventsByOperator: FAILED to DELETE Event details - ", err.Error())
		return err
	}
	log.Println("EventStatusCollection: DeleteEventsByOperator: Matched recoreds Count - ", result.DeletedCount)
	return nil
}

func DeleteEventstatusByEventId(EventId string) error {
	filter := bson.D{{"event_id", EventId}}
	result, err := EventStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("EventStatusCollection: DeleteEventsByEventId: FAILED to DELETE Event details - ", err.Error())
		return err
	}
	log.Println("EventStatusCollection: DeleteEventsByEventId: Matched recoreds Count - ", result.DeletedCount)
	return nil
}

// Get EventStatuss updated in last 5 minutes
func GetUpdatedEventStatuss() ([]models.EventStatus, error) {
	retObjs := []models.EventStatus{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := EventStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("EventStatusCollection: GetUpdatedEventStatuss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.EventStatus{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("EventStatusCollection: GetUpdatedEventStatuss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

// func UpdateEventStatusDtos(eventStatusDtos []models.EventStatus) error {
// 	// IDs := []primitive.ObjectID{}
// 	// for _, eventStatusDto := range eventStatusDtos {
// 	// 	IDs = append(IDs, eventStatusDto.ID)
// 	// }
// 	// filter := bson.M{"_id": bson.M{"$in": IDs}}
// 	// update := bson.D{{"$set", eventStatusDtos}}
// 	// EventStatusCollection.UpdateMany(Ctx, filter, update)
// 	for _, eventStatusDto := range eventStatusDtos {
// 		err := ReplaceEventStatus(eventStatusDto)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
