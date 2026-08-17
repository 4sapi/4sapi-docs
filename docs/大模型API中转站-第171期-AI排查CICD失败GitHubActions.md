---
title: "【大模型API中转站】第171期 AI排查CI/CD失败 | GitHub Actions日志"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - CI/CD
  - GitHub Actions
  - Claude Fable 5
  - DevOps
description: "CI/CD 失败日志很长，最适合用辅助模型先提取失败步骤，再让 Fable 5 判断根因。本文以 GitHub Actions 为例，讲依赖、测试、构建、权限、密钥、部署阶段的 AI 排查方法。"
---

# 【大模型API中转站】第171期 AI排查CI/CD失败 | GitHub Actions日志

本文是【大模型API中转站】系列的第171篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

CI/CD 失败很适合用 AI。

原因简单：

```text
日志长。
噪音多。
真正失败点常常只有一两行。
```

GitHub Actions 失败时，很多人会从头翻几千行日志。

AI 更好的用法是：

```text
辅助模型先提取失败步骤。
Fable 5 判断根因和最小修复。
复核模型检查会不会破坏部署流程。
```

## 1. CI/CD 先分阶段

CI/CD 不是一个步骤。

通常包括：

```text
checkout。
安装依赖。
lint。
typecheck。
test。
build。
docker build。
push image。
deploy。
smoke test。
```

排查第一步：

```text
失败在哪个 step。
```

不要把整条流水线混成一句：

```text
CI 挂了。
```

## 2. 给 AI 的 CI 排错包

```text
【平台】
- GitHub Actions / GitLab CI / Jenkins / 其他：

【失败信息】
- workflow 名称：
- job 名称：
- step 名称：
- exit code：
- 失败日志最后 100 行：

【环境】
- 语言：
- node/python/go/java 版本：
- 包管理器：
- 缓存策略：

【最近变更】
- package lock：
- workflow yaml：
- Dockerfile：
- 测试文件：
- 环境变量或 secrets：

【边界】
- 不输出 secrets
- 不建议直接关闭测试
- 不直接跳过失败步骤
- 先给最小修复建议
```

这份包能让 AI 直接定位。

## 3. 辅助模型先提取失败点

Prompt：

```text
请从这段 CI 日志中提取真正导致失败的步骤。
只输出：
- job
- step
- exit code
- 关键错误行
- 可能相关文件
不要给修复建议。
```

这个任务不需要 Fable 5。

低成本模型就够。

把提取结果交给 Fable 5，再问根因。

## 4. 常见失败一：依赖安装

典型错误：

```text
npm ci failed
lockfile out of sync
pnpm install --frozen-lockfile failed
No matching version found
```

常见根因：

```text
package.json 改了但 lockfile 没提交。
Node 版本不一致。
私有包 token 缺失。
缓存污染。
registry 不可达。
```

AI 判断时要看：

```text
package manager。
lockfile。
Node 版本。
workflow 里的 setup-node。
```

不要让 AI 建议删除 lockfile 作为第一步。

## 5. 常见失败二：测试只在 CI 挂

本地过，CI 挂。

常见原因：

```text
时区不同。
环境变量缺失。
文件路径大小写。
测试依赖顺序。
数据库服务没启动。
浏览器依赖缺失。
```

Prompt：

```text
请判断为什么测试只在 CI 失败、本地不失败。
重点看时区、环境变量、路径大小写、服务依赖和缓存。
不要建议直接跳过测试。
```

这个很适合 Fable 5。

因为需要跨环境判断。

## 6. 常见失败三：Secrets 缺失

日志可能写：

```text
401 Unauthorized
permission denied
invalid token
environment variable is not set
```

这里不要把 secret 发给 AI。

只告诉它：

```text
变量名是否存在。
在哪个环境配置。
当前 job 是否有权限读。
是 PR from fork 还是主仓库。
```

GitHub Actions 里，fork PR 默认拿不到某些 secrets。

AI 很容易漏掉这个点。

可以明确提示：

```text
请检查这个失败是否可能来自 fork PR 无法读取 secrets。
```

## 7. 常见失败四：部署权限

部署阶段失败可能是：

```text
云厂商 token 过期。
Docker registry 登录失败。
SSH key 权限不对。
服务器 known_hosts 缺失。
目标目录权限不对。
```

AI 可以帮你整理权限链路。

但不要让 AI 直接生成新的生产密钥。

让它给：

```text
只读验证。
权限最小化建议。
需要人工确认的密钥轮换步骤。
```

## 8. workflow yaml 要一起给 AI

CI 日志只告诉你哪里失败。

workflow yaml 才告诉你为什么这么跑。

给 AI 时至少贴这几块：

```text
on 触发条件。
jobs 名称。
runs-on。
setup-node / setup-python / setup-go。
cache 配置。
install 命令。
test 命令。
build 命令。
deploy 命令。
permissions。
environment。
```

很多 CI 问题不在代码。

而在 workflow 写法。

例如：

```text
CI 用 Node 18，本地用 Node 20。
cache key 没包含 lockfile。
PR from fork 读不到 secrets。
permissions 没开 id-token。
deploy job 没限制 main 分支。
```

Prompt：

```text
请同时审查失败日志和 workflow yaml。
判断这是代码问题、依赖问题、环境差异、缓存问题、权限问题还是部署流程问题。
不要只根据最后一行 error 下结论。
```

这类跨日志和配置的判断，Fable 5 比低成本模型更稳。

## 9. 偶发失败要单独处理

有些 CI 不是每次都失败。

而是：

```text
十次失败一次。
重跑又好了。
只在某个 runner 上失败。
只在高并发时失败。
```

这种叫 flaky。

AI 排查 flaky test，要给它：

```text
失败频率。
失败测试名。
最近 5 次失败日志。
是否并行执行。
是否依赖时间、网络、随机数、数据库顺序。
```

常见根因：

```text
测试共享状态。
异步没有 await。
时间没有 mock。
随机数据冲突。
端口冲突。
数据库未隔离。
浏览器测试等待条件不稳定。
```

Prompt：

```text
请判断这个 CI 失败是否属于 flaky test。
如果是，请找出共享状态、异步等待、时间依赖、随机数、端口和数据库隔离方面的证据。
不要建议简单提高 timeout，除非先证明是合理等待不足。
```

很多团队处理 flaky 的坏习惯是：

```text
失败就重跑。
```

短期可以。

长期会让 CI 失去可信度。

AI 更适合帮你把 flaky 的共同模式找出来。

## 10. 4SAPI 分工

| 阶段 | 模型 |
| --- | --- |
| 日志压缩 | 低成本模型 |
| 根因判断 | Fable 5 |
| workflow 修改建议 | Fable 5 |
| PR 描述和复盘 | 低成本模型 |

4SAPI 记录：

```text
error_type: cicd_failed
workflow
job
step
model
cost
fix_pr
```

企业团队可以把 CI 失败自动发给辅助模型，生成工单摘要。

真正复杂的再升级 Fable 5。

## 11. AI Prompt

```text
你是 CI/CD 排查助手。

请根据 workflow yaml、失败 step、日志最后 100 行和最近变更，判断根因。

重点检查：
1. 依赖和 lockfile
2. 语言版本
3. 环境变量和 secrets
4. 缓存
5. 测试环境差异
6. Docker build
7. 部署权限

要求：
- 不建议直接跳过测试。
- 不输出 secrets。
- 每个结论引用日志证据。
- 给最小修复建议和验证命令。
```

## 12. 总结

CI/CD 失败不要从头肉眼翻日志。

用 AI 的正确方式：

```text
低成本模型提取失败点。
Fable 5 判断根因。
复核模型检查修复风险。
4SAPI 记录成本和结果。
```

最重要的是：

```text
不要用“跳过测试”解决测试失败。
```

CI 是你上线前的刹车。

AI 应该帮你修刹车，不是帮你拆刹车。
