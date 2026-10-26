import pytest
from unittest.mock import patch
import sys
import os


class TestWrapper:
    """Test Python wrapper core functionality"""

    def test_main_function_exists(self):
        """Test that main function exists and is callable"""
        try:
            from gptcomet import main
            assert callable(main)
        except ImportError:
            pytest.skip("gptcomet module not available")

    def test_subprocess_execution(self):
        """Test that main dispatches to the binary without replacing pytest"""
        try:
            from gptcomet import main
        except ImportError:
            pytest.skip("gptcomet module not available")

        with patch('gptcomet.find_gptcomet_binary', return_value='/tmp/gptcomet'), \
             patch('gptcomet.os.execvpe') as mock_exec, \
             patch.object(sys, 'argv', ['gmsg', '--version']), \
             patch.object(sys, 'platform', 'linux'):
            main()

        mock_exec.assert_called_once()
        assert mock_exec.call_args.args[0] == '/tmp/gptcomet'

    def test_binary_not_found_error(self):
        """Test error handling when binary file is not found"""
        try:
            from gptcomet import main
        except ImportError:
            pytest.skip("gptcomet module not available")

        with patch('gptcomet.find_gptcomet_binary', side_effect=FileNotFoundError("missing")), \
             patch.object(sys, 'argv', ['gmsg', '--version']):
            with pytest.raises(FileNotFoundError):
                main()

    @patch('sys.platform', 'linux')
    def test_linux_platform_detection(self):
        """Test Linux platform detection"""
        try:
            from gptcomet import find_gptcomet_binary
            path = find_gptcomet_binary()
            assert 'linux' in path.lower()
        except (ImportError, AttributeError, FileNotFoundError):
            pytest.skip("find_gptcomet_binary or binary not available")

    @patch('sys.platform', 'darwin')
    def test_macos_platform_detection(self):
        """Test macOS platform detection"""
        try:
            from gptcomet import find_gptcomet_binary
            path = find_gptcomet_binary()
            assert 'mac' in path.lower() or 'darwin' in path.lower()
        except (ImportError, AttributeError, FileNotFoundError):
            pytest.skip("find_gptcomet_binary or binary not available")

    @patch('sys.platform', 'win32')
    def test_windows_platform_detection(self):
        """Test Windows platform detection"""
        try:
            from gptcomet import find_gptcomet_binary
            path = find_gptcomet_binary()
            assert 'win' in path.lower()
        except (ImportError, AttributeError, FileNotFoundError):
            pytest.skip("find_gptcomet_binary or binary not available")

    def test_environment_variables(self):
        """Test environment variables can be accessed"""
        test_value = 'test_value_12345'
        os.environ['GPTCOMET_TEST_VAR'] = test_value
        assert os.environ.get('GPTCOMET_TEST_VAR') == test_value
        # Cleanup
        del os.environ['GPTCOMET_TEST_VAR']

    def test_subprocess_with_args(self):
        """Test the binary receives the original CLI arguments"""
        try:
            from gptcomet import main
        except ImportError:
            pytest.skip("gptcomet module not available")

        with patch('gptcomet.find_gptcomet_binary', return_value='/tmp/gptcomet'), \
             patch('gptcomet.os.execvpe') as mock_exec, \
             patch.object(sys, 'argv', ['gmsg', 'commit', '--help']), \
             patch.object(sys, 'platform', 'linux'):
            main()

        mock_exec.assert_called_once()
        assert mock_exec.call_args.args[0] == '/tmp/gptcomet'
        assert mock_exec.call_args.args[1] == ['/tmp/gptcomet', 'commit', '--help']

    def test_module_import(self):
        """Test that gptcomet module can be imported"""
        try:
            import gptcomet
            assert gptcomet is not None
        except ImportError:
            pytest.skip("gptcomet module not available")
