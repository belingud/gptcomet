import os
from dataclasses import dataclass
from typing import Any

# Change working dir to tests dir
os.chdir(os.path.dirname(os.path.abspath(__file__)))


@dataclass
class ActiveBranch:
    name: str


@dataclass
class MockGit:
    name: str
    show: Any


@dataclass
class Author:
    name: str
    email: str
    conf_name: str = "name"
    conf_email: str = "email"


@dataclass
class MockRepo:
    active_branch: ActiveBranch
    git: MockGit
