package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jgsheppa/golang_website/context"
	"github.com/jgsheppa/golang_website/email"
	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/rand"
	"github.com/jgsheppa/golang_website/views"
)

// User to create a new Users controller
// This function will panic if the templates
// are parsed incorrectly
func NewUser(us models.UserService) *User {
	return &User{
		NewView: views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		DashboardView: views.NewView("bootstrap", "users/dashboard"),
		us: us,
	}
}

type User struct {
	NewView *views.View
	LoginView *views.View
	DashboardView *views.View
	us models.UserService
}

// Used to render the /register HTML form
//
// GET /register
func (u *User) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, r, nil)
}

// GET /dashboard and pass in user data
func (u *User) Dashboard(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, r, nil)
}

type RegistrationForm struct {
	Name string `schema:"name"`
	Email string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the registration form
//
// POST /register
func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form RegistrationForm

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	user := models.User{
		Name: form.Name,
		Email: form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusNotFound)
		return 
	}
	email.SendWelcomeEmail()

	http.Redirect(w, r, "/galleries/new", http.StatusFound)
}


type LoginForm struct {
	Email string `schema:"email"`
	Password string `schema:"password"`
}

func (u *User) Login(w http.ResponseWriter, r *http.Request) {
	vd := views.Data{}
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	user, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("Invalid email address")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return 
	}
	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u *User) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name: "remember_token",
		Value: "",
		Expires: time.Now(),
		HttpOnly: true,
		Secure: true,
	}
	http.SetCookie(w, &cookie)

	user := context.User(r.Context())
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.Update(user)

	http.Redirect(w, r, "/", http.StatusFound)
}


func (u *User) GetUser(w http.ResponseWriter, r *http.Request) (*models.User) {
	cookie, err := r.Cookie("remember_token"); if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return user
} 

func (u *User) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name: "remember_token",
		Value: user.Remember,
		HttpOnly: true,
		Secure: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}

func (u *User) GetUserJson(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())

	json, err := json.Marshal(user)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}