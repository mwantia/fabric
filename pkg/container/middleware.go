package container

import (
	"context"
	"reflect"
)

// MiddlewareService is an interface for services that process resolved instances
// during service resolution. Middlewares can modify, wrap, validate, or perform
// additional operations on services before they are returned to the caller.
//
// Middlewares are executed in the order they are registered and can be used for
// cross-cutting concerns such as logging, caching, validation, or proxying.
//
// Example:
//
//	type LoggingMiddleware struct{}
//
//	func (lm *LoggingMiddleware) Process(ctx context.Context, serviceType reflect.Type, instance any) (any, error) {
//		log.Printf("Resolved service of type: %v", serviceType)
//		return instance, nil
//	}
//
//	type ValidationMiddleware struct{}
//
//	func (vm *ValidationMiddleware) Process(ctx context.Context, serviceType reflect.Type, instance any) (any, error) {
//		if validator, ok := instance.(interface{ Validate() error }); ok {
//			if err := validator.Validate(); err != nil {
//				return nil, fmt.Errorf("service validation failed: %w", err)
//			}
//		}
//		return instance, nil
//	}
type MiddlewareService interface {
	// Process handles the middleware processing and returns the potentially modified instance.
	// The context provides request context, serviceType is the reflect.Type of the service,
	// and instance is the resolved service instance. Returns the processed instance or an error.
	Process(context.Context, reflect.Type, any) (any, error)
}
