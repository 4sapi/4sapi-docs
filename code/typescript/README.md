# TypeScript

This example shows how to call 4SAPI from TypeScript with the OpenAI SDK.

## Requirements

- Node.js 18+
- `tsx`
- OpenAI SDK

```bash
npm install openai
npm install -D tsx typescript
```

## Configuration

Set your 4SAPI API key:

```bash
export FOUR_S_API_KEY="your-api-key"
```

PowerShell:

```powershell
$env:FOUR_S_API_KEY = "your-api-key"
```

Optionally set a model name. If omitted, the example uses `gpt-5.6`.

```bash
export FOUR_S_API_MODEL="gpt-5.6"
```

PowerShell:

```powershell
$env:FOUR_S_API_MODEL = "gpt-5.6"
```

## Run

```bash
npx tsx index.ts
```

The script sends a chat completion request to:

```text
https://4sapi.com/v1/chat/completions
```
