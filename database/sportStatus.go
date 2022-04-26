package database

import (
	"Sp/dto/models"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get SportStatus By SportKey
func GetSportStatus(sportKey string) (models.SportStatus, error) {
	//log.Println("GetSportStatus: Looking for SportKey - ", sportKey)
	sportStatus := models.SportStatus{}
	err := SportStatusCollection.FindOne(Ctx, bson.M{"sport_key": sportKey}).Decode(&sportStatus)
	if err != nil {
		log.Println("GetSportStatus: Failed with error - ", err.Error())
		return sportStatus, err
	}
	return sportStatus, nil
}

// Get Sports By OperatorId
func GetOpSports(operatorId string, partnerId string) ([]models.SportStatus, error) {
	//log.Println("GetOpSports: Looking for OperatorId - ", operatorId)
	// 0. Response object
	sports := []models.SportStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_id"] = operatorId
	if partnerId != "" {
		filter["partner_id"] = partnerId
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := SportStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOpSports: Failed with error - ", err.Error())
		return sports, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.SportStatus{}
		err = cursor.Decode(&sport)
		if err != nil {
			log.Println("GetOpSports: SportStatus Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// Get Sports By ProviderId
func GetPrSports(providerId string) ([]models.SportStatus, error) {
	//log.Println("GetPrSports: Looking for ProviderId - ", providerId)
	// 0. Response object
	sports := []models.SportStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := SportStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetPrSports: Failed with error - ", err.Error())
		return sports, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.SportStatus{}
		err = cursor.Decode(&sport)
		if err != nil {
			log.Println("GetPrSports: SportStatus Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// Get Sports By OperatorId & ProviderId
func GetOpPrSports(operatorId string, partnerId string, providerId string) ([]models.SportStatus, error) {
	//log.Println("GetOpPrSports: Looking for OperatorId - ", operatorId)
	// 0. Response object
	sports := []models.SportStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["operator_id"] = operatorId
	filter["provider_id"] = providerId
	if partnerId != "" {
		filter["partner_id"] = partnerId
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := SportStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOpPrSports: Failed with error - ", err.Error())
		return sports, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.SportStatus{}
		err = cursor.Decode(&sport)
		if err != nil {
			log.Println("GetOpPrSports: SportStatus Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// Get Sports By SportId & ProviderId
func GetOperatorsFromSpPr(sportId string, providerId string) ([]models.SportStatus, error) {
	// 0. Response object
	sports := []models.SportStatus{}
	// 1. Create Filter
	filter := bson.M{}
	filter["sport_id"] = sportId
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := SportStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOperatorsFromSpPr: Failed with error - ", err.Error())
		return sports, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.SportStatus{}
		err = cursor.Decode(&sport)
		if err != nil {
			log.Println("GetOperatorsFromSpPr: SportStatus Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// Bulk insert Sport Staus objects
// Use Cases:
// 1. When added new Operator
// 2. When added new Sport
func InsertManySportStatus(sportStatus []models.SportStatus) error {
	//log.Println("InsertManySportStatus: Adding SportStatus, count is - ", len(sportStatus))
	proStatus := []interface{}{}
	for _, proSts := range sportStatus {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	result, err := SportStatusCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManySportStatus: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertManySportStatus: Inserted count is - ", len(result.InsertedIDs))
	return nil
}

// Update Operator Status
func UpdateOASportStatus(sportKey string, status string) error {
	//if status == "ACTIVE" {
	//	log.Println("UpdateOASportStatus: Unblocking the sport - ", sportKey)
	//} else {
	//	log.Println("UpdateOASportStatus: Status changing for the sportKey: ", sportKey+"-"+status)
	//}
	filter := bson.D{{"sport_key", sportKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"operator_status", status}, {"updated_at", updatedAt}}}}
	opts := options.Update()
	result, err := SportStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateOASportStatus: FAILED to UPDATE operator status - ", err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("Document not found!")
	}
	log.Println("UpdateOASportStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateOASportStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update Provider Status
func UpdatePASportStatus(sportKey string, status string) error {
	//if status == "ACTIVE" {
	//	log.Println("UpdatePASportStatus: Unblocking the sport - ", sportKey)
	//} else {
	//	log.Println("UpdatePASportStatus: Status changing for the sportKey: ", sportKey+"-"+status)
	//}
	filter := bson.D{{"sport_key", sportKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"provider_status", status}, {"updated_at", updatedAt}}}}
	opts := options.Update()
	result, err := SportStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdatePASportStatus: FAILED to UPDATE provider status - ", err.Error())
		return err
	}
	log.Println("UpdatePASportStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdatePASportStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func GetUpdatedSportStatus(sportKeys []string) ([]models.SportStatus, error) {
	//log.Println("GetUpdatedSportStatus: Looking for Keys Count - ", len(sportKeys))
	// 0. Response object
	sportStatuses := []models.SportStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	if len(sportKeys) != 0 {
		filter["sport_key"] = bson.M{"$in": sportKeys}
	}
	// } else {
	// 	findOptions.SetLimit(100)
	// }
	// 3. Execute Query
	cursor, err := SportStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUpdatedSportStatus: Failed with error - ", err.Error())
		return sportStatuses, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sportStatus := models.SportStatus{}
		err = cursor.Decode(&sportStatus)
		if err != nil {
			log.Println("GetUpdatedSportStatus: Decode failed with error - ", err.Error())
			continue
		}
		sportStatuses = append(sportStatuses, sportStatus)
	}
	return sportStatuses, nil
}

// Update Sport Document
func ReplaceSportStatus(sportDto models.SportStatus) error {
	//log.Println("UpdateSessionDetails: Updating Documnet for operatorId - ", operatorDto.OperatorId)
	// opts := options.Update()
	sportDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", sportDto.ID}}
	// update := bson.D{{"$set", operatorDto}}
	result, err := SportStatusCollection.ReplaceOne(Ctx, filter, sportDto)
	if err != nil {
		log.Println("ReplaceSportStatus: FAILED to UPDATE spoty details - ", err.Error())
		return err
	}
	log.Println("ReplaceSportStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceSportStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func DeleteSportsByOperator(operatorId string) error {
	//log.Println("DeleteSportsByOperator: Deleting Sport Status for operatorId - ", operatorId)
	filter := bson.D{{"operator_id", operatorId}}
	result, err := SportStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteSportsByOperator: FAILED to DELETE Sport Status for operatorId - ", operatorId)
		return err
	}
	log.Println("DeleteSportsByOperator: Deleted recoreds Count - ", result.DeletedCount)
	return nil
}

func DeleteSportstatusBySportId(sportId string) error {
	filter := bson.D{{"sport_id", sportId}}
	result, err := SportStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteSportsBySportId: FAILED to DELETE Sport Status for operatorId - ", sportId)
		return err
	}
	log.Println("DeleteSportsBySportId: Deleted recoreds Count - ", result.DeletedCount)
	return nil
}

func GetAllSportStatus() ([]models.SportStatus, error) {
	// 0. Response object
	sportStatuses := []models.SportStatus{}
	// 1. Create Filter
	filter := bson.D{}
	// 2. Execute Query
	cursor, err := SportStatusCollection.Find(Ctx, filter)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllSportStatus: documents NOT FOUND!")
		return sportStatuses, err
	}
	// 3. Iterate through cursor
	for cursor.Next(Ctx) {
		ss := models.SportStatus{}
		err := cursor.Decode(&ss)
		if err != nil {
			log.Println("GetAllSportStatus: Decode failed with error - ", err.Error())
			continue
		}
		sportStatuses = append(sportStatuses, ss)
	}
	return sportStatuses, nil
}

// Get SportStatuss updated in last 5 minutes
func GetUpdatedSportStatuss() ([]models.SportStatus, error) {
	retObjs := []models.SportStatus{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := SportStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedSportStatuss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.SportStatus{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedSportStatuss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

func DeleteSportsBySportKeys(sportKeys []string) error {
	filter := bson.M{}
	// 1 Check if sportKeys is empty
	if len(sportKeys) != 0 {
		filter["sport_key"] = bson.M{"$in": sportKeys}
	}
	// 2. Execute Query
	result, err := SportStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteSportsBySportKeys: FAILED to DELETE Sport Status for sportKey - ", sportKeys)
		return err
	}
	// 3. Check deleted count
	log.Println("DeleteSportsByOperator: Deleted recoreds Count - ", result.DeletedCount)
	return nil
}
