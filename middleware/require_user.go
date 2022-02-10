package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jgsheppa/golang_website/models"
)

type RequireUser struct {
	models.UserService
}

func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if the user is logged in...
		t := time.Now()
		fmt.Println("Fake request timer:", t)
		cookie, err := r.Cookie("remember_token"); if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fmt.Println("User found:", user)
		next(w,r)

	})
}


func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}