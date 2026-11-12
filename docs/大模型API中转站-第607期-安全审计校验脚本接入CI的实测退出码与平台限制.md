---
title: "把智能体安全审计的校验脚本接进 CI：实测退出码与 Windows 限制"
category: 人工智能
tags:
  - 持续集成
  - 代码安全审计
  - Node.js
description: "说明 cloudflare/security-audit-skill 的两个校验脚本能在 CI 里做什么、不能做什么，记录在 Windows 上的实测结果（所有输入都以退出码 1 失败）及其源码原因，并给出尚未在 Linux 上验证的流水线草案。适合想把智能体审计产物纳入合并检查的工程团队。"
brand_mode: none
---

# 把智能体安全审计的校验脚本接进 CI：实测退出码与 Windows 限制

把智能体安全审计的产物纳入持续集成，听上去顺理成章：每次合并前检查 findings.json 是否合规。但这里有两个容易混淆的东西。一个是"让智能体去审计"，它需要编程智能体、子智能体调度和沙箱，不是普通 CI 作业能直接承担的。另一个是"校验审计产物的格式"，这只需要一个 Node 脚本。

本文只讨论后者。内容包括：校验脚本的用法和资源限制，我在 Windows 上实际跑出的结果和原因，一份没有在 Linux 上验证过的流水线草案，以及这一步在整个流程里的价值边界。

信息截止日期为 2026-10-30。实测环境是 Windows，Node.js v24.14.0。Linux 和 macOS 上的行为我没有测过，文中会逐处标明。

## 一、CI 里能放什么，不能放什么

依据仓库 [cloudflare/security-audit-skill](https://github.com/cloudflare/security-audit-skill) 的 README（检索于 2026-10-30）：

| 事项 | 能否放进普通 CI 作业 | 原因 |
| --- | --- | --- |
| 运行完整审计（六个阶段） | 不建议 | README 要求支持工具调用和并行子智能体的编程智能体，以及 OS 级沙箱 |
| 校验 findings.json 的格式 | 可以 | validate-findings.cjs 是零依赖的 Node 脚本 |
| 校验 coverage-ledger.json | 可以 | validate-coverage-ledger.cjs 同样是零依赖脚本 |
| 判断报告里的漏洞是否属实 | 不能 | 校验只证明格式和账本一致性，真伪要人工复核 |

README 的要求里还有一条与 CI 直接相关：运行目标代码的构建、测试、浏览器、模糊测试等，必须在能关闭外部网络、使用白名单环境并限制资源的沙箱里进行。没有这些控制时，技能的规则是不执行目标代码，把线索保留为 needs_validation。

所以 CI 里合理的角色是**门卫**：检查别处产出的审计产物是否合格，而不是**审计员**。

## 二、两个脚本的用法

两个脚本的帮助文本都是一行：

```text
node validate-findings.cjs <path-to-findings.json>
node validate-coverage-ledger.cjs <path-to-coverage-ledger.json>
```

文件位于仓库的 skills/security-audit/ 目录。不带参数运行时，脚本打印 Usage 并以退出码 1 结束，这一点我实测过。

### 资源上限

读源码入口部分可以确认这些限制：

| 脚本 | 限制 |
| --- | --- |
| validate-findings.cjs | 输入不超过 5 MiB；顶层记录不超过 1000 条；嵌套不超过 64 层；错误信息最多输出 100 条 |
| validate-coverage-ledger.cjs | 输入不超过 5 MiB；单元不超过 10000 个；各集合项不超过 1000；字段数不超过 1000；嵌套不超过 64 层；错误最多 100 条 |

这些限制对 CI 有一个直接含义：产物过大会被当作错误，而不是被截断后继续。

## 三、Windows 上的实测

我把脚本和 schema 下载到本地，用 Node.js v24.14.0 在 Windows 上做了几组输入：

| 输入 | 命令 | 输出 | 退出码 |
| --- | --- | --- | --- |
| 空数组 `[]` | `node validate-findings.cjs empty.json` | Failed to read findings JSON: OS no-follow and nonblocking input protection is unavailable | 1 |
| 非法 JSON | `node validate-findings.cjs bad.json` | 同上 | 1 |
| 结构错误 `[{"title":"x"}]` | `node validate-findings.cjs wrong.json` | 同上 | 1 |
| 文件不存在 | `node validate-findings.cjs nope.json` | 同上 | 1 |
| 不带参数 | `node validate-findings.cjs` | Usage: node validate-findings.cjs <path-to-findings.json> | 1 |
| 非法 JSON，账本脚本 | `node validate-coverage-ledger.cjs bad.json` | Failed to read coverage ledger: OS no-follow and nonblocking input protection is unavailable | 1 |

结果是：**在这台 Windows 机器上，两个脚本对任何文件都无法读取，包括格式正确的空数组**。这意味着在这个环境里，退出码 1 不携带任何信息，不能区分"产物不合格"和"脚本无法运行"。

### 原因

我读了 validate-findings.cjs 的读取逻辑。它在打开输入文件时要求同时使用 O_NOFOLLOW 和 O_NONBLOCK 两个标志，目的是不跟随符号链接、不在特殊文件上阻塞，这与技能"不信任目标控制的文件"的设计一致。代码在这两个常量缺失时抛出上面的错误。

我用 Node 直接查询了常量：

```text
node -e "const c=require('fs').constants; console.log(c.O_NOFOLLOW, c.O_NONBLOCK, process.platform)"
输出：undefined undefined win32
```

在 Windows 上这两个常量都是 undefined，所以脚本在读取阶段就拒绝运行。账本脚本的错误信息相同，说明用的是同一类检查，我没有逐行核对它的实现。

### 结论的范围

- 已确认：Node.js v24.14.0 加 Windows 上，两个脚本无法读取任何输入。
- 已确认原因：两个文件打开标志在该平台未定义。
- 未验证：其他 Windows 配置、较旧的 Node 版本、WSL 或容器环境下是否相同。我检查了 WSL 状态，只看到 Docker Desktop 的发行版，没有进一步尝试。
- 未验证：Linux 和 macOS 上能否正常运行。从原理上，这两个标志在这些系统上通常存在，但这是推断，没有实测。

## 四、这对 CI 选型的意义

1. **不要用 Windows 运行器跑这两个脚本。** 至少在我测过的配置上，它们无法工作。
2. **用 Linux 运行器，并先做一次冒烟测试。** 在流水线里加一个已知合格的最小样例，以及一个已知不合格的样例，确认前者退出码为 0、后者为非 0。没有这个冒烟测试，就分不清脚本坏了还是产物坏了。
3. **别把"退出码非 0"直接当作产物不合格。** 要同时检查错误信息里是否含 "input protection is unavailable" 这类环境错误，把它分类为基础设施故障。

关于退出码 0：我没有在任何环境里看到过校验成功的运行，所以"成功返回 0"是依据常见约定和脚本里 `process.exit(run(...))` 的写法做的推断，没有观察到。

## 五、流水线草案（未验证）

下面是一份草案。我没有运行过它。请先在一个测试分支上验证，再接入受保护分支。

```yaml
name: audit-artifacts-check
on:
  pull_request:
    paths:
      - "audits/**"

jobs:
  validate:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<固定到你核验过的提交>
      - uses: actions/setup-node@<固定到你核验过的提交>
        with:
          node-version: "<与你本地验证一致的版本>"
      - name: Fetch validators at a pinned commit
        run: |
          # 把技能仓库固定到你审阅过的提交，不要跟随默认分支
          git clone https://github.com/cloudflare/security-audit-skill /tmp/skill
          git -C /tmp/skill checkout <你审阅过的提交>
      - name: Smoke test with known-good and known-bad fixtures
        run: |
          node /tmp/skill/skills/security-audit/validate-findings.cjs ci/fixtures/good-findings.json
          ! node /tmp/skill/skills/security-audit/validate-findings.cjs ci/fixtures/bad-findings.json
      - name: Validate submitted artifacts
        run: |
          node /tmp/skill/skills/security-audit/validate-findings.cjs audits/findings.json
          node /tmp/skill/skills/security-audit/validate-coverage-ledger.cjs audits/coverage-ledger.json
```

这份草案里的几个设计选择，理由如下：

| 选择 | 理由 |
| --- | --- |
| 固定第三方仓库的提交 | 校验脚本会在你的 CI 里执行，应当是你审阅过的版本 |
| 最小权限（只读内容） | 这个作业只读文件，不需要写权限 |
| 只在 audits 目录变化时触发 | 避免对无关修改产生噪声 |
| 冒烟测试放在前面 | 区分"脚本坏了"和"产物坏了" |
| 不在 CI 里运行审计智能体 | 沙箱、网络和凭据的要求超出普通作业 |

注意这份草案里没有写 `actions/checkout` 和 `setup-node` 的具体版本。请自行核对这些动作当前的版本和提交，并固定到提交哈希。

两个夹具文件（good-findings.json、bad-findings.json）需要你自己准备。good 样例应符合 report-schema.json 的某个结论分支，我没有构造过，因为我无法验证自己构造的样例在脚本里确实通过。可以从技能仓库里的 validate-findings.test.cjs 中寻找官方测试用到的样例，README 称该测试文件包含"与生产者兼容的夹具检查"，我没有读它的内容。

## 六、CI 门卫能带来什么价值

即使脚本在 Linux 上正常工作，它能保证的也很有限。按文档，成功只证明格式和账本一致性。能拦住的问题：

- 字段缺失或多余，结论类型用错了字段。
- 调用路径结构不合规。
- 严重度超过已记录的影响。
- 账本和记录之间的引用不一致。

拦不住的问题：

- 引用的行号是否存在，调用路径是否可达。
- 复现结果是否真实。
- 严重度是否合理。
- 有人手写了一份看起来合规的假产物。

所以这一步适合做**入口卫生**：保证交到人手里的报告是机器可读、结构完整的，减少人工复核时对格式的纠缠。它不能代替人工复核。

## 七、合并策略上的建议

1. **把这个检查设为"必须通过"之前，先观察一段时间。** 用只报告不阻塞的模式跑几周，看冒烟测试是否稳定。
2. **不要让这个检查成为放行依据。** 通过只表示格式合格。是否合并含 confirmed 的修复，仍要看人工复核结论。
3. **保留产物的来源信息。** run-metadata.json 里记录了源提交、配置和范围，建议和 findings.json 一并提交，复核者需要它判断报告对应的是哪个版本的代码。
4. **产物不要放进被审计仓库的受版本控制目录。** README 说明，只有当你明确选择一个被版本控制忽略的目录时，工作流才会写入目标仓库内部。默认输出目录在目标之外。如果你的流程要把产物提交到某处，用单独的仓库或有访问控制的位置，因为里面可能包含未修复漏洞的细节。

第 4 条尤其重要。findings.json 和 FINDINGS-DETAIL.md 描述的是尚未修复的漏洞，放进公开仓库等于公开披露。

## 八、限制与结论

接入 CI 的合理形态是：用 Linux 运行器执行两个零依赖脚本，检查审计产物的格式，并配上冒烟测试来区分脚本故障和产物问题。

本文的限制：实测只覆盖 Windows 和 Node.js v24.14.0，两个脚本在该环境下全部失败；Linux 和 macOS 上的行为未验证；流水线草案未运行；成功路径（退出码 0）未观察到；我没有读完两个脚本各八百多行的校验逻辑，只读了入口、限制常量和文件读取部分。

下一步建议：在一台 Linux 机器上，用 Node.js 下载这两个脚本，对一个空数组和一个故意写错的文件各跑一次，记录退出码和输出。只有这两个数据点成立，才值得把它放进流水线。
