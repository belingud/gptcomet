# from click.testing import CliRunner
# from line_profiler import LineProfiler

# from gptcomet import app as cli


# def run_profiling():
#     profiler = LineProfiler()

#     # add target functions
#     profiler.add_function(cli)
#     profiler.add_function(config_set)
#     # start the profiler
#     profiler.enable_by_count()

#     runner = CliRunner()
#     runner.invoke(cli, ["config", "set", "openai.retries", "3"])

#     profiler.disable_by_count()

#     # Print results
#     profiler.print_stats()


# if __name__ == "__main__":
#     run_profiling()
