package database

import (
	"Sp/dto/models"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
// Get User Statement
func GetOperatorStatement(req portaldto.UserStatementReqDto) ([]models.OperatorLedgerDto, error) {
	records := []models.OperatorLedgerDto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["operator_id"] = req.OperatorId
	filter["user_id"] = req.UserId
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		log.Println("GetOperatorStatement: with date filters - ", req.StartDate)
		filter["transaction_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"transaction_time", -1}})
	// 3. Execute Query
	cursor, err := OperatorLedgerCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetOperatorStatement: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return records, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		record := models.OperatorLedgerDto{}
		err := cursor.Decode(&record)
		if err != nil {
			log.Println("GetOperatorStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		switch strings.ToUpper(req.TxType) {
		case "FUNDS":
			if strings.ToUpper(record.TransactionType) == "DEPOSIT-FUNDS" || strings.ToUpper(record.TransactionType) == "WITHDRAW-FUNDS" {
				records = append(records, record)
			}
		case "BETS":
			if strings.ToUpper(record.TransactionType) != "DEPOSIT-FUNDS" && strings.ToUpper(record.TransactionType) != "WITHDRAW-FUNDS" {
				records = append(records, record)
			}
		default:
			records = append(records, record)
		}
	}
	return records, nil
}
*/
/*
// Get All User Statement
func GetAllOperatorStatement(req portaldto.UserStatementReqDto) ([]models.OperatorLedgerDto, error) {
	records := []models.OperatorLedgerDto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["user_id"] = req.UserId
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		log.Println("GetAllOperatorStatement: with date filters - ", req.StartDate)
		filter["transaction_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"transaction_time", -1}})
	// 3. Execute Query
	cursor, err := OperatorLedgerCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllOperatorStatement: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return records, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		record := models.OperatorLedgerDto{}
		err := cursor.Decode(&record)
		if err != nil {
			log.Println("GetAllOperatorStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		switch strings.ToUpper(req.TxType) {
		case "FUNDS":
			if strings.ToUpper(record.TransactionType) == "DEPOSIT-FUNDS" || strings.ToUpper(record.TransactionType) == "WITHDRAW-FUNDS" {
				records = append(records, record)
			}
		case "BETS":
			if strings.ToUpper(record.TransactionType) != "DEPOSIT-FUNDS" && strings.ToUpper(record.TransactionType) != "WITHDRAW-FUNDS" {
				records = append(records, record)
			}
		default:
			records = append(records, record)
		}
	}
	return records, nil
}
*/
// Insert operator Document
func InsertOperatorLedger(ledger models.OperatorLedgerDto) error {
	//log.Println("InsertOperatorLedger: Adding Documnet for UserKey - ", ledger.UserKey)
	result, err := OperatorLedgerCollection.InsertOne(Ctx, ledger)
	if err != nil {
		log.Println("InsertOperatorLedger: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertOperatorLedger: Document _id is - ", result.InsertedID)
	return nil
}

// Bulk insert ledger objects
func InsertOperatorLedgers(ledgers []models.OperatorLedgerDto) error {
	//log.Println("InsertOperatorLedgers: Adding Legers, count is - ", len(ledgers))
	operatorLedgers := []interface{}{}
	for _, ledger := range ledgers {
		operatorLedgers = append(operatorLedgers, ledger)
	}
	result, err := OperatorLedgerCollection.InsertMany(Ctx, operatorLedgers)
	if err != nil {
		log.Println("InsertOperatorLedgers: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertOperatorLedgers: Inserted Ledger count is - ", len(result.InsertedIDs))
	return nil
}

// Get operators ledger history by operatorkey
func GetOperatorLedgers(operatorId string) ([]models.OperatorLedgerDto, error) {
	// TODO: Sort by transaction time descending (Latest first)
	//log.Println("GetLedgers: Looking for operatorId - ", operatorId)
	result := []models.OperatorLedgerDto{}
	opts := options.Find()
	opts.Sort = bson.D{{"transaction_time", -1}}
	filter := bson.D{{"operator_id", operatorId}}
	cursor, err := OperatorLedgerCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetOperatorLedgers: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		operator := models.OperatorLedgerDto{}
		err = cursor.Decode(&operator)
		if err != nil {
			log.Println("GetOperatorLedgers: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, operator)
	}
	log.Println("GetOperatorLedgers: total operators - ", len(result))
	return result, nil
}

// Get operators ledger history by operatorId and userId
func GetOperatorLedgersByUserId(operatorId, userId, referenceId string, startTime, EndTime int64) ([]models.OperatorLedgerDto, error) {
	result := []models.OperatorLedgerDto{}
	opts := options.Find()
	opts.Sort = bson.D{{"transaction_time", -1}}
	filter := bson.M{}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	} else {
		return []models.OperatorLedgerDto{}, errors.New("operatorId is required")
	}
	if userId != "" {
		filter["user_id"] = userId
	} else {
		return []models.OperatorLedgerDto{}, errors.New("userId is required")
	}
	if referenceId != "" {
		filter["reference_id"] = referenceId
	}
	if startTime > 0 && EndTime > 0 {
		filter["transaction_time"] = bson.M{"$gte": startTime, "$lte": EndTime}
	}
	cursor, err := OperatorLedgerCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetOperatorLedgersByUserId: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		operator := models.OperatorLedgerDto{}
		err = cursor.Decode(&operator)
		if err != nil {
			log.Println("GetOperatorLedgersByUserId: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, operator)
	}
	log.Println("GetOperatorLedgersByUserId: total operators - ", len(result))
	return result, nil
}

// Get operators ledger history by operatorkey for fund statement
func GetOperatorLedgersForFundStatement(operatorId string) ([]models.OperatorLedgerDto, error) {
	// TODO: Sort by transaction time descending (Latest first)
	//log.Println("GetLedgers: Looking for operatorId - ", operatorId)
	result := []models.OperatorLedgerDto{}
	opts := options.Find()
	opts.Sort = bson.D{{"transaction_time", -1}}
	filter := bson.D{{"operator_id", operatorId}}
	// filter transactions for deposit and withdraw
	filter = append(filter, bson.E{Key: "transaction_type", Value: bson.D{{"$in", []string{"Deposit-Funds", "Withdraw-Funds"}}}})
	cursor, err := OperatorLedgerCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetOperatorLedgersForFundStatement: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		operator := models.OperatorLedgerDto{}
		err = cursor.Decode(&operator)
		if err != nil {
			log.Println("GetOperatorLedgersForFundStatement: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, operator)
	}
	log.Println("GetOperatorLedgersForFundStatement: total operators - ", len(result))
	return result, nil
}

func GetAllOperatorLedgers() ([]models.OperatorLedgerDto, error) {
	start := 0
	last := 10000
	pageSize := 40

	options := options.Find()
	options.SetLimit(int64(pageSize)).SetSkip(int64(start))
	count := 0
	ledgerCount := 0
	for i := start; i <= last; i = i + 40 {
		result := []models.OperatorLedgerDto{}
		options.SetSkip(int64(i))

		cursor, err := OperatorLedgerCollection.Find(Ctx, bson.D{}, options)
		if err != nil {
			log.Println("GetAllOperatorLedgers: Failed with error - ", err.Error())
			return result, err
		}
		defer cursor.Close(Ctx)
		for cursor.Next(Ctx) {
			operator := models.OperatorLedgerDto{}
			err = cursor.Decode(&operator)
			if err != nil {
				count += 1
				log.Println("GetAllOperatorLedgers: Decode failed with error - ", err.Error())
				continue
			}
			result = append(result, operator)
		}
		ledgerCount += len(result)
		// log.Println("GetAllOperatorLedgers: Fetched ", ledgerCount, " ledgers")
		if len(result) < 40 {
			break
		}
	}
	log.Println("GetAllOperatorLedgers: Total failed Operator ledger count - ", count)
	return []models.OperatorLedgerDto{}, nil
}
