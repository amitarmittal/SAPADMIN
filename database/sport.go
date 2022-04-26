package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get Sport Document
func GetSport(sportKey string) (models.Sport, error) {
	//log.Println("GetSport: Looking for sport_key - ", sportKey)
	sport := models.Sport{}
	err := SportCollection.FindOne(Ctx, bson.M{"sport_key": sportKey}).Decode(&sport)
	if err != nil {
		log.Println("GetSport: Get Sport FAILED with error - ", err.Error())
		return sport, err
	}
	return sport, nil
}

// Insert Sport Document
func InsertSport(sport models.Sport) error {
	//log.Println("InsertSport: Adding Documnet for sportId - ", sportDto.SportId)
	sport.CreatedAt = time.Now().Unix()
	sport.UpdatedAt = sport.CreatedAt
	_, err := SportCollection.InsertOne(Ctx, sport)
	if err != nil {
		log.Println("InsertSport: FAILED to INSERT sport - ", err.Error())
		return err
	}
	//log.Println("InsertSport: Sport Document _id is - ", result.InsertedID)
	return nil
}

// Bulk insert Sport objects
// Use Cases:
// 2. When added new Sports
func InsertManySports(sports []models.Sport) error {
	//log.Println("InsertManySports: Adding Sports, count is - ", len(sports))
	proStatus := []interface{}{}
	for _, proSts := range sports {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	result, err := SportCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManySports: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertManySports: Inserted count is - ", len(result.InsertedIDs))
	return nil
}

// Update Sport Document
func UpdateSport(sport models.Sport) error {
	//log.Println("UpdateSport: Updating Document for sport_key - ", sport.SportKey)
	opts := options.Update()
	filter := bson.D{{"sport_key", sport.SportKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"status", sport.Status}, {"updated_at", updatedAt}}}}
	result, err := SportCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateSport: FAILED to UPDATE sport - ", err.Error())
		return err
	}
	log.Println("UpdateSport: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateSport: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Sports by ProviderId
func GetSports(providerId string) ([]models.Sport, error) {
	//log.Println("GetSports: Looking Sports for ProviderId : ", providerId)
	// 0. Response object
	sports := []models.Sport{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	//findOptions := options.Find()
	//findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	//cursor, err := SportCollection.Find(Ctx, filter, findOptions)
	cursor, err := SportCollection.Find(Ctx, filter)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetSports: Sports NOT FOUND for ProviderId : ", providerId)
		log.Println("GetSports: SportCollection.Find failed with error : ", err.Error())
		return sports, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.Sport{}
		err := cursor.Decode(&sport)
		if err != nil {
			log.Println("GetSports: Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

func GetAllSports() ([]models.Sport, error) {
	//log.Println("GetUpdatedSports: Return last 100 documents")
	// 0. Response object
	sports := []models.Sport{}
	// 1. Create Filter
	filter := bson.D{}
	// 2. Create Find options - add sort
	//findOptions := options.Find()
	//findOptions.SetSort(bson.D{{"updated_at", -1}})
	//findOptions.SetLimit(100)
	// 3. Execute Query
	//cursor, err := SportCollection.Find(Ctx, filter, findOptions)
	cursor, err := SportCollection.Find(Ctx, filter)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetUpdatedSports: documents NOT FOUND!")
		return sports, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		sport := models.Sport{}
		err := cursor.Decode(&sport)
		if err != nil {
			log.Println("GetUpdatedSports: Decode failed with error - ", err.Error())
			continue
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// func ReplaceSport(sport models.Sport) error {
// 	//log.Println("ReplaceSport: Updating Document for sport_key - ", sport.SportKey)
// 	filter := bson.D{{"_id", sport.ID}}
// 	updatedAt := time.Now().Unix()
// 	update := bson.D{{"$set", bson.D{{"status", sport.Status}, {"updated_at", updatedAt}}}}
// 	result, err := SportCollection.ReplaceOne(Ctx, filter, update)
// 	if err != nil {
// 		log.Println("ReplaceSport: FAILED to UPDATE sport - ", err.Error())
// 		return err
// 	}
// 	log.Println("ReplaceSport: Matched recoreds Count - ", result.MatchedCount)
// 	log.Println("ReplaceSport: Modified recoreds Count - ", result.ModifiedCount)
// 	return nil
// }

func ReplaceSport(sportDto models.Sport) error {
	//log.Println("UpdateSessionDetails: Updating Documnet for operatorId - ", operatorDto.OperatorId)
	// opts := options.Update()
	sportDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", sportDto.ID}}
	result, err := SportCollection.ReplaceOne(Ctx, filter, sportDto)
	if err != nil {
		log.Println("ReplaceSport: FAILED to UPDATE spoty details - ", err.Error())
		return err
	}
	log.Println("ReplaceSport: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceSport: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Sports updated in last 5 minutes
func GetUpdatedSports() ([]models.Sport, error) {
	retObjs := []models.Sport{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := SportCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedSports: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.Sport{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedSports: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

func DeleteSportsBySportId(sportId string) error {
	//log.Println("DeleteSportsByOperator: Deleting Sport Status for operatorId - ", operatorId)
	filter := bson.D{{"sport_id", sportId}}
	result, err := SportCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteSportsBySportId: FAILED to DELETE Sport Status for operatorId - ", sportId)
		return err
	}
	log.Println("DeleteSportsBySportId: Deleted recoreds Count - ", result.DeletedCount)
	return nil
}
