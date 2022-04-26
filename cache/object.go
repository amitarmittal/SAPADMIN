package cache

import (
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto"
)

// Key 		- ObjectType Ex. Operator, Provider, Provider-Sports etc...
// Value 	- ObjectMap (map[ObjectIds]interface{})

var objectCache *ristretto.Cache

func init() {
	objectCache, _ = InitializeCache(1000, 1<<30, 64)
}

// Save Object in Cache
func AddObject(objectType string, objectId string, object interface{}) {
	// 1. Get ObjectMap using objectType
	value, found := objectCache.Get(objectType)
	if found {
		// 1.1 ObjectMap FOUND in cache, add/update objectId and object
		objectMap := value.(map[string]interface{})
		objectMap[objectId] = object
		objectCache.Set(objectType, objectMap, 0)
		objectCache.Wait()
		return
	}
	log.Println("objectCache: ObjectType NOT FOUND in Cache - " + objectType)
	// 2. ObjectMap NOT FOUND in cache, create objectMap and add to cache
	objectMap := make(map[string]interface{})
	objectMap[objectId] = object
	objectCache.Set(objectType, objectMap, 0)
	objectCache.Wait()
	return
}

// Save ObjectMap in Cache
func SetObjectMap(objectType string, objectMap map[string]interface{}) {
	objectCache.Set(objectType, objectMap, 0)
	objectCache.Wait()
	return
}

// Get Object from Cache
func GetObject(objectType string, objectId string) (interface{}, error) {
	// 1. Get ObjectMap using objectType
	value, found := objectCache.Get(objectType)
	if found {
		// ObjectMap FOUND in cache, find object using objectId
		objectMap := value.(map[string]interface{})
		object, found := objectMap[objectId]
		if found {
			// Object FOUND in cache, return object
			return object, nil
		}
		// Object NOT FOUND in cache, return error
		log.Println("objectCache: OBJECT-ID NOT FOUND for objectType - objectId" + objectType + "-" + objectId)
		return nil, fmt.Errorf("OBJECT-ID NOT FOUND for objectType - objectId" + objectType + "-" + objectId)
	}
	// ObjectMap NOT FOUND in cache, return error
	log.Println("objectCache: OBJECT-TYPE NOT FOUND for objectType - objectId" + objectType + "-" + objectId)
	return nil, fmt.Errorf("OBJECT-TYPE NOT FOUND for objectType - objectId" + objectType + "-" + objectId)
}

// Get ObjectMap from Cache
func GetObjectMap(objectType string) (map[string]interface{}, error) {
	// 1. Get ObhectMap using objectType
	value, found := objectCache.Get(objectType)
	if found {
		// ObjectMap FOUND in cache, return ObjectMap
		objectMap := value.(map[string]interface{})
		return objectMap, nil
	}
	// ObjectMap NOT FOUND in cache, return error
	log.Println("objectCache: OBJECT-TYPE NOT FOUND for objectType - " + objectType)
	return nil, fmt.Errorf("OBJECT-TYPE NOT FOUND for objectType - " + objectType)
}
