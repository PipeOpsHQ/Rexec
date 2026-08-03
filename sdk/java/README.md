# Rexec Java SDK

Official Java client for the [Rexec](https://rexec.sh) API (sandboxes, files, terminals).

**Coordinates:** `io.pipeops:rexec:1.0.1` · also usable from Kotlin

## Requirements

- Java 17+
- Maven or Gradle

## Install

### Maven

```xml
<dependency>
    <groupId>io.pipeops</groupId>
    <artifactId>rexec</artifactId>
    <version>1.0.1</version>
</dependency>
```

### Gradle

```groovy
implementation 'io.pipeops:rexec:1.0.1'
```

From source (monorepo):

```bash
cd sdk/java && mvn install -DskipTests
```

## Quick start

```java
import io.pipeops.rexec.*;

public class Example {
    public static void main(String[] args) throws RexecException {
        RexecClient client = new RexecClient(
            System.getenv("REXEC_URL"),
            System.getenv("REXEC_TOKEN")
        );

        System.out.println("count " + client.containers().list().size());

        // Prefer image aliases: ubuntu, debian, alpine (not ubuntu on hosted Rexec)
        Container c = client.containers().create(
            new CreateContainerRequest("ubuntu").setName("java-demo")
        );
        System.out.println(c.getId() + " " + c.getStatus());

        client.containers().get(c.getId());
        client.containers().delete(c.getId());
    }
}
```

## Kotlin

```kotlin
val client = RexecClient(System.getenv("REXEC_URL"), System.getenv("REXEC_TOKEN"))
val list = client.containers().list()
val c = client.containers().create(CreateContainerRequest("ubuntu").setName("kt-demo"))
client.containers().delete(c.id)
```

## Containers

```java
List<Container> containers = client.containers().list();
Container c = client.containers().create(new CreateContainerRequest("ubuntu").setName("demo"));
client.containers().start(c.getId());
client.containers().stop(c.getId());
client.containers().delete(c.getId());
```

> Rexec has no HTTP `exec` API. Use the terminal WebSocket for interactive commands.

## Files

```java
List<FileInfo> files = client.files().list(containerId, "/home");
byte[] content = client.files().read(containerId, "/etc/hostname");
client.files().write(containerId, "/tmp/hello.txt", "hi\n".getBytes());
client.files().delete(containerId, "/tmp/hello.txt");
```

## Auth

Create an API token in the Rexec UI, or use a guest JWT for short tests (see monorepo `docs/SDK.md`).

## Publishing to Maven Central

See monorepo [docs/SDK_PUBLISHING.md](../../docs/SDK_PUBLISHING.md) (Sonatype Central Portal + GPG secrets).  
CI: **Publish SDKs** workflow with `sdks=java`.

## License

MIT — see [LICENSE](LICENSE).
