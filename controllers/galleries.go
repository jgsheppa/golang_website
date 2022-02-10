package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jgsheppa/golang_website/context"
	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/views"
)


func NewGallery(gs models.GalleryService) *Galleries {
	return &Galleries{
		New: views.NewView("bootstrap", "galleries/new"),
		gs: gs,
	}
}

type Galleries struct {
	New *views.View
	gs models.GalleryService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// POST to /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	fmt.Println("Create got the user:", user)
	gallery := models.Gallery{
		Title: form.Title,
		UserId: user.ID,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	fmt.Fprintln(w, gallery)

}