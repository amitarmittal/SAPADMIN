package operator

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	operatordto "Sp/dto/operator"
	"log"
)

func SyncWallets(userKeysRateMap map[string]int32) {
	log.Println("SyncWallets: START!!!")
	// create userKeys list
	userKeys := make([]string, 0, len(userKeysRateMap))
	for k := range userKeysRateMap {
		userKeys = append(userKeys, k)
	}
	// Read UserBalances from database
	operatorsMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println("SyncWallets: cache.GetObjectMap failed with error: ", err.Error())
		return
	}
	users, err := database.GetB2BUsersByKeys(userKeys)
	if err != nil {
		log.Println("SyncWallets: database.GetB2BUsersByKeys failed with error: ", err.Error())
		return
	}
	for _, user := range users {
		opObject := operatorsMap[user.OperatorId]
		operator := opObject.(operatordto.OperatorDTO)
		balance := user.Balance / float64(userKeysRateMap[user.UserKey])
		respObj, err := WalletSync(user.UserId, balance, operator.BaseURL, operator.Keys.PrivateKey)
		if err != nil {
			log.Println("SyncWallets: WalletSync failed with error: ", err.Error())
			continue
		}
		if respObj.Status != "RS_OK" {
			log.Println("SyncWallets: WalletSync failed with Status: ", respObj.Status)
		}
	}
	log.Println("SyncWallets: END!!!")
	return
}
