import pytest
import tempfile
import os
import sys


@pytest.fixture
def mock_binary_path():
    """提供模拟的二进制路径"""
    with tempfile.NamedTemporaryFile(delete=False, suffix='.exe' if os.name == 'nt' else '') as f:
        temp_path = f.name
        # Make it executable
        os.chmod(temp_path, 0o755)
    yield temp_path
    # Cleanup
    if os.path.exists(temp_path):
        os.unlink(temp_path)


@pytest.fixture
def temp_config_dir():
    """提供临时配置目录"""
    temp_dir = tempfile.mkdtemp()
    yield temp_dir
    # Cleanup
    import shutil
    if os.path.exists(temp_dir):
        shutil.rmtree(temp_dir)


@pytest.fixture
def mock_home(monkeypatch, temp_config_dir):
    """模拟HOME目录"""
    monkeypatch.setenv('HOME', temp_config_dir)
    if sys.platform == 'win32':
        monkeypatch.setenv('USERPROFILE', temp_config_dir)
    return temp_config_dir
