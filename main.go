package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/views"
)

var (
	homeView *views.View
	contactView *views.View
	registerView *views.View
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

// func faq(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/html")
// 	must(faqView.Render(w, nil))
// }

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>We couldn't find that page!</h1><a href=\"/\">Home</a>")
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(registerView.Render(w, nil))
}

func main() {
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	registerView = views.NewView("bootstrap", "views/register.gohtml")

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/register", register)

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