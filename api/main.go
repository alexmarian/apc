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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Println("APC api listening on port " + port)
	log.Fatal(srv.ListenAndServe())
}
