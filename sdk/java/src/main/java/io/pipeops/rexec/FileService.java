package io.pipeops.rexec;

import java.util.List;

/**
 * Service for file operations within containers.
 */
public class FileService {
    private final RexecClient client;

    FileService(RexecClient client) {
        this.client = client;
    }

    /**
     * List files in a directory.
     *
     * @param containerId Container ID
     * @param path        Directory path
     */
    public List<FileInfo> list(String containerId, String path) throws RexecException {
        String encodedPath = java.net.URLEncoder.encode(path, java.nio.charset.StandardCharsets.UTF_8);
        FileListResponse response = client.request(
                "GET",
                "/api/containers/" + containerId + "/files/list?path=" + encodedPath,
                null,
                FileListResponse.class
        );
        if (response == null || response.files == null) {
            return List.of();
        }
        return response.files;
    }

    /**
     * Read a file's contents.
     *
     * @param containerId Container ID
     * @param path        File path
     */
    public byte[] read(String containerId, String path) throws RexecException {
        String encodedPath = java.net.URLEncoder.encode(path, java.nio.charset.StandardCharsets.UTF_8);
        return client.requestBytes("GET", "/api/containers/" + containerId + "/files?path=" + encodedPath);
    }

    /**
     * Read a file as string.
     *
     * @param containerId Container ID
     * @param path        File path
     */
    public String readString(String containerId, String path) throws RexecException {
        return new String(read(containerId, path), java.nio.charset.StandardCharsets.UTF_8);
    }

    /**
     * Write content to a file.
     *
     * @param containerId Container ID
     * @param path        File path
     * @param content     Content to write
     */
    public void write(String containerId, String path, byte[] content) throws RexecException {
        WriteFileRequest request = new WriteFileRequest(path, java.util.Base64.getEncoder().encodeToString(content));
        client.request("POST", "/api/containers/" + containerId + "/files", request, Void.class);
    }

    /**
     * Write string content to a file.
     *
     * @param containerId Container ID
     * @param path        File path
     * @param content     Content to write
     */
    public void write(String containerId, String path, String content) throws RexecException {
        write(containerId, path, content.getBytes(java.nio.charset.StandardCharsets.UTF_8));
    }

    /**
     * Delete a file.
     *
     * @param containerId Container ID
     * @param path        File path
     */
    public void delete(String containerId, String path) throws RexecException {
        String encodedPath = java.net.URLEncoder.encode(path, java.nio.charset.StandardCharsets.UTF_8);
        client.request("DELETE", "/api/containers/" + containerId + "/files?path=" + encodedPath, null, Void.class);
    }

    private static class FileListResponse {
        List<FileInfo> files;
    }
}

class WriteFileRequest {
    private final String path;
    private final String content;

    WriteFileRequest(String path, String content) {
        this.path = path;
        this.content = content;
    }
}
