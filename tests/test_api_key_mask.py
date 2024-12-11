import pytest

from gptcomet.utils import api_key_mask


@pytest.mark.parametrize(
    "api_key, show_first, expected",
    [
        (123, 3, 123),  # Non-string input
        (None, 3, None),  # Non-string input
        ([1, 2, 3], 3, [1, 2, 3]),  # Non-string input
        ("sk-or-v1-abcdefghijklmn", 3, "sk-or-v1-abc***********"),  # With prefix sk-or-v1-
        ("sk-abcdefghijklmn", 3, "sk-abc***********"),  # With prefix sk-
        ("gsk_abcdefghijklmn", 3, "gsk_abc***********"),  # With prefix gsk_
        ("xai-abcdefghijklmn", 3, "xai-abc***********"),  # With prefix xai-
        ("abcdefghijklmn", 3, "abc***********"),  # No prefix
        ("abc", 3, "abc"),  # Short API key
        ("", 3, ""),  # Empty API key
        ("sk-abcdefghijklmn", 4, "sk-abcd**********"),  # Custom show_first
        ("sk-abcdefghijklmn", 0, "sk-**************"),  # show_first zero, include prefix
        ("abcdefghijklmn", 0, "**************"),  # show_first zero, no prefix
        ("sk-abcdefghijklmn", -1, "sk-**************"),  # show_first negative, same as zero
        ("abcdefghijklmn", -1, "**************"), # show_first negative, same as zero
        (None, 0, None), # Non-string input with show_first zero, expect same as input
    ],
)
def test_api_key_mask(api_key, show_first, expected):
    assert api_key_mask(api_key, show_first) == expected