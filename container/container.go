package container

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type ServiceContainer struct {
	mu sync.RWMutex

	services   map[reflect.Type]map[string]*RegistrationService
	singletons map[reflect.Type]map[string]any

	lifecycles  []LifecycleService
	middlewares []MiddlewareService
}

func NewContainer() *ServiceContainer {
	return &ServiceContainer{
		services:   make(map[reflect.Type]map[string]*RegistrationService),
		singletons: make(map[reflect.Type]map[string]any),
		lifecycles: make([]LifecycleService, 0),
	}
}

func (sc *ServiceContainer) Cleanup(ctx context.Context) error {
	errs := &Errors{}
	for i := len(sc.lifecycles) - 1; i >= 0; i-- {
		if err := sc.lifecycles[i].Cleanup(ctx); err != nil {
			errs.Add(fmt.Errorf("error during container cleanup: %w", err))
		}
	}

	return errs.Errors()
}

func (sc *ServiceContainer) AddMiddleware(middlewares ...MiddlewareService) {
	sc.mu.Lock()
	sc.middlewares = append(sc.middlewares, middlewares...)
	sc.mu.Unlock()
}

func Register[T any](sc *ServiceContainer, opts ...RegistrationOption) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	options := defaultRegistrationOptions()
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return err
		}
	}

	key := typeKey[T]()
	maps, exists := sc.services[key]
	if exists {
		maps[options.Name] = options
		return nil
	}

	maps = make(map[string]*RegistrationService)
	maps[options.Name] = options

	sc.services[key] = maps
	return nil
}

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

func Resolve[T any](ctx context.Context, sc *ServiceContainer) (T, error) {
	return ResolveName[T](ctx, sc, "")
}

func ResolveNameAs[T any](ctx context.Context, sc *ServiceContainer, name string, t *T) error {
	resolved, err := ResolveName[T](ctx, sc, name)
	if err != nil {
		return err
	}
	*t = resolved
	return nil
}

func ResolveAs[T any](ctx context.Context, sc *ServiceContainer, t *T) error {
	resolved, err := Resolve[T](ctx, sc)
	if err != nil {
		return err
	}
	*t = resolved
	return nil
}

func typeKey[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
