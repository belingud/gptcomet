package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTempConfig creates a temporary config file for testing
func setupTempConfig(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "factory-test")
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
provider: openai
openai:
  api_key: test-api-key
  model: gpt-4
  max_tokens: 2048
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	return configPath, func() {
		os.RemoveAll(tmpDir)
	}
}

func TestNewServiceDependencies(t *testing.T) {
	t.Run("Success with Git", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()

		vcs, cfgManager, err := NewServiceDependencies(ServiceOptions{
			UseSVN:     false,
			ConfigPath: configPath,
			Provider:   "",
		})

		require.NoError(t, err)
		assert.NotNil(t, vcs, "VCS should not be nil")
		assert.NotNil(t, cfgManager, "ConfigManager should not be nil")
	})

	t.Run("Success with SVN", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()

		vcs, cfgManager, err := NewServiceDependencies(ServiceOptions{
			UseSVN:     true,
			ConfigPath: configPath,
			Provider:   "",
		})

		require.NoError(t, err)
		assert.NotNil(t, vcs, "VCS should not be nil")
		assert.NotNil(t, cfgManager, "ConfigPath should not be nil")
	})

	t.Run("Invalid config path", func(t *testing.T) {
		vcs, cfgManager, err := NewServiceDependencies(ServiceOptions{
			UseSVN:     false,
			ConfigPath: "/nonexistent/path/config.yaml",
			Provider:   "",
		})

		assert.Error(t, err, "Should return error for invalid config path")
		assert.Nil(t, vcs, "VCS should be nil on error")
		assert.Nil(t, cfgManager, "ConfigManager should be nil on error")
		assert.Contains(t, err.Error(), "Dependency Creation Failed", "Error should mention dependency creation")
	})
}

func TestNewServiceDependenciesWithClient(t *testing.T) {
	t.Run("Success with all dependencies", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()

		deps, err := NewServiceDependenciesWithClient(ServiceOptions{
			UseSVN:     false,
			ConfigPath: configPath,
			Provider:   "",
		})

		require.NoError(t, err)
		assert.NotNil(t, deps, "Dependencies should not be nil")
		assert.NotNil(t, deps.VCS, "VCS should not be nil")
		assert.NotNil(t, deps.CfgManager, "ConfigManager should not be nil")
		assert.NotNil(t, deps.APIConfig, "APIConfig should not be nil")
		assert.NotNil(t, deps.APIClient, "APIClient should not be nil")

		// Verify APIConfig has expected values
		assert.Equal(t, "openai", deps.APIConfig.Provider, "Provider should be openai")
		assert.Equal(t, "test-api-key", deps.APIConfig.APIKey, "APIKey should match config")
		assert.Equal(t, "gpt-4", deps.APIConfig.Model, "Model should match config")
		assert.Equal(t, 2048, deps.APIConfig.MaxTokens, "MaxTokens should match config")
	})

	t.Run("Invalid config path", func(t *testing.T) {
		deps, err := NewServiceDependenciesWithClient(ServiceOptions{
			UseSVN:     false,
			ConfigPath: "/nonexistent/path/config.yaml",
			Provider:   "",
		})

		assert.Error(t, err, "Should return error for invalid config path")
		assert.Nil(t, deps, "Dependencies should be nil on error")
	})

	t.Run("With provider override", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()

		// Test with default provider (no override)
		deps, err := NewServiceDependenciesWithClient(ServiceOptions{
			UseSVN:     false,
			ConfigPath: configPath,
			Provider:   "",
		})

		require.NoError(t, err)
		assert.NotNil(t, deps, "Dependencies should not be nil")
		assert.Equal(t, "openai", deps.APIConfig.Provider, "Provider should be openai")
		assert.Equal(t, "test-api-key", deps.APIConfig.APIKey, "APIKey should match openai config")
	})
}

func TestNewAPIClient(t *testing.T) {
	t.Run("Success with valid config", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()

		_, cfgManager, err := NewServiceDependencies(ServiceOptions{
			UseSVN:     false,
			ConfigPath: configPath,
			Provider:   "",
		})
		require.NoError(t, err)

		clientConfig, err := cfgManager.GetClientConfig("")
		require.NoError(t, err)

		client, err := NewAPIClient(clientConfig)
		require.NoError(t, err)
		assert.NotNil(t, client, "Client should not be nil")
	})

	t.Run("Nil config should error", func(t *testing.T) {
		client, err := NewAPIClient(nil)

		assert.Error(t, err, "Should return error for nil config")
		assert.Nil(t, client, "Client should be nil on error")
	})
}

func TestServiceOptions(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		opts := ServiceOptions{}

		assert.False(t, opts.UseSVN, "UseSVN should default to false")
		assert.Empty(t, opts.ConfigPath, "ConfigPath should default to empty")
		assert.Empty(t, opts.Provider, "Provider should default to empty")
	})
}

func TestServiceDependencies(t *testing.T) {
	t.Run("All fields can be nil", func(t *testing.T) {
		// This test verifies that a zero-value ServiceDependencies is valid
		// and won't cause nil pointer panans when checked
		var deps ServiceDependencies

		assert.Nil(t, deps.VCS, "VCS should be nil")
		assert.Nil(t, deps.CfgManager, "CfgManager should be nil")
		assert.Nil(t, deps.APIConfig, "APIConfig should be nil")
		assert.Nil(t, deps.APIClient, "APIClient should be nil")
	})
}
