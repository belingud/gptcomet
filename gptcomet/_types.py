import typing as t

from ruamel.yaml import CommentedMap

# Config class types
CacheKeys = t.Literal["default_config", "config"]


class CacheType(t.TypedDict):
    default_config: t.Optional[CommentedMap]
    config: t.Optional[t.Any]


class CompleteParams(t.TypedDict, total=False):
    """complete required params"""

    api_base: str
    api_key: str
    model: str
    messages: t.Union[list[dict], None, list]
    max_tokens: t.Optional[int]
    temperature: t.Optional[float]
    top_p: t.Optional[float]
    frequency_penalty: t.Optional[float]
    extra_headers: t.Optional[dict]


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
