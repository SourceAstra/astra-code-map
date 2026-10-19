# 更新日志 (Changelog)

AstraMap 项目的所有重要版本变更都将记录在此文件中。

遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 规范，并遵守 [语义化版本 2.0.0](https://semver.org/lang/zh-CN/)。

## [v0.2.1] - 2026-10-19

### 🔍 探索检索修复 (`astramap_explore`)
- **camelCase 符号零召回修复**：FTS5 unicode61 分词器不拆分驼峰命名，`resolveContext` 是单一 token，此前精确词匹配在 Go/JS 符号上召回率趋近于零。现对每个检索词做前缀扩展（`resolve*`），稳定命中该 token 前缀。
- **bm25 候选池扩容排序**：先在 4 倍于 `maxFiles` 的候选池上按 bm25 排名，再收敛到 `maxFiles` 个文件，避免词法噪声节点挤掉部分匹配结果。
- **片段符号降噪（保留成员字段）**：探索结果排除局部变量/参数类碎片节点，但保留成员字段——id 含 `#` 的节点（C 的 struct 函数指针字段如 `fireflys_api_s#fireflys_acl_init`、Go 的 struct 字段如 `ProjectSecretStore#Put`）是真实语义符号，不做一刀切排除。
- **结果对称截断**：最终文件列表按 `maxFiles` 硬截断（此前仅限制候选查询，不限制输出文件数）。

### 🕸️ 调用图
- **explore 关系输出扩展**：关系列表在 callers 之外新增 callees（每符号限 20 条）；关系收集改为仅针对保留进结果的符号（原始节点 ID 查询，输出前统一规范化），避免被截断文件的无谓收集。
- **callees 确定性排序**：callees 查询增加 `ORDER BY target, line, id`，输出顺序稳定可复现；新增 `GetCalleesLimited` 支持限额查询（0 = 不限量）。

### 🔧 工具安装 (`amap install`)
- **MCP 配置可移植化**：项目级 `.agents/mcp_config.json` 改用 `--project .` 相对路径；全局级配置不再硬编码 `--project`，服务随当前工作区 CWD 自适应，杜绝配置中的物理绝对路径。
- **Agent 规则文案强化**：注入 AGENTS.md 的 AstraMap 调度规则明确要求通过 `call_mcp_tool(ServerName="astramap", ToolName=...)` 调用，严禁降级为 `grep_search`。

### 📏 函数边界解析修复
- **字符串/注释中的花括号截断修复**：`estimateFunctionEndLine` 重写为跨行状态机，正确跟踪字符串字面量、块注释与反斜杠续行。修复 C 项目 CLI 宏多行语法字符串（如 `firefly_l2_cli.c` 的 `"l2 fdb add {port P | nexthop N} ..."`）中的花括号被误计为代码花括号、导致函数范围提前截断的问题；新增 `astramap/extent_test.go` 回归测试。

### 🛡️ 稳定性
- **索引卡死看门狗**：`amap index` 的 post-SCIP 阶段停滞 10 分钟时，自动导出 goroutine dump 到 `.astramap/hang-dump.txt` 并 exit(1)，避免 SQLite 驱动死锁永久持有数据库写锁、挂起同项目后续所有 amap 命令；新增 `cmd/amap/watchdog_test.go` 触发测试。
- **依赖升级**：Go 1.25 → 1.26；`modernc.org/sqlite` v1.50.1 → v1.60.1。

---

## [v0.2.0] - 2026-07-30

### 🚀 架构设计与核心定位
- **SCIP 语义主导 + Tree-sitter 语法覆盖架构**：明确并确立了系统“SCIP 高精度语义为主，Tree-sitter 实时语法为辅”的双层混合架构。SCIP 提供 95%+ 编译器级的跨文件类型推导与消歧，Tree-sitter 负责毫秒级的文件变更增量热更新。
- **文档全面重构**：重构了 `README.md` 与 `README_EN.md`，开篇突出了高精度代码地图的核心优势、Mermaid 双层架构图、与纯 Tree-sitter 工具的痛点横评以及 Web Dashboard 三大交互视界。

### ⚡ 性能优化与增量索引
- **Watch 模式下的防抖 SCIP 收敛**：重构了 `amap watch` 循环，彻底隔离了高频保存文件时的 SCIP 重新编译开销。高频编辑时仅触发 Tree-sitter 毫秒级补丁，引入 2 分钟静默防抖定时器，仅在系统闲置时自动于后台运行一次高精度 SCIP 收敛。
- **精准日志输出**：优化了增量更新时的终端日志，清晰显示哪些文件发生了语法层更新，并在 SCIP 更新完成时提供明确提示 (`watch successfully refreshed map via SCIP index update`)。

### 🎯 启发式解析 (Heuristic) 与消歧增强
- **同包路径就近消歧**：增强了 `ResolveCrossFileCalls` 的启发式解析。当遇到多义同名短函数时，优先匹配与调用者处于物理同包（同一目录）下的定义节点。
- **Go 隐式函数拦截**：显式拦截了 Go 语言运行时的隐式初始化函数 `init()`，杜绝启发式解析建立错误的 `calls -> init` 连线。
- **成员选择符盲连防御**：对带选择符的调用（`obj.Method()`）增加了消歧阀门。若未通过全限定名 `QualifiedName` 精确命中，且存在多个不同结构体的同名方法时，禁止退化为全库盲目乱连，防止污染调用图。
- **跨平台路径规范化**：修复了 Windows 平台下路径斜杠分隔符不一致导致同包消歧失效的问题，统一采用 `filepath.ToSlash` 进行标准化。

---

## [v0.1.0] - 2026-07-25

### 🎉 首次发布
- AstraMap 语义代码地图引擎首个版本发布。
- 原生 stdio MCP Server，提供 9 大结构化工具（`search`、`explore`、`node`、`callers`、`callees`、`impact`、`trace`、`status`、`files`）。
- 内置 Web Dashboard，提供探索视界、力导向依赖拓扑网与模块理解文档。
- 支持 12 种主流语言的 SCIP 导入与 Tree-sitter 基础解析。
