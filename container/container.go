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

	tagProcessor *TagProcessorManager
}

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

func (sc *ServiceContainer) Cleanup(ctx context.Context) error {
	errs := &Errors{}
	for i := len(sc.lifecycles) - 1; i >= 0; i-- {
		if err := sc.lifecycles[i].Cleanup(ctx); err != nil {
			errs.Add(fmt.Errorf("error during container cleanup: %w", err))
		}
	}

	return errs.Errors()
}

// Add method to register new middleware services
func (sc *ServiceContainer) AddMiddleware(middlewares ...MiddlewareService) {
	sc.mu.Lock()
	sc.middlewares = append(sc.middlewares, middlewares...)
	sc.mu.Unlock()
}

// Add method to register new tag processors
func (sc *ServiceContainer) AddTagProcessor(processor ...TagProcessor) {
	sc.mu.Lock()
	sc.tagProcessor.registerProcessor(processor...)
	sc.mu.Unlock()
}

// Helper function for resolving by type (used by tag processors)
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
