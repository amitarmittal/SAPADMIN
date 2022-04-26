package function

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/models"
	"Sp/dto/responsedto"
	"log"
	"time"
)

// Get Sports by Operator & Provider
// User cases:
// portal service/list-sports by Operator & Provider
func GetOpSports(operatorId string, partnerId string, providerId string) ([]responsedto.SportDto, error) {
	// 0. create default return object
	sportInfos := []responsedto.SportDto{}
	// 1. Get SportStatus by OperatorId & ProviderId from cahce
	mapSS, err := cache.GetOpSportStatus(operatorId, partnerId, providerId)
	if err != nil {
		// 1.1. Return Error
		log.Println("GetOpSports: Failed with error - ", err.Error())
		return sportInfos, err
	}
	// 2. Iterate throuhg map and add only provider allowed sports
	for _, ss := range mapSS {
		if ss.ProviderStatus == "ACTIVE" {
			si := responsedto.SportDto{}
			si.SportId = ss.SportId
			si.SportName = ss.SportName
			si.Status = ss.OperatorStatus
			si.PartnerId = ss.PartnerId
			sportInfos = append(sportInfos, si)
		}
	}
	// 3. Return results
	return sportInfos, nil
}

// Get Active Sports by Operator & Provider
// User cases:
// feed service/list-sports
// core service/get-sportss
func GetOpActiveSports(operatorId string, partnerId string, providerId string) ([]responsedto.SportDto, error) {
	// 0. create default return object
	sportInfos := []responsedto.SportDto{}
	// 1. Get SportStatus by OperatorId & ProviderId from cahce
	mapSS, err := cache.GetOpSportStatus(operatorId, partnerId, providerId)
	if err != nil {
		// 1.1. Return Error
		log.Println("GetOpActiveSports: Failed with error - ", err.Error())
		return sportInfos, err
	}
	// 2. Iterate throuhg map and add only active providers
	//log.Println("GetOpActiveSports: mapSS count is - ", len(mapSS))
	for _, ss := range mapSS {
		if ss.ProviderStatus == "ACTIVE" && ss.OperatorStatus == "ACTIVE" {
			si := responsedto.SportDto{}
			si.SportId = ss.SportId
			si.SportName = ss.SportName
			si.Status = "ACTIVE"
			si.PartnerId = ss.PartnerId
			sportInfos = append(sportInfos, si)
		}
	}
	// 3. Return results
	return sportInfos, nil
}

func IsSportActive(operatorId string, partnerId string, providerId string, sportId string) bool {
	isActive := false
	// 1. Get Sport
	sportKey := providerId + "-" + sportId
	sport, err := cache.GetSport(sportKey)
	if err != nil {
		// 1.1. Sport not found, Return Error
		log.Println("IsSportActive:  GetSport failed with error - ", err.Error())
		return isActive
	}
	// 2. Check Sport Status (Platform level status)
	if sport.Status != "ACTIVE" {
		// 6.2.1 Sport was blocked, Return Error
		log.Println("IsSportActive:  Sport was BLOCKED at Platform Level: ", sport.Status)
		return isActive
	}
	// 3. Get SportStatus for OperatorId
	sportStaus, err := cache.GetSportStatus(operatorId, partnerId, providerId, sportId)
	if err != nil {
		// 3.1. SportStatus not found, Return Error
		log.Println("IsSportActive:  GetSportStatus failed with error - ", err.Error())
		return isActive
	}
	// 4. Check SportStatus (Operator level status)
	if sportStaus.ProviderStatus != "ACTIVE" {
		// 6.4.1. Sport was blocked by Platform admin, Return Error
		log.Println("IsSportActive:  Sport was BLOCKED by Platform Admin: ", sportStaus.ProviderStatus)
		return isActive
	}
	// 5. Check OperatorStatus (Operator level status)
	if sportStaus.OperatorStatus != "ACTIVE" {
		// 5.1. Sport was blocked by Operator Admin, Return Error
		log.Println("IsSportActive:  Sport was BLOCKED by Operator Admin: ", sportStaus.OperatorStatus)
		return isActive
	}
	return true
}

// Get Active Competitions by Operator, Provider & Sport
// User cases:
func GetOpActiveCompetitions(operatorId string, partnerId string, providerId string, sportId string) ([]responsedto.CompetitionDto, error) {
	sportKey := operatorId + "-" + providerId + "-" + sportId
	// 0. create default return object
	competitionInfos := []responsedto.CompetitionDto{}
	// 1. Get Competitions by providerId, sportId
	competitions1, err := cache.GetCompetitions(providerId, sportId)
	// 2. Filter Competitions blocked by PlatformAdmin
	competitions2 := []models.Competition{}
	for _, competition := range competitions1 {
		if competition.Status != "ACTIVE" {
			continue
		}
		competitions2 = append(competitions2, competition)
	}
	// 3. Get CompetitionStatus by OperatorId, ProviderId & SportId from cahce
	listCSs, err := cache.GetCompetitionStatusBySport(operatorId, providerId, sportId)
	if err != nil {
		// 1.1. Return Error
		log.Println("GetOpActiveCompetitions: cache.GetCompetitionStatusBySport failed with error for sportKey - ", sportKey, err.Error())
		return competitionInfos, err
	}
	// 4. Filter Competitions blocked for operator by PlatformAdmin & OperatorAdmin
	for _, competition := range competitions2 {
		isBlocked := false
		for _, cs := range listCSs {
			if competition.CompetitionId == cs.CompetitionId {
				if cs.ProviderStatus != "ACTIVE" || cs.OperatorStatus != "ACTIVE" {
					isBlocked = true
				}
			}
		}
		if isBlocked == true {
			continue
		}
		ci := responsedto.CompetitionDto{}
		ci.SportId = competition.SportId
		ci.SportName = competition.SportName
		ci.CompetitionId = competition.CompetitionId
		ci.CompetitionName = competition.CompetitionName
		ci.Status = "ACTIVE"
		ci.PartnerId = partnerId
		competitionInfos = append(competitionInfos, ci)
	}
	// 3. Return results
	log.Println("GetOpActiveCompetitions: competitionInfos count for sportKey - ", sportKey, len(competitionInfos))
	return competitionInfos, nil
}

func SportStatusMigration() {
	// Read All Operators
	// Read All Sports
	// Read All SportStatus
	// Create existing SportStatus Map by SportKey
	// Create list of SportStatus objects
	// Save in Database

	// Read All Operators
	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("SportStatusMigration: database.GetAllOperators failed with error - ", err.Error())
		return
	}
	log.Println("SportStatusMigration: operators count - ", len(operators))
	// Read All Sports
	sports, err := database.GetAllSports()
	if err != nil {
		log.Println("SportStatusMigration: database.GetAllSports failed with error - ", err.Error())
		return
	}
	log.Println("SportStatusMigration: sports count - ", len(sports))
	// Read All SportStatus
	sportSs, err := database.GetAllSportStatus()
	if err != nil {
		log.Println("SportStatusMigration: database.GetAllSportStatus failed with error - ", err.Error())
		return
	}
	log.Println("SportStatusMigration: existing sportStatus count - ", len(sportSs))
	// Create existing SportStatus Map by SportKey
	mapSS := make(map[string]models.SportStatus)
	for _, ss := range sportSs {
		mapSS[ss.SportKey] = ss
	}
	// Create list of SportStatus objects
	sportStatusList := []models.SportStatus{}
	oldCount := 0
	newCount := 0
	for _, operator := range operators {
		log.Println("SportStatusMigration: operator - ", operator.OperatorId)
		log.Println("SportStatusMigration: partners count - ", len(operator.Partners))
		for _, partner := range operator.Partners {
			for _, sport := range sports {
				sportStatusKey := operator.OperatorId + "-" + partner.PartnerId + "-" + sport.SportKey
				ssKey := operator.OperatorId + "-" + sport.SportKey
				ss, isFound := mapSS[ssKey]
				sportStatus := models.SportStatus{}
				if isFound {
					// create from existing
					sportStatus.SportKey = sportStatusKey
					sportStatus.OperatorId = operator.OperatorId
					sportStatus.OperatorName = operator.OperatorName
					sportStatus.PartnerId = partner.PartnerId
					sportStatus.ProviderId = sport.ProviderId
					sportStatus.ProviderName = sport.ProviderName
					sportStatus.SportId = sport.SportId
					sportStatus.SportName = sport.SportName
					sportStatus.ProviderStatus = ss.ProviderStatus
					sportStatus.OperatorStatus = ss.OperatorStatus
					sportStatus.Favourite = ss.Favourite
					sportStatus.CreatedAt = ss.CreatedAt
					sportStatus.Config = ss.Config
					oldCount++
				} else {
					// create as a new
					log.Println("SportStatusMigration: sportStatusKey - ", ssKey)
					sportStatus.SportKey = sportStatusKey
					sportStatus.OperatorId = operator.OperatorId
					sportStatus.OperatorName = operator.OperatorName
					sportStatus.PartnerId = partner.PartnerId
					sportStatus.ProviderId = sport.ProviderId
					sportStatus.ProviderName = sport.ProviderName
					sportStatus.SportId = sport.SportId
					sportStatus.SportName = sport.SportName
					sportStatus.ProviderStatus = constants.SAP.ObjectStatus.ACTIVE()
					sportStatus.OperatorStatus = constants.SAP.ObjectStatus.ACTIVE()
					sportStatus.Favourite = false
					sportStatus.CreatedAt = time.Now().Unix()
					newCount++
				}
				sportStatusList = append(sportStatusList, sportStatus)
			}
		}
	}
	log.Println("SportStatusMigration: sportStatus count - ", len(sportStatusList))
	log.Println("SportStatusMigration: sportStatus oldcount - ", oldCount)
	log.Println("SportStatusMigration: sportStatus newcount - ", newCount)
	// Save in Database
	// err = database.InsertManySportStatus(sportStatusList)
	// if err != nil {
	// 	log.Println("SportStatusMigration: database.GetAllSports failed with error - ", err.Error())
	// }
}

func SportStatusDelete() {
	// Read All Operators
	// Read All Sports
	// Create list of SportKeys to be deleted
	// Delete SportStatus from Database

	operators, err := database.GetAllOperators()
	if err != nil {
		log.Println("SportStatusDelete: database.GetAllOperators failed with error - ", err.Error())
		return
	}
	log.Println("SportStatusDelete: operators count - ", len(operators))
	sports, err := database.GetAllSports()
	if err != nil {
		log.Println("SportStatusDelete: database.GetAllSports failed with error - ", err.Error())
		return
	}
	log.Println("SportStatusDelete: sports count - ", len(sports))
	sportKeyList := []string{}
	for _, operator := range operators {
		log.Println("SportStatusDelete: operator - ", operator.OperatorId)
		log.Println("SportStatusDelete: partners count - ", len(operator.Partners))
		for _, partner := range operator.Partners {
			for _, sport := range sports {
				sportStatusKey := operator.OperatorId + "-" + partner.PartnerId + "-" + sport.SportKey
				sportKeyList = append(sportKeyList, sportStatusKey)
			}
		}
	}
	log.Println("SportStatusDelete: sportkey count - ", len(sportKeyList))
	log.Println("SportStatusDelete: sportkey - ", sportKeyList)
	err = database.DeleteSportsBySportKeys(sportKeyList)
	if err != nil {
		log.Println("SportStatusDelete: database.GetAllSports failed with error - ", err.Error())
	}
}
