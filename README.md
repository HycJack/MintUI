# MintUI

[MyGo](https://github.com/egoist/mygo) 生态的原生组件库与组件画廊。
没有 WebView：组件由 MyGo 直接绘制（macOS 上是 Metal），状态是调用方的指针，
数据由调用方传入——组件自己不碰磁盘、网络和设备。

## 组件库

`ui/` 下 22 个包，按领域分层，箭头只能向右：

```
theme → core → layout → display / input → overlay / data → chart / media / ...
```

- `theme` — 设计令牌（浅深两套、密度、字号）
- `core` — 每帧入口 `core.Use(c, core.Settings{})`
- `layout` — Container / Board / SplitPane / ScrollArea…
- `display` / `input` / `navigation` / `feedback` — 文本、按钮、选择、侧栏、加载与提示
- `overlay` / `data` — 浮层、表格、看板列
- `chart` / `media` / `code` / `chat` / `finance` / … — 领域组件

## 快速开始

```bash
go run ./cmd/gallery                  # 组件画廊：真实窗口，左侧包列表
go run ./cmd/gallery -check           # 门禁：每页画出承诺的文本，退出码 0
go run ./cmd/gallery -shots out -scale 2 [-dark]   # 全部页面出 PNG
go run ./cmd/gallery -fit             # 每页实际需要的高度
```

`cmd/coverage/main.py` 对 `docs/component-catalogue.baseline.json`
计算组件覆盖率（当前 447/532 = 84%，缺口见 `docs/coverage-report.txt`）。

## 测试

```bash
go test ./... -count=1
```

`ui/showcase/pages/gate_test.go` 是画廊门禁：**没有画廊页的包算不可审查**，
加了新包就要加 `ui/showcase/pages/<pkg>.go`。测试全绿不代表画对了——
尺寸类错误不会 panic，只会画错；出图之后用眼睛真的读一遍。

## 文档

- `docs/design-system.md` — 主题、组件 API、实战坑、分层规则
- `docs/component-catalogue.md` / `.baseline.json` — 组件清单与覆盖率基线

## 用法

```go
core.Use(c, core.Settings{})
theme.Tabs(c, ...)  // 状态是调用方的指针，数据由调用方传入
```

组件把状态视作调用方的指针：组件不存自己的状态，也不持有页面的知识。
