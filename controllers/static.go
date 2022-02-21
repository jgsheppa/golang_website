package controllers

import (
	"net/http"

	"github.com/jgsheppa/golang_website/views"
)


func NewStatic() *Static {
	return &Static{
		Home: Index,
		Contact: views.NewView("bootstrap", "static/contact"),
		About: views.NewView("bootstrap", "static/about"),
	}
}

type Static struct {
	Home func(http.ResponseWriter, *http.Request)
	Contact *views.View
	About *views.View
}


func Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
	return
}