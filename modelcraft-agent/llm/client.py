"""
LLM client factory.

Supports deepseek/openai (OpenAI-compatible SDK) and anthropic.
Switch provider via LLM_PROVIDER env var — agent code never imports provider SDK directly.
"""
from openai import AsyncOpenAI
import config


def get_llm_client() -> AsyncOpenAI:
    """
    Return an async OpenAI-compatible client.

    DeepSeek and OpenAI both use the OpenAI SDK; only base_url differs.
    For Anthropic, swap this to the anthropic SDK in a future iteration.
    """
    if config.LLM_PROVIDER in ("deepseek", "openai"):
        kwargs = {"api_key": config.LLM_API_KEY}
        if config.LLM_BASE_URL:
            kwargs["base_url"] = config.LLM_BASE_URL
        return AsyncOpenAI(**kwargs)

    raise ValueError(
        f"Unsupported LLM_PROVIDER='{config.LLM_PROVIDER}'. "
        "Supported values: deepseek, openai"
    )


def get_model_name() -> str:
    return config.LLM_MODEL
