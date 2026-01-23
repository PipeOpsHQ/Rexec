using System.Text.Json.Serialization;

namespace Rexec;

/// <summary>
/// Represents a Rexec container/sandbox.
/// </summary>
public class Container
{
    /// <summary>
    /// Unique container identifier.
    /// </summary>
    public string Id { get; set; } = "";

    /// <summary>
    /// Container name.
    /// </summary>
    public string? Name { get; set; }

    /// <summary>
    /// Docker image used.
    /// </summary>
    public string Image { get; set; } = "";

    /// <summary>
    /// Current status (created, running, stopped).
    /// </summary>
    public string Status { get; set; } = "";

    /// <summary>
    /// When the container was created.
    /// </summary>
    public string? CreatedAt { get; set; }

    /// <summary>
    /// When the container was started.
    /// </summary>
    public string? StartedAt { get; set; }

    /// <summary>
    /// Container labels.
    /// </summary>
    public Dictionary<string, string>? Labels { get; set; }

    /// <summary>
    /// Environment variables.
    /// </summary>
    public Dictionary<string, string>? Environment { get; set; }

    /// <summary>
    /// Whether the container is running.
    /// </summary>
    [JsonIgnore]
    public bool IsRunning => Status == "running";

    /// <summary>
    /// Whether the container is stopped.
    /// </summary>
    [JsonIgnore]
    public bool IsStopped => Status == "stopped";
}

/// <summary>
/// Request to create a new container.
/// </summary>
public class CreateContainerRequest
{
    /// <summary>
    /// Docker image to use.
    /// </summary>
    public string Image { get; set; }

    /// <summary>
    /// Optional container name.
    /// </summary>
    public string? Name { get; set; }

    /// <summary>
    /// Environment variables.
    /// </summary>
    public Dictionary<string, string>? Environment { get; set; }

    /// <summary>
    /// Container labels.
    /// </summary>
    public Dictionary<string, string>? Labels { get; set; }

    /// <summary>
    /// Create a request with the specified image.
    /// </summary>
    public CreateContainerRequest(string image)
    {
        Image = image;
    }

    /// <summary>
    /// Set environment variable.
    /// </summary>
    public CreateContainerRequest WithEnv(string key, string value)
    {
        Environment ??= new Dictionary<string, string>();
        Environment[key] = value;
        return this;
    }

    /// <summary>
    /// Set label.
    /// </summary>
    public CreateContainerRequest WithLabel(string key, string value)
    {
        Labels ??= new Dictionary<string, string>();
        Labels[key] = value;
        return this;
    }
}

/// <summary>
/// Result of command execution.
/// </summary>
public class ExecResult
{
    /// <summary>
    /// Exit code from the command.
    /// </summary>
    public int ExitCode { get; set; }

    /// <summary>
    /// Standard output.
    /// </summary>
    public string Stdout { get; set; } = "";

    /// <summary>
    /// Standard error.
    /// </summary>
    public string Stderr { get; set; } = "";

    /// <summary>
    /// Whether the command succeeded.
    /// </summary>
    [JsonIgnore]
    public bool IsSuccess => ExitCode == 0;
}

/// <summary>
/// Information about a file in a container.
/// </summary>
public class FileInfo
{
    /// <summary>
    /// File name.
    /// </summary>
    public string Name { get; set; } = "";

    /// <summary>
    /// Full path.
    /// </summary>
    public string Path { get; set; } = "";

    /// <summary>
    /// File size in bytes.
    /// </summary>
    public long Size { get; set; }

    /// <summary>
    /// File mode/permissions.
    /// </summary>
    public string? Mode { get; set; }

    /// <summary>
    /// Last modification time.
    /// </summary>
    public string? ModTime { get; set; }

    /// <summary>
    /// Whether this is a directory.
    /// </summary>
    public bool IsDir { get; set; }

    /// <summary>
    /// Whether this is a file.
    /// </summary>
    [JsonIgnore]
    public bool IsFile => !IsDir;
}

internal class ContainerListResponse
{
    public List<Container>? Containers { get; set; }
}

internal class FileListResponse
{
    public List<FileInfo>? Files { get; set; }
}

internal class ExecRequest
{
    public string[] Command { get; set; } = Array.Empty<string>();
}

internal class WriteFileRequest
{
    public string Path { get; set; } = "";
    public string Content { get; set; } = "";
}
