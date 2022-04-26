package database

import (
	"Sp/dto/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert Provider Document
func InsertProvider(provider models.Provider) error {
	//log.Println("InsertProvider: Adding Documnet for providerId - ", provider.ProviderId)
	provider.CreatedAt = time.Now().Unix()
	provider.UpdatedAt = provider.CreatedAt
	_, err := ProviderCollection.InsertOne(Ctx, provider)
	if err != nil {
		log.Println("InsertProvider: Provider INSERT failed with error - ", err.Error())
		return err
	}
	//log.Println("InsertProvider: Document _id is - ", result.InsertedID)
	return nil
}

// Update Provider status
func UpdateProviderStatus(providerId string, status string) error {
	//if status == "ACTIVE" {
	//	log.Println("UpdateProviderStatus: Unblocking the provider - ", providerId)
	//} else {
	//	log.Println("UpdateProviderStatus: Status changing for the provider: ", providerId+"-"+status)
	//}
	filter := bson.D{{"provider_id", providerId}}
	updatedAt := time.Now().Unix()
	update := bson.D{{"$set", bson.D{{"status", status}, {"updated_at", updatedAt}}}}
	//update := bson.D{{"$set", bson.D{{"status", status}}}}
	opts := options.Update()
	_, err := ProviderCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateProviderStatus: FAILED to UPDATE operator status - ", err.Error())
		return err
	}
	//log.Println("UpdateProviderStatus: Matched recoreds Count - ", result.MatchedCount)
	//log.Println("UpdateProviderStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Get Provider Document
func GetProvider(providerId string) (models.Provider, error) {
	//log.Println("GetProvider: Looking for provider_id - ", providerId)
	provider := models.Provider{}
	err := ProviderCollection.FindOne(Ctx, bson.M{"provider_id": providerId}).Decode(&provider)
	if err != nil {
		log.Println("GetProvider: Provider NOT FOUND - ", err.Error())
		log.Println("GetProvider: provider_id - ", providerId)
		return provider, err
	}
	return provider, nil
}

// Get All Providers
func GetAllProviders() ([]models.Provider, error) {
	//log.Println("GetAllProviders: ")
	providers := []models.Provider{}
	cursor, err := ProviderCollection.Find(Ctx, bson.D{})
	if err != nil {
		log.Println("GetAllProviders: Failed with error - ", err.Error())
		return providers, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		provider := models.Provider{}
		err = cursor.Decode(&provider)
		if err != nil {
			log.Println("GetAllProviders: Decode failed with error - ", err.Error())
			continue
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

// Get NON-ACTIVE Providers
func GetBlockedProviders() ([]models.Provider, error) {
	//log.Println("GetBlockedProviders: ")
	providers := []models.Provider{}
	cursor, err := ProviderCollection.Find(Ctx, bson.D{})
	if err != nil {
		log.Println("GetBlockedProviders: Failed with error - ", err.Error())
		return providers, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		provider := models.Provider{}
		err = cursor.Decode(&provider)
		if err != nil {
			log.Println("GetBlockedProviders: Decode failed with error - ", err.Error())
			continue
		}
		if provider.Status == "BLOCKED" {
			providers = append(providers, provider)
		}
	}
	return providers, nil
}

func ReplaceProvider(provider models.Provider) error {
	//log.Println("ReplaceProvider: Adding Documnet for providerId - ", provider.ProviderId)
	provider.CreatedAt = time.Now().Unix()
	provider.UpdatedAt = provider.CreatedAt
	_, err := ProviderCollection.ReplaceOne(Ctx, bson.M{"_id": provider.ID}, provider)
	if err != nil {
		log.Println("ReplaceProvider: Provider INSERT failed with error - ", err.Error())
		return err
	}
	//log.Println("ReplaceProvider: Document _id is - ", result.InsertedID)
	return nil
}

// Get Providers updated in last 5 minutes
func GetUpdatedProviders() ([]models.Provider, error) {
	retObjs := []models.Provider{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := ProviderCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedProviders: Failed with error - ", err.Error())
		return retObjs, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		obj := models.Provider{}
		err = cursor.Decode(&obj)
		if err != nil {
			log.Println("GetUpdatedProviders: Decode failed with error - ", err.Error())
			continue
		}
		retObjs = append(retObjs, obj)
	}
	return retObjs, nil
}
