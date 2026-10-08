# 组件清单与进度

本项目 `ui/` 要覆盖的组件全集 —— 范围对齐组件库，但**全部用 Callbacks 的设计规范实现**：
`ui/theme` 的配色与密度、`ui/core` 的每帧一次 `Use`，以及「组件不持有状态、
给不了的东西就 panic」这两条约定。

范围来自对组件库的盘点（`component-catalogue.baseline.json`），作为**基准**而非代码来源
—— 一行库代码都没有搬过来。

`[x]` 已实现并有测试覆盖。`_新增_` 是本项目多出来的构件，写别的组件时真正需要的。

## ui/theme — 配色与尺寸　0/0


## ui/core — 每帧解析一次　13/13

- [x] `ControlHeight` _新增_
- [x] `Density` _新增_
- [x] `FontSize` _新增_
- [x] `IsDark` _新增_
- [x] `Messages` _新增_
- [x] `Motion` _新增_
- [x] `Msg` _新增_
- [x] `Reduced` _新增_
- [x] `Tokens` _新增_
- [x] `Use` _新增_
- [x] `WithMessages` _新增_
- [x] `WithReducedMotion` _新增_
- [x] `WithTextScale` _新增_

## ui/layout — 容器与窗口骨架　16/16

- [x] `AppShell`
- [x] `AspectRatio`
- [x] `Container`
- [x] `Divider`
- [x] `Grid`
- [x] `GridCell`
- [x] `ScrollArea`
- [x] `SplitPane`
- [x] `Stack`
- [x] `StatusBar`
- [x] `TitleBar`
- [x] `AvatarSlots` _新增_
- [x] `Board` _新增_
- [x] `PageHeader` _新增_
- [x] `Panel` _新增_
- [x] `TrafficLights` _新增_

## ui/display — 记号与文字　15/15

- [x] `Avatar`
- [x] `AvatarGroup`
- [x] `Badge`
- [x] `EditableText`
- [x] `Heading`
- [x] `Icon`
- [x] `Image`
- [x] `Kbd`
- [x] `Link`
- [x] `Tag`
- [x] `Text`
- [x] `AvatarCluster` _新增_
- [x] `IconClose` _新增_
- [x] `Meter` _新增_
- [x] `PresenceDot` _新增_

## ui/input — 人操作的控件　50/61

- [x] `Banner`
- [x] `BulkActionBar`
- [x] `Button`
- [x] `ButtonGroup`
- [x] `Cascader`
- [x] `Checkbox`
- [x] `CheckboxGroup`
- [x] `ColorPalette`
- [x] `ColorPicker`
- [x] `Combobox`
- [x] `CopyButton`
- [x] `CurrencyInput`
- [x] `FileDropZone`
- [x] `FilePicker`
- [x] `Form`
- [x] `FormField`
- [x] `HoldToConfirm`
- [x] `IconButton`
- [x] `InputGroup`
- [x] `Knob`
- [x] `MaskedInput`
- [x] `Masonry`
- [x] `MentionInput`
- [x] `NumberInput`
- [x] `PasswordInput`
- [x] `PinInput`
- [x] `QRCode`
- [x] `RangeSlider`
- [x] `Rating`
- [x] `ResizablePanelGroup`
- [x] `ScrubInput`
- [x] `SearchInput`
- [x] `SelectSearch`
- [x] `SignaturePad`
- [x] `Slider`
- [x] `Switch`
- [x] `TagsInput`
- [x] `TextArea`
- [x] `TextInput`
- [x] `Transfer`
- [x] `TreeSelect`
- [x] `VerticalSlider`
- [x] `ChoiceChips` _新增_
- [x] `Dock` _新增_
- [x] `MultiSelect` _新增_
- [x] `RadioGroup` _新增_
- [x] `SearchField` _新增_
- [x] `Segmented` _新增_
- [x] `Select` _新增_
- [x] `ToggleGroup` _新增_
- [ ] `DockPanel`, `MultiSelectCheck`, `NumberInputButton`, `SelectCheck`, `SelectField`, `SelectNote`, `SelectPanel`, `SelectRow`, `SliderControl`, `TextInputEditor`, `Toggle`

## ui/navigation — 在窗口里移动　15/17

- [x] `Breadcrumb`
- [x] `CommandPalette`
- [x] `Menubar`
- [x] `NavigationMenu`
- [x] `Pagination`
- [x] `Sidebar`
- [x] `Steps`
- [x] `Tabs`
- [x] `Toolbar`
- [x] `ToolbarButton`
- [x] `ToolbarGroup`
- [x] `ToolbarSeparator`
- [x] `FilterRow` _新增_
- [x] `Group` _新增_
- [x] `Rail` _新增_
- [ ] `SegmentedControl`, `TabsStyle`

## ui/data — 记录与集合　10/14

- [x] `Accordion`
- [x] `Card`
- [x] `Carousel`
- [x] `Collapsible`
- [x] `DescriptionList`
- [x] `ImageViewer`
- [x] `Timeline`
- [x] `DataTable` _新增_
- [x] `List` _新增_
- [x] `StatLine` _新增_
- [ ] `DataError`, `DataPlaceholder`, `DataSelectTheme`, `Statistic`

## ui/overlay — 浮层　11/17

- [x] `AlertDialog`
- [x] `ContextMenu`
- [x] `Dialog`
- [x] `Drawer`
- [x] `DropdownMenu`
- [x] `HoverCard`
- [x] `Popconfirm`
- [x] `Popover`
- [x] `SplitButton`
- [x] `Tooltip`
- [x] `Panel` _新增_
- [ ] `DialogPanel`, `Overlay`, `OverlayAnchored`, `PopoverRoom`, `PopoverWide`, `PopupMenuOpen`

## ui/feedback — 说而不是展示　25/36

- [x] `ActivityFeed`
- [x] `Alert`
- [x] `BarsLoader`
- [x] `BlinkHighlight`
- [x] `CountUp`
- [x] `DotsLoader`
- [x] `InterruptButton`
- [x] `LayoutTransition`
- [x] `LoadingOverlay`
- [x] `NotificationCenter`
- [x] `OrbitLoader`
- [x] `Presence`
- [x] `Progress`
- [x] `PulseLoader`
- [x] `Result`
- [x] `Shimmer`
- [x] `ShimmerText`
- [x] `Skeleton`
- [x] `StatusIndicator`
- [x] `Typewriter`
- [x] `UndoToast`
- [x] `WaveLoader`
- [x] `Empty` _新增_
- [x] `NewStagger` _新增_
- [x] `Toast` _新增_
- [ ] `CostBreakdown`, `DismissToast`, `EmptyState`, `EvalResultTable`, `Meter`, `ShowToast`, `Stagger`, `Toaster`, `TokenUsageChart`, `ToolRegistryPanel`, `TraceViewer`

## ui/datetime — 日期与时间　31/39

- [x] `AgendaView`
- [x] `AttendeeList`
- [x] `AvailabilityPicker`
- [x] `Calendar`
- [x] `CalendarDayView`
- [x] `CalendarMonthView`
- [x] `CalendarWeekView`
- [x] `Countdown`
- [x] `CronEditor`
- [x] `CurrentTimeIndicator`
- [x] `DatePicker`
- [x] `DateRangePicker`
- [x] `DateTimePicker`
- [x] `DurationPicker`
- [x] `EventChip`
- [x] `EventEditor`
- [x] `EventPopover`
- [x] `MonthPicker`
- [x] `RecurrenceEditor`
- [x] `RelativeTime`
- [x] `ReminderPicker`
- [x] `Stopwatch`
- [x] `TimePicker`
- [x] `TimeRangePicker`
- [x] `TimezoneSelect`
- [x] `WeekPicker`
- [x] `YearPicker`
- [x] `CalendarEvents` _新增_
- [x] `CalendarTimeGrid` _新增_
- [x] `CalendarYearView` _新增_
- [x] `Grid` _新增_
- [ ] `CalendarFirstWeekday`, `CountdownText`, `DurationText`, `RecurrenceSummary`, `RelativeTimeText`, `ReminderText`, `WeekViewTitle`, `YearView`

## ui/chat — 对话　58/65

- [x] `ArtifactCard`
- [x] `AttachmentChip`
- [x] `AudioMessage`
- [x] `BranchNavigator`
- [x] `CapabilityCards`
- [x] `CitationBadge`
- [x] `CodeBlock`
- [x] `ContextChips`
- [x] `ContextWindowMeter`
- [x] `ConversationExport`
- [x] `ConversationItem`
- [x] `ConversationList`
- [x] `ConversationSearch`
- [x] `CostEstimator`
- [x] `DateSeparator`
- [x] `DateSeparatorWhen`
- [x] `DragDropOverlay`
- [x] `ErrorMessage`
- [x] `FeedbackForm`
- [x] `FileMessage`
- [x] `ImageGrid`
- [x] `MarkdownView`
- [x] `MessageActions`
- [x] `MessageBubble`
- [x] `MessageEditor`
- [x] `MessageHeader`
- [x] `MessageList`
- [x] `ModeSelector`
- [x] `ModelSelector`
- [x] `ParameterPanel`
- [x] `PasteImagePreview`
- [x] `ProjectList`
- [x] `PromptComposer`
- [x] `PromptLibrary`
- [x] `QuoteReply`
- [x] `RegenerateMenu`
- [x] `ScrollToBottomButton`
- [x] `SearchProgress`
- [x] `SendButton`
- [x] `SlashCommandMenu`
- [x] `SourceCard`
- [x] `SourcesPanel`
- [x] `StopGeneratingButton`
- [x] `StreamingText`
- [x] `SuggestionChips`
- [x] `SystemPromptEditor`
- [x] `TableBlock`
- [x] `ThinkingBlock`
- [x] `ThinkingIndicator`
- [x] `TokenCounter`
- [x] `ToolToggleMenu`
- [x] `TypingIndicator`
- [x] `VoiceInputButton`
- [x] `VoiceWaveform`
- [x] `ChatMessage` _新增_
- [x] `ConversationContainer` _新增_
- [x] `EmptyState` _新增_
- [x] `MentionMenu` _新增_
- [ ] `ChatContainer`, `ChatEmptyState`, `ContextMentionMenu`, `LinkPreviewCard`, `ProjectKnowledgePanel`, `SharedConversationView`, `WelcomeScreen`

## ui/code — 代码与终端　32/37

- [x] `BreakpointList`
- [x] `CallStack`
- [x] `CodeBadge`
- [x] `CodeCaption`
- [x] `CodeHeader`
- [x] `CodeMono`
- [x] `CodePanel`
- [x] `CodeSelectMark`
- [x] `CodeViewer`
- [x] `CommandHistory`
- [x] `CompletionMenu`
- [x] `DebugToolbar`
- [x] `FindWidget`
- [x] `Flamegraph`
- [x] `LogViewerName`
- [x] `Minimap`
- [x] `OutputPanel`
- [x] `ProblemsPanel`
- [x] `ProcessList`
- [x] `SearchPanel`
- [x] `SymbolOutline`
- [x] `Terminal`
- [x] `TerminalSearch`
- [x] `TerminalTabs`
- [x] `VariablesPanel`
- [x] `CodeMatchTokens` _新增_
- [x] `CodeMetrics` _新增_
- [x] `Diff` _新增_
- [x] `HexView` _新增_
- [x] `LineHeight` _新增_
- [x] `LogView` _新增_
- [x] `TokenInk` _新增_
- [ ] `CodeAnsiSpans`, `CodeMatchSpans`, `CodeRows`, `HexViewer`, `LogViewer`

## ui/git — 版本控制　20/20

- [x] `AnsiText`
- [x] `BlameView`
- [x] `BranchList`
- [x] `BranchSelector`
- [x] `ChangesList`
- [x] `CommandBlock`
- [x] `CommitGraph`
- [x] `CommitInput`
- [x] `CommitList`
- [x] `ConflictResolver`
- [x] `DiffStat`
- [x] `DiffViewer`
- [x] `FileHistory`
- [x] `GitStatusBadge`
- [x] `InlineDiff`
- [x] `PullRequestCard`
- [x] `ReviewComment`
- [x] `StashList`
- [x] `TagList`
- [x] `ThreeWayMerge`

## ui/files — 文件　18/18

- [x] `ArchiveViewer`
- [x] `FileExplorer`
- [x] `FileGrid`
- [x] `FileIcon`
- [x] `FileOperationProgress`
- [x] `FilePreview`
- [x] `LiveIndicator`
- [x] `PathBar`
- [x] `PermissionSelect`
- [x] `PresenceAvatars`
- [x] `Reactions`
- [x] `RecentFiles`
- [x] `RemoteCursor`
- [x] `RenameInline`
- [x] `ShareDialog`
- [x] `StorageUsage`
- [x] `TransferQueue`
- [x] `Permissions` _新增_

## ui/chart — 图表　54/86

- [x] `AreaChart`
- [x] `BarChart`
- [x] `BoxPlot`
- [x] `BulletChart`
- [x] `CalendarHeatmap`
- [x] `CandleCountdown`
- [x] `CandlestickChart`
- [x] `ChordDiagram`
- [x] `FunnelChart`
- [x] `Gauge`
- [x] `HeatmapChart`
- [x] `Histogram`
- [x] `LineChart`
- [x] `NetworkGraph`
- [x] `ParallelCoordinates`
- [x] `PieChart`
- [x] `RadarChart`
- [x] `SankeyChart`
- [x] `ScatterChart`
- [x] `SparkBar`
- [x] `Sparkline`
- [x] `Treemap`
- [x] `ViolinPlot`
- [x] `WaterfallChart`
- [x] `AlluvialChart` _新增_
- [x] `Annotation` _新增_
- [x] `AreaMountain` _新增_
- [x] `Axis` _新增_
- [x] `Brush` _新增_
- [x] `BubbleChart` _新增_
- [x] `Crosshair` _新增_
- [x] `DecompositionTree` _新增_
- [x] `DonutChart` _新增_
- [x] `Empty` _新增_
- [x] `EntriesOf` _新增_
- [x] `Frame` _新增_
- [x] `Grid` _新增_
- [x] `LabelHeight` _新增_
- [x] `LabelWidth` _新增_
- [x] `Legend` _新增_
- [x] `LegendHeight` _新增_
- [x] `Measure` _新增_
- [x] `NightingaleChart` _新增_
- [x] `Palette` _新增_
- [x] `ParetoChart` _新增_
- [x] `PercentBar` _新增_
- [x] `Plot` _新增_
- [x] `PolarChart` _新增_
- [x] `RangeHighlight` _新增_
- [x] `StackedBar` _新增_
- [x] `SunburstChart` _新增_
- [x] `Ticks` _新增_
- [x] `Tooltip` _新增_
- [x] `Widest` _新增_
- [ ] `Chart`, `ChartAnnotation`, `ChartAxis`, `ChartBrush`, `ChartCrosshair`, `ChartEmptyState`, `ChartExport`, `ChartGrid`, `ChartLegend`, `ChartTooltip`, `ChartTypeSwitcher`, `ChartValue`, `CompareLegend`, `DensityPlot`, `DepthChart`, `DrawingToolbar`, `HeikinAshiChart`, `IndicatorPane`, `IndicatorParamsField`, `IndicatorSelector`, `IntervalSelector`, `LineQuoteChart`, `MarketHeatmap`, `MultiChartLayout`, `NewChart`, `OHLCChart`, `RealtimeChart`, `Sunburst`, `TimeRangeSelector`, `TrendIndicator`, `VolumeChart`, `VolumeProfile`

## ui/project — 任务与看板　0/19

- [ ] `AssigneePicker`, `BurndownChart`, `DueDatePicker`, `IssueCard`, `IssueIdBadge`, `KanbanBoard`, `LabelManager`, `MilestoneProgress`, `PomodoroTimer`, `PriorityIndicator`, `PriorityName`, `SprintBoard`, `StatusSelect`, `SubtaskList`, `TaskDetailPanel`, `TaskItem`, `TaskList`, `TimeTracker`, `WorkloadView`

## ui/media — 音视频　0/21

- [ ] `AudioPlayer`, `AudioSpectrum`, `AudioWaveform`, `CameraPreview`, `DeviceSelector`, `ImageAnnotator`, `ImageCompare`, `ImageCropper`, `ImageThumbnail`, `Lightbox`, `MediaControls`, `MediaToggleAction`, `MicLevelMeter`, `PlaybackSpeedControl`, `Playlist`, `ScreenRecorderControls`, `SubtitleEditor`, `VideoPlayer`, `VideoScrubber`, `VideoThumbnailStrip`, `VolumeControl`

## ui/messaging — 消息与邮件　0/20

- [ ] `CallControls`, `ChannelList`, `ChatMessage`, `EmojiPicker`, `IncomingCallDialog`, `MailComposer`, `MailList`, `MailReader`, `MemberList`, `OnlineStatus`, `PinnedMessages`, `QuotedText`, `ReadReceipt`, `RecipientInput`, `SnoozePicker`, `StatusSetter`, `ThreadPanel`, `UnreadDivider`, `UserProfileCard`, `VideoCallGrid`

## ui/devtools — 调试面板　0/17

- [ ] `AlertList`, `ApiRequestBuilder`, `DashboardFilterBar`, `IncidentCard`, `JsonEditor`, `JsonViewer`, `KeyValueInput`, `LogStream`, `RefreshIntervalSelector`, `ResourceGauge`, `ResponseViewer`, `SchemaTree`, `ServiceStatus`, `StatCard`, `SystemMonitor`, `TraceWaterfall`, `UptimeBar`

## ui/finance — 行情与交易　0/39

- [ ] `AssetAllocationChart`, `BidAskBar`, `CurrencyConverter`, `DepthLadder`, `EarningsCalendar`, `EconomicCalendar`, `GreeksTable`, `IndexCard`, `Level2Quotes`, `LeverageSlider`, `MarginIndicator`, `MarketOverview`, `MarketStatus`, `NewsFeed`, `OptionChain`, `OrderBook`, `OrderConfirmDialog`, `OrderEntry`, `OrderTable`, `PayoffDiagram`, `PerformanceChart`, `PnLDisplay`, `PortfolioSummary`, `PositionTable`, `PriceChangeBadge`, `PriceText`, `QuickTradeButtons`, `QuoteCard`, `RiskMeter`, `Screener`, `SentimentGauge`, `SpreadIndicator`, `SymbolBadge`, `SymbolSearch`, `TickerTape`, `TimeAndSales`, `TradeHistoryTable`, `TradingSessionClock`, `Watchlist`

## ui/account — 账号与设置　0/22

- [ ] `AccentColorPicker`, `AccountSwitcher`, `ApiKeyManager`, `Coachmark`, `FeedbackWidget`, `KeyboardShortcutsList`, `LoginForm`, `OAuthButtons`, `OnboardingWizard`, `ProfileEditor`, `SessionList`, `SettingsLayout`, `SettingsRow`, `SettingsSearch`, `SettingsSection`, `ShortcutRecorder`, `ThemeSelector`, `TwoFactorInput`, `UsageQuota`, `UserMenu`, `WhatsNewDialog`, `WorkspaceSwitcher`

## ui/agent — Agent 运行过程　0/26

- [ ] `AgentCaption`, `AgentCard`, `AgentCode`, `AgentPlan`, `AgentProgress`, `AgentRows`, `AgentStatus`, `AgentStepList`, `AgentToggle`, `AgentWellRows`, `ArtifactPanel`, `ArtifactVersionSwitcher`, `CheckpointList`, `CommandExecutionCard`, `FileChangeCard`, `HumanInputRequest`, `MCPServerList`, `MemoryPanel`, `MultiFileDiffReview`, `PermissionPrompt`, `SandboxStatus`, `ScreenshotStream`, `SubAgentTree`, `ToolApprovalDialog`, `ToolCallCard`, `ToolCallGroup`

合计 **368/618 = 60%**。
