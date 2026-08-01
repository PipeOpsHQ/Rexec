"""Main Rexec client."""

from typing import Any, Optional
from urllib.parse import urljoin, urlparse

import httpx

from rexec.containers import ContainerService
from rexec.exceptions import RexecAPIError, RexecConnectionError
from rexec.files import FileService
from rexec.terminal import TerminalService

__all__ = ["RexecClient"]


class RexecClient:
    """
    Main client for interacting with Rexec API.

    Example:
        async with RexecClient("https://your-instance.com", "your-token") as client:
            containers = await client.containers.list()
            container = await client.containers.create(image="ubuntu:24.04")

            async with client.terminal.connect(container.id) as term:
                await term.write(b"echo hello\\n")
                async for data in term:
                    print(data.decode(), end="")
    """

    def __init__(
        self,
        base_url: str,
        token: str,
        *,
        timeout: float = 30.0,
        http_client: Optional[httpx.AsyncClient] = None,
    ):
        """
        Initialize the Rexec client.

        Args:
            base_url: Base URL of your Rexec instance.
            token: API token for authentication.
            timeout: Request timeout in seconds (default: 30).
            http_client: Optional custom httpx.AsyncClient.
        """
        self._base_url = base_url.rstrip("/")
        self._token = token
        self._timeout = timeout
        self._own_client = http_client is None
        self._http = http_client or httpx.AsyncClient(
            timeout=timeout,
            headers={
                "Authorization": f"Bearer {token}",
                "Accept": "application/json",
            },
        )

        # Initialize services
        self.containers = ContainerService(self)
        self.files = FileService(self)
        self.terminal = TerminalService(self)

    async def __aenter__(self) -> "RexecClient":
        """Enter async context."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        """Exit async context and close HTTP client."""
        await self.close()

    async def close(self) -> None:
        """Close the HTTP client."""
        if self._own_client:
            await self._http.aclose()

    def _ws_url(self, path: str) -> str:
        """Convert HTTP URL to WebSocket URL."""
        parsed = urlparse(self._base_url)
        ws_scheme = "wss" if parsed.scheme == "https" else "ws"
        return f"{ws_scheme}://{parsed.netloc}{path}"

    async def _request(
        self,
        method: str,
        path: str,
        *,
        json: Optional[dict] = None,
        data: Optional[bytes] = None,
        params: Optional[dict] = None,
    ) -> Any:
        """
        Make an authenticated API request.

        Args:
            method: HTTP method.
            path: API path.
            json: Optional JSON body.
            data: Optional raw body.
            params: Optional query parameters.

        Returns:
            Parsed JSON response or None for empty responses.

        Raises:
            RexecAPIError: For API errors.
            RexecConnectionError: For network errors.
        """
        url = urljoin(self._base_url + "/", path.lstrip("/"))

        try:
            response = await self._http.request(
                method,
                url,
                json=json,
                content=data,
                params=params,
            )
        except httpx.RequestError as e:
            raise RexecConnectionError(f"Request failed: {e}")

        if response.status_code >= 400:
            try:
                error_data = response.json()
                message = error_data.get("error", error_data.get("message", "Unknown error"))
            except Exception:
                message = response.text or "Unknown error"

            raise RexecAPIError(response.status_code, message)

        if response.status_code == 204 or not response.content:
            return None

        return response.json()

    async def _request_bytes(self, method: str, path: str) -> bytes:
        """
        Make an API request and return raw bytes.

        Args:
            method: HTTP method.
            path: API path.

        Returns:
            Raw response bytes.
        """
        url = urljoin(self._base_url + "/", path.lstrip("/"))

        try:
            response = await self._http.request(method, url)
        except httpx.RequestError as e:
            raise RexecConnectionError(f"Request failed: {e}")

        if response.status_code >= 400:
            raise RexecAPIError(response.status_code, "Download failed")

        return response.content
