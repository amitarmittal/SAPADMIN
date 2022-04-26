package sports

import "go.mongodb.org/mongo-driver/bson/primitive"

type OddsData struct {
	OddsKey   string  `bson:"odds_key"` // BEFORE / CURRENT / AFTER
	OddsValue float64 `bson:"odds_value,truncate"`
	OddsAt    string  `bson:"odds_at"` // Date in readable string format in UTC (from UNIX timestamp)
}

type RollbackReqDto struct {
	ReqId          string  `bson:"request_id"`
	ReqTime        int64   `bson:"request_time"`
	RollbackAmount float64 `bson:"rollback_amount,truncate"`
	OperatorAmount float64 `bson:"operator_amount,truncate"`
	PlatformAmount float64 `bson:"platform_amount,truncate"`
	RollbackReason string  `bson:"rollback_reason"`
}

type ResultReqDto struct {
	ReqId          string  `bson:"request_id"`
	ReqTime        int64   `bson:"request_time"`
	CreditAmount   float64 `bson:"credit_amount,truncate"`
	OperatorAmount float64 `bson:"operator_amount,truncate"`
	PlatformAmount float64 `bson:"platform_amount,truncate"`
	RunnerName     string  `bson:"runner_name"`
	SessionOutcome float64 `bson:"session_outcome,truncate"`
}

type BetReqDto struct {
	BetId          string  `bson:"bet_id"` // only for BetFair
	ReqId          string  `bson:"request_id"`
	ReqTime        int64   `bson:"request_time"`
	DebitAmount    float64 `bson:"debit_amount,truncate"`
	OperatorHold   float64 `bson:"operator_hold"`
	PlatformHold   float64 `bson:"platform_hold"`
	OperatorAmount float64 `bson:"operator_amount,truncate"`
	PlatformAmount float64 `bson:"platform_amount,truncate"`
	Rate           int32   `bson:"rate"`
	SizePlaced     float64 `bson:"size_placed"`
	SizeMatched    float64 `bson:"size_matched"`
	SizeRemaining  float64 `bson:"size_remaining"`
	SizeLapsed     float64 `bson:"size_lapsed"`
	SizeCancelled  float64 `bson:"size_cancelled"`
	SizeVoided     float64 `bson:"size_voided"`
	OddsMatched    float64 `bson:"odds_matched"`
}

type BetDetailsDto struct {
	SportName       string  `bson:"sport_name"`
	CompetitionName string  `bson:"competition_name"`
	EventName       string  `bson:"event_name"`
	BetType         string  `bson:"betType"` // BACK/LAY
	OddValue        float64 `bson:"oddValue,truncate"`
	StakeAmount     float64 `bson:"stakeAmount,truncate"`
	MarketType      string  `bson:"marketType"`              // MATCH_ODDS/BOOK_MAKER/GOAL_MARKETS/FANCY_MARKET/
	MarketName      string  `bson:"marketName"`              // for soccer, match_odds or goal_markets, can be avoided
	RunnerId        string  `bson:"runnerId"`                // ?? what is for fancy
	RunnerName      string  `bson:"runnerName"`              // ?? what is for fancy
	SessionOutcome  float64 `bson:"sessionOutcome,truncate"` // fancy scrore ex: 45 NO, 46 YES
	IsUnmatched     bool    `bson:"is_unmatched"`            // For BetFair. Default is false
}

type BetDto struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	EventKey      string             `bson:"event_key"`
	OperatorId    string             `bson:"operator_id"`
	PartnerId     string             `bson:"partner_id"`
	Token         string             `bson:"token"`
	ProviderId    string             `bson:"provider_id"`
	SportId       string             `bson:"sport_id"`
	CompetitionId string             `bson:"competition_id"`
	UserId        string             `bson:"user_id"`
	UserName      string             `bson:"user_name"`
	EventId       string             `bson:"event_id"`
	MarketId      string             `bson:"market_id"`
	BetId         string             `bson:"transaction_id"`
	NetAmount     float64            `bson:"net_amount,truncate"`
	Commission    float64            `bson:"commission"`
	CommLevel     string             `bson:"commission_level"`
	BetReq        BetReqDto          `bson:"bet_req"`
	ResultReqs    []ResultReqDto     `bson:"result_reqs"`
	RollbackReqs  []RollbackReqDto   `bson:"rollback_reqs"`
	BetDetails    BetDetailsDto      `bson:"bet_details"`
	OddsHistory   []OddsData         `bson:"odds_history"`
	Status        string             `bson:"status"` // unmatched / open(rollback) / settled / void / cancel / deleted
	UserIp        string             `bson:"user_ip"`
	CreatedAt     int64              `bson:"created_at"` // DateTime in Unix seconds
	UpdatedAt     int64              `bson:"updated_at"`
}
