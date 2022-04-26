package database

import (
	operatordto "Sp/dto/operator"
	dto "Sp/dto/portal"
	"Sp/dto/reports"
	"Sp/dto/responsedto"
	"Sp/dto/sports"
	utils "Sp/utilities"
	"errors"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get Bet Document
func GetBetDetails(betId string) (sports.BetDto, error) {
	log.Println("GetBetDetails: Looking for betId - ", betId)
	resultDto := sports.BetDto{}
	err := BetCollection.FindOne(Ctx, bson.M{"transaction_id": betId}).Decode(&resultDto)
	if err != nil {
		log.Println("GetBetDetails: Bet Details NOT FOUND for betId - ", betId)
		return resultDto, err
	}
	return resultDto, nil
}

// Get Bet Document
func GetBets(betIds []string) ([]sports.BetDto, error) {
	//log.Println("GetBets: Looking for betsCount - ", len(betIds))
	bets := []sports.BetDto{}
	filter := bson.M{"transaction_id": bson.M{"$in": betIds}}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBets: Find failed with error - ", err.Error())
		log.Println("GetBets: Looking for betsCount - ", len(betIds))
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetBets: totalBets - ", totalBets)
	//log.Println("GetBets: betsCount - ", len(bets))
	return bets, nil
}

// Get Bet Document
func GetAllBets(reqDto operatordto.BetsHistoryReqDto) ([]sports.BetDto, error) {
	//log.Println("GetAllBets")
	bets := []sports.BetDto{}
	filter := bson.M{}
	findOptions := options.Find()
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if reqDto.OperatorId != "" {
		filter["operator_id"] = reqDto.OperatorId
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	if reqDto.SportId != "" && reqDto.EventId != "" {
		eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
		filter["event_key"] = eventKey
	}
	if reqDto.FilterBy == "UpdatedAt" {
		findOptions.SetSort(bson.D{{"updated_at", -1}})
	}
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetAllBets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAllBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetAllBets: totalBets - ", totalBets)
	//log.Println("GetAllBets: betsCount - ", len(bets))
	return bets, nil
}

// Get Bet Document
func GetAllOperatorBets(reqDto operatordto.BetsHistoryReqDto) ([]sports.BetDto, error) {
	//log.Println("GetAllBets")
	bets := []sports.BetDto{}
	filter := bson.M{}
	findOptions := options.Find()
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if reqDto.OperatorId != "" {
		filter["operator_id"] = reqDto.OperatorId
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	if reqDto.SportId != "" && reqDto.EventId != "" {
		eventKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId
		filter["event_key"] = eventKey
	}
	if reqDto.FilterBy == "UpdatedAt" {
		findOptions.SetSort(bson.D{{"updated_at", -1}})
	}
	if reqDto.OperatorId == "" {
		filter["event_key"] = reqDto.OperatorId
	}
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetAllBets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAllBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetAllBets: totalBets - ", totalBets)
	//log.Println("GetAllBets: betsCount - ", len(bets))
	return bets, nil
}

// Get Settled Bet Document
func GetAllSettledBets() ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{"status": "SETTLED"}
	cursor, err := BetCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetAllSettledBets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalSettledBets = 0
	for cursor.Next(Ctx) {
		totalSettledBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAllSettledBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetAllSettledBets: totalSettledBets - ", totalSettledBets)
	//log.Println("GetAllSettledBets: SettledbetsCount - ", len(bets))
	return bets, nil
}

// Get Lapsed Bet Document
func GetAllLapsedBets() ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{"status": "LAPSED"}
	cursor, err := BetCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetAllLapsedBets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalLapsedBets = 0
	for cursor.Next(Ctx) {
		totalLapsedBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAllLapsedBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetAllLapsedBets: totalSettledBets - ", totalLapsedBets)
	//log.Println("GetAllLapsedBets: SettledbetsCount - ", len(bets))
	return bets, nil
}

// Get Cancelled Bet Document
func GetAllCancelledBets() ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{"status": "cancelled"}
	cursor, err := BetCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetAllCancelledBets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalLapsedBets = 0
	for cursor.Next(Ctx) {
		totalLapsedBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAllCancelledBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetAllCancelledBets: totalSettledBets - ", totalLapsedBets)
	//log.Println("GetAllCancelledBets: SettledbetsCount - ", len(bets))
	return bets, nil
}

// Get Open Bets By Market
func GetOpenBetsByMarket(eventKey string, marketId string) ([]sports.BetDto, int, error) {
	//log.Println("GetOpenBetsByMarket: Looking for EventKey - MarketId: ", eventKey+"-"+marketId)
	openbets := []sports.BetDto{}
	cursor, err := BetCollection.Find(Ctx, bson.M{"event_key": eventKey, "market_id": marketId})
	if err != nil {
		log.Println("GetOpenBetsByMarket: Bets Details NOT FOUND for  EventKey - MarketId: ", eventKey+"-"+marketId)
		return openbets, 0, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetOpenBetsByMarket: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status == "OPEN" || bet.Status == "UNMATCHED" || bet.Status == "" {
			openbets = append(openbets, bet)
		}
	}
	//log.Println("GetOpenBetsByMarket: totalBets - ", totalBets)
	//log.Println("GetOpenBetsByMarket: openBets - ", len(openbets))
	return openbets, totalBets, nil
}

// Get Open Bets By ProviderId & Bet Status
func GetBetsByStatus(providerId string, status string) ([]sports.BetDto, error) {
	//log.Println("GetBetsByStatus: Looking for providerId - status: ", providerId+"-"+status)
	bets := []sports.BetDto{}
	filter := bson.M{}
	if providerId != "" {
		filter["provider_id"] = providerId
	}
	filter["status"] = status
	cursor, err := BetCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetBetsByStatus: BetCollection.Find failed with error : ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsByStatus: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	return bets, nil
}

// Get Open Bets By Market
func GetBetsByMarket(eventKey string, marketId string, rollbackType string, start int64, end int64) ([]sports.BetDto, error) {
	//log.Println("GetBetsByMarket: Looking for EventKey - MarketId: ", eventKey+"-"+marketId)
	bets := []sports.BetDto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["event_key"] = eventKey
	filter["market_id"] = marketId
	if start != 0 {
		filter["bet_req.request_time"] = bson.M{"$gte": start}
	}
	if end != 0 {
		filter["bet_req.request_time"] = bson.M{"$lte": end}
	}
	// 2. Execute Query
	cursor, err := BetCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetBetsByMarket: Bets Details NOT FOUND for  EventKey - MarketId: ", eventKey+"-"+marketId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsByMarket: Bet Decode failed with error - ", err.Error())
			continue
		}
		if strings.ToLower(rollbackType) == "rollback" {
			if bet.Status == "SETTLED" || bet.Status == "SETTLED_VOIDED" {
				bets = append(bets, bet)
			}
		} else if strings.ToLower(rollbackType) == "void" {
			if bet.Status != "VOIDED" && bet.Status != "VOID" && bet.Status != "DELETED" && bet.Status != "CANCEL" {
				bets = append(bets, bet)
			}
		} else if strings.ToLower(rollbackType) == "voidrollback" {
			if bet.Status == "VOID" || bet.Status == "VOIDED" || bet.Status == "TIMELY_VOIDED" {
				bets = append(bets, bet)
			}
		}
	}
	//log.Println("GetBetsByMarket: totalBets - ", totalBets)
	//log.Println("GetBetsByMarket: betsCount - ", len(bets))
	return bets, nil
}

// Get Open Bets By Market & Time
func GetBetsByMarketTime(eventKey string, marketId string, start int64, end int64) ([]sports.BetDto, error) {
	//log.Println("GetBetsByMarketTime: Looking for EventKey - MarketId: ", eventKey+"-"+marketId)
	bets := []sports.BetDto{}
	// 1. Create Find filter
	// <operatorId>,ProviderId, <-SportId, <-EventId, UserId, StartDate<->EndDate
	filter := bson.M{}
	filter["event_key"] = eventKey
	filter["market_id"] = marketId
	filter["bet_req.request_time"] = bson.M{"$gte": start, "$lte": end}
	// 2. Execute Query
	cursor, err := BetCollection.Find(Ctx, filter)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetBetsByMarketTime: Bets NOT FOUND for EvenKey-MarketId : ", eventKey+"-"+marketId)
		return bets, err
	}
	// 4. Iterate through cursor
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsByMarketTime: Bet Decode failed with error - ", err.Error())
			continue
		}
		// add only non VOID / DELETED / CANCEL bets
		if bet.Status != "VOID" && bet.Status != "DELETED" && bet.Status != "CANCEL" {
			bets = append(bets, bet)
		}
	}
	//log.Println("GetBetsByMarket: totalBets - ", totalBets)
	//log.Println("GetBetsByMarket: betsCount - ", len(bets))
	return bets, nil
}

// Get Open Bets By User
func GetOpenBetsByUser(eventKey string, operatorId string, userId string) ([]sports.BetDto, int, error) {
	//log.Println("GetOpenBets: Looking for EventKey - OperatorId - UserId: ", eventKey+"-"+operatorId+"-"+userId)
	openbets := []sports.BetDto{}
	cursor, err := BetCollection.Find(Ctx, bson.M{"event_key": eventKey, "operator_id": operatorId, "user_id": userId})
	if err != nil {
		log.Println("GetOpenBets: Bets Details NOT FOUND for  EventKey - OperatorId - UserId: ", eventKey+"-"+operatorId+"-"+userId)
		return openbets, 0, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	var openBetsCount = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetOpenBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status == "OPEN" || bet.Status == "UNMATCHED" || bet.Status == "INPROCESS" || bet.Status == "" {
			openBetsCount++
			openbets = append(openbets, bet)
		}
	}
	//log.Println("GetOpenBets: totalBets - ", totalBets)
	//log.Println("GetOpenBets: openBets - ", openBetsCount)
	return openbets, totalBets, nil
}

// Get All Bets by OperatorId and ProviderId
func GetAllBetsByOperatorIdUserId(operatorId string, userId string) ([]sports.BetDto, error) {
	//log.Println("GetAllBetsByOperatorIdProviderId: Looking for OperatorId - ProviderId: ", operatorId+"-"+providerId)
	bets := []sports.BetDto{}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"bet_req.request_time", -1}})
	cursor, err := BetCollection.Find(Ctx, bson.M{"operator_id": operatorId, "user_id": userId}, findOptions)
	if err != nil {
		log.Println("GetAllBetsByOperatorIdProviderId: Bets Details NOT FOUND for  OperatorId - ProviderId: ", operatorId+"-"+userId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAllBetsByOperatorIdProviderId: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	//log.Println("GetAllBetsByOperatorIdProviderId: totalBets - ", totalBets)
	//log.Println("GetAllBetsByOperatorIdProviderId: betsCount - ", len(bets))
	return bets, nil
}

// Get Bets
func GetBetsByOperator(req operatordto.BetsHistoryReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	// 1. Create Find filter
	// <operatorId>,ProviderId, <-SportId, <-EventId, UserId, StartDate<->EndDate
	filter := bson.M{}
	if req.EventId != "" { // check for eventId
		// add event_key & operator_id
		eventKey := req.ProviderId + "-" + req.SportId + "-" + req.EventId
		//log.Println("GetOpenBets: with eventKey - ", eventKey)
		filter["event_key"] = eventKey
		filter["operator_id"] = req.OperatorId
	} else {
		// add operator_id
		filter["operator_id"] = req.OperatorId
		if req.SportId != "" { // no eventId, check for sportId
			//log.Println("GetOpenBets: with sportId - ", req.SportId)
			// add provider_id & bet_details.sport_name
			sportName := utils.SportsMapById[req.SportId]
			filter["provider_id"] = req.ProviderId
			filter["bet_details.sport_name"] = sportName
		} else if req.ProviderId != "" { // no sportId, check for providerId
			// add provider_id
			//log.Println("GetOpenBets: with ProviderId - ", req.ProviderId)
			filter["provider_id"] = req.ProviderId
		}
	}
	if req.UserId != "" { // check for userId
		// add user_id
		//log.Println("GetOpenBets: with UserId - ", req.UserId)
		filter["user_id"] = req.UserId
	}
	if len(req.BetIds) > 0 {
		filter["transaction_id"] = bson.M{"$in": req.BetIds}
	}
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		// add bet_req.request_time
		//log.Println("GetOpenBets: with date filters - ", req.StartDate)
		if strings.ToLower(req.FilterBy) == "updatedat" {
			filter["updated_at"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
		} else {
			filter["bet_req.request_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
		}
	}
	if req.Status != "" { // check for status
		// add status
		//log.Println("GetOpenBets: with Status - ", req.Status)
		filter["status"] = req.Status
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	if strings.ToLower(req.FilterBy) == "updatedat" {
		findOptions.SetSort(bson.D{{"updated_at", -1}})
	} else {
		findOptions.SetSort(bson.D{{"bet_req.request_time", -1}})
	}
	// 3. Execute Query
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetBetsByOperator: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return bets, err
	}
	// 4. Iterate through cursor
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsByOperator: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

// Insert Bet Document
func InsertBetDetails(betDto sports.BetDto) error {
	//log.Println("InsertBetDetails: Adding Documnet for betId - ", betDto.BetId)
	betDto.CreatedAt = time.Now().UnixMilli()
	betDto.UpdatedAt = betDto.CreatedAt
	_, err := BetCollection.InsertOne(Ctx, betDto)
	if err != nil {
		log.Println("InsertBetDetails: FAILED to INSERT Bet details - ", err.Error())
		return err
	}
	//log.Println("InsertBetDetails: Document _id is - ", result.InsertedID)
	return nil
}

// Update Bet Document
func UpdateBet(betDto sports.BetDto) error {
	//log.Println("UpdateBetDetails: Updating Documnet for betId - ", betDto.BetId)
	betDto.UpdatedAt = time.Now().UnixMilli()
	//opts1 := options.Update()
	opts := options.Replace()
	// var isUpsert bool = true
	// opts.Upsert = &isUpsert
	filter := bson.D{{"transaction_id", betDto.BetId}}
	//update := bson.D{{"$set", bson.D{{"status", betDto.Status}}}}
	result, err := BetCollection.ReplaceOne(Ctx, filter, betDto, opts)
	//result, err := BetCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateBetDetails: FAILED to UPDATE result details - ", err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		log.Println("UpdateBetDetails: ZERO Matched recoreds Count for betid - ", betDto.BetId)
	}
	if result.ModifiedCount == 0 {
		log.Println("UpdateBetDetails: ZERO Modified recoreds Count for betid - ", betDto.BetId)
	}
	return nil
}

// Update bets.
func UpdateBets(bets []sports.BetDto) (int, []string) {
	//log.Println("UpdateBets: Updating user bets for count - ", len(bets))
	msgs := []string{}
	count1 := 0
	count2 := 0
	// TODO: Find a way to update documents in one DB call
	for _, bet := range bets {
		count1++
		err := UpdateBet(bet)
		if err != nil {
			log.Println("UpdateBets: FAILED to UPDATE - ", err.Error())
			// TODO: Handle failures
			msgs = append(msgs, bet.BetId+": "+err.Error())
			continue
		}
		count2++
	}
	//log.Println("UpdateBets: Matched recoreds Count - ", count1)
	//log.Println("UpdateBets: Modified recoreds Count - ", count2)
	return count2, msgs
}

// Update bets.
func UpdateBetsStatus(betIds []string, status string) (int, error) {
	count := 0
	// betIds := []string{}
	// for _, bet := range bets {
	// 	betIds = append(betIds, bet.BetId)
	// }
	filter := bson.M{"transaction_id": bson.M{"$in": betIds}}
	updatedAt := time.Now().UnixMilli()
	update := bson.D{{"$set", bson.D{{"status", status}, {"updated_at", updatedAt}}}}
	//update := bson.D{{"$set", bson.D{{"status", status}}}}
	opts := options.Update()
	result, err := BetCollection.UpdateMany(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateBetsStatus: UpdateMany failed with error - ", err.Error())
		return count, err
	}
	count = int(result.ModifiedCount)
	return count, nil
}

func GetAllbets() ([]sports.BetDto, error) {
	//log.Println("Getbets: Getting user bets")
	start := 0
	last := 10000
	pageSize := 40

	options := options.Find()
	options.SetLimit(int64(pageSize)).SetSkip(int64(start))
	count := 0
	for i := start; i <= last; i = i + 40 {
		bets := []sports.BetDto{}
		options.SetSkip(int64(i))
		cursor, err := BetCollection.Find(Ctx, bson.D{}, options)
		if err != nil {
			log.Println("Getbets: FAILED to get bets - ", err.Error())
			return bets, err
		}
		for cursor.Next(Ctx) {
			bet := sports.BetDto{}
			err := cursor.Decode(&bet)
			if err != nil {
				count += 1
				log.Println("Getbets: FAILED to decode bet - ", err.Error())
				continue
			}
			bets = append(bets, bet)
		}
		// log.Println("Getbets: Fetched ", len(bets), " bets")
		if len(bets) < 40 {
			break
		}
		cursor.Close(Ctx)
	}
	log.Println("Getbets: Total failed Bets - ", count)
	return []sports.BetDto{}, nil
}

// Get Bets
func GetBetReport(startDate int64, endDate int64, page int64, pageSize int64, operatorId string, providerId string, sportId string, competitionId string, eventId string, marketId string, status string, userId string, filterBy string) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	// 1. Create Find filter
	// <operatorId>,ProviderId, <-SportId, <-EventId, UserId, StartDate<->EndDate
	filter := bson.M{}

	if eventId != "" { // check for eventId
		// add event_key & operator_id
		eventKey := providerId + "-" + sportId + "-" + eventId
		//log.Println("GetOpenBets: with eventKey - ", eventKey)
		filter["event_key"] = eventKey
		filter["operator_id"] = operatorId
	} else {
		// add operator_id
		filter["operator_id"] = operatorId
		if sportId != "" { // no eventId, check for sportId
			//log.Println("GetOpenBets: with sportId - ", SportId)
			// add provider_id & bet_details.sport_name
			sportName := utils.SportsMapById[sportId]
			filter["provider_id"] = providerId
			filter["bet_details.sport_name"] = sportName
		} else if providerId != "" { // no sportId, check for providerId
			// add provider_id
			//log.Println("GetOpenBets: with ProviderId - ", ProviderId)
			filter["provider_id"] = providerId
		}
	}
	if userId != "" { // check for userId
		// add user_id
		//log.Println("GetOpenBets: with UserId - ", UserId)
		filter["user_id"] = userId
	}
	if startDate > 0 && endDate > 0 { // check for date filters
		// add bet_request_time
		//log.Println("GetOpenBets: with date filters - ", StartDate)
		if strings.ToLower(filterBy) == "updatedat" {
			filter["updated_at"] = bson.M{"$gte": startDate, "$lte": endDate}
		} else {
			filter["bet_request_time"] = bson.M{"$gte": startDate, "$lte": endDate}
		}
	}
	if status != "" { // check for status
		// add status
		//log.Println("GetOpenBets: with Status - ", Status)
		filter["status"] = status
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	if page != 1 {
		findOptions.SetSkip((page - 1) * pageSize)
	}
	findOptions.SetLimit(pageSize)
	if strings.ToLower(filterBy) == "updatedat" {
		findOptions.SetSort(bson.D{{"updated_at", -1}})
	} else {
		findOptions.SetSort(bson.D{{"bet_request_time", -1}})
	}
	// 3. Execute Query
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetBetsByOperator: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	// 4. Iterate through cursor
	var totalBets = 0
	for cursor.Next(Ctx) {
		totalBets++
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsByOperator: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetList(operatorId string, reqDto reports.BetListReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["created_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.SportName != "" {
		if reqDto.SportName != "all" {
			if reqDto.ProviderId == "Dream" {
				filter["bet_details.sport_name"] = strings.ToLower(reqDto.SportName)
			} else {
				filter["bet_details.sport_name"] = reqDto.SportName
			}
		}
	}
	if reqDto.SportId != "" {
		filter["sport_id"] = reqDto.SportId
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	// filter from Operator Admin
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	// filter from SAP Admin
	if reqDto.OperatorId != "" {
		filter["operator_id"] = reqDto.OperatorId
	}

	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetList: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetList: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetUserStatement(operatorId string, reqDto reports.UserStatementReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["created_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000} //??
	}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUserStatement: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetUserStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetAdminStatement(operatorId string, reqDto reports.AdminStatementReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetAdminStatement: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetAdminStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetMyAccStatement(operatorId string, reqDto reports.MyAccStatementReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetMyAccStatement: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetMyAccStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForGameReport(operatorId string, reqDto reports.GameReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetMyAccStatement: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetMyAccStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForSportReport(operatorId string, reqDto reports.SportReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	if reqDto.EventId != "" {
		filter["event_id"] = reqDto.EventId
	}
	if reqDto.CompetitionId != "" {
		filter["competition_id"] = reqDto.CompetitionId
	}
	if reqDto.SportId != "" {
		filter["sport_id"] = reqDto.SportId
	}
	if reqDto.Status != "" {
		filter["status"] = reqDto.Status
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.MarketId != "" {
		filter["market_id"] = reqDto.MarketId
	}
	filter["status"] = "SETTLED"
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"bet_req.request_time", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetMyAccStatement: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetMyAccStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForPnLReport(operatorId string, reqDto reports.PnLReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsForPnLReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsForPnLReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForProviderPnLReport(operatorId string, reqDto reports.ProviderPnLReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsForProviderPnLReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsForProviderPnLReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForSportPnLReport(operatorId string, reqDto reports.SportPnLReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.SportId != "" {
		filter["sport_id"] = strings.ToLower(reqDto.SportId)
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsForSportPnLReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsForSportPnLReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForCompetitionPnLReport(operatorId string, reqDto reports.CompetitionPnLReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.SportId != "" {
		filter["sport_id"] = strings.ToLower(reqDto.SportId)
	}
	if reqDto.CompetitionId != "" {
		filter["competition_id"] = reqDto.CompetitionId
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsForCompetitionPnLReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsForCompetitionPnLReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForEventPnLReport(operatorId string, reqDto reports.EventPnLReportReqDto) ([]sports.BetDto, error) {
	bets := []sports.BetDto{}
	filter := bson.M{}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["updated_at"] = bson.M{"$gte": reqDto.StartTime * 1000, "$lte": reqDto.EndTime * 1000}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.SportId != "" {
		filter["sport_id"] = strings.ToLower(reqDto.SportId)
	}
	if reqDto.CompetitionId != "" {
		filter["competition_id"] = reqDto.CompetitionId
	}
	if reqDto.EventId != "" {
		filter["event_id"] = reqDto.EventId
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	if reqDto.UserId != "" {
		filter["user_id"] = reqDto.UserId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsForEventPnLReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsForEventPnLReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

// Pagnination Testing
// Get Bets
func GetBetsReportA(page int64, pageSize int64) ([]responsedto.BetData, int64, error) {
	var reportDataA []responsedto.BetData
	var count int64 = 0

	//log.Println("GetBetsReport: page is : ", page)
	//log.Println("GetBetsReport: pageSize is : ", pageSize)
	var skipCount int64 = 0
	if page > 0 {
		skipCount = (page - 1) * pageSize
	}
	//log.Println("GetBetsReport: skipCount is : ", skipCount)

	matchDoc := bson.D{{
		"$match", bson.D{{}}}}
	skipDoc := bson.D{{
		"$skip", skipCount}}
	limitDoc := bson.D{{
		"$limit", pageSize}}
	countDoc := bson.D{{
		"$count", "count"}}

	dataDoc := bson.A{matchDoc, skipDoc, limitDoc}
	totalDoc := bson.A{matchDoc, countDoc}
	facetDoc := bson.D{{"data", dataDoc}, {"total", totalDoc}}
	facetStage := bson.D{{"$facet", facetDoc}}

	// 3. Execute Query
	reportCur, err := BetCollection.Aggregate(Ctx, mongo.Pipeline{facetStage})
	defer reportCur.Close(Ctx)
	if err != nil {
		log.Println("GetBetsReport: BetCollection.Aggregate failed with error : ", err.Error())
		return reportDataA, count, err
	}
	// 4. Iterate through cursor
	err = reportCur.All(Ctx, &reportDataA)
	if err != nil {
		log.Println("GetBetsReport: reportCur.All failed with error : ", err.Error())
		return reportDataA, count, err
	}
	// log.Println("GetBetsReport: reportDataA count is : ", len(reportDataA))
	// log.Println("GetBetsReport: reportDataA[0].Data count is : ", len(reportDataA[0].Data))
	// log.Println("GetBetsReport: reportDataA[0].Total count is : ", len(reportDataA[0].Total))
	// log.Println("GetBetsReport: reportDataA[0].Total[0].Count is : ", reportDataA[0].Total[0].Count)
	return reportDataA, reportDataA[0].Total[0].Count, nil
}

// Pagnination Testing
// Get Bets
func GetBetsReport(page int64, pageSize int64, operatorId string, status string) ([]responsedto.BetData, int64, error) {
	var reportDataA []responsedto.BetData
	var count int64 = 0

	log.Println("GetBetsReport: operatorId is : ", operatorId)
	log.Println("GetBetsReport: status is : ", status)
	log.Println("GetBetsReport: page is : ", page)
	log.Println("GetBetsReport: pageSize is : ", pageSize)
	var skipCount int64 = 0
	if page > 0 {
		skipCount = (page - 1) * pageSize
	}
	log.Println("GetBetsReport: skipCount is : ", skipCount)

	matchDoc := bson.D{{
		"$match", bson.D{{
			"$and", bson.A{
				bson.D{{
					"operator_id", operatorId}},
				bson.D{{
					"status", status}},
			}}}}}
	skipDoc := bson.D{{
		"$skip", skipCount}}
	limitDoc := bson.D{{
		"$limit", pageSize}}
	countDoc := bson.D{{
		"$count", "count"}}

	dataDoc := bson.A{matchDoc, skipDoc, limitDoc}
	totalDoc := bson.A{matchDoc, countDoc}
	facetDoc := bson.D{{"data", dataDoc}, {"total", totalDoc}}
	facetStage := bson.D{{"$facet", facetDoc}}

	// 3. Execute Query
	reportCur, err := BetCollection.Aggregate(Ctx, mongo.Pipeline{facetStage})
	defer reportCur.Close(Ctx)
	if err != nil {
		log.Println("GetBetsReport: BetCollection.Aggregate failed with error : ", err.Error())
		return reportDataA, count, err
	}
	// 4. Iterate through cursor
	err = reportCur.All(Ctx, &reportDataA)
	if err != nil {
		log.Println("GetBetsReport: reportCur.All failed with error : ", err.Error())
		return reportDataA, count, err
	}
	log.Println("GetBetsReport: reportDataA count is : ", len(reportDataA))
	log.Println("GetBetsReport: reportDataA[0].Data count is : ", len(reportDataA[0].Data))
	log.Println("GetBetsReport: reportDataA[0].Total count is : ", len(reportDataA[0].Total))
	log.Println("GetBetsReport: reportDataA[0].Total[0].Count is : ", reportDataA[0].Total[0].Count)
	return reportDataA, reportDataA[0].Total[0].Count, nil
}

func GetRiskReport(operatorId string, reqDto reports.RiskReportReqDto) ([]sports.BetDto, error) {
	var bets []sports.BetDto

	filter := bson.M{}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetRiskReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetRiskReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForMarketRiskReport(operatorId, role string, reqDto reports.RiskReportReqDto) ([]sports.BetDto, error) {
	var bets []sports.BetDto

	filter := bson.M{}
	if operatorId != "" && role != "SAPAdmin" {
		filter["operator_id"] = operatorId
	} else {
		if role != "SAPAdmin" {
			return bets, errors.New("operatorId is required")
		}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	} else {
		return bets, errors.New("providerId is required")
	}
	if reqDto.EventId != "" {
		filter["event_id"] = reqDto.EventId
	} else {
		return bets, errors.New("eventId is required")
	}
	filter["status"] = "OPEN"
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsforMarketRiskReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsforMarketRiskReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

func GetBetsForUserBookReport(operatorId, role string, reqDto reports.UserBookReportReqDto) ([]sports.BetDto, error) {
	var bets []sports.BetDto

	filter := bson.M{}
	if operatorId != "" && role != "SAPAdmin" {
		filter["operator_id"] = operatorId
	} else {
		if role != "SAPAdmin" {
			return bets, errors.New("operatorId is required")
		}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	} else {
		return bets, errors.New("providerId is required")
	}
	if reqDto.EventId != "" {
		filter["event_id"] = reqDto.EventId
	} else {
		return bets, errors.New("eventId is required")
	}
	if reqDto.MarketId != "" {
		filter["market_id"] = reqDto.MarketId
	} else {
		return bets, errors.New("marketId is required")
	}
	filter["status"] = "OPEN"
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsforUserBookReport: Bets NOT FOUND for OperatorId : ", operatorId)
		return bets, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err := cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsforUserBookReport: Bet Decode failed with error - ", err.Error())
			continue
		}
		if bet.Status != "DELETED" {
			bets = append(bets, bet)
		}
	}
	return bets, nil
}

// Get Bet Document
func GetBetsByMarkets(eventKeys []string, marketIds []string, status string) ([]sports.BetDto, error) {
	//log.Println("GetBets: Looking for betsCount - ", len(betIds))
	bets := []sports.BetDto{}
	filter := bson.M{}
	filter["event_key"] = bson.M{"$in": eventKeys}
	filter["market_id"] = bson.M{"$in": marketIds}
	if status != "" {
		filter["status"] = status
	}
	findOptions := options.Find()
	// findOptions.SetSort(bson.D{{"created_at", 1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetBetsByMarkets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetBetsByMarkets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	log.Println("GetBetsByMarkets: betsCount - ", len(bets))
	return bets, nil
}

func GetOpenBets(operatorId, role string, reqDto dto.OpenBetsReqDto) ([]sports.BetDto, error) {
	var bets []sports.BetDto

	filter := bson.M{}
	if operatorId != "" && role != "SAPAdmin" {
		filter["operator_id"] = operatorId
	} else {
		if role != "SAPAdmin" {
			return bets, errors.New("operatorId is required")
		}
	}
	if reqDto.ProviderId != "" {
		filter["provider_id"] = reqDto.ProviderId
	}
	if reqDto.SportId != "" {
		filter["sport_id"] = reqDto.SportId
	}
	if reqDto.EventId != "" {
		filter["event_id"] = reqDto.EventId
	}
	filter["status"] = "OPEN"
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}})
	cursor, err := BetCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetOpenBets: Find failed with error - ", err.Error())
		return bets, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		bet := sports.BetDto{}
		err = cursor.Decode(&bet)
		if err != nil {
			log.Println("GetOpenBets: Bet Decode failed with error - ", err.Error())
			continue
		}
		bets = append(bets, bet)
	}
	log.Println("GetOpenBets: betsCount - ", len(bets))
	return bets, nil
}
