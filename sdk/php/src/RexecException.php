<?php

declare(strict_types=1);

namespace Rexec;

use Exception;

/**
 * Exception thrown by Rexec SDK operations.
 */
class RexecException extends Exception
{
    private int $statusCode;

    /**
     * Create a new exception.
     */
    public function __construct(string $message, int $statusCode = 0, ?Exception $previous = null)
    {
        $this->statusCode = $statusCode;
        $formattedMessage = $statusCode > 0 ? "API error {$statusCode}: {$message}" : $message;
        parent::__construct($formattedMessage, $statusCode, $previous);
    }

    /**
     * Get the HTTP status code.
     */
    public function getStatusCode(): int
    {
        return $this->statusCode;
    }

    /**
     * Check if this is an API error.
     */
    public function isApiError(): bool
    {
        return $this->statusCode > 0;
    }
}
