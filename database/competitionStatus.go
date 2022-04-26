package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert CompetitionStatus Document
func InsertCompetitionStatus(competitionStatus models.CompetitionStatus) error {
	//log.Println("InsertCompetitionStatus: Adding Documnet for eventId - ", eventDto.EventId)
	competitionStatus.CreatedAt = time.Now().Unix()
	competitionStatus.UpdatedAt = competitionStatus.CreatedAt
	_, err := CompetitionStatusCollection.InsertOne(Ctx, competitionStatus)
	if err != nil {
		log.Println("InsertCompetitionStatus: FAILED to INSERT competitionStatus - ", err.Error())
		return err
	}
	//log.Println("InsertCompetitionStatus: Event Document _id is - ", result.InsertedID)
	return nil
}

// Get CompetitionStatus By CompetitionKey
func GetCompetitionStatus(competitionKey string) (models.CompetitionStatus, error) {
	//log.Println("GetCompetitionStatus: Looking for CompetitionKey - ", competitionKey)
	competitionStatus := models.CompetitionStatus{}
	err := CompetitionStatusCollection.FindOne(Ctx, bson.M{"competition_key": competitionKey}).Decode(&competitionStatus)
	if err != nil {
		//log.Println("GetCompetitionStatus: Failed with error - ", err.Error())
		return competitionStatus, err
	}
	return competitionStatus, nil
}

// Get CompetitionStatus By CompetitionKey
func GetCompetitionStatusByKeys(competitionKeys []string) ([]models.CompetitionStatus, error) {
	//log.Println("GetCompetitionStatus: Looking for CompetitionKey - ", competitionKey)
	// 0. Response object
	competitionStatuses := []models.CompetitionStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	if len(competitionKeys) == 0 {
		return competitionStatuses, nil
	}
	filter["competition_key"] = bson.M{"$in": competitionKeys}
	// 3. Execute Query
	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUpdatedCompetitionStatus: Failed with error - ", err.Error())
		return competitionStatuses, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competitionStatus := models.CompetitionStatus{}
		err = cursor.Decode(&competitionStatus)
		if err != nil {
			log.Println("GetUpdatedCompetitionStatus: Decode failed with error - ", err.Error())
			continue
		}
		competitionStatuses = append(competitionStatuses, competitionStatus)
	}
	return competitionStatuses, nil
}

// Get Competitions By OperatorId
// func GetOpCompetitions(operatorId string) ([]models.CompetitionStatus, error) {
// 	//log.Println("GetOpCompetitions: Looking for OperatorId - ", operatorId)
// 	// 0. Response object
// 	competitions := []models.CompetitionStatus{}
// 	// 1. Create Filter
// 	filter := bson.M{}
// 	filter["operator_id"] = operatorId
// 	// 2. Create Find options - add sort
// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"created_at", 1}})
// 	// 3. Execute Query
// 	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetOpCompetitions: Failed with error - ", err.Error())
// 		return competitions, err
// 	}
// 	defer cursor.Close(Ctx)
// 	// 4. Iterate through cursor
// 	for cursor.Next(Ctx) {
// 		competition := models.CompetitionStatus{}
// 		err = cursor.Decode(&competition)
// 		if err != nil {
// 			log.Println("GetOpCompetitions: CompetitionStatus Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		competitions = append(competitions, competition)
// 	}
// 	return competitions, nil
// }

// Get Competitions By ProviderId
// func GetPrCompetitions(providerId string) ([]models.CompetitionStatus, error) {
// 	//log.Println("GetPrCompetitions: Looking for ProviderId - ", providerId)
// 	// 0. Response object
// 	sports := []models.CompetitionStatus{}
// 	// 1. Create Filter
// 	filter := bson.M{}
// 	filter["provider_id"] = providerId
// 	// 2. Create Find options - add sort
// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"created_at", 1}})
// 	// 3. Execute Query
// 	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetPrCompetitions: Failed with error - ", err.Error())
// 		return sports, err
// 	}
// 	defer cursor.Close(Ctx)
// 	// 4. Iterate through cursor
// 	for cursor.Next(Ctx) {
// 		sport := models.CompetitionStatus{}
// 		err = cursor.Decode(&sport)
// 		if err != nil {
// 			log.Println("GetPrCompetitions: CompetitionStatus Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		sports = append(sports, sport)
// 	}
// 	return sports, nil
// }

// Get Competitions By OperatorId & ProviderId & SportId
func GetOpPrCompetitions(operatorId string, providerId string, sportId string) ([]models.CompetitionStatus, error) {
	//log.Println("GetOpPrCompetitions: Looking for OperatorId - ", operatorId)
	// 0. Response object
	competitions := []models.CompetitionStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_id"] = operatorId
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOpPrCompetitions: Failed with error - ", err.Error())
		return competitions, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.CompetitionStatus{}
		err = cursor.Decode(&competition)
		if err != nil {
			log.Println("GetOpPrCompetitions: CompetitionStatus Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

// func GetCompetitionStatusByPrSport(providerId string, sportId string) ([]models.CompetitionStatus, error) {
// 	//log.Println("GetOpPrCompetitions: Looking for OperatorId - ", operatorId)
// 	// 0. Response object
// 	competitions := []models.CompetitionStatus{}
// 	// 1. Create Filter
// 	filter := bson.M{}
// 	filter["provider_id"] = providerId
// 	filter["sport_id"] = sportId
// 	// 2. Create Find options - add sort
// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"created_at", 1}})
// 	// 3. Execute Query
// 	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetOpPrCompetitions: Failed with error - ", err.Error())
// 		return competitions, err
// 	}
// 	defer cursor.Close(Ctx)
// 	// 4. Iterate through cursor
// 	for cursor.Next(Ctx) {
// 		competition := models.CompetitionStatus{}
// 		err = cursor.Decode(&competition)
// 		if err != nil {
// 			log.Println("GetOpPrCompetitions: CompetitionStatus Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		competitions = append(competitions, competition)
// 	}
// 	return competitions, nil
// }

// Get Competitions By OperatorId & ProviderId & SportId
func GetOperatorsFromCompPr(competition_id string, providerId string) ([]models.CompetitionStatus, error) {
	// 0. Response object
	competitions := []models.CompetitionStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["competition_id"] = competition_id
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOperatorsFromCompPr: Failed with error - ", err.Error())
		return competitions, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.CompetitionStatus{}
		err = cursor.Decode(&competition)
		if err != nil {
			log.Println("GetOperatorsFromCompPr: CompetitionStatus Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

// Bulk insert Competition Staus objects
// Use Cases:
// 1. When added new Operator
// 2. When added new Competition
// func InsertManyCompetitionStatus(competitionStatus []models.CompetitionStatus) error {
// 	//log.Println("InsertManyCompetitionStatus: Adding CompetitionStatus, count is - ", len(competitionStatus))
// 	proStatus := []interface{}{}
// 	for _, proSts := range competitionStatus {
// 		proSts.CreatedAt = time.Now().Unix()
// 		proSts.UpdatedAt = proSts.CreatedAt
// 		proStatus = append(proStatus, proSts)
// 	}
// 	result, err := CompetitionStatusCollection.InsertMany(Ctx, proStatus)
// 	if err != nil {
// 		log.Println("InsertManyCompetitionStatus: FAILED to INSERT - ", err.Error())
// 		return err
// 	}
// 	log.Println("InsertManyCompetitionStatus: Inserted count is - ", len(result.InsertedIDs))
// 	return nil
// }

// Update Operator Status
// func UpdateOACompetitionStatus(competitionKey string, status string) error {
// 	//if status == "ACTIVE" {
// 	//	log.Println("UpdateOACompetitionStatus: Unblocking the competition - ", competitionKey)
// 	//} else {
// 	//	log.Println("UpdateOACompetitionStatus: Status changing for the competitionKey: ", competitionKey+"-"+status)
// 	//}
// 	filter := bson.D{{"competition_key", competitionKey}}
// 	updatedAt := time.Now().Unix()
// 	update := bson.D{{"$set", bson.D{{"operator_status", status}, {"updated_at", updatedAt}}}}
// 	opts := options.Update()
// 	result, err := CompetitionStatusCollection.UpdateOne(Ctx, filter, update, opts)
// 	if err != nil {
// 		log.Println("UpdateOACompetitionStatus: FAILED to UPDATE operator status - ", err.Error())
// 		return err
// 	}
// 	log.Println("UpdateOACompetitionStatus: Matched recoreds Count - ", result.MatchedCount)
// 	log.Println("UpdateOACompetitionStatus: Modified recoreds Count - ", result.ModifiedCount)
// 	return nil
// }

// Update Provider Status
// func UpdatePACompetitionStatus(competitionKey string, status string) error {
// 	//if status == "ACTIVE" {
// 	//	log.Println("UpdatePACompetitionStatus: Unblocking the competition - ", competitionKey)
// 	//} else {
// 	//	log.Println("UpdatePACompetitionStatus: Status changing for the competitionKey: ", competitionKey+"-"+status)
// 	//}
// 	filter := bson.D{{"competition_key", competitionKey}}
// 	updatedAt := time.Now().Unix()
// 	update := bson.D{{"$set", bson.D{{"provider_status", status}, {"updated_at", updatedAt}}}}
// 	opts := options.Update()
// 	result, err := CompetitionStatusCollection.UpdateOne(Ctx, filter, update, opts)
// 	if err != nil {
// 		log.Println("UpdatePACompetitionStatus: FAILED to UPDATE provider status - ", err.Error())
// 		return err
// 	}
// 	log.Println("UpdatePACompetitionStatus: Matched recoreds Count - ", result.MatchedCount)
// 	log.Println("UpdatePACompetitionStatus: Modified recoreds Count - ", result.ModifiedCount)
// 	return nil
// }

func GetUpdatedCompetitionStatus(competitionKeys []string) ([]models.CompetitionStatus, error) {
	//log.Println("GetUpdatedCompetitionStatus: Looking for Keys Count - ", len(competitionKeys))
	// 0. Response object
	competitionStatuses := []models.CompetitionStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	if len(competitionKeys) != 0 {
		filter["competition_key"] = bson.M{"$in": competitionKeys}
	} else {
		findOptions.SetLimit(100)
	}
	// 3. Execute Query
	cursor, err := CompetitionStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUpdatedCompetitionStatus: Failed with error - ", err.Error())
		return competitionStatuses, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		cs := models.CompetitionStatus{}
		err = cursor.Decode(&cs)
		if err != nil {
			log.Println("GetUpdatedCompetitionStatus: Decode failed with error - ", err.Error())
			continue
		}
		competitionStatuses = append(competitionStatuses, cs)
	}
	return competitionStatuses, nil
}

// Update Competition Document
func ReplaceCompetitionStatus(competitionDto models.CompetitionStatus) error {
	//log.Println("UpdateSessionDetails: Updating Documnet for operatorId - ", operatorDto.OperatorId)
	// opts := options.Update()
	competitionDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", competitionDto.ID}}
	// update := bson.D{{"$set", operatorDto}}
	result, err := CompetitionStatusCollection.ReplaceOne(Ctx, filter, competitionDto)
	if err != nil {
		log.Println("ReplaceCompetitionStatus: FAILED to UPDATE competition details - ", err.Error())
		return err
	}
	log.Println("ReplaceCompetitionStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceCompetitionStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func DeleteCompetitionsByOperator(operatorId string) error {
	filter := bson.D{{"operator_id", operatorId}}
	result, err := CompetitionStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteCompetitionsByOperator: FAILED to DELETE competition details - ", err.Error())
		return err
	}
	log.Println("DeleteCompetitionsByOperator: Matched recoreds Count - ", result.DeletedCount)
	return nil
}

func DeleteCompetitionstatusBycompetitionId(competition_id string) error {
	filter := bson.D{{"competition_id", competition_id}}
	result, err := CompetitionStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteCompetitionsBycompId: FAILED to DELETE competition details - ", err.Error())
		return err
	}
	log.Println("DeleteCompetitionsBycompId: Matched recoreds Count - ", result.DeletedCount)
	return nil
}

// Get CompetitionStatuss updated in last 5 minutes
func GetUpdatedCompetitionStatuss() ([]models.CompetitionStatus, error) {
	retObjs := []models.CompetitionStatus{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := CompetitionStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedCompetitionStatuss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.CompetitionStatus{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedCompetitionStatuss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}
