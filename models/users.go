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


// UserDB is used to interact with the database.
// As a general rule, any error but ErrNotFound should
// result in a 500 error
type UserDB interface {
	// Methods for querying for single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// CRUD operations for user
	Create(user *User) error
	Update(user *User) error
	Delete (id uint) error
	
	// Migration helpers
	AutoMigrate() error
	DestructiveReset() error

}

func NewUserService(connectionInfo string) (*UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	return &UserService{
	UserDB: &userValidator{
		UserDB: ug,
		},
	}, nil
}

type userValidator struct {
	UserDB
}

type UserService struct {
	UserDB
}

func newUserGorm(connectionInfo string)  (*userGorm, error) {
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
	return &userGorm{
		db: db,
		hmac: hmac,
	}, nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
	hmac hash.HMAC
}

// Look up a user by ID
// Case 1: User, nil
// Case 2: nil, ErrNotFound
// Case 3: nil, OtherError
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// Will create the provided user
func (ug *userGorm) Create(user *User) error {
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
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}
	return ug.db.Create(user).Error
}

// Update will update the provided user with all of the 
// provided data in the user object
func (ug *userGorm) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}
	return ug.db.Save(user).Error
}

func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID  
	}
	user := User{Model: gorm.Model{ID:id}}
	return ug.db.Delete(&user).Error
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User 
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := ug.hmac.Hash(token)
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
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

func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.Migrator().DropTable(&User{}); err != nil {
		return err
	}
	return ug.AutoMigrate()
}

func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}); err != nil {
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