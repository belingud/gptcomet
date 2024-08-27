from unittest.mock import MagicMock

import pytest


@pytest.fixture
def config_manager(tmp_path):
    from gptcomet.config_manager import ConfigManager

    # Mock tmp config file for all fixture
    runtime_config_file = tmp_path / "gptcomet.yaml"
    config_manager = ConfigManager(config_path=runtime_config_file)

    # Mock runtime config file path for CliRunner
    ConfigManager.get_config_path = MagicMock(return_value=config_manager.current_config_path)
    with config_manager.default_config_file.open() as f1, runtime_config_file.open("w") as f2:
        f2.write(f1.read())

    yield config_manager
