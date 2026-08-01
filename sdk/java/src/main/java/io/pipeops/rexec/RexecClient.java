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
 * RexecClient client = new RexecClient("https://your-instance.com", "your-token");
 *
 * Container container = client.containers().create(
 *     new CreateContainerRequest("ubuntu:24.04").setName("my-sandbox")
 * );
 *
 * Terminal terminal = client.terminal().connect(container.getId());
 * terminal.write("echo hello\n");
 * }</pre>
 */
public class RexecClient {
    private final String baseUrl;
    private final String token;
    private final OkHttpClient httpClient;
    private final Gson gson;

    private final ContainerService containers;
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
                            .header("Accept", "application/json");
                    return chain.proceed(builder.build());
                })
                .build();

        this.containers = new ContainerService(this);
        this.files = new FileService(this);
        this.terminal = new TerminalService(this);
    }

    /**
     * Get the container service.
     */
    public ContainerService containers() {
        return containers;
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
        String wsScheme = uri.getScheme().equals("https") ? "wss" : "ws";
        int port = uri.getPort() != -1 ? uri.getPort() : (uri.getScheme().equals("https") ? 443 : 80);
        return wsScheme + "://" + uri.getHost() + ":" + port + path;
    }

    /**
     * Make an API request.
     */
    <T> T request(String method, String path, Object body, Class<T> responseType) throws RexecException {
        String url = baseUrl + path;

        Request.Builder builder = new Request.Builder().url(url);

        if (body != null) {
            String json = gson.toJson(body);
            RequestBody requestBody = RequestBody.create(json, MediaType.parse("application/json"));
            builder.method(method, requestBody);
        } else if (method.equals("POST") || method.equals("PUT") || method.equals("PATCH")) {
            builder.method(method, RequestBody.create("", null));
        } else {
            builder.method(method, null);
        }

        try (Response response = httpClient.newCall(builder.build()).execute()) {
            if (!response.isSuccessful()) {
                String errorBody = response.body() != null ? response.body().string() : "";
                throw new RexecException(response.code(), extractErrorMessage(errorBody));
            }

            if (responseType == Void.class || response.body() == null) {
                return null;
            }

            return gson.fromJson(response.body().string(), responseType);
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
