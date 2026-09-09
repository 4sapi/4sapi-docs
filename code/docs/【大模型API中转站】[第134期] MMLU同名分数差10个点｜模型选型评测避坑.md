---
title: 【大模型API中转站】[第134期] MMLU同名分数差10个点｜模型选型评测避坑
tags: [模型评测, MMLU, 模型选型, harness, grader, dataset-split]
description: 同一供应商、同一模型家族、同一指标名（mmlu）、同一单位，分数仍可能差 10 个点——差别在 runner、grader 与 dataset split。含漂移来源表、自建评测 Python 教程、跨模型对比表模板与验收清单。
---

两个构建若用同一供应商、同一模型家族、同一指标名（如 mmlu）、同一单位，仍可能给出不同分数——差别在 runner（运行器）、grader（评分器）与 dataset split（数据切分）。这一期我把这个坑拆开，并给出从切片到入库的自建评测完整流程。

## 一、先看现象：同名分数差 10 个点

本周技术圈的核心议题，就是这个"同名不同分"。我日常在 4sapi（https://4sapi.com）同时接入多家供应商的模型做选型对比，两周前就撞上过一次：同一家供应商、同一模型家族的两个构建，官方评测页面都写着 mmlu，一个标 74.2，一个标 84.5，单位都是百分数——整整差了 10 个点。若只看这个名字，我几乎要以为后一个构建把模型能力翻了一档。

把两个构建的评测配置调出来一比对，答案立刻清楚了：第一个构建用旧版 runner（运行器）跑 0-shot，grader（评分器）用宽松的包含匹配，数据集用的是 MMLU 的 test 全集；第二个构建用新版 harness 跑 5-shot，grader 换成严格精确匹配，数据集换成了清洗后抽样的一小份子集。四项里三项不同，分数差 10 个点一点都不意外。

这不是个例。任何一次评测结果，本质上都绑定了 runner、grader 与 dataset split 的具体取值；标签上只写一个 "mmlu"，等于把这些变量全部抹掉了。

## 二、痛点：按榜单选型，线上翻车

大多数团队的选型流程是：打开各家榜单 → 按 mmlu 之类的指标排序 → 挑分数最高的模型 → 接入线上。这个流程的翻车率，比想象中高得多。

我见过一个真实案例（示意）：某团队按 mmlu 85.1 选了一个模型做客服意图分类，上线后准确率比旧模型低了 6 个百分点。原因不复杂——榜单上的 85.1 是在通用知识问答的分布上测出来的，而线上任务是多轮对话里的短句意图识别，两类样本的难度与分布完全不同。更隐蔽的是，榜单分数还可能来自不同的 grader：旧模型用的是允许同义改写判对的宽松评分器，新模型用的是逐字精确匹配，同一批输出得出的分数差 3~5 个点都很正常。

还有一类翻车来自 dataset split：同一份评测数据，A 榜用全量 test，B 榜用洗过一遍的子集，B 榜的分数天然会虚高或虚低。按榜单分数选型，等于默认"所有榜单背后是同一套测量程序"——而这个默认前提本身就是错的。

## 三、原理速览：评测 = harness + 数据 + 评分器 + 切分

把一次评测拆开，结果由四部分共同决定：

```text
一次评测结果 ≈
    runner（harness：采样、prompt 模板、批处理、日志）
  + 数据集与 dataset split（科目、切分、题量、清洗）
  + grader（评分器：判分规则、容错方式、判分提示词）
  + 环境（seed、温度、数值精度、硬件型号）
```

runner 决定模型怎么被调用：few-shot 数量、模板措辞、batch 大小、解码参数，全在这里。dataset split 决定考什么：同一套题，抽 100 条和抽 1000 条，分数分布就不同。grader 决定怎么算对：精确匹配、包含匹配、正则、还是让另一个模型当裁判（LLM-as-judge），判分口径直接影响分数。环境决定稳定性：seed 不同、温度不为 0、精度从 bf16 换成 fp16，都会引入抖动。

这四个变量里任意一个不同，两个 "mmlu" 分数就不具备横向可比性。

## 四、同名基准分数漂移来源表

把常见漂移来源列成一张表，选型前对一眼，能挡掉大部分坑：

| 漂移来源 | 具体变量 | 影响机制（示意） | 规避手段 |
| --- | --- | --- | --- |
| runner（harness） | harness 版本、few-shot 数量、prompt 模板、batch、解码参数 | 同一题集在不同模板与采样下，acc 可差数个点 | 锁死 harness 版本与全部参数，写入版本化配置 |
| grader（评分器） | 精确匹配 / 包含匹配 / 正则 / LLM 判分及其提示词 | 判分宽严不一，同一批答案得出不同正确率 | 固化 grader 规则，并留存判分样例供抽查 |
| dataset split | test / val、抽样子集、科目取舍、题量 | 样本不同导致难度分布平移，分数整体偏高或偏低 | 固定切分规则，计算并记录切片指纹 |
| 环境 | seed、temperature、bf16 / fp16、GPU 型号 | 采样噪声与数值精度差异造成抖动 | 固定 seed、temperature=0、固定精度与硬件 |
| 指标口径 | 是否计入空答案、macro / weighted、是否容错 | 同一份原始结果，算出来的数字不同 | 指标定义写进评测报告，不允许口头约定 |
| 单位与展示 | 百分数 / 0~1 小数、保留位数 | 量纲不同却同写一个 "mmlu" | 统一换算成同一单位再比较 |

这张表里最常被忽视的是 grader 与指标口径：多数人只盯数据集，却不知道评分器换一种容错方式就能让分数移动好几个点。

## 五、"mmlu" 是什么：一个数据集家族，不是一套测量程序

理解了上面四要素，就该接受这个结论："mmlu" 这类标签标识的是一个数据集家族，不是一套完整测量程序；同名分数不可直接横向比较。

这个家族里至少包含：原始 MMLU（57 个科目）、MMLU-Pro（更难、更多干扰项）、MMLU-Redux（修正标注噪声的版本）、各语言版本、还有 test 与 val 的不同切分，以及 0-shot / 5-shot 的不同玩法。任何一个版本配上不同的 runner、grader，都会产出一个新的"mmlu 分数"。名字一样，背后的测量程序完全可能是两套。

所以我在选型时把基准分数降级看待：它只能告诉我"这个模型大概处在什么量级"，不能告诉我"它比另一个模型好 3 个点"。

## 六、选型建议方向：vendor 分数只当入场券

基于上面的认识，我的选型策略调整为三条：

第一，vendor 分数只当入场券。用榜单分数做粗筛，划定候选池，比如把 mmlu 低于某个阈值的模型先排除，但绝不根据几个点的差距做最终决策。

第二，用自有任务集复测。把线上真实流量抽样、清洗、标注后做成自有评测集，所有候选模型在同一套数据、同一套流程下重测。这一步做下来，榜单排名经常被推翻。

第三，评测固化 harness、评分器与切分。自有评测必须把 runner 版本、grader 规则、dataset split 全部固化成代码与配置，随仓库版本化，保证任何一次对比都建立在同一套测量程序上。

这三条正好对应 4sapi（https://4sapi.com）这类中转接入方式的优势：一次接入多家模型，用同一套评测脚本挨个调用，天然消除供应商内部评测口径差异。评测与选型都属于合法用途，我只通过合规接入方式调用模型 API，不涉及任何违规代理。

## 七、自建评测教程 1：washed 数据集切片

选型评测的第一步，是准备一份干净、固定、可复现的数据集切片。我管这个过程叫 washed 切片：先清洗（wash），再切片，最后算指纹。这段流程我在 4sapi（https://4sapi.com）上对每家候选模型都要跑一遍，第一步永远是先出切片。代码示意如下：

```python
# 评测准备：washed 数据集切片（示意流程，可照搬到自有任务集）
import hashlib
import json
from datasets import load_dataset

RAW_DATASET = "cais/mmlu"      # 示意：换成自有任务的源数据集
SUBJECT = "abstract_algebra"   # 示意：固定科目子集
SEED = 20240909                # 固定 seed，切分才可复现

def wash(example):
    # 清洗：丢弃空题干、非法选项、坏答案，统一字段格式
    if not example.get("question") or not example.get("choices"):
        return False
    if len(example["choices"]) < 2:
        return False
    return True

def load_washed_slice():
    ds = load_dataset(RAW_DATASET, SUBJECT, split="test")
    clean = ds.filter(wash)
    # 先洗后切，seed 固定，得到稳定切片
    sliced = clean.shuffle(seed=SEED).select(range(100))
    # 记录切片指纹，供入库与复现比对
    blob = json.dumps(
        [{k: r[k] for k in ("question", "choices", "answer")} for r in sliced],
        ensure_ascii=False)
    fingerprint = hashlib.sha256(blob.encode("utf-8")).hexdigest()
    return sliced, fingerprint

slices, fp = load_washed_slice()
print("切片指纹:", fp[:16])
```

三个要点：一是"先洗后切"，清洗规则变了切片就变，所以清洗逻辑也要版本化；二是 seed 必须固定，否则每次抽取的样本不同；三是切片指纹（sha256）是切片版本的唯一标识，任何一次评测都必须带上它。

## 八、自建评测教程 2：固定 harness 跑单个模型

切片就绪后，用固定 harness 跑单个模型。这里的关键是"固定"两个字：harness 版本、few-shot、batch、seed、解码参数全部写死在配置里，不允许运行时临时改。示意如下：

```python
# 固定 harness：版本、模板、采样参数全部锁死（示意）
# 依赖示意：lm-eval-harness 0.4.x
import lm_eval

EVAL_CONFIG = {
    "model": "hf",                       # 固定后端
    "model_args": {"pretrained": "MODEL_ID"},  # 换成被测模型
    "tasks": ["mmlu_abstract_algebra"],  # 与切片对应的 task 名
    "num_fewshot": 5,                    # few-shot 数量锁死
    "batch_size": 8,                     # 固定 batch，避免显存抖动
    "seed": 20240909,
    "device": "cuda:0",
    # 解码参数：贪婪解码，消除采样噪声（示意字段）
    "gen_kwargs": {"temperature": 0, "top_p": 1.0},
}

results = lm_eval.simple_evaluate(**EVAL_CONFIG)
score = results["results"]["mmlu_abstract_algebra"]["acc"]
print("score:", score)
```

注意 few-shot 的提示模板。模板措辞哪怕只差一个标点，也可能移动零点几个点；所以模板文本要连同 harness 版本一起入库，而不是散落在脚本注释里。

## 九、自建评测教程 3：跑跨模型对比

单个模型能跑通后，跨模型对比就是同一切片、同一 harness、不同 model_id 的循环。示意如下：

```python
# 跨模型对比：同一切片 + 同一 harness，逐个模型跑（示意）
MODELS = [
    "vendor-a/model-x",
    "vendor-b/model-y",
    "vendor-c/model-z",
]  # 示意名单，换成实际候选

def run_model(model_id):
    cfg = dict(EVAL_CONFIG)
    cfg["model_args"] = {"pretrained": model_id}
    results = lm_eval.simple_evaluate(**cfg)
    return results["results"]["mmlu_abstract_algebra"]["acc"]

rows = []
for mid in MODELS:
    acc = run_model(mid)
    rows.append({"model": mid, "acc": acc, "slice_fp": fp[:16]})
    print(f"{mid}: {acc:.4f}  slice={fp[:16]}")
```

到这里，跨模型对比已经在同一套测量程序下完成：同一 slice、同一 harness、同一 grader、同一环境。此时再比较分数，才具备可比性。

## 十、自建评测教程 4：记录环境元数据

分数本身没有意义，分数加元数据才有。每次运行都要记录完整环境信息并入库，否则三个月后没人能复现这份结果。示意如下：

```python
# 记录环境元数据：没有元数据的分数等于没有分数（示意）
import platform
import sqlite3
import datetime

ENV_META = {
    "harness_name": "lm-eval-harness",
    "harness_version": "0.4.3",          # 锁版本
    "dataset": "cais/mmlu",
    "subject": "abstract_algebra",
    "slice_fingerprint": fp,              # 上面算出的切片指纹
    "split": "test",
    "num_fewshot": 5,
    "grader": "exact_match",              # 评分器口径
    "seed": 20240909,
    "temperature": 0,
    "precision": "bf16",
    "gpu": "A100-80G",
    "python": platform.python_version(),
    "run_at": datetime.datetime.now().isoformat(),
}

conn = sqlite3.connect("eval_store.db")   # 示意：结果入库
conn.execute("""CREATE TABLE IF NOT EXISTS eval_runs(
    run_id TEXT PRIMARY KEY,
    model TEXT,
    metric TEXT,
    score REAL,
    meta TEXT,
    created_at TEXT)""")
# 写库时把 ENV_META 一并序列化存入 meta 列
```

grader 口径在这里显式记录为 exact_match。如果哪天换成模糊匹配，它就是一条新记录，而不是"同一个 mmlu 分数的更新"。

## 十一、跨模型对比表模板

跑完对比，用一张固定结构的表交付结论。模板如下（分数均为示意值）：

| 模型 ID（供应商） | vendor 宣称 mmlu | 自建评测 mmlu（固定切片） | Δ | seed | harness 版本 | 切片指纹（前 8 位） |
| --- | --- | --- | --- | --- | --- | --- |
| vendor-a/model-x | 84.5 | 76.3 | -8.2 | 20240909 | 0.4.3 | 9f3a2c1b |
| vendor-b/model-y | 82.1 | 79.8 | -2.3 | 20240909 | 0.4.3 | 9f3a2c1b |
| vendor-c/model-z | 85.1 | 74.0 | -11.1 | 20240909 | 0.4.3 | 9f3a2c1b |

这张表把 vendor 宣称值与自建评测值并排，Δ 一列直接暴露口径差距。注意三行共用同一个切片指纹——这才是"同一测量程序"的证据；指纹不一致时，行与行之间同样不可比。

## 十二、评测结果入库与回归评测（示意）

有了 eval_store.db 这张表，评测就不再是一次性活动，而是一条持续流水线。我把它扩成三件事：

第一，结果入库。每次运行生成 run_id（如 20240909-vendor-a-model-x），连同分数与 ENV_META 一起写入。入库是硬性动作：没入库等于没跑。

第二，回归评测。每当供应商发布新构建，就用同一套切片重跑一遍基线，与上次分数做差。我设了一个示意阈值：同一模型家族、同一测量程序下，分数下跌超过 2 个点就告警，人工核查是新构建的退化，还是环境抖动。

第三，基线冻结。一旦某次评测被采纳用于选型决策，就把那次的切片指纹、harness 版本、grader 规则冻结为基线配置。后续所有对比都以基线为准，防止评测配置被悄悄改掉而不自知。

这套做法把"评测"从拍脑袋变成了可审计的工程资产。

## 十三、抖动控制（示意）

即使测量程序完全固定，分数仍有抖动。抖动的主要来源是采样噪声与数值精度，我的控制手段如下：

其一，解码固定。temperature 设为 0、top_p 设为 1.0，做贪婪解码，消除采样随机性。对生成式任务，这一步影响最大。

其二，seed 固定。数据切分、shuffle、以及任何涉及随机的环节都用同一个 seed，保证每次运行吃到的样本顺序一致。

其三，多跑取分布。对重点决策，同一个配置跑 3~5 次（改变 seed），报告 mean±std。示意经验：100 题的小切片，两次运行差 1~2 个点属于正常抖动，差 5 个点就要查环境或配置是否漂移。

其四，硬件与精度固定。bf16 / fp16、GPU 型号、batch 大小都写进元数据；换卡重跑会引入数值精度差异，抖动控制的前提是环境不变。

## 十四、评测报告清单

一份能支撑选型决策的评测报告，至少包含下列条目；缺任何一项，这份报告就不具备复现价值：

- 被测模型 ID、供应商、构建版本（含权重快照标识）
- harness 名称与精确版本号
- 数据集、科目、split、切分规则与切片指纹
- few-shot 数量与 prompt 模板全文
- grader 规则：精确 / 包含 / 正则 / LLM 判分及其提示词
- 解码参数：temperature、top_p、seed
- 硬件与精度：GPU 型号、bf16 / fp16、batch
- 运行日期与原始日志存放位置
- 多次运行的方差（mean±std）
- 与基线配置的差异说明（若有）

我习惯把这份清单做成模板文件，每次评测只填结果、不改结构。

## 十五、验收清单

报告写完不等于评测合格，过一遍验收清单再下结论：

- [ ] vendor 分数仅作入场券使用，未直接用于最终决策
- [ ] 自有任务集复测通过预设阈值（阈值在评测前定好）
- [ ] harness、grader、dataset split 已固化为版本化配置
- [ ] 切片指纹、环境元数据完整，结果可复现
- [ ] 多次运行方差已记录，决策未建立在单次抖动上
- [ ] 结果已入库，回归评测机制已生效
- [ ] 对比行共用同一测量程序（指纹一致）
- [ ] 接入与评测均为合规方式，无违规代理

## 十六、总结

"mmlu" 只是一个数据集家族的名字，不是一套完整测量程序；runner、grader、dataset split 任一不同，同名分数就不可直接横向比较，差 10 个点并不稀奇。应对办法只有一条：把 vendor 分数当入场券，用自有任务集在固定的 harness、grader 与切分下复测，并让每一次评测都可复现、可入库、可回归。我在 4sapi（https://4sapi.com）上把这套流程跑了两周，踩过的坑基本都写在上面了——欢迎在评论区聊聊各自的选型评测经验。
