---
title: "【大模型API中转站】第129期 Codex沙箱修复 | os error 740"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Windows
  - sandbox
  - Browser
  - Computer Use
  - 企业级大模型接入
  - 4SAPI
description: "复盘 Codex Windows 桌面版浏览器控制失败的 sandbox 分支：当 .codex\\.sandbox 日志出现 os error 740，且 config.toml 设置 sandbox = elevated 时，为什么应优先尝试改为 unelevated 并重启验证。文章附 4SAPI 接 Claude 生成排障树和企业治理清单。"
---

# 【大模型API中转站】第129期 Codex沙箱修复 | os error 740

本文是【大模型API中转站】系列的第129篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前两篇讲了 Codex 浏览器自动化失效的两条线：

```text
第127期：先判断 browser-client is not trusted 属于哪一层。
第128期：备份并重建 %LOCALAPPDATA%\OpenAI。
```

这一篇讲第三条线，也是 Windows 上很容易被忽略的一条：

```text
Codex 的 Windows sandbox 设置。
```

在 openai/codex issue #23222 的后续评论里，有用户让 Codex 自己分析日志，最后定位到：

```text
config.toml 里设置了 sandbox = "elevated"，
但浏览器 / node helper 启动时没有拿到匹配的权限，
于是 Windows sandbox setup 失败。
```

日志里出现一个非常关键的错误：

```text
os error 740
```

在 Windows 语境下，这通常和“请求的操作需要提升权限”有关。

最后有效的修复不是反复重装，也不是删除插件缓存，而是把：

```toml
sandbox = "elevated"
```

改成：

```toml
sandbox = "unelevated"
```

然后完全关闭并重启 Codex。

这篇就把这个分支讲透。

## 1. 为什么 sandbox 会影响浏览器控制

Codex 不是一个只会聊天的网页。

它要在本地帮你做工程任务，往往要调用：

```text
Shell
Git
文件读写
Node helper
Browser helper
Computer Use helper
MCP 工具
插件工具
```

这些动作都涉及权限边界。

所以 Codex 需要 sandbox。

sandbox 的作用可以理解为：

```text
限制 Agent 能在哪里运行命令。
限制命令能读写哪些文件。
限制高风险动作是否需要批准。
限制某些 helper 的启动方式。
```

问题在于，Windows 的权限模型比较敏感。

如果某个 helper 需要在普通权限下启动，但配置要求 elevated sandbox，或者相反，就可能出现：

```text
主应用在一个权限层。
helper 在另一个权限层。
native pipe 无法建立。
browser-client 无法成为可信客户端。
```

这时表面现象就是：

```text
浏览器能开，但 Codex 控制不了。
```

## 2. 什么时候怀疑 sandbox

不是所有浏览器自动化失败都和 sandbox 有关。

只有出现下面这些信号时，才优先看这一层。

第一，`.codex\.sandbox` 下有日志。

可以检查：

```powershell
Get-ChildItem -Force "$env:USERPROFILE\.codex\.sandbox" -ErrorAction SilentlyContinue
```

第二，日志里有 Windows sandbox setup 失败。

可以搜索：

```powershell
Select-String -Path "$env:USERPROFILE\.codex\.sandbox\*.log" `
    -Pattern "sandbox failed","os error 740","requires elevation" `
    -ErrorAction SilentlyContinue
```

第三，`config.toml` 里有 Windows sandbox 配置。

```powershell
Get-Content "$env:USERPROFILE\.codex\config.toml" -ErrorAction SilentlyContinue
```

重点看：

```toml
[windows]
sandbox = "elevated"
```

第四，重建 `%LOCALAPPDATA%\OpenAI` 无效。

如果缓存重建后 helper 仍然起不来，那问题可能不在文件释放，而在启动权限。

第五，普通网页、Chrome、网络都没明显问题。

这时继续换网页没有意义。

## 3. 用 4SAPI 接 Claude 做 sandbox 分支判断

我建议把日志交给模型前，先脱敏。

不要把完整用户名、内部项目路径、Key、公司域名原样丢进去。

可以整理成这样：

```text
系统：Windows x64
现象：Codex in-app browser 手动可用，自动化不可用
已做：重建 LocalAppData 后仍失败
config.toml：
[windows]
sandbox = "elevated"

sandbox 日志摘要：
windows sandbox failed during setup
os error 740

问题：
请判断是否可能是 elevated sandbox 导致 browser/helper 启动权限不匹配。
请给出低风险验证步骤、修复步骤和回滚步骤。
```

用 4SAPI 的 OpenAI 兼容接口可以这样问：

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["SAPI_API_KEY"],
    base_url=os.getenv("SAPI_BASE_URL", "https://4sapi.com/v1"),
)

prompt = """
Windows Codex desktop browser automation fails.

Facts:
- In-app browser works manually.
- Browser automation cannot click/type/navigate.
- Rebuilding LocalAppData did not fix it.
- config.toml has [windows] sandbox = "elevated".
- Sandbox log includes os error 740.

Task:
Build a diagnostic tree.
Tell me whether changing sandbox to unelevated is a reasonable next step.
Include backup, edit, restart, verification, rollback.
"""

resp = client.chat.completions.create(
    model=os.getenv("SAPI_MODEL", "claude-sonnet-4-5-20250929"),
    temperature=0.1,
    messages=[
        {
            "role": "system",
            "content": (
                "你是 Codex Windows sandbox 排障助手。"
                "只基于给定证据分析，区分事实、推测和操作建议。"
            ),
        },
        {"role": "user", "content": prompt},
    ],
)

print(resp.choices[0].message.content)
```

理想输出应该告诉你：

```text
这是一个合理怀疑。
先备份 config.toml。
把 elevated 改成 unelevated。
完全关闭并重启 Codex。
用最小 browser 动作验证。
如失败则恢复备份。
```

如果模型直接说“删除所有缓存”，说明它没有抓住 sandbox 证据。

你可以继续追问：

```text
请只讨论 os error 740 和 sandbox = elevated 之间的关系。
不要讨论缓存重建。
```

## 4. 修改前先备份 config.toml

Codex 配置通常在：

```text
C:\Users\<你>\.codex\config.toml
```

也就是：

```powershell
$env:USERPROFILE\.codex\config.toml
```

修改前先备份：

```powershell
$config = Join-Path $env:USERPROFILE ".codex\config.toml"
$backup = Join-Path $env:USERPROFILE (".codex\config.toml.bak-" + (Get-Date -Format "yyyyMMddHHmmss"))

Copy-Item -LiteralPath $config -Destination $backup -Force
Write-Host "Backup saved to $backup"
```

如果文件不存在，不要新建乱写。

先确认你是否真的有这个配置。

```powershell
Test-Path "$env:USERPROFILE\.codex\config.toml"
```

## 5. 把 elevated 改成 unelevated

用记事本打开：

```powershell
notepad "$env:USERPROFILE\.codex\config.toml"
```

找到：

```toml
[windows]
sandbox = "elevated"
```

改成：

```toml
[windows]
sandbox = "unelevated"
```

保存。

然后完全关闭 Codex。

最好确认进程也退出：

```powershell
Stop-Process -Name Codex -ErrorAction SilentlyContinue
```

再重新打开 Codex。

注意：

```text
只改配置不重启，通常不够。
```

这类配置往往在启动时读取。

## 6. 为什么“管理员启动”不一定解决

很多人看到权限错误，就会想：

```text
那我用管理员身份运行 Codex。
```

这个方向不一定错，但 issue 里的反馈很有启发：

```text
仅管理员运行没有真正解决。
改成 sandbox = "unelevated" 后才恢复。
```

原因是：

```text
问题不只是权限高低，
而是主应用、helper、sandbox 策略之间是否匹配。
```

如果配置要求 elevated，而 helper 启动路径不符合这个策略，它依然可能失败。

所以不要把它简化成：

```text
权限不够 -> 管理员启动。
```

更准确的是：

```text
sandbox 策略和 helper 启动方式不匹配。
```

这就是为什么改为 unelevated 可能更合理。

## 7. 改完后怎么验证

还是从最小动作开始。

第一，问工具可用性：

```text
请检查当前 thread 是否能使用 browser / computer-use / node_repl 相关工具。
不要访问外部网站，只报告工具可用性。
```

第二，打开空白页：

```text
请用 in-app browser 打开 about:blank，并报告当前 URL。
```

第三，访问稳定网页：

```text
请打开 https://openai.com，并报告页面标题。
```

第四，做简单交互：

```text
请在浏览器中打开一个搜索页，输入 test，然后不要点击任何广告，只报告搜索结果页是否加载。
```

如果这几步成功，说明 browser helper 这条链路基本恢复。

再去跑你真实任务。

## 8. 如果失败，怎么回滚

关闭 Codex：

```powershell
Stop-Process -Name Codex -ErrorAction SilentlyContinue
```

找到刚才的备份：

```powershell
Get-ChildItem "$env:USERPROFILE\.codex" -Filter "config.toml.bak-*" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 5
```

恢复：

```powershell
$config = Join-Path $env:USERPROFILE ".codex\config.toml"
$backup = Join-Path $env:USERPROFILE ".codex\config.toml.bak-20260629123000"

Copy-Item -LiteralPath $backup -Destination $config -Force
```

把时间戳换成你的实际文件名。

然后重启 Codex。

回滚后继续查：

```text
LocalAppData 是否生成完整。
Browser plugin 是否启用。
当前 thread 是否暴露工具。
安全软件是否拦截 helper。
Windows 用户权限是否异常。
```

## 9. config.toml 不要随便改什么

这篇只建议改一个字段：

```toml
sandbox = "unelevated"
```

不要顺手大改：

```text
模型配置
MCP 配置
插件配置
hooks
approval
网络代理
```

排错最怕同时改五个变量。

最后好了，你也不知道是哪一个修好的。

最后坏了，更不知道是哪一个弄坏的。

正确姿势是：

```text
一次只改一个变量。
每次改前备份。
每次改后最小验证。
每次验证后记录结果。
```

这也是企业排障 SOP 的核心。

## 10. sandbox 和安全边界怎么理解

有人会问：

```text
把 elevated 改成 unelevated，会不会不安全？
```

这个问题要具体看 Codex 的实际权限策略和你的环境。

从排障角度看，`unelevated` 并不等于“无安全边界”。

它更像是让 Windows helper 在普通权限层运行，避免因为提升权限策略导致启动失败。

真正的安全治理不只靠这一个字段。

还要看：

```text
workspace 范围
approval 策略
命令权限
网络权限
AGENTS.md 规则
hooks 检查
MCP 授权
API Key 权限
日志审计
```

企业环境里，建议把 Codex 安全拆成两层：

```text
本地执行安全：sandbox、approval、workspace、hooks。
模型调用安全：4SAPI Key、额度、日志、模型白名单、项目分组。
```

不要只盯着一个 sandbox 字段。

## 11. 企业团队推荐配置思路

如果团队要推广 Codex 桌面版，建议先做试点。

不要一上来全员统一推复杂配置。

推荐流程：

```text
第一批：3-5 个研发试点。
第二步：统一记录 Codex App 版本和 Windows 版本。
第三步：统一 4SAPI 研发助手 Key。
第四步：统一 AGENTS.md 基础规则。
第五步：保留默认或稳定 sandbox 策略。
第六步：遇到问题再按日志分支调整。
```

4SAPI 这边建议分组：

```text
codex-dev：日常研发助手。
codex-debug：日志分析和排障报告。
codex-review：代码审查和 PR 总结。
codex-automation：定时自动化任务。
```

每组单独设置：

```text
预算。
额度。
模型范围。
调用日志。
负责人。
告警阈值。
```

这样即使某一类 Agent 消耗异常，也不会影响生产服务。

## 12. 一份完整排查树

下面这份可以直接收藏。

```text
问题：Codex in-app browser 手动可用，自动化不可用。

第一层：页面层
- 换 about:blank 是否可控？
- 换 openai.com 是否可导航？
- 如果只有某网站失败，再查 DOM、登录态、网络。

第二层：工具层
- 当前 thread 是否暴露 browser 工具？
- 是否暴露 computer-use？
- 是否暴露 node_repl？
- 如果工具不存在，查插件和工具注入。

第三层：本地缓存层
- %LOCALAPPDATA%\OpenAI\Codex\bin 是否存在？
- 是否有 bundled executable relocation 失败？
- 可尝试备份并重建 %LOCALAPPDATA%\OpenAI。

第四层：sandbox层
- .codex\.sandbox 是否有日志？
- 是否出现 os error 740？
- config.toml 是否为 sandbox = "elevated"？
- 可备份后改为 sandbox = "unelevated" 并重启。

第五层：企业安全层
- 安全软件是否拦截 node/helper？
- 是否限制 AppData 写入？
- 是否限制 WindowsApps 包路径？
- 是否有组策略影响本地 pipe？

第六层：上报
- 附版本、日志摘要、配置片段、已尝试动作。
```

这个排查树的好处是：

```text
不会一上来就删缓存。
不会一上来就重装系统。
不会把网页问题和本地 helper 问题混在一起。
```

## 13. 一份给 Codex 的自查提示词

如果你当前 Codex 还能跑命令，可以直接问它：

```text
请帮我排查当前 Windows Codex 桌面版浏览器自动化失效问题。

限制：
1. 不要修改文件。
2. 不要删除目录。
3. 只读检查。

请检查：
1. 当前工作环境是否暴露 browser / computer-use / node_repl 工具。
2. %LOCALAPPDATA%\OpenAI\Codex\bin 是否存在。
3. %USERPROFILE%\.codex\config.toml 是否包含 [windows] sandbox 配置。
4. %USERPROFILE%\.codex\.sandbox 是否有 os error 740。

输出：
按 页面层、工具层、缓存层、sandbox层、企业安全层 给结论。
每个结论必须说明证据。
```

如果 Codex 无法访问浏览器工具，但还能跑 shell，这条提示词仍然有用。

它可以先帮你定位本机状态。

如果 Codex 连 shell 都不可用，就手动执行本文里的 PowerShell 命令。

## 14. 本篇结论

`os error 740` 是一个很强的方向信号。

当你同时看到：

```text
Codex 浏览器手动可用。
自动化不可用。
LocalAppData 重建无效。
.codex\.sandbox 日志有权限提升错误。
config.toml 里是 sandbox = "elevated"。
```

就可以优先尝试：

```text
备份 config.toml。
把 sandbox 改成 unelevated。
完全关闭并重启 Codex。
用 about:blank 和 openai.com 做最小验证。
```

这不是万能修复。

但它是一个有证据支持、风险可控、可回滚的修复分支。

企业团队要记住：

```text
Codex 本地执行靠 sandbox 和工具链治理。
模型调用靠 4SAPI 做 Key、日志和预算治理。
两边都要可审计。
```

一句话总结：

```text
Agent 越能动电脑，越要把权限、日志和回滚做在前面。
```

## 资料来源与延伸阅读

- GitHub issue：openai/codex #23222：https://github.com/openai/codex/issues/23222
- OpenAI Codex 文档：https://developers.openai.com/codex/
- OpenAI Codex In-app browser：https://developers.openai.com/codex/app/browser
- OpenAI Codex Agent Approvals and Security：https://developers.openai.com/codex/agent-approvals-security
- OpenAI Codex Windows：https://developers.openai.com/codex/windows
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
