package container

import "context"

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
