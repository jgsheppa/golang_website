package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/context"
	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/views"
)

// GET /admin
func (u *User) Admin(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := context.User(r.Context())
	if (user.Role != "admin") {
		http.Redirect(w, r, "/galleries", http.StatusFound)
	}

	users, err := u.us.GetAllUsers()
	if err != nil {
		vd.SetAlert(err)
		u.AdminView.Render(w, r, vd)
		return
	}
	
	vd.Yield = users
	u.AdminView.Render(w, r, vd)
}


// DELETE user with /admin/user/:id/delete
func (u *User) UserDelete(w http.ResponseWriter, r *http.Request) {
	user, err := u.userByID(w, r)
	if err != nil {
		return
	}
	
	var vd views.Data
	
	if err := u.us.Delete(user.ID); err != nil {
		vd.SetAlert(err)
		u.AdminView.Render(w, r, vd)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func(u *User) userByID(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	vars := mux.Vars(r)

	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid user ID", http.StatusNotFound)
		return nil, err
	}

	user, err := u.us.ByID(uint(id))
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
	
	return user, nil
}