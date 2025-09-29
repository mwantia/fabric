# Fabric - Type-Safe Dependency Injection for Go

Fabric is a lightweight, type-safe dependency injection container for Go applications. It leverages Go generics to provide compile-time type safety while supporting advanced features like automatic dependency injection via struct tags, lifecycle management, and middleware processing.

## Features

- **Type-Safe**: Uses Go generics for compile-time type checking
- **Automatic Injection**: Fabric tags (`fabric:"inject"`) for automatic dependency injection
- **Flexible Registration**: Support for singletons, factories, instances, and interface mappings
- **Named Services**: Register multiple implementations with different names
- **Lifecycle Management**: Automatic initialization and cleanup of services
- **Middleware Support**: Process services during resolution
- **Thread-Safe**: Concurrent-safe operations for production use

## Installation

```bash
go get github.com/mwantia/fabric
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "github.com/mwantia/fabric/container"
)

// Define your services
type Logger interface {
    Log(message string)
}

type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(message string) {
    log.Println(message)
}

func main() {
    // Create container
    sc := container.NewServiceContainer()
    defer sc.Cleanup(context.Background())

    // Register services
    container.Register[*ConsoleLogger](sc, container.With[Logger]())

    // Resolve and use
    logger, err := container.Resolve[Logger](context.Background(), sc)
    if err != nil {
        log.Fatal(err)
    }

    logger.Log("Hello, Fabric!")
}
```

## Core Concepts

### Service Registration

Register services with various options:

```go
// Basic registration with automatic construction
container.Register[*LoggerService](sc)

// Singleton with interface mapping
container.Register[*DatabaseService](sc,
    container.AsSingleton(),
    container.With[Database]())

// Custom factory function
container.Register[*ConfigService](sc,
    container.AsFactory(func(ctx context.Context, sc *container.ServiceContainer) (any, error) {
        return &ConfigService{DatabaseURL: "localhost:5432"}, nil
    }))

// Pre-created instance
config := &Config{DatabaseURL: "localhost:5432"}
container.Register[*Config](sc, container.WithInstance(config))
```

### Service Resolution

Resolve services by type or name:

```go
// Resolve by interface
logger, err := container.Resolve[Logger](ctx, sc)

// Resolve by concrete type
service, err := container.Resolve[*LoggerService](ctx, sc)

// Resolve named service
pgDB, err := container.ResolveName[Database](ctx, sc, "postgres")

// Resolve into existing variable
var logger Logger
err := container.ResolveAs[Logger](ctx, sc, &logger)
```

### Named Services

Register multiple implementations of the same interface:

```go
// Register multiple database implementations
container.Register[*PostgresDB](sc, container.WithName[Database]("postgres"))
container.Register[*MySQLDB](sc, container.WithName[Database]("mysql"))

// Resolve specific implementation
pgDB, err := container.ResolveName[Database](ctx, sc, "postgres")
myDB, err := container.ResolveName[Database](ctx, sc, "mysql")
```

## Usage Examples

### Fabric Tags (Automatic Dependency Injection)

Use struct tags for automatic dependency injection:

```go
type UserService struct {
    Logger   Logger   `fabric:"inject"`           // Unnamed injection
    Database Database `fabric:"inject"`           // Unnamed injection
    Cache    Database `fabric:"inject:redis"`     // Named injection
    Queue    Database `fabric:"inject:rabbitmq"`  // Named injection
}

// Register the service - dependencies will be automatically resolved
container.Register[*UserService](sc)

// Dependencies are injected automatically during creation
userService, err := container.Resolve[*UserService](ctx, sc)
```

The fabric tags support two formats:
- `fabric:"inject"` - Resolves by type without a name
- `fabric:"inject:name"` - Resolves by type with the specified name

### Lifecycle Management

Services can implement `LifecycleService` for automatic initialization and cleanup:

```go
type DatabaseService struct {
    Connection *sql.DB
}

func (ds *DatabaseService) Init(ctx context.Context) error {
    var err error
    ds.Connection, err = sql.Open("postgres", "connection_string")
    return err
}

func (ds *DatabaseService) Cleanup(ctx context.Context) error {
    if ds.Connection != nil {
        return ds.Connection.Close()
    }
    return nil
}

// Register normally - Init/Cleanup called automatically
container.Register[*DatabaseService](sc, container.AsSingleton())

// Init called during first resolution
db, err := container.Resolve[*DatabaseService](ctx, sc)

// Cleanup called during container cleanup
defer sc.Cleanup(ctx)
```

### Middleware

Process services during resolution:

```go
type LoggingMiddleware struct{}

func (lm *LoggingMiddleware) Process(ctx context.Context, serviceType reflect.Type, instance any) (any, error) {
    log.Printf("Resolved service: %v", serviceType)
    return instance, nil
}

// Register middleware
sc.AddMiddleware(&LoggingMiddleware{})
```

## Registration Options

| Option | Description |
|--------|-------------|
| `WithInstance(instance)` | Register a pre-created instance |
| `AsFactory(factory)` | Register with a factory function |
| `With[I]()` | Map service to interface I |
| `WithName[I](name)` | Map service to named interface I |
| `AsSingleton()` | Register as singleton (default: transient) |

## Advanced Usage

### Custom Tag Processors

Create custom tag processors for specialized dependency injection:

```go
type ConfigTagProcessor struct{}

func (ctp *ConfigTagProcessor) GetPriority() int { return 100 }
func (ctp *ConfigTagProcessor) CanProcess(value string) bool { return value == "config" }

func (ctp *ConfigTagProcessor) Process(ctx context.Context, sc *container.ServiceContainer, field reflect.StructField, value string) (any, error) {
    // Custom logic for config injection
    return resolveConfig(field.Name)
}

// Register custom processor
sc.AddTagProcessor(&ConfigTagProcessor{})

// Use in structs
type Service struct {
    DatabaseURL string `fabric:"config"`
}
```

## Best Practices

1. **Use Interfaces**: Register services with interface mappings for better abstraction
2. **Lifecycle Management**: Implement `LifecycleService` for services that need initialization/cleanup
3. **Singleton Pattern**: Use `AsSingleton()` for expensive-to-create services
4. **Named Services**: Use named registration when you need multiple implementations
5. **Error Handling**: Always check errors from registration and resolution operations
6. **Container Cleanup**: Always call `Cleanup()` in your main function or service shutdown

## API Documentation

For complete API documentation, run:

```bash
go doc github.com/mwantia/fabric/container
```

Or visit the online documentation at [pkg.go.dev](https://pkg.go.dev/github.com/mwantia/fabric/container).

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and development process.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
