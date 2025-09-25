package container

import (
	"context"
	"reflect"
)

type MiddlewareService interface {
	Process(context.Context, reflect.Type, any) (any, error)
}
