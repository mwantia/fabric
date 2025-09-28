package container

import (
	"context"
	"reflect"
)

type RegistrationService struct {
	Name        string
	IsSingleton bool
	Factory     func(context.Context, *ServiceContainer) (any, error)
	Interfaces  map[reflect.Type][]string // Interface type -> list of names (empty string for default)
}

type RegistrationOption func(*RegistrationService) error
type RegistrationFactory func(ctx context.Context, sc *ServiceContainer) (any, error)

func defaultRegistrationOptions() *RegistrationService {
	return &RegistrationService{
		Name:        "",
		IsSingleton: false,
		Factory:     nil,
		Interfaces:  make(map[reflect.Type][]string),
	}
}

func AsSingleton() RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.IsSingleton = true
		return nil
	}
}

func AsFactory(factory RegistrationFactory) RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.Factory = factory
		return nil
	}
}

func WithInstance(instance any) RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.Factory = func(ctx context.Context, sc *ServiceContainer) (any, error) {
			return instance, nil
		}
		return nil
	}
}

// With registers an interface that this concrete type implements (without name)
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

// WithName registers a named interface that this concrete type implements
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
