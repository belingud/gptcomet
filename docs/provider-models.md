# Provider Model Defaults

These are the default models used by `gmsg newprovider` and when a provider
configuration omits `model`. The list was last refreshed on **2026-10-26**.

Model names change frequently. When updating one, update the provider constant,
its tests, this table, and the examples in the READMEs.

| Provider   | Default model                              | Notes |
| ---------- | ------------------------------------------ | ----- |
| `ai21`     | `jamba-1.5-large`                          | Kept; no newer public model verified in this refresh. |
| `azure`    | `gpt-6.1-sol`                              | Use your Azure deployment name; this is the default deployment example. |
| `chatglm`  | `glm-5.3-flash`                            | Zhipu GLM-5.3 Flash. |
| `claude`   | `claude-sonnet-5-5`                        | Anthropic Claude Sonnet 5.5. |
| `cohere`   | `command-a-plus`                           | Cohere Command A+. |
| `deepseek` | `deepseek-v4.1-flash`                      | DeepSeek V4.1 Flash. |
| `gemini`   | `gemini-3.8-flash`                         | Google Gemini 3.8 Flash. |
| `groq`     | `openai/gpt-oss-120b`                      | Groq-hosted GPT-OSS 120B. |
| `hunyuan`  | `hunyuan-a13b`                             | Kept on the existing Hunyuan endpoint; `hy3` requires migrating the API base to TokenHub. |
| `kimi`     | `kimi-k3-1-pro`                            | Moonshot Kimi K3.1 Pro. |
| `longcat`  | `LongCat-2.0`                              | Meituan LongCat 2.0. |
| `minimax`  | `MiniMax-M3.1-Flash`                       | MiniMax M3.1 Flash. |
| `mistral`  | `mistral-large-latest`                     | Rolling alias; stays current automatically. |
| `modelscope` | `deepseek-ai/DeepSeek-V4.1-Flash`        | ModelScope-hosted DeepSeek V4.1 Flash. |
| `ollama`   | `qwen3.8:27b`                              | Local Ollama library tag. |
| `openai`   | `gpt-6.1-sol`                              | OpenAI GPT-6.1 Sol. |
| `openrouter` | `openai/gpt-6.1-sol`                     | OpenRouter namespaced model ID. |
| `sambanova` | `Meta-Llama-4-Maverick-17B-128E-Instruct` | SambaNova-hosted Llama 4 Maverick. |
| `silicon`  | `Qwen/Qwen3.8-2.4T-A95B`                   | SiliconFlow model ID. |
| `tongyi`   | `qwen3.8-max`                              | Alibaba Cloud Qwen3.8 Max. |
| `vertex`   | `gemini-3.8-flash`                         | Vertex AI Gemini 3.8 Flash. |
| `xai`      | `grok-4.7`                                 | xAI Grok 4.7. |
| `yi`       | `yi-lightning`                             | Kept; no newer public model verified in this refresh. |

## Sources

- OpenAI: <https://developers.openai.com/api/docs/models/gpt-6.1-sol>
- Anthropic: <https://platform.claude.com/docs/en/models/sonnet-5-5/overview.md>
- Google Gemini: <https://ai.google.dev/gemini-api/docs/models/gemini-3.8-flash>
- DeepSeek: <https://api-docs.deepseek.com/news/news261017/>
- xAI: <https://docs.x.ai/developers/grok-4-7>
- Qwen: <https://www.alibabacloud.com/help/tc/model-studio/qwen3-8-max>
- Kimi: <https://platform.kimi.com/docs/guide/kimi-k3-1-pro-quickstart>
- MiniMax: <https://platform.minimax.io/docs/api-reference/api-overview>
- GLM: <https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3>
- LongCat: <https://longcat.chat/platform/docs/zh/api/chat>
- Cohere: <https://docs.cohere.com/v1/docs/command-a-plus>
- Groq: <https://console.groq.com/docs/model/openai/gpt-oss-120b>
- SiliconFlow: <https://www.siliconflow.com/models/qwen3-8-2-4t-a95b>
- Ollama: <https://registry.ollama.com/library/qwen3.8>
- Tencent Hunyuan: <https://tokenhub.com/zh/providers/tencent-hunyuan>
- OpenRouter: <https://openrouter.ai/openai/gpt-6.1-sol>

## Refresh checklist

1. Verify the current model ID from the provider's official model documentation.
2. Update the `Default*Model` constant in `internal/llm/<provider>.go`.
3. Update `pkg/config/defaults/defaults.go` if the provider is part of the default config.
4. Update provider tests to use the constant rather than a hard-coded model string.
5. Update this table, `README.md`/`README.zh.md` examples, and `CLAUDE.md`.
6. Run `go test ./...` and `go vet ./...`.
