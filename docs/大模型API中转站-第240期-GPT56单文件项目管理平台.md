---
title: "单文件项目管理应用的状态闭环设计"
category: 人工智能
tags:
  - HTML
  - JavaScript
  - 状态管理
description: "看板、统计与操作历史若各自维护数据，界面很快会互相矛盾。本文说明单一状态源、派生数据、localStorage 恢复和组合筛选的实现与验收。"
---

# 单文件项目管理应用的状态闭环设计

单文件项目管理应用可以快速展示看板、筛选和统计，但每个区域若使用独立的硬编码数据，拖动任务后项目进度、历史和刷新结果就会互相矛盾。本文只解决状态闭环：用一个 `state` 保存事实数据，让统计和进度从任务状态派生；每次有效操作按“更新事实、记录历史、持久化、重绘”执行；最后通过拖拽、组合筛选、刷新恢复和损坏数据测试验证。读者可直接使用文中的状态结构、函数示例和验收任务书。

界面是否完整，关键问题是：

```text
这些区域是不是同一个系统？
```

这是一道适合用于检查 Coding Agent 交付质量的验收题：

```text
模型能否把多个页面做成共享同一状态的产品，而不是一组互不相干的界面。
```

## 1. 先明确单文件任务边界

任务应同时限定：

```text
单个 HTML 文件。
纯原生 HTML、CSS 和 JavaScript。
完全离线。
内置演示数据。
禁止第三方库、外部接口和外部素材。
要求自主设计、开发、测试和修复。
```

这类约束有两个好处。

第一，结果容易检查。在目标浏览器中直接打开文件，不需要安装依赖。

第二，验收边界清楚。任何 CDN、远程字体、网络请求或缺失资源都算失败。

但它也牺牲了模块化、多人协作和长期维护。

所以单文件适合作为模型能力测试、离线工具和产品原型，不应直接当作所有生产后台的默认架构。

## 2. 真正的核心：一个状态源

一个项目管理平台至少有这些实体：

```text
Project
Task
Member
Activity
Settings
```

任务不能在看板里是一份数据，在统计页又是另一份硬编码数据。

推荐状态结构：

```javascript
const state = {
  projects: [],
  tasks: [],
  members: [],
  activities: [],
  filters: {
    query: "",
    projectId: "all",
    assigneeId: "all",
    priority: "all"
  },
  settings: {
    theme: "light"
  }
};
```

所有页面都从 `state` 读取。

所有操作都先修改 `state`，再重新计算和渲染。

## 3. 一次拖拽应该触发什么

把任务从“进行中”拖到“已完成”，至少触发：

```text
1. 更新 task.status。
2. 写入 task.updatedAt。
3. 新增一条 activity。
4. 重新计算所属项目进度。
5. 重新计算全局统计。
6. 保存到 localStorage。
7. 更新当前视图。
```

如果只移动 DOM 卡片，刷新后它会回到原位置。

如果只改任务数据，却没有更新统计，多个视图会相互矛盾。

验收时不要只看卡片能不能拖。

拖完后立刻检查：

```text
看板列数量。
项目完成百分比。
首页统计。
操作历史。
刷新后的状态。
```

## 4. 派生数据不要重复保存

项目进度可以从任务状态计算：

```javascript
function getProjectProgress(projectId) {
  const tasks = state.tasks.filter(task => task.projectId === projectId);
  if (tasks.length === 0) return 0;

  const completed = tasks.filter(task => task.status === "done").length;
  return Math.round((completed / tasks.length) * 100);
}
```

不建议同时保存：

```text
task.status = done
project.progress = 75
dashboard.completed = 12
```

三份值需要同步，任何一步漏掉都会不一致。

更稳的方式是只保存事实状态，把进度和统计视为派生数据。

## 5. localStorage 持久化怎么做

最小流程：

```text
初始化：尝试读取本地数据。
读取失败：加载演示数据。
每次有效操作：序列化并保存。
结构升级：检查 schemaVersion。
用户重置：确认后恢复演示数据。
```

数据包可以包含版本号：

```javascript
const payload = {
  schemaVersion: 1,
  savedAt: new Date().toISOString(),
  data: {
    projects: state.projects,
    tasks: state.tasks,
    members: state.members,
    activities: state.activities,
    settings: state.settings
  }
};

localStorage.setItem("ai-project-manager", JSON.stringify(payload));
```

需要处理解析失败：

```javascript
function isRecord(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function isRecordArray(value) {
  return Array.isArray(value) && value.every(isRecord);
}

function hasSavedDataShape(data) {
  return (
    isRecord(data) &&
    isRecordArray(data.projects) &&
    isRecordArray(data.tasks) &&
    isRecordArray(data.members) &&
    isRecordArray(data.activities) &&
    isRecord(data.settings)
  );
}

function loadSavedData() {
  try {
    const raw = localStorage.getItem("ai-project-manager");
    if (!raw) return null;
    const payload = JSON.parse(raw);
    if (
      !isRecord(payload) ||
      payload.schemaVersion !== 1 ||
      !hasSavedDataShape(payload.data)
    ) {
      return null;
    }
    return payload.data;
  } catch (error) {
    console.error("Failed to load saved data", error);
    return null;
  }
}
```

`schemaVersion` 只说明版本，不能证明 `data` 可渲染。上面的校验还要求四个集合都是由对象组成的数组，`settings` 是对象；无法解析、版本不符或顶层结构损坏时返回 `null`，由初始化流程加载演示数据。

这仍只是容器级校验。如果界面依赖任务的 `id`、`status` 等具体字段，还要为每类记录定义字段校验或版本迁移，不能把任意损坏数据都视为安全输入。

## 6. 操作历史不是控制台日志

用户需要读懂操作历史。

一条记录至少包含：

```text
谁执行。
什么时候执行。
操作对象。
发生了什么变化。
```

例如：

```text
10:32 陈默将“API 权限验收”从“进行中”移动到“已完成”。
```

而不是：

```text
update task #t-18 status=done
```

单文件离线应用没有真实登录系统，可以明确标记“当前演示用户”，不要伪装成企业级身份审计。

## 7. 搜索和筛选必须组合

用户可能同时选择：

```text
项目 A。
高优先级。
负责人小林。
关键词 API。
```

筛选函数应该一次应用全部条件：

```javascript
function getVisibleTasks() {
  const query = state.filters.query.trim().toLowerCase();

  return state.tasks.filter(task => {
    const matchesQuery = !query ||
      task.title.toLowerCase().includes(query) ||
      task.description.toLowerCase().includes(query);

    const matchesProject = state.filters.projectId === "all" ||
      task.projectId === state.filters.projectId;

    const matchesAssignee = state.filters.assigneeId === "all" ||
      task.assigneeId === state.filters.assigneeId;

    const matchesPriority = state.filters.priority === "all" ||
      task.priority === state.filters.priority;

    return matchesQuery && matchesProject && matchesAssignee && matchesPriority;
  });
}
```

验收时要测试单条件、组合条件和清空筛选。

## 8. 原生拖拽还要考虑移动端

HTML Drag and Drop API 在桌面端方便，但触屏体验不一定一致。

至少提供一种移动端替代：

```text
任务编辑弹窗中修改状态。
卡片菜单选择移动到某列。
触摸手势实现拖拽。
```

不能因为桌面鼠标能拖，就宣布响应式已经完成。

## 9. 深色模式不只是换背景

需要检查：

```text
文字对比度。
边框可见性。
图表颜色。
输入框和弹窗。
悬停、选中和禁用状态。
浏览器原生控件。
```

用 CSS 变量统一管理：

```css
:root {
  --bg: #f4f6f8;
  --surface: #ffffff;
  --text: #17202a;
  --muted: #667085;
  --border: #d9dee5;
  --accent: #176b5b;
}

[data-theme="dark"] {
  --bg: #101417;
  --surface: #181e22;
  --text: #eef2f3;
  --muted: #a9b3b8;
  --border: #354047;
  --accent: #52b69a;
}
```

## 10. 一份可复制的完整任务书

```text
请在一个 HTML 文件中，使用纯原生 HTML、CSS 和 JavaScript，从零开发一个可离线运行的项目管理平台。

必须包含：
- 项目总览
- 任务看板
- 新建、编辑、删除任务
- 桌面端原生拖拽
- 移动端修改状态的替代入口
- 项目、负责人、优先级和关键词组合筛选
- 项目进度与任务统计
- 操作历史
- 深色模式
- localStorage 持久化
- 导出和恢复演示数据
- 响应式布局

数据要求：
1. 所有页面使用同一份状态数据。
2. 进度和统计必须从任务数据派生，不得硬编码。
3. 每次有效操作写入历史并保存。
4. 本地数据损坏时回退到安全初始状态，不得白屏。

技术限制：
- 禁止第三方库、外部接口、CDN、远程字体和外部素材。
- 不允许页面产生任何网络请求。
- 禁止使用 eval 和内联字符串事件处理器。

验证要求：
- 实际运行并完成新建、编辑、拖拽、筛选、刷新恢复和重置测试。
- 检查窄屏布局、键盘操作和深色模式。
- 修复发现的问题。

最终只交付一个可直接打开的 HTML 文件，并说明已验证内容、已知限制和数据存储位置。
```

## 11. 浏览器验收脚本

### 数据闭环

```text
1. 新建项目。
2. 新建三条任务。
3. 拖动一条到已完成。
4. 检查项目进度和统计。
5. 刷新页面。
6. 确认状态与历史仍然存在。
```

### 筛选闭环

```text
1. 同时选择项目、负责人和优先级。
2. 输入关键词。
3. 确认结果满足所有条件。
4. 清空筛选。
5. 确认全部任务恢复。
```

### 异常闭环

```text
1. 尝试提交空标题。
2. 删除当前筛选中的任务。
3. 手工破坏 localStorage 数据后刷新。
4. 重复点击保存。
5. 在窄屏下完成状态修改。
```

## 12. 单文件项目的生产边界

这个架构不包含：

```text
真实账号与权限。
多人实时协作。
服务端数据库。
备份与恢复。
跨设备同步。
企业审计日志。
大规模数据查询。
```

如果要上线，至少要拆分：

```text
前端组件。
状态管理。
API 服务。
数据库。
身份认证。
权限策略。
审计与监控。
自动化测试。
```

一句话生成的单文件应用可以验证产品闭环，不能替代生产架构评审。

## 13. 验收清单

```text
[ ] 单个 HTML 文件可离线打开
[ ] 页面没有第三方网络请求
[ ] 所有视图共享同一份任务数据
[ ] 统计和进度由事实数据派生
[ ] 拖拽后状态、统计、历史同步更新
[ ] 刷新后数据恢复
[ ] 无法解析、版本不符或顶层结构损坏时回退到演示数据
[ ] 搜索和多个筛选条件可以组合
[ ] 移动端有不依赖鼠标拖拽的操作方式
[ ] 深色模式覆盖弹窗、输入和交互状态
[ ] 不包含真实密钥或敏感数据
[ ] 已明确单机原型与生产系统的边界
```

## 14. 结论与限制

项目管理平台是否完成，不能看卡片多不多。

要看用户一次操作能否沿着完整状态链传播：

```text
任务变化。
项目进度变化。
统计变化。
历史变化。
本地数据变化。
```

一个可交付的单文件原型必须让任务变化沿同一状态链传播，并在刷新、异常输入和窄屏操作后保持可解释。界面数量、代码行数和文件大小都不能替代这些运行检查。

本文结构只适用于单机、离线原型。它不包含真实身份、多人协作、服务端数据库、跨设备同步、备份和审计能力；进入生产前需要重新设计模块边界、服务端契约、权限、安全、监控与自动化测试。
