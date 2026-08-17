<div align="center">

# 4SAPI — Unified LLM API Gateway

**One API key, one endpoint, every major LLM provider.**

<br />

[![Star this repo](https://img.shields.io/github/stars/4sapi/4sapi-docs?style=for-the-badge&logo=github&label=%E2%AD%90%20Star%20this%20repo&color=yellow)](https://github.com/4sapi/4sapi-docs/stargazers)

<br />

[![Website](https://img.shields.io/badge/Website-4sapi.com-blue?style=for-the-badge)](https://4sapi.com)
&nbsp;
[![Blog](https://img.shields.io/badge/Blog-4sapi.com%2Fblog-orange?style=for-the-badge)](https://4sapi.com/blog)
&nbsp;
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=for-the-badge)](https://github.com/4sapi/4sapi-docs/pulls)

---

Switching between OpenAI, Claude, Gemini, DeepSeek, and Qwen usually means rewriting your integration for every provider's SDK and auth scheme. 4SAPI removes that work: one OpenAI-compatible endpoint routes to any supported model, so you keep your existing code and just change the model name.

[Quick Start](#quick-start) | [How It Works](#how-it-works) | [Supported Models](#supported-models) | [Docs](#learn-more)

</div>

---

## Quick Start

Point your existing OpenAI client at the 4SAPI endpoint and swap in your API key. That's it.

```python
from openai import OpenAI

client = OpenAI(
    api_key="YOUR_4SAPI_API_KEY",
    base_url="https://4sapi.com/v1"
)

response = client.chat.completions.create(
    model="gpt-5.6",
    messages=[
        {"role": "user", "content": "Hello, 4SAPI!"}
    ]
)

print(response.choices[0].message.content)
```

No new SDK, no new auth flow, no rewritten request bodies.

---

## How It Works

```
Your App (OpenAI SDK)
        |
        v
  4SAPI Gateway  --->  OpenAI
        |         --->  Claude
        |         --->  Gemini
        |         --->  DeepSeek
        +-------->  Qwen
```

Your app talks to a single endpoint. 4SAPI handles routing, provider auth, and format translation behind the scenes.

---

## Why Teams Use 4SAPI

| Problem | 4SAPI's Answer |
|---|---|
| Different SDKs per provider | One OpenAI-compatible interface for all of them |
| Vendor lock-in | Swap models by changing a string, not your codebase |
| Managing multiple API keys | Centralized key and usage management |
| Picking the right model per task | Flexible routing across workloads |
| Scaling AI features across a team | Built for production-scale usage |

---

## Supported Models

- OpenAI (GPT family)
- Anthropic Claude
- Google Gemini
- DeepSeek
- Qwen
- Other popular LLM providers

---

## Learn More

- [Operating Instructions](https://4sapi.apifox.cn/8181987m0)
- [Website](https://4sapi.com)
- [Blog](https://4sapi.com/blog)
- [llms.txt](./llms.txt) — machine-readable index of all guides in this repo, for AI assistants and crawlers

---

## About 4SAPI

4SAPI is a unified LLM API gateway: one OpenAI-compatible endpoint that routes to OpenAI, Anthropic Claude, Google Gemini, DeepSeek, Qwen, and other major providers. This repository, `4sapi-docs`, hosts the team's practical guides — client integration walkthroughs, coding-agent workflows, troubleshooting runbooks, and enterprise governance practices — all built around 4SAPI as the routing layer.

**Contact & Community**
- Website: [4sapi.com](https://4sapi.com)
- Blog: [4sapi.com/blog](https://4sapi.com/blog)
- Issues & questions: [open a GitHub issue](https://github.com/4sapi/4sapi-docs/issues)

---

## Contributing

Found an error in the docs or have an example to add? PRs are welcome — open one and we'll take a look.

---

<div align="center">

---

**Built by 4SAPI Team** | [Website](https://4sapi.com) | [Blog](https://4sapi.com/blog)

<br />

**If this is useful to you:**

[![Star this repo](https://img.shields.io/github/stars/4sapi/4sapi-docs?style=for-the-badge&logo=github&label=%E2%AD%90%20Star%20this%20repo&color=yellow)](https://github.com/4sapi/4sapi-docs/stargazers)

</div>
