package database

import (
	operatordto "Sp/dto/operator"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert Operator Document
func InsertOperatorDetails(operatorDto operatordto.OperatorDTO) error {
	//log.Println("InsertOperatorDetails: Adding Documnet for operatorId - ", operatorDto.OperatorId)
	operatorDto.CreatedAt = time.Now().Unix()
	operatorDto.UpdatedAt = operatorDto.CreatedAt
	result, err := OperatorCollection.InsertOne(Ctx, operatorDto)
	if err != nil {
		log.Println("InsertOperatorDetails: FAILED to INSERT session details - ", err.Error())
		return err
	}
	log.Println("InsertOperatorDetails: Document _id is - ", result.InsertedID)
	return nil
}

func UpdateOperators(operators []operatordto.OperatorDTO) (int, []string, error) {
	msgs := []string{}
	count1 := 0
	count2 := 0
	// TODO: Find a way to update documents in one DB call
	for _, operator := range operators {
		count1++
		err := ReplaceOperator(operator)
		if err != nil {
			log.Println("UpdateOperators: FAILED to UPDATE - ", err.Error())
			// TODO: Handle failures
			msgs = append(msgs, operator.OperatorId+": "+err.Error())
			continue
		}
		count2++
	}
	log.Println("UpdateOperators: Matched recoreds Count - ", count1)
	log.Println("UpdateOperators: Modified recoreds Count - ", count2)
	return count2, msgs, nil
}

// Update Operator status
func UpdateOperatorStatus(operatorId string, status string) error {
	log.Println("UpdateOperatorStatus: Status changing for the operator: ", operatorId+"-"+status)
	// 1. Get Operator from Database
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		//log.Println("UpdateOperatorStatus: Operator NOT FOUND - ", err.Error())
		return err
	}
	// 2. Update partner status
	operator.Status = status
	// 3. Replace Operator
	err = ReplaceOperator(operator)
	if err != nil {
		//log.Println("UpdateOperatorStatus: Operator NOT FOUND - ", err.Error())
		return err
	}
	return nil
}

// Update Partner status only SA
func UpdatePartnerStatus(operatorId string, partnerId string, status string) error {
	log.Println("UpdatePartnerStatus: Status changing for the operator + partner: ", operatorId+"-"+partnerId+"-"+status)
	// 1. Get Operator from Database
	operator, err := GetOperatorDetails(operatorId)
	if err != nil {
		//log.Println("UpdatePartnerStatus: Operator NOT FOUND - ", err.Error())
		return err
	}
	// 2. Update partner status
	for i, partner := range operator.Partners {
		if partner.PartnerId == partnerId {
			operator.Partners[i].Status = status
			break
		}
	}
	// 3. Replace Operator
	err = ReplaceOperator(operator)
	if err != nil {
		//log.Println("UpdatePartnerStatus: Operator NOT FOUND - ", err.Error())
		return err
	}
	return nil
}

// Get Operator Document
func GetOperatorDetails(operatorId string) (operatordto.OperatorDTO, error) {
	//log.Println("GetSessionDetailsByToken: Looking for operator_id - ", operatorId)
	operatorDto := operatordto.OperatorDTO{}
	err := OperatorCollection.FindOne(Ctx, bson.M{"operator_id": operatorId}).Decode(&operatorDto)
	if err != nil {
		log.Println("GetOperatorDetails: Operator NOT FOUND - ", err.Error())
		return operatorDto, err
	}
	return operatorDto, nil
}

// Get All Operators
func GetAllOperators() ([]operatordto.OperatorDTO, error) {
	//log.Println("GetAllOperators: ")
	operatorDtos := []operatordto.OperatorDTO{}
	cursor, err := OperatorCollection.Find(Ctx, bson.D{})
	if err != nil {
		log.Println("GetAllOperators: Failed with error - ", err.Error())
		return operatorDtos, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		operator := operatordto.OperatorDTO{}
		err = cursor.Decode(&operator)
		if err != nil {
			log.Println("GetAllOperators: Decode failed with error - ", err.Error())
			continue
		}
		operatorDtos = append(operatorDtos, operator)
	}
	return operatorDtos, nil
}

// Update Operator Document
func ReplaceOperator(operatorDto operatordto.OperatorDTO) error {
	//log.Println("UpdateSessionDetails: Updating Documnet for operatorId - ", operatorDto.OperatorId)
	// opts := options.Update()
	operatorDto.UpdatedAt = time.Now().Unix()
	filter := bson.D{{"_id", operatorDto.ID}}
	// update := bson.D{{"$set", operatorDto}}
	result, err := OperatorCollection.ReplaceOne(Ctx, filter, operatorDto)
	if err != nil {
		log.Println("ReplaceOperator: FAILED to UPDATE operator details - ", err.Error())
		return err
	}
	log.Println("ReplaceOperator: Matched recoreds Count - ", result.MatchedCount)
	log.Println("ReplaceOperator: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update operator balance. delta value can be +ve for deposit, -ve for withdraw
func UpdateOperatorBalance(operatorId string, deltaValue float64) error {
	//log.Println("UpdateOperatorBalance: Updating balance for operatorId - ", operatorId)
	opts := options.Update()
	filter := bson.D{{"operator_id", operatorId}}
	update := bson.D{{"$inc", bson.D{{"balance", deltaValue}}}}
	result, err := OperatorCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateOperatorBalance: FAILED to UPDATE - ", err.Error())
		return err
	}
	log.Println("UpdateOperatorBalance: deltaValue - ", deltaValue)
	log.Println("UpdateOperatorBalance: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateOperatorBalance: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

func DeleteOperatorDetails(operatorId string) error {
	//log.Println("DeleteOperatorDetails: Deleting Documnet for operatorId - ", operatorId)
	filter := bson.D{{"operator_id", operatorId}}
	result, err := OperatorCollection.DeleteOne(Ctx, filter)
	if err != nil {
		log.Println("DeleteOperatorDetails: FAILED to DELETE operator details - ", err.Error())
		return err
	}
	log.Println("DeleteOperatorDetails: Delete recoreds Count - ", result.DeletedCount)
	return nil
}

// Get Operators updated in last 5 minutes
func GetUpdatedOperators() ([]operatordto.OperatorDTO, error) {
	operatorDtos := []operatordto.OperatorDTO{}
	filter := bson.M{}
	afterUpdateAt := time.Now().Add(-1 * 5 * time.Minute).Unix()
	filter["updated_at"] = bson.M{"$gte": afterUpdateAt}
	cursor, err := OperatorCollection.Find(Ctx, filter)
	if err != nil {
		log.Println("GetUpdatedOperators: Failed with error - ", err.Error())
		return operatorDtos, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		operator := operatordto.OperatorDTO{}
		err = cursor.Decode(&operator)
		if err != nil {
			log.Println("GetUpdatedOperators: Decode failed with error - ", err.Error())
			continue
		}
		operatorDtos = append(operatorDtos, operator)
	}
	return operatorDtos, nil
}

// func TestOperatorCalling() ([]operatordto.OperatorDTO, error) {
// 	operatorDtos := []operatordto.OperatorDTO{}
// 	filter := bson.M{}

// 	findOptions := options.Find()
// 	findOptions.SetSkip(2)
// 	findOptions.SetLimit(5)

// 	cursor, err := OperatorCollection.Find(Ctx, filter, findOptions)
// 	if err != nil {
// 		log.Println("GetUpdatedOperators: Failed with error - ", err.Error())
// 		return operatorDtos, err
// 	}
// 	defer cursor.Close(Ctx)
// 	for cursor.Next(Ctx) {
// 		operator := operatordto.OperatorDTO{}
// 		err = cursor.Decode(&operator)
// 		if err != nil {
// 			log.Println("GetUpdatedOperators: Decode failed with error - ", err.Error())
// 			continue
// 		}
// 		operatorDtos = append(operatorDtos, operator)
// 	}
// 	log.Println("GetUpdatedOperators: len operatorDtos - ", len(operatorDtos))
// 	// log.Println("GetUpdatedOperators: operatorDtos - ", operatorDtos)
// 	return operatorDtos, nil
// }

func GetAllOperator() ([]operatordto.OperatorDTO, error) {
	start := 0
	last := 10000
	pageSize := 40
	options := options.Find()

	options.SetLimit(int64(pageSize)).SetSkip(int64(start))
	count := 0
	for i := start; i <= last; i = i + 40 {
		operatorDtos := []operatordto.OperatorDTO{}
		options.SetSkip(int64(i))
		cursor, err := OperatorCollection.Find(Ctx, bson.D{}, options)
		if err != nil {
			log.Println("GetAllOperator: Failed with error - ", err.Error())
			return operatorDtos, err
		}
		defer cursor.Close(Ctx)
		for cursor.Next(Ctx) {
			operator := operatordto.OperatorDTO{}
			err = cursor.Decode(&operator)
			if err != nil {
				count += 1
				log.Println("GetAllOperator: Decode failed with error - ", err.Error())
				continue
			}
			operatorDtos = append(operatorDtos, operator)
		}
		// log.Println("GetAllOperator: Fetched ", len(operatorDtos), " Operators")
		if len(operatorDtos) < 40 {
			break
		}
	}
	log.Println("Getbets: Total failed Operator - ", count)
	return []operatordto.OperatorDTO{}, nil
}
