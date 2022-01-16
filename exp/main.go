package main

import (
	"fmt"

	"github.com/jgsheppa/golang_website/models"
)

const (
	host = "localhost"
	port = 5432
	user = "jamessheppard"
	dbname = "golang"
)
func main() {

	
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", 
	host, port, user, dbname)
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	
	// us.DestructiveReset()
	us.AutoMigrate()
	user := models.User{
		Name: "Jon Calhoun",
		Email: "jon@jon.io",
		Password: "jon",
		Remember: "password",
	}
	err = us.Create(&user)
	if err != nil {
		panic(err)
	}

	user2, err := us.ByRemember("password")
	fmt.Println(user2)
}
