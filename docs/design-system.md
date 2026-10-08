# 原生版 UI 设计规范

适用对象：`ui/` 组件库 + `internal/board/`（业务组装层，MyGo 原生 UI，无 WebView）。

代码里的每个取值都来自 `ui/theme/`，本文与它一一对应 —— 改设计改 `ui/theme/theme.go`，
本文跟着改。组件本身的规范在 `ui/<domain>/doc.go`，本文只说**为什么这样设计**，
以及每条规则落在哪个组件上（见 §14）。

配套的 WebView 版（`web/assets/app.css`）视觉基准相同，本文的「与 WebView 版的差异」一节说明
两版**允许**不一样的地方。

---

## 1. 设计原则

**面板在下，卡片在上。** 整个界面只有两层表面关系：白底（`Canvas`）和灰底（`Panel`）。
侧栏、看板列、图标块是灰；卡片、图标栏、抽屉是白。谁在谁上面，靠这一层关系说清楚，不靠阴影堆。

**颜色少而固定。** 除状态色外不引入品牌色。界面里只有三个"有颜色"的地方：
优先级的红/橙、在线状态的 lime、告警角标的琥珀。其余全是黑白灰。

**数字永远跟着数据走。** 头部统计、侧栏计数、列头数量都由 store 现算，视图不做缓存、不做估算。
一个筛选不会改变头部那行数字（它描述的是整块板，不是一屏）。

**每个控件都能被读出来。** 原生版的意义在于真的控件：Tab 能走到、读屏能念、暗色跟随系统。
所以每个只有图标的元素都必须有 `Label`。

---

## 2. 色彩

### 2.1 主题角色（跟随系统）

角色名和直觉相反，**这是最容易搞错的地方**：

| Token | 浅色主题 | 深色主题 | 用在哪 |
|---|---|---|---|
| `Background` | `#ffffff` 白 | `#111113` 近黑 | 图标栏、卡片、主内容区、抽屉 |
| `Surface` | `#f4f4f5` 浅灰 | `#1c1c1f` 深灰 | 侧栏、看板列、图标块、计数药丸 |
| `SurfaceHover` | `#ececee` | `#26262a` | 悬停态 |
| `SurfacePressed` | `#e4e4e7` | `#2f2f34` | 按下态 |
| `Border` | `#d9d9de` | `#33333a` | 边框、分组分隔线、选中行的计数药丸 |
| `Text` | `#18181b` | `#f4f4f5` | 标题、卡片客户名、金额 |
| `TextMuted` | `#6b6b74` | `#a1a1aa` | 次级文字：工单号、技师名、日期 |
| `TextFaint` | `#8b8b95` | `#71717a` | 分隔符等**记号**，不承载阅读 |

> ⚠️ **反了会怎样**：把列做成白底、卡片做成灰底，卡片就"陷进"列里，整块板发平。
> 这是实际踩过的坑 —— 卡片必须是 `Background`，列必须是 `Surface`。

> ⚠️ **不要读 `ui.Theme.Inverse`**：MyGo 文档写明该字段留零时由工具包从 `Text`/`Background`
> 推导，直接读字段拿到的是**透明色**。深色填充面一律用 `Fill` / `OnFill`。

**对比度是有测试守着的。** `ui/theme/theme_test.go` 断言正文文字在它出现的每一种表面上
≥ 4.5:1、状态药丸的前景/背景 ≥ 4.5:1、边框 ≤ 2:1（边框是边，不是线）。改色值跑一次测试就知道
有没有破坏它，不需要靠眼睛比对截图。

### 2.2 状态色（成对的底 + 字）

| 角色 | 浅色底 / 字 | 深色底 / 字 | 用在哪 |
|---|---|---|---|
| `DangerBg` / `Danger` | `#fdecec` / `#c62b30` | `#3a1f21` / `#ff6b6f` | Critical 药丸、Repeat failures 计数 |
| `WarningBg` / `Warning` | `#fdf1e6` / `#9a5410` | `#3a2a17` / `#e5a04a` | High 药丸、告警角标 |
| `SuccessBg` / `Success` | `#e8f5e9` / `#1f7a3d` | `#16301d` / `#5fc47c` | 已解决 |
| `AccentBg` / `AccentText` | `#e8effd` / `#1d4ed8` | `#16233d` / `#9dc0ff` | 系统强调色，只在控件需要时 |
| `Lively` | `#d9f24b` | `#d9f24b` | 唯一不随主题变的亮色：在线点、空态插图的加号 |

`core.Severity`（Neutral / Info / Success / Warning / Danger）是**唯一**决定
「哪个底配哪个字」的地方，`Severity.Pair(k)` 返回一对。业务代码不自己挑颜色。

### 2.3 深色填充面

| Token | 浅色 | 深色 | 用在哪 |
|---|---|---|---|
| `Fill` | `#1a1a1f` | `#f4f4f5` | 图标栏激活态、主按钮、Toast、头像底、开关滑块 |
| `OnFill` | `#ffffff` | `#18181b` | 上述元素上的文字 |

判深浅用 `Tokens.IsDark()`，实现是比较两个表面的亮度。

---

## 3. 字体与排版

系统字体，负字距收紧标题。半号字号是有意为之：13.5 的次级标签比 13 更立得住，
15.5 能让卡片标题晚一点折行。

| Token | 值 | 用途 |
|---|---|---|
| `DisplaySize` | 40 | 页面大标题 "Open callbacks" |
| `TitleSize` | 24 | 侧栏 "Callbacks" |
| `SheetSize` | 21 | 抽屉标题 |
| `LeadSize` | 17 | 空态标题 |
| `BodySize` | 15.5 | 卡片客户名、列表客户名 |
| `StatSize` | 14.5 | 头部统计行、金额 |
| `RowSize` | 13.5 | 筛选行、侧栏标签、列标题 |
| `MetaSize` | 13 | 卡片副行、表格单元 |
| `CaptionSize` | 12 | 药丸、计数、优先级 |
| `MonoSize` | 11.5 | 头像缩写 |
| `GlyphSize` | 11 | 信号格 |
| `IconSize` | 21 | 图标尺寸（21×21） |

字重：只有 `Bold()`（700）和常规两档。表格表头用 12 + Bold。

---

## 4. 间距与圆角

**间距**走 4 的网格。网格的一步就是 `core.Density(c).Unit()`（`Compact` = 4，
`Comfortable` = 6）—— 组件内部所有的 `u*n` 里的 `n` 都是间距倍数，调用方改一个数字，
整块板的呼吸就跟着变。需要视觉修正的地方再加 1（比如侧栏行高用 10 而不是 12）。
窗口级留白（22 / 28）不是网格的整数倍，属于有意的例外。

**圆角**一个刻度族，大值即药丸：

| Token | 值 | 用在哪 |
|---|---|---|
| `ControlRadius` | 14 | 图标块、筛选行、空态插图底 |
| `CardRadius` | 18 | 卡片、空态卡 |
| `PanelRadius` | 24 | 看板列 |
| `PillRadius` | 999 | 药丸、开关轨道、计数、"N more" |

**别用 `Radius` 的四值形式**，除非确实要做非对称圆角 —— 混用会让曲率看起来不属于同一套。

---

## 5. 布局骨架

```
┌──────┬────────────┬──────────────────────────────────────┐
│ rail │  sidebar   │  main（Canvas 白底）                 │
│  76  │    322     │                                      │
│      │            │  header：面包屑 / 大标题 / 统计 /    │
│      │            │          头像组 + 工具行              │
│      │            ├──────────────────────────────────────┤
│      │            │  board：4 列，列间距 16，列宽 272    │
│      │            │  横向滚动，不压缩卡片                 │
└──────┴────────────┴──────────────────────────────────────┘
```

| 区域 | 尺寸 | 关键约束 |
|---|---|---|
| 图标栏 | 76 宽 | 白底（Canvas），图标块 44×44 / `rMd`，间距 10 |
| 侧栏 | 322 宽 | 灰底（Panel），内边距 22/18/18/18，行间距 14 |
| 主区 | 剩余 | 白底，横向留白 28（`sp6`） |
| 看板列 | **272 宽** | 272 是卡片标题不折行的下限；窗口不够就横向滚动，**不压缩** |
| 列头 | 高约 60 | 标题 + 计数药丸 + "···" 按钮 |
| 列底 | — | "N more" 吸底（`margin-top: auto` 语义，用 Flex 吸底实现） |

> ⚠️ **横向滚动的坑**：外层是 `ScrollHorizontal` 时，内层 Row 必须给 `FillHeight()`，
> 否则列的 `FillHeight` 解析成 0，**卡片整片消失**，只剩列头和 "N more"。

---

## 6. 组件规范

### 6.1 图标栏 — `navigation.Rail`

- 头像 44×44 圆形，`Fill` 底 + `OnFill` 的缩写
- 在线点 11×11 `Lively`，2px `Background` 描边，`Attach(AnchorBottomRight)` 贴住头像右下
- 导航项 44×44 / `ControlRadius`；未选中 `Surface` 底 + `TextMuted` 图标，选中 `Fill` + `OnFill`
- 告警角标 9×9 琥珀圆点，2px 描边贴图标块右上
- 底部（告警 / 设置 / 头像 / 折叠）由 `Spacer` 推到最下
- 调用方只给两组 `RailItem`：上组是「我在哪」，下组是「设置类」，中间的空白是组件自己的事

### 6.2 筛选行 — `navigation.FilterRow`

- 高约 44，内边距 10/14，`ControlRadius`
- **只有选中行有底色**（`Background`），未选中行完全透明，直接坐在侧栏灰底上
  —— 否则侧栏会变成一摞药丸
- 选中时文字 `Text`、计数药丸底 `Surface`；未选中文字 `TextMuted`、药丸底 `Border`
- 计数由调用方给 `*int`；nil 就不画药丸。`Repeat failures` 传红色的那个值 —— 唯一一个彩色计数
- 子项（Views / Branches 里的）调用方传 `Indent: true`，组件加一段 `—` 前缀表示层级

### 6.3 分组（Views / Branches）— `navigation.Group`

可折叠。标题行 10/12 内边距 + 图标 + 标题 + 右侧 chevron；展开时下方是筛选行的子集。
`Group` 不自己存折叠状态 —— 它从 `GroupOptions.Open` 读，把标题的按压通过
`GroupResult.Toggled()` 交回去，由 `App.open`（按分组 id）决定下一帧的 `Open`。
组件没有状态，所以展开时渲染出的子项一定和 `Open` 一致。

### 6.4 头部 — `layout.PageHeader` + `data.StatLine` + `display.AvatarCluster`

- 面包屑：`Callbacks / <当前筛选名>`，`MetaSize`，`TextMuted`
- 大标题 40pt Bold（`DisplaySize`）
- 统计行 14.5pt：数字用 `Text` 加粗、单位用 `TextMuted`，整行一句话读出来
  （`data.StatLine` 用 `ui.RichText`，并且给整行一个无障碍标签）
- 头像组：30 圆形、`Surface` 底、2px `Background` 描边、负间距叠放
  （`AvatarCluster(c, names, total)`：画前 5 个，第 6 个位置画「还剩几个」的数字圆）
- 工具行右对齐：Group by branch 药丸 + 分段控件 + 主按钮 —— 这三个是**页面自己的**控件，
  调用方写在 `HeaderOptions` 之外，不进组件

### 6.5 卡片 — `data.Card`

- `CardRadius`，白底（`Background`），内边距 15/16/14/16，`shadowCard`
- 顶行：优先级药丸（可 Grow 撑开）+ 日期药丸靠右
- 客户名 15.5 Bold，单行省略号
- 副行 `CB-2871 · AC repair` 13pt `TextMuted`，单行省略号
- 底行：28 圆形头像 + 技师名（Grow，省略号）+ 金额右对齐 14.5 Bold
- **拖拽源**：`CardOptions.Draggable` 给一个值（回调 id）就变成拖拽源，`CursorGrab`

`Card` 有三个槽位：`Title` / `Meta` 是文本槽，`Footer func()` 是调用方自己拼的底行
（卡片库不知道技师和金额是什么），`body func()` 夹在副行和底行之间。
底行默认**不画分割线** —— 一行元数据不需要一条线把它和正文分开；
真要放操作时才传 `FooterRule: true`。

### 6.6 优先级药丸 — `core.Severity` + `display.Meter`

| 优先级 | Severity | 信号格 |
|---|---|---|
| Low | Neutral | 1 实 3 虚 |
| Medium | Neutral | 2 实 2 虚 |
| High | Warning | 3 实 1 虚 |
| Critical | Danger | 4 实 |

底和字由 `Severity.Pair(k)` 给，业务代码不挑颜色。信号格是 `display.Meter` 画的
（`GlyphSize` 高的竖条，不是字符 `▮`/`▯`，所以在任何字体下都是同一根线），
**宽度固定为 7 个单位**并且和文字一起放在一个 `Row` 里 —— 它是药丸里的一个记号，
不是一个会撑满整行的元素。`level` 会被 clamp 到 `[0, n]`，`n == 0` 直接 panic。

### 6.7 看板列 — `layout.Column`

- `PanelRadius` 灰底（`Surface`），内边距 16/14
- 列头：标题（单行）+ 计数药丸 + 38 圆形 "···" 按钮
- **卡片区外面套 `Scroll`**；但空列（`ColumnOptions.Empty` 有值时）**不套** ——
  空状态没有可滚内容，而且套上之后按钮命中不到
- "N more" 吸底，48 高，`Border` 底，按压通过 `ColumnResult.Revealed()` 交回调用方
- **列宽固定 272**（`theme.ColumnWidth`）。窗口不够就横向滚动，**不压缩**：
  卡片标题一折行，整块板就不能扫了

`Column` 是泛型的：`ColumnOptions[T]` / `ColumnResult[T]`，因为一条泳道要明确自己接什么
类型的东西。`Dropped()` 返回 `(T, bool)` 而不是 `any`，调用方拿到手就是对的类型。

### 6.8 空状态 — `feedback.Empty`

一张白卡（`CardRadius`）里依次是：

1. 插图区：150 高，`ControlRadius`，`Surface` 底 —— 三块圆角小卡片 + 一个 `Lively` 圆形加号
2. 标题 17pt Bold "Nothing resolved yet today"
3. 说明 13.5 `TextMuted`
4. `Resolve a callback` 深色药丸按钮（`Fill` / `OnFill`），满宽

插图是 `Empty` 自己的默认图；调用方要换就传 `Art`。按钮的按压走
`EmptyResult.Pressed()` —— 布局层不需要知道「没有东西」长什么样。

### 6.9 列表视图

六列：Customer（Grow）/ Issue（Grow）/ Priority（120）/ Technician（120）/ Due（120）/ Cost（Grow，右对齐）。
行高约 48，行间一条 `d.Line` 下边线；悬停时整行铺 `d.Panel`。行同样可拖拽改列。
表格头 12pt Bold `d.InkMuted` 吸顶。

### 6.10 抽屉（Log callback）

`ui.Modal` + `*bool`：点外面或按 Esc 关闭，打开时接管焦点。字段用 `ui.Form` / `ui.Field`
（标签在左、控件在右，与 macOS 系统设置一致）：Customer、Issue、Priority（分段）、
Branch（分段）、Open cost、Technician（下拉）。

主按钮文案是 **Save** 而不是再来一次 "Log callback" —— 抽屉标题已经说了在做什么，
而且背景上那个同名按钮此时是失效的（模态吞掉背景点击）。

### 6.11 Toast — `feedback.Toast`

底部居中，`Fill` 底 / `OnFill` 字，`PillRadius` 药丸，`shadowFloat`，2.2 秒后自动消失。

**高度只占内容那一行。** 用 `FillWidth()` 而不是 `Fill()`：早期版本用 `Fill()`，
toast 那一行跟主内容抢高度，把整块板挤扁了。这是实际踩过的坑。

---

## 7. 状态

| 状态 | 表现 |
|---|---|
| 悬停（卡片） | `shadowCard` → `shadowCardUp` |
| 拖拽中（卡片） | 透明度 0.45，`CursorGrabbing` |
| 拖拽悬停（列） | 列边框 2px `d.InkMuted` |
| 选中（筛选行） | `d.Canvas` 底 + `d.Ink` 字 |
| 悬停（图标块） | 底色加深到 `d.Hover` |
| 焦点 | 用系统焦点环，不自定义 |
| 禁用 | 交给系统控件表现 |

阴影只有四档（`shadowCard` / `shadowCardUp` / `shadowKnob` / `shadowFloat`），不要临时造第五档。

---

## 8. 交互契约

**视图是纯函数。** 每帧从状态重建 UI，状态只存在 App 里；MyGo 按位置或 `Key` 保留焦点、悬停、
滚动、动画这些逐元素状态。

**事件是提问。** `if row.Clicked() { … }` 写在构建该元素的代码旁边。

**改动要推迟到帧末。** 视图里遍历切片时改它会让迭代失效，所以所有状态变更走
`a.deferDo(fn)`，在一帧构建完之后统一执行。**执行完必须 `c.Invalidate()` 请求下一帧** ——
否则界面停在旧状态（这是实际踩过的坑）。

**每帧都要给元素命名。** 兄弟元素顺序会变的地方给 `Key`，让 MyGo 能把状态跟着走。

---

## 9. 无障碍

- 只有图标的元素必须有 `Label`；带文字的按钮也要 `Label`，因为测试器和读屏都按名字找
- 侧栏"+"叫 **New callback**，头部按钮叫 **Log callback**，抽屉提交叫 **Save** ——
  三个名字不许撞车，否则自动化点到的不是你以为的那个
- 表格是有语义的行，不是画出来的网格
- 键盘：Tab 按视觉顺序走；抽屉里循环；Esc 关闭

---

## 10. 响应式

| 断点 | 行为 |
|---|---|
| ≥ 1590 | 四列并排可见（4×272 + 间距 + 留白） |
| < 1590 | 看板横向滚动，列宽不变 |
| 窗口最窄 900 | 由 `WindowOptions.MinWidth` 兜底 |

**列宽是常量不是比例。** 卡片标题在 272 以下会折行，一折行整列的信息密度就塌了，
宁可滚动也不压缩。

---

## 11. 与 WebView 版的差异

同一份数据、同一套数字，以下差异是**有意保留**的：

| | WebView 版 | 原生版 |
|---|---|---|
| 暗色模式 | 无（CSS 写死浅色） | 跟随系统 |
| 筛选/搜索框 | 自绘胶囊 | 系统控件 |
| 空态插图 | 手绘 SVG（卡片+光标+加号） | 圆角色块简化版 |
| 卡片阴影 | 稍重 | 稍轻（`shadowCard`） |
| 主按钮 | 近黑 | 近黑（刻意不用系统强调蓝） |

反过来，**不允许**的差异：数字、计数、筛选语义、预览张数（3/4/3）、"N more" 措辞。

---

## 12. 反模式（都是真踩过的）

1. ❌ 直接读 `t.Inverse` 拿深色面 —— 那是透明色，深色面一律用 `d.Fill`
2. ❌ 卡片用 `d.Panel`、列用 `d.Canvas` —— 反了，整块板发平
3. ❌ 在 `ScrollHorizontal` 里不给内层 Row 的高度 —— 列高塌成 0，卡片消失
4. ❌ 把空状态包在 `Scroll` 里 —— 按钮点不到
5. ❌ 推迟执行状态变更后忘了 `c.Invalidate()` —— 界面不刷新
6. ❌ 从 label 反推筛选 id —— "High impact" 会变成 "high"，点了没反应，筛选 id 要显式传
7. ❌ 用 tab 拼 `<id>\t<姓名>` 给 `ui.Select` —— Select 不认这个分隔符，显示成 "atAndre Thom…"
8. ❌ 分组模式下先整体裁剪再分组 —— 子标题会承诺没有渲染的卡片；要按组各自裁剪
9. ❌ 三个控件用同一个 `Label` —— 自动化和读屏都会点错
10. ❌ Toast 用 `Fill()` 而不是 `FillWidth()` —— 它会跟主内容抢高度
11. ❌ `navigation.Rail` 忘了 `.Children(...)` —— 图标挂到父 Row 上去，图标栏整个错位
12. ❌ 模态没打开也调用 `ui.Modal` —— 它仍然参与布局，关着的抽屉会把主区挤窄
13. ❌ 在 `Column` 里用 `ui.DragOver[any]` —— 载荷是 `any`，调用方每次都要断言；
    `Column[T]` 把它变成 `Dropped() (T, bool)`

---

## 13. 怎么验证

```bash
go test ./internal/board/     # 16 个无头渲染测试
go run ./cmd/snapshots -out shots   # 导出 PNG
```

截图用的是 MyGo 自己的渲染器（`ui.Tester.Image()`），不开窗口、不需要 macOS，
所以改完 UI 立刻能看效果。上面每一条反模式都是被这套流程抓出来的。
---

## 14. 组件 API 与实现位置

设计规则落在哪个包、哪个函数上。**每帧的第一个调用是 `core.Use`**，其余全是组装。

### 14.1 包与职责

| 包 | 职责 | 放什么 | 不放什么 |
|---|---|---|---|
| `ui/theme` | 看起来是什么样 | 两套 `Tokens`、密度档、字号、圆角、`ColumnWidth` | 任何绘制、任何窗口读取 |
| `ui/core` | 这一帧解析成什么 | `Use` / `Tokens` / `Density` / `IsDark` / `Severity` | 业务语义 |
| `ui/display` | 不需要文字就读得懂的记号 | `Avatar`、`AvatarCluster`、`Meter` | 交互 |
| `ui/input` | 人的操作 | `Button`、`IconButton`、`Segmented`、`SearchField`、`Switch` | 业务含义 |
| `ui/navigation` | 在窗口里移动 | `Rail`、`FilterRow`、`Group` | 筛选逻辑 |
| `ui/layout` | 窗口的骨架 | `PageHeader`、`Board`、`Column` | 数据 |
| `ui/data` | 一条记录长什么样 | `Card`、`Stat`、`StatLine` | 记录的业务含义 |
| `ui/feedback` | 应用**说**而不是**展示** | `Empty`、`Toast` | 视图 |
| `internal/board` | 业务组装 | callbacks 数据、交互状态、模态 | 任何绘制原语 |

### 14.2 调用形态

三行之内能看出一个视图的全部：

```go
core.Use(c, core.Settings{})                     // 每帧一次
k, u := core.Tokens(c), core.Density(c).Unit()   // 调色板 + 间距一步
// ……组件组装……
```

三条约定：

1. **组件不持有状态。** 选中的东西是调用方的一个指针
   （`input.Segmented(c, &mode, "▦", "☰")`），按压是通过返回值的方法交回来的
   （`Button(...).Clicked()`、`ColumnResult.Revealed()`、`EmptyResult.Pressed()`）。
   没有 `OnChange`，也就没有「变量和回调不同步」这种状态。
2. **参数走 `XxxOptions`。** 组件返回 `XxxResult`（有交互的）或 `*ui.Element`（没有交互的）。
   `Result` 上只暴露一个方法，对应用户能做的那一个动作。
3. **给不了的东西就 panic，不猜。** 没起名字的按钮、越界的分段选中、
   没调 `core.Use` 就读主题、`Meter(0, …)` —— 都 `panic("pkg: …")`。
   一帧悄悄画错，比直接停下来难找得多。

### 14.3 组件索引

签名里的 `c` 和返回值从略；参数是 `XxxOptions` 的写 `(opts)`。

**`ui/core`** — 每帧解析一次

- `Use(s Settings) → *ui.Element`
- `Tokens(—) → theme.Tokens`
- `Density(—) → theme.Density`
- `IsDark(—) → bool`
- `Msg(key, def string) → string`
- `WithMessages(m map[string]string) → *ui.Context`
- `Messages(—) → map[string]string`
- `Motion(d float32) → float32`
- `Reduced(—) → bool`
- `WithReducedMotion(on bool) → *ui.Context`
- `FontSize(size float32) → float32`
- `WithTextScale(scale float32) → *ui.Context`
- `ControlHeight(—) → float32`

**`ui/layout`** — 容器与窗口骨架

- `Board(columns func() → ) *ui.Element`
- `Container(opts ContainerOptions, children func() → ) *ui.Element`
- `Divider(opts DividerOptions) → *ui.Element`
- `Stack(opts StackOptions, children func() → ) *ui.Element`
- `Grid(opts GridOptions, children func() → ) *ui.Element`
- `GridCell(opts GridCellOptions, build func() → ) *ui.Element`
- `ScrollArea(opts ScrollAreaOptions, children func() → ) ScrollResult`
- `SplitPane(opts SplitPaneOptions, first, second func() → ) SplitPaneResult`
- `AspectRatio(opts AspectRatioOptions, child func() → ) *ui.Element`
- `TitleBar(opts TitleBarOptions) → *ui.Element`
- `TrafficLights(—) → *ui.Element`
- `StatusBar(opts StatusBarOptions) → StatusBarResult`
- `AppShell(collapsed *bool, opts AppShellOptions, sidebar, main, status func() → ) AppShellResult`
- `PageHeader(opts HeaderOptions) → *ui.Element`
- `AvatarSlots(names []string, total int) → *ui.Element`
- `Panel(host *ui.Element, opts PanelOptions, body func() → ) *ui.Element`

**`ui/display`** — 记号与文字

- `Avatar(name string) → *ui.Element`
- `AvatarCluster(names []string, total int) → *ui.Element`
- `Meter(n, level int, severity core.Severity) → *ui.Element`
- `AvatarGroup(names []string, opts AvatarGroupOptions) → AvatarGroupResult`
- `Badge(count int, opts BadgeOptions) → *ui.Element`
- `PresenceDot(opts BadgeOptions) → *ui.Element`
- `Icon(name IconName, opts IconOptions) → *ui.Element`
- `Image(src *ui.Bitmap, opts ImageOptions) → *ui.Element`
- `Heading(title string, opts HeadingOptions) → *ui.Element`
- `Text(s string, opts TextOptions) → *ui.Element`
- `EditableText(value *string, opts EditableTextOptions) → EditableTextView`
- `Kbd(keys ...string) → *ui.Element`
- `Link(label string, opts LinkOptions) → *ui.Element`
- `Tag(label string, opts TagOptions) → TagView`
- `IconClose(u float32) → *ui.Element`

**`ui/input`** — 人操作的控件

- `BulkActionBar(chosen *bool, selected []string, opts BulkActionBarOptions) → BulkActionBarResult`
- `Dock(open *bool, opts DockOptions, body func() → ) DockResult`
- `Button(label string, opts ButtonOptions) → *ui.Element`
- `IconButton(glyph *ui.SVG, name string, opts ButtonOptions) → *ui.Element`
- `ButtonGroup(selected *string, buttons []GroupButton, opts ButtonGroupOptions) → *ui.Element`
- `Checkbox(state *CheckState, label string, opts CheckboxOptions) → *ui.Element`
- `CheckboxGroup(selected *[]string, choices []Choice, opts CheckboxGroupOptions) → *ui.Element`
- `RadioGroup(selected *string, choices []Choice, opts RadioGroupOptions) → *ui.Element`
- `ToggleGroup(selected *int, labels []string, opts ToggleGroupOptions) → *ui.Element`
- `ChoiceChips(selected *[]string, choices []Choice, opts ChoiceChipsOptions) → *ui.Element`
- `Segmented(selected *int, labels ...string) → *ui.Element`
- `SearchField(query *string, placeholder string) → *ui.Element`
- `Switch(on *bool, opts SwitchOptions) → *ui.Element`
- `FilePicker(path *string, entries []FileEntry, opts FilePickerOptions) → FilePickerResult`
- `FileDropZone(opts FileDropZoneOptions) → FileDropZoneResult`
- `Form(opts FormOptions) → FormResult`
- `FormField(opts FormFieldOptions, control func(err string) → *ui.Element) *ui.Element`
- `Banner(opts BannerOptions) → BannerResult`
- `InputGroup(value *string, opts InputGroupOptions) → *ui.Element`
- `CopyButton(text *string, opts CopyButtonOptions) → CopyButtonResult`
- `HoldToConfirm(fired *bool, opts HoldToConfirmOptions) → HoldToConfirmResult`
- `Knob(value *float64, opts KnobOptions) → KnobResult`
- `VerticalSlider(value *float64, opts VerticalSliderOptions) → VerticalSliderResult`
- `Masonry(items []MasonryItem, opts MasonryOptions) → *ui.Element`
- `MentionInput(value *string, opts MentionOptions) → MentionResult`
- `NumberInput(value *float64, opts NumberInputOptions) → NumberInputResult`
- `CurrencyInput(amount *float64, opts CurrencyInputOptions) → CurrencyInputResult`
- `Rating(value *int, opts RatingOptions) → *ui.Element`
- `ColorPalette(selected *ui.Color, swatches []Swatch, opts ColorPaletteOptions) → *ui.Element`
- `ColorPicker(color *ui.Color, opts ColorPickerOptions) → *ui.Element`
- `TagsInput(tags *[]string, opts TagsInputOptions) → *ui.Element`
- `QRCode(text string, opts QRCodeOptions) → QRCodeResult`
- `ResizablePanelGroup(opts ResizablePanelGroupOptions, panes ...func() → ) *ui.Element`
- `ScrubInput(value *float64, opts ScrubOptions) → ScrubResult`
- `SignaturePad(strokes *[]Stroke, opts SignaturePadOptions) → SignaturePadResult`
- `Select(selected *string, choices []Choice, opts SelectOptions) → *ui.Element`
- `SelectSearch(selected, query *string, choices []Choice, opts SelectSearchOptions) → *ui.Element`
- `MultiSelect(selected *[]string, choices []Choice, opts MultiSelectOptions) → MultiSelectResult`
- `Cascader(path *[]string, nodes []Node, opts CascaderOptions) → *ui.Element`
- `Combobox(selected *string, suggestions []string, opts ComboboxOptions) → *ui.Element`
- `TreeSelect(selected *string, nodes []Node, opts TreeSelectOptions) → *ui.Element`
- `Slider(value *float64, opts SliderOptions) → *ui.Element`
- `RangeSlider(low, high *float64, opts RangeSliderOptions) → *ui.Element`
- `TextInput(value *string, opts TextInputOptions) → *ui.Element`
- `TextArea(value *string, opts TextAreaOptions) → *ui.Element`
- `PasswordInput(value *string, opts PasswordInputOptions) → *ui.Element`
- `SearchInput(query *string, opts SearchInputOptions) → SearchInputResult`
- `MaskedInput(value *string, opts MaskedInputOptions) → *ui.Element`
- `PinInput(value *string, opts PinInputOptions) → PinInputResult`
- `Transfer(left, right *[]string, opts TransferOptions) → TransferResult`

**`ui/navigation`** — 在窗口里移动

- `FilterRow(text string, opts FilterRowOptions) → *ui.Element`
- `Group(opts GroupOptions, body func() → ) GroupResult`
- `Menubar(app string, menus []Menu, opts MenubarOptions) → MenubarResult`
- `NavigationMenu(triggerLabel string, items []NavigationMenuItem,
	opts NavigationMenuOptions) → NavigationMenuResult`
- `CommandPalette(open *bool, items []Command, opts CommandPaletteOptions) → CommandPaletteResult`
- `Rail(identity string, items []RailItem, tools []RailItem) → *ui.Element`
- `Sidebar(sections []SidebarSection, opts SidebarOptions) → SidebarResult`
- `Tabs(items []Tab, opts TabsOptions) → *ui.Element`
- `Breadcrumb(items []BreadcrumbItem) → BreadcrumbResult`
- `Steps(steps []Step, opts StepsOptions) → *ui.Element`
- `Pagination(opts PaginationOptions) → PaginationResult`
- `Toolbar(opts ToolbarOptions, children func() → ) *ui.Element`
- `ToolbarSeparator(—) → *ui.Element`
- `ToolbarGroup(opts ToolbarGroupOptions, children func() → ) *ui.Element`
- `ToolbarButton(label string, opts ToolbarButtonOptions) → *ui.Element`

**`ui/data`** — 记录与集合

- `Accordion(opts AccordionOptions, sections ...Section) → AccordionResult`
- `Collapsible(opts CollapsibleOptions, body func() → ) CollapsibleResult`
- `Card(opts CardOptions, body func() → ) *ui.Element`
- `StatLine(stats ...Stat) → *ui.Element`
- `Carousel(opts CarouselOptions) → *ui.Element`
- `DescriptionList(opts DescriptionListOptions, terms ...Term) → *ui.Element`
- `ImageViewer(opts ImageViewerOptions) → ImageViewerResult`
- `List(opts ListOptions, row func(row int) → ) *ui.Element`
- `DataTable(opts DataTableOptions) → DataTableResult`
- `Timeline(opts TimelineOptions, events ...Event) → *ui.Element`

**`ui/overlay`** — 浮层

- `Dialog(open *bool, opts DialogOptions) → *ui.Element`
- `AlertDialog(open *bool, opts AlertDialogOptions) → AlertDialogResult`
- `Drawer(open *bool, opts DrawerOptions) → *ui.Element`
- `HoverCard(anchor *ui.Element, open *bool, opts HoverCardOptions) → *ui.Element`
- `ContextMenu(anchor *ui.Element, opts ContextMenuOptions) → *ui.Element`
- `DropdownMenu(label string, opts DropdownMenuOptions) → *ui.Element`
- `Panel(host *ui.Element, opts PanelOptions, body func() → ) *ui.Element`
- `Popover(anchor *ui.Element, open *bool, opts PopoverOptions) → *ui.Element`
- `Popconfirm(anchor *ui.Element, open *bool, opts PopconfirmOptions) → PopconfirmResult`
- `SplitButton(label string, opts SplitButtonOptions) → SplitButtonResult`
- `Tooltip(anchor *ui.Element, text string) → *ui.Element`

**`ui/feedback`** — 说而不是展示

- `Alert(opts AlertOptions) → AlertResult`
- `Result(opts ResultOptions) → ResultResult`
- `CountUp(opts CountUpOptions) → *ui.Element`
- `Typewriter(opts TypewriterOptions) → *ui.Element`
- `ActivityFeed(opts ActivityFeedOptions) → *ui.Element`
- `Empty(opts EmptyOptions) → EmptyResult`
- `Toast(message string, u float32) → *ui.Element`
- `BlinkHighlight(opts BlinkHighlightOptions, child func() → ) *ui.Element`
- `InterruptButton(opts InterruptButtonOptions) → InterruptButtonResult`
- `BarsLoader(opts LoaderOptions) → *ui.Element`
- `DotsLoader(opts LoaderOptions) → *ui.Element`
- `OrbitLoader(opts LoaderOptions) → *ui.Element`
- `PulseLoader(opts LoaderOptions) → *ui.Element`
- `WaveLoader(opts LoaderOptions) → *ui.Element`
- `NotificationCenter(opts NotificationCenterOptions) → NotificationCenterResult`
- `UndoToast(opts UndoToastOptions) → UndoToastResult`
- `LoadingOverlay(opts LoadingOverlayOptions, child func() → ) LoadingOverlayResult`
- `Skeleton(opts SkeletonOptions) → *ui.Element`
- `Shimmer(opts ShimmerOptions, child func() → ) *ui.Element`
- `ShimmerText(opts ShimmerTextOptions) → *ui.Element`
- `Progress(opts ProgressOptions) → *ui.Element`
- `StatusIndicator(opts StatusIndicatorOptions) → *ui.Element`
- `Presence(opts PresenceOptions) → *ui.Element`
- `LayoutTransition(opts LayoutTransitionOptions, child func() → ) *ui.Element`
- `NewStagger(opts StaggerOptions) → *Stagger`

**`ui/datetime`** — 日期与时间

- `AgendaView(opts AgendaOptions) → *ui.Element`
- `AttendeeList(opts AttendeeOptions) → *ui.Element`
- `AvailabilityPicker(opts AvailabilityOptions) → AvailabilityResult`
- `DurationPicker(opts DurationPickerOptions) → DurationPickerResult`
- `TimezoneSelect(opts TimezoneSelectOptions) → *ui.Element`
- `Calendar(opts CalendarOptions) → CalendarResult`
- `EventChip(opts EventChipOptions) → EventChipResult`
- `CalendarEvents(opts CalendarEventsOptions) → CalendarEventsResult`
- `Grid(opts GridOptions) → GridResult`
- `DatePicker(opts DatePickerOptions) → DatePickerResult`
- `DateRangePicker(opts DateRangePickerOptions) → DateRangePickerResult`
- `TimePicker(opts TimePickerOptions) → TimePickerResult`
- `TimeRangePicker(opts TimeRangePickerOptions) → TimeRangePickerResult`
- `DateTimePicker(opts DateTimePickerOptions) → DateTimePickerResult`
- `MonthPicker(opts MonthPickerOptions) → MonthPickerResult`
- `YearPicker(opts YearPickerOptions) → YearPickerResult`
- `WeekPicker(opts WeekPickerOptions) → WeekPickerResult`
- `RecurrenceEditor(opts RecurrenceEditorOptions) → RecurrenceEditorResult`
- `CronEditor(opts CronEditorOptions) → CronEditorResult`
- `CalendarTimeGrid(opts TimeGridOptions) → TimeGridResult`
- `CurrentTimeIndicator(opts CurrentTimeIndicatorOptions) → *ui.Element`
- `RelativeTime(opts RelativeTimeOptions) → *ui.Element`
- `Countdown(opts CountdownOptions) → *ui.Element`
- `Stopwatch(opts StopwatchOptions) → StopwatchResult`
- `ReminderPicker(opts ReminderPickerOptions) → ReminderPickerResult`
- `EventPopover(opts EventPopoverOptions) → EventPopoverResult`
- `EventEditor(opts EventEditorOptions) → EventEditorResult`
- `CalendarDayView(opts CalendarDayViewOptions) → CalendarDayViewResult`
- `CalendarWeekView(opts CalendarWeekViewOptions) → CalendarWeekViewResult`
- `CalendarMonthView(opts CalendarMonthViewOptions) → CalendarMonthViewResult`
- `CalendarYearView(opts CalendarYearViewOptions) → CalendarYearViewResult`

**`ui/chart`** — 图表

- `Annotation(opts AnnotationOptions) → *ui.Element`
- `Axis(frame FrameResult, opts AxisOptions) → *ui.Element`
- `Brush(opts BrushOptions) → BrushResult`
- `LineChart(opts LineOptions) → *ui.Element`
- `AreaChart(opts AreaOptions) → *ui.Element`
- `AreaMountain(opts AreaMountainOptions) → *ui.Element`
- `BarChart(opts BarOptions) → *ui.Element`
- `StackedBar(opts StackedBarOptions) → *ui.Element`
- `PercentBar(opts StackedBarOptions) → *ui.Element`
- `Histogram(opts HistogramOptions) → *ui.Element`
- `ScatterChart(opts ScatterOptions) → *ui.Element`
- `BubbleChart(opts BubbleOptions) → *ui.Element`
- `WaterfallChart(opts WaterfallOptions) → *ui.Element`
- `ParetoChart(opts ParetoOptions) → *ui.Element`
- `Sparkline(opts SparklineOptions) → *ui.Element`
- `SparkBar(opts SparkBarOptions) → *ui.Element`
- `Crosshair(opts CrosshairOptions) → *ui.Element`
- `BoxPlot(opts BoxPlotOptions) → *ui.Element`
- `ViolinPlot(opts ViolinOptions) → *ui.Element`
- `Empty(opts EmptyOptions) → *ui.Element`
- `FunnelChart(opts FunnelOptions) → *ui.Element`
- `SankeyChart(opts SankeyOptions) → *ui.Element`
- `AlluvialChart(opts AlluvialOptions) → *ui.Element`
- `ChordDiagram(opts ChordOptions) → *ui.Element`
- `Frame(opts FrameOptions, children func(FrameResult) → ) FrameResult`
- `Measure(opts FrameOptions) → Inset`
- `Grid(opts GridOptions) → *ui.Element`
- `EntriesOf(series []Series) → []LegendEntry`
- `Legend(opts LegendOptions) → *ui.Element`
- `LegendHeight(opts LegendOptions) → float32`
- `Palette(n int) → []ui.Color`
- `PieChart(opts PieOptions) → *ui.Element`
- `DonutChart(opts DonutOptions) → *ui.Element`
- `NightingaleChart(opts NightingaleOptions) → *ui.Element`
- `RadarChart(opts RadarOptions) → *ui.Element`
- `PolarChart(opts PolarOptions) → *ui.Element`
- `Gauge(opts GaugeOptions) → *ui.Element`
- `BulletChart(opts BulletOptions) → *ui.Element`
- `Plot(opts PlotOptions) → *ui.Element`
- `Ticks(s Scale, opts TickOptions) → []Tick`
- `LabelWidth(label string, size float32) → float32`
- `LabelHeight(size float32) → float32`
- `Widest(ticks []Tick, size float32) → float32`
- `Tooltip(opts TooltipOptions) → *ui.Element`
### 14.4 测试怎么覆盖

```bash
go test ./...                                 # 全库，每个组件一个测试文件
go run ./cmd/snapshots -out shots -scale 2   # 真实原生渲染的 PNG
```

- **每个组件一个无头渲染测试**，用 `ui.NewTester` 断言看得见的文字和
  布局结果，不断言内部字段。
- **纯函数必须有精确断言**：`Rows` 的稳定排序、`Ticks` 的刻度位置、
  `Match` 的排序、`Sanitize` 的清洗结果、`HumanSize` 的边界 ——
  断言具体值，不是「大于 0」。
- **断言捕获的变量，不断言点击后画出的文字**，原因见 §17.2。
- **对比度由 `ui/theme/theme_test.go` 守着**：正文 4.5:1、状态药丸 4.5:1、
  边框 ≤ 2:1。改色值不用靠眼睛比截图。
- **每个包一个 `example_<pkg>_test.go`**，都是能跑的示例。


- **每个组件一个 `_test.go`**，用 `ui.NewTester` 无头渲染，断言看得见的文字和
  布局结果，不断言内部字段。
- **每个包一个 `example_<pkg>_test.go`**，都是能跑的示例，同时是这一页的活文档。
- **对比度由 `ui/theme/theme_test.go` 守着**：正文 4.5:1、状态药丸 4.5:1、边框 ≤ 2:1。
  改色值不用靠眼睛比截图。
- **拖拽必须分步移动**：一步跳过去不算拖拽（见 `TestColumnDropsValues` 里的 `drag`）。

---

## 15. MyGo 布局的三个坑（都是实测撞出来的）

写 `ui/layout` 时踩到的，记在这里因为下一个组件很可能再踩一次。

### 15.1 `BasisPercent` 不是「按百分比分」

`BasisPercent(p)` 只是给 flex 一个**起始尺寸**，之后两个 `Grow(1)` 的元素会把剩下的
**平分**。结果是设了 25% / 75% 却拿到 50% / 50%。

按比例分一块区域要用 **`WidthPercent(p)` / `HeightPercent(p)`**，并配 `Shrink(0)`
让元素不要在空间不够时先缩水。`SplitPane` 就是这么写的。

### 15.2 空盒子量不出宽度

无头测试里 `tt.Find(label)` 返回的是**元素的盒子**，一个没有内容的 `ui.Box`
宽度是 0 —— 哪怕它占了 200 DIP。

所以想断言「左边占了多少」，要么给盒子内容并断言**内容的位置**，
要么像 `SplitPane` 那样加一个可访问标签后再量。**断言位置比断言宽度可靠。**

### 15.3 `Vertical` 这种名字有两种读法

「垂直分隔条」= 左右并排，「竖着放」= 上下堆叠。写反了不会报错，
只会安静地把侧栏放到页面上面。

所以 `SplitPaneOptions` 用 **`SideBySide`**，按它做什么来命名，不按分隔条的方向。

---

## 16. 实现进度

`docs/component-catalogue.md` 是 532 个组件的完整清单。当前：

| 包 | 状态 |
|---|---|
| `ui/theme` | 完整（配色 + 密度 + 字号 + 圆角 + 框架尺寸，对比度有测试守着） |
| `ui/core` | 完整（`Use` / `Tokens` / `Density` / `Severity` / `Msg` / `Local` / `Motion`） |
| `ui/layout` | 完整（容器、分割线、叠放、网格、滚动、分栏、标题栏、状态栏、App 外壳） |
| `ui/display` `ui/input` `ui/navigation` `ui/data` `ui/feedback` | 看板所需的核心件已完成，余下按清单补 |
| 其余 14 个包 | 待实现 |

每个组件都遵守同一套约定：每帧一次 `core.Use`、参数走 `XxxOptions`、
交互走 `XxxResult` 的单一方法、组件不持有状态、给不了的东西 `panic("pkg: …")`。

---

## 17. MyGo 的另外三个坑（2026-10 补）

前三个见 §15。这三个是写 `ui/input`、`ui/data`、`ui/navigation` 时撞出来的，
都比前三个更难查——因为它们都不报错。

### 17.1 元素归属于「它被创建时当前的那个容器」

```go
img := ui.Image(c, src)      // 归属于窗口
box.Children(func() {        // 此刻的 parent 仍然是 box
    ui.Text(c, "在框里")
})
```

`img` 落在 `box` **外面**，`box` 于是量出来是 `{0 0 0 0}`。
没有报错，只是一个组件安静地什么都没画。

所以：**要先进入容器，再在里面创建元素**。反过来（先建元素，后挂容器）
在 `ui.Row` / `ui.Box` 的链式写法里不会发生，但只要把创建提到闭包外面就会。

### 17.2 一帧最多构建 3 次，settle 之后的 `Clicked()` 是 false

MyGo 为了让帧内的事件结果在帧尾就画出来，会在一帧里把视图跑最多 3 次
（`runtime.go:333`）。无头测试 settle 完之后，最后一趟的
`Clicked()` / `Changed()` / `Submitted()` **全是 false**。

后果是这一类测试必然失败：

```go
tt.Click("Save")
if r.Clicked() { … }        // ❌ 永远不成立
```

正确写法是**在视图闭包里跨趟累积**：

```go
var pressed int
tt := ui.NewTester(func(c *ui.Context) {
    if Button(c, "Save", ButtonOptions{}).Clicked() { pressed++ }
}, 400, 200)
tt.Click("Save")
if pressed != 1 { … }       // ✅
```

**推论：断言捕获的变量，不要断言点击后画出来的文字。**
文字要到下一帧才稳，而帧数是有限的。

### 17.3 `Disabled` 不阻止子控件被点

`Disabled(true)` 只把自己画灰，`pointerDown` 看的仍是元素自己的 flag，
所以被禁用的框里的按钮照样能聚焦、能点。

每个带子控件的容器要在**框和里面的控件上各标一次**。
更重要的是：**不要把 `Disabled` 当作逻辑守卫**。分页的箭头就是例子——
它被禁用时仍能收到点击，所以箭头的代码里必须自己判范围
（`ui/navigation/tabs.go` 的 `arrow`）。

### 17.4 附带记两条

- **`internal.Commas` 只处理 `int`**，金额千分位要另写（整数位和小数位要分开处理）。
  以后 `ui/chart` 也要分组格式化的话，把它提到 `internal`。
- **同一个 `ui.ListState` 不能在一帧里给两个列表** —— MyGo 会 panic。
  每个列表组件各自持有一个。

---

## 18. 分层：为什么 `Panel` 在 layout 而不是 overlay

`overlay` 里的对话框要画按钮，用的是 `input.Button`，所以
**`overlay` 依赖 `input`**。如果 `Panel` 放在 `overlay`，
那么 `input` 想给自己的下拉菜单画一个面板就得反过来 import `overlay` —— 成环。

解法是把 `Panel` 放到 `ui/layout`：它谁都不依赖（只用 `core` 和 `theme`），
所以 `input` 和 `overlay` 都能用。`overlay.Panel` 现在只是
`layout.Panel` 的一层转发，保留是因为所有浮层的调用方都从那里找它。

**这条规则的形状**：`theme` ← `core` ← `layout` ← `display`/`input` ← `overlay`/`data`。
箭头只能向右。发现反向依赖时，把被共享的东西往下沉，而不是在上层互相引用。

---

## 19. 画廊抓到的五个 bug（2026-10 第二轮）

这一轮把每个包画进真实窗口看图，又抓到五个测试全绿时看不见的问题。
它们的共同点：**尺寸类的东西坏了不会 panic，只会画错或者不画。**

### 19.1 竖线 Divider 一个像素都没画出来

```go
// 错的
return e.Width(theme.BorderWidth).Grow(0)
// 对的
return e.Width(theme.BorderWidth).FillHeight()
```

一个 Row 默认 `AlignItems(Center)`。被居中的子元素，**交叉轴尺寸取内容高度**，
而 Divider 没有内容 —— 高度 0。于是它「画了」一个 1px 宽、0px 高的矩形：
树里有、测试量得到宽度、屏幕上什么都没有。

**规则**：交叉轴上要「撑满」的线，必须显式 `FillHeight()` / `FillWidth()`。
`Grow(0)` 只管主轴，碰不到交叉轴。

`ui/layout/divider_extra_test.go` 读真实像素来守它 ——
量盒子测不出来，因为盒子宽度是对的。

### 19.2 相邻的两个右对齐列会贴成一个

`DataTable` 的 cell 铺满整列宽，两列都 `Align: End` 时中间一个单位的空隙都没有，
`今天$184.50` 会连成一串。

**规则**：固定列宽的表格，**除最后一列外每个 cell 都要让出右侧一个 unit 作为列间距**。
表头和表体都要，否则表头的排序箭头会和相邻列的名字撞上。

### 19.3 表格的 `Width` 必须 ≥ 各列宽度之和

列宽加起来 116，表格给了 82 —— 列被压扁，「Andre Thomson」换行，
行高翻倍，然后被固定的 `Height` 从底部裁掉半行。

**规则**：`Width` 是各列之和，不是「大约够宽」。

### 19.4 `SignaturePoint` 是像素，不是 theme 单位

`PointerPosition()` 给的是像素，所以签名点也是像素。签名板 `unit(42)` 宽实际是 167px，
按 42 写点只会画出 17% 宽的一道痕。

**规则**：任何「从指针拿到的坐标」都是像素；任何「布局尺寸」才是单位。
两者混用时，组件在 1x 下看着有内容，在 2x 下就散了。

**这也是一个 API 上的取舍**：签名按像素存，好处是昨天存今天重放不会悄悄被缩放，
代价是签名不能在不同尺寸的签名板之间通用。要通用就得存归一化坐标，
而那样重放时就丢了笔压。两个都要，就存像素 + 板尺寸。

### 19.5 挂在元素上的浮层，默认不该有遮罩

`Popover` / `HoverCard` / `Popconfirm` 原本默认 modal，于是三张卡片把整页压灰。

**规则，按「这个浮层是不是把别的事挡在外面」分开**：

| 浮层 | 默认 | 为什么 |
|---|---|---|
| `Dialog` / `AlertDialog` / `Drawer` | 模态 | 它们是「先把这件事办完」 |
| `Popover` / `HoverCard` / `Popconfirm` | **非模态** | 它们挂在一个人正在用的页面上。遮罩在说「你停下了」，而实际发生的是「有人在解释其中一行」 |
| `CommandPalette` | 模态 | 它是整窗的搜索，Esc 关闭，遮罩是对的 |

`Popconfirm` 有个例外要小心：非模态时 **Esc 不再关闭它**（Esc 只送给模态层），
所以命令面板不能顺手也改成非模态。

`CommandPaletteOptions.NonModal` 是给画廊页的：它要在一帧里同时显示面板和它盖着的侧栏。

---

## 20. 画廊自己的护栏

画廊的 `-check` 一开始只验证「承诺的文字画出来了」。这抓不到尺寸类的问题 ——
一个空盒子、一个被推到页面边上的按钮、一条 0 像素高的线，文字一个不少。
所以它现在有四条护栏：

| 护栏 | 抓什么 |
|---|---|
| `Page.Missing` | 承诺的文字没画出来 |
| `Page.Gutter`（28px） | 任何 `FillWidth` 的小节都不得贴到页面边上 —— 贴边和被裁在图上长得一样 |
| `BottomFlush` | 页面最后 60 像素里有内容 = 这一页比自己声明的高，底部被切了 |
| `Page.Anchored` | 贴窗口边的层（底栏、抽屉）让上面那条测不了，只能由作者自己声明 |

### 20.1 `Height` 由 `-fit` 给，不要手写

```bash
go run ./cmd/gallery -fit      # 每页需要多高
go run ./cmd/gallery -check    # 校验；报 "the bottom is cut" 就是 Height 写小了
```

`-fit` 从页面**自己声明的高度**往上找，不是从 240 二分。
原因：抽屉挂在窗口右边、选择栏挂在窗口底部，这两样东西落在哪完全取决于页面被摆多高。
先量再裁会把它们留在半空中，唯一的正确做法就是把页面摆到刚好够高。

### 20.2 `Anchored` 是「我知道，而且我看过了」

标了 `Anchored` 的页面跳过底部裁切检查。**这不是免检** ——
标了却还是被切的页面，是有 bug 的。

### 20.3 这套护栏抓到过什么

这一轮的四个 bug（竖线 0 像素、表格列贴在一起、表格宽度小于列宽之和、
页面整体贴边）全部是 `-check` 全绿、870 个测试全绿的情况下，靠看图发现的。
