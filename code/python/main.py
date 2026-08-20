import os
from openai import OpenAI


# Read your 4SAPI credentials from environment variables.
api_key = os.environ["FOUR_S_API_KEY"]
model = os.getenv("FOUR_S_API_MODEL", "gpt-5.6")

# The OpenAI SDK can call 4SAPI by pointing base_url to the 4SAPI endpoint.
client = OpenAI(api_key=api_key, base_url="https://4sapi.com/v1")

# Send a standard chat completion request.
response = client.chat.completions.create(
    model=model,
    messages=[{"role": "user", "content": "Introduce 4SAPI in one sentence."}],
)

print(response.choices[0].message.content)
