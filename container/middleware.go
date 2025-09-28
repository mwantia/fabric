package container

import (
	"context"
	"reflect"
)

type MiddlewareService interface {
	// Process handles the middleware processing and returns the changed value
	Process(context.Context, reflect.Type, any) (any, error)
}
