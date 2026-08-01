# frozen_string_literal: true

require "faraday"
require "json"

module Rexec
  # Main client for interacting with Rexec API.
  #
  # @example
  #   client = Rexec::Client.new("https://your-instance.com", "your-token")
  #   containers = client.containers.list
  #
  class Client
    attr_reader :base_url, :containers, :files, :terminal

    # Initialize a new Rexec client.
    #
    # @param base_url [String] Base URL of your Rexec instance
    # @param token [String] API token for authentication
    # @param timeout [Integer] Request timeout in seconds (default: 30)
    def initialize(base_url, token, timeout: 30)
      @base_url = base_url.chomp("/")
      @token = token
      @timeout = timeout

      @http = Faraday.new(url: @base_url) do |f|
        f.request :json
        f.response :json, content_type: /\bjson$/
        f.adapter Faraday.default_adapter
        f.options.timeout = timeout
        f.headers["Authorization"] = "Bearer #{token}"
        f.headers["Accept"] = "application/json"
      end

      @containers = ContainerService.new(self)
      @files = FileService.new(self)
      @terminal = TerminalService.new(self)
    end

    # Make an API request.
    # @api private
    def request(method, path, body: nil, params: nil)
      response = @http.run_request(method, path, body, nil) do |req|
        req.params = params if params
      end

      handle_response(response)
    end

    # Make a raw request and return bytes.
    # @api private
    def request_bytes(method, path)
      raw_http = Faraday.new(url: @base_url) do |f|
        f.adapter Faraday.default_adapter
        f.options.timeout = @timeout
        f.headers["Authorization"] = "Bearer #{@token}"
      end

      response = raw_http.run_request(method, path, nil, nil)
      
      if response.status >= 400
        raise APIError.new(response.status, "Request failed")
      end

      response.body
    end

    # Get WebSocket URL.
    # @api private
    def ws_url(path)
      uri = URI.parse(@base_url)
      ws_scheme = uri.scheme == "https" ? "wss" : "ws"
      "#{ws_scheme}://#{uri.host}:#{uri.port || (uri.scheme == 'https' ? 443 : 80)}#{path}"
    end

    # Get the API token.
    # @api private
    attr_reader :token

    private

    def handle_response(response)
      case response.status
      when 200..299
        response.body
      when 401, 403
        raise AuthError.new(extract_error_message(response))
      when 404
        raise APIError.new(404, extract_error_message(response))
      else
        raise APIError.new(response.status, extract_error_message(response), response.body)
      end
    end

    def extract_error_message(response)
      if response.body.is_a?(Hash)
        response.body["error"] || response.body["message"] || "Unknown error"
      else
        "Unknown error"
      end
    end
  end
end
