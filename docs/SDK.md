# Rexec SDK Documentation

Rexec provides official SDKs for programmatically interacting with your sandboxed environments.

> 📖 **New to Rexec SDKs?** Start with the [Getting Started Guide](SDK_GETTING_STARTED.md) for a step-by-step tutorial.

## Available SDKs

| SDK | Package | Install | Documentation |
|-----|---------|---------|---------------|
| **Go** | `github.com/PipeOpsHQ/rexec-go` | `go get github.com/PipeOpsHQ/rexec-go` | [README](../sdk/go/README.md) |
| **JavaScript/TypeScript** | `@pipeopshq/rexec` | `npm install @pipeopshq/rexec` | [README](../sdk/js/README.md) |
| **Python** | `rexec` (PyPI) | `pip install rexec` | [README](../sdk/python/README.md) |
| **Rust** | `rexec` (crates.io) | `cargo add rexec` | [README](../sdk/rust/README.md) |
| **Ruby** | `rexec` (RubyGems) | `gem install rexec` | [README](../sdk/ruby/README.md) |
| **Java** | `io.pipeops:rexec` (Maven) | [Maven dependency](#java) | [README](../sdk/java/README.md) |
| **C#/.NET** | `Rexec` (NuGet) | `dotnet add package Rexec` | [README](../sdk/dotnet/README.md) |
| **PHP** | `pipeopshq/rexec` (Packagist) | `composer require pipeopshq/rexec` | [README](../sdk/php/README.md) |

## Getting an API Token

Before using the SDK, you need to generate an API token:

1. Log in to your Rexec instance
2. Go to **Settings** → **API Tokens**
3. Click **Generate Token**
4. Copy the token and store it securely

> ⚠️ **Security Note**: Never commit API tokens to version control. Use environment variables or secret management tools.

## Quick Start

### Go

```bash
go get github.com/PipeOpsHQ/rexec-go
```

```go
package main

import (
    "context"
    "fmt"
    "os"

    rexec "github.com/PipeOpsHQ/rexec-go"
)

func main() {
    client := rexec.NewClient(
        os.Getenv("REXEC_URL"),
        os.Getenv("REXEC_TOKEN"),
    )

    ctx := context.Background()

    // Create a sandbox
    container, _ := client.Containers.Create(ctx, &rexec.CreateContainerRequest{
        Image: "ubuntu:24.04",
    })

    // Connect to terminal
    term, _ := client.Terminal.Connect(ctx, container.ID)
    defer term.Close()

    term.Write([]byte("echo 'Hello World'\n"))
}
```

### JavaScript/TypeScript

```bash
npm install @pipeopshq/rexec
```

```typescript
import { RexecClient } from '@pipeopshq/rexec';

const client = new RexecClient({
  baseURL: process.env.REXEC_URL,
  token: process.env.REXEC_TOKEN,
});

// Create a sandbox
const container = await client.containers.create({
  image: 'ubuntu:24.04',
});

// Connect to terminal
const terminal = await client.terminal.connect(container.id);
terminal.write('echo "Hello World"\n');
terminal.onData((data) => console.log(data));
```

### Python

```bash
pip install rexec
```

```python
import asyncio
from rexec import RexecClient

async def main():
    async with RexecClient(
        base_url=os.environ["REXEC_URL"],
        token=os.environ["REXEC_TOKEN"]
    ) as client:
        # Create a sandbox
        container = await client.containers.create("ubuntu:24.04")
        
        # Execute a command
        result = await client.containers.exec(container.id, "echo 'Hello World'")
        print(result.stdout)

asyncio.run(main())
```

### Rust

```bash
cargo add rexec
```

```rust
use rexec::RexecClient;

#[tokio::main]
async fn main() -> Result<(), rexec::Error> {
    let client = RexecClient::new(
        std::env::var("REXEC_URL")?,
        std::env::var("REXEC_TOKEN")?
    )?;

    // Create a sandbox
    let container = client.containers()
        .create("ubuntu:24.04")
        .await?;

    // Connect to terminal
    let mut term = client.terminal()
        .connect(&container.id)
        .await?;

    term.write(b"echo 'Hello World'\n").await?;
    Ok(())
}
```

### Ruby

```bash
gem install rexec
```

```ruby
require 'rexec'

client = Rexec::Client.new(
  ENV['REXEC_URL'],
  ENV['REXEC_TOKEN']
)

# Create a sandbox
container = client.containers.create('ubuntu:24.04')

# Execute a command
result = client.containers.exec(container.id, "echo 'Hello World'")
puts result.stdout
```

### Java

```xml
<dependency>
    <groupId>io.pipeops</groupId>
    <artifactId>rexec</artifactId>
    <version>1.0.0</version>
</dependency>
```

```java
import io.pipeops.rexec.*;

public class Example {
    public static void main(String[] args) throws RexecException {
        RexecClient client = new RexecClient(
            System.getenv("REXEC_URL"),
            System.getenv("REXEC_TOKEN")
        );

        // Create a sandbox
        Container container = client.containers().create("ubuntu:24.04");

        // Execute a command
        ExecResult result = client.containers().exec(
            container.getId(), "echo 'Hello World'"
        );
        System.out.println(result.getStdout());
    }
}
```

### C#/.NET

```bash
dotnet add package Rexec
```

```csharp
using Rexec;

var client = new RexecClient(
    Environment.GetEnvironmentVariable("REXEC_URL"),
    Environment.GetEnvironmentVariable("REXEC_TOKEN")
);

// Create a sandbox
var container = await client.Containers.CreateAsync("ubuntu:24.04");

// Execute a command
var result = await client.Containers.ExecAsync(
    container.Id, "echo 'Hello World'"
);
Console.WriteLine(result.Stdout);
```

### PHP

```bash
composer require pipeopshq/rexec
```

```php
<?php
require 'vendor/autoload.php';

use Rexec\RexecClient;

$client = new RexecClient(
    getenv('REXEC_URL'),
    getenv('REXEC_TOKEN')
);

// Create a sandbox
$container = $client->containers()->create('ubuntu:24.04');

// Execute a command
$result = $client->containers()->exec($container->id, "echo 'Hello World'");
echo $result->stdout;
```

## Core Concepts

### Containers

Containers are isolated Linux environments powered by Docker. Each container:

- Has its own filesystem
- Can run any Linux command
- Is isolated from other containers
- Can be started, stopped, and deleted

### Terminal Sessions

Terminal sessions provide real-time WebSocket connections to containers:

- Full PTY support
- Resizable terminals
- Binary data support for tools like vim, nano, etc.

### Files

The file API allows you to:

- List files and directories
- Upload files to containers
- Download files from containers
- Create directories

## API Endpoints

The SDKs wrap these REST API endpoints:

### Containers

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/containers` | List all containers |
| `POST` | `/api/containers` | Create a container |
| `GET` | `/api/containers/:id` | Get container details |
| `DELETE` | `/api/containers/:id` | Delete a container |
| `POST` | `/api/containers/:id/start` | Start a container |
| `POST` | `/api/containers/:id/stop` | Stop a container |

### Files

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/containers/:id/files/list` | List files |
| `GET` | `/api/containers/:id/files` | Download a file |
| `POST` | `/api/containers/:id/files` | Upload a file |
| `POST` | `/api/containers/:id/files/mkdir` | Create directory |
| `DELETE` | `/api/containers/:id/files` | Delete a file |

### WebSocket

| Endpoint | Description |
|----------|-------------|
| `/ws/terminal/:containerId` | Terminal connection |

## Use Cases

### CI/CD Integration

Run tests in isolated environments with the [JavaScript SDK](../sdk/js/README.md) or [Python SDK](../sdk/python/README.md):

```typescript
// JavaScript/TypeScript
const container = await client.containers.create({
  image: 'node:20',
  environment: { CI: 'true' }
});

await client.containers.start(container.id);
const result = await client.containers.exec(container.id, 'npm install && npm test');

if (result.exitCode !== 0) {
  console.error('Tests failed:', result.stderr);
  process.exit(1);
}

await client.containers.delete(container.id);
```

```python
# Python - Async CI runner
async def run_ci_tests(image: str, commands: list[str]):
    async with RexecClient(base_url, token) as client:
        container = await client.containers.create(image)
        await client.containers.start(container.id)
        
        for cmd in commands:
            result = await client.containers.exec(container.id, cmd)
            if result.exit_code != 0:
                raise Exception(f"Command failed: {cmd}")
        
        await client.containers.delete(container.id)
```

### Remote Development Environments

Provide cloud development environments with the [Go SDK](../sdk/go/README.md):

```go
// Create a full development environment
container, _ := client.Containers.Create(ctx, &rexec.CreateContainerRequest{
    Image: "ubuntu:24.04",
    Name:  fmt.Sprintf("dev-%s", userID),
    Environment: map[string]string{
        "EDITOR": "vim",
        "TERM":   "xterm-256color",
    },
    Labels: map[string]string{
        "user": userID,
        "type": "development",
    },
})

// Pre-install development tools
client.Containers.Exec(ctx, container.ID, []string{
    "apt-get", "update", "&&", 
    "apt-get", "install", "-y", "git", "vim", "curl",
})
```

### Education Platforms

Provide sandboxed coding environments for students with the [JavaScript SDK](../sdk/js/README.md):

```typescript
// Create a container per student with resource limits
async function createStudentSandbox(studentId: string, courseId: string) {
  const container = await client.containers.create({
    image: 'python:3.12',
    name: `student-${studentId}-${courseId}`,
    labels: { 
      student: studentId,
      course: courseId,
      type: 'education'
    }
  });

  // Upload starter code
  await client.files.write(
    container.id,
    '/home/student/assignment.py',
    starterCode
  );

  return container;
}

// Grade submissions
async function gradeSubmission(containerId: string) {
  const result = await client.containers.exec(containerId, 'python -m pytest /home/student/');
  return {
    passed: result.exitCode === 0,
    output: result.stdout,
    errors: result.stderr
  };
}
```

### Automated Testing

Run integration tests in clean environments with the [Go SDK](../sdk/go/README.md) or [Rust SDK](../sdk/rust/README.md):

```go
func TestDatabaseMigrations(t *testing.T) {
    container, err := client.Containers.Create(ctx, &rexec.CreateContainerRequest{
        Image: "postgres:16",
        Environment: map[string]string{
            "POSTGRES_PASSWORD": "test",
        },
    })
    require.NoError(t, err)
    defer client.Containers.Delete(ctx, container.ID)

    client.Containers.Start(ctx, container.ID)
    
    // Wait for postgres to be ready
    time.Sleep(5 * time.Second)
    
    // Run migrations
    result, _ := client.Containers.Exec(ctx, container.ID, []string{
        "psql", "-U", "postgres", "-f", "/migrations/001_init.sql",
    })
    
    assert.Equal(t, 0, result.ExitCode)
}
```

```rust
// Rust - Parallel test execution
async fn run_tests_in_parallel(test_cases: Vec<TestCase>) -> Vec<TestResult> {
    let handles: Vec<_> = test_cases.into_iter().map(|tc| {
        let client = client.clone();
        tokio::spawn(async move {
            let container = client.containers()
                .create(&tc.image)
                .await?;
            
            client.containers().start(&container.id).await?;
            
            let result = client.containers()
                .exec(&container.id, &tc.command)
                .await?;
            
            client.containers().delete(&container.id).await?;
            
            Ok(TestResult {
                name: tc.name,
                passed: result.exit_code == 0,
            })
        })
    }).collect();
    
    futures::future::join_all(handles).await
}
```

### Code Execution Service

Build a code execution API with the [Java SDK](../sdk/java/README.md) or [C# SDK](../sdk/dotnet/README.md):

```java
// Java - Safe code execution endpoint
@PostMapping("/execute")
public ResponseEntity<ExecutionResult> executeCode(@RequestBody CodeRequest request) {
    try {
        Container container = client.containers().create(
            new CreateContainerRequest(request.getLanguageImage())
                .addLabel("type", "execution")
        );
        
        client.containers().start(container.getId());
        
        // Write user code to file
        client.files().write(
            container.getId(),
            "/tmp/code." + request.getExtension(),
            request.getCode()
        );
        
        // Execute with timeout
        ExecResult result = client.containers().exec(
            container.getId(),
            request.getRunCommand()
        );
        
        client.containers().delete(container.getId());
        
        return ResponseEntity.ok(new ExecutionResult(
            result.getStdout(),
            result.getStderr(),
            result.getExitCode()
        ));
    } catch (RexecException e) {
        return ResponseEntity.status(500).body(new ExecutionResult(e.getMessage()));
    }
}
```

```csharp
// C# - Code execution with cancellation
public async Task<ExecutionResult> ExecuteCodeAsync(
    string code, 
    string language,
    CancellationToken cancellationToken = default)
{
    var imageMap = new Dictionary<string, string> {
        ["python"] = "python:3.12",
        ["node"] = "node:20",
        ["go"] = "golang:1.22",
    };

    var container = await client.Containers.CreateAsync(
        imageMap[language], 
        cancellationToken
    );

    try {
        await client.Containers.StartAsync(container.Id, cancellationToken);
        await client.Files.WriteAsync(container.Id, "/tmp/code", code, cancellationToken);
        
        var result = await client.Containers.ExecAsync(
            container.Id,
            GetRunCommand(language),
            cancellationToken
        );
        
        return new ExecutionResult(result.Stdout, result.Stderr, result.ExitCode);
    }
    finally {
        await client.Containers.DeleteAsync(container.Id);
    }
}
```

### DevOps Automation

Automate infrastructure tasks with the [Ruby SDK](../sdk/ruby/README.md) or [PHP SDK](../sdk/php/README.md):

```ruby
# Ruby - Automated backup script
require 'rexec'

client = Rexec::Client.new(ENV['REXEC_URL'], ENV['REXEC_TOKEN'])

# Create a container with backup tools
container = client.containers.create('ubuntu:24.04', 
  name: 'backup-worker',
  environment: {
    'AWS_ACCESS_KEY_ID' => ENV['AWS_ACCESS_KEY_ID'],
    'AWS_SECRET_ACCESS_KEY' => ENV['AWS_SECRET_ACCESS_KEY']
  }
)

client.containers.start(container.id)

# Install AWS CLI and run backup
client.containers.exec(container.id, 'apt-get update && apt-get install -y awscli')
result = client.containers.exec(container.id, 'aws s3 sync /data s3://my-bucket/backup/')

if result.exit_code == 0
  puts "Backup completed successfully"
else
  puts "Backup failed: #{result.stderr}"
end

client.containers.delete(container.id)
```

```php
<?php
// PHP - WordPress maintenance automation
use Rexec\RexecClient;

$client = new RexecClient(getenv('REXEC_URL'), getenv('REXEC_TOKEN'));

// Create WordPress container
$container = $client->containers()->create('wordpress:latest', [
    'name' => 'wp-maintenance',
    'environment' => [
        'WORDPRESS_DB_HOST' => getenv('DB_HOST'),
        'WORDPRESS_DB_PASSWORD' => getenv('DB_PASSWORD'),
    ]
]);

$client->containers()->start($container->id);

// Run WP-CLI commands
$result = $client->containers()->exec($container->id, 
    'wp plugin update --all --allow-root'
);

echo "Plugin updates: " . $result->stdout;

// Cleanup
$client->containers()->delete($container->id);
```

### AI/ML Training Environments

Spin up GPU-enabled training environments with the [Python SDK](../sdk/python/README.md):

```python
# Python - ML training job manager
import asyncio
from rexec import RexecClient

async def run_training_job(model_config: dict):
    async with RexecClient(base_url, token) as client:
        # Create GPU-enabled container
        container = await client.containers.create(
            image="pytorch/pytorch:2.0-cuda11.7-cudnn8-runtime",
            name=f"training-{model_config['name']}",
            environment={
                "CUDA_VISIBLE_DEVICES": "0",
                "WANDB_API_KEY": os.environ["WANDB_API_KEY"],
            },
            labels={
                "job": model_config["name"],
                "type": "training",
            }
        )
        
        await client.containers.start(container.id)
        
        # Upload training script and data
        await client.files.write(
            container.id,
            "/workspace/train.py",
            model_config["script"]
        )
        
        # Run training
        terminal = await client.terminal.connect(container.id)
        await terminal.write(b"python /workspace/train.py\n")
        
        # Stream output
        async for data in terminal:
            print(data.decode(), end="")
        
        # Download trained model
        model_data = await client.files.read(
            container.id,
            "/workspace/model.pth"
        )
        
        await client.containers.delete(container.id)
        return model_data
```

### API Testing & Mocking

Test APIs against real service containers with the [JavaScript SDK](../sdk/js/README.md):

```typescript
// JavaScript - Integration testing with real services
import { RexecClient } from '@pipeopshq/rexec';
import { describe, it, beforeAll, afterAll } from 'vitest';

describe('API Integration Tests', () => {
  let client: RexecClient;
  let redisContainer: Container;
  let postgresContainer: Container;

  beforeAll(async () => {
    client = new RexecClient({ baseURL, token });

    // Spin up test infrastructure
    [redisContainer, postgresContainer] = await Promise.all([
      client.containers.create({ image: 'redis:7' }),
      client.containers.create({ 
        image: 'postgres:16',
        environment: { POSTGRES_PASSWORD: 'test' }
      }),
    ]);

    await Promise.all([
      client.containers.start(redisContainer.id),
      client.containers.start(postgresContainer.id),
    ]);
  });

  afterAll(async () => {
    await Promise.all([
      client.containers.delete(redisContainer.id),
      client.containers.delete(postgresContainer.id),
    ]);
  });

  it('should cache data in Redis', async () => {
    const result = await client.containers.exec(
      redisContainer.id,
      'redis-cli SET test "hello" && redis-cli GET test'
    );
    expect(result.stdout.trim()).toBe('hello');
  });
});
```

## Rate Limits

API requests are rate-limited to ensure fair usage:

- **Container creation**: 10 per minute
- **API requests**: 100 per minute
- **WebSocket connections**: 5 concurrent per user

## Error Handling

All SDKs provide structured error handling:

### Go

```go
container, err := client.Containers.Get(ctx, "invalid-id")
if err != nil {
    if apiErr, ok := err.(*rexec.APIError); ok {
        fmt.Printf("API Error %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

### JavaScript/TypeScript

```typescript
try {
  await client.containers.get('invalid-id');
} catch (error) {
  if (error instanceof RexecError) {
    console.error(`API Error ${error.statusCode}: ${error.message}`);
  }
}
```

### Python

```python
from rexec import RexecException

try:
    container = await client.containers.get("invalid-id")
except RexecException as e:
    if e.status_code:
        print(f"API Error {e.status_code}: {e}")
```

### Rust

```rust
match client.containers().get("invalid-id").await {
    Ok(container) => println!("Got: {}", container.id),
    Err(rexec::Error::Api { status, message }) => {
        eprintln!("API Error {}: {}", status, message);
    }
    Err(e) => eprintln!("Error: {}", e),
}
```

### Ruby

```ruby
begin
  container = client.containers.get("invalid-id")
rescue Rexec::Error => e
  puts "Error: #{e.message}"
end
```

### Java

```java
try {
    Container container = client.containers().get("invalid-id");
} catch (RexecException e) {
    if (e.isApiError()) {
        System.out.println("API Error " + e.getStatusCode() + ": " + e.getMessage());
    }
}
```

### C#/.NET

```csharp
try {
    var container = await client.Containers.GetAsync("invalid-id");
} catch (RexecException ex) when (ex.IsApiError) {
    Console.WriteLine($"API Error {ex.StatusCode}: {ex.Message}");
}
```

### PHP

```php
try {
    $container = $client->containers()->get("invalid-id");
} catch (RexecException $e) {
    if ($e->isApiError()) {
        echo "API Error " . $e->getStatusCode() . ": " . $e->getMessage();
    }
}
```

## Best Practices

1. **Reuse clients**: Create one client instance and reuse it
2. **Handle errors**: Always check for and handle errors appropriately
3. **Clean up**: Delete containers when done to free resources
4. **Use environment variables**: Never hardcode tokens
5. **Set timeouts**: Configure appropriate timeouts for your use case

## See Also

- [Getting Started Guide](SDK_GETTING_STARTED.md) - Step-by-step tutorial for all SDKs
- [SDK Roadmap](SDK_ROADMAP.md) - Future SDK plans and contribution guidelines
- [Security Guide](SECURITY.md) - Security best practices
- [CLI Documentation](CLI.md) - Command-line interface reference

### SDK-Specific Documentation

- [Go SDK](../sdk/go/README.md) - Idiomatic Go with context support
- [JavaScript/TypeScript SDK](../sdk/js/README.md) - Promise-based async API
- [Python SDK](../sdk/python/README.md) - Async/await with httpx
- [Rust SDK](../sdk/rust/README.md) - Tokio-based async runtime
- [Ruby SDK](../sdk/ruby/README.md) - Faraday HTTP with WebSocket support
- [Java SDK](../sdk/java/README.md) - OkHttp with WebSocket
- [C#/.NET SDK](../sdk/dotnet/README.md) - Native async/await patterns
- [PHP SDK](../sdk/php/README.md) - Guzzle HTTP with ReactPHP WebSocket

## Support

- [GitHub Issues](https://github.com/PipeOpsHQ/rexec/issues) - Bug reports and feature requests
- [GitHub Discussions](https://github.com/PipeOpsHQ/rexec/discussions) - Questions and community support
- [GitHub Repository](https://github.com/PipeOpsHQ/rexec) - Source code and contributions
