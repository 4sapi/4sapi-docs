---
title: "Remotion 实战：如何给真人视频叠加 React 组件"
tags:
  - Remotion
  - React
  - 视频合成
description: "从 React 组件、帧号和 OffthreadVideo 出发，搭建真人口播、录屏、字幕、图表和批量数据视频的渲染流程。"
---
# Remotion 实战：如何给真人视频叠加 React 组件

Remotion 最有价值的能力，不是“可以用 React 做动画”，而是它可以把已有视频导入成底层画面，再让 React 组件按照帧号精准叠加在上面。

这件事很适合真人口播、录屏讲解、产品演示和数据视频：底层是一个已经录好的 MP4，上面再加标题卡、字幕、代码高亮、流程图、进度条和数据图表。

这类任务如果只用一个从零绘制 HTML 的框架，通常要先把真人视频切成画面、再做额外合成；Remotion 则把“视频素材”和“组件画面”放进同一个 React 时间轴里。

本文重点讲 Remotion 的工作模型、项目结构、视频导入、组件叠加、批量生成和验收边界。具体命令和授权条件应以当前 Remotion 文档和项目版本为准。

## 一、Remotion 把视频当成 React 组件

在 Remotion 中，一条视频可以被理解成一个随帧号变化的 React 画面：

```text
当前帧 0   -> React 组件状态 0
当前帧 1   -> React 组件状态 1
当前帧 2   -> React 组件状态 2
...
当前帧 N   -> React 组件状态 N
```

组件通过当前帧计算透明度、位置、缩放、字幕和图表状态。渲染器逐帧获得组件结果，再编码为视频。

最小的组件形态类似这样：

```tsx
import { AbsoluteFill, useCurrentFrame, useVideoConfig } from "remotion";

export const TitleCard = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const opacity = Math.min(1, frame / (fps * 0.5));

  return (
    <AbsoluteFill
      style={{
        backgroundColor: "#f7f7f5",
        justifyContent: "center",
        alignItems: "center",
        opacity
      }}
    >
      <h1>代码做视频</h1>
    </AbsoluteFill>
  );
};
```

`useCurrentFrame()` 提供当前帧，`useVideoConfig()` 提供帧率、宽高和时长等配置。不要把浏览器当前时间当作动画依据，否则逐帧渲染时可能出现状态漂移。

## 二、视频底层和 React overlay

Remotion 的关键场景是：底层播放一段已有视频，前景叠加 React 组件。

```tsx
import {
  AbsoluteFill,
  OffthreadVideo,
  useCurrentFrame,
  useVideoConfig
} from "remotion";

export const TalkingHead = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const showTitle = frame >= 0 && frame < fps * 4;

  return (
    <AbsoluteFill>
      <OffthreadVideo src="assets/talking-head.mp4" />

      {showTitle && (
        <div
          style={{
            position: "absolute",
            left: 80,
            bottom: 80,
            padding: "18px 28px",
            background: "rgba(0, 0, 0, 0.72)",
            color: "white"
          }}
        >
          这一段讲清楚一个概念
        </div>
      )}
    </AbsoluteFill>
  );
};
```

这里的 MP4 是画面底层，标题卡由 React 生成。二者共享同一帧号，所以标题可以精确地在指定帧出现和消失。

## 三、项目结构应该先固定

一个长期维护的 Remotion 项目不要把所有逻辑塞在 `App.tsx`：

```text
remotion-video/
├── src/
│   ├── Root.tsx
│   ├── compositions/
│   │   ├── TalkingHead.tsx
│   │   ├── ScreenRecording.tsx
│   │   └── DataVideo.tsx
│   ├── components/
│   │   ├── Subtitle.tsx
│   │   ├── LowerThird.tsx
│   │   ├── CodeBlock.tsx
│   │   └── ProgressBar.tsx
│   └── data/
│       └── examples.ts
├── public/
│   ├── video/
│   ├── audio/
│   ├── images/
│   └── fonts/
├── scripts/
└── package.json
```

Composition 负责一类视频结构，组件负责可复用视觉元素，数据文件负责内容。这样做的好处是：同一个标题卡可以用于真人口播、录屏和产品发布视频，不需要复制整套时间轴。

## 四、把时间轴变成数据

不要把所有时间点散落在组件代码里。可以统一写成场景数据：

```ts
export const scenes = [
  {
    id: "intro",
    startFrame: 0,
    endFrame: 150,
    title: "今天解决一个问题"
  },
  {
    id: "diagram",
    startFrame: 150,
    endFrame: 420,
    title: "把流程拆成三步"
  }
];
```

组件只负责读取数据和计算画面：

```tsx
const active = scenes.find(
  scene => frame >= scene.startFrame && frame < scene.endFrame
);
```

这样当口播稿、场景时长或字幕发生变化时，可以修改数据而不是重写组件。批量视频也可以把同一个 Composition 和不同的输入 JSON 组合起来。

## 五、字幕和口播怎样同步

字幕同步有三种常见精度：句级、词级和逐字级。

句级字幕实现成本最低，适合信息密度不高的视频；词级字幕更适合口播和教程；逐字级字幕需要更准确的时间戳，也更容易暴露 ASR 错误。

字幕数据可以是：

```json
[
  {"text": "今天讲代码视频", "start": 0.0, "end": 2.1},
  {"text": "重点是逐帧渲染", "start": 2.1, "end": 4.5}
]
```

转换成帧时：

```ts
const startFrame = Math.round(start * fps);
const endFrame = Math.round(end * fps);
const visible = frame >= startFrame && frame < endFrame;
```

不要让 React 组件自行猜当前句子。字幕的起止时间应该来自统一数据，音频、字幕和画面都引用同一份时间轴。

## 六、全身出镜和录屏切换

一个常见的视频结构是：

```text
全身出镜：真人视频全屏
录屏讲解：录屏全屏，真人缩成左下角
章节卡：全屏图形动画
回到出镜：真人视频全屏
```

Remotion 可以把这些段落设计成可组合的组件。关键是提前定义每段的帧范围和层级：

```text
0-180 帧：TalkingHeadFull
180-720 帧：ScreenRecording + TalkingHeadSmall
720-900 帧：ChapterCard
```

如果切换点来自口播或字幕，不要凭感觉填秒数。先生成时间轴清单，确认每段视频实际长度，再由组件使用帧范围。

## 七、数据视频和批量生成

Remotion 适合把画面结构和数据分离：

```json
{
  "name": "张三",
  "metric": 82,
  "message": "本周完成了 12 个任务",
  "avatar": "avatars/zhang-san.png"
}
```

同一个组件可以读取不同 JSON，生成不同版本。适合：

- 每个客户一条报告视频；
- 每个城市一条数据视频；
- 每个产品一条发布视频；
- 每个用户一条个性化通知；
- 每周榜单和数据栏目。

批量渲染前要先处理输入校验：姓名为空、数字格式错误、头像不存在和文本过长，都应该在渲染前被发现，而不是等生成几百条视频后再逐个排查。

## 八、本地渲染和分布式渲染

本地渲染适合调试和少量成片。它的优点是素材、日志和错误都在当前机器，修改反馈更直接。

当视频数量或时长增加时，可以考虑分布式渲染。Remotion 提供云端渲染相关能力，但并行并不会自动解决所有问题：

- 素材需要被远程任务访问；
- 字体和浏览器环境必须一致；
- 每个分片的输出要能正确合并；
- 失败任务需要重试和去重；
- 云端费用与并发、时长和存储有关；
- 视频内容可能涉及隐私和授权。

云端渲染适合批量生产，但应该先用固定样片验证完整流程，不能第一次就把大批任务提交到远程环境。

## 九、Remotion 的 AI 使用边界

AI 写 React 比写简单 HTML 更容易遇到错误：

- JSX 括号和组件嵌套不匹配；
- Hook 使用位置不正确；
- 帧率和秒数混用；
- 视频组件层级和尺寸错误；
- 资源路径在开发和渲染环境不一致；
- 数据类型不符合组件预期。

所以给编码 Agent 的任务应该分阶段：

```text
第一轮：只创建 Composition 和静态组件。
第二轮：接入固定数据，不加入视频素材。
第三轮：接入底层 MP4，验证时长和尺寸。
第四轮：加入字幕和 overlay。
第五轮：运行构建、渲染和帧抽查。
```

每一轮都应该能独立运行。不要让 Agent 一次生成项目、安装依赖、接视频、加字幕、部署和批量渲染，出了问题后很难知道是哪一层失败。

## 十、字体、素材和视频解码

Remotion 最终还是要在确定的浏览器和编码环境中渲染。正式渲染前检查：

- 所有字体是否在渲染环境中存在；
- 视频编码是否可以被当前浏览器读取；
- 透明素材是否保持 Alpha；
- 图片尺寸是否过大；
- 音频是否能被正确读取；
- 文件路径是否跨平台；
- 长视频随机 seek 是否稳定。

开发预览能播放，不代表逐帧渲染一定成功。尤其是外部视频、远程字体和网络图片，最好在正式渲染前复制到项目可控的素材目录。

## 十一、授权与成本要单独核对

Remotion 的版本、云渲染能力和商业授权条款可能变化。个人、团队和企业使用时，不能只根据别人文章里的价格或人数限制做判断，应直接查看当前官方授权页面和项目协议。

同样，底层真人视频、音乐、字体、图片和代码截图也有各自的授权边界。框架能把素材合成视频，不代表你自动获得了素材的发布权。

## 十二、验收 Remotion 视频

至少检查：

- Composition 的宽高、帧率和时长；
- 底层视频没有提前结束或黑帧；
- overlay 和字幕在正确帧出现；
- 录屏与真人小窗切换位置正确；
- 文本没有超出安全区域；
- 音频时长和画面时长匹配；
- 批量输入中缺失字段已被拒绝；
- 失败渲染可以定位到具体输入和场景。

## 结论

Remotion 的核心优势是把已有视频、React 组件、字幕、图表和数据放进同一套帧级渲染模型里。

如果你的底层是一段真人口播或录屏，而前景还要持续叠加动态内容，Remotion 往往比纯 HTML 画面框架更自然。它的代价是 React 构建、依赖、渲染时间和版本维护都更复杂。

它适合做成模板资产，而不是每次从零写一套组件。先把一个真实的视频场景跑通，再抽取字幕、标题卡、数据图表和视频布局，批量生产才会有稳定基础。

## 参考资料

- [Remotion 官方文档](https://www.remotion.dev/docs/)，用于核对项目创建、Composition 和渲染流程。
- [OffthreadVideo 文档](https://www.remotion.dev/docs/offthreadvideo)，用于核对视频素材导入和逐帧处理方式。
- [Remotion License 文档](https://www.remotion.dev/license)，用于核对当前授权和商业使用条件。
