//! # Rexec Rust SDK
//!
//! Official Rust SDK for [Rexec](https://github.com/PipeOpsHQ/rexec) — AI-native sandboxes.
//!
//! ## Quick Start
//!
//! ```rust,no_run
//! use rexec::{RexecClient, CreateSandboxRequest};
//!
//! #[tokio::main]
//! async fn main() -> Result<(), rexec::Error> {
//!     let client = RexecClient::new(
//!         "https://rexec.sh",
//!         "your-api-token"
//!     );
//!
//!     // Create a sandbox (preferred). client.containers() still works.
//!     let sandbox = client.sandboxes()
//!         .create(CreateSandboxRequest::new("ubuntu").name("my-sandbox"))
//!         .await?;
//!
//!     println!("Created sandbox: {}", sandbox.id);
//!
//!     let mut term = client.terminal().connect(&sandbox.id).await?;
//!     term.write(b"echo hello\n").await?;
//!
//!     client.sandboxes().delete(&sandbox.id).await?;
//!
//!     Ok(())
//! }
//! ```

mod client;
mod containers;
mod error;
mod files;
mod terminal;
mod types;

pub use client::RexecClient;
pub use containers::SandboxService;
#[allow(deprecated)]
pub use containers::ContainerService;
pub use error::Error;
pub use files::FileService;
pub use terminal::{Terminal, TerminalService};
#[allow(deprecated)]
pub use types::*;
