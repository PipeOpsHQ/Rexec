"""Sandbox service for managing isolated environments.

Wire protocol remains ``/api/containers``. Public names prefer *sandbox*;
``Container*`` aliases remain for backward compatibility.
"""

from typing import TYPE_CHECKING, List, Optional

from rexec.types import (
    Container,
    CreateContainerRequest,
    CreateSandboxRequest,
    ExecResult,
    Sandbox,
    SandboxSnapshot,
    SandboxTemplate,
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
    "ExecResult",
    "SandboxTemplate",
    "SandboxSnapshot",
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
        image: Optional[str] = None,
        *,
        name: Optional[str] = None,
        environment: Optional[dict] = None,
        labels: Optional[dict] = None,
        template_id: Optional[str] = None,
        snapshot_id: Optional[str] = None,
        network_mode: Optional[str] = None,
        egress_allow: Optional[List[str]] = None,
        custom_image: Optional[str] = None,
        idle_timeout_seconds: Optional[int] = None,
        max_lifetime_seconds: Optional[int] = None,
        prefer_warm: Optional[bool] = None,
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
            custom_image=custom_image,
            template_id=template_id,
            snapshot_id=snapshot_id,
            network_mode=network_mode,
            egress_allow=egress_allow,
            idle_timeout_seconds=idle_timeout_seconds,
            max_lifetime_seconds=max_lifetime_seconds,
            prefer_warm=prefer_warm,
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

    async def exec(
        self,
        sandbox_id: str,
        command: Optional[str] = None,
        *,
        cmd: Optional[List[str]] = None,
        workdir: Optional[str] = None,
        env: Optional[List[str]] = None,
        user: Optional[str] = None,
        timeout_seconds: Optional[int] = None,
    ) -> ExecResult:
        """
        Run a non-interactive command in a running sandbox.

        Wire: ``POST /api/containers/:id/exec``.

        Args:
            sandbox_id: Sandbox id (Docker or DB id).
            command: Shell string run via ``sh -c`` (when ``cmd`` is not set).
            cmd: Argv vector; takes precedence over ``command`` when non-empty.
            workdir: Working directory inside the sandbox.
            env: Extra env as ``KEY=VALUE`` strings.
            user: User to run as inside the sandbox.
            timeout_seconds: Timeout (default 60, max 300).

        Returns:
            ExecResult with stdout, stderr, exit_code, etc.

        Example:
            r = await client.sandboxes.exec(sid, "echo hello")
            print(r.stdout, r.exit_code)
            r = await client.sandboxes.exec(sid, cmd=["uname", "-a"])
        """
        body: dict = {}
        if cmd:
            body["cmd"] = cmd
        elif command is not None:
            body["command"] = command
        else:
            raise ValueError("exec requires command or cmd")
        if workdir:
            body["workdir"] = workdir
        if env:
            body["env"] = env
        if user:
            body["user"] = user
        if timeout_seconds is not None:
            body["timeout_seconds"] = timeout_seconds

        data = await self._client._request(
            "POST", f"/api/containers/{sandbox_id}/exec", json=body
        )
        return ExecResult.from_dict(data or {})

    async def list_templates(self) -> List[SandboxTemplate]:
        """List sandbox templates for the current user."""
        data = await self._client._request("GET", "/api/templates")
        items = data if isinstance(data, list) else (data or {}).get("templates") or []
        return [SandboxTemplate.from_dict(t) for t in items]

    async def create_template(
        self,
        name: str,
        from_sandbox_id: str,
        *,
        description: Optional[str] = None,
    ) -> SandboxTemplate:
        """Commit a running sandbox to a reusable template image."""
        body: dict = {"name": name, "from_sandbox_id": from_sandbox_id}
        if description:
            body["description"] = description
        data = await self._client._request("POST", "/api/templates", json=body)
        return SandboxTemplate.from_dict(data or {})

    async def get_template(self, template_id: str) -> SandboxTemplate:
        data = await self._client._request("GET", f"/api/templates/{template_id}")
        return SandboxTemplate.from_dict(data or {})

    async def delete_template(self, template_id: str) -> None:
        await self._client._request("DELETE", f"/api/templates/{template_id}")

    async def snapshot(
        self,
        sandbox_id: str,
        *,
        name: Optional[str] = None,
        description: Optional[str] = None,
    ) -> SandboxSnapshot:
        """Create a point-in-time filesystem snapshot of a sandbox."""
        body: dict = {}
        if name:
            body["name"] = name
        if description:
            body["description"] = description
        data = await self._client._request(
            "POST", f"/api/containers/{sandbox_id}/snapshot", json=body
        )
        return SandboxSnapshot.from_dict(data or {})

    async def list_snapshots(self) -> List[SandboxSnapshot]:
        data = await self._client._request("GET", "/api/snapshots")
        items = data if isinstance(data, list) else (data or {}).get("snapshots") or []
        return [SandboxSnapshot.from_dict(s) for s in items]

    async def get_snapshot(self, snapshot_id: str) -> SandboxSnapshot:
        data = await self._client._request("GET", f"/api/snapshots/{snapshot_id}")
        return SandboxSnapshot.from_dict(data or {})

    async def delete_snapshot(self, snapshot_id: str) -> None:
        await self._client._request("DELETE", f"/api/snapshots/{snapshot_id}")

    async def fork(
        self,
        sandbox_id: str,
        *,
        name: Optional[str] = None,
        network_mode: Optional[str] = None,
        egress_allow: Optional[List[str]] = None,
        idle_timeout_seconds: Optional[int] = None,
        max_lifetime_seconds: Optional[int] = None,
        save_snapshot: bool = False,
        snapshot_name: Optional[str] = None,
        labels: Optional[dict] = None,
    ) -> Sandbox:
        """Commit current FS and create a new sandbox from it."""
        body: dict = {}
        if name:
            body["name"] = name
        if network_mode:
            body["network_mode"] = network_mode
        if egress_allow:
            body["egress_allow"] = egress_allow
        if idle_timeout_seconds is not None:
            body["idle_timeout_seconds"] = idle_timeout_seconds
        if max_lifetime_seconds is not None:
            body["max_lifetime_seconds"] = max_lifetime_seconds
        if save_snapshot:
            body["save_snapshot"] = True
        if snapshot_name:
            body["snapshot_name"] = snapshot_name
        if labels:
            body["labels"] = labels
        data = await self._client._request(
            "POST", f"/api/containers/{sandbox_id}/fork", json=body
        )
        return Sandbox.from_dict(data or {})


# Backward-compatible alias (same class)
ContainerService = SandboxService
