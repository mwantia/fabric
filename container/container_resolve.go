package container

import (
	"context"
	"fmt"
)

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
