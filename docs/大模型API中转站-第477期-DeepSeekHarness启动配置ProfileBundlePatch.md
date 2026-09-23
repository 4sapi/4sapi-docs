---
title: "DeepSeek Harness 启动时 Profile、Bundle 和 Patch 怎么生效"
category: 人工智能
tags:
  - DeepSeek Harness
  - Agent 架构
  - 插件系统
description: "从一次启动配置的解析顺序出发，解释 Profile、Bundle 和 Patch 如何共同得到最终的 Cordis 配置树，并说明自定义运行方案时需要验证什么。"
---

# DeepSeek Harness 启动时 Profile、Bundle 和 Patch 怎么生效

很多人看到 DeepSeek Harness 的插件配置后，第一反应是：为什么不在启动命令里直接列出所有插件？问题在于，一个可运行的 Host 往往需要一整套相互配合的能力。Session、模型、工具、持久化和 Web 服务并不是几个可以随手拼接的开关。

这篇文章只回答一个问题：一次 `dsh --profile <name>` 启动，怎样从一个运行方案得到最终的插件配置。这里的“配置树”指的是待 Cordis 加载的配置，不等同于插件已经成功启动。文中依据 DeepSeek Harness 官方仓库的 profile/bundle 设计说明，信息核对日期为 2026 年 8 月 25 日；具体内置名称和命令参数应以当前版本为准。

## 先看四层关系

可以把启动阶段压缩成下面这条链：

```text
Profile：选择一种运行方案和 Bundle 顺序
    ↓
Bundle：提供成套的 Cordis patch
    ↓
用户 patch：覆盖或补充局部配置
    ↓
最终配置树：交给 Cordis loader
```

Profile 是用户选择的入口。它通常位于 Harness home 下的 `profiles/<name>` 目录，包含自己的 `package.json` 和 `cordis.patch.yml`。其中 `dsh.profile.bundles` 是一个有序列表，明确指定要叠加哪些 Bundle。

Bundle 则是声明了 `dsh.bundle.patch` 的 npm 包。它不是一个“功能按钮”，而是一组默认配置层。例如，官方默认方案把共享底座和 Web 应用拆成不同 Bundle；headless 方案可以复用底座，但不叠加 Web 层。这样，运行形态的差异写在 Profile，公共能力的组合写在 Bundle。

最后才是用户 patch。它可以修改端口、模型提供方、工具开关，或加入 profile 自己安装的插件。patch 的作用是改变配置树，不是绕过插件的依赖检查。

## Profile 文件里通常写什么

一个 Profile 至少要表达两件事：它有哪些依赖，以及按什么顺序叠加 Bundle。下面是帮助理解结构的简化示意，不是可以直接复制到任意版本的完整配置：

```json
{
  "dependencies": {
    "@example/dsh-custom-tool": "^1.0.0"
  },
  "dsh": {
    "profile": {
      "bundles": [
        "@deepseek-ai/dsh-base",
        "@deepseek-ai/dsh-web-app"
      ]
    }
  }
}
```

Profile 自己的 `cordis.patch.yml` 只描述本地覆盖和额外配置。Bundle 的 patch 则由对应包的 `dsh.bundle.patch` 指向。把两类 manifest 分开有一个实际好处：查看 `package.json` 时，可以直接判断这个包是 Profile 还是 Bundle，而不用猜它是否会在启动时被特殊处理。

需要注意两个解析来源。内置 Bundle 要从与当前 `dsh` 相同的安装位置解析，避免用户 Profile 误加载另一份版本；Profile 自己安装的插件则从 Profile 的依赖树解析。这个“双锚点”设计解决的是版本和路径问题，并不等同于允许任意目录中的包自动进入 Host。

## 为什么要有 Bundle

假设 Web 和 headless 都需要 Session、模型注册、工具系统和 Agent Loop。如果两个 Profile 各自复制一份完整插件清单，会出现三个问题：

1. 相同配置散落在多个入口，升级时容易漏改。
2. Web 与 headless 的差异被混在一张很长的列表里，不容易审查。
3. 第三方扩展没有明确的位置贡献一整套运行层。

Bundle 把经常一起出现的配置放进一个可复用包。Profile 只负责选择 Bundle 及其顺序，用户层再处理本地差异。这个边界也解释了为什么“安装了一个 npm 包”不等于“它已经成为 Host 的一部分”：只有被 Profile 的 `bundles` 显式列出、并提供有效 patch 声明的包，才会贡献配置层。

## 顺序为什么必须显式

配置层存在覆盖关系。一个简化的合并过程可以写成：

```text
空配置根
  + base bundle patch
  + web-app 或 headless bundle patch
  + profile 自己的 cordis.patch.yml
  + 命令行指定的 patch overlay
  = 最终配置树
```

这里的顺序不是装饰。若两个层都修改同一个配置项，后应用的层会成为最终结果；若一个层依赖另一个层提供的插件或服务，顺序和依赖关系都需要同时检查。官方设计选择显式的有序 `dsh.profile.bundles`，而不是扫描所有依赖后再按隐含规则排序，原因正是让结果可预测、可 dump、可回归。

这也带来一个容易忽略的限制：Bundle 不会因为被某个已安装包间接依赖，就自动传递式激活。需要参与启动的 Bundle，应直接出现在 Profile 的列表中；想做元 Bundle，也要在自己的 patch 中明确完成组合。

## 一次启动可以怎样被验证

不要只看浏览器是否打开。配置解析至少有四个可观察点：

| 检查点 | 应该确认什么 | 失败通常说明什么 |
| --- | --- | --- |
| Profile 识别 | 当前进程使用的 Profile 名称和目录 | 参数拼写、目录位置或默认 Profile 不对 |
| Bundle 解析 | 每个名称对应到哪一个包、版本和 patch 文件 | 依赖未安装、manifest 缺失或路径解析错误 |
| 最终配置树 | 端口、Provider、工具开关和插件条目 | patch 顺序或字段覆盖关系不符合预期 |
| Cordis 启动 | 必需服务进入 ACTIVE，插件没有 PENDING/FAILED | 配置树虽生成，但依赖图没有闭合 |

官方实现让启动和配置 dump 复用同一条 patch 应用路径。实际排查时，优先保留这份最终配置和启动日志，再去判断是 Profile、Bundle 还是插件代码出了问题。否则只看源 YAML，很容易漏掉后应用层对字段的覆盖。

## 一个可复现的覆盖例子

假设 base Bundle 默认注册 `shell` 和 `files` 两类工具，web-app Bundle 再加入 Web API。用户 patch 要做一个只读入口，可以按下面的思路验证：

```text
base：提供 shell、files、sessions、llm、agent-loop
web-app：提供 API、Web Server、UI
用户 patch：关闭 files 的写入能力，限制 shell 工作目录
最终树：仍有读取和查询能力，但写入路径需要被工具层拒绝
```

这里有两个独立的验收条件：配置树中确实出现了限制，运行时也确实拒绝越界写入。只验证第一项，会把“声明了只读”误当成“已经只读”。工具的拒绝行为应使用测试目录、明确路径和可观察错误完成确认。

## Profile、Preset 不是一回事

两者都像“预设”，但作用范围不同：

| 概念 | 决定什么 | 作用时间 |
| --- | --- | --- |
| Profile | 整个 Harness 以 Web、headless 或其他形态启动 | 进程启动阶段 |
| Preset | 一个会话中的 Agent 拥有哪些工具、提示词和 Skills | 创建会话时 |

因此，换 Profile 通常意味着启动另一套 Host；换 Preset 则可以让同一个 Host 中的不同会话拥有不同 Agent 组合。把这两个概念混用，会导致错误的改造方向：想新增一个只读 Agent 时，不必复制整套 Web Host；想做一个不带 UI 的运行入口，也不应只修改某个 Agent Preset。

## 怎样设计一个自定义 Profile

先写运行目标，而不是先抄一份配置。建议把需求压缩成四项：

```text
入口：Web、headless，还是自定义 CLI
公共能力：Session、模型、工具、持久化是否需要
差异配置：端口、Provider、默认工具和外部依赖
验收结果：启动日志、可访问接口、一次最小任务和退出行为
```

例如，一个“只跑一次只读检查”的方案可能复用 base，再叠加 headless Bundle，并在用户 patch 中关闭写入工具。这里的“关闭”必须以最终配置树和实际拒绝行为为准，不能只看 YAML 中是否出现了某个字段。

创建后至少做三类验证：

1. 配置验证：确认 Bundle 顺序、patch 覆盖结果和插件解析来源。
2. 启动验证：确认所有必需服务都激活，没有插件停留在 PENDING 或 FAILED。
3. 行为验证：用一个无副作用任务检查模型请求、工具列表、权限和退出条件。

结果可以记录成一张很小的验收表：

```text
Profile：readonly-headless
Bundle 顺序：base -> headless
用户 patch：关闭写工具，限制 workspace
启动结果：无 FAILED，必要服务为 ACTIVE
只读任务：列出文件并输出数量
拒绝任务：尝试写入 workspace 外路径，返回权限拒绝
退出条件：任务完成后进程退出，未遗留后台服务
```

这张表比“启动成功”更有用，因为它同时记录了配置来源、能力范围和边界行为。换版本或修改 patch 后，重跑同一张表，就能知道变化来自哪里。

## 升级时最容易出现的漂移

Profile 依赖的 Bundle 可能升级，Bundle 内部的插件名、默认值和配置 schema 也可能变化。常见的漂移包括：

1. 一个字段仍然存在，但含义或默认值改变。
2. 插件包仍能解析，但它注入的服务名已经变化。
3. 用户 patch 仍能合并，却覆盖了新版新增的安全默认值。
4. Web 与 headless 共用的 base 变化，两个入口同时受到影响。

因此，升级验证不应只检查 `pnpm install` 或进程退出码。至少要比较升级前后的最终配置树，并重跑一次只读任务、一次工具调用和一次拒绝用例。若本地 patch 覆盖了大量默认值，最好先删掉可有可无的覆盖项，再做升级回归。

## 出错时先定位配置层

同一个“Agent 启动不了”现象，可能来自不同层。可以先按下面的分界定位：

| 现象 | 优先检查层 | 不要先做的事 |
| --- | --- | --- |
| 找不到 Profile | 启动入口和 Profile 目录 | 修改 Agent Preset |
| Bundle 名称无法解析 | Profile manifest 与依赖安装 | 复制一份完整插件清单 |
| 配置字段存在但值不对 | Bundle 顺序和用户 patch | 直接改插件源码 |
| 插件停留 PENDING | Cordis 服务依赖图 | 反复重启并猜测顺序 |
| 工具能显示但动作被拒绝 | 权限策略和工具层 | 把拒绝当成模型失败 |
| Web 能打开但任务不结束 | Agent Loop、持久化和退出策略 | 只检查 HTTP 端口 |

这个分层也适合写进故障记录。记录“最后一个已确认正常的层”，比只记录一条最终错误更容易复现。例如，若最终配置树已经正确生成，问题就不再是 Profile 解析，而应转向 Cordis 激活、Provider 配置或运行时权限。

## 常见误区

**把 Profile 当成插件清单。** Profile 只表达运行方案；具体能力仍由 Bundle patch 和 Cordis 依赖图共同决定。

**把“依赖已安装”当成“配置已生效”。** npm 依赖只是可解析的代码来源，是否被加载要看 Bundle 声明和最终配置树。

**用用户 patch 覆盖所有默认值。** 覆盖越多，升级时越难知道哪些行为来自官方 Bundle，哪些来自本地层。优先保留默认层，只写必要差异。

**只验证能启动。** Host 能打开并不代表 Agent 能完成任务。至少要继续验证一次模型请求、一次工具调用和一次失败路径。

**把 headless 当成“少一个 UI 的 Web”。** headless Bundle 还可能改变任务接收和进程退出方式。需要一次任务后退出，还是持续监听，都应在 Profile 的验收条件中单独写清。

**把命令行 patch 当成长期配置。** 命令行覆盖适合临时实验，长期运行方案应回收到 Profile 的 patch 文件并纳入版本管理，否则重启时很难复现当时的配置。

## 结论

Profile 负责选择运行形态，Bundle 负责复用成套配置，Patch 负责表达本地差异。三者的价值不在于让启动配置更复杂，而在于让“公共底座”和“部署差异”有各自的归属。

如果只是使用 Harness，先选择现成 Profile 并观察最终配置即可；如果要交付自定义运行入口，再从一个小的 Bundle 组合开始，保留配置 dump、启动日志和最小回归任务。具体字段、内置 Bundle 名称和解析路径会随版本变化，发布前应重新核对官方仓库。

资料来源：

- [Profile 与 Bundle 设计说明所在目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/.agents/notes/implemented/architecture)
- [Bundle base](https://github.com/deepseek-ai/deepseek-harness/tree/master/packages/bundle/base)
- [Bundle web-app](https://github.com/deepseek-ai/deepseek-harness/tree/master/packages/bundle/web-app)
- [DeepSeek Harness 官方仓库](https://github.com/deepseek-ai/deepseek-harness)
