# frozen_string_literal: true

module Rexec
  # Represents a Rexec sandbox (isolated Linux environment).
  class Sandbox
    attr_reader :id, :name, :image, :status, :created_at, :started_at, :labels, :environment

    def initialize(data)
      @id = data["id"]
      @name = data["name"]
      @image = data["image"]
      @status = data["status"]
      @created_at = data["created_at"]
      @started_at = data["started_at"]
      @labels = data["labels"] || {}
      @environment = data["environment"] || {}
    end

    def running?
      status == "running"
    end

    def stopped?
      status == "stopped"
    end
  end

  # @deprecated Use {Sandbox}
  Container = Sandbox

  # Service for managing sandboxes. HTTP paths remain /api/containers.
  class SandboxService
    def initialize(client)
      @client = client
    end

    # List all sandboxes.
    #
    # @return [Array<Sandbox>]
    def list
      data = @client.request(:get, "/api/containers")
      # API returns { "containers" => [...], "count" => N, "limit" => M }
      items = data.is_a?(Array) ? data : (data && data["containers"]) || []
      items.map { |c| Sandbox.new(c) }
    end

    # Get a sandbox by ID.
    #
    # @param id [String] Sandbox ID
    # @return [Sandbox]
    def get(id)
      data = @client.request(:get, "/api/containers/#{id}")
      Sandbox.new(data)
    end

    # Create a new sandbox.
    #
    # @param image [String] Image alias (e.g. "ubuntu")
    # @param name [String, nil] Optional sandbox name
    # @param environment [Hash] Environment variables
    # @param labels [Hash] Labels
    # @return [Sandbox]
    #
    # @example
    #   sandbox = client.sandboxes.create(
    #     image: "ubuntu",
    #     name: "my-sandbox",
    #     environment: { "MY_VAR" => "value" }
    #   )
    def create(image:, name: nil, environment: {}, labels: {})
      body = { image: image }
      body[:name] = name if name
      body[:environment] = environment unless environment.empty?
      body[:labels] = labels unless labels.empty?

      data = @client.request(:post, "/api/containers", body: body)
      Sandbox.new(data)
    end

    # Delete a sandbox.
    #
    # @param id [String] Sandbox ID
    def delete(id)
      @client.request(:delete, "/api/containers/#{id}")
      nil
    end

    # Start a sandbox.
    #
    # @param id [String] Sandbox ID
    def start(id)
      @client.request(:post, "/api/containers/#{id}/start")
      nil
    end

    # Stop a sandbox.
    #
    # @param id [String] Sandbox ID
    def stop(id)
      @client.request(:post, "/api/containers/#{id}/stop")
      nil
    end
  end

  # @deprecated Use {SandboxService}
  ContainerService = SandboxService
end
