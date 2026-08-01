"""Container service for managing sandboxed environments."""

from typing import TYPE_CHECKING

from rexec.types import Container, CreateContainerRequest

if TYPE_CHECKING:
    from rexec.client import RexecClient

# Re-export types for convenience
__all__ = ["ContainerService", "Container", "CreateContainerRequest"]


class ContainerService:
    """Service for managing containers."""

    def __init__(self, client: "RexecClient"):
        self._client = client

    async def list(self) -> list[Container]:
        """
        List all containers for the authenticated user.

        Returns:
            List of Container objects.

        Example:
            containers = await client.containers.list()
            for c in containers:
                print(f"{c.name}: {c.status}")
        """
        data = await self._client._request("GET", "/api/containers")
        # API returns { "containers": [...], "count": N, "limit": M }
        items = data if isinstance(data, list) else (data or {}).get("containers") or []
        return [Container.from_dict(c) for c in items]

    async def get(self, container_id: str) -> Container:
        """
        Get a container by ID.

        Args:
            container_id: The container ID.

        Returns:
            Container object.

        Raises:
            RexecNotFoundError: If container doesn't exist.
        """
        data = await self._client._request("GET", f"/api/containers/{container_id}")
        return Container.from_dict(data)

    async def create(
        self,
        image: str,
        *,
        name: str | None = None,
        environment: dict[str, str] | None = None,
        labels: dict[str, str] | None = None,
    ) -> Container:
        """
        Create a new container.

        Args:
            image: Docker image to use (e.g., "ubuntu:24.04").
            name: Optional container name.
            environment: Optional environment variables.
            labels: Optional labels.

        Returns:
            Created Container object.

        Example:
            container = await client.containers.create(
                image="ubuntu:24.04",
                name="my-sandbox",
                environment={"MY_VAR": "value"}
            )
        """
        request = CreateContainerRequest(
            image=image,
            name=name,
            environment=environment or {},
            labels=labels or {},
        )
        data = await self._client._request("POST", "/api/containers", json=request.to_dict())
        return Container.from_dict(data)

    async def delete(self, container_id: str) -> None:
        """
        Delete a container.

        Args:
            container_id: The container ID to delete.
        """
        await self._client._request("DELETE", f"/api/containers/{container_id}")

    async def start(self, container_id: str) -> None:
        """
        Start a stopped container.

        Args:
            container_id: The container ID to start.
        """
        await self._client._request("POST", f"/api/containers/{container_id}/start")

    async def stop(self, container_id: str) -> None:
        """
        Stop a running container.

        Args:
            container_id: The container ID to stop.
        """
        await self._client._request("POST", f"/api/containers/{container_id}/stop")
