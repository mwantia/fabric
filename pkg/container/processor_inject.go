package container

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// InjectTagProcessor is the default tag processor that handles fabric:"inject" tags.
// It resolves dependencies by their type from the container and injects them into
// struct fields during service creation. Supports both unnamed and named injection:
//   - `fabric:"inject"` - resolves by type without name
//   - `fabric:"inject:name"` - resolves by type with the specified name
type InjectTagProcessor struct{}

// NewInjectTagProcessor creates a new InjectTagProcessor instance.
// This processor is registered by default when creating a new service container.
func NewInjectTagProcessor() *InjectTagProcessor {
	return &InjectTagProcessor{}
}

// GetPriority returns the processing priority for this processor.
// The default inject processor has priority 0 (lowest).
func (itp *InjectTagProcessor) GetPriority() int {
	return 0
}

// CanProcess returns true if this processor can handle the given tag value.
// The InjectTagProcessor handles:
//   - "inject" - for unnamed injection
//   - "inject:name" - for named injection
//
// All matching is case-insensitive.
func (itp *InjectTagProcessor) CanProcess(value string) bool {
	return strings.EqualFold(value, "inject") || strings.HasPrefix(strings.ToLower(value), "inject:")
}

// Process handles the injection of dependencies for fabric:"inject" tags.
// It supports both unnamed and named injection:
//   - "inject" - resolves by type without name
//   - "inject:name" - resolves by type with the specified name
//
// The method parses the tag value to extract the service name and then
// resolves the appropriate service from the container.
func (itp *InjectTagProcessor) Process(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error) {
	// Parse the tag value to extract the service name
	serviceName := ""
	if strings.Contains(value, ":") {
		parts := strings.SplitN(value, ":", 2)
		if len(parts) == 2 {
			serviceName = strings.TrimSpace(parts[1])
		}
	}

	// Try to resolve by name if specified
	if serviceName != "" {
		// Use reflection to call ResolveName with the service name
		return itp.resolveByName(ctx, sc, field.Type, serviceName)
	}

	// Fall back to unnamed resolution
	ok, resolved := sc.ResolveByType(ctx, field.Type)
	if !ok {
		return nil, fmt.Errorf("failed to inject type '%s' for field '%s'", field.Type, field.Name)
	}

	return resolved, nil
}

// resolveByName resolves a service by type and name using the container's internal resolution
func (itp *InjectTagProcessor) resolveByName(ctx context.Context, sc *ServiceContainer, fieldType reflect.Type, name string) (any, error) {
	sc.mu.RLock()
	serviceMaps, exists := sc.services[fieldType]
	sc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no service registered for type '%s'", fieldType)
	}

	service, exists := serviceMaps[name]
	if !exists {
		return nil, fmt.Errorf("no service registered for type '%s' with name '%s'", fieldType, name)
	}

	// Handle singleton resolution
	if service.IsSingleton {
		sc.mu.RLock()
		singletonsMaps, exists := sc.singletons[fieldType]
		sc.mu.RUnlock()

		if exists {
			if singleton, exists := singletonsMaps[name]; exists {
				return singleton, nil
			}
		}
	}

	// Create new instance using factory
	if service.Factory == nil {
		return nil, fmt.Errorf("no factory available for service '%s' with name '%s'", fieldType, name)
	}

	instance, err := service.Factory(ctx, sc)
	if err != nil {
		return nil, fmt.Errorf("failed to create service '%s' with name '%s': %w", fieldType, name, err)
	}

	// Apply middlewares
	for _, middleware := range sc.middlewares {
		instance, err = middleware.Process(ctx, fieldType, instance)
		if err != nil {
			return nil, fmt.Errorf("failed to process middleware for service '%s' with name '%s': %w", fieldType, name, err)
		}
	}

	// Run lifecycle
	if err := sc.runLifecycle(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to run lifecycle for service '%s' with name '%s': %w", fieldType, name, err)
	}

	// Cache singleton if needed
	if service.IsSingleton {
		sc.mu.Lock()
		singletonsMaps, exists := sc.singletons[fieldType]
		if exists {
			singletonsMaps[name] = instance
		} else {
			singletonsMaps = make(map[string]any)
			singletonsMaps[name] = instance
			sc.singletons[fieldType] = singletonsMaps
		}
		sc.mu.Unlock()
	}

	return instance, nil
}
