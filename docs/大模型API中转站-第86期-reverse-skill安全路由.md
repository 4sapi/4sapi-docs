---
title: "【大模型API中转站】第86期 reverse-skill安全路由 | Agent少猜命令多干活"
category: 人工智能
tags:
  - 大模型API中转站
  - reverse-skill
  - Skill Router
  - Cybersecurity
  - MCP
  - 4SAPI
description: "拆解 reverse-skill 这套面向代码 Agent 的安全任务路由系统：为什么安全和逆向任务不能让 Agent 直接猜命令，如何用 RULES、Skill Router、工具索引、MCP 和 field journal，把 APK、二进制、JS、抓包、CTF 等任务变成可路由、可执行、可沉淀的工作流。"
---

# 【大模型API中转站】第86期 reverse-skill安全路由 | Agent少猜命令多干活

本文是【大模型API中转站】系列的第86篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前几篇我们一直在讲一件事：

```text
不要只把 AI 当聊天窗口。
要把它变成能按流程干活的工作台。
```

第81期讲 Claude Project。

第82期讲 Memory 和 Knowledge。

第83期讲 Skills 和 Artifacts。

第84、85期讲 Image2 + Sora2 的长视频生产线。

这一篇换一个更硬核的场景：

```text
安全分析、逆向分析、CTF、抓包、APK、二进制、前端 JS。
```

这类任务最怕什么？

不是模型不会解释。

而是模型一上来就猜命令。

你让它分析 APK。

它可能直接让你跑 jadx。

你让它看前端签名。

它可能直接让你搜 `encrypt`。

你让它做二进制分析。

它可能直接甩一堆 IDA、radare2、Ghidra 名词。

看起来很专业。

但真正执行时经常变成：

```text
工具没检查。
路径没确认。
任务类型没分清。
静态和动态链路混在一起。
输出一堆解释，但没有真正进入工具。
遇到同类问题，下次继续从零踩坑。
```

这就是 `reverse-skill` 想解决的问题。

它的正式定位是：

```text
Cybersecurity Skills Router
面向代码 Agent 的安全任务路由与工具编排系统。
```

一句话说：

```text
先判断任务，再选择 Skill，最后调用真实工具执行。
```

这不是一个单独工具。

也不是一份普通 prompt。

它更像给 Claude Code、Codex CLI、Cursor、Cline、Windsurf 这类代码 Agent 装了一套安全任务工作流操作系统。

## 1. 普通 Agent 做安全任务的问题

普通代码 Agent 写业务代码时，通常还能靠上下文和仓库结构慢慢摸。

但安全和逆向任务不一样。

这类任务天然有几个难点。

第一，目标类型差异太大。

```text
APK
ELF
EXE
DLL
SO
前端 JS
HTTP 抓包
PCAP
固件
CTF 题目
LLM 应用
API 服务
```

这些东西不能用同一条分析链。

APK 可能先看 Manifest 和 Java 层。

二进制可能要先判断架构、导入表、字符串和符号。

前端 JS 可能要先观察请求，再定位签名入口。

抓包任务可能要先整理请求链路。

CTF 题可能要先判断它到底是 Web、Crypto、Reverse、Pwn 还是 Forensics。

第二，工具链分散。

常见工具包括：

```text
jadx
apktool
adb
Frida
IDA Pro
radare2
BurpSuite
Playwright
Graphviz
Nmap
sqlmap
ffuf
hashcat
```

工具在不在本机？

路径在哪里？

是否已经注册成 MCP？

当前系统是 Windows、Kali、Ubuntu 还是 macOS？

这些问题不先确认，后面就是空转。

第三，安全任务有边界。

同一个工具，在不同场景下性质完全不同。

授权环境里的内部安全测试、教学实验、CTF、自己 App 的逆向分析，是合理场景。

未经授权扫描、入侵、绕过他人系统限制，就是红线。

所以安全类 Agent 不应该只追求“更会执行”。

还要先把任务范围、授权上下文、风险门控和输出边界讲清楚。

第四，经验很容易浪费。

比如你今天解决了一个 APK 反编译失败的问题。

下次换一个 APK，Agent 又重新猜一遍。

你今天发现某个工具版本不兼容。

下次换一台机器，还是踩同一个坑。

如果没有经验回写，AI 工作流永远像临时工。

当场能聊。

但不会成长。

## 2. reverse-skill 的核心思路

`reverse-skill` 的核心链路很清楚：

```text
用户任务
  ↓
RULES.md
  ↓
Skill Router
  ↓
目标场景 Skill
  ↓
工具 / MCP / 脚本
  ↓
报告 + field journal
```

换成大白话：

```text
先别急着干。
先判断这是什么任务。
再找对应的方法论。
再检查本机工具。
再调用真实工具。
最后生成报告，并把经验写回去。
```

这个顺序很关键。

普通 prompt 包经常是：

```text
遇到 APK，你可以用 jadx。
遇到 JS，你可以用 AST。
遇到二进制，你可以用 IDA。
```

这类提示有用，但不够。

因为它只告诉模型“知道什么”。

`reverse-skill` 更像告诉模型：

```text
先读哪份规则。
怎么路由。
什么时候检查工具索引。
什么时候进入子 Skill。
什么时候生成报告。
什么时候回写经验。
```

所以它不是让 Agent “知道更多术语”。

而是让 Agent “少猜、少跳步、能落地执行”。

## 3. 它适合什么人

这套东西不适合完全零基础用户直接拿来乱跑。

它更适合三类人。

第一，安全研究和逆向分析人员。

你已经知道 jadx、Frida、IDA、BurpSuite 这些工具大概做什么。

但希望 Agent 帮你把流程串起来。

第二，开发团队和安全团队。

你们内部有授权测试、代码审计、移动端安全验证、API 安全验证、CTF 训练或演练环境。

希望把工具链、报告模板、经验库和模型调用统一起来。

第三，AI Agent 工作流玩家。

你关心的不是某一个安全工具。

而是：

```text
怎么让 Agent 先路由。
怎么让 Agent 调真实工具。
怎么让 Agent 避免空转。
怎么让 Agent 把成功经验沉淀成下一次的上下文。
```

如果你在做企业内训、红蓝队实验室、移动端安全验证、AI 自动化分析平台，这套项目的结构值得研究。

## 4. 它和普通 prompt 包有什么不同

普通 prompt 包通常是一段文本：

```text
你是一个逆向专家。
请按步骤分析 APK。
```

这能提升模型回答风格。

但很难保证执行稳定。

`reverse-skill` 的不同点在于，它有可执行结构。

它至少包含这几层：

```text
RULES.md          全局路由与行为规则
skills/SKILL.md   总控入口
skills/routing.md 路由矩阵
tool-index        本机工具状态
子 Skill           APK、JS、IDA、radare2、CTF 等场景工作流
MCP / 脚本         真实执行入口
field-journal     经验沉淀
docs-generator    报告生成
```

所以它不是单纯告诉 Agent：

```text
你应该更专业。
```

而是逼 Agent 按顺序做事：

```text
先路由。
再查工具。
再执行。
再报告。
再复盘。
```

这对安全任务尤其重要。

因为安全任务一旦跳步，结果通常不是“回答差一点”。

而是：

```text
方向错了。
工具错了。
路径错了。
证据链断了。
报告不可复现。
风险边界不清楚。
```

## 5. 路由矩阵：先分清任务，再进入工具

`skills/routing.md` 是这个项目里很关键的一层。

它要求 Agent 在执行前先完成路由。

不是做着做着再想。

而是先判断三个维度：

```text
目标类型
用户意图
工具链需求
```

举几个例子。

如果目标是 APK / Android App：

```text
优先进入 apk-reverse。
如果核心逻辑在 native .so，再分流到 IDA 或 radare2。
如果需要动态验证，再考虑 Frida。
```

如果目标是前端 JS：

```text
优先进入 js-reverse。
先观察请求。
再捕获参数。
再定位签名入口。
再做本地复现。
```

如果目标是二进制：

```text
可以走 ida-reverse 做深度分析。
也可以走 radare2 做命令行侦察。
```

如果是 CTF：

```text
先进入 CTF-Sandbox-Orchestrator。
再按证据分流到 Web、Reverse、Pwn、Crypto、Forensics 等子方向。
```

这种设计的价值在于：

```text
不是模型想起哪个工具就用哪个。
而是任务类型决定入口。
```

这就是 Skill Router 的意义。

## 6. 工具索引：别让 Agent 猜你的机器

很多 AI 工具链失败，不是模型不会。

而是它不知道你的电脑上到底有什么。

`reverse-skill` 里有一个很实用的设计：

```text
tool-index.md
tool-index.json
```

它们不是固定文档。

而是本机扫描结果。

第一次 clone 后，需要运行刷新脚本生成工具索引。

Windows 侧大致是：

```powershell
powershell -ExecutionPolicy Bypass -File skills/scripts/refresh-tool-index.ps1
```

Linux / macOS 侧大致是：

```bash
bash skills/scripts/refresh-tool-index.sh
```

这个动作会告诉 Agent：

```text
node 在不在。
python 在不在。
jadx 在不在。
apktool 在不在。
adb 在不在。
radare2 在不在。
IDA 路径有没有。
MCP 入口有没有。
```

注意，工具索引只代表当前机器。

不要把别人的 `tool-index.md` 当成你的环境。

迁移到新电脑后，应该重新刷新。

这点非常工程化。

因为很多人写 AI 工作流，只写“应该安装什么”。

但真正执行时，Agent 需要知道：

```text
现在这台机器实际有什么。
```

## 7. Bootstrap：缺工具时不要原地发呆

工具索引发现缺工具后，下一步就是 bootstrap。

项目里区分了不同平台：

```text
Windows       PowerShell 脚本
Kali Linux    Kali 专项 Bash 脚本
Ubuntu/Debian 通用 Bash 脚本
macOS         通用 Bash 脚本 + Homebrew 思路
```

它的思路不是“所有工具都自动装好”。

更现实一点：

```text
能自动装的，按脚本补齐。
不能自动装的，输出结构化手动引导。
装完后刷新工具索引。
再继续任务。
```

比如 jadx、apktool、radare2、Frida 这类工具，可以根据平台做自动或半自动安装。

但 IDA Pro、BurpSuite、Android Build-Tools 这类涉及商业软件、桌面应用或 SDK 环境的工具，通常需要人工确认路径和安装状态。

这个设计比“失败后反复重试”好很多。

正确的 Agent 行为应该是：

```text
发现缺失。
尝试补齐。
验证是否可用。
失败就给出明确步骤。
等待用户确认。
恢复执行。
```

这也是安全类工作流很需要的一种稳态。

## 8. MCP：把真实工具暴露给 Agent

安全任务不能只靠模型脑补。

真正有价值的是让 Agent 调用真实工具。

`reverse-skill` 里提到的 MCP 能力主要包括：

```text
BurpSuite MCP
IDA / idalib MCP
浏览器自动化
anything-analyzer
jshookmcp
```

它们解决的是同一个问题：

```text
让 Agent 不只是说“你可以打开 Burp”。
而是能读取代理历史、分析请求、操作浏览器、调用反编译结果。
```

举一个安全团队内部场景。

你在授权测试环境里排查某个接口的鉴权逻辑。

普通 Agent 可能只能说：

```text
建议检查 JWT、Cookie、权限字段。
```

接入工具后，理想状态是：

```text
浏览器采样请求。
Burp 记录代理历史。
Agent 读取请求和响应。
路由到 API 安全或 Web 分析 Skill。
生成可复现的验证报告。
```

这一步才是从“咨询式 AI”进入“执行式 AI”。

## 9. field journal：让 Agent 记住踩过的坑

我很喜欢这个项目里的 `field-journal` 设计。

很多团队做 AI 自动化，最容易忽略这个环节。

一次任务完成后，如果没有回写机制，Agent 下次还是从零开始。

`field-journal` 的目标是把真实项目里的经验沉淀下来。

它记录的不是漂亮话。

而是：

```text
任务类型
执行链路
踩坑记录
工具链发现
关键命令
适用范围
对路由矩阵的改进建议
```

这意味着系统会越用越像你的环境。

比如：

```text
某版本 jadx 解析失败，换 apktool 先拆资源。
某台 Windows 机器 IDA 路径不在默认位置。
某类前端签名要先看 SourceMap，再补浏览器环境。
某个 CTF 类型应该先进入 forensic，而不是 web-runtime。
```

这些东西如果只留在聊天记录里，很快就丢了。

写进 field journal，下次同类任务就能复用。

这也是 AI Agent 工程化的关键：

```text
让经验从一次对话，变成系统资产。
```

## 10. 报告生成：不要只给聊天结论

安全分析任务最后一定要有交付物。

不是一句：

```text
我发现问题了。
```

而是要能给团队、客户或自己复盘：

```text
目标是什么
授权范围是什么
分析方法是什么
用了哪些工具
关键证据在哪里
风险影响是什么
复现条件是什么
修复建议是什么
哪些内容不能外传
```

`reverse-skill` 里有 `docs-generator` 和 `diagram-generator`。

这两个模块的价值不是“把文字写漂亮”。

而是把执行过程变成正式报告。

比如：

```text
APK 安全分析报告
前端签名分析报告
CTF writeup
攻击路径图
请求链路图
工具执行记录
```

对团队来说，这比一次聊天输出更重要。

因为安全工作需要可复现、可审计、可交接。

如果 Agent 只是当场答得很爽，但不能生成结构化报告，那它还没有进入团队流程。

## 11. 一个典型 APK 分析工作流

假设你在授权环境中，要分析自己 App 的签名校验逻辑。

普通 Agent 可能会直接说：

```text
用 jadx 打开 APK，搜索 sign、verify、check。
```

这不算错。

但太粗。

`reverse-skill` 希望 Agent 走的是：

```text
识别任务：APK / Android / 签名校验
读取路由：进入 apk-reverse
检查工具：jadx、apktool、adb、Frida 是否可用
静态拆解：Manifest、Java 层、资源、native library
判断分流：Java 层足够，还是需要 native 分析
必要时动态验证：只在授权测试设备和范围内进行
输出报告：校验位置、调用链、验证方式、风险说明
经验回写：记录本次工具链和踩坑
```

这条链路更像一个真实安全工程师的工作方式。

不是“会一个命令”。

而是有阶段、有证据、有分流、有报告。

## 12. 一个典型 JS 签名分析工作流

再看前端 JS 场景。

很多网站或内部系统会有请求签名、时间戳、nonce、设备指纹、加密参数。

在授权测试或自有系统排查里，常见需求是：

```text
为什么请求失败？
签名参数在哪里生成？
前端和后端校验有没有不一致？
测试环境的 SDK 是否和生产版本不同？
```

普通 Agent 可能会让你：

```text
搜索 sign、token、encrypt、crypto。
```

但实际项目里，前端经常有打包、混淆、动态导入、运行时环境依赖。

更稳的链路应该是：

```text
先观察真实请求。
再定位目标接口。
再捕获关键参数变化。
再找生成入口。
再用浏览器调试或 Hook 辅助取证。
再做本地最小复现。
最后写清楚参数来源和验证边界。
```

这正是 `js-reverse` 这类子 Skill 的价值。

它让 Agent 不会一上来就在源码里乱搜。

而是从请求证据出发。

## 13. 一个典型 CTF 工作流

CTF 场景也很适合 Skill Router。

因为 CTF 最大的问题是：

```text
题目表面像 Web，实际核心可能是 Crypto。
题目文件像 Reverse，实际还要 Pwn。
抓到一段流量，可能是 Forensics，也可能是 Protocol Reverse。
```

如果 Agent 没有路由，容易被第一眼误导。

`reverse-skill` 中接入了 `CTF-Sandbox-Orchestrator`。

它的思路是先建模：

```text
题目给了什么文件？
暴露了什么服务？
已有证据指向哪个方向？
当前卡点是环境、漏洞点、解码、利用还是报告？
```

然后再进入对应子技能。

这比“看到 CTF 就开始猜 flag”靠谱得多。

对训练和教学来说，也更适合复盘。

因为最后能沉淀出 writeup，而不是只拿到一个答案。

## 14. 和 4SAPI 的关系

`reverse-skill` 本身不是大模型 API 中转站。

它解决的是：

```text
Agent 应该怎么路由任务。
应该怎么调用本机工具。
应该怎么沉淀经验。
```

4SAPI 这类大模型 API 中转站，适合放在模型调用层。

也就是：

```text
代码 Agent / 安全工作流
  ↓
Skill Router
  ↓
工具 / MCP / 脚本
  ↓
模型调用层
  ↓
4SAPI 统一模型入口
```

为什么这里需要中转站？

第一，统一模型入口。

安全分析里可能同时用 Claude、GPT、Gemini 或国产模型。

不同模型适合不同任务。

比如：

```text
长上下文读文档。
代码解释。
报告生成。
日志归纳。
结构化检查清单。
```

统一入口后，工作流里不用到处改模型地址。

第二，记录调用日志。

安全任务需要审计。

谁调用了模型？

调用了什么模型？

花了多少 token？

失败率多少？

这些信息不能只靠聊天窗口记。

第三，做成本治理。

安全分析和报告生成都很吃 token。

如果每一步都用最贵模型，很快就烧钱。

更合理的方式是：

```text
路由判断用轻量模型。
复杂代码解释用强模型。
报告润色用性价比模型。
日志归纳用便宜模型。
```

第四，做团队权限管理。

企业里不能让所有人共用一个 Key。

应该按团队、项目、环境拆分。

中转站可以承接这层：

```text
Key 管理
模型路由
费用统计
错误日志
权限隔离
```

所以 `reverse-skill` 和 4SAPI 的关系不是替代。

而是上下游。

```text
reverse-skill 管 Agent 怎么干活。
4SAPI 管模型怎么调用、怎么计费、怎么审计。
```

## 15. 安全边界必须写在前面

这类文章必须强调边界。

`reverse-skill` 涉及逆向、安全测试、抓包、二进制分析、渗透工具链。

这些能力只能用于：

```text
授权环境中的安全研究
自有应用分析
内部安全测试
教学实验
CTF 竞赛
防护验证
合规审计
```

不要用于：

```text
未经授权访问他人系统
绕过他人服务限制
破坏性测试
真实目标攻击
窃取数据
扩散攻击脚本
规避法律或平台规则
```

安全工具链的价值，不在于“让 AI 更会攻击”。

而在于：

```text
让授权测试更可控。
让分析过程更可复现。
让报告更完整。
让防护验证更系统。
```

如果团队要引入这类 Skill Router，建议先定四条规则：

```text
1. 每个任务必须有授权范围。
2. 高风险动作必须人工确认。
3. 所有模型调用和工具执行要留日志。
4. 输出报告必须区分内部版和可公开版。
```

这不是形式主义。

这是安全工作流能不能进企业环境的前提。

## 16. 最小接入路径

如果你只是想理解和试用，可以按这个路径走。

第一步，先读概览。

```text
OVERVIEW_zh.md
```

它是给人看的。

先理解项目定位、能力边界和整体结构。

第二步，再读 README。

```text
README_zh.md
```

它更偏给 AI Agent 执行 bootstrap。

里面会要求先刷新工具索引，再读规则，再进入路由。

第三步，生成工具索引。

Windows：

```powershell
powershell -ExecutionPolicy Bypass -File skills/scripts/refresh-tool-index.ps1
```

Linux / macOS：

```bash
bash skills/scripts/refresh-tool-index.sh
```

第四步，读规则和路由。

```text
RULES.md
skills/SKILL.md
skills/routing.md
```

第五步，用一个低风险真实任务验证。

比如：

```text
分析一个自己写的测试 APK。
整理一段授权抓包记录。
对一个 CTF 题目生成 writeup。
检查内部测试 API 的请求链路。
```

不要一上来就把它接到高风险生产环境。

先让工作流跑通。

再考虑团队化接入。

## 17. 企业接入时要补什么

如果团队要把这套思路放进内部环境，还需要补几层。

第一，任务准入。

不是所有人都能发起所有安全任务。

要区分：

```text
普通代码分析
内部安全验证
生产环境只读排查
高风险动态测试
对外报告生成
```

第二，数据脱敏。

不要把真实用户数据、密钥、Cookie、内部域名、未公开漏洞细节直接交给模型。

至少要做：

```text
密钥替换
域名脱敏
用户标识脱敏
样本分级
报告分级
```

第三，模型分层。

不是所有步骤都需要最强模型。

建议拆成：

```text
路由模型
分析模型
报告模型
审计模型
```

4SAPI 这类中转站可以在这里做模型路由和成本统计。

第四，日志审计。

要记录：

```text
谁发起任务
任务授权范围
调用了哪些工具
用了哪个模型
输出了什么报告
是否进入 field journal
```

第五，人工确认。

尤其是动态测试、请求重放、漏洞验证、生产环境排查这类任务，不应该完全自动化。

Agent 可以准备计划。

但关键动作要有人点头。

## 18. 最容易踩的坑

第一，把它当成工具安装包。

它不是。

它的核心价值是路由和编排。

工具只是执行面。

第二，只读概览，不读 README 和 RULES。

概览是给人看的。

真正让 Agent 行为改变的是：

```text
README_zh.md
RULES.md
skills/SKILL.md
skills/routing.md
```

第三，不刷新工具索引。

不刷新索引，Agent 就不知道你机器上有什么。

后面很容易猜路径。

第四，MCP 只写配置，不验证可调用。

`tool-index` 显示有 node 或 npx，不代表 MCP 已经在你的客户端启用。

要实际验证。

第五，不做经验回写。

如果不用 field journal，这套系统会少一半价值。

因为它不会从你的真实任务中成长。

第六，把授权边界写得太模糊。

安全任务最怕一句：

```text
帮我测一下这个目标。
```

更好的输入应该是：

```text
这是我方授权测试环境。
范围是 example.internal 的测试子域。
只做只读分析和请求链路整理。
不做破坏性测试。
最后输出内部报告。
```

边界越清楚，Agent 越稳定。

## 19. 最小检查清单

接入 `reverse-skill` 前，先过一遍：

```text
[ ] 是否明确这是授权环境
[ ] 是否知道当前目标类型：APK / JS / 二进制 / 抓包 / CTF / API
[ ] 是否读过 OVERVIEW_zh.md
[ ] 是否读过 README_zh.md
[ ] 是否刷新了 tool-index
[ ] 是否确认本机关键工具可用
[ ] 是否配置并验证 MCP
[ ] 是否读了 RULES.md、skills/SKILL.md、skills/routing.md
[ ] 是否从低风险任务开始验证
[ ] 是否设计了报告输出格式
[ ] 是否启用 field journal 经验回写
[ ] 是否把模型调用接入日志和成本统计
[ ] 团队使用时是否有权限、审计和人工确认
```

如果这些都没有做，先别急着跑复杂任务。

先把地基打稳。

## 20. 最后总结

`reverse-skill` 最值得学习的，不是里面列了多少安全工具。

真正值得学的是这套结构：

```text
先路由。
再检查工具。
再进入子 Skill。
再调用真实执行面。
再生成报告。
再回写经验。
```

普通 Agent 最大的问题，是太容易“看起来很懂”。

安全任务最需要的，却不是看起来懂。

而是：

```text
边界清楚。
步骤稳定。
证据完整。
工具真实。
结果可复现。
经验能沉淀。
```

`reverse-skill` 把这件事做成了一套工作流。

个人用户可以把它当成安全分析和逆向任务的学习框架。

安全团队可以把它当成 Agent 化工具链的参考架构。

企业用户如果要落地，还要加上：

```text
权限控制
数据脱敏
日志审计
人工确认
模型成本治理
报告分级
```

4SAPI 适合放在模型调用层：

```text
统一 Claude、GPT、Gemini 等模型入口。
记录调用日志。
统计 token 成本。
按任务类型做模型路由。
把 Agent 工作流从单人实验推进到团队可控使用。
```

一句话总结：

```text
安全类 Agent 的关键，不是让模型多背工具名。
而是让它先路由、少猜命令、按证据执行、把经验留下来。
```

这就是 `reverse-skill` 这类 Skill Router 最有价值的地方。

## 资料来源与延伸阅读

- reverse-skill: OVERVIEW_zh.md https://github.com/zhaoxuya520/reverse-skill/blob/main/OVERVIEW_zh.md
- reverse-skill: README_zh.md https://github.com/zhaoxuya520/reverse-skill/blob/main/README_zh.md
- reverse-skill: ARCHITECTURE.md https://github.com/zhaoxuya520/reverse-skill/blob/main/ARCHITECTURE.md
- reverse-skill: PLATFORMS.md https://github.com/zhaoxuya520/reverse-skill/blob/main/PLATFORMS.md
- reverse-skill: skills/routing.md https://github.com/zhaoxuya520/reverse-skill/blob/main/skills/routing.md
- reverse-skill 仓库：https://github.com/zhaoxuya520/reverse-skill
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
