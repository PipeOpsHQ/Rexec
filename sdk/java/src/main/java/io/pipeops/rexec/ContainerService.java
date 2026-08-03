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
public class ContainerService {
    private static final Type CONTAINER_LIST_TYPE = new TypeToken<List<Container>>() {}.getType();

    private final RexecClient client;

    ContainerService(RexecClient client) {
        this.client = client;
    }

    /**
     * List all containers.
     * Handles {@code {containers:[...],count,limit}} (including null list) and a bare array.
     */
    public List<Container> list() throws RexecException {
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
     * Get a container by ID.
     */
    public Container get(String containerId) throws RexecException {
        return client.request("GET", "/api/containers/" + containerId, null, Container.class);
    }

    /**
     * Create a new container.
     */
    public Container create(CreateContainerRequest request) throws RexecException {
        return client.request("POST", "/api/containers", request, Container.class);
    }

    /**
     * Create a new container with just an image alias.
     */
    public Container create(String image) throws RexecException {
        return create(new CreateContainerRequest(image));
    }

    /**
     * Start a container.
     */
    public void start(String containerId) throws RexecException {
        client.request("POST", "/api/containers/" + containerId + "/start", null, Void.class);
    }

    /**
     * Stop a container.
     */
    public void stop(String containerId) throws RexecException {
        client.request("POST", "/api/containers/" + containerId + "/stop", null, Void.class);
    }

    /**
     * Delete a container.
     */
    public void delete(String containerId) throws RexecException {
        client.request("DELETE", "/api/containers/" + containerId, null, Void.class);
    }
}
