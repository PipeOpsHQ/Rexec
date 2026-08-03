<?php

declare(strict_types=1);

namespace Rexec;

use GuzzleHttp\Client as HttpClient;
use GuzzleHttp\Exception\GuzzleException;
use GuzzleHttp\Exception\RequestException;

/**
 * Main client for interacting with the Rexec API.
 *
 * @example
 * $client = new RexecClient('https://rexec.sh', $token);
 * $list = $client->containers()->list();
 * $container = $client->containers()->create('ubuntu', ['name' => 'demo']);
 * $client->containers()->delete($container->id);
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
     * @param string $baseUrl Base URL of your Rexec instance
     * @param string $token   API token (Bearer)
     * @param array  $options Additional Guzzle options
     */
    public function __construct(string $baseUrl, string $token, array $options = [])
    {
        $this->baseUrl = rtrim($baseUrl, '/');
        $this->token = $token;

        $defaultOptions = [
            'base_uri' => $this->baseUrl . '/',
            'timeout' => 30,
            'http_errors' => false,
            'headers' => [
                'Authorization' => 'Bearer ' . $token,
                'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'User-Agent' => 'pipeops-rexec-php/1.0.1',
            ],
        ];

        $this->httpClient = new HttpClient(array_merge($defaultOptions, $options));

        $this->containers = new ContainerService($this);
        $this->files = new FileService($this);
        $this->terminal = new TerminalService($this);
    }

    public function containers(): ContainerService
    {
        return $this->containers;
    }

    public function files(): FileService
    {
        return $this->files;
    }

    public function terminal(): TerminalService
    {
        return $this->terminal;
    }

    public function getBaseUrl(): string
    {
        return $this->baseUrl;
    }

    public function getToken(): string
    {
        return $this->token;
    }

    /**
     * Build a WebSocket URL for a path (e.g. /ws/terminal/{id}).
     */
    public function getWebSocketUrl(string $path): string
    {
        $parsed = parse_url($this->baseUrl);
        $scheme = ($parsed['scheme'] ?? 'http') === 'https' ? 'wss' : 'ws';
        $host = $parsed['host'] ?? 'localhost';
        $port = $parsed['port'] ?? null;
        $portPart = $port !== null ? ':' . $port : '';
        if ($path === '' || $path[0] !== '/') {
            $path = '/' . $path;
        }

        return "{$scheme}://{$host}{$portPart}{$path}";
    }

    /**
     * Make a JSON API request.
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

            $response = $this->httpClient->request($method, ltrim($path, '/'), $options);
            $status = $response->getStatusCode();
            $content = (string) $response->getBody();

            if ($status < 200 || $status >= 300) {
                throw RexecException::fromResponse($status, $content);
            }

            if ($content === '') {
                return null;
            }

            $decoded = json_decode($content, true);
            if (json_last_error() !== JSON_ERROR_NONE) {
                throw new RexecException('Invalid JSON response: ' . json_last_error_msg(), $status);
            }

            return $decoded;
        } catch (RexecException $e) {
            throw $e;
        } catch (RequestException $e) {
            $status = $e->getResponse()?->getStatusCode() ?? 0;
            $body = $e->getResponse() ? (string) $e->getResponse()->getBody() : $e->getMessage();
            throw RexecException::fromResponse($status, $body, $e);
        } catch (GuzzleException $e) {
            throw new RexecException($e->getMessage(), 0, $e);
        }
    }

    /**
     * Make a request and return raw response bytes/string.
     *
     * @throws RexecException
     */
    public function requestBytes(string $method, string $path): string
    {
        try {
            $response = $this->httpClient->request($method, ltrim($path, '/'));
            $status = $response->getStatusCode();
            $content = (string) $response->getBody();

            if ($status < 200 || $status >= 300) {
                throw RexecException::fromResponse($status, $content);
            }

            return $content;
        } catch (RexecException $e) {
            throw $e;
        } catch (GuzzleException $e) {
            throw new RexecException($e->getMessage(), 0, $e);
        }
    }
}
