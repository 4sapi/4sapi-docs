#!/usr/bin/env bash
set -euo pipefail

# Require a 4SAPI API key before sending the request.
: "${FOUR_S_API_KEY:?Please set FOUR_S_API_KEY first}"

# Use a default model unless FOUR_S_API_MODEL is provided.
MODEL="${FOUR_S_API_MODEL:-gpt-5.6}"

# Send a chat completion request to the 4SAPI endpoint.
curl --fail-with-body "https://4sapi.com/v1/chat/completions" \
  -H "Authorization: Bearer ${FOUR_S_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"${MODEL}\",\"messages\":[{\"role\":\"user\",\"content\":\"Introduce 4SAPI in one sentence.\"}]}"
