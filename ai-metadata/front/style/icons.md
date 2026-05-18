# ModelCraft 图标体系文档

> 本文档是 ModelCraft 前端图标的**唯一真相源**。
> 记录所有 62 个在用图标的语义分类、使用位置、渲染上下文及 AI 生成 Prompt。
>
> **图标库**：`lucide-react@0.294.0`  
> **自定义图标目录**：`modelcraft-front/public/icons/`  
> **最后更新**：2026-05-17

---

## 概览

| 类别 | 图标数 | 优先级 |
|------|--------|--------|
| 品牌 / 产品身份 | 4 | 🔴 最高——优先替换为自定义图标 |
| 导航侧边栏 | 8 | 🟠 高——构成产品主框架视觉语言 |
| CRUD 操作动词 | 10 | 🟡 中——高频但语义通用 |
| 权限 / 安全 | 7 | 🟠 高——核心业务能力视觉标识 |
| 数据 / 模型结构 | 7 | 🟡 中——模型编辑器专属场景 |
| 状态反馈 | 9 | 🟢 低——约定俗成，lucide 已够用 |
| 用户 / 身份 | 6 | 🟡 中——用户体系相关 |
| 导航辅助 / 方向 | 9 | 🟢 低——纯方向性，无需自定义 |
| UI 控件 | 7 | 🟢 低——通用交互控件 |

---

## 品牌关键词（来自 DESIGN.md）

| 维度 | 约束 |
|------|------|
| 风格定位 | Stripe Dashboard · Precision Tool · 企业运营工具 |
| 主色调 | Action Indigo `#4F46E5` / Cool Blue-Gray 中性调 |
| 线条 | stroke-width 1.5px，简洁几何，无装饰 |
| 禁止 | 渐变、毛玻璃、圆形徽章、光晕、投影装饰 |
| 背景 | 纯色透明，不带框架 |
| 尺寸规范 | 导航 16px (`size-4`)、标准 20px (`size-5`)、强调 24px (`size-6`) |

### AI Prompt 基础模板

```
A single [描述] icon, minimal line-art style, stroke weight 1.5px,
monochrome #4F46E5 indigo on transparent background,
enterprise SaaS precision tool aesthetic similar to Stripe dashboard,
clean geometric shapes, 24x24 pixel grid, no gradients, no shadows,
no decorative elements, vector-style, flat design
```

---

## 一、品牌 / 产品身份

> 最高优先级替换目标。这类图标承载产品识别度，使用 lucide 通用图标会模糊品牌个性。

---

### `Sparkles` — 产品 Logo / AI 能力标记

**语义类别**：品牌  
**当前实现**：lucide `Sparkles`（四角星光芒形态）

**使用位置与渲染上下文**：

| 位置 | 尺寸 | 颜色 | 渲染语境 |
|------|------|------|----------|
| `auth-layout.tsx:43` | `size-6` | `text-primary-foreground`（白色） | 登录页左侧品牌区，深色 `bg-primary` 正方形背景框内，产品 Logo |
| `auth-layout.tsx:115` | `size-5` | `text-primary-foreground`（白色） | 登录页底部 AI 特性卡片徽章，深色背景 |
| `AppLayout.tsx:289` | `size-3.5` | `text-white` | 侧边栏顶部组织名旁，`bg-primary` 圆角方块内，**组织/产品身份徽标** |
| `AppLayout.tsx:503` | `size-4` | 继承 `text-foreground` | 侧边栏底部"快速开始"引导入口的左侧图标 |
| `PermissionsTab.tsx:341` | `size-3` | 继承 | 权限包徽章内，"AI 推荐"角标标记 |
| `FilterPanel.tsx:88` | `11px` | 继承 | 数据筛选面板 AI 过滤器触发按钮 |
| `guide/page.tsx:145` | `size-5` | `text-blue-600` | 项目引导页功能卡片图标 |

**关键上下文**：Logo 形态出现在两个位置——登录页品牌区（白色 on 深色背景）和侧边栏顶部组织徽标（白色 on `bg-primary` 小方块）。这是全项目唯一的品牌 Logo 图标载体。

**AI 生成 Prompt**：
> A single four-pointed sparkle star icon with a larger center star and smaller accent stars, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS precision tool aesthetic, clean geometric shapes, 24x24 pixel grid, no gradients, no shadows, no decorative elements, vector-style, flat design

---

### `Globe` — HTTP 接口 / 网络连接

**语义类别**：品牌特性 / 数据库配置  
**使用位置与渲染上下文**：

| 位置 | 尺寸 | 渲染语境 |
|------|------|----------|
| `auth-layout.tsx:15` | `size-5` | 登录页特性介绍卡片，"数据库 HTTP 接口"功能图标，`bg-primary` 圆角方块内白色显示 |
| `DatabaseConfigFields.tsx:113` | `size-3.5` | 数据库配置表单，主机地址输入框左侧前缀图标，`text-muted-foreground` |

**AI 生成 Prompt**：
> A single globe/earth icon showing latitude and longitude grid lines, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS precision tool aesthetic, clean geometric shapes, 24x24 pixel grid, no gradients, no shadows, flat design

---

### `ShieldCheck` — 权限已验证 / 访问健康

**语义类别**：品牌特性 / 权限状态  
**使用位置与渲染上下文**：

| 位置 | 尺寸 | 渲染语境 |
|------|------|----------|
| `auth-layout.tsx:21` | `size-5` | 登录页"权限管理"特性卡片图标，`bg-emerald-600` 背景内白色显示 |
| `UsersTab.tsx:669` | `size-5` | `text-primary`，用户已拥有有效角色的状态指示 |
| `BundlesTab.tsx:274` | `size-8` | `text-muted-foreground/30`，权限包列表为空时的空状态图标 |
| `workspace/page.tsx:625` | `size-3.5` | `text-muted-foreground/50`，项目卡片权限状态徽章 |
| `end-users/[userId]/page.tsx:331` | `size-4` | 用户详情页已授权角色行内状态图标 |

**AI 生成 Prompt**：
> A single shield icon with a checkmark inside, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS precision tool aesthetic, clean geometric shapes, 24x24 pixel grid, no gradients, no shadows, flat design

---

### `Link2` / `LinkIcon` — 外键关联 / 逻辑关系

**语义类别**：品牌特性 / 数据关系  
**使用位置与渲染上下文**：

| 位置 | 尺寸 | 渲染语境 |
|------|------|----------|
| `auth-layout.tsx:27` (`as LinkIcon`) | `size-5` | 登录页"逻辑外键与枚举"特性卡片图标，`bg-amber-600` 背景内白色显示 |
| `InsertFieldSheet.tsx:543` | `size-3.5` | `text-primary`，插入字段面板关联字段类型选项图标 |
| `OneToManyRelationManagerSection.tsx:287` | `size-4` | `text-primary`，一对多关系管理区标题图标 |
| `ForeignKeyPanel.tsx:171` | `size-6` | `opacity-30`，外键面板为空时的空状态提示图标 |
| `ModelRecordTable.tsx:549` | `size-3.5` | 数据表关联字段列头图标 |

**AI 生成 Prompt**：
> A single chain link icon showing two interlocked oval rings, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS precision tool aesthetic, clean geometric shapes, 24x24 pixel grid, no gradients, no shadows, flat design

---

## 二、导航侧边栏

> 构成产品主框架的视觉语言，16px 尺寸，`text-ink-muted` 默认色，激活时 `text-primary`。

---

### `FolderOpen` — 项目列表

**使用位置**：`AppLayout.tsx` 侧边栏 nav item（标签"项目"）、`ProjectCard.tsx` 项目卡片图标、`workspace/page.tsx` 空状态

**渲染上下文**：侧边栏导航最顶层入口，链接到 `/org/{orgName}/workspace`，配文字标签"项目"。

**AI 生成 Prompt**：
> A single open folder icon with an open top flap, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `Users` — 开发者管理

**使用位置**：`AppLayout.tsx` 侧边栏（标签"开发者"）、`UsersTab.tsx`、`EndUsersManagementTable.tsx`、`developers/members/page.tsx`

**渲染上下文**：侧边栏 Org 区段导航，链接到 `/org/{orgName}/developers`。也用于开发者成员列表的列表头和用户组相关的表格图标。

**AI 生成 Prompt**：
> A single icon showing two overlapping human silhouettes (group of people), minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `KeyRound` — 终端用户授权

**使用位置**：`AppLayout.tsx` 侧边栏（标签"终端用户"和"用户授权"）、`roles/[roleId]/page.tsx`、`roles/permissions/page.tsx`

**渲染上下文**：侧边栏出现两次——Org 级"终端用户管理"和 Project 级"用户授权"。钥匙圆头设计区别于普通 `Key`，暗示访问凭证语义。

**AI 生成 Prompt**：
> A single key icon with a round head/bow and notched blade pointing right, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `Table2` — 数据模型编辑器

**使用位置**：`AppLayout.tsx` 侧边栏（标签"数据模型"）、`ModelSidebar.tsx`、`ModelDetailPanel.tsx`、`ImportModelDialog.tsx`、`data/page.tsx`

**渲染上下文**：Project 区段核心导航，链接到模型编辑器。也用于模型侧边栏模型列表项图标，以及导入弹窗的场景图标。

**AI 生成 Prompt**：
> A single table/grid icon showing rows and columns with header row, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `List` — 枚举管理 / 列表视图

**使用位置**：`AppLayout.tsx` 侧边栏（标签"枚举管理"）、`view-toggle.tsx`（列表视图按钮）、`RecordWorkspace`（视图模式）

**渲染上下文**：双重语义——作为导航图标指向枚举管理页；作为视图切换按钮（与 `LayoutGrid` 配对），切换数据工作区的显示模式。

**AI 生成 Prompt**：
> A single list icon showing three horizontal lines with small dots/bullets on the left, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `Shield` — 访问控制

**使用位置**：`AppLayout.tsx` 侧边栏（标签"访问控制"）、`RoleTable.tsx`（角色列表图标）、`bundles/[bundleId]/page.tsx`

**渲染上下文**：Project 导航区段，链接到角色与权限管理页面，是权限体系的入口图标。

**AI 生成 Prompt**：
> A single shield icon with a slightly pointed bottom, clean outline, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `Settings` — 设置

**使用位置**：`AppLayout.tsx` 侧边栏（标签"组织设置"和"项目设置"）、`UserMenu.tsx`（用户菜单"设置"项）、`FieldEditSheet.tsx`

**渲染上下文**：侧边栏出现两次——Org 级和 Project 级设置。也出现在顶栏用户下拉菜单。

**AI 生成 Prompt**：
> A single gear/cog icon with evenly spaced teeth around a central circle, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

### `PanelLeft` — 侧边栏折叠

**使用位置**：`AppLayout.tsx`（折叠/展开按钮）、`sidebar.tsx`

**渲染上下文**：侧边栏右上角控制按钮，点击后收缩侧边栏为 64px 图标模式。Ghost 按钮，hover 时显示背景。

**AI 生成 Prompt**：
> A single panel-left icon showing a vertical sidebar panel on the left with an arrow or line indicating collapse, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 pixel grid, no gradients, flat design

---

## 三、CRUD 操作动词

> 高频使用，遍布全项目，语义通用，lucide 已足够，自定义优先级低。

| 图标 | 语义 | 典型尺寸 | 主要位置 | 渲染上下文摘要 |
|------|------|----------|----------|----------------|
| `Plus` | 新增 | `size-4` | 全项目 DropdownMenuItem、Toolbar 按钮 | 始终配文字标签如"新增模型"、"添加字段"，放在按钮最左侧 |
| `Edit` | 行内编辑 | `size-4` | ProjectCard、RecordWorkspace、ModelSidebar | DropdownMenuItem 中 `<Edit className="mr-2 size-4" />编辑`，与 Trash2 配对出现 |
| `Pencil` | 详情页编辑 | `size-4` | ProfileOverviewPanel、roles/bundles 详情 | 独立的"编辑"按钮图标，不与下拉菜单结合 |
| `Trash2` | 删除 | `size-4` | 全项目删除操作 | 始终配 `text-destructive` 颜色，在 DropdownMenuItem 中与 Edit 配对 |
| `Copy` | 复制文本 | `size-3.5`~`size-4` | organization-name-input、identity-form-section | 出现在只读输入框右侧，点击复制 API Key / Org Name |
| `Save` | 保存 | `size-4` | InsertFieldSheet、login-settings | 表单提交按钮图标，与"保存"文字并排 |
| `Archive` | 归档 | `size-4` | ModelRecordTable、ModelDetailPanel | 数据记录软删除操作，区别于 Trash2 的永久删除 |
| `Search` | 搜索 | `size-4` | search-input 前缀、AppLayout 全局搜索 | 始终作为输入框左侧前缀图标，`text-muted-foreground` |
| `RefreshCw` | 刷新 | `size-3.5`~`size-4` | RelationPicker、EndUserTables、identity-form-section | 数据刷新/重载按钮，多为 Ghost 图标按钮 |
| `RotateCcw` | 回滚 | `size-3.5` | bundles/[bundleId]（权限包版本回滚）、OnboardingPanel | 权限包版本回滚按钮，配"还原"文字，与 Loader2 切换（加载中状态） |

---

## 四、权限 / 安全

> 核心业务能力的视觉标识，有完整的语义谱系，建议成套设计。

---

### `Shield` 谱系

| 图标 | 语义 | 使用位置 | 渲染上下文 |
|------|------|----------|------------|
| `Shield` | 角色/访问控制入口 | RoleTable、bundles 详情页 | 角色列表行图标，中性色 `text-muted-foreground` |
| `ShieldCheck` | 已授权 / 权限健康 | UsersTab、workspace、end-user 详情 | 绿色/主色强调，用户已有有效角色时显示 |
| `ShieldAlert` | 权限告警 / 无权限 | PermissionsTab 警告角标、no-project-access 页 | `text-destructive` 或警告色，用户访问被拒时的空状态主图标 |
| `ShieldOff` | 已撤权 / 未授权 | EndUserAccessTable、roles/page、end-user 详情 | 灰色/弱化色，表示该用户当前无有效授权 |

**成套 AI Prompt 提示**：四个图标应风格完全一致，仅内部符号不同：
- `Shield`：空盾牌轮廓
- `ShieldCheck`：盾内加 ✓
- `ShieldAlert`：盾内加 !
- `ShieldOff`：盾上加对角斜线

---

### `Key` 与 `KeyRound`

| 图标 | 语义 | 使用位置 | 渲染上下文 |
|------|------|----------|------------|
| `Key` | 主键字段标识 | ModelDetailPanel、ModelRecordTable | `text-warning`，出现在字段行的主键标记小方块内，`size-3`，非常小 |
| `KeyRound` | 终端用户授权 | AppLayout nav、roles/permissions | 导航图标，`size-4`，较 `Key` 更有"访问权限"含义 |

**AI Prompt — `Key`**：
> A single key icon with a diamond/rectangular bow and simple notched blade, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, 24x24 grid, flat design

---

### `Lock` — 加密 / 只读保护

**使用位置**：`PermissionsTab.tsx`（已锁定权限）、`UsersTab.tsx`（锁定状态）、`DatabaseConfigFields.tsx`（密码字段前缀）

**渲染上下文**：出现在密码/敏感字段的输入框前缀（`size-3.5 text-muted-foreground`），以及权限表格中"不可修改"行的状态标记。

**AI Prompt**：
> A single padlock icon in closed/locked position, minimal line-art style, stroke weight 1.5px, monochrome #4F46E5 indigo on transparent background, enterprise SaaS aesthetic, 24x24 grid, flat design

---

## 五、数据 / 模型结构

| 图标 | 语义 | 典型尺寸 | 主要位置 | 渲染上下文摘要 |
|------|------|----------|----------|----------------|
| `Database` | 数据库/集群 | `size-3.5`~`size-7` | workspace 空状态引导卡、ModelSidebar、settings | workspace 中以 `bg-primary` 圆形徽章包裹（`size-7`），空状态引导"创建第一个项目" |
| `Table2` | 数据表/模型 | `size-4` | ModelSidebar 模型列表项、ImportModelDialog | 每个模型的行内图标，与模型名称并排 |
| `Columns` | 字段/列插入 | `size-4` | InsertFieldSheet、ModelRecordInsertMenu | 字段插入面板的场景图标，配"插入字段"标签 |
| `Tags` | 枚举类型字段 | `size-4` | FieldEditSheet 字段类型标记 | 字段类型为 ENUM 时显示，标识枚举类型 |
| `Link2` | 外键/关联字段 | `size-3.5`~`size-6` | InsertFieldSheet、ForeignKeyPanel | 字段类型为外键时的图标；ForeignKeyPanel 空状态大图标（`opacity-30`） |
| `Unlink` | 断开关联 | `size-4` | OneToManySection、RecordRelationManagerDialog | "解除关联"操作按钮，通常与 Link2 配套出现 |
| `ExternalLink` | 跳转关联记录 | `size-3.5` | EndUsersManagementTable、ModelEditorView | 在表格行内提供"在新上下文查看"的跳转入口 |

---

## 七、用户 / 身份

| 图标 | 语义 | 典型位置 | 渲染上下文摘要 |
|------|------|----------|----------------|
| `User` | 单个用户 / 用户名 | DatabaseConfigFields 用户名前缀、end-user 详情 | 数据库连接配置表单的用户名输入框前缀（`size-3.5 text-muted-foreground`） |
| `Users` | 用户组 / 成员列表 | AppLayout nav（开发者）、UsersTab、developers/members | 见导航侧边栏章节 |
| `CircleUserRound` | 当前用户头像占位 | `UserMenu.tsx` 顶部用户下拉菜单 | 用户菜单 DropdownMenuItem，配"个人资料"文字，`size-4 mr-2` |
| `Building2` | 组织切换器 | `organization-switcher.tsx` | 组织切换下拉菜单的组织图标占位，当组织无自定义头像时显示 |
| `LogIn` | 登录入口 | *(已 import，无直接 JSX 渲染，疑为条件渲染或遗留)* | — |
| `LogOut` | 退出登录 | `user-menu.tsx`、`UserMenu.tsx` | 用户下拉菜单底部"退出登录"操作，`size-4 mr-2` |

---

## 八、导航辅助 / 方向

> 纯方向性图标，无业务语义，无需自定义。

| 图标 | 语义 | 典型位置 |
|------|------|----------|
| `ChevronDown` | 折叠展开 / 下拉触发 | select 组件、AppLayout 可展开导航组、UserMenu |
| `ChevronUp` | 折叠收起 | select 组件（展开态）、OnboardingPanel |
| `ChevronLeft` | 多步骤上一步 | ImportModelDialog、CreatePermissionSheet |
| `ChevronRight` | 面包屑分隔 / 行跳转 | breadcrumb（分隔符）、AppLayout 子菜单箭头、UsersTab 行跳转 |
| `ChevronsUpDown` | 可展开下拉（Combobox）| RelationSelector、RecordRelationManagerDialog、ModelSidebar |
| `ArrowLeft` | 返回上级页面 | roles/[roleId]、permissions、bundles/[bundleId]、end-users/[userId] |
| `ArrowRight` | 进入详情 / 下一步 | BundlesTab 列表行、CreateEndUserDialog 下一步、guide 页导航 |
| `ArrowUpDown` | 排序触发 | `SortPopover.tsx` 排序弹出面板触发按钮 |
| `Circle` | 单选指示器（空心圆） | `dropdown-menu.tsx` Radio Item 未选中态 |

---

## 九、UI 控件

| 图标 | 语义 | 典型位置 | 渲染上下文摘要 |
|------|------|----------|----------------|
| `LayoutGrid` | 网格视图切换 | `view-toggle.tsx` | 与 `List` 配对的 ToggleGroup，`aria-label="网格视图"`，`size-4` |
| `MoreHorizontal` | 更多操作菜单（⋮） | ProjectCard、EndUserAccessTable、EndUsersManagementTable、breadcrumb | 表格行 hover 时出现，`opacity-0 group-hover:opacity-100` 动画，触发 DropdownMenu |
| `Eye` | 查看密码 / 预览 | password-input、DatabaseConfigFields、ErrorHistoryDialog | 密码框右侧切换图标；ErrorHistoryDialog 中"展开详情"按钮 |
| `EyeOff` | 隐藏密码 | password-input、DatabaseConfigFields | 密码可见时的切换图标，与 `Eye` 互斥显示 |
| `HelpCircle` | 帮助提示 tooltip | InsertFieldSheet、AppLayout | `size-3.5 cursor-help text-muted-foreground`，悬停展开 TooltipContent，如"UUID v7 说明"、"关联字段说明" |
| `Settings2` | 高级设置 | *(已 import，无直接 JSX 渲染，可能遗留)* | — |
| `History` | 错误历史记录 | `ErrorHistoryDialog.tsx` | 对话框标题图标（`size-5`）+ 空状态大图标（`size-12 opacity-50`），配"暂无错误记录"文字 |

---

## 自定义替换优先级清单

| 优先级 | 图标 | 理由 |
|--------|------|------|
| 🔴 P0 | `Sparkles` | 产品 Logo，当前用通用图标，品牌辨识度弱 |
| 🔴 P0 | `Globe` + `ShieldCheck` + `Link2` | 登录页品牌特性三件套，需风格统一 |
| 🟠 P1 | 导航侧边栏 8 个图标 | 用户每次使用都会看到，品牌感知高 |
| 🟡 P2 | `Database`, `Table2`, `Columns` | 模型编辑器核心视觉语言 |
| 🟢 P3 | 其余通用操作/状态图标 | lucide 已足够，无需替换 |
