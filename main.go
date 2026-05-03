package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"photos.com/controllers"
	"photos.com/models"
)

func main() {
	cfg := LoadConfig(false)
	dbCfg := cfg.Database

	db, err := models.Open(dbCfg.ConnectionInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// create NewUserService
	userDBS := models.UserDBService{
		DB: db,
	}
	usersC := controllers.NewUserController(userDBS)
	staticC := controllers.NewStaticController()

	mux := http.NewServeMux()
	// asset files
	fs := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.Handle("/", staticC.Home)
	mux.Handle("/contact", staticC.Contact)
	mux.Handle("/about", staticC.About)
	mux.HandleFunc("GET /signup", usersC.New)
	mux.HandleFunc("POST /signup", usersC.Create)

	mux.HandleFunc("GET /signin", usersC.SignIn)
	mux.HandleFunc("POST /signin", usersC.ProcessSignIn)
	mux.HandleFunc("GET /verify", usersC.VerifyOTP)
	mux.HandleFunc("POST /verify", usersC.ProcessOTP)

	mux.HandleFunc("GET /userhome", controllers.RequireSession(usersC.Home))
	mux.HandleFunc("GET /logout", usersC.Logout)

	srv := newServer(mux)
	_, stop := setupGracefulShutdown(srv)
	defer stop()
	startServer(srv)
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    ":9000",
		Handler: handler,
	}
}

func setupGracefulShutdown(srv *http.Server) (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		<-ctx.Done()
		log.Println("received shutdown signal")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()
	return ctx, stop
}

func startServer(srv *http.Server) {
	log.Println("starting the server...")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
