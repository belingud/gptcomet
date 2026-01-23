package factory

import (
	"github.com/belingud/gptcomet/internal/client"
	"github.com/belingud/gptcomet/internal/config"
	gptErrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/git"
	"github.com/belingud/gptcomet/pkg/types"
)

// ServiceDependencies contains the core dependencies needed by command services.
// It provides VCS for version control operations, ConfigManager for configuration
// access, and APIConfig for LLM client initialization.
type ServiceDependencies struct {
	VCS         git.VCS
	CfgManager  config.ManagerInterface
	APIConfig   *types.ClientConfig
	APIClient   client.ClientInterface
}

// ServiceOptions contains configuration options for service creation.
type ServiceOptions struct {
	UseSVN     bool
	ConfigPath string
	Provider   string
}

// NewServiceDependencies creates the core service dependencies (VCS and ConfigManager).
// This is useful when you need to initialize services before applying command-line
// overrides or creating the API client.
//
// Parameters:
//   - options: ServiceOptions containing VCS type and configuration path
//
// Returns:
//   - git.VCS: Version control system instance (Git or SVN)
//   - config.ManagerInterface: Configuration manager for accessing config values
//   - error: Error if VCS or config initialization fails
func NewServiceDependencies(options ServiceOptions) (git.VCS, config.ManagerInterface, error) {
	vcsType := git.Git
	if options.UseSVN {
		vcsType = git.SVN
	}

	vcs, err := git.NewVCS(vcsType)
	if err != nil {
		return nil, nil, gptErrors.VCSCreationError(string(vcsType), err)
	}

	cfgManager, err := config.New(options.ConfigPath)
	if err != nil {
		return nil, nil, gptErrors.DependencyCreationError("config manager", err)
	}

	return vcs, cfgManager, nil
}

// NewServiceDependenciesWithClient creates all service dependencies including the API client.
// It initializes VCS, ConfigManager, retrieves client configuration, and creates the API client.
//
// This is the most complete factory function and should be used when you need all dependencies
// at once, including a configured API client for LLM operations.
//
// Parameters:
//   - options: ServiceOptions containing VCS type, config path, and provider override
//
// Returns:
//   - *ServiceDependencies: All initialized service dependencies
//   - error: Error if any initialization step fails
func NewServiceDependenciesWithClient(options ServiceOptions) (*ServiceDependencies, error) {
	vcs, cfgManager, err := NewServiceDependencies(options)
	if err != nil {
		return nil, err
	}

	clientConfig, err := cfgManager.GetClientConfig(options.Provider)
	if err != nil {
		return nil, gptErrors.DependencyCreationError("client config", err)
	}

	apiClient, err := client.New(clientConfig)
	if err != nil {
		return nil, gptErrors.DependencyCreationError("API client", err)
	}

	return &ServiceDependencies{
		VCS:       vcs,
		CfgManager: cfgManager,
		APIConfig:  clientConfig,
		APIClient:  apiClient,
	}, nil
}

// NewAPIClient creates an API client with the given configuration.
// This is useful when you already have a ClientConfig and need to create or recreate
// the API client (for example, after modifying the config).
//
// Parameters:
//   - clientConfig: Configuration for the API client
//
// Returns:
//   - client.ClientInterface: Configured API client
//   - error: Error if client creation fails
func NewAPIClient(clientConfig *types.ClientConfig) (client.ClientInterface, error) {
	apiClient, err := client.New(clientConfig)
	if err != nil {
		return nil, gptErrors.DependencyCreationError("API client", err)
	}
	return apiClient, nil
}
