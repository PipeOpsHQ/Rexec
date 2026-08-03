package io.pipeops.rexec;

/**
 * Deprecated alias for {@link CreateSandboxRequest}.
 * @deprecated use {@link CreateSandboxRequest}
 */
@Deprecated
public class CreateContainerRequest extends CreateSandboxRequest {
    public CreateContainerRequest(String image) {
        super(image);
    }
}
