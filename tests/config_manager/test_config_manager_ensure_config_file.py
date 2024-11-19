def test_ensure_configfile_not_exist(config_manager, tmp_path):
    config_manager.current_config_path = tmp_path / "not_exist" / "gptcomet.yaml"
    assert not config_manager.current_config_path.exists()

    config_manager.ensure_config_file()
    assert config_manager.current_config_path.exists()


def test_ensure_configfile_exist(config_manager):
    with config_manager.current_config_path.open() as f:
        before_content = f.read()
    config_manager.ensure_config_file()
    with config_manager.current_config_path.open() as f:
        after_content = f.read()
    assert before_content == after_content
