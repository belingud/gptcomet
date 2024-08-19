import os
import sys
from pathlib import Path

import pyinstrument

from gptcomet.cli import cli

os.chdir(Path(__file__).resolve().parent.parent.parent)
profiler = pyinstrument.Profiler()
sys.argv = ["gptcomet", "generate", "commit"]
profiler.start()
# ===========================
# entry point of the program
# ===========================

cli()

# ===========================
# exit program
# ===========================
session = profiler.stop()
profiler.write_html("pyinstrument.html", timeline=True)
