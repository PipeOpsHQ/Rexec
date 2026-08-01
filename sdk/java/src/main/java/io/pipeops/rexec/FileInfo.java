package io.pipeops.rexec;

/**
 * Information about a file in a container.
 */
public class FileInfo {
    private String name;
    private String path;
    private long size;
    private String mode;
    private String modTime;
    private boolean isDir;

    public String getName() {
        return name;
    }

    public String getPath() {
        return path;
    }

    public long getSize() {
        return size;
    }

    public String getMode() {
        return mode;
    }

    public String getModTime() {
        return modTime;
    }

    public boolean isDirectory() {
        return isDir;
    }

    public boolean isFile() {
        return !isDir;
    }
}
