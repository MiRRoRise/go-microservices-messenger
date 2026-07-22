// @title Chat Service API
// @version 1.0
// @description Chat service for messenger
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MiRRoRise/chat-service/internal/config"
	deliveryHTTP "github.com/MiRRoRise/chat-service/internal/delivery/http"
	"github.com/MiRRoRise/chat-service/internal/repository"
	"github.com/MiRRoRise/chat-service/internal/usecase"
	"github.com/MiRRoRise/chat-service/pkg/jwt"
	"github.com/MiRRoRise/chat-service/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.New()
	logger := logger.New()

	connStr := "user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" host=" + cfg.DBHost +
		" port=" + cfg.DBPort +
		" sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("failed to connect to db: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping DB: ", err)
	}
	logger.Info("connected to PostgreSQL")

	if err := runMigrations(db, cfg.DBName); err != nil {
		logger.Fatal("failed to run migrations", err)
	}
	logger.Info("migrations completed")

	chatRepo := repository.NewChatRepo(db)
	messageRepo := repository.NewMessageRepo(db)

	chatUseCase := usecase.NewChatUseCase(chatRepo)
	messageUseCase := usecase.NewMessageUseCase(messageRepo, chatRepo)
	
	manager := jwt.NewManager(cfg.JWTSecret)

	handler := deliveryHTTP.NewHandler(chatUseCase, messageUseCase, manager)
	router := handler.RegisterRoutes()

	srv := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("chat-service started on port: ", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server: ", err)
		}
	}()

	<-stop
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", err)
	}
	logger.Info("server stopped")
}

func runMigrations(db *sql.DB, DBname string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		DBname,
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
