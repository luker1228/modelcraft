---
name: icon-manager
description: >
  ModelCraft 项目图标管理技能，覆盖图标全生命周期。当用户提到"图标"、"icon"、"收集图标"、
  "扫描图标"、"生成图标 prompt"、"切图"、"替换图标"、"图标库"、"图标审计"、"更新 logo"
  时，必须立即触发此技能。适用场景：(1) 盘点项目中所有图标使用情况；(2) 根据品牌调性生成
  AI 绘图 Prompt；(3) 将 AI 生成的图片切分为单独透明 PNG；(4) 将新图标替换到项目代码中。
---

# Icon Manager — ModelCraft 图标管理

本技能分四个阶段，每个阶段都可以独立执行。根据用户意图，跳到对应阶段即可。

---

## 阶段一：盘点 — 收集所有图标使用情况

**目标：** 产出一份结构化的图标清单，记录每个图标的名称、用途、使用位置。

### 执行步骤

1. **扫描 lucide-react 用法**

```bash
grep -rh "lucide-react" modelcraft-front/src --include="*.tsx" --include="*.ts" \
  | grep -o 'import {[^}]*}' | grep -o '[A-Z][a-zA-Z0-9]*' | sort | uniq
```

2. **定位每个图标的使用上下文**

对关键图标运行：

```bash
grep -rn "IconName" modelcraft-front/src --include="*.tsx" | head -20
```

重点关注以下语义区域：
- 品牌/Logo：`Sparkles`（当前作为 Logo 图标使用，出现在 auth-layout、AppLayout）
- 导航侧边栏：`FolderOpen`（项目）、`Users`（开发者）、`KeyRound`（终端用户/用户授权）、`Table2`（数据模型）、`List`（枚举）、`Shield`（访问控制）、`Settings`（设置）
- 操作动词：`Plus`（新增）、`Edit/Pencil`（编辑）、`Trash2`（删除）、`Copy`（复制）、`Save`（保存）、`RefreshCw`（刷新）、`RotateCcw`（重置）、`Archive`（归档）、`Search`（搜索）
- 状态反馈：`Check/CheckCircle2`（成功）、`AlertCircle/AlertTriangle/ShieldAlert`（警告/错误）、`Loader2`（加载中）、`X/XCircle`（关闭/失败）
- 权限安全：`Shield`、`ShieldCheck`、`ShieldOff`、`Lock`、`Key`、`KeyRound`
- 数据/模型：`Database`、`Table2`、`Columns`、`Tags`、`ExternalLink`、`Link2`、`Unlink`
- 导航辅助：`ChevronLeft/Right/Up/Down`、`ChevronsUpDown`、`ArrowLeft/Right`、`ArrowUpDown`
- 用户/身份：`User`、`Users`、`CircleUserRound`、`Building2`（组织）、`Globe`（HTTP/全球）
- UI 控件：`MoreHorizontal`（更多菜单）、`PanelLeft`（侧边栏折叠）、`LayoutGrid/List`（视图切换）、`Eye/EyeOff`（密码显隐）、`LogIn/LogOut`（登录登出）
- 辅助功能：`HelpCircle`（帮助）、`Settings2`（高级设置）、`History`（历史）、`Bug`（调试）、`FolderOpen`（文件夹）

3. **输出清单格式**

以 Markdown 表格输出，字段：图标名 | 语义类别 | 使用位置 | 备注

---

## 阶段二：生成 AI 绘图 Prompt

**目标：** 根据品牌调性（DESIGN.md）为每个图标生成标准化 AI 生成 prompt。

### 品牌关键词（从 DESIGN.md 提炼）

| 维度 | 关键词 |
|------|--------|
| 风格 | Stripe Dashboard、Precision Tool、企业级、运营工具 |
| 主色调 | Action Indigo `#4F46E5`、Cool Blue-Gray 中性色 |
| 设计哲学 | 克制、精准、通过减法体现高级感 |
| 线条 | 细线条（stroke-width 1.5）、简洁几何 |
| 不允许 | 渐变、毛玻璃、圆形徽章、装饰性动画 |
| 背景 | 纯色透明背景，不带任何框架或光晕 |

### Prompt 模板

```
A single [图标语义描述] icon, minimal line-art style, 
stroke weight 1.5px, monochrome using #4F46E5 indigo on transparent background,
enterprise SaaS precision tool aesthetic similar to Stripe dashboard,
clean geometric shapes, 24x24 pixel grid, no gradients, no shadows,
no decorative elements, vector-style, flat design
```

### 生成规则

- 每个图标一条 prompt
- 语义描述要精准（e.g. "database cylinder" 而非 "storage"）
- 品牌相关图标（Logo 等）可以使用带背景版本
- 对于语义接近的图标（如 Shield 系列），添加区分描述词：
  - `ShieldCheck` → "shield with check mark inside"
  - `ShieldAlert` → "shield with exclamation mark"
  - `ShieldOff` → "shield with diagonal strikethrough"

### 执行指令

读取阶段一输出的图标清单，为每一行生成对应 prompt，输出格式：

```markdown
### [图标名] — [语义类别]
**用途：** [使用场景]
**Prompt：**
> A single [描述] icon, minimal line-art...
```

---

## 阶段三：图片切分 — 将 AI 生成图片切割为单个图标

**目标：** 将 AI 生成的包含多个图标的图片，切分为单独的透明 PNG 文件。

### 使用切分脚本

```bash
python3 .claude/skills/icon-manager/scripts/slice_icons.py \
  --input <图片路径> \
  --output <输出目录> \
  --rows <行数> \
  --cols <列数> \
  --names <icon1,icon2,...> \
  --padding 4
```

参数说明：
- `--input`: AI 生成的图片路径（PNG / JPG / WebP）
- `--output`: 输出目录，默认 `modelcraft-front/public/icons/`
- `--rows` / `--cols`: 图标网格行列数
- `--names`: 对应图标名称列表（逗号分隔，按行优先顺序）
- `--padding`: 每个图标四周额外裁掉的像素（去除间隔线）

### 命名规范

输出文件命名格式：`icon-[功能名]-[变体].png`

示例：
- `icon-sparkles.png` — Brand Logo 图标
- `icon-folder-open.png` — 项目导航图标
- `icon-database.png` — 数据模型图标
- `icon-shield-check.png` — 权限验证状态图标

### 处理规则

- 自动转换为 RGBA 模式（支持透明通道）
- 白色背景自动转透明（阈值 240）
- 输出尺寸统一为 24×24（可通过 `--size` 参数调整为 32/48/64）

---

## 阶段四：应用 — 将图标替换到项目代码

**目标：** 将自定义图标 PNG 引入前端代码，替换对应的 lucide-react 图标。

### 替换策略

**方式 A：直接 img 标签（推荐用于自定义品牌图标）**

```tsx
// 替换前（使用 lucide-react）
<Sparkles className="size-6 text-primary-foreground" strokeWidth={1.5} />

// 替换后（使用自定义 PNG）
<img src="/icons/icon-sparkles.png" alt="ModelCraft" className="size-6" />
```

**方式 B：创建封装组件（推荐用于频繁使用的图标）**

创建 `modelcraft-front/src/web/components/ui/icon.tsx`：

```tsx
interface IconProps {
  name: string
  size?: number
  className?: string
  alt?: string
}

export function Icon({ name, size = 24, className, alt = '' }: IconProps) {
  return (
    <img
      src={`/icons/icon-${name}.png`}
      alt={alt}
      width={size}
      height={size}
      className={className}
    />
  )
}
```

用法：
```tsx
<Icon name="sparkles" size={24} />
<Icon name="database" size={16} className="opacity-70" />
```

### 替换步骤

1. 确认图标文件已存在于 `modelcraft-front/public/icons/`
2. 找到要替换的文件位置（使用阶段一的清单）
3. 替换 import 和 JSX（移除 lucide import，添加 img 或 Icon 组件）
4. 保留原有 className 中的尺寸类（`size-4` / `size-6` 等）
5. 运行前端检查：`cd modelcraft-front && npm run lint`

### 回滚方式

所有修改均可通过 git 回滚：
```bash
git diff --stat HEAD
git checkout -- modelcraft-front/src/path/to/file.tsx
```

---

## 快速参考：当前项目图标体系

### 品牌图标
| 图标 | 当前实现 | 位置 |
|------|----------|------|
| 产品 Logo | `Sparkles` (lucide) | auth-layout、AppLayout header |
| 特性图标-HTTP | `Globe` | auth-layout 功能介绍 |
| 特性图标-权限 | `ShieldCheck` | auth-layout 功能介绍 |
| 特性图标-外键 | `Link2` (as LinkIcon) | auth-layout 功能介绍 |

### 导航图标
| 图标 | 对应功能 |
|------|----------|
| `FolderOpen` | 项目列表 |
| `Users` | 开发者管理 |
| `KeyRound` | 终端用户 / 用户授权 |
| `Table2` | 数据模型编辑器 |
| `List` | 枚举管理 |
| `Shield` | 访问控制 |
| `Settings` | 设置（组织/项目） |

---

## 注意事项

- 公共 icons 目录：`modelcraft-front/public/icons/`（已添加到 `.gitignore` 则需手动排除）
- 自定义图标优先用于品牌相关场景，通用 UI 图标（箭头、加减号等）继续使用 lucide-react
- 替换时保持语义一致，不要仅因为"有了新图标"就替换掉正常运作的 lucide 图标
- 图标尺寸遵循项目规范：导航 `size-4`(16px)，标准 `size-5`(20px)，强调 `size-6`(24px)
