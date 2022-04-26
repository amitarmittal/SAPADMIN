package request

import (
	"Sp/constants"
	"Sp/dto/requestdto"
	utils "Sp/utilities"
	"log"
	"strconv"
	"strings"
)

type MarketVersion struct {
	Version float64 `json:"version"`
}

type LimitOrder struct {
	Size            float64 `json:"size"`
	Price           float64 `json:"price"`
	PersistencaType string  `json:"persistenceType"`
	TimeInForce     string  `json:"timeInForce"`
}

type LimitOnCloseOrder struct {
	Liability float64 `json:"liability"`
	Price     float64 `json:"price"`
}

type MarketOnCloseOrder struct {
	Liability float64 `json:"liability"`
}

type PlaceInstruction struct {
	OrderType   string     `json:"orderType"`
	SelectionId float64    `json:"selectionId"`
	Handicap    float64    `json:"handicap"`
	Side        string     `json:"side"`
	LimitOrder  LimitOrder `json:"limitOrder"`
	//LimitOnCloseOrder  LimitOnCloseOrder  `json:"limitOnCloseOrder"`
	//MarketOnCloseOrder MarketOnCloseOrder `json:"marketOnCloseOrder"`
	CustomerOrderRef string `json:"customerOrderRef"`
}

type PlaceOrderReq struct {
	MarketId     string             `json:"marketId"`
	Instructions []PlaceInstruction `json:"instructions"`
	CustomerRef  string             `json:"customerRef"`
	//MarketVer           MarketVersion      `json:"marketVersion"`
	//CustomerStrategyRef string `json:"customerStrategyRef"`
	//IsAsync             bool   `json:"async"`
}

func NewPlaceOrderReq(reqDto requestdto.PlaceBetReqDto, customerOrderRef string) PlaceOrderReq {
	placeOrderReq := PlaceOrderReq{}
	placeOrderReq.MarketId = reqDto.MarketId
	//placeOrderReq.CustomerRef = ""         // ???
	//placeOrderReq.CustomerStrategyRef = "" // ???
	//placeOrderReq.IsAsync = false
	// Place Instructions
	placeOrderReq.Instructions = []PlaceInstruction{}
	placeInstruction := PlaceInstruction{}
	placeInstruction.CustomerOrderRef = customerOrderRef
	var err error = nil
	placeInstruction.SelectionId, err = strconv.ParseFloat(reqDto.RunnerId, 64)
	if err != nil {
		log.Println("NewPlaceOrderReq: ParseFloat failed for RunnerId with error - ", err.Error())
	}
	placeInstruction.Handicap = 0
	placeInstruction.Side = strings.ToUpper(reqDto.BetType)
	//placeInstruction.CustomerOrderRef = ""
	// Limit Order
	placeInstruction.OrderType = "LIMIT"
	placeInstruction.LimitOrder = LimitOrder{}
	placeInstruction.LimitOrder.Size = utils.Truncate64(reqDto.StakeAmount)
	placeInstruction.LimitOrder.Price = reqDto.OddValue
	if reqDto.MarketType == constants.SAP.MarketType.LINE_ODDS() {
		placeInstruction.LimitOrder.Price = reqDto.SessionOutcome
	}
	// if reqDto.IsUnmatched {
	placeInstruction.LimitOrder.PersistencaType = "PERSIST"
	// } else {
	// 	placeInstruction.LimitOrder.PersistencaType = "LAPSE"
	// 	placeInstruction.LimitOrder.TimeInForce = "FILL_OR_KILL"
	// }

	placeOrderReq.Instructions = append(placeOrderReq.Instructions, placeInstruction)

	// Market Version (Optional)
	//placeOrderReq.MarketVer = MarketVersion{}
	//placeOrderReq.MarketVer.Version = 0

	return placeOrderReq
}

func NewPlaceOrderReqAsync(reqDto requestdto.PlaceBetReqDto, customerOrderRef string, customerRef string) PlaceOrderReq {
	placeOrderReq := PlaceOrderReq{}
	placeOrderReq.MarketId = reqDto.MarketId
	//placeOrderReq.CustomerRef = customerRef // ???
	//placeOrderReq.CustomerStrategyRef = "" // ???
	//placeOrderReq.IsAsync = false
	// Place Instructions
	placeOrderReq.Instructions = []PlaceInstruction{}
	placeInstruction := PlaceInstruction{}
	placeInstruction.CustomerOrderRef = customerOrderRef
	var err error = nil
	placeInstruction.SelectionId, err = strconv.ParseFloat(reqDto.RunnerId, 64)
	if err != nil {
		log.Println("NewPlaceOrderReq: ParseFloat failed for RunnerId with error - ", err.Error())
	}
	placeInstruction.Handicap = 0
	placeInstruction.Side = strings.ToUpper(reqDto.BetType)
	//placeInstruction.CustomerOrderRef = ""
	// Limit Order
	placeInstruction.OrderType = "LIMIT"
	placeInstruction.LimitOrder = LimitOrder{}
	placeInstruction.LimitOrder.Size = utils.Truncate64(reqDto.StakeAmount)
	placeInstruction.LimitOrder.Price = reqDto.OddValue
	if reqDto.MarketType == constants.SAP.MarketType.LINE_ODDS() {
		placeInstruction.LimitOrder.Price = reqDto.SessionOutcome
	}
	//if reqDto.IsUnmatched {
	placeInstruction.LimitOrder.PersistencaType = "PERSIST"
	// } else {
	// 	placeInstruction.LimitOrder.PersistencaType = "LAPSE"
	// 	placeInstruction.LimitOrder.TimeInForce = "FILL_OR_KILL"
	// }

	placeOrderReq.Instructions = append(placeOrderReq.Instructions, placeInstruction)

	// Market Version (Optional)
	//placeOrderReq.MarketVer = MarketVersion{}
	//placeOrderReq.MarketVer.Version = 0

	return placeOrderReq
}
