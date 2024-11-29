import typing as t

from ruamel.yaml import CommentedMap

# Config class types
CacheKeys = t.Literal["default_config", "config"]


class CacheType(t.TypedDict):
    default_config: t.Optional[CommentedMap]
    config: t.Optional[CommentedMap]


class Message(t.TypedDict):
    """Chat message format"""

    role: t.Literal["user", "assistant", "system"]
    content: str


class ChatUsage(t.TypedDict):
    """API usage information"""

    prompt_tokens: int
    completion_tokens: int
    total_tokens: int


class ChatResponse(t.TypedDict):
    """API response format"""

    choices: list[dict[str, t.Any]]
    usage: t.Optional[ChatUsage]


class CompleteParams(t.TypedDict, total=False):
    """complete required params"""

    api_base: str
    api_key: str
    model: str
    messages: list[Message]
    max_tokens: t.Optional[int]
    temperature: t.Optional[float]
    top_p: t.Optional[float]
    frequency_penalty: t.Optional[float]
    extra_headers: t.Optional[dict[str, str]]


class Provider(t.TypedDict, total=False):
    """provider setting dict"""

    api_base: str
    api_key: str
    model: str
    max_tokens: int
    temperature: t.Optional[float]
    top_p: t.Optional[float]
    frequency_penalty: t.Optional[float]
    presence_penalty: t.Optional[float]
    extra_headers: t.Optional[str]
    proxy: t.Optional[str]
    retries: t.Optional[int]
