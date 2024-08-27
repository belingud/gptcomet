import typing as t
from pathlib import Path

from ruamel.yaml import YAML, CommentedMap
from ruamel.yaml.comments import CommentedSeq

yaml = YAML()

config_file = Path(__file__).resolve().parent.parent / "gptcomet" / "gptcomet.yaml"

print(f"Reading default config file: {config_file}")
with config_file.open() as f:
    config: CommentedMap = yaml.load(f.read())

print("Generating support keys")

keys = []
# lines: list[tuple[str, str]] = []
max_key_and_comment_length = 0
max_key_length = 0
_comment = None


def gen_commented_key(_config: t.Union[CommentedMap, CommentedSeq], _prefix: str):
    keys_and_comments = []
    cas = _config.ca.items
    for k in _config:
        v: t.Union[CommentedMap, CommentedSeq] = _config[k]
        comment = cas.get(k, "")
        if comment != "" and comment[2] is not None:
            comment = comment[2].value.strip()
        full_key = k.replace("openai", "{provider}")
        if _prefix:
            full_key = f"{_prefix}.{full_key}"
        if not isinstance(v, dict):
            keys_and_comments.append((full_key, comment))
        global max_key_and_comment_length
        if comment and len(full_key) > max_key_and_comment_length:
            max_key_and_comment_length = len(full_key)
        if isinstance(v, dict):
            keys_and_comments.extend(gen_commented_key(v, full_key))
        # if comment and keys_and_comments and len(''.join(keys_and_comments[-1])) > max_key_and_comment_length:
        #     max_key_and_comment_length = len(''.join(keys_and_comments[-1]))
    return keys_and_comments


lines = gen_commented_key(config, "")
# for k, v in config:
#     k: t.Optional
#     v: t.Any
#     # Top level comments
#     if isinstance(v, Comment):
#         _comment = str(v)
#         continue
#     if k:
#         full_key = k.key.replace("openai", "{provider}")
#         if isinstance(v, Table):
#             # Deal with nested tables, generate keys like `prompt.brief_commit_message`
#             for subk, subv in v.value.body:
#                 # Also sub table keys allow comments
#                 if isinstance(subv, Comment):
#                     _comment = str(subv)
#                     continue
#                 if subk is None and isinstance(subv, Whitespace):
#                     continue
#                 lines.append([f"{full_key}.{subk.key}", _comment])
#                 if _comment and len(f"{full_key}.{subk.key}") > max_key_and_comment_length:
#                     max_key_and_comment_length = len(f"{full_key}.{subk}")
#                 _comment = None
#         else:
#             lines.append([full_key, _comment])
#             if _comment and len(full_key) > max_key_and_comment_length:
#                 max_key_and_comment_length = len(full_key)
#             _comment = None
#     # ignore no comment lines
#     if _comment and lines and len(''.join(lines[-1])) > max_key_and_comment_length:
#         max_key_and_comment_length = len(lines[-1])

max_key_and_comment_length += 2
code = '''SUPPORT_KEYS: str = """\\\n'''

for line in lines:
    c = "" if not line[1] else " " * (max_key_and_comment_length - len(line[0]) + 1) + line[1]
    key_comment = f"{line[0]}{c}\n"
    code += key_comment

code += '"""'

support_keys_file = Path(__file__).resolve().parent.parent / "gptcomet" / "support_keys.py"
with open(support_keys_file, "w") as f:
    f.write(code)
    f.write("\n")

print(f"Saved SUPPORT_KEYS to {support_keys_file}")
