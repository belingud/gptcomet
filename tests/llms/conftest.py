"""Test fixtures for LLM tests."""
import pytest


@pytest.fixture
def mock_response():
    """Mock response fixture."""

    class MockResponse:
        def __init__(self, json_data, status_code=200):
            self.json_data = json_data
            self.status_code = status_code
            self.text = str(json_data)

        def json(self):
            return self.json_data

        def raise_for_status(self):
            if self.status_code >= 400:
                raise Exception(f"HTTP Error: {self.status_code}")

    return MockResponse


@pytest.fixture
def base_config():
    """Base config fixture."""
    return {
        "api_key": "test-api-key",
        "max_tokens": 100,
        "temperature": 0.7,
        "top_p": 0.9,
        "frequency_penalty": 0.0,
        "presence_penalty": 0.0,
    }


@pytest.fixture
def sample_message():
    """Sample message fixture."""
    return "Hello, how are you?"


@pytest.fixture
def sample_history():
    """Sample chat history fixture."""
    return [
        {"role": "user", "content": "What is AI?"},
        {
            "role": "assistant",
            "content": "AI is a branch of computer science that aims to create intelligent machines.",
        },
    ]
