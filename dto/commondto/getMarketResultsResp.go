package commondto

import (
	sportradar "Sp/dto/providers/sportradar"
)

/*
  {
    "id": 37,
    "createdDate": 1639293649534,
    "updatedAt": 1639293649534,
    "sportId": "sr:sport:4",
    "eventId": "sr:match:28274864",
    "marketId": "sr:match:28274864:mr:16:sp:hcp=1.5",
    "marketName": "Handicap",
    "marketType": "MATCH_ODDS",
    "marketStatus": "RESULT_SETTLEMENT",
    "results": [
      {
        "receivedAt": 1639293649377,
        "runners": [
          {
            "runnerId": "sr:match:28274864:mr:16:sp:hcp=1.5:rn:1714",
            "runnerName": "San Jose Sharks (+1.5)",
            "result": "Won",
            "deadHeatFactor": 1,
            "voidFactor": 0
          },
          {
            "runnerId": "sr:match:28274864:mr:16:sp:hcp=1.5:rn:1715",
            "runnerName": "Dallas Stars (-1.5)",
            "result": "Lost",
            "deadHeatFactor": 1,
            "voidFactor": 0
          }
        ]
      }
    ],
    "rollbacks": []
  }
*/
type Result struct {
	ReceivedAt    int64                     `json:"receivedDate"`
	RunnerResults []sportradar.RunnerResult `json:"runners"`
}

type Rollback struct {
	NamedValueId          int    `json:"namedValueId"`
	NamedValueDescription string `json:"namedValueDescription"`
	RollbackType          string `json:"rollbackType"`
	RollbackReason        string `json:"rollbackReason"`
	ReceivedAt            int64  `json:"receivedDate"`
	StartTime             int64  `json:"startTime"`
	EndTime               int64  `json:"endTime"`
}

type MarketResult struct {
	// Id            string                    `json:"id"`
	CreatedAt    int64      `json:"createdDate"`
	UpdatedAt    int64      `json:"updatedAt"`
	SportId      string     `json:"sportId"`
	EventId      string     `json:"eventId"`
	MarketId     string     `json:"marketId"`
	MarketName   string     `json:"marketName"`
	MarketType   string     `json:"marketType"`
	MarketStatus string     `json:"marketStatus"`
	Results      []Result   `json:"results"`
	Rollbacks    []Rollback `json:"rollbacks"`
}
type GetMarketResultsResp struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	MarketResults    []MarketResult `josn:"data"`
}
