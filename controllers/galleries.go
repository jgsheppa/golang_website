package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/context"
	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/redis"
	"github.com/jgsheppa/golang_website/views"
)

const (
	ShowGallery = "show_gallery"
	EditGallery = "edit_gallery"

	maxMultipartMemory = 1 << 20 // 1 megabyte
)

func NewGallery(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		New: views.NewView("bootstrap", http.StatusFound, "galleries/new"),
		ShowView: views.NewView("bootstrap", http.StatusFound, "galleries/show"),
		EditView: views.NewView("bootstrap", http.StatusFound, "galleries/edit"),
		IndexView: views.NewView("bootstrap", http.StatusFound, "galleries/index"),
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
	var vd views.Data
	
	rememberCookie, err := r.Cookie("remember_token")
	if err != nil {
		vd.SetAlert(err)
		return
	}

	id := mux.Vars(r)["id"]
	redis, err := redis.NewRedis()
	
	if err != nil {
		vd.SetAlert(err)
		return
	}

	val, err := redis.GetGalleryID(r.Context(), id)
	if err == nil {		
		vd.Yield = val
		g.ShowView.Render(w, r, vd)
		return
	}

	
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	gallery.UserToken = "remember_token=" + rememberCookie.Value
	
	_ = redis.SetGalleryID(r.Context(), gallery) 
	
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
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	user := context.User(r.Context())
	
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
		http.Redirect(w, r, "/galleries", http.StatusFound)
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
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	gallery.Images = images
	vd.Yield = gallery

	g.EditView.Render(w, r, vd)

}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)

	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Whoops! Something went wrong.", http.StatusNotFound)
		}
		return nil, err
	}
	images, _ := g.is.ByGalleryID(gallery.ID)
	gallery.Images = images
	return gallery, nil
}

// POST /galleries/:id/images/:filename/delete
func (g *Galleries) ImageDelete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserId != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}
	
	filename := mux.Vars(r)["filename"]
	
	userImage := models.Image{
		Filename: filename,
		GalleryID: gallery.ID,
	}

	err = g.is.Delete(&userImage)
	if err != nil {
		var vd views.Data
		vd.Yield = gallery
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Galleries) GetGalleryJson(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	redis, err := redis.NewRedis()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	val, err := redis.GetGalleryID(r.Context(), id)
	if err == nil {		
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(val)
		return
	}

	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	_ = redis.SetGalleryID(r.Context(), gallery) 

	for image := range gallery.Images {
		// TODO change host depending on environment
		host := os.Getenv("HOST")
		gallery.Images[image].URL = host + "/images/galleries/" + strconv.FormatUint(uint64(gallery.Images[image].GalleryID), 10) + "/" + gallery.Images[image].Filename
	}

	json, err := json.Marshal(gallery)

	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
