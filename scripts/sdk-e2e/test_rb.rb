require "json"
$LOAD_PATH.unshift File.expand_path("../../sdk/ruby/lib", __dir__)
require "rexec"

env = File.read(File.expand_path(".env", __dir__)).lines.map { |l| l.strip.split("=", 2) }.to_h
client = Rexec::Client.new(env["URL"], env["TOKEN"])
puts "[rb] list..."
before = client.containers.list
puts "[rb] list count #{before.length}"
puts "[rb] create..."
c = client.containers.create(image: "ubuntu", name: "rb-e2e-#{Time.now.to_i}")
puts "[rb] created #{c.id} #{c.status} #{c.image}"
after = client.containers.list
puts "[rb] list count #{after.length}"
got = client.containers.get(c.id)
puts "[rb] get #{got.id} #{got.status}"
client.containers.delete(c.id)
puts "[rb] OK"
