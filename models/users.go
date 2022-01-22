package models

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
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
	ErrRememberTokenTooShort = errors.New("models: remember token must be at least 32 bytes")
	ErrRememberRequired = errors.New("models: remember token required")
	ErrEmailRequired = errors.New("Email address is required")
	ErrEmailInvalid = errors.New("Email address is not valid")
	ErrEmailTaken = errors.New("models: email address is already taken")
)

const userPwPepper = "?3o!yM$LmRKmQhDD"


// User model which stores user name, email address, 
// password hash, and remember hash in the PSQL database.
type User struct {
	gorm.Model
	Name string
	Email string `gorm:"not null;unique"`
	Password string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember string `gorm:"-"`
	RememberHash string `gorm:"not null;unique"`

}

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

// UserService is a set of methods used to manipulate and work with the user model
type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(connectionInfo string, hmacKey string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacKey)
	uv := newUserValidator(ug, hmac)
	
	return &userService{
		UserDB: uv,
	}, nil
}

type userValFunc func(*User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// This pattern ensures the type 
// and the pointer are always aligned
var _ UserDB = &userValidator{}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB: udb,
		hmac: hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

type userValidator struct {
	UserDB
	hmac hash.HMAC
	emailRegex *regexp.Regexp
}

// User to normalize the email on login and creation of user
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFuncs(&user, uv.setRememberIfUnset, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

// Will create the provided user
func (uv *userValidator) Create(user *User) error {
	err := runUserValFuncs(
		user, 
		uv.bcryptPassword,  
		uv.setRememberIfUnset,
		uv.rememberMinBytes, 
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvailable,
		); 
	if err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

// Update will update the provided user with all of the 
// provided data in the user object
func (uv *userValidator) Update(user *User) error {
	err := runUserValFuncs(user, 
		uv.bcryptPassword, 
		uv.rememberMinBytes, 
		uv.hmacRemember, 
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvailable,
		) 
	if err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValFuncs(&user, uv.idGreaterThanZero)
	if err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

func (uv *userValidator) idGreaterThanZero(user *User) error {
	if user.ID <= 0 {
		return ErrInvalidID
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		// Email is not taken; proceed with login/register process
		return nil
	}
	if err != nil {
		return err
	}
	// We found a user with this email address
	// If the user id is the same then we treat this as an update
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}

	return nil
}


// Used during user validation to hash the user's submitted password
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Store hash as a string in the user model
	user.PasswordHash = string(hashedBytes)
	// Not required, but this prevents accidental printing of logs
	user.Password = ""
	
	return nil
}

// There is no remember token, this function provides one
func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}


func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.Remember == "" {
		return ErrRememberRequired
	}
	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == ""{
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTokenTooShort
	}
	return nil
}

var _ UserDB = &userService{}

type userService struct {
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
	return &userGorm{
		db: db,
	}, nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
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

// Will be used to create the provided user
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update will update the provided user with all of the 
// provided data in the user object
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID:id}}
	return ug.db.Delete(&user).Error
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User 
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// This method will be used to match a user with their remember token
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (us *userService) Authenticate(email, password string) (*User, error) {
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


// Used to delete an old database tables and entries in development
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.Migrator().DropTable(&User{}); err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// Used to migrate the model
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}); err != nil {
		return err
	}
	return nil
}
