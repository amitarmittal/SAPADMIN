package cache

import (
	"Sp/cache"
	"Sp/database"
	dto "Sp/dto/operator"
	"errors"
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
)

var portalSessionCache *ristretto.Cache

func init() {
	portalSessionCache, _ = cache.InitializeCache(1e7, 1<<30, 64)
}

func GetPortalSessionDetails(userId string, tokenString string) (dto.PortalSession, error) {
	// Get from Cache
	value, found := portalSessionCache.Get(tokenString)
	// If not found get from DB and update to cache
	if !found {
		//log.Println("PortalSession details for token : " + userId + " not present in Cache")
		// Include method to fetch from DB
		sessModel, err := database.GetSessionTokenDetails(tokenString, userId)
		//log.Printf("sessModel : %+v\n", sessModel)
		// if present in DB
		if err == nil {
			// copy values from DB to cache structure
			// Save Value from DB in Cache
			SetPortalSessionDetails(sessModel)
			return sessModel, nil
		} else {
			log.Println(err)
			log.Println("PortalSession details for token : " + userId + " not present in DB and Cache")
			return sessModel, errors.New("PortalSession details not present for token: " + userId)
		}
	}
	sessModel := value.(dto.PortalSession)
	//log.Printf("sessModel : %+v\n", sessModel)
	// Check for portal session Validity
	if !isPortalSessionValid(sessModel) {
		log.Println("portal session is not valid ")
		// Delete Token from cache and send empty response with error
		DeletePortalSessionfromCache(userId)
		return dto.PortalSession{}, errors.New("portal session expired")
	}
	return sessModel, nil
}

//Set Portal Session Details
func SetPortalSessionDetails(sessDTO dto.PortalSession) bool {
	resp := portalSessionCache.SetWithTTL(sessDTO.JWTToken, sessDTO, 1, 8*time.Hour)
	if !resp {
		log.Println("Error Saving PortalSession to Cache")
	}
	return resp
}

//Delete Portal Session from Cache
func DeletePortalSessionfromCache(userId string) {
	portalSessionCache.Del(userId)
}

//Check the validity of session - set to 8 hrs.
func isPortalSessionValid(sessDTO dto.PortalSession) bool {
	createdTime := time.Unix(0, sessDTO.ExpiresAt)
	return time.Now().After(createdTime)
}
