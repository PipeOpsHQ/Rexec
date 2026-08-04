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

    image: str
    name: Optional[str] = None
    environment: dict[str, str] = field(default_factory=dict)
    labels: dict[str, str] = field(default_factory=dict)

    def to_dict(self) -> dict:
        """Convert to API request dict."""
        data: dict = {"image": self.image}
        if self.name:
            data["name"] = self.name
        if self.environment:
            data["environment"] = self.environment
        if self.labels:
            data["labels"] = self.labels
        return data


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
