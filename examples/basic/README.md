# Basic Fabric Usage Example

This example demonstrates the core features of the Fabric dependency injection container.

## Features Demonstrated

1. **Basic Service Registration** - Register services with interface mappings
2. **Custom Factory Functions** - Use factory functions for complex service creation
3. **Fabric Tags** - Automatic dependency injection using struct tags
4. **Lifecycle Management** - Services that implement Init/Cleanup
5. **Named Services** - Multiple implementations with different names
6. **Named Injection** - Fabric tags with named service resolution
7. **Singleton Services** - Services created once and cached
8. **Automatic Pointer Construction** - Proper handling of pointer types in registration

## Running the Example

```bash
go run examples/basic/main.go
```

## Expected Output

```
=== Example 1: Basic Registration ===
[APP] Application started
Connecting to PostgreSQL: postgres://localhost:5432/myapp
[APP] Query returned 2 results

=== Example 2: Fabric Tags ===
[APP] Creating user: John Doe
Connecting to PostgreSQL: postgres://localhost:5432/myapp
[APP] User John Doe created successfully

=== Example 3: Lifecycle Management ===
Initializing configuration...
Configuration initialized
Config - Database URL: postgres://localhost:5432/myapp
Config - API Key: secret-api-key

=== Example 4: Named Services ===
Connecting to PostgreSQL: postgres://prod:5432/app
[APP] Connected to production database

=== Example 5: Named Injection with Fabric Tags ===
[APP] Generating report...
Connecting to PostgreSQL: postgres://localhost:5432/myapp
Connecting to PostgreSQL: redis://cache:6379
[APP] Report generated with 2 records

=== Application finished ===
Cleaning up configuration...
Configuration cleaned up
```

## Code Structure

- **Interfaces**: `Logger` and `Database` define contracts
- **Concrete Types**: `ConsoleLogger` and `PostgresDB` implement interfaces
- **Service with Dependencies**: `UserService` uses fabric tags for injection
- **Lifecycle Service**: `ConfigService` implements automatic init/cleanup
- **Container Setup**: Registration and resolution examples

## Key Concepts

### Interface Mapping
```go
container.Register[*ConsoleLogger](sc, container.With[Logger]())
```
This allows the concrete type to be resolved by its interface.

### Factory Functions
```go
container.AsFactory(func(ctx context.Context, sc *container.ServiceContainer) (any, error) {
    return &PostgresDB{connectionString: "postgres://localhost:5432/myapp"}, nil
})
```
Custom creation logic for services.

### Fabric Tags
```go
type UserService struct {
    Logger   Logger   `fabric:"inject"`
    Database Database `fabric:"inject"`
}

type ReportService struct {
    Logger    Logger   `fabric:"inject"`
    PrimaryDB Database `fabric:"inject"`       // Default database
    CacheDB   Database `fabric:"inject:cache"` // Named injection
}
```
Dependencies are automatically resolved and injected. Named injection allows you to specify which specific implementation to inject when multiple services implement the same interface.

### Lifecycle Management
```go
func (c *ConfigService) Init(ctx context.Context) error {
    // Initialization logic
}

func (c *ConfigService) Cleanup(ctx context.Context) error {
    // Cleanup logic
}
```
Automatic lifecycle management for services.