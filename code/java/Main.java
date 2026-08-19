import java.net.URI;
import java.net.http.*;

public class Main {
  public static void main(String[] args) throws Exception {
    // Read the API key from the environment instead of storing it in source code.
    String key = System.getenv("FOUR_S_API_KEY");
    if (key == null || key.isBlank()) {
      throw new IllegalStateException("FOUR_S_API_KEY is required");
    }

    // Use a configurable model name, with a default for quick testing.
    String model = System.getenv().getOrDefault("FOUR_S_API_MODEL", "gpt-5.6");

    // 4SAPI accepts the OpenAI-compatible Chat Completions request format.
    String json = "{\"model\":\"" + model + "\",\"messages\":[{\"role\":\"user\",\"content\":\"用一句话介绍 4SAPI。\"}]}";

    // Send the request to the 4SAPI OpenAI-compatible endpoint.
    HttpRequest request = HttpRequest.newBuilder(
            URI.create("https://4sapi.com/v1/chat/completions"))
        .header("Authorization", "Bearer " + key)
        .header("Content-Type", "application/json")
        .POST(HttpRequest.BodyPublishers.ofString(json))
        .build();
    HttpResponse<String> response = HttpClient.newHttpClient().send(
        request,
        HttpResponse.BodyHandlers.ofString());

    // Surface non-success responses with the body to simplify troubleshooting.
    if (response.statusCode() >= 300) {
      throw new RuntimeException(response.statusCode() + ": " + response.body());
    }

    // Print the raw response so the example needs no JSON parsing dependency.
    System.out.println(response.body());
  }
}
