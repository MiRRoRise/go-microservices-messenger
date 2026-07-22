// @title Auth Service API
// @version 1.0
// @description Authentication service for messenger
// @host localhost:8080
// @BasePath /api/v1
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

	_ "github.com/MiRRoRise/auth-service/docs"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/MiRRoRise/auth-service/internal/config"
	deliveryHTTP "github.com/MiRRoRise/auth-service/internal/delivery/http"
	"github.com/MiRRoRise/auth-service/internal/repository"
	"github.com/MiRRoRise/auth-service/internal/usecase"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	"github.com/MiRRoRise/auth-service/pkg/logger"
	"github.com/MiRRoRise/auth-service/pkg/password"
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

	userRepo := repository.NewUserRepo(db)
	hasher := password.NewBcryptHasher(0)
	tokenManager := jwt.NewManager(cfg.JWTSecret)

	authUseCase := usecase.NewUserUseCase(userRepo, hasher, tokenManager)

	handler := deliveryHTTP.NewHandler(authUseCase)
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
		logger.Info("auth-service started on port: ", cfg.ServerPort)
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
