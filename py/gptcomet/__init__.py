import os
import platform
import subprocess
import sys


def find_gptcomet_binary():
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
        msg = f"gptcomet binary not found for {platform_name}-{arch}, please open an issue on github, thanks."
        raise FileNotFoundError(msg)
    return binary_path


def main():
    binary = find_gptcomet_binary()
    args = [binary] + sys.argv[1:]
    if sys.platform == "win32":
        # no need shell=True for windows
        process = subprocess.run(args)  # noqa: S603
        sys.exit(process.returncode)
    else:
        os.execvpe(binary, [binary, *sys.argv[1:]], env=os.environ)


if __name__ == "__main__":
    main()
