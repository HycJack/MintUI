package pages

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	input_selected = "board"
	text           = "callbacks://cb-2871"
	dockOpen       = true
	on             = input.Checked
	states         = []input.CheckState{input.Unchecked, input.Checked, input.Indeterminate}
	input_picked   = []string{"ac"}
	branch         = "north"
	seg            = 1
	idx            = 0
	chips          = []string{"ac", "leak"}
	enabled        = true
	name           = "Riverside Clinic"
	secret         = "cb-2871-secret"
	revealed       = false
	pin            = "4"
	masked         = "2026-10-07"
	input_query    = "riverside"
	input_url      = "https://example.invalid"
	note           = "交给 @Mi 处理"
	cost           = 184.5
	count          = 12.0
	rate           = 0.62
	scrub          = 1400.0
	knob           = 0.55
	vol            = 0.4
	stars          = 3
	who            = "ravi"
	multi          = []string{"andre", "mia"}
	input_search   = "mia"
	sug            = "Mia"
	leaf           = "oven"
	customer       = "Riverside Clinic"
	price          = 184.5
	swatch         = ui.Hex("#1d4ed8")
	tags           = []string{"AC", "Riverside"}
	fractions      = []float32{0.3, 0.45, 0.25}
	chosen         = true
	left           = []string{"andre", "ravi", "sam"}
	right          = []string{"mia", "lena"}
)

func init() {
	showcase.Register(showcase.Page{
		Package: "input",
		Title:   "ui/input — 人操作的控件",
		Note:    "按钮、选择、文本、数字、日期以外的一切：开关、拖拽、颜色、二维码、签名、穿梭",
		Width:   1000,
		Height:  3080,
		// Anchored: this page has a selection bar along the bottom and a
		// drawer down the right, both pinned to the window, so the page
		// is the only thing that knows how tall it is.
		Anchored: true,
		Want: []string{
			"普通",
			"主要",
			"危险",
			"禁用",
			"看板",
			"列表",
			"按住删除",
			"复制链接",
			"已选 2 条",
			"指派",
			"归档",
			"未选",
			"选中",
			"半选",
			"North",
			"Harbour",
			"Riverside Clinic",
			"问题描述",
			"搜客户",
			"Log callback",
			"Customer",
			"Issue",
			"Open cost",
			"金额要大于 0",
			"Save",
			"Cancel",
			"Browse",
			"Ravi Patel",
			"Oven",
			"Harbour / Oven",
			"0.62",
			"20 – 80",
			"Sign here",
			"Clear signature",
			"AC",
			"Riverside",
			"未指派",
			"本班次",
		},
		Render: func(c *ui.Context) {
			inputPage(c)
		},
	})
}

// inputPage draws every control a person can operate. Most of them take a
// pointer to the value they edit, so the state a component needs is declared
// right here in the render: a bool, a string, a slice. Nothing is stored in
// the component, which is why a page can be thrown away and rebuilt.
func inputPage(c *ui.Context) {
	buttonSection(c)
	choiceSection(c)
	inputTextSection(c)
	numberSection(c)
	selectSection(c)
	formSection(c)
	pickerSection(c)
	advancedSection(c)
	transferSection(c)
	bulkBarSection(c)
}

// buttonSection shows the press buttons, the icon-only one, the group, the
// split one and the two that need a moment of holding.
func buttonSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Button 与 IconButton — 一个按下去的动作")
	showcase.Stack(c, 1, func() {
		input.Button(c, "普通", input.ButtonOptions{})
		input.Button(c, "主要", input.ButtonOptions{Primary: true})
		input.Button(c, "危险", input.ButtonOptions{Danger: true})
		input.Button(c, "禁用", input.ButtonOptions{Disabled: true})
	})
	showcase.Stack(c, 1, func() {
		glyph := iconGlyph(c, "check")
		input.IconButton(c, glyph, "标记已完成", input.ButtonOptions{})
		input.IconButton(c, glyph, "主要图标按钮", input.ButtonOptions{Primary: true})
		input.IconButton(c, glyph, "禁用的图标按钮", input.ButtonOptions{Disabled: true})
		input.IconButton(c, glyph, "带图标的按钮", input.ButtonOptions{Icon: glyph})
	})

	showcase.Field(c, "ButtonGroup — 选一个，选中值是调用方的字符串")
	group := &input.GroupButton{
		Value: "board", Label: "看板", Icon: iconGlyph(c, "panel"),
	}
	input.ButtonGroup(c, &input_selected, []input.GroupButton{
		*group,
		{Value: "list", Label: "列表", Icon: iconGlyph(c, "sliders")},
		{Value: "group", Label: "按分支分组", Tip: "子项会自己缩进", Disabled: true},
	}, input.ButtonGroupOptions{Label: "视图", Wrap: true})

	showcase.Field(c, "HoldToConfirm 与 CopyButton — 要按住才生效的那个")
	var held bool
	showcase.Stack(c, 1, func() {
		input.HoldToConfirm(c, &held, input.HoldToConfirmOptions{
			Label: "按住删除", HoldFor: 1200000000, Done: "松手即删", Danger: true,
		})
		input.CopyButton(c, &text, input.CopyButtonOptions{Label: "复制链接"})
	})

	// Dock fills its parent, so it is given a frame with a height of its own
	// rather than a whole page.
	showcase.Field(c, "Dock — 挂在窗口一条边上的一层，内容由调用方给")
	ui.Box(c).Width(unit(c, 56)).Height(unit(c, 26)).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		input.Dock(c, &dockOpen, input.DockOptions{
			Title: "技术员", Subtitle: "这一单派给谁", Side: input.DockLeft,
			Closeable: true,
		}, func() {
			ui.Text(c, "Dock 的内容由调用方给").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		})
	})
}

// choiceSection shows the four ways of choosing, plus the on/off control
// that is not a choice between things.
func choiceSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Checkbox、Radio、Toggle、Chips — 选与不选")
	ui.Row(c).FillWidth().Wrap().Gap(unit(c, 4) * 2).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "Checkbox — 可以全不选").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			// The three positions a box can be in, each named beside itself
			// so the row says which is which without a second line.
			for i := range states {
				input.Checkbox(c, &states[i], boxStateName(i), input.CheckboxOptions{})
			}
			input.Checkbox(c, &on, "禁用", input.CheckboxOptions{Disabled: true})
		})

		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "CheckboxGroup — 多选，选中是切片").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.CheckboxGroup(c, &input_picked, []input.Choice{
				{Value: "ac", Label: "AC"},
				{Value: "leak", Label: "Water leak"},
				{Value: "oven", Label: "Oven"},
			}, input.CheckboxGroupOptions{})
		})

		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "RadioGroup — 恰好一个").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.RadioGroup(c, &branch, []input.Choice{
				{Value: "north", Label: "North"},
				{Value: "harbour", Label: "Harbour"},
				{Value: "closed", Label: "Closed"},
			}, input.RadioGroupOptions{})
		})
	})

	ui.Row(c).FillWidth().Wrap().Gap(unit(c, 4) * 2).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "ToggleGroup — 视图切换").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.ToggleGroup(c, &idx, []string{"看板", "列表", "日历"},
				input.ToggleGroupOptions{})
		})
		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "ChoiceChips — 还能当筛选器用").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.ChoiceChips(c, &chips, []input.Choice{
				{Value: "ac", Label: "AC"},
				{Value: "leak", Label: "Water leak"},
				{Value: "oven", Label: "Oven"},
				{Value: "boiler", Label: "Boiler"},
			}, input.ChoiceChipsOptions{Max: 3, Wrap: true})
		})
		ui.Column(c).Width(unit(c, 40)).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "Segmented 与 Switch").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.Segmented(c, &seg, "▦", "☰", "≡")
			showcase.Stack(c, 1, func() {
				input.Switch(c, &enabled, input.SwitchOptions{Label: "启用通知"})
				input.Switch(c, &enabled, input.SwitchOptions{Label: "标签在前", LabelFirst: true})
			})
		})
	})
}

// textSection shows every field that holds text, from one line to a code
// and a PIN.
func inputTextSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "文本 — 单行、多行、密码、搜索、掩码、PIN")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		issue := "客诉机不制冷，上午报修"
		ui.Column(c).Width(unit(c, 46)).Gap(unit(c, 2)).Children(func() {
			input.TextInput(c, &name, input.TextInputOptions{Label: "客户名", Placeholder: "客户名"})
			input.TextArea(c, &issue, input.TextAreaOptions{Label: "问题描述", Lines: 3})
		})
		ui.Column(c).Width(unit(c, 46)).Gap(unit(c, 2)).Children(func() {
			input.PasswordInput(c, &secret, input.PasswordInputOptions{
				Label: "访问令牌", Placeholder: "令牌", Revealed: &revealed,
			})
			input.PinInput(c, &pin, input.PinInputOptions{Label: "验证码", Length: 6})
			input.MaskedInput(c, &masked, input.MaskedInputOptions{
				Label: "日期", Mask: "YYYY-MM-DD",
			})
		})
	})

	showcase.Field(c, "SearchInput、InputGroup、Banner")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 46)).Gap(unit(c, 2)).Children(func() {
			input.SearchInput(c, &input_query, input.SearchInputOptions{
				Label: "搜客户", Placeholder: "客户名或工单号",
			})
			input.InputGroup(c, &input_url, input.InputGroupOptions{
				Label: "回访链接", Prefix: "https://", Placeholder: "example.invalid",
			})
		})
		ui.Column(c).Width(unit(c, 72)).Gap(unit(c, 2)).Children(func() {
			input.Banner(c, input.BannerOptions{
				Text: "3 条回访因为网络原因没能同步。", Severity: core.Warning,
				Action: "重试", Dismissable: true,
			})
			input.Banner(c, input.BannerOptions{
				Text: "已切到离线模式，改动会等网络回来再送。", Severity: core.Accent,
			})
		})
	})

	showcase.Field(c, "MentionInput — @ 一个人")
	ui.Box(c).Width(unit(c, 62)).Children(func() {
		input.MentionInput(c, &note, input.MentionInputOptions{
			Label: "内部备注", Placeholder: "写点什么",
			People: []string{"Mia Chen", "Ravi Patel", "Lena Ford", "Sam Ortiz"},
		})
	})
	_ = k
}

// numberSection shows the ways a number is entered and adjusted: typed,
// stepped, dragged, scrubbed, turned on a knob.
func numberSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "数字与滑块 — 敲进去的和拖出来的")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 46)).Gap(unit(c, 2)).Children(func() {
			input.NumberInput(c, &count, input.NumberInputOptions{
				Label: "配件数量", Min: ptr(1.0), Max: ptr(99.0), Step: 1,
				Prefix: "×", Suffix: "件",
			})
			input.CurrencyInput(c, &cost, input.CurrencyInputOptions{
				Label: "未结金额", Symbol: "$", Min: ptr(0.0),
			})
		})
		ui.Column(c).Width(unit(c, 46)).Gap(unit(c, 2)).Children(func() {
			input.Slider(c, &rate, input.SliderOptions{
				Min: 0, Max: 1, Step: 0.01, ShowValue: true, Label: "完成度",
			})
			lo := showcase.State(c, "input.341.lo", 20.0)
			hi := showcase.State(c, "input.341.hi", 80.0)
			input.RangeSlider(c, lo, hi, input.RangeSliderOptions{
				Min: 0, Max: 100, Step: 1, ShowValue: true, Label: "预算区间",
			})
			input.ScrubInput(c, &scrub, input.ScrubOptions{
				Min: ptr(0.0), Max: ptr(4000.0), Step: 10,
				Prefix: "$", Format: "%.0f", Label: "左右拖动改预算",
			})
		})
	})

	showcase.Field(c, "Knob 与 VerticalSlider — 旋钮和竖着的滑块")
	ui.Row(c).FillWidth().Gap(unit(c, 5)).AlignItems(ui.Center).
		MarginY(unit(c, 2)).Children(func() {
		ui.Box(c).Width(unit(c, 30)).Children(func() {
			input.Knob(c, &knob, input.KnobOptions{
				Min: ptr(0.0), Max: ptr(1.0), ShowValue: true, Size: unit(c, 22),
				Format: "%.0f%%", Label: "音量",
			})
		})
		ui.Box(c).Width(unit(c, 30)).Height(unit(c, 34)).Children(func() {
			input.VerticalSlider(c, &vol, input.VerticalSliderOptions{
				Height: unit(c, 30), ShowValue: true, Label: "推子",
			})
		})
		ui.Column(c).Grow(1).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "Rating").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.Rating(c, &stars, input.RatingOptions{Max: 5, Label: "回访评分"})
		})
	})
	ui.Box(c).FillWidth().MarginY(unit(c, 1)).Children(func() {
		ui.Text(c, "Masonry — 每一列自己往下堆").TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Bold()
		ui.Box(c).FillWidth().Height(unit(c, 22)).Children(func() {
			input.Masonry(c, []input.MasonryItem{
				{Label: "3 张", Height: unit(c, 10)},
				{Label: "1 张", Height: unit(c, 16)},
				{Label: "2 张", Height: unit(c, 13)},
				{Label: "4 张", Height: unit(c, 18)},
			}, input.MasonryOptions{Columns: 4, Gap: unit(c, 1), Column: unit(c, 24)})
		})
	})
}

// selectSection shows the pickers: a list to choose from, one you can input_search
// in, several at once, and the ones that walk a hierarchy.
func selectSection(c *ui.Context) {
	k := core.Tokens(c)
	choices := []input.Choice{
		{Value: "andre", Label: "Andre Thomson"},
		{Value: "mia", Label: "Mia Chen"},
		{Value: "ravi", Label: "Ravi Patel"},
		{Value: "lena", Label: "Lena Ford"},
		{Value: "sam", Label: "Sam Ortiz"},
	}
	showcase.Section(c, "Select 与它的亲戚 — 从一份名单里挑")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 30)).Gap(unit(c, 2)).Children(func() {
			input.Select(c, &who, choices, input.SelectOptions{
				Label: "技术员", Placeholder: "还没指派",
			})
			input.Select(c, &who, choices, input.SelectOptions{
				Label: "禁用", Placeholder: "锁住了", Disabled: true,
			})
		})
		ui.Column(c).Width(unit(c, 34)).Gap(unit(c, 2)).Children(func() {
			input.MultiSelect(c, &multi, choices, input.MultiSelectOptions{
				Label: "一组技术员", Placeholder: "选了 2 个人", Named: 1,
			})
			input.SelectSearch(c, &who, &input_search, choices, input.SelectSearchOptions{
				Label: "边搜边选", Placeholder: "搜人", Search: "搜技术员",
			})
		})
		ui.Column(c).Width(unit(c, 30)).Gap(unit(c, 2)).Children(func() {
			input.Combobox(c, &sug, []string{"Mia Chen", "Ravi Patel", "Lena Ford"},
				input.ComboboxOptions{Label: "Combobox"})
			input.TreeSelect(c, &leaf, sampleNodes(), input.TreeSelectOptions{
				Label: "TreeSelect", Placeholder: "选一台设备",
			})
		})
	})

	showcase.Field(c, "Cascader — 沿着层级走下来")
	path := []string{"harbour", "oven"}
	input.Cascader(c, &path, sampleNodes(), input.CascaderOptions{
		Label: "设备", Placeholder: "还没选设备", Separator: " / ",
	})
	_ = k
}

// formSection shows the arrangement every record in this interface is edited
// in: a column of fields with a submit row under it.
func formSection(c *ui.Context) {
	showcase.Section(c, "Form 与 FormField — 标签、控件、说明、错误")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		issue := "客诉机不制冷，上午报修"
		ui.Box(c).Width(unit(c, 62)).Children(func() {
			input.Form(c, input.FormOptions{
				Label: "回访记录", Title: "Log callback",
				Description: "记一条回访，提交后会进未处理列。",
				Divider:     true,
				Fields: func() {
					input.FormField(c, input.FormFieldOptions{
						Label: "Customer", Required: true,
					}, func(err string) *ui.Element {
						return input.TextInput(c, &customer, input.TextInputOptions{
							Label: "Customer", Error: err,
						})
					})
					input.FormField(c, input.FormFieldOptions{
						Label: "Issue", Description: "一句话说清对方报了什么",
					}, func(err string) *ui.Element {
						return input.TextArea(c, &issue, input.TextAreaOptions{
							Label: "Issue", Lines: 2, Error: err,
						})
					})
					input.FormField(c, input.FormFieldOptions{
						Label: "Open cost", Error: "金额要大于 0",
					}, func(err string) *ui.Element {
						return input.CurrencyInput(c, &price, input.CurrencyInputOptions{
							Label: "Open cost", Error: err,
						}).Element
					})
				},
				Actions: func() {
					showcase.Stack(c, 1, func() {
						input.Button(c, "Save", input.ButtonOptions{Primary: true})
						input.Button(c, "Cancel", input.ButtonOptions{})
					})
				},
			})
		})

		ui.Column(c).Width(unit(c, 34)).Gap(unit(c, 2)).Children(func() {
			ui.Text(c, "FilePicker 与 FileDropZone").TextColor(core.Tokens(c).Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			path := "/var/lib/callbacks"
			input.FilePicker(c, &path, []input.FileEntry{
				{Path: "/var/lib/callbacks/2026", Name: "2026", Dir: true},
				{Path: "/var/lib/callbacks/photos", Name: "photos", Dir: true, Tip: "12 个文件"},
				{Path: "/var/lib/callbacks/readme.md", Name: "readme.md"},
				{Path: "/var/lib/callbacks/old.csv", Name: "old.csv", Disabled: true},
			}, input.FilePickerOptions{
				Label: "附件目录", Placeholder: "选一个目录", Clearable: true, Tip: "只读目录",
			})
			ui.Box(c).Width(unit(c, 62)).Children(func() {
				input.FileDropZone(c, input.FileDropZoneOptions{
					Label: "把照片拖到这里", Text: "支持 PNG 与 JPG", Multiple: true,
					Height: unit(c, 20),
				})
			})
		})
	})
}

// pickerSection shows the controls that pick a thing out of the world rather
// than out of a list: a colour, a QR code, a signature, a file.
func pickerSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "颜色、二维码、签名、标签")
	// Not FillWidth: a row that fills would hand every spare pixel to the
	// Grow below and push the signature pad's right border onto the page
	// edge, where a picture of the page cannot tell it from a clipped one.
	// The pad is the one control on this page that draws its own edge, so it
	// has to be the one with room left on both sides.
	ui.Row(c).Width(unit(c, 136)).Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 34)).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "ColorPalette 与 ColorPicker").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			input.ColorPalette(c, &swatch, []input.Swatch{
				{Name: "强调", Color: ui.Hex("#1d4ed8")},
				{Name: "成功", Color: ui.Hex("#1f7a3d")},
				{Name: "告警", Color: ui.Hex("#9a5410")},
				{Name: "危险", Color: ui.Hex("#c62b30")},
				{Name: "灰", Color: ui.Hex("#6b6b74")},
			}, input.ColorPaletteOptions{})
			input.ColorPicker(c, &swatch, input.ColorPickerOptions{Label: "自定义颜色"})
		})

		ui.Column(c).Width(unit(c, 56)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "QRCode — 纠错等级").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			qr := "https://example.invalid/cb-2871"
			showcase.Stack(c, 1, func() {
				for i, level := range []input.QRErrorCorrection{
					input.QRLow, input.QRMedium, input.QRHigh, input.QRHighest,
				} {
					qr := qr
					ui.Column(c).AlignItems(ui.Center).Gap(unit(c, 0.5)).Children(func() {
						input.QRCode(c, qr, input.QRCodeOptions{
							Label: "纠错等级样例", Size: unit(c, 13), ErrorCorrection: level,
						})
						ui.Text(c, []string{"Low", "Medium", "High", "Highest"}[i]).
							TextColor(k.TextFaint).FontSize(core.FontSize(c, 9))
					})
				}
			})
		})

		ui.Column(c).Width(unit(c, 42)).Gap(unit(c, 1.5)).Children(func() {
			ui.Text(c, "SignaturePad 与 TagsInput").TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			// Handed real points, not an empty pad: the pad has to offset
			// them by its own box, and the only way a picture proves that
			// happened is by drawing a signature that belongs inside it.
			// It used to paint these at the top left of the window, which is
			// why input/scrub_extra_test.go now reads real pixels back.
			// SignaturePoint is in **pixels relative to the pad**, not in
			// theme units — PointerPosition hands back pixels, so a
			// signature saved yesterday and replayed today has to replay at
			// the same pixel size or it has been silently rescaled. The pad
			// below is unit(42) by unit(18), which comes out 167×72 here,
			// so these points are written against that.
			at := func(x, y float32) input.SignaturePoint {
				return input.SignaturePoint{X: x, Y: y}
			}
			strokes := []input.Stroke{
				{Points: []input.SignaturePoint{ // the name
					at(18, 44), at(30, 20), at(44, 48), at(58, 20),
					at(72, 44), at(84, 30), at(96, 30),
				}},
				{Points: []input.SignaturePoint{ // the A, left stem
					at(22, 60), at(16, 24),
				}},
				{Points: []input.SignaturePoint{ // the A, right stem
					at(112, 60), at(104, 24),
				}},
				{Points: []input.SignaturePoint{ // the A's crossbar
					at(30, 48), at(92, 48),
				}},
			}
			input.SignaturePad(c, &strokes, input.SignaturePadOptions{
				Label: "签名", Width: unit(c, 42), Height: unit(c, 18), Clearable: true,
			})
			// The empty one beside it is the state the pad is in for most of
			// its life, and the only state in which it says what it wants
			// written in it. A page that only showed a signed pad would have
			// a component nobody had ever seen waiting for a person.
			var blank []input.Stroke
			input.SignaturePad(c, &blank, input.SignaturePadOptions{
				Label: "空签名板", Width: unit(c, 42), Height: unit(c, 12),
			})
			ui.Box(c).Width(unit(c, 42)).Children(func() {
				input.TagsInput(c, &tags, input.TagsInputOptions{
					Label: "标签", Placeholder: "加一个标签",
					Suggestions: []string{"urgent", "warranty"},
				})
			})
		})
	})
}

// advancedSection shows the two layout controls a person resizes with, and
// the resizable panes.
func advancedSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "ResizablePanelGroup — 拖把手改大小")
	ui.Box(c).FillWidth().Height(unit(c, 38)).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, k.Border).Clip().Children(func() {
		input.ResizablePanelGroup(c, input.ResizablePanelGroupOptions{
			Fractions: &fractions,
			Labels:    []string{"列表", "详情", "记录"},
			Min:       0.15, Gutter: unit(c, 0.5),
		}, func() {
			panelBlock(c, "列表", k.Surface)
		}, func() {
			panelBlock(c, "详情", k.Background)
		}, func() {
			panelBlock(c, "记录", k.Surface)
		})
	})
}

// bulkBarSection is last on the page on purpose. BulkActionBar draws itself
// in an overlay pinned to the bottom of the window rather than taking a row
// out of the layout, so anywhere else on a page it would float over the
// middle of a stranger's section. Here it floats over the list it belongs
// to, which is the only place a person would ever see it.
func bulkBarSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "BulkActionBar — 选中之后浮在列表上面，不挤动任何一行")
	rows := []string{"Riverside Clinic", "Northgate Dental", "Harbour Cafe", "Elm Street Gym"}
	ui.Box(c).FillWidth().Padding(unit(c, 1)).Radius(theme.ControlRadius).
		Background(k.Surface).Children(func() {
		ui.Column(c).FillWidth().Children(func() {
			for i, r := range rows {
				ui.Row(c).FillWidth().Padding(unit(c, 1.25), unit(c, 1.5)).
					AlignItems(ui.Center).Gap(unit(c, 1.5)).Children(func() {
					ui.Box(c).Size(unit(c, 3), unit(c, 3)).Radius(unit(c, 1.5)).
						Background(k.Border)
					ui.Text(c, r).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.BodySize)).Grow(1)
					ui.Text(c, "CB-287"+strconv.Itoa(i)).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize))
				})
			}
		})
	})
	input.BulkActionBar(c, &chosen, rows[:2], input.BulkActionBarOptions{
		Label: "已选 2 条", Noun: "回访", Clearable: true,
		Actions: []input.BulkAction{
			{Label: "指派"},
			{Label: "归档"},
			{Label: "删除", Danger: true},
			{Label: "导出", Disabled: true},
		},
	})
}

// transferSection shows the two-column chooser.
func transferSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "Transfer — 名单在两边之间搬")
	ui.Box(c).Width(unit(c, 74)).Children(func() {
		input.Transfer(c, &left, &right, input.TransferOptions{
			Left:   "未指派",
			Right:  "本班次",
			Height: unit(c, 30),
			Column: unit(c, 30),
			Choices: []input.Choice{
				{Value: "andre", Label: "Andre Thomson"},
				{Value: "ravi", Label: "Ravi Patel"},
				{Value: "sam", Label: "Sam Ortiz"},
				{Value: "mia", Label: "Mia Chen"},
				{Value: "lena", Label: "Lena Ford"},
			},
			InOrder: true,
		})
	})
	_ = k
}

// panelBlock is one pane's content in the resizable sample. The height is
// stated rather than filled: a pane is a plain box in a row, and a box
// asking a row to fill it is asking the wrong container.
func panelBlock(c *ui.Context, name string, bg ui.Color) {
	k := core.Tokens(c)
	ui.Box(c).FillWidth().Height(unit(c, 37)).Padding(unit(c, 1.5)).Background(bg).Children(func() {
		ui.Text(c, name).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// glyph is the one mark this page's buttons draw. The few input controls
// that take a raw *ui.SVG rather than an icon name get this one; it is the
// same 24×24 stroke style display.Icon uses, so a glyph in a button and a
// glyph in a sidebar are two marks of the same hand.
var glyph = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
	`fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" ` +
	`stroke-linejoin="round"><path d="m5 12.5 4.5 4.5L19 7"/></svg>`))

func iconGlyph(c *ui.Context, name string) *ui.SVG { return glyph }

// sampleNodes is a fixed hierarchy for the tree-shaped pickers.
func sampleNodes() []input.Node {
	return []input.Node{
		{Value: "north", Label: "North", Children: []input.Node{
			{Value: "north-ac", Label: "AC"},
			{Value: "north-leak", Label: "Water leak"},
		}},
		{Value: "harbour", Label: "Harbour", Children: []input.Node{
			{Value: "oven", Label: "Oven"},
			{Value: "fridge", Label: "Fridge", Children: []input.Node{
				{Value: "sealer", Label: "Sealer"},
			}},
		}},
		{Value: "closed", Label: "Closed"},
	}
}

func ptr(v float64) *float64 { return &v }

// boxStateName is what the three positions of a check box are called beside
// it, so the row says which is which without a second line of caption.
func boxStateName(i int) string {
	return [...]string{"未选", "选中", "半选"}[i]
}
