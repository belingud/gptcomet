"""
This script is used to update CHANGELOG.md file.

It uses `git cliff` to generate a changelog from git commit messages.
"""

import io
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

root_dir = Path(__file__).parent.parent
latest_tag = (
    subprocess.check_output(["git", "describe", "--tags", "--abbrev=0"]).decode("utf-8").strip()  # noqa: S607, S603
)
if not latest_tag:
    raise ValueError("latest tag not found")  # noqa: TRY003

print(f"Latest tag: {latest_tag}")


def get_changelog_content() -> str:
    os.chdir(root_dir)
    p = subprocess.run(["git", "cliff", "--tag", latest_tag], stdout=subprocess.PIPE)  # noqa: S607, S603
    return p.stdout.decode("utf-8")


def format_changelog(changelog_content: str) -> str:
    return changelog_content


def write_changelog(changelog_content: str) -> None:
    """
    Write changelog content to CHANGELOG.md just after second line.

    Args:
        changelog_content (str): generated changelog content.

    Returns:
        None
    """
    latest_version_num = latest_tag[1:]
    tmp = tempfile.NamedTemporaryFile(delete=True)
    cc_container = io.StringIO(changelog_content)
    print(f"Start writing changelog of {latest_tag} to {root_dir / 'CHANGELOG.md'}")
    with open(root_dir / "CHANGELOG.md", "r+") as f, open(tmp.name, "r+") as tmp_f:
        # header -> # Changelog
        # note -> All notable changes to this project will xxx
        tmp_f.write(f.readline())
        tmp_f.write(f.readline())
        tmp_f.write("\n")

        while True:
            cc_content = cc_container.read(1024 * 1024)
            if not cc_content:
                break
            tmp_f.write(cc_content)

        while True:
            origin_content: str = f.read(1024 * 1024)
            if not origin_content:
                break
            if latest_version_num in origin_content:
                # truncate all changes
                raise ValueError(  # noqa: TRY003
                    f"Latest tag {latest_tag} already exists in CHANGELOG.md, aborting."
                )
            tmp_f.write(origin_content)
        f.seek(0)
        f.truncate(0)
        tmp_f.seek(0)
        shutil.copyfileobj(tmp_f, f)
    print("Done")


def main() -> None:
    changelog_content = get_changelog_content()
    changelog_content = format_changelog(changelog_content)
    try:
        write_changelog(changelog_content)
    except ValueError as e:
        print(e)


if __name__ == "__main__":
    main()
