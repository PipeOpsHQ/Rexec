<?php

declare(strict_types=1);

namespace Rexec;

/**
 * Service for managing sandboxes/containers.
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
     * Handles production payload `{ containers: [...], count, limit }`
     * (including `containers: null`) and a bare JSON array.
     *
     * @return Container[]
     * @throws RexecException
     */
    public function list(): array
    {
        $response = $this->client->request('GET', '/api/containers');
        $items = self::normalizeList($response);

        return array_map(static fn(array $data) => new Container($data), $items);
    }

    /**
     * Get a container by ID.
     *
     * @throws RexecException
     */
    public function get(string $containerId): Container
    {
        $response = $this->client->request('GET', "/api/containers/{$containerId}");
        if (!is_array($response)) {
            throw new RexecException('Unexpected empty response for get container', 500);
        }

        return new Container($response);
    }

    /**
     * Create a new container.
     *
     * Prefer image aliases such as `ubuntu`, `debian`, or `alpine`
     * (not `ubuntu` on hosted Rexec).
     *
     * @param string $image Image alias or name
     * @param array{name?: string, environment?: array<string, string>, labels?: array<string, string>} $options
     * @throws RexecException
     */
    public function create(string $image, array $options = []): Container
    {
        $body = array_merge(['image' => $image], $options);
        $response = $this->client->request('POST', '/api/containers', $body);
        if (!is_array($response)) {
            throw new RexecException('Unexpected empty response for create container', 500);
        }

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
     * @return list<array<string, mixed>>
     */
    private static function normalizeList(mixed $response): array
    {
        if ($response === null) {
            return [];
        }
        if (is_array($response) && array_is_list($response)) {
            /** @var list<array<string, mixed>> $response */
            return $response;
        }
        if (is_array($response) && array_key_exists('containers', $response)) {
            $containers = $response['containers'];
            if ($containers === null) {
                return [];
            }
            if (is_array($containers)) {
                /** @var list<array<string, mixed>> $containers */
                return array_values($containers);
            }
        }

        return [];
    }
}
