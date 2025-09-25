# Fabric

A type-safe dependency injection container for Go applications.

⚠️ **Active Development**: This library is currently under active development. APIs may change before the first stable release.

## Installation

```bash
go get github.com/mwantia/fabric
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/mwantia/fabric/container"
)

type UserService struct {
    name string
}

func (u *UserService) GetUser() string {
    return u.name
}

func main() {
    // Create a new container
    sc := container.NewContainer()

    // Register a service with a factory
    container.Register[*UserService](sc,
        container.WithConstructor(func(ctx context.Context, sc *container.ServiceContainer) (*UserService, error) {
            return &UserService{name: "John Doe"}, nil
        }),
        container.AsSingleton(),
    )

    // Resolve the service
    ctx := context.Background()
    userService, err := container.Resolve[*UserService](ctx, sc)
    if err != nil {
        panic(err)
    }

    fmt.Println(userService.GetUser()) // Output: John Doe
}
```

## Features

- **Type Safety**: Compile-time type checking with Go generics
- **Singleton Support**: Register services as singletons or transient instances
- **Named Services**: Register multiple implementations with different names
- **Middleware Support**: Process services during resolution
- **Lifecycle Management**: Automatic cleanup of resources
- **Context Support**: Full context.Context integration

## Usage Examples

### Basic Registration

```go
// Register with instance
container.Register[string](sc, container.WithInstance("hello world"))

// Register with factory
container.Register[*Database](sc,
    container.WithFactory(func(ctx context.Context, sc *container.ServiceContainer) (any, error) {
        return &Database{connectionString: "localhost:5432"}, nil
    }),
)

// Register as singleton
container.Register[*Logger](sc,
    container.WithConstructor(newLogger),
    container.AsSingleton(),
)
```

### Named Services

```go
// Register multiple implementations
container.Register[Database](sc,
    container.WithName("postgres"),
    container.WithConstructor(newPostgresDB),
)

container.Register[Database](sc,
    container.WithName("mysql"),
    container.WithConstructor(newMySQLDB),
)

// Resolve by name
pgDB, err := container.ResolveName[Database](ctx, sc, "postgres")
```

### Resolution

```go
// Direct resolution
service, err := container.Resolve[*MyService](ctx, sc)

// Resolution with pointer assignment
var service *MyService
err := container.ResolveAs[*MyService](ctx, sc, &service)

// Named resolution with pointer assignment
err := container.ResolveNameAs[Database](ctx, sc, "postgres", &db)
```

### Cleanup

```go
defer func() {
    if err := sc.Cleanup(context.Background()); err != nil {
        log.Printf("Cleanup error: %v", err)
    }
}()
```

## Registration Options

| Option | Description |
|--------|-------------|
| `WithInstance(instance)` | Register a pre-created instance |
| `WithFactory(factory)` | Register with a factory function |
| `WithConstructor[T](ctor)` | Register with a type-safe constructor |
| `WithName(name)` | Register with a specific name |
| `AsSingleton()` | Register as singleton (default: transient) |

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Author

Created by [mwantia](https://github.com/mwantia)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
