# Java

This example uses the built-in `java.net.http.HttpClient`, so it requires JDK 11+ and has no third-party dependencies.

Set your 4SAPI API key before running the example. You can optionally set `FOUR_S_API_MODEL`; it defaults to `gpt-5.6`.

```bash
export FOUR_S_API_KEY="your-api-key"
export FOUR_S_API_MODEL="gpt-5.6"
```

On Windows PowerShell:

```powershell
$env:FOUR_S_API_KEY = "your-api-key"
$env:FOUR_S_API_MODEL = "gpt-5.6"
```

Compile and run:

```bash
javac Main.java && java Main
```

The example prints the raw JSON response. This makes it easy to verify your API key, model name, and network connection without adding a JSON library.
