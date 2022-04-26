package providers

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/commondto"
	dto "Sp/dto/core"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	sessdto "Sp/dto/session"
	sportsDto "Sp/dto/sports"
	"Sp/operator"
	utils "Sp/utilities"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	OperatorHttpReqTimeout time.Duration = 15
	DREAM_SPORT                          = "Dream"
	BETFAIR                              = "BetFair"
	SPORT_RADAR                          = "SportRadar"
)

type CommissionConfig struct {
	CommPercentage float64
	CommLevel      string
}

func SportsFeedCall(reqBody []byte, reqType string, url string, reqtimeout time.Duration) ([]byte, error) {
	log.Println("SportsFeedCall: URL is - ", url)
	bReqBody := bytes.NewBuffer(reqBody)
	req, err := http.NewRequest(reqType, url, bReqBody)
	if err != nil {
		log.Println("SportsFeedCall: Failed to create HTTP Req object")
		return []byte{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	// 4. Create HTTP Client with Timeout
	client := &http.Client{
		Timeout: reqtimeout * time.Second,
	}
	// 5. Make Request
	resp, err := client.Do(req)
	//log.Println("SportsFeedCall: Time after request is :  ", time.Now().String())
	if err != nil {
		log.Println("SportsFeedCall: Request Failed with error - ", err.Error())
		return []byte{}, err
	}
	defer resp.Body.Close()
	// 6. Read response body
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("SportsFeedCall: Response ReadAll failed with error - ", err.Error())
		return []byte{}, err
	}
	//log.Println(string(respbody))
	return respbody, nil
}

func SyncPartnerStatus() {
	// 1. Get all operators
	operators, err := database.GetAllOperators()
	if err != nil {
		// 1.1. Error, return
		log.Println("SyncPartnerStatus: GetAllOperators failed with error - ", err.Error())
		return
	}
	if len(operators) == 0 {
		// 1.2. ZERO Operators, return
		log.Println("SyncPartnerStatus: GetAllOperators returned ZERO documents - ", len(operators))
		return
	}
	// 2. Get all Providers
	providers, err := database.GetAllProviders()
	if err != nil {
		// 2.1. Error, return
		log.Println("SyncPartnerStatus: GetAllProviders failed with error - ", err.Error())
		return
	}
	if len(providers) == 0 {
		// 2.2. ZERO Providers, return
		log.Println("SyncPartnerStatus: GetAllProviders returned ZERO documents - ", len(providers))
		return
	}
	// 3. Get all PartnerStatus
	partnerStatus, err := database.GetAllPartnerStatus()
	if err != nil {
		// 3.1. Error, return
		log.Println("SyncPartnerStatus: GetAllPartnerStatus failed with error - ", err.Error())
		return
	}
	if len(partnerStatus) == 0 {
		// 3.2. ZERO PartnerStatus, informational
		log.Println("SyncPartnerStatus: GetAllPartnerStatus returned ZERO documents - ", len(partnerStatus))
	}
	// 4. Add missing partnerStatus documents to collection
	// 4.1. Create empty PartnerStatus list
	pss := []models.PartnerStatus{}
	// 4.2. Iterator through all operators
	for _, operator := range operators {
		// 4.2.1. Iterator through all partners
		for _, partner := range operator.Partners {
			// 4.2.1.1 Iterator through all providers
			for _, provider := range providers {
				psKey := operator.OperatorId + "-" + partner.PartnerId + "-" + provider.ProviderId // OperatorId+"-"+PartnerId+"-"+ProviderId
				isFound := false
				// 4.2.1.1.1 Iterate through all PartnerStatus
				for _, ps := range partnerStatus {
					if ps.PartnerKey == psKey {
						// Found, break the loop
						isFound = true
						break
					}
				}
				if isFound {
					// Found, skip the sport
					continue
				}
				// Not found, create and add to the missing list
				ps := models.PartnerStatus{}
				ps.PartnerKey = psKey
				ps.OperatorId = operator.OperatorId
				ps.OperatorName = operator.OperatorName
				ps.PartnerId = partner.PartnerId
				ps.ProviderId = provider.ProviderId
				ps.ProviderName = provider.ProviderName
				ps.ProviderStatus = "BLOCKED"
				ps.OperatorStatus = "ACTIVE"
				ps.Favourite = false
				ps.CreatedAt = time.Now().Unix()
				pss = append(pss, ps)
			}
		}
	}
	//log.Println("SyncPartnerStatus: PartnerStatus count added to db & cache is - ", len(pss))
	if len(pss) == 0 {
		return
	}
	// do bulk insert PartnerStatus documents in to DB
	err = database.InsertManyPartnerStatus(pss)
	if err != nil {
		log.Println("SyncPartnerStatus: InsertManyPartnerStatus failed with error - ", err.Error())
		return
	}
	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
	for _, ss := range pss {
		cache.SetPartnerStatus(ss)
	}
}

// Sync Sports to Cache & DB if not present
func SyncSports(sports []commondto.SportDto, providerId string, providerName string) {
	//log.Println("AddSports: Sports count is - ", len(sports))
	// 1. Get all Sports by ProviderId
	dbsports, err := database.GetSports(providerId)
	if err != nil {
		// 2.1. Error, return
		log.Println("AddSports: GetSports failed with error - ", err.Error())
		return
	}
	//log.Println("AddSports: GetSports returned documents - ", len(dbsports))
	// 2. Add missing Sport documents to collection
	// 2.1. Create empty Sport list
	newsports := []models.Sport{}
	// 2.2. Iterator through all sports
	for _, sport := range sports {
		sportKey := providerId + "-" + sport.SportId
		isFound := false
		// 4.2.1. Iterator through all sports from DB
		for _, dbsport := range dbsports {
			if dbsport.SportKey == sportKey {
				// Found, break the loop
				isFound = true
				break
			}
		}
		if isFound {
			// Found, skip the sport
			continue
		}
		// Not found, create and add to the missing list
		newsport := models.Sport{}
		newsport.SportKey = sportKey
		newsport.ProviderId = providerId
		newsport.ProviderName = providerName
		newsport.SportId = sport.SportId
		newsport.SportName = sport.SportName
		newsport.Status = "BLOCKED"
		newsport.Favourite = false
		newsport.CreatedAt = time.Now().Unix()
		newsports = append(newsports, newsport)
	}
	//log.Println("AddSports: Sports count added to db & cache is - ", len(newsports))
	if len(newsports) == 0 {
		return
	}
	// do bulk insert sportStatus documents in to DB
	err = database.InsertManySports(newsports)
	if err != nil {
		log.Println("AddSports: InsertManySports failed with error - ", err.Error())
		return
	}
	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
	for _, ns := range newsports {
		cache.SetSport(ns)
	}
}

func SyncSportStatus(providerId string) {
	// 1. Get all operators
	operators, err := database.GetAllOperators()
	if err != nil {
		// 1.1. Error, return
		log.Println("SyncSportStatus: GetAllOperators failed with error - ", err.Error())
		return
	}
	if len(operators) == 0 {
		// 1.2. ZERO Operators, return
		log.Println("SyncSportStatus: GetAllOperators returned ZERO documents - ", len(operators))
		return
	}
	// 2. Get all Sports by ProviderId
	sports, err := database.GetSports(providerId)
	if err != nil {
		// 2.1. Error, return
		log.Println("SyncSportStatus: GetAllOperators failed with error - ", err.Error())
		return
	}
	if len(sports) == 0 {
		// 2.2. ZERO Operators, return
		log.Println("SyncSportStatus: GetSports returned ZERO documents - ", len(sports))
		return
	}
	// 3. Get all SportStatus by ProviderId
	sportStatuses, err := database.GetPrSports(providerId)
	if err != nil {
		// 3.1. Error, return
		log.Println("SyncSportStatus: GetPrSports failed with error - ", err.Error())
		return
	}
	if len(sportStatuses) == 0 {
		// 3.2. ZERO Operators, informational
		log.Println("SyncSportStatus: GetPrSports returned ZERO documents - ", len(sportStatuses))
	}
	// 4. Add missing SportStatus documents to collection
	// 4.1. Create empty SportsStaus list
	sportStatsuDtos := []models.SportStatus{}
	// 4.2. Iterator through all operators
	for _, operator := range operators {
		// 4.2.1. Iterator through all sports
		for _, sport := range sports {
			sportKey := operator.OperatorId + "-" + sport.SportKey // OperatorId+"-"+ProviderId+"-"+SportId
			isFound := false
			// 4.2.1.1. Iterate through all sportStatus
			for _, ss := range sportStatuses {
				if ss.SportKey == sportKey {
					// Found, break the loop
					isFound = true
					break
				}
			}
			if isFound {
				// Found, skip the sport
				continue
			}
			// Not found, create and add to the missing list
			sportStatus := models.SportStatus{}
			sportStatus.SportKey = sportKey
			sportStatus.OperatorId = operator.OperatorId
			sportStatus.OperatorName = operator.OperatorName
			sportStatus.ProviderId = sport.ProviderId
			sportStatus.ProviderName = sport.ProviderName
			sportStatus.SportId = sport.SportId
			sportStatus.SportName = sport.SportName
			sportStatus.ProviderStatus = "BLOCKED"
			sportStatus.OperatorStatus = "BLOCKED"
			sportStatus.Favourite = false
			sportStatus.CreatedAt = time.Now().Unix()
			sportStatsuDtos = append(sportStatsuDtos, sportStatus)
		}
	}
	//log.Println("SyncSportStatus: SportStatus count added to db & cache is - ", len(sportStatsuDtos))
	if len(sportStatsuDtos) == 0 {
		return
	}
	// do bulk insert sportStatus documents in to DB
	err = database.InsertManySportStatus(sportStatsuDtos)
	if err != nil {
		log.Println("SyncSportStatus: InsertManySportStatus failed with error - ", err.Error())
		return
	}
	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
	for _, ss := range sportStatsuDtos {
		cache.SetSportStatus(ss)
	}
}

// Sync Competitions to Cache & DB if not present
func SyncCompetitions(competitions []commondto.CompetitionDto, sportMap map[string]models.Sport) {
	if len(sportMap) == 0 {
		return
	}
	providerId := ""
	providerName := ""
	for _, sport := range sportMap {
		providerId = sport.ProviderId
		providerName = sport.ProviderName
		break
	}
	// 0. Create unique list of CompetitionKeys to query database
	compKeys := []string{}
	set := make(map[string]bool) // New empty set
	for _, comp := range competitions {
		compKey := providerId + "-" + comp.SportId + "-" + comp.CompetitionId
		exists := set[compKey] // Membership
		if exists {
			continue
		}
		set[compKey] = true
		compKeys = append(compKeys, compKey)
	}
	//log.Println("SyncCompetitions: Competitions count is - ", len(competitions))
	// 1. Get all Competitions by ProviderId
	dbcompetitions, err := database.GetCompetitionsByKeys(compKeys)
	if err != nil {
		// 2.1. Error, return
		log.Println("SyncCompetitions: GetCompetitionsByKeys failed with error - ", err.Error())
		return
	}
	//log.Println("SyncCompetitions: GetCompetitionsByKeys returned documents - ", len(dbcompetitions))
	// 2. Add missing Competition documents to collection
	// 2.1. Create empty Competition list
	newcompetitions := []models.Competition{}
	// 2.2. Iterator through all competitions
	for _, competition := range competitions {
		competitionKey := providerId + "-" + competition.SportId + "-" + competition.CompetitionId
		isFound := false
		// 4.2.1. Iterator through all competitions from DB
		for _, dbcompetition := range dbcompetitions {
			if dbcompetition.CompetitionKey == competitionKey {
				// Found, break the loop
				isFound = true
				break
			}
		}
		if isFound {
			// Found, skip the competition
			continue
		}
		// Not found, create and add to the missing list
		newcompetition := models.Competition{}
		newcompetition.CompetitionKey = competitionKey
		newcompetition.ProviderId = providerId
		newcompetition.ProviderName = providerName
		newcompetition.SportId = competition.SportId
		newcompetition.SportName = sportMap[competition.SportId].SportName
		newcompetition.CompetitionId = competition.CompetitionId
		newcompetition.CompetitionName = competition.CompetitionName
		newcompetition.Status = "BLOCKED"
		if providerId == constants.SAP.ProviderType.Dream() {
			newcompetition.Status = "ACTIVE"
		}
		newcompetition.Favourite = false
		newcompetition.CreatedAt = time.Now().Unix()
		newcompetitions = append(newcompetitions, newcompetition)
	}
	//log.Println("SyncCompetitions: Competitions count added to db & cache is - ", len(newcompetitions))
	if len(newcompetitions) == 0 {
		return
	}
	// do bulk insert competitionStatus documents in to DB
	err = database.InsertManyCompetitions(newcompetitions)
	if err != nil {
		log.Println("SyncCompetitions: InsertManyCompetitions failed with error - ", err.Error())
		return
	}
	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
	for _, ns := range newcompetitions {
		cache.SetCompetition(ns)
	}
	// SyncCompetitionStatus(newcompetitions)
}

// Sync Events to Cache & DB if not present
func SyncEvents(events []dto.EventDto, providerId string, sportId string) {
	//log.Println("SyncEvents: Events count for "+providerId+" "+sportId+" is - ", len(events))
	if len(events) > 0 {
		providerId = events[0].ProviderId
		sportId = events[0].SportId
	}
	eventDtos := []models.Event{}
	// 1. Get all Events by ProviderId & SportId
	dbevents, err := database.GetEvents(providerId, sportId, "", false)
	if err != nil {
		// 2.1. Error, return
		log.Println("SyncEvents: GetEvents failed with error - ", err.Error())
		return
	}
	//log.Println("SyncEvents: GetEvents returned documents - ", len(dbevents))
	// 2. Add missing Event documents to collection
	// 2.2. Iterator through all events
	for _, event := range events {
		// If competitionId is empty or -1 then skip
		// if event.CompetitionId == "" || event.CompetitionId == "-1" {
		// 	continue
		// }
		eventKey := providerId + "-" + sportId + "-" + event.EventId
		isFound := false
		// 4.2.1. Iterator through all events from DB
		for _, dbevent := range dbevents {
			if dbevent.EventKey == eventKey {
				// Found, break the loop
				isFound = true
				break
			}
		}
		if isFound {
			// Found, skip the event
			continue
		}
		// Not found, create and add to the missing list
		eventDto := GetSportsEventDto(event)
		eventDtos = append(eventDtos, eventDto)
	}
	if len(eventDtos) == 0 {
		return
	}
	//log.Println("SyncEvents: Events count added to db is - ", len(eventDtos))
	// do bulk insert competitionStatus documents in to DB
	err = database.InsertManyEvents(eventDtos)
	if err != nil {
		log.Println("AddEvents: InsertManyEvents failed with error - ", err.Error())
		// if len(eventDtos) < 10 {
		// 	InsertEventsInBatches(eventDtos, 1)
		// } else if len(eventDtos) < 100 {
		// 	InsertEventsInBatches(eventDtos, 10)
		// } else if len(eventDtos) < 1000 {
		// 	InsertEventsInBatches(eventDtos, 100)
		// } else {
		// 	InsertEventsInBatches(eventDtos, 1000)
		// }
		return
	}
	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
	for _, ns := range eventDtos {
		cache.SetEvent(ns)
	}
	// Sync EventStatus
	// SyncEventStatus2(eventDtos)
}

// Insert the Events in Batched, Commented in the above function as of now
func InsertEventsInBatches(eventDtos []models.Event, batchSize int) {
	var batches [][]models.Event
	for i := 0; i < len(eventDtos); i += batchSize {
		batches = append(batches, eventDtos[i:min(i+batchSize, len(eventDtos))])
	}

	for _, batch := range batches {
		err := database.InsertManyEvents(batch)
		if err != nil {
			log.Println("AddEvents: Failed to insert events: ", err.Error())
		} else {
			log.Println("AddEvents: Inserted events: ", len(batch))
		}
		time.Sleep(200 * time.Millisecond) // Sleep for 0.2 seconds
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// func SyncEventStatus2(events []models.Event) {
// 	//log.Println("SyncEventStatus2: Events count is - ", len(events))
// 	if len(events) == 0 {
// 		return
// 	}
// 	// 1. Get all operators
// 	operators, err := database.GetAllOperators()
// 	if err != nil {
// 		// 1.1. Error, return
// 		log.Println("SyncEventStatus2: GetAllOperators failed with error - ", err.Error())
// 		return
// 	}
// 	if len(operators) == 0 {
// 		// 1.2. ZERO Operators, return
// 		log.Println("SyncEventStatus2: GetAllOperators returned ZERO documents - ", len(operators))
// 		return
// 	}
// 	//log.Println("SyncEventStatus2: Operators count - ", len(operators))
// 	//log.Println("SyncEventStatus2: Events count - ", len(events))
// 	// 3. Create unique eventKeys list
// 	eventKeys := []string{}
// 	set := make(map[string]bool) // New empty set
// 	for _, operator := range operators {
// 		for _, event := range events {
// 			eventKey := operator.OperatorId + "-" + event.EventKey
// 			exists := set[eventKey] // Membership
// 			if exists {
// 				continue
// 			}
// 			set[eventKey] = true
// 			eventKeys = append(eventKeys, eventKey)
// 		}
// 	}
// 	//log.Println("SyncEventStatus: eventKeys count - ", len(eventKeys))
// 	// 4. Get all EventStatus by eventKeys from db
// 	eventStatuses, err := database.GetUpdatedEventStatus(eventKeys)
// 	if err != nil {
// 		// 4.1. Error, return
// 		log.Println("SyncEventStatus2: GetUpdatedEventStatus failed with error - ", err.Error())
// 		return
// 	}
// 	//log.Println("SyncEventStatus: GetUpdatedEventStatus documents count - ", len(eventStatuses))
// 	// 4. Add missing EventStatus documents to collection
// 	// 4.1. Create empty EventsStaus list
// 	eventStatsuDtos := []models.EventStatus{}
// 	// 4.2. Iterator through all operators
// 	for _, operator := range operators {
// 		// 4.2.1. Iterator through all events
// 		for _, event := range events {
// 			eventKey := operator.OperatorId + "-" + event.EventKey // OperatorId+"-"+ProviderId+"-"+SportId+"-"+EventId
// 			isFound := false
// 			// 4.2.1.1. Iterate through all eventStatus
// 			for _, ss := range eventStatuses {
// 				if ss.EventKey == eventKey {
// 					// Found, break the loop
// 					isFound = true
// 					break
// 				}
// 			}
// 			if isFound {
// 				// Found, skip the event
// 				continue
// 			}
// 			// Not found, create and add to the missing list
// 			eventStatus := models.EventStatus{}
// 			eventStatus.EventKey = eventKey
// 			eventStatus.OperatorId = operator.OperatorId
// 			eventStatus.OperatorName = operator.OperatorName
// 			eventStatus.ProviderId = capatizeProviderId(event.ProviderId)
// 			eventStatus.ProviderName = event.ProviderName
// 			eventStatus.SportId = event.SportId
// 			eventStatus.SportName = event.SportName
// 			eventStatus.CompetitionId = event.CompetitionId
// 			eventStatus.CompetitionName = event.CompetitionName
// 			eventStatus.EventId = event.EventId
// 			eventStatus.EventName = event.EventName
// 			eventStatus.ProviderStatus = "ACTIVE"
// 			eventStatus.OperatorStatus = "ACTIVE"
// 			eventStatus.Favourite = false
// 			eventStatus.CreatedAt = time.Now().Unix()
// 			eventStatus.UpdatedAt = eventStatus.CreatedAt
// 			eventStatsuDtos = append(eventStatsuDtos, eventStatus)
// 		}
// 	}
// 	//log.Println("SyncEventStatus: EventStatus count added to db & cache is - ", len(eventStatsuDtos))
// 	if len(eventStatsuDtos) == 0 {
// 		return
// 	}
// 	//log.Println("SyncEventStatus2: Events count added to db is - ", len(eventStatsuDtos))
// 	// do bulk insert eventStatus documents in to DB
// 	err = database.InsertManyEventStatus(eventStatsuDtos)
// 	if err != nil {
// 		log.Println("SyncEventStatus2: InsertManyEventStatus failed with error - ", err.Error())
// 		// if len(eventStatsuDtos) < 10 {
// 		// 	InsertEventsStatusInBatches(eventStatsuDtos, 1)
// 		// } else if len(eventStatsuDtos) < 100 {
// 		// 	InsertEventsStatusInBatches(eventStatsuDtos, 10)
// 		// } else if len(eventStatsuDtos) < 1000 {
// 		// 	InsertEventsStatusInBatches(eventStatsuDtos, 100)
// 		// } else {
// 		// 	InsertEventsStatusInBatches(eventStatsuDtos, 1000)
// 		// }
// 		return
// 	}
// 	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
// 	for _, ss := range eventStatsuDtos {
// 		cache.SetEventStatus(ss)
// 	}
// }

// Insert the Events in Batched, Commented in the above function as of now
// func InsertEventsStatusInBatches(eventStatsuDtos []models.EventStatus, batchSize int) {
// 	var batches [][]models.EventStatus
// 	for i := 0; i < len(eventStatsuDtos); i += batchSize {
// 		batches = append(batches, eventStatsuDtos[i:min(i+batchSize, len(eventStatsuDtos))])
// 	}

// 	for _, batch := range batches {
// 		err := database.InsertManyEventStatus(batch)
// 		if err != nil {
// 			log.Println("AddEvents: Failed to insert event status: ", err.Error())
// 		} else {
// 			log.Println("AddEvents: Inserted event status: ", len(batch))
// 		}
// 		time.Sleep(200 * time.Millisecond) // Sleep for 0.2 seconds
// 	}
// }

/*
// Add Events to Cache & DB if not present
func AddEvents(events []dto.EventDto) {
	// 1. Iterate through event collection
	//log.Println("AddEvents: Events count is - ", len(events))
	eventDtos := []models.Event{}
	for _, event := range events {
		eventKey := event.ProviderId + "-" + event.SportId + "-" + event.EventId
		//log.Println("AddEvents: event key is - ", eventKey)
		// 2. Get from Cahce
		_, err := cache.GetEvent(eventKey)
		if err == nil {
			// 2.1. Found the Event (in either in chace or added to cache from DB)
			continue
		}
		//log.Println("AddEvents: Event not found in cache & db - ", err.Error())
		// 3. Convert Sp/dto/core/EventDto coming from sports feed to Sp/dto/sports/EventDto stored in cache & db
		eventDto := GetSportsEventDto(event)
		eventDtos = append(eventDtos, eventDto)
	}
	if len(eventDtos) > 0 {
		//log.Println("AddEvents: Events count added to db is - ", len(eventDtos))
		err := database.InsertManyEvents(eventDtos)
		if err != nil {
			// 2.1. Found the Event (in either in chace or added to cache from DB)
			log.Println("AddEvents: InsertManyEvents failed with error - ", err.Error())
		}
		for _, eventDto := range eventDtos {
			cache.SetEvent(eventDto)
		}
	}
}
*/
// Object convertion Sp/dto/core/EventDto -> Sp/dto/sports/EventDto
func GetSportsEventDto(event dto.EventDto) models.Event {
	eventDto := models.Event{}
	eventDto.EventKey = event.ProviderId + "-" + event.SportId + "-" + event.EventId
	eventDto.ProviderId = capatizeProviderId(event.ProviderId)
	eventDto.SportId = event.SportId
	eventDto.SportName = event.SportName
	eventDto.CompetitionId = event.CompetitionId
	eventDto.CompetitionName = event.CompetitionName
	eventDto.EventId = event.EventId
	eventDto.EventName = event.EventName
	eventDto.OpenDate = event.OpenDate
	eventDto.Status = "ACTIVE"
	eventDto.Favourite = false
	eventDto.CreatedAt = time.Now().Unix()
	eventDto.UpdatedAt = eventDto.CreatedAt
	// Get ProviderName from cache
	provider := models.Provider{}
	provider, err := cache.GetProvider(event.ProviderId)
	if err != nil {
		provider.ProviderName = event.ProviderId
	} else {
		eventDto.ProviderName = provider.ProviderName
	}
	return eventDto
}

// Add Operator Ledger transaction
func OperatorLedgerTx(operator operatordto.OperatorDTO, txType string, amount float64, refId string) error {
	// Save in Operator Ledger
	operatorLedger := GetOperatorLedgerTx(operator, txType, amount, refId)
	// TODO: Add retry Mechanism
	err := database.InsertOperatorLedger(operatorLedger)
	if err != nil {
		// inserting ledger document failed
		log.Println("OperatorLedgerTx: InsertOperatorLedger failed with error - ", err.Error())
		//return err
	}
	// TODO: Add retry Mechanism
	// Update Operator Balance balance and save
	err = database.UpdateOperatorBalance(operatorLedger.OperatorId, operatorLedger.Amount)
	if err != nil {
		// updating operator balance failed
		log.Println("OperatorLedgerTx: UpdateOperatorBalance failed with error - ", err.Error())
		return err
	}
	return nil
}

// Add Operator Ledger transaction
func GetOperatorLedgerTx(operator operatordto.OperatorDTO, txType string, amount float64, refId string) models.OperatorLedgerDto {
	// Save in Operator Ledger
	operatorLedger := models.OperatorLedgerDto{}
	operatorLedger.OperatorId = operator.OperatorId
	operatorLedger.OperatorName = operator.OperatorName
	operatorLedger.TransactionType = txType
	operatorLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	operatorLedger.ReferenceId = refId
	operatorLedger.Amount = amount // -ve value means, debited from user account
	return operatorLedger
}

// Transfer Wallet - Bet
func PlaceBet_Transfer(betDto sportsDto.BetDto) (float64, error) {
	var userBalance float64 = 0
	userKey := betDto.OperatorId + "-" + betDto.UserId
	// 6.2.1. Get user balance
	user, err := database.GetB2BUser(userKey)
	if err != nil {
		// 6.2.1.1. Get user balance failed
		log.Println("PlaceBet_Transfer: get user balance failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	userBalance = user.Balance
	// 6.2.2 Check for sufficient balance
	if betDto.BetReq.DebitAmount > userBalance {
		// 6.2.2.1. Get user balance failed
		log.Println("PlaceBet_Transfer: Insufficient Balance - GOING INTO NEGATIVE BALANCE", user.Balance)
		//respDto.ErrorDescription = "Insufficient Balance"
		//respDto.Balance = user.Balance
		//return c.Status(fiber.StatusOK).JSON(respDto)
		//return userBalance, fmt.Errorf("Insufficient Balance!")
	}
	// 6.2.3. Save in User Ledger
	userLedger := models.UserLedgerDto{}
	userLedger.UserKey = betDto.OperatorId + "-" + betDto.UserId
	userLedger.OperatorId = betDto.OperatorId
	userLedger.UserId = betDto.UserId
	userLedger.TransactionType = constants.SAP.LedgerTxType.BETPLACEMENT() // "Bet-Placement"
	userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	userLedger.ReferenceId = betDto.BetId
	userLedger.Amount = betDto.BetReq.DebitAmount * -1 // -ve value means, debited from user account
	userLedger.Remark = utils.GetRemark(betDto)
	userLedger.CompetitionName = betDto.BetDetails.CompetitionName
	userLedger.EventName = betDto.BetDetails.EventName
	userLedger.MarketType = betDto.BetDetails.MarketType
	userLedger.MarketName = betDto.BetDetails.MarketName
	err = database.InsertLedger(userLedger)
	if err != nil {
		// 6.2.3.1. inserting ledger document failed
		log.Println("PlaceBet_Transfer: insert ledger failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	// 6.2.4. Debit amount from user balance and save
	err = database.UpdateB2BUserBalance(userLedger.UserKey, userLedger.Amount)
	if err != nil {
		// 6.2.4.1. updating user balance failed
		log.Println("PlaceBet_Transfer: update user balance failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	// 6.2.5. set balance to resp object
	userBalance = userBalance - betDto.BetReq.DebitAmount
	// Sync Wallet
	userKeyRateMap := make(map[string]int32)
	userKeyRateMap[userLedger.UserKey] = betDto.BetReq.Rate
	go operator.SyncWallets(userKeyRateMap)
	return userBalance, nil
}

// Seamless Wallet - Bet
func PlaceBet_Seamless(betDto sportsDto.BetDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto) (float64, error) {
	var userBalance float64 = 0
	// 6.2.2. Do operator's wallet call
	sessionDto.BaseURL = operatorDto.BaseURL
	opRespDto, err := operator.WalletBet(betDto, sessionDto, operatorDto.Keys.PrivateKey)
	if err != nil {
		// 6.2.2.1. Wallet call failed
		log.Println("PlaceBet: PlaceBet for Dream failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	if opRespDto.Status != "RS_OK" {
		// 6.2.2.2. Wallet call returned error
		log.Println("PlaceBet: Operator returned error for Wallet Bet. error - ", opRespDto.Status)
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, fmt.Errorf(opRespDto.Status)
	}
	// 6.2.3. Set Balance
	userBalance = opRespDto.Balance
	return userBalance, nil
}

// Seamless Wallet - Bet
func CancelBet_Seamless(operatorDto operatordto.OperatorDTO, cancelledBets []sportsDto.BetDto) ([]sportsDto.BetDto, error) {
	//return betDtos, fmt.Errorf("NOT IMPLEMENTED - CancelBet_Seamless")
	log.Println("CancelBet_Seamless: Bet Settlement Started for operator - ", operatorDto.OperatorId)
	log.Println("CancelBet_Seamless: Operator Base URL is - ", operatorDto.BaseURL)
	count := 0
	for i, bet := range cancelledBets {
		opResp, rollBackReq, err := operator.WalletRollback(constants.BetFair.BetStatus.CANCELLED(), bet, operatorDto.BaseURL, operatorDto.Keys.PrivateKey)
		if err != nil {
			log.Println("CancelBet_Seamless: WalletRollback failed with error - ", err.Error())
			log.Println("CancelBet_Seamless: WalletRollback failed for betId - ", bet.BetId)
			cancelledBets[i].Status = bet.Status + "-failed"
			log.Println("CancelBet_Seamless: Rollback Request is - ", rollBackReq)
			continue
		}
		if opResp.Status != "RS_OK" {
			log.Println("CancelBet_Seamless: Rollback Failed. Status is - ", opResp.Status)
			cancelledBets[i].Status = bet.Status + "-failed"
			continue
		}
		count++
		log.Println("CancelBet_Seamless: Rollback Successfully completed for betId - ", bet.BetId)
	}
	log.Println("CancelBet_Seamless: cancelled success bets count is - ", count)
	count2, msgs := database.UpdateBets(cancelledBets)
	if len(msgs) > 0 {
		log.Println("CancelBet_Seamless: Total bets     are - ", len(cancelledBets))
		log.Println("CancelBet_Seamless: Total success  are - ", count2)
		log.Println("CancelBet_Seamless: Total failures are - ", len(msgs))
		log.Println("CancelBet_Seamless: Error messages are:")
		for _, msg := range msgs {
			log.Println("CancelBet_Seamless: *** ERROR *** - ", msg)
		}
	}
	log.Println("CancelBet_Seamless: Bet Cancellation Ended!!!")
	return cancelledBets, nil
}

// Transfer Wallet - Bet
func CancelBet_Transfer(betDtos []sportsDto.BetDto) ([]sportsDto.BetDto, error) {
	userLedgerTxs := []models.UserLedgerDto{}
	userKeyRateMap := make(map[string]int32)
	usersMap := make(map[string]float64)
	for _, betDto := range betDtos {
		userKey := betDto.OperatorId + "-" + betDto.UserId
		delta, ok := usersMap[userKey]
		if !ok {
			usersMap[userKey] = 0
			delta = 0
		}
		// create a user ledger tx
		userLedger := models.UserLedgerDto{}
		userLedger.UserKey = userKey
		userLedger.OperatorId = betDto.OperatorId
		userLedger.UserId = betDto.UserId
		userLedger.TransactionType = constants.SAP.LedgerTxType.BETCANCEL() // "CANCELLATION"
		userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
		userLedger.ReferenceId = betDto.BetId
		userLedger.Amount = betDto.BetReq.DebitAmount // -ve value means, debited from user account
		userLedger.Remark = utils.GetRemark(betDto)
		userLedger.CompetitionName = betDto.BetDetails.CompetitionName
		userLedger.EventName = betDto.BetDetails.EventName
		userLedger.MarketType = betDto.BetDetails.MarketType
		userLedger.MarketName = betDto.BetDetails.MarketName
		userLedgerTxs = append(userLedgerTxs, userLedger)
		// Add a rollback req to the bet
		// add rollback amount to the user delta
		usersMap[userKey] = delta + betDto.BetReq.DebitAmount
		userKeyRateMap[userLedger.UserKey] = betDto.BetReq.Rate
	}
	err := database.InsertLedgers(userLedgerTxs)
	if err != nil {
		// 6.2.3.1. inserting ledger document failed
		log.Println("CancelBet_Transfer: database.InsertLedgers failed with error - ", err.Error())
		//return err
	}
	for userKey, delta := range usersMap {
		err = database.UpdateB2BUserBalance(userKey, delta)
		if err != nil {
			// 6.2.3.1. inserting ledger document failed
			log.Println("CancelBet_Transfer: database.UpdateB2BUserBalance failed with error - ", err.Error())
			//return err
		}
	}
	// Sync Wallet
	go operator.SyncWallets(userKeyRateMap)
	return betDtos, nil
}

// func ComputeResult(openBet sportsDto.BetDto, result models.Result) (sportsDto.ResultReqDto, error) {
// 	resultReq := sportsDto.ResultReqDto{}
// 	resultReq.ReqId = uuid.New().String()
// 	resultReq.ReqTime = time.Now().UnixNano() / int64(time.Millisecond)
// 	resultReq.CreditAmount = 0
// 	resultReq.RunnerName = result.RunnerName
// 	resultReq.SessionOutcome = result.SessionOutcome
// 	// TODO: Review the logic ????
// 	switch strings.ToUpper(openBet.BetDetails.MarketType) {
// 	case "MATCH_ODDS", "BOOKMAKER":
// 		if openBet.BetDetails.BetType == "BACK" && openBet.BetDetails.RunnerId == result.RunnerId {
// 			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
// 		}
// 		if openBet.BetDetails.BetType == "LAY" && openBet.BetDetails.RunnerId != result.RunnerId {
// 			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
// 		}
// 	case "FANCY":
// 		if openBet.BetDetails.BetType == "BACK" && result.SessionOutcome >= openBet.BetDetails.SessionOutcome {
// 			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
// 		}
// 		if openBet.BetDetails.BetType == "LAY" && result.SessionOutcome < openBet.BetDetails.SessionOutcome {
// 			resultReq.CreditAmount = openBet.BetDetails.OddValue * float64(openBet.BetDetails.StakeAmount)
// 		}
// 	default:
// 		log.Println("ComputeResult: Unexpected MarketType - ", openBet.BetDetails.MarketType)
// 		return resultReq, fmt.Errorf("Invalid Market Type - " + openBet.BetDetails.MarketType)
// 	}
// 	resultReq.CreditAmount = utils.Truncate64(resultReq.CreditAmount)
// 	return resultReq, nil
// }

func ComputeRollback(openBet sportsDto.BetDto, reason string) sportsDto.RollbackReqDto {
	rollbackReq := sportsDto.RollbackReqDto{}
	rollbackReq.ReqId = uuid.New().String()
	rollbackReq.ReqTime = time.Now().UnixMilli()
	rollbackReq.RollbackReason = reason
	rollbackReq.RollbackAmount = 0 // positive means, deposit to user, negative means, deduct from user
	for _, result := range openBet.ResultReqs {
		rollbackReq.RollbackAmount -= result.CreditAmount
		rollbackReq.OperatorAmount -= result.OperatorAmount
		rollbackReq.PlatformAmount -= result.PlatformAmount
	}
	for _, rollback := range openBet.RollbackReqs {
		rollbackReq.RollbackAmount -= rollback.RollbackAmount
		rollbackReq.OperatorAmount -= rollback.OperatorAmount
		rollbackReq.PlatformAmount -= rollback.PlatformAmount
	}
	switch strings.ToLower(reason) {
	case "void", "voided", "cancelled", "deleted", "lapsed", "expired", "timelyvoid":
		rollbackReq.RollbackAmount += openBet.BetReq.DebitAmount
		rollbackReq.OperatorAmount += openBet.BetReq.OperatorAmount
		rollbackReq.PlatformAmount += openBet.BetReq.PlatformAmount
	case "rollback", "voidrollback", "timelyvoidrollback":
		log.Println("ComputeRollback: ResultType - ", reason)
		// nothing to do here
	default:
		log.Println("ComputeRollback: Unexpected ResultType - ", reason)
	}
	rollbackReq.RollbackAmount = utils.Truncate4Decfloat64(rollbackReq.RollbackAmount)
	rollbackReq.OperatorAmount = utils.Truncate4Decfloat64(rollbackReq.OperatorAmount)
	rollbackReq.PlatformAmount = utils.Truncate4Decfloat64(rollbackReq.PlatformAmount)
	return rollbackReq
}

// // Sync Data Methods
// func SyncCompetitionStatus(competitions []models.Competition) {
// 	if len(competitions) == 0 {
// 		log.Println("SyncCompetitionStatus: ZERO competitions provided")
// 		return
// 	}
// 	// 1. Get all operators
// 	operators, err := database.GetAllOperators()
// 	if err != nil {
// 		// 1.1. Error, return
// 		log.Println("SyncCompetitionStatus: GetAllOperators failed with error - ", err.Error())
// 		return
// 	}
// 	if len(operators) == 0 {
// 		// 1.2. ZERO Operators, return
// 		log.Println("SyncCompetitionStatus: GetAllOperators returned ZERO documents - ", len(operators))
// 		return
// 	}
// 	// 2. Create unique list of CompetitionStatusKeys to query database
// 	compKeys := []string{}
// 	set := make(map[string]bool) // New empty set
// 	for _, operator := range operators {
// 		for _, competition := range competitions {
// 			compKey := operator.OperatorId + "-" + competition.CompetitionKey
// 			exists := set[compKey] // Membership
// 			if exists {
// 				continue
// 			}
// 			set[compKey] = true
// 			compKeys = append(compKeys, compKey)
// 		}
// 	}
// 	// 3. Get all CompetitionStatus by CompetitionKeys from db
// 	competitionStatuses, err := database.GetCompetitionStatusByKeys(compKeys)
// 	if err != nil {
// 		// 3.1. Error, return
// 		log.Println("SyncCompetitionStatus: GetCompetitionStatusByKeys failed with error - ", err.Error())
// 		return
// 	}
// 	//log.Println("SyncCompetitionStatus: GetPrCompetitions returned documents - ", len(competitionStatuses))
// 	// 4. Add missing CompetitionStatus documents to collection
// 	// 4.1. Create empty CompetitionsStaus list
// 	competitionStatsuDtos := []models.CompetitionStatus{}
// 	// 4.2. Iterator through all operators
// 	for _, operator := range operators {
// 		// 4.2.1. Iterator through all competitions
// 		for _, competition := range competitions {
// 			competitionKey := operator.OperatorId + "-" + competition.CompetitionKey // OperatorId+"-"+ProviderId+"-"+SportId+"-"+CompetitionId
// 			isFound := false
// 			// 4.2.1.1. Iterate through all competitionStatus
// 			for _, ss := range competitionStatuses {
// 				if ss.CompetitionKey == competitionKey {
// 					// Found, break the loop
// 					isFound = true
// 					break
// 				}
// 			}
// 			if isFound {
// 				// Found, skip the competition
// 				continue
// 			}
// 			// Not found, create and add to the missing list
// 			competitionStatus := models.CompetitionStatus{}
// 			competitionStatus.CompetitionKey = competitionKey
// 			competitionStatus.OperatorId = operator.OperatorId
// 			competitionStatus.OperatorName = operator.OperatorName
// 			competitionStatus.ProviderId = competition.ProviderId
// 			competitionStatus.ProviderName = competition.ProviderName
// 			competitionStatus.SportId = competition.SportId
// 			competitionStatus.SportName = competition.SportName
// 			competitionStatus.CompetitionId = competition.CompetitionId
// 			competitionStatus.CompetitionName = competition.CompetitionName
// 			competitionStatus.ProviderStatus = "ACTIVE"
// 			competitionStatus.OperatorStatus = "ACTIVE"
// 			competitionStatus.Favourite = false
// 			competitionStatus.CreatedAt = time.Now().Unix()
// 			switch competitionStatus.ProviderId {
// 			case constants.SAP.ProviderType.BetFair():
// 				if competitionStatus.ProviderName != constants.SAP.ProviderName.BetFairName() {
// 					log.Println("SyncCompetitionStatus: BetFair ProviderId & ProviderName mismatch occured for competitionkey - ", competitionKey)
// 					continue
// 				}
// 			case constants.SAP.ProviderType.Dream():
// 				if competitionStatus.ProviderName != constants.SAP.ProviderName.DreamName() {
// 					log.Println("SyncCompetitionStatus: Dream ProviderId & ProviderName mismatch occured for competitionkey - ", competitionKey)
// 					continue
// 				}
// 			case constants.SAP.ProviderType.SportRadar():
// 				if competitionStatus.ProviderName != constants.SAP.ProviderName.SportRadarName() {
// 					log.Println("SyncCompetitionStatus: SportRadar ProviderId & ProviderName mismatch occured for competitionkey - ", competitionKey)
// 					continue
// 				}
// 			default:
// 				log.Println("SyncCompetitionStatus: Invalid ProviderType - ", competitionStatus.ProviderId)
// 				continue
// 			}
// 			competitionStatsuDtos = append(competitionStatsuDtos, competitionStatus)
// 		}
// 	}
// 	//log.Println("SyncCompetitionStatus: CompetitionStatus count added to db & cache is - ", len(competitionStatsuDtos))
// 	if len(competitionStatsuDtos) == 0 {
// 		return
// 	}
// 	// do bulk insert competitionStatus documents in to DB
// 	err = database.InsertManyCompetitionStatus(competitionStatsuDtos)
// 	if err != nil {
// 		log.Println("SyncCompetitionStatus: InsertManyCompetitionStatus failed with error - ", err.Error())
// 		return
// 	}
// 	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
// 	for _, ss := range competitionStatsuDtos {
// 		cache.SetCompetitionStatus(ss)
// 	}
// }

// func SyncEventStatus() {
// 	// 1. Get all operators
// 	operators, err := database.GetAllOperators()
// 	if err != nil {
// 		// 1.1. Error, return
// 		log.Println("SyncEventStatus: GetAllOperators failed with error - ", err.Error())
// 		return
// 	}
// 	if len(operators) == 0 {
// 		// 1.2. ZERO Operators, return
// 		log.Println("SyncEventStatus: GetAllOperators returned ZERO documents - ", len(operators))
// 		return
// 	}
// 	//log.Println("SyncEventStatus: Operators count - ", len(operators))
// 	// 2. Get all Events by ProviderId from db
// 	events, err := database.GetLatestEvents()
// 	if err != nil {
// 		// 2.1. Error, return
// 		log.Println("SyncEventStatus: GetLatestEvents failed with error - ", err.Error())
// 		return
// 	}
// 	if len(events) == 0 {
// 		// 2.2. ZERO Operators, return
// 		log.Println("SyncEventStatus: GetLatestEvents returned ZERO documents - ", len(events))
// 		return
// 	}
// 	//log.Println("SyncEventStatus: Events count - ", len(events))
// 	// 3. Create unique eventKeys list
// 	eventKeys := []string{}
// 	set := make(map[string]bool) // New empty set
// 	for _, operator := range operators {
// 		for _, event := range events {
// 			eventKey := operator.OperatorId + "-" + event.EventKey
// 			exists := set[eventKey] // Membership
// 			if exists {
// 				continue
// 			}
// 			set[eventKey] = true
// 			eventKeys = append(eventKeys, eventKey)
// 		}
// 	}
// 	//log.Println("SyncEventStatus: eventKeys count - ", len(eventKeys))
// 	// 4. Get all EventStatus by eventKeys from db
// 	eventStatuses, err := database.GetUpdatedEventStatus(eventKeys)
// 	if err != nil {
// 		// 4.1. Error, return
// 		log.Println("SyncEventStatus: GetUpdatedEventStatus failed with error - ", err.Error())
// 		return
// 	}
// 	//log.Println("SyncEventStatus: GetUpdatedEventStatus documents count - ", len(eventStatuses))
// 	// 4. Add missing EventStatus documents to collection
// 	// 4.1. Create empty EventsStaus list
// 	eventStatsuDtos := []models.EventStatus{}
// 	// 4.2. Iterator through all operators
// 	for _, operator := range operators {
// 		// 4.2.1. Iterator through all events
// 		for _, event := range events {
// 			eventKey := operator.OperatorId + "-" + event.EventKey // OperatorId+"-"+ProviderId+"-"+SportId+"-"+EventId
// 			isFound := false
// 			// 4.2.1.1. Iterate through all eventStatus
// 			for _, ss := range eventStatuses {
// 				if ss.EventKey == eventKey {
// 					// Found, break the loop
// 					isFound = true
// 					break
// 				}
// 			}
// 			if isFound {
// 				// Found, skip the event
// 				continue
// 			}
// 			// Not found, create and add to the missing list
// 			eventStatus := models.EventStatus{}
// 			eventStatus.EventKey = eventKey
// 			eventStatus.OperatorId = operator.OperatorId
// 			eventStatus.OperatorName = operator.OperatorName
// 			eventStatus.ProviderId = capatizeProviderId(event.ProviderId)
// 			eventStatus.ProviderName = event.ProviderName
// 			eventStatus.SportId = event.SportId
// 			eventStatus.SportName = event.SportName
// 			eventStatus.CompetitionId = event.CompetitionId
// 			eventStatus.CompetitionName = event.CompetitionName
// 			eventStatus.EventId = event.EventId
// 			eventStatus.EventName = event.EventName
// 			eventStatus.ProviderStatus = "ACTIVE"
// 			eventStatus.OperatorStatus = "ACTIVE"
// 			eventStatus.Favourite = false
// 			eventStatus.CreatedAt = time.Now().Unix()
// 			eventStatus.UpdatedAt = eventStatus.CreatedAt
// 			eventStatsuDtos = append(eventStatsuDtos, eventStatus)
// 		}
// 	}
// 	//log.Println("SyncEventStatus: EventStatus count added to db & cache is - ", len(eventStatsuDtos))
// 	if len(eventStatsuDtos) == 0 {
// 		return
// 	}
// 	// do bulk insert eventStatus documents in to DB
// 	err = database.InsertManyEventStatus(eventStatsuDtos)
// 	if err != nil {
// 		log.Println("SyncEventStatus: InsertManyEventStatus failed with error - ", err.Error())
// 		return
// 	}
// 	// add to cache - acn be optimized, but, because it is cache and less frequent thing, not an important one
// 	for _, ss := range eventStatsuDtos {
// 		cache.SetEventStatus(ss)
// 	}
// }

// Init Cache Methods
func InitOperatorCache() {
	// 1. Get all latest updated 100 documents from collection
	operators, err := database.GetAllOperators()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitOperatorCache: GetAllOperators failed with error - ", err.Error())
		return
	}
	if len(operators) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitOperatorCache: GetAllOperators documents count - ", len(operators))
	objectMap := make(map[string]interface{})
	for _, operator := range operators {
		cache.SetOperatorDetails(operator)
		objectMap[operator.OperatorId] = operator
	}
	cache.SetObjectMap(constants.SAP.ObjectTypes.OPERATOR(), objectMap)
}

func InitProviderCache() {
	// 1. Get all latest updated 100 documents from collection
	providers, err := database.GetAllProviders()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitProviderCache: GetAllProviders failed with error - ", err.Error())
		return
	}
	if len(providers) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitProviderCache: GetAllProviders documents count - ", len(providers))
	objectMap := make(map[string]interface{})
	for _, provider := range providers {
		cache.SetProvider(provider)
		objectMap[provider.ProviderId] = provider
	}
	cache.SetObjectMap(constants.SAP.ObjectTypes.PROVIDER(), objectMap)
}

func InitPartnerStatusCache() {
	// 1. Get all latest updated 100 documents from PartnerStatus collection
	partnerStatuses, err := database.GetAllPartnerStatus()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitPartnerStatusCache: database.GetAllPartnerStatus failed with error - ", err.Error())
		return
	}
	if len(partnerStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitPartnerStatusCache: GetAllPartnerStatus documents count - ", len(partnerStatuses))
	// TODO: Can be set at tope level key to improve
	for _, ps := range partnerStatuses {
		cache.SetPartnerStatus(ps)
	}
}

func InitSportCache() {
	// 1. Get all latest updated 100 documents from collection
	sports, err := database.GetAllSports()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitSportCache: GetAllSports failed with error - ", err.Error())
		return
	}
	if len(sports) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitSportCache: GetAllSports documents count - ", len(sports))
	for _, sport := range sports {
		cache.SetSport(sport)
	}
}

func InitSportStatusCache() {
	// 1. Get all latest updated 100 documents from SportStatus collection
	sportStatuses, err := database.GetAllSportStatus()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitSportStatusCache: GetAllSportStatus failed with error - ", err.Error())
		return
	}
	if len(sportStatuses) == 0 {
		// 1.2. No records found
		return
	}
	log.Println("InitSportStatusCache: GetAllSportStatus documents count - ", len(sportStatuses))
	// 2. Update cache
	for _, ss := range sportStatuses {
		cache.SetSportStatus(ss)
	}
}

func InitCompetitionCache() {
	// 1. Get all latest updated 100 documents from collection
	competitions, err := database.GetUpdatedCompetitions()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitCompetitionCache: GetUpdatedCompetitions failed with error - ", err.Error())
		return
	}
	if len(competitions) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitCompetitionCache: GetUpdatedCompetitions documents count - ", len(competitions))
	for _, competition := range competitions {
		cache.SetCompetition(competition)
	}
}

func InitCompetitionStatusCache() {
	// 1. Get all latest updated 100 documents from collection
	competitionStatuses, err := database.GetUpdatedCompetitionStatus([]string{})
	if err != nil {
		// 1.1. Error, return
		log.Println("InitCompetitionStatusCache: GetUpdatedCompetitionStatus failed with error - ", err.Error())
		return
	}
	if len(competitionStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitCompetitionStatusCache: GetUpdatedCompetitionStatus documents count - ", len(competitionStatuses))
	for _, cs := range competitionStatuses {
		cache.SetCompetitionStatus(cs)
	}
}

func InitEventCache() {
	// 1. Get all latest updated 100 documents from collection
	events, err := database.GetUpdatedEvents()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitEventCache: GetUpdatedEvents failed with error - ", err.Error())
		return
	}
	if len(events) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitEventCache: GetUpdatedEvents documents count - ", len(events))
	for _, event := range events {
		cache.SetEvent(event)
	}
}

func InitEventStatusCache() {
	// 1. Get all latest updated 100 documents from EventStatus collection
	eventStatuses, err := database.GetUpdatedEventStatus([]string{})
	if err != nil {
		// 1.1. Error, return
		log.Println("InitEventStatusCache: GetUpdatedEventStatus failed with error - ", err.Error())
		return
	}
	if len(eventStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitEventStatusCache: GetUpdatedEventStatus documents count - ", len(eventStatuses))
	for _, ss := range eventStatuses {
		cache.SetEventStatus(ss)
	}
}

func InitMarketCache() {
	// 1. Get all latest updated 100 documents from collection
	markets, err := database.GetUpdatedMarkets()
	if err != nil {
		// 1.1. Error, return
		log.Println("InitMarketCache: database.GetUpdatedMarkets failed with error - ", err.Error())
		return
	}
	if len(markets) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitMarketCache: GetUpdatedMarkets documents count - ", len(markets))
	for _, market := range markets {
		cache.SetMarket(market)
	}
}

func InitMarketStatusCache() {
	// 1. Get all latest updated 100 documents from MarketStatus collection
	marketStatuses, err := database.GetUpdatedMarketStatus([]string{})
	if err != nil {
		// 1.1. Error, return
		log.Println("InitMarketStatusCache: database.GetUpdatedMarketStatus failed with error - ", err.Error())
		return
	}
	if len(marketStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("InitMarketStatusCache: GetUpdatedMarketStatus documents count - ", len(marketStatuses))
	for _, ms := range marketStatuses {
		cache.SetMarketStatus(ms)
	}
}

// Refresh Cache Methods
func RefreshOperatorCache() {
	// 0. Dump cache metrics
	//cache.GetOperatorCacheMetrics()
	// 1. Get all latest updated 100 documents from collection
	operators, err := database.GetAllOperators()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshOperatorCache: GetAllOperators failed with error - ", err.Error())
		return
	}
	if len(operators) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshOperatorCache: GetUpdatedOperators documents count - ", len(operators))
	objectMap := make(map[string]interface{})
	for _, operator := range operators {
		cache.SetOperatorDetails(operator)
		objectMap[operator.OperatorId] = operator
	}
	cache.SetObjectMap(constants.SAP.ObjectTypes.OPERATOR(), objectMap)
}

func RefreshProviderCache() {
	// 0. Dump cache metrics
	//cache.GetProviderCacheMetrics()
	// 1. Get all latest updated 100 documents from collection
	providers, err := database.GetAllProviders()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshProviderCache: GetAllProviders failed with error - ", err.Error())
		return
	}
	if len(providers) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshProviderCache: GetUpdatedProviders documents count - ", len(providers))
	objectMap := make(map[string]interface{})
	for _, provider := range providers {
		cache.SetProvider(provider)
		objectMap[provider.ProviderId] = provider
	}
	cache.SetObjectMap(constants.SAP.ObjectTypes.PROVIDER(), objectMap)
}

func RefreshPartnerStatusCache() {
	// 1. Get all latest updated 100 documents from PartnerStatus collection
	partnerStatuses, err := database.GetUpdatedPartners()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshPartnerStatusCache: database.GetUpdatedPartners failed with error - ", err.Error())
		return
	}
	if len(partnerStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshPartnerStatusCache: GetUpdatedPartners documents count - ", len(partnerStatuses))
	// TODO: Can be set at tope level key to improve
	for _, ps := range partnerStatuses {
		cache.SetPartnerStatus(ps)
	}
}

func RefreshSportCache() {
	// 1. Get all latest updated 100 documents from collection
	sports, err := database.GetUpdatedSports()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshSportCache: GetUpdatedSports failed with error - ", err.Error())
		return
	}
	if len(sports) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshSportCache: GetUpdatedSports documents count - ", len(sports))
	for _, sport := range sports {
		cache.SetSport(sport)
	}
}

func RefreshSportStatusCache() {
	// 1. Get all latest updated 100 documents from SportStatus collection
	sportStatuses, err := database.GetUpdatedSportStatuss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshSportStatusCache: GetUpdatedSportStatuss failed with error - ", err.Error())
		return
	}
	if len(sportStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshSportStatusCache: GetUpdatedSportStatuss documents count - ", len(sportStatuses))
	for _, ss := range sportStatuses {
		cache.SetSportStatus(ss)
	}
}

func RefreshCompetitionCache() {
	// 1. Get all latest updated 100 documents from collection
	competitions, err := database.GetUpdatedCompetitionss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshCompetitionCache: GetUpdatedCompetitionss failed with error - ", err.Error())
		return
	}
	if len(competitions) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshCompetitionCache: GetUpdatedCompetitionss documents count - ", len(competitions))
	for _, competition := range competitions {
		cache.SetCompetition(competition)
	}
}

func RefreshCompetitionStatusCache() {
	// 1. Get all latest updated 100 documents from collection
	competitionStatuses, err := database.GetUpdatedCompetitionStatuss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshCompetitionStatusCache: GetUpdatedCompetitionStatuss failed with error - ", err.Error())
		return
	}
	if len(competitionStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshCompetitionStatusCache: GetUpdatedCompetitionStatuss documents count - ", len(competitionStatuses))
	for _, cs := range competitionStatuses {
		cache.SetCompetitionStatus(cs)
	}
}

func RefreshEventCache() {
	// 1. Get all latest updated 100 documents from collection
	events, err := database.GetUpdatedEventss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshEventCache: GetUpdatedEventss failed with error - ", err.Error())
		return
	}
	if len(events) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshEventCache: GetUpdatedEventss documents count - ", len(events))
	for _, event := range events {
		cache.SetEvent(event)
	}
}

func RefreshEventStatusCache() {
	// 1. Get all latest updated 100 documents from EventStatus collection
	eventStatuses, err := database.GetUpdatedEventStatuss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshEventStatusCache: GetUpdatedEventStatuss failed with error - ", err.Error())
		return
	}
	if len(eventStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshEventStatusCache: GetUpdatedEventStatuss documents count - ", len(eventStatuses))
	for _, ss := range eventStatuses {
		cache.SetEventStatus(ss)
	}
}

func RefreshMarketCache() {
	// 1. Get all latest updated 100 documents from collection
	markets, err := database.GetUpdatedMarketss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshMarketCache: database.GetUpdatedMarketss failed with error - ", err.Error())
		return
	}
	if len(markets) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshMarketCache: GetUpdatedMarketss documents count - ", len(markets))
	for _, market := range markets {
		cache.SetMarket(market)
	}
}

func RefreshMarketStatusCache() {
	// 1. Get all latest updated 100 documents from MarketStatus collection
	marketStatuses, err := database.GetUpdatedMarketStatuss()
	if err != nil {
		// 1.1. Error, return
		log.Println("RefreshMarketStatusCache: database.GetUpdatedMarketStatuss failed with error - ", err.Error())
		return
	}
	if len(marketStatuses) == 0 {
		// 1.2. No records found
		return
	}
	// 2. Update cache
	log.Println("RefreshMarketStatusCache: GetUpdatedMarketStatuss documents count - ", len(marketStatuses))
	for _, ms := range marketStatuses {
		cache.SetMarketStatus(ms)
	}
}

// Functions: Checking Status
// Provider Status
func IsProviderActive(operatorId string, partnerId string, providerId string) bool {
	// 1. Get Provider from Cache
	provider, err := cache.GetProvider(providerId)
	if err != nil {
		log.Println("IsProviderActive: GetProvider for "+providerId+" failed with error - ", err.Error())
		return false
	}
	// 2. Check Platform Level Status
	if constants.SAP.ObjectStatus.BLOCKED() == provider.Status {
		return false
	}
	// 3. Get ProviderStatus from Cache
	if partnerId == "" {
		return true
	}
	providerStatus, err := cache.GetPartnerStatusFromCache(operatorId, partnerId, providerId)
	if err != nil {
		log.Println("IsProviderActive: GetProviderStatus for "+providerId+" failed with error - ", err.Error())
		return false
	}
	// 4. Check ProviderStatus at Operator Level
	if constants.SAP.ObjectStatus.BLOCKED() == providerStatus.ProviderStatus {
		return false
	}
	// 5. Check OperatorStatus at Operator Level
	if constants.SAP.ObjectStatus.BLOCKED() == providerStatus.OperatorStatus {
		return false
	}
	return true
}

// Sport Status
func IsSportActive(operatorId string, partnerId string, providerId string, sportId string) bool {
	// 1. Get Sport from Cache
	sportKey := providerId + "-" + sportId
	sport, err := cache.GetSportFromCache(sportKey)
	if err != nil {
		log.Println("IsSportActive: GetSport for "+sportKey+" failed with error - ", err.Error())
		return false
	}
	// 2. Check Platform Level Status
	if constants.SAP.ObjectStatus.BLOCKED() == sport.Status {
		return false
	}
	// 3. Get SportStatus from Cache
	sportStatus, err := cache.GetSportStatus(operatorId, partnerId, providerId, sportId)
	if err != nil {
		ssKey := operatorId + "-" + sportKey
		log.Println("IsSportActive: GetSportStatus "+ssKey+" failed with error - ", err.Error())
		return false
	}
	// 4. Check ProviderStatus at Operator Level
	if constants.SAP.ObjectStatus.BLOCKED() == sportStatus.ProviderStatus {
		return false
	}
	// 5. Check OperatorStatus at Operator Level
	if constants.SAP.ObjectStatus.BLOCKED() == sportStatus.OperatorStatus {
		return false
	}
	return true
}

// Competition Status
func IsCompetitionActive(operatorId string, providerId string, sportId string, competitionId string) bool {
	// 1. Get Competition from Cache
	competitionKey := providerId + "-" + sportId + "-" + competitionId
	competition, err := cache.GetCompetition(providerId, sportId, competitionId)
	if err != nil {
		log.Println("IsCompetitionActive: GetCompetition for "+competitionKey+" failed with error - ", err.Error())
		return false
	}
	// 2. Check Platform Level Status
	if constants.SAP.ObjectStatus.BLOCKED() == competition.Status {
		log.Println("IsCompetitionActive: Competition BLOCKED for - ", competitionKey)
		return false
	}
	// 3. Get CompetitionStatus from Cache
	csKey := operatorId + "-" + competitionKey
	competitionStatus, err := cache.GetCompetitionStatus(operatorId, providerId, sportId, competitionId)
	if err == nil {
		// 4. Check ProviderStatus at Operator Level
		if constants.SAP.ObjectStatus.BLOCKED() == competitionStatus.ProviderStatus {
			log.Println("IsCompetitionActive: CompetitionStatus BLOCKED by Platform for - ", csKey)
			return false
		}
		// 5. Check OperatorStatus at Operator Level
		if constants.SAP.ObjectStatus.BLOCKED() == competitionStatus.OperatorStatus {
			log.Println("IsCompetitionActive: CompetitionStatus BLOCKED by Operator for - ", csKey)
			return false
		}
		// log.Println("IsCompetitionActive: GetCompetitionStatus for "+csKey+" failed with error - ", err.Error())
		// return false
	}
	return true
}

// Event Status
func IsEventActive(operatorId string, providerId string, sportId string, eventId string) bool {
	// 1. Get Event from Cache
	eventKey := providerId + "-" + sportId + "-" + eventId
	event, err := cache.GetEvent(providerId, sportId, eventId)
	if err != nil {
		log.Println("IsEventActive: GetEvent for "+eventKey+" failed with error - ", err.Error())
		return false
	}
	// 2. Check Platform Level Status
	if constants.SAP.ObjectStatus.BLOCKED() == event.Status {
		log.Println("IsEventActive: Event BLOCKED for - ", eventKey)
		return false
	}
	// 3. Get EventStatus from Cache
	esKey := operatorId + "-" + eventKey
	eventStatus, err := cache.GetEventStatus(operatorId, providerId, sportId, eventId)
	if err == nil {
		// 4. Check ProviderStatus at Operator Level
		if constants.SAP.ObjectStatus.BLOCKED() == eventStatus.ProviderStatus {
			log.Println("IsEventActive: EventStatus BLOCKED by Platform for - ", esKey)
			return false
		}
		// 5. Check OperatorStatus at Operator Level
		if constants.SAP.ObjectStatus.BLOCKED() == eventStatus.OperatorStatus {
			log.Println("IsEventActive: EventStatus BLOCKED by Operator for - ", esKey)
			return false
		}
		// log.Println("IsEventActive: GetEventStatus for "+esKey+" failed with error - ", err.Error())
		// return false
	}
	return true
}

func capatizeProviderId(providerId string) string {
	if strings.EqualFold(providerId, BETFAIR) {
		return BETFAIR
	}
	if strings.EqualFold(providerId, DREAM_SPORT) {
		return DREAM_SPORT
	}
	if strings.EqualFold(providerId, SPORT_RADAR) {
		return SPORT_RADAR
	}
	return providerId
}

// Get UserDto from DB
func GetB2BUser(operatorId string, userId string) models.B2BUserDto {
	userKey := operatorId + "-" + userId
	user, err := database.GetB2BUser(userKey)
	if err != nil {
		log.Println("GetB2BUser: database.GetB2BUser for "+userKey+" failed with error - ", err.Error())
	}
	return user
}

// Get OperatorDto from DB
func GetOperator(operatorId string) operatordto.OperatorDTO {
	operator, err := database.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("GetB2BUser: database.GetOperatorDetails for "+operatorId+" failed with error - ", err.Error())
	}
	return operator
}

type Config struct {
	MinBetValue int32
	MaxBetValue int32
	BetDelay    float64
	OddsLimit   int32
	Hold        float64
}

func GetOperatorConfig(marketType string, ms models.MarketStatus, es models.EventStatus, cs models.CompetitionStatus, ss models.SportStatus, ps models.PartnerStatus, op operatordto.OperatorDTO) (Config, string) {
	config := Config{}
	level := "NONE"
	// Get MinBetValue
	if ms.Config.IsSet {
		config.MinBetValue = GetMinBetValue(marketType, ms.Config)
	} else if es.Config.IsSet {
		config.MinBetValue = GetMinBetValue(marketType, es.Config)
	} else if cs.Config.IsSet {
		config.MinBetValue = GetMinBetValue(marketType, cs.Config)
	} else if ss.Config.IsSet {
		config.MinBetValue = GetMinBetValue(marketType, ss.Config)
	} else if ps.Config.IsSet {
		config.MinBetValue = GetMinBetValue(marketType, ps.Config)
		// } else if op.Config.IsSet {
		// 	config.MinBetValue = GetMinBetValue(marketType, op.Config)
	}
	// Get MaxBetValue
	if ms.Config.IsSet && GetMaxBetValue(marketType, ms.Config) > 0 {
		config.MaxBetValue = GetMaxBetValue(marketType, ms.Config)
	} else if es.Config.IsSet && GetMaxBetValue(marketType, es.Config) > 0 {
		config.MaxBetValue = GetMaxBetValue(marketType, es.Config)
	} else if cs.Config.IsSet && GetMaxBetValue(marketType, cs.Config) > 0 {
		config.MaxBetValue = GetMaxBetValue(marketType, cs.Config)
	} else if ss.Config.IsSet && GetMaxBetValue(marketType, ss.Config) > 0 {
		config.MaxBetValue = GetMaxBetValue(marketType, ss.Config)
	} else if ps.Config.IsSet && GetMaxBetValue(marketType, ps.Config) > 0 {
		config.MaxBetValue = GetMaxBetValue(marketType, ps.Config)
		// } else if op.Config.IsSet && GetMaxBetValue(marketType, op.Config) > 0 {
		// 	config.MaxBetValue = GetMaxBetValue(marketType, op.Config)
	}
	// Get BetDelay
	if ms.Config.IsSet {
		config.BetDelay = GetBetDelay(marketType, ms.Config)
	} else if es.Config.IsSet {
		config.BetDelay = GetBetDelay(marketType, es.Config)
	} else if cs.Config.IsSet {
		config.BetDelay = GetBetDelay(marketType, cs.Config)
	} else if ss.Config.IsSet {
		config.BetDelay = GetBetDelay(marketType, ss.Config)
	} else if ps.Config.IsSet {
		config.BetDelay = GetBetDelay(marketType, ps.Config)
		// } else if op.Config.IsSet {
		// 	config.BetDelay = GetBetDelay(marketType, op.Config)
	}
	// Get OddsLimit
	if ms.Config.IsSet && GetOddsLimit(marketType, ms.Config) > 0 {
		config.OddsLimit = GetOddsLimit(marketType, ms.Config)
	} else if es.Config.IsSet && GetOddsLimit(marketType, es.Config) > 0 {
		config.OddsLimit = GetOddsLimit(marketType, es.Config)
	} else if cs.Config.IsSet && GetOddsLimit(marketType, cs.Config) > 0 {
		config.OddsLimit = GetOddsLimit(marketType, cs.Config)
	} else if ss.Config.IsSet && GetOddsLimit(marketType, ss.Config) > 0 {
		config.OddsLimit = GetOddsLimit(marketType, ss.Config)
	} else if ps.Config.IsSet && GetOddsLimit(marketType, ps.Config) > 0 {
		config.OddsLimit = GetOddsLimit(marketType, ps.Config)
		// } else if op.Config.IsSet && GetOddsLimit(marketType, op.Config) > 0 {
		// 	config.OddsLimit = GetOddsLimit(marketType, op.Config)
	}
	// Get Hold
	if ms.Config.IsSet {
		config.Hold = ms.Config.Hold
		level = "MARKET"
		//log.Println("PlaceBet: OP Hold value from event config - ", es.EventId)
	} else if es.Config.IsSet {
		config.Hold = es.Config.Hold
		level = "EVENT"
		//log.Println("PlaceBet: OP Hold value from event config - ", es.EventId)
	} else if cs.Config.IsSet {
		config.Hold = cs.Config.Hold
		level = "COMPETITION"
		//log.Println("PlaceBet: OP Hold value from competition config - ", cs.CompetitionId)
	} else if ss.Config.IsSet {
		config.Hold = ss.Config.Hold
		level = "SPORT"
		//log.Println("PlaceBet: OP Hold value from sport config - ", ss.SportId)
	} else if ps.Config.IsSet {
		config.Hold = ps.Config.Hold
		level = "PROVIDER"
		//log.Println("PlaceBet: OP Hold value from partner config - ", ps.PartnerId)
		// } else if op.Config.IsSet {
		// 	config.Hold = op.Config.Hold
		// 	log.Println("PlaceBet: OP Hold value from operator config - ", op.OperatorId)
	} else {
		log.Println("PlaceBet: OP Hold value not configured in any level - ", op.OperatorId, ps.PartnerId, ss.SportId, cs.CompetitionId, es.EventId, marketType)
	}
	return config, level
}

func GetSapConfig(marketType string, ms models.Market, es models.Event, cs models.Competition, ss models.Sport, os operatordto.OperatorDTO, ps models.Provider) (Config, string) {
	config := Config{}
	config.Hold = 0
	config.MinBetValue = 0
	config.MaxBetValue = 100000 // Default 1L INR
	config.OddsLimit = 10
	config.BetDelay = 0
	level := "NONE"
	// Get MinBetValue
	if ms.Config.IsSet && GetMinBetValue(marketType, ms.Config) > 0 { // market
		config.MinBetValue = GetMinBetValue(marketType, ms.Config)
	} else if es.Config.IsSet && GetMinBetValue(marketType, es.Config) > 0 { // event
		config.MinBetValue = GetMinBetValue(marketType, es.Config)
	} else if cs.Config.IsSet && GetMinBetValue(marketType, cs.Config) > 0 { // competition
		config.MinBetValue = GetMinBetValue(marketType, cs.Config)
	} else if ss.Config.IsSet && GetMinBetValue(marketType, ss.Config) > 0 { // sport
		config.MinBetValue = GetMinBetValue(marketType, ss.Config)
	} else if os.Config.IsSet && GetMinBetValue(marketType, os.Config) > 0 { // operator
		config.MinBetValue = GetMinBetValue(marketType, os.Config)
	} else if ps.Config.IsSet && GetMinBetValue(marketType, ps.Config) > 0 { // provider
		config.MinBetValue = GetMinBetValue(marketType, ps.Config)
	}
	// Get MaxBetValue
	if ms.Config.IsSet && GetMaxBetValue(marketType, ms.Config) > 0 { // market
		config.MaxBetValue = GetMaxBetValue(marketType, ms.Config)
	} else if es.Config.IsSet && GetMaxBetValue(marketType, es.Config) > 0 { // event
		config.MaxBetValue = GetMaxBetValue(marketType, es.Config)
	} else if cs.Config.IsSet && GetMaxBetValue(marketType, cs.Config) > 0 { // competition
		config.MaxBetValue = GetMaxBetValue(marketType, cs.Config)
	} else if ss.Config.IsSet && GetMaxBetValue(marketType, ss.Config) > 0 { // sport
		config.MaxBetValue = GetMaxBetValue(marketType, ss.Config)
	} else if os.Config.IsSet && GetMaxBetValue(marketType, os.Config) > 0 { // operator
		config.MaxBetValue = GetMaxBetValue(marketType, os.Config)
	} else if ps.Config.IsSet && GetMaxBetValue(marketType, ps.Config) > 0 { // provider
		config.MaxBetValue = GetMaxBetValue(marketType, ps.Config)
	}
	// Get BetDelay
	if ms.Config.IsSet && GetBetDelay(marketType, ms.Config) > 0 { // market
		config.BetDelay = GetBetDelay(marketType, ms.Config)
	} else if es.Config.IsSet && GetBetDelay(marketType, es.Config) > 0 { // event
		config.BetDelay = GetBetDelay(marketType, es.Config)
	} else if cs.Config.IsSet && GetBetDelay(marketType, cs.Config) > 0 { // competition
		config.BetDelay = GetBetDelay(marketType, cs.Config)
	} else if ss.Config.IsSet && GetBetDelay(marketType, ss.Config) > 0 { // sport
		config.BetDelay = GetBetDelay(marketType, ss.Config)
	} else if os.Config.IsSet && GetBetDelay(marketType, os.Config) > 0 { // operator
		config.BetDelay = GetBetDelay(marketType, os.Config)
	} else if ps.Config.IsSet && GetBetDelay(marketType, ps.Config) > 0 { // provider
		config.BetDelay = GetBetDelay(marketType, ps.Config)
	}
	// Get OddsLimit
	if ms.Config.IsSet && GetOddsLimit(marketType, ms.Config) > 0 { // market
		config.OddsLimit = GetOddsLimit(marketType, ms.Config)
	} else if es.Config.IsSet && GetOddsLimit(marketType, es.Config) > 0 { // event
		config.OddsLimit = GetOddsLimit(marketType, es.Config)
	} else if cs.Config.IsSet && GetOddsLimit(marketType, cs.Config) > 0 { // competition
		config.OddsLimit = GetOddsLimit(marketType, cs.Config)
	} else if ss.Config.IsSet && GetOddsLimit(marketType, ss.Config) > 0 { // sport
		config.OddsLimit = GetOddsLimit(marketType, ss.Config)
	} else if os.Config.IsSet && GetOddsLimit(marketType, os.Config) > 0 { // operator
		config.OddsLimit = GetOddsLimit(marketType, os.Config)
	} else if ps.Config.IsSet && GetOddsLimit(marketType, ps.Config) > 0 { // provider
		config.OddsLimit = GetOddsLimit(marketType, ps.Config)
	}
	// Get Hold
	if ms.Config.IsSet {
		config.Hold = ms.Config.Hold
		level = "MARKET"
		log.Println("PlaceBet: SAP Hold value from event config - ", ms.MarketId)
	} else if es.Config.IsSet {
		config.Hold = es.Config.Hold
		level = "EVENT"
		log.Println("PlaceBet: SAP Hold value from event config - ", es.EventId)
	} else if cs.Config.IsSet {
		config.Hold = cs.Config.Hold
		level = "COMPETITION"
		log.Println("PlaceBet: SAP Hold value from competition config - ", cs.CompetitionId)
	} else if ss.Config.IsSet {
		config.Hold = ss.Config.Hold
		level = "SPORT"
		log.Println("PlaceBet: SAP Hold value from sport config - ", ss.SportId)
	} else if os.Config.IsSet {
		config.Hold = os.Config.Hold
		level = "OPERATOR"
		log.Println("PlaceBet: SAP Hold value from operator config - ", os.OperatorId)
	} else if ps.Config.IsSet {
		config.Hold = ps.Config.Hold
		level = "PROVIDER"
		log.Println("PlaceBet: SAP Hold value from provider config - ", ps.ProviderId)
	}
	return config, level
}

func GetMinBetValue(mt string, config commondto.ConfigDto) int32 {
	switch mt {
	case constants.SAP.MarketType.MATCH_ODDS():
		return config.MatchOdds.Min
	case constants.SAP.MarketType.BOOKMAKER():
		return config.Bookmaker.Min
	case constants.SAP.MarketType.FANCY():
		return config.Fancy.Min
	default:
		return 0
	}
}

func GetMaxBetValue(mt string, config commondto.ConfigDto) int32 {
	switch mt {
	case constants.SAP.MarketType.MATCH_ODDS():
		return config.MatchOdds.Max
	case constants.SAP.MarketType.BOOKMAKER():
		return config.Bookmaker.Max
	case constants.SAP.MarketType.FANCY():
		return config.Fancy.Max
	default:
		return 0
	}
}

func GetBetDelay(mt string, config commondto.ConfigDto) float64 {
	// BetDelay is stored in 1000 ms in DB so we need to convert it to seconds
	switch mt {
	case constants.SAP.MarketType.MATCH_ODDS():
		return config.MatchOdds.Delay / 1000
	case constants.SAP.MarketType.BOOKMAKER():
		return config.Bookmaker.Delay / 1000
	case constants.SAP.MarketType.FANCY():
		return config.Fancy.Delay / 1000
	default:
		return 0
	}
}

func GetOddsLimit(mt string, config commondto.ConfigDto) int32 {
	switch mt {
	case constants.SAP.MarketType.MATCH_ODDS():
		return config.MatchOdds.OddLimits
	case constants.SAP.MarketType.BOOKMAKER():
		return config.Bookmaker.OddLimits
	case constants.SAP.MarketType.FANCY():
		return config.Fancy.OddLimits
	default:
		return 0
	}
}

func GetCommissionConfig(userId, marketId, eventId, competitionId, sportId, partnerId, operatorId, providerId string) CommissionConfig {
	commissionConfig := CommissionConfig{}
	commissionConfig.CommPercentage = 2
	commissionConfig.CommLevel = "DEFAULT"
	// 1. User Level - SKIP for now. TODO: Implement Cache first.

	// 2. Market Level
	market, err := cache.GetMarketStatus(operatorId, providerId, sportId, eventId, marketId)
	if err != nil {
		// TODO: Log
	}
	if market.IsCommission == true {
		if market.WinCommission > 2 {
			commissionConfig.CommPercentage = market.WinCommission
		}
		commissionConfig.CommLevel = "MARKET"
		return commissionConfig
	}
	// 3. Event Level
	event, err := cache.GetEventStatus(operatorId, providerId, sportId, eventId)
	if err != nil {
		// TODO: Log
	}
	if event.IsCommission == true {
		if event.WinCommission > 2 {
			commissionConfig.CommPercentage = event.WinCommission
		}
		commissionConfig.CommLevel = "EVENT"
		return commissionConfig
	}
	// 4. Competition Level
	competition, err := cache.GetCompetitionStatus(operatorId, providerId, sportId, competitionId)
	if err != nil {
		// TODO: Log
	}
	if competition.IsCommission == true {
		if competition.WinCommission > 2 {
			commissionConfig.CommPercentage = competition.WinCommission
		}
		commissionConfig.CommLevel = "COMPETITION"
		return commissionConfig
	}
	// 5. Sport Level
	sport, err := cache.GetSportStatus(operatorId, "", providerId, sportId)
	if err != nil {
		// TODO: Log
	}
	if sport.IsCommission == true {
		if sport.WinCommission > 2 {
			commissionConfig.CommPercentage = sport.WinCommission
		}
		commissionConfig.CommLevel = "SPORT"
		return commissionConfig
	}
	// 6. Partner Level
	partner, err := cache.GetPartnerStatus(operatorId, partnerId, providerId)
	if err != nil {
		// TODO: Log
	}
	if partner.IsCommission == true {
		if partner.WinCommission > 2 {
			commissionConfig.CommPercentage = partner.WinCommission
		}
		commissionConfig.CommLevel = "PARTNER"
		return commissionConfig
	}
	// 7. Operator Level
	// operator, err := cache.GetOperatorDetails(operatorId)
	// if err != nil {
	// 	// TODO: Log
	// }
	// if operator.IsCommission == true {
	// 	if operator.WinCommission > 2 {
	// 		commissionConfig.CommPercentage = operator.WinCommission
	// 	}
	// 	commissionConfig.CommLevel = "OPERATOR"
	// 	return commissionConfig
	// }
	return commissionConfig
}
