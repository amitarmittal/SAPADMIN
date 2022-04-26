package portalsvc

import (
	mainCache "Sp/cache"
	cache "Sp/cache/operator"
	"Sp/database"
	"Sp/dto/commondto"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	portaldto "Sp/dto/portal"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	OperatorHttpReqTimeout time.Duration = 5
	ErrorUserIdMissing     string        = "Bad request - UserId is missing!"
	ErrorOperatorIdMissing string        = "Bad request - OperatorId is missing!"
	ErrorProviderIdMissing string        = "Bad request - ProviderId is missing!"
	ErrorSportIdMissing    string        = "Bad request - SportId is missing!"
	ErrorTxIdMissing       string        = "Bad request - TransactionId is missing!"
	ErrorStartDateMissing  string        = "Bad request - StartDate is missing!"
	ErrorEndDateMissing    string        = "Bad request - EndDate is missing!"
	ErrorInvalidRange      string        = "Bad request - Filter value out of range!"
	DREAM_SPORT                          = "Dream"
	BETFAIR                              = "BetFair"
	SPORT_RADAR                          = "SportRadar"
)

var (
	// ouAPIList contains all the API whitelisted for Operator-User
	ouAPIList []string = []string{"getbets", "getbet", "userstatement", "getusers", "blockuser", "unblockuser", "getoaproviders",
		"blockoaprovider", "unblockoaprovider", "listusers", "listsports", "listevents", "getProviders", "getProviderStatus",
		"updateOperators", "getSportsListForOP", "getCompetitionsListForOP", "getEventsListForOP", "blockSportsForOP",
		"unblockSportsForOP", "blockCompetitionForOP", "unblockCompetitionForOP", "blockEventForOP", "unblockEventForOP", "getBetsForOP",
		"/api/v1/portal/opadmin/get-operator-details", "/api/v1/portal/opadmin/get-config",
		"/api/v1/portal/opadmin/reset-password", "/api/v1/portal/opadmin/get-markets", "/api/v1/portal/opadmin/get-balance",
		"/api/v1/portal/opadmin/get-partner-ids", "/api/v1/reports/get-bet-list", "/api/v1/reports/get-user-statement",
		"/api/v1/reports/get-admin-statement", "/api/v1/reports/get-game-report", "/api/v1/reports/get-pnl-report", "/api/v1/reports/get-provider-pnl-report",
		"/api/v1/reports/get-sport-pnl-report", "/api/v1/reports/get-competition-pnl-report", "/api/v1/reports/get-event-pnl-report",
		"/api/v1/reports/get-my-account-statement", "/api/v1/reports/get-user-audit-report",
		"/api/v1/portal/opadmin/block-market",
		"/api/v1/portal/opadmin/unblock-market",
		"/api/v1/portal/opadmin/recent-compititions",
		"/api/v1/reports/get-sport-report",
		"/api/v1/reports/get-bet-detail-report",
		"/api/v1/portal/sapadmin/delete-bets",
		"/api/v1/reports/get-operator-risk-report",
		"/api/v1/reports/get-user-book-report",
		"/api/v1/portal/opadmin/get-open-bets",
		"/api/v1/reports/get-transfer-user-statement",
	}
	// oaAPIList contains all the API whitlisted for Operator-Admin
	oaAPIList []string = []string{"getbets", "getbet", "userstatement", "getusers", "blockuser", "unblockuser", "getoaproviders",
		"blockoaprovider", "unblockoaprovider", "listusers", "listsports", "listevents", "getProviders", "getProviderStatus",
		"updateOperators", "getSportsListForOP", "getCompetitionsListForOP", "getEventsListForOP", "blockSportsForOP",
		"unblockSportsForOP", "blockCompetitionForOP", "unblockCompetitionForOP", "blockEventForOP", "unblockEventForOP", "getBetsForOP",
		"/api/v1/portal/opadmin/get-operator-details", "/api/v1/portal/opadmin/get-config", "/api/v1/portal/opadmin/set-config",
		"/api/v1/portal/opadmin/reset-password", "/api/v1/portal/opadmin/get-markets", "/api/v1/portal/opadmin/get-balance",
		"/api/v1/portal/opadmin/get-partner-ids", "/api/v1/reports/get-bet-list", "/api/v1/reports/get-user-statement",
		"/api/v1/reports/get-admin-statement", "/api/v1/reports/get-game-report", "/api/v1/reports/get-pnl-report", "/api/v1/reports/get-provider-pnl-report",
		"/api/v1/reports/get-sport-pnl-report", "/api/v1/reports/get-competition-pnl-report", "/api/v1/reports/get-event-pnl-report",
		"/api/v1/reports/get-my-account-statement", "/api/v1/reports/get-user-audit-report",
		"/api/v1/portal/opadmin/block-market",
		"/api/v1/portal/opadmin/unblock-market",
		"/api/v1/portal/opadmin/recent-compititions",
		"/api/v1/reports/get-sport-report",
		"/api/v1/reports/get-bet-detail-report",
		"/api/v1/portal/sapAdmin/delete-bets",
		"/api/v1/reports/get-operator-risk-report",
		"/api/v1/reports/get-user-book-report",
		"/api/v1/portal/opadmin/get-open-bets",
		"/api/v1/reports/get-transfer-user-statement",
	}

	// saAPIList contains all the API whitlisted for SP-Admin
	spAPIList []string = []string{"createuser", "getoperators", "getpartners", "blockoperator", "unblockoperator", "blockpartner", "unblockpartner", "getsaproviders",
		"blocksaprovider", "unblocksaprovider", "getProviders", "getProviderStatus", "updateProviderStatus", "listevents",
		"updateOperators", "blockoaprovider", "unblockoaprovider", "getSportsListForSAP", "getSportsListForOP",
		"getCompetitionsListForSAP", "getEventsListForSAP", "blockSportsForSAP", "unblockSportsForSAP", "blockSportsForSAP",
		"unblockSportsForSAP", "blockCompetitionForSAP", "unblockCompetitionForSAP", "blockEventForSAP", "unblockEventForSAP",
		"getBetsForSAP", "getUsersForSAP", "userstatement",
		"/api/v1/portal/sapadmin/user-statement", "/api/v1/portal/sapadmin/operator-details", "/api/v1/portal/sapadmin/operator-status-block",
		"/api/v1/portal/sapadmin/get-settled-bets", "/api/v1/portal/sapadmin/get-lapsed-bets", "/api/v1/portal/sapadmin/get-cancelled-bets",
		"/api/v1/portal/sapadmin/operator-status-unblock", "/api/v1/portal/sapadmin/provider-operator-status", "/api/v1/portal/sapadmin/sport-operator-status",
		"/api/v1/portal/sapadmin/competition-operator-status", "/api/v1/portal/sapadmin/event-operator-status",
		"/api/v1/portal/sapadmin/get-config", "/api/v1/portal/sapadmin/set-config", "/api/v1/portal/sapadmin/sync-sports",
		"/api/v1/portal/sapadmin/sync-competitions", "/api/v1/portal/sapadmin/sync-events", "/api/v1/portal/sapadmin/get-portal-user",
		"/api/v1/portal/sapadmin/reset-password", "/api/v1/portal/sapadmin/reset-op-password", "/api/v1/portal/sapadmin/add-partner",
		"/api/v1/portal/sapadmin/get-markets", "/api/v1/portal/sapadmin/block-portal-user",
		"/api/v1/portal/sapadmin/unblock-portal-user", "/api/v1/portal/sapadmin/replace-partner", "/api/v1/portal/sapadmin/deposit-funds", "/api/v1/portal/sapadmin/withdraw-funds",
		"/api/v1/portal/sapadmin/get-partner-ids", "/api/v1/portal/sapadmin/matched-sports", "/api/v1/portal/sapadmin/unmatched-sports",
		"/api/v1/portal/sapadmin/all-sports", "/api/v1/portal/sapadmin/update-sport-card", "/api/v1/portal/sapadmin/get-sport-radar-cards",
		"/api/v1/portal/sapadmin/get-event", "/api/v1/reports/get-bet-list", "/api/v1/reports/get-user-statement",
		"/api/v1/reports/get-admin-statement", "/api/v1/reports/get-game-report", "/api/v1/reports/get-pnl-report", "/api/v1/reports/get-provider-pnl-report",
		"/api/v1/reports/get-sport-pnl-report", "/api/v1/reports/get-competition-pnl-report", "/api/v1/reports/get-event-pnl-report",
		"/api/v1/reports/get-my-account-statement", "/api/v1/reports/get-user-audit-report",
		"/api/v1/portal/sapadmin/block-market",
		"/api/v1/portal/sapadmin/unblock-market",
		"/api/v1/portal/sapadmin/suspend-market",
		"/api/v1/portal/sapadmin/resume-market",
		"/api/v1/portal/sapadmin/get-op-markets",
		"/api/v1/portal/sapadmin/block-op-market",
		"/api/v1/portal/sapadmin/unblock-op-market",
		"/api/v1/portal/sapadmin/recent-compititions",
		"/api/v1/portal/sapadmin/view-fund-statement",
		"/api/v1/portal/sapadmin/get-role",
		"/api/v1/reports/get-sport-report",
		"/api/v1/reports/get-bet-detail-report",
		"/api/v1/portal/sapAdmin/delete-bets",
		"/api/v1/reports/get-operator-risk-report",
		"/api/v1/reports/get-user-book-report",
		"/api/v1/portal/sapadmin/get-open-bets",
		"/api/v1/reports/get-transfer-user-statement",
	}

	// suAPIList contains all the API whitlisted for SP-User
	suAPIList []string = []string{"createuser", "getoperators", "getpartners", "blockoperator", "unblockoperator", "blockpartner", "unblockpartner", "getsaproviders",
		"blocksaprovider", "unblocksaprovider", "getProviders", "getProviderStatus", "updateProviderStatus", "listevents",
		"updateOperators", "blockoaprovider", "unblockoaprovider", "getSportsListForSAP", "getSportsListForOP",
		"getCompetitionsListForSAP", "getEventsListForSAP", "blockSportsForSAP", "unblockSportsForSAP", "blockSportsForSAP",
		"unblockSportsForSAP", "blockCompetitionForSAP", "unblockCompetitionForSAP", "blockEventForSAP", "unblockEventForSAP",
		"getBetsForSAP", "getUsersForSAP", "userstatement",
		"/api/v1/portal/sapadmin/user-statement", "/api/v1/portal/sapadmin/operator-details", "/api/v1/portal/sapadmin/operator-status-block",
		"/api/v1/portal/sapadmin/get-settled-bets", "/api/v1/portal/sapadmin/get-lapsed-bets", "/api/v1/portal/sapadmin/get-cancelled-bets",
		"/api/v1/portal/sapadmin/operator-status-unblock", "/api/v1/portal/sapadmin/provider-operator-status", "/api/v1/portal/sapadmin/sport-operator-status",
		"/api/v1/portal/sapadmin/competition-operator-status", "/api/v1/portal/sapadmin/event-operator-status",
		"/api/v1/portal/sapadmin/get-config", "/api/v1/portal/sapadmin/sync-sports",
		"/api/v1/portal/sapadmin/sync-competitions", "/api/v1/portal/sapadmin/sync-events", "/api/v1/portal/sapadmin/get-portal-user",
		"/api/v1/portal/sapadmin/reset-password", "/api/v1/portal/sapadmin/reset-op-password", "/api/v1/portal/sapadmin/add-partner",
		"/api/v1/portal/sapadmin/get-markets", "/api/v1/portal/sapadmin/block-portal-user",
		"/api/v1/portal/sapadmin/unblock-portal-user", "/api/v1/portal/sapadmin/replace-partner", "/api/v1/portal/sapadmin/deposit-funds", "/api/v1/portal/sapadmin/withdraw-funds",
		"/api/v1/portal/sapadmin/get-partner-ids", "/api/v1/portal/sapadmin/matched-sports", "/api/v1/portal/sapadmin/unmatched-sports",
		"/api/v1/portal/sapadmin/all-sports", "/api/v1/portal/sapadmin/update-sport-card", "/api/v1/portal/sapadmin/get-sport-radar-cards",
		"/api/v1/portal/sapadmin/get-event", "/api/v1/reports/get-bet-list", "/api/v1/reports/get-user-statement",
		"/api/v1/reports/get-admin-statement", "/api/v1/reports/get-game-report", "/api/v1/reports/get-pnl-report", "/api/v1/reports/get-provider-pnl-report",
		"/api/v1/reports/get-sport-pnl-report", "/api/v1/reports/get-competition-pnl-report", "/api/v1/reports/get-event-pnl-report",
		"/api/v1/reports/get-my-account-statement", "/api/v1/reports/get-user-audit-report",
		"/api/v1/portal/sapadmin/block-market",
		"/api/v1/portal/sapadmin/unblock-market",
		"/api/v1/portal/sapadmin/suspend-market",
		"/api/v1/portal/sapadmin/resume-market",
		"/api/v1/portal/sapadmin/get-op-markets",
		"/api/v1/portal/sapadmin/block-op-market",
		"/api/v1/portal/sapadmin/unblock-op-market",
		"/api/v1/portal/sapadmin/recent-compititions",
		"/api/v1/portal/sapadmin/view-fund-statement",
		"/api/v1/portal/sapadmin/get-role",
		"/api/v1/reports/get-sport-report",
		"/api/v1/reports/get-bet-detail-report",
		"/api/v1/portal/sapAdmin/delete-bets",
		"/api/v1/reports/get-operator-risk-report",
		"/api/v1/reports/get-user-book-report",
		"/api/v1/portal/sapadmin/get-open-bets",
		"/api/v1/reports/get-transfer-user-statement",
	}

	// remainingAPILists contains all the Remaining APIs which will have Auth token but will not requied any role checking.
	remainingAPIList []string = []string{"logout"}
)

func Pagination(page int, number int, total int) (int, int) {
	start := 0
	end := 0
	if page == 0 && number == 0 {
		return 1, total
	}
	if total <= number {
		return 1, total
	}
	if total >= page*number {
		start = ((page * number) - number) + 1
		end = page * number
	} else {
		start = (total - int(math.Mod(float64(total), float64(number)))) + 1
		end = total
	}
	return start, end
}

func CheckPasswordHash(password, hash string) bool {
	cacheHash, found := mainCache.GetPasswordHash(password)
	if found {
		if hash == cacheHash {
			return true
		}
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		mainCache.SetPasswordHash(password, hash)
	}
	return err == nil
}

func jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return []byte("SECRET"), nil
}

type TokenMetadata struct {
	Username   string
	UserId     string
	Role       string
	OperatorId string
	IP         string
	Expires    int64
}

func Authenticate(c *fiber.Ctx) (*TokenMetadata, bool) {
	Tknmeta := new(TokenMetadata)
	bearerToken := c.Get("Authorization")
	tokenString := ""
	log.Println("Bearer Token: ", bearerToken)
	// Normally Authorization HTTP header.
	onlyToken := strings.Split(bearerToken, " ")
	if len(onlyToken) == 2 {
		tokenString = onlyToken[1]
	} else {
		log.Println("Failed in splitting the Token")
		return Tknmeta, false
	}
	token, err := jwt.Parse(tokenString, jwtKeyFunc)
	if err != nil {
		log.Println("Failed in convert string to Jwt token")
		return Tknmeta, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		Tknmeta.OperatorId = claims["operatorId"].(string)
		Tknmeta.Role = string(claims["role"].(string))
		Tknmeta.Username = string(claims["username"].(string))
		Tknmeta.UserId = string(claims["userId"].(string))
		Tknmeta.Expires = int64(claims["exp"].(float64))
		// Tknmeta.IP = string(claims["ip"].(string))
		session, err := cache.GetPortalSessionDetails(Tknmeta.Username, tokenString)

		if err == nil {
			if session.UserId == Tknmeta.Username && session.OperatorId == Tknmeta.OperatorId {
				return Tknmeta, true
			}
			log.Println("session: ", session)
			log.Println("Tknmeta: ", Tknmeta)
		}
		log.Println(err)
		log.Println("Error in GetPortalSessionDetails.")
	} else {
		log.Println("Failed in Claims, ok", claims)
		return Tknmeta, false
	}
	return Tknmeta, false
}

func CheckGetBet(req portaldto.GetBetReqDto) error {
	// 1. OperatorId must have
	if req.OperatorId == "" {
		return fmt.Errorf(ErrorOperatorIdMissing)
	}
	// 2. ProviderId need if SportId provided
	if req.TxId == "" {
		return fmt.Errorf(ErrorTxIdMissing)
	}
	return nil
}

func CheckUserStatement(req portaldto.UserStatementReqDto) error {
	// 1. OperatorId must have
	if req.OperatorId == "" {
		return fmt.Errorf(ErrorOperatorIdMissing)
	}
	// 2. ProviderId need if SportId provided
	if req.UserId == "" {
		return fmt.Errorf(ErrorUserIdMissing)
	}
	// 3. EndDate needed if StartDate provided
	if req.StartDate != 0 && req.EndDate == 0 {
		return fmt.Errorf(ErrorEndDateMissing)
	}
	// 4. StartDate needed if EndDate provided
	if req.EndDate != 0 && req.StartDate == 0 {
		return fmt.Errorf(ErrorStartDateMissing)
	}
	// 5. -ve values check
	if req.Page < 0 || req.PageSize < 0 || req.StartDate < 0 || req.EndDate < 0 {
		return fmt.Errorf(ErrorInvalidRange)
	}
	return nil
}

func CheckAllUserStatement(req portaldto.UserStatementReqDto) error {

	// 1. ProviderId need if SportId provided
	if req.UserId == "" {
		return fmt.Errorf(ErrorUserIdMissing)
	}
	// 2. EndDate needed if StartDate provided
	if req.StartDate != 0 && req.EndDate == 0 {
		return fmt.Errorf(ErrorEndDateMissing)
	}
	// 3. StartDate needed if EndDate provided
	if req.EndDate != 0 && req.StartDate == 0 {
		return fmt.Errorf(ErrorStartDateMissing)
	}
	// 4. -ve values check
	if req.Page < 0 || req.PageSize < 0 || req.StartDate < 0 || req.EndDate < 0 {
		return fmt.Errorf(ErrorInvalidRange)
	}
	return nil
}

//TODO: Optimize the code below to check role based operations.
func IsApplicable(Tknmeta *TokenMetadata, apiName string) bool {
	role := Tknmeta.Role

	// sends true if the apiName is in remainingAPIList
	for _, api := range remainingAPIList {
		if api == apiName {
			return true
		}
	}

	// sends true if the apiName is in suAPIList
	if role == "SAPUser" {
		for _, api := range suAPIList {
			if api == apiName {
				return true
			}
		}
	}

	// sends true if the apiName is in ouAPIList
	if role == "OperatorUser" {
		for _, api := range ouAPIList {
			if api == apiName {
				return true
			}
		}
	}

	// sends true if the apiName is in oaAPIList
	if role == "OperatorAdmin" {
		for _, api := range oaAPIList {
			if api == apiName {
				return true
			}
		}
	}

	// sends true if the apiName is in spAPIList
	if role == "SAPAdmin" {
		for _, api := range spAPIList {
			if api == apiName {
				return true
			}
		}
	}

	// Returns false if no match found
	return false
}

func GetUserDto(b2bUser models.B2BUserDto) portaldto.UserDto {
	user := portaldto.UserDto{}
	user.UserId = b2bUser.UserId
	user.UserName = b2bUser.UserName
	return user
}

func GetUser(b2bUser models.B2BUserDto) portaldto.OperatorUser {
	user := portaldto.OperatorUser{}
	user.UserId = b2bUser.UserId
	user.UserName = b2bUser.UserName
	user.Balance = b2bUser.Balance
	user.Status = b2bUser.Status
	return user
}

func GetProvider(partnerStatus models.PartnerStatus, operator operatordto.OperatorDTO) portaldto.Provider {
	provider := portaldto.Provider{}
	provider.PartnerId = partnerStatus.PartnerId
	provider.ProviderId = partnerStatus.ProviderId
	provider.ProviderName = partnerStatus.ProviderName
	provider.Status = partnerStatus.OperatorStatus
	for _, partner := range operator.Partners {
		if partner.PartnerId == partnerStatus.PartnerId {
			provider.Currency = partner.Currency
			provider.Rate = partner.Rate
			break
		}
	}
	return provider
}

func GetSportStatus(sportStatus models.SportStatus) portaldto.Sport {
	Sport := portaldto.Sport{}
	Sport.SportId = sportStatus.SportId
	Sport.SportName = sportStatus.SportName
	Sport.Status = sportStatus.OperatorStatus
	Sport.PartnerId = sportStatus.PartnerId
	return Sport
}

func GetCompetitionsStatus(competitionStatus models.CompetitionStatus) portaldto.Competition {
	competition := portaldto.Competition{}
	competition.CompetitionId = competitionStatus.CompetitionId
	competition.CompetitionName = competitionStatus.CompetitionName
	competition.Status = competitionStatus.OperatorStatus
	return competition
}

func GetCompetitionsStatus2(competition1 models.Competition) portaldto.Competition {
	competition := portaldto.Competition{}
	competition.CompetitionId = competition1.CompetitionId
	competition.CompetitionName = competition1.CompetitionName
	competition.Status = competition1.Status
	return competition
}

func GetEventsStatus(eventStatus models.EventStatus) portaldto.Event {
	event := portaldto.Event{}
	event.EventId = eventStatus.EventId
	event.EventName = eventStatus.EventName
	event.CompetitionId = eventStatus.CompetitionId
	event.CompetitionName = eventStatus.CompetitionName
	event.Status = eventStatus.OperatorStatus
	eventKey := eventStatus.ProviderId + "-" + eventStatus.SportId + "-" + eventStatus.EventId
	eventForOpenDate, err := mainCache.GetEvent(eventStatus.ProviderId, eventStatus.SportId, eventStatus.EventId)
	if err != nil {
		log.Println("GetEventsListForOP: Get Event List failed with error - ", err.Error())
		log.Println("GetEventsListForOP: openDate for eventKey - ", eventKey, " will be empty")
		event.OpenDate = 0
	} else {
		event.OpenDate = eventForOpenDate.OpenDate
	}
	return event
}

func GetEventsStatus2(event1 models.Event) portaldto.Event {
	event := portaldto.Event{}
	event.EventId = event1.EventId
	event.EventName = event1.EventName
	event.CompetitionId = event1.CompetitionId
	event.CompetitionName = event1.CompetitionName
	event.Status = event1.Status
	event.OpenDate = event1.OpenDate
	// eventKey := event1.ProviderId + "-" + event1.SportId + "-" + event1.EventId
	// eventForOpenDate, err := mainCache.GetEvent(event1.ProviderId, event1.SportId, event1.EventId)
	// if err != nil {
	// 	log.Println("GetEventsListForOP: Get Event List failed with error - ", err.Error())
	// 	log.Println("GetEventsListForOP: openDate for eventKey - ", eventKey, " will be empty")
	// 	event.OpenDate = 0
	// } else {
	// 	event.OpenDate = eventForOpenDate.OpenDate
	// }
	return event
}

func GetSAProvider(saProvider models.Provider) portaldto.Provider {
	provider := portaldto.Provider{}
	provider.ProviderId = saProvider.ProviderId
	provider.ProviderName = saProvider.ProviderName
	provider.Status = saProvider.Status
	return provider
}

func GetOperator(opDto operatordto.OperatorDTO) portaldto.Operator {
	operator := portaldto.Operator{}
	operator.OperatorId = opDto.OperatorId
	operator.OperatorName = opDto.OperatorName
	operator.BaseURL = opDto.BaseURL
	operator.Balance = opDto.Balance
	operator.Currency = "PTS"
	if len(opDto.Partners) > 0 {
		operator.Balance = operator.Balance / float64(opDto.Partners[0].Rate)
		operator.Currency = opDto.Partners[0].Currency
	}
	operator.Status = opDto.Status
	operator.WalletType = opDto.WalletType
	operator.OperatorKey = opDto.Keys.OperatorKey
	operator.PublicKey = opDto.Keys.PublicKey
	operator.Config = opDto.Config
	return operator
}

func GetPartner(opDto operatordto.OperatorDTO, partnerdto operatordto.Partner) portaldto.Partner {
	partner := portaldto.Partner{}
	partner.OperatorId = opDto.OperatorId
	partner.OperatorName = opDto.OperatorName
	partner.PartnerId = partnerdto.PartnerId
	partner.Currency = partnerdto.Currency
	partner.Rate = partnerdto.Rate
	partner.Commission = partnerdto.Commisssion
	partner.Status = partnerdto.Status
	partner.WalletType = opDto.WalletType
	return partner
}

func GetEvent(eventDto models.Event) portaldto.Event {
	event := portaldto.Event{}
	event.EventId = eventDto.EventId
	event.EventName = eventDto.EventName
	event.OpenDate = eventDto.OpenDate
	return event
}

func GetTransaction(ul models.UserLedgerDto) portaldto.UserTransaction {
	ut := portaldto.UserTransaction{}
	ut.TxTime = ul.TransactionTime
	ut.TxType = ul.TransactionType
	ut.RefId = ul.ReferenceId
	ut.Amount = ul.Amount
	ut.Remark = ul.Remark
	ut.CompetitionName = ul.CompetitionName
	ut.EventName = ul.EventName
	ut.MarketType = ul.MarketType
	ut.MarketName = ul.MarketName
	return ut
}

func updateOperatorDTOFromRequest(operatorDTO *operatordto.OperatorDTO, reqDto operatordto.OperatorDTO) {
	if reqDto.OperatorId != "" {
		operatorDTO.OperatorId = reqDto.OperatorId
	}
	if reqDto.OperatorName != "" {
		operatorDTO.OperatorName = reqDto.OperatorName
	}
	if reqDto.BaseURL != "" {
		operatorDTO.BaseURL = reqDto.BaseURL
	}
	if reqDto.Status != "" {
		operatorDTO.Status = reqDto.Status
	}
	if reqDto.Keys.OperatorKey != "" {
		operatorDTO.Keys.OperatorKey = reqDto.Keys.OperatorKey
	}
	if reqDto.Keys.PrivateKey != "" {
		operatorDTO.Keys.PrivateKey = reqDto.Keys.PrivateKey
	}
	if reqDto.Keys.PublicKey != "" {
		operatorDTO.Keys.PublicKey = reqDto.Keys.PublicKey
	}
	if reqDto.Keys.PublicKey != "" {
		operatorDTO.Keys.PublicKey = reqDto.Keys.PublicKey
	}
	if reqDto.WalletType != "" {
		operatorDTO.WalletType = reqDto.WalletType
	}
	if reqDto.Config.IsSet == true {
		operatorDTO.Config = reqDto.Config
	}
}

//Function Getting Config -- OP
//Opertor Config
// func getOperatorConfigOP(operatorId string) (commondto.ConfigDto, string, error) {
// 	lvl := "operator"
// 	// Get Operator Dto.
// 	opDTO, err := mainCache.GetOperatorDetails(operatorId)
// 	if err != nil {
// 		log.Println("getOperatorConfigOP: Get Operator Details failed with error - ", err.Error())
// 		return DefaultConfig(), lvl, err
// 	}
// 	// Default Hold will always be -1.0
// 	if opDTO.Config.IsSet {
// 		opDTO.Config = getConfigDelay(opDTO.Config)
// 		return opDTO.Config, lvl, nil
// 	}
// 	return DefaultConfig(), lvl, nil
// }

func getPartnerConfigOP(operatorId string, partnerId string, providerId string) (commondto.ConfigDto, string, error) {
	lvl := "partner"
	// Get Partner Dto.
	partnerStatusKey := operatorId + "-" + partnerId + "-" + providerId
	partnerDTO, err := database.GetPartnerStatus(partnerStatusKey)
	// partnerDTO, err := mainCache.GetPartnerStatus(operatorId, partnerId, providerId)
	if err != nil {
		log.Println("getPartnerConfigOP: Get Partner Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return partnerDTO.Config, lvl, nil
	// Default Hold will always be -1.0
	// if partnerDTO.Config.IsSet {
	// 	partnerDTO.Config = getConfigDelay(partnerDTO.Config)
	// 	return partnerDTO.Config, lvl, nil
	// }
	// return DefaultConfig(), lvl, nil
}

// Sport Config
func getSportConfigOP(operatorId string, providerId string, sportId string) (commondto.ConfigDto, string, error) {
	lvl := "sport"
	// Get Sport Dto.
	sportKey := operatorId + "-" + providerId + "-" + sportId
	sportDTO, err := database.GetSportStatus(sportKey)
	// sportDTO, err := mainCache.GetSportStatus(operatorId, providerId, sportId)
	if err != nil {
		log.Println("getSportConfigOP: Get Sport Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return sportDTO.Config, lvl, nil
	// Default Hold will always be -1.0
	// if sportDTO.Config.IsSet {
	// 	sportDTO.Config = getConfigDelay(sportDTO.Config)
	// 	return sportDTO.Config, lvl, nil
	// Get Operator Config
	// opConfig, lvl, err := getPartnerConfigOP(operatorId, partnerId, providerId)
	// if err != nil {
	// 	log.Println("getSportConfigOP: Get Operator Details failed with error - ", err.Error())
	// 	return DefaultConfig(), lvl, err
	// }
	// return DefaultConfig(), lvl, nil
}

// Competition Config
func getCompetitionConfigOP(operatorId string, providerId string, sportId string, competitionId string) (commondto.ConfigDto, string, error) {
	lvl := "competition"
	// Get Competition Dto.
	competitionKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId
	competitionDTO, err := database.GetCompetitionStatus(competitionKey)
	// competitionDTO, err := mainCache.GetCompetitionStatus(operatorId, providerId, sportId, competitionId)
	if err != nil {
		// log.Println("getCompetitionConfigOP: Get Competition Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return competitionDTO.Config, lvl, nil
	// // Default Hold will always be -1.0
	// if competitionDTO.Config.IsSet {
	// 	competitionDTO.Config = getConfigDelay(competitionDTO.Config)
	// 	return competitionDTO.Config, lvl, nil
	// }

	// // Get Sport Config
	// sportConfig, lvl, err := getSportConfigOP(operatorId, providerId, sportId)
	// if err != nil {
	// 	log.Println("getCompetitionConfigOP: Get Sport Details failed with error - ", err.Error())
	// 	return DefaultConfig(), lvl, err
	// }
	// // Default Hold will always be -1.0
	// if sportConfig.IsSet {
	// 	return sportConfig, lvl, nil
	// }
	// return DefaultConfig(), lvl, nil
}

// Event Config
func getEventConfigOP(operatorId string, providerId string, sportId string, competitionId string, eventId string) (commondto.ConfigDto, string, error) {
	lvl := "event"
	// Get Event Dto.
	eventKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
	eventDTO, err := database.GetEventStatus(eventKey)
	// eventDTO, err := mainCache.GetEventStatus(operatorId, providerId, sportId, eventId)
	if err != nil {
		//log.Println("getEventConfigOP: Get Event Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return eventDTO.Config, lvl, nil
	// // Default Hold will always be -1.0
	// if eventDTO.Config.IsSet {
	// 	eventDTO.Config = getConfigDelay(eventDTO.Config)
	// 	eventJson, _ := json.Marshal(eventDTO)
	// 	log.Println("getEventConfigOP: eventJson - ", string(eventJson))
	// 	return eventDTO.Config, lvl, nil
	// }

	// // Get Competition Config
	// competitionConfig, lvl, err := getCompetitionConfigOP(operatorId, providerId, sportId, competitionId)
	// if err != nil {
	// 	log.Println("getEventConfigOP: Get Competition Details failed with error - ", err.Error())
	// 	return DefaultConfig(), lvl, err
	// }
	// // Default Hold will always be -1.0
	// if competitionConfig.IsSet {
	// 	return competitionConfig, lvl, nil
	// }
	// return DefaultConfig(), lvl, nil
}

//Function Setting Config -- OP
//Opertor Config
func setOperatorConfigOP(operatorId string, config commondto.ConfigDto) error {
	// Get Operator Dto.
	opDTO, err := mainCache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("setOperatorConfigOP: Get Operator Details failed with error - ", err.Error())
		return err
	}
	config.IsSet = true
	opDTO.Config = config
	// update Operator DB
	err = database.ReplaceOperator(opDTO)
	if err != nil {
		log.Println("setOperatorConfigOP: Update Operator Details failed with error - ", err.Error())
		return err
	}
	// Update Operator Dto.
	mainCache.SetOperatorDetails(opDTO)
	return nil
}

func setPartnerStatus(operatorId string, partnerId string, providerId string, config commondto.ConfigDto) error {
	// Get Partner Dto.
	partnerKey := operatorId + "-" + partnerId + "-" + providerId
	partner, err := database.GetPartnerStatus(partnerKey)
	if err != nil {
		log.Println("SetPartnerStatus: database.GetPartnerStatus failed with error - ", err.Error())
		return err
	}
	//config.IsSet = true
	partner.Config = config
	// update Partner DB
	err = database.ReplacePartner(partner)
	if err != nil {
		log.Println("SetPartnerStatus: Update Partner Details failed with error - ", err.Error())
		return err
	}
	// Update Partner Dto.
	mainCache.SetPartnerStatus(partner)
	return nil
}

// Sport Config
func setSportConfigOP(operatorId string, partnerId string, providerId string, sportId string, config commondto.ConfigDto) error {
	// Get Sport Dto.
	sportKey := operatorId + "-" + providerId + "-" + sportId
	// if partnerId != "" {
	// 	sportKey = operatorId + "-" + partnerId + "-" + providerId + "-" + sportId
	// }
	sportDTO, err := database.GetSportStatus(sportKey)
	if err != nil {
		log.Println("setSportConfigOP: Get Sport Details failed with error - ", err.Error())
		return err
	}
	//config.IsSet = true
	sportDTO.Config = config
	// update Sport DB
	err = database.ReplaceSportStatus(sportDTO)
	if err != nil {
		log.Println("setSportConfigOP: Update Sport Details failed with error - ", err.Error())
		return err
	}
	// Update Sport Dto.
	mainCache.SetSportStatus(sportDTO)
	return nil
}

// Competition Config
func setCompetitionConfigOP(operatorId string, providerId string, sportId string, competitionId string, config commondto.ConfigDto) error {
	// Get Competition Dto.
	competitionDTO, err := mainCache.GetCompetitionStatus(operatorId, providerId, sportId, competitionId)
	if err != nil {
		// CompetitionStatus is missing in cache & db
		// get competition details and add competitionstatus
		err = mainCache.AddCompetitionStatusConfig(operatorId, providerId, sportId, competitionId, config)
		if err != nil {
			competitionStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + competitionId
			log.Println("setCompetitionConfigOP: mainCache.AddEventStatusConfig failed with error for competitionStatusKey - ", err.Error(), competitionStatusKey)
			return err
		}
		return nil
	}
	//config.IsSet = true
	competitionDTO.Config = config
	// update Competition DB
	err = database.ReplaceCompetitionStatus(competitionDTO)
	if err != nil {
		log.Println("setCompetitionConfigOP: Update Competition Details failed with error - ", err.Error())
		return err
	}
	// Update Competition Dto.
	mainCache.SetCompetitionStatus(competitionDTO)
	return nil
}

// Event Config
func setEventConfigOP(operatorId string, providerId string, sportId string, eventId string, config commondto.ConfigDto) error {
	// Get Event Dto.
	eventDTO, err := mainCache.GetEventStatus(operatorId, providerId, sportId, eventId)
	if err != nil {
		// EventStatus is missing in cache & db
		// get event details and add eventstatus
		err = mainCache.AddEventStatusConfig(operatorId, providerId, sportId, eventId, config)
		if err != nil {
			eventStatusKey := operatorId + "-" + providerId + "-" + sportId + "-" + eventId
			log.Println("setEventConfigOP: mainCache.AddEventStatusConfig failed with error for eventStatusKey - ", err.Error(), eventStatusKey)
			return err
		}
		return nil
	}
	//config.IsSet = true
	eventDTO.Config = config
	// update Event DB
	err = database.ReplaceEventStatus(eventDTO)
	if err != nil {
		log.Println("setEventConfigOP: Update Event Details failed with error - ", err.Error())
		return err
	}
	// Update Event Dto.
	mainCache.SetEventStatus(eventDTO)
	return nil
}

//Function getting Provider Config -- SAP
func getProviderConfigSAP(providerId string) (commondto.ConfigDto, string, error) {
	lvl := "provider"
	// providerConfig, err := mainCache.GetProvider(providerId)
	providerConfig, err := database.GetProvider(providerId)
	if err != nil {
		log.Println("getProviderConfigSAP: Get Provider Details failed with error - ", err.Error())
		return DefaultConfig(), "Provider", err
	}
	// Default Hold will always be -1.0
	if providerConfig.Config.IsSet {
		providerConfig.Config = getConfigDelay(providerConfig.Config)
		return providerConfig.Config, lvl, nil
	}
	return DefaultConfig(), lvl, nil
}

//Function getting Sport Config -- SAP
func getSportConfigSAP(providerId string, sportId string) (commondto.ConfigDto, string, error) {
	lvl := "sport"
	// sportConfig, err := mainCache.GetSport(providerId + "-" + sportId)
	sportConfig, err := database.GetSport(providerId + "-" + sportId)
	if err != nil {
		log.Println("getSportConfigSAP: Get Sport Details failed with error - ", err.Error())
		return DefaultConfig(), "Sport", err
	}
	// Default Hold will always be -1.0
	if sportConfig.Config.IsSet {
		sportConfig.Config = getConfigDelay(sportConfig.Config)
		return sportConfig.Config, lvl, nil
	}
	sapConfig, lvl, err := getProviderConfigSAP(providerId)
	if err != nil {
		log.Println("getSportConfigSAP: Get Provider Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return sapConfig, lvl, nil
}

func getCompetitionConfigSAP(providerId string, sportId string, competitionId string) (commondto.ConfigDto, string, error) {
	lvl := "competition"
	// competitionConfig, err := mainCache.GetCompetition(providerId + "-" + sportId + "-" + competitionId)
	competitionConfig, err := database.GetCompetition(providerId + "-" + sportId + "-" + competitionId)
	if err != nil {
		log.Println("getCompetitionConfigSAP: Get Competition Details failed with error - ", err.Error())
		return DefaultConfig(), "Competition", err
	}
	// Default Hold will always be -1.0
	if competitionConfig.Config.IsSet {
		competitionConfig.Config = getConfigDelay(competitionConfig.Config)
		return competitionConfig.Config, lvl, nil
	}
	sapConfig, lvl, err := getSportConfigSAP(providerId, sportId)
	if err != nil {
		log.Println("getCompetitionConfigSAP: Get Sport Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return sapConfig, lvl, nil
}

func getEventConfigSAP(providerId string, sportId string, competitionId string, eventId string) (commondto.ConfigDto, string, error) {
	lvl := "event"
	// eventConfig, err := mainCache.GetEvent(providerId + "-" + sportId + "-" + eventId)
	eventConfig, err := database.GetEventDetails(providerId + "-" + sportId + "-" + eventId)
	if err != nil {
		log.Println("getEventConfigSAP: Get Event Details failed with error - ", err.Error())
		return DefaultConfig(), "Event", err
	}
	// Default Hold will always be -1.0
	if eventConfig.Config.IsSet {
		eventConfig.Config = getConfigDelay(eventConfig.Config)
		return eventConfig.Config, lvl, nil
	}
	sapConfig, lvl, err := getCompetitionConfigSAP(providerId, sportId, competitionId)
	if err != nil {
		log.Println("getEventConfigSAP: Get Competition Details failed with error - ", err.Error())
		return DefaultConfig(), lvl, err
	}
	return sapConfig, lvl, nil
}

func setProviderConfigSAP(providerId string, config commondto.ConfigDto) error {
	// Get Provider Dto.
	providerDTO, err := mainCache.GetProvider(providerId)
	if err != nil {
		log.Println("setProviderConfigSAP: Get Provider Details failed with error - ", err.Error())
		return err
	}
	//config.IsSet = true
	providerDTO.Config = config
	// update Provider DB
	err = database.ReplaceProvider(providerDTO)
	if err != nil {
		log.Println("setProviderConfigSAP: Update Provider Details failed with error - ", err.Error())
		return err
	}
	// Update Provider Dto.
	mainCache.SetProvider(providerDTO)
	return nil
}

func setSportConfigSAP(providerId string, sportId string, config commondto.ConfigDto) error {
	// Get Sport Dto.
	sportDTO, err := mainCache.GetSport(providerId + "-" + sportId)
	if err != nil {
		log.Println("setSportConfigSAP: Get Sport Details failed with error - ", err.Error())
		return err
	}
	//config.IsSet = true
	sportDTO.Config = config
	// update Sport DB
	err = database.ReplaceSport(sportDTO)
	if err != nil {
		log.Println("setSportConfigSAP: Update Sport Details failed with error - ", err.Error())
		return err
	}
	// Update Sport Dto.
	mainCache.SetSport(sportDTO)
	return nil
}

func setCompetitionConfigSAP(providerId string, sportId string, competitionId string, config commondto.ConfigDto) error {
	// Get Competition Dto.
	competitionDTO, err := mainCache.GetCompetition(providerId, sportId, competitionId)
	if err != nil {
		log.Println("setCompetitionConfigSAP: Get Competition Details failed with error - ", err.Error())
		return err
	}
	//config.IsSet = true
	competitionDTO.Config = config
	// update Competition DB
	err = database.ReplaceCompetition(competitionDTO)
	if err != nil {
		log.Println("setCompetitionConfigSAP: Update Competition Details failed with error - ", err.Error())
		return err
	}
	// Update Competition Dto.
	mainCache.SetCompetition(competitionDTO)
	return nil
}

func setEventConfigSAP(providerId string, sportId string, eventId string, config commondto.ConfigDto) error {
	// Get Event Dto.
	eventDTO, err := mainCache.GetEvent(providerId, sportId, eventId)
	if err != nil {
		log.Println("setEventConfigSAP: Get Event Details failed with error - ", err.Error())
		return err
	}
	//config.IsSet = true
	eventDTO.Config = config
	// update Event DB
	err = database.ReplaceEvent(eventDTO)
	if err != nil {
		log.Println("setEventConfigSAP: Update Event Details failed with error - ", err.Error())
		return err
	}
	// Update Event Dto.
	mainCache.SetEvent(eventDTO)
	return nil
}

//Set Default Config
func DefaultConfig() commondto.ConfigDto {
	var config commondto.ConfigDto
	var features commondto.Features
	features.Min = 0
	features.Max = 1
	features.Delay = 0
	config.Hold = 0
	config.MatchOdds = features
	config.Fancy = features
	config.Bookmaker = features
	config.IsSet = false
	return config
}

func PortalAudit(Tknmeta TokenMetadata, c *fiber.Ctx, ObId, ObType string) {
	audit := models.Audit{}
	audit.UserId = Tknmeta.UserId
	audit.Operation = c.OriginalURL()
	audit.IP = Tknmeta.IP
	audit.OperatorId = Tknmeta.OperatorId
	audit.ObjectId = ObId
	audit.ObjectType = ObType
	audit.Payload = string(c.Body())
	audit.UserRole = Tknmeta.Role
	err := database.InsertPortalAudit(audit)
	if err != nil {
		log.Println("PortalAudit: Insert Portal Audit failed with error - ", err.Error())
		log.Printf("PortalAudit: Audit model : %+v\n", audit)
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// Convert config delay from milliseconds to seconds
func getConfigDelay(config commondto.ConfigDto) commondto.ConfigDto {
	/*
		if config.Bookmaker.Delay > 1000 {
			config.Bookmaker.Delay = config.Bookmaker.Delay / 1000
		}
		if config.Fancy.Delay > 1000 {
			config.Fancy.Delay = config.Fancy.Delay / 1000
		}
		if config.MatchOdds.Delay > 1000 {
			config.MatchOdds.Delay = config.MatchOdds.Delay / 1000
		}
	*/
	return config
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
