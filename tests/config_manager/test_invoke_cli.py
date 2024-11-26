from typer.testing import CliRunner

from gptcomet import app as cli
from gptcomet.const import GPTCOMET_PRE


def test_config_get(config_manager):
    runner = CliRunner()

    result = runner.invoke(cli, ["config", "get", "provider"])
    assert result.exit_code == 0
    assert result.output.strip() == "provider: " + config_manager.config["provider"]


def test_config_set(config_manager):
    runner = CliRunner()
    result = runner.invoke(cli, ["config", "set", "provider", "openai"])
    assert result.exit_code == 0
    assert result.output.strip() == f"{GPTCOMET_PRE} Set provider to openai."


def test_config_append(config_manager):
    runner = CliRunner()
    result = runner.invoke(cli, ["config", "append", "file_ignore", "README.md"])
    assert result.exit_code == 0
    assert result.output.strip() == f"\x1b[32m{GPTCOMET_PRE} Appended README.md to file_ignore.\x1b[0m"


# def test_config_remove_value_error(config_manager):
#     runner = CliRunner()
#     result = runner.invoke(cli, ["config", "remove", "file_ignore", "README.md"])
#     assert result.output.strip() == f"{GPTCOMET_PRE} value not found: README.md"
#     assert result.exit_code == 0


def test_config_remove(config_manager):
    runner = CliRunner()
    result = runner.invoke(cli, ["config", "remove", "file_ignore", "yarn.lock"])
    assert result.exit_code == 0
    assert result.output.strip() == f"\x1b[32m{GPTCOMET_PRE} Removed yarn.lock from file_ignore\x1b[0m"
