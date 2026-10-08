# MintUI 对齐 MujicaUI：设计规格

日期：2026-10-08
状态：已批准（待复核）

## 1. 目标与成功标准

把 MintUI 的**组件清单与能力**对齐
[`ZacharyZhang-NY/MujicaUI`](https://github.com/ZacharyZhang-NY/MujicaUI)（同生态的另一个
MyGo 原生组件库），同时**保持 MintUI 自己的实现规范**（`docs/design-system.md`）。

成功标准：

1. MujicaUI 487 项目录中的每一个组件，在 MintUI 里都有对应实现，且拼写一致。
2. Form 与 Data 两个包的**能力**与 MujicaUI 持平（校验、异步校验、dirty 追踪、
   Reset、提交生命周期、列管理、过滤态、行类型泛型）。
3. MintUI 的两条立身契约不被破坏：组件不持有状态；`XxxResult` 只暴露一个方法。
4. 每一批改动都有可自动验证的门禁，且门禁本身可信。

非目标见 §9。

## 2. 现状实测

两库同框架（`github.com/egoist/mygo`）、同 Go 版本（1.27.1）。实测数据：

| 项 | MintUI | MujicaUI |
|---|---|---|
| 包布局 | `ui/<domain>/` 嵌套，22 个库包 | 顶层平铺 |
| 导出函数（去重） | **791** | 目录记录 **487** 项组件 |
| 框架版本 | `mygo v0.2.16` | `mygo v0.2.10` |
| 包文档 | §14.3 只索引 10 包 / 230 项 | `docs/components/` 487 篇，每包一篇 |
| 视觉回归 | `ui.Tester` 无头文本断言 | 1020 张黄金 PNG + 子进程隔离 |
| i18n | `core.Msg(key, def)` 单张 map | 30+ `messages_*.go` 目录 |

结论：框架版本**不是**差异来源（MintUI 反而更新）。差异集中在组件清单、
能力深度、实现规范三处。

`docs/upstream.md` 记录了 MujicaUI 撞到的 14 处 MyGo 上游限制及绕行方案
（`ui.Stepper` 禁用仍可步进、`PopoverBase` 背板吞外点、`SelectBase` 无高亮接口、
`TextInputBase` 无只读、路径遮罩按 1/4 像素缓存导致快照须跑独立子进程…）。
两库撞的是同一堵墙，本设计不重复解决这些问题，只在 §7 记录依赖。

## 3. 对齐契约

以下规则贯穿全部批次，新代码与改动后的代码都必须满足：

| 维度 | 规则 |
|---|---|
| 组件清单 | 以 MujicaUI 487 项目录为基准逐项落齐；已有的**一律改名对齐**，不留同义异名 |
| 能力 | 按 MujicaUI 的能力清单补齐，不做子集 |
| API 形态 | 参数走 `XxxOptions`；返回 `XxxResult`（有交互）或 `*ui.Element`（无交互） |
| Result 方法数 | **恰好一个**，对应用户能做的**那一个**动作 |
| 状态 | 一律调用方指针。视图是纯函数，每帧从状态重建 |
| 命名 | 包路径用 MintUI 惯例 `ui/<domain>/`；组件名用 MujicaUI 拼写（如 `QRCode` 非 `QrCode`） |
| 视觉 | 颜色/字号/圆角全部取 `ui/theme`；间距走 `core.Density(c).Unit()` |
| 每帧首调用 | `core.Use(c, core.Settings{})` |
| 无障碍 | 纯图标元素必须有 `Label`；同一帧内不得有重复 `Label` |
| 焦点 | 兄弟顺序会变处必须给 `Key` |
| 失效 | 给不了的东西 `panic("pkg: …")`，不猜 |

### 3.1 冲突裁决（本项目与 MujicaUI 正面冲突处）

MujicaUI 有两处做法与 MintUI 契约直接冲突。经批准，**能力优先、形式不让**：

1. **不引入内部缓存。** MujicaUI 的 `DataTableState[K]` 持有未导出的 `cache`/`rows`/`sel`。
   MintUI 的对应设计把这些全部外置为调用方指针，排序后的行窗口每帧现算。
2. **拆分多方法 Result。** MujicaUI 的 `FormParts` 有 2 个方法、`DataTableView` 有 4 个。
   MintUI 的对应设计把它们收敛为单个语义命名的方法
   （例：表格的 4 个信号收进 `TableResult.Interaction() TableInteraction`）。

代价：与 MujicaUI 的调用代码不能逐字互换。这是有意接受的。

## 4. 缺口清单

### 4.1 改名对齐（16 项，不新增功能）

| MujicaUI 名 | MintUI 现名 | 位置 |
|---|---|---|
| `chart-annotation` | `chart.Annotation` | `ui/chart/annotation.go` |
| `chart-axis` | `chart.Axis` | `ui/chart/axis.go` |
| `chart-brush` | `chart.Brush` | `ui/chart/brush.go` |
| `chart-crosshair` | `chart.Crosshair` | `ui/chart/crosshair.go` |
| `chart-empty-state` | `chart.Empty` | `ui/chart/empty.go` |
| `chart-grid` | `chart.Grid` | `ui/chart/grid.go` |
| `chart-legend` | `chart.Legend` | `ui/chart/legend.go` |
| `chart-tooltip` | `chart.Tooltip` | `ui/chart/tooltip.go` |
| `qr-code` | `input.QRCode` | `ui/input/qrcode.go` |
| `segmented-control` | `input.Segmented` | `ui/input/choice.go` |
| `spinner` | `feedback.BarsLoader`/`DotsLoader`/`OrbitLoader`/`PulseLoader`/`WaveLoader` | `ui/feedback/loaders.go` |
| `stagger` | `feedback.NewStagger` | `ui/feedback/transition.go` |
| `statistic` | `data.StatLine` | `ui/data/card.go` |
| `year-view` | `datetime.CalendarYearView` | `ui/datetime/views.go` |
| `pnl-display` | `finance.PnLDisplay` | `ui/finance/` |
| `sunburst` | `chart.Sunburst` | `ui/chart/hierarchy.go` |

改名前须确认调用点（含 `ui/showcase/pages/` 与各 `example_*_test.go`），
并同步 `docs/design-system.md §14.3`。

### 4.2 真缺（35 项，需新增）

**交易/金融（15）**

`CompareSymbol`、`CurrencyConverter`、`DepthChart`、`HeikinAshiChart`、
`LineQuoteChart`、`MarketHeatmap`、`NewsFeed`、`OhlcChart`、
`OrderConfirmDialog`、`PriceAlertLine`、`Screener`、`TradeHistoryTable`、
`TrendIndicator`、`VolumeChart`、`VolumeProfile`

**图表进阶（12）**

`ChartExport`、`ChartSync`、`ChartTypeSwitcher`、`DensityPlot`、`DrawingToolbar`、
`DrawingTools`、`IndicatorOverlay`、`IndicatorPane`、`IndicatorSelector`、
`MultiChartLayout`、`RealtimeChart`、`TimeRangeSelector`

**零散（8）**

`ApiRequestBuilder`、`IndexCard`、`IntervalSelector`、`McpServerList`、
`OAuthButtons`、`Playlist`、`PriorityIndicator`、`ReorderTransition`

> `IntervalSelector` 曾被误判为改名项。MintUI 只有 `devtools.RefreshIntervalSelector`
> （刷新间隔，devtools 专用），与图表时间粒度选择器不是同一件事，故计入真缺。

## 5. 分批计划

严格按序执行。批次 0 未完成不得进入批次 1。

| 批次 | 内容 | 依赖 |
|---|---|---|
| 0 | 门禁修复 + 上轮审核的 17 条 Critical | 无 |
| 1 | Form 引擎 + `FormItem[T]` + `FormField` 自动无障碍 | 批次 0 |
| 2 | `DataTable` 泛型化 + 列管理 + `TreeTable` gutter | 批次 0 |
| 3 | 16 项改名对齐 | 可与 1/2 并行 |
| 4 | 35 项真缺，按 trading → chart → 零散 | 批次 2 |
| 5 | 文档与工具收尾 | 全部 |

### 5.0 关于分解

本规格横跨 6 个批次，**不足以支撑单一实施计划**。实施计划按批次分别编写，
首个计划只覆盖**批次 0**。批次 1、2 各自在开工前另起一轮规格复核，
因为它们是 breaking change，且在批次 0 的门禁修好之前无法验证。
批次 4 的 35 项在批次 1、2 完成后需重新评估是否仍要全做。

审核中评为 Important 但不属于 17 条 Critical 的项（`input/knob.go` 与
`display/avatargroup.go`、`display/text.go` 的同帧重复 `Label`、
`ui/feedback/transition.go` 的 `Stagger` 无界增长、
`ui/chat`/`ui/code`/`ui/devtools`/`ui/project`/`ui/files` 的其余 Important 项）
不在批次 0 范围内，随对应领域批次一并处理。

### 5.1 批次 0：地基

**门禁（3 项）**

1. `cmd/gallery/main.go` `shotsTo`：把 `p.Missing(p.Draw())` 移到 `render()` 之前。
   现状 `-shots` 在干净树上就 exit 1（overlay 页 HoverCard 自关闭导致第二次 draw 缺文本）。
2. `ui/showcase/pages/gate_test.go` + `cmd/gallery/main.go`：实现 §20 文档承诺但代码里
   不存在的 `Gutter` 护栏（`Page.Gutter` 28px）；`-check` 增加 `len(Want) >= 3` 下限，
   堵掉空 `Want` 空过；`Missing` 只认真正画出的文字，不认 `Label()`。
3. `ui/showcase/`：补 `c.Invalidate()`（现为 0 处），并把包级 `var` 演示状态迁到
   `showcase.State`。参考消费者不能教错写法。

**Critical（17 项，已逐条复核）**

| 位置 | 问题 |
|---|---|
| `ui/core/core.go:95-117` | `Use` 构造 `&ui.Theme{}` 时漏 `FontSize`/`Font`，系统文字缩放被丢弃 |
| `ui/theme/theme.go:67-69` | `IsDark()` 是指针相等，任何非 `#ffffff` 的浅色板都被判成深色 |
| `ui/core/state.go:85-107` | 全库 `c.Preferences()` 调用数为 0，`Reduced`/`FontSize` 忽略桌面无障碍偏好 |
| `ui/agent/diff.go:632` | 审批写入 `Options` 值拷贝，签署结果帧尾蒸发 |
| `ui/finance/order.go:143` | `kindIndex(&opts.Kind)` 指向副本，Market/Limit/Stop 永远选不了 |
| `ui/code/tools.go:279,319` | 查找/替换框绑定局部 `string`，调用方 query 永不改变 |
| `ui/devtools/response.go:320` | 日志级别过滤写入值副本，过滤失效 |
| `ui/overlay/layer.go:38,49` | `.Modal()` 无条件，`NonModal` 选项无效 |
| `ui/overlay/dialog.go:161` | Esc 投递被 `layer()` 提前消费，`Chosen()` 拿不到 cancel 索引 |
| `ui/navigation/palette.go:195` | `NonModal` 承诺「Esc 总会关」但不生效 |
| `ui/datetime/timers.go:262` | Reset 无条件 `*opts.Laps = nil`，`Laps` 可选时 nil 解引用 |
| `ui/datetime/format.go:266` | `DurationText` 自递归，`math.MinInt64` 栈溢出杀进程 |
| `ui/navigation/tabs.go:343` | `windowAround` 在 `cur > total` 时算出负容量 |
| `ui/input/number.go:428` | `ParseFloat` 接受 `NaN`，`clampTo` 修不了，NaN 永久写入 |
| `ui/input/masonry.go:126` | 默认路径返回 `nil`，`Masonry(...).Grow(1)` nil panic |
| `ui/input/shared.go:260` | `optionRow` 不调 `.Label()`，全库下拉选项行无名 |
| `ui/input/shared.go`/全库 | 选项行无 `Key`，筛选收窄时焦点跳到别的项 |

### 5.2 批次 1：Form

外壳保留 `FormOptions`，内部换成校验引擎。

```go
// 状态由调用方持有；只有注册表是状态，校验结果每帧现算，不缓存视图。
type FormState struct {
    Fields map[string]*FormFieldState
    Order  []string
}

type FormItemOptions[T any] struct {
    Label, Description string
    Required           bool
    Initial            T
    Validate           func(v T) string
    Async              func(v T, version uint64)
}

func Form(c *ui.Context, st *FormState, opts FormOptions, fields func(*FormScope)) FormResult
func FormItem[T any](c *ui.Context, f *FormScope, name string, value *T,
                    opts FormItemOptions[T], control func() *ui.Element) *ui.Element
func (r FormResult) Submitted() bool
func FormField(c *ui.Context, label string, opts FormFieldOptions,
               control func() *ui.Element) *ui.Element
```

要点：

- `FormField` 的 `label` 提为位置参数；`control` 不再收 `err string`，
  改由 `FormField` 自动 `ctrl.Label(label).Description(...).Error(...)` 串联。
  这修掉 MintUI 现状「调用方忘挂 `Label` 就静默失去无障碍」。
- `FormFieldOptions.Horizontal`：标签在左（macOS 系统设置样式）。MintUI 现状无此项。
- dirty 追踪用值快照 diff；异步校验用单调版本号，`Resolve(name, version, err) bool`
  丢弃过期结果。
- 提交生命周期：`Submitting()` / `EndSubmit(formError string)` /
  `Awaiting`。settle 后定位首个错误字段、`c.Announce`、`c.Invalidate()`。
- Reset：恢复 `Initial` 并递增 `generation` 换 `Key`，丢弃控件草稿与错误。
- `EndSubmit` 在无提交进行中时 `panic("input: …")`。

按 §3.1，`Submitted()` 是 `FormResult` 上唯一的方法；Reset 由调用方持有的
`*bool` 或独立入口承接，不额外开第二个方法。

### 5.3 批次 2：Data

```go
type DataColumn[R any] struct {
    ID, Title       string
    Width, MinWidth float32
    Align           ui.Align
    Sortable        bool
    Compare         func(a, b R) int
    Text            func(r R) string
    Cell            func(r R)
}

type DataTableOptions[R any, K comparable] struct {
    Columns []DataColumn[R]
    Rows    []R
    Key     func(r R) K
    Multiple, Controlled bool
    Version int
    Empty   string
    Status  DataStatus
    Error   string
}

func DataTable[R any, K comparable](c *ui.Context, opts DataTableOptions[R, K]) TableResult
```

要点：

- 列自带 `Compare`/`Text`/`Cell`，取代 MintUI 的 `data.Rows`/`Toggle`/`ByText`/`ByNumber`
  辅助函数（这四个随之作废）。
- 排序、列序、列宽、显隐、选中**全部外置**为调用方指针
  （`*ui.SortOrder`、`*ColumnLayout`、选区指针）。不引入 `DataTableState` 内部缓存。
- 行窗口每帧由排序后的切片现算，不缓存。
- 列管理菜单走 `ui.PopoverBase`：排序、左右移、显隐、调宽。最后一个可见列不可隐藏。
- 虚拟化每列表必须独占一个 `ui.ListState`（同帧共用会 panic）。
- 内建 empty / loading / error 三态。
- `TreeTable` 一并修 §19.2 的表体 gutter 缺失（`ui/data/treetable.go:141`），
  否则新增列管理会放大该缺陷。
- 表格 4 个交互信号按 §3.1 收进单个 `TableResult.Interaction() TableInteraction`。

### 5.4 批次 5：文档与工具

- `docs/design-system.md §14.3` 重写为**全 22 包索引**。现状只索引 10 包 / 230 项，
  实际 791 项，漏掉 12 个领域包。
- `cmd/coverage/main.py`：当前只遍历 `sorted(base)`，新增包被静默排除；
  正则 `func ([A-Z]\w*)\(` 漏掉全部泛型（`data.Rows`/`Tree`/`TreeTable`、`layout.Column`）；
  `norm()` 别名规则把 Options/Result **类型名**计入「改名」额度，虚高覆盖率；
  且永远 exit 0，不设门。三项一并修正，并统一 `docs/component-catalogue.md`、
  `docs/coverage-report.txt`、README 的分母（现状 618 与 532 并存）。
- 把 22.7 MB 的 `gallery` 二进制移出 git 索引并加入 `.gitignore`。
- 清理 `design-system.md` 里指向已删除路径的引用：`internal/board/`（§1、§13、§14.1）、
  `web/assets/app.css`（§1）、`go run ./cmd/snapshots`（§13、§14.4）；
  §14.4 重复的清单；§20.3 的「870 个测试」需按实际更新。

## 6. 验证策略

不引入快照体系（已批准）。沿用 MintUI 现有门禁：

| 层 | 手段 |
|---|---|
| 纯函数 | 精确值断言，不用「大于 0」。边界必测：空输入、单元素、全等值、负值、极值 |
| 组件 | `ui.NewTester` 无头渲染，断言**可见文字与布局结果**，不断言内部字段 |
| 交互 | 按 §17.2 在视图闭包里跨趟累积计数，**不断言点击后画出的文字** |
| 拖拽 | 必须分步移动（Press → Move → Move → Release），一步跳过不算 |
| 无障碍 | 每个纯图标元素断言其 `Label` 存在；同帧 `Label` 唯一性 |
| 包门禁 | `ui/showcase/pages/gate_test.go`：新增 `ui/*` 包而无画廊页即失败 |
| 目视 | `-shots` 出图后**人眼逐页读一遍**（§20 既有要求，尺寸类缺陷不 panic） |

`-fit` 给出的高度不得手写。`Anchored` 仅用于确知并已目视过的贴边层。

## 7. 已知上游依赖

以下来自 MujicaUI `docs/upstream.md`，本设计不重复解决，但会在实现时碰到：

| MyGo 限制 | 影响本项目的组件 |
|---|---|
| 自定义元素无法设 `expanded`/`mixed`/数值范围与值 | `input.Checkbox` 半选、`MultiSelect`、`TreeSelect`、`RangeSlider`、`Rating`、各日期时间弹出按钮 |
| 无 `menu`/`menuitem` 角色 | `navigation.Menubar`、`overlay.DropdownMenu`、`ContextMenu` |
| 输入法组合状态不公开 | `input.TextInput`、`TextArea`、`SearchInput`、`overlay.Dialog`、`AlertDialog` |
| 浮层焦点恢复与浮层快捷键不公开 | `overlay.Dialog`、`Drawer` |
| `PopoverBase` 背板吞外部点击；只有上下两个位置 | `overlay.Popover`、`Popconfirm`、`DropdownMenu`、`HoverCard` |
| `SelectBase` 无高亮设置接口 | `input.Select` |
| `TextInputBase` 无只读开关 | `input.TextInput` 只读模式 |
| `ui.Stepper` 禁用仍可步进、固定 9 位取整 | `input.NumberInput` |
| `SliderBase` 方向键步长固定 1% | `input.Slider` |

`ui/overlay/layer.go` 的 `NonModal` 修复（批次 0）与最后一行
「`PopoverBase` 背板吞外部点击」是同一堵墙的不同侧面，实现时须一致处理。

## 8. 风险

| 风险 | 缓解 |
|---|---|
| 批次 1/2 是 breaking change，画廊页与测试需同步改 | 每批结束跑 `-check` + `-shots` + `go test ./...`；分批提交，每批可独立回滚 |
| 34 项真缺横跨 5 个领域，工作量远大于 Form/Data | 严格按批次推进，批次 1、2 完成后重新评估批次 4 是否仍要全做 |
| `design-system.md §20` 已承诺不存在的 `Gutter` 护栏，说明文档与实现存在系统性漂移 | 批次 0 补齐护栏；批次 5 重写 §14.3；此后每批改动须同步对应文档节 |
| 升级到 `mygo v0.3.x` 会把 `Element` 的 builder 方法搬到未导出的 `*node` 上，影响全部 791 个导出函数 | **本设计全程留在 `v0.2.16`**，不升级。需要的新原语已确认在 v0.2.16 全部存在 |

## 9. 不做的事

- 不升级 `mygo` 到 v0.3.x（见 §8）。
- 不引入黄金快照 / 像素级回归体系（已批准不做）。
- 不改包布局（保持 `ui/<domain>/` 嵌套）。
- 不引入 MujicaUI 的 `messages_*.go` 分包 i18n 目录；沿用 `core.Msg` + 单张 map，
  仅在新增组件需要新文案时扩充。
- 不追求与 MujicaUI 调用代码逐字互换（见 §3.1）。
- 不做视觉风格对齐（Court-gothic 主题不引入）；沿用 MintUI 现有
  「面板在下、卡片在上、黑白灰」设计语言。

## 10. 验收标准

批次全部完成后，以下每一条都必须成立：

1. `go build ./...`、`go vet ./...`、`go test ./... -count=1` 全绿。
2. `go run ./cmd/gallery -check` 在干净树上 exit 0。
3. `go run ./cmd/gallery -shots <dir>` 在干净树上 exit 0（**现状不成立**）。
4. `Page.Gutter` 护栏实际存在且所有页面通过。
5. `-check` 无法因空 `Want` 而通过。
6. MujicaUI 487 项目录逐项可在 MintUI 找到对应实现，无同义异名。
7. 无任何组件在内部持有可变状态；无任何 `XxxResult` 暴露多于一个方法。
8. `docs/design-system.md §14.3` 索引全部 22 个库包，与代码零漂移。
9. 覆盖率工具能发现新增包与泛型函数，且分母在三个文件中一致。
10. git 索引中不含构建产物。