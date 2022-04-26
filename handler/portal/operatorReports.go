package portalsvc

import (
	"Sp/cache"
	"Sp/constants"
	"Sp/database"
	"Sp/dto/models"
	operatordto "Sp/dto/operator"
	"Sp/dto/reports"
	"Sp/dto/sports"
	"Sp/handler"
	utils "Sp/utilities"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Get Bet List Report API
// @Summary      Bet List Report API
// @Description  Bet List Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization    header    string                         true  "Bearer Token"
// @Param        BetList        body      reports.BetListReqDto  true  "BetListReqDto model is used"
// @Success      200            {object}  reports.BetListRespDto
// @Failure      503            {object}  reports.BetListRespDto
// @Router       /reports/get-bet-list [post]
func BetList(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.BetListRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BetList: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("BetList: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.BetListReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetBets: Body Parsing failed")
		log.Println("GetBets: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	objMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println("BetList: No Operator UserId - ", Tknmeta.UserId)
	}

	// create currency map
	bets, err := database.GetBetList(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("BetList: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	partnerMap := make(map[string]string)
	for _, bet := range bets {
		betList := reports.BetList{}
		betList.BetTime = bet.BetReq.ReqTime
		betList.UserName = bet.UserName
		betList.OperatorId = bet.OperatorId
		betList.SportName = bet.BetDetails.SportName
		betList.EventName = bet.BetDetails.EventName
		betList.MarketType = bet.BetDetails.MarketType
		betList.MarketName = bet.BetDetails.MarketName
		betList.RunnerName = bet.BetDetails.RunnerName
		betList.PartnerId = bet.PartnerId // Partner Id
		cur, found := partnerMap[bet.OperatorId+"-"+bet.PartnerId]
		if !found {
			for _, partner := range objMap[bet.OperatorId].(operatordto.OperatorDTO).Partners {
				if partner.PartnerId == bet.PartnerId {
					cur = partner.Currency
				}
				partnerMap[bet.OperatorId+"-"+partner.PartnerId] = partner.Currency
			}

		}
		betList.Currency = cur
		betList.Rate = bet.BetReq.Rate
		betList.Odds = utils.Truncate64(bet.BetDetails.OddValue)
		if bet.BetReq.OddsMatched > 0 {
			betList.Odds = utils.Truncate64(bet.BetReq.OddsMatched)
		}
		betList.BetType = bet.BetDetails.BetType
		betList.Stake = utils.Truncate64(bet.BetDetails.StakeAmount)
		if bet.ProviderId == constants.SAP.ProviderType.BetFair() && bet.BetReq.SizeCancelled > 0 {
			if bet.BetReq.SizePlaced != bet.BetReq.SizeCancelled {
				stakeAmount := bet.BetReq.SizePlaced - bet.BetReq.SizeCancelled
				stakeAmount = stakeAmount * 10
				stakeAmount = stakeAmount * 100 / (100 - bet.BetReq.PlatformHold)
				stakeAmount = stakeAmount * 100 / (100 - bet.BetReq.OperatorHold)
				betList.Stake = utils.Truncate64(stakeAmount)
			}
		}
		betList.TransactionId = bet.BetId
		betList.Status = bet.Status
		betList.OperatorHold = bet.BetReq.OperatorHold
		betList.OperatorAmount = utils.Truncate64(bet.BetReq.OperatorAmount)
		betList.NetAmount = utils.Truncate64(bet.NetAmount)
		betList.SessionOutCome = bet.BetDetails.SessionOutcome
		respDto.BetLists = append(respDto.BetLists, betList)
	}
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Bet Details Report Report API
// @Summary      Get Bet Details Report Report API
// @Description  Get Bet Details Report Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                 true  "Bearer Token"
// @Param        BetDetailReport  body      reports.BetDetailReportReqDto  true  "BetDetailReportReqDto model is used"
// @Success      200              {object}  reports.BetDetailReportRespDto
// @Failure      503              {object}  reports.BetDetailReportRespDto
// @Router       /reports/get-bet-detail-report [post]
func BetDetailReport(c *fiber.Ctx) error {

	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.BetDetailReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("BetDetailReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("BetDetailReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.BetDetailReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("BetDetailReport: Body Parsing failed")
		log.Println("BetDetailReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	// 5. Get Bet Data from DB
	betDetail, err := database.GetBetDetails(reqDto.TransactionId)
	if err != nil {
		log.Println("BetDetailReport: Error in getting bet details - ", err.Error())
		respDto.ErrorDescription = "Failed to get bet details!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 6. Get Market Result from DB
	marketId := []string{betDetail.MarketId}
	markets, err := database.GetMarketsByMarketIds(marketId)
	if err != nil {
		log.Println("BetDetailReport: Error in getting market result - ", err.Error())
		respDto.ErrorDescription = "Failed to get market result!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 6. Create bet report detail
	betReportDetail := reports.BetReportDetail{}
	betReportDetail.BetId = betDetail.BetId
	betReportDetail.BetStatus = betDetail.Status
	if len(betDetail.ResultReqs) > 0 {
		betReportDetail.ResultTime = betDetail.ResultReqs[len(betDetail.ResultReqs)-1].ReqTime
	}
	if len(betDetail.RollbackReqs) > 0 {
		betReportDetail.LapsedTime = betDetail.RollbackReqs[len(betDetail.RollbackReqs)-1].ReqTime
		betReportDetail.LapsedAmount = betDetail.RollbackReqs[len(betDetail.RollbackReqs)-1].RollbackAmount

	}
	betReportDetail.BetType = betDetail.BetDetails.BetType
	betReportDetail.Odds = betDetail.BetDetails.OddValue
	if len(markets) > 0 {
		marketResult := markets[0].Results
		if len(marketResult) > 0 {
			marketResult := marketResult[len(marketResult)-1]
			marketResult.ResultTime = marketResult.ResultTime * 1000
			betReportDetail.MarketResult = marketResult
		}
	}
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.BetReportDetail = betReportDetail
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get User Statement API
// @Summary      User Statement API
// @Description  User Statement API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                       true  "Bearer Token"
// @Param        Statement      body      reports.UserStatementReqDto  true  "UserStatementReqDto model is used"
// @Success      200            {object}  reports.UserStatementRespDto
// @Failure      503            {object}  reports.UserStatementRespDto
// @Router       /reports/get-user-statement [post]
func Statement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.UserStatementRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("Statement: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("Statement: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.UserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("Statement: Body Parsing failed")
		log.Println("Statement: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	bets, err := database.GetUserStatement(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("Statement: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	for _, bet := range bets {
		userStatement := reports.UserStatement{}
		if len(bet.ResultReqs) > 0 {
			userStatement.SettlementTime = bet.ResultReqs[len(bet.ResultReqs)-1].ReqTime
			userStatement.CreditAmount = bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount
		}
		userStatement.EventName = bet.BetDetails.EventName
		userStatement.MarketType = bet.BetDetails.MarketType
		userStatement.DebitAmount = bet.BetReq.DebitAmount
		userStatement.TransactionId = bet.BetId
		respDto.UserStatements = append(respDto.UserStatements, userStatement)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)

}

// Get Admin Statement API
// @Summary      Admin Statement API
// @Description  Admin Statement API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization   header    string                        true  "Bearer Token"
// @Param        AdminStatement  body      reports.AdminStatementReqDto  true  "AdminStatementReqDto model is used"
// @Success      200             {object}  reports.AdminStatementRespDto
// @Failure      503             {object}  reports.AdminStatementRespDto
// @Router       /reports/get-admin-statement [post]
func AdminStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.AdminStatementRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("AdminStatement: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("AdminStatement: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.AdminStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("AdminStatement: Body Parsing failed")
		log.Println("AdminStatement: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetAdminStatement(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("AdminStatement: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	for _, bet := range bets {
		adminStatement := reports.AdminStatement{}
		if len(bet.ResultReqs) > 0 {
			adminStatement.SettlementTime = bet.ResultReqs[len(bet.ResultReqs)-1].ReqTime
		}
		adminStatement.FromTo = bet.ProviderId + "-" + bet.OperatorId
		adminStatement.Points = bet.BetReq.DebitAmount * float64(bet.BetReq.Rate)
		adminStatement.Status = bet.Status
		adminStatement.Amount = bet.BetReq.DebitAmount
		adminStatement.MyShare = bet.BetReq.OperatorAmount
		adminStatement.TransactionId = bet.BetId
		respDto.AdminStatements = append(respDto.AdminStatements, adminStatement)
	}
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get My Account Statement API
// @Summary      My Account Statement API
// @Description  My Account Statement API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                        true  "Bearer Token"
// @Param        MyAccountStatement  body      reports.MyAccStatementReqDto  true  "MyAccStatementReqDto model is used"
// @Success      200                 {object}  reports.MyAccStatementRespDto
// @Failure      503                 {object}  reports.MyAccStatementRespDto
// @Router       /reports/get-my-account-statement [post]
func MyAccountStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.MyAccStatementRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("MyAccountStatement: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("MyAccountStatement: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.MyAccStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("MyAccountStatement: Body Parsing failed")
		log.Println("MyAccountStatement: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetMyAccStatement(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("AdminStatement: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	for _, bet := range bets {
		myAccStatement := reports.MyAccStatement{}
		myAccStatement.SettlementTime = bet.UpdatedAt
		myAccStatement.FromTo = bet.ProviderId + "-" + bet.OperatorId
		myAccStatement.Points = bet.BetReq.DebitAmount * float64(bet.BetReq.Rate)
		myAccStatement.Amount = bet.BetReq.DebitAmount
		myAccStatement.MyShare = bet.BetReq.OperatorAmount
		myAccStatement.TransactionId = bet.BetId
		myAccStatement.Status = bet.Status
		respDto.MyAccStatement = append(respDto.MyAccStatement, myAccStatement)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get My Game Report API
// @Summary      My Game Report API
// @Description  My Game Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                    true  "Bearer Token"
// @Param        GameReport     body      reports.GameReportReqDto  true  "GameReportReqDto model is used"
// @Success      200            {object}  reports.GameReportRespDto
// @Failure      503            {object}  reports.GameReportRespDto
// @Router       /reports/get-game-report [post]
func GameReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.GameReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GameReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GameReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.GameReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GameReport: Body Parsing failed")
		log.Println("GameReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForGameReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("AdminStatement: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	gamePlayed := make(map[string]reports.GameReport)
	for _, bet := range bets {
		if _, ok := gamePlayed[bet.BetDetails.SportName]; !ok {
			gamePlayed[bet.BetDetails.SportName] = reports.GameReport{
				GameName:  bet.BetDetails.SportName,
				BetCount:  0,
				WinCount:  0,
				LoseCount: 0,
				VoidCount: 0,
				WinAmount: bet.NetAmount,
			}
		} else {
			gReport := gamePlayed[bet.BetDetails.SportName]
			gReport.BetCount++
			gReport.WinAmount += bet.NetAmount
			if bet.Status == "VOIDED" || bet.Status == "CANCELLED" ||
				bet.Status == "VOIDED-failed" || bet.Status == "CANCELLED-failed" {
				gReport.VoidCount++
			} else {
				if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
					if len(bet.ResultReqs) > 0 {
						if bet.BetReq.DebitAmount-bet.ResultReqs[len(bet.ResultReqs)-1].CreditAmount > 0 {
							gReport.WinCount++
						} else {
							gReport.LoseCount++
						}
					}
				}
			}
			gamePlayed[bet.BetDetails.SportName] = gReport
		}
	}

	for _, game := range gamePlayed {
		gr := reports.GameReport{}
		gr.GameName = game.GameName
		gr.BetCount = game.BetCount
		gr.WinCount = game.WinCount
		gr.LoseCount = game.LoseCount
		gr.WinAmount = game.WinAmount
		respDto.GameReports = append(respDto.GameReports, gr)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.UserName = reqDto.UserName
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func SportReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")
	// 1. Create Default Response
	respDto := reports.SportReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("SportReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("SportReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.SportReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SportReport: Body Parsing failed")
		log.Println("SportReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForSportReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("AdminStatement: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	objMap, err := cache.GetObjectMap(constants.SAP.ObjectTypes.OPERATOR())
	if err != nil {
		log.Println("BetList: No Operator UserId - ", Tknmeta.UserId)
	}

	sportPlayed := make(map[string]reports.SportReport)
	partnerMap := make(map[string]string)
	// filter bets by eventName
	for _, bet := range bets {
		statement := reports.MarketStatement{}
		statement.BetTime = bet.BetReq.ReqTime
		statement.BetType = bet.BetDetails.BetType
		statement.MarketName = bet.BetDetails.MarketName
		statement.MarketType = bet.BetDetails.MarketType
		if bet.BetReq.OddsMatched != 0 {
			statement.OddValue = bet.BetReq.OddsMatched
		} else {
			statement.OddValue = bet.BetDetails.OddValue
		}
		statement.RunnerName = bet.BetDetails.RunnerName
		statement.SessionOutCome = bet.BetDetails.SessionOutcome
		statement.StakeAmount = bet.BetDetails.StakeAmount
		statement.Status = bet.Status
		statement.Returns = bet.NetAmount
		cur, found := partnerMap[bet.OperatorId+"-"+bet.PartnerId]
		if !found {
			for _, partner := range objMap[bet.OperatorId].(operatordto.OperatorDTO).Partners {
				if partner.PartnerId == bet.PartnerId {
					cur = partner.Currency
				}
				partnerMap[bet.OperatorId+"-"+partner.PartnerId] = partner.Currency
			}

		}
		statement.Currency = cur
		statement.Rate = bet.BetReq.Rate
		statement.EventName = bet.BetDetails.EventName
		statement.TransactionId = bet.BetId
		statement.UserId = bet.UserId
		statement.UserName = bet.UserName
		statement.SportName = bet.BetDetails.SportName
		statement.SportId = bet.SportId
		if _, ok := sportPlayed[bet.MarketId]; !ok {
			sportPlayed[bet.MarketId] = reports.SportReport{
				EventName:  bet.BetDetails.EventName,
				EventId:    bet.EventId,
				ProviderId: bet.ProviderId,
				BetCount:   1,
				NetAmount:  bet.NetAmount,
				Bets:       []reports.MarketStatement{statement},
			}
		} else {
			sReport := sportPlayed[bet.MarketId]
			sReport.BetCount += 1
			sReport.NetAmount += bet.NetAmount
			sReport.Bets = append(sReport.Bets, statement)
			sportPlayed[bet.MarketId] = sReport
		}
	}

	for _, sport := range sportPlayed {
		sr := reports.SportReport{}
		sr.EventName = sport.EventName
		sr.BetCount = sport.BetCount
		sr.ProviderId = sport.ProviderId
		sr.NetAmount = sport.NetAmount
		sr.Bets = sport.Bets
		respDto.SportReports = append(respDto.SportReports, sr)

	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.UserId = reqDto.UserId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Profit and Loss Report API
// @Summary      Profit and Loss Report API
// @Description  Profit and Loss Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                   true  "Bearer Token"
// @Param        GameReport     body      reports.PnLReportReqDto  true  "PnLReportReqDto model is used"
// @Success      200            {object}  reports.PnLReportRespDto
// @Failure      503            {object}  reports.PnLReportRespDto
// @Router       /reports/get-pnl-report [post]
func PnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.PnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("PnLReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("PnLReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.PnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("PnLReport: Body Parsing failed")
		log.Println("PnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForPnLReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("PnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	sportPlayed := make(map[string]map[string]reports.PnLReport)
	for _, bet := range bets {
		if _, ok := sportPlayed[bet.BetDetails.SportName]; !ok {
			sportPlayed[bet.BetDetails.SportName] = make(map[string]reports.PnLReport)
		}
		if _, ok := sportPlayed[bet.BetDetails.SportName][bet.BetDetails.MarketType]; !ok {
			sportPlayed[bet.BetDetails.SportName][bet.BetDetails.MarketType] = reports.PnLReport{
				SportName:  bet.BetDetails.SportName,
				MarketType: bet.BetDetails.MarketType,
				ProfitLoss: bet.NetAmount,
			}
		} else {
			gReport := sportPlayed[bet.BetDetails.SportName][bet.BetDetails.MarketType]
			gReport.ProfitLoss += bet.NetAmount
			sportPlayed[bet.BetDetails.SportName][bet.BetDetails.MarketType] = gReport
		}
	}

	for _, sport := range sportPlayed {
		for _, market := range sport {
			pl := reports.PnLReport{}
			pl.SportName = market.SportName
			pl.MarketType = market.MarketType
			pl.ProfitLoss = market.ProfitLoss
			respDto.PnLReports = append(respDto.PnLReports, pl)
		}
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.UserName = reqDto.UserName
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Provider Pnl Report API
// @Summary      Provider Pnl Report API
// @Description  Provider Pnl Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization      header    string                           true  "Bearer Token"
// @Param        ProviderPnLReport  body      reports.ProviderPnLReportReqDto  true  "ProviderPnLReportReqDto model is used"
// @Success      200                {object}  reports.ProviderPnLReportRespDto
// @Failure      503                {object}  reports.ProviderPnLReportRespDto
// @Router       /reports/get-provider-pnl-report [post]
func ProviderPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.ProviderPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("ProviderPnLReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("ProviderPnLReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.ProviderPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("ProviderPnLReport: Body Parsing failed")
		log.Println("ProviderPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForProviderPnLReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("ProviderPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	providerPlayed := make(map[string]reports.ProviderPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := providerPlayed[bet.SportId]; !ok {
				providerPlayed[bet.SportId] = reports.ProviderPnLReport{
					ProviderId: bet.ProviderId,
					SportName:  bet.BetDetails.SportName,
					SportId:    bet.SportId,
					ProfitLoss: bet.NetAmount,
					BetCount:   1,
				}
			} else {
				gReport := providerPlayed[bet.SportId]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				providerPlayed[bet.SportId] = gReport
			}
		}
	}

	for _, provider := range providerPlayed {
		pl := reports.ProviderPnLReport{}
		pl.ProviderId = provider.ProviderId
		pl.SportName = provider.SportName
		pl.SportId = provider.SportId
		pl.ProfitLoss = provider.ProfitLoss
		pl.BetCount = provider.BetCount
		respDto.ProviderPnLReports = append(respDto.ProviderPnLReports, pl)
	}
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.ProviderId = reqDto.ProviderId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Sports Pnl Report API
// @Summary      Sports Pnl Report API
// @Description  Sports Pnl Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization   header    string                        true  "Bearer Token"
// @Param        SportPnLReport  body      reports.SportPnLReportReqDto  true  "SportPnLReportReqDto model is used"
// @Success      200             {object}  reports.SportPnLReportRespDto
// @Failure      503             {object}  reports.SportPnLReportRespDto
// @Router       /reports/get-sport-pnl-report [post]
func SportPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.SportPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("SportPnLReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("SportPnLReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.SportPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("SportPnLReport: Body Parsing failed")
		log.Println("SportPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForSportPnLReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("SportPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	competitionPlayed := make(map[string]reports.SportPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := competitionPlayed[bet.CompetitionId]; !ok {
				competitionPlayed[bet.CompetitionId] = reports.SportPnLReport{
					SportName:       bet.BetDetails.SportName,
					SportId:         bet.SportId,
					CompetitionName: bet.BetDetails.CompetitionName,
					CompetitionId:   bet.CompetitionId,
					ProfitLoss:      bet.NetAmount,
					BetCount:        1,
				}
			} else {
				gReport := competitionPlayed[bet.CompetitionId]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				competitionPlayed[bet.CompetitionId] = gReport
			}
		}
	}

	for _, competition := range competitionPlayed {
		pl := reports.SportPnLReport{}
		pl.SportName = competition.SportName
		pl.SportId = competition.SportId
		pl.CompetitionName = competition.CompetitionName
		pl.CompetitionId = competition.CompetitionId
		pl.ProfitLoss = competition.ProfitLoss
		pl.BetCount = competition.BetCount
		respDto.SportPnLReports = append(respDto.SportPnLReports, pl)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.SportId = reqDto.SportId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Competition Pnl Report API
// @Summary      Competition Pnl Report API
// @Description  Competition Pnl Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization         header    string                              true  "Bearer Token"
// @Param        CompetitionPnLReport  body      reports.CompetitionPnLReportReqDto  true  "CompetitionPnLReportReqDto model is used"
// @Success      200                   {object}  reports.CompetitionPnLReportRespDto
// @Failure      503                   {object}  reports.CompetitionPnLReportRespDto
// @Router       /reports/get-competition-pnl-report [post]
func CompetitionPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.CompetitionPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("CompetitionPnLReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("CompetitionPnLReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.CompetitionPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("CompetitionPnLReport: Body Parsing failed")
		log.Println("CompetitionPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForCompetitionPnLReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("CompetitionPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	EventPlayed := make(map[string]reports.CompetitionPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := EventPlayed[bet.EventId]; !ok {
				EventPlayed[bet.EventId] = reports.CompetitionPnLReport{
					CompetitionName: bet.BetDetails.CompetitionName,
					CompetitionId:   bet.CompetitionId,
					EventName:       bet.BetDetails.EventName,
					EventId:         bet.EventId,
					ProfitLoss:      bet.NetAmount,
					BetCount:        1,
				}
			} else {
				gReport := EventPlayed[bet.EventId]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				EventPlayed[bet.EventId] = gReport
			}
		}
	}

	for _, event := range EventPlayed {
		pl := reports.CompetitionPnLReport{}
		pl.CompetitionName = event.CompetitionName
		pl.CompetitionId = event.CompetitionId
		pl.EventName = event.EventName
		pl.EventId = event.EventId
		pl.ProfitLoss = event.ProfitLoss
		pl.BetCount = event.BetCount
		respDto.CompetitionPnLReports = append(respDto.CompetitionPnLReports, pl)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.CompetitionId = reqDto.CompetitionId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Event Pnl Report API
// @Summary      Event Pnl Report API
// @Description  Event Pnl Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization   header    string                        true  "Bearer Token"
// @Param        EventPnLReport  body      reports.EventPnLReportReqDto  true  "EventPnLReportReqDto model is used"
// @Success      200             {object}  reports.EventPnLReportRespDto
// @Failure      503             {object}  reports.EventPnLReportRespDto
// @Router       /reports/get-competition-pnl-report [post]
func EventPnLReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.EventPnLReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("EventPnLReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("EventPnLReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.EventPnLReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("EventPnLReport: Body Parsing failed")
		log.Println("EventPnLReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForEventPnLReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("EventPnLReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	marketTypePlayed := make(map[string]reports.EventPnLReport)
	for _, bet := range bets {
		if bet.Status == constants.BetFair.BetStatus.SETTLED() || bet.Status == "SETTLED-failed" {
			if _, ok := marketTypePlayed[bet.BetDetails.MarketName]; !ok {
				marketTypePlayed[bet.BetDetails.MarketName] = reports.EventPnLReport{
					EventId:    bet.EventId,
					EventName:  bet.BetDetails.EventName,
					MarketType: bet.BetDetails.MarketType,
					MarketName: bet.BetDetails.MarketName,
					MarketId:   bet.MarketId,
					ProfitLoss: bet.NetAmount,
					BetCount:   1,
				}
			} else {
				gReport := marketTypePlayed[bet.BetDetails.MarketName]
				gReport.ProfitLoss += bet.NetAmount
				gReport.BetCount += 1
				marketTypePlayed[bet.BetDetails.MarketName] = gReport
			}
		}
	}

	for _, marketType := range marketTypePlayed {
		pl := reports.EventPnLReport{}
		pl.EventId = marketType.EventId
		pl.EventName = marketType.EventName
		pl.MarketType = marketType.MarketType
		pl.MarketName = marketType.MarketName
		pl.MarketId = marketType.MarketId
		pl.ProfitLoss = marketType.ProfitLoss
		pl.BetCount = marketType.BetCount
		respDto.EventPnLReports = append(respDto.EventPnLReports, pl)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.EventId = reqDto.EventId
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get User Audit Report API
// @Summary      Get User Audit Report API
// @Description  Get User Audit Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization       header    string                         true  "Bearer Token"
// @Param        GetUserAuditReport  body      reports.UserAuditReportReqDto  true  "UserAuditReportReqDto model is used"
// @Success      200                 {object}  reports.UserAuditReportRespDto
// @Failure      503                 {object}  reports.UserAuditReportRespDto
// @Router       /reports/get-user-audit-report [post]
func GetUserAuditReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.UserAuditReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetUserAuditReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetUserAuditReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.UserAuditReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetUserAuditReport: Body Parsing failed")
		log.Println("GetUserAuditReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	audits, err := database.GetUserAuditReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("GetUserAuditReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	for _, audit := range audits {
		userAuditReport := reports.UserAuditReport{}
		userAuditReport.Time = audit.Time.UnixNano() / int64(time.Millisecond)
		userAuditReport.UserName = audit.UserId
		opSplit := strings.Split(audit.Operation, "/")
		userAuditReport.Operation = strings.Title(strings.Replace(opSplit[len(opSplit)-1], "-", " ", -1))

		var jsonMap map[string]interface{}
		json.Unmarshal([]byte(audit.Payload), &jsonMap)
		if _, ok := jsonMap["SportId"]; ok {
			sportId := jsonMap["SportId"].(string)
			jsonMap["SportName"] = utils.SportsMapById[sportId]
		}
		userAuditReport.Payload = jsonMap
		userAuditReport.IP = audit.IP
		respDto.UserAuditReport = append(respDto.UserAuditReport, userAuditReport)
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GetRiskReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.RiskReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetRiskReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetRiskReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.RiskReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetRiskReport: Body Parsing failed")
		log.Println("GetRiskReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetRiskReport(Tknmeta.OperatorId, reqDto)
	if err != nil {
		log.Println("GetRiskReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// Get events for the bets
	events := make(map[string]string)
	eventIds := []string{}
	marketIds := []string{}
	for _, bet := range bets {
		if _, ok := events[bet.EventId]; !ok {
			events[bet.EventId] = bet.BetDetails.EventName
			eventIds = append(eventIds, bet.EventId)
			marketIds = append(marketIds, bet.MarketId)
		}
	}

	// Get Market for all Events from DB
	markets, err := database.GetMarketsByMarketIds(marketIds)
	if err != nil {
		log.Println("GetRiskReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get matched status!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetRiskReport: Market length - ", len(markets))

	risk := make(map[string]map[string]float64)
	for _, bet := range bets {
		for _, market := range markets {
			if market.MarketId == bet.MarketId {
				if _, ok := risk[bet.EventId]; !ok {
					risk[bet.EventId] = make(map[string]float64)
				}
				risk[bet.EventId][market.MarketId] = 0
			}
		}
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	return c.Status(fiber.StatusOK).JSON(respDto)
}

func GenerateExcel(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.GenerateExcelRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GenerateExcelReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GenerateExcelReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.GenerateReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GenerateExcelReport: Body Parsing failed")
		log.Println("GenerateExcelReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}
	log.Println("GenerateExcelReport: Request Body is - ", reqDto)
	// map of lists of strings
	// sheetMap := make(map[string][]string)
	// for _, sheet := range reqDto.Excel.([]interface{}) {
	// 	sheetMap[sheet.SheetName] = sheet.SheetData
	// }
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""

	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Market Risk Report API
// @Summary      Get Market Risk Report API
// @Description  Get Market Risk Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization        header    string                    true  "Bearer Token"
// @Param        GetMarketRiskReport  body      reports.RiskReportReqDto  true  "RiskReportReqDto model is used"
// @Success      200                  {object}  reports.RiskReportRespDto
// @Failure      503                  {object}  reports.RiskReportRespDto
// @Router       /reports/get-operator-risk-report [post]
func GetMarketRiskReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.RiskReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetMarketRiskReport: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetMarketRiskReport: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.RiskReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetMarketRiskReport: Body Parsing failed")
		log.Println("GetMarketRiskReport: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForMarketRiskReport(Tknmeta.OperatorId, Tknmeta.Role, reqDto)
	if err != nil {
		log.Println("GetMarketRiskReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get bets!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// Get Market for all Events from DB
	dbMarkets, err := database.GetMarketsByEventId(reqDto.EventId, reqDto.ProviderId)
	if err != nil {
		log.Println("GetMarketRiskReport: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get Markets!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	markets := []models.Market{}
	for _, market := range dbMarkets {
		if market.MarketType == constants.SAP.MarketType.MATCH_ODDS() || market.MarketType == constants.SAP.MarketType.BOOKMAKER() {
			markets = append(markets, market)
		}
	}

	mIdRunRisk := []reports.MarketIdToRunnerRisk{}
	for _, market := range markets {
		runnerRisks := singleMarketRisk(Tknmeta.OperatorId, bets, market)
		mIdRunRisk = append(mIdRunRisk, reports.MarketIdToRunnerRisk{MarketId: market.MarketId, RunnerRisks: runnerRisks})
	}
	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.EventId = reqDto.EventId

	for _, dbMarket := range markets {
		respMarket := reports.Market{}
		respDto.EventName = dbMarket.EventName
		respMarket.MarketId = dbMarket.MarketId
		respMarket.MarketName = dbMarket.MarketName
		respMarket.MarketType = dbMarket.MarketType
		respRunners := []reports.Runner{}
		runRisks := make(map[string]float64)
		for _, mm := range mIdRunRisk {
			if mm.MarketId == dbMarket.MarketId {
				runRisks = mm.RunnerRisks
			}
		}
		for _, runner := range dbMarket.Runners {
			respRunners = append(respRunners, reports.Runner{
				RunnerName: runner.RunnerName,
				RunnerId:   runner.RunnerId,
				RunnerRisk: runRisks[runner.RunnerId],
			})
		}
		respMarket.Runners = respRunners
		respDto.Markets = append(respDto.Markets, respMarket)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

func singleMarketRisk(operatorId string, bets []sports.BetDto, market models.Market) map[string]float64 {
	runnerRisks := make(map[string]float64)
	for _, bet := range bets {
		if market.MarketId == bet.MarketId {
			var UserWonAmt float64 = 0
			var UserLostAmt float64 = 0
			var oddValue float64 = 0
			var stakeAmt float64 = 0
			if bet.ProviderId == constants.SAP.ProviderType.BetFair() {
				// Check the cancel bet amount
				if bet.BetReq.SizeCancelled != 0 {
					stakeAmt = stakeAmt - bet.BetReq.SizeCancelled
				}
				// If betFair we calc the operator amount
				stakeAmt = ((bet.BetDetails.StakeAmount - stakeAmt) * bet.BetReq.OperatorHold) / 100
				if operatorId == "" {
					// If betfair and operatorId is empty i.e, SAPAdmin is calling in StakeAmt we calc platform hold
					stakeAmt = ((bet.BetDetails.StakeAmount - stakeAmt) * bet.BetReq.PlatformHold) / 100
				}

			} else {
				// for others we use 100% in StakeAmt
				stakeAmt = bet.BetDetails.StakeAmount
			}
			// If bet matched to BetFair we use betfair's odds
			if bet.BetReq.OddsMatched != 0 {
				oddValue = bet.BetReq.OddsMatched
			} else {
				oddValue = bet.BetDetails.OddValue
			}
			if bet.BetDetails.BetType == "BACK" {
				if market.MarketType == constants.SAP.MarketType.MATCH_ODDS() {
					// Using stakeAmt here because we are calculating the risk for BetFair and Betfair's stakeAmt is variable.
					// For Betfair we use betfair's stakeAmt and for others we use 100% stakeAmt
					UserWonAmt = stakeAmt * (oddValue - 1) * -1
					UserLostAmt = stakeAmt
				} else if market.MarketType == constants.SAP.MarketType.BOOKMAKER() {
					// Using the actual value of StakeAmount as Bookmaker's stakeAmount percentage is fixed.
					UserWonAmt = (oddValue * bet.BetDetails.StakeAmount) / 100 * -1
					UserLostAmt = bet.BetDetails.StakeAmount
				}

			} else {
				if market.MarketType == constants.SAP.MarketType.MATCH_ODDS() {
					// Using stakeAmt here because we are calculating the risk for BetFair and Betfair's stakeAmt is variable.
					// For Betfair we use betfair's stakeAmt and for others we use 100% stakeAmt
					UserWonAmt = stakeAmt * -1
					UserLostAmt = stakeAmt * (oddValue - 1)
				} else if market.MarketType == constants.SAP.MarketType.BOOKMAKER() {
					// Using the actual value of StakeAmount as Bookmaker's stakeAmount percentage is fixed.
					UserWonAmt = bet.BetDetails.StakeAmount * -1
					UserLostAmt = oddValue
				}
			}
			for _, runner := range market.Runners {
				if _, ok := runnerRisks[runner.RunnerId]; !ok {
					runnerRisks[runner.RunnerId] = 0 // Initialize the value to 0
				}
				if runner.RunnerId == bet.BetDetails.RunnerId {
					if bet.BetDetails.BetType == "BACK" {
						runnerRisks[runner.RunnerId] += UserWonAmt // Calculate if back bet won by user
					} else {
						runnerRisks[runner.RunnerId] += UserLostAmt // Calculate if lay bet lost by user
					}
				} else {
					if bet.BetDetails.BetType == "LAY" {
						runnerRisks[runner.RunnerId] += UserWonAmt // Calculate if back bet won by user
					} else {
						runnerRisks[runner.RunnerId] += UserLostAmt // Calculate if back bet lost by user
					}
				}
			}
		}
	}
	return runnerRisks
}

func SingleUserRisk(operatorId, userId string, bets []sports.BetDto, market models.Market) map[string]float64 {
	runnerRisks := make(map[string]float64)
	for _, bet := range bets {
		if bet.UserId == userId {
			var UserWonAmt float64 = 0
			var UserLostAmt float64 = 0
			var oddValue float64 = 0
			var stakeAmt float64 = bet.BetDetails.StakeAmount
			// if bet.ProviderId == constants.SAP.ProviderType.BetFair() {
			// 	// If betFair we calc the operator amount
			// 	stakeAmt = (bet.BetDetails.StakeAmount * bet.BetReq.OperatorHold) / 100
			// 	if operatorId == "" {
			// 		// If betfair and operatorId is empty i.e, SAPAdmin is calling in StakeAmt we calc platform hold
			// 		stakeAmt = ((bet.BetDetails.StakeAmount - stakeAmt) * bet.BetReq.PlatformHold) / 100
			// 	}
			// } else {
			// 	// for others we use 100% in StakeAmt
			// 	stakeAmt = bet.BetDetails.StakeAmount
			// }
			// Check the cancel bet amount
			if bet.BetReq.SizeCancelled != 0 {
				stakeAmt = stakeAmt - bet.BetReq.SizeCancelled
			}

			// If bet matched to BetFair we use betfair's odds
			if bet.BetReq.OddsMatched != 0 {
				oddValue = bet.BetReq.OddsMatched
			} else {
				oddValue = bet.BetDetails.OddValue
			}
			if bet.BetDetails.BetType == "BACK" {
				if market.MarketType == constants.SAP.MarketType.MATCH_ODDS() {
					// Using stakeAmt here because we are calculating the risk for BetFair and Betfair's stakeAmt is variable.
					// For Betfair we use betfair's stakeAmt and for others we use 100% stakeAmt
					UserWonAmt = stakeAmt * (oddValue - 1)
					UserLostAmt = stakeAmt * -1
				} else if market.MarketType == constants.SAP.MarketType.BOOKMAKER() {
					// Using the actual value of StakeAmount as Bookmaker's stakeAmount percentage is fixed.
					UserWonAmt = (oddValue * stakeAmt) / 100
					UserLostAmt = stakeAmt * -1
				}

			} else {
				if market.MarketType == constants.SAP.MarketType.MATCH_ODDS() {
					// Using stakeAmt here because we are calculating the risk for BetFair and Betfair's stakeAmt is variable.
					// For Betfair we use betfair's stakeAmt and for others we use 100% stakeAmt
					UserWonAmt = stakeAmt
					UserLostAmt = stakeAmt * (oddValue - 1) * -1
				} else if market.MarketType == constants.SAP.MarketType.BOOKMAKER() {
					// Using the actual value of StakeAmount as Bookmaker's stakeAmount percentage is fixed.
					UserWonAmt = stakeAmt
					UserLostAmt = oddValue * -1
				}
			}
			for _, runner := range market.Runners {
				if _, ok := runnerRisks[runner.RunnerId]; !ok {
					runnerRisks[runner.RunnerId] = 0 // Initialize the value to 0
				}
				if runner.RunnerId == bet.BetDetails.RunnerId {
					if bet.BetDetails.BetType == "BACK" {
						runnerRisks[runner.RunnerId] += UserWonAmt // Calculate if back bet won by user
					} else {
						runnerRisks[runner.RunnerId] += UserLostAmt // Calculate if lay bet lost by user
					}
				} else {
					if bet.BetDetails.BetType == "LAY" {
						runnerRisks[runner.RunnerId] += UserWonAmt // Calculate if back bet won by user
					} else {
						runnerRisks[runner.RunnerId] += UserLostAmt // Calculate if back bet lost by user
					}
				}
			}
		}
	}
	return runnerRisks
}

// Get User Book Report API
// @Summary      Get User Book Report API
// @Description  Get User Book Report API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization          header    string                               true  "Bearer Token"
// @Param        GetUserBookReport  body      reports.UserBookReportReqDto  true  "UserBookReportReqDto model is used"
// @Success      200                {object}  reports.UserBookReportRespDto
// @Failure      503                {object}  reports.UserBookReportRespDto
// @Router       /reports/get-user-book-report [post]
func GetUserBookReport(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.UserBookReportRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// 2. Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// 2.1. Token validaton failed.
		log.Println("GetUserBook: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// 3. Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("GetUserBook: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// 4. Get Request Body
	reqDto := reports.UserBookReportReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("GetUserBook: Body Parsing failed")
		log.Println("GetUserBook: Request Body is - ", err)
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	bets, err := database.GetBetsForUserBookReport(Tknmeta.OperatorId, Tknmeta.Role, reqDto)
	if err != nil {
		log.Println("GetUserBook: Error in getting bets - ", err.Error())
		respDto.ErrorDescription = "Failed to get bets!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	log.Println("GetUserBook: len of bets - ", len(bets))
	marketKey := reqDto.ProviderId + "-" + reqDto.SportId + "-" + reqDto.EventId + "-" + reqDto.MarketId
	market, err := database.GetMarket(marketKey)
	if err != nil {
		log.Println("GetUserBook: Error in getting matched status - ", err.Error())
		respDto.ErrorDescription = "Failed to get Markets!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	if market.MarketType != constants.SAP.MarketType.MATCH_ODDS() && market.MarketType != constants.SAP.MarketType.BOOKMAKER() {
		log.Println("GetUserBook: MarketType is Match Odds or Bookmaker")
		respDto.ErrorDescription = "MarketType is Match Odds or Bookmaker"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	uniqueUserIds := []string{}
	uniqueUserNameMap := map[string]string{}
	for _, bet := range bets {
		if !contains(uniqueUserIds, bet.UserId) {
			uniqueUserIds = append(uniqueUserIds, bet.UserId)
			uniqueUserNameMap[bet.UserId] = bet.UserName
		}
	}

	mIdRunRisk := []reports.UserIdToRunnerRisk{}
	for _, userId := range uniqueUserIds {
		runnerRisks := SingleUserRisk(Tknmeta.OperatorId, userId, bets, market)
		mIdRunRisk = append(mIdRunRisk, reports.UserIdToRunnerRisk{UserId: userId, RunnerRisks: runnerRisks})
	}

	// Success Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.EventId = reqDto.EventId
	respDto.EventName = market.EventName
	respDto.MarketId = reqDto.MarketId
	respDto.MarketName = market.MarketName
	respDto.MarketType = market.MarketType
	for _, userId := range uniqueUserIds {
		respUser := reports.UserBook{}
		respUser.UserId = userId
		respUser.UserName = uniqueUserNameMap[userId]
		runRisks := make(map[string]float64)
		for _, mm := range mIdRunRisk {
			if mm.UserId == userId {
				runRisks = mm.RunnerRisks
			}
		}
		for _, runner := range market.Runners {
			respUser.UserBookRunners = append(respUser.UserBookRunners, reports.UserBookRunner{
				RunnerName: runner.RunnerName,
				RunnerId:   runner.RunnerId,
				RunnerRisk: runRisks[runner.RunnerId],
			})
		}
		respDto.UserBooks = append(respDto.UserBooks, respUser)
	}
	return c.Status(fiber.StatusOK).JSON(respDto)
}

// Get Transfer User Statement API
// @Summary      Get Transfer User Statement API
// @Description  Get Transfer User Statement API
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        Authorization      header    string                        true  "Bearer Token"
// @Param        TransferUserStatement  body      reports.TransferUserStatementReqDto  true  "TransferUserStatementReqDto model is used"
// @Success      200                    {object}  reports.TransferUserStatementRespDto
// @Failure      503                    {object}  reports.TransferUserStatementRespDto
// @Router       /reports/get-transfer-user-statement [post]
func TransferUserStatement(c *fiber.Ctx) error {
	c.Accepts("json", "text")
	c.Accepts("application/json")

	respDto := reports.TransferUserStatementRespDto{}
	respDto.ErrorDescription = "Generic Error"
	respDto.Status = "RS_ERROR"

	// Validation Token
	Tknmeta, ok := Authenticate(c)
	if !ok {
		// Token validaton failed.
		log.Println("TransferUserStatement: Token Validation Failed")
		respDto.ErrorDescription = INVALID_TOKEN_ERROR_DESC
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Check Role Permissions
	if !IsApplicable(Tknmeta, c.OriginalURL()) {
		log.Println("TransferUserStatement: User not Permitted to access the API")
		respDto.ErrorDescription = UNAUTH_ACCESS
		return c.Status(fiber.StatusOK).JSON(respDto)
	}

	// Get Request Body
	reqDto := reports.TransferUserStatementReqDto{}
	if err := c.BodyParser(&reqDto); err != nil {
		log.Println("TransferUserStatement: Body Parsing failed - ", err.Error())
		log.Println("TransferUserStatement: Request Body is - ", string(c.Body()))
		respDto.ErrorDescription = "Invalid Request"
		return c.Status(fiber.StatusBadRequest).JSON(respDto)
	}

	statement, user, err := handler.GetTransferUserStatement(Tknmeta.OperatorId, reqDto.UserId, reqDto.ReferenceId, reqDto.StartTime, reqDto.EndTime)
	if err != nil {
		log.Println("TransferUserStatement: Error in getting bets - ", err.Error())
		respDto.ErrorDescription = "Failed to Generate User Statement!"
		return c.Status(fiber.StatusOK).JSON(respDto)
	}
	// Return Response
	respDto.Status = OK_STATUS
	respDto.ErrorDescription = ""
	respDto.UserId = reqDto.UserId
	respDto.UserName = user.UserName
	respDto.UserBalance = user.Balance
	respDto.Statement = statement

	return c.Status(fiber.StatusOK).JSON(respDto)
}
