"""Custom exceptions for Rexec SDK."""

from typing import Optional


class RexecError(Exception):
    """Base exception for all Rexec errors."""

    pass


class RexecAPIError(RexecError):
    """API error with status code and message."""

    def __init__(self, status_code: int, message: str, details: Optional[dict] = None):
        self.status_code = status_code
        self.message = message
        self.details = details or {}
        super().__init__(f"API error {status_code}: {message}")


class RexecConnectionError(RexecError):
    """Connection error (network, timeout, etc.)."""

    pass


class RexecAuthError(RexecAPIError):
    """Authentication error (401/403)."""

    def __init__(self, message: str = "Authentication failed"):
        super().__init__(401, message)


class RexecNotFoundError(RexecAPIError):
    """Resource not found (404)."""

    def __init__(self, resource: str, resource_id: str):
        super().__init__(404, f"{resource} '{resource_id}' not found")
        self.resource = resource
        self.resource_id = resource_id


class RexecValidationError(RexecAPIError):
    """Validation error (400)."""

    def __init__(self, message: str, errors: Optional[dict] = None):
        super().__init__(400, message, errors)
        self.errors = errors or {}
