package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Caesarsage/bankflow/identity-service/internal/handlers"
	"github.com/Caesarsage/bankflow/identity-service/internal/kafka"
	"github.com/Caesarsage/bankflow/identity-service/internal/middleware"
	"github.com/Caesarsage/bankflow/identity-service/internal/repository"
	"github.com/Caesarsage/bankflow/identity-service/internal/service"
	"github.com/Caesarsage/bankflow/identity-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	port := getEnv("PORT", "8001")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "bankflow")
	dbPassword := getEnv("DB_PASSWORD", "bankflow123")
	dbName := getEnv("DB_NAME", "identity_db")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	jwtExpiry := getEnv("JWT_EXPIRY", "15m")
	refreshExpiry := getEnv("REFRESH_TOKEN_EXPIRY", "168h")

	// Kafka configuration
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "identity-events")

	// Parse durations
	jwtDuration, err := time.ParseDuration(jwtExpiry)
	if err != nil {
		log.Fatalf("Invalid JWT_EXPIRY: %v", err)
	}

	refreshDuration, err := time.ParseDuration(refreshExpiry)
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_EXPIRY: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Falied to ping database: ", err)
	}
	log.Println("Connected to database")

	// Initialize Kafka producer
	brokers := strings.Split(kafkaBrokers, ",")
	kafkaProducer := kafka.NewProducer(brokers, kafkaTopic)
	defer kafkaProducer.Close()
	log.Println("Connected to Kafka")

	// Initialize dependencies
	jwtManager := jwt.NewJWTManager(jwtSecret, jwtDuration, refreshDuration)
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtManager, kafkaProducer)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", authHandler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	authHandler.RegisterRoutes(v1, jwtManager)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf(" Identity Service starting on port %s", port)
		log.Printf(" Environment: %s", getEnv("ENV", "development"))
		log.Printf(" Database: %s:%s/%s", dbHost, dbPort, dbName)
		log.Printf(" Kafka: %s (topic: %s)", kafkaBrokers, kafkaTopic)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
