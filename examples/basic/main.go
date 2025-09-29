package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mwantia/fabric/pkg/container"
)

// Define interfaces for abstraction
type Logger interface {
	Log(message string)
}

type Database interface {
	Connect() error
	Query(sql string) ([]string, error)
	Close() error
}

// Concrete implementations
type ConsoleLogger struct {
	prefix string
}

func (c *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", c.prefix, message)
}

type PostgresDB struct {
	connectionString string
	connected        bool
}

func (p *PostgresDB) Connect() error {
	fmt.Printf("Connecting to PostgreSQL: %s\n", p.connectionString)
	p.connected = true
	return nil
}

func (p *PostgresDB) Query(sql string) ([]string, error) {
	if !p.connected {
		return nil, fmt.Errorf("database not connected")
	}
	return []string{"result1", "result2"}, nil
}

func (p *PostgresDB) Close() error {
	fmt.Println("Closing PostgreSQL connection")
	p.connected = false
	return nil
}

// Service with fabric tag dependency injection
type UserService struct {
	Logger   Logger   `fabric:"inject"`
	Database Database `fabric:"inject"`
}

// Service demonstrating named injection
type ReportService struct {
	Logger    Logger   `fabric:"inject"`
	PrimaryDB Database `fabric:"inject"`       // Default database
	CacheDB   Database `fabric:"inject:cache"` // Named injection
}

func (u *UserService) CreateUser(name string) error {
	u.Logger.Log(fmt.Sprintf("Creating user: %s", name))

	if err := u.Database.Connect(); err != nil {
		return err
	}

	_, err := u.Database.Query("INSERT INTO users (name) VALUES (?)")
	if err != nil {
		return err
	}

	u.Logger.Log(fmt.Sprintf("User %s created successfully", name))
	return nil
}

func (r *ReportService) GenerateReport() error {
	r.Logger.Log("Generating report...")

	// Use primary database for main data
	if err := r.PrimaryDB.Connect(); err != nil {
		return err
	}
	data, err := r.PrimaryDB.Query("SELECT * FROM users")
	if err != nil {
		return err
	}

	// Use cache database for caching results
	if err := r.CacheDB.Connect(); err != nil {
		return err
	}
	_, err = r.CacheDB.Query("CACHE report_data")
	if err != nil {
		return err
	}

	r.Logger.Log(fmt.Sprintf("Report generated with %d records", len(data)))
	return nil
}

// Service with lifecycle management
type ConfigService struct {
	DatabaseURL string
	APIKey      string
}

func (c *ConfigService) Init(ctx context.Context) error {
	fmt.Println("Initializing configuration...")
	c.DatabaseURL = "postgres://localhost:5432/myapp"
	c.APIKey = "secret-api-key"
	fmt.Println("Configuration initialized")
	return nil
}

func (c *ConfigService) Cleanup(ctx context.Context) error {
	fmt.Println("Cleaning up configuration...")
	c.APIKey = "" // Clear sensitive data
	fmt.Println("Configuration cleaned up")
	return nil
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

	// Example 1: Basic registration and resolution
	fmt.Println("=== Example 1: Basic Registration ===")

	// Register logger with interface mapping
	err := container.Register[*ConsoleLogger](sc,
		container.WithInstance(&ConsoleLogger{prefix: "APP"}),
		container.With[Logger]())
	if err != nil {
		log.Fatal(err)
	}

	// Register database with custom factory
	err = container.Register[*PostgresDB](sc,
		container.AsFactory(func(ctx context.Context, sc *container.ServiceContainer) (any, error) {
			return &PostgresDB{connectionString: "postgres://localhost:5432/myapp"}, nil
		}),
		container.With[Database](),
		container.AsSingleton())
	if err != nil {
		log.Fatal(err)
	}

	// Resolve and use services
	logger, err := container.Resolve[Logger](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	db, err := container.Resolve[Database](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log("Application started")
	db.Connect()
	results, _ := db.Query("SELECT * FROM users")
	logger.Log(fmt.Sprintf("Query returned %d results", len(results)))

	fmt.Println()

	// Example 2: Fabric tags with automatic dependency injection
	fmt.Println("=== Example 2: Fabric Tags ===")

	// Register UserService - dependencies will be automatically injected
	err = container.Register[*UserService](sc)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve UserService - fabric tags automatically inject dependencies
	userService, err := container.Resolve[*UserService](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	// Use the service - dependencies are already injected
	userService.CreateUser("John Doe")

	fmt.Println()

	// Example 3: Lifecycle management
	fmt.Println("=== Example 3: Lifecycle Management ===")

	// Register service that implements LifecycleService
	err = container.Register[*ConfigService](sc)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve service - Init will be called automatically
	config, err := container.Resolve[*ConfigService](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Config - Database URL: %s\n", config.DatabaseURL)
	fmt.Printf("Config - API Key: %s\n", config.APIKey)

	fmt.Println()

	// Example 4: Named services
	fmt.Println("=== Example 4: Named Services ===")

	// Register multiple database implementations
	err = container.Register[*PostgresDB](sc,
		container.WithInstance(&PostgresDB{connectionString: "postgres://prod:5432/app"}),
		container.WithName[Database]("postgres"))
	if err != nil {
		log.Fatal(err)
	}

	// We could register MySQL here too:
	// container.Register[*MySQLDB](sc, container.WithName[Database]("mysql"))

	// Resolve specific implementation by name
	prodDB, err := container.ResolveName[Database](ctx, sc, "postgres")
	if err != nil {
		log.Fatal(err)
	}

	prodDB.Connect()
	logger.Log("Connected to production database")

	fmt.Println()

	// Example 5: Named injection with fabric tags
	fmt.Println("=== Example 5: Named Injection with Fabric Tags ===")

	// Register a cache database with a specific name
	err = container.Register[*PostgresDB](sc,
		container.WithInstance(&PostgresDB{connectionString: "redis://cache:6379"}),
		container.WithName[Database]("cache"))
	if err != nil {
		log.Fatal(err)
	}

	// Register ReportService - it will use both unnamed and named injection
	err = container.Register[*ReportService](sc)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve ReportService - fabric tags will inject the correct services
	reportService, err := container.Resolve[*ReportService](ctx, sc)
	if err != nil {
		log.Fatal(err)
	}

	// Use the service - it will use both primary (unnamed) and cache (named) databases
	reportService.GenerateReport()

	fmt.Println("\n=== Application finished ===")
	// Cleanup will be called automatically due to defer
}
