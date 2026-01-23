<?php

declare(strict_types=1);

namespace Rexec;

/**
 * Service for managing containers.
 */
class ContainerService
{
    private RexecClient $client;

    public function __construct(RexecClient $client)
    {
        $this->client = $client;
    }

    /**
     * List all containers.
     *
     * @return Container[]
     * @throws RexecException
     */
    public function list(): array
    {
        $response = $this->client->request('GET', '/api/containers');
        $containers = $response['containers'] ?? [];

        return array_map(fn($data) => new Container($data), $containers);
    }

    /**
     * Get a container by ID.
     *
     * @throws RexecException
     */
    public function get(string $containerId): Container
    {
        $response = $this->client->request('GET', "/api/containers/{$containerId}");
        return new Container($response);
    }

    /**
     * Create a new container.
     *
     * @param string $image Docker image to use
     * @param array $options Optional: name, environment, labels
     * @throws RexecException
     */
    public function create(string $image, array $options = []): Container
    {
        $body = array_merge(['image' => $image], $options);
        $response = $this->client->request('POST', '/api/containers', $body);
        return new Container($response);
    }

    /**
     * Start a container.
     *
     * @throws RexecException
     */
    public function start(string $containerId): void
    {
        $this->client->request('POST', "/api/containers/{$containerId}/start");
    }

    /**
     * Stop a container.
     *
     * @throws RexecException
     */
    public function stop(string $containerId): void
    {
        $this->client->request('POST', "/api/containers/{$containerId}/stop");
    }

    /**
     * Delete a container.
     *
     * @throws RexecException
     */
    public function delete(string $containerId): void
    {
        $this->client->request('DELETE', "/api/containers/{$containerId}");
    }

    /**
     * Execute a command in a container.
     *
     * @param string $containerId Container ID
     * @param string|array $command Command to execute
     * @throws RexecException
     */
    public function exec(string $containerId, string|array $command): ExecResult
    {
        if (is_string($command)) {
            $command = ['/bin/sh', '-c', $command];
        }

        $response = $this->client->request('POST', "/api/containers/{$containerId}/exec", [
            'command' => $command,
        ]);

        return new ExecResult($response);
    }
}
