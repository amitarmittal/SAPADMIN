package function

import (
	"Sp/cache"
	"Sp/dto/responsedto"
	"log"
)

// Get Active Providers by OperatorId & PartnerID
// User cases:
// feed service/list-providers
// core service/get-providers
func GetOpActiveProviders(operatorId string, partnerId string) ([]responsedto.ProviderDto, error) {
	// 0. create default return object
	providerDtos := []responsedto.ProviderDto{}
	// 1. Get ProviderStatus by OperatorId from cahce
	mapPS, err := cache.GetOpPartnerStatus(operatorId, partnerId)
	if err != nil {
		// 1.1. Return Error
		log.Println("GetOpActiveProviders: Failed with error - ", err.Error())
		return providerDtos, err
	}
	// 2. Iterate throuhg map and add only active providers
	for _, ps := range mapPS {
		if ps.ProviderStatus == "ACTIVE" && ps.OperatorStatus == "ACTIVE" {
			pi := responsedto.ProviderDto{}
			pi.ProviderId = ps.ProviderId
			pi.ProviderName = ps.ProviderName
			pi.Status = "ACTIVE"
			providerDtos = append(providerDtos, pi)
		}
	}
	// 3. Return results
	return providerDtos, nil
}
