<?php

declare(strict_types=1);

namespace Rexec;

use Exception;
use Throwable;

/**
 * Exception thrown by Rexec SDK operations.
 */
class RexecException extends Exception
{
    private int $statusCode;

    public function __construct(string $message, int $statusCode = 0, ?Throwable $previous = null)
    {
        $this->statusCode = $statusCode;
        $formattedMessage = $statusCode > 0 ? "API error {$statusCode}: {$message}" : $message;
        parent::__construct($formattedMessage, $statusCode, $previous);
    }

    public static function fromResponse(int $statusCode, string $body, ?Throwable $previous = null): self
    {
        $message = $body;
        $decoded = json_decode($body, true);
        if (is_array($decoded)) {
            $message = (string) ($decoded['error'] ?? $decoded['message'] ?? $body);
        }
        if ($message === '') {
            $message = 'Unknown error';
        }

        return new self($message, $statusCode, $previous);
    }

    public function getStatusCode(): int
    {
        return $this->statusCode;
    }

    public function isApiError(): bool
    {
        return $this->statusCode > 0;
    }
}
