# frozen_string_literal: true

require "uri"

module Rexec
  # Represents file metadata.
  class FileInfo
    attr_reader :name, :path, :size, :mode, :mod_time, :is_dir

    def initialize(data)
      @name = data["name"]
      @path = data["path"]
      @size = data["size"] || 0
      @mode = data["mode"]
      @mod_time = data["mod_time"]
      @is_dir = data["is_dir"] || false
    end

    alias directory? is_dir

    def file?
      !is_dir
    end
  end

  # Service for file operations in containers.
  class FileService
    def initialize(client)
      @client = client
    end

    # List files in a directory.
    #
    # @param container_id [String] Container ID
    # @param path [String] Directory path
    # @return [Array<FileInfo>]
    def list(container_id, path = "/")
      encoded_path = URI.encode_www_form_component(path)
      data = @client.request(:get, "/api/containers/#{container_id}/files/list?path=#{encoded_path}")
      data.map { |f| FileInfo.new(f) }
    end

    # Download a file.
    #
    # @param container_id [String] Container ID
    # @param path [String] File path
    # @return [String] File contents
    def download(container_id, path)
      encoded_path = URI.encode_www_form_component(path)
      @client.request_bytes(:get, "/api/containers/#{container_id}/files?path=#{encoded_path}")
    end

    # Create a directory.
    #
    # @param container_id [String] Container ID
    # @param path [String] Directory path
    def mkdir(container_id, path)
      @client.request(:post, "/api/containers/#{container_id}/files/mkdir", body: { path: path })
      nil
    end

    # Delete a file or directory.
    #
    # @param container_id [String] Container ID
    # @param path [String] Path to delete
    def delete(container_id, path)
      encoded_path = URI.encode_www_form_component(path)
      @client.request(:delete, "/api/containers/#{container_id}/files?path=#{encoded_path}")
      nil
    end
  end
end
