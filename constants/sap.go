package constants

const ()

type sap struct {
	BetType         sapBetType
	BetStatus       sapbetStatus
	ObjectStatus    sapObjectStatus
	WalletType      sapWalletType
	MarketType      sapMarketType
	LedgerTxType    sapLedgerTxType
	ProviderType    sapProviderType
	ProviderName    sapProviderName
	OperationKeys   sapOperationKeys
	OperationStatus sapOperationStatus
	ObjectTypes     sapObjectTypes
}
type sapBetType struct {
	back string
	lay  string
}
type sapbetStatus struct {
	open           string
	unmatched      string
	settled        string
	voided         string
	settled_voided string
	timely_voided  string
	lapsed         string
	cancelled      string
	rollback       string
	expired        string
	//inprocess      string
}

type sapObjectStatus struct {
	active  string
	blocked string
}

type sapWalletType struct {
	seamless string
	transfer string
	feed     string
}

type sapMarketType struct {
	match_odds string
	bookmaker  string
	fancy      string
	line_odds  string
}

type sapLedgerTxType struct {
	deposit          string
	withdraw         string
	betplacement     string
	betresult        string
	betrollback      string
	betcancel        string
	marketcommission string
}

type sapProviderType struct {
	dream      string
	betfair    string
	sportradar string
}

type sapProviderName struct {
	dreamname      string
	betfairname    string
	sportradarname string
}

type sapOperationKeys struct {
	betfairkeepalive               string
	betfailrholbetssettlement      string
	betfairclearedorderssettlement string
	betfairmarketcommission        string
	betfaircompetitionidsync       string
	userledgerbalancesync          string
}

type sapOperationStatus struct {
	free string
	busy string
	hold string
}

type sapObjectTypes struct {
	operator string
	provider string
}

func newSAPConstants() sap {
	sc := sap{}
	// BetType
	sc.BetType.back = back
	sc.BetType.lay = lay
	// BetStatus
	sc.BetStatus.open = open
	sc.BetStatus.unmatched = unmatched
	sc.BetStatus.settled = settled
	sc.BetStatus.voided = voided
	sc.BetStatus.settled_voided = settled_voided
	sc.BetStatus.timely_voided = timely_voided
	sc.BetStatus.lapsed = lapsed
	sc.BetStatus.cancelled = cancelled
	sc.BetStatus.rollback = rollback
	sc.BetStatus.expired = expired
	//sc.BetStatus.inprocess = inprocess
	// Object Status
	sc.ObjectStatus.active = active
	sc.ObjectStatus.blocked = blocked
	// Wallet Types
	sc.WalletType.seamless = seamless
	sc.WalletType.transfer = transfer
	sc.WalletType.feed = feed
	// Market Types
	sc.MarketType.match_odds = match_odds
	sc.MarketType.bookmaker = bookmaker
	sc.MarketType.fancy = fancy
	sc.MarketType.line_odds = line_odds
	// Ledger Tx Types
	sc.LedgerTxType.deposit = deposit
	sc.LedgerTxType.withdraw = withdraw
	sc.LedgerTxType.betplacement = betplacement
	sc.LedgerTxType.betresult = betresult
	sc.LedgerTxType.betrollback = betrollback
	sc.LedgerTxType.betcancel = betcancel
	sc.LedgerTxType.marketcommission = marketcommission
	// Provider IDs
	sc.ProviderType.dream = dream
	sc.ProviderType.betfair = betfair
	sc.ProviderType.sportradar = sportradar
	// Provider Names
	sc.ProviderName.dreamname = dreamname
	sc.ProviderName.betfairname = betfairname
	sc.ProviderName.sportradarname = sportradarname
	// OperationLock Keys
	sc.OperationKeys.betfairkeepalive = betfair_keepalive
	sc.OperationKeys.betfailrholbetssettlement = betfair_holdbets_settlement
	sc.OperationKeys.betfairclearedorderssettlement = betfair_clearedorders_settlement
	sc.OperationKeys.betfairmarketcommission = betfair_market_commission
	sc.OperationKeys.betfaircompetitionidsync = betfair_competitionid_sync
	sc.OperationKeys.userledgerbalancesync = user_ledger_balance_sync
	// OperationLock Status
	sc.OperationStatus.free = free
	sc.OperationStatus.busy = busy
	sc.OperationStatus.hold = hold
	// Object Types
	sc.ObjectTypes.operator = operator
	sc.ObjectTypes.provider = provider
	return sc
}

// BetType methods
func (s sapBetType) BACK() string {
	return s.back
}
func (s sapBetType) LAY() string {
	return s.lay
}

// BetStatus methods
func (s sapbetStatus) OPEN() string {
	return s.open
}
func (s sapbetStatus) UNMATCHED() string {
	return s.unmatched
}
func (s sapbetStatus) SETTLED() string {
	return s.settled
}
func (s sapbetStatus) VOIDED() string {
	return s.voided
}
func (s sapbetStatus) SETTLED_VOIDED() string {
	return s.settled_voided
}
func (s sapbetStatus) TIMELY_VOIDED() string {
	return s.timely_voided
}
func (s sapbetStatus) LAPSED() string {
	return s.lapsed
}
func (s sapbetStatus) CANCELLED() string {
	return s.cancelled
}
func (s sapbetStatus) ROLLBACK() string {
	return s.rollback
}
func (s sapbetStatus) EXPIRED() string {
	return s.expired
}

// func (s sapbetStatus) INPROCESS() string {
// 	return s.inprocess
// }

// ObjectStatus
func (s sapObjectStatus) ACTIVE() string {
	return s.active
}
func (s sapObjectStatus) BLOCKED() string {
	return s.blocked
}

// WalletType
func (s sapWalletType) Seamless() string {
	return s.seamless
}
func (s sapWalletType) Transfer() string {
	return s.transfer
}
func (s sapWalletType) Feed() string {
	return s.feed
}

// MarketType
func (s sapMarketType) MATCH_ODDS() string {
	return s.match_odds
}
func (s sapMarketType) BOOKMAKER() string {
	return s.bookmaker
}
func (s sapMarketType) FANCY() string {
	return s.fancy
}
func (s sapMarketType) LINE_ODDS() string {
	return s.line_odds
}

// MarketType
func (s sapLedgerTxType) DEPOSIT() string {
	return s.deposit
}
func (s sapLedgerTxType) WITHDRAW() string {
	return s.withdraw
}
func (s sapLedgerTxType) BETPLACEMENT() string {
	return s.betplacement
}
func (s sapLedgerTxType) BETRESULT() string {
	return s.betresult
}
func (s sapLedgerTxType) BETROLLBACK() string {
	return s.betrollback
}
func (s sapLedgerTxType) BETCANCEL() string {
	return s.betcancel
}
func (s sapLedgerTxType) MARKETCOMMISSION() string {
	return s.marketcommission
}

// ProviderType
func (s sapProviderType) Dream() string {
	return s.dream
}
func (s sapProviderType) BetFair() string {
	return s.betfair
}
func (s sapProviderType) SportRadar() string {
	return s.sportradar
}

// ProviderType
func (s sapProviderName) DreamName() string {
	return s.dreamname
}
func (s sapProviderName) BetFairName() string {
	return s.betfairname
}
func (s sapProviderName) SportRadarName() string {
	return s.sportradarname
}

// OperationLock Keys
func (s sapOperationKeys) BETFAIR_KEEPALIVE() string {
	return s.betfairkeepalive
}
func (s sapOperationKeys) BETFAIR_HOLDBETS_SETTLEMENT() string {
	return s.betfailrholbetssettlement
}
func (s sapOperationKeys) BETFAIR_CLEAREDORDERS_SETTLEMENT() string {
	return s.betfairclearedorderssettlement
}
func (s sapOperationKeys) BETFAIR_MARKET_COMMISSION() string {
	return s.betfairmarketcommission
}
func (s sapOperationKeys) BETFAIR_COMPETITION_SYNC() string {
	return s.betfaircompetitionidsync
}
func (s sapOperationKeys) USER_LEDGER_BALANCE_SYNC() string {
	return s.userledgerbalancesync
}

// OperationLock Status
func (s sapOperationStatus) FREE() string {
	return s.free
}
func (s sapOperationStatus) BUSY() string {
	return s.busy
}
func (s sapOperationStatus) HOLD() string {
	return s.hold
}

// Object Types
func (s sapObjectTypes) OPERATOR() string {
	return s.operator
}
func (s sapObjectTypes) PROVIDER() string {
	return s.provider
}
