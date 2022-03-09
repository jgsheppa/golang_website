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
func NewUser(us models.UserService, emailer *email.Client) *User {
	return &User{
		NewView: views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		ProfileView: views.NewView("bootstrap", "users/profile"),
		AdminView: views.NewView("bootstrap", "admin/index"),
		us: us,
		emailer: emailer,
	}
}

type User struct {
	NewView *views.View
	LoginView *views.View
	ProfileView *views.View
	AdminView *views.View
	us models.UserService
	emailer *email.Client
}

type RegistrationForm struct {
	Name string `schema:"name"`
	Email string `schema:"email"`
	Password string `schema:"password"`
}

// Used to render the /register HTML form
//
// GET /register
func (u *User) New(w http.ResponseWriter, r *http.Request) {
	var form RegistrationForm
	parseURLParams(r, &form)
	u.NewView.Render(w, r, form)
}


// Create is used to process the registration form
//
// POST /register
func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form RegistrationForm
	vd.Yield = &form

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
	user.Role = "user"
	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	// TODO use real user emails once I am done testing with Mailgun sandbox
	u.emailer.Welcome(user.Name, "jgsheppard92@gmail.com")

	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusNotFound)
		return 
	}

	alert := views.Alert{
		Level: views.AlertLevelSuccess,
		Message: "Welcome to Schnup! You've successfully created your account.",
	}
	views.RedirectAlert(w, r, "/galleries/new", http.StatusFound, alert)
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

	if user.Role != "admin" {
		http.Redirect(w, r, "/galleries", http.StatusFound)
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
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

// GET /profile
func (u *User) Profile(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	
	var vd views.Data
	vd.Yield = user
	u.ProfileView.Render(w, r, vd)
}

// DELETE /profile
func (u *User) ProfileDelete(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := context.User(r.Context())
	
	if err := u.us.Delete(user.ID); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	cookie := http.Cookie{
		Name: "remember_token",
		Value: "",
		Expires: time.Now(),
		HttpOnly: true,
		Secure: true,
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/login", http.StatusFound)
}

