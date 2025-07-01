package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pendeploy-simple/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConnection represents a database connection
type DBConnection struct {
	DB     *gorm.DB
	Name   string
	DbURL  string
	Models []interface{}
}

// NewDBConnection creates a new database connection
func NewDBConnection(name, dbURL string) (*DBConnection, error) {
	if dbURL == "" {
		return nil, errors.New("database URL cannot be empty")
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
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %v", name, err)
	}

	// Get and configure the underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB for %s: %v", name, err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("✅ Connected to %s database", name)

	// Print connection info
	rows, err := sqlDB.Query("SELECT version()")
	if err == nil {
		var version string
		if rows.Next() {
			if err := rows.Scan(&version); err == nil {
				log.Printf("📊 %s Database: %s", name, version)
			}
		}
		rows.Close()
	}

	// Define models to migrate
	models := []interface{}{
		&models.Registry{},
		&models.User{},
		&models.Project{},
		&models.Environment{},
		&models.Service{},
		&models.Deployment{},
	}

	return &DBConnection{
		DB:     db,
		Name:   name,
		DbURL:  dbURL,
		Models: models,
	}, nil
}

// Migrate migrates the database schema
func (c *DBConnection) Migrate() error {
	log.Printf("Migrating %s database schema...", c.Name)
	err := c.DB.AutoMigrate(c.Models...)
	if err != nil {
		return fmt.Errorf("failed to migrate %s database: %v", c.Name, err)
	}
	log.Printf("✅ %s database schema migrated", c.Name)
	return nil
}

// MigrateDataBetweenDatabases migrates data from source to target
func MigrateDataBetweenDatabases(source, target *DBConnection) error {
	log.Println("Starting data migration from source to target...")

	// // Step 1: Migrate Registries
	// var registries []models.Registry
	// if err := source.DB.Find(&registries).Error; err != nil {
	// 	return fmt.Errorf("failed to fetch registries: %v", err)
	// }
	// log.Printf("Found %d registries to migrate", len(registries))
	// if len(registries) > 0 {
	// 	if err := target.DB.Create(&registries).Error; err != nil {
	// 		return fmt.Errorf("failed to migrate registries: %v", err)
	// 	}
	// }

	// // Step 2: Migrate Users
	// var users []models.User
	// if err := source.DB.Find(&users).Error; err != nil {
	// 	return fmt.Errorf("failed to fetch users: %v", err)
	// }
	// log.Printf("Found %d users to migrate", len(users))
	// if len(users) > 0 {
	// 	if err := target.DB.Create(&users).Error; err != nil {
	// 		return fmt.Errorf("failed to migrate users: %v", err)
	// 	}
	// }

	// // Step 3: Migrate Projects
	// var projects []models.Project
	// if err := source.DB.Find(&projects).Error; err != nil {
	// 	return fmt.Errorf("failed to fetch projects: %v", err)
	// }
	// log.Printf("Found %d projects to migrate", len(projects))
	// if len(projects) > 0 {
	// 	if err := target.DB.Create(&projects).Error; err != nil {
	// 		return fmt.Errorf("failed to migrate projects: %v", err)
	// 	}
	// }

	// Step 4: Migrate Environments
	var environments []models.Environment
	if err := source.DB.Find(&environments).Error; err != nil {
		return fmt.Errorf("failed to fetch environments: %v", err)
	}
	log.Printf("Found %d environments to migrate", len(environments))
	if len(environments) > 0 {
		if err := target.DB.Create(&environments).Error; err != nil {
			return fmt.Errorf("failed to migrate environments: %v", err)
		}
	}

	// Step 5: Migrate Services
	var services []models.Service
	if err := source.DB.Find(&services).Error; err != nil {
		return fmt.Errorf("failed to fetch services: %v", err)
	}
	log.Printf("Found %d services to migrate", len(services))
	if len(services) > 0 {
		if err := target.DB.Create(&services).Error; err != nil {
			return fmt.Errorf("failed to migrate services: %v", err)
		}
	}

	// Step 6: Migrate Deployments
	var deployments []models.Deployment
	if err := source.DB.Find(&deployments).Error; err != nil {
		return fmt.Errorf("failed to fetch deployments: %v", err)
	}
	log.Printf("Found %d deployments to migrate", len(deployments))
	if len(deployments) > 0 {
		if err := target.DB.Create(&deployments).Error; err != nil {
			return fmt.Errorf("failed to migrate deployments: %v", err)
		}
	}

	log.Println("✅ Data migration completed successfully!")
	return nil
}
