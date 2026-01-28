# LongCat Provider 实施计划

## 概述

本文档描述了为 GPTComet 添加 LongCat 提供商的实施计划。LongCat 是一个 OpenAI 兼容的 API 提供商，支持思考模型（reasoning models），类似于 DeepSeek 的实现。

## 提供商信息

- **提供商名称**: longcat
- **API 地址**: https://api.longcat.chat/openai
- **默认模型**: LongCat-Flash-Chat
- **API 兼容性**: OpenAI 兼容
- **特殊功能**: 支持思考模型（reasoning models）

## 参考实现

基于现有的 DeepSeek 实现作为参考，因为两者都：
1. 使用 OpenAI 兼容的 API 格式
2. 支持思考模型（reasoning models）
3. 使用标准的 BaseLLM 实现

参考文件：
- [`internal/llm/deepseek.go`](internal/llm/deepseek.go) - DeepSeek 实现
- [`internal/llm/openai.go`](internal/llm/openai.go) - OpenAI 标准实现
- [`internal/llm/builder.go`](internal/llm/builder.go) - 配置构建器

## 实施步骤

### 步骤 1: 创建 LongCat LLM 实现

**文件**: `internal/llm/longcat.go`

**任务**:
1. 创建 `LongCatLLM` 结构体，嵌入 `*BaseLLM`
2. 定义常量：
   - `DefaultLongCatAPIBase = "https://api.longcat.chat/openai"`
   - `DefaultLongCatModel = "LongCat-Flash-Chat"`
3. 实现 `NewLongCatLLM()` 构造函数
   - 使用 `BuildStandardConfigSimple()` 设置默认配置
   - 返回 `LongCatLLM` 实例
4. 实现 `Name()` 方法，返回 `"longcat"`
5. 实现 `GetRequiredConfig()` 方法
   - 返回标准配置要求（api_base, api_key, model, max_tokens）
   - 使用适当的提示消息
6. 实现 `MakeRequest()` 方法
   - 调用 `BaseLLM.MakeRequest()` 处理请求

**代码结构**:
```go
package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultLongCatAPIBase = "https://api.longcat.chat/openai"
	DefaultLongCatModel   = "LongCat-Flash-Chat"
)

// LongCatLLM implements the LLM interface for LongCat
type LongCatLLM struct {
	*BaseLLM
}

// NewLongCatLLM creates a new LongCatLLM
func NewLongCatLLM(config *types.ClientConfig) *LongCatLLM {
	BuildStandardConfigSimple(config, DefaultLongCatAPIBase, DefaultLongCatModel)
	return &LongCatLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (l *LongCatLLM) Name() string {
	return "longcat"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (l *LongCatLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultLongCatAPIBase,
			PromptMessage: "Enter LongCat API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultLongCatModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the LongCat API
func (l *LongCatLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return l.BaseLLM.MakeRequest(ctx, client, l, message, stream)
}
```

**预期结果**: 创建约 60 行的 `longcat.go` 文件

---

### 步骤 2: 注册 LongCat 提供商

**文件**: `internal/llm/provider.go`

**任务**:
1. 在 `init()` 函数中添加 LongCat 提供商注册
2. 按字母顺序插入到合适位置（在 Kimi 和 MiniMax 之间）
3. 使用 `RegisterProvider()` 函数注册构造函数

**修改位置**: 在 `init()` 函数中，约第 136-142 行之间

**代码变更**:
```go
// Kimi
RegisterProvider("kimi", func(config *types.ClientConfig) LLM {
	return NewKimiLLM(config)
})

// LongCat
RegisterProvider("longcat", func(config *types.ClientConfig) LLM {
	return NewLongCatLLM(config)
})

// MiniMax
RegisterProvider("minimax", func(config *types.ClientConfig) LLM {
	return NewMinimaxLLM(config)
})
```

**预期结果**: 在 provider.go 中添加 4 行代码

---

### 步骤 3: 创建单元测试

**文件**: `internal/llm/longcat_test.go`

**任务**:
1. 创建测试文件，参考 `deepseek_test.go` 的结构
2. 实现以下测试用例：
   - `TestNewLongCatLLM` - 测试构造函数
   - `TestLongCatLLM_Name` - 测试 Name() 方法
   - `TestLongCatLLM_GetRequiredConfig` - 测试配置要求
   - `TestLongCatLLM_BuildURL` - 测试 URL 构建
   - `TestLongCatLLM_BuildHeaders` - 测试请求头构建
   - `TestLongCatLLM_FormatMessages` - 测试消息格式化
   - `TestLongCatLLM_ParseResponse` - 测试响应解析

**代码结构**:
```go
package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewLongCatLLM(t *testing.T) {
	config := &types.ClientConfig{
		Provider: "longcat",
	}
	llm := NewLongCatLLM(config)
	
	assert.NotNil(t, llm)
	assert.Equal(t, DefaultLongCatAPIBase, config.APIBase)
	assert.Equal(t, DefaultLongCatModel, config.Model)
}

func TestLongCatLLM_Name(t *testing.T) {
	config := &types.ClientConfig{}
	llm := NewLongCatLLM(config)
	
	assert.Equal(t, "longcat", llm.Name())
}

func TestLongCatLLM_GetRequiredConfig(t *testing.T) {
	config := &types.ClientConfig{}
	llm := NewLongCatLLM(config)
	
	requiredConfig := llm.GetRequiredConfig()
	
	assert.Contains(t, requiredConfig, "api_base")
	assert.Contains(t, requiredConfig, "api_key")
	assert.Contains(t, requiredConfig, "model")
	assert.Contains(t, requiredConfig, "max_tokens")
	
	assert.Equal(t, DefaultLongCatAPIBase, requiredConfig["api_base"].DefaultValue)
	assert.Equal(t, DefaultLongCatModel, requiredConfig["model"].DefaultValue)
}

// Additional tests following the pattern from deepseek_test.go
```

**预期结果**: 创建约 150-200 行的测试文件

---

### 步骤 4: 更新文档

**文件**: 
- `README.md`
- `AGENTS.md`

**任务**:

#### 4.1 更新 README.md
1. 在支持的提供商列表中添加 LongCat
2. 更新提供商数量（从 22 个增加到 23 个）
3. 添加 LongCat 配置示例（如果有专门的配置章节）

**修改位置**: 在提供商列表部分，按字母顺序添加

#### 4.2 更新 AGENTS.md
1. 在 "LLM Provider System" 部分更新提供商列表
2. 在 "Supported providers" 列表中添加 longcat
3. 更新提供商数量说明

**修改位置**: 约第 22 行和第 62 行

**预期结果**: 两个文档文件各添加 1-2 行

---

### 步骤 5: 集成测试验证

**任务**:
1. 运行单元测试验证新提供商
2. 运行集成测试确保注册正确
3. 手动测试提供商初始化

**测试命令**:
```bash
# 运行 LongCat 单元测试
go test ./internal/llm/longcat_test.go -v

# 运行所有 LLM 包测试
go test ./internal/llm/... -v

# 运行提供商注册测试
go test ./internal/llm/provider_test.go -v
go test ./internal/llm/registry_test.go -v

# 运行集成测试
go test ./tests/integration/provider_integration_test.go -v

# 运行所有测试
just test
```

**验证检查清单**:
- [ ] 所有单元测试通过
- [ ] 提供商注册测试通过
- [ ] 集成测试通过
- [ ] 无 linting 警告
- [ ] 代码格式符合标准

---

### 步骤 6: 手动功能测试

**任务**:
1. 使用 `gmsg newprovider` 命令测试交互式配置
2. 使用 `gmsg config` 命令测试配置管理
3. 测试实际的 API 调用（如果有可用的 API key）

**测试场景**:

#### 6.1 交互式配置测试
```bash
# 启动交互式提供商配置
gmsg newprovider

# 选择 longcat
# 输入 API key
# 输入模型名称（或使用默认值）
# 验证配置保存成功
```

#### 6.2 配置管理测试
```bash
# 查看 LongCat 配置
gmsg config get provider
gmsg config get api_base
gmsg config get model

# 设置 LongCat 为默认提供商
gmsg config set provider longcat

# 列出所有配置
gmsg config list
```

#### 6.3 实际 API 调用测试（可选）
```bash
# 如果有可用的 API key，测试实际调用
gmsg commit --provider longcat

# 测试代码审查功能
gmsg review --provider longcat
```

**预期结果**: 所有手动测试场景成功执行

---

## 技术细节

### API 兼容性

LongCat 使用 OpenAI 兼容的 API 格式，因此：
1. 使用标准的 `BaseLLM` 实现
2. 消息格式与 OpenAI 相同
3. 响应解析与 OpenAI 相同
4. 支持流式响应（SSE）

### 思考模型支持

类似于 DeepSeek，LongCat 支持思考模型：
1. 模型可能返回带有思考过程的响应
2. `BaseLLM` 已经处理了思考标签的清理
3. 无需额外的特殊处理

### 配置结构

LongCat 使用标准配置结构：
```yaml
provider: longcat
api_base: https://api.longcat.chat/openai
api_key: your-api-key-here
model: LongCat-Flash-Chat
max_tokens: 1024
```

---

## 文件清单

### 新增文件
1. `internal/llm/longcat.go` (~60 行) - LongCat 实现
2. `internal/llm/longcat_test.go` (~150-200 行) - 单元测试

### 修改文件
1. `internal/llm/provider.go` (+4 行) - 注册提供商
2. `README.md` (+1-2 行) - 更新文档
3. `AGENTS.md` (+1-2 行) - 更新文档

### 总计
- 新增代码: ~210-260 行
- 修改代码: ~8-12 行
- 涉及文件: 5 个

---

## 依赖关系

### 无外部依赖
- LongCat 实现不需要新的外部依赖
- 使用现有的 `BaseLLM` 和配置系统
- 遵循现有的提供商模式

### 内部依赖
- `internal/llm/base.go` - BaseLLM 实现
- `internal/llm/builder.go` - 配置构建器
- `internal/llm/registry.go` - 提供商注册表
- `pkg/types` - 类型定义
- `pkg/config` - 配置类型

---

## 测试策略

### 单元测试
- 测试构造函数正确设置默认值
- 测试 Name() 返回正确的提供商名称
- 测试 GetRequiredConfig() 返回正确的配置要求
- 测试 URL 构建
- 测试请求头构建
- 测试消息格式化
- 测试响应解析

### 集成测试
- 验证提供商注册成功
- 验证提供商可以被正确初始化
- 验证提供商在列表中可见
- 验证配置切换功能

### 手动测试
- 交互式配置流程
- 配置管理命令
- 实际 API 调用（如果可用）

---

## 成功标准

实施完成的标准：
- [ ] `longcat.go` 文件创建并实现所有必需方法
- [ ] `longcat_test.go` 文件创建并包含完整测试
- [ ] 提供商在 `provider.go` 中注册
- [ ] 所有单元测试通过
- [ ] 所有集成测试通过
- [ ] 无 linting 警告
- [ ] 代码格式符合项目标准
- [ ] 文档已更新
- [ ] 手动测试验证功能正常

---

## 风险和注意事项

### 潜在风险
1. **API 兼容性**: 虽然声称 OpenAI 兼容，但可能存在细微差异
2. **思考模型行为**: 思考模型的输出格式可能与 DeepSeek 不同
3. **API 限制**: 可能有特定的速率限制或配额限制

### 缓解措施
1. 使用 `BaseLLM` 的标准实现，已经处理了大多数边缘情况
2. 参考 DeepSeek 的实现，它已经处理了思考模型
3. 在文档中说明任何已知的限制

### 回滚策略
如果实施出现问题：
1. 删除 `longcat.go` 和 `longcat_test.go`
2. 从 `provider.go` 中移除注册代码
3. 恢复文档更改
4. 所有更改都是增量的，不影响现有功能

---

## 实施时间线

### 预估时间
- 步骤 1 (创建实现): 30 分钟
- 步骤 2 (注册提供商): 5 分钟
- 步骤 3 (单元测试): 45 分钟
- 步骤 4 (更新文档): 15 分钟
- 步骤 5 (集成测试): 20 分钟
- 步骤 6 (手动测试): 30 分钟

**总计**: 约 2.5 小时

### 建议顺序
1. 先完成步骤 1-3（核心实现和测试）
2. 运行测试确保基本功能正常
3. 完成步骤 4-6（文档和验证）

---

## 后续工作

### 可选增强
1. 添加 LongCat 特定的配置选项（如果需要）
2. 优化思考模型的输出处理
3. 添加更多的错误处理和重试逻辑
4. 创建 LongCat 使用示例和最佳实践文档

### 维护考虑
1. 监控 LongCat API 的更新和变化
2. 根据用户反馈调整默认配置
3. 定期测试 API 兼容性

---

## 参考资料

### 内部参考
- [`internal/llm/deepseek.go`](internal/llm/deepseek.go) - 思考模型参考实现
- [`internal/llm/openai.go`](internal/llm/openai.go) - OpenAI 标准实现
- [`internal/llm/base.go`](internal/llm/base.go) - BaseLLM 基础实现
- [`internal/llm/builder.go`](internal/llm/builder.go) - 配置构建工具
- [`plans/01-provider-and-config-refactoring.md`](plans/01-provider-and-config-refactoring.md) - 重构计划参考

### 项目文档
- [`AGENTS.md`](AGENTS.md) - 项目架构和规范
- [`README.md`](README.md) - 用户文档
- [`CONTRIBUTING.rst`](CONTRIBUTING.rst) - 贡献指南

---

## 附录

### A. 代码风格指南
- 遵循 Go 标准代码风格
- 使用 `gofmt` 和 `goimports` 格式化代码
- 注释使用英文
- 常量使用大写字母和下划线
- 私有方法使用小写字母开头

### B. 测试最佳实践
- 使用表驱动测试（table-driven tests）
- 使用 `testutils` 包中的 mock 对象
- 测试覆盖正常流程和边缘情况
- 使用有意义的测试名称

### C. 提交信息格式
```
feat(llm): add LongCat provider support

- Add LongCat LLM implementation with OpenAI-compatible API
- Register LongCat provider in provider registry
- Add comprehensive unit tests
- Update documentation with LongCat support

Closes #XXX
```
