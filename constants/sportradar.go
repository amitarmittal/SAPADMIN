package constants

type sportRadar struct {
	MarketStatus srMarketSettledStatus
	RollbackType srRollbackType
	ResultType   srResultType
}

type srMarketSettledStatus struct {
	voided  string
	open    string
	settled string
}

// Rollback / TimelyVoid / TimelyVoidRollback
type srRollbackType struct {
	rollback           string
	timelyVoid         string
	timelyVoidRollback string
}

// RESULT_SETTLEMENT / VOID_SETTLEMENT
type srResultType struct {
	result_settlement string
	void_settlement   string
}

func newSportRadarConstants() sportRadar {
	sr := sportRadar{}
	// MarketStatus
	sr.MarketStatus.voided = voided
	sr.MarketStatus.open = open
	sr.MarketStatus.settled = settled
	// RollbackType
	sr.RollbackType.rollback = rollback2
	sr.RollbackType.timelyVoid = timelyVoid
	sr.RollbackType.timelyVoidRollback = timelyVoidRollback
	// ResultType
	sr.ResultType.result_settlement = result_settlement
	sr.ResultType.void_settlement = void_settlement
	return sr
}

// MarketSettledStatus methods
func (s srMarketSettledStatus) VOIDED() string {
	return s.voided
}
func (s srMarketSettledStatus) OPEN() string {
	return s.open
}
func (s srMarketSettledStatus) SETTLED() string {
	return s.settled
}

// RollbackType methods
func (s srRollbackType) Rollback() string {
	return s.rollback
}
func (s srRollbackType) TimelyVoid() string {
	return s.timelyVoid
}
func (s srRollbackType) TimelyVoidRollback() string {
	return s.timelyVoidRollback
}

// ResultType methods
func (s srResultType) RESULT_SETTLEMENT() string {
	return s.result_settlement
}
func (s srResultType) VOID_SETTLEMENT() string {
	return s.void_settlement
}
