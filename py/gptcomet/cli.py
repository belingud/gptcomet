"""
Command-line interface for gptcomet
"""
# ruff: noqa

import os
import platform
import subprocess
import sys
from pathlib import Path


def get_binary_name():
    """Get the binary name based on the platform"""
    system = platform.system().lower()
    machine = platform.machine().lower()

    # Map architecture names
    arch_map = {
        "x86_64": "amd64",
        "amd64": "amd64",
        "arm64": "arm64",
        "aarch64": "arm64",
    }

    arch = arch_map.get(machine, machine)

    if system == "windows":
        return f"gptcomet_{system}_{arch}.exe"
    else:
        return f"gptcomet_{system}_{arch}"

def find_binary():
    """Find the gptcomet binary"""
    # First, check if gptcomet is in PATH
    try:
        if platform.system().lower() == "windows":
            result = subprocess.run(["where", "gptcomet"], capture_output=True, text=True)
        else:
            result = subprocess.run(["which", "gptcomet"], capture_output=True, text=True)
        if result.returncode == 0:
            return result.stdout.strip()
    except Exception:
        pass

    # If not found in PATH, check if binary is in the same directory as this script
    package_dir = Path(__file__).parent
    binary_name = get_binary_name()
    binary_path = package_dir / binary_name

    if binary_path.exists():
        return str(binary_path)

    raise FileNotFoundError(f"Could not find gptcomet binary: {binary_name}")

def main():
    """Main entry point for the CLI"""
    try:
        binary_path = find_binary()
        # Execute the binary with all arguments
        os.execv(binary_path, ["gptcomet"] + sys.argv[1:])
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error executing gptcomet: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
