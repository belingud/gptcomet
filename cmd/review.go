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
	"github.com/belingud/gptcomet/pkg/config/defaults"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/spf13/cobra"
)

// ReviewOptions contains the configuration settings for the review operation.
type ReviewOptions struct {
	RepoPath   string
	UseSVN     bool
	ConfigPath string
	Stream     bool // Currently not used, consider removing if no plans to implement
}

// Validate checks if required fields are set and returns an error if not
func (o *ReviewOptions) Validate() error {
	if o.RepoPath == "" {
		return fmt.Errorf("repository path is required")
	}
	if o.ConfigPath == "" {
		return fmt.Errorf("config path is required")
	}
	return nil
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

	clientConfig, err := cfgManager.GetClientConfig()
	if err != nil {
		return nil, err
	}

	return &ReviewService{
		vcs:              vcs,
		client:           client.New(clientConfig),
		cfgManager:       cfgManager,
		options:          options,
		editor:           &TerminalEditor{},
		markdownRenderer: &GlamourRenderer{}, // Inject the renderer
	}, nil
}

// Execute performs the review operation
func (s *ReviewService) Execute() error {
	diff, err := s.getDiff()
	if err != nil {
		return err
	}

	reviewComment, err := s.generateReviewComment(diff)
	if err != nil {
		return err
	}

	formattedComment, err := s.formatReviewComment(reviewComment)
	if err != nil {
		return err
	}

	fmt.Println(formattedComment)
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

	fmt.Println(formatRemindMessage("Reviwing, may take a few seconds..."))
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

	cmd.Flags().BoolVar(&options.UseSVN, "svn", false, "Use SVN instead of Git")
	// cmd.Flags().BoolVar(&options.Stream, "stream", false, "Stream output") // Currently not used

	return cmd
}
