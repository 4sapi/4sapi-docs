---
title: "【大模型API中转站】第24期 Hermes飞书模型切换 | 自定义厂商别跑偏"
category: 人工智能
tags:
  - 大模型API中转站
  - Hermes
  - 飞书机器人
  - Custom Provider
  - Claude
  - 4SAPI
description: "拆解 Hermes 飞书机器人使用 /model 切换模型时为什么会误走官方厂商，并给出厂商前缀、models 声明、per-platform 覆盖三种解决方案。"
---

# 【大模型API中转站】第24期 Hermes飞书模型切换 | 自定义厂商别跑偏

本文是【大模型API中转站】系列的第24篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型 API 的任督二脉。建议先收藏，随用随查。

上一篇讲了 Hermes Desktop 在 macOS 上的 GUI、Gateway 和会话持久化。这一篇继续解决一个更容易踩的坑：飞书机器人里明明已经配置了 4SAPI 自定义厂商，为什么输入 `/model claude-sonnet-4-6` 后，模型却跑到了官方 Anthropic？

这个问题的本质不是飞书，也不是 Claude，而是 Hermes 的模型路由规则：当你只输入一个裸模型名时，系统到底应该走哪个 provider？

## 1. 一个让人困惑的问题

你在飞书里和 Hermes 机器人聊得挺好，突然想把模型从 Opus 切到 Sonnet，于是在对话框输入：

```text
/model claude-sonnet-4-6
```

但返回的却像是官方 Anthropic 的响应，而不是你预期的 4SAPI 中转。明明 `config.yaml` 里已经配了自定义厂商，为什么它不认？

通常原因只有一个：你配置了 provider，但没有把这个模型名声明到该 provider 的 `models` 列表里。

## 2. Hermes 的模型路由优先级

可以把 Hermes 的模型选择理解成下面这条路：

```text
用户输入 /model
    ↓
解析模型名
    ↓
是否带厂商前缀？
    ├─ 是：按 provider/model 精确匹配
    └─ 否：按裸模型名自动推断
             ↓
       先看 custom_providers.models 是否声明
             ↓
       再看是否匹配官方厂商模型名
```

举例：

```text
/model 4s/claude-sonnet-4-6
```

这个写法带了 `4s/` 前缀，Hermes 会明确走 `4s` provider。

但如果你只写：

```text
/model claude-sonnet-4-6
```

Hermes 就要自己判断。假如 `custom_providers` 的 `4s.models` 里没有声明 `claude-sonnet-4-6`，它就可能按 `claude-*` 走官方 Anthropic 路由。

## 3. 为什么明明配了 4SAPI 还是跑偏

假设你的配置是这样：

```yaml
custom_providers:
- name: 4s
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: anthropic_messages
  models:
    claude-opus-4-7:
      name: claude-opus-4-7

model:
  default: claude-opus-4-7
  provider: 4s
```

这份配置里，`4s` provider 是存在的，默认模型也是 `claude-opus-4-7`。但 `models` 下面只声明了 `claude-opus-4-7`，没有 `claude-sonnet-4-6`。

所以当你发：

```text
/model claude-sonnet-4-6
```

Hermes 在 `4s.models` 里找不到这个模型名，就会继续按官方模型名推断。结果就是：你以为在走 4SAPI，实际可能走到了官方 Anthropic。

核心规律：

```text
只有声明在 custom_providers.models 里的模型，才会被稳定识别为该自定义厂商的模型。
```

## 4. 解法一：临时切换时加厂商前缀

最快的办法，是在飞书里明确告诉 Hermes 走哪个 provider：

```text
/model 4s/claude-sonnet-4-6
```

格式是：

```text
/model <provider>/<model>
```

优点：

- 不用改配置。
- 适合临时试模型。
- 能绕开裸模型名自动推断。

缺点：

- 团队成员要记住 provider 名。
- 会话重置或 Gateway 重启后，可能回到默认模型。
- 长期使用不如配置里声明清楚。

这适合排查问题，也适合临时对比模型效果。

## 5. 先确认你的 provider 名和模型名

在真正改配置前，建议先做一个小检查：你以为 provider 叫 `4s`，配置里可能叫 `4sapi`；你以为模型叫 `claude-sonnet-4-6`，4SAPI 模型广场里可能有更完整的名称。

建议把团队可用模型整理成一张表：

| 用途 | provider | 模型名 | 备注 |
| --- | --- | --- | --- |
| 默认飞书回复 | `4s` | `claude-sonnet-4-6` | 成本和效果均衡 |
| 深度推理 | `4s` | `claude-opus-4-7` | 低频使用 |
| 快速问答 | `4s` | `deepseek-chat` | 日常低成本 |
| GPT 兼容任务 | `4s-openai` | `gpt-4o` | 视协议拆分 |

这张表不需要复杂，但要统一。团队里最怕的是每个人都按自己的习惯命名，最后飞书命令、配置文件、文档全都对不上。

如果你不确定当前配置里到底有哪些 provider，可以先打开：

```bash
sed -n '1,220p' ~/.hermes/config.yaml
```

重点看三块：

```text
custom_providers.name
custom_providers.base_url
custom_providers.models
```

确认之后，再决定是用 `/model 4s/模型名` 临时切，还是把模型补进 `models`。

## 6. 解法二：把模型加入 4SAPI 厂商配置

更推荐的做法，是把常用模型都写进 `custom_providers.models`。

编辑：

```bash
~/.hermes/config.yaml
```

把 Sonnet、Opus、Haiku 等需要的模型都声明进去：

```yaml
custom_providers:
- name: 4s
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: anthropic_messages
  models:
    claude-opus-4-7:
      name: claude-opus-4-7
    claude-sonnet-4-6:
      name: claude-sonnet-4-6
    claude-haiku-3-5:
      name: claude-haiku-3-5

model:
  default: claude-opus-4-7
  provider: 4s
```

改完后重启 Gateway：

```bash
hermes gateway restart
```

或者在飞书里使用你已有的重启命令。重启后再发：

```text
/model claude-sonnet-4-6
```

这时 Hermes 能在 `4s.models` 里找到它，就会走 4SAPI。

## 7. 多协议时建议拆 provider

如果你同时接 Claude 类模型和 GPT 类模型，不建议为了省事全塞进一个 provider。不同模型可能使用不同协议格式，混在一起容易出现“模型找到了，但请求格式不对”的问题。

可以按协议拆成两个 provider：

```yaml
custom_providers:
- name: 4s-claude
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: anthropic_messages
  models:
    claude-sonnet-4-6:
      name: claude-sonnet-4-6
    claude-opus-4-7:
      name: claude-opus-4-7

- name: 4s-openai
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: openai_chat
  models:
    gpt-4o:
      name: gpt-4o
    deepseek-chat:
      name: deepseek-chat
```

然后飞书里切模型时写清楚：

```text
/model 4s-claude/claude-sonnet-4-6
/model 4s-openai/gpt-4o
```

这样配置略长，但排错更清楚。看到 provider 名，就知道走的是哪种协议。

如果你的 Hermes 版本对 `api_mode` 命名不同，以实际文档和当前配置为准。这里的重点不是字段名，而是“按协议拆 provider”这个思路。

## 8. 解法三：给飞书单独设置默认模型

如果你想让飞书默认用 Sonnet，本地桌面默认用 Opus，可以给平台单独设置模型。

示例：

```yaml
gateway:
  platforms:
    feishu:
      model: claude-sonnet-4-6
      provider: 4s
```

适合这类场景：

- 飞书群里多人共用一个机器人。
- 想让团队默认使用成本更稳的模型。
- 本地开发想保留更强或更贵的模型。
- 不希望每个人都手动 `/model` 切换。

这里的 `provider: 4s` 很关键。只写模型名，不写 provider，仍然可能回到裸模型名推断的问题。

如果你按协议拆了 provider，平台默认配置也要跟着改：

```yaml
gateway:
  platforms:
    feishu:
      model: claude-sonnet-4-6
      provider: 4s-claude
```

不要 provider 叫 `4s-claude`，平台里还写 `provider: 4s`。这类小错最难查，因为模型名看起来没问题，真正错的是路由入口。

## 9. 完整验证流程

配置完成后，建议按下面顺序验证。

```text
# 查看当前模型
/model

# 临时指定 4SAPI provider
/model 4s/claude-sonnet-4-6

# 如果已在 models 中声明，可以直接裸名切换
/model claude-sonnet-4-6

# 切回默认 Opus
/model 4s/claude-opus-4-7
```

如果切换成功，机器人通常会回复当前模型名。后续对话就会走对应模型处理。

如果还是跑偏，按这个顺序查：

```text
1. provider 名是不是 4s。
2. base_url 是否为 https://4sapi.com/v1。
3. api_key 是否有效。
4. api_mode 是否和模型协议匹配。
5. models 里是否声明了完整模型名。
6. 改配置后是否重启 Gateway。
```

更完整的验证可以分三步。

第一步，验证 provider 前缀：

```text
/model 4s/claude-sonnet-4-6
```

如果这一步成功，说明 provider、Key、base_url 大概率没问题。

第二步，验证裸模型名：

```text
/model claude-sonnet-4-6
```

如果前缀能成功、裸名失败，基本就是 `models` 声明或路由优先级问题。

第三步，验证平台默认：

```text
/new
/model
```

看新会话是否回到你期望的飞书默认模型。如果没有，检查 `gateway.platforms.feishu`。

## 10. 多模型配置时的几个细节

第一，不要手打模型名。

模型名建议从 4SAPI 模型广场复制。少一个版本号、大小写不一致、横线写错，都可能导致 Hermes 匹配不到。

第二，不同协议要分清。

Claude 类模型可能走 `anthropic_messages`，GPT 类模型可能走 OpenAI-compatible。具体以你的 Hermes 版本和 4SAPI 后台支持的协议为准。

第三，团队最好统一 provider 命名。

比如统一叫 `4s`，不要有人叫 `4sapi`、有人叫 `proxy`、有人叫 `claude4s`。provider 名不统一，飞书命令和配置都会变乱。

第四，常用模型提前声明。

不要等到群里临时切模型才发现没配。可以把团队常用模型一次性写进 `models`：

```yaml
models:
  claude-opus-4-7:
    name: claude-opus-4-7
  claude-sonnet-4-6:
    name: claude-sonnet-4-6
  gpt-4o:
    name: gpt-4o
  deepseek-chat:
    name: deepseek-chat
```

真实名称仍然以 4SAPI 模型广场为准。

## 11. 排错矩阵：看现象定位问题

| 现象 | 大概率原因 | 处理方式 |
| --- | --- | --- |
| `/model 4s/xxx` 失败 | provider、Key、base_url 有问题 | 查 `custom_providers.name/base_url/api_key` |
| `/model 4s/xxx` 成功，`/model xxx` 失败 | `models` 没声明或裸名被官方路由抢走 | 在 `models` 补模型，或坚持用前缀 |
| 切换后回复格式异常 | `api_mode` 和模型协议不匹配 | 按 Claude/GPT 协议拆 provider |
| 重启后又变回旧模型 | 只是会话级切换 | 写入 `model.default` 或平台覆盖 |
| 飞书和本地模型不一致 | per-platform 生效 | 检查 `gateway.platforms` |
| 404/model not found | 模型名不完整 | 从 4SAPI 模型广场复制 |
| 401/403 | Key 或权限问题 | 换项目 Key，检查额度和权限 |

这张表建议直接放到团队内部文档里。以后飞书机器人“突然换错模型”，不用每次都从头猜。

## 12. 团队配置模板

如果是团队共用飞书机器人，可以参考下面这个模板：

```yaml
custom_providers:
- name: 4s-claude
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: anthropic_messages
  models:
    claude-sonnet-4-6:
      name: claude-sonnet-4-6
    claude-opus-4-7:
      name: claude-opus-4-7

- name: 4s-openai
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: openai_chat
  models:
    gpt-4o:
      name: gpt-4o
    deepseek-chat:
      name: deepseek-chat

model:
  default: claude-sonnet-4-6
  provider: 4s-claude

gateway:
  platforms:
    feishu:
      model: claude-sonnet-4-6
      provider: 4s-claude
```

这个模板的目标不是让你一字不改照抄，而是建立一个清晰结构：

```text
provider 按协议拆
models 明确声明
默认模型写全
平台覆盖写 provider
Key 由 4SAPI 项目级令牌提供
```

团队里如果要限制成本，可以在 4SAPI 里给飞书机器人单独建一个项目 Key，只开放日常模型；Opus 这类高成本模型放到另一个 review 或 admin Key 里，避免群聊里随手切到高价模型。

## 13. 常见问题

**Q：每个自定义厂商都要配 models 列表吗？**

是的。每个 provider 都有自己的模型声明。用 4SAPI 的好处是，一个 `base_url`、一个项目 Key 可以统一管理多模型，但 Hermes 侧仍然需要知道哪些模型属于这个 provider。

**Q：为什么不一直用 `/model 4s/模型名`？**

临时排查可以，长期不推荐。团队成员会忘，命令会变长，也不利于默认模型管理。工程上更稳的方式，是把常用模型声明到 `models`，再用 platform 默认模型控制。

**Q：Gateway 重启后 `/model` 切换会保留吗？**

通常会话级切换不应该当成永久配置。想长期默认用某个模型，写进 `model.default` 或 `gateway.platforms.feishu`。

**Q：GPT 模型也能这么配吗？**

可以，但要注意协议和 `api_mode`。如果同一个中转站同时支持 Claude 和 GPT，建议按协议拆 provider，或者严格按 Hermes 支持的配置方式写。

**Q：能不能给模型起别名？**

如果 Hermes 当前版本支持 alias 或映射，可以用别名降低命令复杂度；如果不支持，不要在团队文档里创造“口头别名”。最稳的是 provider 和 model 都使用配置文件里的真实名称。

**Q：为什么建议用 4SAPI 项目 Key，而不是个人 Key？**

因为飞书机器人是团队入口。项目 Key 可以单独限额、单独停用、单独看日志；个人 Key 泄露或超额时影响面更难控制。

## 14. 总结

这类问题一句话就能讲清楚：

```text
裸模型名容易被 Hermes 自动路由到官方厂商；
想稳定走 4SAPI，要么加 provider 前缀，要么在 custom_providers.models 里声明。
```

推荐优先级：

| 场景 | 推荐做法 |
| --- | --- |
| 临时试模型 | `/model 4s/模型名` |
| 长期稳定使用 | 在 `custom_providers.models` 声明模型 |
| 飞书团队统一默认模型 | `gateway.platforms.feishu` 指定 model 和 provider |
| 多模型统一管理 | 用 4SAPI 做统一 `base_url` 和项目 Key |

如果你已经用 Hermes 接飞书机器人，建议尽早把模型配置整理成“provider + models + platform override”三层结构。再往前一步，按协议拆 provider，用 4SAPI 项目 Key 做额度和权限隔离。Hermes 负责多入口 Agent 工作流，4SAPI 负责多模型 API 接入、账单和路由。这样后面无论切 Claude、GPT、GLM 还是 DeepSeek，都不会在飞书里越切越乱。

参考资料：

- Hermes 官方网站：https://hermes-agent.nousresearch.com
- 4SAPI 接入地址：https://4sapi.com/v1
