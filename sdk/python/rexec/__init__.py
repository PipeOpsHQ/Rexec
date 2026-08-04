"""
Rexec Python SDK — official client for AI-native sandboxes.

Example usage:
    from rexec import RexecClient

    async with RexecClient("https://rexec.sh", "your-token") as client:
        sandbox = await client.sandboxes.create(image="ubuntu")
        # Legacy alias still works: client.containers
        async with client.terminal.connect(sandbox.id) as term:
            await term.write(b"echo hello\\n")
            async for data in term:
                print(data.decode())
"""

from rexec.client import RexecClient
from rexec.containers import (
    Container,
    ContainerService,
    CreateContainerRequest,
    CreateSandboxRequest,
    ExecResult,
    Sandbox,
    SandboxService,
)
from rexec.exceptions import RexecError, RexecAPIError, RexecConnectionError
from rexec.files import FileInfo, FileService
from rexec.terminal import Terminal, TerminalService

__version__ = "1.1.0"
__all__ = [
    "RexecClient",
    "Sandbox",
    "SandboxService",
    "CreateSandboxRequest",
    "ExecResult",
    # Backward-compatible aliases
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
