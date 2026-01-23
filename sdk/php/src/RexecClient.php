<?php

declare(strict_types=1);

namespace Rexec;

use GuzzleHttp\Client as HttpClient;
use GuzzleHttp\Exception\GuzzleException;

/**
 * Main client for interacting with Rexec API.
 *
 * @example
 * $client = new RexecClient('https://your-instance.com', 'your-token');
 * $container = $client->containers()->create('ubuntu:24.04');
 * $client->containers()->start($container->id);
 */
class RexecClient
{
    private HttpClient $httpClient;
    private string $baseUrl;
    private string $token;

    private ContainerService $containers;
    private FileService $files;
    private TerminalService $terminal;

    /**
     * Create a new Rexec client.
     *
     * @param string $baseUrl Base URL of your Rexec instance
     * @param string $token API token for authentication
     * @param array $options Additional Guzzle options
     */
    public function __construct(string $baseUrl, string $token, array $options = [])
    {
        $this->baseUrl = rtrim($baseUrl, '/');
        $this->token = $token;

        $defaultOptions = [
            'base_uri' => $this->baseUrl,
            'timeout' => 30,
            'headers' => [
                'Authorization' => 'Bearer ' . $token,
                'Accept' => 'application/json',
                'Content-Type' => 'application/json',
            ],
        ];

        $this->httpClient = new HttpClient(array_merge($defaultOptions, $options));

        $this->containers = new ContainerService($this);
        $this->files = new FileService($this);
        $this->terminal = new TerminalService($this);
    }

    /**
     * Get the container service.
     */
    public function containers(): ContainerService
    {
        return $this->containers;
    }

    /**
     * Get the file service.
     */
    public function files(): FileService
    {
        return $this->files;
    }

    /**
     * Get the terminal service.
     */
    public function terminal(): TerminalService
    {
        return $this->terminal;
    }

    /**
     * Get the base URL.
     */
    public function getBaseUrl(): string
    {
        return $this->baseUrl;
    }

    /**
     * Get the API token.
     */
    public function getToken(): string
    {
        return $this->token;
    }

    /**
     * Get WebSocket URL for a path.
     */
    public function getWebSocketUrl(string $path): string
    {
        $parsed = parse_url($this->baseUrl);
        $scheme = ($parsed['scheme'] ?? 'http') === 'https' ? 'wss' : 'ws';
        $host = $parsed['host'] ?? 'localhost';
        $port = $parsed['port'] ?? (($parsed['scheme'] ?? 'http') === 'https' ? 443 : 80);

        return "{$scheme}://{$host}:{$port}{$path}";
    }

    /**
     * Make an API request.
     *
     * @throws RexecException
     */
    public function request(string $method, string $path, ?array $body = null): mixed
    {
        try {
            $options = [];
            if ($body !== null) {
                $options['json'] = $body;
            }

            $response = $this->httpClient->request($method, $path, $options);
            $content = $response->getBody()->getContents();

            if (empty($content)) {
                return null;
            }

            return json_decode($content, true);
        } catch (GuzzleException $e) {
            $statusCode = method_exists($e, 'getCode') ? $e->getCode() : 0;
            throw new RexecException($e->getMessage(), $statusCode, $e);
        }
    }

    /**
     * Make a request and return raw bytes.
     *
     * @throws RexecException
     */
    public function requestBytes(string $method, string $path): string
    {
        try {
            $response = $this->httpClient->request($method, $path);
            return $response->getBody()->getContents();
        } catch (GuzzleException $e) {
            throw new RexecException($e->getMessage(), 0, $e);
        }
    }
}
