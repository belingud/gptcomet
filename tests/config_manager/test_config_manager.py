import pytest

from gptcomet.config_manager import ConfigManager


@pytest.fixture
def mock_config_manager(tmp_path):
    config_file = tmp_path / "test_config.yaml"
    # Create a base configuration file
    config_file.write_text("""
provider: openai
openai:
    api_key: sk-test
    api_base: https://api.openai.com/v1
    model: gpt-3.5-turbo
    max_tokens: 1000
anthropic:
    api_key: sk-ant-test
    api_base: https://api.anthropic.com/v1
    model: claude-2
    max_tokens: 2000
    """)
    return ConfigManager(config_file)


class TestSetCliOverrides:
    def test_set_provider_only(self, mock_config_manager):
        """Test setting provider only"""
        mock_config_manager.set_cli_overrides(provider="anthropic")
        assert mock_config_manager.get("provider") == "anthropic"

    def test_set_api_config_only(self, mock_config_manager):
        """Test setting API configuration only"""
        api_config = {"api_key": "new-key", "model": "gpt-4"}
        mock_config_manager.set_cli_overrides(api_config=api_config)
        assert mock_config_manager.get("openai.api_key") == "new-key"
        assert mock_config_manager.get("openai.model") == "gpt-4"
        # Ensure other configurations remain unchanged
        assert mock_config_manager.get("openai.api_base") == "https://api.openai.com/v1"

    def test_set_both_provider_and_config(self, mock_config_manager):
        """Test setting both provider and API configuration"""
        api_config = {"api_key": "new-ant-key", "model": "claude-3"}
        mock_config_manager.set_cli_overrides(provider="anthropic", api_config=api_config)
        assert mock_config_manager.get("provider") == "anthropic"
        assert mock_config_manager.get("anthropic.api_key") == "new-ant-key"
        assert mock_config_manager.get("anthropic.model") == "claude-3"

    def test_invalid_provider_type(self, mock_config_manager):
        """Test invalid provider type"""
        with pytest.raises(TypeError, match="Provider must be a string"):
            mock_config_manager.set_cli_overrides(provider=123)

    def test_missing_provider(self, mock_config_manager):
        """Test setting API configuration without a provider"""
        mock_config_manager.config["provider"] = None
        with pytest.raises(ValueError, match="Provider is required"):
            mock_config_manager.set_cli_overrides(api_config={"api_key": "test"})

    def test_empty_api_config(self, mock_config_manager):
        """Test empty API configuration"""
        original_config = mock_config_manager.config.copy()
        mock_config_manager.set_cli_overrides(api_config={})
        assert mock_config_manager.config == original_config

    def test_none_values_in_api_config(self, mock_config_manager):
        """Test None values in API configuration are filtered"""
        api_config = {"api_key": "new-key", "model": None, "max_tokens": None}
        mock_config_manager.set_cli_overrides(api_config=api_config)
        assert mock_config_manager.get("openai.api_key") == "new-key"
        assert mock_config_manager.get("openai.model") == "gpt-3.5-turbo"  # Keep original value
        assert mock_config_manager.get("openai.max_tokens") == 1000  # Keep original value

    def test_new_provider_config(self, mock_config_manager):
        """Test setting configuration for a new provider"""
        api_config = {
            "api_key": "new-key",
            "model": "new-model",
            "api_base": "https://api.new-provider.com",
        }
        mock_config_manager.set_cli_overrides(provider="new-provider", api_config=api_config)
        assert mock_config_manager.get("provider") == "new-provider"
        assert mock_config_manager.get("new-provider.api_key") == "new-key"
        assert mock_config_manager.get("new-provider.model") == "new-model"
        assert mock_config_manager.get("new-provider.api_base") == "https://api.new-provider.com"

    def test_partial_update_existing_config(self, mock_config_manager):
        """Test partial update of existing configuration"""
        api_config = {"model": "gpt-4", "temperature": 0.8}
        mock_config_manager.set_cli_overrides(api_config=api_config)
        assert mock_config_manager.get("openai.model") == "gpt-4"
        assert mock_config_manager.get("openai.temperature") == 0.8
        assert mock_config_manager.get("openai.api_key") == "sk-test"  # Keep original value

    def test_config_persistence(self, mock_config_manager):
        """Test configuration overrides don't affect the original config file"""
        original_content = mock_config_manager.current_config_path.read_text()
        mock_config_manager.set_cli_overrides(
            provider="anthropic", api_config={"api_key": "new-key"}
        )
        assert mock_config_manager.current_config_path.read_text() == original_content
