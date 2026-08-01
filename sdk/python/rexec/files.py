"""File service for managing files in containers."""

from typing import TYPE_CHECKING
from urllib.parse import quote

from rexec.types import FileInfo

if TYPE_CHECKING:
    from rexec.client import RexecClient

__all__ = ["FileService", "FileInfo"]


class FileService:
    """Service for file operations in containers."""

    def __init__(self, client: "RexecClient"):
        self._client = client

    async def list(self, container_id: str, path: str = "/") -> list[FileInfo]:
        """
        List files in a container directory.

        Args:
            container_id: The container ID.
            path: Directory path to list (default: "/").

        Returns:
            List of FileInfo objects.

        Example:
            files = await client.files.list(container.id, "/home")
            for f in files:
                print(f"{f.name} - {'dir' if f.is_dir else 'file'}")
        """
        encoded_path = quote(path, safe="")
        data = await self._client._request(
            "GET", f"/api/containers/{container_id}/files/list?path={encoded_path}"
        )
        return [FileInfo.from_dict(f) for f in data]

    async def download(self, container_id: str, path: str) -> bytes:
        """
        Download a file from a container.

        Args:
            container_id: The container ID.
            path: Path to the file.

        Returns:
            File contents as bytes.

        Example:
            content = await client.files.download(container.id, "/etc/passwd")
            print(content.decode())
        """
        encoded_path = quote(path, safe="")
        return await self._client._request_bytes(
            "GET", f"/api/containers/{container_id}/files?path={encoded_path}"
        )

    async def upload(self, container_id: str, path: str, content: bytes) -> None:
        """
        Upload a file to a container.

        Args:
            container_id: The container ID.
            path: Destination path.
            content: File contents as bytes.

        Example:
            await client.files.upload(
                container.id,
                "/home/script.py",
                b"print('hello')"
            )
        """
        await self._client._request(
            "POST",
            f"/api/containers/{container_id}/files",
            data=content,
            params={"path": path},
        )

    async def mkdir(self, container_id: str, path: str) -> None:
        """
        Create a directory in a container.

        Args:
            container_id: The container ID.
            path: Directory path to create.

        Example:
            await client.files.mkdir(container.id, "/home/mydir")
        """
        await self._client._request(
            "POST", f"/api/containers/{container_id}/files/mkdir", json={"path": path}
        )

    async def delete(self, container_id: str, path: str) -> None:
        """
        Delete a file or directory from a container.

        Args:
            container_id: The container ID.
            path: Path to delete.
        """
        encoded_path = quote(path, safe="")
        await self._client._request(
            "DELETE", f"/api/containers/{container_id}/files?path={encoded_path}"
        )
