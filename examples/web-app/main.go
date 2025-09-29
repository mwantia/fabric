package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mwantia/fabric/pkg/container"
)

// Interfaces
type Logger interface {
	Info(message string)
	Error(message string)
}

type Database interface {
	GetUser(id string) (*User, error)
	CreateUser(user *User) error
}

type UserRepository interface {
	FindByID(id string) (*User, error)
	Save(user *User) error
}

// Models
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Implementations
type ConsoleLogger struct{}

func (c *ConsoleLogger) Info(message string) {
	log.Printf("[INFO] %s", message)
}

func (c *ConsoleLogger) Error(message string) {
	log.Printf("[ERROR] %s", message)
}

type InMemoryDatabase struct {
	users map[string]*User
}

func (db *InMemoryDatabase) Init(ctx context.Context) error {
	db.users = make(map[string]*User)
	// Seed some data
	db.users["1"] = &User{ID: "1", Name: "John Doe", Email: "john@example.com"}
	db.users["2"] = &User{ID: "2", Name: "Jane Smith", Email: "jane@example.com"}
	return nil
}

func (db *InMemoryDatabase) Cleanup(ctx context.Context) error {
	db.users = nil
	return nil
}

func (db *InMemoryDatabase) GetUser(id string) (*User, error) {
	if user, exists := db.users[id]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found: %s", id)
}

func (db *InMemoryDatabase) CreateUser(user *User) error {
	db.users[user.ID] = user
	return nil
}

type DatabaseUserRepository struct {
	Logger   Logger   `fabric:"inject"`
	Database Database `fabric:"inject"`
}

func (r *DatabaseUserRepository) FindByID(id string) (*User, error) {
	r.Logger.Info(fmt.Sprintf("Finding user by ID: %s", id))
	return r.Database.GetUser(id)
}

func (r *DatabaseUserRepository) Save(user *User) error {
	r.Logger.Info(fmt.Sprintf("Saving user: %s", user.Name))
	return r.Database.CreateUser(user)
}

type UserService struct {
	Logger     Logger         `fabric:"inject"`
	Repository UserRepository `fabric:"inject"`
}

func (s *UserService) GetUser(id string) (*User, error) {
	s.Logger.Info(fmt.Sprintf("Getting user: %s", id))
	return s.Repository.FindByID(id)
}

func (s *UserService) CreateUser(name, email string) (*User, error) {
	s.Logger.Info(fmt.Sprintf("Creating user: %s", name))

	user := &User{
		ID:    fmt.Sprintf("user_%d", len(name)), // Simple ID generation
		Name:  name,
		Email: email,
	}

	if err := s.Repository.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

type UserHandler struct {
	Logger      Logger       `fabric:"inject"`
	UserService *UserService `fabric:"inject"`
}

func (h *UserHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.GetUser(userID)
	if err != nil {
		h.Logger.Error(fmt.Sprintf("Failed to get user: %v", err))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id":"%s","name":"%s","email":"%s"}`, user.ID, user.Name, user.Email)
}

func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	if name == "" || email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.CreateUser(name, email)
	if err != nil {
		h.Logger.Error(fmt.Sprintf("Failed to create user: %v", err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id":"%s","name":"%s","email":"%s"}`, user.ID, user.Name, user.Email)
}

type WebServer struct {
	Logger  Logger       `fabric:"inject"`
	Handler *UserHandler `fabric:"inject"`
	server  *http.Server
}

func (ws *WebServer) Init(ctx context.Context) error {
	ws.Logger.Info("Initializing web server...")

	mux := http.NewServeMux()
	mux.HandleFunc("/users", ws.Handler.handleGetUser)
	mux.HandleFunc("/users/create", ws.Handler.handleCreateUser)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	ws.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ws.Logger.Info("Web server initialized on :8080")
	return nil
}

func (ws *WebServer) Start() error {
	ws.Logger.Info("Starting web server...")
	return ws.server.ListenAndServe()
}

func (ws *WebServer) Cleanup(ctx context.Context) error {
	ws.Logger.Info("Shutting down web server...")
	return ws.server.Shutdown(ctx)
}

func main() {
	// Create service container
	sc := container.NewServiceContainer()
	defer func() {
		if err := sc.Cleanup(context.Background()); err != nil {
			log.Printf("Cleanup error: %v", err)
		}
	}()

	ctx := context.Background()

	// Register services with dependency injection

	// Infrastructure
	err := container.Register[*ConsoleLogger](sc,
		container.With[Logger](),
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	err = container.Register[*InMemoryDatabase](sc,
		container.With[Database](),
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	// Repository layer
	err = container.Register[*DatabaseUserRepository](sc,
		container.With[UserRepository](),
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	// Service layer
	err = container.Register[*UserService](sc,
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	// Web layer
	err = container.Register[*UserHandler](sc,
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	err = container.Register[*WebServer](sc,
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	// Start the application
	logger, err := container.Resolve[Logger](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	webServer, err := container.Resolve[*WebServer](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Application starting...")
	logger.Info("Available endpoints:")
	logger.Info("  GET /users?id=1 - Get user by ID")
	logger.Info("  POST /users/create - Create new user (name, email)")
	logger.Info("  GET /health - Health check")

	// Start server (this blocks)
	if err := webServer.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
