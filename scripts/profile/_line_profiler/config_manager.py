import tempfile
from pathlib import Path

from line_profiler import LineProfiler

from gptcomet.config_manager import ConfigManager


def config_set():
    profiler = LineProfiler()
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir) / "getcomment.toml"
        cfg_manager = ConfigManager(config_path=config_path)
        profiler.add_function(cfg_manager.save_config)
        profiler.enable_by_count()
        cfg_manager.set("openai.retries", "2")
        profiler.disable_by_count()
        print("=============config set line profiler==============")
        profiler.print_stats()
        print("=============config set line profiler==============")


def config_keys():
    profiler = LineProfiler()
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir) / "getcomment.toml"
        cfg_manager = ConfigManager(config_path=config_path)
        profiler.add_function(cfg_manager.list_keys)
        profiler.enable_by_count()
        keys = cfg_manager.list_keys()
        profiler.disable_by_count()
        print("=============config keys line profiler==============")
        profiler.print_stats()
        print("=============config keys line profiler==============")


def config_get():
    profiler = LineProfiler()
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir) / "getcomment.toml"
        cfg_manager = ConfigManager(config_path=config_path)
        profiler.add_function(cfg_manager.get)
        profiler.enable_by_count()
        value = cfg_manager.get("openai.retries")
        profiler.disable_by_count()
        print("=============config get line profiler==============")
        profiler.print_stats()
        print("=============config get line profiler==============")


def config_list():
    profiler = LineProfiler()
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir) / "getcomment.toml"
        cfg_manager = ConfigManager(config_path=config_path)
        profiler.add_function(cfg_manager.list)
        profiler.enable_by_count()
        cfg_manager.ensure_config_file()
        value = cfg_manager.list()
        profiler.disable_by_count()
        print("=============config list line profiler==============")
        profiler.print_stats()
        print("=============config list line profiler==============")


def config_reset():
    profiler = LineProfiler()
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir) / "getcomment.toml"
        cfg_manager = ConfigManager(config_path=config_path)
        profiler.add_function(cfg_manager.reset)
        profiler.enable_by_count()
        cfg_manager.reset()
        profiler.disable_by_count()
        print("=============config reset line profiler==============")
        profiler.print_stats()
        print("=============config reset line profiler==============")


def config_load_config():
    profiler = LineProfiler()
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir) / "getcomment.toml"
        cfg_manager = ConfigManager(config_path=config_path)
        profiler.add_function(cfg_manager.load_config)
        profiler.enable_by_count()
        value = cfg_manager.load_config()
        profiler.disable_by_count()
        print("=============config path line profiler==============")
        profiler.print_stats()
        print("=============config path line profiler==============")


def main():
    config_set()
    config_keys()
    config_get()
    config_list()
    config_reset()
    config_load_config()


if __name__ == "__main__":
    main()
