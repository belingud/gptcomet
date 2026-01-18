package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/belingud/gptcomet/internal/client"
	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/internal/git"
	"github.com/belingud/gptcomet/internal/ui"
	"github.com/belingud/gptcomet/pkg/config/defaults"
	"github.com/belingud/gptcomet/pkg/types"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ReviewOptions contains the configuration settings for the review operation.
type ReviewOptions struct {
	RepoPath         string
	UseSVN           bool
	ConfigPath       string
	Stream           bool
	APIBase          string
	APIKey           string
	MaxTokens        int
	Retries          int
	Model            string
	AnswerPath       string
	CompletionPath   string
	Proxy            string
	FrequencyPenalty float64
	Temperature      float64
	TopP             float64
	Provider         string
}

// MarkdownRenderer interface for mocking in tests
type MarkdownRenderer interface {
	Render(text string, style string) (string, error)
}

// GlamourRenderer is the default implementation of MarkdownRenderer
type GlamourRenderer struct{}

func (r *GlamourRenderer) Render(text string, style string) (string, error) {
	return glamour.Render(text, style)
}

// ReviewService handles the logic for reviewing code changes
// while integrating with the GPT service for generating review comments.
type ReviewService struct {
	vcs              git.VCS
	client           client.ClientInterface
	cfgManager       config.ManagerInterface
	options          ReviewOptions
	editor           TextEditor
	markdownRenderer MarkdownRenderer
	clientConfig     *types.ClientConfig
}

const defaultReviewLanguage = "en"

// NewReviewService creates a new ReviewService instance with the provided options.
func NewReviewService(options ReviewOptions) (*ReviewService, error) {
	vcs, cfgManager, err := createServiceDependencies(struct {
		UseSVN     bool
		ConfigPath string
	}{
		UseSVN:     options.UseSVN,
		ConfigPath: options.ConfigPath,
	})
	if err != nil {
		return nil, err
	}

	clientConfig, err := cfgManager.GetClientConfig(options.Provider)
	if err != nil {
		return nil, err
	}

	// Overwrite client config with command line flags
	if options.APIBase != "" {
		clientConfig.APIBase = options.APIBase
	}
	if options.APIKey != "" {
		clientConfig.APIKey = options.APIKey
	}
	if options.MaxTokens > 0 {
		clientConfig.MaxTokens = options.MaxTokens
	}
	if options.Retries > 0 {
		clientConfig.Retries = options.Retries
	}
	if options.Model != "" {
		clientConfig.Model = options.Model
	}
	if options.AnswerPath != "" {
		clientConfig.AnswerPath = options.AnswerPath
	}
	if options.CompletionPath != "" {
		clientConfig.CompletionPath = &options.CompletionPath
	}
	if options.Proxy != "" {
		clientConfig.Proxy = options.Proxy
	}
	if options.FrequencyPenalty != 0 {
		clientConfig.FrequencyPenalty = options.FrequencyPenalty
	}
	if options.Temperature != 0 {
		clientConfig.Temperature = options.Temperature
	}
	if options.TopP != 0 {
		clientConfig.TopP = options.TopP
	}

	return &ReviewService{
		vcs:              vcs,
		client:           client.New(clientConfig),
		cfgManager:       cfgManager,
		options:          options,
		editor:           &TerminalEditor{},
		markdownRenderer: &GlamourRenderer{}, // Inject the renderer
		clientConfig:     clientConfig,
	}, nil
}

// Execute performs the review operation
func (s *ReviewService) Execute() error {
	// Get verbose setting
	verbose := s.getVerboseSetting()

	// Initialize progress tracking if verbose
	var progress *ui.Progress
	if verbose {
		progress = ui.NewProgress(true)
		progress.AddStages("Fetching diff", "Generating review")
	}

	if progress != nil {
		progress.Start("Fetching diff")
	}

	diff, err := s.getDiff()
	if err != nil {
		if progress != nil {
			progress.Error("Fetching diff", err)
		}
		return err
	}

	if progress != nil {
		progress.Complete("Fetching diff")
		progress.StartWithNewLine("Generating review")
	}

	// Get provider and model from configuration
	fmt.Printf("Discovered provider: %s, model: %s\n", s.clientConfig.Provider, s.clientConfig.Model)

	// Use streaming mode if the option is enabled
	if s.options.Stream {
		if progress != nil {
			// Complete current stage in new line, have a checkmark prefix
			progress.CompleteInNewLine("Generating review")
		}
		return s.ExecuteStream(diff)
	}

	// Otherwise use the standard non-streaming mode
	reviewComment, err := s.generateReviewComment(diff)
	if err != nil {
		if progress != nil {
			progress.Error("Generating review", err)
		}
		return err
	}

	if progress != nil {
		progress.CompleteInNewLine("Generating review")
	}

	formattedComment, err := s.formatReviewComment(reviewComment)
	if err != nil {
		return err
	}

	fmt.Println(formattedComment)
	return nil
}

// getVerboseSetting retrieves the console.verbose configuration
func (s *ReviewService) getVerboseSetting() bool {
	if val, ok := s.cfgManager.GetNestedValue([]string{"console", "verbose"}); ok {
		if verbose, ok := val.(bool); ok {
			return verbose
		}
	}
	return false
}

// ExecuteStream performs the review operation with streaming output
func (s *ReviewService) ExecuteStream(diff string) error {
	if diff == "" {
		return fmt.Errorf("empty diff provided")
	}

	prompt := s.cfgManager.GetReviewPrompt()
	if prompt == "" {
		return fmt.Errorf("empty review prompt configured")
	}

	reviewLang, err := s.getConfiguredReviewLanguage()
	if err != nil {
		return fmt.Errorf("failed to get review language: %w", err)
	}

	prompt = strings.ReplaceAll(prompt, "{{ output.review_lang }}", reviewLang)
	debug.Printf("Generating streaming review comment for diff length: %d\n", len(diff))

	fmt.Println(formatRemindMessage("Reviewing, streaming results as they arrive..."))

	// Use a buffer to accumulate the response for formatting at the end
	var responseBuffer strings.Builder

	// Define the callback function that will be called with each chunk of the response
	err = s.client.GenerateReviewCommentStream(diff, prompt, func(chunk string) error {
		// Print the chunk directly to the console
		fmt.Print(chunk)
		// Also accumulate it for final formatting
		responseBuffer.WriteString(chunk)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to generate streaming review comment: %w", err)
	}

	return nil
}

// getDiff retrieves the diff either from piped input or staged changes
func (s *ReviewService) getDiff() (string, error) {
	if s.isInputFromPipe() {
		return readPipedInput()
	}
	return s.getStagedDiff()
}

// getStagedDiff retrieves and returns filtered diff of staged changes
func (s *ReviewService) getStagedDiff() (string, error) {
	hasStagedChanges, err := s.vcs.HasStagedChanges(s.options.RepoPath)
	if err != nil {
		return "", fmt.Errorf("failed to check staged changes: %w", err)
	}
	if !hasStagedChanges {
		return "", fmt.Errorf("no staged changes found")
	}

	diff, err := s.vcs.GetStagedDiffFiltered(s.options.RepoPath, s.cfgManager)
	debug.Printf("Got staged diff length: %d\n", len(diff))
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	if diff == "" {
		return "", fmt.Errorf("no changes found after filtering")
	}

	return diff, nil
}

// generateReviewComment creates a review comment using the provided diff
// and configured prompt template.
func (s *ReviewService) generateReviewComment(diff string) (string, error) {
	if diff == "" {
		return "", fmt.Errorf("empty diff provided")
	}

	prompt := s.cfgManager.GetReviewPrompt()
	if prompt == "" {
		return "", fmt.Errorf("empty review prompt configured")
	}

	reviewLang, err := s.getConfiguredReviewLanguage()
	if err != nil {
		return "", fmt.Errorf("failed to get review language: %w", err)
	}

	prompt = strings.ReplaceAll(prompt, "{{ output.review_lang }}", reviewLang)
	debug.Printf("Generating review comment for diff length: %d\n", len(diff))

	fmt.Println(formatRemindMessage("Reviwing, may take a few seconds, you can set --stream/-s to stream the results..."))
	comment, err := s.client.GenerateReviewComment(diff, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate review comment: %w", err)
	}

	return comment, nil
}

// formatReviewComment formats the review comment using the configured markdown theme.
func (s *ReviewService) formatReviewComment(comment string) (string, error) {
	markdownTheme := s.getConfiguredMarkdownTheme()
	out, err := s.markdownRenderer.Render(comment, markdownTheme)
	if err != nil {
		// return original comment if rendering fails.
		return formatHighlightedMessage(comment), err
	}
	return formatHighlightedMessage(out), nil
}

// getConfiguredMarkdownTheme retrieves the configured markdown theme or returns the default style
func (s *ReviewService) getConfiguredMarkdownTheme() string {
	markdownThemeValue := s.cfgManager.GetWithDefault(MARKDOWN_THEME, defaults.DefaultConfig[MARKDOWN_THEME])

	markdownTheme, ok := markdownThemeValue.(string)
	if !ok {
		debug.Printf("Invalid markdown theme type %T, using default", markdownThemeValue)
		return styles.AutoStyle
	}

	if markdownTheme == "" {
		debug.Printf("Empty markdown theme configured, using default")
		return styles.AutoStyle
	}

	return markdownTheme
}

// getConfiguredReviewLanguage retrieves the review language from configuration
func (s *ReviewService) getConfiguredReviewLanguage() (string, error) {
	reviewLangValue, ok := s.cfgManager.Get(REVIEW_LANG_KEY)
	if !ok {
		debug.Printf("No review language configured, using default '%s'", defaultReviewLanguage)
		return config.OutputLanguageMap[defaultReviewLanguage], nil
	}

	reviewLang, ok := reviewLangValue.(string)
	if !ok {
		return "", fmt.Errorf("invalid review language configuration: expected string, got %T (%v)", reviewLangValue, reviewLangValue)
	}

	if reviewLang == "" {
		debug.Printf("Empty review language configured, using default '%s'", defaultReviewLanguage)
		return config.OutputLanguageMap[defaultReviewLanguage], nil
	}

	return config.OutputLanguageMap[reviewLang], nil
}

// isInputFromPipe checks if the program is receiving input from a pipe.
func (s *ReviewService) isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
}

// readPipedInput reads input from the standard input stream if it is a pipe
func readPipedInput() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewReviewCmd returns a new cobra.Command for the "review" subcommand.
func NewReviewCmd() *cobra.Command {
	options := ReviewOptions{}

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Generate and display review comments for code changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}
			options.ConfigPath = configPath

			service, err := NewReviewService(options)
			if err != nil {
				return err
			}

			return service.Execute()
		},
	}

	var generalFlags = pflag.NewFlagSet("General Flag", pflag.ExitOnError)
	var advancedFlags = pflag.NewFlagSet("Overwrite Flag", pflag.ExitOnError)

	// General Flags
	generalFlags.StringVarP(&options.RepoPath, "repo", "r", ".", "Path to the repository")
	generalFlags.BoolVarP(&options.UseSVN, "svn", "v", false, "Use SVN instead of Git")
	generalFlags.BoolVarP(&options.Stream, "stream", "s", false, "Stream output as it arrives from the LLM")
	generalFlags.StringVarP(&options.ConfigPath, "config", "c", "", "Path to the configuration file")

	// Advanced Flags
	advancedFlags.StringVar(&options.APIBase, "api-base", "", "Override API base URL")
	advancedFlags.StringVar(&options.APIKey, "api-key", "", "Override API key")
	advancedFlags.IntVar(&options.MaxTokens, "max-tokens", 0, "Override maximum tokens")
	advancedFlags.IntVar(&options.Retries, "retries", 0, "Override retry count")
	advancedFlags.StringVar(&options.Model, "model", "", "Override model name")
	advancedFlags.StringVar(&options.AnswerPath, "answer-path", "", "Override answer path")
	advancedFlags.StringVar(&options.CompletionPath, "completion-path", "", "Override completion path")
	advancedFlags.StringVar(&options.Proxy, "proxy", "", "Override proxy URL")
	advancedFlags.Float64Var(&options.FrequencyPenalty, "frequency-penalty", 0, "Override frequency penalty")
	advancedFlags.Float64Var(&options.Temperature, "temperature", 0, "Override temperature")
	advancedFlags.Float64Var(&options.TopP, "top-p", 0, "Override top_p value")
	advancedFlags.StringVar(&options.Provider, "provider", "", "Override AI provider (openai/deepseek)")

	// Add flag groups to command
	cmd.Flags().AddFlagSet(generalFlags)
	cmd.Flags().AddFlagSet(advancedFlags)

	// Organize flags in help output
	cmd.Flags().SetInterspersed(false)
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
		fmt.Println("\nUsage:")
		fmt.Printf("  %s\n", cmd.UseLine())
		fmt.Println("\nGeneral Flags:")
		generalFlags.PrintDefaults()
		fmt.Println("\nOverwrite Flags:")
		advancedFlags.PrintDefaults()
		fmt.Println()
		fmt.Println(`Global Flags:
  -d, --debug           Enable debug mode`)
	})

	return cmd
}
