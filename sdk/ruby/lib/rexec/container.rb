# frozen_string_literal: true

module Rexec
  # Represents a Rexec container/sandbox.
  class Container
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

  # Service for managing containers.
  class ContainerService
    def initialize(client)
      @client = client
    end

    # List all containers.
    #
    # @return [Array<Container>]
    def list
      data = @client.request(:get, "/api/containers")
      # API returns { "containers" => [...], "count" => N, "limit" => M }
      items = data.is_a?(Array) ? data : (data && data["containers"]) || []
      items.map { |c| Container.new(c) }
    end

    # Get a container by ID.
    #
    # @param id [String] Container ID
    # @return [Container]
    def get(id)
      data = @client.request(:get, "/api/containers/#{id}")
      Container.new(data)
    end

    # Create a new container.
    #
    # @param image [String] Docker image to use
    # @param name [String, nil] Optional container name
    # @param environment [Hash] Environment variables
    # @param labels [Hash] Container labels
    # @return [Container]
    #
    # @example
    #   container = client.containers.create(
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
      Container.new(data)
    end

    # Delete a container.
    #
    # @param id [String] Container ID
    def delete(id)
      @client.request(:delete, "/api/containers/#{id}")
      nil
    end

    # Start a container.
    #
    # @param id [String] Container ID
    def start(id)
      @client.request(:post, "/api/containers/#{id}/start")
      nil
    end

    # Stop a container.
    #
    # @param id [String] Container ID
    def stop(id)
      @client.request(:post, "/api/containers/#{id}/stop")
      nil
    end
  end
end
