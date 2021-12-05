package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/jgsheppa/golang_website/views"
)

// User to create a new Users controller
// This function will panic if the templates
// are parsed incorrectly
func NewUser() *User {
	return &User{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

type User struct {
	NewView *views.View
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
	Email string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the registration form
//
// POST /register
var decoder = schema.NewDecoder()

func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	var form RegistrationForm
	
	if err := decoder.Decode(&form, r.PostForm); err != nil {
		panic(err)
	}

	fmt.Fprintln(w, form)
}