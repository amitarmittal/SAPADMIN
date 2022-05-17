package routines

import (
	"Sp/cache"
	"Sp/providers"
	"log"
	"time"
)

func InitCache() {
	log.Println("InitCache: Start Time is: ", time.Now())
	// 0. Init Operators
	providers.InitOperatorCache()
	// 1. Init Providers
	providers.InitProviderCache()
	// // 1.1. Init PartnerStatus
	providers.InitPartnerStatusCache()
	// // 2. Init Sports
	providers.InitSportCache()
	// // 2.1. Init SportStatus
	providers.InitSportStatusCache()
	// 3. Init Competitions
	// providers.InitCompetitionCache()
	// // 3.1. Init CompetitionStatus
	// providers.InitCompetitionStatusCache()
	// // 4. Init Events
	// providers.InitEventCache()
	// // 4.1. Init EventStatus
	// providers.InitEventStatusCache()
	// // 5. Init Markets
	// providers.InitMarketCache()
	// // 5.1. Init MarketStatus
	// providers.InitMarketStatusCache()
	cache.DumpMetrics()
	log.Println("InitCache: End Time is: ", time.Now())
}

// runs for every miute at 40th second
func RefreshCache() {
	log.Println("RefreshCache: Start Time is: ", time.Now())
	cache.DumpMetrics()
	// 0. Refresh Operators
	providers.RefreshOperatorCache()
	// 1. Refresh Providers
	providers.RefreshProviderCache()
	// 1.1. Refresh PartnerStatus
	providers.RefreshPartnerStatusCache()
	// 2. Refresh Sports
	providers.RefreshSportCache()
	// 2.1. Refresh SportStatus
	providers.RefreshSportStatusCache()
	// 3. Refresh Competitions
	// providers.RefreshCompetitionCache()
	// // 3.1. Refresh CompetitionStatus
	// providers.RefreshCompetitionStatusCache()
	// // 4. Refresh Events
	// providers.RefreshEventCache()
	// // 4.1. Refresh EventStatus
	// providers.RefreshEventStatusCache()
	// // 5. Refresh Markets
	// providers.RefreshMarketCache()
	// // 5.1. Refresh MarketStatus
	// providers.RefreshMarketStatusCache()
	log.Println("RefreshCache: End Time is: ", time.Now())
}
