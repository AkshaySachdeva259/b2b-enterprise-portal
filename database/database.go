package database

import (
	"log"
	"os"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() *gorm.DB {
	dsn := "postgresql://postgres:Jetpac_global@db.yzsmzrumvcikprukumud.supabase.co:5432/postgres"

	logLevel := logger.Silent
	if os.Getenv("APP_ENV") != "production" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("failed to initialise gorm: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	ensureTables(db)

	DB = db
	log.Println("database connected successfully")
	return db
}

func ensureTables(db *gorm.DB) {
	modelsToEnsure := []struct {
		model   interface{}
		columns []string
	}{
		{model: &models.Destination{}},
		{model: &models.Catalog{}},
		{model: &models.Cart{}},
		{model: &models.CartItem{}},
		{model: &models.Esim{}, columns: []string{"TenantID"}},
		{model: &models.B2BAllocation{}},
	}

	for _, item := range modelsToEnsure {
		if db.Migrator().HasTable(item.model) {
			ensureColumns(db, item.model, item.columns)
			continue
		}

		if err := db.AutoMigrate(item.model); err != nil {
			log.Fatalf("failed to auto-migrate database: %v", err)
		}
	}
}

func ensureColumns(db *gorm.DB, model interface{}, fields []string) {
	for _, field := range fields {
		if db.Migrator().HasColumn(model, field) {
			continue
		}

		if err := db.Migrator().AddColumn(model, field); err != nil {
			log.Fatalf("failed to auto-migrate database: %v", err)
		}
	}
}
