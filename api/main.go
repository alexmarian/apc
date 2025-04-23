package main

import (
	"database/sql"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET environment variable is not set")
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
		log.Println("Connected to database!")
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))

	mux.Handle("POST /v1/api/admin/user", http.FileServer(http.Dir("static")))

	mux.HandleFunc("POST /v1/api/users", handlers.HandleCreateUser(apiCfg))
	mux.HandleFunc("PUT /v1/api/users", apiCfg.MiddlewareAuth(handlers.HandleUpdateUser(apiCfg)))
	mux.HandleFunc("POST /v1/api/login", handlers.HandleLogin(apiCfg))
	mux.HandleFunc("POST /v1/api/refresh", handlers.HandleRefresh(apiCfg))

	mux.HandleFunc("GET /v1/api/associations", apiCfg.MiddlewareAuth(handlers.HandleGetUserAssociations(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetUserAssociation(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationBuildings(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationBuilding(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnits(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnit(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/buildings/{%s}/units/{%s}", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateBuildingUnit(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/owners", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnitOwner(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/ownerships", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetBuildingUnitOwnerships(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/owners", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleCreateOwner(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/buildings/{%s}/units/{%s}/ownerships", handlers.AssociationIdPathValue, handlers.BuildingIdPathValue, handlers.UnitIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleCreateUnitOwnership(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/owners/report", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetOwnerReport(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/ownerships/{ownershipId}/disable",
		handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDisableOwnership(apiCfg)))

	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/owners", handlers.AssociationIdPathValue), apiCfg.MiddlewareAssociationResource(handlers.HandleGetAssociationOwners(apiCfg)))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Categories
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/categories", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateCategory(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/categories/{categoryId}/deactivate", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDeactivateCategory(apiCfg)))

	// Expenses
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/expenses", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetExpenses(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("POST /v1/api/associations/{%s}/expenses", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleCreateExpense(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("PUT /v1/api/associations/{%s}/expenses/{expenseId}", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleUpdateExpense(apiCfg)))
	mux.HandleFunc(fmt.Sprintf("DELETE /v1/api/associations/{%s}/expenses/{expenseId}", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleDeleteExpense(apiCfg)))

	// Expense Reports
	mux.HandleFunc(fmt.Sprintf("GET /v1/api/associations/{%s}/expenses/report", handlers.AssociationIdPathValue),
		apiCfg.MiddlewareAssociationResource(handlers.HandleGetExpenseReport(apiCfg)))

	log.Println("APC api listening on port " + port)
	log.Fatal(srv.ListenAndServe())
}
