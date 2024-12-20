import json
from typer.testing import CliRunner

from gptcomet import app as cli
from gptcomet.const import GPTCOMET_PRE


def test_config_get(config_manager):
    runner = CliRunner()

    result = runner.invoke(cli, ["config", "get", "provider"])
    assert result.exit_code == 0
    assert f"provider: "+ json.dumps(config_manager.config["provider"], indent=2) in result.output


def test_config_set(config_manager):
    runner = CliRunner()
    result = runner.invoke(cli, ["config", "set", "provider", "openai"])
    assert result.exit_code == 0
    assert f"{GPTCOMET_PRE} Set provider to openai." in result.output


def test_config_append(config_manager):
    runner = CliRunner()
    result = runner.invoke(cli, ["config", "append", "file_ignore", "README.md"])
    assert result.exit_code == 0
    assert f"{GPTCOMET_PRE} Appended README.md to file_ignore" in result.output


def test_config_remove(config_manager):
    runner = CliRunner()
    result = runner.invoke(cli, ["config", "remove", "file_ignore", "yarn.lock"])
    assert result.exit_code == 0
    assert f"{GPTCOMET_PRE} Removed yarn.lock from file_ignore" in result.output
