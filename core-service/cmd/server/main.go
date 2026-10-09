package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"core-service/internal/config"
	"core-service/internal/database"
	"core-service/internal/handler"
	"core-service/internal/repository"
	"core-service/internal/router"
	"core-service/internal/security"
	"core-service/internal/service"
	"core-service/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("core-service: database connected and migrated")

	jwtManager := security.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Repositories.
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	// Services.
	userService := service.NewUserService(db, userRepo, walletRepo, jwtManager)
	walletService := service.NewWalletService(walletRepo)
	txService := service.NewTransactionService(db, userRepo, walletRepo, txRepo)

	// Handlers.
	handlers := router.Handlers{
		User:        handler.NewUserHandler(userService),
		Wallet:      handler.NewWalletHandler(walletService),
		Transaction: handler.NewTransactionHandler(txService),
		Internal:    handler.NewInternalHandler(txService),
	}

	// Background workers.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workerMgr := worker.NewManager(txService)
	workerMgr.Start(ctx)

	engine := router.New(handlers, jwtManager, cfg.InternalKey)
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: engine,
	}

	go func() {
		log.Printf("core-service listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("core-service: shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("core-service: stopped")
}
