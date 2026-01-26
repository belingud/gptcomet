package llm

import (
	"fmt"
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	err := registry.Register("test", func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	})
	if err != nil {
		t.Errorf("Register() error = %v, want nil", err)
	}

	// Test empty name
	err = registry.Register("", func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	})
	if err == nil {
		t.Error("Register() with empty name should return error")
	}

	// Test nil constructor
	err = registry.Register("test2", nil)
	if err == nil {
		t.Error("Register() with nil constructor should return error")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	// Register a provider
	testConstructor := func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	}
	registry.Register("test", testConstructor)

	// Test Get existing provider
	constructor, ok := registry.Get("test")
	if !ok {
		t.Error("Get() should find registered provider")
	}
	if constructor == nil {
		t.Error("Get() should return non-nil constructor")
	}

	// Test Get non-existing provider
	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("Get() should not find non-existent provider")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Register multiple providers
	providers := []string{"provider1", "provider2", "provider3"}
	for _, name := range providers {
		registry.Register(name, func(config *types.ClientConfig) LLM {
			return &DefaultLLM{}
		})
	}

	// Get list
	list := registry.List()
	if len(list) != len(providers) {
		t.Errorf("List() returned %d providers, want %d", len(list), len(providers))
	}

	// Verify list is sorted
	for i := 1; i < len(list); i++ {
		if list[i-1] > list[i] {
			t.Error("List() should return sorted provider names")
		}
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()

	// Register a provider
	registry.Register("test", func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	})

	// Test Has with existing provider
	if !registry.Has("test") {
		t.Error("Has() should return true for registered provider")
	}

	// Test Has with non-existing provider
	if registry.Has("nonexistent") {
		t.Error("Has() should return false for non-existent provider")
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	// Initially empty
	if registry.Count() != 0 {
		t.Errorf("Count() = %d, want 0", registry.Count())
	}

	// Register providers
	registry.Register("provider1", func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	})
	registry.Register("provider2", func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	})

	if registry.Count() != 2 {
		t.Errorf("Count() = %d, want 2", registry.Count())
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Note: This test uses the global registry which is shared across tests
	// In a real scenario, we might want to reset it or use a separate instance

	// Test RegisterProvider
	testProviderName := "test-global-provider"
	err := RegisterProvider(testProviderName, func(config *types.ClientConfig) LLM {
		return &DefaultLLM{}
	})
	if err != nil {
		t.Errorf("RegisterProvider() error = %v, want nil", err)
	}

	// Test HasProvider
	if !HasProvider(testProviderName) {
		t.Error("HasProvider() should return true for registered provider")
	}

	// Test GetProviderConstructor
	constructor, ok := GetProviderConstructor(testProviderName)
	if !ok {
		t.Error("GetProviderConstructor() should find registered provider")
	}
	if constructor == nil {
		t.Error("GetProviderConstructor() should return non-nil constructor")
	}

	// Test GetProviders includes our test provider
	providersList := GetProviders()
	found := false
	for _, name := range providersList {
		if name == testProviderName {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetProviders() should include registered test provider")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Test concurrent registration and access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			providerName := fmt.Sprintf("provider%d", id)
			registry.Register(providerName, func(config *types.ClientConfig) LLM {
				return &DefaultLLM{}
			})
			done <- true
		}(i)
	}

	// Wait for all registrations
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all providers are registered
	if registry.Count() != 10 {
		t.Errorf("Count() = %d, want 10 after concurrent registration", registry.Count())
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			providerName := fmt.Sprintf("provider%d", id)
			if !registry.Has(providerName) {
				t.Errorf("Has(%s) should return true", providerName)
			}
			done <- true
		}(i)
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}
}
