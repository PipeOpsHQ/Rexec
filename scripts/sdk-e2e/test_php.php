<?php

declare(strict_types=1);

/**
 * E2E smoke: list → create → get → delete for the PHP SDK.
 *
 * Usage:
 *   URL=https://rexec.sh TOKEN=... php test_php.php
 */

$root = dirname(__DIR__, 2) . '/sdk/php';
require $root . '/vendor/autoload.php';

use Rexec\RexecClient;
use Rexec\RexecException;

$url = getenv('URL') ?: getenv('REXEC_URL') ?: 'https://rexec.sh';
$token = getenv('TOKEN') ?: getenv('REXEC_TOKEN') ?: '';
if ($token === '') {
    fwrite(STDERR, "TOKEN or REXEC_TOKEN required\n");
    exit(1);
}

$client = new RexecClient($url, $token);

try {
    echo "[php] list...\n";
    $before = $client->containers()->list();
    echo '[php] list count ' . count($before) . "\n";

    $name = 'php-e2e-' . time();
    echo "[php] create...\n";
    $c = $client->containers()->create('ubuntu', ['name' => $name]);
    echo "[php] created {$c->id} {$c->status} {$c->image}\n";
    if ($c->id === '') {
        throw new RuntimeException('create returned empty id');
    }

    $after = $client->containers()->list();
    echo '[php] list count ' . count($after) . "\n";

    $got = $client->containers()->get($c->id);
    echo "[php] get {$got->id} {$got->status}\n";

    $client->containers()->delete($c->id);
    echo "[php] OK\n";
} catch (RexecException $e) {
    fwrite(STDERR, '[php] FAIL ' . $e->getMessage() . "\n");
    exit(1);
}
