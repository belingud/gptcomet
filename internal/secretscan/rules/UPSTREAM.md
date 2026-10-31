# Upstream — gitleaks rule data

The files in this directory are copied from [gitleaks](https://github.com/gitleaks/gitleaks),
which is MIT-licensed (see `LICENSE`).

| Item | Value |
| --- | --- |
| Upstream project | https://github.com/gitleaks/gitleaks |
| Upstream version | v8.30.1 |
| Upstream commit | 83d9cd684c87d95d656c1458ef04895a7f1cbd8e |
| Source file | `config/gitleaks.toml` |
| Vendored on | 2026-10-31 |
| TOML SHA-256 | `e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf` (97,731 bytes) |
| Rules | 222 + one global allowlist |

`internal/secretscan/rules_gen.go` is generated from `gitleaks.toml` by
`scripts/gen_secret_rules.py` and embeds the rule data as Go structures, so the runtime does not
need a TOML parser. The scanner implements the subset of the gitleaks matcher that the default
rule set actually uses (regex + `secretGroup`, keyword prefilter, Shannon entropy, path rules,
rule/global allowlists). It does **not** implement base64/hex decode passes, archive or directory
scanning, git history scanning, `.gitleaksignore`/baselines, or `required` proximity rules.

## Updating the rules

1. Download `config/gitleaks.toml` from the target upstream release tag.
2. Copy it over `gitleaks.toml`; copy the upstream `LICENSE` if it changed.
3. Update the table above (version, commit, SHA-256, date).
4. Run:

   ```bash
   python3 scripts/gen_secret_rules.py \
       --version vX.Y.Z --commit <upstream-commit> [--toml <path>]
   gofmt -w internal/secretscan/rules_gen.go
   ```

   or simply `just sync-secret-rules` after updating the script defaults.
5. Run `go test ./internal/secretscan/` — the tests assert every generated regex compiles and the
   rule count is non-zero.
