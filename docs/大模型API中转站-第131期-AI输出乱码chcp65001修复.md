---
title: "【大模型API中转站】第131期 AI输出乱码修复 | chcp 65001实战"
category: 人工智能
tags:
  - 大模型API中转站
  - AI输出乱码
  - chcp 65001
  - UTF-8
  - Windows
  - Codex
  - Claude Code
  - 企业级大模型接入
  - 4SAPI
description: "Windows 上使用 Codex、Claude Code、OpenCode 等 AI CLI 时，中文输出经常变成问号、锟斤拷或 æµ‹è¯•。本文以 chcp 65001 为主线，讲清终端显示乱码和文件内容损坏的区别、快速修复命令、PowerShell 编码补强、字体排查和企业级 API 接入时的日志审计清单。"
---

# 【大模型API中转站】第131期 AI输出乱码修复 | chcp 65001实战

本文是【大模型API中转站】系列的第131篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这一篇讲一个很小，但很要命的问题：

```text
Windows 上用 AI CLI 时，中文输出乱码。
```

比如你让 Codex、Claude Code、OpenCode 或其他命令行 AI 工具读取项目，终端里突然出现：

```text
???
锟斤拷
æµ‹è¯•
ä¸­æ–‡
```

如果只是看起来难受，还能忍。

真正麻烦的是，AI 可能会把终端里看到的乱码当成“源码坏了”，然后认真地帮你把正常中文改成乱码。

这就从显示问题变成了文件损坏。

所以这篇先给结论：

```powershell
chcp 65001
```

在 Windows 中文环境里，遇到 AI 输出乱码，先在当前终端运行这条命令，把控制台代码页切到 UTF-8。

很多场景下，问题会立刻消失。

## 1. 先判断：显示乱码，还是文件真坏了

不要一看到乱码就让 AI 改文件。

先分清楚两种情况。

第一种是终端显示乱码：

```text
终端里乱码，
但 VS Code 打开源文件正常，
git diff 也没有出现乱码替换。
```

这种通常只是 Windows 控制台编码不匹配。

优先处理终端，不要动源码。

第二种是文件内容损坏：

```text
VS Code 打开文件也乱码，
git diff 里能看到正常中文被替换成乱码，
提交记录里出现了大量非预期中文变更。
```

这种就不是简单显示问题了。

要先停下来，回滚 AI 的错误修改，或者从 Git 里恢复文件，再排查编码链路。

我个人建议把这个原则写进项目的 AI 规则里：

```text
不要根据终端显示乱码直接修改源文件。
修改含中文文件前，先确认文件本身是否损坏。
```

这句话很土，但能救命。

## 2. 为什么 Windows 上 AI 输出会乱码

根因通常不是 AI 模型不会中文。

也不是 Codex、Claude Code 或 OpenCode 故意抽风。

更常见的原因是：

```text
AI 工具输出 UTF-8，
Windows 中文终端按 CP936 / GBK 去解释，
两边编码对不上，
中文就变成乱码。
```

在很多中文 Windows 环境里，传统控制台默认代码页是 936，也就是常说的 GBK / CP936 体系。

而现在的大模型工具、Node.js 工具、Python 工具、Git 工具，越来越默认使用 UTF-8。

于是链路变成这样：

```text
AI CLI 输出 UTF-8 文本
-> PowerShell / cmd 接收输出
-> 当前 code page 仍然是 936
-> UTF-8 字节按 GBK 解码
-> 终端显示乱码
-> AI 看到乱码后可能误判文件坏了
```

这个时候，先不要怀疑模型。

先看终端编码。

## 3. 30 秒快速修复：chcp 65001

打开你正在使用的 Windows 终端，先运行：

```powershell
chcp
```

如果看到类似：

```text
活动代码页: 936
```

或者英文环境下是：

```text
Active code page: 936
```

说明当前终端还在用中文默认代码页。

这时运行：

```powershell
chcp 65001
```

正常会看到：

```text
Active code page: 65001
```

或者：

```text
活动代码页: 65001
```

65001 代表 UTF-8 代码页。

切换完成后，再运行你的 AI CLI：

```powershell
codex
```

或者：

```powershell
claude
```

再或者：

```powershell
opencode
```

然后做一个中文烟雾测试：

```powershell
Write-Host "中文测试：AI 输出不应该乱码"
```

如果这行能正常显示，再让 AI 读取中文文件或生成中文说明。

很多时候，到这里就够了。

## 4. 我自己的常用启动方式

如果只是临时排查，我会这样做：

```powershell
chcp 65001
codex
```

如果是 Claude Code：

```powershell
chcp 65001
claude
```

如果是 OpenCode：

```powershell
chcp 65001
opencode
```

如果你经常在一个项目里反复开 AI CLI，也可以把它写成一行：

```powershell
chcp 65001; codex
```

或者：

```powershell
chcp 65001; claude
```

这不是最“工程化”的最终方案，但非常适合先救急。

因为它有三个好处：

```text
第一，不改系统全局配置。
第二，不需要先写脚本。
第三，出了问题容易回到原状态。
```

对于个人开发者、内容创作者、临时排障任务来说，这条命令是性价比最高的第一步。

## 5. PowerShell 进阶补强

有些场景下，只运行 `chcp 65001` 还不够。

尤其是 AI 工具通过 PowerShell 子进程、管道、脚本去调用命令时，可能还会受到 PowerShell 自身编码变量影响。

这时可以在当前 PowerShell 里补上：

```powershell
chcp 65001
[Console]::InputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
```

这几行分别处理：

```text
chcp 65001：切换控制台代码页到 UTF-8。
InputEncoding：让控制台输入按 UTF-8 处理。
OutputEncoding：让控制台输出按 UTF-8 处理。
$OutputEncoding：让 PowerShell 管道输出按 UTF-8 处理。
```

然后再启动 AI 工具：

```powershell
codex
```

如果你不想每次都手打，可以先放进一个临时启动脚本，比如：

```powershell
# start-codex-utf8.ps1
chcp 65001
[Console]::InputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
codex
```

以后进入项目目录后运行：

```powershell
powershell -ExecutionPolicy Bypass -File .\start-codex-utf8.ps1
```

如果你用的是 Claude Code，就把最后一行改成：

```powershell
claude
```

如果你用的是 OpenCode，就改成：

```powershell
opencode
```

## 6. 为什么我不建议新手一上来就写一键脚本

参考文章里给了一套很完整的方案：

```text
安装 PowerShell 7。
创建 codexu / claudeu / opencodeu wrapper。
配置 PowerShell profile。
配置 Git UTF-8。
写入 AGENTS.md / CLAUDE.md 编码规则。
必要时迁移到 WSL2。
```

这套方案适合长期使用 AI CLI 的开发者，也适合团队统一环境。

但如果你只是第一次遇到乱码，我不建议一上来就复制一大段脚本执行。

原因很简单：

```text
你还不知道问题到底在 code page、PowerShell、字体、Git，还是文件本身。
```

我的建议顺序是：

```text
第一步：chcp 看当前代码页。
第二步：chcp 65001 临时切到 UTF-8。
第三步：重新启动 AI CLI。
第四步：用中文测试命令验证。
第五步：确认文件没有被 AI 改坏。
第六步：如果每天都要用，再做 wrapper 或 profile 固化。
```

先把问题缩小，再决定要不要自动化。

这比一上来改很多地方更稳。

## 7. 字体问题也会伪装成乱码

有时你已经切到了 UTF-8，但终端仍然显示方框、空白、问号。

这不一定是编码问题。

也可能是字体不支持中文、日文、韩文或 emoji。

可以用这行测试：

```powershell
Write-Host "中文测试 😀 こんにちは 한국어"
```

如果中文正常，但 emoji 是方框，可能只是字体缺字。

如果中文、日文、韩文都成方框，优先换终端字体。

Windows Terminal 里可以选择：

```text
Cascadia Code
Microsoft YaHei Mono
Sarasa Mono SC
```

VS Code 集成终端可以检查：

```json
{
  "terminal.integrated.fontFamily": "Cascadia Code, Microsoft YaHei Mono, Sarasa Mono SC"
}
```

编码负责“字节怎么解释”。

字体负责“字符怎么画出来”。

两个环节都正常，终端才会正常显示。

## 8. Git diff 是最后一道保险

修复乱码后，我强烈建议跑一次：

```powershell
git diff
```

重点看两件事。

第一，AI 有没有把正常中文改成乱码：

```diff
- 中文测试
+ æµ‹è¯•
```

如果看到这种变化，不要提交。

先回滚这部分修改。

第二，AI 有没有因为终端乱码误改文档、注释、配置文件：

```text
.md
.txt
.json
.yaml
.toml
.ps1
.py
.ts
.tsx
```

这些文件里只要有中文，都值得多看一眼。

特别是 Markdown 教程、中文注释、提示词文件、AGENTS.md、CLAUDE.md、README.md。

AI 写坏代码逻辑已经够烦了。

AI 把文档编码写坏，更隐蔽。

## 9. Codex / Claude Code / OpenCode 的区别

这类乱码问题不只出现在某一个工具。

不同 AI CLI 的表现不一样，但底层链路相似。

Codex 常见场景：

```text
读取中文文件后，终端输出变成乱码。
运行 PowerShell 命令时，命令结果乱码。
AI 误判中文文件内容异常。
```

Claude Code 常见场景：

```text
PowerShell 输出乱码。
Git Bash 正常，PowerShell 不正常。
中文日志在命令行里显示异常。
```

OpenCode 常见场景：

```text
Windows native 环境乱码。
WSL2 环境正常。
终端字体或 shell 编码导致显示异常。
```

所以不要把它理解成“某个模型中文能力差”。

更准确的说法是：

```text
Windows shell 编码链路没有统一到 UTF-8。
```

## 10. 企业级大模型接入时要怎么处理

个人电脑上，`chcp 65001` 是快速修复。

但企业环境不能只靠每个人手动敲命令。

如果你的团队正在做企业级大模型接入，把 Claude、GPT、Gemini、DeepSeek 等模型接入内部系统、SaaS 产品、Agent 工作流、客服系统或知识库平台，我建议把编码问题纳入上线检查。

至少要检查五件事。

第一，开发终端是否统一：

```text
Windows Terminal / PowerShell 7 / WSL2 / Git Bash 的推荐组合是什么。
```

第二，AI CLI 启动方式是否统一：

```text
是否统一使用 UTF-8 wrapper。
是否要求启动前执行 chcp 65001。
是否在项目 AGENTS.md / CLAUDE.md 里写明编码规则。
```

第三，日志是否能追踪：

```text
AI 调用日志。
终端执行日志。
模型请求和响应日志。
失败原因。
文件修改记录。
```

第四，网关侧是否能审计：

```text
通过企业级 API 网关或大模型 API 统一入口，
记录模型、Key、项目、团队、时间、Token、错误码。
```

第五，提交前是否有检查：

```text
git diff 人工确认。
CI 检查文件编码。
关键文档变更必须 review。
```

这也是为什么我在企业接入里一直建议用统一入口，而不是每个人本地随便配一堆 Key。

通过 4SAPI 这类大模型 API 中转站或企业 API 网关接入时，你至少可以把模型调用、Key 权限、成本统计、失败日志、团队项目维度放在一个地方管理。

编码问题不一定发生在网关层。

但一旦出现“AI 输出异常、文件被改坏、日志看不懂”，统一日志和调用追踪能帮你更快定位是哪一层出了问题。

## 11. 推荐排查流程

我自己的排查顺序如下：

```text
1. 先运行 chcp，确认当前 code page。
2. 如果不是 65001，运行 chcp 65001。
3. 重新启动 Codex / Claude Code / OpenCode。
4. 用中文烟雾测试确认终端显示正常。
5. 如果仍乱码，补 PowerShell InputEncoding / OutputEncoding。
6. 如果显示方框，检查终端字体。
7. 如果文件也乱码，立刻看 git diff，不要让 AI 继续改。
8. 高频使用时，再做 wrapper、profile 或 WSL2 固化。
9. 团队使用时，把编码规则写进项目规范和 AI 使用规范。
```

浓缩成一句话：

```text
先救显示，再保文件，最后固化环境。
```

## 12. 一段可以直接复制给 AI 的规则

你可以把下面这段放进项目的 AGENTS.md、CLAUDE.md 或团队 AI 使用规范：

```text
Windows encoding rules:

- 在 Windows PowerShell / cmd 中执行中文相关命令前，先运行 chcp 65001。
- 不要根据终端显示乱码直接判断源文件损坏。
- 修改含中文文件前，先检查 git diff 和编辑器里的真实文件内容。
- 如果终端显示乱码，优先排查 code page、PowerShell OutputEncoding 和终端字体。
- 保持源码、Markdown、JSON、YAML、TOML 等文本文件使用 UTF-8。
```

这段规则的价值不是“高级”。

它的价值是让 AI 别自作聪明。

## 13. 参考命令合集

查看当前代码页：

```powershell
chcp
```

切换到 UTF-8：

```powershell
chcp 65001
```

查看 PowerShell 编码：

```powershell
[Console]::InputEncoding.WebName
[Console]::OutputEncoding.WebName
$OutputEncoding.WebName
```

设置 PowerShell UTF-8：

```powershell
[Console]::InputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
```

中文烟雾测试：

```powershell
Write-Host "中文测试 😀 こんにちは 한국어"
```

启动 Codex：

```powershell
chcp 65001; codex
```

启动 Claude Code：

```powershell
chcp 65001; claude
```

启动 OpenCode：

```powershell
chcp 65001; opencode
```

检查 AI 是否改坏文件：

```powershell
git diff
```

## 14. 常见误区

误区一：

```text
看到乱码，就让 AI 修复全部中文。
```

这很危险。

终端显示乱码不代表文件真的坏了。

误区二：

```text
以为换模型就能解决乱码。
```

多数时候换 GPT、Claude、DeepSeek 都没用，因为问题在本地终端编码。

误区三：

```text
只改 VS Code 编码，不改终端编码。
```

VS Code 打开文件正常，不代表 PowerShell 管道输出正常。

误区四：

```text
只看截图，不看 git diff。
```

截图只能证明你看到了乱码，不能证明文件被写坏。

误区五：

```text
一上来运行复杂脚本。
```

复杂脚本适合固化方案，不适合第一轮诊断。

先用 `chcp 65001` 确认方向。

## 15. 什么时候需要升级到长期方案

如果你只是偶尔用一次 AI CLI：

```text
chcp 65001
```

基本够用。

如果你每天都用 Codex、Claude Code 或 OpenCode：

```text
建议做 UTF-8 启动 wrapper。
```

如果你带团队做企业级大模型接入：

```text
建议统一终端、统一启动脚本、统一 AI 规则、统一日志审计。
```

如果你在 Windows native 环境怎么调都不稳定：

```text
可以考虑 WSL2。
```

WSL2 的 Linux 环境天然更接近 UTF-8 工作流，对命令行 AI 工具更友好。

但我仍然建议第一步从 `chcp 65001` 开始。

因为它最快，也最容易验证。

## 16. 参考资料

本文参考了 DeepSeek 技术社区文章《Windows 上彻底解决 Codex / Claude / OpenCode 中文乱码》的诊断思路，尤其是“先区分终端显示乱码和文件内容损坏”这一点。

同时，`chcp` 命令的作用可参考 Microsoft Windows 命令文档：它用于显示或设置当前控制台代码页；65001 对应 UTF-8 代码页。

参考链接：

```text
https://deepseek.csdn.net/6a31f80f10ee7a33f27e33aa.html
https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/chcp
https://learn.microsoft.com/en-us/windows/win32/intl/code-page-identifiers
```

## 总结

Windows 上 AI 输出乱码，先不要急着怪模型，也不要急着让 AI 改源码。

先运行：

```powershell
chcp 65001
```

再启动你的 AI CLI。

如果能恢复正常，说明问题大概率在终端代码页。

如果仍然乱码，再继续排查 PowerShell 编码、终端字体、Git diff 和文件真实内容。

对于个人用户，这是一条最快的救急命令。

对于企业团队，它应该变成 AI 工具接入规范的一部分：

```text
统一 UTF-8。
统一日志。
统一审计。
统一提交前检查。
```

这样，AI 才是在帮你写代码，而不是帮你制造一堆看不懂的乱码。
