package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Hello Go!</h1><p>What a crazy world this is!</p>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Contact</h1><a href=\"bob@example.com\">bob@example.com</a>")
}

func pageNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>We could not find the page you were looking for :(</h1><a href=\"/\">Home</a>")
}

// Prints out the route's path in the browser
func pathHandler (w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	default:
		pageNotFoundHandler(w, r)
	}
}

func main() {
	http.HandleFunc("/", pathHandler)
	fmt.Println("Starting the development server on port 3000...")
	http.ListenAndServe(":3000", nil)
}