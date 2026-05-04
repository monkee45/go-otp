package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"photos.com/controllers"
	"photos.com/loggerctx"
	"photos.com/models"
	"photos.com/utils"
)

func main() {
	cfg := LoadConfig(false)
	dbCfg := cfg.Database
	// setup the logger
	logger := utils.NewLogger("logfile")
	ctx := loggerctx.WithLogger(context.Background(), logger)
	slog.SetDefault(logger)

	db, err := models.Open(dbCfg.ConnectionInfo())
	if err != nil {
		logger.Error("Failed to open Database", "Error:", err)
		panic(err)
	}
	defer db.Close()
	// create NewUserService
	userService := models.UserService{
		DB:     db,
		Logger: logger,
	}
	logger.Info("Setting up the Controllers")
	usersC := controllers.NewUserController(userService)
	staticC := controllers.NewStaticController()

	logger.Info("Setting up the Mux")
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
	logger.Info("calling setupGracefulShutdown")
	_, stop := setupGracefulShutdown(ctx, srv)
	defer stop()
	startServer(ctx, srv)
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    ":9000",
		Handler: handler,
	}
}

func setupGracefulShutdown(logctx context.Context, srv *http.Server) (context.Context, context.CancelFunc) {
	logger := loggerctx.LoggerFromContext(logctx)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		<-ctx.Done()
		logger.Info("received shutdown signal...")
		// log.Println("received shutdown signal")
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("Shutdown error ", "Error: ", err)
		}
		logger.Info("HTTP server stopped")
	}()
	return ctx, stop
}

func startServer(ctx context.Context, srv *http.Server) {
	logger := loggerctx.LoggerFromContext(ctx)
	logger.Info("starting the server...")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
