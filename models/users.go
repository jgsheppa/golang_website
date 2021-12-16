package models

import (
	"errors"

	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("models: resource not found")
	ErrInvalidID = errors.New("models: ID provided was invalid")
)

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return &UserService{
		db: db,
	}, nil
}

type UserService struct {
	db *gorm.DB
}

// Look up a user by ID
// Case 1: User, nil
// Case 2: nil, ErrNotFound
// Case 3: nil, OtherError
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// Will create the provided user
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Update will update the provided user with all of the 
// provided data in the user object
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID:id}}
	return us.db.Delete(&user).Error
}

func (us *UserService) ByEmail(email string) (*User, error) {
	var user User 
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
	
}

func first(db *gorm.DB, dst interface{}) error {	
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

func (us *UserService) DestructiveReset() {
	us.db.Migrator().DropTable(&User{})
	us.db.AutoMigrate(&User{})
}

type User struct {
	gorm.Model
	Name string
	Email string `gorm:"not null;unique_index`
}