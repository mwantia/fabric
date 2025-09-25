package container

import "context"

type LifecycleService interface {
	Init(context.Context) error
	Cleanup(context.Context) error
}

func (sc *ServiceContainer) runLifecycle(ctx context.Context, singleton any) error {
	if lifecycle, ok := singleton.(LifecycleService); ok {
		if err := lifecycle.Init(ctx); err != nil {
			return err
		}
		sc.lifecycles = append(sc.lifecycles, lifecycle)
	}

	return nil
}
