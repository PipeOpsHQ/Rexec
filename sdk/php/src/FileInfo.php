<?php

declare(strict_types=1);

namespace Rexec;

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
        $this->name = (string) ($data['name'] ?? '');
        $this->path = (string) ($data['path'] ?? '');
        $this->size = (int) ($data['size'] ?? 0);
        $this->mode = isset($data['mode']) ? (string) $data['mode'] : null;
        $this->modTime = isset($data['mod_time'])
            ? (string) $data['mod_time']
            : (isset($data['modTime']) ? (string) $data['modTime'] : null);
        $this->isDir = (bool) ($data['is_dir'] ?? $data['isDir'] ?? false);
    }

    public function isFile(): bool
    {
        return !$this->isDir;
    }
}
