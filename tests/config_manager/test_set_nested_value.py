def test_set_nested_value_single_key(config_manager):
    doc = {"a": 0}
    config_manager.set_nested_value(doc, "a", 1)
    assert doc == {"a": 1}


def test_set_nested_value_multiple_keys(config_manager):
    doc = {"a": {"b": {"c": 0}}}
    config_manager.set_nested_value(doc, ["a", "b", "c"], 1)
    assert doc == {"a": {"b": {"c": 1}}}


def test_set_nested_value_create_new_dict(config_manager):
    doc = {"a": {"b": {}}}
    config_manager.set_nested_value(doc, ["a", "b", "c"], 1)
    assert doc == {"a": {"b": {"c": 1}}}
