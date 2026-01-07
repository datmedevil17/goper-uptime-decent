package database

import (
	"log"
	"time"
	
	"github.com/datmedevil17/gopher-uptime/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, err
	}

	// Get underlying SQL DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connected successfully")
	return db, nil
}

// AutoMigrate creates all tables
func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running auto-migration...")
	
	err := db.AutoMigrate(
		&models.User{},
		&models.Validator{},
		&models.Website{},
		&models.WebsiteTick{},
		&models.PayoutTransaction{},
	)
	
	if err != nil {
		return err
	}
	
	log.Println("✅ Migration completed successfully")
	return nil
}