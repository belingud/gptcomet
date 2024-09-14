from unittest.mock import Mock

import pytest

from gptcomet.clis.commit import gen_output
from gptcomet.const import COMMIT_OUTPUT_TEMPLATE


def test_gen_output(repo, commit):
    output = gen_output(repo, commit)
    assert output.startswith(
        COMMIT_OUTPUT_TEMPLATE.format(
            author=":construction_worker: [green]John Doe[/]",
            email="[blue]john@example.com[/blue]",
            branch="master",
            commit_hash="123456",
            commit_msg="test commit message",
            git_show_stat=repo.git.show(),
        )
    )


def test_gen_output_rich_false(repo, commit):
    output = gen_output(repo, commit, rich=False)
    assert output.startswith(
        COMMIT_OUTPUT_TEMPLATE.format(
            author="John Doe",
            email="john@example.com",
            branch="master",
            commit_hash="123456",
            commit_msg="test commit message",
            git_show_stat=repo.git.show(),
        )
    )


def test_gen_output_commit_msg_empty(repo, commit):
    commit.message = ""
    output = gen_output(repo, commit)
    assert output.startswith(
        COMMIT_OUTPUT_TEMPLATE.format(
            author=":construction_worker: [green]John Doe[/]",
            email="[blue]john@example.com[/blue]",
            branch="master",
            commit_hash="123456",
            commit_msg="",
            git_show_stat=repo.git.show(),
        )
    )


def test_gen_output_author_email_empty(repo, commit):
    commit.author.name = None
    commit.author.email = None
    output = gen_output(repo, commit)
    form = COMMIT_OUTPUT_TEMPLATE.format(
        author=":construction_worker: [green]No Author[/]",
        email="[blue]No Email[/blue]",
        branch="master",
        commit_hash="123456",
        commit_msg="test commit message",
        git_show_stat=repo.git.show(),
    )
    assert output == form


def test_gen_output_repo_none():
    with pytest.raises(AttributeError):
        gen_output(None, Mock())


def test_gen_output_commit_none():
    with pytest.raises(AttributeError):
        gen_output(Mock(), None)
