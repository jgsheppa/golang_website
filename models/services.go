package models

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewServices(connectionInfo string) (*Services, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
				SlowThreshold:              time.Second,   // Slow SQL threshold
				LogLevel:                   logger.Info, // Log level
				IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,          // Disable color
			},
		)
		
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic(err)
	}
	return &Services{User: NewUserService(db), Gallery: NewGalleryService(db), db: db}, nil
}

type Services struct {
	Gallery GalleryService
	User UserService
	db *gorm.DB
}

// Used to delete old database tables and entries in development
func (s *Services) DestructiveReset() error {
	err := s.db.Migrator().DropTable(&User{}, &Gallery{})
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

// Will attempt to automigrate all database tables
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{})
}