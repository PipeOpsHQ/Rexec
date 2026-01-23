<?php

declare(strict_types=1);

namespace Rexec;

/**
 * Service for file operations within containers.
 */
class FileService
{
    private RexecClient $client;

    public function __construct(RexecClient $client)
    {
        $this->client = $client;
    }

    /**
     * List files in a directory.
     *
     * @return FileInfo[]
     * @throws RexecException
     */
    public function list(string $containerId, string $path): array
    {
        $encodedPath = urlencode($path);
        $response = $this->client->request('GET', "/api/containers/{$containerId}/files/list?path={$encodedPath}");
        $files = $response['files'] ?? [];

        return array_map(fn($data) => new FileInfo($data), $files);
    }

    /**
     * Read a file's contents.
     *
     * @throws RexecException
     */
    public function read(string $containerId, string $path): string
    {
        $encodedPath = urlencode($path);
        return $this->client->requestBytes('GET', "/api/containers/{$containerId}/files?path={$encodedPath}");
    }

    /**
     * Write content to a file.
     *
     * @param string $containerId Container ID
     * @param string $path File path
     * @param string $content Content to write
     * @throws RexecException
     */
    public function write(string $containerId, string $path, string $content): void
    {
        $this->client->request('POST', "/api/containers/{$containerId}/files", [
            'path' => $path,
            'content' => base64_encode($content),
        ]);
    }

    /**
     * Delete a file.
     *
     * @throws RexecException
     */
    public function delete(string $containerId, string $path): void
    {
        $encodedPath = urlencode($path);
        $this->client->request('DELETE', "/api/containers/{$containerId}/files?path={$encodedPath}");
    }
}
