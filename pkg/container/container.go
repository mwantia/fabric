// Package container provides a type-safe dependency injection framework for Go applications.
//
// The container package implements a comprehensive dependency injection system with support for:
//
//   - Type-safe service registration and resolution using Go generics
//   - Automatic dependency injection via struct tags (fabric:"inject")
//   - Flexible registration options including singletons, factories, and instances
//   - Interface mapping for dependency abstraction
//   - Named service resolution for multiple implementations
//   - Middleware processing during service resolution
//   - Lifecycle management for proper resource cleanup
//
// # Basic Usage
//
// Create a new service container and register services:
//
//	container := NewServiceContainer()
//
//	// Register a service with automatic construction
//	err := Register[*LoggerService](container)
//
//	// Register with interface mapping
//	err = Register[*DatabaseService](container,
//		With[Database](),
//		AsSingleton())
//
//	// Resolve services
//	logger, err := Resolve[*LoggerService](ctx, container)
//	db, err := Resolve[Database](ctx, container)
//
// # Fabric Tags
//
// Use struct tags for automatic dependency injection:
//
//	type UserService struct {
//		Logger   *LoggerService   `fabric:"inject"`
//		Database Database         `fabric:"inject"`
//		Cache    Database         `fabric:"inject:redis"`
//	}
//
//	// Register the service - dependencies will be automatically injected
//	err := Register[*UserService](container, With[UserService]())
//
// # Registration Options
//
// The package provides several registration options:
//
//   - AsSingleton(): Register as singleton (single instance shared across resolutions)
//   - AsFactory(factory): Use custom factory function for service creation
//   - WithInstance(instance): Register pre-created instance
//   - With[Interface](): Map service to interface type
//   - WithName[Interface](name): Map service to named interface
//
// # Named Services
//
// Register multiple implementations of the same interface:
//
//	Register[*PostgresDB](container, WithName[Database]("postgres"))
//	Register[*MySQLDB](container, WithName[Database]("mysql"))
//
//	// Resolve specific implementation
//	pgDB, err := ResolveName[Database](ctx, container, "postgres")
//
// # Error Handling
//
// All registration and resolution operations return detailed errors.
// The package provides custom error types for different failure scenarios.
package container

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// ServiceContainer is the main dependency injection container that manages service
// registrations, resolutions, and lifecycle. It provides thread-safe operations
// for concurrent access and supports singleton management, middleware processing,
// and automatic dependency injection via fabric tags.
//
// The container maintains separate maps for services and singletons, organized by
// type and optional names, enabling both unnamed and named service resolution.
type ServiceContainer struct {
	// mu provides thread-safe access to container state
	mu sync.RWMutex

	// services stores service registrations indexed by type and name
	services map[reflect.Type]map[string]*RegistrationService

	// singletons caches singleton instances to ensure single instance per registration
	singletons map[reflect.Type]map[string]any

	// lifecycles contains services that implement cleanup functionality
	lifecycles []LifecycleService

	// middlewares contains services that process resolved instances
	middlewares []MiddlewareService

	// tagProcessor manages fabric tag processing for automatic dependency injection
	tagProcessor *TagProcessorManager
}

// NewServiceContainer creates a new dependency injection container with default
// configuration. The container is initialized with:
//
//   - Empty service and singleton maps
//   - Default tag processor manager
//   - Automatic registration of the inject tag processor for fabric:"inject" tags
//
// The returned container is ready for service registration and resolution.
//
// Example:
//
//	container := NewServiceContainer()
//	defer container.Cleanup(context.Background())
func NewServiceContainer() *ServiceContainer {
	sc := &ServiceContainer{
		services:     make(map[reflect.Type]map[string]*RegistrationService),
		singletons:   make(map[reflect.Type]map[string]any),
		lifecycles:   make([]LifecycleService, 0),
		tagProcessor: NewTagProcessorManager(),
	}
	// Register the inject processor by default when creating a new container
	sc.AddTagProcessor(NewInjectTagProcessor())

	return sc
}

// Cleanup performs cleanup of all registered lifecycle services in reverse order
// of registration. This ensures that services are cleaned up in the opposite
// order they were registered, maintaining proper dependency cleanup order.
//
// The method collects all cleanup errors and returns them as a single error.
// If no errors occur during cleanup, it returns nil.
//
// It should typically be called when the application shuts down to ensure
// proper resource cleanup.
//
// Example:
//
//	defer func() {
//		if err := container.Cleanup(context.Background()); err != nil {
//			log.Printf("Cleanup errors: %v", err)
//		}
//	}()
func (sc *ServiceContainer) Cleanup(ctx context.Context) error {
	errs := &Errors{}
	for i := len(sc.lifecycles) - 1; i >= 0; i-- {
		if err := sc.lifecycles[i].Cleanup(ctx); err != nil {
			errs.Add(fmt.Errorf("error during container cleanup: %w", err))
		}
	}

	return errs.Errors()
}

// AddMiddleware registers one or more middleware services that will process
// resolved instances during service resolution. Middlewares are executed in
// the order they are registered.
//
// Middleware services can modify, wrap, or validate resolved instances before
// they are returned to the caller. Common use cases include logging, caching,
// validation, and proxying.
//
// Example:
//
//	container.AddMiddleware(
//		&LoggingMiddleware{},
//		&ValidationMiddleware{},
//	)
func (sc *ServiceContainer) AddMiddleware(middlewares ...MiddlewareService) {
	sc.mu.Lock()
	sc.middlewares = append(sc.middlewares, middlewares...)
	sc.mu.Unlock()
}

// AddTagProcessor registers one or more custom tag processors that handle
// fabric tag processing during service creation. Tag processors enable
// automatic dependency injection based on struct field tags.
//
// The default InjectTagProcessor handles the "inject" tag, but custom
// processors can be added to support additional tag types and behaviors.
//
// Example:
//
//	container.AddTagProcessor(&CustomTagProcessor{})
func (sc *ServiceContainer) AddTagProcessor(processor ...TagProcessor) {
	sc.mu.Lock()
	sc.tagProcessor.registerProcessor(processor...)
	sc.mu.Unlock()
}

// ResolveByType resolves a service by its reflect.Type. This method is primarily
// used internally by tag processors during fabric tag processing.
//
// It returns a boolean indicating whether the service was found and resolved,
// and the resolved instance if successful. If the service is not found or
// resolution fails, it returns false and nil.
//
// This method handles singleton caching automatically and applies the service's
// factory function if needed.
func (sc *ServiceContainer) ResolveByType(ctx context.Context, t reflect.Type) (bool, any) {
	sc.mu.RLock()
	serviceMaps, exists := sc.services[t]
	sc.mu.RUnlock()

	if !exists {
		return false, nil
	}

	service, exists := serviceMaps[""]
	if !exists {
		return false, nil
	}

	if service.IsSingleton {
		sc.mu.RLock()
		singletonsMaps, exists := sc.singletons[t]
		sc.mu.RUnlock()

		if exists {
			if singleton, exists := singletonsMaps[""]; exists {
				return true, singleton
			}
		}
	}

	if service.Factory == nil {
		return false, nil
	}

	factory, err := service.Factory(ctx, sc)
	if err != nil {
		return false, nil
	}

	return true, factory
}
