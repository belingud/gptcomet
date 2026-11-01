# Python Wrapper — Development Guidelines

`py/gptcomet/` is a **distribution shell**, not an implementation. It locates the correct
platform binary and execs it. All behavior (commands, flags, output, errors) is implemented in Go
— see [../go/index.md](../go/index.md).

## Pre-Development Checklist

1. Read this file — the wrapper contract below is small and easy to break.
2. If you touch lint/test tooling or CI, read [quality-guidelines.md](./quality-guidelines.md).
3. Shared thinking guides: [../guides/index.md](../guides/index.md).

Any request that sounds like "the Python wrapper should do X" is almost always a Go change.
Only platform/arch detection, process launch and PyPI packaging belong here.

## What the Wrapper Owns

`py/gptcomet/__init__.py` exports exactly two functions (`pyproject.toml` entry points:
`gmsg = "gptcomet:main"`, `gptcomet = "gptcomet:main"`):

### `find_gptcomet_binary()`

Maps `(sys.platform, platform.machine())` to a bundled binary and returns its absolute path:

| Platform | Arch | Binary |
| --- | --- | --- |
| `win32` | arm64 / amd64 | `bin/gptcomet_<arch>.exe` |
| `darwin` | arm64 / amd64 | `bin/gptcomet_<arch>_mac` |
| `linux` | arm64 / amd64 | `bin/gptcomet_<arch>_linux` |

- Arch aliases: `arm64`/`aarch64`/`arm` → `arm64`; `x86_64`/`amd64`/`x64`/`i386`/`x86` → `amd64`.
- Unknown platform/arch → `OSError` with the offending value; missing binary →
  `FileNotFoundError` that says to open an issue.
- Binaries are build artifacts dropped in by `just build` (which calls
  `scripts/get_binary_path.py`); only `bin/.gitkeep` is tracked.

### `main()`

```python
subcommand = sys.argv[1] if len(sys.argv) > 1 else ""
if subcommand == "update":
    print("Installed by pypi does not support update command, please update by the way you install gptcomet.")
    sys.exit(1)
binary = find_gptcomet_binary()
if sys.platform == "win32":
    process = subprocess.run([binary, *sys.argv[1:]])
    sys.exit(process.returncode)
os.execvpe(binary, [binary, *sys.argv[1:]], env=os.environ)
```

Contract to preserve:

- **Argument passthrough is verbatim** — no parsing, no rewriting, no default args.
- **POSIX replaces the process** (`os.execvpe`) so signals and exit codes are the binary's; only
  Windows uses `subprocess.run` and forwards `returncode`.
- **`update` is blocked** for PyPI installs with the exact message above and exit code 1.
- **Environment is inherited untouched** (`env=os.environ`); config comes from the Go side.

## Packaging (`pyproject.toml`)

- Build backend `pdm-backend`, `package-dir = "py"`, `is-purelib = false` — the wheel ships
  binaries, not pure Python.
- `requires-python = ">=3.9,<4.0"`, **no runtime dependencies** (stdlib only).
- Wheel platform mapping is in `[tool.pdm.build.wheel-data]`; adding a target arch means adding
  the Go build target (`.goreleaser.yaml` / `just build`) *and* the wheel mapping.
- `source-includes = ["LICENSE", "README.md", "pyproject.toml"]` — README is the top-level one.

## Rules

- Keep the module importable with **zero side effects** (`import gptcomet` must not run a binary).
  `main()` is only called from the console-script entry point or `__main__`.
- No new dependencies; no `subprocess` calls beyond the single launch in `main()`.
- No business logic, no config parsing, no output formatting — that belongs in `cmd/`.
- Public functions carry Google-style docstrings (`Args`/`Returns`/`Raises`); this module does not
  use type hints — match the file's existing style instead of introducing a third convention.
- Any change to binary names or supported platforms must be reflected in `py/README.md`, the
  packaging block in `pyproject.toml`, and `.goreleaser.yaml` in the same commit.
