package operatordto

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProviderConfig struct {
	OperatorHold float64          `bson:"operator_hold"`
	PlatformHold float64          `bson:"platform_hold"`
	PlatformComm float64          `bson:"platform_comm"`
	BetDelay     map[string]int32 `bson:"bet_delay"`
}

type KeysDto struct {
	PrivateKey  string `bson:"private_key"`  // Our Private Key used for signing
	PublicKey   string `bson:"public_key"`   // Our Public Key to share with Operator
	OperatorKey string `bson:"operator_key"` // Operator Public Key used for verfication
}

type Currency struct {
	Currency    string  `bson:"currency"` // INR / HKD / USD
	Commisssion float64 `bson:"commission"`
}

type Partner struct {
	PartnerId   string  `bson:"partner_id" json:"partnerId"`
	Status      string  `bson:"status" josn:"status"`     // ACTIVE / BLOCKED / SUSPENDED / DEACTIVATED
	Currency    string  `bson:"currency" json:"currency"` // INR / HKD / USD
	Rate        int32   `bson:"rate" json:"rate"`
	Commisssion float64 `bson:"commission" json:"commission"`
}

type OperatorDTO struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty"`
	OperatorId    string              `bson:"operator_id"`
	OperatorName  string              `bson:"operator_name"`
	BaseURL       string              `bson:"base_url"`
	DreamFeed     string              `bson:"dream_feed"`    // if empty, uses SAP feed
	BetFairFeed   string              `bson:"betfair_feed"`  // if empty, uses SAP feed
	Status        string              `bson:"status"`        // ACTIVE / BLOCKED / SUSPENDED / DEACTIVATED
	WalletType    string              `bson:"wallet_type"`   // Seamless / Transfer / Feed
	BetFairPlus   bool                `bson:"betfair_plus"`  // BetFair MO + Dream BM + Dream Fancy
	MarketResult  bool                `bson:"market_result"` // true - to notify market results
	Signature     bool                `bson:"signature"`     // true - to enforce signature validation
	NewSession    bool                `bson:"new_session"`   // true - to create new session for every login call
	BetUpdates    bool                `bson:"bet_updates"`   // true - to call wallet update bet endpoint to notify changes
	BetLock       bool                `bson:"bet_lock"`      // true - to stop bets
	Keys          KeysDto             `bson:"keys"`
	Currencies    []Currency          `bson:"currencies"`
	Partners      []Partner           `bson:"partners"`
	Config        commondto.ConfigDto `bson:"config"` // key is providerId
	Ips           []string            `bson:"ips"`
	Balance       float64             `bson:"balance,truncate"`
	CreatedAt     int64               `bson:"created_at"`     // DateTime in Unix seconds
	UpdatedAt     int64               `bson:"updated_at"`     // DateTime in Unix seconds
	IsCommission  bool                `bson:"is_commission"`  // Default false, NO commission
	WinCommission float64             `bson:"win_commission"` // if enabled, default (minimum) is 2%
}

//OperatorUser's DTO
type PortalUser struct {
	UserKey    string `bson:"user_key"` // OperatorId+UserId
	UserId     string `bson:"user_id"`
	OperatorId string `bson:"operator_id"`
	Password   string `bson:"password"` //Encrypted
	UserName   string `bson:"user_name"`
	Status     string `bson:"status"` // active / blocked / deleted
	Role       string `bson:"role"`   // SPAdmin / OperatorAdmin
}

type PortalUserReq struct {
	UserId     string `json:"user_id"`
	OperatorId string `json:"operator_id"`
	Password   string `json:"password"` //Encrypted
	UserName   string `json:"user_name"`
	Status     string `json:"status"` // active / blocked / deleted
	Role       string `json:"role"`   // SPAdmin / OperatorAdmin
}

// SessionToken model: Used for Maintaining user sessions for Portal
type PortalSession struct {
	UserId     string `bson:"user_id"`
	OperatorId string `bson:"operator_id"`
	JWTToken   string `bson:"jwt_token"`
	ExpiresAt  int64  `bson:"exp_time"`
}
