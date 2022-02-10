package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/controllers"
	"github.com/jgsheppa/golang_website/middleware"
	"github.com/jgsheppa/golang_website/models"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>We couldn't find that page!</h1><a href=\"/\">Home</a>")
}


const (
	host = "127.0.0.1"
	port = 5432
	user = "jamessheppard"
	dbname = "golang_website"
	password = "password"
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", 
	host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo)
	if err != nil {
	panic(err)
	}
	must(err)

	services.DestructiveReset()
	services.AutoMigrate()

	staticController := controllers.NewStatic()
	userController := controllers.NewUser(services.User)
	galleriesC := controllers.NewGallery(services.Gallery)

	r := mux.NewRouter()
	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.Handle("/about", staticController.About).Methods("GET")
	r.HandleFunc("/register", userController.New).Methods("GET")
	r.HandleFunc("/register", userController.Create).Methods("POST")
	r.Handle("/login", userController.LoginView).Methods("GET")
	r.HandleFunc("/login", userController.Login).Methods("POST")
	r.HandleFunc("/cookie", userController.CookieTest).Methods("GET")
	
	// Gallery routes
	requiredUserMW := middleware.RequireUser{
		UserService: services.User,
	}
	r.Handle("/galleries/new", requiredUserMW.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", requiredUserMW.ApplyFn(galleriesC.Create)).Methods("POST")
	r.Handle("/dashboard", requiredUserMW.Apply(userController.DashboardView)).Methods("GET")


	// HandlerFunc converts notFound to the correct type
	r.NotFoundHandler = http.HandlerFunc(notFound)
	fmt.Println("Starting the development server on port 3000...")
	http.ListenAndServe(":3000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}