package constants

type betFair struct {
	Side         side
	OrderStatus  orderStatus
	BetStatus    betStatus
	BetOutcome   betOutcome
	MarketStatus marketStatus
	RunnerStatus runnerStatus
}
type side struct {
	back string
	lay  string
}
type orderStatus struct {
	execution_complete string
	executable         string
	expired            string
	pending            string
}
type betStatus struct {
	settled   string
	voided    string
	lapsed    string
	cancelled string
}
type betOutcome struct {
	won   string
	lose  string
	place string
}
type marketStatus struct {
	inactive  string
	open      string
	suspended string
	closed    string
}
type runnerStatus struct {
	active         string
	winner         string
	loser          string
	placed         string
	removed_vacant string
	removed        string
	hidden         string
}

func newBetFairConstants() betFair {
	bf := betFair{}
	// Side
	bf.Side.back = back
	bf.Side.lay = lay
	// OrderStatus
	bf.OrderStatus.execution_complete = execution_complete
	bf.OrderStatus.executable = executable
	bf.OrderStatus.expired = expired
	bf.OrderStatus.pending = pending
	// BetStatus
	bf.BetStatus.settled = settled
	bf.BetStatus.voided = voided
	bf.BetStatus.lapsed = lapsed
	bf.BetStatus.cancelled = cancelled
	// BetOutcome
	bf.BetOutcome.won = won
	bf.BetOutcome.lose = lose
	bf.BetOutcome.place = place
	// MarketStatus
	bf.MarketStatus.inactive = inactive
	bf.MarketStatus.open = open
	bf.MarketStatus.suspended = suspended
	bf.MarketStatus.closed = closed
	// RunnerStatus
	bf.RunnerStatus.active = active
	bf.RunnerStatus.winner = winner
	bf.RunnerStatus.loser = loser
	bf.RunnerStatus.placed = placed
	bf.RunnerStatus.removed_vacant = removed_vacant
	bf.RunnerStatus.removed = removed
	bf.RunnerStatus.hidden = hidden
	return bf
}

// Side methods
func (s side) BACK() string {
	return s.back
}
func (s side) LAY() string {
	return s.lay
}

// OrderStatsu methods
func (s orderStatus) EXECUTION_COMPLETE() string {
	return s.execution_complete
}
func (s orderStatus) EXECUTABLE() string {
	return s.executable
}
func (s orderStatus) EXPIRED() string {
	return s.expired
}
func (s orderStatus) PENDING() string {
	return s.pending
}

// BetStatus methods
func (s betStatus) SETTLED() string {
	return s.settled
}
func (s betStatus) VOIDED() string {
	return s.voided
}
func (s betStatus) LAPSED() string {
	return s.lapsed
}
func (s betStatus) CANCELLED() string {
	return s.cancelled
}

// BetOutcome methods
func (s betOutcome) WON() string {
	return s.won
}
func (s betOutcome) LOSE() string {
	return s.lose
}
func (s betOutcome) PLACE() string {
	return s.place
}

// MarketStatus methods
func (s marketStatus) INACTIVE() string {
	return s.inactive
}
func (s marketStatus) OPEN() string {
	return s.open
}
func (s marketStatus) SUSPENDED() string {
	return s.suspended
}
func (s marketStatus) CLOSED() string {
	return s.closed
}

// RunnerStatus methods
func (s runnerStatus) ACTIVE() string {
	return s.active
}
func (s runnerStatus) WINNER() string {
	return s.winner
}
func (s runnerStatus) LOSER() string {
	return s.loser
}
func (s runnerStatus) PLACED() string {
	return s.placed
}
func (s runnerStatus) REMOVED_VACANT() string {
	return s.removed_vacant
}
func (s runnerStatus) REMOVED() string {
	return s.removed
}
func (s runnerStatus) HIDDEN() string {
	return s.hidden
}
