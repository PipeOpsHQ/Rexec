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
        $encodedPath = rawurlencode($path);
        $response = $this->client->request(
            'GET',
            "/api/containers/{$containerId}/files/list?path={$encodedPath}"
        );

        $files = [];
        if (is_array($response)) {
            if (array_is_list($response)) {
                $files = $response;
            } elseif (isset($response['files']) && is_array($response['files'])) {
                $files = $response['files'];
            }
        }

        return array_map(static fn(array $data) => new FileInfo($data), $files);
    }

    /**
     * Read a file's contents.
     *
     * @throws RexecException
     */
    public function read(string $containerId, string $path): string
    {
        $encodedPath = rawurlencode($path);
        return $this->client->requestBytes(
            'GET',
            "/api/containers/{$containerId}/files?path={$encodedPath}"
        );
    }

    /**
     * Write content to a file (content is base64-encoded for transport).
     *
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
        $encodedPath = rawurlencode($path);
        $this->client->request(
            'DELETE',
            "/api/containers/{$containerId}/files?path={$encodedPath}"
        );
    }
}
