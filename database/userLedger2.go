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
func GetStatement2(req portaldto.UserStatementReqDto) ([]models.UserLedger2Dto, error) {
	records := []models.UserLedger2Dto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["operator_id"] = req.OperatorId
	filter["user_id"] = req.UserId
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		log.Println("GetStatement2: with date filters - ", req.StartDate)
		filter["transaction_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"transaction_time", -1}})
	// 3. Execute Query
	cursor, err := Ledger2Collection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetStatement2: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return records, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		record := models.UserLedger2Dto{}
		err := cursor.Decode(&record)
		if err != nil {
			log.Println("GetStatement2: Bet Decode failed with error - ", err.Error())
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
func GetAllStatement2(req portaldto.UserStatementReqDto) ([]models.UserLedger2Dto, error) {
	records := []models.UserLedger2Dto{}
	// 1. Create Find filter
	filter := bson.M{}
	filter["user_id"] = req.UserId
	if req.StartDate > 0 && req.EndDate > 0 { // check for date filters
		log.Println("GetAllStatement2: with date filters - ", req.StartDate)
		filter["transaction_time"] = bson.M{"$gte": req.StartDate, "$lte": req.EndDate}
	}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"transaction_time", -1}})
	// 3. Execute Query
	cursor, err := Ledger2Collection.Find(Ctx, filter, findOptions)
	defer cursor.Close(Ctx)
	if err != nil {
		log.Println("GetAllStatement2: Bets NOT FOUND for OperatorId : ", req.OperatorId)
		return records, err
	}
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		record := models.UserLedger2Dto{}
		err := cursor.Decode(&record)
		if err != nil {
			log.Println("GetAllStatement2: Bet Decode failed with error - ", err.Error())
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
func InsertLedger2(ledger models.UserLedger2Dto) error {
	result, err := Ledger2Collection.InsertOne(Ctx, ledger)
	if err != nil {
		log.Println("InsertLedger2: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertLedger2: Document _id is - ", result.InsertedID)
	return nil
}

// Bulk insert ledger objects
func InsertManyLedgers2(ledgers []models.UserLedger2Dto) error {
	userLedgers := []interface{}{}
	for _, ledger := range ledgers {
		userLedgers = append(userLedgers, ledger)
	}
	result, err := Ledger2Collection.InsertMany(Ctx, userLedgers)
	if err != nil {
		log.Println("InsertLedgers2: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertLedgers2: Inserted Ledger count is - ", len(result.InsertedIDs))
	return nil
}

// Get users ledger history by userkey
func GetLedgers2(userKey, referenceId string, startTime, endTime int64) ([]models.UserLedger2Dto, error) {
	// TODO: Sort by transaction time descending (Latest first)
	result := []models.UserLedger2Dto{}
	opts := options.Find()
	filter := bson.D{{"user_key", userKey}}
	if referenceId != "" {
		filter = append(filter, bson.E{Key: "reference_id", Value: referenceId})
	}
	if startTime > 0 && endTime > 0 {
		filter = append(filter, bson.E{Key: "transaction_time", Value: bson.M{"$gte": startTime, "$lte": endTime}})
	}

	opts.SetSort(bson.D{{"transaction_time", -1}})
	cursor, err := Ledger2Collection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetLedgers2: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := models.UserLedger2Dto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetLedgers2: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	log.Println("GetLedgers2: total users - ", len(result))
	return result, nil
}

func GetAllUserLedgers2() ([]models.UserLedger2Dto, error) {

	start := 0
	last := 10000
	pageSize := 40

	options := options.Find()
	options.SetLimit(int64(pageSize)).SetSkip(int64(start))
	count := 0
	ledgerCount := 0
	for i := start; i <= last; i = i + 40 {
		result := []models.UserLedger2Dto{}
		options.SetSkip(int64(i))
		cursor, err := Ledger2Collection.Find(Ctx, bson.D{}, options)
		if err != nil {
			log.Println("GetAllUserLedgers2: Failed with error - ", err.Error())
			return result, err
		}
		defer cursor.Close(Ctx)
		for cursor.Next(Ctx) {
			user := models.UserLedger2Dto{}
			err = cursor.Decode(&user)
			if err != nil {
				count += 1
				log.Println("GetAllUserLedgers2: Decode failed with error - ", err.Error())
				continue
			}
			result = append(result, user)
		}
		ledgerCount += len(result)
		if len(result) < 40 {
			break
		}
		cursor.Close(Ctx)
	}
	log.Println("GetAllUserLedgers2:  Total failed User ledger count - ", count)
	return []models.UserLedger2Dto{}, nil
}

func GetLedgersByOperatorId2(operatorId string) ([]models.UserLedger2Dto, error) {
	result := []models.UserLedger2Dto{}
	opts := options.Find()
	filter := bson.D{{"operator_id", operatorId}}
	opts.SetSort(bson.D{{"transaction_time", -1}})
	cursor, err := LedgerCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetLedgersByOperatorId2: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := models.UserLedger2Dto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetLedgersByOperatorId2: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	log.Println("GetLedgersByOperatorId2: total users - ", len(result))
	return result, nil
}
