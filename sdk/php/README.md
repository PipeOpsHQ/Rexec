# Rexec PHP SDK

Official PHP client for the [Rexec](https://rexec.sh) API (sandboxes, files, terminals).

**Package:** [`pipeopshq/rexec`](https://packagist.org/packages/pipeopshq/rexec) · **Version:** 1.0.1  
**Packagist source repo:** [PipeOpsHQ/rexec-php](https://github.com/PipeOpsHQ/rexec-php) (synced from this monorepo path)

## Requirements

- PHP 8.1+
- Composer
- ext-json, ext-curl (via Guzzle)

## Install

```bash
composer require pipeopshq/rexec
```

Until Packagist is linked, install from the monorepo path:

```bash
composer config repositories.rexec path ../path/to/rexec/sdk/php
composer require pipeopshq/rexec:@dev
```

## Quick start

```php
<?php

require 'vendor/autoload.php';

use Rexec\RexecClient;
use Rexec\RexecException;

$client = new RexecClient(
    getenv('REXEC_URL') ?: 'https://rexec.sh',
    getenv('REXEC_TOKEN') ?: ''
);

try {
    $list = $client->containers()->list();
    echo 'count: ' . count($list) . PHP_EOL;

    // Prefer image aliases: ubuntu, debian, alpine (not ubuntu on hosted Rexec)
    $container = $client->containers()->create('ubuntu', [
        'name' => 'php-demo',
    ]);
    echo "created {$container->id} status={$container->status}\n";

    $got = $client->containers()->get($container->id);
    echo "get {$got->id}\n";

    $client->containers()->delete($container->id);
    echo "deleted\n";
} catch (RexecException $e) {
    fwrite(STDERR, $e->getMessage() . PHP_EOL);
    exit(1);
}
```

## Containers

```php
$containers = $client->containers()->list();
$c = $client->containers()->create('ubuntu', ['name' => 'demo']);
$client->containers()->get($c->id);
$client->containers()->start($c->id);
$client->containers()->stop($c->id);
$client->containers()->delete($c->id);
```

> There is no HTTP `exec` API on Rexec. Run commands over the terminal WebSocket
> (`$client->terminal()->connect($id)`).

## Files

```php
$files = $client->files()->list($containerId, '/home');
$content = $client->files()->read($containerId, '/etc/hostname');
$client->files()->write($containerId, '/tmp/hello.txt', "hi\n");
$client->files()->delete($containerId, '/tmp/hello.txt');
```

## Terminal (WebSocket)

```php
use React\EventLoop\Loop;

$terminal = $client->terminal()->connect($containerId);
$terminal->onData(fn ($data) => print($data));
$terminal->open();
$terminal->write("echo hello\n");
Loop::run();
$terminal->close();
```

## Auth

Create an API token in the Rexec UI (**Settings → API Tokens**), or use a guest JWT:

```bash
export REXEC_URL=https://rexec.sh
export REXEC_TOKEN=$(curl -sS -X POST "$REXEC_URL/api/auth/guest" \
  -H 'Content-Type: application/json' \
  -d '{"username":"php_demo","email":"you@example.com"}' \
  | php -r 'echo json_decode(stream_get_contents(STDIN))->token;')
```

## License

MIT — see [LICENSE](LICENSE).
