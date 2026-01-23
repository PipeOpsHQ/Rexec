# Rexec Java SDK

Official Java SDK for [Rexec](https://github.com/PipeOpsHQ/rexec) - Terminal as a Service.

## Requirements

- Java 17 or later
- Maven or Gradle

## Installation

### Maven

```xml
<dependency>
    <groupId>io.pipeops</groupId>
    <artifactId>rexec</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Gradle

```groovy
implementation 'io.pipeops:rexec:1.0.0'
```

## Quick Start

```java
import io.pipeops.rexec.*;

public class Example {
    public static void main(String[] args) throws RexecException {
        // Create client
        RexecClient client = new RexecClient(
            "https://your-instance.com",
            "your-api-token"
        );

        // Create a container
        Container container = client.containers().create("ubuntu:24.04");
        System.out.println("Created: " + container.getId());

        // Start it
        client.containers().start(container.getId());

        // Execute a command
        ExecResult result = client.containers().exec(container.getId(), "echo 'Hello from Java!'");
        System.out.println(result.getStdout());

        // Clean up
        client.containers().delete(container.getId());
    }
}
```

## Features

### Container Management

```java
// List all containers
List<Container> containers = client.containers().list();

// Create with options
Container container = client.containers().create(
    new CreateContainerRequest("python:3.12")
        .setName("my-python-sandbox")
        .addEnv("PYTHONPATH", "/app")
        .addLabel("project", "demo")
);

// Lifecycle
client.containers().start(containerId);
client.containers().stop(containerId);
client.containers().delete(containerId);

// Execute commands
ExecResult result = client.containers().exec(containerId, "python --version");
if (result.isSuccess()) {
    System.out.println(result.getStdout());
}
```

### File Operations

```java
FileService files = client.files();

// List directory
List<FileInfo> entries = files.list(containerId, "/app");
for (FileInfo file : entries) {
    System.out.println(file.getName() + " - " + (file.isDirectory() ? "DIR" : file.getSize() + " bytes"));
}

// Read file
String content = files.readString(containerId, "/etc/hostname");

// Write file
files.write(containerId, "/app/script.py", "print('Hello!')");

// Delete file
files.delete(containerId, "/tmp/scratch.txt");
```

### Interactive Terminal

```java
Terminal terminal = client.terminal().connect(containerId);

// Set up handlers
terminal.onData(data -> System.out.print(data))
        .onClose(() -> System.out.println("Disconnected"))
        .onError(e -> e.printStackTrace());

// Send commands
terminal.write("ls -la\n");
terminal.write("cd /app && python main.py\n");

// Resize terminal
terminal.resize(120, 40);

// Clean up
terminal.close();
```

## Error Handling

```java
try {
    Container container = client.containers().get("invalid-id");
} catch (RexecException e) {
    if (e.isApiError()) {
        System.out.println("API error " + e.getStatusCode() + ": " + e.getMessage());
    } else {
        System.out.println("Network error: " + e.getMessage());
    }
}
```

## Building from Source

```bash
cd sdk/java
mvn clean install
```

## License

MIT License - see [LICENSE](../../LICENSE) for details.
