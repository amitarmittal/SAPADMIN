package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get PartnerStatus By PartnerKey
func GetPartnerStatus(partnerKey string) (models.PartnerStatus, error) {
	//log.Println("GetPartnerStatus: Looking for PartnerKey - ", partnerKey)
	partnerStatus := models.PartnerStatus{}
	err := PartnerStatusCollection.FindOne(Ctx, bson.M{"partner_key": partnerKey}).Decode(&partnerStatus)
	if err != nil {
		log.Println("GetPartnerStatus: Failed with error - ", err.Error())
		return partnerStatus, err
	}
	return partnerStatus, nil
}

// Get Providers By OperatorId
func GetProvidersPS(operatorId string) ([]models.PartnerStatus, error) {
	//log.Println("GetProviders: Looking for OperatorId - ", operatorId)
	partnerStatus := []models.PartnerStatus{}
	cursor, err := PartnerStatusCollection.Find(Ctx, bson.M{"operator_id": operatorId})
	if err != nil {
		log.Println("GetProviders: Failed with error - ", err.Error())
		return partnerStatus, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		ps := models.PartnerStatus{}
		err = cursor.Decode(&ps)
		if err != nil {
			log.Println("GetProviders: PartnerStatus Decode failed with error - ", err.Error())
			continue
		}
		partnerStatus = append(partnerStatus, ps)
	}
	return partnerStatus, nil
}

func GetPartnerProviders(operatorId string, partnerId string) ([]models.PartnerStatus, error) {
	//log.Println("GetPartners: Looking for OperatorId - ", operatorId)
	partners := []models.PartnerStatus{}
	cursor, err := PartnerStatusCollection.Find(Ctx, bson.M{"operator_id": operatorId, "partner_id": partnerId})
	if err != nil {
		log.Println("GetPartners: Failed with error - ", err.Error())
		return partners, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		partner := models.PartnerStatus{}
		err = cursor.Decode(&partner)
		if err != nil {
			log.Println("GetPartners: PartnerStatus Decode failed with error - ", err.Error())
			continue
		}
		partners = append(partners, partner)
	}
	return partners, nil
}

func GetPartnersByProviderId(providerId string) ([]models.PartnerStatus, error) {
	//log.Println("GetPartnersByProviderId: Looking for providerId - ", providerId)
	partners := []models.PartnerStatus{}
	cursor, err := PartnerStatusCollection.Find(Ctx, bson.M{"provider_id": providerId})

	if err != nil {
		log.Println("GetPartnersByProviderId: Failed with error - ", err.Error())
		return partners, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		partner := models.PartnerStatus{}
		err = cursor.Decode(&partner)
		if err != nil {
			log.Println("GetPartnersByProviderId: PartnerStatus Decode failed with error - ", err.Error())
			continue
		}
		partners = append(partners, partner)
	}
	return partners, nil
}

// Get Active Partners By OperatorId
func GetPartnerActiveProviders(operatorId string, partnerId string) ([]models.PartnerStatus, error) {
	//log.Println("GetPartnerActiveProviders: Looking for OperatorId - ", operatorId)
	partners := []models.PartnerStatus{}
	filter := bson.M{}
	filter["operator_id"] = operatorId
	filter["partner_id"] = partnerId
	filter["partner_status"] = "ACTIVE"
	filter["operator_status"] = "ACTIVE"

	cursor, err := PartnerStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetPartnerActiveProviders: Failed with error - ", err.Error())
		return partners, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		partner := models.PartnerStatus{}
		err = cursor.Decode(&partner)
		if err != nil {
			log.Println("GetPartnerActiveProviders: PartnerStatus Decode failed with error - ", err.Error())
			continue
		}
		partners = append(partners, partner)
	}
	return partners, nil
}

// Bulk insert Partner Staus objects
func InsertManyPartnerStatus(partnerStatus []models.PartnerStatus) error {
	//log.Println("InsertManyPartnerStatus: Adding PartnerStatus, count is - ", len(partnerStatus))
	proStatus := []interface{}{}
	for _, proSts := range partnerStatus {
		proSts.CreatedAt = time.Now().Unix()
		proSts.UpdatedAt = proSts.CreatedAt
		proStatus = append(proStatus, proSts)
	}
	result, err := PartnerStatusCollection.InsertMany(Ctx, proStatus)
	if err != nil {
		log.Println("InsertManyPartnerStatus: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertManyPartnerStatus: Inserted count is - ", len(result.InsertedIDs))
	return nil
}

// Update Partner Status
func UpdatePartnerProviderStatus(operatorId string, partnerId string, providerId string, status string) error {
	partnerKey := operatorId + "-" + partnerId + "-" + providerId
	nowTime := time.Now().Unix()
	filter := bson.D{{"partner_key", partnerKey}}
	update := bson.D{{"$set", bson.D{{"provider_status", status}, {"updated_at", nowTime}}}}
	opts := options.Update()
	result, err := PartnerStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdatePartnerProviderStatus: FAILED to UPDATE operator status - ", err.Error())
		return err
	}
	log.Println("UpdatePartnerProviderStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdatePartnerProviderStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update Partner Status
func UpdatePartnerOperatorStatus(operatorId string, partnerId string, providerId string, status string) error {
	partnerKey := operatorId + "-" + partnerId + "-" + providerId
	nowTime := time.Now().Unix()
	filter := bson.D{{"partner_key", partnerKey}}
	update := bson.D{{"$set", bson.D{{"operator_status", status}, {"updated_at", nowTime}}}}
	opts := options.Update()
	result, err := PartnerStatusCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdatePartnerOperatorStatus: FAILED to UPDATE operator status - ", err.Error())
		return err
	}
	log.Println("UpdatePartnerOperatorStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdatePartnerOperatorStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func GetUpdatedPartnerStatus(partnerKeys []string) ([]models.PartnerStatus, error) {
	//log.Println("GetUpdatedPartnerStatus: Looking for Keys Count - ", len(partnerKeys))
	// 0. Response object
	partnerStatuses := []models.PartnerStatus{}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	if len(partnerKeys) != 0 {
		filter["partner_key"] = bson.M{"$in": partnerKeys}
	} else {
		findOptions.SetLimit(100)
	}
	// 3. Execute Query
	cursor, err := PartnerStatusCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUpdatedPartnerStatus: Failed with error - ", err.Error())
		return partnerStatuses, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		partnerStatus := models.PartnerStatus{}
		err = cursor.Decode(&partnerStatus)
		if err != nil {
			log.Println("GetUpdatedPartnerStatus: Decode failed with error - ", err.Error())
			continue
		}
		partnerStatuses = append(partnerStatuses, partnerStatus)
	}
	return partnerStatuses, nil
}

// Get Active Partners By OperatorId
func GetActiveProvidersPS(operatorId string, partnerId string) ([]models.PartnerStatus, error) {
	//log.Println("GetPartnerActiveProviders: Looking for OperatorId - ", operatorId)
	partners := []models.PartnerStatus{}
	filter := bson.M{}
	filter["operator_id"] = operatorId
	filter["partner_id"] = partnerId
	filter["provider_status"] = "ACTIVE"
	filter["operator_status"] = "ACTIVE"

	cursor, err := PartnerStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetPartnerActiveProviders: Failed with error - ", err.Error())
		return partners, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		partner := models.PartnerStatus{}
		err = cursor.Decode(&partner)
		if err != nil {
			log.Println("GetPartnerActiveProviders: PartnerStatus Decode failed with error - ", err.Error())
			continue
		}
		partners = append(partners, partner)
	}
	return partners, nil
}

// Get Active Providers by OperatorId
// Get All PartnerStatuses
func GetAllPartnerStatus() ([]models.PartnerStatus, error) {
	//log.Println("GetAllPartnerStatus: ")
	partnerStatus := []models.PartnerStatus{}
	cursor, err := PartnerStatusCollection.Find(Ctx, bson.D{})
	if err != nil {
		log.Println("GetAllPartnerStatus: Failed with error - ", err.Error())
		return partnerStatus, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		ps := models.PartnerStatus{}
		err = cursor.Decode(&ps)
		if err != nil {
			log.Println("GetAllPartnerStatus: Decode failed with error - ", err.Error())
			continue
		}
		partnerStatus = append(partnerStatus, ps)
	}
	return partnerStatus, nil
}

// Update Partner Document
func ReplacePartner(partnerStatus models.PartnerStatus) error {
	partnerStatus.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", partnerStatus.ID}}
	result, err := PartnerStatusCollection.ReplaceOne(Ctx, filter, partnerStatus)
	if err != nil {
		log.Println("ReplacePartner: FAILED to UPDATE operator details - ", err.Error())
		return err
	}
	log.Println("ReplacePartner: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplacePartner: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Partners updated in last 5 minutes
func GetUpdatedPartners() ([]models.PartnerStatus, error) {
	retObjs := []models.PartnerStatus{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := PartnerStatusCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedPartners: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.PartnerStatus{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedPartners: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}

func DeletePartnersByOperator(operatorId string) error {
	filter := bson.M{}
	filter["operator_id"] = operatorId
	result, err := PartnerStatusCollection.DeleteMany(Ctx, filter)
	if err != nil {
		log.Println("DeletePartnersByOperatorId: Failed to delete partner status - ", err.Error())
		return err
	}
	log.Println("DeletePartnersByOperatorId: Deleted recoreds Count - ", result.DeletedCount)
	return nil
}
