---
title: "【大模型API中转站】第128期 Codex缓存重建 | AppData修复法"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Claude
  - Windows
  - AppData
  - Browser
  - 企业级API
  - 4SAPI
description: "围绕 Codex 桌面版 Windows in-app browser 自动化失效，讲清 .codex 与 %LOCALAPPDATA%\\OpenAI 的区别，给出备份、重命名、重启、验证和回滚流程，并演示如何用 4SAPI 接 Claude 生成排障报告。"
---

# 【大模型API中转站】第128期 Codex缓存重建 | AppData修复法

本文是【大模型API中转站】系列的第128篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一期讲了 Codex 浏览器自动化失效时，为什么不能只盯着网页本身。

这一期讲一个更具体的修复动作：

```text
关闭 Codex，
备份并重建 %LOCALAPPDATA%\OpenAI，
让 Codex 桌面版重新生成本地运行缓存。
```

这个办法来自 openai/codex issue #23222 里的真实反馈。

有用户遇到：

```text
Codex in-app browser 可以手动用，
但 Codex 无法自动点击、输入、导航。
```

重装 Codex、重建 `.codex` 不一定解决。

但把 Windows 本地应用数据目录改名，让 Codex 重新生成，问题恢复了。

这篇不是鼓励你乱删目录。

相反，重点是：

```text
什么时候该动 LocalAppData。
动之前怎么备份。
动之后怎么验证。
失败后怎么回滚。
企业环境怎么记录证据。
```

## 1. 先搞清楚两个目录

Windows 上排查 Codex，最容易混淆两个位置。

第一个是：

```text
C:\Users\<你>\.codex
```

第二个是：

```text
C:\Users\<你>\AppData\Local\OpenAI
```

也就是：

```text
%USERPROFILE%\.codex
%LOCALAPPDATA%\OpenAI
```

它们都和 Codex 有关，但角色不同。

你可以粗略理解为：

```text
.codex：用户侧配置和扩展状态。
LocalAppData\OpenAI：桌面应用本地运行状态。
```

`.codex` 常见内容包括：

```text
config.toml
插件缓存
技能缓存
MCP 配置
某些 thread 或工具配置
```

`%LOCALAPPDATA%\OpenAI` 更接近：

```text
桌面 App 缓存
本地 helper 释放目录
Codex bin
运行时中间文件
Windows App 本地状态
```

如果你是模型调用失败、Key 配置错误、MCP 配置错误，优先看 `.codex`。

如果你是 helper path、native pipe、bundled executable relocation 这类问题，`%LOCALAPPDATA%\OpenAI` 的优先级更高。

## 2. 为什么重装不一定有效

很多 Windows 用户的直觉是：

```text
坏了就重装。
```

但桌面应用经常不是这么简单。

重装应用时，系统可能会保留一部分用户本地数据。

这对正常升级是好事。

因为你不希望每次升级都丢配置、丢缓存、丢登录状态。

但如果坏掉的正是本地缓存，重装就可能出现：

```text
应用文件是新的。
坏缓存还是旧的。
错误继续复现。
```

issue #23222 里就有类似现象：

```text
重新安装 Codex 后，浏览器自动化仍然失败。
```

这就是为什么需要区分：

```text
安装包问题
用户配置问题
本地运行态问题
权限策略问题
```

不要把所有问题都丢给“重装”。

## 3. 什么时候适合重建 LocalAppData

如果你看到下面这些信号，可以考虑重建 `%LOCALAPPDATA%\OpenAI`。

第一类：

```text
Codex 浏览器能手动打开，但 Agent 自动化不可用。
```

第二类：

```text
Computer Use helper path 不可用。
```

第三类：

```text
bundled executable relocation 失败。
```

第四类：

```text
node.exe、node_repl.exe、rg.exe、codex.exe 这类内置工具释放失败。
```

第五类：

```text
重装 Codex 无效，但错误仍指向本地 helper 或 bin。
```

这时重建 LocalAppData 是一个相对低风险的动作。

因为不是直接删除，而是改名备份。

你随时可以回滚。

## 4. 动手前先做三件事

第一，关掉 Codex。

不要在 Codex 运行时改它的本地目录。

第二，记录版本和环境。

建议先运行：

```powershell
Get-ComputerInfo | Select-Object WindowsProductName, WindowsVersion, OsBuildNumber
```

再记录 Codex 版本。

如果你能打开 Codex，可以从 About Codex 里看。

第三，导出关键目录状态。

```powershell
$stamp = Get-Date -Format "yyyyMMddHHmmss"
$out = "$env:USERPROFILE\Desktop\codex-debug-$stamp.txt"

"== Codex local state ==" | Out-File $out -Encoding UTF8
"USERPROFILE=$env:USERPROFILE" | Out-File $out -Append -Encoding UTF8
"LOCALAPPDATA=$env:LOCALAPPDATA" | Out-File $out -Append -Encoding UTF8

"`n== .codex ==" | Out-File $out -Append -Encoding UTF8
Get-ChildItem -Force "$env:USERPROFILE\.codex" -ErrorAction SilentlyContinue |
    Select-Object Name, Length, LastWriteTime |
    Format-Table -AutoSize |
    Out-String |
    Out-File $out -Append -Encoding UTF8

"`n== Local OpenAI ==" | Out-File $out -Append -Encoding UTF8
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI" -ErrorAction SilentlyContinue |
    Select-Object Name, Length, LastWriteTime |
    Format-Table -AutoSize |
    Out-String |
    Out-File $out -Append -Encoding UTF8

Write-Host "Saved debug note to $out"
```

这一步很适合企业团队。

以后报障时，不要只说“浏览器坏了”。

把版本、目录、日志、修复动作一起留下。

## 5. 标准修复命令

确认 Codex 已关闭后，执行：

```powershell
Stop-Process -Name Codex -ErrorAction SilentlyContinue

$src = Join-Path $env:LOCALAPPDATA "OpenAI"
$stamp = Get-Date -Format "yyyyMMddHHmmss"
$backupName = "OpenAI.old-$stamp"

if (Test-Path $src) {
    Rename-Item -LiteralPath $src -NewName $backupName
    Write-Host "Renamed $src to $backupName"
} else {
    Write-Host "No existing OpenAI local app data folder found."
}
```

这条命令做的事很简单：

```text
不删除。
只改名。
让 Codex 下次启动时重新创建 OpenAI 本地目录。
```

如果你更谨慎，可以先复制一份：

```powershell
$src = Join-Path $env:LOCALAPPDATA "OpenAI"
$copy = Join-Path $env:LOCALAPPDATA ("OpenAI.copy-" + (Get-Date -Format "yyyyMMddHHmmss"))

if (Test-Path $src) {
    Copy-Item -LiteralPath $src -Destination $copy -Recurse -Force
}
```

复制完成后再改名。

企业环境建议用复制加改名。

个人环境一般改名就够。

## 6. 重启后怎么验证

重新打开 Codex 后，不要直接让它做复杂任务。

先做四个最小验证。

第一：

```text
请检查当前是否有 browser / computer-use / node_repl 相关工具可用。
只报告可用性，不要操作网页。
```

第二：

```text
请打开 in-app browser，并访问 about:blank。
```

第三：

```text
请导航到 https://openai.com，只报告当前 URL 和页面标题。
```

第四：

```text
请关闭刚刚打开的浏览器 tab。
```

如果这四步都能成功，说明浏览器控制桥基本恢复。

再去测试：

```text
搜索关键词。
点击按钮。
填写表单。
读取页面文本。
截图验证。
```

排错时要从小动作开始。

不要一上来就让它：

```text
打开某网站，登录，搜索，点击第三个结果，截图，再写报告。
```

动作越多，失败点越多。

你很难判断到底是哪一步坏了。

## 7. 如果新目录生成失败

如果重启 Codex 后，`%LOCALAPPDATA%\OpenAI` 没有重新生成，或者 `Codex\bin` 仍然缺失，要继续查权限。

先看目录：

```powershell
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI" -ErrorAction SilentlyContinue
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI\Codex" -ErrorAction SilentlyContinue
Get-ChildItem -Force "$env:LOCALAPPDATA\OpenAI\Codex\bin" -ErrorAction SilentlyContinue
```

再测试当前用户是否能创建目录：

```powershell
$test = Join-Path $env:LOCALAPPDATA "OpenAI-write-test"
New-Item -ItemType Directory -Path $test -Force
Remove-Item -LiteralPath $test -Force
```

如果这里都失败，问题就不是 Codex 独有。

可能是：

```text
用户目录权限异常。
企业安全策略限制 AppData 写入。
安全软件拦截可执行文件释放。
WindowsApps 包访问限制。
磁盘或路径状态异常。
```

这时不要继续删目录。

应该把错误交给 IT 或安全团队排查。

## 8. 如果修复无效，怎么回滚

回滚很简单。

先关闭 Codex。

然后删除新生成的 OpenAI 目录，或者改名保留：

```powershell
Stop-Process -Name Codex -ErrorAction SilentlyContinue

$current = Join-Path $env:LOCALAPPDATA "OpenAI"
$failed = Join-Path $env:LOCALAPPDATA ("OpenAI.failed-" + (Get-Date -Format "yyyyMMddHHmmss"))

if (Test-Path $current) {
    Rename-Item -LiteralPath $current -NewName (Split-Path $failed -Leaf)
}
```

再把原备份改回来：

```powershell
Get-ChildItem -Path $env:LOCALAPPDATA -Directory |
    Where-Object { $_.Name -like "OpenAI.old-*" } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1
```

找到你要恢复的目录后：

```powershell
Rename-Item -LiteralPath "$env:LOCALAPPDATA\OpenAI.old-20260629123000" -NewName "OpenAI"
```

把时间戳换成你的实际备份名。

这就是为什么我一直说：

```text
改名，不要直接删除。
```

排错动作必须可回滚。

## 9. 用 4SAPI 让 Claude 生成排障报告

重建缓存后，建议让 AI 帮你写一份排障报告。

企业里这一步很有用。

它可以把零散动作整理成：

```text
问题现象
影响范围
已观察证据
执行动作
验证结果
风险判断
后续建议
```

可以用 4SAPI 接 Claude 来做。

示例脚本：

```python
import os
from pathlib import Path
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["SAPI_API_KEY"],
    base_url=os.getenv("SAPI_BASE_URL", "https://4sapi.com/v1"),
)

debug_note = Path(os.environ["CODEX_DEBUG_NOTE"]).read_text(encoding="utf-8")

prompt = f"""
请把下面的 Windows Codex 排障记录整理成企业内部故障报告。

要求：
1. 不输出完整 API Key。
2. 不夸大结论。
3. 区分已验证事实和推测。
4. 给出下一步低风险检查。
5. 用中文输出。

排障记录：
{debug_note}
"""

resp = client.chat.completions.create(
    model=os.getenv("SAPI_MODEL", "claude-sonnet-4-5-20250929"),
    temperature=0.1,
    messages=[
        {"role": "system", "content": "你是企业桌面应用排障报告助手。"},
        {"role": "user", "content": prompt},
    ],
)

print(resp.choices[0].message.content)
```

运行时：

```powershell
$env:SAPI_API_KEY="sk-你的4SAPI令牌"
$env:SAPI_BASE_URL="https://4sapi.com/v1"
$env:SAPI_MODEL="你的模型名称"
$env:CODEX_DEBUG_NOTE="$env:USERPROFILE\Desktop\codex-debug-20260629123000.txt"
python .\codex_report.py
```

注意：

```text
不要把完整 Key 写进 debug_note。
不要把企业内部敏感路径、用户名、项目名原样发出去。
需要脱敏后再让模型分析。
```

4SAPI 在这里的作用不是“绕过问题”。

它是统一模型入口：

```text
Claude 负责分析。
Codex 负责执行和验证。
4SAPI 负责 Key、日志、预算和调用记录。
```

## 10. 为什么企业团队要分 Key

很多团队图省事，会把一个 Key 到处填。

这很危险。

排障 Agent 和生产业务不应该共用 Key。

建议至少拆成三类：

```text
业务生产 Key：只给线上服务。
研发助手 Key：给 Codex、Claude Code、Cursor 等工程工具。
排障分析 Key：只用于日志分析、故障报告、知识库问答。
```

这样做的好处是：

```text
出了问题能查是谁调用。
排障成本能单独统计。
某个 Key 泄露可以单独停用。
不同用途可以设置不同预算。
敏感任务可以限制模型和权限。
```

企业级大模型接入，不是把 API Key 填进工具就结束。

真正的落地是：

```text
每类工具有自己的 Key。
每个 Key 有自己的额度。
每次调用有日志。
每次异常能追踪。
```

## 11. 常见误区

误区一：

```text
浏览器能打开，所以 browser automation 一定可用。
```

不对。

手动浏览和 Agent 控制不是一条链路。

误区二：

```text
重装 Codex 一定会清理所有坏状态。
```

不一定。

本地应用数据可能被保留。

误区三：

```text
改 .codex 就等于重置 Codex。
```

不完整。

`.codex` 和 `%LOCALAPPDATA%\OpenAI` 是两类状态。

误区四：

```text
AI 给了一个修复命令，直接跑。
```

不建议。

先让 AI 输出证据和回滚方案。

误区五：

```text
排障日志随便丢给模型。
```

更不建议。

日志里可能有用户名、路径、项目名、Key 片段、内部域名。

先脱敏，再分析。

## 12. 一份企业内部 SOP

可以把下面这份 SOP 放进团队文档。

```text
标题：Codex Windows 浏览器自动化失效排查

触发条件：
1. in-app browser 可手动使用。
2. Codex 无法自动点击、输入或导航。
3. 日志出现 browser-client、native pipe、helper path 或 relocation 相关错误。

一级检查：
1. 记录 Codex 版本和 Windows 版本。
2. 确认当前 thread 是否有 browser / computer-use / node_repl 工具。
3. 尝试 about:blank 最小导航。

二级检查：
1. 检查 %LOCALAPPDATA%\OpenAI\Codex\bin。
2. 检查 .codex\.sandbox 日志。
3. 检查 config.toml 的 windows sandbox。

低风险修复：
1. 关闭 Codex。
2. 备份并重命名 %LOCALAPPDATA%\OpenAI。
3. 重启 Codex。
4. 验证 about:blank、openai.com、tab title。

回滚：
1. 关闭 Codex。
2. 改名新生成目录。
3. 恢复旧 OpenAI 目录。

上报材料：
1. 版本。
2. 错误摘要。
3. 目录状态。
4. 执行动作。
5. 验证结果。
```

这套 SOP 的关键是：

```text
先观察，再备份，再修复，再验证。
```

不要边猜边删。

## 13. 本篇结论

Codex 桌面版在 Windows 上出现浏览器自动化失效时，重建 `%LOCALAPPDATA%\OpenAI` 是一个值得尝试的低风险修复。

但它适用的前提是：

```text
错误指向本地 helper、native pipe、bundled executable 或 bin 释放失败。
```

它不适合用来解决所有问题。

如果你的错误是 Key、模型、API base_url、网络请求失败，那应该回到 API 排查。

如果你的错误是 sandbox 权限提升，例如出现 `os error 740`，那就要继续看下一层：

```text
config.toml 里的 windows sandbox 设置。
```

下一篇继续讲：

```text
sandbox = "elevated" 为什么可能让 Browser helper 起不来，
以及什么时候应该改成 sandbox = "unelevated"。
```

## 资料来源与延伸阅读

- GitHub issue：openai/codex #23222：https://github.com/openai/codex/issues/23222
- OpenAI Codex 文档：https://developers.openai.com/codex/
- OpenAI Codex In-app browser：https://developers.openai.com/codex/app/browser
- OpenAI Codex Agent Approvals and Security：https://developers.openai.com/codex/agent-approvals-security
- OpenAI Codex Windows：https://developers.openai.com/codex/windows
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
