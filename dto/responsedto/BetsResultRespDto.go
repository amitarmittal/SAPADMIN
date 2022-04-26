package responsedto

import (
	"Sp/constants"
	"Sp/dto/sports"
)

type BetResult struct {
	OperatorId     string  `json:"operatorId"`
	Token          string  `json:"token"`
	UserId         string  `json:"userId"`
	TransactionId  string  `json:"transactionId"`
	EventId        string  `json:"eventId"`  // EventId map to gameId in JSON to be inline with GAP
	MarketId       string  `json:"marketId"` // MarketId map to roundId in JSON to be inline iwth GAP
	Status         string  `json:"status"`
	ReqId          string  `json:"reqId"`
	CreditAmount   float64 `json:"creditAmount"`   // SETTLED Bets
	RollbackAmount float64 `json:"rollbackAmount"` // +ve value to deposit to user, -ve value to deduct from user
	//RollbackType   string  `json:"rollbackType"`   // rollback / void / cancelled / expired / lapsed / deleted
	RollbackReason string `json:"rollbackReason"` // wrong settlement / market closed / invalid odds / time exceeded
}

type BetsResultRespDto struct {
	Status           string      `json:"status"`
	ErrorDescription string      `json:"errorDescription"`
	BetsResult       []BetResult `json:"betsResult"`
}

func GetBetResult(betDto sports.BetDto) BetResult {
	betResult := BetResult{}
	betResult.OperatorId = betDto.OperatorId
	betResult.Token = betDto.Token
	betResult.UserId = betDto.UserId
	betResult.TransactionId = betDto.BetId
	betResult.EventId = betDto.EventId
	betResult.MarketId = betDto.MarketId
	betResult.Status = betDto.Status
	if betResult.Status == constants.SAP.BetStatus.SETTLED() || betResult.Status == "SETTLED-failed" {
		if len(betDto.ResultReqs) > 0 {
			betResult.ReqId = betDto.ResultReqs[len(betDto.ResultReqs)-1].ReqId
			betResult.CreditAmount = betDto.ResultReqs[len(betDto.ResultReqs)-1].CreditAmount
			betResult.Status = constants.SAP.BetStatus.SETTLED()
		} else {
			// TODO: Log
		}
	}
	if betResult.Status == constants.SAP.BetStatus.ROLLBACK() || betResult.Status == "Rollback-failed" {
		if len(betDto.RollbackReqs) > 0 {
			betResult.ReqId = betDto.RollbackReqs[len(betDto.RollbackReqs)-1].ReqId
			betResult.RollbackAmount = betDto.RollbackReqs[len(betDto.RollbackReqs)-1].RollbackAmount
			//betResult.RollbackType = betDto.RollbackReqs[len(betDto.RollbackReqs) - 1].
			betResult.RollbackReason = betDto.RollbackReqs[len(betDto.RollbackReqs)-1].RollbackReason
			betResult.Status = constants.SAP.BetStatus.ROLLBACK()
		} else {
			// TODO: Log
		}
	}
	return betResult
}
