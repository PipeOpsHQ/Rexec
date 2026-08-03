package io.pipeops.rexec;

import java.util.HashMap;
import java.util.Map;

/**
 * Request to create a new container.
 * Prefer image aliases such as {@code ubuntu}, {@code debian}, {@code alpine}.
 */
public class CreateContainerRequest {
    private final String image;
    private String name;
    private Map<String, String> environment;
    private Map<String, String> labels;

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

    public String getImage() {
        return image;
    }

    public String getName() {
        return name;
    }
}
