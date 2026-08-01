"""
Rexec Python SDK - Official SDK for Rexec Terminal as a Service.

Example usage:
    from rexec import RexecClient

    async with RexecClient("https://your-instance.com", "your-token") as client:
        container = await client.containers.create(image="ubuntu:24.04")
        async with client.terminal.connect(container.id) as term:
            await term.write(b"echo hello\\n")
            async for data in term:
                print(data.decode())
"""

from rexec.client import RexecClient
from rexec.containers import Container, ContainerService, CreateContainerRequest
from rexec.exceptions import RexecError, RexecAPIError, RexecConnectionError
from rexec.files import FileInfo, FileService
from rexec.terminal import Terminal, TerminalService

__version__ = "1.0.1"
__all__ = [
    "RexecClient",
    "Container",
    "ContainerService", 
    "CreateContainerRequest",
    "FileInfo",
    "FileService",
    "Terminal",
    "TerminalService",
    "RexecError",
    "RexecAPIError",
    "RexecConnectionError",
]
