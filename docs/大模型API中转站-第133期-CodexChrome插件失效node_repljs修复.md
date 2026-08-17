---
title: "【大模型API中转站】第133期 Codex Chrome插件失效 | node_repl js修复"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Chrome插件
  - MCP
  - node_repl
  - Browser
  - 开发调优
  - 软件开发
  - 企业级大模型接入
  - 4SAPI
description: "复盘 Mac 端 Codex Chrome 插件失效、提示当前线程没有 mcp__node_repl__js 的排障思路。文章讲清 Control Chrome skill、Node REPL js 工具、browser-client.mjs 和 Chrome 扩展之间的关系，并给出可复制给模型的修复提示词、验证清单和企业团队排障规范。"
---

# 【大模型API中转站】第133期 Codex Chrome插件失效 | node_repl js修复

本文是【大模型API中转站】系列的第133篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前面几篇连续写了 Codex 浏览器自动化排障：

```text
第127期：trusted bridge 报错怎么分层。
第128期：Windows AppData 缓存怎么重建。
第129期：sandbox = elevated 导致 os error 740 怎么处理。
第130期：完全访问权限打开后意味着什么。
第131期：Windows 中文输出乱码怎么先用 chcp 65001 排查。
```

第132期单独写了 WorkBuddy 完结篇。

这一期再回到 Codex 排障线。

这一篇换到另一个很细，但非常实用的问题：

```text
Mac 端 Codex 的 Chrome 插件突然失效。
模型提示当前线程没有 mcp__node_repl__js。
```

这个问题来自 Linux.do 上一个真实讨论线索：

```text
https://linux.do/t/topic/2513640
```

原帖的核心经验很短：

```text
告诉大模型去找 mcp__node_repl.js 就可以解决。
猜测是 Codex 更新以后没有同步更新 Control Chrome 的 skill。
```

这句话看起来像一句小技巧。

但它背后其实牵出了一条完整的 Codex Chrome 控制链路：

```text
Codex 当前线程
-> 工具发现 / MCP 工具注册
-> Node REPL 的 js 执行工具
-> Control Chrome skill
-> browser-client.mjs
-> Chrome 扩展
-> 用户自己的 Chrome 标签页
```

所以这篇不只写“让模型去找工具”。

我会把它拆开讲：

```text
1. 这个报错到底说明什么。
2. 为什么 Chrome 插件没坏，但 Codex 仍然控制不了。
3. mcp__node_repl__js 和 mcp__node_repl.js 是什么关系。
4. 复制给模型的修复提示词怎么写。
5. Mac 端该怎么排查，哪些 Windows 修复不要照搬。
6. 企业团队怎么把这类问题做成排障卡片。
```

如果你只是想要结论，可以先看这一段。

遇到：

```text
当前线程没有 mcp__node_repl__js
无法使用 Chrome 插件
Chrome extension unavailable
Control Chrome skill 无法启动
```

不要第一反应就重装 Chrome，也不要直接让模型换 Playwright。

先告诉模型：

```text
当前任务需要控制用户自己的 Chrome 插件。
请先通过工具发现搜索 node_repl js，找到 Node REPL 的 js 执行工具。
它通常叫 mcp__node_repl__js，有些界面会显示成 mcp__node_repl.js。
找到以后再按 Control Chrome skill 初始化 browser-client.mjs。
不要用 js_reset 或 js_add_node_module_dir 代替 js 工具。
如果仍找不到，请明确说明当前线程没有暴露 Node REPL js 工具。
```

很多时候，问题就卡在这里。

## 1. 先分清：这是 Chrome 坏了，还是线程没拿到工具

这类问题最容易误判。

用户看到“Chrome 插件失效”，第一反应通常是：

```text
Chrome 扩展坏了。
Chrome 版本太新。
Codex 更新后扩展不兼容。
浏览器权限没开。
需要重新安装插件。
```

这些都有可能。

但如果报错里明确出现：

```text
current thread does not have mcp__node_repl__js
当前线程没有 mcp__node_repl__js
```

那优先级就变了。

这不是网页按钮点不到，也不是 Chrome 标签页打不开。

它更像是：

```text
Codex 想走 Chrome 插件控制链路，
但当前线程没有拿到启动这条链路的 js 执行工具。
```

也就是说，Chrome 可能完全正常。

插件可能也已经安装。

你的登录态、Cookie、扩展权限、Chrome Profile 也可能都没问题。

真正断掉的是：

```text
Codex 线程 -> Node REPL js 工具 -> browser-client.mjs
```

这一段。

所以第一步不是重装。

第一步是让模型重新发现工具。

## 2. 为什么是 node_repl 的 js 工具

Codex 里控制 Chrome 插件，不是随便调用一个浏览器自动化工具就行。

它和普通 Playwright、内置浏览器、网页搜索不是同一个入口。

大致可以这样理解：

```text
Playwright / in-app browser：
控制 Codex 自己开的浏览器或测试浏览器。

Control Chrome：
控制用户本机 Chrome，依赖 Chrome 扩展和用户当前 Chrome Profile。
```

这两条线能力不同。

如果你要测试一个公开网页，用 Codex 的内置浏览器就够了。

如果你要操作用户已经登录的 Chrome，例如：

```text
打开用户 Chrome 里已有的网页。
使用用户已经登录的网站。
读取当前 Chrome 标签页可见内容。
在用户 Chrome Profile 下继续工作流。
依赖 Chrome 扩展建立控制桥。
```

就需要 Control Chrome 这条线。

而 Control Chrome 的核心初始化动作，要通过 Node REPL 的 `js` 工具来跑。

本机插件说明里有几个关键信息：

```text
Chrome 控制入口不是内置 browser-client 包，
而是插件目录里的 scripts/browser-client.mjs。

初始化需要通过 Node REPL 的 js 工具执行。

这个 js 工具通常暴露为 mcp__node_repl__js。

只有这个 Node REPL js 工具能控制 Chrome extension 这一面。
```

所以当当前线程没有 `mcp__node_repl__js` 时，模型就算知道“应该控制 Chrome”，也缺少真正的手。

它可能会误走几条路：

```text
改用 mcp__playwright。
改用 in-app browser。
要求用户重装 Chrome。
要求用户重新登录网站。
调用 js_reset。
调用 js_add_node_module_dir。
```

这些都不一定能解决。

因为问题不是 JavaScript 状态没清，也不是 Node 包路径没加。

问题是：

```text
当前线程没有把 js 执行工具暴露给模型，
或者模型没有主动通过工具发现去找它。
```

## 3. mcp__node_repl__js 和 mcp__node_repl.js 到底差在哪

这里还有一个非常容易混淆的细节。

不同界面、日志、提示词里，对同一个工具的显示方式可能不一样。

你可能看到：

```text
mcp__node_repl__js
```

也可能看到：

```text
mcp__node_repl.js
```

两者想表达的通常是同一件事：

```text
node_repl 这个 MCP server 里的 js 工具。
```

只是展示格式不同。

在 Codex 工具命名里，常见形式是：

```text
mcp__服务名__工具名
```

所以更严格的写法是：

```text
mcp__node_repl__js
```

而有些 UI 或文章为了好读，会写成：

```text
mcp__node_repl.js
```

也就是：

```text
node_repl.js
```

读者在排障时不用卡在这个点上。

你真正要告诉模型的是：

```text
去找 node_repl 的 js 执行工具。
```

如果模型只搜索完整字符串 `mcp__node_repl__js`，有时可能搜不到。

更稳的是让它搜：

```text
node_repl js
```

这也是为什么这类提示词里，不建议只写一句“调用 mcp__node_repl__js”。

更好的写法是：

```text
如果当前工具列表没有直接显示 mcp__node_repl__js，
请使用工具发现搜索 node_repl js。
```

这句话给了模型一条恢复路径。

## 4. 复制给模型的修复提示词

下面这段可以直接复制给 Codex。

适用场景是：

```text
你已经安装并启用了 Codex 的 Chrome 插件。
你希望 Codex 操作用户自己的 Chrome。
但模型说当前线程没有 mcp__node_repl__js。
```

提示词如下：

```text
当前任务需要使用 Codex 的 Control Chrome 能力，控制用户本机 Chrome 插件，而不是使用 in-app browser 或普通 Playwright。

请先检查当前线程是否有 Node REPL 的 js 执行工具。这个工具通常叫 mcp__node_repl__js，有些界面可能显示为 mcp__node_repl.js。

如果当前工具列表没有直接显示，请使用工具发现搜索 node_repl js，不要只搜索完整工具名。找到 js 工具后，再按 Control Chrome skill 的要求初始化插件目录下的 scripts/browser-client.mjs。

注意：js_reset 只是重置 Node REPL 状态，js_add_node_module_dir 只是添加模块搜索路径，它们都不能代替 js 执行工具。不要用 mcp__playwright 或 in-app browser 替代 Control Chrome，因为我要控制的是用户自己的 Chrome Profile。

如果搜索后仍然没有 js 工具，请明确告诉我：当前线程没有暴露 Node REPL js 工具。然后建议我新开线程、重启 Codex、检查插件是否安装启用，或等待 Codex/插件缓存刷新。
```

如果你想更短一点，可以用这版：

```text
Chrome 插件控制需要 Node REPL 的 js 工具。请先通过工具发现搜索 node_repl js，找到 mcp__node_repl__js 或界面显示的 mcp__node_repl.js，再初始化 Control Chrome 的 browser-client.mjs。不要用 js_reset、js_add_node_module_dir、Playwright 或 in-app browser 代替。
```

这段提示词的关键不是“魔法字符串”。

关键是给模型明确四件事：

```text
我要的是用户 Chrome，不是内置浏览器。
我要的是 Node REPL 的 js 工具，不是 reset/helper。
如果工具没显示，请先工具发现。
找不到就报告线程工具缺失，不要乱修 Chrome。
```

## 5. 正确排障顺序

遇到这个问题时，我建议按下面顺序查。

不要一上来就删缓存。

不要一上来就重装扩展。

先做低风险判断。

### 第一步：确认报错关键词

先看模型或日志里有没有这些词：

```text
mcp__node_repl__js
node_repl js
current thread
tool not available
Control Chrome
Chrome extension
```

如果核心报错是：

```text
当前线程没有 mcp__node_repl__js
```

那就先按“工具发现”处理。

如果核心报错是：

```text
Chrome extension disconnected
browser extension not installed
cannot connect to extension
```

那才更偏向 Chrome 插件连接问题。

如果核心报错是：

```text
browser-client is not trusted
native pipe failed
helper path failed
```

那就回到第127期那条 trusted bridge / helper 排障线。

不要把所有“Chrome 控制失败”都当成同一种问题。

### 第二步：让模型搜索 node_repl js

给模型一个明确动作：

```text
请使用工具发现搜索 node_repl js。
```

如果第一次没有结果，再让它放宽：

```text
请搜索 node_repl js，返回更多结果，例如 limit: 10。
```

这里有一个很重要的点：

```text
不要让模型调用 js_reset 来“找 js”。
```

`js_reset` 的作用是清理 Node REPL 会话状态。

它不是 JavaScript 执行入口。

`js_add_node_module_dir` 的作用是添加 Node 模块搜索目录。

它也不是 JavaScript 执行入口。

真正需要的是：

```text
js
```

也就是能执行：

```js
await import("...")
nodeRepl.write(...)
```

这一类代码的工具。

### 第三步：确认 browser-client.mjs 是否存在

Control Chrome 的核心入口通常在插件目录里：

```text
scripts/browser-client.mjs
```

如果这个文件不存在，那就不是简单提示词能解决的。

这说明插件缓存、插件安装、版本更新或 bundle 内容可能有问题。

模型应该停止瞎猜，并报告：

```text
当前 Chrome 插件缺少 scripts/browser-client.mjs。
```

这时候再考虑：

```text
重启 Codex。
刷新插件。
重新安装插件。
检查插件版本。
检查 Codex 更新后缓存是否同步。
```

但只要 `browser-client.mjs` 存在，且 `node_repl js` 工具能找到，很多场景就不需要动 Chrome 本体。

### 第四步：确认是不是新线程问题

这个报错里有一个词很关键：

```text
current thread
```

当前线程没有，不代表整个 Codex 永久没有。

可能只是：

```text
这个线程启动时工具列表没暴露。
Codex 更新后旧线程没有刷新工具能力。
Control Chrome skill 更新了，但线程上下文仍按旧能力运行。
插件缓存更新了，但当前会话没有重新加载。
```

所以简单修复常常是：

```text
新开一个线程。
重启 Codex App。
确认插件启用。
再让模型通过工具发现找 node_repl js。
```

这个顺序比重装 Chrome 更轻。

也更容易保留你的浏览器登录状态和工作现场。

## 6. 为什么 Codex 更新后容易出现这个问题

这个问题很像“版本错位”。

不是说某一方一定坏了。

更像是几层东西更新节奏不一致：

```text
Codex App 更新了。
Chrome 插件或 openai-bundled 插件更新了。
Control Chrome skill 文档更新了。
当前线程的工具暴露状态没有同步。
模型仍按旧工具列表理解环境。
```

于是就出现一种尴尬情况：

```text
新的 skill 知道应该用 node_repl js。
但当前线程看不到 mcp__node_repl__js。
模型不知道应该先做工具发现。
用户只看到 Chrome 插件失效。
```

从用户视角看，这像是“Chrome 插件挂了”。

从工程视角看，它更像是“工具路由没接上”。

这也是为什么一句：

```text
去找 mcp__node_repl.js
```

会有效。

这句话不是修复 Chrome。

它是在提醒模型：

```text
你先把正确的工具入口找回来。
```

一旦入口找回来，后面的 Control Chrome 初始化流程才有机会继续。

## 7. Mac 端不要照搬 Windows 修复

前面几篇写过 Windows 的 Codex 排障。

但这篇的场景是 Mac 端 Chrome 插件失效。

所以不要直接照搬这些动作：

```text
重建 %LOCALAPPDATA%\OpenAI。
修改 Windows sandbox = elevated / unelevated。
排查 os error 740。
运行 chcp 65001。
```

这些都是 Windows 语境里的修复。

Mac 端遇到当前线程没有 `mcp__node_repl__js`，更应该优先查：

```text
当前线程是否暴露 Node REPL js 工具。
Control Chrome skill 是否被正确加载。
Chrome 插件是否启用。
Codex App 更新后是否需要重启。
旧线程是否需要新开。
插件缓存是否完成刷新。
```

如果你在 Mac 上让模型执行 Windows 修复命令，轻则没用，重则把排障方向带偏。

一句话：

```text
Windows helper / sandbox 问题，看第127-129期。
Mac 当前线程缺少 node_repl js，先看本篇。
```

## 8. 常见错误做法

这个问题里，最常见的错误有五类。

第一类：只重装 Chrome。

```text
Chrome 本体正常。
扩展也正常。
只是 Codex 当前线程没拿到 node_repl js。
```

这时候重装 Chrome 只是绕远路。

第二类：把 Control Chrome 换成 Playwright。

Playwright 能打开网页，但它不等于用户自己的 Chrome。

如果你的任务依赖：

```text
用户登录态
用户 Chrome Profile
用户已打开标签页
Chrome 扩展桥
```

那换成 Playwright 可能会让任务跑到另一个浏览器环境里。

第三类：调用 `js_reset` 当修复。

`js_reset` 可以清空 Node REPL 会话。

但如果当前线程根本没有 `js` 工具，reset 不能变出一个 `js`。

第四类：让模型安装一堆 Node 包。

Control Chrome 不是缺 npm 包。

它需要的是插件目录里自带的 `browser-client.mjs`，以及能执行它的 Node REPL js 工具。

第五类：把工具名写死。

如果你只写：

```text
调用 mcp__node_repl__js。
```

模型可能回答：

```text
当前没有这个工具。
```

更稳的写法是：

```text
搜索 node_repl js。
```

因为这样模型可以通过工具发现把入口找出来。

## 9. 企业团队怎么做成排障规范

如果团队里多人使用 Codex、Claude Code、Chrome 插件和企业级大模型接入，建议把这类问题做成一张内部排障卡。

卡片可以这样写：

```text
问题名称：
Codex Chrome 插件控制失败，提示当前线程没有 mcp__node_repl__js。

适用场景：
需要控制用户自己的 Chrome，而不是普通网页自动化。

第一动作：
让模型通过工具发现搜索 node_repl js。

关键判断：
mcp__node_repl__js 是 Node REPL 的 js 执行工具。
mcp__node_repl.js 可能是同一工具的界面显示写法。

不要先做：
不要重装 Chrome。
不要切到 Playwright。
不要调用 js_reset 代替 js。
不要执行 Windows 专用修复。

升级处理：
如果找不到 js 工具，新开线程、重启 Codex、检查插件启用、确认插件目录下是否有 scripts/browser-client.mjs。
```

企业环境里，建议再补两条：

```text
不要让模型读取 Chrome Cookie、密码、Local Storage 或会话文件。
不要把浏览器登录态、私密页面截图、API Key 直接贴给第三方模型。
```

Chrome 控制能力很方便，但它控制的是用户真实浏览器。

这和普通测试浏览器不一样。

要把权限边界讲清楚。

## 10. 和 4SAPI / 企业 API 网关有什么关系

有人会问：

```text
这不是 Codex 插件问题吗，为什么放在大模型API中转站系列里？
```

原因很简单。

现在企业接入大模型，已经不是“一个 Key 调一个模型”这么简单。

真实工作流经常是：

```text
Codex / Claude Code / Cursor 负责执行工程任务。
4SAPI 或企业 API 网关负责统一模型入口。
Chrome 插件负责接管用户真实浏览器现场。
MCP 工具负责连接本地能力和外部系统。
日志审计负责记录每次模型调用与工具调用。
```

任何一层错位，都会变成“AI 不能干活”。

但排障时不能只盯模型输出。

要分清：

```text
模型能力问题。
API 网关问题。
MCP 工具暴露问题。
插件 skill 更新问题。
浏览器扩展连接问题。
操作系统权限问题。
当前线程上下文问题。
```

用 4SAPI 这类企业级 API 入口时，也可以把这类排障纳入日志体系：

```text
模型当时使用的是哪个供应商。
上下文里是否出现 Control Chrome skill。
工具列表里是否有 node_repl js。
模型是否尝试了错误工具。
是否发生了不必要的重试和成本浪费。
最终是提示词修复，还是重启 / 新线程修复。
```

这样下次同事再遇到，不需要从零问。

直接查排障卡和历史日志即可。

## 11. 最小验证清单

最后给一个最小验证清单。

当你按本文修复后，让模型做下面这些低风险动作：

```text
1. 确认它找到的是 Node REPL 的 js 工具。
2. 确认它没有把 js_reset 当成 js。
3. 确认它使用的是 Control Chrome，而不是 in-app browser。
4. 确认它能读取 Chrome 插件文档或初始化 browser-client.mjs。
5. 确认它能列出或接管一个用户允许的 Chrome 标签页。
6. 完成任务后释放或关闭不需要保留的标签页。
```

如果只是为了验证连接，不要让模型直接操作敏感站点。

可以先让它打开一个普通页面，例如：

```text
https://example.com
```

或者让它读取当前空白页状态。

确认控制桥通了，再去做真正任务。

## 12. 一句话总结

Mac 端 Codex Chrome 插件失效，如果报错是：

```text
当前线程没有 mcp__node_repl__js
```

优先不要怀疑 Chrome。

先告诉模型：

```text
去工具发现里找 node_repl js。
```

因为 Chrome 插件控制链路的关键入口，不是 Playwright，不是 in-app browser，也不是 `js_reset`。

而是：

```text
Node REPL 的 js 执行工具
```

也就是通常写作：

```text
mcp__node_repl__js
```

或者在某些界面里显示成：

```text
mcp__node_repl.js
```

这类问题看似小，其实很适合做成团队内部排障模板。

因为真正节省时间的，不是记住某个神秘工具名。

而是形成稳定判断：

```text
先分层。
先找入口。
先验证工具。
再动缓存和安装。
```

这就是 Codex、Chrome 插件、MCP 工具和企业级大模型接入一起使用时，最值得养成的排障习惯。
