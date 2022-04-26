package database

import (
	dto "Sp/dto/session"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert Session Document
func InsertSessionDetails(sessDto dto.B2BSessionDto) error {
	//log.Println("InsertSessionDetails: Adding Documnet for token - ", sessDto.Token)
	_, err := B2bSessionCollection.InsertOne(Ctx, sessDto)
	if err != nil {
		log.Println("InsertSessionDetails: FAILED to INSERT session details - ", err.Error())
		return err
	}
	//log.Println("InsertSessionDetails: Document _id is - ", result.InsertedID)
	return nil
}

// Update Session Document
func UpdateSessionDetails(sessDto dto.B2BSessionDto) error {
	//log.Println("UpdateSessionDetails: Updating Documnet for token - ", sessDto.Token)
	opts := options.Update()
	filter := bson.D{{"token", sessDto.Token}}
	update := bson.D{{"$set", bson.D{{"expireAt", sessDto.ExpireAt}}}}
	_, err := B2bSessionCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateSessionDetails: FAILED to UPDATE session details - ", err.Error())
		return err
	}
	//log.Println("UpdateSessionDetails: Matched recoreds Count - ", result.MatchedCount)
	//log.Println("UpdateSessionDetails: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func GetSessionDetailsByToken(token string) (dto.B2BSessionDto, error) {
	//log.Println("GetSessionDetailsByToken: Looking for token - ", token)
	b2bSessionDto := dto.B2BSessionDto{}
	//var sessionDto bson.M
	err := B2bSessionCollection.FindOne(Ctx, bson.M{"token": token}).Decode(&b2bSessionDto)
	if err != nil {
		log.Println("GetSessionDetailsByToken: Session Details NOT FOUND for token - ", token)
		return b2bSessionDto, err
	}
	return b2bSessionDto, nil
}

func GetSessionDetailsByUserKey(userKey string) (dto.B2BSessionDto, error) {
	//log.Println("GetSessionDetailsByUserKey: Looking for userKey - ", userKey)
	b2bSessionDto := dto.B2BSessionDto{}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"createdAt", -1}})
	cur, err := B2bSessionCollection.Find(Ctx, bson.M{"userKey": userKey}, findOptions)
	if err != nil {
		log.Println("GetSessionDetailsByUserKey: Session Details NOT FOUND 1 for token - ", userKey)
		return b2bSessionDto, err
	}
	defer cur.Close(Ctx)

	for cur.Next(Ctx) {
		err = cur.Decode(&b2bSessionDto)
		if err != nil {
			log.Println("GetSessionDetailsByUserKey: FAILED to DECODE BSON - ")
			return b2bSessionDto, fmt.Errorf("FAILED to DECODE BSON")
		}
		//log.Println("GetSessionDetailsByUserKey: Session CreatedAt - ", b2bSessionDto.CreatedAt)
		return b2bSessionDto, nil
	}
	//log.Println("GetSessionDetailsByUserKey: Session Details NOT FOUND 2 for token - ", userKey)
	return b2bSessionDto, fmt.Errorf("NOT FOUND in DB")
}
