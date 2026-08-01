use rexec::{CreateContainerRequest, RexecClient};
use std::env;
use std::time::{SystemTime, UNIX_EPOCH};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let url = env::var("URL")?;
    let token = env::var("TOKEN")?;
    let client = RexecClient::new(url, token);
    println!("[rs] list...");
    let before = client.containers().list().await?;
    println!("[rs] list count {}", before.len());
    let name = format!(
        "rs-e2e-{}",
        SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs()
    );
    println!("[rs] create...");
    let c = client
        .containers()
        .create(CreateContainerRequest::new("ubuntu").name(name))
        .await?;
    println!("[rs] created {} {} {}", c.id, c.status, c.image);
    let after = client.containers().list().await?;
    println!("[rs] list count {}", after.len());
    let got = client.containers().get(&c.id).await?;
    println!("[rs] get {} {}", got.id, got.status);
    client.containers().delete(&c.id).await?;
    println!("[rs] OK");
    Ok(())
}
