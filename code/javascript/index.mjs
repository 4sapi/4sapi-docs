import OpenAI from "openai";

// Read credentials from environment variables.
const client = new OpenAI({
  apiKey: process.env.FOUR_S_API_KEY,
  baseURL: "https://4sapi.com/v1",
});

// Call the 4SAPI chat completions endpoint through the OpenAI SDK.
const response = await client.chat.completions.create({
  model: process.env.FOUR_S_API_MODEL || "gpt-5.6",
  messages: [{ role: "user", content: "Introduce 4SAPI in one sentence." }],
});

console.log(response.choices[0].message.content);
