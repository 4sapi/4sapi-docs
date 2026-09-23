---
title: "Coding Agent 的能力检查与外部权限层"
category: "人工智能"
tags:
  - AI Agent
  - 安全工程
  - 软件工程
description: "用能力注册表和外部策略预检不同 Coding Agent，明确降级、拒绝和人工确认的边界。"
brand_mode: none
---

# Coding Agent 的能力检查与外部权限层

不同 Coding Agent 都能“读文件、改代码、跑命令”的印象，常常掩盖了它们在工具能力与权限模型上的差异。

有的运行时能产出结构化补丁，有的只能写回整个文件；有的能运行浏览器，有的只能使用 Shell。

如果控制层假设这些能力完全等价，任务切换时就可能跳过测试、扩大写入范围或直接失败。

本文介绍一个最小的能力注册表和外部权限层，让任务在启动前明确判断“可运行、可降级、需确认或拒绝”。

读者将得到能力声明、任务需求、策略配置和预检输出的参考格式。

文中配置是自建控制层示例，不会替代操作系统权限、容器隔离或代码审查。

## 适用场景

这套方法适合需要在多个 Harness 之间路由任务的代码仓库。

它也适合团队希望把文件范围、网络访问和 Shell 命令统一审查的情况。

如果所有任务都只读且没有工具调用，能力注册表的价值有限。

如果任务涉及生产环境、真实密钥或不可逆数据操作，仍需使用组织既有的身份、审批和隔离系统。

## 先分开三个概念

能力表示 Harness 理论上能做什么。

权限表示当前任务被允许做什么。

策略表示控制层实际会放行或拦截什么。

例如某个 Harness 具备运行 Shell 的能力，不代表本次任务能运行任意命令。

同样，任务需要运行测试，不代表策略允许安装依赖或访问网络。

把三者混在一个“权限等级”里，会让失败原因难以解释。

## 定义通用能力名称

能力名要描述结果，而不是某个供应商的 API。

以下列表足够覆盖多数代码维护任务。

```yaml
# .agent/capabilities.yaml
capabilities:
  read_files:
    description: 在授权范围内读取文件
  write_patch:
    description: 以补丁或等价方式修改文件
  run_shell:
    description: 在受控环境执行命令
  run_tests:
    description: 运行任务声明的测试命令
  structured_result:
    description: 返回可解析的统一结果对象
  browser:
    description: 使用浏览器完成任务
  subagents:
    description: 委派子任务
```

`write_patch` 不要求某个工具必须使用特定 diff API。

只要它能在策略允许的文件范围内产生可审查的变更，就可以映射到该能力。

能力定义应版本化，避免适配器和任务文件对同名能力有不同理解。

## 为每个适配器登记能力

适配器的声明应该保守。

不确定是否支持的能力，应当不声明，而不是乐观填入。

```yaml
# .agent/harnesses/codex.yaml
id: codex
adapter: adapters/codex
capabilities:
  - read_files
  - write_patch
  - run_shell
  - run_tests
  - structured_result
limits:
  supports_network_policy: true
  supports_subagents: false
```

另一个适配器可以有不同能力集合。

```yaml
# .agent/harnesses/review-only.yaml
id: review-only
adapter: adapters/review-only
capabilities:
  - read_files
  - structured_result
limits:
  supports_network_policy: false
  supports_subagents: false
```

该声明不是对上游产品的宣传或评价。

它是团队适配器在当前版本、当前部署方式下愿意承担的合同。

## 在任务中声明必需与可选能力

任务不能只写“使用最佳工具”。

它应声明完成验收所必需的能力，以及缺失时可接受的降级项。

```yaml
requirements:
  required:
    - read_files
    - write_patch
    - run_tests
    - structured_result
  optional:
    - browser
    - subagents
```

缺少任一必需能力时，预检应拒绝运行或选择另一个适配器。

缺少可选能力时，预检可以选择串行执行、跳过非关键步骤，或要求人工提供替代输入。

降级行为必须写入运行结果，不能悄悄发生。

## 写出可解释的预检规则

预检至少比较任务需求、适配器能力和当前策略。

```text
missing_required = task.required - harness.capabilities
missing_optional = task.optional - harness.capabilities

if missing_required is not empty:
    reject with missing_required
if policy denies a required operation:
    reject with policy reason
if missing_optional is not empty:
    continue with a recorded degradation
else:
    allow
```

预检的输出应当对人可读，也能被程序解析。

例如“拒绝：缺少 `run_tests`”比“运行失败”更有可操作性。

## 将权限放在 Harness 之外

提示词中的“不要读取密钥”属于行为期望，不是访问控制。

可靠的控制层至少需要在工具调用前判断路径、命令和网络请求。

对于本地环境，这可能由包装器、受限用户、容器或沙箱共同实现。

对于远程运行时，这可能由工作区挂载、代理网关和临时凭证实现。

具体技术可以不同，但决策应来自同一份策略。

## 一个最小文件策略

```yaml
# .agent/policies.yaml
filesystem:
  read:
    - services/payment/**
    - tests/payment/**
    - .agent/**
  write:
    - services/payment/**
    - tests/payment/**
    - .agent/runs/**
  deny:
    - .env
    - .git/**
    - secrets/**
    - migrations/**
```

策略应优先匹配 `deny`，再判断读写白名单。

控制层必须在规范化路径之后再匹配，防止 `../` 绕过。

符号链接、挂载点和大小写不敏感文件系统也需要在实现时单独处理。

仅依赖字符串前缀匹配通常不够安全。

## Shell 策略不应只有允许或拒绝

Shell 命令存在不同风险等级。

测试、格式化和静态检查可以放在允许列表。

安装依赖、修改锁文件或执行迁移可以要求人工确认。

推送远程分支、删除文件或读取环境变量应有更严格限制。

```yaml
shell:
  allow:
    - npm test -- payment
    - npm run lint
  confirm:
    - npm install
    - npm run db:migrate
  deny:
    - git push
    - rm -rf
    - printenv
network: deny
```

命令匹配不能只比较字符串开头。

例如允许 `npm test` 不应自动允许通过 Shell 拼接额外删除命令。

更稳妥的做法是让控制层调用预定义命令 ID，而不是执行任意文本。

## 用命令 ID 减少解析风险

任务中引用命令 ID，而不是直接提交一段 Shell。

```yaml
acceptance:
  checks:
    - type: command
      command_id: test_payment
    - type: command
      command_id: lint_payment
```

控制层从受审查的 `commands.yaml` 获取具体命令和工作目录。

这样 Agent 即使建议其他命令，也不会自动获得执行权限。

需要临时命令时，运行记录应标记为 `confirmation_required`，由人确认后再执行。

## 处理网络与环境变量

网络访问和环境变量读取经常被忽略。

即使文件权限正确，工具也可能通过网络把信息带出工作区。

因此，任务策略应显式写入网络状态。

环境变量也应按白名单注入，而不是把完整宿主环境交给执行者。

```yaml
environment:
  allow:
    - NODE_ENV
    - CI
  deny_all_others: true
network:
  default: deny
```

实际执行环境能否做到完全隔离，取决于底层运行时。

当适配器无法落实某条强制策略时，应拒绝任务，而不是只在提示词中提醒。

## 预检输出示例

```json
{
  "status": "allowed_with_degradation",
  "harness": "codex",
  "required_capabilities": ["read_files", "write_patch", "run_tests"],
  "missing_required": [],
  "missing_optional": ["subagents"],
  "policy": {
    "network": "denied",
    "write_roots": ["services/payment", "tests/payment"]
  },
  "degradations": ["独立检查将串行执行"]
}
```

`allowed_with_degradation` 不是成功完成任务。

它只说明在当前能力与策略下可以开始执行，并且某些可选行为会改变。

最终结果还必须记录测试和人工检查的状态。

## 如何验证权限层

首先创建一个只读评审任务，确认预检拒绝写入能力要求。

然后请求写入 `migrations/` 中的文件，预检或工具包装器应拦截该动作。

再尝试执行一个未登记命令，控制层应返回需要确认或拒绝的结果。

最后切换到不支持 `run_tests` 的适配器。

该任务应在启动前被拒绝，而不是执行后才发现没有测试证据。

验证时使用无敏感信息的临时仓库，避免将真实凭证作为测试样本。

## 常见失败与排查

### 失败一：能力名称过于具体

把能力命名为某个供应商的工具调用，会让每个任务都绑定实现细节。

优先描述目标能力，例如 `write_patch` 或 `run_tests`。

适配器负责把目标映射为下游工具。

### 失败二：适配器夸大能力

如果适配器只能写文件却不能限制路径，就不应把自己声明为可执行受限写入。

应先补齐外部包装器或降低能力声明。

保守声明会增加拒绝次数，但能避免错误的安全假设。

### 失败三：确认动作没有记录

人工确认后执行了安装或迁移，却没有在运行记录中留下谁确认、确认什么。

这会让后续复盘无法区分自动执行和人工授权。

确认事件应和命令、时间、运行 ID 一起保存。

### 失败四：将策略当作质量保证

权限策略可以限制副作用，不能保证代码逻辑正确。

测试、代码审查和验收仍是独立环节。

不要因为权限收紧，就省略变更验证。

## 结论

多 Harness 管理的重点不是让所有执行者看起来一样。

更可靠的做法是显式登记能力，在任务开始前完成匹配，并由 Harness 外部的策略层决定文件、命令、网络和确认边界。

无法满足必需能力或强制策略时，拒绝运行比静默降级更可控。
