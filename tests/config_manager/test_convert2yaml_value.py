def test_convert2yaml_value_bool(config_manager):
    assert config_manager.convert2yaml_value("true") is True
    assert config_manager.convert2yaml_value("false") is False
    assert config_manager.convert2yaml_value("no") is False
    assert config_manager.convert2yaml_value("0") is False
    assert config_manager.convert2yaml_value("1") is True
    assert config_manager.convert2yaml_value("yes") is True


def test_convert2yaml_value_none(config_manager):
    assert config_manager.convert2yaml_value("none") is None
    assert config_manager.convert2yaml_value("null") is None


def test_convert2yaml_value_float(config_manager):
    assert config_manager.convert2yaml_value("3.14") == 3.14


def test_convert2yaml_value_int(config_manager):
    assert config_manager.convert2yaml_value("42") == 42


def test_convert2yaml_value_str(config_manager):
    assert config_manager.convert2yaml_value("hello") == "hello"


# def test_convert2yaml_value_list(config_manager):
#     assert config_manager.convert2yaml_value("[1, 2, 3]") == [1, 2, 3]
