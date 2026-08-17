---
title: "【大模型API中转站】第127期 Codex浏览器报错 | trusted bridge排查"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Claude
  - Windows
  - Browser
  - Computer Use
  - 企业级大模型接入
  - 4SAPI
description: "基于 GitHub openai/codex issue #23222，复盘 Windows Codex 桌面版浏览器能手动打开、但无法被 Agent 自动控制的问题。文章给出 browser-client is not trusted、native pipe helper paths、bundled executable relocation failed 三类报错的排查顺序，并演示如何用 4SAPI 接 Claude/Codex 分析日志。"
---

# 【大模型API中转站】第127期 Codex浏览器报错 | trusted bridge排查

本文是【大模型API中转站】系列的第127篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这一篇写一个非常具体、也很容易误判的问题：

```text
Codex 桌面版里的浏览器能手动打开，
但让 Codex 自动搜索、点击、导航时全部失败。
```

这个问题来自 GitHub 上 openai/codex 的一个真实 issue：

```text
openai/codex #23222：In-app browser commands stopped working
```

用户的现象很典型：

```text
浏览器窗格是打开的。
手动输入网址正常。
手动点页面正常。
但 Codex 不能通过浏览器自动化去点击、输入、跳转。
```

错误里最刺眼的一句是：

```text
browser-client is not trusted
```

如果你没做过 Codex、Browser、Computer Use、Node helper 这一层排查，很容易第一反应就去重装 Chrome、重装扩展、换网络、清 Cookie。

但这类问题的关键通常不在网页，也不在 Google、OpenAI 官网或某个具体站点。

它更像是：

```text
Codex App 的可视浏览器在，
但 Agent 到浏览器之间的可信控制桥断了。
```

这篇我按实战排查来写。

我会把它拆成三层：

```text
第一层：页面问题还是控制桥问题。
第二层：Browser / Computer Use helper 是否能启动。
第三层：Windows 本地缓存和 sandbox 配置是否卡住 helper。
```

中间我也会放一段我常用的做法：

```text
用 4SAPI 接 Claude，
把 issue、日志、配置片段丢进去，
让模型先做一次“工程排错分流”。
```

不要直接让 AI 猜修复命令。

要让 AI 先判断：

```text
这个报错属于哪一层。
哪些证据支持。
哪些动作风险低。
哪些动作应该放到最后。
```

这才是企业级大模型接入里更稳的用法。

## 1. 先判断：这不是网页点击失败

很多人看到“浏览器不能点”，会把它理解成：

```text
网页选择器变了。
Google 页面有反爬。
按钮没有加载出来。
浏览器版本不兼容。
```

但 issue #23222 的现象更早。

它还没进入页面级别的点击逻辑，就已经在浏览器客户端初始化阶段失败。

判断标准很简单：

```text
如果任何网页都不能控制，
包括打开 about:blank、访问 openai.com、点 Google Images，
那就不要先查网页 DOM。
```

这时候应该先查：

```text
Codex 是否拿到了浏览器自动化工具。
Browser helper 是否可用。
Computer Use native pipe 是否启动。
node_repl / browser-client 是否在可信环境里运行。
```

你可以把这类故障理解成三段链路：

```text
Codex Agent
  ↓
可信工具层：browser / computer-use / node_repl
  ↓
浏览器窗格或 Chrome
```

如果中间的可信工具层没注入，或者 native pipe 没起来，页面再正常也没用。

这就是为什么有些用户会看到：

```text
浏览器能手动用，
但 Codex 不能操作。
```

两者不是同一条路径。

手动浏览器只需要窗口能打开。

Agent 浏览器控制需要额外的可信桥。

## 2. 这类报错的三个关键词

从 issue 里看，最重要的信号有三个。

第一个是：

```text
browser-client is not trusted
```

这句话不要翻译成“浏览器不可信”。

更准确的理解是：

```text
当前运行 browser-client 的 Node 环境，
没有拿到 Codex 授权的 privileged native pipe。
```

换句话说，如果你在普通 PowerShell 里直接 import 某个 browser-client 文件，它本来就可能报这个错。

因为普通 Node 不是 Codex 的可信工具运行时。

第二个信号是：

```text
Windows Computer Use helper paths are unavailable
```

这说明不是页面点击失败，而是 Computer Use 这一层 helper 路径没有准备好。

在 Windows 上，Codex 桌面版通常需要一些本地 helper、node、node_repl、浏览器桥接文件配合工作。

如果这些路径没有被正确释放、复制、注册或暴露给当前 thread，浏览器工具就会“看起来存在，但实际不可用”。

第三个信号是：

```text
bundled_executable_relocation_failed
```

这个信号很关键。

它指向的是：

```text
Codex 打包自带的可执行文件，
没有成功搬到本地可运行目录。
```

issue 中提到受影响的包括：

```text
rg.exe
codex.exe
node.exe
node_repl.exe
```

如果 node_repl 都没正确就位，浏览器自动化桥就更难正常工作。

所以这不是“换一个网页试试”的问题。

它是 Codex 桌面版在 Windows 本地运行态、缓存态、工具注入态之间的连接问题。

## 3. 我是怎么用 4SAPI 接 Claude 分析这个问题的

这里不是让 Claude 代替你修电脑。

更合理的用法是：

```text
把报错、系统环境、已尝试动作和日志片段整理成结构化材料，
让 Claude 先给出排查树。
```

比如我会把材料整理成这样：

```text
目标：
分析 Windows Codex 桌面版 in-app browser 自动化失效。

现象：
浏览器可手动打开，Codex 无法点击、输入、导航。

关键错误：
1. browser-client is not trusted
2. Windows Computer Use helper paths are unavailable
3. bundled_executable_relocation_failed

已尝试：
重装 Codex、重启电脑、重新打开浏览器，仍失败。

要求：
不要泛泛建议重装。
请按“页面层、工具层、缓存层、权限层、sandbox层”给排查顺序。
每一步说明证据、风险和回滚方式。
```

如果你用 4SAPI 的 OpenAI 兼容接口，可以用下面这个最小脚本。

注意：模型名称以你的 4SAPI 后台可用模型为准，下面只是写法示例。

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["SAPI_API_KEY"],
    base_url=os.getenv("SAPI_BASE_URL", "https://4sapi.com/v1"),
)

issue_context = """
Windows Codex desktop in-app browser can be used manually,
but Agent browser automation cannot click, type, or navigate.

Key errors:
- browser-client is not trusted
- Windows Computer Use helper paths are unavailable
- bundled_executable_relocation_failed

Already tried reinstalling Codex and restarting Windows.
Please build a diagnostic tree.
"""

resp = client.chat.completions.create(
    model=os.getenv("SAPI_MODEL", "claude-sonnet-4-5-20250929"),
    temperature=0.2,
    messages=[
        {
            "role": "system",
            "content": (
                "你是 Windows 桌面应用和 Agent 工具链排错工程师。"
                "请根据证据分层判断，不要跳步，不要编造官方结论。"
            ),
        },
        {
            "role": "user",
            "content": issue_context,
        },
    ],
)

print(resp.choices[0].message.content)
```

我建议你让模型输出这几栏：

```text
可能原因
证据
反证
低风险验证
修复动作
回滚方式
```

这比直接问：

```text
这个报错怎么修？
```

要稳很多。

因为这类问题不是一个错误码对应一个答案。

同样是浏览器自动化失败，可能来自：

```text
本地缓存损坏。
工具没有注入当前 thread。
Chrome 扩展装了但 runtime 没暴露。
Codex helper relocation 失败。
Windows sandbox 权限不匹配。
企业安全软件阻止本地 pipe。
```

AI 的价值不是替你乱试。

它的价值是帮你把排错路径排成顺序。

## 4. 第一轮排查：先确认工具面是不是可用

如果你正在用 Codex 桌面版，可以先问它：

```text
请检查当前线程里是否有 browser、chrome、node_repl 或 computer-use 相关工具可用。
不要操作网页，只报告工具可用性和失败原因。
```

你要看的不是它能不能打开 Google。

你要看的是：

```text
当前 thread 是否暴露了 browser 工具。
当前 thread 是否暴露了 computer-use 工具。
是否存在 node_repl 这类可信执行工具。
是否能创建或连接浏览器 tab。
```

如果 Codex 回复类似：

```text
没有浏览器控制工具。
找不到 node_repl。
无法连接 native pipe。
```

那就不要继续让它点网页了。

继续点只会得到更多无效失败。

这时排查方向应该从页面转到本地运行态。

## 5. 第二轮排查：看 Windows 本地目录

在 Windows 上，可以检查几个目录。

先看 Codex 用户目录：

```powershell
Get-ChildItem -Force "$env:USERPROFILE\.codex"
```

再看 Codex 本地应用数据：

```powershell
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI"
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI\Codex"
```

如果报错里出现 bundled executable relocation failed，就重点看：

```powershell
Test-Path "$env:LOCALAPPDATA\OpenAI\Codex\bin"
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI\Codex\bin" -ErrorAction SilentlyContinue
```

正常情况下，这里应该能看到 Codex 运行需要的一些本地可执行文件。

如果目录不存在、空、权限异常，或者文件释放不完整，浏览器 helper 就可能启动失败。

注意，这一步不要急着删除。

先观察。

企业环境尤其要先截图或导出目录状态，方便回滚和报障。

## 6. 第三轮排查：区分 .codex 和 LocalAppData

很多人排查 Codex，会第一时间改：

```text
C:\Users\<你>\.codex
```

这里确实很重要。

它通常放：

```text
config.toml
插件缓存
技能配置
MCP 配置
线程相关缓存
```

但 issue #23222 里有用户反馈，重命名或重建 `.codex` 并没有解决问题。

真正起作用的是重建：

```text
%LOCALAPPDATA%\OpenAI
```

为什么？

因为这里更接近 Windows App 本地运行数据。

包括：

```text
桌面应用缓存
打包 helper 的释放目录
本地 bin
运行时状态
工具桥接需要的部分本地文件
```

简单说：

```text
.codex 更像用户配置。
%LOCALAPPDATA%\OpenAI 更像 App 本地运行态。
```

如果报错是配置找不到，优先看 `.codex`。

如果报错是 helper path、native pipe、bundled executable relocation，优先看 `%LOCALAPPDATA%\OpenAI`。

## 7. 一个低风险修复：重建 OpenAI 本地应用数据

根据 issue 里的反馈，第一种可以尝试的修复是：

```text
关闭 Codex。
把 %LOCALAPPDATA%\OpenAI 改名成 OpenAI.old。
重新打开 Codex，让它重建目录。
```

PowerShell 写法如下：

```powershell
Stop-Process -Name Codex -ErrorAction SilentlyContinue

$src = Join-Path $env:LOCALAPPDATA "OpenAI"
$bak = Join-Path $env:LOCALAPPDATA ("OpenAI.old-" + (Get-Date -Format "yyyyMMddHHmmss"))

if (Test-Path $src) {
    Rename-Item -LiteralPath $src -NewName (Split-Path $bak -Leaf)
}
```

然后重新打开 Codex。

再测试一个最小动作：

```text
打开 in-app browser。
让 Codex 访问 about:blank。
再让它打开 https://openai.com。
```

不要一上来就测试复杂网页。

最小验证应该是：

```text
能否创建 tab。
能否导航。
能否读取 URL/title。
能否关闭 tab。
```

如果这些都可以，再去测试搜索、点击、输入。

## 8. 为什么不建议先重装 Chrome

因为这个 issue 的错误证据不指向 Chrome 页面本身。

如果你看到的是：

```text
Chrome 扩展未安装。
native host manifest 不正确。
浏览器版本不兼容。
```

那当然要查 Chrome 扩展。

但这里的主线是：

```text
Codex runtime 没有拿到可信工具桥。
Windows helper path 不可用。
bundled executable 没能 relocation。
```

这三个都在 Codex App 本地运行层。

重装 Chrome 大概率只是绕远路。

同理，也不要第一时间删除插件缓存。

如果插件文件存在，但当前 thread 没有暴露工具，问题可能在工具注入或 sandbox。

删除插件缓存未必能解决，反而会让变量更多。

## 9. 如果重建 LocalAppData 仍不行

issue 里也有用户反馈：

```text
重建 %LOCALAPPDATA%\OpenAI 不一定对所有人有效。
```

这很正常。

因为同一类表象可能有不同根因。

如果重建后仍失败，下一步要查：

```text
1. Codex config.toml 里的 windows sandbox 设置。
2. .codex\.sandbox 下是否有 os error 740。
3. 当前线程是否暴露 browser / node_repl 工具。
4. 企业安全软件是否阻止 helper 启动。
5. Microsoft Store 包路径是否存在访问限制。
```

尤其是：

```text
os error 740
```

这个错误通常和 Windows 权限提升有关。

它会把排查方向从“缓存损坏”转到“sandbox 权限不匹配”。

这个我会放到第129期专门讲。

## 10. 企业团队怎么沉淀这类排查

个人用户可以靠经验试。

企业团队不能这么做。

因为团队里有几十台 Windows 机器时，同一个“Codex 浏览器不能用”，背后可能是：

```text
版本不一致。
配置不一致。
权限策略不一致。
安全软件拦截不一致。
AppData 缓存状态不一致。
代理、网络、证书策略不一致。
```

所以建议把排查沉淀成一个模板。

最少包含：

```text
Codex App 版本
Windows 版本
安装来源
是否 Microsoft Store
是否管理员启动
config.toml 的 windows sandbox
%LOCALAPPDATA%\OpenAI\Codex\bin 是否存在
.codex\.sandbox 是否有错误日志
当前 thread 是否暴露 browser / node_repl
是否安装 Chrome plugin
是否能手动打开浏览器
是否能自动打开 about:blank
```

然后用 4SAPI 给排查 Agent 单独开一个 Key。

这个 Key 只用于：

```text
分析日志。
归类错误。
生成排查报告。
输出低风险修复步骤。
```

不要把生产业务 Key 给排障 Agent。

这就是企业级 API 网关的意义：

```text
同样是 Claude 或 Codex 类模型能力，
业务调用、排障调用、研发助手调用要分 Key、分日志、分预算。
```

## 11. 一份可直接复制的 AI 排障提示词

你可以把下面这段复制到接了 4SAPI 的 Claude、GPT 或 Codex 类 Agent 里。

```text
你是 Windows Codex 桌面版排障工程师。

我遇到的问题：
Codex in-app browser 可以手动打开网页，
但 Codex 不能通过 browser automation 点击、输入或导航。

关键错误：
1. browser-client is not trusted
2. Windows Computer Use helper paths are unavailable
3. bundled_executable_relocation_failed

请按以下结构分析：
1. 这是不是网页 DOM 问题？
2. 这是不是 Browser/Computer Use helper 问题？
3. 应该先检查哪些目录和配置？
4. 哪些动作低风险，可以先做？
5. 哪些动作有破坏性，必须备份？
6. 如果修复失败，下一步看什么日志？

输出要求：
不要直接让我重装系统。
不要让我删除目录，除非先给备份命令。
每个步骤都要说明验证方式。
```

如果模型只回答“重装 Codex”，说明提示词还不够约束。

你可以继续追问：

```text
请只基于我给出的三个错误信号，
把可能原因按概率排序，
并说明每个原因对应的最小验证命令。
```

这类追问非常重要。

AI 排错不是问一次就结束。

正确姿势是：

```text
先让它分层。
再让它列证据。
再让它给最小验证。
最后才执行修复动作。
```

## 12. 本篇结论

Codex 浏览器自动化失效，不要一上来就看网页。

先判断它卡在哪一层：

```text
页面层：选择器、按钮、加载、跨域。
工具层：browser、computer-use、node_repl 是否可用。
本地运行层：helper、native pipe、bin 文件是否可用。
缓存层：%LOCALAPPDATA%\OpenAI 是否损坏。
权限层：Windows sandbox、管理员权限、os error 740。
```

issue #23222 这类报错的重点是：

```text
浏览器能手动用，不代表 Agent 能控制。
```

如果看到：

```text
browser-client is not trusted
Windows Computer Use helper paths are unavailable
bundled_executable_relocation_failed
```

优先查 Codex 本地工具桥和 Windows 本地运行态。

低风险第一步通常是：

```text
备份并重建 %LOCALAPPDATA%\OpenAI。
```

但如果日志里出现：

```text
os error 740
```

就要进一步查 sandbox 配置。

这就是下一篇要讲的内容：

```text
为什么 sandbox = "elevated" 可能导致浏览器 helper 起不来。
```

## 资料来源与延伸阅读

- GitHub issue：openai/codex #23222：https://github.com/openai/codex/issues/23222
- OpenAI Codex 文档：https://developers.openai.com/codex/
- OpenAI Codex In-app browser：https://developers.openai.com/codex/app/browser
- OpenAI Codex Agent Approvals and Security：https://developers.openai.com/codex/agent-approvals-security
- OpenAI Codex Windows：https://developers.openai.com/codex/windows
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
