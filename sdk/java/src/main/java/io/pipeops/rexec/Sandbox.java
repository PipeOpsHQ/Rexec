package io.pipeops.rexec;

import com.google.gson.annotations.SerializedName;

import java.util.Collections;
import java.util.Map;

/**
 * Represents a Rexec sandbox.
 */
public class Sandbox {
    private String id;
    private String name;
    private String image;
    private String status;

    @SerializedName("created_at")
    private String createdAt;

    @SerializedName("started_at")
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
        return labels != null ? labels : Collections.emptyMap();
    }

    public Map<String, String> getEnvironment() {
        return environment != null ? environment : Collections.emptyMap();
    }

    public boolean isRunning() {
        return "running".equals(status);
    }

    public boolean isStopped() {
        return "stopped".equals(status);
    }
}
