<?php

declare(strict_types=1);

namespace Rexec;

/**
 * Service for managing sandboxes.
 */
class SandboxService
{
    private RexecClient $client;

    public function __construct(RexecClient $client)
    {
        $this->client = $client;
    }

    /**
     * List all sandboxes.
     *
     * Handles production payload `{ containers: [...], count, limit }`
     * (including `containers: null`) and a bare JSON array.
     * Wire path remains `/api/containers`.
     *
     * @return Sandbox[]
     * @throws RexecException
     */
    public function list(): array
    {
        $response = $this->client->request('GET', '/api/containers');
        $items = self::normalizeList($response);

        return array_map(static fn(array $data) => new Sandbox($data), $items);
    }

    /**
     * Get a sandbox by ID.
     *
     * @throws RexecException
     */
    public function get(string $sandboxId): Sandbox
    {
        $response = $this->client->request('GET', "/api/containers/{$sandboxId}");
        if (!is_array($response)) {
            throw new RexecException('Unexpected empty response for get sandbox', 500);
        }

        return new Sandbox($response);
    }

    /**
     * Create a new sandbox.
     *
     * Prefer image aliases such as `ubuntu`, `debian`, or `alpine`.
     *
     * @param string $image Image alias or name
     * @param array{name?: string, environment?: array<string, string>, labels?: array<string, string>} $options
     * @throws RexecException
     */
    public function create(string $image, array $options = []): Sandbox
    {
        $body = array_merge(['image' => $image], $options);
        $response = $this->client->request('POST', '/api/containers', $body);
        if (!is_array($response)) {
            throw new RexecException('Unexpected empty response for create sandbox', 500);
        }

        return new Sandbox($response);
    }

    /**
     * Start a sandbox.
     *
     * @throws RexecException
     */
    public function start(string $sandboxId): void
    {
        $this->client->request('POST', "/api/containers/{$sandboxId}/start");
    }

    /**
     * Stop a sandbox.
     *
     * @throws RexecException
     */
    public function stop(string $sandboxId): void
    {
        $this->client->request('POST', "/api/containers/{$sandboxId}/stop");
    }

    /**
     * Delete a sandbox.
     *
     * @throws RexecException
     */
    public function delete(string $sandboxId): void
    {
        $this->client->request('DELETE', "/api/containers/{$sandboxId}");
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
