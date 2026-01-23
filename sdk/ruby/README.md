# Rexec Ruby SDK

Official Ruby SDK for [Rexec](https://github.com/PipeOpsHQ/rexec) - Terminal as a Service.

## Installation

Add to your Gemfile:

```ruby
gem 'rexec'
```

Or install directly:

```bash
gem install rexec
```

## Quick Start

```ruby
require 'rexec'

client = Rexec::Client.new("https://your-instance.com", "your-token")

# Create a container
container = client.containers.create(
  image: "ubuntu:24.04",
  name: "my-sandbox"
)
puts "Created container: #{container.id}"

# Connect to terminal
terminal = client.terminal.connect(container.id)

terminal.on_data { |data| print data }
terminal.on_close { puts "\nConnection closed" }

terminal.write("echo 'Hello from Rexec!'\n")

sleep 2  # Wait for output

# Clean up
terminal.close
client.containers.delete(container.id)
```

## API Reference

### Client

```ruby
require 'rexec'

# Create client
client = Rexec::Client.new(
  "https://your-instance.com",
  "your-api-token",
  timeout: 30  # optional
)
```

### Containers

```ruby
# List all containers
containers = client.containers.list
containers.each do |c|
  puts "#{c.name}: #{c.status}"
end

# Get a specific container
container = client.containers.get("container-id")

# Create a container
container = client.containers.create(
  image: "ubuntu:24.04",
  name: "my-container",
  environment: { "MY_VAR" => "value" },
  labels: { "project" => "demo" }
)

# Start a container
client.containers.start("container-id")

# Stop a container
client.containers.stop("container-id")

# Delete a container
client.containers.delete("container-id")
```

### Files

```ruby
# List files in a directory
files = client.files.list("container-id", "/home")
files.each do |f|
  icon = f.directory? ? "📁" : "📄"
  puts "#{icon} #{f.name}"
end

# Download a file
content = client.files.download("container-id", "/etc/passwd")
puts content

# Create a directory
client.files.mkdir("container-id", "/home/mydir")

# Delete a file
client.files.delete("container-id", "/home/file.txt")
```

### Terminal

```ruby
# Connect to terminal
terminal = client.terminal.connect("container-id", cols: 120, rows: 40)

# Send commands
terminal.write("ls -la\n")

# Handle output
terminal.on_data do |data|
  print data
end

# Handle close
terminal.on_close do
  puts "Connection closed"
end

# Handle errors
terminal.on_error do |error|
  puts "Error: #{error}"
end

# Resize terminal
terminal.resize(150, 50)

# Close connection
terminal.close
```

## Examples

### Run a Script

```ruby
def run_script(client, container_id, script)
  output = []
  done = false
  
  terminal = client.terminal.connect(container_id)
  
  terminal.on_data { |data| output << data }
  terminal.on_close { done = true }
  
  terminal.write("#{script}\nexit\n")
  
  sleep 0.1 until done
  
  output.join
end

result = run_script(client, container.id, "apt update && apt install -y curl")
puts result
```

### Batch Operations

```ruby
require 'concurrent'

def create_batch(client, count)
  futures = (0...count).map do |i|
    Concurrent::Future.execute do
      client.containers.create(
        image: "ubuntu:24.04",
        name: "worker-#{i}"
      )
    end
  end
  
  futures.map(&:value)
end

containers = create_batch(client, 5)
```

## Error Handling

```ruby
begin
  container = client.containers.get("invalid-id")
rescue Rexec::APIError => e
  puts "API Error #{e.status_code}: #{e.message}"
rescue Rexec::ConnectionError => e
  puts "Connection Error: #{e.message}"
end
```

## Requirements

- Ruby 3.0+
- `faraday` for HTTP requests
- `websocket-client-simple` for terminal connections

## License

MIT License - see [LICENSE](../../LICENSE) for details.
