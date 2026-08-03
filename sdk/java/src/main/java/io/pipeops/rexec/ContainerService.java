package io.pipeops.rexec;

/**
 * Deprecated alias for {@link SandboxService}. Prefer sandboxes().
 * @deprecated use {@link SandboxService}
 */
@Deprecated
public class ContainerService extends SandboxService {
    ContainerService(RexecClient client) {
        super(client);
    }
}
