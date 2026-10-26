# GPTComet 未完成功能修复计划

> 分支：`fix/incomplete-features`
> 基线：`master` @ `bb5578c`（v2.5.1）
> 状态：已确认，实施中（计划提交只新增本文件；后续按工作项独立提交）
> 范围：仅修复代码审查中确认的、代码层可见的未完成/半成品问题。不包含 VSCode 扩展、Git hooks、JetBrains/Vim/Emacs、团队/SaaS、覆盖率门禁、ARM64 发布矩阵等路线图或工程化大项。

## 1. 背景

当前核心 CLI（`commit` / `review` / `config` / `newprovider` / `update`）可以正常构建，`go test ./...` 全绿，但存在几处“文档或配置层已经承诺、实际链路没有接上”的问题：

- 默认翻译 prompt 使用 `{{ output.lang }}` / `{{ placeholder }}`，`TranslateMessage` 却用 `fmt.Sprintf`，默认模板会被拼坏。
- `output.rich_template` 没有被替换，自定义 rich 模板不生效。
- `newprovider` 写入的 Azure / Vertex / Claude 专属配置字段没有被 `GetClientConfig` 读取；`retries` 因 YAML 整数类型而始终使用默认值。
- `gmsg config keys` 漏掉 `output.markdown_theme` 和 `prompt.review`。
- 错误提示建议用户设置环境变量 API key，但配置加载从不读取环境变量。
- SVN 的 `file_ignore` 参数被接收但未使用。
- 流式 review 的 `responseBuffer` 只写不读，流结束后的收尾没有完成。
- 没有 `gmsg config validate`，非法配置只能在真正调用 API 时才暴露。

本计划的目标是把这些“半成品”补成可验证、可回归测试的完整行为。

## 2. 原则与非目标

### 原则

1. **先修正确性，再补体验**：P0 只处理会导致请求内容错误或配置失效的问题。
2. **保持 CLI / 配置兼容**：不删除现有 flag、配置键或 provider 名称；默认值只在明确错误时修正。
3. **每个修复先写失败测试**：至少覆盖“修复前会失败、修复后通过”的回归用例。
4. **小步提交**：每个工作项独立 commit，便于回滚和 review。
5. **不扩大范围**：不顺手重构无关模块，不新增 provider。

### 非目标（另开分支/另开 PR）

- VSCode 扩展、Git hooks、IDE 集成、团队协作、SaaS/企业功能、插件系统。
- CI 覆盖率门禁、ARM64 发布矩阵、PyPI wheel 平台补齐。
- `BaseLLM.MakeRequest` 的流式分支重构（实际 review 流式走 `Client.Stream`，不在本次范围）。
- 新增 provider 或修改现有 provider 的默认模型/默认 API base。

## 3. 工作项

### P0-1 统一 prompt 渲染：翻译与 rich_template

**优先级**：P0（用户可见的正确性问题）

**现状**

- `Client.TranslateMessage` 使用 `fmt.Sprintf(prompt, message, lang)`（[internal/client/client.go](../internal/client/client.go)）。
- 默认翻译 prompt 使用 `{{ output.lang }}` 与 `{{ placeholder }}`（[pkg/config/defaults/defaults.go](../pkg/config/defaults/defaults.go)）。
- 当模板不含 `%s` 时，`fmt.Sprintf` 会追加 `%!(EXTRA string=..., string=...)`，消息没有落到 `{{ placeholder }}` 位置。
- `GenerateCommitMessage` 只替换 `{{ placeholder }}`，rich prompt 中的 `{{ output.rich_template }}` 从未被替换。

**修复方案**

1. 新增 `internal/prompt` 包：
   - `Render(template string, vars map[string]string) string`
   - 使用 `regexp.MustCompile(`\{\{\s*([\w.]+)\s*\}\}`)` 匹配占位符；只替换 `vars` 中存在的 key，未知占位符原样保留。
   - `HasPlaceholder(template string) bool` 供兼容分支使用。
2. `Client.TranslateMessage`：
   - 先调用 `prompt.Render`，变量为 `placeholder`、`output.lang`。
   - 如果模板没有任何 `{{ ... }}` 且包含 `%s`，保留旧的 `fmt.Sprintf` 行为，兼容用户自定义 `%s` 模板。
   - 两种路径都必须保证 message 和 lang 进入最终 prompt。
3. `CommitService.generateCommitMessage`：
   - 在调用 `client.GenerateCommitMessage` 前，读取 `output.rich_template`（缺省使用默认值），用 `prompt.Render` 替换 `{{ output.rich_template }}`。
   - 新增 `defaults.DefaultRichTemplate` 常量，`defaultConfig` 与读取逻辑共用，避免字符串散落。
4. review 路径保持现有行为；如低风险，可将 `{{ output.review_lang }}` 的替换也切到 `prompt.Render`，但不得改变输出。

**涉及文件**

- 新增：`internal/prompt/render.go`、`internal/prompt/render_test.go`
- 修改：`internal/client/client.go`、`cmd/commit_action.go`、`pkg/config/defaults/defaults.go`
- 测试：`internal/client/client_test.go`、`cmd/commit_test.go`

**测试**

- 默认翻译 prompt：渲染结果包含原始 commit message 和目标语言，不包含 `%!`。
- 自定义 `%s` 翻译 prompt：仍按旧格式工作。
- 含空格的 `{{ output.lang }}`：能正确替换。
- rich prompt：自定义模板出现在最终 prompt 中；默认模板也正确。
- 未知占位符：保持原样，不 panic。

**验收**

- 用默认配置、`output.lang=zh-cn` 走单测时，发给 mock LLM 的 prompt 不再出现 `%!(EXTRA ...)`。
- `--rich` 且自定义 `output.rich_template` 时，mock client 收到的 prompt 含自定义模板内容。

---

### P0-2 Provider 专属配置映射与 retries

**优先级**：P0（配置写入后不生效）

**现状**

- `newprovider` 会保存 `api_version`、`deployment_name`、`project_id`、`location`、`anthropic_version` 等字段。
- `Manager.GetClientConfig` 只映射通用字段（[internal/config/config.go](../internal/config/config.go)），上述字段不会进入 `types.ClientConfig`。
- Azure 构造函数用空的 `DeploymentName` 拼出 `deployments//chat/completions`；`APIVersion` 为空时 URL 与 header 都没有有效版本。
- `retries` 只在类型为 `float64` 时读取；YAML 整数是 `int`，因此 `retries: 5` 被忽略，始终得到默认值 3。
- 现有测试甚至断言了 `retries: 5` 仍返回 3 的错误行为（[internal/config/config_test.go](../internal/config/config_test.go)）。

**修复方案**

1. `GetClientConfig` 增加字段映射：
   - `api_version` -> `APIVersion`
   - `deployment_name` -> `DeploymentName`
   - `project_id` -> `ProjectID`
   - `location` -> `Location`
   - `anthropic_version` -> `AnthropicVersion`
   - `retries` 改用 `getIntValue(providerConfig, "retries", defaults.DefaultRetries)`
2. 为字符串字段增加一个小型 helper（如 `getStringValue`），避免重复类型断言。
3. `NewAzureLLM`：
   - 若 `DeploymentName` 为空，使用 `Model` 作为 deployment。
   - 若 `APIVersion` 为空，设置默认值（与 `GetRequiredConfig` 的默认值一致）。
   - 先确定 deployment，再拼接 `CompletionPath`。
4. 保持现有默认值不变：Vertex 仍使用 `default-project` / `us-central1`；只有用户显式配置时才覆盖。
5. 更新 `TestManager_GetClientConfig` 中 retries 的期望值，并新增 Azure/Vertex/Claude 配置映射用例。

**涉及文件**

- 修改：`internal/config/config.go`、`internal/config/parser.go`（如需 helper）、`internal/llm/azure.go`
- 测试：`internal/config/config_test.go`、`internal/llm/azure_test.go`、`internal/llm/vertex_test.go`（按需）

**测试**

- YAML 中 `retries: 5` -> `ClientConfig.Retries == 5`。
- YAML 中 Azure 的 `api_version`、`deployment_name` -> `ClientConfig` 字段正确。
- YAML 中 Vertex 的 `project_id`、`location` -> `ClientConfig` 字段正确。
- YAML 中 Claude 的 `anthropic_version` -> `ClientConfig.AnthropicVersion` 正确。
- `NewAzureLLM` 在 `DeploymentName` 为空、`Model` 非空时，`CompletionPath` 使用 model；`APIVersion` 为空时使用默认版本；`BuildURL` 不含空 deployment 或空 api-version。
- 现有未配置这些字段的测试仍通过。

**验收**

- 用 `newprovider` 生成的 Azure / Vertex 配置走 `GetClientConfig` 后，关键字段不再丢失。
- `retries` 的配置值在配置文件和命令行两条路径上一致生效。

---

### P1-3 补齐 `config keys`

**优先级**：P1

**现状**

- `GetSupportedKeys` 的 `outputKeys` 缺 `markdown_theme`，`promptKeys` 缺 `review`。
- 两者都在代码中使用，README 也记录了 `output.markdown_theme`。

**修复方案**

- 在 `internal/config/accessor.go` 的 `outputKeys` 加入 `markdown_theme`。
- 在 `promptKeys` 加入 `review`。
- README 的支持键表格补充 `prompt.review`（当前只有 brief/rich/translation）。
- 更新 `TestManager_GetSupportedKeys`，明确断言这两个 key 存在。

**涉及文件**

- 修改：`internal/config/accessor.go`、`README.md`
- 测试：`internal/config/accessor_test.go`

**验收**

- `gmsg config keys` 输出包含 `output.markdown_theme` 和 `prompt.review`。
- key 列表仍然排序、仍然使用 `<provider>` 占位符。

---

### P1-4 环境变量 API key

**优先级**：P1（错误提示已承诺，但实际不生效）

**现状**

- `APIKeyNotSetError` 生成 `export OPENAI_API_KEY=...` 建议（[internal/errors/templates.go](../internal/errors/templates.go)）。
- `GetClientConfig` 只读取 YAML 的 `api_key`，没有 `os.Getenv`。
- `getAPIKeyEnvVar` 当前是 errors 包私有函数，默认拼出 `unknown_API_KEY` 这种大小写混合的名字。

**修复方案**

1. 将 provider -> 环境变量名映射导出为 `errors.GetAPIKeyEnvVar(provider)`，错误提示与配置加载共用。
2. 默认规则改为 `strings.ToUpper(provider) + "_API_KEY"`；保留 OpenAI/Azure/Claude/Gemini/Cohere 的特殊名称。
3. `GetClientConfig`：
   - 先读配置文件 `api_key`。
   - 配置为空时，读取 `os.Getenv(GetAPIKeyEnvVar(provider))`。
   - Ollama 仍然不需要 key。
   - 两者都为空时才返回 `APIKeyNotSetError`。
4. 在 README 的 Configuration Methods 中补充环境变量说明，并列出常用变量名。
5. 更新 `TestGetAPIKeyEnvVar` 对 `unknown` 的期望为 `UNKNOWN_API_KEY`。

**涉及文件**

- 修改：`internal/errors/templates.go`、`internal/config/config.go`、`README.md`
- 测试：`internal/errors/errors_test.go`、`internal/config/config_test.go`

**测试**

- 配置无 api_key + `t.Setenv("OPENAI_API_KEY", "env-key")` -> 加载成功且 key 为 `env-key`。
- 配置有 api_key + 环境变量同时存在 -> 配置优先。
- 两者都没有 -> 仍返回 `API Key Not Configured`。
- Ollama 无 key 不报错。
- 错误提示中的变量名与实际读取的变量名一致。

**验收**

- 用户按错误提示 `export OPENAI_API_KEY=...` 后，不需要改 YAML 即可运行。

---

### P1-5 SVN 的 file_ignore

**优先级**：P1

**现状**

- `SVNVCS.GetStagedDiffFiltered(repoPath, cfgManager)` 接收 `cfgManager`，但完全没有使用。
- Git 路径已按 `file_ignore` 过滤，SVN 路径没有，行为不一致。

**修复方案**

1. 在 `internal/git` 中提取纯函数：
   - `FilterIgnoredFiles(files []string, patterns []string) []string`
   - 复用现有 `ShouldIgnoreFile`。
2. Git 与 SVN 的 `GetStagedDiffFiltered` 共用该函数；Git 现有行为不变。
3. SVN：
   - 先 `GetStagedFiles`。
   - 过滤后为空则返回 `""`（上层按无变更处理）。
   - 对以 `-` 开头的文件名做参数保护（优先用 `svn diff -- <files...>`，实现前先本地验证 svn 是否接受该分隔符；不接受则逐文件调用）。
4. 不改变 `svn status` 的解析规则。

**涉及文件**

- 修改：`internal/git/svn.go`、`internal/git/git.go`（如需复用）
- 测试：`internal/git/git_test.go`、`internal/git/svn_test.go`

**测试**

- 纯函数：精确匹配、通配符、全部被忽略、空列表。
- SVN 不依赖真实 `svn` 的单元测试：验证过滤后的文件列表；如果现有测试环境有 `svn`，再补一个端到端用例（保持可跳过）。
- Git 原有 `file_ignore` 测试不回归。

**验收**

- SVN 仓库中 `*.lock` 等被忽略文件不会出现在 `svn diff` 参数里。
- 全部文件被忽略时不会执行 `svn diff`。

---

### P2-6 流式 review 收尾与安全重试

**优先级**：P2

**现状**

- `ReviewService.ExecuteStream` 中的 `responseBuffer` 只写不读（[cmd/review.go](../cmd/review.go)）。
- `Client.Stream` 没有重试；网络瞬断时整个 review 直接失败。
- 流式输出结束后没有统一换行处理。

**修复方案**

1. 删除未使用的 `responseBuffer` 和误导性注释；明确流式模式就是实时输出，不做二次 markdown 渲染。
2. 在 `ExecuteStream` 中跟踪最后一个 chunk 是否以 `\n` 结尾，结束时只补一个换行，避免重复空行。
3. `Client.Stream` 增加“首个内容块之前”的安全重试：
   - 抽出 `streamOnce`。
   - 记录是否已经向 callback 输出过非空内容。
   - 仅在“尚未输出任何内容”且错误为网络类型（`GPTCometError.Type == ErrTypeNetwork`）时重试。
   - `context.Canceled` / `context.DeadlineExceeded` 不重试。
   - 一旦已经输出内容，后续错误直接返回，避免重复片段和错乱输出。
   - 退避复用 `constants.BaseRetryDelay` / `constants.MaxJitterPercent`，与 `Chat` 保持一致。
4. 不改变 `--stream` 的 flag 语义，不引入“先缓冲再打印”。

**涉及文件**

- 修改：`cmd/review.go`、`internal/client/client.go`
- 测试：`cmd/review_test.go`、`internal/client/client_test.go`

**测试**

- 服务端第一次网络失败、第二次成功：在未输出内容时重试并成功。
- 服务端返回 200 后发送一个 chunk 再断开：不重试，返回错误，且已输出内容不重复。
- 正常流：chunk 顺序、最终换行符合预期。
- `context.Canceled`：不重试。
- `ExecuteStream` 的 prompt 仍包含替换后的 `output.review_lang` 和 diff。

**验收**

- 流式输出仍然是实时的；失败场景不会出现重复片段。
- 网络瞬断在“尚未输出”阶段可自动恢复。

---

### P2-7 `gmsg config validate`

**优先级**：P2

**现状**

- `gmsg config` 没有 `validate` 子命令。
- `config set` 只校验 `output.lang` / `output.review_lang` / `output.translate_title`。
- 非法的 `output.markdown_theme`、缺失的 provider 字段、错误的数值范围，都要等到真正请求 API 才暴露。

**修复方案**

1. 在 `internal/config` 新增 `Validate() []string`（返回人类可读的问题列表，空表示通过）。
2. 校验规则（第一版保持保守，避免误报）：
   - `provider` 非空，且对应配置段存在且是 map。
   - 非 Ollama provider 必须有 `api_key`（YAML 或环境变量），否则报告。
   - `output.lang` / `output.review_lang` 必须是 `OutputLanguageMap` 中的值。
   - `output.markdown_theme` 必须是已知主题：`auto`、`ascii`、`dark`、`dracula`、`light`、`tokyo-night`、`notty`、`pink`。
   - `output.translate_title` 必须是 bool。
   - `max_tokens`、`retries` 必须是正整数；`top_p` 在 `[0,1]`；`temperature` 在 `[0,2]`；`frequency_penalty` 在 `[-2,2]`。
   - `file_ignore` 必须是字符串数组。
   - `prompt.brief_commit_message`、`prompt.rich_commit_message`、`prompt.translation`、`prompt.review` 如存在必须是非空字符串。
   - provider 专属：Azure 检查 `deployment_name`（允许用 `model` 作为 deployment 回退）；`api_version` 缺失不报错（构造函数有默认值）；Vertex 要求 `project_id` 和 `location`；Claude 不强制 `anthropic_version`（构造函数有默认值）。
3. `gmsg config validate`：
   - 读取当前配置，打印每个问题。
   - 有问题时返回错误（进程退出码非 0），无问题时打印通过信息。
   - 支持全局 `--config`。
4. `config set`：
   - 增加 `output.markdown_theme` 的合法性校验，非法值直接返回错误。
   - 不引入严格的未知 key 拒绝，避免破坏自定义 provider 场景。
5. `cmd.ConfigManager` 接口增加 `Validate() []string`；`testutils.MockConfigManager` 同步实现。
6. README / CLAUDE.md 的命令列表补充 `config validate`。

**涉及文件**

- 修改：`internal/config/config.go`、`internal/config/validator.go`、`cmd/config.go`、`cmd/config_test.go`、`internal/testutils/mock_config.go`、`README.md`、`CLAUDE.md`
- 测试：`internal/config/validator_test.go`、`internal/config/config_test.go`、`cmd/config_test.go`

**测试**

- 合法配置：返回空列表，命令成功。
- 缺 provider / 缺 api_key / 非法 lang / 非法 markdown theme / 负 retries / 非数组 file_ignore：分别报告对应问题。
- Azure/Vertex 缺必填字段：报告。
- `config set output.markdown_theme invalid`：返回错误；合法值成功。
- mock manager 的 validate 命令行为符合预期。

**验收**

- `gmsg config validate` 对配置了 API key（YAML 或环境变量）的配置返回成功；空配置明确报告缺少 API key。
- 构造一个非法配置时，命令打印具体问题并以非 0 退出。
- 不影响现有 `config set` / `config get` / `config keys` 行为。

---

## 4. 实施顺序

1. **Phase 1（P0）**：P0-1 prompt 渲染；P0-2 provider 配置与 retries。
2. **Phase 2（P1）**：P1-3 config keys；P1-4 环境变量 API key；P1-5 SVN file_ignore。
3. **Phase 3（P2）**：P2-6 流式 review；P2-7 config validate。
4. **Phase 4**：全量测试、文档同步、手动验收。

每个 phase 结束时跑一次 `go test ./...`，P0 完成后额外跑 `go test -race ./internal/client ./cmd/...`。

## 5. 提交拆分

建议 commit 粒度（每个都可独立回滚）：

1. `fix(prompt): render translation and rich template placeholders`
2. `fix(config): load provider-specific fields and retries`
3. `fix(config): list markdown theme and review prompt keys`
4. `feat(config): support API keys from environment variables`
5. `fix(svn): honor file_ignore when building diffs`
6. `fix(review): clean up streaming output and retry before first chunk`
7. `feat(config): add validate command`
8. `docs: document new config behavior and commands`

不强制 squash；如果最终 PR 需要整洁历史，可在合并前整理。

## 6. 测试与验收矩阵

### 自动化

```bash
go test ./...
go test -race ./...
go vet ./...
staticcheck ./...   # 如环境已安装
gofmt -l .          # 应无输出
```

Python 侧本次不改，仍建议跑一次：

```bash
pytest tests/py_tests/ -q
```

### 手动

```bash
# 1. 新分支构建
go build -o /tmp/gmsg .

# 2. 新配置目录，验证 keys
HOME=$(mktemp -d) /tmp/gmsg config keys | grep -E 'markdown_theme|prompt.review'

# 3. 验证 validate（默认配置缺 API key 时应报错；提供 key 后应通过）
HOME=$(mktemp -d) OPENAI_API_KEY=test-key /tmp/gmsg config validate

# 4. 构造非法配置，验证非 0 退出
#    （非法 lang 或 markdown_theme）

# 5. 验证翻译 prompt（单测覆盖，不发真实请求）
go test ./internal/client -run TestTranslateMessage -v
```

### 总体验收

- [ ] 所有 P0/P1/P2 工作项完成。
- [ ] `go test ./...` 全绿，新增回归测试覆盖每个修复。
- [ ] `go vet ./...` 无新增问题。
- [ ] `gofmt` 无变更。
- [ ] `gmsg config validate` 可用；提供 API key 后通过，空配置明确报错。
- [ ] `gmsg config keys` 包含 `output.markdown_theme`、`prompt.review`。
- [ ] 默认翻译 prompt 不再出现 `%!(EXTRA ...)`。
- [ ] Azure/Vertex/Claude 的专属字段能从 YAML 进入 `ClientConfig`。
- [ ] SVN 的 `file_ignore` 生效。
- [ ] 流式 review 不重复输出，且“首块前”网络错误可重试。
- [ ] README / CLAUDE.md 与行为一致。

## 7. 风险与回滚

| 风险 | 影响 | 缓解 |
| --- | --- | --- |
| 修改 prompt 渲染导致请求内容变化 | 中 | 保留 `%s` 兼容分支；用默认模板和自定义模板分别测试；不改变 review 的 `review_lang` 替换语义 |
| Provider 字段映射改变现有默认行为 | 中 | 只在字段存在时覆盖；保留 Vertex/Azure 默认值；更新并新增配置测试 |
| 环境变量优先级争议 | 低 | 明确“配置文件优先，环境变量兜底”；在 README 写清楚 |
| SVN diff 参数变化 | 中 | 纯函数测试 + 现有 SVN 测试（有 svn 时）；保留原有文件列表解析 |
| 流式重试导致重复输出 | 中 | 仅在尚未输出任何内容时重试；已有内容后立即返回错误；增加“部分输出后断开”测试 |
| config validate 误报 | 中 | 第一版只校验明确规则；provider 专属缺失用 warning/error 分级；未知 key 不拒绝 |

回滚策略：每个工作项独立 commit，出现问题可单独 revert；不涉及配置迁移、数据写入或发布流程。

## 8. 完成定义（Definition of Done）

- 代码、测试、文档三部分都完成。
- 新行为有自动化回归测试，且测试在修复前能复现问题。
- 不引入新的 lint/vet 警告。
- PR 描述中列出每个问题的“根因 / 修复 / 验证方式”。
- 不包含非范围项（VSCode、hooks、覆盖率门禁、ARM 发布矩阵）。
