package container

import (
	"context"
	"fmt"
	"reflect"
)

type RegistrationService struct {
	Name        string
	IsSingleton bool
	Factory     func(context.Context, *ServiceContainer) (any, error)
}

type RegistrationOption func(*RegistrationService) error
type RegistrationConstructor[T any] func(ctx context.Context, sc *ServiceContainer) (T, error)
type RegistrationFactory func(ctx context.Context, sc *ServiceContainer) (any, error)

func defaultRegistrationOptions() *RegistrationService {
	return &RegistrationService{
		Name:        "",
		IsSingleton: false,
		Factory:     nil,
	}
}

func WithName(name string) RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.Name = name
		return nil
	}
}

func AsSingleton() RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.IsSingleton = true
		return nil
	}
}

func WithFactory(factory RegistrationFactory) RegistrationOption {
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

func WithConstructor[T any](ctor RegistrationConstructor[T]) RegistrationOption {
	return func(rs *RegistrationService) error {
		rs.Factory = func(ctx context.Context, sc *ServiceContainer) (any, error) {
			return ctor(ctx, sc)
		}
		return nil
	}
}


func createFabricTagFactory[T any]() RegistrationFactory {
	return func(ctx context.Context, sc *ServiceContainer) (any, error) {
		var zero T
		t := reflect.TypeOf(zero)

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.Kind() != reflect.Struct {
			return zero, fmt.Errorf("fabric tags can only be used with struct types, got %T", zero)
		}

		val := reflect.New(t)
		structVal := val.Elem()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldVal := structVal.Field(i)

			if !fieldVal.CanSet() {
				continue
			}

			tag := field.Tag.Get("fabric")
			if tag == "inject" {
				fieldType := field.Type

				resolved, err := resolveByType(ctx, sc, fieldType)
				if err != nil {
					return zero, fmt.Errorf("failed to inject field %s: %w", field.Name, err)
				}

				fieldVal.Set(reflect.ValueOf(resolved))
			}
		}

		return val.Interface(), nil
	}
}

func hasFabricTags[T any]() bool {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("fabric"); tag == "inject" {
			return true
		}
	}

	return false
}

func resolveByType(ctx context.Context, sc *ServiceContainer, t reflect.Type) (any, error) {
	sc.mu.RLock()
	serviceMaps, exists := sc.services[t]
	sc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no registration found for type %s", t)
	}

	service, exists := serviceMaps[""]
	if !exists {
		return nil, fmt.Errorf("no default registration found for type %s", t)
	}

	if service.IsSingleton {
		sc.mu.RLock()
		singletonsMaps, exists := sc.singletons[t]
		sc.mu.RUnlock()

		if exists {
			if singleton, exists := singletonsMaps[""]; exists {
				return singleton, nil
			}
		}
	}

	return service.Factory(ctx, sc)
}
