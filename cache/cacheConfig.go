package cache

import (
	"log"

	"github.com/dgraph-io/ristretto"
)

func InitializeCache(numCounters int64, maxCost int64, numKeys int64) (*ristretto.Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: numCounters, // number of keys to track frequency of (10M).
		MaxCost:     maxCost,     // maximum cost of cache (1GB).
		BufferItems: numKeys,     // number of keys per Get buffer.
		Metrics:     true,
	})
}

func DumpMetrics() {
	log.Println("DumpMetrics: Metrics Provider - ", providerCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Operator - ", operatorCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Partner - ", partnerStatusCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Sport - ", sportCache.Metrics.String())
	log.Println("DumpMetrics: Metrics SportStatus - ", sportStatusCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Compitition - ", competitionCache.Metrics.String())
	log.Println("DumpMetrics: Metrics CompititionStatus - ", competitionStatusCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Event - ", eventCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Event Status - ", eventStatusCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Market - ", marketCache.Metrics.String())
	log.Println("DumpMetrics: Metrics Market Status - ", marketStatusCache.Metrics.String())
}
