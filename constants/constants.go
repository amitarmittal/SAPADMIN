package constants

const (
	// Side
	back = "BACK"
	lay  = "LAY"
	// OrderStatus
	execution_complete = "EXECUTION_COMPLETE"
	executable         = "EXECUTABLE"
	expired            = "EXPIRED"
	pending            = "PENDING"
	// BetStatus
	settled        = "SETTLED"
	voided         = "VOIDED"
	settled_voided = "SETTLED_VOIDED"
	timely_voided  = "TIMELY_VOIDED"
	lapsed         = "LAPSED"
	cancelled      = "CANCELLED"
	unmatched      = "UNMATCHED"
	rollback       = "ROLLBACK"
	//inprocess      = "INPROCESS"
	// BetOutcome
	won   = "WON"
	lose  = "LOSE"
	place = "PLACE"
	// MarketStatus
	inactive  = "INACTIVE"
	open      = "OPEN"
	suspended = "SUSPENDED"
	closed    = "CLOSED"
	// RunnerStatus
	active         = "ACTIVE"
	winner         = "WINNER"
	loser          = "LOSER"
	placed         = "PLACED"
	removed_vacant = "REMOVED_VACANT"
	removed        = "REMOVED"
	hidden         = "HIDDEN"
	// Provider/Operator/Sports/Competitions/Events Status
	blocked = "BLOCKED"
	// Portal Audit
	USER        = "USER"
	SPORT       = "SPORT"
	COMPETITION = "COMPETITION"
	EVENT       = "EVENT"
	MARKET      = "MARKET"
	OPERATOR    = "OPERATOR"
	PROVIDER    = "PROVIDER"
	PORTAL_USER = "PORTAL USER"
	// wallet types
	seamless = "Seamless"
	transfer = "Transfer"
	feed     = "Feed"
	// Market Types
	match_odds = "MATCH_ODDS"
	bookmaker  = "BOOKMAKER"
	fancy      = "FANCY"
	line_odds  = "LINE_ODDS"
	// Ledger Tx Types
	deposit          = "DEPOSIT"
	withdraw         = "WITHDRAW"
	betplacement     = "BETPLACEMENT"
	betresult        = "BETRESULT"
	betrollback      = "BETROLLBACK"
	betcancel        = "BETCANCEL"
	marketcommission = "MARKETCOMMISSION"
	// Provider Ids
	dream      = "Dream"
	betfair    = "BetFair"
	sportradar = "SportRadar"
	// Provider Names
	dreamname      = "Dream Feed"
	betfairname    = "Bet Fair"
	sportradarname = "Sport Radar"
	// Rollback Types (sportradar)
	//ROLLBACK  ("Rollback"),
	//TIMELY_VOID("TimelyVoid"),
	//TIMELY_VOID_ROLLBACK("TimelyVoidRollback");
	rollback2          = "Rollback"
	timelyVoid         = "TimelyVoid"
	timelyVoidRollback = "TimelyVoidRollback"
	// Result Type (sportradar)
	result_settlement = "RESULT_SETTLEMENT"
	void_settlement   = "VOID_SETTLEMENT"
	// OperationLock Keys
	betfair_keepalive                = "BETFAIR_KEEPALIVE"
	betfair_holdbets_settlement      = "BETFAIR_HOLDBETS_SETTLEMENT"
	betfair_clearedorders_settlement = "BETFAIR_CLEAREDORDERS_SETTLEMENT"
	betfair_market_commission        = "BETFAIR_MARKET_COMMISSION"
	betfair_competitionid_sync       = "BETFAIR_COMPETITIONID_SYNC"
	user_ledger_balance_sync         = "USER_LEDGER_BALANCE_SYNC"
	// OperationLock Status
	free = "FREE"
	busy = "BUSY"
	hold = "HOLD"
	// Object Types
	operator = "OPERATOR"
	provider = "PROVIDER"
)

var (
	BetFair    betFair    = newBetFairConstants()
	SAP        sap        = newSAPConstants()
	SportRadar sportRadar = newSportRadarConstants()
)
