# Rexec .NET SDK

Official .NET SDK for [Rexec](https://github.com/PipeOpsHQ/rexec) - Terminal as a Service.

## Requirements

- .NET 8.0 or later

## Installation

### NuGet

```bash
dotnet add package Rexec
```

### Package Manager

```powershell
Install-Package Rexec
```

## Quick Start

```csharp
using Rexec;

// Create client
var client = new RexecClient("https://your-instance.com", "your-api-token");

// Create a container
var container = await client.Containers.CreateAsync("ubuntu:24.04");
Console.WriteLine($"Created: {container.Id}");

// Start it
await client.Containers.StartAsync(container.Id);

// Execute a command
var result = await client.Containers.ExecAsync(container.Id, "echo 'Hello from .NET!'");
Console.WriteLine(result.Stdout);

// Clean up
await client.Containers.DeleteAsync(container.Id);
```

## Features

### Container Management

```csharp
// List all containers
var containers = await client.Containers.ListAsync();

// Create with options
var container = await client.Containers.CreateAsync(
    new CreateContainerRequest("python:3.12")
        .WithEnv("PYTHONPATH", "/app")
        .WithLabel("project", "demo")
);

// Lifecycle
await client.Containers.StartAsync(containerId);
await client.Containers.StopAsync(containerId);
await client.Containers.DeleteAsync(containerId);

// Execute commands
var result = await client.Containers.ExecAsync(containerId, "python --version");
if (result.IsSuccess)
{
    Console.WriteLine(result.Stdout);
}
```

### File Operations

```csharp
// List directory
var files = await client.Files.ListAsync(containerId, "/app");
foreach (var file in files)
{
    Console.WriteLine($"{file.Name} - {(file.IsDir ? "DIR" : $"{file.Size} bytes")}");
}

// Read file
var content = await client.Files.ReadStringAsync(containerId, "/etc/hostname");

// Write file
await client.Files.WriteAsync(containerId, "/app/script.py", "print('Hello!')");

// Delete file
await client.Files.DeleteAsync(containerId, "/tmp/scratch.txt");
```

### Interactive Terminal

```csharp
await using var terminal = await client.Terminal.ConnectAsync(containerId);

// Set up handlers
terminal.OnData += data => Console.Write(data);
terminal.OnClose += () => Console.WriteLine("Disconnected");
terminal.OnError += ex => Console.WriteLine($"Error: {ex.Message}");

// Send commands
await terminal.WriteAsync("ls -la\n");
await terminal.WriteAsync("cd /app && python main.py\n");

// Resize terminal
await terminal.ResizeAsync(120, 40);

// The using statement handles cleanup automatically
```

## Error Handling

```csharp
try
{
    var container = await client.Containers.GetAsync("invalid-id");
}
catch (RexecException ex) when (ex.IsApiError)
{
    Console.WriteLine($"API error {ex.StatusCode}: {ex.Message}");
}
catch (RexecException ex)
{
    Console.WriteLine($"Network error: {ex.Message}");
}
```

## Async/Await and Cancellation

All methods support cancellation tokens:

```csharp
using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(30));

try
{
    var containers = await client.Containers.ListAsync(cts.Token);
}
catch (OperationCanceledException)
{
    Console.WriteLine("Operation timed out");
}
```

## Building from Source

```bash
cd sdk/dotnet
dotnet build
dotnet pack
```

## License

MIT License - see [LICENSE](../../LICENSE) for details.
