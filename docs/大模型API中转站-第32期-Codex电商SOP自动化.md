---
title: "【大模型API中转站】第32期 Codex电商SOP | 从表格到自动化"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - 电商自动化
  - SOP
  - 4SAPI
description: "用一个最小可落地的电商 SOP 仓库，演示 Codex 如何从表格读取、生成运营报告、输出上架草稿、风险清单和验收记录，并通过 4SAPI 管理模型调用。"
---

# 【大模型API中转站】第32期 Codex电商SOP | 从表格到自动化

本文是【大模型API中转站】系列的第32篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇讲的是思路：电商不是让 Codex 替你开店，而是把运营数据流整理起来。

这一篇更偏实操：假设你有一个 2-5 人的小电商团队，日常有商品表、订单表、广告表、库存表、客服记录，希望 Codex 每天帮你生成运营日报、上架草稿、库存预警和利润复盘，应该怎么搭第一套 SOP。

这套方案不追求一上来全自动。目标只有一个：让 Codex 先稳定做“只读分析”和“草稿生成”，人工确认后再执行真实业务动作。

## 1. 为什么先做 SOP，而不是先做系统

很多团队一提 AI 自动化，就想直接做一个后台：

```text
自动抓竞品
自动写标题
自动调价格
自动投广告
自动回复客服
```

听起来很完整，风险也很完整。

对电商来说，价格、广告、售后、发货、财务都直接影响收入和责任。没有稳定规则前，越自动越容易出问题。更务实的顺序应该是：

```text
表格标准化
  -> Codex 只读分析
  -> 人工复核
  -> 半自动生成草稿
  -> 小范围自动提醒
  -> 再考虑系统化
```

4SAPI 这类大模型API中转站适合放在这个阶段的模型入口：先把 Key、模型、日志、额度统一起来，避免团队成员各自用不同账号、不同模型、不同提示词，最后没人知道成本花在哪里。

## 2. 建一个最小仓库

建议把电商运营资料放进一个普通 Git 仓库。不是为了让运营学复杂开发，而是为了保留文件版本、提示词版本和报告记录。

目录结构：

```text
ecommerce-codex-sop/
  AGENTS.md
  README.md
  data/
    products.csv
    inventory.csv
    orders.csv
    ads.csv
    competitors.csv
    customer_service.csv
  prompts/
    daily_report.md
    listing_generator.md
    ad_review.md
    inventory_warning.md
    profit_review.md
  reports/
  scripts/
    validate_csv.py
    summarize_inputs.py
```

每个目录只做一件事：

| 目录 | 作用 |
| --- | --- |
| data | 放原始导出表格或脱敏后的运营数据 |
| prompts | 固定提示词模板 |
| reports | 放 Codex 生成的日报、周报、复盘 |
| scripts | 放简单校验脚本，不做危险动作 |
| AGENTS.md | 固定 Codex 的工作边界 |

如果团队没有开发基础，`scripts` 目录可以先不做。只要 `data`、`prompts`、`reports` 和 `AGENTS.md` 存在，就能跑第一版。

### 2.1 每个文件都要有“责任人”

SOP 仓库不是资料堆放处。建议给每类文件标清楚责任人：

| 文件 | 维护人 | 更新频率 | 说明 |
| --- | --- | --- | --- |
| products.csv | 商品运营 | 新品或改价时 | 商品基础信息和风险字段 |
| ads.csv | 投流负责人 | 每日 | 广告计划、素材和转化数据 |
| inventory.csv | 仓库或运营 | 每日 | 库存、安全线、在途库存 |
| customer_service.csv | 客服负责人 | 每周或每日 | 脱敏后的问题记录 |
| profit.csv | 财务或店主 | 每周 | 成本、佣金、退款和利润 |
| AGENTS.md | 店主或负责人 | 规则变化时 | Codex 工作边界 |
| prompts/*.md | 运营负责人 | 复盘后迭代 | 固定提示词模板 |

谁维护、多久更新、能不能被 Codex 修改，都要写清楚。否则自动化跑起来以后，最常见的问题不是模型错，而是数据已经过期。

## 3. 先写 AGENTS.md，把边界钉住

Codex 做电商运营，最怕的不是它不会写，而是它写得太像“可以直接发布”。所以第一步不是写提示词，而是写规则。

可以在仓库根目录创建：

```markdown
# 电商运营 Codex 工作规则

## 数据处理

1. 默认只读取 data/ 目录，不直接修改原始数据。
2. 如需生成结果，写入 reports/ 目录。
3. 发现字段缺失时先在报告中说明，不要猜测补全。

## 内容边界

1. 商品标题、主图文案、详情页内容必须避免绝对化、夸大和无法证明的表述。
2. 医疗、保健、功效、材质、认证、适用人群等内容必须标注“人工复核”。
3. 售后、退款、发货承诺必须以平台规则和店铺政策为准。

## 隐私与合规

1. 不输出用户姓名、手机号、地址、订单号等敏感信息。
2. 客服记录只做分类和摘要。
3. 不编写绕过平台限制、批量骚扰用户或规避风控的方案。

## 经营决策

1. Codex 可以给建议，但不直接决定选品、定价、预算和赔付。
2. 所有涉及花钱、发布、承诺和下架的动作，都必须人工确认。
```

这份文件的价值是让 Codex 每次进仓库都先读规则。提示词可以变化，但底线不应该每次重新说。

还可以加一段“输出格式约束”：

```markdown
## 输出格式

1. 所有报告必须包含“数据来源”“关键发现”“风险清单”“建议动作”“人工复核”。
2. 表格中的建议动作只写建议，不写“已执行”。
3. 如果数据缺失，必须列出缺失字段和影响。
4. 如果样本不足，必须标注“样本不足，不建议据此决策”。
5. 不允许编造数据、链接、销量、评论或平台政策。
```

这段很实用。它会逼着 Codex 把“我根据什么得出结论”写出来，方便人复核。

## 4. 商品表：让上架内容从同一份底稿生成

先准备 `data/products.csv`：

```csv
product_id,product_name,category,core_keywords,target_user,material,specs,cost_price,sale_price,risk_notes
P001,折叠收纳箱,家居收纳,"收纳箱;折叠;车载","租房人群;车主",PP,"30L;50L",28,59,"承重表述需复核"
P002,便携筋膜球,运动放松,"筋膜球;便携;放松","久坐办公;运动人群",硅胶,"单只装;双只装",12,39,"避免医疗承诺"
```

提示词模板 `prompts/listing_generator.md`：

```markdown
请根据 data/products.csv 生成商品上架草稿。

输出到 reports/listing-draft.md，包含：
1. 每个商品 3 个标题
2. 5 条主图卖点
3. 商品属性表
4. 详情页结构
5. FAQ
6. 需要人工复核的风险点

要求：
- 不使用绝对化词语
- 不夸大功效
- 不承诺无法保证的结果
- 保留核心关键词
- 用表格输出
```

你给 Codex 的任务可以很短：

```text
请读取 prompts/listing_generator.md，按规则生成上架草稿。
```

这样做的好处是：运营同事以后只改模板，不需要每次重新写一大段提示词。

### 4.1 上架草稿要分“可发布”和“需确认”

建议让 Codex 输出两张表。

第一张是可发布草稿：

| product_id | 标题 | 主图卖点 | 详情页模块 |
| --- | --- | --- | --- |

第二张是需确认事项：

| product_id | 风险字段 | 原始内容 | 为什么要复核 | 建议找谁确认 |
| --- | --- | --- | --- | --- |

比如“承重 100kg”“食品级”“适合儿童”“缓解疼痛”这类内容，都不应该直接进入发布稿。它们可以出现在“需确认事项”里，让运营去找供应商、质检或平台规则确认。

## 5. 广告表：让复盘关注“下一步测试”

`data/ads.csv` 可以这样设计：

```csv
date,campaign,creative_id,product_id,spend,impressions,clicks,orders,revenue,refund_amount
2026-06-16,收纳箱测试,A01,P001,300,12000,420,18,1062,0
2026-06-16,收纳箱测试,A02,P001,260,9000,180,6,354,0
2026-06-16,筋膜球测试,B01,P002,180,8000,300,5,195,39
```

复盘提示词：

```markdown
请分析 data/ads.csv 并生成 reports/ad-review.md。

请计算：
1. CTR = clicks / impressions
2. CVR = orders / clicks
3. CPA = spend / orders
4. ROI = revenue / spend

请输出：
1. 点击率高但转化低的素材
2. ROI低但消耗高的计划
3. 可能值得加预算的素材
4. 下一轮测试建议
5. 需要停止、降预算或重写素材的对象

注意：
不要只看点击率。
如果订单数太少，请标注样本不足。
```

投流数据最怕“看起来有结论”。样本太小、退款未回流、归因延迟，都可能让复盘偏掉。所以模板里要明确要求标注样本不足。

### 5.1 加一列素材假设

广告表如果只有数字，复盘会很机械。建议加一张 `creative_notes.csv`：

```csv
creative_id,hypothesis,hook,main_selling_point,target_user
A01,"强调大容量能提升点击","小房间也能多收纳","折叠大容量","租房人群"
A02,"强调车载场景能提升转化","后备箱不再乱","车载收纳","车主"
B01,"强调久坐放松能提升点击","下班后放松一下","便携放松","办公人群"
```

然后让 Codex 把素材假设和投放结果合并分析：

```markdown
请结合 ads.csv 和 creative_notes.csv 分析素材测试结果。

要求：
1. 判断每个素材假设是否被数据支持
2. 标出点击率高但假设未必成立的素材
3. 给出下一轮要保留、修改、放弃的假设
4. 样本不足时不要下结论
```

这会让投流复盘从“哪个素材数据好”升级为“哪个假设值得继续测试”。

## 6. 库存表：用规则找风险，不让模型自由发挥

`data/inventory.csv`：

```csv
sku,product_id,warehouse_stock,safety_stock,avg_daily_sales,in_transit_stock,last_restock_date
P001-30L,P001,35,30,8,100,2026-06-10
P001-50L,P001,12,25,6,0,2026-06-08
P002-1PC,P002,80,50,12,0,2026-06-12
```

库存预警提示词：

```markdown
请分析 data/inventory.csv，生成 reports/inventory-warning.md。

规则：
1. warehouse_stock < safety_stock，标为库存低于安全线
2. warehouse_stock / avg_daily_sales < 7，标为 7 天内可能缺货
3. in_transit_stock = 0 且库存风险高，标为需要人工确认补货
4. avg_daily_sales 为空时，不做缺货天数判断

输出：
- 风险 SKU
- 风险原因
- 建议动作
- 需要人工确认的问题
```

库存预警不需要模型猜，它更像一个规则引擎。Codex 的作用是把表格读懂、把异常讲清楚、把报告写出来。

### 6.1 先用脚本算，再让模型解释

库存、利润、CTR、ROI 这类计算最好先用脚本完成，再让模型解释。比如 `scripts/validate_csv.py` 可以只做最基础的字段检查：

```python
import csv
from pathlib import Path

REQUIRED_COLUMNS = {
    "data/ads.csv": {"date", "campaign", "creative_id", "product_id", "spend", "impressions", "clicks", "orders", "revenue"},
    "data/inventory.csv": {"sku", "product_id", "warehouse_stock", "safety_stock", "avg_daily_sales"},
    "data/products.csv": {"product_id", "product_name", "core_keywords", "risk_notes"},
}

for file_name, required in REQUIRED_COLUMNS.items():
    path = Path(file_name)
    if not path.exists():
        print(f"MISS {file_name}")
        continue
    with path.open("r", encoding="utf-8-sig", newline="") as f:
        reader = csv.DictReader(f)
        columns = set(reader.fieldnames or [])
    missing = sorted(required - columns)
    if missing:
        print(f"BAD {file_name}: missing {', '.join(missing)}")
    else:
        print(f"OK {file_name}")
```

这段脚本不碰业务数据，只检查字段。你可以让 Codex 每次生成报告前先跑它，避免模型拿缺字段的数据硬分析。

## 7. 客服表：先脱敏，再分析

客服数据要特别注意隐私。建议导出后先删掉姓名、手机号、地址、订单号等字段，只保留问题内容和商品编号。

`data/customer_service.csv`：

```csv
date,product_id,channel,question,after_sales_result
2026-06-16,P001,店铺客服,"收到后发现比想象小，能换大号吗","引导换货"
2026-06-16,P002,店铺客服,"用了以后没有明显感觉，可以退吗","按平台规则处理"
```

分析提示词：

```markdown
请分析 data/customer_service.csv，生成 reports/customer-service-review.md。

输出：
1. 高频问题分类
2. 每类问题数量
3. 标准回复草稿
4. 可能需要优化详情页的点
5. 可能需要反馈给供应商的问题
6. 必须人工介入的问题

要求：
不要输出任何个人隐私。
回复模板不要承诺平台规则以外的赔付。
涉及功效、质量争议、投诉风险时标注人工复核。
```

客服自动化的目标不是让用户感觉被机器敷衍，而是把重复问题标准化，把严重问题更快推给人处理。

### 7.1 客服模板要区分“自动回复”和“人工介入”

建议让 Codex 输出两类模板：

| 类型 | 适合场景 | 处理方式 |
| --- | --- | --- |
| 自动回复草稿 | 尺寸咨询、发货时间、安装说明 | 可进入 FAQ，但上线前人工审核 |
| 人工介入 | 退款争议、投诉、质量问题、平台纠纷 | 只生成处理要点，不自动回复 |

提示词里可以明确写：

```markdown
如果问题涉及退款争议、质量投诉、赔付、平台申诉，不要生成直接发送给用户的话术。
请只输出客服主管内部处理要点，并标注“人工介入”。
```

这样客服自动化不会越界。

## 8. 利润表：让 Codex 生成“需复核”的经营报告

利润分析建议单独放一张表，不要只看商品销售额。

`data/profit.csv`：

```csv
sku,product_id,quantity,sale_price,purchase_cost,platform_fee,ad_cost,shipping_cost,package_cost,coupon_cost,refund_amount
P001-30L,P001,18,59,28,4,300,6,1,2,0
P002-1PC,P002,5,39,12,3,180,5,1,1,39
```

利润复盘提示词：

```markdown
请分析 data/profit.csv，生成 reports/profit-review.md。

请计算：
1. 总销售额
2. 总采购成本
3. 平台费用
4. 广告成本
5. 物流和包材成本
6. 退款金额
7. 单件净利
8. 净利率

请标出：
1. 销量高但净利低的 SKU
2. 广告成本占比过高的 SKU
3. 退款后可能亏损的 SKU
4. 需要人工复核的财务口径

注意：
所有经营结论都必须写“需人工复核”。
```

这里的底线是：Codex 可以算账，不能替你拍板。尤其涉及税费、跨境平台佣金、汇率、退款周期、仓储费时，必须以财务口径为准。

### 8.1 利润复盘建议输出“利润桥”

普通利润表只告诉你赚没赚钱，利润桥能告诉你钱消失在哪里：

```text
销售额
  - 商品成本
  - 平台佣金
  - 广告成本
  - 物流成本
  - 包材成本
  - 优惠券
  - 退款损失
  = 预估净利
```

可以让 Codex 按 SKU 输出：

| SKU | 销售额 | 商品成本 | 广告 | 物流包材 | 退款 | 预估净利 | 最大拖累项 |
| --- | --- | --- | --- | --- | --- | --- | --- |

运营看这张表，会比只看“净利率 8%”更容易采取动作：是广告太贵、退款太高，还是售价本来就没空间。

## 9. 接入 4SAPI：把模型调用放到统一入口

如果你的流程只是手动让 Codex 读表生成报告，可以先不写接口。但当你要把日报、库存预警、广告复盘做成定时任务时，就需要统一模型入口。

4SAPI 的常见接入方式是 OpenAI 兼容接口：

```text
Base URL: https://4sapi.com/v1
API Key: 4SAPI 工作台令牌
Model: 从模型广场复制完整模型名
```

Python 示例：

```python
from openai import OpenAI

client = OpenAI(
    api_key="YOUR_4SAPI_KEY",
    base_url="https://4sapi.com/v1",
)

response = client.chat.completions.create(
    model="your-model-name",
    messages=[
        {"role": "system", "content": "你是电商运营数据分析助手，只输出可复核的运营建议。"},
        {"role": "user", "content": "请根据今天的广告数据生成日报摘要。"},
    ],
)

print(response.choices[0].message.content)
```

Node.js 示例：

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.SAPI_KEY,
  baseURL: "https://4sapi.com/v1",
});

const response = await client.chat.completions.create({
  model: "your-model-name",
  messages: [
    { role: "system", content: "你是电商运营数据分析助手，只输出可复核的运营建议。" },
    { role: "user", content: "请根据库存表生成风险清单。" },
  ],
});

console.log(response.choices[0].message.content);
```

建议按任务拆 Key：

| Key | 用途 | 额度建议 |
| --- | --- | --- |
| ecommerce-report | 日报、周报 | 中等额度 |
| ecommerce-listing | 上架草稿 | 低额度 |
| ecommerce-risk | 库存、客服风险 | 中等额度 |
| ecommerce-test | 新提示词测试 | 很低额度 |

这样即使某个脚本写错，也不至于把所有额度一次耗完。对小团队来说，统一入口最大的价值不是“炫技”，而是成本和日志可追踪。

### 9.1 记录每次模型调用

如果要把 SOP 跑成长期流程，建议保存一份 `reports/model-usage.csv`：

```csv
date,task,key_name,model,input_tokens,output_tokens,status,accepted,notes
2026-06-17,daily_report,ecommerce-report,your-model-name,12000,1800,ok,yes,"日报被采纳"
2026-06-17,listing,ecommerce-listing,your-model-name,5000,1600,ok,partial,"标题需人工改平台词"
```

这张表的重点不是精确到每一分钱，而是回答三个问题：

1. 哪类任务最耗 Token。
2. 哪类输出最容易被人工采纳。
3. 哪些提示词经常返工。

4SAPI 负责统一入口和日志，仓库里的 `model-usage.csv` 负责把模型调用和业务结果对应起来。两边结合，才能知道“花的模型成本有没有换来运营时间节省”。

### 9.2 模型选择不要只看单价

电商 SOP 里可以把任务分成三档：

| 档位 | 任务 | 选择原则 |
| --- | --- | --- |
| 低成本 | 字段清洗、分类、FAQ 初稿 | 便宜、稳定、速度快 |
| 中成本 | 日报、广告复盘、库存风险 | 输出结构稳定，计算解释清楚 |
| 高能力 | 选品风险、复杂利润复盘、跨表归因 | 少量调用，重点看准确和可复核 |

便宜模型如果导致返工，最后未必便宜；强模型如果拿去做简单分类，也是一种浪费。统一接入 4SAPI 后，最适合做的就是按任务切模型，而不是全流程一把梭。

## 10. 每日运行方式：先半自动

第一阶段可以完全手动：

```text
1. 运营从后台导出 CSV
2. 放入 data/ 目录
3. 让 Codex 读取 prompts/daily_report.md
4. 生成 reports/日期-daily-report.md
5. 人工复核后执行动作
```

第二阶段再加入脚本：

```text
1. validate_csv.py 检查字段是否齐全
2. summarize_inputs.py 生成输入摘要
3. 调用模型生成日报
4. 把报告发到团队群
5. 仍然人工确认调价、投流和售后动作
```

日报结构建议固定：

```markdown
# 电商运营日报

## 今日概览

## 销售与利润

## 广告投放

## 库存与发货

## 客服与售后

## 竞品变化

## 风险清单

## 明日建议

## 必须人工复核
```

固定结构会让团队更容易比较每天变化，也方便后续做周报和月报。

### 10.1 daily_report.md 模板可以这样写

```markdown
请生成今天的电商运营日报。

读取范围：
- data/products.csv
- data/ads.csv
- data/inventory.csv
- data/customer_service.csv
- data/profit.csv

输出到 reports/YYYY-MM-DD-daily-report.md。

报告结构：
1. 今日概览：最多 5 条
2. 销售与利润：说明增长、下滑和亏损 SKU
3. 广告投放：按素材输出 CTR、CVR、CPA、ROI
4. 库存与发货：标出 7 天内可能断货 SKU
5. 客服与售后：高频问题和差评风险
6. 竞品变化：如果没有数据，写“今日无竞品数据”
7. 风险清单：表格输出，必须含人工复核列
8. 明日建议：最多 5 条，按优先级排序

要求：
- 不修改 data/ 原始文件
- 样本不足必须标注
- 所有调价、投流、售后、财务结论都写“需人工复核”
- 不输出用户隐私
```

模板固定后，你每天只需要更新数据，报告质量会稳定很多。

## 11. 常见坑

第一，把表格原始数据直接丢进聊天窗口。

如果包含用户隐私、订单号、地址、手机号，这个习惯很危险。先脱敏，再分析。

第二，让模型直接给广告预算结论。

投流数据受样本量、归因周期、退款、季节和活动影响很大。Codex 可以指出异常，预算必须人定。

第三，让模型改原始表格。

早期最好只读数据、输出报告。等流程稳定，再考虑让脚本写入派生文件。

第四，用一把 Key 跑所有任务。

日报、客服、上架、测试应该分开。通过 4SAPI 拆项目 Key、看日志和控额度，会比后面追账单轻松很多。

第五，不写平台规则。

不同平台对标题、功效、素材、售后承诺都有要求。AGENTS.md 和提示词里必须把这些边界写清楚，发布前仍要人工核对。

## 12. 验收清单：报告能不能用

不要只看 Codex 有没有生成报告，要看报告能不能进入运营动作。可以用这张清单验收：

| 检查项 | 通过标准 |
| --- | --- |
| 数据来源 | 报告写清楚读取了哪些文件 |
| 字段缺失 | 缺字段时没有硬编结论 |
| 指标计算 | CTR、CVR、CPA、ROI 公式正确 |
| 样本提示 | 小样本结论有风险提示 |
| 隐私处理 | 没有输出手机号、地址、订单号 |
| 内容合规 | 没有绝对化、夸大或违规承诺 |
| 人工复核 | 调价、预算、售后、财务都有复核标记 |
| 建议可执行 | 每条建议能对应负责人或下一步动作 |

如果一份报告只有“总体不错、建议优化”，那就不合格。好的 Codex 报告应该能让团队知道明天先处理哪三件事。

## 13. 灰度上线：从一个品类开始

不要全店铺一起上。更稳的做法是先选一个品类或 5 个 SKU：

```text
第1周：只做日报，不给自动建议
第2周：加入库存预警和广告复盘
第3周：加入上架草稿和客服 FAQ
第4周：复盘采纳率、返工率和模型成本
```

可以记录三个指标：

| 指标 | 说明 |
| --- | --- |
| 报告采纳率 | 报告建议中有多少被人工采纳 |
| 返工率 | 输出需要大改的比例 |
| 节省时间 | 运营每周少做多少整理工作 |

这三个指标比“模型回答得像不像人”更重要。电商 SOP 的价值最终要落到少返工、少漏看、少踩坑。

## 14. 一周落地计划

如果你想快速试一轮，可以按 7 天执行：

| 天数 | 目标 | 产出 |
| --- | --- | --- |
| Day 1 | 整理商品、订单、广告、库存字段 | data 模板 |
| Day 2 | 写 AGENTS.md 和提示词 | prompts 模板 |
| Day 3 | 生成第一版商品上架草稿 | listing-draft.md |
| Day 4 | 生成广告复盘和库存预警 | ad-review.md、inventory-warning.md |
| Day 5 | 分析客服记录和差评风险 | customer-service-review.md |
| Day 6 | 接入 4SAPI 测试模型调用 | API Key、Base URL、测试脚本 |
| Day 7 | 固定日报模板并人工复盘 | daily-report.md |

这一周的目标不是追求“全自动”，而是验证三件事：

- 数据字段够不够用。
- Codex 生成的报告是否能帮运营少花时间。
- 4SAPI 的模型调用成本和稳定性是否符合预期。

## 15. 总结：先让流程变清楚，再谈自动化

电商团队用 Codex，最值得做的不是让它替你拍脑袋，而是让它变成一个稳定的运营助手：

- 读表格。
- 查异常。
- 写草稿。
- 生成报告。
- 提醒风险。
- 沉淀 SOP。

当这些基础动作跑顺后，再考虑接入更多平台、做定时任务、自动推送日报，甚至把供应链、客服、投流和利润数据串起来。

4SAPI 在这套流程里扮演的是模型网关：统一 Base URL、统一 Key、统一模型选择、统一日志和额度。Codex 负责执行和整理，人负责判断和确认。这个分工清楚了，电商 AI 才会从“试试看”变成真正能复用的生产力。

参考资料：

- 4SAPI官网：https://blog.4sapi.com/zh
- 4SAPI接入教程：https://4sapi.apifox.cn/
- 4SAPI接入地址：https://4sapi.com/v1
- OpenAI Codex文档：https://developers.openai.com/codex
