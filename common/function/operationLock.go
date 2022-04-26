package function

import (
	"Sp/constants"
	"Sp/database"
	"log"
	"time"
)

func AquireLock(operationKey string) bool {
	logkey := "OperationLock: AquireLock: " + operationKey + ": "
	log.Println(logkey + "START!!!")
	// 1. Get OperationLock Documnet
	opLock, err := database.GetOperationLock(operationKey)
	if err != nil {
		log.Println(logkey+"database.GetOperationLock failed with error - ", err.Error())
		log.Println(logkey + "FAILED END 1 !!!")
		return false
	}
	// 2. If Lock is free & beyond grace period, aquire lock
	if opLock.Status == constants.SAP.OperationStatus.FREE() {
		// 2.1. Grace Period Check
		var gracePeriod int = 30 * 1 // 30 seconds
		switch operationKey {
		case constants.SAP.OperationKeys.BETFAIR_KEEPALIVE():
			gracePeriod = 55 * 60 * 1 // 55 minutes
		default:
			log.Println(logkey+"GracePeriod not defined for key - ", operationKey)
		}
		// 3.2. Get Last Updated Time
		updatedAt, err := time.Parse(time.RFC3339Nano, opLock.UpdatedAt)
		if err != nil {
			log.Println(logkey+"time.Parse failed with error: ", err.Error())
			log.Println(logkey + "FAILED END 2 !!!")
			return false
		}
		// 3.3. Add graceperiod to the last updated time
		updatedAt2 := updatedAt.Add(time.Duration(gracePeriod) * time.Second)
		curTime := time.Now()
		// log.Println(logkey+"Update Time 1: updated at:  - ", updatedAt.Format(time.RFC3339Nano))
		// log.Println(logkey+"Update Time 2: updated at2  - ", updatedAt2.Format(time.RFC3339Nano))
		// log.Println(logkey+"Update Time 3: current time - ", curTime.Format(time.RFC3339Nano))
		// 3.4. Check if current time is greater than idletimeout
		if curTime.After(updatedAt2) == true {
			// 3.4.1. Aquire Locl
			log.Println(logkey + "Acquiring Lock...")
			// TODO: Aquire Lock and set updatedAt and return true
			err = database.UpdateOperationLockStatus(operationKey, constants.SAP.OperationStatus.BUSY())
			if err != nil {
				log.Println(logkey+"database.UpdateOperationLockStatus failed with error: ", err.Error())
				log.Println(logkey + "FAILED END 3 !!!")
				return false
			}
			log.Println(logkey + "SUCCESS END 4 !!!")
			return true
		}
		log.Println(logkey+"WITHIN THE GRACE PERIOD - FAILED END 5 !!!", gracePeriod)
		return false
	}
	// 3. If Lock is busy
	if opLock.Status == constants.SAP.OperationStatus.BUSY() {
		// 3.1. Know idleTimeOut
		var idleTimeOut int = 60 * 60 * 1 // 60 minutes
		switch operationKey {
		case constants.SAP.OperationKeys.BETFAIR_KEEPALIVE():
			idleTimeOut = 2 * 60 * 60 * 1 // 2 hours
		case constants.SAP.OperationKeys.BETFAIR_HOLDBETS_SETTLEMENT():
			idleTimeOut = 30 * 60 * 1 // 30 minutes
		case constants.SAP.OperationKeys.BETFAIR_CLEAREDORDERS_SETTLEMENT():
			idleTimeOut = 30 * 60 * 1 // 30 minutes
		default:
			log.Println(logkey+"IdelTimeOut not defined for key - ", operationKey)
		}
		// 3.2. Get Last Updated Time
		updatedAt, err := time.Parse(time.RFC3339Nano, opLock.UpdatedAt)
		if err != nil {
			log.Println(logkey+"time.Parse failed with error: ", err.Error())
			log.Println(logkey + "FAILED END 6 !!!")
			return false
		}
		// 3.3. Add idletimeout to the last updated time
		updatedAt2 := updatedAt.Add(time.Duration(idleTimeOut) * time.Second)
		curTime := time.Now()
		// log.Println(logkey+"Update Time 1: updated at:  - ", updatedAt.Format(time.RFC3339Nano))
		// log.Println(logkey+"Update Time 2: updated at2  - ", updatedAt2.Format(time.RFC3339Nano))
		// log.Println(logkey+"Update Time 3: current time - ", curTime.Format(time.RFC3339Nano))
		// 3.4. Check if current time is greater than idletimeout
		if curTime.After(updatedAt2) == true {
			// 3.4.1. Aquire Locl
			log.Println(logkey+"Lock stuck with Busy for more than IdleTimeOut!!!", idleTimeOut)
			// TODO: Aquire Lock and set updatedAt and return true
			err = database.UpdateOperationLockStatus(operationKey, constants.SAP.OperationStatus.BUSY())
			if err != nil {
				log.Println(logkey+"database.UpdateOperationLockStatus failed with error: ", err.Error())
				log.Println(logkey + "FAILED END 7 !!!")
				return false
			}
			log.Println(logkey + "SUCCESS END 8 !!!")
			return true
		}
		// 3.5. Lock is holding by the another instance within the ideltimeout
		log.Println(logkey+"WITHIN THE IDLE TIMEOUT - FAILED END 9 !!!", idleTimeOut)
		return false
	}
	// 4. Lock is neither FREE nor BUSY. Means, HOLD. Return false
	log.Println(logkey+"opLock.Status is not FREE. The current status is - ", opLock.Status)
	log.Println(logkey + "FAILED END 10 !!!")
	return false
}

func ReleaseLock(operationKey string) {
	logkey := "OperationLock: ReleaseLock: " + operationKey + ": "
	err := database.UpdateOperationLockStatus(operationKey, constants.SAP.OperationStatus.FREE())
	if err != nil {
		log.Println(logkey+"database.UpdateOperationLockStatus failed with error: ", err.Error())
	}
}
