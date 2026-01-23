# Rexec PHP SDK

Official PHP SDK for [Rexec](https://github.com/PipeOpsHQ/rexec) - Terminal as a Service.

## Requirements

- PHP 8.1 or later
- Composer

## Installation

```bash
composer require pipeopshq/rexec
```

## Quick Start

```php
<?php

require 'vendor/autoload.php';

use Rexec\RexecClient;

// Create client
$client = new RexecClient('https://your-instance.com', 'your-api-token');

// Create a container
$container = $client->containers()->create('ubuntu:24.04');
echo "Created: {$container->id}\n";

// Start it
$client->containers()->start($container->id);

// Execute a command
$result = $client->containers()->exec($container->id, "echo 'Hello from PHP!'");
echo $result->stdout;

// Clean up
$client->containers()->delete($container->id);
```

## Features

### Container Management

```php
// List all containers
$containers = $client->containers()->list();

// Create with options
$container = $client->containers()->create('python:3.12', [
    'name' => 'my-python-sandbox',
    'environment' => ['PYTHONPATH' => '/app'],
    'labels' => ['project' => 'demo'],
]);

// Lifecycle
$client->containers()->start($containerId);
$client->containers()->stop($containerId);
$client->containers()->delete($containerId);

// Execute commands
$result = $client->containers()->exec($containerId, 'python --version');
if ($result->isSuccess()) {
    echo $result->stdout;
}

// Execute with array command
$result = $client->containers()->exec($containerId, ['python', '-c', 'print("Hello")']);
```

### File Operations

```php
// List directory
$files = $client->files()->list($containerId, '/app');
foreach ($files as $file) {
    $type = $file->isDir ? 'DIR' : "{$file->size} bytes";
    echo "{$file->name} - {$type}\n";
}

// Read file
$content = $client->files()->read($containerId, '/etc/hostname');

// Write file
$client->files()->write($containerId, '/app/script.py', "print('Hello!')");

// Delete file
$client->files()->delete($containerId, '/tmp/scratch.txt');
```

### Interactive Terminal

```php
use React\EventLoop\Loop;

$terminal = $client->terminal()->connect($containerId);

// Set up handlers
$terminal->onData(function ($data) {
    echo $data;
});

$terminal->onClose(function () {
    echo "Disconnected\n";
});

$terminal->onError(function ($e) {
    echo "Error: {$e->getMessage()}\n";
});

// Open connection
$terminal->open();

// Send commands
$terminal->write("ls -la\n");
$terminal->write("cd /app && python main.py\n");

// Resize terminal
$terminal->resize(120, 40);

// Run the event loop
Loop::run();

// Clean up
$terminal->close();
```

## Error Handling

```php
use Rexec\RexecException;

try {
    $container = $client->containers()->get('invalid-id');
} catch (RexecException $e) {
    if ($e->isApiError()) {
        echo "API error {$e->getStatusCode()}: {$e->getMessage()}\n";
    } else {
        echo "Network error: {$e->getMessage()}\n";
    }
}
```

## Building from Source

```bash
cd sdk/php
composer install
composer test
```

## License

MIT License - see [LICENSE](../../LICENSE) for details.
