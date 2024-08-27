from pathlib import Path
from unittest.mock import PropertyMock, patch

import pytest
import yaml

from gptcomet.config_manager import ConfigManager


@pytest.fixture
def mock_config_path(tmp_path):
    return tmp_path / "config.yaml"


@pytest.fixture
def mock_config():
    return {"key": "value"}


def test_save_config_success(mock_config_path, mock_config):
    with (
        patch.object(
            ConfigManager,
            "current_config_path",
            new_callable=PropertyMock,
        ) as mock_current_config_path,
        patch.object(ConfigManager, "config", new_callable=PropertyMock) as mock_config_property,
    ):
        mock_config_property.return_value = mock_config
        mock_current_config_path.return_value = mock_config_path
        config_manager = ConfigManager(config_path=mock_config_path)
        config_manager.save_config()
        assert mock_config_path.exists()
        with mock_config_path.open("r") as f:
            assert yaml.safe_load(f) == mock_config


def test_save_config_fail_path_not_exists(mock_config):
    with (
        patch.object(
            ConfigManager,
            "current_config_path",
            new_callable=PropertyMock,
        ) as mock_current_config_path,
        patch.object(ConfigManager, "config", new_callable=PropertyMock) as mock_config_property,
    ):
        mock_config_property.return_value = mock_config
        mock_current_config_path.return_value = Path("/non/existent/path")
        config_manager = ConfigManager(config_path=Path("/non/existent/path"))
        with pytest.raises(FileNotFoundError):
            config_manager.save_config()


def test_save_config_new_file(mock_config_path, mock_config):
    with (
        patch.object(
            ConfigManager,
            "current_config_path",
            new_callable=PropertyMock,
        ) as mock_current_config_path,
        patch.object(ConfigManager, "config", new_callable=PropertyMock) as mock_config_property,
        patch("builtins.open", side_effect=IOError),
    ):
        mock_config_property.return_value = mock_config
        mock_current_config_path.return_value = mock_config_path
        assert not mock_config_path.exists()
        config_manager = ConfigManager(config_path=mock_config_path)
        config_manager.save_config()
        assert mock_config_path.exists()


def test_save_config_permission_error(mock_config_path, mock_config):
    with (
        patch.object(
            ConfigManager,
            "current_config_path",
            new_callable=PropertyMock,
        ) as mock_current_config_path,
        patch.object(ConfigManager, "config", new_callable=PropertyMock) as mock_config_property,
        patch("io.open", side_effect=PermissionError),
    ):
        mock_config_property.return_value = mock_config
        mock_current_config_path.return_value = mock_config_path
        assert not mock_config_path.exists()
        config_manager = ConfigManager(config_path=mock_config_path)
        with pytest.raises(PermissionError):
            config_manager.save_config()
