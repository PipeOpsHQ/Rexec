package io.pipeops.rexec;

/**
 * Exception thrown by Rexec SDK operations.
 */
public class RexecException extends Exception {
    private final int statusCode;

    /**
     * Create an exception with status code and message.
     */
    public RexecException(int statusCode, String message) {
        super("API error " + statusCode + ": " + message);
        this.statusCode = statusCode;
    }

    /**
     * Create an exception with message only.
     */
    public RexecException(String message) {
        super(message);
        this.statusCode = -1;
    }

    /**
     * Create an exception with message and cause.
     */
    public RexecException(String message, Throwable cause) {
        super(message, cause);
        this.statusCode = -1;
    }

    /**
     * Get the HTTP status code, or -1 if not an API error.
     */
    public int getStatusCode() {
        return statusCode;
    }

    /**
     * Check if this is an API error (has status code).
     */
    public boolean isApiError() {
        return statusCode > 0;
    }
}
