namespace Rexec;

/// <summary>
/// Exception thrown by Rexec SDK operations.
/// </summary>
public class RexecException : Exception
{
    /// <summary>
    /// HTTP status code if this is an API error.
    /// </summary>
    public int StatusCode { get; }

    /// <summary>
    /// Whether this is an API error (has status code).
    /// </summary>
    public bool IsApiError => StatusCode > 0;

    /// <summary>
    /// Create an exception with status code and message.
    /// </summary>
    public RexecException(int statusCode, string message)
        : base($"API error {statusCode}: {message}")
    {
        StatusCode = statusCode;
    }

    /// <summary>
    /// Create an exception with message only.
    /// </summary>
    public RexecException(string message) : base(message)
    {
        StatusCode = -1;
    }

    /// <summary>
    /// Create an exception with message and cause.
    /// </summary>
    public RexecException(string message, Exception innerException)
        : base(message, innerException)
    {
        StatusCode = -1;
    }
}
