package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/provider"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProviderProber records calls instead of performing network I/O.
type fakeProviderProber struct {
	response *types.CompletionResponse
	err      error
	calls    int
	message  string
}

func (f *fakeProviderProber) Chat(_ context.Context, message string, _ []types.Message) (*types.CompletionResponse, error) {
	f.calls++
	f.message = message
	return f.response, f.err
}

// withFakeProber replaces the package-level prober factory for one test and
// returns a pointer that receives the client config passed to it.
func withFakeProber(t *testing.T, fake *fakeProviderProber) **types.ClientConfig {
	t.Helper()

	original := newProviderProber
	t.Cleanup(func() { newProviderProber = original })

	var captured *types.ClientConfig
	newProviderProber = func(cfg *types.ClientConfig) (providerProber, error) {
		captured = cfg
		return fake, nil
	}
	return &captured
}

func TestRunProviderTest_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
  api_base: https://api.example/v1
  model: gpt-test
  retries: 5
`)

	fake := &fakeProviderProber{response: &types.CompletionResponse{Content: "pong"}}
	captured := withFakeProber(t, fake)

	result, err := runProviderTest(context.Background(), manager, provider.NewService(manager), "openai", nil)
	require.NoError(t, err)

	assert.Equal(t, "pong", result.Response)
	assert.Equal(t, "openai", result.Info.Name)
	assert.Equal(t, "gpt-test", result.ClientConfig.Model)
	assert.Equal(t, "https://api.example/v1", result.ClientConfig.APIBase)
	assert.Equal(t, 1, fake.calls)
	assert.NotEmpty(t, fake.message)
	assert.Equal(t, 1, (*captured).Retries, "probe must force a single attempt")
	assert.GreaterOrEqual(t, result.Latency, time.Duration(0))
}

func TestRunProviderTest_Failure(t *testing.T) {
	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
  model: gpt-test
`)

	fake := &fakeProviderProber{err: errors.New("request failed with status 401: Unauthorized")}
	withFakeProber(t, fake)

	result, err := runProviderTest(context.Background(), manager, provider.NewService(manager), "openai", nil)
	require.Error(t, err)

	require.NotNil(t, result, "a failed probe still reports latency")
	assert.Equal(t, 1, fake.calls)

	var gptErr *gptcometerrors.GPTCometError
	require.True(t, errors.As(err, &gptErr))
	assert.Equal(t, gptcometerrors.ErrTitleProviderTest, gptErr.Title)
	assert.Contains(t, err.Error(), "did not respond successfully")
	assert.Contains(t, err.Error(), "status 401")
}

func TestRunProviderTest_MissingProviderConfig(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
`)

	fake := &fakeProviderProber{response: &types.CompletionResponse{Content: "pong"}}
	withFakeProber(t, fake)

	_, err := runProviderTest(context.Background(), manager, provider.NewService(manager), "deepseek", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Provider Not Configured")
	assert.Equal(t, 0, fake.calls, "no request should be made without a config block")
}

func TestProviderTestCmd_SuccessOutsideGitRepo(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Chdir(t.TempDir())

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
  api_base: https://api.example/v1
  model: gpt-test
`)

	fake := &fakeProviderProber{response: &types.CompletionResponse{Content: "pong"}}
	withFakeProber(t, fake)

	cmd := newProviderTestCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	require.NoError(t, cmd.Execute())

	output := buf.String()
	for _, want := range []string{
		"Testing provider 'openai'...",
		"model:    gpt-test",
		"api base: https://api.example/v1",
		"api key:  config (test****)",
		"Connected in",
		`response: "pong"`,
	} {
		assert.Contains(t, output, want)
	}
}

func TestProviderTestCmd_NamedProvider(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
deepseek:
  api_key: deepseek-key
  model: deepseek-chat
  api_base: https://api.deepseek.com/v1
`)

	fake := &fakeProviderProber{response: &types.CompletionResponse{Content: "pong"}}
	captured := withFakeProber(t, fake)

	cmd := newProviderTestCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"deepseek"})

	require.NoError(t, cmd.Execute())

	assert.Contains(t, buf.String(), "Testing provider 'deepseek'...")
	assert.Equal(t, "deepseek-chat", (*captured).Model)
	assert.Equal(t, 1, fake.calls)
}

func TestProviderTestCmd_NoProviderSelected(t *testing.T) {
	manager, _ := newProviderTestManager(t, `
output:
  lang: en
`)

	cmd := newProviderTestCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No Provider Selected")
}

func TestProviderTestCmd_Failure(t *testing.T) {
	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
  model: gpt-test
`)

	fake := &fakeProviderProber{err: errors.New("request failed with status 503: unavailable")}
	withFakeProber(t, fake)

	cmd := newProviderTestCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Provider Test Failed")

	output := buf.String()
	assert.Contains(t, output, "Testing provider 'openai'...")
	assert.Contains(t, output, "model:    gpt-test")
}

func TestFormatProbeResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		want          string
		wantTruncated bool
	}{
		{name: "empty", response: "", want: ""},
		{name: "whitespace only", response: "  \n\t ", want: ""},
		{name: "simple", response: "pong", want: `"pong"`},
		{name: "collapses newlines", response: "line one\nline two", want: `"line one line two"`},
		{
			name:          "truncates long response",
			response:      strings.Repeat("a", providerProbePreviewLimit+20),
			wantTruncated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatProbeResponse(tt.response)
			if tt.wantTruncated {
				assert.Contains(t, got, "…")
				assert.LessOrEqual(t, len([]rune(got)), providerProbePreviewLimit+3)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunProviderTest_UnknownProvider(t *testing.T) {
	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
`)

	fake := &fakeProviderProber{response: &types.CompletionResponse{Content: "pong"}}
	withFakeProber(t, fake)

	_, err := runProviderTest(context.Background(), manager, provider.NewService(manager), "no-such-provider", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown Provider")
	assert.Equal(t, 0, fake.calls)
}

func TestProviderTestCmd_NormalizesProviderName(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: test-key
deepseek:
  api_key: deepseek-key
  model: deepseek-chat
  api_base: https://api.deepseek.com/v1
`)

	fake := &fakeProviderProber{response: &types.CompletionResponse{Content: "pong"}}
	captured := withFakeProber(t, fake)

	cmd := newProviderTestCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"DeepSeek"})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "Testing provider 'deepseek'...")
	assert.Equal(t, "deepseek-chat", (*captured).Model)
}
