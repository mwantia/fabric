package container

import (
	"context"
	"fmt"
	"testing"
)

type LoggerEngine interface {
	Debug(msg string, args ...any)
}

type LoggerService struct{}

func (ls *LoggerService) Debug(msg string, args ...any) {
	fmt.Printf(msg, args...)
}

type EncryptEngine interface {
	Encrypt(ctx context.Context, plain []byte) ([]byte, error)
}

type EncryptService struct{}

func (es *EncryptService) Encrypt(ctx context.Context, plain []byte) ([]byte, error) {
	return plain, nil
}

type Agent struct {
	Logger  LoggerEngine  `fabric:"inject"`
	Encrypt EncryptEngine `fabric:"inject"`
}

func TestFabricTagsInjection(t *testing.T) {
	sc := NewServiceContainer()
	ctx := t.Context()

	errs := &Errors{}

	errs.Add(Register[*Agent](sc,
		AsSingleton()))

	errs.Add(Register[*LoggerService](sc,
		With[LoggerEngine](),
		AsSingleton()))

	errs.Add(Register[*EncryptService](sc,
		With[EncryptEngine](),
		AsSingleton()))

	if err := errs.Errors(); err != nil {
		t.Fatalf("Failed to complete service registration: %v", err)
	}

	agent, err := Resolve[*Agent](ctx, sc)
	if err != nil {
		t.Fatalf("Failed to resolve agent: %v", err)
	}

	if agent.Logger == nil {
		t.Error("Logger was not successfully injected")
	}
	if agent.Encrypt == nil {
		t.Error("Encrypt was not successfully injected")
	}
}
