package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/context"
	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/views"
)

const (
	ShowGallery = "show_gallery"
	EditGallery = "edit_gallery"

	maxMultipartMemory = 1 << 20 // 1 megabyte
)

func NewGallery(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		New: views.NewView("bootstrap", "galleries/new"),
		ShowView: views.NewView("bootstrap", "galleries/show"),
		EditView: views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs: gs,
		is: is,
		r: r,
	}
}

type Galleries struct {
	New *views.View
	ShowView *views.View
	EditView *views.View
	IndexView *views.View
	gs models.GalleryService
	is models.ImageService
	r *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// GET /gallers/:id	
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, r, vd)
}

// GET /galleries
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	

	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}

// POST to /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	gallery := models.Gallery{
		Title: form.Title,
		UserId: user.ID,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	// This is used to get the URL to show galleries and return
	// the correct gallery given a valid ID
	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		// TODO make this go to the index page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)

}

// DELETE /galleries/:id/delete
func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserId != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}
	
	var vd views.Data
	err = g.gs.Delete(gallery.ID)
	if err != nil {
		vd.SetAlert(err)
		vd.Yield = gallery
		g.EditView.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserId != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, r, vd)
}

// POST /galleries/:id/update
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserId != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}

	var form GalleryForm
	vd.Yield = gallery
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	gallery.Title = form.Title

	err = g.gs.Update(gallery)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	vd.Alert = &views.Alert{
		Level: views.AlertLevelSuccess, 
		Message: "Gallery successfully updated!",
	}
	g.EditView.Render(w, r, vd)
}

// POST /galleries/:id/images
func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserId != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}
	
	// TODO: parse a multipart form
	var vd views.Data
	vd.Yield = gallery
	err = r.ParseMultipartForm(maxMultipartMemory)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer file.Close()

		err = g.is.Create(gallery.ID, file, f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

	}
	images, err := g.is.ByGalleryID(gallery.ID)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, "Files:", images)

}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)

	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Whoops! Something went wrong.", http.StatusNotFound)
		}
		return nil, err
	}
	images, err := g.is.ByGalleryID(gallery.ID)
	if err != nil {
		gallery.Images = []string{}
	} else {
		gallery.Images = images
	}
	return gallery, nil
}

func (g *Galleries) GetGalleryJson(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	json, err := json.Marshal(gallery)

	fmt.Println("ERROR", json)

	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}