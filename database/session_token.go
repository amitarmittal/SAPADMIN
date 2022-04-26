package database

import (
	operatorDto "Sp/dto/operator"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

// Insert Session Token
func InsertSessionTokenDetails(portalSession operatorDto.PortalSession) error {
	//log.Println("InsertSessionTokenDetails: Adding Document for Session - ", portalSession)
	result, err := SessionTokenCollection.InsertOne(Ctx, portalSession)
	if err != nil {
		log.Println("InsertOperatorDetails: FAILED to INSERT session details - ", err.Error())
		return err
	}
	log.Println("InsertOperatorDetails: Document _id is - ", result.InsertedID)
	return nil
}

// Get Session Token
func GetSessionTokenDetails(jwt_token string, userId string) (operatorDto.PortalSession, error) {
	//log.Println("GetSessionTokenDetails: Looking for user_id - ", userId)
	portalSession := operatorDto.PortalSession{}
	err := SessionTokenCollection.FindOne(Ctx, bson.M{"jwt_token": jwt_token}).Decode(&portalSession)
	log.Println(portalSession.JWTToken)
	if err != nil {
		log.Println("GetSessionDetailsByToken: Portal Session NOT FOUND - ", err.Error())
		return portalSession, err
	}
	return portalSession, nil
}
