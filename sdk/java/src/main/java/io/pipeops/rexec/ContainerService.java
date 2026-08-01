package io.pipeops.rexec;

import java.util.List;

/**
 * Service for managing containers.
 */
public class ContainerService {
    private final RexecClient client;

    ContainerService(RexecClient client) {
        this.client = client;
    }

    /**
     * List all containers.
     */
    public List<Container> list() throws RexecException {
        ContainerListResponse response = client.request("GET", "/api/containers", null, ContainerListResponse.class);
        return response != null ? response.containers : List.of();
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
     * Create a new container with just an image name.
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

    /**
     * Execute a command in a container (non-interactive).
     */
    public ExecResult exec(String containerId, String[] command) throws RexecException {
        ExecRequest request = new ExecRequest(command);
        return client.request("POST", "/api/containers/" + containerId + "/exec", request, ExecResult.class);
    }

    /**
     * Execute a shell command in a container.
     */
    public ExecResult exec(String containerId, String command) throws RexecException {
        return exec(containerId, new String[]{"/bin/sh", "-c", command});
    }

    private static class ContainerListResponse {
        List<Container> containers;
    }
}

class ExecRequest {
    private final String[] command;

    ExecRequest(String[] command) {
        this.command = command;
    }
}

/**
 * Result of command execution.
 */
class ExecResult {
    private int exitCode;
    private String stdout;
    private String stderr;

    public int getExitCode() {
        return exitCode;
    }

    public String getStdout() {
        return stdout;
    }

    public String getStderr() {
        return stderr;
    }

    public boolean isSuccess() {
        return exitCode == 0;
    }
}
