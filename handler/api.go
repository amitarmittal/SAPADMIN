package handler

import (
	"github.com/gofiber/fiber/v2"
)

// Hello hanlde api status
func Hello(c *fiber.Ctx) error {
	/*
		newRQ := sports.ResultQueueDto{}
		newRQ.EventKey = "Test EventKey 6"
		newRQ.MarketId = "Test Market ID"
		newRQ.MarketName = "Test Market Name"
		newRQ.Status = "in-queue"
		result, err := database.InsertResultsQueue(newRQ)
		if err != nil {
			log.Println("Hello: insert failed wtih error - ", err.Error())
			return c.JSON(fiber.Map{"status": "error", "message": "Hello i'm ok! E1", "data": nil})
		}
		log.Println("Hello: Successfully inserted the record. Record _id is - ", result.InsertedID)
		objId := result.InsertedID.(primitive.ObjectID)
		log.Println("Hello: objId is - ", objId)
		getRQ, err := database.GetResultsQueueById(objId)
		if err != nil {
			log.Println("Hello: get failed wtih error - ", err.Error())
			return c.JSON(fiber.Map{"status": "error", "message": "Hello i'm ok! E2", "data": nil})
		}
		log.Println("Hello: getRq _id is - ", getRQ.ID)
		log.Println("Hello: getRq status is - ", getRQ.Status)
		log.Println("Hello: getRq eventkey is - ", getRQ.EventKey)
		getRQ.Status = "in-progress"
		result2, err := database.UpdateResultsQueueStatus(getRQ.ID, getRQ.Status)
		if err != nil {
			log.Println("Hello: get failed wtih error - ", err.Error())
			return c.JSON(fiber.Map{"status": "error", "message": "Hello i'm ok! E3", "data": nil})
		}
		log.Println("Hello: result2 modified count is - ", result2.ModifiedCount)
	*/
	return c.JSON(fiber.Map{"status": "success", "message": "Hello i'm ok!", "data": nil})
}
