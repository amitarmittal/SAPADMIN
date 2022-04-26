package database

import (
	operatorDto "Sp/dto/operator"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert Portal Users
func InsertPortalUserDetails(portalUser operatorDto.PortalUser) error {
	//log.Println("InsertPortalUserDetails: Adding Portal User - ", portalUser)
	result, err := PortalUsersCollection.InsertOne(Ctx, portalUser)
	if err != nil {
		log.Println("InsertPortalUserDetails: FAILED to INSERT Portal User details - ", err.Error())
		return err
	}
	log.Println("InsertPortalUserDetails: Document _id is - ", result.InsertedID)
	return nil
}

// Get Portal Users
func GetPortalUserDetailsByUserId(userId string) (operatorDto.PortalUser, error) {
	portalUser := operatorDto.PortalUser{}
	err := PortalUsersCollection.FindOne(Ctx, bson.M{"user_id": userId}).Decode(&portalUser)
	if err != nil {
		log.Println("GetPortalUserDetailsByUserId: User NOT FOUND - ", err.Error())
		return portalUser, err
	}
	return portalUser, nil
}

// Get Portal User from UserName
func GetPortalUserDetailsByUserName(userName string) (operatorDto.PortalUser, error) {
	portalUser := operatorDto.PortalUser{}
	err := PortalUsersCollection.FindOne(Ctx, bson.M{"user_name": userName}).Decode(&portalUser)
	if err != nil {
		log.Println("GetPortalUserDetailsByUserName: User NOT FOUND - ", err.Error())
		return portalUser, err
	}
	return portalUser, nil
}

// Get Portal User from UserKey
func GetPortalUserDetailsByUserKey(userKey string) (operatorDto.PortalUser, error) {
	portalUser := operatorDto.PortalUser{}
	err := PortalUsersCollection.FindOne(Ctx, bson.M{"user_key": userKey}).Decode(&portalUser)
	if err != nil {
		log.Println("GetPortalUserDetailsByUserKey: User NOT FOUND - ", err.Error())
		return portalUser, err
	}
	return portalUser, nil
}

// Update Portal Users
func UpdatePortalUserDetails(portalUser operatorDto.PortalUser) error {
	//log.Println("UpdatePortalUserDetails: Updating Portal User - ", portalUser)
	_, err := PortalUsersCollection.UpdateOne(Ctx, bson.M{"user_id": portalUser.UserId}, bson.M{"$set": portalUser})
	if err != nil {
		log.Println("UpdatePortalUserDetails: FAILED to UPDATE Portal User details - ", err.Error())
		return err
	}
	return nil
}

// Get operators ledger history by operatorkey
func GetPortalUsers() ([]operatorDto.PortalUser, error) {
	// TODO: Sort by transaction time descending (Latest first)
	//log.Println("GetLedgers: Looking for operatorId - ", operatorId)
	result := []operatorDto.PortalUser{}
	opts := options.Find()
	filter := bson.D{{}}
	cursor, err := PortalUsersCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetOperatorLedgers: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := operatorDto.PortalUser{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetOperatorLedgers: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	log.Println("GetOperatorLedgers: total operators - ", len(result))
	return result, nil
}
