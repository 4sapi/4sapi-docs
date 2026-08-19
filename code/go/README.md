# Go

This example requires Go 1.20+ and uses only the Go standard library, so no additional dependencies are needed.

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

Run the example:

```bash
go run main.go
```

The program sends a Chat Completions request to `https://4sapi.com/v1/chat/completions` and prints the assistant's response.
