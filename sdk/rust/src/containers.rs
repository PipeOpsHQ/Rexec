//! Container service for managing sandboxed environments.

use reqwest::Method;
use std::sync::Arc;

use crate::client::ClientInner;
use crate::error::Result;
use crate::types::{CreateSandboxRequest, Sandbox};

/// Service for managing sandboxes.
pub struct SandboxService {
    client: Arc<ClientInner>,
}

impl SandboxService {
    pub(crate) fn new(client: Arc<ClientInner>) -> Self {
        Self { client }
    }

    /// List all containers for the authenticated user.
    ///
    /// # Example
    ///
    /// ```rust,no_run
    /// # use rexec::RexecClient;
    /// # async fn example() -> Result<(), rexec::Error> {
    /// let client = RexecClient::new("https://example.com", "token");
    /// let containers = client.sandboxes().list().await?;
    /// for c in containers {
    ///     println!("{}: {}", c.name, c.status);
    /// }
    /// # Ok(())
    /// # }
    /// ```
    pub async fn list(&self) -> Result<Vec<Sandbox>> {
        // Production API: { "containers": [...], "count": N, "limit": M }
        // `containers` may be null when empty.
        #[derive(serde::Deserialize)]
        struct ListResponse {
            #[serde(default)]
            containers: Option<Vec<Sandbox>>,
        }
        let resp: ListResponse = self.client.request(Method::GET, "/api/containers").await?;
        Ok(resp.containers.unwrap_or_default())
    }

    /// Get a container by ID.
    ///
    /// # Arguments
    ///
    /// * `id` - Container ID
    pub async fn get(&self, id: &str) -> Result<Sandbox> {
        self.client
            .request(Method::GET, &format!("/api/containers/{}", id))
            .await
    }

    /// Create a new container.
    ///
    /// # Example
    ///
    /// ```rust,no_run
    /// # use rexec::{RexecClient, CreateSandboxRequest};
    /// # async fn example() -> Result<(), rexec::Error> {
    /// let client = RexecClient::new("https://example.com", "token");
    /// let container = client.sandboxes()
    ///     .create(CreateSandboxRequest::new("ubuntu")
    ///         .name("my-sandbox")
    ///         .env("MY_VAR", "value"))
    ///     .await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn create(&self, request: CreateSandboxRequest) -> Result<Sandbox> {
        self.client
            .request_with_body(Method::POST, "/api/containers", &request)
            .await
    }

    /// Delete a container.
    ///
    /// # Arguments
    ///
    /// * `id` - Container ID to delete
    pub async fn delete(&self, id: &str) -> Result<()> {
        self.client
            .request_empty(Method::DELETE, &format!("/api/containers/{}", id))
            .await
    }

    /// Start a stopped container.
    ///
    /// # Arguments
    ///
    /// * `id` - Container ID to start
    pub async fn start(&self, id: &str) -> Result<()> {
        self.client
            .request_empty(Method::POST, &format!("/api/containers/{}/start", id))
            .await
    }

    /// Stop a running container.
    ///
    /// # Arguments
    ///
    /// * `id` - Container ID to stop
    pub async fn stop(&self, id: &str) -> Result<()> {
        self.client
            .request_empty(Method::POST, &format!("/api/containers/{}/stop", id))
            .await
    }
}

/// Deprecated alias for [`SandboxService`].
#[deprecated(since = "1.1.0", note = "use SandboxService")]
pub type ContainerService = SandboxService;
