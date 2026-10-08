package pages

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chat"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	edited    = answerOne
	streaming = "Replacing the contactor takes about 0.9 hours, plus the " +
		"drive time to Ridgeway."
	chat_draft  = "把 HC-2A5 的接线图画出来，顺带说说接触器为什么会烧"
	chat_picker = 1
	slashSel    = 0
	mentionSel  = 1
	model       = "claude-opus-4-1"
	mode        = 0
	librarySel  = 0
	system      = "你是 callbacks 的维修助手。只根据工具返回的内容回答，" +
		"不知道就说不知道。"
	v           = 0.3
	collapsed   = false
	query       = "riverside"
	state       = ui.ListState{}
	sel         = 0
	chatProject = "callbacks"
	comment     = "答得对，但漏了停机的等待时间。"
)

// The page is one conversation from the first question to the cost of the
// answer, drawn with every component this package has. The order is the order
// a person reads a conversation in — the turns, what an answer is made of,
// the streaming, the box you type in, the list of other conversations, where
// the facts came from, what a turn can carry — and not the order the files are
// written in, because a gallery is read as a thing and not as an index.
//
// Everything on the page is fixed: the same three turns, the same Go snippet,
// the same four log-shaped lines, the same conversation names. A page that
// drew differently each time it was opened could not be compared against the
// last one, and a screenshot that changes between runs is a screenshot nobody
// reviews.

func init() {
	showcase.Register(showcase.Page{
		Package: "chat",
		Title:   "ui/chat — 一段对话从头到尾",
		Note:    "气泡、转写、markdown、代码块、流式、输入框、会话列表、来源、用量、附件",
		Width:   1000,
		Height:  5700,
		Want: []string{
			// the four roles
			"我, user", "claude opus, assistant", "system message", "tool message",
			"客诉机不制冷了，今天报的修。",
			"我看了一下压缩机接触器，万用表量出来是 24V，线圈额定 220V，所以是接触器本身烧了，不是制冷剂。",
			"已于 14:02 把这条派给 Andre Thomson。",
			"// a comment inside a tool bubble",
			// heads, marks, quotes, editing
			"edited", "Copy", "Retry", "Good answer", "Delete",
			"我: 客诉机不制冷了，今天报的修。", "Edit message", "Save", "Cancel",
			// markdown, and the three things an answer embeds
			"结论", "接触器线圈烧了，不是制冷剂。线圈额定 220V，实测 24V，读数见 docs/design-system.md §2.1。",
			"func ohms(measured, rated float32) float32 {",
			"工时与配件", "HC-2A5 接触器", "bench-wiring.svg", "接线端子",
			// streaming
			"正在想", "读了 ui/core/state.go，看它怎么算密度",
			"Replacing the contactor takes about 0.9 hours, plus the drive time to Ridgeway.",
			"claude opus is typing", "Send", "Stop generating",
			"New messages", "Search finished", "3 of 7",
			// the box you type in
			"PromptComposer", "Read", "Write", "Bash",
			"/compact", "/", "@andre", "Tools", "Hold to talk", "voice note",
			"Claude Opus 4.1", "Mode", "Ask", "Plan", "Build",
			"根因", "System prompt", "Reset all", "temperature",
			// the other conversations
			"Conversations", "Search conversations", "2 of 5", "Riverside Clinic",
			"Maple Street Bakery", "Lakeside Vet", "Export as Markdown",
			"Previous answer", "2 / 3", "Next answer", "asked the same thing again",
			// where the facts came from
			"SourcesPanel", "docs/design-system.md", "[1]", "[3] of 3",
			"接触器线圈额定 220V，通电后应为满压。", "context", "58.2k", "$0.41",
			// what else a turn can carry
			"0:23", "现场接线，线圈那端已经黑了", "The model refused to answer",
			"Feedback", "Report", "report.html", "Capabilities",
			"还没有对话", "Empty", "Drop files to attach",
			// the half that needs no window
			"heading   para   list   code   table", "none   left   center   right",
			"user → user   assistant → assistant   system → system   tool → tool   nonsense → user",
			"浅 #1d4ed8   深 #9dc0ff", "$12.40", "0:01   FormatDuration(3725) 1:02:05",
		},
		Render: func(c *ui.Context) {
			chatPage(c)
		},
	})
}

// The conversation every section on this page is a piece of: a callback about
// an air-conditioning unit that is not cooling. It is written once, here, so a
// name that appears in a bubble, a source card and a conversation item is the
// same name in all three.
const (
	askText   = "客诉机不制冷了，今天报的修。"
	answerOne = "我看了一下压缩机接触器，万用表量出来是 24V，线圈额定 220V，" +
		"所以是接触器本身烧了，不是制冷剂。"
	answerTwo = "更换接触器，型号 HC-2A5。仓库有一个，明早带到现场就行。"
	toolNote  = "// a comment inside a tool bubble"
	toolOut   = "matched 3 files in ui/core, ui/theme and ui/code"
)

func chatPage(c *ui.Context) {
	chatTurnsSection(c)
	chatBlocksSection(c)
	chatStreamSection(c)
	chatComposerSection(c)
	chatConversationSection(c)
	chatSourceSection(c)
	chatPayloadSection(c)
	chatPureSection(c)
}

// ── the turns ──────────────────────────────────────────────────────────────

func chatTurnsSection(c *ui.Context) {
	showcase.Section(c, "转写 · MessageBubble / MessageHeader / MessageActions / QuoteReply")

	showcase.Field(c, "MessageBubble — 四个角色只差位置、墨色和尾巴")
	ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 84)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleUser, Author: "我", Time: "14:02",
			}, func() {
				ui.Text(c, askText)
			})
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleAssistant, Author: "claude opus", Time: "14:03",
			}, func() {
				ui.Column(c).Gap(unit(c, 1)).Children(func() {
					ui.Text(c, answerOne)
					ui.Text(c, answerTwo)
				})
			})
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleSystem, Time: "14:02",
			}, func() {
				ui.Text(c, "已于 14:02 把这条派给 Andre Thomson。")
			})
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleTool, Time: "14:03", Monospace: true, Muted: true,
			}, func() {
				ui.Column(c).Gap(unit(c, 0.5)).Children(func() {
					ui.Text(c, toolNote)
					ui.Text(c, toolOut)
				})
			})
		})

		// The right column holds everything that hangs off a turn rather than
		// being one: its head, the marks under it, the turn it answers, and
		// the editor that replaces it.
		ui.Column(c).Width(unit(c, 60)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.MessageHeader(c, chat.MessageHeaderOptions{
				Author: "claude opus", Time: "14:03", Role: chat.RoleAssistant,
				Badge: "edited",
			})
			chat.MessageHeader(c, chat.MessageHeaderOptions{
				Author: "我", Time: "14:02", Role: chat.RoleUser,
			})
			chat.MessageHeader(c, chat.MessageHeaderOptions{
				Time: "14:02", Role: chat.RoleSystem,
			})
			chat.MessageActions(c, []chat.MessageAction{
				{Label: "Copy", Icon: display.IconCopy},
				{Label: "Retry", Icon: display.IconRefresh},
				{Label: "Good answer", Icon: display.IconStar, Active: true},
				{Label: "Delete", Icon: display.IconTrash, Tone: core.Danger},
			})
			chat.QuoteReply(c, chat.QuoteReplyOptions{
				Author: "我", Text: askText, Lines: 2,
			})
			// The assistant's turn, not mine: a user bubble is the dark fill,
			// and the editor's field draws its text in the window's ordinary
			// ink, which on that fill is ink on ink.
			chat.MessageEditor(c, &edited, chat.MessageEditorOptions{
				Role: chat.RoleAssistant, Label: "Edit message", Lines: 3,
			})
		})
	})

	showcase.Field(c, "MessageList — 转写，DateSeparator 在里面跟着滚")
	chat.MessageList(c, chat.MessageListOptions{
		Messages: 6, Height: 420, Label: "Riverside Clinic · callbacks",
		Top: func() {
			chat.DateSeparator(c, "今天")
		},
		Message: func(i int) {
			switch i {
			case 0:
				chat.MessageBubble(c, chat.MessageBubbleOptions{
					Role: chat.RoleUser, Author: "我", Time: "14:02",
				}, func() { ui.Text(c, askText) })
			case 1:
				chat.MessageBubble(c, chat.MessageBubbleOptions{
					Role: chat.RoleAssistant, Author: "claude opus", Time: "14:03",
				}, func() {
					ui.Column(c).Gap(unit(c, 1)).Children(func() {
						ui.Text(c, answerOne)
						ui.Text(c, answerTwo)
					})
				})
			case 2:
				chat.MessageBubble(c, chat.MessageBubbleOptions{
					Role: chat.RoleSystem, Time: "14:02",
				}, func() { ui.Text(c, "已于 14:02 把这条派给 Andre Thomson。") })
			case 3:
				chat.MessageBubble(c, chat.MessageBubbleOptions{
					Role: chat.RoleUser, Author: "我", Time: "14:04", Selected: true,
				}, func() { ui.Text(c, "仓库有现货吗？") })
			case 4:
				chat.MessageBubble(c, chat.MessageBubbleOptions{
					Role: chat.RoleAssistant, Author: "claude opus", Time: "14:04",
				}, func() { ui.Text(c, "有一个 HC-2A5，明早带过去就行。") })
			default:
				chat.MessageBubble(c, chat.MessageBubbleOptions{
					Role: chat.RoleTool, Time: "14:05", Monospace: true, Muted: true,
				}, func() {
					ui.Column(c).Gap(unit(c, 0.5)).Children(func() {
						ui.Text(c, toolNote)
						ui.Text(c, toolOut)
					})
				})
			}
		},
		Bottom: func() {
			chat.DateSeparator(c, "昨天")
		},
	})
	chat.MessageList(c, chat.MessageListOptions{
		Messages: 0, Height: 210, Label: "no turns yet",
		Message: func(int) {},
		Empty: func() {
			chat.EmptyState(c, chat.EmptyStateOptions{
				Title: "还没有对话",
				Body:  "把一条工单拖进来，或者先问一句。",
			})
		},
	})
}

// ── what an answer is made of ──────────────────────────────────────────────

func chatBlocksSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "正文 · MarkdownView / CodeBlock / TableBlock / ImageGrid")

	showcase.Field(c, "MarkdownView — 一段真的答案，源码由调用方给")
	chatMarkdown(c, "### 结论\n\n接触器线圈烧了，**不是制冷剂**。线圈额定 220V，"+
		"实测 24V，读数见 `docs/design-system.md` §2.1。\n\n"+
		"1. 断电后量线圈两端\n"+
		"2. 换 HC-2A5\n"+
		"3. 通电观察压缩机是否起转\n\n"+
		"> 这台机上次换过压缩机，线圈还是原厂的。\n\n"+
		"---\n\n"+
		"```go\nfunc ohms(measured, rated float32) float32 {\n"+
		"\t// a coil under a tenth of its rating is a dead coil\n"+
		"\treturn measured / rated\n}\n```\n")

	showcase.Field(c, "CodeBlock / TableBlock / ImageGrid — 模型答案里嵌的三样东西")
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 108)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.CodeBlock(c, chat.CodeBlockOptions{
				Lang: "go", Title: "ohms.go", LineNumbers: true, MaxHeight: 120,
				Code: "func ohms(measured, rated float32) float32 {\n" +
					"\t// a coil under a tenth of its rating is a dead coil\n" +
					"\treturn measured / rated\n}\n",
			})
			chat.TableBlock(c, chat.TableBlockOptions{
				Caption: "工时与配件",
				Head:    []string{"part", "need", "hours"},
				Rows: [][]string{
					{"HC-2A5 接触器", "1", "0.9"},
					{"N 端 4mm² 线鼻", "4", "0.3"},
					{"氟利昂 R32", "—", "0.4"},
				},
				Aligns: []chat.Align{chat.AlignLeft, chat.AlignCenter, chat.AlignRight},
			})
		})
		ui.Column(c).Width(unit(c, 116)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.ImageGrid(c, chat.ImageGridOptions{
				Columns: 3, Size: 108, Label: "现场照片",
				Items: []chat.ImageItem{
					{Alt: "bench-wiring.svg", Draw: func() {
						chatPhoto(c, k.SurfaceHover, k.TextFaint, "接线端子")
					}},
					{Alt: "coil.svg", Draw: func() {
						chatPhoto(c, k.DangerBg.Alpha(0.5), k.Danger, "烧黑的线圈")
					}},
					{Alt: "plate.svg", Draw: func() {
						chatPhoto(c, k.AccentBg.Alpha(0.5), k.AccentText, "铭牌 HC-2A5")
					}},
				},
			})
			chat.AttachmentChip(c, chat.AttachmentChipOptions{
				Name: "bench-wiring.svg", Meta: "1.8 MB · SVG", Mark: "图",
				Removable: true, Openable: true,
			})
			chat.AttachmentChip(c, chat.AttachmentChipOptions{
				Name: "wiring-old.pdf", Meta: "装不上", Mark: "PDF", Tone: core.Warning,
			})
		})
	})
}

// chatPhoto is what a pasted or attached picture looks like when there is no
// file behind it: a tinted tile with a caption. A gallery that left the tiles
// blank would be showing the frame of a picture and nothing of the picture.
func chatPhoto(c *ui.Context, bg, ink ui.Color, caption string) {
	ui.Box(c).Fill().Background(bg).Center().Children(func() {
		ui.Text(c, caption).TextColor(ink).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// ── the answer arriving ────────────────────────────────────────────────────

func chatStreamSection(c *ui.Context) {
	showcase.Section(c, "流式 · StreamingText / ThinkingBlock / TypingIndicator / SendButton")

	showcase.Field(c, "StreamingText — 光标跟着流，ThinkingBlock 就在它上面")
	ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 146)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleAssistant, Author: "claude opus", Time: "14:04",
			}, func() {
				ui.Column(c).Gap(unit(c, 1)).Children(func() {
					chat.ThinkingBlock(c, chat.ThinkingBlockOptions{
						Summary: "正在想", Busy: true, Duration: "1.4s",
						Open:   true,
						Detail: "读了 ui/core/state.go，看它怎么算密度",
					})
					chat.StreamingText(c, &streaming, chat.StreamingTextOptions{
						Streaming: true, Lines: 3,
					})
				})
			})
			chat.ThinkingBlock(c, chat.ThinkingBlockOptions{
				Summary: "正在想", Duration: "2.1s",
				Detail: "算了三种口径：铭牌电流 6.2A，实测 0.8A",
			})
		})
		ui.Column(c).Width(unit(c, 84)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.TypingIndicator(c, chat.TypingIndicatorOptions{Who: "claude opus"})
			chat.TypingIndicator(c, chat.TypingIndicatorOptions{
				Who: "Andre Thomson", Bubbles: 3,
			})
			chat.ThinkingIndicator(c, chat.ThinkingIndicatorOptions{
				Label: "正在想", Detail: "读了 3 个文件",
			})
			chat.SearchProgress(c, chat.SearchProgressOptions{
				Query: "contact", Found: 7, Busy: true, Match: 3,
			})
			chat.SearchProgress(c, chat.SearchProgressOptions{
				Query: "接触器", Found: 2, Match: 1,
			})
			ui.Row(c).FillWidth().Gap(unit(c, 1.5)).AlignItems(ui.Center).Children(func() {
				chat.SendButton(c, chat.SendButtonOptions{Label: "Send"})
				chat.StopGeneratingButton(c, chat.StopGeneratingButtonOptions{
					Label: "Stop generating", Busy: true,
				})
			})
			ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
				chat.ScrollToBottomButton(c, chat.ScrollToBottomButtonOptions{
					Pending: 12, Label: "New messages",
				})
			})
			chat.ScrollToBottomButton(c, chat.ScrollToBottomButtonOptions{
				Pending: 0, Hidden: true, Label: "New messages",
			})
		})
	})
}

// ── the box you type in ────────────────────────────────────────────────────

func chatComposerSection(c *ui.Context) {
	showcase.Section(c, "输入 · PromptComposer / 菜单 / 选择器 / 参数")

	showcase.Field(c, "PromptComposer — 草稿、上下文、工具、发送，全部是调用方的指针")
	on := []string{"Read", "Grep"}
	ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Start).Children(func() {
		chat.PromptComposer(c, chat.PromptComposerOptions{
			Text: &chat_draft, Label: "PromptComposer", Lines: 3,
			Placeholder: "Ask about this callback",
			Chips: []chat.ContextChip{
				{Name: "Riverside Clinic", Detail: "CB-2871", Selected: true},
				{Name: "ui/chat", Detail: "14 files"},
			},
			Tools: []string{"Read", "Write", "Bash"}, ToolsOn: &on,
			Selected: &chat_picker, ShowVoice: true,
		})
		ui.Column(c).Width(unit(c, 70)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.SlashCommandMenu(c, chat.SlashCommandOptions{
				Commands: []string{"/compact", "/model", "/clear", "/export"},
				Query:    "/co", Selected: &slashSel, MaxHeight: 110,
			})
			chat.MentionMenu(c, chat.MentionOptions{
				Items: []string{"@andre", "@sam", "@dante", "@riverside"},
				Query: "@a", Selected: &mentionSel, MaxHeight: 90,
			})
			chat.ToolToggleMenu(c, chat.ToolToggleMenuOptions{
				Tools: []string{"Read", "Write", "Bash"}, On: &on, Label: "Tools",
			})
		})
	})

	showcase.Field(c, "ModelSelector / ModeSelector / VoiceInputButton / VoiceWaveform")
	listening := true
	levels := []float32{0.2, 0.55, 0.9, 0.4, 0.7, 0.3, 0.8, 0.5, 0.2, 0.6, 0.35, 0.75}
	// 88 + 68 + 68 plus two gaps of 8 is 940 of the page's 944. The row is
	// given three columns whose widths add up rather than a fourth that takes
	// what is left: a column that grows is the one whose contents end up
	// pushed against the edge of the window.
	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 88)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.ModelSelector(c, chat.ModelSelectorOptions{
				Selected: &model, Label: "Model",
				Models: []string{
					"claude-opus-4-1: Claude Opus 4.1",
					"claude-sonnet-4-5: Claude Sonnet 4.5",
					"gpt-5.1: GPT-5.1",
				},
			})
			chat.ModeSelector(c, chat.ModeSelectorOptions{
				Selected: &mode, Label: "Mode",
				Modes: []string{"Ask", "Plan", "Build"},
			})
		})
		ui.Column(c).Width(unit(c, 68)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.VoiceInputButton(c, chat.VoiceInputButtonOptions{
				Label: "Hold to talk", Listening: listening, Levels: levels,
			})
			chat.VoiceInputButton(c, chat.VoiceInputButtonOptions{Label: "Hold to talk"})
		})
		ui.Column(c).Width(unit(c, 68)).Shrink(0).Gap(unit(c, 2)).Children(func() {
			chat.VoiceWaveform(c, chat.VoiceWaveformOptions{
				Levels: levels, Bars: 12, Height: 34,
				Label: "voice note", Color: core.Tokens(c).Accent,
			})
			chat.VoiceWaveform(c, chat.VoiceWaveformOptions{
				Levels: levels[:4], Bars: 24, Height: 34, Label: "quiet",
			})
		})
	})

	showcase.Field(c, "PasteImagePreview / SuggestionChips / PromptLibrary")
	// SuggestionChips is one row that does not wrap, so it is given the page's
	// own width rather than a column too narrow for it: three chips in a
	// 280-point column is two chips and one off the edge of the window.
	chat.SuggestionChips(c, chat.SuggestionChipsOptions{
		Title: "Suggestions", Max: 3,
		Suggestions: []string{
			"返修: 查一下同型号的返修记录",
			"工时: 把这次上门估成工时",
			"通知: 给客户写一条明早八点到场的短信",
		},
	})
	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 76)).Shrink(0).Children(func() {
			chat.PasteImagePreview(c, chat.PasteImagePreviewOptions{
				Images: []chat.ImageItem{
					{Alt: "bench-wiring.svg"},
					{Alt: "coil.svg"},
					{Alt: "plate.svg"},
				},
			})
		})
		chat.PromptLibrary(c, chat.PromptLibraryOptions{
			Title: "PromptLibrary", Height: 220, Selected: &librarySel,
			Prompts: []string{
				"根因: 总结这条工单的根因，并给出换件清单",
				"报警: 把这台机过去 90 天的报警代码列出来",
				"通知: 给客户写一条明早八点到场的通知",
			},
		})
	})

	showcase.Field(c, "SystemPromptEditor 与 ParameterPanel — 模型这一侧给的东西")
	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 122)).Shrink(0).Children(func() {
			chat.SystemPromptEditor(c, chat.SystemPromptEditorOptions{
				Prompt: &system, Label: "system prompt", Title: "System prompt",
				Original: "你是 callbacks 的维修助手。", Lines: 4,
			})
		})
		ui.Column(c).Width(unit(c, 110)).Shrink(0).Children(func() {
			chat.ParameterPanel(c, chat.ParameterPanelOptions{
				Title: "Parameters", ResetAll: "Reset all",
				Parameters: []chat.Parameter{
					{Name: "temperature", Hint: "0.0 - 1.0", Control: func() {
						input.Slider(c, &v, input.SliderOptions{
							Label: "temperature", Min: 0, Max: 1, Step: 0.1,
						})
					}},
					{Name: "max_tokens", Hint: "4096", Control: func() {
						ui.Text(c, "4096").TextColor(core.Tokens(c).Text).
							FontSize(core.FontSize(c, theme.RowSize))
					}},
					{Name: "tools", Hint: "Read / Grep / Bash", Control: func() {
						on := []string{"Read", "Grep"}
						chat.ToolToggleMenu(c, chat.ToolToggleMenuOptions{
							Tools: []string{"Read", "Grep", "Bash"}, On: &on, Label: "tools",
						})
					}},
				},
			})
		})
	})
}

// ── the other conversations ────────────────────────────────────────────────

// The conversation names the sidebar is drawn with. Fixed, like everything
// else here, so the page looks the same on every run.
var chatConvNames = []struct{ title, preview, when string }{
	{"Riverside Clinic", "接触器线圈烧了，明早带 HC-2A5", "4m"},
	{"Maple Street Bakery", "两台柜机都报 F02，先查冷媒压力", "1h"},
	{"Corner Studio", "排期挪到周四下午，工程师有空", "3h"},
	{"Northgate Dental", "问过保修范围，等厂家回话", "昨天"},
	{"Lakeside Vet", "新装两台，要先确认电源是三相的", "3 天前"},
}

func chatConversationSection(c *ui.Context) {
	showcase.Section(c, "会话 · ConversationContainer / ConversationList / BranchNavigator")

	showcase.Field(c, "ConversationContainer — 侧栏 322，转写拿剩下的")
	ui.Box(c).Width(unit(c, 236)).Height(360).Shrink(0).Children(func() {
		chat.ConversationContainer(c, &collapsed, chat.ConversationContainerOptions{
			SidebarLabel: "Conversations",
		}, func() {
			chat.ConversationSearch(c, chat.ConversationSearchOptions{
				Query: &query, Placeholder: "Search conversations",
				Found: 2, Total: 5,
			})
			chat.ConversationList(c, chat.ConversationListOptions{
				Conversations: len(chatConvNames), Height: 290,
				Selected: &sel, Choice: &data.Selectable{}, State: &state,
				Conversation: func(row int) {
					n := chatConvNames[row]
					chat.ConversationItem(c, chat.ConversationItemOptions{
						Title: n.title, Preview: n.preview, When: n.when,
						Unread: row, Pinned: row == 0, Selected: row == sel,
						Icon: "callbacks", Tinted: row == 0,
					})
				},
			})
		}, func() {
			chat.MessageList(c, chat.MessageListOptions{
				Messages: 3, Height: 360, Label: "Riverside Clinic · callbacks",
				Message: func(i int) {
					switch i {
					case 0:
						chat.MessageBubble(c, chat.MessageBubbleOptions{
							Role: chat.RoleUser, Author: "我", Time: "14:02",
						}, func() { ui.Text(c, askText) })
					case 1:
						chat.MessageBubble(c, chat.MessageBubbleOptions{
							Role: chat.RoleAssistant, Author: "claude opus", Time: "14:03",
						}, func() { ui.Text(c, answerOne) })
					default:
						chat.MessageBubble(c, chat.MessageBubbleOptions{
							Role: chat.RoleTool, Time: "14:03", Monospace: true, Muted: true,
						}, func() {
							ui.Column(c).Gap(unit(c, 0.5)).Children(func() {
								ui.Text(c, toolNote)
								ui.Text(c, toolOut)
							})
						})
					}
				},
			})
		})
	})

	showcase.Field(c, "ConversationItem / ProjectList / ConversationExport / 分支与重问")
	ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 68)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			for i, n := range chatConvNames {
				chat.ConversationItem(c, chat.ConversationItemOptions{
					Title: n.title, Preview: n.preview, When: n.when,
					Unread: i, Pinned: i == 0, Selected: i == 0,
					Icon: "customers", Tinted: i < 2,
				})
			}
		})
		ui.Column(c).Width(unit(c, 60)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.ProjectList(c, chat.ProjectListOptions{
				Projects: []string{"callbacks", "ui", "internal"}, Selected: &chatProject,
				Height: 120, Query: "cal",
			})
			chat.ProjectList(c, chat.ProjectListOptions{
				Projects: []string{"callbacks", "ui", "internal"}, Selected: &chatProject,
				Height: 70, Query: "zzz",
				Empty: func() {
					chat.EmptyState(c, chat.EmptyStateOptions{
						Title: "没有匹配的项目", Body: "换个关键词，或者新建一个。",
					})
				},
			})
			chat.ConversationExport(c, chat.ConversationExportOptions{
				Button: "Export as Markdown", Format: chat.ExportMarkdown,
				Transcript: chat.Transcript{Title: "Riverside Clinic", Turns: chatTurns()},
			})
			chat.ConversationExport(c, chat.ConversationExportOptions{
				Button: "Export as JSON", Format: chat.ExportJSON,
				Transcript: chat.Transcript{Title: "Riverside Clinic", Turns: chatTurns()},
			})
		})
		ui.Column(c).Width(unit(c, 62)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.BranchNavigator(c, chat.BranchNavigatorOptions{Branch: 2, Of: 3})
			chat.RegenerateMenu(c, chat.RegenerateOptions{
				Label: "asked the same thing again",
				Items: []string{"换个说法", "更简短", "带上工时", "只给结论"},
			})
			chat.DateSeparatorWhen(c, strPtr("2026-10-07"), strPtr("2026-10-07"))
			chat.DateSeparatorWhen(c, strPtr("2026-10-07"), strPtr("2026-10-06"))
			chat.DateSeparatorWhen(c, strPtr("2026-10-07"), strPtr("2026-10-07"))
		})
	})
}

func chatTurns() []chat.Turn {
	return []chat.Turn{
		{Role: chat.RoleUser, Who: "我", When: "14:02", Text: askText},
		{Role: chat.RoleAssistant, Who: "claude opus", When: "14:03",
			Text: answerOne + answerTwo, Quote: askText},
		{Role: chat.RoleTool, Who: "Grep", When: "14:03", Text: toolOut},
	}
}

func strPtr(s string) *string { return &s }

// hexOf writes a colour the way a designer reads it, which is the only reason
// the pure-function section can show what TokenColor returned at all: a
// function that returns a colour cannot be checked by printing its name.
func hexOf(col ui.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", col.R, col.G, col.B)
}

// ── where the facts came from ──────────────────────────────────────────────

func chatSourceSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "来源与用量 · CitationBadge / SourcesPanel / TokenCounter")

	showcase.Field(c, "SourcesPanel 与 SourceCard — 一句结论背后的三份文件")
	// 128 + 56 + 50 plus two gaps of 8 is 928: the widths add up on purpose, so
	// that the third column's right-hand edge is a number rather than
	// "whatever is left" — which is how the usage column ended up off the page.
	ui.Row(c).Width(unit(c, 232)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 128)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.SourcesPanel(c, chat.SourcesPanelOptions{
				Title: "SourcesPanel", Height: 230, Selected: 1,
				Quotation: "接触器线圈额定 220V，实测 24V。",
				Sources: []chat.Source{
					{Title: "docs/design-system.md",
						Snippet: "接触器线圈额定 220V，通电后应为满压。",
						Where:   "docs/design-system.md §2.1", Score: 0.92, Icon: "reports"},
					{Title: "ui/core/state.go",
						Snippet: "func Density(c *ui.Context) theme.Density",
						Where:   "ui/core/state.go:41", Score: 0.71, Icon: "inventory"},
					{Title: "internal/store/service.go",
						Snippet: "coilOhms := measured / rated",
						Where:   "internal/store/service.go:212", Score: 0.48, Icon: "inventory"},
				},
			})
			chat.SourcesPanel(c, chat.SourcesPanelOptions{
				Title: "SourcesPanel", Height: 90,
				Empty: func() {
					ui.Text(c, "这次回答没有查任何文件。").TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				},
			})
		})
		ui.Column(c).Width(unit(c, 56)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.SourceCard(c, chat.SourceCardOptions{
				N: 1, Title: "docs/design-system.md", Where: "docs/design-system.md §2.1",
				Snippet: "接触器线圈额定 220V，通电后应为满压。", Icon: "reports",
				Quotation: "线圈额定 220V", Selected: true, Lines: 2,
			})
			chat.SourceCard(c, chat.SourceCardOptions{
				N: 2, Title: "ui/core/state.go", Where: "ui/core/state.go:41",
				Snippet: "func Density(c *ui.Context) theme.Density", Icon: "inventory", Lines: 2,
			})
			ui.Row(c).FillWidth().Gap(unit(c, 1)).AlignItems(ui.Center).Children(func() {
				chat.CitationBadge(c, chat.CitationBadgeOptions{N: 1, Count: 3})
				chat.CitationBadge(c, chat.CitationBadgeOptions{N: 2, Count: 3})
				chat.CitationBadge(c, chat.CitationBadgeOptions{N: 3, Count: 3, Selected: true})
			})
		})
		ui.Column(c).Width(unit(c, 50)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			// One chip: ContextChips is a row of pills that does not wrap, and
			// three of them are 300 points wide in a 200-point column.
			chat.ContextChips(c, []chat.ContextChip{
				{Name: "Riverside Clinic", Detail: "CB-2871", Selected: true},
			})
			chat.ContextChips(c, []chat.ContextChip{
				{Name: "ui/chat", Detail: "14 files", Tone: core.Accent},
			})
			chat.ContextWindowMeter(c, chat.ContextWindowMeterOptions{
				Used: 58240, Window: 200000, WarnAt: 0.6, DangerAt: 0.85,
				Label: "context",
			})
			chat.ContextWindowMeter(c, chat.ContextWindowMeterOptions{
				Used: 186000, Window: 200000, WarnAt: 0.6, DangerAt: 0.85,
			})
			chat.TokenCounter(c, chat.TokenCounterOptions{
				Tokens: 58240, Cost: 0.41, ShowCost: true, Label: "58.2k",
			})
			chat.TokenCounter(c, chat.TokenCounterOptions{Tokens: 812})
			chat.CostEstimator(c, chat.CostEstimatorOptions{
				InTokens: 41200, OutTokens: 17040,
				InRate: 15, OutRate: 75, Model: "claude-opus-4-1", Breakdown: true,
			})
		})
	})
}

// ── what else a turn can carry ─────────────────────────────────────────────

func chatPayloadSection(c *ui.Context) {
	showcase.Section(c, "载荷 · AudioMessage / ErrorMessage / FeedbackForm / ArtifactCard")

	showcase.Field(c, "一句话、音频、文件、错误 —— 一条 turn 能带的四种东西")
	wave := []float32{0.2, 0.55, 0.9, 0.4, 0.7, 0.3, 0.8, 0.5, 0.2, 0.6, 0.35, 0.75,
		0.45, 0.85, 0.25, 0.55}
	rating := 4
	// 76 + 78 + 76 plus two gaps of 8 is 944, the page's own content width.
	// The widths are written out because a column that grows is a column whose
	// right-hand contents end up outside the window.
	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 76)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleUser, Author: "我", Time: "14:06",
			}, func() {
				// No transcript here on purpose: a transcript is drawn in the
				// muted ink, and a user bubble's ink is the dark one, so the
				// words would be dark grey on near-black.
				chat.AudioMessage(c, chat.AudioMessageOptions{
					Title: "voice note", Seconds: 23.4, Playing: true, Wave: wave,
				})
			})
			chat.MessageBubble(c, chat.MessageBubbleOptions{
				Role: chat.RoleUser, Author: "我", Time: "14:06",
			}, func() {
				chat.FileMessage(c, chat.FileMessageOptions{
					Role: chat.RoleUser, Author: "我", Time: "14:06",
					Attachment: chat.AttachmentChipOptions{
						Name: "bench-wiring.svg", Meta: "1.8 MB", Mark: "SVG",
					},
					Caption:  "现场接线，线圈那端已经黑了",
					Selected: true,
				})
			})
		})
		ui.Column(c).Width(unit(c, 78)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.ErrorMessage(c, chat.ErrorMessageOptions{
				Title: "The model refused to answer", Severity: core.Danger,
				Detail: "This account has no access to that machine's history. " +
					"The rest of the transcript is unaffected.",
				Retry: "Retry", Dismiss: "Dismiss",
			})
			chat.ErrorMessage(c, chat.ErrorMessageOptions{
				Title: "The tool call timed out", Severity: core.Warning,
				Detail: "Grep did not answer in 30s.",
				Retry:  "Retry", Dismiss: "Dismiss",
			})
			chat.FeedbackForm(c, chat.FeedbackFormOptions{
				Rating: rating, Comment: &comment, CommentLabel: "哪里不对？",
				Title: "Feedback", Submit: "Report", Dismiss: "Dismiss",
			})
		})
		ui.Column(c).Width(unit(c, 76)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.ArtifactCard(c, chat.ArtifactCardOptions{
				Title: "report.html", Kind: "HTML", Ready: true, Selected: true,
				Body: "换件清单与工时估算，12 kB",
				Open: "Open", Copy: "Copy",
			})
			chat.ArtifactCard(c, chat.ArtifactCardOptions{
				Title: "wiring.svg", Kind: "SVG",
				Body: "还在画接线图", Open: "Open",
			})
			chat.CapabilityCards(c, chat.CapabilityCardsOptions{
				Title: "Capabilities", Columns: 2,
				Capabilities: []chat.Capability{
					{Title: "查返修记录", Body: "同型号 90 天内的工单", Icon: "search"},
					{Title: "估工时", Body: "按站点和距离算", Icon: "clock"},
					{Title: "写客户通知", Body: "短信或邮件两种语气", Icon: "bell"},
				},
			})
		})
	})

	showcase.Field(c, "ChatMessage — 上面几样东西按 kind 拼成一条 turn")
	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 58)).Shrink(0).Children(func() {
			chat.ChatMessage(c, chat.ChatMessageOptions{
				Kind: chat.KindText, Role: chat.RoleUser, Author: "我", Time: "14:02",
				Text: askText,
			})
		})
		ui.Column(c).Width(unit(c, 92)).Shrink(0).Children(func() {
			chat.ChatMessage(c, chat.ChatMessageOptions{
				Kind: chat.KindCode, Role: chat.RoleAssistant,
				Author: "claude opus", Time: "14:03", Quoted: askText,
				Code: "func ohms(measured, rated float32) float32 {\n\treturn measured / rated\n}\n",
				Lang: "go", CodeTitle: "ohms.go", LineNumbers: true, MaxCodeHeight: 110,
				Streaming: true,
				Actions: []chat.MessageAction{
					{Label: "Copy", Icon: display.IconCopy},
					{Label: "Retry", Icon: display.IconRefresh},
				},
			})
		})
		ui.Column(c).Width(unit(c, 74)).Shrink(0).Children(func() {
			chat.ChatMessage(c, chat.ChatMessageOptions{
				Kind: chat.KindError, Role: chat.RoleAssistant,
				Author: "claude opus", Time: "14:05",
				Err: chat.ErrorMessageOptions{
					Title: "The model refused to answer", Severity: core.Danger,
					Detail: "No access to that machine's history.",
					Retry:  "Retry", Dismiss: "Dismiss",
				},
				Audio: chat.AudioMessageOptions{
					Title: "voice note", Seconds: 23.4, Wave: wave,
				},
				File: chat.FileMessageOptions{
					Attachment: chat.AttachmentChipOptions{
						Name: "bench-wiring.svg", Meta: "1.8 MB", Mark: "SVG",
					},
				},
				Thinking: chat.ThinkingBlockOptions{Summary: "正在想", Detail: "读了 3 个文件"},
			})
		})
	})

	showcase.Field(c, "EmptyState 与 DragDropOverlay — 什么都没有的那一屏，以及拖进来的时候")
	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 2)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(unit(c, 92)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			ui.Column(c).Width(unit(c, 72)).Shrink(0).Children(func() {
				host := ui.Box(c).FillWidth().Height(unit(c, 34)).Shrink(0).
					Radius(theme.CardRadius).Background(core.Tokens(c).Surface).
					Border(theme.BorderWidth, core.Tokens(c).Border).Label("drop host")
				host.Children(func() {
					chat.DragDropOverlay(c, chat.DragDropOverlayOptions{Host: host})
					ui.Text(c, "Drop files to attach").TextColor(core.Tokens(c).TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
			ui.Text(c, "DragDropOverlay — 只在 Host 悬停时出现，静止时它是空的").
				TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Column(c).Width(unit(c, 140)).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			chat.EmptyState(c, chat.EmptyStateOptions{
				Title: "还没有对话",
				Body:  "把一条工单拖进来，或者从下面挑一个开始。",
			})
			chat.EmptyState(c, chat.EmptyStateOptions{
				Title:       "Empty",
				Body:        "没有建议：这段对话太短，还不知道该推荐什么。",
				Suggestions: []string{},
			})
		})
	})
}

// ── the half that needs no window ──────────────────────────────────────────

func chatPureSection(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "纯函数 · 不需要窗口就能算的那几个")

	// A call and its result, in a column 76 units wide: the call names itself
	// and the result wraps under it rather than being given a width of its own
	// and running off the right-hand edge of the column.
	line := func(call, result string) {
		ui.Column(c).FillWidth().Gap(unit(c, 0.25)).Children(func() {
			ui.Text(c, call).TextColor(k.Text).FillWidth().
				FontSize(core.FontSize(c, theme.RowSize))
			ui.Text(c, result).TextColor(k.TextMuted).FillWidth().
				FontSize(core.FontSize(c, theme.RowSize))
		})
	}

	blocks := chat.Parse("### 结论\n\n线圈烧了。\n\n" +
		"- 换接触器\n- 复测\n\n```go\nreturn 1\n```\n\n" +
		"| part | need |\n|---|---|\n| 线圈 | 1 |\n")
	kinds := ""
	for i, b := range blocks {
		if i > 0 {
			kinds += "   "
		}
		kinds += string(b.Kind)
	}
	aligns := ""
	for i, a := range []chat.Align{
		chat.AlignNone, chat.AlignLeft, chat.AlignCenter, chat.AlignRight,
	} {
		if i > 0 {
			aligns += "   "
		}
		aligns += a.String()
	}
	roles := ""
	for i, s := range []string{"user", "assistant", "system", "tool", "nonsense"} {
		if i > 0 {
			roles += "   "
		}
		roles += s + " → " + chat.RoleOf(s).String()
	}
	tokens := chat.Highlight("if x > 1 { return \"hi\" }", "go")
	kinds2 := ""
	for i, t := range tokens {
		if i > 0 {
			kinds2 += " "
		}
		kinds2 += t.Kind.String()
	}
	ink, _ := chat.TokenColor(chat.TokKeyword, false)
	dark, _ := chat.TokenColor(chat.TokKeyword, true)

	ui.Row(c).Width(unit(c, 236)).Shrink(0).Gap(unit(c, 4)).
		AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("Parse(src)", kinds)
			line("BlockKind", string(blocks[0].Kind)+" … "+string(blocks[len(blocks)-1].Kind)+
				"   共 "+itoaPage(len(blocks))+" 块")
			line("Align", aligns)
			line("RoleOf", roles)
		})
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("Highlight(src, \"go\")", kinds2)
			line("TokenText(tokens)", chat.TokenText(tokens))
			line("TokenColor(TokKeyword)", "浅 "+hexOf(ink)+"   深 "+hexOf(dark))
			line("FormatTokens(58240)", chat.FormatTokens(58240))
		})
		ui.Column(c).WidthPercent(share(3)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			line("FormatCost(0.4117)", chat.FormatCost(0.4117))
			line("FormatCost(12.4)", chat.FormatCost(12.4))
			line("FormatDuration(23.4)", chat.FormatDuration(23.4))
			line("FormatDuration(0.9)", chat.FormatDuration(0.9)+
				"   FormatDuration(3725) "+chat.FormatDuration(3725))
		})
	})
}

// chatMarkdown draws the same answer twice: once parsed from a string, once
// from blocks the caller already has. A page that only showed Parse's output
// would be showing the parser and not the view.
func chatMarkdown(c *ui.Context, src string) {
	ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(unit(c, 118)).Shrink(0).Children(func() {
			chat.MarkdownView(c, chat.MarkdownViewOptions{Source: src})
		})
		ui.Column(c).Width(unit(c, 118)).Shrink(0).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "MarkdownView 直接吃 Parse 出来的块").TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			chat.MarkdownView(c, chat.MarkdownViewOptions{
				Blocks: chat.Parse(src), Caption: "parsed blocks", Tight: true,
			})
		})
	})
}
