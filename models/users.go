package models

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/jgsheppa/golang_website/hash"
	"github.com/jgsheppa/golang_website/rand"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrNotFound = errors.New("models: resource not found")
	ErrInvalidID = errors.New("models: ID provided was invalid")
	ErrInvalidPassword = errors.New("models: incorrect password")
)

const userPwPepper = "?3o!yM$LmRKmQhDD"
const hmacSecretKey = "secret-hmac-key"

func NewUserService(connectionInfo string) (*UserService, error) {
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
	hmac := hash.NewHMAC(hmacSecretKey)
	return &UserService{
		db: db,
		hmac: hmac,
	}, nil
}

type UserService struct {
	db *gorm.DB
	hmac hash.HMAC
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
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Store hash as a string in the user model
	user.PasswordHash = string(hashedBytes)
	// Not required, but this prevents accidental printing of logs
	user.Password = ""
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
	return us.db.Create(user).Error
}

// Update will update the provided user with all of the 
// provided data in the user object
func (us *UserService) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
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

func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := us.hmac.Hash(token)
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password + userPwPepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}
	return foundUser, nil
}

func first(db *gorm.DB, dst interface{}) error {	
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

func (us *UserService) DestructiveReset() error {
	if err := us.db.Migrator().DropTable(&User{}); err != nil {
		return err
	}
	return us.AutoMigrate()
}

func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}); err != nil {
		return err
	}
	return nil
}

type User struct {
	gorm.Model
	Name string
	Email string `gorm:"not null;unique"`
	Password string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember string `gorm:"-"`
	RememberHash string `gorm:"not null;unique"`

}