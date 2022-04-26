package utils

import (
	"Sp/constants"
	sportsdto "Sp/dto/sports"
	"fmt"
	"log"
	"math"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	SportsMapByName = map[string]string{"cricket": "4", "soccer": "1", "tennis": "2"}
	SportsMapById   = map[string]string{"4": "cricket", "1": "soccer", "2": "tennis"}
	ProviderMapById = map[string]string{"Dream": "Dream Sports", "BetFair": "Bet Fair", "SportRadar": "Sport Radar"}
)

func Truncate64(value float64) float64 {
	return float64(math.Round(value*100)) / 100
	//return float64(int(value*100)) / 100
}

func Truncate4Decfloat64(value float64) float64 {
	return float64(math.Round(value*10000)) / 10000
}

func Truncate1Decfloat64(value float64) float64 {
	return float64(math.Round(value*10)) / 10
}

func ObjectiveIdToString(objectId primitive.ObjectID) string {
	return objectId.Hex()
}

func StringToObjectiveId(objectId string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(objectId)
}

func GetRemark(bet sportsdto.BetDto) string {
	return bet.BetDetails.EventName + " - " + bet.BetDetails.BetType + " - " + bet.BetDetails.RunnerName + " - " + fmt.Sprintf("%.2f", bet.BetDetails.OddValue) + " - " + fmt.Sprintf("%.2f", bet.BetDetails.StakeAmount)
}

func GetOddsFactor(oddValue float64, marketType string) float64 {
	log.Println("GetOddsFactor: Data is - ", marketType, oddValue)
	if strings.ToUpper(marketType) == constants.SAP.MarketType.BOOKMAKER() {
		log.Println("GetOddsFactor: BOOKMAKER bet!!!")
		oddValue = 1.0 + oddValue*0.01
	}
	if strings.ToUpper(marketType) == constants.SAP.MarketType.FANCY() {
		log.Println("GetOddsFactor: FANCY bet!!!")
		oddValue = 1.0 + oddValue*0.01
	}
	log.Println("GetOddsFactor: oddValue is - ", oddValue)
	return oddValue
}
