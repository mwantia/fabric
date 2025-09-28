package container

import (
	"context"
	"fmt"
)

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
				var instance T
				return instance, nil
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
