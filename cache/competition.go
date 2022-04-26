package cache

import (
	"Sp/database"
	"Sp/dto/models"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

// 	[ProviderId]								mapKey 1, mapValue1
//		[SportId]								mapKey 2, mapValue2		=> Level 2 Data
//			[CompititionId] -> Compitition		key 3, value 			=> Level 3 Data

// Level 1:
// mapKey1: ProviderId
// mapValue1: Map[SportId]Level2Data

// Level 2:
// mapKey2: SportId
// mapValue2: Map[CompetitionId]Level3Data

// Level 3:
// Key3: CompititionId
// Value3: models.Compitition

var competitionCache *ristretto.Cache

func init() {
	competitionCache, _ = InitializeCache(1000, 1<<12, 100)
}

// Save Competition in Cache
func SetCompetition(competition models.Competition) {
	competitionKey := competition.ProviderId + "-" + competition.SportId + "-" + competition.CompetitionId
	// Level 1 - ProviderId
	mapValue1, isFound := competitionCache.Get(competition.ProviderId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, add to the cache
		log.Println("competitionCache: (SET) Provider NOT FOUND in Cache - " + competitionKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Competition) // initializing map[CompititionId]compitition
		level3Data[competition.CompetitionId] = competition

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Competition) // initializing map[SportId]level3Data
		level2Data[competition.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		competitionCache.Set(competition.ProviderId, level2Data, 0)
		competitionCache.Wait()

		return
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Competition)
	level3Data, isFound := level2Data[competition.SportId]
	if false == isFound {
		// Leve 2 Key not found
		log.Println("competitionCache: (SET) Sport NOT FOUND in Cache - " + competitionKey)

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Competition) // initializing map[CompititionId]compitition
		level3Data[competition.CompetitionId] = competition

		// Level 2 Data - add
		level2Data[competition.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		competitionCache.Set(competition.ProviderId, level2Data, 0)
		competitionCache.Wait()

		return
	}
	// Level 3 - CompititionId

	// Level 3 Data -  add or update
	level3Data[competition.CompetitionId] = competition

	// Level 2 Data - update
	level2Data[competition.SportId] = level3Data

	// Level 1 Data - Update it in Cache
	competitionCache.Set(competition.ProviderId, level2Data, 0)
	competitionCache.Wait()

	return
}

// Get Competition from Cache
func GetCompetition(providerId string, sportId string, competitionId string) (models.Competition, error) {
	// 0. Create resp object
	competition := models.Competition{}
	competitionKey := providerId + "-" + sportId + "-" + competitionId

	// Level 1 - ProviderId
	mapValue1, isFound := competitionCache.Get(providerId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		log.Println("competitionCache: (GET) Provider NOT FOUND in Cache - " + competitionKey)

		// Get from DB
		competition, err := database.GetCompetition(competitionKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("competitionCache: Competition NOT FOUND in DB - ", err.Error())
			log.Println("competitionCache: CompetitionKey is - ", competitionKey)
			return competition, fmt.Errorf("Competition NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Competition) // initializing map[CompititionId]compitition
		level3Data[competition.CompetitionId] = competition

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Competition) // initializing map[SportId]level3Data
		level2Data[competition.SportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		competitionCache.Set(competition.ProviderId, level2Data, 0)
		competitionCache.Wait()

		return competition, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Competition)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		log.Println("competitionCache: (GET) Sport NOT FOUND in Cache - " + competitionKey)

		// Get from DB
		competition, err := database.GetCompetition(competitionKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("competitionCache: Competition NOT FOUND in DB - ", err.Error())
			log.Println("competitionCache: CompetitionKey is - ", competitionKey)
			return competition, fmt.Errorf("Competition NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data = make(map[string]models.Competition) // initializing map[CompititionId]compitition
		level3Data[competition.CompetitionId] = competition

		// Level 2 Data - add
		level2Data[competition.SportId] = level3Data

		// Level 1 Data - Update it in Cache
		competitionCache.Set(competition.ProviderId, level2Data, 0)
		competitionCache.Wait()

		return competition, nil
	}
	// Level 3 - CompititionId
	competition, isFound = level3Data[competitionId]
	if false == isFound {
		// Level 3 Key not found, Get from DB and add to the cache
		log.Println("competitionCache: (GET) Competition NOT FOUND in Cache - " + competitionKey)

		// Get from DB
		competition, err := database.GetCompetition(competitionKey)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("competitionCache: Competition NOT FOUND in DB - ", err.Error())
			log.Println("competitionCache: CompetitionKey is - ", competitionKey)
			return competition, fmt.Errorf("Competition NOT FOUND!")
		}

		// Level 3 Data - add
		level3Data[competition.CompetitionId] = competition

		// Level 2 Data - update
		level2Data[competition.SportId] = level3Data

		// Level 1 Data - update it in Cache
		competitionCache.Set(competition.ProviderId, level2Data, 0)
		competitionCache.Wait()

		return competition, nil
	}

	return competition, nil
}

// Get Competition from Cache
func GetCompetitions(providerId string, sportId string) ([]models.Competition, error) {
	// 0. Create resp object
	competitions := []models.Competition{}
	sportKey := providerId + "-" + sportId
	// Level 1 - ProviderId
	mapValue1, isFound := competitionCache.Get(providerId) // returns map[SportId]level2Data
	if false == isFound {
		// Level 1 Key not found, Get from DB and add to the cache
		log.Println("competitionCache: (GET) Provider NOT FOUND in Cache - " + sportKey)

		// Get from DB
		competitions, err := database.GetCompetitionsbySport(providerId, sportId)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("competitionCache: Competitions NOT FOUND in DB for sportKey - ", sportKey, err.Error())
			return competitions, fmt.Errorf("Competitions NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Competition) // initializing map[CompititionId]compitition
		for _, competition := range competitions {
			level3Data[competition.CompetitionId] = competition
		}

		// Level 2 Data - init & add
		level2Data := make(map[string]map[string]models.Competition) // initializing map[SportId]level3Data
		level2Data[sportId] = level3Data

		// Level 1 Data - Save it in Cache (insert)
		competitionCache.Set(providerId, level2Data, 0)
		competitionCache.Wait()

		return competitions, nil
	}
	// Level 2 - SportId
	level2Data := mapValue1.(map[string]map[string]models.Competition)
	level3Data, isFound := level2Data[sportId]
	if false == isFound {
		// Level 2 Key not found, Get from DB and add to the cache
		log.Println("competitionCache: (GET) Sport NOT FOUND in Cache - " + sportKey)

		// Get from DB
		competitions, err := database.GetCompetitionsbySport(providerId, sportId)
		if err != nil {
			// 2.1 NOT FOUND in DB, return error
			log.Println("competitionCache: Competitions NOT FOUND in DB for sportKey - ", sportKey, err.Error())
			return competitions, fmt.Errorf("Competitions NOT FOUND!")
		}

		// Level 3 Data - init & add
		level3Data := make(map[string]models.Competition) // initializing map[CompititionId]compitition
		for _, competition := range competitions {
			level3Data[competition.CompetitionId] = competition
		}

		// Level 2 Data - add
		level2Data[sportId] = level3Data

		// Level 1 Data - Update it in Cache
		competitionCache.Set(providerId, level2Data, 0)
		competitionCache.Wait()

		return competitions, nil
	}
	for _, competition := range level3Data {
		competitions = append(competitions, competition)
	}
	return competitions, nil
}
