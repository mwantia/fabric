package container

import "context"

// LifecycleService is an interface for services that require initialization
// and cleanup during their lifecycle. Services implementing this interface
// will have their Init method called after creation and their Cleanup method
// called during container shutdown.
//
// The Init method is called immediately after a service is created and before
// it is returned to the caller. The Cleanup method is called in reverse order
// of registration during container cleanup.
//
// Example:
//
//	type DatabaseService struct {
//		Connection *sql.DB
//	}
//
//	func (ds *DatabaseService) Init(ctx context.Context) error {
//		var err error
//		ds.Connection, err = sql.Open("postgres", "connection_string")
//		return err
//	}
//
//	func (ds *DatabaseService) Cleanup(ctx context.Context) error {
//		if ds.Connection != nil {
//			return ds.Connection.Close()
//		}
//		return nil
//	}
type LifecycleService interface {
	// Init is called after service creation to perform initialization
	Init(context.Context) error

	// Cleanup is called during container shutdown to perform cleanup
	Cleanup(context.Context) error
}

// runLifecycle checks if the provided service implements LifecycleService and,
// if so, calls its Init method and registers it for cleanup during container shutdown.
// This method is called internally during service resolution.
func (sc *ServiceContainer) runLifecycle(ctx context.Context, singleton any) error {
	if lifecycle, ok := singleton.(LifecycleService); ok {
		if err := lifecycle.Init(ctx); err != nil {
			return err
		}
		sc.lifecycles = append(sc.lifecycles, lifecycle)
	}

	return nil
}
