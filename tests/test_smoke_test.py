import pytest
from typer.testing import CliRunner

from gptcomet import app

runner = CliRunner()

@pytest.mark.parametrize("command", [
    "config get -h",
    "config set -h",
    "config path -h",
    "config keys -h",
    "config list -h",
    "config reset -h",
    "gen commit -h"
])
def test_gmsg_commands(command):
    result = runner.invoke(app, command.split())
    assert result.exit_code == 0
    assert "Error" not in result.output
