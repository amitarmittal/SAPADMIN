package sportradar

import (
	"Sp/dto/requestdto"
	"Sp/dto/sports"
	"time"
)

/*
{
  "betId": "string",
  "betPlacedTime": 0,
  "currency": "string",
  "eventId": "string",
  "marketId": "string",
  "odds": 0,
  "runnerId": "string",
  "sportId": "string",
  "stakeAmount": 0,
  "ticketId": "string",
  "userId": "string",
  "userIp": "string"
}*/
type PlaceOrderReq struct {
	BetId         string  `json:"betId"`
	BetPlacedTime int64   `json:"betPlacedTime"`
	Currency      string  `json:"currency"`
	EventId       string  `json:"eventId"`
	MarketId      string  `json:"marketId"`
	Odds          float64 `json:"odds"`
	RunnerId      string  `json:"runnerId"`
	SportId       string  `json:"sportId"`
	StakeAmount   float64 `json:"stakeAmount"`
	TicketId      string  `json:"ticketId"`
	UserId        string  `json:"userId"`
	UserIp        string  `json:"userIp"`
	SenderChannel string  `json:"senderChannel"`
}

func NewPlaceOrderReq(reqDto requestdto.PlaceBetReqDto, betDto sports.BetDto) PlaceOrderReq {
	placeOrderReq := PlaceOrderReq{}
	placeOrderReq.BetId = betDto.BetId
	placeOrderReq.BetPlacedTime = time.Now().UnixNano() / 1000000
	placeOrderReq.Currency = "HKD"
	placeOrderReq.EventId = reqDto.EventId
	placeOrderReq.MarketId = reqDto.MarketId
	placeOrderReq.Odds = reqDto.OddValue
	placeOrderReq.RunnerId = reqDto.RunnerId
	placeOrderReq.SportId = reqDto.SportId
	placeOrderReq.StakeAmount = reqDto.StakeAmount
	placeOrderReq.TicketId = betDto.BetReq.BetId
	placeOrderReq.UserId = betDto.UserId
	placeOrderReq.UserIp = betDto.UserIp
	placeOrderReq.SenderChannel = reqDto.UserAgent
	return placeOrderReq
}
