package database

import (
	"Sp/dto/models"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserDelta struct {
	UserKey string
	Delta   float64
}

// Insert user Document
func InsertB2BUser(b2bUser models.B2BUserDto) error {
	//log.Println("InsertB2BUser: Adding Documnet for UserKey - ", b2bUser.UserKey)
	result, err := B2BUserCollection.InsertOne(Ctx, b2bUser)
	if err != nil {
		log.Println("InsertB2BUser: FAILED to INSERT - ", err.Error())
		return err
	}
	log.Println("InsertB2BUser: Document _id is - ", result.InsertedID)
	return nil
}

// Update user Document
func UpdateB2BUserStatus(userKey string, status string) error {
	//log.Println("UpdateB2BUserStatus: Updating Documnet for UserKey - ", userKey)
	opts := options.Update()
	filter := bson.D{{"user_key", userKey}}
	update := bson.D{{"$set", bson.D{{"status", status}}}}
	result, err := B2BUserCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateB2BUserStatus: FAILED to UPDATE - ", err.Error())
		return err
	}
	log.Println("UpdateB2BUserStatus: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateB2BUserStatus: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update user balance. delta value can be +ve for deposit, -ve for withdraw
func UpdateB2BUserBalance(userKey string, deltaValue float64) error {
	//log.Println("UpdateB2BUserBalance: Updating balance for UserKey - ", userKey)
	opts := options.Update()
	filter := bson.D{{"user_key", userKey}}
	update := bson.D{{"$inc", bson.D{{"balance", deltaValue}}}}
	result, err := B2BUserCollection.UpdateOne(Ctx, filter, update, opts)
	if err != nil {
		log.Println("UpdateB2BUserBalance: FAILED to UPDATE - ", err.Error())
		return err
	}
	log.Println("UpdateB2BUserBalance: Matched recoreds Count - ", result.MatchedCount)
	log.Println("UpdateB2BUserBalance: Modified recoreds Count - ", result.ModifiedCount)
	return nil
}

// Update user balance. delta value can be +ve for deposit, -ve for withdraw
func UpdateB2BUserBalances(users []UserDelta) (int, error) {
	//log.Println("UpdateB2BUserBalance: Updating user balances for count - ", len(users))
	count1 := 0
	count2 := 0
	// TODO: Find a way to update documents in one DB call
	for _, user := range users {
		count1++
		err := UpdateB2BUserBalance(user.UserKey, user.Delta)
		if err != nil {
			log.Println("UpdateB2BUserBalance: FAILED to UPDATE - ", err.Error())
			// TODO: Handle failures
			continue
		}
		count2++
	}
	log.Println("UpdateB2BUserBalance: Matched recoreds Count - ", count1)
	log.Println("UpdateB2BUserBalance: Modified recoreds Count - ", count2)
	return count2, nil
}

// Get user document
func GetB2BUser(userKey string) (models.B2BUserDto, error) {
	//log.Println("GetB2BUser: Looking for UserKey - ", userKey)
	result := models.B2BUserDto{}
	err := B2BUserCollection.FindOne(Ctx, bson.M{"user_key": userKey}).Decode(&result)
	if err != nil {
		log.Println("GetB2BUser: Failed with error - ", err.Error())
		return result, err
	}
	return result, nil
}

// Get users by operator
func GetB2BUsers(operatorId, partialUserName string) ([]models.B2BUserDto, error) {
	//log.Println("GetB2BUsers: Looking for operatorId - ", operatorId)
	result := []models.B2BUserDto{}
	opts := options.Find()
	// sort by user_name
	opts.SetSort(bson.D{{"user_name", 1}})
	filter := bson.D{}
	if partialUserName != "" {
		filter = bson.D{{"operator_id", operatorId}, {"$or", bson.A{bson.D{{"user_name", bson.M{"$regex": partialUserName, "$options": "i"}}}, bson.D{{"user_key", bson.M{"$regex": partialUserName, "$options": "i"}}}}}}
		opts.SetLimit(10)
	} else {
		filter = bson.D{{"operator_id", operatorId}}
	}
	cursor, err := B2BUserCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetB2BUsers: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := models.B2BUserDto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetB2BUsers: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	//log.Println("GetB2BUsers: total users - ", len(result))
	return result, nil
}

// Get All users
func GetAllB2BUsers() ([]models.B2BUserDto, error) {
	//log.Println("GetB2BUsers: Looking for All B2B Users")
	result := []models.B2BUserDto{}
	opts := options.Find()
	filter := bson.D{}
	cursor, err := B2BUserCollection.Find(Ctx, filter, opts)
	if err != nil {
		log.Println("GetB2BUsers: Failed with error - ", err.Error())
		return result, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		user := models.B2BUserDto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetB2BUsers: Decode failed with error - ", err.Error())
			continue
		}
		result = append(result, user)
	}
	//log.Println("GetB2BUsers: total users - ", len(result))
	return result, nil
}

func GetAllB2BUsersForAudit() ([]models.B2BUserDto, error) {
	//log.Println("GetB2BUsers: Looking for All B2B Users")
	start := 0
	last := 10000
	pageSize := 40
	filter := bson.D{}

	options := options.Find()
	options.SetLimit(int64(pageSize)).SetSkip(int64(start))
	count := 0
	ledgerCount := 0
	for i := start; i <= last; i = i + 40 {
		result := []models.B2BUserDto{}
		options.SetSkip(int64(i))

		cursor, err := B2BUserCollection.Find(Ctx, filter, options)
		if err != nil {
			log.Println("GetB2BUsers: Failed with error - ", err.Error())
			return result, err
		}
		defer cursor.Close(Ctx)
		for cursor.Next(Ctx) {
			user := models.B2BUserDto{}
			err = cursor.Decode(&user)
			if err != nil {
				count += 1
				log.Println("GetB2BUsers: Decode failed with error - ", err.Error())
				continue
			}
			result = append(result, user)
		}
		ledgerCount += len(result)
		// log.Println("GetB2BUsers: Fetched ", ledgerCount, " ledgers")
		if len(result) < 40 {
			break
		}
	}
	log.Println("GetB2BUsers: Total failed B2BUsers count - ", count)
	return []models.B2BUserDto{}, nil
}

// Get B2BUser Document
func GetB2BUsersByKeys(userKeys []string) ([]models.B2BUserDto, error) {
	//log.Println("GetB2BUser: Looking for user_key - ", userKey)
	users := []models.B2BUserDto{}
	if len(userKeys) == 0 {
		return users, nil
	}
	// 1. Create Filter
	filter := bson.M{}
	// 2. Create Find options - add sort
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"updated_at", -1}})
	filter["user_key"] = bson.M{"$in": userKeys}
	cursor, err := B2BUserCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetB2BUsersByKeys: Failed with error - ", err.Error())
		return users, err
	}
	defer cursor.Close(Ctx)
	// 4. Iterate through cursor
	for cursor.Next(Ctx) {
		user := models.B2BUserDto{}
		err = cursor.Decode(&user)
		if err != nil {
			log.Println("GetB2BUsersByKeys: Decode failed with error - ", err.Error())
			continue
		}
		users = append(users, user)
	}
	return users, nil
}
