---
title: "AGPL 的商用边界：从 Macro 的许可变更看自托管与闭源改造"
category: 人工智能
tags:
  - 开源许可证
  - AGPL
  - 自托管
description: "一个项目在三个月内从 BSL 换成 AGPLv3，商用到底受什么约束？本文用 Macro 的许可变更作样本，逐句读第 13 条网络交互义务、区分修改与使用、说明对应源码包含什么，再对照它自己承认的自托管门槛（第三方组件授权、移动端审批、SOC 2 在托管版）与贡献者协议，给出一份选型前的自查清单。不构成法律意见。"
brand_mode: none
---

# AGPL 的商用边界：从 Macro 的许可变更看自托管与闭源改造

2026 年 3 月 2 日，仓库里放进了 Business Source License 1.1；5 月 31 日，同一位置换成 AGPLv3 全文。不到三个月。

Macro 就是这么变的。当前 LICENSE.txt 是标准 AGPLv3，约 5,535 词；README 里写得很直接（译文）："Macro 是完整的开源，不是 open core，采用 GNU AGPL v3.0。"

换许可证意味着两个具体动作要重新算：能不能架起来提供服务？改完要不要开源？答案不一样，也都不在"能商用吗"这个笼统问题里。

下面把条文和项目说法摆出来逐段看。**本文只做条文复述与事实整理，不构成法律意见；你的业务是否构成分发或需要开源，请找律师。英文引文附原文。**

## 一、从 BSL 到 AGPL 这一步

BSL 常见结构是：源码公开，但限制生产环境的商业使用，按版本定转换日期，到期自动转为某个开源许可。它给项目方留了商业窗口，给使用方留下"哪天之后我能干什么"。

这个仓库 3 月到 5 月的 LICENSE.md 就是 BSL 1.1。改成 AGPLv3 之后，时间条件消失了。项目在 FAQ 里的解释是（译文）：

> AGPL 是 copyleft 许可证，衍生作品也需要开源。也就是说，如果你在 Macro 之上构建，你需要按 AGPLv3 保持代码开源。如果你想在其上构建而不使用 AGPLv3（即闭源或不兼容的许可证），需要从我们这里获得另一份许可。

这段话是骨架：路径两条，走 AGPL 就开源，或者谈另一份许可。

## 二、第 13 条：网络交互义务

很多人的印象是"云服务也要开源"，来源是第 13 条。原文：

> 13. Remote Network Interaction; Use with the GNU General Public License. Notwithstanding any other provision of this License, if you modify the Program, your modified version must prominently offer all users interacting with it remotely through a computer network (if your version supports such interaction) an opportunity to receive the Corresponding Source of your version by providing access to the Corresponding Source from a network server at no charge, through some standard or customary means of facilitating copying of software. This Corresponding Source shall include the Corresponding Source for any work covered by version 3 of the GNU General Public License that is incorporated pursuant to the following paragraph.

标题容易被误读：它不全是新增义务。第二段讲的是**许可**——允许把 AGPL 作品与 GPLv3 作品链接或合并成一个作品再传递，AGPL 条款继续适用于原本受其覆盖的部分。

第一段才是义务，结构是"两个前提 + 一条后果"：

1. 前提一：你**修改**了这个程序。
2. 前提二：你的版本支持通过网络被远程交互。
3. 后果：你必须向这些远程用户**显著地**提供获取"你这个版本对应源码"的机会，方式是从网络服务器免费提供。

第一个前提最容易被含糊过去。改配置、换主题、加一条自己的业务逻辑、把登录换成公司的 SSO——哪些算修改，哪些算部署，不是靠感觉定的。第二个前提也不像看上去那么简单：条文写的是 interacting with it remotely through a computer network，没有"对外"或"第三方"的限定。

我不在这里下结论，这也不是文本能解决的事。能确定的是：**把 AGPL 当成"内部用就随便"是一种赌，赌注是自己公司的代码。** 真要押，先问一次律师比事后补救便宜。

## 三、"对应源码"到底要给什么

真走到给源码那一步，给到什么程度？协议有定义：

> The "Corresponding Source" for a work in object code form means all the source code needed to generate, install, and (for an executable work) run the object code and to modify the work, including scripts to control those activities. … For example, Corresponding Source includes interface definition files associated with source files for the work, and the source code for shared libraries and dynamically linked subprograms that the work is specifically designed to require…

三个动词加一个补充：生成、安装、运行（条文限定为可执行作品），以及修改所需的源码与控制这些活动的脚本。例子有两个：接口定义文件，以及作品特意依赖的共享库与动态链接子程序。

排除项原文是：不含该作品的系统库、通用工具，或在上述活动中原样使用、属于普遍可获得的自由程序但不属于该作品的部分。另有一句常被忽略：**用户可以自己从对应源码的其他部分自动再生成的内容，不必包含在内。**

实践里麻烦的往往不是源码目录，而是构建脚本、数据库迁移、部署配置——它们属于"安装与运行所需"。这个仓库把运行与构建脚本放在仓库里（根部 justfile 有建网络、起数据库、初始化数据库这些目标），自托管方交付时反而容易齐全；换成把部署脚本留在内部的项目，就得专门去要。

## 四、双许可：那条要谈的路

README 只写了两句：想以其他许可证在其上构建，联系 licensing 邮箱；托管或商业安排，联系 self-host 邮箱。没有价目表，也没有条款。

按一般双许可实践（**这一句不是这个项目的说法**）：

- 条款是谈出来的，各家不同，通常带范围限制：用户数、部署形态、能否再分发。
- 只部署、不改代码的情形通常不需要商业许可，因为义务的触发条件没成立。
- 改得越深、越靠近核心，越可能需要另一份许可。

最常见的误判是以为买了商业许可就一劳永逸。签之前确认三件事：范围、期限、终止条件。

## 五、自托管的真实门槛：项目自己承认的部分

许可允许自托管，不等于能跑起来。FAQ 在这点上比较诚实，把说法整理成表：

| 事项 | 官方说法 |
| --- | --- |
| 能不能自托管 | 可以，按 AGPL 许可 |
| 是不是重点 | 截至 2026 年 6 月不是重点；当前重点是把 Macro 做好，希望今年晚些时候转向改善自托管 |
| 移动端 | 托管版才支持 iOS / Android 应用 |
| 集成 | 托管版具备接入 PR 所需的 Apple、Google Mail、GitHub 批准 |
| 第三方组件 | sublicense 了 LiveKit（视频）、FusionAuth（认证）、PostHog（分析）等；自托管方需自行维护这些许可，或禁用 |
| 合规 | 托管版在 AWS 上运行并维持 SOC 2 Type II 审计 |
| 特殊客户 | 需要自有基础设施（如 FedRAMP 场景）可联系其 self-host 邮箱 |
| 数据驻留 | EU 驻留在托管版也可选，自托管不是前置条件 |

第三、四行重点。

FAQ 只说"托管版具备那些批准"，**没说自托管方必须重新申请**。反推下来，自托管方大概要自己走一遍平台审批——这是推断，不是原文。

第五行更实际。视频、认证、分析这些能力是**转授权**来的，自托管时不会跟着代码一起过来。要么自己谈，要么关掉。关掉意味着通话、登录方式、统计面板可能直接不可用——这不是配置项，是功能缺失。

把这几行放在一起看：**AGPL 给的是法律许可，不是工程可运行。** 这两件事差着几个月的活。

## 六、贡献者协议：另一侧的约束

打算回馈代码，还有一道手续。贡献文档写着：组织外的 PR 要先签 Macro CLA，一次性签名、覆盖之后的贡献；每个 PR 会跑名为 `cla` 的必过检查，未签会变红，项目方说明这不阻塞 review；贡献按 AGPLv3 授权。

CLA 不是 AGPL 的要求，是项目方的安排，作用通常是保留再许可的能力——也就是双许可能成立的前提。签之前想一件事：你贡献的代码之后可能被以商业许可卖给别的公司。可接受，但要知情。

## 七、选型自查清单

如果明天要引入一个 AGPL 组件，我会按这张表逐条过，每一项都能在一次会议里问完。

| # | 问题 | 为什么问 |
| --- | --- | --- |
| 1 | 我们只是部署，还是会改代码？ | 第 13 条的两个前提里，第一个是"修改" |
| 2 | 改动是配置级，还是逻辑级？ | 边界模糊时按"可能算修改"准备 |
| 3 | 这个服务给谁用？ | 条文没限定"对外"，别自己加限定 |
| 4 | 如果需要开源，我们愿意开哪部分？ | 提前划边界，比事后争论便宜 |
| 5 | 有没有商业许可可谈，范围与期限如何？ | 有些场景谈许可比改架构便宜 |
| 6 | 自托管依赖哪些转授权组件？ | 决定功能能不能齐 |
| 7 | 移动端与第三方集成要不要过平台审批？ | 决定交付周期 |
| 8 | 谁负责跟进上游许可证变更？ | 三个月换一次许可这种事会再发生 |

第 8 行是私心。这个仓库从 BSL 1.1 到 AGPLv3 只隔了不到三个月；下个项目也可能反过来收紧。**把许可证当成会变的依赖项，不是签一次就完的合同。**

## 八、四个常见误读

有四句话在讨论里反复出现，都不是没道理，但每句都少了个条件。

**"AGPL 不能商用。"** 按协议文本，商业使用不在禁止之列，也没有营收门槛；它约束的是源码义务。

**"只要不发布就没事。"** GPL 时代这句大体成立，触发点是"传递"（conveying）。第 13 条把它扩到通过网络远程交互，正是云服务场景。"只是放服务器上给人用"在 AGPL 下不足以自证安全。

**"用 Docker 跑就不算修改。"** 装和改是两件事，但"装"里有多少配置算修改，取决于具体情形。换主题、改默认值、写插件、把认证接到公司 SSO，实务里都可能被争论。跑容器不是护身符。

**"内部系统不算。"** 条文写的是 remotely through a computer network，没限定外部用户。员工算不算远程用户，需要解释——而解释的人不该是你。

四句都在把有条件的话压成结论，而条文恰恰是条件句组成的。

## 九、结论

两个问题再答。

**能不能商用？** 条文里没有禁止商业使用的条款；被约束的是修改后通过网络提供服务时的源码义务——**这里没有"对外"两个字**，别自己加。

**改完要不要开源？** 取决于你是否修改、服务是否通过网络被交互。都是事实判断，不是态度问题。判断不了找律师，别在群里问。

这个样本把两条路都摆明了：走 AGPL 就开源，要闭源就谈许可；自托管法律上可以，工程上要自己补第三方授权与平台审批。这比一句"完全开源"信息量大。

限制：条文与引文来自 LICENSE.txt 全文，项目说法来自 README、CONTRIBUTING.md 与官方 FAQ（检索日期 2026-11-06，英文引文为原文，中文为译文）；没有部署，也没就任何具体业务做过法律分析；FAQ 的时间点与状态属项目自述，会变；SOC 2 与 ISO 徽章未核实。

如果只带走一句：**先确定要动它哪一步，再读许可证。** 顺序反了，读十遍也读不出结论。

## 十、来源

- 许可证全文：仓库 `LICENSE.txt`（https://github.com/macro-inc/macro/blob/main/LICENSE.txt）
- 项目协议与自托管说明：https://docs.macro.com/faq
- 贡献者协议：仓库 `CONTRIBUTING.md`（https://github.com/macro-inc/macro/blob/main/CONTRIBUTING.md）
- 许可变更时间线：仓库对 `LICENSE.md` / `LICENSE.txt` 的提交记录（2026-03-02 加入 BSL 1.1，2026-05-31 换成 AGPLv3）
