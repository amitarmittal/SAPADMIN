package commondto

type Features struct {
	Min       int32   `json:"min" bson:"min"`
	Max       int32   `json:"max" bson:"max"`
	Delay     float64 `json:"delay" bson:"delay"`
	OddLimits int32   `json:"oddLimits" bson:"odd_limits"`
}

type ConfigDto struct {
	Hold      float64  `json:"hold" bson:"hold"`
	MatchOdds Features `json:"matchOdds" bson:"match_odds"`
	Fancy     Features `json:"fancy" bson:"fancy"`
	Bookmaker Features `json:"bookmaker" bson:"bookmaker"`
	IsSet     bool     `json:"isSet" bson:"is_set"`
}

// GetConfigRespDto represents response body of this API
type GetConfigRespDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	Level            string    `json:"level"`
	Config           ConfigDto `json:"config"`
}
