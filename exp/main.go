package main

import (
	"fmt"

	"github.com/jgsheppa/golang_website/models"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
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

	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold:              time.Second,   // Slow SQL threshold
	// 		LogLevel:                   logger.Info, // Log level
	// 		IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
	// 		Colorful:                  true,          // Disable color
	// 	},
	// )

	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	us.DestructiveReset()
	user := models.User {
		Name: "Michael Scott",
		Email: "michael@dundermifflin.com",
	}
	if err := us.Create(&user); err != nil {
		panic(err)
	}

	user.Email = "dwight@dundermifflin.com"
	if err := us.Update(&user); err != nil {
		panic(err)
	}

	userByEmail, err := us.ByEmail("dwight@dundermifflin.com")
	if err != nil {
		panic(err)
	}
	fmt.Println("USER", userByEmail)	
	
	if err := us.Delete(user.ID); err != nil {
		panic(err)
	}
	
	userByID, err := us.ByID(1)
	if err != nil {
		panic(err)
	}
	fmt.Println("USER", userByID)	
	
}