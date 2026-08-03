# frozen_string_literal: true

require_relative "rexec/version"
require_relative "rexec/error"
require_relative "rexec/client"
require_relative "rexec/container"
require_relative "rexec/file_service"
require_relative "rexec/terminal"

# Rexec Ruby SDK — official client for AI-native sandboxes.
#
# @example Basic usage
#   client = Rexec::Client.new("https://your-instance.com", "your-token")
#
#   sandbox = client.sandboxes.create(image: "ubuntu")
#   puts "Created: #{sandbox.id}"
#
#   terminal = client.terminal.connect(sandbox.id)
#   terminal.write("echo hello\n")
#   terminal.on_data { |data| puts data }
#
#   client.sandboxes.delete(sandbox.id)
#   # Legacy: client.containers is the same service
#
module Rexec
  class << self
    # Create a new Rexec client.
    #
    # @param base_url [String] Base URL of your Rexec instance
    # @param token [String] API token for authentication
    # @return [Rexec::Client]
    def new(base_url, token, **options)
      Client.new(base_url, token, **options)
    end
  end
end
