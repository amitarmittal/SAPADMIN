package coresvc

import (
	opdto "Sp/dto/operator"
	"Sp/dto/responsedto"
	sessdto "Sp/dto/session"
	sportsdto "Sp/dto/sports"
	"Sp/operator"
	"fmt"
	"log"
)

// Construct openBet object from Bet Object
func GetBetsStatusDto(bets []sportsdto.BetDto) []responsedto.BetStatus {
	betsStatus := []responsedto.BetStatus{}
	for _, bet := range bets {
		betStatus := responsedto.BetStatus{}
		betStatus.BetId = bet.BetId
		betStatus.Status = bet.Status
		betsStatus = append(betsStatus, betStatus)
	}
	return betsStatus
}

func OpBalanceCall(sessDto sessdto.B2BSessionDto, priKey string) (opdto.OperatorRespDto, error) {
	opBalRespDto := opdto.OperatorRespDto{}
	opBalRespDto, err := operator.WalletBalance(sessDto, priKey)
	if err != nil {
		log.Println("OpBalanceCall: Failed in getting user balance using Wallet Balance call! - ", err.Error())
		return opBalRespDto, err
	}
	if opBalRespDto.Status != "RS_OK" {
		log.Println("OpBalanceCall: Operator's Wallet Balance call returned error - !", opBalRespDto.Status)
		// TODO: Error code mapping
		return opBalRespDto, fmt.Errorf("Failed to get balance!")
	}
	log.Println("OpBalanceCall: User Balance is - ", opBalRespDto.Balance)
	return opBalRespDto, nil
}
