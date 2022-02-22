package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/jgsheppa/golang_website/controllers"
	"github.com/jgsheppa/golang_website/email"
	"github.com/jgsheppa/golang_website/middleware"
	"github.com/jgsheppa/golang_website/models"
	"github.com/jgsheppa/golang_website/rand"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>We couldn't find that page!</h1><a href=\"/\">Home</a>")
}



func main() {
	rand.CheckForEnvFile()

	// Necessary for Heroku deploy
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	sentryDsn := os.Getenv("SENTRY_SDK")
	// User to log any errors to Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDsn,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	psqlInfo := os.Getenv("DATABASE_URL")

	services, err := models.NewServices(psqlInfo)
	if err != nil {
	panic(err)
	}
	must(err)

	// services.DestructiveReset()
	services.AutoMigrate()

	domain := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_PRIVATE_KEY")

	emailer := email.NewClient(
		email.WithSender("Schnup Support", "jgsheppard92@gmail.com"),
		email.WithMailgun(domain, apiKey),
	)
	
	r := mux.NewRouter()
	staticController := controllers.NewStatic()
	userController := controllers.NewUser(services.User, emailer)
	galleriesC := controllers.NewGallery(services.Gallery, services.Image, r)

	// TODO: Update to config var
	prodEnv := os.Getenv("IS_PROD")
	isProd, err := strconv.ParseBool(prodEnv)

	if err != nil {
			log.Fatal(err)
	}
	b, err := rand.Bytes(32)
	must(err)
	csrfMw := csrf.Protect(b, csrf.Secure(isProd))
	// Middleware to protect routes
	userMw := middleware.User{
		UserService: services.User,
	}
	requiredUserMW := middleware.RequireUser{
		User: userMw,
	}

	noUserMW := middleware.NoUser{
		User: userMw,
	}

	r.HandleFunc("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	// r.Handle("/about", staticController.About).Methods("GET")
	r.HandleFunc("/register", userController.New).Methods("GET")
	r.HandleFunc("/register", userController.Create).Methods("POST")
	r.Handle("/login", noUserMW.Apply(userController.LoginView)).Methods("GET")
	r.HandleFunc("/login", noUserMW.ApplyFn(userController.Login)).Methods("POST")

	// JSON Routes
	r.HandleFunc("/me", requiredUserMW.ApplyFn(userController.GetUserJson)).Methods("GET")
	r.HandleFunc("/me/galleries/{id:[0-9]+}", requiredUserMW.ApplyFn(galleriesC.GetGalleryJson)).Methods("GET")
	
	// Image routes
	imageHandler := http.FileServer(http.Dir("./images"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Assets	
	assetHandler := http.FileServer(http.Dir("./assets/"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetHandler))
	
	// Gallery Views
	r.Handle("/galleries", requiredUserMW.ApplyFn(galleriesC.Index)).Methods("GET")
	r.Handle("/galleries/new", requiredUserMW.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", requiredUserMW.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requiredUserMW.ApplyFn(galleriesC.Edit)).Methods("GET").Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requiredUserMW.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requiredUserMW.ApplyFn(galleriesC.ImageUpload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requiredUserMW.ApplyFn(galleriesC.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requiredUserMW.ApplyFn(galleriesC.ImageDelete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGallery)
	r.Handle("/dashboard", requiredUserMW.Apply(userController.DashboardView)).Methods("GET")

	// Logout action
	r.HandleFunc("/logout", requiredUserMW.ApplyFn(userController.Logout)).Methods("POST")

	// HandlerFunc converts notFound to the correct type
	r.NotFoundHandler = http.HandlerFunc(notFound)
	fmt.Println("Starting the development server on port" + port)
	http.ListenAndServe(":" + port, csrfMw(userMw.Apply(r)))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}