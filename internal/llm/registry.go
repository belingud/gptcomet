// Package llm provides registry functionality for LLM provider management.
package llm

import (
	"fmt"
	"sort"
	"sync"

	"github.com/belingud/gptcomet/pkg/types"
)

// ProviderConstructor is a function that creates a new LLM instance
type ProviderConstructor func(config *types.ClientConfig) LLM

// Registry manages LLM provider registration and lookup
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProviderConstructor
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ProviderConstructor),
	}
}

// Register registers a new LLM provider constructor
func (r *Registry) Register(name string, constructor ProviderConstructor) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if constructor == nil {
		return fmt.Errorf("constructor cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[name] = constructor
	return nil
}

// Get retrieves a provider constructor by name
func (r *Registry) Get(name string) (ProviderConstructor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	constructor, ok := r.providers[name]
	return constructor, ok
}

// List returns a sorted list of all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Has checks if a provider is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.providers[name]
	return ok
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}

// Global registry instance
var defaultRegistry = NewRegistry()

// RegisterProvider registers a new LLM provider constructor in the global registry
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return defaultRegistry.Register(name, constructor)
}

// GetProviders returns a list of all registered providers from the global registry
func GetProviders() []string {
	return defaultRegistry.List()
}

// HasProvider checks if a provider is registered in the global registry
func HasProvider(name string) bool {
	return defaultRegistry.Has(name)
}

// GetProviderConstructor retrieves a provider constructor from the global registry
func GetProviderConstructor(name string) (ProviderConstructor, bool) {
	return defaultRegistry.Get(name)
}

// ResetRegistry resets the global registry (for testing purposes only)
func ResetRegistry() {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.providers = make(map[string]ProviderConstructor)
}
