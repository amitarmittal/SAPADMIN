package cache

import (
	"Sp/database"
	"Sp/dto/models"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

// Key: SportKey = ProviderId+"-"+SportId
// Value: models.Sport
var sportCache *ristretto.Cache

func init() {
	sportCache, _ = InitializeCache(5000, 1<<12, 500)
}

// Save Sport in Cache
func SetSport(sport models.Sport) {
	// 1. Update it in Cache
	sportCache.Set(sport.SportKey, sport, 0)
	sportCache.Wait()
	_, found := sportCache.Get(sport.SportKey)
	if found {
		//log.Println("sportCache: Sport SAVE is SUCCESS in Cache - " + sport.SportKey)
		return
	}
	log.Println("sportCache: Sport NOT SAVED in Cache - " + sport.SportKey)
}

// Get Sport from Cache
func GetSport(sportKey string) (models.Sport, error) {
	// 0. Create resp object
	sport := models.Sport{}
	// 1. Get from Cache
	value, found := sportCache.Get(sportKey)
	if found {
		// 1.1 FOUND in cache, retrun object
		sport = value.(models.Sport)
		return sport, nil
	}
	log.Println("sportCache: Sport NOT FOUND in Cache - " + sportKey)
	// 2. NOT FOUND in cache, get from DB and update to cache
	sport, err := database.GetSport(sportKey)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("sportCache: Sport NOT FOUND in DB - ", err.Error())
		return sport, fmt.Errorf("Sport NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	SetSport(sport)
	// 4. return object
	return sport, nil
}

// Get Sport from Cache
func GetSportFromCache(sportKey string) (models.Sport, error) {
	// 0. Create resp object
	sport := models.Sport{}
	// 1. Get from Cache
	value, found := sportCache.Get(sportKey)
	if found {
		// 1.1 FOUND in cache, retrun object
		sport = value.(models.Sport)
		return sport, nil
	}
	return sport, fmt.Errorf("Sport NOT FOUND!")
}
