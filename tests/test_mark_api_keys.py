import pytest

from gptcomet.utils import mask_api_keys


@pytest.mark.parametrize(
    "data, show_first, expected",
    [
        ({"api_key": "sk-abcdef1234567890"}, 3, {"api_key": "sk-abc*************"}),
        (
            {"api_key": "sk-abcdef1234567890", "other": "value"},
            3,
            {"api_key": "sk-abc*************", "other": "value"},
        ),
        (
            {"nested": {"api_key": "sk-abcdef1234567890"}},
            3,
            {"nested": {"api_key": "sk-abc*************"}},
        ),
        (
            {"nested": {"api_key": "sk-abcdef1234567890", "other": "value"}},
            3,
            {"nested": {"api_key": "sk-abc*************", "other": "value"}},
        ),
        (
            [{"api_key": "sk-abcdef1234567890"}],
            3,
            [{"api_key": "sk-abc*************"}],
        ),
        (["sk-abcdef1234567890"], 3, ["sk-abc*************"]),  # String in list
        ("sk-abcdef1234567890", 3, "sk-abc*************"),  # String input
        (12345, 3, 12345),  # Non-dict/list/string input
        (None, 3, None),  # None input
        # Test with different show_first values
        ({"api_key": "sk-abcdef1234567890"}, 0, {"api_key": "sk-****************"}),
        ({"api_key": "sk-abcdef1234567890"}, 5, {"api_key": "sk-abcde***********"}),
        ({"api_key": "sk-abcdef1234567890"}, -1, {"api_key": "sk-****************"}),
    ],
)
def test_mask_api_keys(data, show_first, expected):
    assert mask_api_keys(data, show_first) == expected
