# modelcraft-agent/agents/tools.py
"""All @tool functions shared between admin and end-user agents."""
import json
from functools import lru_cache
from typing import Annotated

from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from langgraph.prebuilt import InjectedState

import config
from agents.shared import AgentState, make_client, log_tool_errors


@lru_cache(maxsize=1)
def _get_llm() -> ChatOpenAI:
    return ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    )


@tool
@log_tool_errors
async def list_projects(
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all projects in the current organization.

    Returns:
        JSON array of projects with id, slug, title, description, status.
    """
    result = await make_client(state).list_projects(org_name=state["org_name"])
    if "errors" in result and result["errors"]:
        return f"GraphQL error: {result['errors']}"
    projects = result.get("data", {}).get("projects", [])
    return json.dumps(projects, ensure_ascii=False)


@tool
@log_tool_errors
async def list_databases(
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all databases available in the current project's cluster.
    Call this before list_models to know which database names exist.

    Returns:
        JSON array of database names, e.g. ["maindb", "analyticsdb"].
    """
    result = await make_client(state).list_databases(
        org_name=state["org_name"],
        project_slug=state["project_slug"],
    )
    if "errors" in result and result["errors"]:
        return f"GraphQL error: {result['errors']}"
    edges = result.get("data", {}).get("listDatabases", {}).get("edges", [])
    names = [e["node"]["name"] for e in edges]
    return json.dumps(names, ensure_ascii=False)


@tool
@log_tool_errors
async def list_models(
    database_name: str,
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all data models in a database of the current project.

    Args:
        database_name: Database name, e.g. "maindb"

    Returns:
        JSON array of models with id, name, title, description, databaseName, displayField.
    """
    result = await make_client(state).list_models(
        org_name=state["org_name"],
        project_slug=state["project_slug"],
        database_name=database_name,
    )
    if "errors" in result and result["errors"]:
        return f"GraphQL error: {result['errors']}"
    data = result.get("data", {}).get("models", {})
    return json.dumps(data, ensure_ascii=False)


@tool
@log_tool_errors
async def get_model_fields(
    model_id: str,
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    Get the field definitions of a model by its ID.
    Use this before query_model to know what fields are available.

    Args:
        model_id: Model ID (from list_models)

    Returns:
        JSON array of fields with name, title, schemaType, format, isPrimary, isUnique, etc.
    """
    result = await make_client(state).get_model_fields(
        org_name=state["org_name"],
        project_slug=state["project_slug"],
        model_id=model_id,
    )
    if "errors" in result and result["errors"]:
        return f"GraphQL error: {result['errors']}"
    fields = result.get("data", {}).get("fields", [])
    return json.dumps(fields, ensure_ascii=False)


@tool
@log_tool_errors
async def query_model(
    db_name: str,
    model_name: str,
    fields: list[str],
    take: int,
    state: Annotated[AgentState, InjectedState()],
    where: dict | None = None,
    skip: int = 0,
) -> str:
    """
    Query records from a ModelCraft data model.

    Args:
        db_name: Database name, e.g. "maindb"
        model_name: Model name, e.g. "users"
        fields: Field names to return, e.g. ["id", "name", "createdAt"]
        take: Max records to return (1-200)
        where: Optional filter JSON, e.g. {"name": {"contains": "张"}}
        skip: Records to skip for pagination (default 0)

    Returns:
        JSON with items array and totalCount.
    """
    take = max(1, min(take, 200))
    result = await make_client(state).find_many(
        org_name=state["org_name"],
        project_slug=state["project_slug"],
        db_name=db_name,
        model_name=model_name,
        fields=fields,
        where=where,
        take=take,
        skip=skip,
    )
    if "errors" in result and result["errors"]:
        return f"GraphQL error: {result['errors']}"
    data = result.get("data", {}).get("findMany", {})
    return json.dumps(data, ensure_ascii=False)


@tool
@log_tool_errors
async def nl2filter(
    natural_language: str,
    field_names: list[str],
) -> str:
    """
    Convert a natural language filter description into a ModelCraft where JSON.

    Args:
        natural_language: User's filter intent, e.g. "名字包含张的且年龄大于18"
        field_names: Available field names in the model, e.g. ["id", "name", "age"]

    Returns:
        A JSON string representing the ModelCraft where clause,
        e.g. {"AND": [{"name": {"contains": "张"}}, {"age": {"gt": 18}}]}
    """
    llm = _get_llm()
    system_prompt = f"""You are a filter JSON generator for ModelCraft.
Convert the user's natural language filter description into a valid ModelCraft where JSON.

Available fields: {field_names}

ModelCraft where JSON rules:
- Top level: {{"AND": [...]}}, {{"OR": [...]}}, or a single field condition
- String operators: contains, startsWith, endsWith, equals, not
- Number operators: equals, not, gt, gte, lt, lte
- Boolean: {{"active": {{"equals": true}}}}
- Combined: {{"AND": [{{"name": {{"contains": "张"}}}}, {{"age": {{"gte": 18}}}}]}}

Return ONLY the raw JSON object, no explanation, no markdown."""

    response = await llm.ainvoke([
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": natural_language},
    ])
    raw = response.content.strip()
    json.loads(raw)  # validate
    return raw
