<?php

declare(strict_types=1);

namespace Rexec;

/**
 * Represents a Rexec container/sandbox.
 */
class Container
{
    public string $id;
    public ?string $name;
    public string $image;
    public string $status;
    public ?string $createdAt;
    public ?string $startedAt;
    public array $labels;
    public array $environment;

    public function __construct(array $data)
    {
        $this->id = $data['id'] ?? '';
        $this->name = $data['name'] ?? null;
        $this->image = $data['image'] ?? '';
        $this->status = $data['status'] ?? '';
        $this->createdAt = $data['createdAt'] ?? null;
        $this->startedAt = $data['startedAt'] ?? null;
        $this->labels = $data['labels'] ?? [];
        $this->environment = $data['environment'] ?? [];
    }

    public function isRunning(): bool
    {
        return $this->status === 'running';
    }

    public function isStopped(): bool
    {
        return $this->status === 'stopped';
    }
}

/**
 * Information about a file in a container.
 */
class FileInfo
{
    public string $name;
    public string $path;
    public int $size;
    public ?string $mode;
    public ?string $modTime;
    public bool $isDir;

    public function __construct(array $data)
    {
        $this->name = $data['name'] ?? '';
        $this->path = $data['path'] ?? '';
        $this->size = $data['size'] ?? 0;
        $this->mode = $data['mode'] ?? null;
        $this->modTime = $data['modTime'] ?? null;
        $this->isDir = $data['isDir'] ?? false;
    }

    public function isFile(): bool
    {
        return !$this->isDir;
    }
}

/**
 * Result of command execution.
 */
class ExecResult
{
    public int $exitCode;
    public string $stdout;
    public string $stderr;

    public function __construct(array $data)
    {
        $this->exitCode = $data['exitCode'] ?? 0;
        $this->stdout = $data['stdout'] ?? '';
        $this->stderr = $data['stderr'] ?? '';
    }

    public function isSuccess(): bool
    {
        return $this->exitCode === 0;
    }
}
