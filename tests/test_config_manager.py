from pathlib import Path
from unittest.mock import MagicMock

import pytest
from tomlkit import TOMLDocument, document, key

from gptcomet.config_manager import ConfigManager


class MockConfigManager(MagicMock, ConfigManager):
    def __init__(self):
        super().__init__()
        default_doc = document()
        default_doc.add("test_key", "test_value")
        default_doc.append(key("default_key"), "default_value")

        test_doc = document()
        test_doc.append(key("test_key"), "test_value")
        self._cache = {
            "config": test_doc,
            "default_config": default_doc,
        }


@pytest.fixture
def config_manager():
    return MockConfigManager()


def test_config_property(config_manager):
    config_manager._cache["config"] = TOMLDocument()
    assert config_manager.config == config_manager._cache["config"]


def test_default_config_property(config_manager):
    config_manager._cache["default_config"] = TOMLDocument()
    assert config_manager.default_config == config_manager._cache["default_config"]


def test_get_config_value(config_manager):
    # config_manager.load_config = MagicMock(return_value=TOMLDocument({"test_key": "test_value"}))
    config_manager.is_valid_key = MagicMock(return_value=True)
    assert config_manager.get("test_key") == "test_value"


def test_save_config(config_manager):
    config_manager.current_config_path = Path("test_config.toml")
    config_manager.is_valid_key = MagicMock(return_value=True)
    config_manager.save_config()
    assert Path("test_config.toml").exists()
    Path("test_config.toml").unlink()
