---
name: mcp-init
description: >
  帮助用户初始化 ModelCraft 项目所需的 MCP 服务。
  当用户提到“MCP 初始化”、“新环境需要哪些 MCP”、“本项目依赖哪些 MCP”、
  “.mcp.json 怎么配”时使用此 skill。
  优先检查项目根 .mcp.json、MCP 审批配置，以及 codegraph / playwright 是否可用。
---

# ModelCraft MCP 初始化

## 必需 MCP

| MCP 服务 | 建议服务名 | 用途 | 作用域建议 | 工具前缀 |
|----------|------------|------|------------|----------|
| **codegraph** | `codegraph` | 代码图谱、语义检索、跨文件理解 | `project` | `mcp__codegraph__` |
| **Playwright MCP** | `playwright` | 浏览器交互、截图、页面验收、前端联调验证 | `user` / `local` | `mcp__playwright__` |

## 常用补充 MCP

| MCP 服务 | 用途 | 工具前缀 |
|----------|------|----------|
| **gongfeng** | 工蜂代码管理平台：项目/分支/Issue/MR/代码评审 | `mcp__gongfeng__` |
| **chrome-devtools** | 浏览器开发工具（页面交互、截图、网络请求、控制台、性能分析、内存快照等） | `mcp__chrome-devtools__` |

## 初始化检查清单

1. **检查项目根 `.mcp.json`**
   - 项目级 MCP 配置放在仓库根目录 `.mcp.json`
   - 不要把 `mcpServers` 写到 `.codebuddy/settings.json`
   - 当前项目至少应声明 `codegraph`

2. **检查 MCP 审批设置**
   - 可在 `.codebuddy/settings.local.json` 或用户级 settings 中启用：
     - `"enableAllProjectMcpServers": true`
     - 或 `"enabledMcpjsonServers": ["codegraph"]`

3. **确认 Playwright MCP 可用**
   - 当前项目前端验收依赖 Playwright MCP
   - 会话里能看到 `mcp__playwright__*` 工具才算就绪
   - 若只做后端/文档工作，可暂不启用；若涉及 UI 联调或截图验证，优先确认它可用

4. **按需补充其他 MCP**
   - 需要代码托管 / MR / 评审时启用 `gongfeng`
   - 需要 DevTools 级调试时启用 `chrome-devtools`

## 项目约定

- `codegraph` 是本项目依赖的项目级 MCP
- `Playwright MCP` 是本项目依赖的前端验证 MCP
- 初始化本项目 MCP 时，优先确认以上两个服务，再补充其他集成
