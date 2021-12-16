package controllers

import "github.com/jgsheppa/golang_website/views"


func NewStatic() *Static {
	return &Static{
		Home: views.NewView("bootstrap", "static/home"),
		Contact: views.NewView("bootstrap", "static/contact"),
		About: views.NewView("bootstrap", "static/about"),
	}
}

type Static struct {
	Home *views.View
	Contact *views.View
	About *views.View
}