package controllers

import (
	"fmt"
	"net/http"

	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/views"
)

// User to create a new Users controller
// This function will panic if the templates
// are parsed incorrectly
func NewUser(us *models.UserService) *User {
	return &User{
		NewView: views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us: us,
	}
}

type User struct {
	NewView *views.View
	LoginView *views.View
	us *models.UserService
}

// Used to render the /register HTML form
//
// GET /register
func (u *User) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}

type RegistrationForm struct {
	Name string `schema:"name"`
	Email string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the registration form
//
// POST /register
func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	var form RegistrationForm

	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	user := models.User{
		Name: form.Name,
		Email: form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, user)
}


type LoginForm struct {
	Email string `schema:"email"`
	Password string `schema:"password"`
}

func (u *User) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	fmt.Fprintln(w, form)
}