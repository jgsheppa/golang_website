package controllers

import (
	"fmt"
	"net/http"

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

// Create is used to process the registration form
//
// POST /register
func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is a temporary response")
}