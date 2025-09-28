package container

import (
	"context"
	"fmt"
	"reflect"
	"sort"
)

type TagProcessor interface {
	// GetPriority returns the processing priority (higher = processed first)
	GetPriority() int

	// CanProcess returns true if this processor can handle and process the given tag value
	CanProcess(value string) bool

	// Process handles the tag processing and returns the value to inject
	Process(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error)
}

type TagProcessorManager struct {
	processors []TagProcessor
}

func NewTagProcessorManager() *TagProcessorManager {
	return &TagProcessorManager{
		processors: make([]TagProcessor, 0),
	}
}

func (tpm *TagProcessorManager) registerProcessor(processor ...TagProcessor) {
	tpm.processors = append(tpm.processors, processor...)

	sort.Slice(tpm.processors, func(i, j int) bool {
		return tpm.processors[i].GetPriority() > tpm.processors[j].GetPriority()
	})
}

func (tpm *TagProcessorManager) processField(ctx context.Context, sc *ServiceContainer, field reflect.StructField, value string) (any, error) {
	for _, processor := range tpm.processors {
		if processor.CanProcess(value) {
			return processor.Process(ctx, sc, field, value)
		}
	}

	return nil, fmt.Errorf("no processor found for tag value: %s", value)
}

func (tpm *TagProcessorManager) hasProcessorFor(value string) bool {
	for _, processor := range tpm.processors {
		if processor.CanProcess(value) {
			return true
		}
	}

	return false
}
