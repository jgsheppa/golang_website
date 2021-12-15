package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host = "localhost"
	port = 5432
	user = "jamessheppard"
	dbname = "golang"
)

type User struct {
	gorm.Model
	Name string
	Email string `gorm:"not null;unique_index"`
	Color string
}

func main () {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", 
	host, port, user, dbname)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:              time.Second,   // Slow SQL threshold
			LogLevel:                   logger.Info, // Log level
			IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,          // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&User{})
	var users []User
	db.Find(&users)
	fmt.Println(users)
}