package main

import (
	"log"
	"os"

	"github.com/pendeploy-simple/database"
)

func main() {
	log.Println("Starting database migration...")

	// Source database (current Railway database)
	sourceDBURL := os.Getenv("SOURCE_DATABASE_URL")
	if sourceDBURL == "" {
		sourceDBURL = "postgresql://postgres:aQRjEOonbXrIWKgnaUFzIveFNHkwAXME@hopper.proxy.rlwy.net:35713/railway"
		log.Println("Using default SOURCE_DATABASE_URL from Railway")
	}

	// Target database (new isacitra.com database)
	targetDBURL := os.Getenv("TARGET_DATABASE_URL")
	if targetDBURL == "" {
		// Try with port 443 and SSL mode
		targetDBURL = "postgresql://user_alegf6a6:ls6nC1KRm3+QuS59@postgres-core-52bda5.managed.app.isacitra.com:5432/db_fwfcxlfq"
		
	}

	// Connect to source database
	sourceDB, err := database.NewDBConnection("source", sourceDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to source database: %v", err)
	}

	// Connect to target database
	targetDB, err := database.NewDBConnection("target", targetDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to target database: %v", err)
	}

	// Ensure target database schema is migrated
	if err := targetDB.Migrate(); err != nil {
		log.Fatalf("Failed to migrate target database schema: %v", err)
	}

	// Migrate data from source to target
	if err := database.MigrateDataBetweenDatabases(sourceDB, targetDB); err != nil {
		log.Fatalf("Data migration failed: %v", err)
	}

	log.Println("Database migration completed successfully!")
}
