package io.pipeops.rexec;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.reflect.TypeToken;

import java.lang.reflect.Type;
import java.util.Collections;
import java.util.List;

/**
 * Service for managing sandboxes/containers.
 */
public class SandboxService {
    private static final Type CONTAINER_LIST_TYPE = new TypeToken<List<Sandbox>>() {}.getType();

    private final RexecClient client;

    SandboxService(RexecClient client) {
        this.client = client;
    }

    /**
     * List all sandboxes.
     * Handles {@code {containers:[...],count,limit}} (including null list) and a bare array.
     * Wire path remains {@code /api/containers}.
     */
    public List<Sandbox> list() throws RexecException {
        String raw = client.requestRaw("GET", "/api/containers", null);
        if (raw == null || raw.isBlank()) {
            return Collections.emptyList();
        }

        JsonElement root = JsonParser.parseString(raw);
        if (root.isJsonArray()) {
            return client.getGson().fromJson(root, CONTAINER_LIST_TYPE);
        }
        if (root.isJsonObject()) {
            JsonObject obj = root.getAsJsonObject();
            if (!obj.has("containers") || obj.get("containers").isJsonNull()) {
                return Collections.emptyList();
            }
            JsonElement containers = obj.get("containers");
            if (containers.isJsonArray()) {
                return client.getGson().fromJson(containers, CONTAINER_LIST_TYPE);
            }
        }
        return Collections.emptyList();
    }

    /**
     * Get a sandbox by ID.
     */
    public Sandbox get(String sandboxId) throws RexecException {
        return client.request("GET", "/api/containers/" + sandboxId, null, Sandbox.class);
    }

    /**
     * Create a new sandbox.
     */
    public Sandbox create(CreateSandboxRequest request) throws RexecException {
        return client.request("POST", "/api/containers", request, Sandbox.class);
    }

    /**
     * Create a new sandbox with just an image alias.
     */
    public Sandbox create(String image) throws RexecException {
        return create(new CreateSandboxRequest(image));
    }

    /**
     * Start a sandbox.
     */
    public void start(String sandboxId) throws RexecException {
        client.request("POST", "/api/containers/" + sandboxId + "/start", null, Void.class);
    }

    /**
     * Stop a sandbox.
     */
    public void stop(String sandboxId) throws RexecException {
        client.request("POST", "/api/containers/" + sandboxId + "/stop", null, Void.class);
    }

    /**
     * Delete a sandbox.
     */
    public void delete(String sandboxId) throws RexecException {
        client.request("DELETE", "/api/containers/" + sandboxId, null, Void.class);
    }
}
