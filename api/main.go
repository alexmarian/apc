package main

import (
	"database/sql"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/logging"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Default log level
	}

	logFile := os.Getenv("LOG_FILE")
	isDevelopment := os.Getenv("ENVIRONMENT") != "production"

	if err := logging.Initialize(logLevel, logFile, isDevelopment); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Logger.Sync()
	err := godotenv.Load(".env")
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "warning: assuming default configuration. .env unreadable", zap.Error(err))
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET environment variable is not set")
	}
	uiOrigin := os.Getenv("UI_ORIGIN")
	if uiOrigin == "" {
		log.Fatal("UI_ORIGIN environment variable is not set")
	}

	apiCfg := &handlers.ApiConfig{
		Secret: secret,
	}
	dbURL := os.Getenv("DB_PATH")
	if dbURL == "" {
		log.Println("DB_PATH environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("sqlite3", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		dbQueries := database.New(db)
		apiCfg.Db = dbQueries
		logging.Logger.Log(zap.InfoLevel, "Connected to database!")
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))

	mux.HandleFunc("POST /v1/api/users", handlers.HandleCreateUserWithToken(apiCfg))
	mux.HandleFunc("POST /v1/api/admin/tokens", apiCfg.MiddlewareAdminOnly(handlers.HandleCreateRegistrationToken(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/admin/tokens/{%s}/revoke", handlers.RegistrationTokenPathValue),
		apiCfg.MiddlewareAdminOnly(handlers.HandleRevokeRegistrationToken(apiCfg)))
	mux.HandleFunc("GET /v1/api/admin/tokens",
		apiCfg.MiddlewareAdminOnly(handlers.HandleGetAllRegistrationTokens(apiCfg)))
	mux.Handle("POST /v1/api/admin/password-reset/request", apiCfg.MiddlewareAdminOnly(handlers.HandleRequestPasswordReset(apiCfg)))
	loginRateLimiter := handlers.NewRateLimiter(5, 5*time.Minute) // 5 attempts per 5 minutes
	// Existing user endpoints
	mux.Handle("PUT /v1/api/users", apiCfg.MiddlewareAuth(handlers.HandleUpdateUser(apiCfg)))
	mux.Handle("POST /v1/api/login", logging.LoggingMiddleware(handlers.MiddlewareRateLimit(loginRateLimiter)(http.HandlerFunc(handlers.HandleLogin(apiCfg)))))

	//resetRateLimiter := handlers.NewRateLimiter(3, 10*time.Minute)
	//mux.Handle("POST /v1/api/password-reset/request", handlers.MiddlewareRateLimit(resetRateLimiter)(handlers.HandleRequestPasswordReset(apiCfg)))
	mux.Handle("POST /v1/api/password-reset/reset", handlers.HandleResetPassword(apiCfg))

	mux.Handle("POST /v1/api/refresh", logging.LoggingMiddleware(handlers.HandleRefresh(apiCfg)))

	mux.HandleFunc("GET /v1/api/associations", apiCfg.MiddlewareAuth(handlers.HandleGetUserAssociations(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetUserAssociation(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationBuildings(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationBuilding(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnits(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnit(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/buildings/{%s}/units/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateBuildingUnit(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/owners", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnitOwner(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/ownerships", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnitOwnerships(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/ownerships/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue, handlers.OwnershipIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetUnitOwnership(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/ownerships/{%s}/voting", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue, handlers.OwnershipIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleUnitVotingOwnership(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/report",
		handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetUnitReport(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/owners", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleCreateOwner(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/owners/{%s}", handlers.AssociationIdPathValue, handlers.OwnerIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateAssociationOwner(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/ownerships", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleCreateUnitOwnership(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/owners/report", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetOwnerReport(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/owners/voters", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetVotersReport(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/owners/{%s}", handlers.AssociationIdPathValue, handlers.OwnerIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationOwner(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/ownerships/{ownershipId}/disable",
		handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDisableOwnership(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/owners", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationOwners(apiCfg)))

	// Categories
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/categories", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetActiveCategories(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/categories/{%s}", handlers.AssociationIdPathValue, handlers.CategoryIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetCategory(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/categories", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateCategory(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/categories/{%s}/deactivate", handlers.AssociationIdPathValue, handlers.CategoryIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDeactivateCategory(apiCfg)))

	// Expenses
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/expenses", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetExpenses(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/expenses/{%s}", handlers.AssociationIdPathValue, handlers.ExpensesIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetExpense(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/expenses", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateExpense(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/expenses/{%s}", handlers.AssociationIdPathValue, handlers.ExpensesIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateExpense(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("DELETE /v1/api/associations/{%s}/expenses/{%s}", handlers.AssociationIdPathValue, handlers.ExpensesIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDeleteExpense(apiCfg)))

	// Expense Reports
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/expenses/report", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetExpenseReport(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/expenses/distribution", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleExpenseDistributionReport(apiCfg)))

	// Accounts
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/accounts", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationAccounts(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/accounts/{%s}", handlers.AssociationIdPathValue, handlers.AccountIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetAccount(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/accounts", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateAccount(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/accounts/{%s}", handlers.AssociationIdPathValue, handlers.AccountIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateAccount(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/accounts/{%s}/disable", handlers.AssociationIdPathValue, handlers.AccountIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDisableAccount(apiCfg)))

	// Gatherings
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetGatherings(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/gatherings", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateGathering(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetGathering(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/gatherings/{%s}/status", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateGatheringStatus(apiCfg)))

	// Voting Matters
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/matters", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetVotingMatters(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/gatherings/{%s}/matters", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateVotingMatter(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/gatherings/{%s}/matters/{%s}", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue, handlers.VotingMatterIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateVotingMatter(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("DELETE /v1/api/associations/{%s}/gatherings/{%s}/matters/{%s}", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue, handlers.VotingMatterIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDeleteVotingMatter(apiCfg)))

	// Participants
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/participants", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetParticipants(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/gatherings/{%s}/participants", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleAddParticipant(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/gatherings/{%s}/participants/{%s}/checkin", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue, handlers.ParticipantIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCheckInParticipant(apiCfg)))

	// Voting (Ballot submission)
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/gatherings/{%s}/ballot", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleSubmitBallot(apiCfg)))

	// Results
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/results", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetVoteResults(apiCfg)))

	// Utility endpoints
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/eligible-voters", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetEligibleVoters(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/qualified-units", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetQualifiedUnits(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/non-participating-owners", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetNonParticipatingOwners(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/stats", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetGatheringStats(apiCfg)))

	// Notifications
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/gatherings/{%s}/notifications", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleSendNotification(apiCfg)))

	// Audit logs
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/gatherings/{%s}/audit-logs", handlers.AssociationIdPathValue, handlers.GatheringIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetAuditLogs(apiCfg)))

	// Ballot verification (public endpoint)
	mux.HandleFunc("POST /v1/api/ballot/verify", handlers.HandleVerifyBallot(apiCfg))

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", uiOrigin)
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux),
	}

	logging.Logger.Log(zap.InfoLevel, "APC api listening on port "+port)
	log.Fatal(srv.ListenAndServe())
}
