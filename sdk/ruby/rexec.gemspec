# frozen_string_literal: true

Gem::Specification.new do |spec|
  spec.name = "rexec"
  spec.version = "1.0.0"
  spec.authors = ["PipeOpsHQ"]
  spec.email = ["support@pipeops.io"]

  spec.summary = "Official Ruby SDK for Rexec - Terminal as a Service"
  spec.description = "Ruby SDK for interacting with Rexec sandboxed environments. Create containers, execute commands, manage files, and connect to terminals."
  spec.homepage = "https://github.com/PipeOpsHQ/rexec"
  spec.license = "MIT"
  spec.required_ruby_version = ">= 3.0.0"

  spec.metadata["homepage_uri"] = spec.homepage
  spec.metadata["source_code_uri"] = "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/ruby"
  spec.metadata["changelog_uri"] = "https://github.com/PipeOpsHQ/rexec/blob/main/CHANGELOG.md"

  spec.files = Dir["lib/**/*", "README.md", "LICENSE"]
  spec.require_paths = ["lib"]

  spec.add_dependency "faraday", "~> 2.0"
  spec.add_dependency "faraday-multipart", "~> 1.0"
  spec.add_dependency "websocket-client-simple", "~> 0.8"

  spec.add_development_dependency "rspec", "~> 3.0"
  spec.add_development_dependency "rubocop", "~> 1.0"
  spec.add_development_dependency "yard", "~> 0.9"
end
