package container

import (
	"context"
	"fmt"
	"reflect"
	"sort"
)

// TagProcessor is an interface for handling fabric tag processing during service creation.
// Tag processors enable automatic dependency injection by processing struct field tags
// and resolving the appropriate dependencies from the container.
//
// Multiple processors can be registered, and they are executed in priority order
// (higher priority first). The first processor that can handle a tag value will
// be used to process that field.
//
// Example:
//
//	type CustomTagProcessor struct{}
//
//	func (ctp *CustomTagProcessor) GetPriority() int { return 100 }
//
//	func (ctp *CustomTagProcessor) CanProcess(value string) bool {
//		return value == "custom"
//	}
//
//	func (ctp *CustomTagProcessor) Process(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error) {
//		// Custom processing logic here
//		return customResolve(ctx, sc, field)
//	}
type TagProcessor interface {
	// GetPriority returns the processing priority (higher = processed first)
	GetPriority() int

	// CanProcess returns true if this processor can handle and process the given tag value
	CanProcess(value string) bool

	// Process handles the tag processing and returns the value to inject into the field.
	// It receives the context, container, field information, and tag value.
	Process(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error)
}

// TagProcessorManager manages a collection of tag processors and handles
// the processing of fabric tags during service creation. It maintains
// processors in priority order and routes tag processing to the appropriate
// processor based on tag values.
type TagProcessorManager struct {
	processors []TagProcessor
}

// NewTagProcessorManager creates a new TagProcessorManager with an empty
// processor collection.
func NewTagProcessorManager() *TagProcessorManager {
	return &TagProcessorManager{
		processors: make([]TagProcessor, 0),
	}
}

// registerProcessor adds one or more tag processors to the manager and
// sorts them by priority (higher priority first).
func (tpm *TagProcessorManager) registerProcessor(processor ...TagProcessor) {
	tpm.processors = append(tpm.processors, processor...)

	sort.Slice(tpm.processors, func(i, j int) bool {
		return tpm.processors[i].GetPriority() > tpm.processors[j].GetPriority()
	})
}

// processField processes a struct field with the given fabric tag value.
// It iterates through registered processors in priority order and uses
// the first processor that can handle the tag value.
func (tpm *TagProcessorManager) processField(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error) {
	for _, processor := range tpm.processors {
		if processor.CanProcess(value) {
			return processor.Process(ctx, sc, field, value)
		}
	}

	return nil, fmt.Errorf("no processor found for tag value: %s", value)
}

// hasProcessorFor checks if there is a registered processor that can handle
// the given tag value. This is used during service registration validation.
func (tpm *TagProcessorManager) hasProcessorFor(value string) bool {
	for _, processor := range tpm.processors {
		if processor.CanProcess(value) {
			return true
		}
	}

	return false
}
