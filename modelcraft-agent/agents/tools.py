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


_PAGE_KNOWLEDGE: dict[str, dict[str, str]] = {
    "model-editor": {
        "name": "模型编辑器",
        "description": "管理项目的数据模型结构。左侧栏：数据库选择器 + 模型列表；主区域：选中模型后展示数据记录；右侧抽屉：字段详情与编辑。",
        "workflow": (
            "前提：数据库选择器需要有已接管的数据库才能使用。\n"
            "  - 若下拉框为空 → 先去「数据库管理」页面接管数据库（/databases）\n"
            "  - 若还没有数据库连接信息 → 先去「项目设置」配置集群（/settings）\n"
            "1. 用数据库选择器选择目标数据库\n"
            "2. 在左侧模型列表中点击模型，右侧主区域显示该模型的数据记录\n"
            "3. 点击模型行末的设置图标（→）打开右侧字段管理抽屉\n"
            "4. 在抽屉中：新建字段（+ 按钮）、编辑字段、废弃字段、删除已废弃字段\n"
            "5. 支持在主区域直接新建 / 编辑数据记录"
        ),
    },
    "databases": {
        "name": "数据库管理",
        "description": "接管项目使用的 MySQL 数据库，设置访问模式（托管/自建），触发同步将数据库表导入为数据模型。",
        "workflow": (
            "前提：接管数据库前，必须先在「项目设置」配置数据库集群连接信息（MySQL 主机、端口、账号密码）。\n"
            "1. 查看已接管的数据库列表\n"
            "2. 点击「接管数据库」，选择目标数据库名\n"
            "3. 选择访问模式：托管（ModelCraft 管理）或自建（用户自己的 MySQL）\n"
            "4. 触发同步：将数据库中的表自动导入为数据模型\n"
            "5. 同步后可在「模型编辑器」中看到导入的模型"
        ),
    },
    "enums": {
        "name": "枚举管理",
        "description": "管理项目中的枚举类型（Enum）。枚举可绑定到字段，限制字段的可选值范围。",
        "workflow": (
            "1. 查看现有枚举列表\n"
            "2. 新建枚举：填写枚举名称，添加枚举值\n"
            "3. 编辑枚举：修改枚举值列表\n"
            "4. 删除枚举前需确保无字段依赖它"
        ),
    },
    "roles": {
        "name": "角色权限管理（RBAC）",
        "description": "管理项目内的角色与权限包。角色可分配给终端用户，控制其对数据的增删改查权限。",
        "workflow": (
            "1. 查看角色列表，点击角色进入详情\n"
            "2. 在角色详情中配置权限包（Permission Bundle）\n"
            "3. 创建新角色，绑定权限包\n"
            "4. 权限包版本化：快照只读，最多 5 个版本，支持回滚"
        ),
    },
    "end-user-access": {
        "name": "用户授权",
        "description": "为终端用户（EndUser）分配项目访问权限和角色。终端用户在 Org 级别创建，在项目级别授权。",
        "workflow": (
            "1. 查看已被授权访问本项目的终端用户列表\n"
            "2. 添加授权：选择已有终端用户，分配角色\n"
            "3. 修改角色：更改用户在本项目中的权限角色\n"
            "4. 撤销授权：移除用户对本项目的访问"
        ),
    },
    "identity-settings": {
        "name": "身份认证设置",
        "description": "配置终端用户的登录认证方式，例如邮箱密码、第三方 OAuth 等。",
        "workflow": (
            "1. 查看当前启用的认证方式\n"
            "2. 开启 / 关闭认证方式\n"
            "3. 配置 OAuth 应用参数（Client ID / Secret）"
        ),
    },
    "settings": {
        "name": "项目设置",
        "description": "修改项目基本信息，包括名称、描述、状态；配置数据库集群连接（MySQL 主机、端口、账号密码），是使用数据库接管功能的前提。",
        "workflow": (
            "1. 修改项目名称或描述，点击保存\n"
            "2. 配置数据库集群：填写 MySQL host / port / user / password\n"
            "3. 归档项目：项目将不再出现在活跃列表中\n"
            "4. 删除项目：不可恢复，需二次确认"
        ),
    },
    "workspace": {
        "name": "所有项目",
        "description": "组织下的项目列表。支持网格 / 列表视图切换，可搜索、新建、编辑、归档项目。",
        "workflow": (
            "1. 点击项目卡片进入该项目（跳转到 model-editor）\n"
            "2. 新建项目：填写名称、slug、数据库连接信息\n"
            "3. 项目卡片右上角菜单：编辑 / 归档 / 删除"
        ),
    },
    "end-users": {
        "name": "终端用户管理",
        "description": "管理组织下的所有终端用户账号。终端用户是通过 API 访问数据的外部用户（非管理员）。",
        "workflow": (
            "1. 查看组织下所有终端用户\n"
            "2. 新建终端用户：设置用户名、邮箱、密码\n"
            "3. 查看用户详情：已授权的项目及角色\n"
            "4. 重置密码、禁用账号"
        ),
    },
    "developers": {
        "name": "成员管理",
        "description": "管理组织内的开发者成员账号与角色（Owner / Admin / Member）。",
        "workflow": (
            "1. 查看成员列表\n"
            "2. 邀请新成员：输入邮箱发送邀请\n"
            "3. 修改成员角色\n"
            "4. 移除成员"
        ),
    },
    "cluster": {
        "name": "数据库集群",
        "description": "管理项目连接的数据库集群。每个集群对应一个 MySQL 数据库连接，模型在集群下的数据库中创建。",
        "workflow": (
            "1. 查看已配置的数据库集群列表\n"
            "2. 新建集群：填写 host / port / user / password / database\n"
            "3. 测试连接：验证连接信息是否正确\n"
            "4. 删除集群前需确保无模型依赖"
        ),
    },
}

_PAGE_KNOWLEDGE_INDEX = ", ".join(_PAGE_KNOWLEDGE.keys())


@tool
async def get_page_knowledge(page: str) -> str:
    """
    Get the operation guide for a specific page in ModelCraft.
    Call this when the user asks how to use the current page or needs
    step-by-step instructions for a page's workflow.

    Args:
        page: Route segment of the page, e.g. "model-editor", "enums", "settings",
              "databases", "roles", "end-user-access", "identity-settings",
              "workspace", "end-users", "developers", "cluster"

    Returns:
        JSON with name, description, and workflow steps for the page.
    """
    knowledge = _PAGE_KNOWLEDGE.get(page)
    if not knowledge:
        available = _PAGE_KNOWLEDGE_INDEX
        return json.dumps({"error": f"No knowledge for page '{page}'. Available: {available}"}, ensure_ascii=False)
    return json.dumps(knowledge, ensure_ascii=False)


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
