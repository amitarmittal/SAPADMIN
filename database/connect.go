package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	//DBRead  *mongo.Client
	DBWrite *mongo.Client
	Ctx     context.Context
	// Database
	Sports01DB *mongo.Database
	// Prod Environment
	// ClusterName    string = "sap"
	// Sports01DBName string = "sapdb"
	// UserName       string = "satya"
	// UserPswd       string = "AI5GEOdtVpKL23MK"
	// ConnectionURL  string = "mongodb+srv://satya:AI5GEOdtVpKL23MK@sap.0ps0y.mongodb.net/sapdb?retryWrites=true&w=majority"
	// ConnectionURL2 string = os.Getenv("DB_CONN")
	// ConnectionURL2 string = "mongodb+srv://%s:%s@%s.0ps0y.mongodb.net/%s?retryWrites=true&w=majority"
	//                       mongodb+srv://<username>:<password>@aggregation.0ps0y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority
	ClusterName    string = os.Getenv("DATABASE_CLUSTER")
	Sports01DBName string = os.Getenv("DATABASE_NAME")
	UserName       string = os.Getenv("DATABASE_USERNAME")
	UserPswd       string = os.Getenv("DATABASE_PWD")
	ConnectionURL2 string = "mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority"
	// Dev Environment
	/*
		ClusterName    string = "devcluster"
		Sports01DBName string = "sports01"
		UserName       string = "app-user-01"
		UserPswd       string = "6JPVhTqmyTQj5q51"
		ConnectionURL  string = "mongodb+srv://app-user-01:6JPVhTqmyTQj5q51@devcluster.3hcqa.mongodb.net/sports01?retryWrites=true&w=majority"
		ConnectionURL2 string = "mongodb+srv://%s:%s@%s.3hcqa.mongodb.net/%s?retryWrites=true&w=majority"
	*/
	// Collections
	// 1. B2B Session
	B2bSessionCollection     *mongo.Collection
	B2bSessionCollectionName string = "b2b_session"
	// 2. Operators
	OperatorCollection     *mongo.Collection
	OperatorCollectionName string = "operators"
	// 3. Events
	EventCollection     *mongo.Collection
	EventCollectionName string = "events"
	// 4. Results
	//ResultCollection     *mongo.Collection
	//ResultCollectionName string = "results"
	// 5. Bets
	BetCollection     *mongo.Collection
	BetCollectionName string = "bets"
	// 6. ResultQueue
	//ResultQueueCollection     *mongo.Collection
	//ResultQueueCollectionName string = "result_queue"
	// 7. User + Wallet
	B2BUserCollection     *mongo.Collection
	B2BUserCollectionName string = "b2b_users"
	// 8. Session Token
	SessionTokenCollection     *mongo.Collection
	SessionTokenCollectionName string = "session_token"
	// 9. Portal Users
	PortalUsersCollection     *mongo.Collection
	PortalUsersCollectionName string = "portal_users"
	// 10. User Ledger
	LedgerCollection     *mongo.Collection
	LedgerCollectionName string = "users_ledger"

	Ledger2Collection     *mongo.Collection
	Ledger2CollectionName string = "users_ledger2"

	UserBalance     *mongo.Collection
	UserBalanceName string = "users_balance"
	// 11. Provider Status
	//ProviderStatusCollection     *mongo.Collection
	//ProviderStatusCollectionName string = "provider_status"
	// 12. Providers
	ProviderCollection     *mongo.Collection
	ProviderCollectionName string = "providers"
	// 13. Sports
	SportCollection     *mongo.Collection
	SportCollectionName string = "sports"
	// 14. Sport Status
	SportStatusCollection     *mongo.Collection
	SportStatusCollectionName string = "sport_status"
	// 15. User Bet Status
	UserBetStatusCollection     *mongo.Collection
	UserBetStatusCollectionName string = "userbet_status"
	// 16. Failed Bets
	FailedBetCollection     *mongo.Collection
	FailedBetCollectionName string = "failed_bets"
	// 17. Competitions
	CompetitionCollection     *mongo.Collection
	CompetitionCollectionName string = "competitions"
	// 18. Competition Status
	CompetitionStatusCollection     *mongo.Collection
	CompetitionStatusCollectionName string = "competition_status"
	// 19. Event Status
	EventStatusCollection     *mongo.Collection
	EventStatusCollectionName string = "event_status"
	// 20. Operator Ledger
	OperatorLedgerCollection     *mongo.Collection
	OperatorLedgerCollectionName string = "operator_ledger"
	// 21.Audit Log
	PortalAuditCollection     *mongo.Collection
	PortalAuditCollectionName string = "portal_audit"
	// 22. Partner Status
	PartnerStatusCollection     *mongo.Collection
	PartnerStatusCollectionName string = "partner_status"
	// 23. Markets
	MarketCollection     *mongo.Collection
	MarketCollectionName string = "markets"
	// 24. Market Status
	MarketStatusCollection     *mongo.Collection
	MarketStatusCollectionName string = "market_status"
	// 25. Operation Locks
	OperationLockCollection     *mongo.Collection
	OperationLockCollectionName string = "operation_locks"
	// 26. Load Test
	LoadTestCollection     *mongo.Collection
	LoadTestCollectionName string = "load_test"
	// 27. BetFair Session
	BFSessionCollection     *mongo.Collection
	BFSessionCollectionName string = "betfair_session"
	// 28. User Market
	UserMarketCollection     *mongo.Collection
	UserMarketCollectionName string = "user_market"
	// 29. Operator Market
	OperatorMarketCollection     *mongo.Collection
	OperatorMarketCollectionName string = "operator_market"
)

// ConnectDB connect to db
func ConnectDB() {
	var err error

	// MongoDB Connection
	connectionURL := fmt.Sprintf(ConnectionURL2, UserName, UserPswd, ClusterName, Sports01DBName)
	clientOptions := options.Client().
		ApplyURI(connectionURL)
	Ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	DBWrite, err = mongo.Connect(Ctx, clientOptions)
	if err != nil {
		log.Fatal(err.Error())
		panic("failed to connect database")
	}
	Sports01DB = DBWrite.Database(Sports01DBName)
	B2bSessionCollection = Sports01DB.Collection(B2bSessionCollectionName)
	OperatorCollection = Sports01DB.Collection(OperatorCollectionName)
	EventCollection = Sports01DB.Collection(EventCollectionName)
	//ResultCollection = Sports01DB.Collection(ResultCollectionName)
	BetCollection = Sports01DB.Collection(BetCollectionName)
	//ResultQueueCollection = Sports01DB.Collection(ResultQueueCollectionName)
	B2BUserCollection = Sports01DB.Collection(B2BUserCollectionName)
	SessionTokenCollection = Sports01DB.Collection(SessionTokenCollectionName)
	PortalUsersCollection = Sports01DB.Collection(PortalUsersCollectionName)
	LedgerCollection = Sports01DB.Collection(LedgerCollectionName)
	Ledger2Collection = Sports01DB.Collection(Ledger2CollectionName)
	UserBalance = Sports01DB.Collection(UserBalanceName)
	//ProviderStatusCollection = Sports01DB.Collection(ProviderStatusCollectionName)
	ProviderCollection = Sports01DB.Collection(ProviderCollectionName)
	SportCollection = Sports01DB.Collection(SportCollectionName)
	SportStatusCollection = Sports01DB.Collection(SportStatusCollectionName)
	UserBetStatusCollection = Sports01DB.Collection(UserBetStatusCollectionName)
	FailedBetCollection = Sports01DB.Collection(FailedBetCollectionName)
	CompetitionCollection = Sports01DB.Collection(CompetitionCollectionName)
	CompetitionStatusCollection = Sports01DB.Collection(CompetitionStatusCollectionName)
	EventStatusCollection = Sports01DB.Collection(EventStatusCollectionName)
	PortalAuditCollection = Sports01DB.Collection(PortalAuditCollectionName)
	OperatorLedgerCollection = Sports01DB.Collection(OperatorLedgerCollectionName)
	PartnerStatusCollection = Sports01DB.Collection(PartnerStatusCollectionName)
	MarketCollection = Sports01DB.Collection(MarketCollectionName)
	MarketStatusCollection = Sports01DB.Collection(MarketStatusCollectionName)
	OperationLockCollection = Sports01DB.Collection(OperationLockCollectionName)
	LoadTestCollection = Sports01DB.Collection(LoadTestCollectionName)
	BFSessionCollection = Sports01DB.Collection(BFSessionCollectionName)
	UserMarketCollection = Sports01DB.Collection(UserMarketCollectionName)
	OperatorMarketCollection = Sports01DB.Collection(OperatorMarketCollectionName)
	log.Println("Connection Opened to Database")
}
