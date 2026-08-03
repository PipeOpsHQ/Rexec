"""Sandbox service for managing isolated environments.

Wire protocol remains ``/api/containers``. Public names prefer *sandbox*;
``Container*`` aliases remain for backward compatibility.
"""

from typing import TYPE_CHECKING

from rexec.types import (
    Container,
    CreateContainerRequest,
    CreateSandboxRequest,
    Sandbox,
)

if TYPE_CHECKING:
    from rexec.client import RexecClient

__all__ = [
    "SandboxService",
    "ContainerService",
    "Sandbox",
    "Container",
    "CreateSandboxRequest",
    "CreateContainerRequest",
]


class SandboxService:
    """Service for managing sandboxes."""

    def __init__(self, client: "RexecClient"):
        self._client = client

    async def list(self) -> list[Sandbox]:
        """
        List all sandboxes for the authenticated user.

        Returns:
            List of Sandbox objects.

        Example:
            sandboxes = await client.sandboxes.list()
            for s in sandboxes:
                print(f"{s.name}: {s.status}")
        """
        data = await self._client._request("GET", "/api/containers")
        # API returns { "containers": [...], "count": N, "limit": M }
        items = data if isinstance(data, list) else (data or {}).get("containers") or []
        return [Sandbox.from_dict(c) for c in items]

    async def get(self, sandbox_id: str) -> Sandbox:
        """
        Get a sandbox by ID.

        Args:
            sandbox_id: The sandbox ID (historically called container id).

        Returns:
            Sandbox object.
        """
        data = await self._client._request("GET", f"/api/containers/{sandbox_id}")
        return Sandbox.from_dict(data)

    async def create(
        self,
        image: str,
        *,
        name: str | None = None,
        environment: dict[str, str] | None = None,
        labels: dict[str, str] | None = None,
    ) -> Sandbox:
        """
        Create a new sandbox.

        Args:
            image: Image alias (e.g., "ubuntu").
            name: Optional sandbox name.
            environment: Optional environment variables.
            labels: Optional labels.

        Returns:
            Created Sandbox object (status may be ``creating``).

        Example:
            sandbox = await client.sandboxes.create(
                image="ubuntu",
                name="my-sandbox",
                environment={"MY_VAR": "value"},
            )
        """
        request = CreateSandboxRequest(
            image=image,
            name=name,
            environment=environment or {},
            labels=labels or {},
        )
        data = await self._client._request("POST", "/api/containers", json=request.to_dict())
        return Sandbox.from_dict(data)

    async def delete(self, sandbox_id: str) -> None:
        """Delete a sandbox."""
        await self._client._request("DELETE", f"/api/containers/{sandbox_id}")

    async def start(self, sandbox_id: str) -> None:
        """Start a stopped sandbox."""
        await self._client._request("POST", f"/api/containers/{sandbox_id}/start")

    async def stop(self, sandbox_id: str) -> None:
        """Stop a running sandbox."""
        await self._client._request("POST", f"/api/containers/{sandbox_id}/stop")


# Backward-compatible alias (same class)
ContainerService = SandboxService
