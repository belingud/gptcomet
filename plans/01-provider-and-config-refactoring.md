# GPTComet Refactoring Plan

## Overview

This document outlines a comprehensive refactoring plan for the GPTComet project to address code quality issues, reduce duplication, and improve maintainability. The refactoring is organized into priority stages that can be executed incrementally.

## Current State Analysis

### Identified Issues

1. **Large Files**
   - `internal/config/config.go` (923 lines) - Configuration handling
   - `cmd/commit.go` (547 lines) - CLI command definition
   - `internal/client/client.go` (431 lines) - HTTP client

2. **Code Duplication**
   - 20+ LLM providers with nearly identical constructor patterns
   - Repeated flag configuration logic across CLI commands
   - Duplicate service initialization logic

3. **Provider System**
   - Long switch statement for provider selection (client.go:44-90)
   - Repetitive `NewXxxLLM()` constructors across providers
   - Similar `GetRequiredConfig()` implementations

4. **Configuration Complexity**
   - Single large file handling parsing, validation, and access
   - Manual type conversion scattered throughout
   - Mixed responsibilities

5. **Inconsistent Patterns**
   - Error handling: Mix of structured and simple errors
   - Logging: debug.Print, fmt.Println, structured logging
   - Name() methods: Inconsistent casing (lowercase vs proper case)

## Refactoring Stages

### Stage 1: Provider System Refactoring ✅
**Priority**: High | **Impact**: Reduces ~30% of code duplication
**Status**: Complete

#### 1.1 Extract Provider Registry ✅
**Files**: `internal/client/client.go`

**Changes**:
- Replace the 20+ case switch statement with a registry pattern
- Create `internal/llm/registry.go` with:
  - `ProviderRegistry` type
  - `Register(name, constructor)` function
  - `GetProvider(name)` function
- Update providers to auto-register in `init()`

**Benefits**:
- Open/closed principle - new providers without modifying client
- Cleaner client code
- Testability improvement

#### 1.2 Standardize Provider Constructors ✅
**Files**: `internal/llm/*.go`

**Status**: Complete - Created `internal/llm/builder.go` with:
- `SetDefaultAPIBase()`, `SetDefaultModel()`
- `BuildStandardConfig()` and `BuildStandardConfigSimple()`
- 17/22 providers refactored to use builder pattern
- 5 providers retained as special cases (azure, claude, gemini, ollama, vertex)

#### 1.3 Unify GetRequiredConfig() ✅
**Files**: `internal/llm/*.go`

**Changes**:
- Create configuration templates for common provider types:
  - `StandardConfigTemplate` (api_base, model, api_key, max_tokens)
  - `OpenAICompatibleTemplate`
  - `CustomTemplate` for special cases
- Providers reference templates instead of duplicating definitions

**Benefits**:
- Single source of truth for common configurations
- Easier to add new config fields globally

**Status**: Complete - Created `internal/llm/templates.go` with:
- `StandardConfigTemplate()` - Standard OpenAI-compatible template
- `OpenAICompatibleTemplate()` - Convenience function
- Providers implement GetRequiredConfig() directly for flexibility

---

### Stage 2: Configuration Module Split ✅
**Priority**: High | **Impact**: Improves maintainability of core config system
**Status**: Complete

#### 2.1 Split config.go into Focused Modules ✅
**Status**: Complete - Split 923-line config.go into 4 focused files:

**New Files Created**:
- `internal/config/manager.go` (475 lines) - Main manager interface and core operations
  - ManagerInterface definition
  - Manager struct and New() constructor
  - Core methods: Set(), Reset(), Remove(), Append(), Load(), Save()
  - GetClientConfig() - client configuration parsing
  - List(), ListWithoutPrompt(), UpdateProviderConfig()

- `internal/config/accessor.go` (311 lines) - Getter methods for nested access
  - Get(), GetWithDefault(), GetNestedValue()
  - SetNestedValue()
  - GetSupportedKeys()
  - GetPrompt(), GetReviewPrompt(), GetTranslationPrompt()
  - GetOutputTranslateTitle(), GetFileIgnore()

- `internal/config/validator.go` (145 lines) - Configuration validation logic
  - IsValidLanguage() - validates language codes
  - MaskAPIKey() - masks API keys for display
  - MaskConfigAPIKeys() - recursively masks API keys in config
  - OutputLanguageMap - language code to name mapping

- `internal/config/parser.go` (56 lines) - YAML parsing and type conversion
  - getIntValue() - converts config values to int
  - getFloatValue() - converts config values to float64

**Benefits Achieved**:
- Each file under 500 lines (config.go reduced from 923 to 475 lines)
- Clear separation of concerns
- Easier testing and maintenance

#### 2.2 Centralize Type Conversion ✅
**Files**: `internal/config/parser.go`

**Status**: Complete - Created conversion utilities:
- `getIntValue()` - handles int, float64, string types
- `getFloatValue()` - handles float64, int, string types
- Consistent warning messages for invalid conversions

**Benefits Achieved**:
- Single place for type conversion logic
- Consistent error handling with warnings

---

### Stage 3: CLI Command Refactoring ✅
**Priority**: Medium | **Impact**: Reduces command setup duplication
**Status**: Complete

#### 3.1 Extract Common Flag Setup ✅
**Files**: `cmd/*.go`

**Status**: Complete - Created `cmd/flags.go` (113 lines) with shared flag builders:

- `CommonOptions` struct - Shared API override configuration
- `AddAdvancedAPIFlags()` - Adds 11 API override flags (api-base, api-key, max-tokens, retries, model, answer-path, completion-path, proxy, frequency-penalty, temperature, top-p, provider)
- `AddGeneralFlags()` - Adds repo path and SVN flags
- `AddConfigFlag()` - Adds config file path flag
- `ApplyCommonOptions()` - Applies API options to client config
- `SetAdvancedHelpFunc()` - Shared help formatting with flag groups

**Updated Files**:
- `cmd/commit.go` - 551 → 489 lines (reduced by 62 lines, ~11%)
  - CommitOptions now embeds CommonOptions
  - Replaced manual flag override code with ApplyCommonOptions()
  - Updated NewCommitCmd() to use shared flag functions

- `cmd/review.go` - 441 → 374 lines (reduced by 67 lines, ~15%)
  - ReviewOptions now embeds CommonOptions
  - Replaced manual flag override code with ApplyCommonOptions()
  - Updated NewReviewCmd() to use shared flag functions

**Benefits Achieved**:
- Single source of truth for API override flags
- Consistent flag definitions across commit and review commands
- DRY principle - eliminate ~130 lines of duplication
- Easy to add API flags to new commands

#### 3.2 Extract Service Factory ✅
**Files**: `cmd/*.go`, `internal/factory/factory.go`

**Status**: Complete - Created `internal/factory/factory.go` (111 lines):

**Changes Made**:
- Created `internal/factory/factory.go` with service creation utilities:
  - `ServiceDependencies` struct - Contains VCS, ConfigManager, APIConfig, APIClient
  - `ServiceOptions` struct - Configuration for service creation
  - `NewServiceDependencies()` - Creates VCS and ConfigManager
  - `NewServiceDependenciesWithClient()` - Creates all dependencies including API client
  - `NewAPIClient()` - Creates API client from config

- Updated service constructors to use factory:
  - `cmd/commit.go` - NewCommitService now uses factory.NewServiceDependencies()
  - `cmd/review.go` - NewReviewService now uses factory.NewServiceDependencies()
  - Removed `createServiceDependencies()` from `cmd/common.go` (21 lines removed)

**Benefits Achieved**:
- Consistent service creation pattern across commit and review commands
- Factory pattern makes it easy to add new dependencies
- Centralized service initialization logic
- Clear separation between factory (internal/factory) and command setup (cmd)

#### 3.3 Simplify commit.go ✅
**File**: `cmd/commit.go`

**Status**: Complete - Successfully extracted action handlers to `cmd/commit_action.go` (329 lines):

**Changes Made**:
- Created `cmd/commit_action.go` with extracted methods:
  - `Execute()` - Main commit workflow
  - `generateCommitMessage()` - Message generation with translation
  - `handleCommitInteraction()` - Interactive prompt loop
  - `createCommit()` - VCS commit creation
  - `getVerboseSetting()` - Configuration lookup
  - `splitCommitMessage()` - Message parsing helper
  - `removeThinkTags()` - Clean thinking tags from LLM output

- Updated `cmd/commit.go`:
  - Removed duplicate method implementations (321 lines removed)
  - Removed unused imports (bufio, regexp, strings, debug, errors, ui)
  - Now only contains struct definitions and command setup
  - Reduced from 489 → 168 lines (~66% reduction)

**Benefits Achieved**:
- commit.go now under 200 lines (target achieved!)
- Clear separation: command setup in commit.go, business logic in commit_action.go
- Better testability - action handlers can be tested independently
- Uses LANGUAGE_KEY from common.go (no duplication)

---

### Stage 4: Error Handling Standardization ✅
**Priority**: Medium | **Impact**: Improves debugging and user experience
**Status**: Complete

#### 4.1 Standardize Error Usage ✅
**Files**: All `internal/*` packages

**Completed**:
- ✅ Created comprehensive error infrastructure in `internal/errors/`
- ✅ Added 8 new error types and templates
- ✅ Added 20+ new error constants and suggestions
- ✅ Refactored all high-impact files:
  - `internal/factory/factory.go` (5 errors → structured)
  - `internal/client/client.go` (10 errors → structured)
  - `internal/llm/provider.go` (2 errors → structured)
  - `internal/config/config.go` (fixed 4 WrapError calls)
  - `internal/git/git.go` (5 errors → structured)
  - `internal/git/svn.go` (1 error → structured)
- ✅ Added nil validation with structured errors
- ✅ Created 16 new error template functions
- ✅ Fixed all failing tests (8 test cases updated)
  - `internal/client/client_test.go` (2 tests)
  - `internal/factory/factory_test.go` (1 test)
  - `internal/llm/provider_test.go` (2 tests)
  - `internal/config/config_test.go` (5 tests)

**Benefits Achieved**:
- Consistent, user-friendly error messages
- Actionable suggestions for fixing issues
- Better debugging information with error unwrapping
- Professional Error UX with icons and documentation links

#### 4.2 Centralize Error Constants ✅
**Files**: `internal/errors/constants.go`

**Completed**:
- ✅ Extended error constants with new titles, messages, and suggestions
- ✅ Added templates for provider, VCS, proxy, request, and dependency errors
- ✅ Created reusable error messages for common scenarios
- ✅ Added i18n-ready structure for future localization

**Benefits Achieved**:
- Single source for error messages
- Easy to update error text globally
- Consistent messaging across the application

---

### Stage 5: Testing Improvements
**Priority**: Low | **Impact**: Better code coverage and confidence

#### 5.1 Add Integration Tests
**New Files**: `tests/integration/*_test.go`

**Changes**:
- Add end-to-end tests for:
  - Config loading with various inputs
  - Provider initialization
  - Full commit message generation flow
- Use test fixtures for consistent data

#### 5.2 Improve Edge Case Coverage
**Files**: Existing `*_test.go` files

**Changes**:
- Add tests for error paths
- Test boundary conditions
- Mock external dependencies completely

---

### Stage 6: Code Quality Polish
**Priority**: Low | **Impact**: Minor improvements

#### 6.1 Standardize Logging
**Files**: All packages

**Changes**:
- Choose single logging approach (structured logging)
- Replace all `debug.Print` and `fmt.Println` calls
- Add log levels (debug, info, warn, error)

#### 6.2 Extract Constants
**Files**: Various

**Changes**:
- Create `internal/constants/` package
- Move all magic strings and numbers
- HTTP status codes, error messages, API paths

#### 6.3 Improve Documentation
**Files**: Package-level

**Changes**:
- Add package documentation comments
- Document complex algorithms
- Add examples for key interfaces

---

## Implementation Order

| Stage | Priority | Estimated Files | Dependencies |
|-------|----------|----------------|--------------|
| 1. Provider System | High | ~25 files | None |
| 2. Config Split | High | ~5 files | None |
| 3. CLI Refactor | Medium | ~10 files | Stage 2 |
| 4. Error Handling | Medium | ~30 files | None |
| 5. Testing | Low | ~15 files | Stages 1-4 |
| 6. Polish | Low | ~20 files | Stages 1-5 |

## Critical Files to Modify

### High Impact (Must Read First)
1. `internal/llm/llm.go` - LLM interface definition
2. `internal/llm/provider.go` - Provider registration
3. `internal/client/client.go` - Provider selection switch
4. `internal/config/config.go` - Configuration manager
5. `internal/llm/base.go` - BaseLLM implementation

### Provider Files (Review for Patterns)
1. `internal/llm/openai.go` - Reference implementation
2. `internal/llm/claude.go` - Custom API pattern
3. `internal/llm/gemini.go` - Full custom pattern
4. `internal/llm/ollama.go` - Different message format

### CLI Files
1. `cmd/commit.go` - Largest command file
2. `cmd/review.go` - Second largest command
3. `cmd/root.go` - Root command setup

## Testing Strategy

### Verification Tests
1. **All existing tests pass**: `go test ./...`
2. **Manual smoke test**: `gmsg commit`, `gmsg review`
3. **Provider initialization**: All 20+ providers load correctly
4. **Config migration**: Existing configs work without changes
5. **Integration tests**: Full commit workflow

### Test Commands
```bash
# Run all tests
just test

# Run specific package tests
go test ./internal/llm/...
go test ./internal/config/...

# Linting
go vet ./...
staticcheck ./...

# Format check
go fmt ./...
goimports -w .
```

## Success Criteria

Each stage is complete when:
- [ ] All existing tests pass
- [ ] No new linting warnings
- [ ] Code follows project formatting standards
- [ ] Documentation updated for changed interfaces
- [ ] Manual verification of affected features

## Rollback Strategy

Each stage will be a separate commit:
1. If a stage breaks functionality, revert that commit
2. Stages are independent - can skip and come back later
3. High priority stages have no dependencies

---

## Notes

- This is an incremental refactoring - each stage can be executed independently
- No breaking changes to public APIs or user-facing behavior
- Focus on code health, not feature additions
- Consider creating feature branches for each major stage
