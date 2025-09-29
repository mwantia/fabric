# Web Application Example

This example demonstrates how to use Fabric for dependency injection in a web application with layered architecture.

## Architecture

The application follows a clean architecture pattern:

- **Web Layer**: HTTP handlers (`UserHandler`, `WebServer`)
- **Service Layer**: Business logic (`UserService`)
- **Repository Layer**: Data access (`UserRepository`, `DatabaseUserRepository`)
- **Infrastructure**: Cross-cutting concerns (`Logger`, `Database`)

## Features Demonstrated

1. **Layered Architecture** with dependency injection
2. **Interface-based Design** for testability and flexibility
3. **Fabric Tag Injection** throughout all layers
4. **Lifecycle Management** for server startup/shutdown
5. **Singleton Services** for shared resources
6. **Clean Separation of Concerns**

## Running the Example

```bash
go run examples/web-app/main.go
```

## Testing the API

Once running, test the endpoints:

### Get User by ID
```bash
curl "http://localhost:8080/users?id=1"
```

### Create New User
```bash
curl -X POST -d "name=Alice Johnson&email=alice@example.com" \
  "http://localhost:8080/users/create"
```

### Health Check
```bash
curl "http://localhost:8080/health"
```

## Code Structure

### Service Registration
```go
// Infrastructure layer
container.Register[*ConsoleLogger](sc, container.With[Logger](), container.AsSingleton())
container.Register[*InMemoryDatabase](sc, container.With[Database](), container.AsSingleton())

// Repository layer
container.Register[*DatabaseUserRepository](sc, container.With[UserRepository](), container.AsSingleton())

// Service layer
container.Register[*UserService](sc, container.AsSingleton())

// Web layer
container.Register[*UserHandler](sc, container.AsSingleton())
container.Register[*WebServer](sc, container.AsSingleton())
```

### Dependency Injection Chain

```
WebServer
├── Logger (injected)
└── UserHandler (injected)
    ├── Logger (injected)
    └── UserService (injected)
        ├── Logger (injected)
        └── UserRepository (injected)
            ├── Logger (injected)
            └── Database (injected)
```

### Fabric Tags Usage

Each service declares its dependencies using fabric tags:

```go
type UserService struct {
    Logger     Logger         `fabric:"inject"`
    Repository UserRepository `fabric:"inject"`
}
```

The container automatically resolves and injects these dependencies during service creation.

## Benefits of This Approach

1. **Testability**: Each layer can be tested in isolation with mocked dependencies
2. **Flexibility**: Easy to swap implementations (e.g., database backends)
3. **Maintainability**: Clear separation of concerns
4. **Type Safety**: Compile-time checking of dependencies
5. **Automatic Wiring**: No manual dependency management

## Production Considerations

In a production application, you would typically:

1. Use a real database instead of in-memory storage
2. Add proper error handling and logging
3. Implement authentication and authorization
4. Add input validation and sanitization
5. Use configuration management for settings
6. Add metrics and monitoring
7. Implement graceful shutdown handling