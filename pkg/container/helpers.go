package container

import "reflect"

// typeKey returns the reflect.Type for type T, handling both concrete types
// and interfaces correctly. This is used internally by the container for
// type-based lookups and service registration.
func typeKey[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
