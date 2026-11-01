# Python Quality Guidelines

Linting: `ruff`. Tests: `pytest` with `unittest.mock`. Both run against `py/` only — the rest of
the repository is Go.

## Commands

| Need | Command | Notes |
| --- | --- | --- |
| Lint | `uv run ruff check py/` (also `just ci-check`) | CI runs `ruff check py/` |
| Format check | `ruff format --check py/` | CI gate; local fix with `just format-py` |
| Test | `just test-py` (`uv run pytest tests/py_tests/ -v`) | |
| Coverage | `just test-py-cov` | `--cov=py/gptcomet` |

## Ruff Configuration (`pyproject.toml`)

- `target-version = "py39"`, `line-length = 100`, `fix = true`, `src = ["gptcomet"]`.
- `exclude = ["tests/*"]` — **the test suite is not linted**, so imports/mocks there can be loose;
  production code under `py/` has no such excuse.
- Enabled rule families: `YTT` (flake8-2020), `S` (bandit), `B` (bugbear), `A` (builtins),
  `C4` (comprehensions), `T10` (debugger), `SIM` (simplify), `I` (isort), `C90` (mccabe),
  `E`/`W` (pycodestyle), `F` (pyflakes), `PGH` (pygrep-hooks), `UP` (pyupgrade), `RUF`, `TRY`
  (tryceratops). Only `E501` (long lines) and `E731` (lambda assignment) are ignored.
- Consequences you will actually hit: `S603`/`S606` on `subprocess.run`/`os.execvpe` — the
  existing `# noqa: S603` / `# noqa: S606` comments in `main()` are intentional; any new
  subprocess call needs a justified `noqa` and a reviewer's eyes.

## Tests (`tests/py_tests/`)

- `conftest.py` puts `py/` on `sys.path` so the source tree imports without installation — do not
  rely on an installed package.
- Fixtures: `mock_binary_path` (executable temp file), `temp_config_dir`, `mock_home`
  (monkeypatches `HOME`, and `USERPROFILE` on Windows).
- The suite mocks the boundary, not the OS: `patch('gptcomet.find_gptcomet_binary')`,
  `patch('gptcomet.os.execvpe')`, `patch.object(sys, 'platform', 'linux')`,
  `patch.object(sys, 'argv', ['gmsg', '--version'])`.
- Platform-dependent tests `pytest.skip(...)` when the module or bundled binary is unavailable
  (e.g. no `bin/gptcomet_amd64_linux` on macOS). Keep the skip guards — CI runs on Linux where
  the binary may be absent.
- Assertions are plain `assert` / `pytest.raises`; no `assertEqual`-style unittest classes.
- Never spawn the real binary or hit the network in tests.

## Checklist

- [ ] `uv run ruff check py/` clean (no blanket `# noqa: ...` without a rule-specific code).
- [ ] `ruff format --check py/` clean.
- [ ] `just test-py` green.
- [ ] New public function has a docstring and a test in `tests/py_tests/test_wrapper.py`.
- [ ] No new runtime dependency in `pyproject.toml`.
- [ ] If a platform/arch mapping changed: `py/README.md` and `.goreleaser.yaml` updated too.
