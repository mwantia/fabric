package container

import (
	"context"
	"fmt"
)

// ResolveName resolves a service of type T with the specified name from the container.
// This method supports named service resolution, allowing you to resolve specific
// implementations when multiple services implement the same interface.
//
// The resolution process:
//  1. Looks up the service registration by type and name
//  2. For singletons, checks if an instance already exists and returns it
//  3. Otherwise, calls the service's factory function to create a new instance
//  4. Applies any registered middlewares to the instance
//  5. Runs lifecycle initialization if the service implements LifecycleService
//  6. For singletons, caches the instance for future resolutions
//
// Example:
//
//	// Register named services
//	Register[*PostgresDB](container, WithName[Database]("postgres"))
//	Register[*MySQLDB](container, WithName[Database]("mysql"))
//
//	// Resolve specific implementation
//	pgDB, err := ResolveName[Database](ctx, container, "postgres")
//	myDB, err := ResolveName[Database](ctx, container, "mysql")
func ResolveName[T any](ctx context.Context, sc *ServiceContainer, name string) (T, error) {
	var zero T
	key := typeKey[T]()

	sc.mu.RLock()
	serviceMaps, exists := sc.services[key]
	sc.mu.RUnlock()

	if !exists {
		return zero, fmt.Errorf("registration for '%T' not found", zero)
	}

	service, exists := serviceMaps[name]
	if !exists {
		return zero, fmt.Errorf("registration for '%T' and name '%s' not found", zero, name)
	}

	if service.IsSingleton {
		sc.mu.RLock()
		singletonsMaps, exists := sc.singletons[key]
		sc.mu.RUnlock()

		if exists {
			if singleton, exists := singletonsMaps[name]; exists {
				if typed, ok := singleton.(T); ok {
					return typed, nil
				}
				return zero, fmt.Errorf("failed to cast stored singleton to %T", zero)
			}
		}

		// Double mutex lock checking
		if singletonsMaps, exists := sc.singletons[key]; exists {
			if singleton, exists := singletonsMaps[name]; exists {
				if typed, ok := singleton.(T); ok {
					return typed, nil
				}
				return zero, fmt.Errorf("failed to cast stored singleton to %T", zero)
			}
		}
	}

	singleton, err := service.Factory(ctx, sc)
	if err != nil {
		return zero, err
	}

	_, ok := singleton.(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast singleton to '%T': %s", zero, err)
	}

	for _, middleware := range sc.middlewares {
		singleton, err = middleware.Process(ctx, key, singleton)
		if err != nil {
			return zero, fmt.Errorf("failed to process middleware during singleton creator of '%T': %s", zero, err)
		}
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if err := sc.runLifecycle(ctx, singleton); err != nil {
		return zero, err
	}

	if service.IsSingleton {
		singletonsMaps, exists := sc.singletons[key]
		if exists {
			singletonsMaps[name] = singleton
		} else {
			singletonsMaps = make(map[string]any)
			singletonsMaps[name] = singleton

			sc.singletons[key] = singletonsMaps
		}
	}

	return singleton.(T), nil
}

// Resolve resolves a service of type T from the container using an empty name.
// This is a convenience method for resolving services that were registered
// without a specific name (the default case).
//
// This method is equivalent to calling ResolveName[T](ctx, sc, "").
//
// Example:
//
//	// Register service without name
//	Register[*LoggerService](container, With[Logger]())
//
//	// Resolve without name
//	logger, err := Resolve[Logger](ctx, container)
func Resolve[T any](ctx context.Context, sc *ServiceContainer) (T, error) {
	return ResolveName[T](ctx, sc, "")
}

// ResolveNameAs resolves a named service of type T and assigns it to the provided pointer.
// This method is useful when you want to avoid declaring a new variable and prefer
// to assign directly to an existing variable reference.
//
// Example:
//
//	var pgDB Database
//	err := ResolveNameAs[Database](ctx, container, "postgres", &pgDB)
//	if err != nil {
//		return err
//	}
func ResolveNameAs[T any](ctx context.Context, sc *ServiceContainer, name string, t *T) error {
	resolved, err := ResolveName[T](ctx, sc, name)
	if err != nil {
		return err
	}
	*t = resolved
	return nil
}

// ResolveAs resolves a service of type T and assigns it to the provided pointer.
// This is a convenience method equivalent to ResolveNameAs with an empty name.
//
// Example:
//
//	var logger Logger
//	err := ResolveAs[Logger](ctx, container, &logger)
//	if err != nil {
//		return err
//	}
func ResolveAs[T any](ctx context.Context, sc *ServiceContainer, t *T) error {
	resolved, err := Resolve[T](ctx, sc)
	if err != nil {
		return err
	}
	*t = resolved
	return nil
}
