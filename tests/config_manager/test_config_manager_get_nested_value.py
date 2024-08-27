def test_get_nested_value_dict(config_manager):
    doc = {"a": {"b": {"c": "value"}}}
    keys = "a.b.c"
    assert config_manager.get_nested_value(doc, keys) == "value"


def test_get_nested_value_commented_map(config_manager):
    doc = {"a": {"b": {"c": "value"}}}
    keys = "a.b.c"
    assert config_manager.get_nested_value(doc, keys) == "value"


def test_get_nested_value_not_found(config_manager):
    doc = {"a": {"b": {"c": "value"}}}
    keys = "a.b.d"
    assert config_manager.get_nested_value(doc, keys) is None


def test_get_nested_value_default(config_manager):
    doc = {"a": {"b": {"c": "value"}}}
    keys = "a.b.d"
    default = "default_value"
    assert config_manager.get_nested_value(doc, keys, default) == "default_value"


def test_get_nested_value_keys_list(config_manager):
    doc = {"a": {"b": {"c": "value"}}}
    keys = ["a", "b", "c"]
    assert config_manager.get_nested_value(doc, keys) == "value"


def test_get_nested_value_keys_string(config_manager):
    doc = {"a": {"b": {"c": "value"}}}
    keys = "a.b.c"
    assert config_manager.get_nested_value(doc, keys) == "value"
