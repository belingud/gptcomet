import typing as t

from ruamel.yaml import CommentedMap

# Config class types
CacheKeys = t.Literal["default_config", "config"]


class CacheType(t.TypedDict):
    default_config: t.Optional[CommentedMap]
    config: t.Optional[CommentedMap]


class CompleteParams(t.TypedDict):
    api_base: str
    api_key: str
    model: str
    retries: int
    messages: t.Union[list[dict], None, list]
    max_tokens: t.Optional[int]
    temperature: t.Optional[float]
    top_p: t.Optional[float]
    frequency_penalty: t.Optional[float]
    extra_headers: t.Optional[dict]
