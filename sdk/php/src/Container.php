<?php

declare(strict_types=1);

namespace Rexec;

/**
 * Represents a Rexec container/sandbox.
 *
 * Accepts both snake_case (production API) and camelCase keys.
 */
class Container
{
    public string $id;
    public ?string $name;
    public string $image;
    public string $status;
    public ?string $createdAt;
    public ?string $startedAt;
    /** @var array<string, string> */
    public array $labels;
    /** @var array<string, string> */
    public array $environment;

    public function __construct(array $data)
    {
        $this->id = (string) self::pick($data, 'id', '');
        $this->name = self::pick($data, 'name', null);
        if ($this->name !== null) {
            $this->name = (string) $this->name;
        }
        $this->image = (string) self::pick($data, 'image', '');
        $this->status = (string) self::pick($data, 'status', '');
        $this->createdAt = self::stringOrNull(self::pick($data, 'created_at', self::pick($data, 'createdAt', null)));
        $this->startedAt = self::stringOrNull(self::pick($data, 'started_at', self::pick($data, 'startedAt', null)));
        $this->labels = is_array($data['labels'] ?? null) ? $data['labels'] : [];
        $this->environment = is_array($data['environment'] ?? null) ? $data['environment'] : [];
    }

    public function isRunning(): bool
    {
        return $this->status === 'running';
    }

    public function isStopped(): bool
    {
        return $this->status === 'stopped';
    }

    private static function pick(array $data, string $key, mixed $default): mixed
    {
        return array_key_exists($key, $data) ? $data[$key] : $default;
    }

    private static function stringOrNull(mixed $value): ?string
    {
        if ($value === null || $value === '') {
            return null;
        }
        return (string) $value;
    }
}
