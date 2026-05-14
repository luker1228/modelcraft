"""
Gateway GraphQL client.

All requests go through Gateway(:8090), never directly to backend(:8080).
Gateway validates JWT, injects X-User-ID, preserves Authorization header,
then forwards to Backend.
"""
import json
from typing import Any

import httpx

import config


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
            f"{config.GATEWAY_URL}/graphql/org/{org_name}"
            f"/project/{project_slug}/db/{db_name}/model/{model_name}"
        )

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

        Returns the full GraphQL response dict, e.g.:
        {"data": {"findMany": {"items": [...], "totalCount": N}}, "errors": [...]}
        """
        fields_str = " ".join(fields) if fields else "id"
        where_arg = f", where: {json.dumps(where)}" if where else ""
        query = (
            f"{{ findMany(take: {take}, skip: {skip}{where_arg}) "
            f"{{ items {{ {fields_str} }} totalCount }} }}"
        )

        url = self._build_url(org_name, project_slug, db_name, model_name)
        headers = {
            "Content-Type": "application/json",
            "Authorization": self._authorization,
        }
        payload = {"query": query}

        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(url, headers=headers, json=payload)
            response.raise_for_status()
            return response.json()
