"""Type definitions for Rexec SDK."""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional


@dataclass
class Sandbox:
    """Represents a Rexec sandbox (isolated Linux environment)."""

    id: str
    name: str
    image: str
    status: str
    created_at: datetime
    started_at: Optional[datetime] = None
    labels: dict[str, str] = field(default_factory=dict)
    environment: dict[str, str] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, data: dict) -> "Sandbox":
        """Create Sandbox from API response dict."""
        return cls(
            id=data["id"],
            name=data.get("name", ""),
            image=data.get("image", ""),
            status=data.get("status", "unknown"),
            created_at=datetime.fromisoformat(data["created_at"].replace("Z", "+00:00"))
            if data.get("created_at")
            else datetime.now(),
            started_at=datetime.fromisoformat(data["started_at"].replace("Z", "+00:00"))
            if data.get("started_at")
            else None,
            labels=data.get("labels") or {},
            environment=data.get("environment") or {},
        )


# Backward-compatible alias
Container = Sandbox


@dataclass
class CreateSandboxRequest:
    """Request to create a new sandbox. Prefer image aliases (e.g. ubuntu)."""

    image: Optional[str] = None
    name: Optional[str] = None
    custom_image: Optional[str] = None
    template_id: Optional[str] = None
    network_mode: Optional[str] = None  # default | none | restricted
    egress_allow: Optional[list] = None  # extra hosts for restricted mode
    idle_timeout_seconds: Optional[int] = None
    max_lifetime_seconds: Optional[int] = None
    prefer_warm: Optional[bool] = None
    environment: dict[str, str] = field(default_factory=dict)
    labels: dict[str, str] = field(default_factory=dict)

    def to_dict(self) -> dict:
        """Convert to API request dict."""
        data: dict = {}
        if self.image:
            data["image"] = self.image
        if self.name:
            data["name"] = self.name
        if self.custom_image:
            data["custom_image"] = self.custom_image
        if self.template_id:
            data["template_id"] = self.template_id
        if self.network_mode:
            data["network_mode"] = self.network_mode
        if self.egress_allow:
            data["egress_allow"] = self.egress_allow
        if self.idle_timeout_seconds is not None:
            data["idle_timeout_seconds"] = self.idle_timeout_seconds
        if self.max_lifetime_seconds is not None:
            data["max_lifetime_seconds"] = self.max_lifetime_seconds
        if self.prefer_warm is not None:
            data["prefer_warm"] = self.prefer_warm
        if self.environment:
            data["environment"] = self.environment
        if self.labels:
            data["labels"] = self.labels
        return data


@dataclass
class SandboxTemplate:
    """Saved sandbox template (committed image)."""

    id: str
    user_id: str
    name: str
    docker_image: str
    description: str = ""
    base_image: str = ""
    source_container_id: str = ""
    status: str = "ready"
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict) -> "SandboxTemplate":
        return cls(
            id=data.get("id", ""),
            user_id=data.get("user_id", ""),
            name=data.get("name", ""),
            docker_image=data.get("docker_image", ""),
            description=data.get("description") or "",
            base_image=data.get("base_image") or "",
            source_container_id=data.get("source_container_id") or "",
            status=data.get("status") or "ready",
            created_at=data.get("created_at"),
            updated_at=data.get("updated_at"),
        )


# Backward-compatible alias
CreateContainerRequest = CreateSandboxRequest


@dataclass
class ExecResult:
    """Result of a non-interactive sandbox exec."""

    stdout: str
    stderr: str
    output: str
    exit_code: int
    duration_ms: Optional[int] = None
    truncated: bool = False
    command: Optional[str] = None
    cmd: Optional[list] = None

    @classmethod
    def from_dict(cls, data: dict) -> "ExecResult":
        return cls(
            stdout=data.get("stdout") or "",
            stderr=data.get("stderr") or "",
            output=data.get("output") or "",
            exit_code=int(data.get("exit_code") if data.get("exit_code") is not None else 0),
            duration_ms=data.get("duration_ms"),
            truncated=bool(data.get("truncated")),
            command=data.get("command"),
            cmd=data.get("cmd"),
        )


@dataclass
class FileInfo:
    """File or directory metadata."""

    name: str
    path: str
    size: int
    mode: str
    mod_time: datetime
    is_dir: bool

    @classmethod
    def from_dict(cls, data: dict) -> "FileInfo":
        """Create FileInfo from API response dict."""
        return cls(
            name=data["name"],
            path=data["path"],
            size=data.get("size", 0),
            mode=data.get("mode", ""),
            mod_time=datetime.fromisoformat(data["mod_time"].replace("Z", "+00:00"))
            if data.get("mod_time")
            else datetime.now(),
            is_dir=data.get("is_dir", False),
        )
