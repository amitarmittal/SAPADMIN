package betfairmodule

import (
	"encoding/json"
	"log"
	"time"
)

type MetricSummary struct {
	Request       string `json:"request"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Count         int    `json:"count"`
	MinRespTime   int64  `json:"minRespTime"`
	MaxRespTime   int64  `json:"maxRespTime"`
	AvgRespTime   int64  `json:"avgRespTime"`
	TotalRespTime int64  `json:"totalRespTime"`
}

type ResponseDetails struct {
	StartTime     string `json:"startTime"`
	ExecutionTime int64  `json:"endTime"`
	Description   string `json:"description"`
}
type BetFairMetrics struct {
	Request   string `json:"request"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	// Count         int               `json:"count"`
	// MinRespTime   int64             `json:"minRespTime"`
	// MaxRespTime   int64             `json:"maxRespTime"`
	// AvgRespTime   int64             `json:"avgRespTime"`
	// TotalRespTime int64             `json:"totalRespTime"`
	Details []ResponseDetails `json:"details"`
}

func (ms MetricSummary) ToString() string {
	jsonBytes, err := json.Marshal(ms)
	if err != nil {
		return err.Error()
	}
	return string(jsonBytes)
}

func InitBetFairMetrics() {
	// PlaceOrders
	PlOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	PlOMetrics.Details = []ResponseDetails{}
	// CurrentOrders
	CuOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	CuOMetrics.Details = []ResponseDetails{}
	// ClearedOrders
	ClOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	ClOMetrics.Details = []ResponseDetails{}
	// CancelOrders
	CaOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	CaOMetrics.Details = []ResponseDetails{}
}

func DumpMetrics() {
	logkey := "BetFairModule: DumpMetrics: "
	// PlaceOrders
	ploMetrics := BetFairMetrics{Request: PLACE_ORDER}
	ploMetrics.StartTime = PlOMetrics.StartTime
	ploMetrics.EndTime = time.Now().Format(time.RFC3339Nano)
	ploMetrics.Details = []ResponseDetails{}
	ploMetrics.Details = append(ploMetrics.Details, PlOMetrics.Details...)
	PlOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	PlOMetrics.Details = []ResponseDetails{}
	// CurrentOrders
	cuoMetrics := BetFairMetrics{Request: LIST_CURRENT_ORDERS}
	cuoMetrics.StartTime = CuOMetrics.StartTime
	cuoMetrics.EndTime = time.Now().Format(time.RFC3339Nano)
	cuoMetrics.Details = []ResponseDetails{}
	cuoMetrics.Details = append(cuoMetrics.Details, CuOMetrics.Details...)
	CuOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	CuOMetrics.Details = []ResponseDetails{}
	// ClearedOrders
	cloMetrics := BetFairMetrics{Request: LIST_CLEARED_ORDER}
	cloMetrics.StartTime = ClOMetrics.StartTime
	cloMetrics.EndTime = time.Now().Format(time.RFC3339Nano)
	cloMetrics.Details = []ResponseDetails{}
	cloMetrics.Details = append(cloMetrics.Details, ClOMetrics.Details...)
	ClOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	ClOMetrics.Details = []ResponseDetails{}
	// CancelOrders
	caoMetrics := BetFairMetrics{Request: CANCEL_ORDERS}
	caoMetrics.StartTime = CaOMetrics.StartTime
	caoMetrics.EndTime = time.Now().Format(time.RFC3339Nano)
	caoMetrics.Details = []ResponseDetails{}
	caoMetrics.Details = append(caoMetrics.Details, CaOMetrics.Details...)
	CaOMetrics.StartTime = time.Now().Format(time.RFC3339Nano)
	CaOMetrics.Details = []ResponseDetails{}
	// Get Summaries
	ploSummary := GetSummary(ploMetrics)
	log.Println(logkey+"PLACEORDERS - ", ploSummary.ToString())
	cuoSummary := GetSummary(cuoMetrics)
	log.Println(logkey+"CURRENTORDERS - ", cuoSummary.ToString())
	cloSummary := GetSummary(cloMetrics)
	log.Println(logkey+"CLEAREDORDERS - ", cloSummary.ToString())
	caoSummary := GetSummary(caoMetrics)
	log.Println(logkey+"CANCELORDERS - ", caoSummary.ToString())
}

func GetMetrics() {
	logkey := "BetFairModule: GetMetrics: "
	// Get Summaries
	ploSummary := GetSummary(PlOMetrics)
	log.Println(logkey+"PLACEORDERS - ", ploSummary.ToString())
	cuoSummary := GetSummary(CuOMetrics)
	log.Println(logkey+"CURRENTORDERS - ", cuoSummary.ToString())
	cloSummary := GetSummary(ClOMetrics)
	log.Println(logkey+"CLEAREDORDERS - ", cloSummary.ToString())
	caoSummary := GetSummary(CaOMetrics)
	log.Println(logkey+"CANCELORDERS - ", caoSummary.ToString())
}

func GetSummary(bfMetrics BetFairMetrics) MetricSummary {
	metricSummary := MetricSummary{}
	metricSummary.Request = bfMetrics.Request
	metricSummary.StartTime = bfMetrics.StartTime
	metricSummary.EndTime = bfMetrics.EndTime
	metricSummary.Count = len(bfMetrics.Details)
	if metricSummary.Count == 0 {
		return metricSummary
	}
	metricSummary.MinRespTime = bfMetrics.Details[0].ExecutionTime
	metricSummary.MaxRespTime = bfMetrics.Details[0].ExecutionTime
	for _, rd := range bfMetrics.Details {
		if rd.ExecutionTime < metricSummary.MinRespTime {
			metricSummary.MinRespTime = rd.ExecutionTime
		} else if rd.ExecutionTime > metricSummary.MaxRespTime {
			metricSummary.MaxRespTime = rd.ExecutionTime
		}
		metricSummary.TotalRespTime += rd.ExecutionTime
	}
	metricSummary.AvgRespTime = metricSummary.TotalRespTime / int64(metricSummary.Count)
	return metricSummary
}
