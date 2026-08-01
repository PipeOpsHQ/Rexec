# frozen_string_literal: true

require "json"

module Rexec
  # WebSocket terminal connection.
  class Terminal
    attr_reader :closed

    def initialize(ws)
      @ws = ws
      @closed = false
      @data_handlers = []
      @close_handlers = []
      @error_handlers = []

      setup_handlers
    end

    # Send data to the terminal.
    #
    # @param data [String] Data to send
    def write(data)
      raise TerminalClosedError if @closed

      @ws.send(data)
    end

    # Resize the terminal.
    #
    # @param cols [Integer] Number of columns
    # @param rows [Integer] Number of rows
    def resize(cols, rows)
      raise TerminalClosedError if @closed

      msg = { type: "resize", cols: cols, rows: rows }.to_json
      @ws.send(msg)
    end

    # Register a handler for incoming data.
    #
    # @yield [data] Block called with each chunk of data
    def on_data(&block)
      @data_handlers << block
    end

    # Register a handler for connection close.
    #
    # @yield Block called when connection closes
    def on_close(&block)
      @close_handlers << block
    end

    # Register a handler for errors.
    #
    # @yield [error] Block called with error
    def on_error(&block)
      @error_handlers << block
    end

    # Close the terminal connection.
    def close
      return if @closed

      @closed = true
      @ws.close
    end

    alias closed? closed

    private

    def setup_handlers
      @ws.on :message do |msg|
        @data_handlers.each { |h| h.call(msg.data) }
      end

      @ws.on :close do
        @closed = true
        @close_handlers.each(&:call)
      end

      @ws.on :error do |e|
        @error_handlers.each { |h| h.call(e) }
      end
    end
  end

  # Service for terminal WebSocket connections.
  class TerminalService
    def initialize(client)
      @client = client
    end

    # Connect to a container's terminal.
    #
    # @param container_id [String] Container ID
    # @param cols [Integer] Terminal width (default: 80)
    # @param rows [Integer] Terminal height (default: 24)
    # @return [Terminal]
    #
    # @example
    #   terminal = client.terminal.connect(container.id)
    #   terminal.on_data { |data| print data }
    #   terminal.write("ls -la\n")
    def connect(container_id, cols: 80, rows: 24)
      begin
        require "websocket-client-simple"
      rescue LoadError
        raise LoadError, "websocket-client-simple is required for terminal connections (gem install websocket-client-simple)"
      end
      ws_url = @client.ws_url("/ws/terminal/#{container_id}")

      ws = WebSocket::Client::Simple.connect(ws_url, headers: {
        "Authorization" => "Bearer #{@client.token}"
      })

      terminal = Terminal.new(ws)

      # Wait for connection
      sleep 0.1 until ws.open?

      # Set initial size
      terminal.resize(cols, rows)

      terminal
    end
  end
end
