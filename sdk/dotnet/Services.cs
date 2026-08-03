using System.Web;

namespace Rexec;

/// <summary>
/// Service for managing sandboxes. HTTP paths remain /api/containers.
/// </summary>
public class SandboxService
{
    private readonly RexecClient _client;

    internal SandboxService(RexecClient client)
    {
        _client = client;
    }

    /// <summary>
    /// List all sandboxes.
    /// </summary>
    public async Task<List<Sandbox>> ListAsync(CancellationToken cancellationToken = default)
    {
        var response = await _client.RequestAsync<ContainerListResponse>(HttpMethod.Get, "/api/containers", null, cancellationToken);
        return response?.Containers ?? new List<Sandbox>();
    }

    /// <summary>
    /// Get a sandbox by ID.
    /// </summary>
    public async Task<Sandbox?> GetAsync(string sandboxId, CancellationToken cancellationToken = default)
    {
        return await _client.RequestAsync<Sandbox>(HttpMethod.Get, $"/api/containers/{sandboxId}", null, cancellationToken);
    }

    /// <summary>
    /// Create a new sandbox.
    /// </summary>
    public async Task<Sandbox?> CreateAsync(CreateSandboxRequest request, CancellationToken cancellationToken = default)
    {
        return await _client.RequestAsync<Sandbox>(HttpMethod.Post, "/api/containers", request, cancellationToken);
    }

    /// <summary>
    /// Create a new sandbox with just an image alias.
    /// </summary>
    public Task<Sandbox?> CreateAsync(string image, CancellationToken cancellationToken = default)
    {
        return CreateAsync(new CreateSandboxRequest(image), cancellationToken);
    }

    /// <summary>
    /// Start a container.
    /// </summary>
    public async Task StartAsync(string containerId, CancellationToken cancellationToken = default)
    {
        await _client.RequestAsync<object>(HttpMethod.Post, $"/api/containers/{containerId}/start", null, cancellationToken);
    }

    /// <summary>
    /// Stop a container.
    /// </summary>
    public async Task StopAsync(string containerId, CancellationToken cancellationToken = default)
    {
        await _client.RequestAsync<object>(HttpMethod.Post, $"/api/containers/{containerId}/stop", null, cancellationToken);
    }

    /// <summary>
    /// Delete a container.
    /// </summary>
    public async Task DeleteAsync(string containerId, CancellationToken cancellationToken = default)
    {
        await _client.RequestAsync<object>(HttpMethod.Delete, $"/api/containers/{containerId}", null, cancellationToken);
    }

    /// <summary>
    /// Execute a command in a container.
    /// </summary>
    public async Task<ExecResult?> ExecAsync(string containerId, string[] command, CancellationToken cancellationToken = default)
    {
        var request = new ExecRequest { Command = command };
        return await _client.RequestAsync<ExecResult>(HttpMethod.Post, $"/api/containers/{containerId}/exec", request, cancellationToken);
    }

    /// <summary>
    /// Execute a shell command in a container.
    /// </summary>
    public Task<ExecResult?> ExecAsync(string containerId, string command, CancellationToken cancellationToken = default)
    {
        return ExecAsync(containerId, new[] { "/bin/sh", "-c", command }, cancellationToken);
    }
}

/// <summary>
/// Deprecated alias for <see cref="SandboxService"/>.
/// </summary>
[Obsolete("Use SandboxService / RexecClient.Sandboxes instead")]
public class ContainerService : SandboxService
{
    internal ContainerService(RexecClient client) : base(client) { }
}

/// <summary>
/// Service for file operations within sandboxes.
/// </summary>
public class FileService
{
    private readonly RexecClient _client;

    internal FileService(RexecClient client)
    {
        _client = client;
    }

    /// <summary>
    /// List files in a directory.
    /// </summary>
    public async Task<List<FileInfo>> ListAsync(string containerId, string path, CancellationToken cancellationToken = default)
    {
        var encodedPath = HttpUtility.UrlEncode(path);
        var response = await _client.RequestAsync<FileListResponse>(
            HttpMethod.Get,
            $"/api/containers/{containerId}/files/list?path={encodedPath}",
            null,
            cancellationToken
        );
        return response?.Files ?? new List<FileInfo>();
    }

    /// <summary>
    /// Read a file's contents.
    /// </summary>
    public async Task<byte[]> ReadAsync(string containerId, string path, CancellationToken cancellationToken = default)
    {
        var encodedPath = HttpUtility.UrlEncode(path);
        return await _client.RequestBytesAsync(HttpMethod.Get, $"/api/containers/{containerId}/files?path={encodedPath}", cancellationToken);
    }

    /// <summary>
    /// Read a file as string.
    /// </summary>
    public async Task<string> ReadStringAsync(string containerId, string path, CancellationToken cancellationToken = default)
    {
        var bytes = await ReadAsync(containerId, path, cancellationToken);
        return System.Text.Encoding.UTF8.GetString(bytes);
    }

    /// <summary>
    /// Write content to a file.
    /// </summary>
    public async Task WriteAsync(string containerId, string path, byte[] content, CancellationToken cancellationToken = default)
    {
        var request = new WriteFileRequest
        {
            Path = path,
            Content = Convert.ToBase64String(content)
        };
        await _client.RequestAsync<object>(HttpMethod.Post, $"/api/containers/{containerId}/files", request, cancellationToken);
    }

    /// <summary>
    /// Write string content to a file.
    /// </summary>
    public Task WriteAsync(string containerId, string path, string content, CancellationToken cancellationToken = default)
    {
        return WriteAsync(containerId, path, System.Text.Encoding.UTF8.GetBytes(content), cancellationToken);
    }

    /// <summary>
    /// Delete a file.
    /// </summary>
    public async Task DeleteAsync(string containerId, string path, CancellationToken cancellationToken = default)
    {
        var encodedPath = HttpUtility.UrlEncode(path);
        await _client.RequestAsync<object>(HttpMethod.Delete, $"/api/containers/{containerId}/files?path={encodedPath}", null, cancellationToken);
    }
}
