package container

import (
	"context"
	"testing"
)

// Mock services for testing
type LoggerService struct {
	Message string
}

type EncryptService struct {
	Algorithm string
}

type StorageService struct {
	Logger  *LoggerService  `fabric:"inject"`
	Encrypt *EncryptService `fabric:"inject"`
}

func TestFabricTagsInjection(t *testing.T) {
	sc := NewContainer()
	ctx := context.Background()

	// Register dependencies
	err := Register[*LoggerService](sc, WithInstance(&LoggerService{Message: "test logger"}))
	if err != nil {
		t.Fatalf("Failed to register LoggerService: %v", err)
	}

	err = Register[*EncryptService](sc, WithInstance(&EncryptService{Algorithm: "AES256"}))
	if err != nil {
		t.Fatalf("Failed to register EncryptService: %v", err)
	}

	// Register StorageService with fabric tags (enabled by default)
	err = Register[*StorageService](sc)
	if err != nil {
		t.Fatalf("Failed to register StorageService: %v", err)
	}

	// Resolve StorageService and verify injection
	storage, err := Resolve[*StorageService](ctx, sc)
	if err != nil {
		t.Fatalf("Failed to resolve StorageService: %v", err)
	}

	if storage.Logger == nil {
		t.Error("Logger was not injected")
	} else if storage.Logger.Message != "test logger" {
		t.Errorf("Expected logger message 'test logger', got '%s'", storage.Logger.Message)
	}

	if storage.Encrypt == nil {
		t.Error("EncryptService was not injected")
	} else if storage.Encrypt.Algorithm != "AES256" {
		t.Errorf("Expected encryption algorithm 'AES256', got '%s'", storage.Encrypt.Algorithm)
	}
}

func TestFabricTagsWithCustomOptions(t *testing.T) {
	sc := NewContainer()
	ctx := context.Background()

	// Register dependencies
	err := Register[*LoggerService](sc, WithInstance(&LoggerService{Message: "custom logger"}))
	if err != nil {
		t.Fatalf("Failed to register LoggerService: %v", err)
	}

	err = Register[*EncryptService](sc, WithInstance(&EncryptService{Algorithm: "RSA"}))
	if err != nil {
		t.Fatalf("Failed to register EncryptService: %v", err)
	}

	// Register StorageService as singleton - fabric tags are auto-detected
	err = Register[*StorageService](sc, AsSingleton())
	if err != nil {
		t.Fatalf("Failed to register StorageService: %v", err)
	}

	// Resolve twice to verify singleton behavior
	storage1, err := Resolve[*StorageService](ctx, sc)
	if err != nil {
		t.Fatalf("Failed to resolve StorageService first time: %v", err)
	}

	storage2, err := Resolve[*StorageService](ctx, sc)
	if err != nil {
		t.Fatalf("Failed to resolve StorageService second time: %v", err)
	}

	// Verify it's the same instance (singleton)
	if storage1 != storage2 {
		t.Error("Expected same instance for singleton, got different instances")
	}

	// Verify injection worked
	if storage1.Logger == nil || storage1.Logger.Message != "custom logger" {
		t.Error("Logger injection failed in singleton")
	}
	if storage1.Encrypt == nil || storage1.Encrypt.Algorithm != "RSA" {
		t.Error("Encrypt service injection failed in singleton")
	}
}

func TestWithoutFabricTags(t *testing.T) {
	sc := NewContainer()
	ctx := context.Background()

	// Define a service without fabric tags
	type SimpleService struct {
		Logger  *LoggerService
		Encrypt *EncryptService
	}

	// Register SimpleService with custom constructor
	err := Register[*SimpleService](sc,
		WithConstructor(func(ctx context.Context, sc *ServiceContainer) (*SimpleService, error) {
			return &SimpleService{}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Failed to register SimpleService: %v", err)
	}

	// Resolve SimpleService
	simple, err := Resolve[*SimpleService](ctx, sc)
	if err != nil {
		t.Fatalf("Failed to resolve SimpleService: %v", err)
	}

	// Verify no injection occurred
	if simple.Logger != nil {
		t.Error("Logger should not be injected when no fabric tags are present")
	}
	if simple.Encrypt != nil {
		t.Error("EncryptService should not be injected when no fabric tags are present")
	}
}