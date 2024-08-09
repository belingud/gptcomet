import typing as t
from pathlib import Path

import tomlkit as toml
from tomlkit.items import Comment, SingleKey, Table, Whitespace

config_file = Path(__file__).resolve().parent.parent / "aicommit" / "aicommit.toml"

print(f"Reading default config file: {config_file}")
config = toml.load(config_file.open())

print("Generating support keys")

keys = []
lines: list[list[str]] = []
max_key_and_comment_length = 0
max_key_length = 0
_comment = None

for k, v in config.body:
    k: t.Optional[SingleKey]
    v: t.Any
    # Top level comments
    if isinstance(v, Comment):
        _comment = str(v)
        continue
    if k:
        full_key = k.key.replace("openai", "{provider}")
        if isinstance(v, Table):
            # Deal with nested tables, generate keys like `prompt.brief_commit_message`
            for subk, subv in v.value.body:
                # Also sub table keys allow comments
                if isinstance(subv, Comment):
                    _comment = str(subv)
                    continue
                if subk is None and isinstance(subv, Whitespace):
                    continue
                lines.append([f"{full_key}.{subk.key}", _comment])
                if _comment and len(f"{full_key}.{subk.key}") > max_key_and_comment_length:
                    max_key_and_comment_length = len(f"{full_key}.{subk}")
                _comment = None
        else:
            lines.append([full_key, _comment])
            if _comment and len(full_key) > max_key_and_comment_length:
                max_key_and_comment_length = len(full_key)
            _comment = None
    # ignore no comment lines
    if _comment and lines and len(''.join(lines[-1])) > max_key_and_comment_length:
        max_key_and_comment_length = len(lines[-1])

code = '''SUPPORT_KEYS: str = """\\\n'''

for line in lines:
    c = "" if line[1] is None else " " * (max_key_and_comment_length - len(line[0]) + 1) + line[1]
    key_comment = f"{line[0]}{c}\n"
    code += key_comment

code += '"""'

support_keys_file = Path(__file__).resolve().parent.parent / "aicommit" / "support_keys.py"
with open(support_keys_file, "w") as f:
    f.write(code)
    f.write("\n")

print(f"Saved SUPPORT_KEYS to {support_keys_file}")
