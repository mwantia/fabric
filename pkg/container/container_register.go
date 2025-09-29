package container

import (
	"context"
	"fmt"
	"reflect"
)

// Register registers a service of type T with the container using the provided options.
// This is the primary method for registering services and supports:
//
//   - Automatic service construction using Go's zero value initialization
//   - Custom factory functions via AsFactory()
//   - Pre-created instances via WithInstance()
//   - Singleton lifecycle via AsSingleton()
//   - Interface mappings via With[I]() and WithName[I](name)
//   - Automatic dependency injection for services with fabric tags
//
// The registration process automatically detects and configures fabric tag processing
// if the service struct contains fabric:"inject" tags, enabling automatic dependency
// injection during service creation.
//
// Examples:
//
//	// Basic registration with automatic construction
//	err := Register[*LoggerService](container)
//
//	// Singleton with interface mapping
//	err = Register[*DatabaseService](container,
//		AsSingleton(),
//		With[Database]())
//
//	// Named interface with custom factory
//	err = Register[*PostgresDB](container,
//		WithName[Database]("postgres"),
//		AsFactory(func(ctx context.Context, sc *ServiceContainer) (any, error) {
//			config, _ := Resolve[*Config](ctx, sc)
//			return NewPostgresDB(config.PostgresURL), nil
//		}))
//
//	// Service with fabric tag dependencies
//	type UserService struct {
//		Logger *LoggerService `fabric:"inject"`
//		DB     Database       `fabric:"inject"`
//	}
//	err = Register[*UserService](container, With[UserService]())
func Register[T any](sc *ServiceContainer, opts ...RegistrationOption) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	options := defaultRegistrationOptions()
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return fmt.Errorf("failed to complete registration: %w", err)
		}
	}

	// If no factory is provided, create one automatically
	if options.Factory == nil {
		if hasFabricTags[T]() {
			ok, err := validateFabricTags[T](sc)
			if err != nil {
				return fmt.Errorf("failed to validate fabric tags: %w", err)
			}

			if !ok {
				return fmt.Errorf("no valid factory found during registration")
			}

			options.Factory = createFabricTagFactory[T]()
		} else {
			// Default factory - create instance using Go's zero value constructor
			options.Factory = func(ctx context.Context, sc *ServiceContainer) (any, error) {
				var zero T
				t := reflect.TypeOf(zero)

				// Handle pointer types by creating a new instance
				if t != nil && t.Kind() == reflect.Ptr {
					// Create a new instance of the pointed-to type
					elemType := t.Elem()
					if elemType.Kind() == reflect.Struct {
						val := reflect.New(elemType)
						return val.Interface(), nil
					}
				}

				// For non-pointer types, return the zero value
				return zero, nil
			}
		}
	}

	// Register the concrete type
	concreteKey := typeKey[T]()
	maps, exists := sc.services[concreteKey]
	if !exists {
		maps = make(map[string]*RegistrationService)
		sc.services[concreteKey] = maps
	}
	maps[options.Name] = options

	// Register all interface mappings
	for ifaceType, names := range options.Interfaces {
		ifaceMaps, exists := sc.services[ifaceType]
		if !exists {
			ifaceMaps = make(map[string]*RegistrationService)
			sc.services[ifaceType] = ifaceMaps
		}

		for _, name := range names {
			ifaceMaps[name] = options
		}
	}

	return nil
}
