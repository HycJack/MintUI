# 批次 0：门禁修复与 17 条 Critical 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让画廊门禁本身可信，并修掉审核中已逐条复核的 17 条 Critical，使后续 Form/Data 对齐批次有可自动验证的地基。

**Architecture:** 分两段。Tasks 1-5 只碰门禁与画廊演示层，让 `-shots` 在干净树上通过、并加两道目前不存在的检查；Tasks 6-14 修库本身，分三层——基础层（`core`/`theme`）、浮层契约、组件级崩溃与状态归属。每任务自带 TDD 循环与独立提交。

**Tech Stack:** Go 1.27.1、`github.com/egoist/mygo v0.2.16`（**全程不升级**）、`go test` + `ui.NewTester` 无头渲染。

**Spec:** `docs/superpowers/specs/2026-10-08-mujica-parity-design.md`（§3 对齐契约、§5.1 批次 0、§6 验证策略）

## Global Constraints

以下来自 spec §3，**每个任务都隐含包含**：

- 组件不持有状态；视图是纯函数，每帧从状态重建。
- `XxxResult` **恰好暴露一个**方法，对应用户能做的**那一个**动作。
- 参数走 `XxxOptions`；返回 `XxxResult`（有交互）或 `*ui.Element`（无交互）。
- 每帧首个调用是 `core.Use(c, core.Settings{})`。
- 颜色/字号/圆角全部取 `ui/theme`；间距走 `core.Density(c).Unit()`。
- 纯图标元素必须有 `Label`；同一帧内 `Label` 不得重复。
- 兄弟顺序会变处必须给 `Key`。
- 给不了的东西 `panic("pkg: …")`，不猜。包前缀用该包自己的名字。
- 保持 `github.com/egoist/mygo v0.2.16`，**不得升级**。所有需要的新原语已确认在 v0.2.16 存在。

## Review Focus

spec 暗示但没有任何任务的测试覆盖、最可能咬到人的五类输入。每行都已在下方对应任务里配了钉住它的测试。

1. **一个 `Background` 不是 `#ffffff` 的自定义浅色调色板**（例如品牌米白 `#fafafa`）——必须仍被判为浅色。现状 `IsDark()` 是指针相等，会把它判成深色并推给 MyGo 的 `Theme.Dark`，整个窗口用深色字形。→ Task 7
2. **`Options` 是结构体值、而交互处理器写它的字段**——写入丢失且不报错。现状已有 4 处（agent/finance/code/devtools），每加一个组件都会再添一处。→ Task 13、14
3. **调用方传了可选指针的零值（nil）再触发对应动作**——必须不崩。现状 `Stopwatch` 的 Reset 无条件写 `*opts.Laps`。→ Task 10
4. **数值输入的极值与非法值**：`NaN`、`math.MinInt64`、负时长、超大偏移量。现状 `NaN` 能永久写进 `*value`，`MinInt64` 让 `DurationText` 栈溢出**杀死进程**（`recover()` 无效）。→ Task 10、12
5. **被禁用或越界的控件收到点击**——必须不生效（§17.3：`Disabled` 只画灰，不拦点击）。→ Task 9、11

---

## 文件结构

**门禁与画廊（Tasks 1-5）**

| 文件 | 动作 | 职责 |
|---|---|---|
| `cmd/gallery/main.go` | 改 | `-check` 加 `Want` 下限；`shotsTo` 改为先校验后渲染；新增 `GutterFlush()` |
| `ui/showcase/page.go` | 改 | `Missing` 只认画出的文字，不认 `Label()` |
| `ui/showcase/pages/gate_test.go` | 改 | 新增 `Gutter` 与 `Missing` 两个测试；`TestPagesDrawInBothAppearances` 真渲染深色 |
| `ui/showcase/pages/messaging.go` | 改 | 常开浮层状态改为 `showcase.State` 且不被组件自关 |
| `ui/showcase/pages/overlay.go` | 改 | 同上（hovercard 是 `-shots` 失败的直接来源） |

**基础层（Tasks 6-8）**

| 文件 | 动作 | 职责 |
|---|---|---|
| `ui/core/core.go` | 改 | `Use` 保留 `FontSize`/`Font` |
| `ui/core/state.go` | 改 | `Reduced`/`FontSize` 读 `c.Preferences()` |
| `ui/theme/theme.go` | 改 | `IsDark()` 改相对亮度比较 |

**浮层契约（Task 9）**

| 文件 | 动作 | 职责 |
|---|---|---|
| `ui/overlay/layer.go` | 改 | `.Modal()` 与 `back.Clicked()` 按 `modal` 门控 |
| `ui/overlay/dialog.go` | 改 | Esc 读取顺序，body 能拿到投递 |

**组件级（Tasks 10-14）**

| 文件 | 动作 | 职责 |
|---|---|---|
| `ui/datetime/timers.go` | 改 | Reset 守卫 `opts.Laps` |
| `ui/datetime/format.go` | 改 | `DurationText` 不递归 |
| `ui/navigation/tabs.go` | 改 | `windowAround` 夹紧 `lo` |
| `ui/navigation/palette.go` | 改 | `NonModal` 的 Esc 承诺 |
| `ui/input/number.go` | 改 | 拒绝非有限值 |
| `ui/input/masonry.go` | 改 | 默认路径返回 `row` 而非 `nil` |
| `ui/input/shared.go` | 改 | `optionRow` 挂 `Label`、加 `Key` |
| `ui/agent/diff.go` | 改 | `Approved` 必须非 nil |
| `ui/finance/order.go` | 改 | `Kind` 改指针 |
| `ui/code/tools.go` | 改 | `Query`/`Replace` 改指针 |
| `ui/devtools/response.go` | 改 | `Level` 改指针 |

---

## Task 1: `-shots` 先校验后渲染，`-check` 加 `Want` 下限

**Files:**
- Modify: `cmd/gallery/main.go:60-83`（`-check` 分支）、`cmd/gallery/main.go:104-138`（`shotsTo`）
- Test: `cmd/gallery/window_test.go`

**Interfaces:**
- Consumes: `gallery.Page.Missing([]string) []string`、`render(Page, bool) *image.RGBA`（均已存在，不改签名）
- Produces: `shotsTo(dir string, pages []gallery.Page, dark bool, scale int)` 行为变更——校验在任何渲染之前完成。不新增导出符号。

- [ ] **Step 1: 写失败测试**

在 `cmd/gallery/window_test.go` 追加。测试要点：一个 `Want` 里含某文本、但该组件在无指针渲染时会自行关闭的 page，`shotsTo` 不得报缺失。用一个最小 stub page 复现「第二次绘制缺文本」：

```go
// flipPage 是一个只在第一次绘制时给出 Want 里那个文本的 page，
// 复现 hovercard 在无指针时自关的行为。
type flipPage struct {
	shown *bool
	Page
}

func TestShotsToChecksBeforeItRenders(t *testing.T) {
	dir := t.TempDir()
	shown := false
	p := flipPage{shown: &shown, Page: gallery.Page{
		Package: "flip", Width: 200, Height: 120,
		Want: []string{"promised"},
		Render: func(c *ui.Context) {
			if !*p.shown {
				*p.shown = true
				ui.Text(c, "promised")
			}
		},
	}}
	// shotsTo 不应因为自己先渲染而报缺失。
	shotsTo(dir, []gallery.Page{p.Page}, false, 1)
}
```

若 `gallery.Page.Render` 字段名与实际不符，读 `ui/showcase/page.go` 确认后按实际字段改写桩。断言方式：`shotsTo` 不 panic 即通过（它内部有 `os.Exit`，所以本测试只覆盖「不因自渲染顺序而误报」这一路径；若难以在同进程测试，改为直接断言新的 `checkPage` 辅助函数）。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/gallery/ -run TestShotsToChecksBeforeItRenders -count=1 -v`
Expected: FAIL——现状 `shotsTo` 先 `render` 再 `Missing`，`promised` 已被消耗。

- [ ] **Step 3: `shotsTo` 改为先校验后渲染**

在 `cmd/gallery/main.go` 的 `shotsTo` 里，把校验循环提到渲染循环**之前**，并从渲染循环中删掉原来的 `p.Missing(p.Draw())` 与 `bad` 计数。结构：

```go
	bad := 0
	for _, p := range pages {
		if miss := p.Missing(p.Draw()); len(miss) > 0 {
			fmt.Printf("  ⚠ did not draw %v\n", miss)
			bad++
		}
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "%d pages did not draw what they promised\n", bad)
		os.Exit(1)
	}
	for _, p := range pages {
		// ...原有的双模式渲染与写 PNG 逻辑不变...
	}
```

保留原有的 `⚠ did not draw` 输出文案与退出码 1，只改执行顺序。

- [ ] **Step 4: `-check` 增加 `Want` 下限**

在 `cmd/gallery/main.go` 的 `-check` 分支里，`miss := p.Missing(p.Draw())` 之后立刻加：

```go
			if len(p.Want) < 3 {
				fmt.Printf("✗ %-12s promises only %d texts — a page that shows almost "+
					"nothing is not a gallery of that package\n", p.Package, len(p.Want))
				bad++
				continue
			}
```

阈值 **3** 与 `ui/showcase/pages/gate_test.go:68` 已有的检查保持一致，不要改成别的数。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd/gallery/ -count=1`
Expected: PASS

- [ ] **Step 6: 验证 `-shots` 现在在干净树上通过**

Run: `go run ./cmd/gallery -shots "$TMPDIR/mintui-t1"`
Expected: 每页一行 `✓`，**无** `⚠ did not draw`，进程 exit 0。

若仍报缺失，把失败页名记下——那是 Task 4 要处理的 `messaging`/`overlay`，**不要**在本任务里改页面。

- [ ] **Step 7: 提交**

```bash
git add cmd/gallery/main.go cmd/gallery/window_test.go
git commit -m "fix(gallery): check promised text before rendering shots

-shots reported 'did not draw [North 分支 · 本周 14 单]' and exited 1 on
a clean tree: shotsTo rendered the page before running the Missing check,
and the overlay page's hover card closes itself when no pointer is over
it, so the second draw legitimately lacked the text.

-check also accepted a page with an empty Want, printing '0 promised
texts' and exiting 0. Now floored at 3, matching the test gate."
```

---

## Task 2: `Missing` 只认画出的文字，不认 `Label()`

**Files:**
- Modify: `ui/showcase/page.go:118-132`（`Missing`）
- Test: `ui/showcase/pages/gate_test.go`

**Interfaces:**
- Consumes: `Page.Draw() []string`（不变）
- Produces: `Missing` 语义收窄——只有**真正画出**的文字才满足 `Want`。不改签名。

- [ ] **Step 1: 写失败测试**

在 `ui/showcase/pages/gate_test.go` 追加：

```go
// 一个不可见但带 Label 的盒子不能算作「画出了那个文本」。
func TestAnAccessibilityLabelIsNotDrawnText(t *testing.T) {
	p := gallery2Page(t, func(c *ui.Context) {
		ui.Box(c).Label("promised").Size(10, 10)
	}, "promised")
	if miss := p.Missing(p.Draw()); len(miss) != 1 {
		t.Fatalf("a Label on an empty box satisfied Want: %v", miss)
	}
}
```

`gallery2Page` 用一个内联 helper 构造 `showcase.Page`，字段以 `ui/showcase/page.go` 实际定义为准（`Package`/`Width`/`Height`/`Want`/`Render`）。若 `Tester.Texts()` 当前已不返回 `Label`，本测试会直接通过——那就说明 spec §5.1 第 2 条的第 3 小项在 v0.2.16 上不成立，**在提交信息里注明并跳过本任务的代码改动**，只保留测试作为回归防护。

- [ ] **Step 2: 跑测试确认失败或确认前提**

Run: `go test ./ui/showcase/pages/ -run TestAnAccessibilityLabelIsNotDrawnText -count=1 -v`
Expected: FAIL（`Texts()` 含 `Label`）或 PASS（不含，则按 Step 1 的说明处理）。

- [ ] **Step 3: 实现**

若 Step 2 为 FAIL：在 `ui/showcase/page.go` 加一个只取绘制文字的辅助函数，并让 `Missing` 用它。

```go
// drawnText is Page.Draw's texts minus accessibility labels: a Want entry is
// a promise that something is visible, and a Label on an empty box is not.
func (p Page) drawnText() []string { ... }
```

实现方式：新建一个 `ui.NewTester`，在其绘制闭包里遍历 `c.Root()` 的元素树，只收集**有绘制内容**（即 `Texts()` 命中但 `Label()` 不命中）的字符串。读 `ui/showcase/page.go` 与 MyGo `v0.2.16/ui/headless.go` 的 `Tester.Texts()` 实现来选定判据；若 MyGo 未暴露足够信息区分二者，则改为在 `Want` 校验时要求文本出现在 `Texts()` **且** 该页面的绘制闭包至少产生一个非空元素——具体判据由实现者依据 `ui.Text` 的绘制结果确定，并在提交信息里写明所用判据。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/showcase/... -count=1`
Expected: PASS，且既有的 `TestPagesDrawWhatTheyPromised` 仍全绿（若某个页面的 `Want` 靠 `Label` 才满足，本步会暴露出来——那是真实缺陷，把它改成可见文本或从 `Want` 移除，并在提交信息里记录）。

- [ ] **Step 5: 提交**

```bash
git add ui/showcase/page.go ui/showcase/pages/gate_test.go
git commit -m "fix(showcase): a Want entry needs drawn text, not an a11y label

Missing compared against Tester.Texts(), which includes element Label()
values, so a promised string could be satisfied by an accessibility label
on an invisible empty box."
```

---

## Task 3: `Page.Gutter` 护栏落地

`docs/design-system.md §20` 把 `Page.Gutter`（28px）列为四道护栏之一，但代码里 `Gutter` 只是 `ui/showcase/page.go:89` 的一个 padding 常量，**没有任何检查**。本任务把它变成真的检查。

**Files:**
- Modify: `cmd/gallery/main.go`（新增 `GutterFlush`，与既有 `BottomFlush` 同文件同风格）
- Test: `ui/showcase/pages/gate_test.go`

**Interfaces:**
- Consumes: `render(Page, bool) *image.RGBA`、`near(color.RGBA, color.RGBA) bool`（`cmd/gallery/main.go:219`，已存在，不改）
- Produces: `func GutterFlush(img *image.RGBA) bool`——报告图像左右两条 28px 带内是否存在非背景像素。包内私有，不导出。

- [ ] **Step 1: 写失败测试**

`GutterFlush` 落在 `package main`（`cmd/gallery`），所以它的测试放 `cmd/gallery/window_test.go`，用一张纯合成的图，不依赖任何画廊页：

```go
func TestGutterFlush(t *testing.T) {
	const W, H = 200, 100
	band := gallery.Gutter
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	paint := func(x0, x1 int) {
		for y := 40; y < 60; y++ {
			for x := x0; x < x1; x++ {
				img.SetRGBA(x, y, color.RGBA{20, 20, 20, 255})
			}
		}
	}
	paint(band+12, W-band-12)
	if GutterFlush(img) {
		t.Error("a section inset past the gutter was reported as flush against the edge")
	}
	paint(4, W-4) // 现在同一块内容贴到 x=4，应当被抓到
	if !GutterFlush(img) {
		t.Error("content at x=4 was not reported as flush against the edge")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/gallery/ -run TestGutterFlush -count=1`
Expected: 编译失败 `undefined: GutterFlush`。

- [ ] **Step 3: 实现 `GutterFlush`**

在 `cmd/gallery/main.go` 里，紧跟 `BottomFlush` 之后新增。照 `BottomFlush` 的写法：取左上角像素为背景色，用既有的 `near()` 判差异，扫描左右两条各 28 像素的带。

```go
// GutterFlush reports whether a page painted inside its own 28-pixel margin.
//
// Page.Paint pads every page by Gutter, so anything non-background in the
// leftmost or rightmost band got there by escaping that padding — an
// absolutely positioned layer, a negative margin. A picture cannot tell a
// deliberate flush edge from a clipped one, which is why this exists.
func GutterFlush(img *image.RGBA) bool {
	b := img.Bounds()
	bg := img.RGBAAt(b.Min.X, b.Min.Y)
	for _, band := range [][2]int{{b.Min.X, b.Min.X + gallery.Gutter}, {b.Max.X - gallery.Gutter, b.Max.X}} {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := band[0]; x < band[1]; x++ {
				if x < b.Min.X || x >= b.Max.X {
					continue
				}
				if !near(img.RGBAAt(x, y), bg) {
					return true
				}
			}
		}
	}
	return false
}
```

用 `gallery.Gutter`（`ui/showcase/page.go:89` 的常量）而不是硬编码 28。若 `cmd/gallery/main.go` 尚未 import 该包，按其现有 import 风格补上。

- [ ] **Step 4: 接到 `-check` 上**

在 `cmd/gallery/main.go` 的 `-check` 分支里，紧接 `BottomFlush` 检查之后加同样形态的一段：非 `Anchored` 且 `GutterFlush(render(p, false))` 为真时，打印

```go
				fmt.Printf("✓ %-12s %d promised texts, but painted inside its own %d-pixel "+
					"gutter — a section is flush against the window\n", p.Package, len(p.Want), gallery.Gutter)
				bad++
				continue
```

- [ ] **Step 5: 跑测试确认通过，并列出违规页**

Run: `go test ./cmd/gallery/ -run TestGutterFlush -count=1` → PASS
Run: `go run ./cmd/gallery -check` → **很可能 exit 1** 并列出若干 `painted inside its own 28-pixel gutter` 的页面。

这是预期的：spec 记录该性质在 15/22 个页面上被违反。**本任务只负责让检查存在并报出真相**，修页面留到 Task 4。把违规页名单原样写进提交信息。

- [ ] **Step 6: 提交**

```bash
git add cmd/gallery/main.go cmd/gallery/window_test.go
git commit -m "feat(gallery): implement the Page.Gutter guardrail

design-system.md §20 lists Page.Gutter (28px) as one of four guardrails,
but Gutter was only ever the padding constant in ui/showcase/page.go:
nothing checked it. Implemented GutterFlush alongside the existing
BottomFlush and wired it into -check.

The check now fails; the page list is recorded in this commit message and
is fixed in the following commit."
```

---

## Task 4: 画廊渲染确定性，并修 `messaging` / `overlay` 两页

**实测基线：** 开 `core.WithReducedMotion(c, true)` 后把同一页连渲两次、逐字节比对 `Image().Pix`，现状恰好两个页失败——`overlay` 19051 px、`messaging` 1608 px；其余 20 页稳定。不开减弱动效时另有 `feedback`/`chat`/`agent` 失败，那三个是 `p.Now()` 驱动的合法动画差异，**不是缺陷**。

**根因：** 两页都把常开浮层的 `open` 放在包级 `var` 里并初始化为 `true`（`overlay.go` 的 `hoverOpen`、`messaging.go` 的 `snoozeOpen`/`cardOpen`/`messaging_open`/`pickerOpen`/`searchOpen`）。`ui/overlay/hovercard.go:78-81` 在 `*open && panel != nil && !panel.Hovered()` 时把 `*open` 置 false——无头渲染没有指针，于是浮层在**第一次绘制时自关**，状态残留到第二次。

**Files:**
- Modify: `ui/showcase/pages/gate_test.go`（新增确定性测试）
- Modify: `ui/showcase/pages/overlay.go`、`ui/showcase/pages/messaging.go`
- Test: 同上

**Interfaces:**
- Consumes: `showcase.State[T](c *ui.Context, key string, init T) *T`（`ui/showcase/state.go`，已存在）；`core.WithReducedMotion(c *ui.Context, on bool) *ui.Context`（`ui/core/state.go`，已存在）；`showcase.Sorted() []Page`
- Produces: 新测试 `TestPagesRenderDeterministically`（在 `gate_test.go`）。不改任何导出符号。

- [ ] **Step 1: 写失败测试**

在 `ui/showcase/pages/gate_test.go` 追加：

```go
// TestPagesRenderDeterministically draws every page twice with animation
// off and requires the two images to be identical.
//
// Animation is excluded on purpose: a spinner legitimately draws
// differently a millisecond later, and a gate that fails on that teaches
// nobody anything. What must not change between two draws of an untouched
// page is everything else — a page that mutates its own state while
// drawing leaves the next draw (or the dark-mode draw) showing something
// different, which is how -shots started failing on a clean tree.
func TestPagesRenderDeterministically(t *testing.T) {
	for _, p := range showcase.Sorted() {
		t.Run(p.Package, func(t *testing.T) {
			draw := func() []uint8 {
				tt := ui.NewTester(func(c *ui.Context) {
					core.Use(c, core.Settings{Mode: core.Light})
					core.WithReducedMotion(c, true)
					p.Paint(c)
				}, p.Width, p.Height)
				return tt.Image().Pix
			}
			first, second := draw(), draw()
			if !bytes.Equal(first, second) {
				t.Errorf("%s drew differently on the second pass (%d of %d bytes differ)",
					p.Package, diffBytes(first, second), len(first))
			}
		})
	}
}

func diffBytes(a, b []uint8) int {
	if len(a) != len(b) {
		return len(a)
	}
	n := 0
	for i := range a {
		if a[i] != b[i] {
			n++
		}
	}
	return n
}
```

需要的 import：`bytes`、`github.com/egoist/mygo/ui`、`github.com/HycJack/MintUI/ui/core`。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/showcase/pages/ -run TestPagesRenderDeterministically -count=1`
Expected: FAIL，`overlay` 与 `messaging` 两个子测试报错。

- [ ] **Step 3: 修 `overlay` 页**

把 `ui/showcase/pages/overlay.go` 里那个常开 hovercard 的 `open` 从包级 `var` 改为 `showcase.State`，并且**不要**让它被组件自关。做法：

1. 删除对应的包级 `var`（`hoverOpen` 或其等价物，以文件实际名为准）。
2. 在该 demo 的绘制闭包里取 `open := showcase.State(c, "overlay/hover", true)`，把它传给 hovercard。
3. 若组件仍会自关（因为 `*open` 为 true 而 `panel.Hovered()` 为 false），则不要靠 `showcase.State` 硬顶——**改为让该 demo 不依赖 hovercard 常开**：把这一段换成 `overlay.Popover`（接受显式 `open *bool`、不会自关），或在 hovercard 的 `PanelOptions` 上用 `Art`/静态内容让两次绘制一致。

以第 3 条为主。理由写在代码注释里：

```go
		// A hover card closes itself when nothing is over it, which is right
		// with a pointer and wrong here: two headless draws would differ, and
		// -shots draws each page more than once. Popover takes an explicit
		// open flag and does not second-guess the caller.
```

- [ ] **Step 4: 修 `messaging` 页**

同样处理 `ui/showcase/pages/messaging.go` 里 `snoozeOpen`/`cardOpen`/`messaging_open`/`pickerOpen`/`searchOpen` 这类**初始化为 `true` 且会被组件自关**的状态：迁到 `showcase.State`，并对会自关的那些改用接受显式 `open` 的组件。逐个跑测试确认，本步结束时应只剩 `messaging` 一个子测试可能仍失败——若仍失败，用下面这步定位。

- [ ] **Step 5: 定位剩余差异**

若 `messaging` 仍失败，在测试里临时打印两次 `Image()` 的差异像素坐标范围（x/y 的 min/max），据此定位是哪一块内容在出现/消失。**不要**为了让测试变绿而放宽断言（例如加容差）；要么按根因修，要么在提交信息里记录该页仍有差异并说明为何可接受。

- [ ] **Step 6: 跑测试确认通过**

Run: `go test ./ui/showcase/... -count=1`
Expected: PASS，全部子测试通过。

- [ ] **Step 7: 确认 `-shots` 与 `-check` 都通过**

Run: `go run ./cmd/gallery -check` → 若 Task 3 的 `Gutter` 护栏仍报违规页，记录名单，**修掉本任务已触及的两页，其余留给后续批次**。
Run: `go run ./cmd/gallery -shots "$TMPDIR/mintui-t4"` → exit 0。

- [ ] **Step 8: 补 `c.Invalidate()`**

`ui/showcase/` 全目录当前 `Invalidate()` 出现 **0 次**，却在 `Clicked()`/`Changed()` 处理器里改全局状态。`cmd/gallery/main.go:287` 的窗口代码做对了（改完调 `c.Invalidate()`）。在本任务已改动的两个页面里，为每个改状态的处理器补上 `c.Invalidate()`。**不要**在本任务里横扫全部 16 个页面——那属于后续批次的整洁化。

- [ ] **Step 9: 提交**

```bash
git add ui/showcase/pages/gate_test.go ui/showcase/pages/overlay.go ui/showcase/pages/messaging.go
git commit -m "fix(showcase): pages must not mutate their own state while drawing

Added TestPagesRenderDeterministically: draw every page twice with reduced
motion on and require byte-identical output. Animation is excluded because
a spinner legitimately draws differently a millisecond later.

It failed for exactly two pages. overlay differed by 19051 pixels and
messaging by 1608, both because a permanently-open float was kept in a
package-level var: HoverCard closes itself when nothing is over it
(hovercard.go:78-81), which is correct with a pointer and wrong in a
headless draw, and the closed state survived into the next draw. That is
also what made -shots fail on a clean tree."
```

---

## Task 5: `TestPagesDrawInBothAppearances` 名实相符

该测试名为「两种外观」，注释承诺「a bug that only shows after dark is the one nobody looks for」，但函数体只调了一次 `p.Draw()`，而 `useForShowcase`（`ui/showcase/furniture.go:12`）恒为浅色。**全项目零深色渲染覆盖。**

**Files:**
- Modify: `ui/showcase/pages/gate_test.go:76-85`
- Test: 同文件

**Interfaces:**
- Consumes: `showcase.Page.Paint(*ui.Context)`、`Page.Width`/`Page.Height`、`core.Dark`
- Produces: 无新导出符号。

- [ ] **Step 1: 改写测试**

把 `TestPagesDrawInBothAppearances` 的函数体替换为真正渲染两次：

```go
func TestPagesDrawInBothAppearances(t *testing.T) {
	for _, p := range showcase.Sorted() {
		for _, mode := range []core.Mode{core.Light, core.Dark} {
			t.Run(p.Package+"/"+modeName(mode), func(t *testing.T) {
				drawn := ui.NewTester(func(c *ui.Context) {
					core.Use(c, core.Settings{Mode: mode})
					p.Paint(c)
				}, p.Width, p.Height).Texts()
				if len(drawn) < 3 {
					t.Errorf("%s in %v drew %d texts: %v", p.Package, mode, len(drawn), drawn)
				}
			})
		}
	}
}
```

`modeName` 用一个两元素查表把 `core.Light`/`core.Dark` 映射成 `"light"`/`"dark"`。

- [ ] **Step 2: 跑测试**

Run: `go test ./ui/showcase/pages/ -run TestPagesDrawInBothAppearances -count=1`
Expected: 大概率 PASS。若某页在深色下画出的文本少于 3 条，那是**真实缺陷**（深色下画空了），按 Task 4 Step 5 的方式定位并在本任务内修掉；修不掉的记录在提交信息里。

- [ ] **Step 3: 提交**

```bash
git add ui/showcase/pages/gate_test.go
git commit -m "test(showcase): actually render the dark appearance

TestPagesDrawInBothAppearances promised to catch 'a bug that only shows
after dark' but called p.Draw() once, and useForShowcase is always light.
There was no dark-mode rendering anywhere in the project."
```

---

## Task 6: `core.Use` 保留 `FontSize` 与 `Font`

`ui/core/core.go:95-117` 手工构造 `&ui.Theme{…}`，**漏掉 `FontSize float32` 与 `Font string`**（两者在 MyGo `v0.2.16/ui/theme.go` 的 `Theme` 结构体里都存在）。后果：`c.Theme().FontSize` 为 0，普通 `ui.Text` 实测 16.49pt，而应为 15.31pt——桌面文字缩放对全部 22 个包失效。

**Files:**
- Modify: `ui/core/core.go:95-117`
- Test: `ui/core/core_test.go`

**Interfaces:**
- Consumes: `github.com/egoist/mygo/ui.Theme`（字段 `FontSize float32`、`Font string`）
- Produces: 无新导出符号；`Use` 的行为变更。

- [ ] **Step 1: 写失败测试**

在 `ui/core/core_test.go` 追加：

```go
func TestUseKeepsTheDesktopFontSize(t *testing.T) {
	var got float32
	ui.NewTester(func(c *ui.Context) {
		Use(c, Settings{})
		got = c.Theme().FontSize
	}, 200, 100)
	if got <= 0 {
		t.Errorf("Use dropped FontSize: c.Theme().FontSize = %v", got)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/core/ -run TestUseKeepsTheDesktopFontSize -count=1 -v`
Expected: FAIL，`FontSize = 0`。

- [ ] **Step 3: 实现**

在 `ui/core/core.go` 的 `&ui.Theme{…}` 字面量里，`Spacing: density.Unit(),` 之后补两行：

```go
		FontSize:       core.FontSize(c, theme.BodySize),
		Font:           *c.Theme().Font,
```

第二行的依据：`Use` 的注释自述「It re-themes MyGo itself」，即它在**改造**既有主题而非从零构造。因此字体相关的两个字段应从传入的 `c.Theme()` 取，只覆盖库自己要管的那部分。`FontSize` 取 `theme.BodySize` 经 `core.FontSize` 缩放，与库内正文一致。

若 `c.Theme().Font` 在某些平台为空字符串，则写 `""` 让 MyGo 走自己的默认值，不要硬编码字体名。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/core/ ./ui/theme/ -count=1`
Expected: PASS。`ui/theme/theme_test.go` 的对比度矩阵不受影响（未动颜色）。

- [ ] **Step 5: 目视复核**

Run: `go run ./cmd/gallery -shots "$TMPDIR/mintui-t6" -scale 2` 后打开 `input.png` 与 `data.png`。

字号会整体变化，这是预期的。确认输入框、正文、次级标签仍清晰且没有文字被裁切。若某处被裁切，说明该处硬编码了行高——记下来交给 Task 8（`FontSize` 接入后行高应随之而来）。

- [ ] **Step 6: 提交**

```bash
git add ui/core/core.go ui/core/core_test.go
git commit -m "fix(core): Use must not drop the desktop font size

Use builds a ui.Theme by hand and omitted FontSize and Font, so
c.Theme().FontSize was 0 after every frame's first call and a plain
ui.Text laid out at 16.49pt instead of 15.31pt. The desktop's text-scale
setting was silently discarded for all 22 packages."
```

---

## Task 7: `IsDark()` 改相对亮度比较

`ui/theme/theme.go:67-69` 的实现是 `k.Background != Light().Background`（`ui.Color` 是可比较结构体，故为相等比较）。文档（`theme.go:66` 与 `design-system.md §2.3`）说的是「by comparing the surfaces it was built with」「比较两个表面的亮度」。后果：任何 `Background` 不是 `#ffffff` 的自定义浅色调色板（品牌米白 `#fafafa` 即是）会被判成深色，经 `ui/core/core.go:96` 推给 MyGo 的 `Theme.Dark`，整窗用深色字形。

**Files:**
- Modify: `ui/theme/theme.go:67-69`
- Test: `ui/theme/theme_test.go`

**Interfaces:**
- Consumes: `github.com/egoist/mygo/ui.Color`
- Produces: `Tokens.IsDark() bool` 语义修正，签名不变。

- [ ] **Step 1: 写失败测试**

在 `ui/theme/theme_test.go` 追加：

```go
func TestIsDarkComparesLuminanceNotIdentity(t *testing.T) {
	brand := Light()
	brand.Background = ui.Hex("#fafafa") // 仍是浅色窗口
	if brand.IsDark() {
		t.Error("a near-white custom palette was classified as dark")
	}
	night := Dark()
	night.Background = ui.Hex("#101012")
	if !night.IsDark() {
		t.Error("a near-black palette was classified as light")
	}
	if !Dark().IsDark() || Light().IsDark() {
		t.Error("the two shipped palettes must classify as dark and light respectively")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/theme/ -run TestIsDarkComparesLuminanceNotIdentity -count=1 -v`
Expected: FAIL，第一条报错。

- [ ] **Step 3: 实现**

把 `ui/theme/theme.go` 的 `IsDark` 换成亮度比较。相对亮度用 ITU-R BT.601 的加权（与该文件既有的对比度测试同源，读 `ui/theme/theme_test.go` 里已有的 lumada 辅助函数并复用它）：

```go
func (k Tokens) IsDark() bool {
	return relativeLuminance(k.Background) < 0.5
}
```

若 `theme_test.go` 里没有可复用的亮度函数，就在 `ui/theme/theme.go` 内加一个未导出的 `relativeLuminance(ui.Color) float64`，并在 `ui/theme/theme_test.go` 里给 `TextMuted`/`SurfaceHover`、`TextFaint`/`SurfaceHover` 补上断言——实测前者 4.47:1（低于该测试自己在别处强制的 4.5 下限）、后者 2.86:1。补断言后若失败，调整这两个色值直到满足，并把新值写进 `docs/design-system.md §2.1` 的表格。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/theme/ ./ui/core/ -count=1`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add ui/theme/theme.go ui/theme/theme_test.go docs/design-system.md
git commit -m "fix(theme): IsDark must compare luminance, not palette identity

IsDark was 'Background != Light().Background'. Any caller-supplied light
palette whose background is not exactly #ffffff — a brand off-white, say —
was classified dark and pushed into MyGo's Theme.Dark, so the whole window
drew dark glyphs on a light desktop.

Also closed two gaps in the contrast matrix: TextMuted on SurfaceHover was
4.47:1 and TextFaint on SurfaceHover 2.86:1, neither asserted."
```

---

## Task 8: `Reduced` 与 `FontSize` 接入桌面偏好

`ui/core/state.go:85-107` 的 `Reduced`/`FontSize` 只读 `WithReducedMotion`/`WithTextScale` 写的 `ui.Local` 槽位。**全仓库 `c.Preferences()` 调用数为 0。** 两者的文档分别写着「whether the desktop asked for reduced motion」和「the desktop's text scale applied to a library size」——都不成立。

**Files:**
- Modify: `ui/core/state.go:85-107`
- Test: `ui/core/core_test.go`

**Interfaces:**
- Consumes: `github.com/egoist/mygo/ui.Context.Preferences()`（返回 `ui.Preferences`；`Tester.SetPreferences(ui.Preferences)` 可注入）；MyGo `v0.2.16/ui/headless.go` 中 `Tester` 有 `SetPreferences`
- Produces: `Reduced`、`FontSize` 的默认来源改为桌面偏好；`WithReducedMotion`/`WithTextScale` 仍作为测试与无桌面窗口场景的显式覆盖，优先级最高。

- [ ] **Step 1: 写失败测试**

在 `ui/core/core_test.go` 追加：

```go
func TestReducedAndFontSizeFollowTheDesktop(t *testing.T) {
	var reduced bool
	var scaled float32
	tt := ui.NewTester(func(c *ui.Context) {
		core_Use(c)
		reduced = Reduced(c)
		scaled = FontSize(c, 20)
	}, 200, 100)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1.5})
	tt.Frame()
	if !reduced {
		t.Error("ReduceMotion was set but Reduced() reported false")
	}
	if scaled != 30 {
		t.Errorf("TextScale 1.5 on a 20pt size gave %v, want 30", scaled)
	}
}
```

`core_Use` 用测试文件里既有的调用 `Use(c, Settings{})`（若该测试文件在 `package core_test`，写 `core.Use`）。字段名以 MyGo `v0.2.16/ui` 里 `Preferences` 结构体的实际定义为准——**先读它**：`grep -n -A 12 "type Preferences struct" $(go env GOMODCACHE)/github.com/egoist/mygo@v0.2.16/ui/*.go`。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/core/ -run TestReducedAndFontSizeFollowTheDesktop -count=1 -v`
Expected: FAIL，两条断言都错。

- [ ] **Step 3: 实现**

把两个读取函数的初值闭包改为读桌面偏好：

```go
func Reduced(c *ui.Context) bool {
	return *ui.Local(c.Root(), reducedKey{}, func() bool { return c.Preferences().ReduceMotion })
}
```

`FontSize` 同理，初值闭包返回 `c.Preferences().TextScale`，并在返回前处理 `<= 0`（沿用现有的「非正数视为 1.0」的兜底）。`WithReducedMotion`/`WithTextScale` 不动——它们写的是同一个槽位，因此仍是显式覆盖，优先级最高，这正是测试与无桌面窗口场景需要的。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/core/ -count=1` → PASS

- [ ] **Step 5: 检查硬编码绕过点**

审核发现 `core.FontSize(c, …)` 在 14 处被绕过，包括 `ui/navigation/rail.go:42` 硬编码 `17`、`ui/feedback/feedback.go:106` 硬编码 `22`、`ui/data/card.go:62,65,68,108`、`ui/navigation/filter.go:59,64,110`、`ui/feedback/feedback.go:78,81,87,106,131`。

```bash
grep -rn "FontSize(theme\.\|FontSize([0-9]" ui/ --include='*.go' | grep -v _test
```

把其中的裸 `FontSize(...)` 改成 `core.FontSize(c, ...)`，使其跟随桌面缩放。**只改这一类绕过点，不做其他清理。** 每改一处跑一次 `go test ./ui/... -count=1`。

- [ ] **Step 6: 目视复核**

Run: `go run ./cmd/gallery -shots "$TMPDIR/mintui-t8" -scale 2`

字号来源现在统一了。若某处文字溢出容器，是该处硬编码了容器高度/宽度——记在提交信息里，不要在本任务扩大范围。

- [ ] **Step 7: 提交**

```bash
git add ui/core/state.go ui/core/core_test.go ui/navigation/rail.go ui/navigation/filter.go ui/feedback/feedback.go ui/data/card.go
git commit -m "fix(core): honour the desktop's reduce-motion and text-scale

Reduced and FontSize read only the slots WithReducedMotion and
WithTextScale write; c.Preferences() appeared nowhere in the repository.
Both doc comments claimed to report what the desktop asked for, so the
accessibility settings did nothing.

Also routed the 14 call sites that bypassed core.FontSize back through it,
including a hardcoded 17 in navigation/rail.go and 22 in feedback."
```

---

## Task 9: `overlay` 模态契约——`NonModal` 生效，Esc 顺序修正

两处耦合，必须一起改：

1. `ui/overlay/layer.go:38,49` 两个分支都**无条件** `.Modal()`，且 `layer.go:60` 的 `back.Clicked()` 在 `modal` 之外被读。所以 `Dialog`/`Drawer` 的 `NonModal` 选项无效：非模态层照样让整窗失焦、照样吞掉背景点击。直接推翻 `dialog.go:29-31` 的文档承诺。
2. `ui/overlay/dialog.go:161` 读不到 Esc：`layer()` 在 `layer.go:60` 就读了 `back.OverlayShortcut(0, ui.KeyEscape)`，**早于** `back.Children` 里 body 的读取，投递已被消费。实测 Esc 能关窗但 `Chosen()` 返回 `-1` 而非 Cancel 索引。既有的 `TestAlertDialogEscapeCancels`（`ui/overlay/overlay_test.go:353`）只断言 `*open` 与画出的文字，正好漏掉。

**Files:**
- Modify: `ui/overlay/layer.go:38,49,60`、`ui/overlay/dialog.go:161`
- Test: `ui/overlay/overlay_test.go`

**Interfaces:**
- Consumes: `ui.Box(c).Modal()`、`ui.Element.OverlayShortcut(uint, ui.Key) bool`、`ui.Element.Clicked() bool`
- Produces: `NonModal` 在 `Dialog`/`Drawer` 上真正生效；`AlertDialogResult.Chosen()` 在 Esc 下返回 Cancel 索引。签名不变。

- [ ] **Step 1: 写失败测试**

在 `ui/overlay/overlay_test.go` 追加两条：

```go
func TestANonModalDialogLeavesThePageBehindItClickable(t *testing.T) {
	behind := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		behindPressed := input.Button(c, "behind", input.ButtonOptions{}).Clicked()
		if behindPressed {
			behind++
		}
		open := true
		Dialog(c, &open, DialogOptions{NonModal: true, Title: "t", Body: ui.Text(c, "in")})
	}, 400, 300)
	tt.Click("behind")
	if behind != 1 {
		t.Errorf("a NonModal dialog swallowed the press behind it (%d)", behind)
	}
}

func TestAlertDialogEscapeChoosesCancel(t *testing.T) {
	chosen := -1
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		open := true
		r := AlertDialog(c, &open, AlertDialogOptions{
			Title: "t", Message: "m",
			Actions: []string{"Keep", "Discard"},
		})
		if c := r.Chosen(); c >= 0 {
			chosen = c
		}
	}, 400, 300)
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if chosen != 1 {
		t.Errorf("Escape chose %d, want the Cancel index 1", chosen)
	}
}
```

`Chosen()` 的索引基准以 `ui/overlay/dialog.go` 的 `AlertDialogResult.Chosen` 实现为准——若它返回的是 `-1` 表示无选择、`0` 起算的 `Actions` 下标，则 Cancel 是 `len(Actions)-1`；按实际调整期望值，并在提交信息里写明。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/overlay/ -run 'TestANonModalDialog|TestAlertDialogEscapeChoosesCancel' -count=1 -v`
Expected: 两条都 FAIL。

- [ ] **Step 3: `layer.go` 按 `modal` 门控**

在 `ui/overlay/layer.go` 里，把两处 `.Modal()` 改为条件调用，并给 `back.Clicked()` 加上同样的条件：

```go
		back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Center()
		if modal {
			back = back.Modal()
		}
		if anchor != ui.Center {
			back = ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).AlignItems(anchor)
			if modal {
				back = back.Modal()
			}
		}
		if modal && back.Clicked() {
			*open = false
		}
		if modal && back.OverlayShortcut(0, ui.KeyEscape) {
			*open = false
		}
```

`layer.go:90-99` 已有的 `anchored` 函数是 Popover/HoverCard/Popconfirm 的另一条路径，它已经正确地按 `modal` 分流（`if !modal { return PopoverBase(...) }`）——**不要动它**。

- [ ] **Step 4: `dialog.go` 的 Esc 顺序**

`layer()` 的签名是 `layer(c, open, modal, anchor, body func(back, panel *ui.Element))`，且 `back.Children(...)` 在 `layer.go` 内于 Esc 读取**之后**执行。要让 body 能读到 Esc，必须让 Esc 的读取发生在 body 之后。

改法：把 Esc 的处理从 `layer()` 移进 `Children` 之后。具体做法是在 `layer()` 的 `back.Children` 闭包**结束后**再读一次 Esc——但同一帧内两次 `OverlayShortcut` 只有一次有效投递，所以正确做法是**只保留一处读取，且放在最后**。

因此：在 `layer()` 里删掉 `layer.go:60` 的 Esc 读取，改成在 `back.Children(func(){...})` 调用**之后**读：

```go
		back.Children(func() {
			// ...原有的 body 与 scrim 绘制...
		})
		// Escape is read after the body so a body that handles Escape itself
		// (AlertDialog, which must also learn which action was chosen) gets
		// the delivery first. Reading it earlier consumed the event and the
		// body's read always came back false.
		if modal && back.OverlayShortcut(0, ui.KeyEscape) {
			*open = false
		}
```

`AlertDialog` 侧的 `dialog.go:161` 那行 `back.OverlayShortcut(0, ui.KeyEscape)` 保持在 body 内不动——它现在会先于 `layer()` 的兜底读取拿到投递。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./ui/overlay/ -count=1`
Expected: PASS，含既有的 `TestAlertDialogEscapeCancels`。

- [ ] **Step 6: 验证 Popconfirm 的 Esc 语义没有回归**

`design-system.md §19.5` 记录：非模态的 `Popconfirm` **不**响应 Esc（Esc 只送给模态层）。跑：

Run: `go test ./ui/overlay/ -run Popconfirm -count=1 -v` → PASS
Run: `go run ./cmd/gallery -check` → 与 Task 4 之后的状态一致，不新增失败。

若 `Popconfirm` 的 Esc 行为变了，说明 `anchored` 路径被误改——回退 Step 3 里关于 `anchored` 的部分。

- [ ] **Step 7: 提交**

```bash
git add ui/overlay/layer.go ui/overlay/dialog.go ui/overlay/overlay_test.go
git commit -m "fix(overlay): make NonModal mean something, and let AlertDialog see Escape

layer() called .Modal() unconditionally in both branches and read
back.Clicked() outside the modal guard, so a NonModal Dialog or Drawer
still made the window inert and swallowed presses on the page behind it.

layer() also read OverlayShortcut(Escape) before back.Children ran the
body, consuming the delivery. AlertDialog reads Escape in its body to
learn which action was chosen, so Chosen() always came back -1: Escape
closed the alert but reported no choice. The existing test only asserted
the flag and the drawn text, so it never saw this.

The anchored path Popover/HoverCard/Popconfirm use already gated on modal
correctly and is left alone, preserving design-system.md §19.5: a
non-modal Popconfirm must not close on Escape."
```

---

## Task 10: `datetime` 两处崩溃

**Files:**
- Modify: `ui/datetime/timers.go:262`、`ui/datetime/format.go:265-290`
- Test: `ui/datetime/datetime_test.go`

**Interfaces:**
- Consumes: 无新依赖
- Produces: `DurationText(time.Duration) string` 不再递归、对 `math.MinInt64` 安全、对亚秒返回非空；`Stopwatch` 的 Reset 在 `opts.Laps == nil` 时不崩。签名不变。

- [ ] **Step 1: 写失败测试**

在 `ui/datetime/datetime_test.go` 追加：

```go
func TestDurationTextSurvivesTheMostNegativeDuration(t *testing.T) {
	if got := DurationText(math.MinInt64); got == "" {
		t.Error("DurationText(math.MinInt64) returned an empty string")
	}
}

func TestDurationTextSaysSomethingAboutSubSecondDurations(t *testing.T) {
	if got := DurationText(500 * time.Millisecond); got == "" {
		t.Error("DurationText(500ms) returned an empty string, but its doc says " +
			"zero is \"0m\" rather than an empty string")
	}
}

func TestStopwatchResetWithoutALapListDoesNotPanic(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		start := time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)
		Stopwatch(c, StopwatchOptions{Start: &start, Now: start})
	}, 400, 300).Click("Reset")
}
```

`Click` 的目标标签以 `ui/datetime/timers.go` 里 Reset 按钮的 `Label` 为准（该文件里已有 `reset` 常量，先读出来确认它的字面值）。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/datetime/ -run 'TestDurationText|TestStopwatchReset' -count=1 -v`
Expected: `TestDurationTextSurvivesTheMostNegativeDuration` 会**杀死测试进程**（`fatal error: stack overflow`，`recover()` 无效）；`SubSecond` 返回空串 FAIL；`StopwatchReset` panic。

栈溢出那一条会中断整个包的测试——**先只跑后两条**确认它们失败，再单独跑第一条并预期进程死掉，把这个事实记录下来。

- [ ] **Step 3: `DurationText` 不递归**

把 `ui/datetime/format.go:265-268` 的自递归改成对绝对值操作：

```go
func DurationText(d time.Duration) string {
	neg := d < 0
	// Work on the magnitude. Negating math.MinInt64 overflows back to itself,
	// and recursing on that value never terminated.
	if neg {
		d = -d
	}
	// ...原有的取模/整除逻辑不变...
	// 最后拼上 neg 的 "-" 前缀
}
```

同时补亚秒兜底：在既有的「小时/分钟」分解之前，若 `d > 0 && d < time.Minute`，返回毫秒或秒的形式（按现有输出的风格，例如 `"0.5s"` 或 `"500ms"`——**选一种并保持全函数一致**）。文档注释同步更新。

- [ ] **Step 4: `Stopwatch` 的 Reset 守卫**

`ui/datetime/timers.go:262` 的 `*opts.Laps = nil` 改为：

```go
			if b := input.Button(c, reset, input.ButtonOptions{Label: reset}); b.Clicked() {
				*opts.Start = time.Time{}
				if opts.Laps != nil {
					// Truncate rather than nil: nilling hands the caller a nil
					// slice and loses their backing array.
					*opts.Laps = (*opts.Laps)[:0]
				}
				r.reset = true
			}
```

用 `[:0]` 而非赋 `nil`，理由写进注释——这与 `MyGo` 对 nil slice 的处理无关，是 MintUI 自己的选择：调用方的底层数组不该被一次 Reset 夺走。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./ui/datetime/ -count=1`
Expected: PASS，包含三条新测试。栈溢出那条现在必须能正常返回。

- [ ] **Step 6: 提交**

```bash
git add ui/datetime/timers.go ui/datetime/format.go ui/datetime/datetime_test.go
git commit -m "fix(datetime): two reachable crashes

DurationText recursed on itself for negative durations, so
math.MinInt64 negated back to itself and the process died with a stack
overflow that recover() cannot catch. It now works on the magnitude.

DurationText also returned \"\" for sub-second durations, contradicting its
own doc and drawing an empty duration line in EventPopover and
EventEditor.

Stopwatch's Reset wrote *opts.Laps unconditionally while Laps is
documented optional, so StopwatchOptions{Start: &t} plus Reset was a nil
dereference. Every existing test passes Laps, so nothing saw it."
```

---

## Task 11: `navigation` 两处

**Files:**
- Modify: `ui/navigation/tabs.go:330-347`、`ui/navigation/palette.go:195`
- Test: `ui/navigation/tabs_test.go`、`ui/navigation/navigation_more_test.go`

**Interfaces:**
- Consumes: 无新依赖
- Produces: `windowAround(cur, total, w int) []int` 在 `cur >= total` 时不再算出负容量；`CommandPaletteOptions.NonModal` 的 Esc 承诺成立。均为包内函数/既有字段，签名不变。

- [ ] **Step 1: 写失败测试**

```go
func TestWindowAroundSurvivesAPageBeyondTheEnd(t *testing.T) {
	got := windowAround(9, 3, 2)
	for _, p := range got {
		if p < 0 || p >= 3 {
			t.Fatalf("windowAround(9, 3, 2) = %v, which is out of range", got)
		}
	}
}
```

放在 `package navigation`（不是 `navigation_test`）的测试文件里，因为 `windowAround` 未导出。

再加一条针对 palette 的：把 `TestCommandPaletteEscapeClosesWhenNonModal` 写进 `ui/navigation/navigation_more_test.go`，断言 `NonModal: true` 时按 Esc 面板关闭。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/navigation/ -run TestWindowAroundSurvivesAPageBeyondTheEnd -count=1 -v`
Expected: panic，`makeslice: cap out of range`。现状 `hi` 被夹到 `total-1 = 2`，而 `lo = cur - w = 7`，于是 `make([]int, 0, 2-7+1)`。

- [ ] **Step 3: 夹紧 `lo`**

`ui/navigation/tabs.go` 的 `windowAround` 改为：

```go
func windowAround(cur, total, w int) []int {
	if total <= 0 {
		return nil
	}
	// Clamp both ends: a filter that shrinks the result set under a stale
	// page leaves cur past the end, and an unclamped lo made a negative
	// slice capacity, which panics.
	lo, hi := cur-w, cur+w
	if lo < 0 {
		lo = 0
	}
	if hi >= total {
		hi = total - 1
	}
	if lo > hi {
		lo = hi
	}
	out := make([]int, 0, hi-lo+1)
	for i := lo; i <= hi; i++ {
		out = append(out, i)
	}
	return out
}
```

- [ ] **Step 4: `Pagination` 越界页不静默画空**

`ui/navigation/tabs.go:268` 附近：`Pagination` 对 `Pages <= 0` 已 panic，但对越界的 `*Page` 既不夹紧也不 panic，1–2 页超出时 `windowAround` 返回空切片，于是**两个箭头之间一个页码都没有**。加上夹紧：把越界的 `*Page` 收到 `[0, Pages-1]`，并在越界时按 §14.2.3 `panic("navigation: …")`——因为越界页意味着调用方的状态被外部改坏，MintUI 的契约是停下来而不是猜。

先读该函数确认 `*Page` 的处理位置，再改。

- [ ] **Step 5: `CommandPalette` 的 NonModal**

`ui/navigation/palette.go:195` 的文档承诺「Escape closes it either way」，实测不成立（`layer.go:60` 把 Esc 门控在 `modal` 上）。按 spec §7 的记录，`PopoverBase` 的背板吞外部点击是同一堵墙。

改法：`NonModal` 为真时，在 palette 自己的 body 里额外读一次 Esc 并关闭，与 `layer()` 的兜底并存（同一帧两处 `OverlayShortcut`，body 里的先读到）。

- [ ] **Step 6: 跑测试确认通过**

Run: `go test ./ui/navigation/ -count=1`
Expected: PASS。`TestWindowAroundKeepsTheCurrentPageVisible` 仍绿。

- [ ] **Step 7: 提交**

```bash
git add ui/navigation/tabs.go ui/navigation/palette.go ui/navigation/tabs_test.go ui/navigation/navigation_more_test.go
git commit -m "fix(navigation): clamp the page window, and honour NonModal Escape

windowAround clamped hi but not lo, so cur past the end — which a filter
shrinking the result set produces — gave a negative slice capacity and
panicked with makeslice: cap out of range. The existing test only ever fed
cur inside the range.

Pagination also drew no page numbers at all for a page one or two past the
end: two arrows and nothing between them. It now clamps, and panics per
§14.2.3 when the caller hands it an out-of-range page.

CommandPaletteOptions.NonModal promised Escape would close it either way;
layer.go gates Escape on modal, so it did not."
```

---

## Task 12: `input` 三处

**Files:**
- Modify: `ui/input/number.go:428-437`、`ui/input/masonry.go:126-127`、`ui/input/shared.go:260-280`
- Test: `ui/input/input_test.go`、`ui/input/extra_test.go`

**Interfaces:**
- Consumes: `strconv.ParseFloat`、`math.IsNaN`/`IsInf`、`ui.Element.Key(string)`
- Produces: `Masonry` 默认路径返回 `*ui.Element` 而非 `nil`；`optionRow` 产出的行带 `Label` 与 `Key`；`NumberInput`/`CurrencyInput` 拒绝非有限值。`Masonry` 的签名不变（返回类型本就是 `*ui.Element`，只是内容错了）。

- [ ] **Step 1: 写失败测试**

```go
func TestMasonryReturnsAnElementOnItsDefaultPath(t *testing.T) {
	var el *ui.Element
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// MasonryItem is {Label string; Height float32}.
		items := []MasonryItem{{Label: "a", Height: 40}, {Label: "b", Height: 40}}
		el = Masonry(c, MasonryOptions{Items: items})
	}, 400, 300)
	if el == nil {
		t.Fatal("Masonry returned nil with MaxHeight unset, so the caller cannot size it")
	}
}

// Every option row in every dropdown must be nameable, so the drawer is
// reachable by what a reader calls it rather than by sibling position.
func TestOptionRowsCarryTheirLabel(t *testing.T) {
	sel, query := "ac", ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// SelectSearch(c, selected, query *string, choices []Choice, opts SelectSearchOptions)
		SelectSearch(c, &sel, &query, []Choice{
			{Value: "ac", Label: "Air conditioning"},
			{Value: "leak", Label: "Water leak"},
		}, SelectSearchOptions{Label: "Kind"})
	}, 400, 300)
	tt.Click("Kind")
	tt.Open("Kind")
	if !tt.HasText("Air conditioning") {
		t.Fatal("the option list did not open")
	}
}
```

`optionRow` 的 `Key` 修复由既有测试覆盖不到（焦点跳位需要连续两次击键），在 Step 4 之后手工验证：`tt.Type("wa")` 收窄选项后，焦点仍应停在 "Water leak" 上。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/input/ -run TestMasonryReturnsAnElementOnItsDefaultPath -count=1 -v`
Expected: FAIL，`el == nil`。

- [ ] **Step 3: `Masonry` 返回 `row`**

`ui/input/masonry.go:126-127` 的

```go
	build()
	return nil
```

改为让 `build` 返回那个 row：

```go
	build := func() *ui.Element { … ; return row }
	if opts.MaxHeight > 0 {
		return layout.ScrollArea(c, …, func() { build() }).Element
	}
	// Returning the row is what lets the caller size it: Masonry(...).Grow(1)
	// on the nil we used to return was a nil dereference, and nothing in the
	// tests touched the return value.
	return build()
```

注意 `layout.ScrollArea` 的 `children` 是 `func()`，所以那条分支仍用 `func() { build() }`。

- [ ] **Step 4: `optionRow` 挂 `Label` 与 `Key`**

`ui/input/shared.go:260-280` 的 `optionRow(c, label, f)`：

1. 在链上补 `.Label(label)`——这修掉「全库每一个下拉选项行都是无名可聚焦控件」，也让 `ui/input/transfer.go:192` 那处不一致的 `.Tooltip(...)` 补丁可以删掉。
2. 补 `Key`。选项行按 `matchingChoices(...)` 的顺序生成，而该顺序**每次击键都会变**，所以焦点/控件状态会跳到别的项上。给 `optionRow` 增加一个 `key string` 参数（放在 `label` 之后），调用点（`ui/input/select.go` 的三处、`ui/input/transfer.go:187`）传该选项的稳定标识（`Choice.Value`，或调用方已有的等价字段）。

```go
func optionRow(c *ui.Context, label, key string, f optionFace) *ui.Element {
	row := ui.ButtonBase(c).FillWidth().Justify(ui.Start).AlignItems(ui.Center).Gap(u).
		Padding(u*0.75, u*1.5).Radius(theme.SmallRadius).TextColor(k.Text)
	if key != "" {
		// Option rows are built in match order, which changes on every
		// keystroke. Without a key, MyGo carries focus by sibling position
		// and narrowing a SelectSearch moves the focus onto a different
		// option.
		row = row.Key(key)
	}
	row = row.Label(label)
	…
}
```

- [ ] **Step 5: `number.go` 拒绝非有限值**

`ui/input/number.go:428-437` 的 `readPlainNumber` 在 `strconv.ParseFloat` 成功后加一道守卫：

```go
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		// ParseFloat accepts "NaN" and "Inf" without error, and clamping
		// cannot repair a NaN: NaN < min is false. Typing NaN would write
		// NaN into the caller's value permanently and the field would
		// render "NaN" forever.
		return 0, false
	}
```

第二返回值按该函数既有的错误约定填写（先读它的签名确认 `0, false` 是否是正确的失败表示）。

- [ ] **Step 6: 跑测试确认通过**

Run: `go test ./ui/input/ -count=1`
Expected: PASS。既有测试全绿——若某个既有测试依赖 `Masonry` 返回 `nil`，那是它编码了 bug，改测试并在提交信息里说明。

- [ ] **Step 7: 提交**

```bash
git add ui/input/masonry.go ui/input/number.go ui/input/shared.go ui/input/select.go ui/input/transfer.go ui/input/input_test.go ui/input/extra_test.go
git commit -m "fix(input): Masonry returned nil, option rows were unnamed, NaN stuck

Masonry's default path (MaxHeight unset) rendered the row and then
returned nil, so Masonry(...).Grow(1) was a nil dereference and the
caller could not size it. The tests only exercised packMasonry and the
panic paths, never the return value.

optionRow never called .Label, so every option row in every dropdown was
an unnamed focusable control leaning on MyGo's innerText fallback;
transfer.go had patched one call site with a Tooltip because of it. Rows
are also built in match order, which changes per keystroke, and had no
Key, so narrowing a SelectSearch moved focus to a different option.

readPlainNumber accepted ParseFloat's NaN and Inf, and clamping cannot
repair NaN because NaN < min is false, so typing NaN wrote NaN into the
caller's value permanently."
```

---

## Task 13: `agent` 审批写入 `Options` 副本

`ui/agent/diff.go:632-635`：nil 的 `Approved` map 被赋成一个**局部** map，于是 `opts.Approved[f.Path] = …` 写的是 `Options` 值拷贝，签署结果在帧尾蒸发。`ui/agent/diff.go:579` 的文档却写着「writes into the caller's set」。

**Files:**
- Modify: `ui/agent/diff.go:632-635`
- Test: `ui/agent/agent_test.go`

**Interfaces:**
- Consumes: `ui/agent.ReviewOptions.Approved map[string]bool`
- Produces: `Approved` 必须由调用方传入非 nil map，否则 `panic("agent: …")`。不新增字段。

- [ ] **Step 1: 写失败测试**

```go
func TestReviewApproveWritesIntoTheCallersSet(t *testing.T) {
	approved := map[string]bool{}
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MultiFileDiffReview(c, MultiFileDiffReviewOptions{
			Files:    []ReviewFile{{Path: "a.go"}},
			Approved: approved,
		})
	}, 600, 400).Click("Approve")
	if !approved["a.go"] {
		t.Error("the approval did not reach the caller's map")
	}
}
```

被点按钮的 `Label` 以 `ui/agent/diff.go` 内该 demo 的实际字面量为准（先 grep 出 `Label:` 的值再改测试）。

**先判定清楚是哪种坏法**：把测试跑两遍——一遍传非 nil 的 `map[string]bool{}`（预期**通过**，因为非 nil 时 `opts.Approved` 的 map 头与调用方共享），一遍传 nil（预期失败）。若第一遍就失败，说明写入路径本身也断了，那么本任务同时要修非 nil 路径。**在提交信息里写明实际是哪种。**

`MultiFileDiffReview(c, MultiFileDiffReviewOptions)`（`ui/agent/diff.go:583`），其 Options 有 `Selected *int` 与 `Approved map[string]bool`（`:559`）；文件类型是 `ReviewFile`（`:510`）。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/agent/ -run TestReviewApproveWritesIntoTheCallersSet -count=1 -v`
Expected: 至少 nil 那一遍 FAIL。

- [ ] **Step 3: 实现**

把 nil 兜底换成 panic：

```go
				}, func() {
					if opts.Approved == nil {
						// Options arrives by value, so a map we allocate here
						// would be written into this frame's copy and lost at
						// frame's end. Ask for the map instead of inventing one.
						panic("agent: Review needs a non-nil Approved map to write into")
					}
					opts.Approved[f.Path] = !opts.Approved[f.Path]
					r.approved = f.Path
				})
```

在组件入口处（`Review` 函数体开头）也做一次同样的检查，让 panic 早于任何绘制发生。

- [ ] **Step 4: 修调用点**

`ui/showcase/pages/agent.go` 若传了 nil map，改为传 `map[string]bool{}`。用

```bash
grep -rn "Approved:" ui/ --include='*.go' | grep -v _test
```

找齐全部调用点，逐个修好。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./ui/agent/ ./ui/showcase/... -count=1`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add ui/agent/diff.go ui/agent/agent_test.go ui/showcase/pages/agent.go
git commit -m "fix(agent): an approval has to land in the caller's map

Review's approve handler replaced a nil Approved map with a fresh local
one. Options arrives by value, so the write went into this frame's copy
and was gone at frame's end, while the doc at diff.go:579 promised it
writes into the caller's set. A nil map now panics with the package
prefix."
```

---

## Task 14: `finance` / `code` / `devtools` 三处改指针（破坏性）

三处同源：`Options` 按值传入，处理器写 `opts.X` 就是写副本。确切签名（已核实）：

| 位置 | 组件 | 现状字段 | 改为 |
|---|---|---|---|
| `ui/finance/order.go:143,209` | `OrderEntry(c, OrderEntryOptions)` | `Kind OrderKind`（`kindIndex(&opts.Kind)` 返回指向局部 `i` 的地址，Market/Limit/Stop **永远选不了**） | `Kind *OrderKind` |
| `ui/code/tools.go:279,319` | `FindWidget(c, FindOptions)` | `Query string`、`Replace string`（`ui.TextInputBase(c, &opts.Query)` 绑定局部副本，`Closed()` 撒谎） | `Query *string`、`Replace *string` |
| `ui/devtools/response.go:320,331-336` | `LogStream(c, LogStreamOptions)` | `Level LogLevel`（级别按钮写值副本，过滤**完全失效**） | `Level *LogLevel` |

`LogLevel` 是 `int`，定义在 `ui/devtools/metrics.go:287`。`OrderKind` 是 `int`，定义在 `ui/finance/order.go:19`。

**Files:**
- Modify: 上述三处 + 各自 Options 定义处
- Modify: 全部调用点（`ui/showcase/pages/`、`ui/finance/finance_test.go`、`ui/code/code_test.go`、`ui/devtools/*_test.go`、各 `example_*_test.go`）
- Test: 同上

**Interfaces:**
- Consumes: `input.Segmented(c, selected *int, labels ...string) *ui.Element`（`ui/input/controls.go:17`——把 `*int` 原样转交 MyGo 并**就地改写**，不回传值）；`ui.TextInputBase(c, value *string)`；`ui.Local(root, key, func() T) *T`（`ui/showcase/state.go` 的用法）
- Produces: `finance.OrderEntryOptions.Kind *OrderKind`、`code.FindOptions.Query *string`、`code.FindOptions.Replace *string`、`devtools.LogStreamOptions.Level *LogLevel`。**这三个是 breaking change。**

- [ ] **Step 1: 写失败测试**

```go
func TestOrderKindSegmentActuallyChangesTheOrderKind(t *testing.T) {
	kind := OrderMarket
	price, size := 10.0, 1.0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		OrderEntry(c, OrderEntryOptions{
			Instrument: "AC-1", Kind: &kind, Price: &price, Size: &size,
			Currency: "USD", Label: "Entry",
		})
	}, 500, 400)
	tt.Click("Limit")
	if kind != OrderLimit {
		t.Errorf("clicking Limit left Kind = %v, want %v", kind, OrderLimit)
	}
}
```

`code` 与 `devtools` 同理：类型进查找框后改 `*query`，断言它真的变了；点级别按钮后断言 `*level` 变了。按钮的 `Label` 以各自文件里的实际字面量为准（先 grep 出 `Label:` 的值）。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/finance/ ./ui/code/ ./ui/devtools/ -run 'ActuallyChanges' -count=1 -v`
Expected: FAIL。

- [ ] **Step 3: 改字段类型与内部使用**

逐包改。原则：**凡是需要被交互写入的字段，一律是指针**；只读的保持原样。每处改动后立刻 `go build ./...`。

`finance/order.go` 的 `kindIndex` **删掉**，换成 `ui.Local` 持稳再回写。`Segmented` 只认 `*int` 且不回传值，所以每帧重新播种一个局部 int 是拿不回点击结果的（局部变量帧尾即死，这正是现状的 bug）：

```go
	// Segmented only speaks *int and hands nothing back, so the index needs
	// somewhere to live that outlives the frame. ui.Local keeps one per root;
	// kindIndex used to return the address of a local, so a click wrote into
	// a variable that was gone before the next frame read it.
	idx := *ui.Local(c.Root(), "finance/orderEntry/kind", func() int { return int(*opts.Kind) })
	input.Segmented(c, &idx, "Market", "Limit", "Stop")
	*opts.Kind = OrderKind(idx)
```

`opts.Kind` 为 nil 时 `panic("finance: OrderEntry needs a Kind")`，与 `input.Segmented` 既有的 nil/越界 panic 风格一致。`opts.Kind != OrderMarket` 那处判断相应改为 `*opts.Kind != OrderMarket`。

`code/tools.go` 与 `devtools/response.go` 只需把字段改指针并把 `opts.Query` 等改为 `*opts.Query`；nil 时 `panic("code: FindWidget needs a Query")` / `panic("devtools: LogStream needs a Level")`。

- [ ] **Step 4: 改全部调用点**

```bash
grep -rn "Kind:\|Query:\|Replace:\|Level:" ui/ --include='*.go' | grep -vE "_test\.go|^[^:]+:[0-9]+:(//|type|func)"
```

逐个改为取地址。`ui/showcase/pages/finance.go`、`code.go`、`devtools.go` 里的演示状态改成 `showcase.State` 取指针（参考 Task 4 的做法）。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./... -count=1`
Expected: 全绿。

- [ ] **Step 6: 提交**

这是 breaking change，提交信息要写清迁移方式：

```bash
git add -A ui/finance ui/code ui/devtools ui/showcase
git commit -m "fix!: three controls that looked interactive and were not

Three Options structs are passed by value, so writing a field from an
interaction handler wrote this frame's copy. All three controls rendered
correctly and silently did nothing:

  finance.OrderEntryOptions.Kind — kindIndex handed Segmented the address
    of a local int, so a click was written into a variable that died at
    frame's end and Market/Limit/Stop could not be selected at all.
    Segmented returns *ui.Element and hands nothing back, so the index
    now lives in ui.Local and is copied into the caller's OrderKind.
  code.FindOptions.Query/Replace — bound the find bar to a local string,
    so the caller's query never changed and Closed() lied about it.
  devtools.LogStreamOptions.Level — the level buttons wrote the copy, so
    the level filter never filtered.

BREAKING: these fields are pointers now. Callers write
  OrderEntryOptions{Kind: &kind}  instead of  {Kind: kind}
  FindOptions{Query: &q}          instead of  {Query: q}
  LogStreamOptions{Level: &l}     instead of  {Level: l}"
```

---

## Task 15: 收尾——全量验证与文档同步

**Files:**
- Modify: `docs/design-system.md`（补记本批修复的行为变更）
- Test: 全量

**Interfaces:**
- Consumes: 无
- Produces: 无新符号。

- [ ] **Step 1: 全量验证**

```bash
go build ./... && \
go vet ./... && \
go test ./... -count=1 && \
go run ./cmd/gallery -check && \
go run ./cmd/gallery -shots "$TMPDIR/mintui-final"
```

Expected: 全部 exit 0。若 `-check` 因 Task 3 的 `Gutter` 护栏报出未修的页面，把它作为**已知失败**记录在提交信息里，并列出页名——不要为了让 `-check` 变绿而删掉或放宽护栏。

- [ ] **Step 2: 记录未修完的部分**

把 Task 3 报出的 `Gutter` 违规页、以及审核中评为 Important 但不在本批的项（同帧重复 `Label`、`Stagger` 无界增长、其余领域包的 Important 项），整理成一份清单追加到 `docs/design-system.md` 末尾，标题为「已知未修（批次 0 之后）」。每条一行：`位置 — 问题 — 影响`。

这一步是为了让下一批接手的人不必重跑一遍审核。

- [ ] **Step 3: 同步文档中已被本批推翻的描述**

`docs/design-system.md` 里以下几处现在与代码不符，逐条改正：

- §20 的护栏表：`Gutter` 从「不存在」变成「已实现」，措辞对齐 Task 3 的实现。
- §2.3「实现是比较两个表面的亮度」——现在**才**是真的了（Task 7 之前是错的）。
- §17.2 关于 `Clicked()` 在最后一趟为 false 的说明——核对本批新增的测试是否遵守，遵守则不动。

- [ ] **Step 4: 提交**

```bash
git add -A docs
git commit -m "docs: record batch 0 outcomes and the known-unfixed list

Full suite, -check and -shots all pass. Lists the pages the new Gutter
guardrail still reports, and the Important-severity findings that are not
in this batch, so the next batch does not have to re-audit."
```

---

## 验收

本计划全部完成后，以下每一条都必须成立（与 spec §10 一致）：

1. `go build ./...`、`go vet ./...`、`go test ./... -count=1` 全绿。
2. `go run ./cmd/gallery -check` 在干净树上 exit 0（或已知失败清单已写入文档）。
3. `go run ./cmd/gallery -shots <dir>` 在干净树上 exit 0 ——**这是本批的核心交付**。
4. `Page.Gutter` 护栏实际存在并被调用。
5. `-check` 无法因空 `Want` 而通过。
6. 开减弱动效时每页重绘字节一致。
7. 深色外观真的被渲染过。
8. `IsDark()` 对非 `#ffffff` 的浅色板判为浅色。
9. `c.Preferences()` 被实际调用，`Reduced`/`FontSize` 跟随桌面设置。
10. `NonModal` 在 `Dialog`/`Drawer` 上生效；`AlertDialog` 的 Esc 返回 Cancel 索引。
11. 四处崩溃不可达：`Stopwatch` Reset、`DurationText(MinInt64)`、`windowAround` 越界、`NumberInput` NaN。
12. 四处 Options 值拷贝改为指针或 panic。
13. 文档中被本批推翻的描述已更正，未修项已列明。