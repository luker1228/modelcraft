"""ModelCraft Python SDK — RLS-aware data API client."""

import requests
from typing import Optional, Dict, Any, List


class Client:
    """Client for the ModelCraft open data API with RLS context support.

    Usage:
        client = Client(
            endpoint="https://modelcraft.example.com",
            api_token="mc_pat_xxx",
            org="my-org",
            project="my-project",
        )

        orders = client.model("orders").list(
            db="main",
            limit=10,
            user_id="customer_123",
            user_name="zhangsan",
            user_roles="admin,manager",
        )
    """

    def __init__(
        self,
        endpoint: str,
        api_token: str,
        org: str,
        project: str,
    ):
        self.endpoint = endpoint.rstrip("/")
        self.api_token = api_token
        self.org = org
        self.project = project

    def _headers(
        self,
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> Dict[str, str]:
        return {
            "Authorization": f"Bearer {self.api_token}",
            "Content-Type": "application/json",
            "X-MC-User-ID": user_id,
            "X-MC-User-Name": user_name,
            "X-MC-User-Roles": user_roles,
        }

    def model(self, model_name: str) -> "ModelQuery":
        return ModelQuery(self, model_name)


class ModelQuery:
    """Query helper for a specific model."""

    def __init__(self, client: Client, model_name: str):
        self._client = client
        self._model = model_name

    def _url(self, db: str) -> str:
        return (
            f"{self._client.endpoint}/end-user/graphql"
            f"/org/{self._client.org}/project/{self._client.project}"
            f"/db/{db}/model/{self._model}"
        )

    def list(
        self,
        db: str,
        limit: int = 10,
        where: Optional[Dict] = None,
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> List[Dict[str, Any]]:
        """List model records with RLS filtering."""
        query = """
        query List($limit: Int!, $where: JSON) {
            findMany(limit: $limit, where: $where) {
                items
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={
                "query": query,
                "variables": {"limit": limit, "where": where or {}},
            },
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        data = resp.json()
        if "errors" in data:
            raise RuntimeError(data["errors"])
        return data.get("data", {}).get("findMany", {}).get("items", [])

    def create(
        self,
        db: str,
        data: Dict[str, Any],
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> Dict[str, Any]:
        """Create a model record with RLS check validation."""
        query = """
        mutation Create($data: JSON!) {
            createOne(data: $data) {
                item
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={"query": query, "variables": {"data": data}},
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        result = resp.json()
        if "errors" in result:
            raise RuntimeError(result["errors"])
        return result.get("data", {}).get("createOne", {}).get("item", {})

    def update(
        self,
        db: str,
        where: Dict[str, Any],
        data: Dict[str, Any],
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> Dict[str, Any]:
        """Update a model record with RLS check validation."""
        query = """
        mutation Update($where: JSON!, $data: JSON!) {
            updateOne(where: $where, data: $data) {
                item
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={"query": query, "variables": {"where": where, "data": data}},
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        result = resp.json()
        if "errors" in result:
            raise RuntimeError(result["errors"])
        return result.get("data", {}).get("updateOne", {}).get("item", {})

    def delete(
        self,
        db: str,
        where: Dict[str, Any],
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> bool:
        """Delete a model record with RLS check."""
        query = """
        mutation Delete($where: JSON!) {
            deleteOne(where: $where) {
                item { id }
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={"query": query, "variables": {"where": where}},
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        return "errors" not in resp.json()
