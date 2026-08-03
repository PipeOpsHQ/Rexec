import io.pipeops.rexec.Container;
import io.pipeops.rexec.CreateContainerRequest;
import io.pipeops.rexec.RexecClient;
import io.pipeops.rexec.RexecException;

import java.util.List;

/**
 * E2E smoke: list → create → get → delete for the Java SDK.
 *
 * Requires the SDK installed in the local Maven repo:
 *   (cd ../../../sdk/java && mvn install -DskipTests -q)
 *   URL=... TOKEN=... mvn -q exec:java
 */
public class E2E {
    public static void main(String[] args) throws RexecException {
        String url = env("URL", env("REXEC_URL", "https://rexec.sh"));
        String token = env("TOKEN", env("REXEC_TOKEN", ""));
        if (token.isEmpty()) {
            System.err.println("TOKEN or REXEC_TOKEN required");
            System.exit(1);
        }

        RexecClient client = new RexecClient(url, token);

        System.out.println("[java] list...");
        List<Container> before = client.containers().list();
        System.out.println("[java] list count " + before.size());

        System.out.println("[java] create...");
        Container c = client.containers().create(
                new CreateContainerRequest("ubuntu").setName("java-e2e-" + System.currentTimeMillis())
        );
        System.out.println("[java] created " + c.getId() + " " + c.getStatus() + " " + c.getImage());
        if (c.getId() == null || c.getId().isEmpty()) {
            throw new IllegalStateException("create returned empty id");
        }

        List<Container> after = client.containers().list();
        System.out.println("[java] list count " + after.size());

        Container got = client.containers().get(c.getId());
        System.out.println("[java] get " + got.getId() + " " + got.getStatus());

        client.containers().delete(c.getId());
        System.out.println("[java] OK");
    }

    private static String env(String key, String fallback) {
        String v = System.getenv(key);
        return v == null || v.isEmpty() ? fallback : v;
    }
}
