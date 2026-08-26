using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

// Read credentials from environment variables.
var key = Environment.GetEnvironmentVariable("FOUR_S_API_KEY")
    ?? throw new InvalidOperationException("FOUR_S_API_KEY is required");

// Use a default model unless FOUR_S_API_MODEL is provided.
var model = Environment.GetEnvironmentVariable("FOUR_S_API_MODEL") ?? "gpt-5.6";

using var http = new HttpClient();
http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", key);

// Build the JSON payload for the chat completion request.
var payload = new
{
    model,
    messages = new[]
    {
        new
        {
            role = "user",
            content = "Introduce 4SAPI in one sentence.",
        },
    },
};

var json = JsonSerializer.Serialize(payload);

// Send the request to the 4SAPI chat completions endpoint.
using var response = await http.PostAsync(
    "https://4sapi.com/v1/chat/completions",
    new StringContent(json, Encoding.UTF8, "application/json"));

var body = await response.Content.ReadAsStringAsync();
response.EnsureSuccessStatusCode();

Console.WriteLine(body);
