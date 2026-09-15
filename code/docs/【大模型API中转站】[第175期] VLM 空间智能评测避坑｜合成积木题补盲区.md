---
title: "【大模型API中转站】[第175期] VLM 空间智能评测避坑｜合成积木题补盲区"
tags:
  - 多模态
  - 模型评测
  - 空间智能
  - 合成数据
  - API 中转
description: "把空间能力评测补进多模态选型流程：用合成积木题与拓扑用例，提前暴露文档版面、UI 定位与空间问答里的翻车点。"
---

# 【大模型API中转站】[第175期] VLM 空间智能评测避坑｜合成积木题补盲区

多模态模型的公开榜单分数越来越好看，真正接进业务的那一刻，翻车往往发生在没人考过的空间能力上。这一期把空间评测的盲区拆开来看，并给出一条用合成积木题补齐它的接入路线。

## 一、榜单分数很漂亮，业务一接就翻车

我近期给一个文档处理业务选多模态模型，又碰到了熟悉的场景：公开评测集上几乎全绿，接进真实业务后问题一个接一个——版面解析把两栏文本的阅读顺序排错；UI 截图定位，框点到了相邻按钮上；空间问答里问"左边的表格在图的上半部分还是下半部分"，模型开始一本正经地胡说。

这些任务有一个共同点：全都要模型在内部维护一个空间结构。哪个框套着哪个框，哪块区域换个视角会不会重叠，哪个元素压在另一个元素上面。公开 benchmark 里考感知、考 OCR、考常识问答的题很多，考空间关系的题很少，拓扑关系那一层更是接近空白。所以分数高只说明考过的部分过关，空间能力过没过关，榜单没有给出答案。

我平时挑模型、跑评测、接业务，中间层一直走 4sapi（https://4sapi.com）的 OpenAI 兼容接口：模型随便换，评测脚本和业务代码一行不用改。这一整套空间评测基建也是挂在这层中转上跑的，后面展开。

## 二、空间智能到底在测什么

空间智能可以拆成两层，评测设计也该跟着拆。

第一层是度量属性：距离、角度、大小、形状这类可以用数值刻画的性质。"这两个物体隔多远""这个夹角是锐角还是钝角"，考的是度量。

第二层是拓扑属性：邻接、包含、连通这类关系。它们有一个数学特征——在连续形变下保持不变。一张纸揉皱了，纸上两个点是否相邻、一个区域是否还包着另一个区域、一条线是否还连着，这些关系不变。认知科学里有一条基础判断：拓扑关系是空间理解的根基，人对空间的把握先于度量、始于拓扑。

落到业务上，拓扑出现得比度量更频繁：文档里"这个文本框属于哪个分组框"，流程图里"这两个节点之间有没有连线"，UI 里"这个弹窗盖在哪个容器上"，全是拓扑问题。度量错一点通常还能容忍，拓扑一错，业务语义直接崩——分组归属错了，抽取出来的字段就会挂到错误的实体上。

把两层拆开还有个工程上的好处：考题可以分开建库、分开计分，哪一层的分数掉了一目了然，定位问题不用翻全套日志。

## 三、MindTopo：基础模型评测缺的那一维

我注意到一项叫 MindTopo 的工作，做的就是这件事：评测基础模型在拓扑空间里的推理，具体考连续形变下保持不变的关系，比如邻接、包含、连通。它背后有一个值得所有接入方记住的判断：认知科学早已把拓扑关系当作空间理解的基础，而面向基础模型的评测普遍缺失这个维度。

缺位的代价最后由谁承担？由接入方承担。拓扑推理弱的模型，接到文档层级解析、流程图理解、组织架构图问答这类业务上，错误率会在真实数据里集中爆发；而且这种错误往往不报错、只出错——模型会非常自信地给出一个错误的层级归属，下游照单全收。把拓扑用例单独列出来考，相当于在选型阶段就把这层风险量出来，比上线后救火便宜得多。

## 四、SpatialBlock：用一万五千道积木题补盲区

另一项叫 SpatialBlock 的工作，把补盲区的办法做成了可复用的工程：提出 SpatialBlock-15k 合成数据集，一共 1.5 万道积木堆叠题，覆盖 3D→2D 投影、视点变换与结构组合三类空间技能，并用受控色块作视觉锚点提示。

有几个点值得单独展开：

- **合成生成，不是人工标注。** 题目和标准答案由程序同时产出，1.5 万道题没有一道需要人去标答案。这是合成数据最大的杠杆：标注成本被压成生成成本。
- **受控色块是视觉锚点。** 色块颜色固定、数量可控，模型答题时有明确的视觉抓手，题目本身不引入无关变量。自建评测时照搬这个思路，题目质量能稳住。
- **两种用法都有效。** 用这份数据直接训练的 LVLM，和做推理式预测的 LVLM，都显著超过基线，并且能泛化到真实空间任务。泛化这一条对接入方最关键：玩具任务上的提升能迁移到真实场景，合成评测集才值得建。
- **代码与数据已开源。** 开源意味着可以直接拿它当回归评测集的底子，不必从零造题。

接入方视角再补一句：直接训练是模型厂的事，接入方真正能复用的是出题思路——受控色块、结构化玩具任务、答案同步产出，这三件事拿 Python 就能搭出来。

## 五、合成评测流水线长什么样

把两项工作的思路合起来，接入方可以搭一条完整的流水线：

```text
业务场景（文档版面 / UI 定位 / 空间问答）
      |
      v
拆解空间技能（投影 / 视点 / 组合 / 拓扑 / 计数 / 定位）
      |
      v
合成题生成（积木堆叠 + 受控色块锚点，题目与标准答案同步产出）
      |
      v
模型作答（图片 base64 输入，走 OpenAI 兼容接口）
      |
      v
分技能计分（按技能维度聚合，不合成一个笼统总分）
      |
      v
回归基线入库（换模型、升版本、改提示词后重跑，分数进台账）
```

这条流水线里最容易被忽略的是最后两步。分技能计分的意义在于：总分 85 的模型可能是投影 95、拓扑 60，也可能是六项均匀的 85——两种模型适合的业务完全不同，前者适合版面抽取，后者才敢接空间问答。回归基线入库的意义在于：多模态模型更新很勤，供应商可能静默升级版本，没有回归集，退化只能在业务事故里被发现。

整条流水线里没有任何一步需要人工标注，人力全部花在写生成器和看失败样本上，这两件事才是真正值钱的工程投入。

## 六、动手生成积木式合成评测题

思路完全照搬 SpatialBlock：用 Pillow 画受控色块的等距积木堆叠图，程序画图的同时顺手把标准答案算出来，题目、答案、技能标签一起写进 JSONL。下面这份生成器可以直接跑，先 `pip install pillow`，目录会自动创建：

```python
# gen_blocks.py：合成积木评测题生成器
# 依赖：pip install pillow，其余为标准库
import json
import os
import random
from dataclasses import dataclass
from PIL import Image, ImageDraw

# 受控色板：颜色即视觉锚点，固定五种，题目文本直接引用色名
COLORS = {"红": "#E63946", "蓝": "#457B9D", "绿": "#2A9D8F", "黄": "#E9C46A", "紫": "#9B5DE5"}
GRID_W, GRID_L, MAX_LAYER = 4, 4, 3   # 底面 4x4 网格，最多 3 层，保证可数
CELL, HEIGHT = 52, 30                 # 网格单元边长与单层高度（像素）

@dataclass
class Block:
    layer: int     # 0 为底层
    grid_x: int
    grid_y: int
    color: str     # 色名，对应 COLORS 的键

def _shade(hex_color, factor):
    """按比例调暗颜色画侧面，制造立体感"""
    r, g, b = (int(hex_color[i:i + 2], 16) for i in (1, 3, 5))
    return (int(r * factor), int(g * factor), int(b * factor))

def _proj(gx, gy, gz, ox, oy):
    """等距投影：网格坐标映射到画面坐标，gz 沿层高方向"""
    return ox + (gx - gy) * CELL, oy + (gx + gy) * CELL * 0.5 - gz * HEIGHT

def _occupied(blocks):
    """当前堆叠里所有被占用的格位，含层号"""
    return {(b.grid_x, b.grid_y, b.layer) for b in blocks}

def make_stack(rng):
    """生成一摞合法积木：每块必须落在底层或另一块正上方"""
    blocks, occupied = [], set()
    for layer in range(MAX_LAYER):
        for _ in range(rng.randint(2, 5)):
            gx, gy = rng.randrange(GRID_W), rng.randrange(GRID_L)
            if layer > 0 and (gx, gy, layer - 1) not in occupied:
                continue  # 悬空积木会让标准答案失去物理意义
            if (gx, gy, layer) not in occupied:
                blocks.append(Block(layer, gx, gy, rng.choice(list(COLORS))))
                occupied.add((gx, gy, layer))
    return blocks

def render(blocks, path):
    """远处先画、近处后画，等距投影下天然形成遮挡"""
    img = Image.new("RGB", (640, 480), "#F8F9FA")
    d = ImageDraw.Draw(img)
    ox, oy = 320, 110
    for b in sorted(blocks, key=lambda t: (t.grid_x + t.grid_y, t.layer)):
        top = [_proj(b.grid_x, b.grid_y, b.layer + 1, ox, oy),
               _proj(b.grid_x + 1, b.grid_y, b.layer + 1, ox, oy),
               _proj(b.grid_x + 1, b.grid_y + 1, b.layer + 1, ox, oy),
               _proj(b.grid_x, b.grid_y + 1, b.layer + 1, ox, oy)]
        left = [_proj(b.grid_x, b.grid_y + 1, b.layer + 1, ox, oy),
                _proj(b.grid_x + 1, b.grid_y + 1, b.layer + 1, ox, oy),
                _proj(b.grid_x + 1, b.grid_y + 1, b.layer, ox, oy),
                _proj(b.grid_x, b.grid_y + 1, b.layer, ox, oy)]
        right = [_proj(b.grid_x + 1, b.grid_y, b.layer + 1, ox, oy),
                 _proj(b.grid_x + 1, b.grid_y + 1, b.layer + 1, ox, oy),
                 _proj(b.grid_x + 1, b.grid_y + 1, b.layer, ox, oy),
                 _proj(b.grid_x + 1, b.grid_y, b.layer, ox, oy)]
        base = COLORS[b.color]
        d.polygon(top, fill=_shade(base, 1.0), outline="#333333")
        d.polygon(left, fill=_shade(base, 0.72), outline="#333333")
        d.polygon(right, fill=_shade(base, 0.50), outline="#333333")
    img.save(path)

def build_questions(blocks, image_name):
    """出题与标准答案同步产出：合成评测零标注成本的关键"""
    items, occupied = [], _occupied(blocks)
    red = [b for b in blocks if b.color == "红"]
    if red:
        items.append({"image": image_name, "skill": "projection",
                      "question": "红色积木最高的一块在第几层？（从 0 开始数）",
                      "answer": str(max(b.layer for b in red))})
    columns = {(b.grid_x, b.grid_y) for b in blocks}
    items.append({"image": image_name, "skill": "viewpoint",
                  "question": "从正上方俯视，能看到几块积木的顶面？",
                  "answer": str(len(columns))})
    pressed = sum(1 for b in blocks if (b.grid_x, b.grid_y, b.layer + 1) in occupied)
    items.append({"image": image_name, "skill": "combination",
                  "question": "有几块积木上面直接压着别的积木？",
                  "answer": str(pressed)})
    items.append({"image": image_name, "skill": "counting",
                  "question": "图里一共有多少块积木？被挡住的也要算。",
                  "answer": str(len(blocks))})
    return items

def main():
    rng = random.Random(2026)
    os.makedirs("images", exist_ok=True)
    with open("questions.jsonl", "w", encoding="utf-8") as fout:
        for i in range(50):
            blocks = make_stack(rng)
            image_name = f"stack_{i:04d}.png"
            render(blocks, f"images/{image_name}")
            for item in build_questions(blocks, image_name):
                fout.write(json.dumps(item, ensure_ascii=False) + "\n")

if __name__ == "__main__":
    main()
```

每张图最多出四道题，分别打在投影、视点、组合、计数四个技能上。计数题刻意保留了遮挡，模型需要按"每块都必须被支撑"这条物理约束去推断看不见的部分——这正是真实业务里遮挡理解的玩具版。拓扑技能也能在同一副骨架上继续加：比如"某种颜色的积木是否全部连通"，用网格坐标做四邻域 BFS 就能算出标准答案，MindTopo 考的邻接与连通在这个玩具场景里都有对应物。

## 七、接入判分链路：从图片到分技能分数

生成器只解决出题，判分要把模型接进来。骨架分三段：图片转 base64 拼进 messages，走 OpenAI 兼容客户端发请求，答案归一化后逐题比对。中转层用 4sapi（https://4sapi.com）的 OpenAI 兼容接口，好处是换模型只改一个环境变量，评测脚本零改动；中转层还可以给评测任务单独挂一份配额，回归测试再怎么跑也吃不掉业务调用的预算。判分环节全部在本地完成，标准答案不需要出内网，出网的只有题目图片和问题文本。

```python
# run_eval.py：自动判分骨架
# 依赖：pip install openai
import base64
import json
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OPENAI_API_KEY"],
    base_url=os.environ["OPENAI_BASE_URL"],  # 中转地址从环境变量读取，不硬编码进仓库
)

def ask_model(image_path, question, model):
    """图片走 base64 输入，OpenAI 兼容格式"""
    with open(image_path, "rb") as f:
        b64 = base64.b64encode(f.read()).decode()
    resp = client.chat.completions.create(
        model=model,
        temperature=0,
        max_tokens=32,
        messages=[{
            "role": "user",
            "content": [
                {"type": "text", "text": f"{question}\n只输出最终答案，不要解释。"},
                {"type": "image_url",
                 "image_url": {"url": f"data:image/png;base64,{b64}"}},
            ],
        }],
    )
    return resp.choices[0].message.content.strip()

def normalize(text):
    """答案归一化：去空白、全角数字转半角、统一小写"""
    text = text.strip().replace(" ", "").replace("　", "").lower()
    return text.translate(str.maketrans("０１２３４５６７８９", "0123456789"))

def run_eval(dataset_path, image_dir, model):
    by_skill, wrong = {}, []
    for line in open(dataset_path, encoding="utf-8"):
        item = json.loads(line)
        pred = normalize(ask_model(os.path.join(image_dir, item["image"]),
                                   item["question"], model))
        gold = normalize(item["answer"])
        ok = pred == gold
        by_skill.setdefault(item["skill"], []).append(ok)
        if not ok:
            wrong.append({**item, "pred": pred})
    report = {s: round(sum(v) / len(v), 4) for s, v in by_skill.items()}
    return report, wrong

if __name__ == "__main__":
    report, wrong = run_eval("questions.jsonl", "images", os.environ["EVAL_MODEL"])
    print(json.dumps(report, ensure_ascii=False, indent=2))
    with open("wrong_cases.jsonl", "w", encoding="utf-8") as f:
        for w in wrong:
            f.write(json.dumps(w, ensure_ascii=False) + "\n")
```

跑完拿到的是按技能拆开的分数表。示例里的归一化刻意从简：模型答"红色"而标准答案是"红"会判错，生产化时再加一层同义归并即可。失败样本一定要回看原图，多数时候能直接定位错在哪一层——是把投影关系看反了，还是压根没数对块数，两种错的处置动作完全不同。

## 八、空间能力验收清单

把空间评测做进选型与验收流程，我固定按这份清单走：

1. **按业务拆技能。** 先把业务里的空间操作列全——阅读顺序、层级归属、元素定位、遮挡关系——再映射到投影、视点、组合、拓扑、计数、定位六类技能上。
2. **合成题覆盖每个技能。** 每个技能至少配一个生成器，题量按技能均衡分配，别让某一类题占掉八成。
3. **拓扑用例单列。** 邻接、包含、连通三类题单独成组、单独计分，不与度量题混在一起平均，否则最容易缺的那一维永远藏在总分里。
4. **真实样本抽检对照。** 合成分数出来后，用几百条真实业务样本做对照，确认两条分数曲线同向，防止"玩具分高、实战拉胯"。
5. **版本升级回归。** 模型版本、提示词、预处理管线任何一项变更都全部重跑并入库，退化当场可见，不靠业务事故报警。

## 九、成本对比：三种路线怎么选

三种评测路线的成本结构完全不同，下表按"每千题"口径估算（人工标注为估算值；合成路线的成本主要是一次性写生成器的人力）：

| 路线 | 每千题成本（估算口径） | 标准答案质量 | 技能覆盖可控性 | 适用阶段 |
| --- | --- | --- | --- | --- |
| 真实场景人工标注 | 高，按每题几元到十几元的人工口径推算 | 高，但依赖标注员的空间理解 | 低，受已有数据分布限制 | 上线前验收、抽检 |
| 合成评测集 | 低，脚本跑一遍近乎零边际成本 | 高，答案由程序同步产出 | 高，技能与难度都可参数化 | 选型对比、版本回归 |
| 只看公开 benchmark | 零 | 不可控，且不含自家业务技能 | 低，考什么是别人定的 | 初筛，缩小候选范围 |

三列成本差出几个数量级，所以合理做法是分层组合：合成评测集负责量大管饱的日常回归，真实样本抽检负责最后把关，公开榜单只负责初筛。把公开榜单当验收依据，是接入方最常见的评测错位。

## 十、风险与合规提示

合成评测不是免费的午餐，三个坑要提前绕开：

- **分布偏窄会让模型过拟合题型。** 积木题考的是空间技能，但如果所有题长一个样，模型可能学到的是这类图的答题套路而不是空间能力本身。缓解办法：技能维度固定，表面形式放开——底面网格大小、层数、色块数量、遮挡程度全部参数化随机；再拿真实样本抽检做同向性对照，两条证据都对上才采信。
- **评测集要防泄漏进训练数据。** 生成器开源之后，同样的题可能被别人拿去训练；自家业务评测集更要管住分发：评测图片不进公开仓库、不带可被爬取的标识，台账里记录每份评测集的流转范围。这是最基本的数据卫生——评测题一旦混进训练语料，分数会虚高到失去参考意义，最后误伤的是自己的选型决策。
- **分数不能当业务验收的唯一依据。** 合成评测是选型与回归的仪表盘，不是验收报告。上线前的最终判断要落在真实业务样本的抽检上，合成与真实两条证据链都对上，再决定切多少流量。

顺带把边界说清楚：这里讨论的全部是接入方视角的合法评测与数据治理——管好自己的评测集、审计好自己的调用、按任务挂配额，不涉及任何绕过官方限制的操作；做评测的目的，是让模型在合规接入的前提下被更准确地度量。

## 总结

这一期把"看图干活"业务里最容易翻车的空间能力拆成了度量与拓扑两层：MindTopo 补上了拓扑推理这块基础模型评测长期缺失的维度，SpatialBlock-15k 证明了合成积木题能把标注成本打成生成成本，并且能泛化到真实空间任务。落到工程上，是一条从技能拆解、合成出题、base64 输入、自动判分到分技能回归的完整链路，配一份验收清单和三个风险提示。整套链路继续跑在 4sapi（https://4sapi.com）的 OpenAI 兼容接口上，模型随便换，评测台账不动。欢迎在评论区聊聊手头业务里空间能力的翻车经历。
