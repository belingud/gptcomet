// Package provider exposes provider inventory information and the switch logic
// behind `gmsg provider list` and `gmsg provider use`.
//
// It performs no network I/O and reads configuration only through
// config.ManagerInterface, so the command layer stays thin and the rules are
// testable with a temporary config file.
package provider

import (
	"sort"
	"strings"

	"github.com/belingud/gptcomet/internal/config"
	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/llm"
	pkgconfig "github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// nonProviderRootKeys lists top-level configuration sections that are not
// provider configuration blocks.
var nonProviderRootKeys = map[string]struct{}{
	"provider":    {},
	"file_ignore": {},
	"output":      {},
	"console":     {},
	"prompt":      {},
}

// Info describes one provider for display purposes. Model and APIBase fall back
// to the provider's built-in defaults when the config file does not set them,
// so the values are explanatory rather than authoritative.
type Info struct {
	Name              string
	Active            bool
	Registered        bool
	Configured        bool
	Model             string
	ModelFromConfig   bool
	APIBase           string
	APIBaseFromConfig bool
	KeySource         config.APIKeySource
	KeyMasked         string
	KeyEnvVar         string
}

// SwitchResult describes the outcome of Service.Use.
type SwitchResult struct {
	Provider      string
	AlreadyActive bool
	Registered    bool
	Seeded        bool
	KeySource     config.APIKeySource
	KeyMasked     string
	KeyEnvVar     string
	MissingFields []string
}

// Service answers provider inventory questions and switches the default
// provider.
type Service struct {
	cfg config.ManagerInterface
}

// NewService creates a provider service backed by cfg.
func NewService(cfg config.ManagerInterface) *Service {
	return &Service{cfg: cfg}
}

// CurrentName returns the configured default provider, or "" when unset.
func (s *Service) CurrentName() string {
	raw, ok := s.cfg.Get("provider")
	if !ok {
		return ""
	}
	name, _ := raw.(string)
	return strings.TrimSpace(name)
}

// List returns one Info per known provider, sorted by name. Known providers are
// all registered LLM providers plus any custom provider that has a config
// section.
func (s *Service) List() []Info {
	names := make(map[string]struct{})
	for _, name := range llm.GetProviders() {
		names[name] = struct{}{}
	}
	for name, raw := range s.cfg.ListWithoutPrompt() {
		if _, reserved := nonProviderRootKeys[name]; reserved {
			continue
		}
		if _, ok := raw.(map[string]interface{}); ok {
			names[name] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)

	infos := make([]Info, 0, len(ordered))
	for _, name := range ordered {
		infos = append(infos, s.Get(name))
	}
	return infos
}

// Get returns the display information for a single provider.
func (s *Service) Get(name string) Info {
	section, configured := s.section(name)

	info := Info{
		Name:       name,
		Active:     name == s.CurrentName(),
		Registered: llm.HasProvider(name),
		Configured: configured,
	}

	if configured {
		if model, ok := section["model"].(string); ok && model != "" {
			info.Model = model
			info.ModelFromConfig = true
		}
		if apiBase, ok := section["api_base"].(string); ok && apiBase != "" {
			info.APIBase = apiBase
			info.APIBaseFromConfig = true
		}
	}

	if info.Registered && !info.ModelFromConfig {
		info.Model = providerDefault(name, "model")
	}
	if info.Registered && !info.APIBaseFromConfig {
		info.APIBase = providerDefault(name, "api_base")
	}

	key, source := s.cfg.ResolveAPIKey(name)
	info.KeySource = source
	info.KeyEnvVar = gptcometerrors.GetAPIKeyEnvVar(name)
	if source == config.APIKeySourceConfig {
		info.KeyMasked = config.MaskAPIKey(key, 4)
	}

	return info
}

// NormalizeName returns the canonical spelling of a provider name: surrounding
// spaces are trimmed and a known provider is lowercased ("OpenAI" -> "openai").
// Unknown names are returned trimmed so custom providers stay usable.
func (s *Service) NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || llm.HasProvider(name) || s.hasSection(name) {
		return name
	}
	lower := strings.ToLower(name)
	if llm.HasProvider(lower) || s.hasSection(lower) {
		return lower
	}
	return name
}

// Use switches the default provider to name.
//
// A registered provider without a config section gets one seeded from its
// GetRequiredConfig() defaults so later commands can resolve a client config.
// An existing section is never modified. Unregistered providers are accepted
// only when they already have a config section (they use the built-in
// OpenAI-compatible fallback at runtime).
func (s *Service) Use(name string) (*SwitchResult, error) {
	name = s.NormalizeName(name)
	registered := llm.HasProvider(name)
	section, configured := s.section(name)

	if name == "" || (!registered && !configured) {
		return nil, gptcometerrors.ProviderNotFoundError(name)
	}

	current := s.CurrentName()
	seeded := false
	if !configured {
		seededSection := seedSection(name)
		if len(seededSection) > 0 {
			s.cfg.SetNestedValue([]string{name}, seededSection)
			section = seededSection
			seeded = true
		}
	}

	if err := s.cfg.Set("provider", name); err != nil {
		return nil, err
	}

	key, source := s.cfg.ResolveAPIKey(name)
	result := &SwitchResult{
		Provider:      name,
		AlreadyActive: current == name,
		Registered:    registered,
		Seeded:        seeded,
		KeySource:     source,
		KeyEnvVar:     gptcometerrors.GetAPIKeyEnvVar(name),
	}
	if source == config.APIKeySourceConfig {
		result.KeyMasked = config.MaskAPIKey(key, 4)
	}
	if registered {
		result.MissingFields = missingRequiredFields(name, section)
	}

	return result, nil
}

// section returns the provider's config map, if it exists.
func (s *Service) section(name string) (map[string]interface{}, bool) {
	raw, ok := s.cfg.GetNestedValue([]string{name})
	if !ok {
		return nil, false
	}
	section, ok := raw.(map[string]interface{})
	return section, ok
}

func (s *Service) hasSection(name string) bool {
	_, ok := s.section(name)
	return ok
}

// providerDefault returns the provider's built-in default for a required config
// field, using the same GetRequiredConfig() source as `gmsg provider new`.
func providerDefault(name, key string) string {
	requirements := requiredConfig(name)
	if requirements == nil {
		return ""
	}
	requirement, ok := requirements[key]
	if !ok {
		return ""
	}
	return requirement.DefaultValue
}

// seedSection builds the initial config block for a registered provider from
// the non-empty defaults it advertises.
func seedSection(name string) map[string]interface{} {
	section := make(map[string]interface{})
	requirements := requiredConfig(name)
	if requirements == nil {
		return section
	}

	keys := make([]string, 0, len(requirements))
	for key := range requirements {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		requirement := requirements[key]
		if key == "api_key" {
			// Persist an explicit empty key so the config documents which field
			// to fill; GetClientConfig still falls back to the environment.
			section["api_key"] = ""
			continue
		}
		if requirement.DefaultValue == "" {
			continue
		}
		section[key] = requirement.DefaultValue
	}

	return section
}

// missingRequiredFields reports required config fields whose default is empty
// (so they cannot be seeded) and which the config still does not provide.
func missingRequiredFields(name string, section map[string]interface{}) []string {
	requirements := requiredConfig(name)
	if requirements == nil {
		return nil
	}

	missing := make([]string, 0)
	for key, requirement := range requirements {
		if key == "api_key" || requirement.DefaultValue != "" {
			continue
		}
		value, _ := section[key].(string)
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

// requiredConfig builds a provider instance with empty config and returns its
// required-configuration metadata.
func requiredConfig(name string) map[string]pkgconfig.ConfigRequirement {
	constructor, ok := llm.GetProviderConstructor(name)
	if !ok {
		return nil
	}
	instance := constructor(&types.ClientConfig{})
	if instance == nil {
		return nil
	}
	return instance.GetRequiredConfig()
}
