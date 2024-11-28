import pytest
from prompt_toolkit.document import Document
from prompt_toolkit.validation import ValidationError

from gptcomet._validator import KEYS_VALIDATOR, RequiredValidator, URLValidator


def test_keys_validator_retries():
    assert KEYS_VALIDATOR["retries"]["validator"]("123")
    assert not KEYS_VALIDATOR["retries"]["validator"]("12.3")
    assert not KEYS_VALIDATOR["retries"]["validator"]("-123")
    assert not KEYS_VALIDATOR["retries"]["validator"]("abc")


def test_keys_validator_max_tokens():
    assert KEYS_VALIDATOR["max_tokens"]["validator"]("1000")
    assert not KEYS_VALIDATOR["max_tokens"]["validator"]("12.3")
    assert not KEYS_VALIDATOR["max_tokens"]["validator"]("-123")
    assert not KEYS_VALIDATOR["max_tokens"]["validator"]("abc")


def test_keys_validator_top_p():
    validator = KEYS_VALIDATOR["top_p"]["validator"]
    assert validator("0")
    assert validator("0.5")
    assert validator("1")
    assert not validator("-0.1")
    assert not validator("1.1")
    assert not validator("abc")


def test_keys_validator_temperature():
    validator = KEYS_VALIDATOR["temperature"]["validator"]
    assert validator("0.1")
    assert validator("0.5")
    assert validator("1")
    assert not validator("0")
    assert not validator("1.1")
    assert not validator("abc")


def test_keys_validator_frequency_penalty():
    validator = KEYS_VALIDATOR["frequency_penalty"]["validator"]
    assert validator("-2")
    assert validator("0")
    assert validator("2")
    assert not validator("-2.1")
    assert not validator("2.1")
    assert not validator("abc")


def test_required_validator():
    validator = RequiredValidator("test_field")
    
    # Test empty input
    with pytest.raises(ValidationError) as exc:
        validator.validate(Document(""))
    assert "test_field is required" in str(exc.value)
    
    # Test valid input
    validator.validate(Document("some value"))  # Should not raise


def test_url_validator():
    validator = URLValidator()
    
    # Test empty input
    with pytest.raises(ValidationError) as exc:
        validator.validate(Document(""))
    assert "URL cannot be empty" in str(exc.value)
    
    # Test invalid URL format (no protocol)
    with pytest.raises(ValidationError) as exc:
        validator.validate(Document("example.com/"))
    assert "URL must start with http:// or https://" in str(exc.value)
    
    # Test invalid URL format (no trailing slash)
    with pytest.raises(ValidationError) as exc:
        validator.validate(Document("http://example.com"))
    assert "URL must end with /" in str(exc.value)
    
    # Test valid URLs
    validator.validate(Document("http://example.com/"))
    validator.validate(Document("https://example.com/"))
