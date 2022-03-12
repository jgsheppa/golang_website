package controllers

import (
	"net/http"

	"github.com/jgsheppa/golang_website/views"
)


func NewStatic() *Static {
	return &Static{
		Home: Index,
		Contact: views.NewView("bootstrap", http.StatusFound, "static/contact"),
		About: views.NewView("bootstrap", http.StatusFound, "static/about"),
		NotFound: views.NewView("bootstrap", http.StatusNotFound, "static/404"),
	}
}

type Static struct {
	Home func(http.ResponseWriter, *http.Request)
	NotFound *views.View
	Contact *views.View
	About *views.View
}


func Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
	return
}
