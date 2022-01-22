package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/controllers"
	"github.com/jgsheppa/golang_website/models"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>We couldn't find that page!</h1><a href=\"/\">Home</a>")
}


const (
	host = "localhost"
	port = 5432
	user = "jamessheppard"
	dbname = "golang"
)

const hmacSecretKey = "secret"

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", 
	host, port, user, dbname)
	us, err := models.NewUserService(psqlInfo, hmacSecretKey)
	if err != nil {
	panic(err)
	}
	must(err)
	
// us.DestructiveReset()
	us.AutoMigrate()

	staticController := controllers.NewStatic()
	userController := controllers.NewUser(us)

	r := mux.NewRouter()
	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.Handle("/about", staticController.About).Methods("GET")
	r.Handle("/register", userController.NewView).Methods("GET")
	r.HandleFunc("/register", userController.Create).Methods("POST")
	r.Handle("/login", userController.LoginView).Methods("GET")
	r.HandleFunc("/login", userController.Login).Methods("POST")
	r.HandleFunc("/cookie", userController.CookieTest).Methods("GET")
	r.Handle("/dashboard", userController.DashboardView).Methods("GET")


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