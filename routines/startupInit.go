package routines

import (
	"log"
	"time"
)

func Initialization() {
	log.Println("Initialization: Cur Time is: ", time.Now())
	InitCache()
}
