package database

import (
	"Sp/dto/sports"
	"log"
)

// Insert Bet Document
func InsertFailedBet(betDto sports.BetDto) error {
	//log.Println("InsertFailedBet: Adding Documnet for betId - ", betDto.BetId)
	result, err := FailedBetCollection.InsertOne(Ctx, betDto)
	if err != nil {
		log.Println("InsertFailedBet: FAILED to INSERT Bet details - ", err.Error())
		return err
	}
	log.Println("InsertFailedBet: Document _id is - ", result.InsertedID)
	return nil
}
