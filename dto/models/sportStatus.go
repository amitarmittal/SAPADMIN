package models

import (
	"Sp/dto/commondto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 1. Many to Many mapping between Operator and Sport.
// 2. Controlling at three levels
//   2.1. Platform Level (Handled at Sports Table)
//   2.2. for selected Operators - ProviderStatus
//   2.3. Operator Level - OperatorStatus
// Old SportKey: OperatorId+"-"+ProviderId+"-"+SportId
// New SportKey: OperatorId+"-"+PartnerId+"-"+ProviderId+"-"+SportId
type SportStatus struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty"`
	SportKey       string              `bson:"sport_key"` // OperatorId+"-"+ProviderId+"-"+SportId //OperatorId+"-"+PartnerId+"-"+ProviderId+"-"+SportId
	OperatorId     string              `bson:"operator_id"`
	OperatorName   string              `bson:"operator_name"`
	PartnerId      string              `bson:"partner_id"`
	ProviderId     string              `bson:"provider_id"`     // Dream / BetFair / SportRadar
	ProviderName   string              `bson:"provider_name"`   // Dream Sports / Bet Fair / Sport Radar
	SportId        string              `bson:"sport_id"`        // 1 / 2 / 4
	SportName      string              `bson:"sport_name"`      // Soccer / Tennis / Cricket
	ProviderStatus string              `bson:"provider_status"` // Provider swtich to block to a particular operator, Default BLOCKED
	OperatorStatus string              `bson:"operator_status"` // Operator swatich to block for themselve, Default BLOCKED
	Favourite      bool                `bson:"favourite"`       // False - non-favourite. Default is False
	CreatedAt      int64               `bson:"created_at"`      // DateTime in Unix seconds
	UpdatedAt      int64               `bson:"updated_at"`      // DateTime in Unix seconds
	Config         commondto.ConfigDto `bson:"config"`          // Sport specific configuration
	IsCommission   bool                `bson:"is_commission"`   // Default false, NO commission
	WinCommission  float64             `bson:"win_commission"`  // if enabled, default (minimum) is 2%
}
