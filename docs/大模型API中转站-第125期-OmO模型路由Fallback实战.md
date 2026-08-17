---
title: "【大模型API中转站】第125期 OmO模型路由 | fallback省钱"
category: 人工智能
tags:
  - 大模型API中转站
  - oh-my-openagent
  - 模型路由
  - fallback
  - runtime_fallback
  - OpenCode
  - 成本治理
  - 4SAPI
description: "用 OpenCode + oh-my-openagent + 4SAPI 讲清企业级模型路由：Agent 用强模型，Category 用任务分层，fallback_models 处理模型不可用，runtime_fallback 处理错误码，background_task 控制并发，再用 4SAPI 日志反推哪些任务该降级、哪些任务必须保留强模型。"
---

# 【大模型API中转站】第125期 OmO模型路由 | fallback省钱

本文是【大模型API中转站】系列的第125篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇讲了 OmO Team Mode 怎么防止成本失控。

这一篇讲更底层的能力：

```text
模型路由和 fallback。
```

企业用 Agent，最容易犯两个极端错误。

第一个极端：

```text
所有任务都用最强模型。
```

结果当然稳一点。

但成本很快上去。

第二个极端：

```text
所有任务都用便宜模型。
```

看起来省钱。

但复杂任务会反复失败，最后更贵。

真正靠谱的做法是：

```text
按任务分层选模型。
失败时自动 fallback。
用日志反推路由策略。
```

这就是 OmO + 4SAPI 组合很适合企业级大模型接入的地方。

OmO 负责 Agent 和 Category 分工。

4SAPI 负责统一 Key、模型、日志、成本和错误排查。

## 1. 先理解三层路由

OmO 里的模型路由，可以分三层。

第一层，Agent 路由。

```text
Sisyphus 用什么模型？
Hephaestus 用什么模型？
Oracle 用什么模型？
Librarian 用什么模型？
Explore 用什么模型？
```

第二层，Category 路由。

```text
quick 用什么模型？
deep 用什么模型？
ultrabrain 用什么模型？
writing 用什么模型？
visual-engineering 用什么模型？
```

第三层，fallback 路由。

```text
主模型失败后，按什么顺序切备用模型？
哪些错误触发 fallback？
最多尝试几次？
多久冷却？
```

这三层叠在一起，才是企业模型治理。

不要只改一个 `model` 字段。

## 2. Agent 路由：按角色选模型

Agent 是角色。

不同角色需要的模型能力不同。

建议这样理解：

```text
Sisyphus：主调度，适合强推理。
Hephaestus：深度执行，适合代码模型。
Oracle：架构和 Debug 咨询，适合强推理或不同厂商模型。
Librarian：文档和外部搜索，适合快模型。
Explore：代码库搜索，适合快模型。
Prometheus：规划，适合强推理。
Momus / Metis：计划评审，适合强推理。
```

所以一个企业初始配置可以是：

```jsonc
{
  "agents": {
    "sisyphus": {
      "model": "4sapi/reasoning-model"
    },
    "hephaestus": {
      "model": "4sapi/coding-model"
    },
    "oracle": {
      "model": "4sapi/reasoning-model"
    },
    "librarian": {
      "model": "4sapi/fast-model"
    },
    "explore": {
      "model": "4sapi/fast-model"
    }
  }
}
```

这里的模型名只是占位。

实际要从 4SAPI 模型广场复制。

关键思路是：

```text
搜索和整理不要占用最强模型。
执行和推理才用更强模型。
```

## 3. Category 路由：按任务选模型

Category 比 Agent 更适合做成本治理。

因为很多时候，主 Agent 会把任务委托给某个类别。

比如：

```text
quick：小修小改。
deep：复杂执行。
ultrabrain：硬逻辑和架构。
writing：文档和文章。
visual-engineering：前端和 UI。
```

可以这样配：

```jsonc
{
  "categories": {
    "quick": {
      "model": "4sapi/fast-model",
      "description": "小修小改、拼写、单文件任务"
    },
    "deep": {
      "model": "4sapi/coding-model",
      "description": "跨文件实现、复杂执行"
    },
    "ultrabrain": {
      "model": "4sapi/reasoning-model",
      "description": "架构判断、复杂逻辑、疑难问题"
    },
    "writing": {
      "model": "4sapi/writing-model",
      "description": "中文文档、博客、报告"
    }
  }
}
```

这比只在 Agent 层配模型更灵活。

因为同一个 Sisyphus 可以根据任务类型，把工作分给不同 Category。

## 4. fallback_models：主模型失败后的顺序

OmO 支持给 Agent 或 Category 配：

```text
fallback_models
```

最简单写法：

```jsonc
{
  "agents": {
    "sisyphus": {
      "model": "4sapi/reasoning-model",
      "fallback_models": [
        "4sapi/coding-model",
        "4sapi/fast-model"
      ]
    }
  }
}
```

这表示：

```text
先用 reasoning-model。
失败后切 coding-model。
再失败切 fast-model。
```

但注意：

```text
fallback 不是降智链。
```

有些任务不能从强模型降到弱模型。

比如架构评审失败，不一定应该 fallback 到 fast 模型。

更合理的是：

```text
同能力不同供应商。
同档位备用模型。
再考虑降级输出部分结果。
```

比如：

```jsonc
{
  "categories": {
    "ultrabrain": {
      "model": "4sapi/reasoning-model-a",
      "fallback_models": [
        "4sapi/reasoning-model-b",
        "4sapi/coding-model"
      ]
    }
  }
}
```

这样比直接掉到 fast 模型更靠谱。

## 5. fallback_models 的对象写法

OmO 文档里提到，`fallback_models` 可以混合字符串和对象。

对象可以带单独设置。

例如：

```jsonc
{
  "agents": {
    "oracle": {
      "model": "4sapi/reasoning-model-a",
      "fallback_models": [
        {
          "model": "4sapi/reasoning-model-b",
          "temperature": 0.2,
          "maxTokens": 8192
        },
        {
          "model": "4sapi/coding-model",
          "temperature": 0.1
        }
      ]
    }
  }
}
```

对象写法适合：

```text
给备用模型设置更低 temperature。
给备用模型限制输出长度。
给不同模型设置不同推理档位。
```

但不要乱写不支持的参数。

OmO 会做能力兼容归一化。

但企业里最好还是按模型实际支持来配。

## 6. runtime_fallback：哪些错误触发切换

`fallback_models` 是备用模型列表。

`runtime_fallback` 是触发机制。

基础写法：

```jsonc
{
  "runtime_fallback": {
    "enabled": true,
    "retry_on_errors": [429, 500, 502, 503, 504],
    "max_fallback_attempts": 3,
    "cooldown_seconds": 60,
    "timeout_seconds": 30,
    "notify_on_fallback": true
  }
}
```

这适合处理：

```text
429 限流。
500 / 502 / 503 / 504 上游错误。
请求超时。
模型短时不可用。
```

不要一上来把所有 4xx 都加进去。

因为：

```text
401 可能是 Key 错。
403 可能是权限不足。
404 可能是模型名错。
400 可能是请求格式错。
```

这些不一定应该 fallback。

否则你会用备用模型掩盖配置错误。

企业建议先在 4SAPI 日志里看真实错误码。

再决定：

```text
哪些错误应该自动切。
哪些错误应该立即暴露。
```

## 7. Proxy API 的错误码策略

如果你使用的是中转或网关，错误码可能和官方 API 不完全一样。

有些网关会把：

```text
模型不存在
无权限
上游不可用
额度不足
```

都映射成不同的 4xx。

所以可以准备一个“观察期配置”：

```jsonc
{
  "runtime_fallback": {
    "enabled": true,
    "retry_on_errors": [429, 500, 502, 503, 504],
    "max_fallback_attempts": 2,
    "cooldown_seconds": 30,
    "timeout_seconds": 20,
    "notify_on_fallback": true
  }
}
```

观察一周后，再决定是否加入：

```text
400
401
403
404
```

加入前必须问：

```text
这个错误在 4SAPI 日志里通常代表短时故障，还是配置错误？
```

短时故障可以 fallback。

配置错误不应该 fallback。

## 8. background_task 并发控制

多 Agent 和后台任务会带来并发。

OmO 配置里可以限制 provider 和 model 并发。

示例：

```jsonc
{
  "background_task": {
    "providerConcurrency": {
      "4sapi": 4
    },
    "modelConcurrency": {
      "4sapi/fast-model": 8,
      "4sapi/coding-model": 3,
      "4sapi/reasoning-model": 1
    }
  }
}
```

这个思路很重要。

便宜快模型可以并发高一点。

强推理模型并发低一点。

这样可以避免：

```text
多个后台任务同时打爆高成本模型。
```

如果你的 OpenCode Provider 名不是 `4sapi`，按实际 provider id 写。

## 9. 三套推荐路由模板

### 9.1 个人开发模板

适合个人和小项目。

```jsonc
{
  "agents": {
    "sisyphus": {
      "model": "4sapi/coding-model",
      "fallback_models": ["4sapi/fast-model"]
    },
    "oracle": {
      "model": "4sapi/reasoning-model"
    },
    "librarian": {
      "model": "4sapi/fast-model"
    },
    "explore": {
      "model": "4sapi/fast-model"
    }
  },
  "categories": {
    "quick": { "model": "4sapi/fast-model" },
    "deep": { "model": "4sapi/coding-model" },
    "ultrabrain": { "model": "4sapi/reasoning-model" },
    "writing": { "model": "4sapi/writing-model" }
  }
}
```

特点：

```text
简单。
成本可控。
强模型只给 Oracle / ultrabrain。
```

### 9.2 团队开发模板

适合 5 到 20 人团队。

```jsonc
{
  "agents": {
    "sisyphus": {
      "model": "4sapi/reasoning-model-a",
      "fallback_models": ["4sapi/reasoning-model-b"]
    },
    "hephaestus": {
      "model": "4sapi/coding-model-a",
      "fallback_models": ["4sapi/coding-model-b"]
    },
    "oracle": {
      "model": "4sapi/reasoning-model-b"
    },
    "librarian": {
      "model": "4sapi/fast-model"
    },
    "explore": {
      "model": "4sapi/fast-model"
    }
  },
  "categories": {
    "quick": {
      "model": "4sapi/fast-model"
    },
    "deep": {
      "model": "4sapi/coding-model-a",
      "fallback_models": ["4sapi/coding-model-b"]
    },
    "ultrabrain": {
      "model": "4sapi/reasoning-model-a",
      "fallback_models": ["4sapi/reasoning-model-b"]
    },
    "writing": {
      "model": "4sapi/writing-model"
    }
  },
  "runtime_fallback": {
    "enabled": true,
    "retry_on_errors": [429, 500, 502, 503, 504],
    "max_fallback_attempts": 3,
    "cooldown_seconds": 60,
    "timeout_seconds": 30,
    "notify_on_fallback": true
  }
}
```

特点：

```text
同档位 fallback。
不轻易降级。
适合日常研发。
```

### 9.3 成本优先模板

适合内容、文档、轻代码任务。

```jsonc
{
  "agents": {
    "sisyphus": {
      "model": "4sapi/fast-model",
      "fallback_models": ["4sapi/coding-model"]
    },
    "oracle": {
      "model": "4sapi/coding-model"
    },
    "librarian": {
      "model": "4sapi/fast-model"
    },
    "explore": {
      "model": "4sapi/fast-model"
    }
  },
  "categories": {
    "quick": { "model": "4sapi/fast-model" },
    "deep": { "model": "4sapi/coding-model" },
    "ultrabrain": { "model": "4sapi/reasoning-model" },
    "writing": { "model": "4sapi/writing-model" }
  }
}
```

特点：

```text
默认便宜。
只有 ultrabrain 用强模型。
适合内容工厂和轻量自动化。
```

## 10. 4SAPI 日志怎么反推路由

配置不是一次写完就不动。

要看日志。

每周看一次：

```text
哪个 Key 消耗最高？
哪个模型请求最多？
哪个模型失败率最高？
fallback 是否频繁触发？
fast 模型是否承担了过多复杂任务？
reasoning 模型是否被拿来做小任务？
Team Mode 是否触发了并发峰值？
```

然后调整。

如果发现：

```text
Explore 消耗很高。
```

说明搜索类任务太多，可能要缩小上下文或改快模型。

如果发现：

```text
ultrabrain 经常 fallback。
```

说明强模型不稳定，应该换同档备用。

如果发现：

```text
quick 任务经常失败后升级到 coding。
```

说明 quick 模型太弱，或者任务分类太宽。

如果发现：

```text
reasoning 模型请求数异常多。
```

说明路由过度使用强模型。

## 11. 不要让 fallback 变成无感烧钱

fallback 的风险是：

```text
用户以为只调用了一次。
实际上背后试了三次。
```

所以企业里建议：

```text
notify_on_fallback: true
```

并且每周看 4SAPI 日志。

如果 fallback 频繁发生，要查原因。

常见原因：

```text
主模型限流。
主模型权限不足。
模型名拼错。
请求上下文太长。
任务分类不合理。
网关上游不稳定。
```

不要简单把 fallback_attempts 提高。

那只是让错误更贵。

## 12. Key 拆分和路由一起设计

模型路由要和 4SAPI Key 设计放在一起。

建议：

```text
4sapi-omo-fast
用途：Explore、Librarian、quick。
```

```text
4sapi-omo-coding
用途：Hephaestus、deep、常规代码任务。
```

```text
4sapi-omo-reasoning
用途：Oracle、ultrabrain、Review。
```

```text
4sapi-omo-teammode
用途：Team Mode。
```

这样成本会非常清楚。

你能知道：

```text
是搜索贵。
还是代码修改贵。
还是 Review 贵。
还是 Team Mode 贵。
```

如果所有模型都走一把 Key，日志可读性会差很多。

## 13. 上线前测试

配置完不要直接上大任务。

按这个顺序测：

```text
1. fast 模型跑只读目录扫描。
2. coding 模型跑单文件小修改。
3. reasoning 模型跑架构评审，只读。
4. 手动制造一个可 fallback 的错误。
5. 确认 OmO 有 fallback 通知。
6. 确认 4SAPI 日志能看到请求和错误。
7. 限制并发后跑两个后台任务。
8. 观察是否超出预算。
```

测试通过后，再给团队用。

不要在真实业务任务里第一次测试 fallback。

## 14. 常见坑

第一，模型名不一致。

OpenCode 里显示的是：

```text
4sapi/model-id
```

OmO 里就要按实际显示写。

不要自己猜。

第二，fallback 到能力差太多的模型。

强推理任务不要直接 fallback 到 fast。

第三，把 401 / 403 / 404 全都自动 fallback。

这可能掩盖配置错误。

第四，没有并发限制。

后台任务和 Team Mode 会叠加消耗。

第五，不看日志。

没有日志，路由永远靠感觉。

第六，所有 Agent 一个模型。

这会浪费 OmO 的真正价值。

## 15. 最后总结

OmO 的模型路由，不是为了好看。

它解决的是企业级大模型接入里的核心问题：

```text
不同任务用不同模型。
失败时有备用模型。
高成本模型不被滥用。
并发不会打爆预算。
日志能反推策略。
```

4SAPI 的价值也不只是一个 Base URL。

它让你能看到：

```text
哪个 Key 在花钱。
哪个模型在失败。
哪个任务在重试。
哪个 Agent 应该降级。
哪个模型必须保留。
```

一句话：

```text
OmO 负责把任务分给合适的模型。
4SAPI 负责让每一次模型调用可查、可控、可复盘。
```

下一篇继续讲：

```text
OmO Ultimate 和 Codex Light 到底怎么选，企业是装 OpenCode 版、Codex 版，还是两个都装。
```

## 资料来源与延伸阅读

- oh-my-openagent GitHub：https://github.com/code-yeongyu/oh-my-openagent
- oh-my-openagent 配置文档：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/reference/configuration.md
- oh-my-openagent 安装指南：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/installation.md
- OpenCode Provider 文档：https://opencode.ai/docs/providers/
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
