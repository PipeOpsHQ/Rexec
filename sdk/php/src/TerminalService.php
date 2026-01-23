<?php

declare(strict_types=1);

namespace Rexec;

use Ratchet\Client\WebSocket;
use Ratchet\Client\Connector;
use React\EventLoop\Loop;

/**
 * Service for terminal WebSocket connections.
 */
class TerminalService
{
    private RexecClient $client;

    public function __construct(RexecClient $client)
    {
        $this->client = $client;
    }

    /**
     * Connect to a container's terminal.
     *
     * @param string $containerId Container ID
     * @param int $cols Terminal columns
     * @param int $rows Terminal rows
     */
    public function connect(string $containerId, int $cols = 80, int $rows = 24): Terminal
    {
        $url = $this->client->getWebSocketUrl("/ws/terminal/{$containerId}?cols={$cols}&rows={$rows}");
        return new Terminal($url, $this->client->getToken());
    }
}

/**
 * Interactive terminal connection to a container.
 */
class Terminal
{
    private string $url;
    private string $token;
    private ?WebSocket $connection = null;
    private bool $connected = false;

    /** @var callable|null */
    private $onData = null;
    /** @var callable|null */
    private $onClose = null;
    /** @var callable|null */
    private $onError = null;

    public function __construct(string $url, string $token)
    {
        $this->url = $url;
        $this->token = $token;
    }

    /**
     * Open the WebSocket connection.
     *
     * @throws RexecException
     */
    public function open(): void
    {
        $loop = Loop::get();
        $connector = new Connector($loop);

        $connector($this->url, [], [
            'Authorization' => 'Bearer ' . $this->token,
        ])->then(
            function (WebSocket $conn) {
                $this->connection = $conn;
                $this->connected = true;

                $conn->on('message', function ($msg) {
                    if ($this->onData) {
                        ($this->onData)((string)$msg);
                    }
                });

                $conn->on('close', function () {
                    $this->connected = false;
                    if ($this->onClose) {
                        ($this->onClose)();
                    }
                });
            },
            function (\Exception $e) {
                if ($this->onError) {
                    ($this->onError)($e);
                }
                throw new RexecException('Failed to connect: ' . $e->getMessage());
            }
        );
    }

    /**
     * Write data to the terminal.
     */
    public function write(string $data): void
    {
        if ($this->connected && $this->connection) {
            $this->connection->send($data);
        }
    }

    /**
     * Resize the terminal.
     */
    public function resize(int $cols, int $rows): void
    {
        $this->write(json_encode([
            'type' => 'resize',
            'cols' => $cols,
            'rows' => $rows,
        ]));
    }

    /**
     * Set handler for data received.
     */
    public function onData(callable $handler): self
    {
        $this->onData = $handler;
        return $this;
    }

    /**
     * Set handler for close event.
     */
    public function onClose(callable $handler): self
    {
        $this->onClose = $handler;
        return $this;
    }

    /**
     * Set handler for errors.
     */
    public function onError(callable $handler): self
    {
        $this->onError = $handler;
        return $this;
    }

    /**
     * Check if the terminal is connected.
     */
    public function isConnected(): bool
    {
        return $this->connected;
    }

    /**
     * Close the terminal connection.
     */
    public function close(): void
    {
        $this->connected = false;
        if ($this->connection) {
            $this->connection->close();
            $this->connection = null;
        }
    }
}
