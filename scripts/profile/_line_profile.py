import click
from click.testing import CliRunner
from line_profiler import LineProfiler

from gptcomet.cli import cli, config_set  # 导入你的命令行工具和你想分析的函数


def run_profiling():
    profiler = LineProfiler()

    # add target functions
    profiler.add_function(cli)
    profiler.add_function(config_set)
    # start the profiler
    profiler.enable_by_count()

    runner = CliRunner()
    result = runner.invoke(cli, ["config", "set", "openai.retries", "3"])

    profiler.disable_by_count()

    # 打印结果
    profiler.print_stats()


if __name__ == "__main__":
    run_profiling()
