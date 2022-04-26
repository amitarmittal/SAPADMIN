package testsvc

import (
	"Sp/cache"
	"Sp/database"
	"Sp/dto/requestdto"
	"Sp/dto/responsedto"
	betfairmodule "Sp/providers/betfairModule"
	"Sp/providers/betfairModule/request"
	"Sp/providers/betfairModule/response"
	dreamsvc "Sp/providers/dream"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// @Summary      AverageLoad is a function to test system with 50ms average response time
// @Description  Will respond back after 50ms sleep.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        AverageLoad  body      requestdto.DefaultReqDto  true  "DefaultReqDto model is used"
// @Success      200           {object}  responsedto.DefaultRespDto{}
// @Failure      503           {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/average-load [post]
func AverageLoad(c *fiber.Ctx) error {
	log.Println("LoadTest: AverageLoad: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.DefaultReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: AverageLoad: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Parse request body to Request Object
	time.Sleep(50 * time.Millisecond)
	log.Println("LoadTest: AverageLoad: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      CacheRead is a function to test system with cache read
// @Description  Will respond back after cache read.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        CacheRead  body      requestdto.DefaultReqDto  true  "DefaultReqDto model is used"
// @Success      200        {object}  responsedto.DefaultRespDto{}
// @Failure      503        {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/cache-read [post]
func CacheRead(c *fiber.Ctx) error {
	log.Println("LoadTest: CacheRead: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.DefaultReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: CacheRead: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Get Operator DTO
	_, err = cache.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("LoadTest: CacheRead: cache.GetOperatorDetails failed with error - ", err.Error())
		log.Println("LoadTest: CacheRead: cache.GetOperatorDetails failed for operator - ", reqDto.OperatorId)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: CacheRead: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      DatabaseRead is a function to test system with database read
// @Description  Will respond back after database read.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        DatabaseRead  body      requestdto.DefaultReqDto  true  "DefaultReqDto model is used"
// @Success      200    {object}  responsedto.DefaultRespDto{}
// @Failure      503    {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/database-read [post]
func DatabaseRead(c *fiber.Ctx) error {
	log.Println("LoadTest: DatabaseRead: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.DefaultReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: DatabaseRead: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Get Operator DTO
	_, err = database.GetOperatorDetails(reqDto.OperatorId)
	if err != nil {
		// 4.1. Return Error
		log.Println("LoadTest: DatabaseRead: database.GetOperatorDetails failed with error - ", err.Error())
		log.Println("LoadTest: DatabaseRead: database.GetOperatorDetails failed for operator - ", reqDto.OperatorId)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: DatabaseRead: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      DatabaseWrite is a function to test system with database write
// @Description  Will respond back after database write.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        DatabaseWrite  body      requestdto.DatabaseWriteReqDto  true  "DatabaseWriteReqDto model is used"
// @Success      200            {object}  responsedto.DefaultRespDto{}
// @Failure      503            {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/database-write [post]
func DatabaseWrite(c *fiber.Ctx) error {
	log.Println("LoadTest: DatabaseWrite: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.DatabaseWriteReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: DatabaseWrite: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Check for input fields
	operatorId := reqDto.OperatorId
	if operatorId == "" {
		log.Println("LoadTest: DatabaseWrite: operatorId cannot be null!!!")
		respDto.ErrorDescription = "operatorId cannot be null!!!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	message := reqDto.Message
	if message == "" {
		message = "EMPTY MESSAGE"
	}
	// 4. Update in database
	err = database.UpdateLoadTest(operatorId, message)
	if err != nil {
		// 4.1. Return Error
		log.Println("LoadTest: DatabaseWrite: database.UpdateLoadTest failed with error - ", err.Error())
		log.Println("LoadTest: DatabaseWrite: database.UpdateLoadTest failed for operator - ", reqDto.OperatorId)
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: DatabaseWrite: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      EndToEnd is a function to test system with L3-L2-L1 latency
// @Description  Will respond back getting response from L1->L2.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Success      200  {object}  responsedto.DefaultRespDto{}
// @Failure      503  {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/end-to-end [post]
func EndToEnd(c *fiber.Ctx) error {
	log.Println("LoadTest: EndToEnd: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	// 3. Parse request body to Request Object
	resp, err := dreamsvc.EndToEnd()
	if err != nil {
		log.Println("LoadTest: EndToEnd: dreamsvc.EndToEnd failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if resp.Status != "Success" {
		log.Println("LoadTest: EndToEnd: dreamsvc.EndToEnd failed with Status - ", resp.Status)
		log.Println("LoadTest: EndToEnd: dreamsvc.EndToEnd failed with Error - ", resp.ErrorDescription)
		respDto.Status = resp.Status
		respDto.ErrorDescription = resp.ErrorDescription
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: EndToEnd: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      LayerTwo is a function to test system with L3-L2 latency
// @Description  Will respond back getting response from L2.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Success      200  {object}  responsedto.DefaultRespDto{}
// @Failure      503  {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/layer-two [post]
func LayerTwo(c *fiber.Ctx) error {
	log.Println("LoadTest: LayerTwo: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	// 3. Parse request body to Request Object
	resp, err := dreamsvc.LayerTwo()
	if err != nil {
		log.Println("LoadTest: LayerTwo: dreamsvc.LayerTwo failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if resp.Status != "Success" {
		log.Println("LoadTest: LayerTwo: dreamsvc.LayerTwo failed with Status - ", resp.Status)
		log.Println("LoadTest: LayerTwo: dreamsvc.LayerTwo failed with Error - ", resp.ErrorDescription)
		respDto.Status = resp.Status
		respDto.ErrorDescription = resp.ErrorDescription
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: LayerTwo: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      LayerOneBetFair is a function to test system with L3-L1 latency
// @Description  Will respond back getting response from L1.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Success      200  {object}  responsedto.DefaultRespDto{}
// @Failure      503  {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/layer-one-betFair [post]
func LayerOneBetFair(c *fiber.Ctx) error {
	log.Println("LoadTest: LayerOneBetFair: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	// 3. Parse request body to Request Object
	resp, err := dreamsvc.LayerOneBetFair()
	if err != nil {
		log.Println("LoadTest: LayerOneBetFair: dreamsvc.LayerOneBetFair failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if resp.Status != "Success" {
		log.Println("LoadTest: LayerOneBetFair: dreamsvc.LayerOneBetFair failed with Status - ", resp.Status)
		log.Println("LoadTest: LayerOneBetFair: dreamsvc.LayerOneBetFair failed with Error - ", resp.ErrorDescription)
		respDto.Status = resp.Status
		respDto.ErrorDescription = resp.ErrorDescription
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: LayerOneBetFair: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      LayerOneDream is a function to test system with L3-L1 latency
// @Description  Will respond back getting response from L1.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Success      200  {object}  responsedto.DefaultRespDto{}
// @Failure      503  {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/layer-one-dream [post]
func LayerOneDream(c *fiber.Ctx) error {
	log.Println("LoadTest: LayerOneDream: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	// 3. Parse request body to Request Object
	resp, err := dreamsvc.LayerOneDream()
	if err != nil {
		log.Println("LoadTest: LayerOneDream: dreamsvc.LayerOneDream failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if resp.Status != "Success" {
		log.Println("LoadTest: LayerOneDream: dreamsvc.LayerOneDream failed with Status - ", resp.Status)
		log.Println("LoadTest: LayerOneDream: dreamsvc.LayerOneDream failed with Error - ", resp.ErrorDescription)
		respDto.Status = resp.Status
		respDto.ErrorDescription = resp.ErrorDescription
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: LayerOneDream: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      LayerOneSportRadar is a function to test system with L3-L1 latency
// @Description  Will respond back getting response from L1.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Success      200  {object}  responsedto.DefaultRespDto{}
// @Failure      503  {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/layer-one-sportradar [post]
func LayerOneSportRadar(c *fiber.Ctx) error {
	log.Println("LoadTest: LayerOneSportRadar: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	// 3. Parse request body to Request Object
	resp, err := dreamsvc.LayerOneSportRadar()
	if err != nil {
		log.Println("LoadTest: LayerOneSportRadar: dreamsvc.LayerOneSportRadar failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	if resp.Status != "Success" {
		log.Println("LoadTest: LayerOneSportRadar: dreamsvc.LayerOneSportRadar failed with Status - ", resp.Status)
		log.Println("LoadTest: LayerOneSportRadar: dreamsvc.LayerOneSportRadar failed with Error - ", resp.ErrorDescription)
		respDto.Status = resp.Status
		respDto.ErrorDescription = resp.ErrorDescription
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("LoadTest: LayerOneSportRadar: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      Login is a function to test system with 50ms average response time
// @Description  Will respond back after 50ms sleep.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        Login  body      requestdto.DefaultReqDto  true  "DefaultReqDto model is used"
// @Success      200           {object}  responsedto.DefaultRespDto{}
// @Failure      503           {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/login [post]
func Login(c *fiber.Ctx) error {
	log.Println("LoadTest: Login: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := requestdto.DefaultReqDto{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: Login: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Parse request body to Request Object
	betfairmodule.Login2()
	log.Println("LoadTest: Login: - END")
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      CurrentOrders is a function to test system with 50ms average response time
// @Description  Will respond back after 50ms sleep.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        CurrentOrders  body      request.ListCurrentOrdersReq  true  "ListCurrentOrdersReq model is used"
// @Success      200            {object}  response.BFCurrentOrdersResp{}
// @Failure      503            {object}  response.BFCurrentOrdersResp{}
// @Router       /loadtest/current-orders [post]
func CurrentOrders(c *fiber.Ctx) error {
	log.Println("LoadTest: Login: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := response.BFCurrentOrdersResp{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := request.ListCurrentOrdersReq{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: CurrentOrders: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Parse request body to Request Object
	//listCOR := request.ListCurrentOrdersReq{}
	resp, err := betfairmodule.CurrentOrders(reqDto)
	if err != nil {
		log.Println("LoadTest: CurrentOrders: betfairmodule.CurrentOrders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.CurrentOrdersResp = resp
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	log.Println("LoadTest: CurrentOrders: - END")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      ClearedOrders is a function to test system with 50ms average response time
// @Description  Will respond back after 50ms sleep.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        ClearedOrders  body      request.ListClearedOrdersReq  true  "ListClearedOrdersReq model is used"
// @Success      200            {object}  response.BFClearedOrdersResp{}
// @Failure      503            {object}  response.BFClearedOrdersResp{}
// @Router       /loadtest/cleared-orders [post]
func ClearedOrders(c *fiber.Ctx) error {
	log.Println("LoadTest: Login: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := response.BFClearedOrdersResp{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := request.ListClearedOrdersReq{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: ClearedOrders: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Parse request body to Request Object
	// listCOR := request.ListClearedOrdersReq{}
	// listCOR.BetStatus = reqDto.BetStatus
	// listCOR.BetIds = []string{}
	// listCOR.BetIds = append(listCOR.BetIds, reqDto.BetIds...)
	resp, err := betfairmodule.ClearedOrders(reqDto)
	if err != nil {
		log.Println("LoadTest: ClearedOrders: betfairmodule.ClearedOrders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.ClearedOrdersResp = resp
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	log.Println("LoadTest: ClearedOrders: - END")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      CancelOrders is a function to test system with 50ms average response time
// @Description  Will respond back after 50ms sleep.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        CancelOrders  body      request.ListCancelOrdersReq  true  "ListCancelOrdersReq model is used"
// @Success      200           {object}  response.BFCancelOrdersResp{}
// @Failure      503           {object}  response.BFCancelOrdersResp{}
// @Router       /loadtest/cancel-orders [post]
func CancelOrders(c *fiber.Ctx) error {
	log.Println("LoadTest: Login: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := response.BFCancelOrdersResp{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"
	// 2. Parse request body to Request Object
	reqDto := request.ListCancelOrdersReq{}
	err := c.BodyParser(&reqDto)
	if err != nil {
		log.Println("LoadTest: CancelOrders: Body Parsing failed with error - ", err.Error())
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	// 3. Parse request body to Request Object
	// listCOR := request.ListCancelOrdersReq{}
	// listCOR.MarketId = reqDto.MarketId
	// listCOR.CustomerRef = reqDto.CustomerRef
	// listCOR.Instructions = []request.CancelInstruction{}
	// listCOR.Instructions = append(listCOR.Instructions, reqDto.Instructions...)
	resp, err := betfairmodule.CancelOrders(reqDto)
	if err != nil {
		log.Println("LoadTest: CancelOrders: betfairmodule.CancelOrders failed with error - ", err.Error())
		respDto.ErrorDescription = err.Error()
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	respDto.CancelOrdersResp = resp
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	log.Println("LoadTest: CancelOrders: - END")
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// @Summary      GetBFMetrics is a function to test system with 50ms average response time
// @Description  Will respond back after 50ms sleep.
// @Tags         loadtest
// @Accept       json
// @Produce      json
// @Param        GetBFMetrics  body      requestdto.DefaultReqDto  true  "DefaultReqDto model is used"
// @Success      200          {object}  responsedto.DefaultRespDto{}
// @Failure      503          {object}  responsedto.DefaultRespDto{}
// @Router       /loadtest/get-bfmetrics [post]
func GetBFMetrics(c *fiber.Ctx) error {
	log.Println("LoadTest: GetBFMetrics: - START")
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response Object
	respDto := responsedto.DefaultRespDto{}
	betfairmodule.GetMetrics()
	respDto.Status = "RS_OK"
	respDto.ErrorDescription = ""
	log.Println("LoadTest: CancelOrders: - END")
	return c.Status(fiber.StatusOK).JSON(respDto)
}
