package cache

import (
	"Sp/database"
	dto "Sp/dto/session"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
)

var tokenSessionCache *ristretto.Cache
var userSessionCache *ristretto.Cache

//const PrefixCoreSessionDetails = "CORE_SESSION_DETAILS_"

func init() {
	tokenSessionCache, _ = InitializeCache(1e7, 1<<30, 64)
	userSessionCache, _ = InitializeCache(1e7, 1<<30, 64)
}

// Save in Cache
func SetSessionDetails(sessDto dto.B2BSessionDto) {
	//log.Println("SetSessionDetails: Updating Cache - ", sessDto.Token+", "+sessDto.UserKey)
	// 1. Update it in token Session Cache
	tokenSessionCache.Set(sessDto.Token, sessDto, 64)
	// 2. Update it in user Session Cache
	userSessionCache.Set(sessDto.UserKey, sessDto, 64)
}

// Get Session Details from Cache using Session Token
// All iFrame communicaitons will use token
func GetSessionDetailsByToken(keyToken string) (dto.B2BSessionDto, error) {
	//log.Println("GetSessionDetailsByToken: looking for token - ", keyToken)
	// 0. Create resp object
	b2bSessionDto := dto.B2BSessionDto{}
	// 1. Get from Cache
	value, found := tokenSessionCache.Get(keyToken)
	if found {
		// 1.1 Token FOUND in cache, retrun session object
		b2bSessionDto = value.(dto.B2BSessionDto)
		return b2bSessionDto, nil
	}
	//log.Println("GetSessionDetailsByUserKey: Token NOT FOUND in Cache - " + keyToken)
	// 2. Token NOT FOUND in cache, get from DB and update to cache
	b2bSessionDto, err := database.GetSessionDetailsByToken(keyToken)
	if err != nil {
		// 2.1 Token NOT FOUND in DB, return error
		log.Println("GetSessionDetailsByToken: Token NOT FOUND in DB for token - ", err.Error(), keyToken)
		return b2bSessionDto, fmt.Errorf("Token NOT FOUND!")
	}
	// 3. Token FOUND in DB, add to Cache
	SetSessionDetails(b2bSessionDto)
	// 4. return session object
	return b2bSessionDto, nil
}

// Get Session Details from Cache using UserKey (operatorId + userId)
func GetSessionDetailsByUserKey(operatorId string, userId string) (dto.B2BSessionDto, error) {
	keyUser := operatorId + "-" + userId
	//log.Println("GetSessionDetailsByUserKey: looking token for userKey - ", keyUser)
	// 0. Create resp object
	b2bSessionDto := dto.B2BSessionDto{}
	// 1. Get from Cache
	value, found := userSessionCache.Get(keyUser)
	if found {
		// 1.1 Token FOUND in cache, retrun session object
		b2bSessionDto = value.(dto.B2BSessionDto)
		return b2bSessionDto, nil
	}
	//log.Println("GetSessionDetailsByUserKey: UserKey NOT FOUND in Cache - " + keyUser)
	// 2. Token NOT FOUND in cache, get from DB and update to cache
	b2bSessionDto, err := database.GetSessionDetailsByUserKey(keyUser)
	if err != nil {
		// 2.1 Token NOT FOUND in DB, return error
		log.Println("GetSessionDetailsByUserKey: Token NOT FOUND in DB - ", err.Error())
		return b2bSessionDto, fmt.Errorf("Token NOT FOUND!")
	}
	// 3. Token FOUND in DB, add to Cache
	SetSessionDetails(b2bSessionDto)
	// 4. return session object
	return b2bSessionDto, nil
}

// Is Session Valid
func IsSessionValid(sessDTO dto.B2BSessionDto) bool {
	expireAt := sessDTO.ExpireAt
	expTime := time.Unix(expireAt, 0)
	if time.Now().Before(expTime) {
		return true
	}
	log.Println("IsSessionValid: Session EXPIRED for - ", sessDTO.OperatorId, sessDTO.PartnerId, sessDTO.UserId, sessDTO.Token, sessDTO.UserIp)
	if sessDTO.OperatorId == "kiaexch" {
		return true
	}
	return false
}

// Is in Grace Period (Less than 30 Mins)
func ExtendValidityIfRequired(sessDTO dto.B2BSessionDto) dto.B2BSessionDto {
	expireAt := sessDTO.ExpireAt
	expTime := time.Unix(expireAt, 0)
	graceTime := time.Now().Add(time.Minute * 30)
	if graceTime.Before(expTime) {
		// Token has more than 30 mins validity
		return sessDTO
	}
	//log.Println("ExtendValidityIfRequired: In GRACE PERIOD!, extending the validity for 8 hours from expiry")
	// 1. Add 8 hours to expireAt
	expTime = expTime.Add(time.Minute * 60 * 8)
	sessDTO.ExpireAt = expTime.Unix()
	// 2. Update it in DB
	err := database.UpdateSessionDetails(sessDTO)
	if err != nil {
		// TODO: Handle error
		log.Println("ExtendValidityIfRequired: DB Insertion FAILED with error - ", err.Error())
	}
	// 3. Update it in Cache
	SetSessionDetails(sessDTO)
	return sessDTO
}

// Is in Grace Period (Less than 30 Mins)
func ExtendValidity(sessDTO dto.B2BSessionDto) dto.B2BSessionDto {
	newExpTime := time.Now()
	// 1. Add 8 hours to expireAt
	expTime := newExpTime.Add(time.Minute * 60 * 8)
	sessDTO.ExpireAt = expTime.Unix()
	// 2. Update it in DB
	err := database.UpdateSessionDetails(sessDTO)
	if err != nil {
		// TODO: Handle error
		log.Println("ExtendValidity: DB Insertion FAILED with error - ", err.Error())
	}
	// 3. Update it in Cache
	SetSessionDetails(sessDTO)
	return sessDTO
}
