package cache

import (
	"Sp/database"
	"Sp/dto/models"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

var providerCache *ristretto.Cache

func init() {
	providerCache, _ = InitializeCache(1024, 1<<12, 64)
}

// Save Provider in Cache 
func SetProvider(provider models.Provider) {
	// 1. Update it in Cache
	providerCache.Set(provider.ProviderId, provider, 0)
	providerCache.Wait()
	_, found := providerCache.Get(provider.ProviderId)
	if found {
		//log.Println("providerCache: Provider SAVE is SUCCESS in Cache - " + provider.ProviderId)
		return
	}
	log.Println("providerCache: Provider NOT SAVED in Cache - " + provider.ProviderId)
}

// Get Provider from Cache
func GetProvider(providerId string) (models.Provider, error) {
	// 0. Create resp object
	provider := models.Provider{}
	// 1. Get from Cache
	value, found := providerCache.Get(providerId)
	if found {
		// 1.1 FOUND in cache, retrun object
		provider = value.(models.Provider)
		return provider, nil
	}
	log.Println("providerCache: Provider NOT FOUND in Cache - " + providerId)
	// 2. NOT FOUND in cache, get from DB and update to cache
	provider, err := database.GetProvider(providerId)
	if err != nil {
		// 2.1 NOT FOUND in DB, return error
		log.Println("providerCache: Provider NOT FOUND in DB - ", err.Error())
		return provider, fmt.Errorf("Provider NOT FOUND!")
	}
	// 3. FOUND in DB, add to Cache
	SetProvider(provider)
	// 4. return object
	return provider, nil
}

// Get Provider from Cache
func GetProviderFromCache(providerId string) (models.Provider, error) {
	// 0. Create resp object
	provider := models.Provider{}
	// 1. Get from Cache
	value, found := providerCache.Get(providerId)
	if found {
		// 1.1 FOUND in cache, retrun object
		provider = value.(models.Provider)
		return provider, nil
	}
	return provider, fmt.Errorf("Provider NOT FOUND!")
}

func GetProviderCacheMetrics() {
	log.Println("providerCache: Metrics - ", providerCache.Metrics.String())
}
