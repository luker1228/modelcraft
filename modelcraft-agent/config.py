import os
from dotenv import load_dotenv

load_dotenv()

# LLM provider settings
LLM_PROVIDER: str = os.environ.get("LLM_PROVIDER", "deepseek")
LLM_MODEL: str = os.environ.get("LLM_MODEL", "deepseek-chat")
LLM_API_KEY: str = os.environ.get("LLM_API_KEY", "")
LLM_BASE_URL: str = os.environ.get("LLM_BASE_URL", "https://api.deepseek.com")

# Gateway settings — all GraphQL calls MUST go through gateway, never direct to backend
GATEWAY_URL: str = os.environ.get("GATEWAY_URL", "http://localhost:8090")

# Server settings
PORT: int = int(os.environ.get("PORT", "8000"))
