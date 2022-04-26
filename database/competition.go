package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get Competition Document
func GetCompetition(competitionKey string) (models.Competition, error) {
	//log.Println("GetCompetition: Looking for competition_key - ", competitionKey)
	competition := models.Competition{}
	err := CompetitionCollection.FindOne(Ctx, bson.M{"competition_key": competitionKey}).Decode(&competition)
	if err != nil {
		log.Println("GetCompetition: Get Competition FAILED with error - ", err.Error())
		return competition, err
	}
	return competition, nil
}

// Get Competition Document
func GetCompetitionsByKeys(competitionKeys []string) ([]models.Competition, error) {
	//log.Println("GetCompetition: Looking for competition_key - ", competitionKey)
	competitions := []models.Competition{}
	if len(competitionKeys) == 0 {
		return competitions, nil
	}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	filter["competition_key"] = bson.M{"$in": competitionKeys}
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetCompetitionsByKeys: Failed with error - ", err.Error())
		return competitions, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err = cursor.Decode(&competition)
		if err != nil {
			log.Println("GetCompetitionsByKeys: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

// Insert Competition Document
func InsertCompetition(competition models.Competition) error {
	//log.Println("InsertCompetition: Adding Documnet for eventId - ", eventDto.EventId)
	competition.CreatedAt = time.Now().Unix()
	competition.UpdatedAt = competition.CreatedAt
	_, err := CompetitionCollection.InsertOne(Ctx, competition)
	if err != nil {
		log.Println("InsertCompetition: FAILED to INSERT competition - ", err.Error())
		return err
	}
	//log.Println("InsertCompetition: Event Document _id is - ", result.InsertedID)
	return nil
}

// Bulk insert Competition objects
// Use Cases:
// 2. When added new Competitions
func InsertManyCompetitions(competitions []models.Competition) error {
	//log.Println("InsertManyCompetitions: Adding Competitions, count is - ", len(competitions))
	proStatus := []interface{}{}
	for _, proSts := range competitions {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	result, err := CompetitionCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManyCompetitions: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertManyCompetitions: Inserted count is - ", len(result.InsertedIDs))
	return nil
}

// Update Competition Document
func UpdateCompetition(competition models.Competition) error {
	//log.Println("UpdateCompetition: Updating Documnet for competition_key - ", competition.CompetitionKey)
	opts := options.Update()
	filter := bson.D{{"competition_key", competition.CompetitionKey}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"status", competition.Status}, {"updated_at", updatedAt}}}}
	result, err := CompetitionCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateCompetition: FAILED to UPDATE competition - ", err.Error())
		return err
	}
	log.Println("UpdateCompetition: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateCompetition: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Competitions by ProviderId
func GetCompetitionsbySport(providerId string, sportId string) ([]models.Competition, error) {
	//log.Println("GetCompetitions: Looking Competitions for ProviderId : ", providerId)
	// 0. Response object
	competitions := []models.Competition{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetCompetitions: Competitions NOT FOUND for ProviderId : ", providerId)
		return competitions, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err := cursor.Decode(&competition)
		if err != nil {
			log.Println("GetCompetitions: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

// Get Competitions by ProviderId
func GetCompetitionsbySportLast10(providerId string, sportId string) ([]models.Competition, error) {
	//log.Println("GetCompetitions: Looking Competitions for ProviderId : ", providerId)
	// 0. Response object
	competitions := []models.Competition{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	filter["sport_id"] = sportId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	findOptions.SetLimit(10)
	// 3. Execute Query
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetCompetitions: Competitions NOT FOUND for ProviderId : ", providerId)
		return competitions, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err := cursor.Decode(&competition)
		if err != nil {
			log.Println("GetCompetitions: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

// Get Competitions by ProviderId & SportId
func GetCompetitions(providerId string) ([]models.Competition, error) {
	//log.Println("GetCompetitions: Looking Competitions for ProviderId : ", providerId)
	// 0. Response object
	competitions := []models.Competition{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	// 3. Execute Query
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetCompetitions: Competitions NOT FOUND for ProviderId : ", providerId)
		return competitions, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err := cursor.Decode(&competition)
		if err != nil {
			log.Println("GetCompetitions: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

func GetLatestCompetitions() ([]models.Competition, error) {
	//log.Println("GetLatestCompetitions: Return last 100 competitions")
	// 0. Response object
	competitions := []models.Competition{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	findOptions.SetLimit(100)
	// 3. Execute Query
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetLatestCompetitions: Competitions NOT FOUND!")
		return competitions, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err := cursor.Decode(&competition)
		if err != nil {
			log.Println("GetLatestEvents: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

func GetUpdatedCompetitions() ([]models.Competition, error) {
	//log.Println("GetUpdatedCompetitions: Return last 100 documents")
	// 0. Response object
	competitions := []models.Competition{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	findOptions.SetLimit(1000)
	// 3. Execute Query
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetUpdatedCompetitions: documents NOT FOUND!")
		return competitions, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err := cursor.Decode(&competition)
		if err != nil {
			log.Println("GetUpdatedCompetitions: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

func GetAllCompetitions(providerId string) ([]models.Competition, error) {
	//log.Println("GetAllCompetitions: Return all competitions")
	// 0. Response object
	competitions := []models.Competition{}
	// 1. Create Filter
	filter := bson.M{}
	filter["provider_id"] = providerId
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", -1}})
	// 3. Execute Query
	cursor, err := CompetitionCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllCompetitions: documents NOT FOUND!")
		return competitions, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		competition := models.Competition{}
		err := cursor.Decode(&competition)
		if err != nil {
			log.Println("GetAllCompetitions: Decode failed with error - ", err.Error())
			continue
		}
		competitions = append(competitions, competition)
	}
	return competitions, nil
}

// func ReplaceCompetition
func ReplaceCompetition(competitionDto models.Competition) error {
	//log.Println("UpdateSessionDetails: Updating Documnet for operatorId - ", operatorDto.OperatorId)
	// opts := options.Update()
	competitionDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", competitionDto.ID}}
	result, err := CompetitionCollection.ReplaceOne(Ctx, filter, competitionDto)
	if err != nil {
		log.Println("ReplaceCompetition: FAILED to UPDATE spoty details - ", err.Error())
		return err
	}
	log.Println("ReplaceCompetition: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceCompetition: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Competitions updated in last 5 minutes
func GetUpdatedCompetitionss() ([]models.Competition, error) {
	retObjs := []models.Competition{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := CompetitionCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedCompetitionss: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.Competition{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedCompetitionss: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

func DeleteCompetitionsBycompetitionId(competition_id string) error {
	filter := bson.D{{"competition_id", competition_id}}
	result, err := CompetitionCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeleteCompetitionsBycompetitionId: FAILED to DELETE competition details - ", err.Error())
		return err
	}
	log.Println("DeleteCompetitionsBycompetitionId: Matched recoreds Count - ", result.DeletedCount)
	return nil
}
