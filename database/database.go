package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	// Open via lib/pq so pq.StringArray works correctly for text[] columns.
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	logLevel := logger.Silent
	if os.Getenv("APP_ENV") != "production" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("failed to initialise gorm: %v", err)
	}

	DB = db
	log.Println("database connected successfully")
	return db
}
