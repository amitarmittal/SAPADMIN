package cache

import "github.com/dgraph-io/ristretto"

var passwordCache *ristretto.Cache

func init() {
	passwordCache, _ = InitializeCache(1024, 1<<12, 64)
}

// Set PartnerStatus by OperatorId & PartnerId
func SetPasswordHash(password, hash string) {
	passwordCache.Set(password, hash, 64)
	passwordCache.Wait()
}

func GetPasswordHash(password string) (string, bool) {
	value, found := passwordCache.Get(password)
	if found {
		return value.(string), true
	}
	return "", false
}
