# GPTComet：AI 驱动的 Git 提交信息生成和代码审查工具

<p align="center">
  <img src="artwork/logo.png" width="150" height="150" alt="GPTComet Logo">
</p>

<a href="https://www.producthunt.com/posts/gptcomet?embed=true&utm_source=badge-featured&utm_medium=badge&utm_source=badge-gptcomet" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=774818&theme=neutral&t=1747386848397" alt="GPTComet - GPTComet&#0058;&#0032;AI&#0045;Powered&#0032;Git&#0032;Commit&#0032;Message&#0032;Generator | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

[![PyPI version](https://img.shields.io/pypi/v/gptcomet?style=for-the-badge)](https://pypi.org/project/gptcomet/)
![GitHub Release](https://img.shields.io/github/v/release/belingud/gptcomet?style=for-the-badge)
[![License](https://img.shields.io/github/license/belingud/gptcomet.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/belingud/gptcomet?style=for-the-badge)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/belingud/gptcomet/release.yml?style=for-the-badge)
![PyPI - Downloads](https://img.shields.io/pypi/dm/gptcomet?logo=pypi&style=for-the-badge)
![Pepy Total Downloads](https://img.shields.io/pepy/dt/gptcomet?style=for-the-badge&logo=python)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/belingud/gptcomet/total?style=for-the-badge&label=Release%20Download)

[English](README.md)

<!-- TOC -->

- [GPTComet：AI 驱动的 Git 提交信息生成和代码审查工具](#gptcometai-驱动的-git-提交信息生成和代码审查工具)
  - [💡 概览](#-概览)
  - [✨ 功能](#-功能)
  - [⬇️ 安装](#️-安装)
  - [📕 使用](#-使用)
  - [🔧 设置](#-设置)
    - [配置方式](#配置方式)
    - [提供商设置指南](#提供商设置指南)
      - [OpenAI](#openai)
      - [Gemini](#gemini)
      - [Claude/Anthropic](#claudeanthropic)
      - [Vertex](#vertex)
      - [Azure](#azure)
      - [Ollama](#ollama)
      - [其他支持的提供商](#其他支持的提供商)
    - [手动设置提供商](#手动设置提供商)
  - [⌨️ 命令](#️-命令)
  - [⚙ 配置](#-配置)
    - [file\_ignore](#file_ignore)
    - [provider](#provider)
    - [output](#output)
    - [Markdown 主题](#markdown-主题)
    - [支持的语言](#支持的语言)
    - [console](#console)
  - [🔦 支持的键](#-支持的键)
  - [📃 示例](#-示例)
    - [基础用法](#基础用法)
    - [增强错误信息](#增强错误信息)
    - [进度提示](#进度提示)
  - [💻 开发](#-开发)
    - [要求](#要求)
    - [设置](#设置)
    - [运行测试](#运行测试)
      - [Go 测试](#go-测试)
      - [Python 测试](#python-测试)
    - [代码质量](#代码质量)
      - [Go](#go)
      - [Python](#python)
    - [构建](#构建)
  - [📩 联系](#-联系)
  - [☕️ 赞助](#️-赞助)
  - [📜 许可证](#-许可证)

<!-- /TOC -->

## 💡 概览

GPTComet 是一个 AI 驱动的开发者工具，用自动生成提交信息和智能代码审查来改进 Git 工作流。

## ✨ 功能

GPTComet 使用大语言模型自动处理重复工作，并帮助改进开发流程。核心功能包括：

-   **自动生成提交信息**：GPTComet 可以根据代码变更生成 Git 提交信息。
-   **智能代码审查**：获取 AI 生成的代码审查结果、反馈和修改建议。
-   **进度提示**：可选的详细模式会显示长时间操作的实时进度。
-   **多语言支持**：GPTComet 支持英语、中文等多种语言。
-   **可配置**：GPTComet 允许用户自定义配置，例如 LLM 模型和 prompt。
-   **富文本提交信息**：GPTComet 支持包含标题、摘要和详细说明的富文本提交信息。
-   **多提供商支持**：GPTComet 支持 OpenAI、Gemini、Claude/Anthropic、Vertex、Azure、Ollama 等提供商。
-   **支持 SVN 和 Git**：GPTComet 同时支持 SVN 和 Git 仓库。

## ⬇️ 安装

可以使用 Homebrew 安装 GPTComet：

```bash
brew install belingud/tap/gptcomet
```

使用 Homebrew 安装后，请用下面的命令升级：

```bash
brew upgrade gptcomet
```

也可以从 [GitHub release](https://github.com/belingud/gptcomet/releases/latest) 下载，或者使用安装脚本：

```bash
curl -sSL https://cdn.jsdelivr.net/gh/belingud/gptcomet@master/install.sh | bash
```

Windows：

```powershell
irm https://cdn.jsdelivr.net/gh/belingud/gptcomet@master/install.ps1 | iex
```

安装指定版本时，可以使用下面的脚本：

```bash
curl -sSL https://cdn.jsdelivr.net/gh/belingud/gptcomet@master/install.sh | bash -s -- -v 0.4.2
```

```powershell
irm https://cdn.jsdelivr.net/gh/belingud/gptcomet@master/install.ps1 | iex -CommandArgs @("-v", "0.4.2")
```

也可以用 Python 方式安装。PyPI 包已经包含对应平台的二进制文件。

```shell
pip install gptcomet

# Using pipx
pipx install gptcomet

# Using uv
uv tool install gptcomet
Resolved 1 package in 1.33s
Installed 1 package in 8ms
 + gptcomet==0.1.6
Installed 2 executables: gmsg, gptcomet
```

## 📕 使用

使用 GPTComet 的基本步骤如下：

1.  **安装 GPTComet**：通过 Homebrew、安装脚本或 PyPI 安装 GPTComet。
2.  **配置 GPTComet**：参考[设置](#-设置)，配置 `api_key` 以及其他必需配置项，例如：

-   `provider`：语言模型提供商，默认值为 `openai`。
-   `api_base`：API 基础地址，默认值为 `https://api.openai.com/v1`。
-   `api_key`：提供商的 API 密钥。
-   `model`：用于生成提交信息的模型，默认值为 `gpt-4o`。

3.  **运行 GPTComet**：执行 `gmsg commit`。

使用 `openai` 提供商，并且已经设置 `api_key` 后，可以直接运行 `gmsg commit`。

## 🔧 设置

### 配置方式

1. **直接配置**

    - 直接编辑 `~/.config/gptcomet/gptcomet.yaml`。

2. **交互式设置**
    - 使用 `gmsg newprovider` 命令完成引导式设置。

### 提供商设置指南

![Made with VHS](https://vhs.charm.sh/vhs-6019QMIveifvh9vGKc2ZZ8.gif)

```bash
gmsg newprovider

    Select Provider

  > 1. azure
    2. chatglm
    3. claude
    4. cohere
    5. deepseek
    6. gemini
    7. groq
    8. kimi
    9. mistral
    10. ollama
    11. openai
    12. openrouter
    13. sambanova
    14. silicon
    15. tongyi
    16. vertex
    17. xai
    18. Input Manually

    ↑/k up • ↓/j down • ? more
```

#### OpenAI

OpenAI API key 页面：https://platform.openai.com/api-keys

```shell
gmsg newprovider

Selected provider: openai
Configure provider:

Previous inputs:
  Enter OpenAI API base: https://api.openai.com/v1
  Enter API key: sk-abc*********************************************
  Enter max tokens: 1024

Enter Enter model name (default: gpt-4o):
> gpt-4o


Provider openai configured successfully!
```

#### Gemini

Gemini API key 页面：https://aistudio.google.com/u/1/apikey

```shell
gmsg newprovider
Selected provider: gemini
Configure provider:

Previous inputs:
  Enter Gemini API base: https://generativelanguage.googleapis.com/v1beta/models
  Enter API key: AIz************************************
  Enter max tokens: 1024

Enter Enter model name (default: gemini-1.5-flash):
> gemini-2.0-flash-exp

Provider gemini already has a configuration. Do you want to overwrite it? (y/N): y

Provider gemini configured successfully!
```

#### Claude/Anthropic

我还没有 Anthropic 账号，请参考 [Anthropic console](https://console.anthropic.com)。

#### Vertex

Vertex 控制台页面：https://console.cloud.google.com

```shell
gmsg newprovider
Selected provider: vertex
Configure provider:

Previous inputs:
  Enter Vertex AI API Base URL: https://us-central1-aiplatform.googleapis.com/v1
  Enter API key: sk-awz*********************************************
  Enter location (e.g., us-central1): us-central1
  Enter max tokens: 1024
  Enter model name: gemini-1.5-pro

Enter Enter Google Cloud project ID:
> test-project


Provider vertex configured successfully!
```

#### Azure

```shell
gmsg newprovider

Selected provider: azure
Configure provider:

Previous inputs:
  Enter Azure OpenAI endpoint: https://gptcomet.openai.azure.com
  Enter API key: ********************************
  Enter API version: 2024-02-15-preview
  Enter Azure OpenAI deployment name: gpt4o
  Enter max tokens: 1024

Enter Enter deployment name (default: gpt-4o):
> gpt-4o


Provider azure configured successfully!
```

#### Ollama

```shell
gmsg newprovider
Selected provider: ollama
Configure provider:

Previous inputs:
  Enter Ollama API Base URL: http://localhost:11434/api
  Enter max tokens: 1024

Enter Enter model name (default: llama2):
> llama2


Provider ollama configured successfully!
```

#### 其他支持的提供商

-   Groq
-   Mistral
-   Tongyi/Qwen
-   XAI
-   Sambanova
-   Silicon
-   Deepseek
-   ChatGLM
-   KIMI
-   LongCat
-   Cohere
-   OpenRouter
-   Hunyuan
-   ModelScope
-   MiniMax
-   Yi (lingyiwanwu)

暂不支持：

-   Baidu ERNIE

### 手动设置提供商

也可以手动输入提供商名称并配置。

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  ...
  vertex
> Input manually

Enter provider name: test
Enter OpenAI API Base URL [https://api.openai.com/v1]:
Enter model name [gpt-4o]:
Enter API key: ************************************
Enter max tokens [1024]:
[GPTComet] Provider test configured successfully.
```

某些特殊提供商可能需要自定义配置，例如 `cloudflare`。

> 注意：Cloudflare API 不会使用模型名称。

```shell
$ gmsg newprovider

Selected provider: cloudflare
Configure provider:

Previous inputs:
  Enter API Base URL: https://api.cloudflare.com/client/v4/accounts/<account_id>/ai/run
  Enter model name: llama-3.3-70b-instruct-fp8-fast
  Enter API key: abc*************************************

Enter Enter max tokens (default: 1024):
> 1024

Provider cloudflare already has a configuration. Do you want to overwrite it? (y/N): y

Provider cloudflare configured successfully!

$ gmsg config set cloudflare.completion_path @cf/meta/llama-3.3-70b-instruct-fp8-fast
$ gmsg config set cloudflare.answer_path result.response
```

## ⌨️ 命令

GPTComet 提供以下命令：

-   `gmsg config`：配置管理命令组。
    -   `get <key>`：获取配置键的值。
    -   `list`：列出完整配置内容。
    -   `reset`：将配置重置为默认值，也可以通过 `--prompt` 只重置 prompt 部分。
    -   `set <key> <value>`：设置配置值。
    -   `path`：获取配置文件路径。
    -   `remove <key> [value]`：删除配置键，或者从列表中删除某个值。仅用于列表值，例如 `fileignore`。
    -   `append <key> <value>`：向列表配置追加值。仅用于列表值，例如 `fileignore`。
    -   `keys`：列出所有支持的配置键。
-   `gmsg commit`：根据变更或 diff 生成提交信息。
    -   `--svn`：为 SVN 生成提交信息。
    -   `--dry-run`：试运行，不真正生成提交信息。
    -   `-y/--yes`：跳过确认提示。
    -   `--no-verify`：跳过 Git hooks 校验，效果类似 `git commit --no-verify`。
    -   `--repo`：仓库路径，默认值为 `.`。
    -   `--answer-path`：覆盖 answer path。
    -   `--api-base`：覆盖 API base URL。
    -   `--api-key`：覆盖 API key。
    -   `--completion-path`：覆盖 completion path。
    -   `--frequency-penalty`：覆盖 frequency penalty。
    -   `--max-tokens`：覆盖最大 token 数。
    -   `--model`：覆盖模型名称。
    -   `--provider`：覆盖 AI 提供商，例如 `openai` 或 `deepseek`。
    -   `--proxy`：覆盖代理 URL。
    -   `--retries`：覆盖重试次数。
    -   `--temperature`：覆盖 temperature。
    -   `--top-p`：覆盖 top_p。
-   `gmsg newprovider`：添加一个新的提供商。
-   `gmsg review`：审查 staged diff，也可以通过管道输入给 `gmsg review`。
    -   `--svn`：从 SVN 获取 diff。
    -   `--stream`：以流式方式输出 LLM 返回内容。
    -   `--repo`：仓库路径，默认值为 `.`。
    -   `--answer-path`：覆盖 answer path。
    -   `--api-base`：覆盖 API base URL。
    -   `--api-key`：覆盖 API key。
    -   `--completion-path`：覆盖 completion path。
    -   `--frequency-penalty`：覆盖 frequency penalty。
    -   `--max-tokens`：覆盖最大 token 数。
    -   `--model`：覆盖模型名称。
    -   `--provider`：覆盖 AI 提供商，例如 `openai` 或 `deepseek`。
    -   `--proxy`：覆盖代理 URL。
    -   `--retries`：覆盖重试次数。
    -   `--temperature`：覆盖 temperature。
    -   `--top-p`：覆盖 top_p。

全局参数：

```shell
  -c, --config string   Config file path
  -d, --debug           Enable debug mode
```

## ⚙ 配置

主要配置键如下：

| Key                            | Description                                                | Default Value                     |
| :----------------------------- | :--------------------------------------------------------- | :-------------------------------- |
| `provider`                     | 要使用的 LLM 提供商名称。                                 | `openai`                          |
| `file_ignore`                  | 在 diff 中忽略的文件模式列表。                            | 参考 [file_ignore](#file_ignore) |
| `output.lang`                  | 生成提交信息时使用的语言。                                | `en`                              |
| `output.rich_template`         | 富文本提交信息使用的模板。                                | `<title>:<summary>\n\n<detail>`   |
| `output.translate_title`       | 是否翻译提交信息标题。                                    | `false`                           |
| `output.review_lang`           | 生成代码审查信息时使用的语言。                            | `en`                              |
| `output.markdown_theme`        | 显示 markdown 内容时使用的主题。                          | `auto`                            |
| `console.verbose`              | 启用详细输出，包含进度提示和详细错误信息。                | `true`                            |
| `<provider>.api_base`          | 提供商 API 基础地址。                                     | 由提供商决定                      |
| `<provider>.api_key`           | 提供商 API 密钥。                                         |                                   |
| `<provider>.model`             | 要使用的模型名称。                                        | 由提供商决定                      |
| `<provider>.retries`           | API 请求重试次数。                                        | `2`                               |
| `<provider>.proxy`             | 需要时使用的代理 URL。                                    |                                   |
| `<provider>.max_tokens`        | 生成内容的最大 token 数。                                 | `2048`                            |
| `<provider>.top_p`             | 核采样的 top-p 值。                                       | `0.7`                             |
| `<provider>.temperature`       | 控制随机性的 temperature 值。                             | `0.7`                             |
| `<provider>.frequency_penalty` | frequency penalty 值。                                    | `0`                               |
| `<provider>.extra_headers`     | 请求中包含的额外 header，JSON 字符串。                    | `{}`                              |
| `<provider>.extra_body`        | 请求中包含的额外 body，JSON 字符串。                      | `{}`                              |
| `<provider>.completion_path`   | completion 请求的 API 路径。                              | 由提供商决定                      |
| `<provider>.answer_path`       | 从 API 响应中提取答案的 JSON path。                       | 由提供商决定                      |
| `prompt.brief_commit_message`  | 生成简短提交信息的 prompt 模板。                          | 参考 `defaults/defaults.go`       |
| `prompt.rich_commit_message`   | 生成富文本提交信息的 prompt 模板。                        | 参考 `defaults/defaults.go`       |
| `prompt.translation`           | 翻译提交信息的 prompt 模板。                              | 参考 `defaults/defaults.go`       |

**注意：**`<provider>` 应替换为实际提供商名称，例如 `openai`、`gemini`、`claude`。

部分提供商需要特定配置键，例如 Vertex 需要 project ID、location 等。

GPTComet 的配置文件是 `gptcomet.yaml`。

`output.translate_title` 用于决定是否翻译提交信息标题。

例如 `output.lang: zh-cn` 时，提交信息标题是 `feat: Add new feature`。

如果 `output.translate_title` 设置为 `true`，提交信息会翻译成 `功能：新增功能`。
否则会翻译成 `feat: 新增功能`。

某些场景可以把 `complation_path` 设置为空字符串，例如 `<provider>.completion_path: ""`，此时会直接使用 `api_base` 作为端点。

### file_ignore

生成提交信息时要忽略的文件。默认值如下：

```yaml
- bun.lockb
- Cargo.lock
- composer.lock
- Gemfile.lock
- package-lock.json
- pnpm-lock.yaml
- poetry.lock
- yarn.lock
- pdm.lock
- Pipfile.lock
- "*.py[cod]"
- go.sum
- uv.lock
```

可以使用 `gmsg config append file_ignore <xxx>` 添加更多忽略项。
`<xxx>` 使用与 `gitignore` 相同的语法，例如用 `*.so` 忽略所有 `.so` 后缀文件。

### provider

语言模型的提供商配置。

默认提供商是 `openai`。

提供商配置示例：

```yaml
provider: openai
openai:
    api_base: https://api.openai.com/v1
    api_key: YOUR_API_KEY
    model: gpt-4o
    retries: 2
    max_tokens: 1024
    temperature: 0.7
    top_p: 0.7
    frequency_penalty: 0
    extra_headers: {}
    answer_path: choices.0.message.content
    completion_path: /chat/completions
```

使用 `openai` 时，保持 `api_base` 默认值即可。在配置中设置 `api_key`。

使用 OpenAI 兼容接口的提供商时，可以把 provider 设置为 `openai`，然后设置自定义 `api_base`、`api_key` 和 `model`。

例如 OpenRouter 的 API 接口兼容 OpenAI，可以把 provider 设置为 `openai`，把 `api_base` 设置为 `https://openrouter.ai/api/v1`，把 `api_key` 设置为 [keys 页面](https://openrouter.ai/settings/keys)获取到的密钥，并把 `model` 设置为 `meta-llama/llama-3.1-8b-instruct:free` 或其他需要的模型。

```shell
gmsg config set openai.api_base https://openrouter.ai/api/v1
gmsg config set openai.api_key YOUR_API_KEY
gmsg config set openai.model meta-llama/llama-3.1-8b-instruct:free
gmsg config set openai.max_tokens 1024
```

Silicon 提供的接口也与 OpenRouter 类似，因此可以把 provider 设置为 `openai`，并把 `api_base` 设置为 `https://api.siliconflow.cn/v1`。

**注意：max tokens 在不同模型中可能不同，设置过大时接口会返回错误。**

### output

提交信息输出配置。

默认输出配置如下：

```yaml
output:
    lang: en
    rich_template: "<title>:<summary>\n\n<detail>"
    translate_title: false
    review_lang: "en"
    markdown_theme: "auto"
```

可以设置 `rich_template` 来修改富文本提交信息模板，也可以设置 `lang` 来修改提交信息语言。

### Markdown 主题

支持的 markdown 主题：

-   `auto`：自动检测 markdown 主题，默认值。
-   `ascii`：ASCII 风格。
-   `dark`：深色主题。
-   `dracula`：Dracula 主题。
-   `light`：浅色主题。
-   `tokyo-night`：Tokyo Night 主题。
-   `notty`：Notty 风格，不渲染。
-   `pink`：粉色主题。

不设置 `markdown_theme` 时会自动检测 markdown 主题。
使用浅色终端时，markdown 主题会是 `dark`；使用深色终端时，markdown 主题会是 `light`。

GPTComet 使用 glamour 渲染 markdown，可以在 [glamour preview](https://github.com/charmbracelet/glamour/tree/master/styles/gallery#glamour-style-section) 预览主题效果。

### 支持的语言

`output.lang` 和 `output.review_lang` 支持以下语言：

-   `en`：英语
-   `zh-cn`：简体中文
-   `zh-tw`：繁体中文
-   `fr`：法语
-   `vi`：越南语
-   `ja`：日语
-   `ko`：韩语
-   `ru`：俄语
-   `tr`：土耳其语
-   `id`：印度尼西亚语
-   `th`：泰语
-   `de`：德语
-   `es`：西班牙语
-   `pt`：葡萄牙语
-   `it`：意大利语
-   `ar`：阿拉伯语
-   `hi`：印地语
-   `el`：希腊语
-   `pl`：波兰语
-   `nl`：荷兰语
-   `sv`：瑞典语
-   `fi`：芬兰语
-   `hu`：匈牙利语
-   `cs`：捷克语
-   `ro`：罗马尼亚语
-   `bg`：保加利亚语
-   `uk`：乌克兰语
-   `he`：希伯来语
-   `lt`：立陶宛语
-   `la`：拉丁语
-   `ca`：加泰罗尼亚语
-   `sr`：塞尔维亚语
-   `sl`：斯洛文尼亚语
-   `mk`：马其顿语
-   `lv`：拉脱维亚语

### console

控制台输出配置。

默认控制台配置如下：

```yaml
console:
    verbose: true
```

启用 `verbose`（`true`）后，GPTComet 会提供更完整的使用体验：

- **进度提示**：展示生成提交信息和代码审查时的实时进度。
  ```
  [1/2] Fetching git diff...
  ✓ Fetching git diff (0.07s)
  Discovered provider: mistral, model: codestral-latest
  [2/2] Generating message...
  ✓ Generating message (13.24s)
  ```

- **详细操作信息**：显示正在使用的提供商和模型。

- **增强错误信息**：所有错误都包含：
  - 清楚的问题描述
  - 具体修复建议
  - 相关文档链接
  - 用于快速识别的提示符号

禁用 `verbose`（`false`）后，GPTComet 会以静默模式运行，只输出最少内容，适合脚本和自动化流程。

## 🔦 支持的键

可以使用 `gmsg config keys` 查看支持的键。

## 📃 示例

下面是 GPTComet 的使用示例。

### 基础用法

1.  首次使用 `gmsg config set openai.api_key YOUR_API_KEY` 设置 OpenAI KEY 时，会在 `~/.local/gptcomet/gptcomet.yaml` 生成配置文件，内容包括：

```
provider: "openai"
openai:
  api_base: "https://api.openai.com/v1"
  api_key: "YOUR_API_KEY"
  model: "gpt-4o"
  retries: 2
output:
  lang: "en"
```

2.  运行 `gmsg commit` 生成提交信息。
3.  GPTComet 会根据代码变更生成提交信息，并显示在控制台中。

### 增强错误信息

GPTComet 会提供包含可执行建议的错误信息：

```bash
$ gmsg commit

❌ API Key Not Configured

Provider 'openai' requires an API key, but none was found.

What to do:
  • Set API key: gmsg config set openai.api_key <your-key>
  • Or set env var: export OPENAI_API_KEY=<your-key>
  • Check provider: gmsg config get openai

Docs: https://github.com/belingud/gptcomet#configuration
```

### 进度提示

启用 `console.verbose`（默认启用）后，会看到实时进度：

```bash
$ gmsg commit

[1/2] Fetching git diff...(0.07s)
Discovered provider: mistral, model: codestral-latest
[2/2] Generating message...
📤 Sending request to mistral...
Token usage> prompt: 1341, completion: 10, total: 1,351
✓ Generating message (13.24s)

feat: add user authentication feature
```

禁用进度提示并以静默模式运行：

```bash
gmsg config set console.verbose false
```

注意：请将 `YOUR_API_KEY` 替换为对应提供商的真实 API key。

## 💻 开发

### 要求

- Go 1.25+
- Python 3.9+
- just 命令运行器
- pytest，用于 Python 测试

### 设置

欢迎 fork 本项目并提交 pull request。

第一步，fork 项目并克隆你的仓库。

```shell
git clone https://github.com/<yourname>/gptcomet
```

第二步，确认已经安装 `uv`。可以通过 `pip`、`brew` 或其他方式安装，参考它的[安装文档](https://docs.astral.sh/uv/getting-started/installation/)。

使用 `just` 安装依赖：

```shell
just install
```

### 运行测试

#### Go 测试

```bash
# Run all Go tests
go test ./...

# Run specific package tests
go test ./internal/llm/

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Using just
just test              # Run tests with coverage
just test-coverage     # Generate coverage report
just test-cover-func   # Show coverage by function
```

#### Python 测试

```bash
# Run Python wrapper tests
just test-py

# Run with coverage
just test-py-cov

# Or manually with uv
uv run pytest tests/py_tests/ -v
uv run pytest tests/py_tests/ --cov=py/gptcomet --cov-report=html
```

### 代码质量

#### Go

```bash
# Static analysis
go vet ./...
staticcheck ./...

# Using just
just check             # Run go vet and staticcheck
just format            # Format Go code
```

#### Python

```bash
# Code linting
ruff check py/

# Formatting
ruff format py/
```

### 构建

```bash
# Build Go binary
just build

# Build all platforms
just build-all

# Build Python wheel
just build-py
```

## 📩 联系

如有问题或建议，欢迎联系。

## ☕️ 赞助

如果喜欢 GPTComet，可以请我喝杯咖啡支持项目。任何支持都能帮助项目继续前进。

[Buy Me A Coffee](./SPONSOR.md)

## 📜 许可证

GPTComet 使用 MIT License。

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbelingud%2Fgptcomet.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbelingud%2Fgptcomet?ref=badge_large&issueType=license)
