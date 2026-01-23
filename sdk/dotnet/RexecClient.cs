using System.Net.Http.Headers;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Rexec;

/// <summary>
/// Main client for interacting with Rexec API.
/// </summary>
/// <example>
/// <code>
/// var client = new RexecClient("https://your-instance.com", "your-token");
/// var container = await client.Containers.CreateAsync("ubuntu:24.04");
/// await client.Containers.StartAsync(container.Id);
/// </code>
/// </example>
public class RexecClient : IDisposable
{
    private readonly HttpClient _httpClient;
    private readonly string _baseUrl;
    private readonly string _token;
    private readonly JsonSerializerOptions _jsonOptions;
    private bool _disposed;

    /// <summary>
    /// Container management service.
    /// </summary>
    public ContainerService Containers { get; }

    /// <summary>
    /// File operations service.
    /// </summary>
    public FileService Files { get; }

    /// <summary>
    /// Terminal connection service.
    /// </summary>
    public TerminalService Terminal { get; }

    /// <summary>
    /// Create a new Rexec client.
    /// </summary>
    /// <param name="baseUrl">Base URL of your Rexec instance</param>
    /// <param name="token">API token for authentication</param>
    public RexecClient(string baseUrl, string token)
    {
        _baseUrl = baseUrl.TrimEnd('/');
        _token = token;

        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
        };

        _httpClient = new HttpClient
        {
            BaseAddress = new Uri(_baseUrl),
            Timeout = TimeSpan.FromSeconds(30)
        };
        _httpClient.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", token);
        _httpClient.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));

        Containers = new ContainerService(this);
        Files = new FileService(this);
        Terminal = new TerminalService(this);
    }

    internal string BaseUrl => _baseUrl;
    internal string Token => _token;
    internal JsonSerializerOptions JsonOptions => _jsonOptions;

    internal string GetWebSocketUrl(string path)
    {
        var uri = new Uri(_baseUrl);
        var wsScheme = uri.Scheme == "https" ? "wss" : "ws";
        var port = uri.Port != -1 ? uri.Port : (uri.Scheme == "https" ? 443 : 80);
        return $"{wsScheme}://{uri.Host}:{port}{path}";
    }

    internal async Task<T?> RequestAsync<T>(HttpMethod method, string path, object? body = null, CancellationToken cancellationToken = default)
    {
        var request = new HttpRequestMessage(method, path);

        if (body != null)
        {
            request.Content = JsonContent.Create(body, options: _jsonOptions);
        }

        var response = await _httpClient.SendAsync(request, cancellationToken);

        if (!response.IsSuccessStatusCode)
        {
            var errorBody = await response.Content.ReadAsStringAsync(cancellationToken);
            var errorMessage = ExtractErrorMessage(errorBody);
            throw new RexecException((int)response.StatusCode, errorMessage);
        }

        if (typeof(T) == typeof(object) || response.Content.Headers.ContentLength == 0)
        {
            return default;
        }

        return await response.Content.ReadFromJsonAsync<T>(_jsonOptions, cancellationToken);
    }

    internal async Task<byte[]> RequestBytesAsync(HttpMethod method, string path, CancellationToken cancellationToken = default)
    {
        var request = new HttpRequestMessage(method, path);
        var response = await _httpClient.SendAsync(request, cancellationToken);

        if (!response.IsSuccessStatusCode)
        {
            throw new RexecException((int)response.StatusCode, "Download failed");
        }

        return await response.Content.ReadAsByteArrayAsync(cancellationToken);
    }

    private string ExtractErrorMessage(string body)
    {
        try
        {
            var error = JsonSerializer.Deserialize<ErrorResponse>(body, _jsonOptions);
            if (!string.IsNullOrEmpty(error?.Error))
            {
                return error.Error;
            }
        }
        catch { }

        return string.IsNullOrEmpty(body) ? "Unknown error" : body;
    }

    public void Dispose()
    {
        if (!_disposed)
        {
            _httpClient.Dispose();
            _disposed = true;
        }
        GC.SuppressFinalize(this);
    }

    private class ErrorResponse
    {
        public string? Error { get; set; }
        public string? Message { get; set; }
    }
}
