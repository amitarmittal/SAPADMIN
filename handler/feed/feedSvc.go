package feedsvc

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"Sp/handler"
	utils "Sp/utilities"
	"log"

	"github.com/gofiber/fiber/v2"
)

var (
	ERROR_STATUS = "RS_ERROR"
	OK_STATUS    = "RS_OK"
)

// Add Market API
// @Summary      Add Market API
// @Description  To notify SAP to add Market in the system if not present. This helps to maintain the market status
// @Tags         Feed-Service
// @Accept       json
// @Produce      json
// @Param        AddMarket  body      requestdto.AddMarket  true  "AddMarket model is used"
// @Success      200        {object}  responsedto.DefaultRespDto
// @Failure      503        {object}  responsedto.DefaultRespDto
// @Router       /feed/add-market [post]
func AddMarket(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = ERROR_STATUS
	bodyStr := string(c.Body())
	log.Println("AddMarket: request body is - ", bodyStr)
	// 2. Parse request body to Request Object
	reqDto := requestdto.AddMarket{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("AddMarket: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Request Check

	// 4. Get Operator Details
	operatorId := reqDto.OperatorId
	operatorDto, err := cache.GetOperatorDetails(operatorId)
	if err != nil {
		log.Println("AddMarket: Failed to get Operator Details: ", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if operatorDto.Status != "ACTIVE" {
		log.Println("AddMarket: Operator account was not Active: ", operatorDto.Status)
		respDto.ErrorDescription = "Something went wrong. Please contact your Provider.!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	if operatorDto.WalletType != constants.SAP.WalletType.Feed() {
		log.Println("AddMarket: Operator wallet type is not feed: ", operatorDto.WalletType)
		respDto.ErrorDescription = "Operation not permitted!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 5. Verify Signature
	if operatorDto.Signature == true {
		signature := c.Request().Header.Peek("Signature")
		//log.Println("Signature", string(signature))
		pubKey, err := utils.ParseRsaPublicKeyFromPemStr(operatorDto.Keys.OperatorKey)
		if err != nil {
			log.Println("AddMarket: Parsing public key failed: ", err.Error())
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		signValid := utils.VerifySignature(string(signature), string(c.Body()), *pubKey)
		if !signValid {
			log.Println("AddMarket: Signature verification failed : ")
			respDto.ErrorDescription = "Bad Request.!"
			return c.Status(fiber.StatusBadRequest).JSON(respDto)
		}
	}
	// 6. AddMarket in DB
	market, err := cache.GetMarket(reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId)
	if err != nil {
		log.Println("AddMarket: cache.GetMarket failed with error - ", err.Error())
		if err.Error() == "Market NOT FOUND!" {
			// insert market
			err = handler.InsertMarket(reqDto.OperatorId, reqDto.ProviderId, reqDto.SportId, reqDto.EventId, reqDto.MarketId, reqDto.MarketType)
			if err != nil {
				log.Println("AddMarket: InsertMarket failed with error - ", err.Error())
				respDto.ErrorDescription = err.Error()
				return c.Status(fiber.StatusOK).JSON(respDto)
			}
			respDto.Status = OK_STATUS
			respDto.ErrorDescription = ""
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	} else {
		log.Println("AddMarket: market already in system for marketKey - ", market.MarketKey)
	}
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}
