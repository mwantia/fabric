package container

import (
	"context"
	"reflect"
)

// RegistrationService holds the configuration for a registered service,
// including its name, lifecycle, factory function, and interface mappings.
// This struct is used internally by the container to manage service registrations.
type RegistrationService struct {
	// Name is the optional name for this service registration, used for named resolution
	Name string

	// IsSingleton indicates whether this service should be created once and cached
	IsSingleton bool

	// Factory is the function used to create instances of this service
	Factory func(context.Context, *ServiceContainer) (any, error)

	// Interfaces maps interface types to their associated names for this service
	Interfaces map[reflect.Type][]string
}

// RegistrationOption is a function type used to configure service registrations.
// Options are applied during service registration to modify the registration's
// behavior, such as setting it as a singleton or providing a custom factory.
type RegistrationOption func(*RegistrationService) error

// RegistrationFactory is a function type for creating service instances.
// It receives a context and the service container, and returns the created
// instance or an error if creation fails.
type RegistrationFactory func(ctx context.Context, sc *ServiceContainer) (any, error)

// defaultRegistrationOptions creates a RegistrationService with default settings.
// The default configuration creates transient (non-singleton) services with
// no custom factory or interface mappings.
func defaultRegistrationOptions() *RegistrationService {
	return &RegistrationService{
		Name:        "",
		IsSingleton: false,
		Factory:     nil,
		Interfaces:  make(map[reflect.Type][]string),
	}
}

// AsSingleton configures a service registration to use singleton lifecycle.
// Singleton services are created once and the same instance is returned
// for all subsequent resolutions.
//
// Example:
//
//	Register[*DatabaseService](container, AsSingleton())
func AsSingleton() RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.IsSingleton = true
		return nil
	}
}

// AsFactory configures a service registration to use a custom factory function
// for creating instances. The factory function receives the current context
// and service container, allowing for complex initialization logic.
//
// Example:
//
//	Register[*DatabaseService](container,
//		AsFactory(func(ctx context.Context, sc *ServiceContainer) (any, error) {
//			config, _ := Resolve[*Config](ctx, sc)
//			return &DatabaseService{
//				ConnectionString: config.DBConnectionString,
//			}, nil
//		}))
func AsFactory(factory RegistrationFactory) RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.Factory = factory
		return nil
	}
}

// WithInstance configures a service registration to use a pre-created instance.
// The provided instance will always be returned when this service is resolved,
// effectively making it a singleton with the specific instance.
//
// This is useful for registering configuration objects, pre-configured clients,
// or other instances that should be shared across the application.
//
// Example:
//
//	config := &Config{DatabaseURL: "localhost:5432"}
//	Register[*Config](container, WithInstance(config))
func WithInstance(instance any) RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.Factory = func(ctx context.Context, sc *ServiceContainer) (any, error) {
			return instance, nil
		}
		return nil
	}
}

// With registers an interface that this concrete type implements, allowing
// the service to be resolved by its interface type without a name.
// This enables dependency abstraction and makes code more testable.
//
// Example:
//
//	type UserRepository interface {
//		GetUser(id string) (*User, error)
//	}
//
//	type DatabaseUserRepository struct {}
//
//	Register[*DatabaseUserRepository](container, With[UserRepository]())
//
//	// Later resolve by interface:
//	repo, err := Resolve[UserRepository](ctx, container)
func With[I any]() RegistrationOption {
	return func(rs *RegistrationService) error {
		ifaceType := reflect.TypeOf((*I)(nil)).Elem()
		if _, exists := rs.Interfaces[ifaceType]; !exists {
			rs.Interfaces[ifaceType] = make([]string, 0)
		}
		rs.Interfaces[ifaceType] = append(rs.Interfaces[ifaceType], "")
		return nil
	}
}

// WithName registers a named interface mapping that this concrete type implements.
// This allows multiple implementations of the same interface to be registered
// with different names, enabling selection of specific implementations at resolution time.
//
// Example:
//
//	type Database interface {
//		Connect() error
//	}
//
//	Register[*PostgresDB](container, WithName[Database]("postgres"))
//	Register[*MySQLDB](container, WithName[Database]("mysql"))
//
//	// Later resolve specific implementation:
//	pgDB, err := ResolveName[Database](ctx, container, "postgres")
//	myDB, err := ResolveName[Database](ctx, container, "mysql")
func WithName[I any](name string) RegistrationOption {
	return func(rs *RegistrationService) error {
		ifaceType := reflect.TypeOf((*I)(nil)).Elem()
		if _, exists := rs.Interfaces[ifaceType]; !exists {
			rs.Interfaces[ifaceType] = make([]string, 0)
		}
		rs.Interfaces[ifaceType] = append(rs.Interfaces[ifaceType], name)
		return nil
	}
}
