---
title: "DeepSeek Harness 插件如何通过 Cordis 依赖图协同"
category: 人工智能
tags:
  - DeepSeek Harness
  - Cordis
  - 插件系统
description: "用服务、依赖注入、事件和 effect 四个概念，解释 DeepSeek Harness 的插件如何协同启动、参与流程并在卸载时清理资源。"
---

# DeepSeek Harness 插件如何通过 Cordis 依赖图协同

“一切皆插件”听起来像把系统拆成很多零件，但零件多并不会自动变成架构。真正的问题是：插件怎样知道谁已经准备好，怎样替换实现，又怎样在移除后不留下半截状态？

DeepSeek Harness 使用 Cordis 处理这件事。本文不讨论某个具体插件的安装，而是回答一个更基础的问题：插件如何通过服务依赖、事件和生命周期组成可运行的 Host。示例语法来自官方 Cordis 教程；它用于说明机制，不代表某个产品插件的完整实现。

## 服务是稳定的协作接口

一个插件可以向上下文提供具名服务，其他插件只依赖这个名字，而不直接导入提供方的实现。Harness 中常见的 `ctx.sessions`、`ctx.llm`、`ctx.tools` 和 `ctx.agents` 都可以按这个思路理解。

提供方的简化写法如下：

```ts
import { Service, type Context } from '@deepseek-ai/cordis'

export class GreeterService extends Service {
  constructor(ctx: Context) {
    super(ctx, 'greeter')
  }

  greet(who: string) {
    return `Hello, ${who}!`
  }
}

export function apply(ctx: Context) {
  ctx.plugin(GreeterService)
}
```

`super(ctx, 'greeter')` 会把服务注册到当前上下文。注册本身是一个 effect，提供方卸载时，服务也会被撤销。

这里有一个经常被忽略的区分：服务名称是运行时契约，TypeScript 类型声明只是编译期辅助。官方教程会用 `declare module '@deepseek-ai/cordis'` 把 `greeter` 加入 `Context` 接口；没有这段声明时，运行时仍可能找到服务，但消费方失去了类型检查。自定义 Harness 插件应同时维护这两层，否则“代码能编译”和“服务实际存在”会变成两套互不相干的判断。

消费方则声明 `inject`：

```ts
export const name = 'consumer'
export const inject = ['greeter']

export function apply(ctx: Context) {
  console.log(ctx.greeter.greet('world'))
}
```

只要 `greeter` 尚未出现，consumer 就停留在 `PENDING`，不会抢先执行。等依赖满足后，它才进入加载流程。如果提供方后来被卸载，消费方也会随之卸载；服务恢复后，依赖它的插件可以重新加载。

这意味着配置文件的行顺序不是主要依据。真正决定插件何时激活的是依赖图。替换一个 `shell` 提供方时，所有注入 `shell` 的插件都可以复用，而不必逐个修改消费方。

## 依赖图和作用域

服务不是永远挂在一个全局表里。Cordis Context 具有父子关系，插件在什么 Context 中注册，决定了哪些消费方能看到它。一个 Agent Preset 可以为单个 Agent 注册工具和提示词；宿主平面的 Agent Loop 则需要依赖跨会话共享的 `sessions`、`llm` 和 `tools` 服务。

这解释了一个常见故障：插件代码没有报错，但 Agent 看不到它注册的工具。可能原因不是工具实现失败，而是工具注册到了另一个作用域。排查时要同时确认：

```text
提供方挂载在哪个 Context？
消费方从哪个 Context 注入服务？
两者之间是否存在可见的父子关系？
这个服务应该是进程级共享，还是每个 Agent 一份？
```

如果一个本应逐 Agent 隔离的服务被注册到根 Context，第二个 Agent 可能遇到名称冲突；反过来，一个宿主 API 需要的共享服务若被放进某个 Agent 子树，宿主就会一直等待一个永远不会出现的依赖。作用域是配置设计的一部分，不是加载成功后的实现细节。

## PENDING 不是启动失败

排查 Harness 时，要把三种状态分开：

| 状态 | 含义 | 常见原因 |
| --- | --- | --- |
| PENDING | 声明存在，但硬依赖尚未满足 | 提供方没加载或名称不匹配 |
| ACTIVE | `apply` 已完成，插件正在工作 | 依赖和配置均已通过 |
| FAILED | 加载或配置校验抛出异常 | 参数错误、代码异常或版本不兼容 |

PENDING 不一定会打印明显错误，尤其是在一个没有其他活跃任务的最小组合里，进程可能直接结束。看到“插件没反应”时，先检查它注入的服务是否真的由当前 Profile/Bundle 提供，而不是立即修改业务代码。

## 事件让插件从侧面参与流程

服务适合“我需要调用谁”，事件适合“流程走到这里，谁想观察或介入”。一个工具执行过程，可以由工具插件负责实际动作，再由审批、权限或遥测插件监听事件：

```text
工具准备执行
  -> 审批插件决定是否需要确认
  -> 权限插件检查路径和身份
  -> 工具插件执行
  -> 遥测插件记录结果和耗时
```

这些监听方不需要修改 Agent Loop，也不必彼此导入。Cordis 的 `ctx.on()` 同样属于 effect，监听器会在插件卸载时自动移除。

事件有不同的分发语义，不能把所有监听都当成普通广播：

| 方式 | 用途 | 是否等待监听器 |
| --- | --- | --- |
| `emit` | 同步通知 | 不等待异步结果 |
| `parallel` | 并发执行多个监听 | 等待全部完成 |
| `serial` | 按顺序执行并等待 | 支持首个有效结果胜出 |
| `waterfall` | 允许转换结果或短路 | 由监听器决定是否继续 |

其中，waterfall 最容易造成误判。只负责记录日志的监听器必须调用 `next()`；如果忘记调用，后续默认逻辑会被悄悄截断。只有确实要拒绝或替换结果时，才应该短路。

可以用“审批”来区分两种监听器：

```text
遥测监听器：记录工具名、结果类别和耗时，然后调用 next()
审批监听器：判断当前调用是否需要人工确认；拒绝时有意不调用 next()
```

如果两个插件都想给出决定，就要确认事件采用的是 `serial` 还是 waterfall，以及谁拥有最终裁决。不要用多个普通事件监听器返回布尔值，再期待框架自动合并；事件模式本身就是协作协议。

## 服务替换的完整过程

把一个 Provider 换成另一个 Provider 时，系统通常经历四步：

1. 卸载旧提供方，并撤销它注册的服务和相关 effect。
2. 依赖旧服务的消费方进入卸载或 PENDING 状态。
3. 挂载新提供方，使用同一个稳定服务名注册实现。
4. 依赖图重新满足后，消费方重新加载。

消费方只依赖 `llm`、`shell` 或 `tools` 这样的职责名，因此可以不改业务逻辑。代价是替换过程必须有清晰的停顿边界：正在执行的请求、后台任务和外部副作用不能因为服务名相同就被当成可以无缝切换。对有写入动作的 Provider，应先在隔离项目演练卸载、重载和失败回滚。

## effect 解决卸载问题

插件往往不只注册服务，还会创建定时器、连接、watcher 或后台任务。Cordis 要求这类资源通过 `ctx.effect()` 返回清理函数：

```ts
export function apply(ctx: Context) {
  ctx.effect(() => {
    const timer = setInterval(() => console.log('tick'), 1000)
    return () => clearInterval(timer)
  })
}
```

当插件被卸载、热替换，或硬依赖消失时，清理函数会执行。内置注册 API（事件监听、子插件、服务注册）本身也遵循同样的 effect 语义。

资源清理是插件化系统的底线。如果插件只从列表中消失，却没有停止后台任务、撤销工具注册或释放连接，下一次重新加载就可能得到重复监听、旧状态和不可解释的行为。

Cordis 的插件实例可以理解为一个有生命周期的 fiber：

```text
PENDING -> LOADING -> ACTIVE -> UNLOADING -> DISPOSED
                         \-> FAILED
```

`PENDING` 表示依赖未满足，`FAILED` 表示加载或校验抛错，`DISPOSED` 则表示清理已经完成。不要把 FAILED 当成“稍后一定会自动恢复”；只有依赖消失后再恢复，或配置热替换触发重新装载时，才可能重新走一遍生命周期。

清理顺序也值得测试。Cordis 会在插件卸载时执行 disposer，异步 disposer 可能并发运行；如果资源释放存在先后依赖，应在同一个 disposer 中显式等待，而不是假设多个 disposer 会按业务顺序串行完成。

## Agent Loop 为什么只依赖服务名

在 Harness 中，Agent Loop 需要把多个能力串起来，但它不应该知道每个能力的具体包名。它只声明需要 `agents`、`sessions`、`llm`、`tools` 和 `systemPrompt` 等服务：

```text
Agent Loop
  ├─ ctx.agents：找到并管理 Agent
  ├─ ctx.sessions：读取会话与事件
  ├─ ctx.llm：选择模型提供方并发起请求
  ├─ ctx.tools：校验、审批、执行工具
  └─ ctx.systemPrompt：组装提示词和运行时信息
```

因此，多个模型 Provider 可以注册到 `ctx.llm`，多个工具可以贡献到 `ctx.tools`。前者通常是运行时选择一个候选实现，后者则可能同时贡献一组能力。这两个“多个插件”不是同一种关系，设计自定义插件时要先分清楚。

可以用一个最小依赖图检查 Agent Loop 是否具备启动条件：

```text
AgentLoop
  requires: agents, sessions, llm, tools, systemPrompt

agents       <- Agent registry
sessions     <- Session service
llm          <- one or more Provider registrations
tools        <- file/shell/search/... contributions
systemPrompt <- persona, tool descriptions, runtime variables
```

只要其中一个硬依赖没有提供，Agent Loop 就不应进入 ACTIVE。相比“按固定槽位把五个包塞进去”，依赖图允许不同 Profile 选择不同实现，也让缺失能力以 PENDING 的形式暴露出来。

## 一个最小的排查顺序

当自定义插件没有按预期工作时，可以沿着下面顺序排查：

1. 确认它是否出现在最终配置树，而不是只安装在 `node_modules`。
2. 检查 `inject` 中的服务名是否有提供方，是否因为作用域不同而不可见。
3. 查看插件是否处于 PENDING、ACTIVE 或 FAILED。
4. 若依赖已满足，再检查 `apply` 中的配置校验和 effect 初始化。
5. 若流程被截断，检查 waterfall 监听器是否错误地没有调用 `next()`。
6. 卸载后重复加载一次，确认服务、工具、事件和后台资源都被清理。

## 给插件写一组最小测试

一个插件至少需要覆盖正常、缺依赖和卸载三条路径。可以在测试配置中安排：

| 测试 | 操作 | 预期 |
| --- | --- | --- |
| 正常加载 | 同时挂载提供方和消费方 | 消费方进入 ACTIVE，能调用服务 |
| 缺失依赖 | 只挂载消费方 | 消费方保持 PENDING，不执行部分逻辑 |
| 提供方替换 | 卸载旧提供方，再挂载同名新实现 | 消费方不会持有旧引用，恢复后调用新实现 |
| 事件短路 | 让审批监听器拒绝一次调用 | 下游默认动作不发生，拒绝原因可观察 |
| 资源清理 | 重复挂载和卸载插件 | 没有重复监听、残留定时器或重复注册 |

如果插件接触文件、网络或外部系统，还要增加拒绝用例和取消用例。仅仅看到一次“Hello, world!”，只能证明最简单的提供与消费成功，不能证明生命周期安全。

## 什么时候不该拆成插件

“一切皆插件”不等于所有代码都要拆出去。以下内容通常更适合留在运行时或稳定策略层：

```text
插件加载、依赖解析和生命周期管理
任务取消、权限确认和状态清理
所有 Agent 都必须遵守的事件顺序
产品层默认导航和最低限度的可用体验
```

相反，变化频繁、依赖独立、需要单独权限或可能被替换的能力，更适合插件化。把安全边界交给任意业务插件自行决定，或者把内部资料连接器硬编码到内核，都会让后续维护变困难。

## 结论

Cordis 把插件协同拆成三种关系：服务依赖保证“需要谁时谁已就绪”，事件让插件在不互相引用的情况下参与流程，effect 保证卸载和替换不会遗留资源。

这套机制的收益是可替换，代价是需要理解作用域、依赖和生命周期。写插件时，先定义稳定的服务职责，再补上事件边界、拒绝路径和清理逻辑，比单纯增加一个可调用函数更接近 Harness 的设计方式。

资料来源：

- [Cordis 官方教程目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/docs/cordis-tutorial)
- [Agent Loop 官方包目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/packages/core/agent-loop)
