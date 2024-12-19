"""Test VertexLLM."""
import pytest

from gptcomet.llms.vertexllm import VertexLLM


@pytest.fixture
def vertex_config(base_config):
    """Vertex config fixture."""
    config = base_config.copy()
    config.update({
        "project_id": "test-project",
        "location": "us-central1",
    })
    return config


def test_vertex_init(vertex_config):
    """Test VertexLLM initialization."""
    llm = VertexLLM(vertex_config)
    assert llm.api_base == "https://us-central1-aiplatform.googleapis.com/v1"
    assert llm.model == "gemini-pro"
    assert llm.project_id == "test-project"
    assert llm.location == "us-central1"
    assert "generateContent" in llm.completion_path
    assert llm.answer_path == "candidates.0.content.parts.0.text"


def test_vertex_init_missing_project_id(base_config):
    """Test VertexLLM initialization with missing project_id."""
    with pytest.raises(ValueError, match="project_id is required for Vertex AI"):
        VertexLLM(base_config)


def test_vertex_build_headers(vertex_config):
    """Test build_headers method."""
    llm = VertexLLM(vertex_config)
    headers = llm.build_headers()
    assert headers["Authorization"] == "Bearer test-api-key"


def test_vertex_build_payload_simple(vertex_config, sample_message):
    """Test build_payload method with simple message."""
    llm = VertexLLM(vertex_config)
    payload = llm.format_messages(sample_message)

    assert len(payload["contents"]) == 1
    assert payload["contents"][0]["author"] == "user"
    assert payload["contents"][0]["content"] == sample_message
    assert payload["generation_config"]["max_output_tokens"] == vertex_config["max_tokens"]
    assert payload["generation_config"]["temperature"] == vertex_config["temperature"]
    assert payload["generation_config"]["top_p"] == vertex_config["top_p"]


def test_vertex_build_payload_with_history(vertex_config, sample_message, sample_history):
    """Test build_payload method with chat history."""
    llm = VertexLLM(vertex_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    assert len(payload["contents"]) == len(sample_history) + 1
    for i, msg in enumerate(sample_history):
        assert payload["contents"][i]["author"] == msg["role"]
        assert payload["contents"][i]["content"] == msg["content"]

    last_message = payload["contents"][-1]
    assert last_message["author"] == "user"
    assert last_message["content"] == sample_message


def test_vertex_build_payload_with_penalties(vertex_config, sample_message):
    """Test build_payload method with frequency and presence penalties."""
    config = vertex_config.copy()
    config.update({
        "frequency_penalty": 0.5,
        "presence_penalty": 0.3,
    })

    llm = VertexLLM(config)
    payload = llm.format_messages(sample_message)

    assert payload["generation_config"]["frequency_penalty"] == 0.5
    assert payload["generation_config"]["presence_penalty"] == 0.3


def test_vertex_get_required_config():
    """Test get_required_config method."""
    config = VertexLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "project_id" in config
    assert "location" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://us-central1-aiplatform.googleapis.com/v1"
    assert config["model"][0] == "gemini-pro"
    assert config["location"][0] == "us-central1"
