import os
import sys
from pathlib import Path

import pyinstrument
from click.testing import CliRunner

os.chdir(Path(__file__).resolve().parent.parent.parent)
profiler = pyinstrument.Profiler()
runner = CliRunner()
sys.argv = ["gptcomet", "config", "set", "openai.retries", "n"]
profiler.start()
# ===========================
# entry point of the program
# ===========================
# cmd = ["config", "get", "openai.retries"]
# runner.invoke(cli, cmd, input="n")

# ===========================
# exit program
# ===========================
session = profiler.stop()
# profiler.write_html(Path(__file__).parent / f"pyinstrument_{'-'.join(cmd)}.html", timeline=True)
