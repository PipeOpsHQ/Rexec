# frozen_string_literal: true

module Rexec
  # Base error class for all Rexec errors.
  class Error < StandardError; end

  # API error with status code.
  class APIError < Error
    attr_reader :status_code, :response_body

    def initialize(status_code, message, response_body = nil)
      @status_code = status_code
      @response_body = response_body
      super("API error #{status_code}: #{message}")
    end
  end

  # Authentication error (401/403).
  class AuthError < APIError
    def initialize(message = "Authentication failed")
      super(401, message)
    end
  end

  # Resource not found (404).
  class NotFoundError < APIError
    attr_reader :resource, :resource_id

    def initialize(resource, resource_id)
      @resource = resource
      @resource_id = resource_id
      super(404, "#{resource} '#{resource_id}' not found")
    end
  end

  # Connection error.
  class ConnectionError < Error; end

  # Terminal connection closed.
  class TerminalClosedError < Error
    def initialize
      super("Terminal connection closed")
    end
  end
end
