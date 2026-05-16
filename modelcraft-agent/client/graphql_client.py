"""
Gateway GraphQL client.

All requests go through the gateway, never directly to backend.
Three endpoint types:
  - Org GraphQL:     /graphql/org/{orgName}
  - Project GraphQL: /graphql/org/{orgName}/project/{slug}
  - Runtime GraphQL: /graphql/org/{orgName}/project/{slug}/db/{db}/model/{model}
"""
import re
import time
from typing import Any
from urllib.parse import quote

import httpx

import config
from logging_setup import get_logger

_FIELD_NAME_RE = re.compile(r'^[_A-Za-z][_0-9A-Za-z]*$')


class GraphQLClient:
    """Async GraphQL client that forwards Authorization header to Gateway."""

    def __init__(self, authorization: str):
        self._authorization = authorization

    # ------------------------------------------------------------------
    # URL builders
    # ------------------------------------------------------------------

    def _org_url(self, org_name: str) -> str:
        return f"{config.GATEWAY_URL}/graphql/org/{quote(org_name, safe='')}"

    def _project_url(self, org_name: str, project_slug: str) -> str:
        return (
            f"{config.GATEWAY_URL}/graphql/org/{quote(org_name, safe='')}"
            f"/project/{quote(project_slug, safe='')}"
        )

    def _runtime_url(self, org_name: str, project_slug: str, db_name: str, model_name: str) -> str:
        return (
            f"{config.GATEWAY_URL}/graphql/org/{quote(org_name, safe='')}"
            f"/project/{quote(project_slug, safe='')}"
            f"/db/{quote(db_name, safe='')}/model/{quote(model_name, safe='')}"
        )

    # ------------------------------------------------------------------
    # Low-level HTTP executor
    # ------------------------------------------------------------------

    async def _execute(self, url: str, query: str, variables: dict | None = None, operation: str = "query") -> dict[str, Any]:
        headers = {
            "Content-Type": "application/json",
            "Authorization": self._authorization,
        }
        payload: dict[str, Any] = {"query": query}
        if variables:
            payload["variables"] = variables

        log = get_logger()
        log.info("graphql.call.start", url=url, operation=operation)
        start = time.perf_counter()
        response = None
        result = None
        try:
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(url, headers=headers, json=payload)
                response.raise_for_status()
                result = response.json()
        except Exception:
            duration_ms = round((time.perf_counter() - start) * 1000, 2)
            status_code = response.status_code if response is not None else 0
            log.exception("error", url=url, operation=operation, duration_ms=duration_ms)
            log.info("graphql.call.end", url=url, duration_ms=duration_ms, has_errors=True, status_code=status_code)
            raise

        duration_ms = round((time.perf_counter() - start) * 1000, 2)
        has_errors = bool(result.get("errors"))
        log.info("graphql.call.end", url=url, duration_ms=duration_ms, has_errors=has_errors, status_code=response.status_code)
        return result

    # ------------------------------------------------------------------
    # Org-level operations
    # ------------------------------------------------------------------

    async def list_projects(self, org_name: str) -> dict[str, Any]:
        """List all projects in an org. Returns {data: {projects: [...]}}"""
        query = """
query ListProjects {
  projects {
    id
    slug
    title
    description
    status
    createdAt
    updatedAt
  }
}
"""
        return await self._execute(self._org_url(org_name), query, operation="listProjects")

    # ------------------------------------------------------------------
    # Project-level operations
    # ------------------------------------------------------------------

    async def list_models(self, org_name: str, project_slug: str, database_name: str, page_index: int = 1, page_size: int = 50) -> dict[str, Any]:
        """List models in a project database. Returns {data: {models: {items: [...], hasNextPage}}}"""
        query = """
query ListModels($input: ModelQueryInput) {
  models(input: $input) {
    items {
      id
      name
      title
      description
      databaseName
      storageType
      displayField
      createdAt
      updatedAt
    }
    hasNextPage
  }
}
"""
        variables = {"input": {"databaseName": database_name, "pageIndex": page_index, "pageSize": page_size}}
        return await self._execute(self._project_url(org_name, project_slug), query, variables, operation="listModels")

    async def get_model_fields(self, org_name: str, project_slug: str, model_id: str) -> dict[str, Any]:
        """Get all fields of a model. Returns {data: {fields: [...]}}"""
        query = """
query GetFields($modelID: ID!) {
  fields(modelID: $modelID) {
    name
    title
    schemaType
    format
    nonNull
    required
    isUnique
    isPrimary
    isArray
    description
    enumName
  }
}
"""
        return await self._execute(self._project_url(org_name, project_slug), query, {"modelID": model_id}, operation="getFields")

    async def get_model_by_name(self, org_name: str, project_slug: str, model_name: str, database_name: str) -> dict[str, Any]:
        """Get a model by name. Returns {data: {modelByName: {model: {...}, error: ...}}}"""
        query = """
query GetModelByName($name: String!, $databaseName: String!) {
  modelByName(name: $name, databaseName: $databaseName) {
    model {
      id
      name
      title
      description
      databaseName
      displayField
      fields {
        name
        title
        schemaType
        format
        nonNull
        isPrimary
        isUnique
      }
    }
    error {
      ... on ResourceNotFound { message }
      ... on InvalidInput { message }
    }
  }
}
"""
        variables = {"name": model_name, "databaseName": database_name}
        return await self._execute(self._project_url(org_name, project_slug), query, variables, operation="getModelByName")

    # ------------------------------------------------------------------
    # Runtime (data) operations
    # ------------------------------------------------------------------

    @staticmethod
    def _validate_fields(fields: list[str]) -> None:
        for f in fields:
            if not _FIELD_NAME_RE.match(f):
                raise ValueError(f"Invalid GraphQL field name: {f!r}")

    async def find_many(
        self,
        org_name: str,
        project_slug: str,
        db_name: str,
        model_name: str,
        fields: list[str],
        where: dict | None = None,
        take: int = 20,
        skip: int = 0,
    ) -> dict[str, Any]:
        """Query records from a runtime model endpoint."""
        validated_fields = fields if fields else ["id"]
        self._validate_fields(validated_fields)
        fields_str = " ".join(validated_fields)
        where_type = f"T{model_name}WhereInput"
        query = f"""
query FindMany($take: Int, $skip: Int, $where: {where_type}) {{
  findMany(take: $take, skip: $skip, where: $where) {{
    items {{ {fields_str} }}
    totalCount
  }}
}}
"""
        variables: dict[str, Any] = {"take": take, "skip": skip}
        if where is not None:
            variables["where"] = where
        return await self._execute(self._runtime_url(org_name, project_slug, db_name, model_name), query, variables, operation="findMany")

    async def create_record(
        self,
        org_name: str,
        project_slug: str,
        db_name: str,
        model_name: str,
        data: dict[str, Any],
        return_fields: list[str],
    ) -> dict[str, Any]:
        """Create a record via runtime GraphQL. Returns {data: {createOne: {...}}}"""
        self._validate_fields(return_fields)
        fields_str = " ".join(return_fields)
        data_type = f"T{model_name}CreateInput"
        query = f"""
mutation CreateOne($data: {data_type}!) {{
  createOne(data: $data) {{
    {fields_str}
  }}
}}
"""
        return await self._execute(self._runtime_url(org_name, project_slug, db_name, model_name), query, {"data": data}, operation="createOne")

    async def update_record(
        self,
        org_name: str,
        project_slug: str,
        db_name: str,
        model_name: str,
        id: str,
        data: dict[str, Any],
        return_fields: list[str],
    ) -> dict[str, Any]:
        """Update a record by id. Returns {data: {updateOne: {...}}}"""
        self._validate_fields(return_fields)
        fields_str = " ".join(return_fields)
        data_type = f"T{model_name}UpdateInput"
        query = f"""
mutation UpdateOne($id: ID!, $data: {data_type}!) {{
  updateOne(id: $id, data: $data) {{
    {fields_str}
  }}
}}
"""
        return await self._execute(self._runtime_url(org_name, project_slug, db_name, model_name), query, {"id": id, "data": data}, operation="updateOne")

    async def delete_record(
        self,
        org_name: str,
        project_slug: str,
        db_name: str,
        model_name: str,
        id: str,
    ) -> dict[str, Any]:
        """Delete a record by id. Returns {data: {deleteOne: {id}}}"""
        query = """
mutation DeleteOne($id: ID!) {
  deleteOne(id: $id) {
    id
  }
}
"""
        return await self._execute(self._runtime_url(org_name, project_slug, db_name, model_name), query, {"id": id}, operation="deleteOne")
