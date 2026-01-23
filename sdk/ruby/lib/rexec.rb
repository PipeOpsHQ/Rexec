# frozen_string_literal: true

require_relative "rexec/version"
require_relative "rexec/error"
require_relative "rexec/client"
require_relative "rexec/container"
require_relative "rexec/file_service"
require_relative "rexec/terminal"

# Rexec Ruby SDK - Official SDK for Rexec Terminal as a Service.
#
# @example Basic usage
#   client = Rexec::Client.new("https://your-instance.com", "your-token")
#   
#   container = client.containers.create(image: "ubuntu:24.04")
#   puts "Created: #{container.id}"
#   
#   terminal = client.terminal.connect(container.id)
#   terminal.write("echo hello\n")
#   terminal.on_data { |data| puts data }
#   
#   client.containers.delete(container.id)
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
