package middleware

import (
	"net/http"
	"strings"

	"github.com/jgsheppa/golang_website/context"
	"github.com/jgsheppa/golang_website/models"
)

type User struct {
	models.UserService
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}


func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// If the user is requesting a static asset
		// we will not need to look up the user in the DB
		if strings.HasPrefix(path, "/assets/") || 
		strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}
		// if the user is logged in then pass the user for the navbar
		cookie, err := r.Cookie("remember_token"); if err != nil {
			next(w, r)
			return
		}
		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next(w,r)
	})
}

// Assumes that user has already been run, 
// otherwise it will not work correctly
type RequireUser struct {
	User
}

// Assumes that user has already been run, 
// otherwise it will not work correctly
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w,r)
	})
}

// Assumes that user has already been run, 
// otherwise it will not work correctly
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}