using Rexec;

var url = Environment.GetEnvironmentVariable("URL")!;
var token = Environment.GetEnvironmentVariable("TOKEN")!;
using var client = new RexecClient(url, token);

Console.WriteLine("[cs] list...");
var before = await client.Containers.ListAsync();
Console.WriteLine($"[cs] list count {before.Count}");

Console.WriteLine("[cs] create...");
var c = await client.Containers.CreateAsync(new CreateContainerRequest("ubuntu") { Name = $"cs-e2e-{DateTimeOffset.UtcNow.ToUnixTimeSeconds()}" });
Console.WriteLine($"[cs] created {c?.Id} {c?.Status} {c?.Image}");

var after = await client.Containers.ListAsync();
Console.WriteLine($"[cs] list count {after.Count}");

var got = await client.Containers.GetAsync(c!.Id);
Console.WriteLine($"[cs] get {got?.Id} {got?.Status}");

await client.Containers.DeleteAsync(c.Id);
Console.WriteLine("[cs] OK");
