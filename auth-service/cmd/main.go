package main

import (
	"context"
	"database/sql"
	"net"
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
	grpcDelivery "github.com/MiRRoRise/auth-service/internal/delivery/grpc"
	deliveryHTTP "github.com/MiRRoRise/auth-service/internal/delivery/http"
	"github.com/MiRRoRise/auth-service/internal/kafka"
	"github.com/MiRRoRise/auth-service/internal/repository"
	"github.com/MiRRoRise/auth-service/internal/usecase"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	"github.com/MiRRoRise/auth-service/pkg/logger"
	"github.com/MiRRoRise/auth-service/pkg/password"
	pb "github.com/MiRRoRise/auth-service/proto/auth"
	"google.golang.org/grpc"
)

// @title Auth Service API
// @version 1.0
// @description Authentication service for messenger
// @host localhost:9080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
		logger.Fatal("failed to connect to db", err)
	}
	defer db.Close()

	if pingErr := db.Ping(); pingErr != nil {
		logger.Fatal("failed to ping DB", pingErr)
	}
	logger.Info("connected to PostgreSQL")

	if migErr := runMigrations(db, cfg.DBName); migErr != nil {
		logger.Fatal("failed to run migrations", migErr)
	}
	logger.Info("migrations completed")

	userRepo := repository.NewUserRepo(db)
	hasher := password.NewBcryptHasher(0)
	tokenManager := jwt.NewManager(cfg.JWTSecret)

	kafkaProducer, err := kafka.NewProducer([]string{cfg.KafkaBrokers})
	if err != nil {
		logger.Fatal("failed to create kafka producer", err)
	}
	defer kafkaProducer.Close()
	logger.Info("connected to kafka")

	authUseCase := usecase.NewUserUseCase(userRepo, hasher, tokenManager, kafkaProducer, logger)

	handler := deliveryHTTP.NewHandler(authUseCase, tokenManager)
	router := handler.RegisterRoutes()

	srv := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	grpcServer := grpc.NewServer()
	grpcHandler := grpcDelivery.NewServer(authUseCase)
	pb.RegisterAuthServiceServer(grpcServer, grpcHandler)

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCPort)
		if err != nil {
			logger.Fatal("failed to listen grpc", err)
		}

		logger.Info("grpc server started", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("failed to serve grpc", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("auth-service started", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", err)
		}
	}()

	<-stop
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

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
