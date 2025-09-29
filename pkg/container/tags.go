package container

import (
	"context"
	"fmt"
	"reflect"
)

func hasFabricTags[T any]() bool {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return false
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("fabric"); tag == "inject" {
			return true
		}
	}

	return false
}

func validateFabricTags[T any](sc *ServiceContainer) (bool, error) {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false, nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("fabric"); tag != "" {
			if !sc.tagProcessor.hasProcessorFor(tag) {
				return false, fmt.Errorf("no processor registered for fabric tag '%s' on field '%s'", tag, field.Name)
			}
		}
	}

	return true, nil
}

func createFabricTagFactory[T any]() RegistrationFactory {
	return func(ctx context.Context, sc *ServiceContainer) (any, error) {
		var zero T
		t := reflect.TypeOf(zero)
		if t == nil {
			return nil, fmt.Errorf("fabric tags not defined")
		}

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.Kind() != reflect.Struct {
			return zero, fmt.Errorf("fabric tags can only be used with struct types, got %T", zero)
		}

		val := reflect.New(t)
		structVal := val.Elem()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldVal := structVal.Field(i)

			if !fieldVal.CanSet() {
				continue
			}

			tag := field.Tag.Get("fabric")
			if tag != "" {
				resolved, err := sc.tagProcessor.processField(ctx, sc, field, tag)
				if err != nil {
					return zero, fmt.Errorf("failed to process fabric tag for field '%s': %w", field.Name, err)
				}

				if resolved != nil {
					fieldVal.Set(reflect.ValueOf(resolved))
				}
			}
		}

		v := val.Interface()
		return v, nil
	}
}
