package container

import "reflect"

func typeKey[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
