package cachesvc

import (
	"Sp/cache"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

var (
	TestOperator string = "TestOperator"
	TestProvider string = "Dream"
	TestToken    string = "f23004ee-2bf4-49b9-a98b-f0189520b795"
	// BaseUrl      string = "https://stage.mysportsfeed.io"
	BaseUrl  string = os.Getenv("IFRAME_URL")
	LobbyUrl string = "%s/auth?token=%s&operatorId=%s&partnerId=%s&providerId=%s"
)

// GetCache API
// @Summary      Get Cache Data
// @Description  Endpoint to read cache data
// @Tags         Cache-Service
// @Accept       json
// @Produce      json
// @Param        Request  body      requestdto.GetCacheReq  true  "GetCacheReq model is used"
// @Success      200      {object}  responsedto.GetCacheResp
// @Failure      503      {object}  responsedto.GetCacheResp
// @Router       /cache/get-cache [post]
func GetCache(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.GetCacheResp{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	respDto.RespJson = ""
	// 2. Parse request body to Request Object
	bodyStr := string(c.Body())
	log.Println("GetCache: Auth Req. Body is - ", bodyStr)
	reqDto := requestdto.GetCacheReq{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("GetCache: Body Parsing failed")
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	switch reqDto.CacheType {
	case "SportStatus":
		respDto.RespJson, err = cache.GetSportStatusByKey(reqDto.Key)
		if err != nil {
			log.Println("GetCache: cache.GetSportStatusByKey failed with error - ", err.Error())
			respDto.ErrorDescription = err.Error()
			return c.Status(fiber.StatusOK).JSON(respDto)
		}
	default:
		log.Println("GetCache: INVALID KEY - ", reqDto.Key)
		respDto.ErrorDescription = "INVALID KEY"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}
