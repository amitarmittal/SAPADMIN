package router

import (
	cachesvc "Sp/handler/cache"
	operatorsvc "Sp/handler/operator"
	portalsvc "Sp/handler/portal"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())
	app.Get("/docs/*", swagger.Handler) // default

	apiv1 := api.Group(("/v1"))

	app.Use(logger.New(logger.Config{
		Format:     "${cyan}[${time}] ${green}${pid} ${red}${status} ${blue}[${method}] ${white}${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "Asia/Kolkata",
	}))

	// Auth - IFrame Intregration for operators
	auth := apiv1.Group("/auth")
	auth.Post("/login", operatorsvc.Authentication)

	// reports
	reports := apiv1.Group("/reports")
	reports.Post("/bet-list", operatorsvc.GetBets)
	reports.Post("/get-providers", operatorsvc.GetProviders)

	reports.Post("/get-bet-list", portalsvc.BetList)
	reports.Post("/get-bet-detail-report", portalsvc.BetDetailReport)
	reports.Post("/get-user-statement", portalsvc.Statement)
	reports.Post("/get-admin-statement", portalsvc.AdminStatement)
	reports.Post("/get-my-account-statement", portalsvc.MyAccountStatement)
	reports.Post("/get-game-report", portalsvc.GameReport)
	reports.Post("/get-sport-report", portalsvc.SportReport)
	reports.Post("/get-pnl-report", portalsvc.PnLReport)
	reports.Post("/get-user-audit-report", portalsvc.GetUserAuditReport)
	reports.Post("/get-operator-risk-report", portalsvc.GetMarketRiskReport)
	reports.Post("/get-user-book-report", portalsvc.GetUserBookReport)
	reports.Post("/get-transfer-user-statement", portalsvc.TransferUserStatement)
	//BO PNL APIs
	reports.Post("/get-provider-pnl-report", portalsvc.ProviderPnLReport)
	reports.Post("/get-sport-pnl-report", portalsvc.SportPnLReport)
	reports.Post("/get-competition-pnl-report", portalsvc.CompetitionPnLReport)
	reports.Post("/get-event-pnl-report", portalsvc.EventPnLReport)

	// Transfer Wallet
	wallet := apiv1.Group("/wallet")
	wallet.Post("/user-balance", operatorsvc.GetUserBalance)
	wallet.Post("/deposit-funds", operatorsvc.Deposit)
	wallet.Post("/withdraw-funds", operatorsvc.Withdraw)

	// Portal
	portal := apiv1.Group("/portal")
	portal.Post("/createuser", portalsvc.CreatePortalUser)
	portal.Post("/login", portalsvc.Login)
	//portal.Post("/createprovider", portalsvc.CreateProvider)
	portal.Post("/createoperator", portalsvc.CreateOperator)
	portal.Post("/deleteoperator", portalsvc.DeleteOperator)
	portal.Post("/delete-event-junk-data", portalsvc.DeleteEventJunkData)
	portal.Post("/delete-competition-junk-data", portalsvc.DeleteCompetitionJunkData)
	portal.Post("/delete-sport-junk-data", portalsvc.DeleteSportJunkData)
	portal.Post("/close-events", portalsvc.CloseEvents)
	// SAP Admin
	portal.Post("/get-operators", portalsvc.GetOperators)
	portal.Post("/get-partners", portalsvc.GetPartners)
	portal.Post("/replace-operators", portalsvc.ReplaceOperators)
	portal.Post("/get-providers", portalsvc.GetProviders)
	portal.Post("/get-provider-status", portalsvc.GetProviderStatus)
	portal.Post("/block-operator", portalsvc.BlockOperator)
	portal.Post("/unblock-operator", portalsvc.UnblockOperator)
	portal.Post("/get-saproviders", portalsvc.GetSAProviders)
	portal.Post("/block-saprovider", portalsvc.BlockSAProvider)
	portal.Post("/unblock-saprovider", portalsvc.UnblockSAProvider)
	portal.Post("/block-provider-status", portalsvc.BlockProviderStatus)
	portal.Post("/unblock-provider-status", portalsvc.UnblockProviderStatus)
	// portal.Post("/block-operator-status", portalsvc.BlockOperatorStatus)
	// portal.Post("/unblock-operator-status", portalsvc.UnblockOperatorStatus)
	// Operator Admin
	portal.Post("/list-users", portalsvc.ListUsers)
	portal.Post("/list-sports", portalsvc.ListSports)
	portal.Post("/list-events", portalsvc.ListEvents)
	portal.Post("/get-bets", portalsvc.GetBets)
	portal.Post("/get-bet", portalsvc.GetBet)
	portal.Post("/user-statement", portalsvc.UserStatement)
	portal.Post("/get-users", portalsvc.GetUsersList)
	portal.Post("/block-user", portalsvc.BlockUser)
	portal.Post("/unblock-user", portalsvc.UnblockUser)
	portal.Post("/get-oaproviders", portalsvc.GetOAProviders)
	portal.Post("/get-providers-for-tab", portalsvc.GetProvidersForTabs)
	portal.Post("/block-oaprovider", portalsvc.BlockOAProvider)
	portal.Post("/unblock-oaprovider", portalsvc.UnblockOAProvider)

	sapAdmin := portal.Group("/sapadmin")
	sapAdmin.Post("/block-partner", portalsvc.BlockPartner)
	sapAdmin.Post("/unblock-partner", portalsvc.UnblockPartner)
	sapAdmin.Post("/sports", portalsvc.GetSportsListForSAP)
	sapAdmin.Post("/block-sport", portalsvc.BlockSportForSAP)
	sapAdmin.Post("/unblock-sport", portalsvc.UnblockSportForSAP)
	sapAdmin.Post("/compititions", portalsvc.GetCompetitionsListForSAP)
	sapAdmin.Post("/recent-compititions", portalsvc.RecentCompetitionsForSAP)
	sapAdmin.Post("/block-competition", portalsvc.BlockCompetitionForSAP)
	sapAdmin.Post("/unblock-competition", portalsvc.UnblockCompetitionForSAP)
	sapAdmin.Post("/events", portalsvc.GetEventsListForSAP)
	sapAdmin.Post("/get-event", portalsvc.GetEventSAP)
	sapAdmin.Post("/block-event", portalsvc.BlockEventForSAP)
	sapAdmin.Post("/unblock-event", portalsvc.UnblockEventForSAP)
	sapAdmin.Post("/get-bets", portalsvc.GetBetsForSAP)
	sapAdmin.Post("/get-settled-bets", portalsvc.GetSettledBetsForSAP)
	sapAdmin.Post("/get-lapsed-bets", portalsvc.GetLapsedBetsForSAP)
	sapAdmin.Post("/get-cancelled-bets", portalsvc.GetCancelledBetsForSAP)
	sapAdmin.Post("/get-users", portalsvc.GetUsersForSAP)
	sapAdmin.Post("/user-statement", portalsvc.UserStatementForSAP)
	sapAdmin.Post("/operator-details", portalsvc.OperatorDetailsForSAP)
	sapAdmin.Post("/operator-status-unblock", portalsvc.OperatorStatusUnblockForSAP)
	sapAdmin.Post("/provider-operator-status", portalsvc.OperatorStatusInProviderForSAP)
	sapAdmin.Post("/sport-operator-status", portalsvc.OperatorStatusInSportForSAP)
	sapAdmin.Post("/competition-operator-status", portalsvc.OperatorStatusInCompetitionForSAP)
	sapAdmin.Post("/event-operator-status", portalsvc.OperatorStatusInEventForSAP)
	sapAdmin.Post("/get-config", portalsvc.GetConfigForSAP)
	sapAdmin.Post("/set-config", portalsvc.SetConfigForSAP)
	sapAdmin.Post("/sync-sports", portalsvc.SyncSports)
	sapAdmin.Post("/sync-competitions", portalsvc.SyncCompetitions)
	sapAdmin.Post("/sync-events", portalsvc.SyncEvents)
	sapAdmin.Post("/get-portal-user", portalsvc.GetPortalUsers)
	sapAdmin.Post("/reset-password", portalsvc.ResetPasswordForSAP)
	sapAdmin.Post("/reset-op-password", portalsvc.ResetOPPasswordForSAP)
	sapAdmin.Post("/add-partner", portalsvc.AddPartner)
	sapAdmin.Post("/get-role", portalsvc.GetRole)
	// Market APIs
	sapAdmin.Post("/get-markets", portalsvc.GetMarketsForSAP)
	sapAdmin.Post("/block-market", portalsvc.BlockMarketForSAP)
	sapAdmin.Post("/unblock-market", portalsvc.UnblockMarketForSAP)
	sapAdmin.Post("/suspend-market", portalsvc.SuspendMarketForSAP)
	sapAdmin.Post("/resume-market", portalsvc.ResumeMarketForSAP)
	sapAdmin.Post("/get-op-markets", portalsvc.GetOpMarketsForSAP)
	sapAdmin.Post("/block-op-market", portalsvc.BlockOPMarketForSAP)
	sapAdmin.Post("/unblock-op-market", portalsvc.UnblockOPMarketForSAP)

	sapAdmin.Post("/block-portal-user", portalsvc.BlockPortalUserForSAP)
	sapAdmin.Post("/unblock-portal-user", portalsvc.UnblockPortalUserForSAP)
	sapAdmin.Post("/replace-partner", portalsvc.ReplacePartnerForSAP)
	sapAdmin.Post("/deposit-funds", portalsvc.OperatorDeposit)
	sapAdmin.Post("/withdraw-funds", portalsvc.OperatorWithdraw)
	sapAdmin.Post("/view-fund-statement", portalsvc.ViewFundStatement)
	sapAdmin.Post("/get-partner-ids", portalsvc.GetPartnerIds)
	sapAdmin.Post("/matched-sports", portalsvc.MatchedSports)
	sapAdmin.Post("/unmatched-sports", portalsvc.UnmatchedSports)
	sapAdmin.Post("/all-sports", portalsvc.AllSports)
	sapAdmin.Post("/update-sport-card", portalsvc.UpdateSportCard)
	sapAdmin.Post("/get-sport-radar-cards", portalsvc.GetSportsRadatCards)
	sapAdmin.Post("/get-new-sports", portalsvc.GetNewSports)
	sapAdmin.Post("/delete-bets", portalsvc.DeleteBets)
	sapAdmin.Post("/get-open-bets", portalsvc.GetOpenBetsForSAP)

	opadmin := portal.Group("/opadmin")
	opadmin.Post("/get-operator-details", portalsvc.GetOperatorDetails)
	opadmin.Post("/sports", portalsvc.GetSportsListForOP)
	opadmin.Post("/block-sport", portalsvc.BlockSportsForOP)
	opadmin.Post("/unblock-sport", portalsvc.UnblockSportsForOP)
	opadmin.Post("/compititions", portalsvc.GetCompetitionsListForOP)
	opadmin.Post("/recent-compititions", portalsvc.RecentCompetitionsForOP)
	opadmin.Post("/block-competition", portalsvc.BlockCompetitionForOP)
	opadmin.Post("/unblock-competition", portalsvc.UnblockCompetitionForOP)
	opadmin.Post("/events", portalsvc.GetEventsListForOP)
	opadmin.Post("/get-bets", portalsvc.GetBetsForOP)
	opadmin.Post("/block-event", portalsvc.BlockEventForOP)
	opadmin.Post("/unblock-event", portalsvc.UnblockEventForOP)
	opadmin.Post("/user-statement", portalsvc.UserStatement)
	opadmin.Post("/get-config", portalsvc.GetConfigForOP)
	opadmin.Post("/set-config", portalsvc.SetConfigForOP)
	opadmin.Post("/reset-password", portalsvc.ResetPasswordForOP)
	// Market APIs
	opadmin.Post("/get-markets", portalsvc.GetMarketsForOP)
	opadmin.Post("/block-market", portalsvc.BlockMarketForOP)
	opadmin.Post("/unblock-market", portalsvc.UnblockMarketForOP)

	opadmin.Post("/get-balance", portalsvc.GetBalanceForOP)
	opadmin.Post("/get-partner-ids", portalsvc.GetPartnerIdsForOP)
	opadmin.Post("/get-open-bets", portalsvc.GetOpenBetsForOP)

	// Get All API Methods and URLs
	app.Get("/stack", func(c *fiber.Ctx) error { return c.JSON(c.App().Stack()) })

	// To monitor Go-Fiber app.
	app.Get("/monitor", monitor.New())

	// Test
	cache := apiv1.Group("/cache")
	cache.Post("/get-cache", cachesvc.GetCache)
}
