"""
Gateway GraphQL client.

All requests go through Gateway(:8090), never directly to backend(:8080).
Gateway validates JWT, injects X-User-ID, preserves Authorization header,
then forwards to Backend.
"""
import re
from typing import Any
from urllib.parse import quote

import httpx

import config

_FIELD_NAME_RE = re.compile(r'^[_A-Za-z][_0-9A-Za-z]*$')


class GraphQLClient:
    """Async GraphQL client that forwards Authorization header to Gateway."""

    def __init__(self, authorization: str):
        """
        Args:
            authorization: The full 'Authorization: Bearer <token>' value
                           received from the incoming request. Forwarded as-is.
        """
        self._authorization = authorization

    def _build_url(self, org_name: str, project_slug: str, db_name: str, model_name: str) -> str:
        return (
            f"{config.GATEWAY_URL}/graphql/org/{quote(org_name, safe='')}"
            f"/project/{quote(project_slug, safe='')}"
            f"/db/{quote(db_name, safe='')}/model/{quote(model_name, safe='')}"
        )

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
        """
        Execute a findMany query on the runtime GraphQL endpoint.

        Uses GraphQL variables for `where` to avoid JSON/GraphQL syntax mismatch.

        Returns the full GraphQL response dict, e.g.:
        {"data": {"findMany": {"items": [...], "totalCount": N}}, "errors": [...]}
        """
        validated_fields = fields if fields else ["id"]
        self._validate_fields(validated_fields)
        fields_str = " ".join(validated_fields)

        # Use GraphQL variables — avoids JSON/GraphQL syntax incompatibility
        # where JSON must be passed as a variable, not inlined as a literal
        query = """
query FindMany($take: Int, $skip: Int, $where: JSON) {
  findMany(take: $take, skip: $skip, where: $where) {
    items { %s }
    totalCount
  }
}
""" % fields_str

        variables: dict[str, Any] = {"take": take, "skip": skip}
        if where is not None:
            variables["where"] = where

        url = self._build_url(org_name, project_slug, db_name, model_name)
        headers = {
            "Content-Type": "application/json",
            "Authorization": self._authorization,
        }
        payload = {"query": query, "variables": variables}

        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(url, headers=headers, json=payload)
            response.raise_for_status()
            return response.json()
