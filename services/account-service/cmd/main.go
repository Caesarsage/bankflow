package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Caesarsage/bankflow/account-service/internal/handlers"
	"github.com/Caesarsage/bankflow/account-service/internal/kafka"
	"github.com/Caesarsage/bankflow/account-service/internal/repository"
	"github.com/Caesarsage/bankflow/account-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "bankflow")
	dbPassword := getEnv("DB_PASSWORD", "bankflow123")
	dbName := getEnv("DB_NAME", "account_db")
	port := getEnv("PORT", "8002")

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "account-events")

	// Initialize database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Initialize Kafka producer
	brokers := strings.Split(kafkaBrokers, ",")
	producer := kafka.NewProducer(brokers, kafkaTopic)
	defer producer.Close()
	log.Println("Kafka producer initialized")

	// Initialize repository, service, and handler
	repo := repository.NewAccountRepository(db)
	svc := service.NewAccountService(repo, producer)
	handler := handlers.NewAccountHandler(svc)

	// Setup Gin router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "account-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	// Start server
	log.Printf("Account service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
