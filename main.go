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

	// cfg := elasticsearch.Config{
	// 	Addresses: []string{
	// 		"http://localhost:9200",
	// 		"http://localhost:9201",
	// 	},
	// 	Username: "foo",
  // 	Password: "bar",
	// }
	// es, err := elasticsearch.NewClient(cfg)
	// fmt.Println(es)
	// if err != nil {
	// 	log.Fatalf("Error creating the client: %s", err)
	// }
	

	

	services, err := models.NewServices(psqlInfo)
	if err != nil {
	panic(err)
	}
	must(err)

	// services.DestructiveReset()
	services.AutoMigrate()

	emailAddress := os.Getenv("ADMIN_EMAIL")
	name := os.Getenv("ADMIN_NAME")
	password := os.Getenv("ADMIN_PW")

	user := models.User{
		Email: emailAddress,
		Name: name,
		Password: password,
		Role: "admin",
	}

	services.User.Create(&user)

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
	r.Handle("/profile", requiredUserMW.ApplyFn(userController.Profile)).Methods("GET")
	r.Handle("/profile/delete", requiredUserMW.ApplyFn(userController.ProfileDelete)).Methods("POST")


	// Admin Views
	r.Handle("/admin", requiredUserMW.ApplyFn(userController.Admin)).Methods("GET")
	r.HandleFunc("/admin/user/{id:[0-9]+}/delete", requiredUserMW.ApplyFn(userController.UserDelete)).Methods("POST")

	// Logout action
	r.HandleFunc("/logout", requiredUserMW.ApplyFn(userController.Logout)).Methods("POST")

	// HandlerFunc converts notFound to the correct type
	r.NotFoundHandler = staticController.NotFound
	fmt.Println("Starting the development server on port" + port)
	http.ListenAndServe(":" + port, csrfMw(userMw.Apply(r)))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}