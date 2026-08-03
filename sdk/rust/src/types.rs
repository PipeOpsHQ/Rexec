//! Type definitions for Rexec SDK.

use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Sandbox status values (string form also used on the wire).
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum SandboxStatus {
    Running,
    Stopped,
    Creating,
    Error,
    #[serde(other)]
    Unknown,
}

/// Deprecated alias for [`SandboxStatus`].
#[deprecated(since = "1.1.0", note = "use SandboxStatus")]
pub type ContainerStatus = SandboxStatus;

/// Represents a Rexec sandbox (isolated Linux environment).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sandbox {
    /// Sandbox ID.
    pub id: String,
    /// Sandbox name.
    #[serde(default)]
    pub name: String,
    /// Image alias.
    #[serde(default)]
    pub image: String,
    /// Current status.
    #[serde(default)]
    pub status: String,
    /// Creation timestamp.
    #[serde(default)]
    pub created_at: String,
    /// Start timestamp (if running).
    #[serde(default)]
    pub started_at: Option<String>,
    /// Labels.
    #[serde(default)]
    pub labels: HashMap<String, String>,
    /// Environment variables.
    #[serde(default)]
    pub environment: HashMap<String, String>,
}

/// Deprecated alias for [`Sandbox`].
#[deprecated(since = "1.1.0", note = "use Sandbox")]
pub type Container = Sandbox;

/// Request to create a new sandbox. Prefer image aliases (e.g. `ubuntu`).
#[derive(Debug, Clone, Serialize, Default)]
pub struct CreateSandboxRequest {
    /// Image alias to use.
    pub image: String,
    /// Optional sandbox name.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// Environment variables.
    #[serde(skip_serializing_if = "HashMap::is_empty", default)]
    pub environment: HashMap<String, String>,
    /// Labels.
    #[serde(skip_serializing_if = "HashMap::is_empty", default)]
    pub labels: HashMap<String, String>,
}

/// Deprecated alias for [`CreateSandboxRequest`].
#[deprecated(since = "1.1.0", note = "use CreateSandboxRequest")]
pub type CreateContainerRequest = CreateSandboxRequest;

impl CreateSandboxRequest {
    /// Create a new request with the given image alias.
    pub fn new(image: impl Into<String>) -> Self {
        Self {
            image: image.into(),
            ..Default::default()
        }
    }

    /// Set the sandbox name.
    pub fn name(mut self, name: impl Into<String>) -> Self {
        self.name = Some(name.into());
        self
    }

    /// Add an environment variable.
    pub fn env(mut self, key: impl Into<String>, value: impl Into<String>) -> Self {
        self.environment.insert(key.into(), value.into());
        self
    }

    /// Add a label.
    pub fn label(mut self, key: impl Into<String>, value: impl Into<String>) -> Self {
        self.labels.insert(key.into(), value.into());
        self
    }
}

/// File or directory metadata.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileInfo {
    /// File name.
    pub name: String,
    /// Full path.
    pub path: String,
    /// File size in bytes.
    pub size: u64,
    /// File mode (permissions).
    pub mode: String,
    /// Modification time.
    pub mod_time: String,
    /// Whether this is a directory.
    pub is_dir: bool,
}

/// Terminal resize message.
#[derive(Debug, Clone, Serialize)]
pub struct ResizeMessage {
    #[serde(rename = "type")]
    pub msg_type: String,
    pub cols: u16,
    pub rows: u16,
}

impl ResizeMessage {
    pub fn new(cols: u16, rows: u16) -> Self {
        Self {
            msg_type: "resize".into(),
            cols,
            rows,
        }
    }
}
