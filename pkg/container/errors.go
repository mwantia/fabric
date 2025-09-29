package container

import (
	"errors"
	"sync"
)

// Errors is a thread-safe collection of errors that can be accumulated
// and then joined into a single error. This is used internally by the
// container for collecting multiple errors during operations like cleanup.
type Errors struct {
	mutex  sync.Mutex
	errors []error
}

// Add appends an error to the collection. Nil errors are ignored.
// This method is thread-safe and can be called concurrently.
func (e *Errors) Add(err error) {
	if err == nil {
		return
	}

	e.mutex.Lock()
	e.errors = append(e.errors, err)
	e.mutex.Unlock()
}

// Errors returns a joined error containing all accumulated errors,
// or nil if no errors have been added. This method is thread-safe.
func (e *Errors) Errors() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if len(e.errors) == 0 {
		return nil
	}

	return errors.Join(e.errors...)
}
