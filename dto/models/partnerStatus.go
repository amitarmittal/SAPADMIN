package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 1. Many to Many mapping between Operator+Partner and Provider.
// 2. Controlling at three levels
//   2.1. Platform Level
//   2.2. for selected Operators
//   2.3. Operator Level
type PartnerStatus struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty"`
	PartnerKey     string              `bson:"partner_key"` // OperatorId+"-"+PartnerId+"-"+ProviderId
	OperatorId     string              `bson:"operator_id"`
	OperatorName   string              `bson:"operator_name"`
	PartnerId      string              `bson:"partner_id"`
	ProviderId     string              `bson:"provider_id"`     // Dream / BetFair / SportRadar
	ProviderName   string              `bson:"provider_name"`   // Dream Sports / Bet Fair / Sport Radar
	ProviderStatus string              `bson:"provider_status"` // Provider switch to block to all operators
	OperatorStatus string              `bson:"operator_status"` // Operator swatich to block for themselve
	Favourite      bool                `bson:"favourite"`       // False - non-favourite. Default is False
	CreatedAt      int64               `bson:"created_at"`      // DateTime in Unix seconds
	UpdatedAt      int64               `bson:"updated_at"`      // DateTime in Unix seconds
	Config         commondto.ConfigDto `bson:"config"`          // Configuration for the provider
	IsCommission   bool                `bson:"is_commission"`   // Default false, NO commission
	WinCommission  float64             `bson:"win_commission"`  // if enabled, default (minimum) is 2%
}
