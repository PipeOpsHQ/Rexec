package io.pipeops.rexec;

import java.util.HashMap;
import java.util.Map;

/**
 * Represents a Rexec container/sandbox.
 */
public class Container {
    private String id;
    private String name;
    private String image;
    private String status;
    private String createdAt;
    private String startedAt;
    private Map<String, String> labels;
    private Map<String, String> environment;

    public String getId() {
        return id;
    }

    public String getName() {
        return name;
    }

    public String getImage() {
        return image;
    }

    public String getStatus() {
        return status;
    }

    public String getCreatedAt() {
        return createdAt;
    }

    public String getStartedAt() {
        return startedAt;
    }

    public Map<String, String> getLabels() {
        return labels != null ? labels : new HashMap<>();
    }

    public Map<String, String> getEnvironment() {
        return environment != null ? environment : new HashMap<>();
    }

    public boolean isRunning() {
        return "running".equals(status);
    }

    public boolean isStopped() {
        return "stopped".equals(status);
    }
}

/**
 * Request to create a new container.
 */
class CreateContainerRequest {
    private final String image;
    private String name;
    private Map<String, String> environment;
    private Map<String, String> labels;

    /**
     * Create a request with the specified image.
     */
    public CreateContainerRequest(String image) {
        this.image = image;
    }

    public CreateContainerRequest setName(String name) {
        this.name = name;
        return this;
    }

    public CreateContainerRequest setEnvironment(Map<String, String> environment) {
        this.environment = environment;
        return this;
    }

    public CreateContainerRequest setLabels(Map<String, String> labels) {
        this.labels = labels;
        return this;
    }

    public CreateContainerRequest addEnv(String key, String value) {
        if (this.environment == null) {
            this.environment = new HashMap<>();
        }
        this.environment.put(key, value);
        return this;
    }

    public CreateContainerRequest addLabel(String key, String value) {
        if (this.labels == null) {
            this.labels = new HashMap<>();
        }
        this.labels.put(key, value);
        return this;
    }
}
