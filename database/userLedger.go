package database

import (
	"Sp/constants"
	"Sp/dto/models"
	portaldto "Sp/dto/portal"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get User Statement
func GetStatement(req portaldto.UserStatementReqDto) ([]models.UserLedgerDto, error) {
	records := []models.UserLedgerDto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["operator_id"] = req.OperatorId
	filter["user_id"] = req.UserId
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		log.Println("GetStatement: with date filters - ", req.StartDate)
		filter["transaction_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"transaction_time", -1}})
	// 3. Execute Query
	cursor, err := LedgerCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetStatement: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return records, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		record := models.UserLedgerDto{}
		err := cursor.Decode(&record)
		if err != nil {
			log.Println("GetStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		switch strings.ToUpper(req.TxType) {
		case "FUNDS":
			if strings.ToUpper(record.TransactionType) == "DEPOSIT-FUNDS" || strings.ToUpper(record.TransactionType) == "WITHDRAW-FUNDS" || record.TransactionType == constants.SAP.LedgerTxType.DEPOSIT() || record.TransactionType == constants.SAP.LedgerTxType.WITHDRAW() {
				records = append(records, record)
			}
		case "BETS":
			if strings.ToUpper(record.TransactionType) != "DEPOSIT-FUNDS" && strings.ToUpper(record.TransactionType) != "WITHDRAW-FUNDS" && record.TransactionType != constants.SAP.LedgerTxType.DEPOSIT() && record.TransactionType != constants.SAP.LedgerTxType.WITHDRAW() {
				records = append(records, record)
			}
		default:
			records = append(records, record)
		}
	}
	return records, nil
}

// Get All User Statement
func GetAllStatement(req portaldto.UserStatementReqDto) ([]models.UserLedgerDto, error) {
	records := []models.UserLedgerDto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["user_id"] = req.UserId
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		log.Println("GetStatement: with date filters - ", req.StartDate)
		filter["transaction_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"transaction_time", -1}})
	// 3. Execute Query
	cursor, err := LedgerCollection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetStatement: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return records, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		record := models.UserLedgerDto{}
		err := cursor.Decode(&record)
		if err != nil {
			log.Println("GetStatement: Bet Decode failed with error - ", err.Error())
			continue
		}
		switch strings.ToUpper(req.TxType) {
		case "FUNDS":
			if strings.ToUpper(record.TransactionType) == "DEPOSIT-FUNDS" || strings.ToUpper(record.TransactionType) == "WITHDRAW-FUNDS" || record.TransactionType == constants.SAP.LedgerTxType.DEPOSIT() || record.TransactionType == constants.SAP.LedgerTxType.WITHDRAW() {
				records = append(records, record)
			}
		case "BETS":
			if strings.ToUpper(record.TransactionType) != "DEPOSIT-FUNDS" && strings.ToUpper(record.TransactionType) != "WITHDRAW-FUNDS" && record.TransactionType != constants.SAP.LedgerTxType.DEPOSIT() && record.TransactionType != constants.SAP.LedgerTxType.WITHDRAW() {
				records = append(records, record)
			}
		default:
			records = append(records, record)
		}
	}
	return records, nil
}

// Insert user Document
func InsertLedger(ledger models.UserLedgerDto) error {
	//log.Println("InsertLedger: Adding Documnet for UserKey - ", ledger.UserKey)
	result, err := LedgerCollection.InsertOne(Ctx, ledger)
	if err != nil {
		log.Println("InsertLedger: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertLedger: Document _id is - ", result.InsertedID)
	return nil
}

// Bulk insert ledger objects
func InsertLedgers(ledgers []models.UserLedgerDto) error {
	//log.Println("InsertLedgers: Adding Legers, count is - ", len(ledgers))
	userLedgers := []interface{}{}
	for _, ledger := range ledgers {
		userLedgers = append(userLedgers, ledger)
	}
	result, err := LedgerCollection.InsertMany(Ctx, userLedgers)
	if err != nil {
		log.Println("InsertLedgers: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertLedgers: Inserted Ledger count is - ", len(result.InsertedIDs))
	return nil
}

// Get users ledger history by userkey
func GetLedgers(userKey, referenceId string, startTime, endTime int64) ([]models.UserLedgerDto, error) {
	// TODO: Sort by transaction time descending (Latest first)
	//log.Println("GetLedgers: Looking for operatorId - ", userKey)
	result := []models.UserLedgerDto{}
	opts := options.Find()
	filter := bson.D{{"user_key", userKey}}
	if referenceId != "" {
		filter = append(filter, bson.E{Key: "reference_id", Value: referenceId})
	}
	if startTime > 0 && endTime > 0 {
		filter = append(filter, bson.E{Key: "transaction_time", Value: bson.M{"$gte": startTime, "$lte": endTime}})
	}

	opts.SetSort(bson.D{{"transaction_time", -1}})
	cursor, err := LedgerCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetLedgers: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := models.UserLedgerDto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetLedgers: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	log.Println("GetLedgers: total users - ", len(result))
	return result, nil
}

func GetAllUserLedgers() ([]models.UserLedgerDto, error) {

	start := 0
	last := 10000
	pageSize := 40

	options := options.Find()
	options.SetLimit(int64(pageSize)).SetSkip(int64(start))
	count := 0
	ledgerCount := 0
	for i := start; i <= last; i = i + 40 {
		result := []models.UserLedgerDto{}
		options.SetSkip(int64(i))
		cursor, err := LedgerCollection.Find(Ctx, bson.D{}, options)
		if err != nil {
			log.Println("GetAllUserLedgers: Failed with error - ", err.Error())
			return result, err
		}
		defer cursor.Close(Ctx)
		for cursor.Next(Ctx) {
			user := models.UserLedgerDto{}
			err = cursor.Decode(&user)
			if err != nil {
				count += 1
				log.Println("GetAllUserLedgers: Decode failed with error - ", err.Error())
				continue
			}
			result = append(result, user)
		}
		ledgerCount += len(result)
		// log.Println("GetAllUserLedgers: Fetched ", ledgerCount, " Ledger")
		if len(result) < 40 {
			break
		}
		cursor.Close(Ctx)
	}
	log.Println("GetAllUserLedgers:  Total failed User ledger count - ", count)
	return []models.UserLedgerDto{}, nil
}

func GetLedgersByOperatorId(operatorId string) ([]models.UserLedgerDto, error) {
	result := []models.UserLedgerDto{}
	opts := options.Find()
	filter := bson.D{{"operator_id", operatorId}}
	opts.SetSort(bson.D{{"transaction_time", -1}})
	cursor, err := LedgerCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetLedgersByOperatorId: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := models.UserLedgerDto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetLedgersByOperatorId: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	log.Println("GetLedgersByOperatorId: total users - ", len(result))
	return result, nil
}
