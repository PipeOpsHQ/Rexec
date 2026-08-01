package io.pipeops.rexec;

import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;

import java.net.URI;
import java.nio.ByteBuffer;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

/**
 * Service for terminal WebSocket connections.
 */
public class TerminalService {
    private final RexecClient client;

    TerminalService(RexecClient client) {
        this.client = client;
    }

    /**
     * Connect to a container's terminal.
     *
     * @param containerId Container ID
     */
    public Terminal connect(String containerId) throws RexecException {
        return connect(containerId, 80, 24);
    }

    /**
     * Connect to a container's terminal with specified size.
     *
     * @param containerId Container ID
     * @param cols        Terminal columns
     * @param rows        Terminal rows
     */
    public Terminal connect(String containerId, int cols, int rows) throws RexecException {
        String url = client.getWebSocketUrl("/ws/terminal/" + containerId + "?cols=" + cols + "&rows=" + rows);
        return new Terminal(url, client.getToken());
    }
}

/**
 * Interactive terminal connection to a container.
 */
class Terminal {
    private final RexecWebSocketClient wsClient;
    private Consumer<String> onData;
    private Consumer<byte[]> onBinaryData;
    private Runnable onClose;
    private Consumer<Exception> onError;
    private volatile boolean connected = false;

    Terminal(String url, String token) throws RexecException {
        Map<String, String> headers = new HashMap<>();
        headers.put("Authorization", "Bearer " + token);

        try {
            URI uri = new URI(url);
            CountDownLatch connectLatch = new CountDownLatch(1);
            final Exception[] connectError = {null};

            this.wsClient = new RexecWebSocketClient(uri, headers) {
                @Override
                public void onOpen(ServerHandshake handshake) {
                    connected = true;
                    connectLatch.countDown();
                }

                @Override
                public void onMessage(String message) {
                    if (onData != null) {
                        onData.accept(message);
                    }
                }

                @Override
                public void onMessage(ByteBuffer bytes) {
                    if (onBinaryData != null) {
                        byte[] data = new byte[bytes.remaining()];
                        bytes.get(data);
                        onBinaryData.accept(data);
                    } else if (onData != null) {
                        byte[] data = new byte[bytes.remaining()];
                        bytes.get(data);
                        onData.accept(new String(data));
                    }
                }

                @Override
                public void onClose(int code, String reason, boolean remote) {
                    connected = false;
                    if (onClose != null) {
                        onClose.run();
                    }
                    connectLatch.countDown();
                }

                @Override
                public void onError(Exception ex) {
                    connectError[0] = ex;
                    if (onError != null) {
                        onError.accept(ex);
                    }
                    connectLatch.countDown();
                }
            };

            wsClient.connect();

            // Wait for connection
            if (!connectLatch.await(10, TimeUnit.SECONDS)) {
                throw new RexecException("Connection timeout");
            }

            if (connectError[0] != null) {
                throw new RexecException("Connection failed", connectError[0]);
            }

            if (!connected) {
                throw new RexecException("Failed to connect to terminal");
            }

        } catch (Exception e) {
            if (e instanceof RexecException) {
                throw (RexecException) e;
            }
            throw new RexecException("Failed to create terminal connection", e);
        }
    }

    /**
     * Write data to the terminal.
     */
    public void write(String data) {
        if (connected && wsClient != null) {
            wsClient.send(data);
        }
    }

    /**
     * Write binary data to the terminal.
     */
    public void write(byte[] data) {
        if (connected && wsClient != null) {
            wsClient.send(data);
        }
    }

    /**
     * Resize the terminal.
     */
    public void resize(int cols, int rows) {
        String resizeMsg = String.format("{\"type\":\"resize\",\"cols\":%d,\"rows\":%d}", cols, rows);
        write(resizeMsg);
    }

    /**
     * Set handler for text data received.
     */
    public Terminal onData(Consumer<String> handler) {
        this.onData = handler;
        return this;
    }

    /**
     * Set handler for binary data received.
     */
    public Terminal onBinaryData(Consumer<byte[]> handler) {
        this.onBinaryData = handler;
        return this;
    }

    /**
     * Set handler for close event.
     */
    public Terminal onClose(Runnable handler) {
        this.onClose = handler;
        return this;
    }

    /**
     * Set handler for errors.
     */
    public Terminal onError(Consumer<Exception> handler) {
        this.onError = handler;
        return this;
    }

    /**
     * Check if the terminal is connected.
     */
    public boolean isConnected() {
        return connected;
    }

    /**
     * Close the terminal connection.
     */
    public void close() {
        connected = false;
        if (wsClient != null) {
            wsClient.close();
        }
    }
}

/**
 * Custom WebSocket client with authorization header support.
 */
abstract class RexecWebSocketClient extends WebSocketClient {
    RexecWebSocketClient(URI serverUri, Map<String, String> httpHeaders) {
        super(serverUri, httpHeaders);
    }
}
