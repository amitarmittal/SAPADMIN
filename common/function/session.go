package function

import (
	"Sp/cache"
	sessdto "Sp/dto/session"
	"fmt"
	"log"
)

func GetSession(token string) (sessdto.B2BSessionDto, error) {
	// 1. Get Session Details
	sessionDto, err := cache.GetSessionDetailsByToken(token)
	if err != nil {
		// 1.1. Return Session NOT FOUND Error
		log.Println("GetSession: Session NOT FOUND - for token ", err.Error(), token)
		return sessdto.B2BSessionDto{}, fmt.Errorf("Session NOT FOUND!")
	}
	// 2. Check Session Validity
	if !cache.IsSessionValid(sessionDto) {
		// 2.1. Return Session EXPIRED Error
		log.Println("GetSession: Session EXPIRED!")
		return sessionDto, fmt.Errorf("Session EXPIRED!")
	}
	// 3. Extend the Session Validity
	sessionDto = cache.ExtendValidityIfRequired(sessionDto)
	return sessionDto, nil
}
