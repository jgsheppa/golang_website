package controllers

import (
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

func (u *User) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}