package database

import (
	"log"
	"os"
	"time"

	"github.com/pendeploy-simple/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up the GORM database connection
func Initialize() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/pendeploy"
		log.Println("⚠️ No DATABASE_URL environment variable set, using default")
	}

	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	// Connect to database
	var err error
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get and configure the underlying SQL DB
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate models
	err = DB.AutoMigrate(
		&models.Registry{},
		&models.User{},
		&models.Project{},
		&models.Environment{},
		&models.Service{},
		&models.Deployment{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	log.Println("✅ Connected to database")

	// Print connection info
	rows, err := sqlDB.Query("SELECT version()")
	if err == nil {
		var version string
		if rows.Next() {
			if err := rows.Scan(&version); err == nil {
				log.Printf("📊 Database: %s", version)
			}
		}
		rows.Close()
	}
}
