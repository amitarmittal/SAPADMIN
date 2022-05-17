package main

import (
	"Sp/database"
	_ "Sp/docs"
	"Sp/router"
	"Sp/routines"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/robfig/cron"
)

func loggerFunction() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	//if err != nil {
	//	log.Println(err.Error())
	//}
	//log.SetOutput(file)
	log.SetOutput(os.Stdout)
}

// @title        SAP
// @version      1.0
// @description  This is an API Documentation for Sports Aggregation Platform (SAP)

// @contact.name   Amit
// @contact.email  amit.m@outlook.com

// @BasePath  /api/v1
func main() {
	loggerFunction()
	app := fiber.New()
	app.Use(cors.New())
	app.Use(pprof.New())
	database.ConnectDB()
	routines.Initialization()
	c := cron.New()
	c.AddFunc("40 0-59/1 * * * *", func() { routines.RefreshCache() }) // runs for every minute, 40th Second
	c.Start()
	// Setup Routes
	router.SetupRoutes(app)
	// Start app
	log.Fatal(app.Listen(":3000"))
	defer database.DBWrite.Disconnect(database.Ctx)
}
