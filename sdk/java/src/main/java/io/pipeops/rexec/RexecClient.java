package io.pipeops.rexec;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import okhttp3.*;

import java.io.IOException;
import java.net.URI;
import java.util.concurrent.TimeUnit;

/**
 * Main client for interacting with Rexec API.
 *
 * <pre>{@code
 * RexecClient client = new RexecClient("https://rexec.sh", token);
 *
 * Sandbox sandbox = client.sandboxes().create(
 *     new CreateSandboxRequest("ubuntu").setName("my-sandbox")
 * );
 * // Legacy: client.containers() is the same service
 *
 * client.sandboxes().delete(sandbox.getId());
 * }</pre>
 */
public class RexecClient {
    private final String baseUrl;
    private final String token;
    private final OkHttpClient httpClient;
    private final Gson gson;

    private final SandboxService sandboxes;
    private final FileService files;
    private final TerminalService terminal;

    /**
     * Create a new Rexec client.
     *
     * @param baseUrl Base URL of your Rexec instance
     * @param token   API token for authentication
     */
    public RexecClient(String baseUrl, String token) {
        this(baseUrl, token, 30);
    }

    /**
     * Create a new Rexec client with custom timeout.
     *
     * @param baseUrl        Base URL of your Rexec instance
     * @param token          API token for authentication
     * @param timeoutSeconds Request timeout in seconds
     */
    public RexecClient(String baseUrl, String token, int timeoutSeconds) {
        this.baseUrl = baseUrl.replaceAll("/$", "");
        this.token = token;
        this.gson = new GsonBuilder().create();

        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(timeoutSeconds, TimeUnit.SECONDS)
                .readTimeout(timeoutSeconds, TimeUnit.SECONDS)
                .writeTimeout(timeoutSeconds, TimeUnit.SECONDS)
                .addInterceptor(chain -> {
                    Request original = chain.request();
                    Request.Builder builder = original.newBuilder()
                            .header("Authorization", "Bearer " + token)
                            .header("Accept", "application/json")
                            .header("User-Agent", "pipeops-rexec-java/1.1.0");
                    return chain.proceed(builder.build());
                })
                .build();

        this.sandboxes = new SandboxService(this);
        this.files = new FileService(this);
        this.terminal = new TerminalService(this);
    }

    /**
     * Get the sandbox service (preferred).
     */
    public SandboxService sandboxes() {
        return sandboxes;
    }

    /**
     * Get the sandbox service.
     * @deprecated use {@link #sandboxes()}
     */
    @Deprecated
    public SandboxService containers() {
        return sandboxes;
    }

    /**
     * Get the file service.
     */
    public FileService files() {
        return files;
    }

    /**
     * Get the terminal service.
     */
    public TerminalService terminal() {
        return terminal;
    }

    // Internal methods

    String getBaseUrl() {
        return baseUrl;
    }

    String getToken() {
        return token;
    }

    OkHttpClient getHttpClient() {
        return httpClient;
    }

    Gson getGson() {
        return gson;
    }

    /**
     * Get WebSocket URL for a path.
     */
    String getWebSocketUrl(String path) {
        URI uri = URI.create(baseUrl);
        String wsScheme = "https".equals(uri.getScheme()) ? "wss" : "ws";
        StringBuilder sb = new StringBuilder();
        sb.append(wsScheme).append("://").append(uri.getHost());
        if (uri.getPort() != -1) {
            sb.append(':').append(uri.getPort());
        }
        sb.append(path);
        return sb.toString();
    }

    /**
     * Make an API request and parse JSON into {@code responseType}.
     */
    <T> T request(String method, String path, Object body, Class<T> responseType) throws RexecException {
        String raw = requestRaw(method, path, body);
        if (responseType == Void.class || raw == null || raw.isBlank()) {
            return null;
        }
        return gson.fromJson(raw, responseType);
    }

    /**
     * Make an API request and return the raw response body.
     */
    String requestRaw(String method, String path, Object body) throws RexecException {
        String url = baseUrl + path;

        Request.Builder builder = new Request.Builder().url(url);

        if (body != null) {
            String json = gson.toJson(body);
            RequestBody requestBody = RequestBody.create(json, MediaType.parse("application/json"));
            builder.method(method, requestBody);
        } else if (method.equals("POST") || method.equals("PUT") || method.equals("PATCH")) {
            builder.method(method, RequestBody.create(new byte[0], MediaType.parse("application/json")));
        } else {
            builder.method(method, null);
        }

        try (Response response = httpClient.newCall(builder.build()).execute()) {
            String responseBody = response.body() != null ? response.body().string() : "";
            if (!response.isSuccessful()) {
                throw new RexecException(response.code(), extractErrorMessage(responseBody));
            }
            return responseBody;
        } catch (IOException e) {
            throw new RexecException("Request failed: " + e.getMessage(), e);
        }
    }

    /**
     * Make a request and return raw bytes.
     */
    byte[] requestBytes(String method, String path) throws RexecException {
        String url = baseUrl + path;

        Request request = new Request.Builder()
                .url(url)
                .method(method, null)
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new RexecException(response.code(), "Download failed");
            }

            return response.body() != null ? response.body().bytes() : new byte[0];
        } catch (IOException e) {
            throw new RexecException("Request failed: " + e.getMessage(), e);
        }
    }

    private String extractErrorMessage(String body) {
        try {
            ErrorResponse error = gson.fromJson(body, ErrorResponse.class);
            if (error != null && error.error != null) {
                return error.error;
            }
        } catch (Exception ignored) {
        }
        return body.isEmpty() ? "Unknown error" : body;
    }

    private static class ErrorResponse {
        String error;
        String message;
    }
}
