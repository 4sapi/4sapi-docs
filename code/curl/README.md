# cURL

This example shows how to call 4SAPI with cURL.

## Requirements

- cURL
- Bash, if you want to run `request.sh`

## Configuration

Set your 4SAPI API key:

```bash
export FOUR_S_API_KEY="your-api-key"
```

Optionally set a model name. If omitted, the script uses `gpt-5.6`.

```bash
export FOUR_S_API_MODEL="gpt-5.6"
```

## Run

```bash
./request.sh
```

The script sends a chat completion request to:

```text
https://4sapi.com/v1/chat/completions
```

## PowerShell

On Windows PowerShell, you can run the same request with `curl.exe`:

```powershell
$env:FOUR_S_API_KEY = "your-api-key"
$env:FOUR_S_API_MODEL = "gpt-5.6"

curl.exe --fail-with-body "https://4sapi.com/v1/chat/completions" `
  -H "Authorization: Bearer $env:FOUR_S_API_KEY" `
  -H "Content-Type: application/json" `
  -d "{`"model`":`"$env:FOUR_S_API_MODEL`",`"messages`":[{`"role`":`"user`",`"content`":`"Introduce 4SAPI in one sentence.`"}]}"
```
