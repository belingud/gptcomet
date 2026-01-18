import os
import platform
import subprocess
import sys


def find_gptcomet_binary():
    """Find the appropriate gptcomet binary for the current platform and architecture.

    This function determines the correct binary name based on the system platform
    (Windows, macOS, Linux) and architecture (ARM64, AMD64), then locates the binary
    in the package's bin directory.

    Returns:
        str: The full path to the gptcomet binary

    Raises:
        OSError: If the platform or architecture is not supported
        FileNotFoundError: If the binary file is not found in the expected location

    Examples:
        >>> find_gptcomet_binary()
        '/path/to/gptcomet/bin/gptcomet_amd64_mac'
    """
    platform_name = sys.platform

    machine = platform.machine().lower()
    if machine in ("arm64", "aarch64", "arm"):
        arch = "arm64"
    elif machine in ("x86_64", "amd64", "x64", "i386", "x86"):
        arch = "amd64"
    else:
        msg = f"Unsupported architecture: {machine}."
        raise OSError(msg)

    if platform_name == "win32":
        binary_name = f"gptcomet_{arch}.exe"
    elif platform_name == "darwin":
        binary_name = f"gptcomet_{arch}_mac"
    elif platform_name == "linux":
        binary_name = f"gptcomet_{arch}_linux"
    else:
        msg = f"Unsupported platform: {platform_name}"
        raise OSError(msg)

    binary_path = os.path.join(os.path.dirname(__file__), "bin", binary_name)
    if not os.path.isfile(binary_path):
        msg = f"gptcomet binary ({binary_name}) not found, please open an issue on github, thanks."
        raise FileNotFoundError(msg)
    return binary_path


def main():
    """Launch and execute the gptcomet binary with provided arguments.

    The function behaves differently on Windows vs other platforms:
    - On Windows: Uses subprocess.run() to execute the binary
    - On other platforms: Uses os.execvpe() for direct process replacement

    Raises:
        SystemExit: With code 1 if update command is used, otherwise with binary's return code
    """
    subcommand = sys.argv[1] if len(sys.argv) > 1 else ""
    if subcommand == "update":
        print(
            "Installed by pypi does not support update command, please update by the way you install gptcomet."
        )
        sys.exit(1)
    binary = find_gptcomet_binary()
    args = [binary, *sys.argv[1:]]
    if sys.platform == "win32":
        # no need shell=True for windows
        process = subprocess.run(args)  # noqa: S603
        sys.exit(process.returncode)
    else:
        os.execvpe(binary, [binary, *sys.argv[1:]], env=os.environ)  # noqa: S606


if __name__ == "__main__":
    main()
