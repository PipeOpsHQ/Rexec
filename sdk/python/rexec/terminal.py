"""Terminal service for WebSocket connections to containers."""

import asyncio
import json
from typing import TYPE_CHECKING, AsyncIterator, Optional

import websockets
from websockets.client import WebSocketClientProtocol

from rexec.exceptions import RexecConnectionError

if TYPE_CHECKING:
    from rexec.client import RexecClient

__all__ = ["TerminalService", "Terminal"]


class Terminal:
    """
    WebSocket terminal connection to a container.

    Use as an async context manager or manually manage the connection.

    Example:
        async with client.terminal.connect(container.id) as term:
            await term.write(b"ls -la\\n")
            async for data in term:
                print(data.decode(), end="")
    """

    def __init__(self, ws: WebSocketClientProtocol):
        self._ws = ws
        self._closed = False

    async def write(self, data: bytes | str) -> None:
        """
        Send data to the terminal.

        Args:
            data: Data to send (bytes or string).
        """
        if self._closed:
            raise RexecConnectionError("Terminal connection is closed")

        if isinstance(data, str):
            data = data.encode()
        await self._ws.send(data)

    async def read(self) -> bytes:
        """
        Read data from the terminal.

        Returns:
            Bytes received from the terminal.

        Raises:
            RexecConnectionError: If connection is closed.
        """
        if self._closed:
            raise RexecConnectionError("Terminal connection is closed")

        try:
            data = await self._ws.recv()
            if isinstance(data, str):
                return data.encode()
            return data
        except websockets.ConnectionClosed:
            self._closed = True
            raise RexecConnectionError("Terminal connection closed")

    async def resize(self, cols: int, rows: int) -> None:
        """
        Resize the terminal.

        Args:
            cols: Number of columns.
            rows: Number of rows.
        """
        msg = json.dumps({"type": "resize", "cols": cols, "rows": rows})
        await self._ws.send(msg)

    async def close(self) -> None:
        """Close the terminal connection."""
        if not self._closed:
            self._closed = True
            await self._ws.close()

    @property
    def closed(self) -> bool:
        """Check if the connection is closed."""
        return self._closed

    def __aiter__(self) -> AsyncIterator[bytes]:
        """Iterate over terminal output."""
        return self

    async def __anext__(self) -> bytes:
        """Get next chunk of terminal output."""
        try:
            return await self.read()
        except RexecConnectionError:
            raise StopAsyncIteration

    async def __aenter__(self) -> "Terminal":
        """Enter async context."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        """Exit async context."""
        await self.close()


class TerminalService:
    """Service for terminal WebSocket connections."""

    def __init__(self, client: "RexecClient"):
        self._client = client

    async def connect(
        self,
        container_id: str,
        *,
        cols: int = 80,
        rows: int = 24,
        timeout: float = 30.0,
    ) -> Terminal:
        """
        Connect to a container's terminal.

        Args:
            container_id: The container ID.
            cols: Terminal width in columns (default: 80).
            rows: Terminal height in rows (default: 24).
            timeout: Connection timeout in seconds (default: 30).

        Returns:
            Terminal object for reading/writing.

        Example:
            term = await client.terminal.connect(container.id)
            try:
                await term.write(b"echo hello\\n")
                data = await term.read()
                print(data.decode())
            finally:
                await term.close()
        """
        ws_url = self._client._ws_url(f"/ws/terminal/{container_id}")

        try:
            ws = await asyncio.wait_for(
                websockets.connect(
                    ws_url,
                    additional_headers={"Authorization": f"Bearer {self._client._token}"},
                ),
                timeout=timeout,
            )
        except asyncio.TimeoutError:
            raise RexecConnectionError(f"Connection timeout after {timeout}s")
        except Exception as e:
            raise RexecConnectionError(f"Failed to connect: {e}")

        terminal = Terminal(ws)

        # Set initial size
        if cols and rows:
            await terminal.resize(cols, rows)

        return terminal
