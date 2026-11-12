---
title: "逆向 Codex 怎样验证结论：搭建版本固定的协议实验室"
category: 人工智能
tags:
  - Codex
  - 协议测试
  - 软件工程
description: "把 Codex 源码阅读变成可复核实验，通过版本指纹、Schema 对照、初始化探针和任务矩阵，区分声明变化、实际运行结果与尚未验证的推断。"
brand_mode: none
---

# 逆向 Codex 怎样验证结论：搭建版本固定的协议实验室

源码里看到一个分支，接着写“Codex 就是这样运行的”，中间往往少了一步：证明你测的版本与分析对象是什么关系。
同样，一个初始化探针成功，也不能说明工具、压缩和恢复全部正常。
做运行机制研究时，最容易丢掉的不是代码细节，而是结论的适用范围。

本文给出一个小型实验室的组织方式。
它围绕版本指纹、协议声明、实际消息和任务产物建立证据，适合研究客户端行为、维护自己的接入层，或者复核一篇技术分析。
示例重点是离线 Schema 检查和本机初始化，不需要先搭建复杂网关。

资料核对日期为 2026-10-20。
本机版本为 `codex-cli 0.141.0`，源码参考提交为 `46d252f8bf8b09f4ef765bc72e40f02c229da3e6`。
两者分别记录；本文没有编译源码，也没有据此断言二进制与该提交一一对应。

## 先定义研究问题，再收集日志

“研究 Codex”太宽，很容易收集一大堆无法回答问题的材料。
可以先把问题写成一个可判定的句子：

- 当前二进制能否完成 stdio 初始化？
- 升级后 `turn/start` 的声明是否变化？
- 某次调用是否真的读取了测试文件？
- 用户中断后，当前回合和子进程分别是什么状态？
- 压缩后是否仍满足指定文件范围的约束？

这些问题需要不同证据。
初始化看请求响应，协议变化看 Schema，读取看工具记录和随机标记，中断要看终态与副作用。
不要用一份“最终回复成功”的截图回答全部问题。

## 用目录把材料和结论放在一起

在独立研究目录准备如下结构。
这是本文建议的实验文件布局，不是 Codex 自身的目录格式。

```text
codex-lab/
  manifest.json
  fixture/
  schema-before/
  schema-after/
  runs/
    handshake/
      outgoing.jsonl
      incoming.jsonl
      stderr.log
      summary.json
    read-marker/
  findings.md
```

`fixture` 放合成材料，`runs` 保存单次实验，`findings.md` 写带证据的结论。
不要把真实业务工作区复制进测试目录；随机文本和小仓库已经能验证多数协议主干。
这样既减少变量，也方便重复运行。

## 指纹需要覆盖二进制和源码两条线

版本号用于识别发行信息，文件哈希用于识别具体二进制。
源码提交用于限定分析对象，配置摘要用于说明实验条件。
它们各有用途，不要用一个值代替其他值。

```json
{
  "observed_on": "2026-10-20",
  "cli_version": "codex-cli 0.141.0",
  "binary_sha256": "执行本地命令后填写",
  "source_commit": "46d252f8bf8b09f4ef765bc72e40f02c229da3e6",
  "os": "填写实际系统",
  "transport": "stdio",
  "config_summary": "仅记录与本次实验有关的非敏感条件",
  "model_turn_submitted": false
}
```

占位字段不是实验结果。
如果无法确认二进制构建来源，保留这个限制，别用提交日期接近来推断它们相同。
本机哈希可以这样获取：

```powershell
$codexBinary = 'C:\research\bin\codex.exe'
Get-FileHash -LiteralPath $codexBinary -Algorithm SHA256
```

路径必须替换为实际原生可执行文件。
若 `Get-Command codex` 指向包装脚本，先定位它启动的程序；包装脚本哈希不能代替程序哈希。
这一步只读取指定文件，不需要解析整个用户配置目录。

## 先验证协议声明，再安排运行实验

本机帮助和导出命令可以提供当前二进制的协议声明：

```powershell
codex --version
codex app-server --help
codex app-server generate-json-schema --help
codex app-server generate-json-schema --out .\schema-before
```

本次已实际核对这些命令，并成功导出 Schema。
公开仓库也提供版本化协议定义，适合与固定提交对照。
来源：[app-server 协议 Schema](https://github.com/openai/codex/tree/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json)。

导出 Schema 不需要通过一次模型任务来证明成功。
它证明的是当前程序能给出声明，并不证明某个账号可调用全部方法。
某项定义存在、环境允许访问和运行成功，仍是三个要分别检查的问题。

## 写一个离线的初始化规格检查器

下面的 Node.js 脚本只读本地文件，不启动 Codex，不连接模型服务。
保存为 `check-initialize.mjs`，在含有导出 Schema 的研究目录运行。
前置条件是可运行当前脚本语法的 Node.js，以及由目标版本导出的文件。

```javascript
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const root = process.argv[2] ?? 'schema-before';
const load = (name) => JSON.parse(
  readFileSync(join(root, name), 'utf8')
);

const initialize = load('v1/InitializeParams.json');
assert(initialize.required.includes('clientInfo'));

const info = initialize.definitions.ClientInfo;
assert(info.required.includes('name'));
assert(info.required.includes('version'));

const notifications = load('ClientNotification.json');
const methods = notifications.oneOf.flatMap(
  (item) => item.properties?.method?.enum ?? []
);
assert(methods.includes('initialized'));

console.log(JSON.stringify({
  initializeClientInfoRequired: true,
  initializedNotificationDeclared: true,
  networkRequests: 0,
}));
```

执行方式：

```powershell
node check-initialize.mjs .\schema-before
```

本次同等断言已对本机导出结果执行通过。
断言失败时，应先查看导出结构是否变化，不要先修改程序使它“总能打印成功”。
这只是规格检查，不是完整 JSON Schema 验证器，也没有覆盖所有参数条件。

## 升级对照要区分文本差异与语义差异

升级前后分别导出 `schema-before` 和 `schema-after`。
先查看新增方法、删除方法、必需字段变化，再逐项决定是否影响自己的客户端。
仅有文件差异不能自动判定协议不兼容。

| 差异 | 可能影响 | 后续验证 |
| --- | --- | --- |
| 新增可选字段 | 严格解析器可能拒绝 | 未知字段处理 |
| 新增必需字段 | 旧请求可能不完整 | 最小请求回归 |
| 枚举增加值 | 穷举处理可能遗漏 | 未知状态展示 |
| 方法删除或改名 | 调用失败 | 替代路径与版本支持 |
| 描述文字变化 | 不一定改变行为 | 阅读实现与说明 |

可以先用 Git 的目录比较查看文本变化：

```powershell
git diff --no-index -- .\schema-before .\schema-after
```

有差异时，该命令返回非零退出码属于正常比较结果。
如果用于自动化脚本，要把“有差异”和“命令执行失败”分开处理。
本文没有对升级前后两个发行版本做完整兼容性回归，因此不报告兼容或不兼容结论。

## 用握手探针验证消息边界

初始化探针只需完成一条很短的链路：启动进程、发送 `initialize`、匹配响应、发送 `initialized`、退出并保存记录。
实际客户端还有更多工作，但这条链路已经能检查路径、分帧和消息配对。
初始化字段来自固定提交的定义。
来源：[InitializeParams](https://github.com/openai/codex/blob/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json/v1/InitializeParams.json)。

本次探针的实际结果是初始化通过，摘要标明没有提交模型回合。
生成记录包括出站消息、入站消息、stderr 和汇总 JSON。
原始初始化结果包含本机目录等环境字段，因此对外展示时只保留字段名和必要状态。

判定规则可以写得明确一些：

```text
通过：目标请求 ID 收到 result，且没有 error。
失败：收到 error、解析异常、超时或进程提前退出。
范围：只证明初始化，不包含模型与工具能力。
```

超时是失败分类，不是服务必然不兼容的最终根因。
需要继续看启动日志、程序路径和环境要求。
服务端通知数量也可能因环境变化，不能把一次测到的数量写成固定协议断言。

## 模型任务另外建立实验矩阵

握手通过后，才有必要安排更接近真实任务的检查。
每次只增加一个变量，保存实际版本、输入、事件和产物。
以下矩阵是待执行方案，本次没有运行这些模型任务。

| 用例 | 输入设计 | 通过依据 |
| --- | --- | --- |
| 纯文本回合 | 要求返回短标记 | 对应回合终态和实际文本 |
| 文件读取 | 文件中放随机 receipt | 读取证据与 receipt 匹配 |
| 允许范围内修改 | 明确一个文件和目标 | 差异、测试与回合状态 |
| 受限写入 | 使用独立受限测试目标 | 策略结果与文件未变 |
| 中断 | 可控的长任务 | 中断结果和副作用检查 |
| 上下文处理 | 分阶段任务与明确约束 | 压缩证据及约束保持情况 |

受限写入应在你自己的测试环境中设计。
如果没有出现拒绝，先排查环境和权限前置条件，不要直接宣布发现绕过。
对于中断，回合终止和所有子进程消失也不能合并成一个未经检查的结论。

## 事件录制不能只保存最后一句话

运行时的生命周期有线程、回合和 item 等粒度。
如果使用 `codex exec --json`，按它自己的事件结构消费；如果使用 app-server，就按对应通知与请求定义处理。
官方非交互文档说明前者输出 JSONL 事件，公开协议定义列出了后者的消息。
来源：[Non-interactive mode](https://developers.openai.com/codex/noninteractive)、[ServerNotification](https://github.com/openai/codex/blob/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json/ServerNotification.json)，核对日期 2026-10-20。

录制器至少应保留方向、原始消息、接收顺序和本地时间。
需要跨来源关联时，再提取线程、回合、item 或调用身份。
不要给本来没有某个字段的事件补一个猜测值。

## 写结论时给出证据强度

研究笔记可以用下面几种句式，分别表达事实与限制：

| 证据 | 合适的表述 |
| --- | --- |
| 代码分支 | 该提交包含此路径 |
| Schema | 该版本声明此字段或方法 |
| 本机握手 | 本环境初始化通过 |
| 单次任务 | 该输入与配置下观察到该结果 |
| 重复任务集 | 在明确样本和条件内报告结果 |

不要把“源码有实现”升级成“所有入口都已启用”，也不要把一次失败升级成“功能完全不可用”。
对未知部分留空，比补上一段听起来合理的机制更有研究价值。

一个可复核的 Codex 实验室不需要复杂设备。
固定对象、保存原始证据、把声明和运行分开，就能让后来的读者检查你的结论，也让下一次升级后的变化有处可查。
