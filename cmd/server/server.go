package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	httpadapter "github.com/vini0012/go-client-server-api/internal/server/adapters/inbound/http"
	"github.com/vini0012/go-client-server-api/internal/server/adapters/outbound/awesomeapi"
	"github.com/vini0012/go-client-server-api/internal/server/adapters/outbound/sqlite"
	"github.com/vini0012/go-client-server-api/internal/server/application"
)

const (
	serverAddress    = ":8080"
	databasePath     = "quotations.db"
	migrationTimeout = 5 * time.Second
)

func main() {
	logger := log.New(os.Stdout, "server: ", log.LstdFlags)

	db, err := sqlite.Open(databasePath)
	if err != nil {
		logger.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	repository := sqlite.NewQuotationRepository(db)

	migrationCtx, cancelMigration := context.WithTimeout(
		context.Background(),
		migrationTimeout,
	)
	err = repository.Migrate(migrationCtx)
	cancelMigration()
	if err != nil {
		logger.Fatalf("failed to initialize database: %v", err)
	}

	provider := awesomeapi.NewRateProvider(
		&http.Client{},
		awesomeapi.DefaultEndpoint,
	)

	getQuotation := application.NewGetQuotation(
		provider,
		repository,
	)

	quotationHandler := httpadapter.NewQuotationHandler(
		getQuotation,
		logger,
	)

	mux := http.NewServeMux()
	mux.Handle("/cotacao", quotationHandler)

	server := &http.Server{
		Addr:              serverAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Printf("listening on http://localhost%s", serverAddress)

	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("server stopped: %v", err)
	}
}
