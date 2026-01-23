using System.Net.WebSockets;
using System.Text;
using System.Text.Json;

namespace Rexec;

/// <summary>
/// Service for terminal WebSocket connections.
/// </summary>
public class TerminalService
{
    private readonly RexecClient _client;

    internal TerminalService(RexecClient client)
    {
        _client = client;
    }

    /// <summary>
    /// Connect to a container's terminal.
    /// </summary>
    public Task<Terminal> ConnectAsync(string containerId, CancellationToken cancellationToken = default)
    {
        return ConnectAsync(containerId, 80, 24, cancellationToken);
    }

    /// <summary>
    /// Connect to a container's terminal with specified size.
    /// </summary>
    public async Task<Terminal> ConnectAsync(string containerId, int cols, int rows, CancellationToken cancellationToken = default)
    {
        var url = _client.GetWebSocketUrl($"/ws/terminal/{containerId}?cols={cols}&rows={rows}");
        var terminal = new Terminal(url, _client.Token);
        await terminal.ConnectAsync(cancellationToken);
        return terminal;
    }
}

/// <summary>
/// Interactive terminal connection to a container.
/// </summary>
public class Terminal : IAsyncDisposable
{
    private readonly ClientWebSocket _webSocket;
    private readonly string _url;
    private readonly string _token;
    private CancellationTokenSource? _receiveCts;
    private Task? _receiveTask;

    /// <summary>
    /// Event raised when data is received.
    /// </summary>
    public event Action<string>? OnData;

    /// <summary>
    /// Event raised when binary data is received.
    /// </summary>
    public event Action<byte[]>? OnBinaryData;

    /// <summary>
    /// Event raised when the connection closes.
    /// </summary>
    public event Action? OnClose;

    /// <summary>
    /// Event raised when an error occurs.
    /// </summary>
    public event Action<Exception>? OnError;

    /// <summary>
    /// Whether the terminal is connected.
    /// </summary>
    public bool IsConnected => _webSocket.State == WebSocketState.Open;

    internal Terminal(string url, string token)
    {
        _url = url;
        _token = token;
        _webSocket = new ClientWebSocket();
    }

    internal async Task ConnectAsync(CancellationToken cancellationToken)
    {
        _webSocket.Options.SetRequestHeader("Authorization", $"Bearer {_token}");

        try
        {
            await _webSocket.ConnectAsync(new Uri(_url), cancellationToken);
            _receiveCts = new CancellationTokenSource();
            _receiveTask = ReceiveLoopAsync(_receiveCts.Token);
        }
        catch (Exception ex)
        {
            throw new RexecException("Failed to connect to terminal", ex);
        }
    }

    /// <summary>
    /// Write data to the terminal.
    /// </summary>
    public async Task WriteAsync(string data, CancellationToken cancellationToken = default)
    {
        if (!IsConnected) return;

        var bytes = Encoding.UTF8.GetBytes(data);
        await _webSocket.SendAsync(new ArraySegment<byte>(bytes), WebSocketMessageType.Text, true, cancellationToken);
    }

    /// <summary>
    /// Write binary data to the terminal.
    /// </summary>
    public async Task WriteAsync(byte[] data, CancellationToken cancellationToken = default)
    {
        if (!IsConnected) return;

        await _webSocket.SendAsync(new ArraySegment<byte>(data), WebSocketMessageType.Binary, true, cancellationToken);
    }

    /// <summary>
    /// Resize the terminal.
    /// </summary>
    public Task ResizeAsync(int cols, int rows, CancellationToken cancellationToken = default)
    {
        var resizeMsg = JsonSerializer.Serialize(new { type = "resize", cols, rows });
        return WriteAsync(resizeMsg, cancellationToken);
    }

    /// <summary>
    /// Close the terminal connection.
    /// </summary>
    public async Task CloseAsync(CancellationToken cancellationToken = default)
    {
        _receiveCts?.Cancel();

        if (_webSocket.State == WebSocketState.Open)
        {
            try
            {
                await _webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing", cancellationToken);
            }
            catch { }
        }

        if (_receiveTask != null)
        {
            try
            {
                await _receiveTask;
            }
            catch { }
        }

        OnClose?.Invoke();
    }

    private async Task ReceiveLoopAsync(CancellationToken cancellationToken)
    {
        var buffer = new byte[4096];

        try
        {
            while (!cancellationToken.IsCancellationRequested && IsConnected)
            {
                var result = await _webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), cancellationToken);

                if (result.MessageType == WebSocketMessageType.Close)
                {
                    break;
                }

                var data = new byte[result.Count];
                Array.Copy(buffer, data, result.Count);

                if (result.MessageType == WebSocketMessageType.Text)
                {
                    OnData?.Invoke(Encoding.UTF8.GetString(data));
                }
                else
                {
                    OnBinaryData?.Invoke(data);
                }
            }
        }
        catch (OperationCanceledException) { }
        catch (Exception ex)
        {
            OnError?.Invoke(ex);
        }
        finally
        {
            OnClose?.Invoke();
        }
    }

    public async ValueTask DisposeAsync()
    {
        await CloseAsync();
        _webSocket.Dispose();
        _receiveCts?.Dispose();
        GC.SuppressFinalize(this);
    }
}
