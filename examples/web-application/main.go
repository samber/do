package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	dochi "github.com/samber/do/http/chi/v2"
	"github.com/samber/do/v2"
)

// Server represents the HTTP server.
type Server struct {
	Config        *Configuration
	DB            *Database
	UserHandler   *UserHandler
	OrderHandler  *OrderHandler
	HealthHandler *HealthHandler
	Logger        *Logger
	injector      do.Injector
}

func (s *Server) Start() error {
	s.Logger.Log("Starting HTTP server...")

	// Connect to database
	s.Logger.Log("Connecting to database...")
	if err := s.DB.Connect(); err != nil {
		return err
	}

	// Setup router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Routes
	router.Get("/users/{id}", s.UserHandler.GetUser)
	router.Get("/orders/{id}", s.OrderHandler.GetOrder)
	router.Get("/health", s.HealthHandler.HealthCheck)

	// Debug endpoint for DI container
	dochi.Use(router, "/debug/do", s.injector)

	s.Logger.Log(fmt.Sprintf("Server listening on port %d", s.Config.Port))
	s.Logger.Log("Available endpoints:")
	s.Logger.Log("  GET /users/{id} - Get user by ID")
	s.Logger.Log("  GET /orders/{id} - Get order by ID")
	s.Logger.Log("  GET /health - Health check")
	s.Logger.Log("  GET /debug/do - DI container debug UI")

	return http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Port), router) //nolint:gosec
}

func (s *Server) Shutdown() error {
	s.Logger.Log("Shutting down server...")
	return nil
}

func main() {
	// Create injector with options
	injector := do.NewWithOpts(&do.InjectorOpts{
		Logf: func(format string, args ...any) {
			fmt.Printf("[DI] "+format+"\n", args...)
		},
		HealthCheckTimeout: 5 * time.Second,
	})

	fmt.Println("=== Web Application Example ===")

	// Register services
	do.Provide(injector, func(i do.Injector) (*Configuration, error) {
		return &Configuration{
			AppName: "WebApp",
			Port:    8080,
			Debug:   true,
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*Database, error) {
		config := do.MustInvoke[*Configuration](i)
		return &Database{
			Config: config,
			URL:    fmt.Sprintf("postgres://localhost:5432/%s", config.AppName),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*Cache, error) {
		return &Cache{
			Data: make(map[string]interface{}),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*Logger, error) {
		config := do.MustInvoke[*Configuration](i)
		level := "INFO"
		if config.Debug {
			level = "DEBUG"
		}
		return &Logger{
			Config: config,
			Level:  level,
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*UserService, error) {
		return &UserService{
			DB:     do.MustInvoke[*Database](i),
			Cache:  do.MustInvoke[*Cache](i),
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*OrderService, error) {
		return &OrderService{
			DB:     do.MustInvoke[*Database](i),
			Cache:  do.MustInvoke[*Cache](i),
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*UserHandler, error) {
		return &UserHandler{
			UserService: do.MustInvoke[*UserService](i),
			Logger:      do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*OrderHandler, error) {
		return &OrderHandler{
			OrderService: do.MustInvoke[*OrderService](i),
			Logger:       do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*HealthHandler, error) {
		return &HealthHandler{
			DB:     do.MustInvoke[*Database](i),
			Cache:  do.MustInvoke[*Cache](i),
			Logger: do.MustInvoke[*Logger](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*Server, error) {
		return &Server{
			Config:        do.MustInvoke[*Configuration](i),
			DB:            do.MustInvoke[*Database](i),
			UserHandler:   do.MustInvoke[*UserHandler](i),
			OrderHandler:  do.MustInvoke[*OrderHandler](i),
			HealthHandler: do.MustInvoke[*HealthHandler](i),
			Logger:        do.MustInvoke[*Logger](i),
			injector:      i,
		}, nil
	})

	fmt.Println("=== Service Registration Complete ===")
	fmt.Println("Available services:", injector.ListProvidedServices())

	// Start server
	server := do.MustInvoke[*Server](injector)

	fmt.Println("\n=== Starting Server ===")
	if err := server.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}

	fmt.Println("\n=== Shutting Down ===")
	_, err := injector.ShutdownOnSignals()
	if err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}
}
