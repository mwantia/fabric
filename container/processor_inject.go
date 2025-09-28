package container

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type InjectTagProcessor struct{}

func NewInjectTagProcessor() *InjectTagProcessor {
	return &InjectTagProcessor{}
}

func (itp *InjectTagProcessor) GetPriority() int {
	return 0
}

func (itp *InjectTagProcessor) CanProcess(value string) bool {
	return strings.EqualFold(value, "inject")
}

func (itp *InjectTagProcessor) Process(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error) {
	ok, resolved := sc.ResolveByType(ctx, field.Type)
	if !ok {
		return nil, fmt.Errorf("failed to inject type '%s'", field.Name)
	}

	return resolved, nil
}
