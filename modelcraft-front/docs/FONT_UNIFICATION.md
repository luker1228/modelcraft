# 字体统一UI实施文档

## 概述

本文档记录了ModelCraft前端字体系统统一的完整实施过程，基于 **Inter + Space Grotesk + Fira Code** 三字体组合，配合shadcn/ui组件库。

**实施日期**: 2026-02-22  
**目标**: 统一所有字体使用，规范字重使用，建立可维护的字体系统

---

## 一、字体系统架构

### 1.1 字体家族

| 字体 | Tailwind类 | 用途 | 加载配置 |
|------|-----------|------|----------|
| **Inter** | `font-sans` | UI文本、正文、表单 | `next/font/google` |
| **Space Grotesk** | `font-heading` | 标题、显示文本 | `next/font/google` |
| **Fira Code** | `font-mono` | 代码、技术标识符 | `next/font/google` |

### 1.2 字重规范 (CRITICAL)

| 字重 | Tailwind类 | 数值 | **严格用途** |
|------|-----------|------|-------------|
| Normal | `font-normal` | 400 | ✅ 正文、按钮、输入框、普通标签 (90%文本) |
| Medium | `font-medium` | 500 | ⚠️ **仅用于技术标识符** (代码、枚举名、API标识) |
| Semibold | `font-semibold` | 600 | ✅ 标题、卡片标题、表单标签 |
| Bold | `font-bold` | 700 | ✅ 页面主标题、强调文本 |

**关键原则**: `font-medium` (500) 是特例，仅用于技术内容！

---

## 二、核心文件

### 2.1 字体常量库

**文件**: `src/lib/typography.ts`

提供语义化排版组合：

```typescript
import { TYPOGRAPHY } from '@/lib/typography';

// 推荐用法：直接使用语义化组合
<h1 className={TYPOGRAPHY.pageTitle}>页面标题</h1>
<p className={TYPOGRAPHY.body}>正文内容</p>
<code className={TYPOGRAPHY.code}>UserRole</code>
<Button className={TYPOGRAPHY.button}>提交</Button>
```

**主要导出**:
- `FONT_FAMILIES`: 字体家族类 (`sans`, `heading`, `mono`)
- `FONT_WEIGHTS`: 字重类 (`normal`, `medium`, `semibold`, `bold`)
- `TEXT_SIZES`: 文字大小类 (`xs` - `3xl`)
- `TYPOGRAPHY`: 语义化组合 (40+ 预定义样式)

### 2.2 设计系统文档

**文件**: `.claude/skills/design-palette/SKILL.md`

新增 **Typography System** 章节，包含：
- 字体家族表格
- 字重使用规则
- shadcn/ui组件推荐用法
- 决策流程图
- 迁移示例

### 2.3 配置文件 (已有)

**`src/app/layout.tsx`**: 使用 `next/font/google` 加载三款字体
```typescript
const inter = Inter({ subsets: ['latin'], weight: ['300', '400', '500', '600', '700'], variable: '--font-inter' })
const spaceGrotesk = SpaceGrotesk({ subsets: ['latin'], weight: ['400', '500', '600', '700'], variable: '--font-space-grotesk' })
const firaCode = FiraCode({ subsets: ['latin'], weight: ['400', '500', '600'], variable: '--font-fira-code' })
```

**`tailwind.config.ts`**: 定义字体家族扩展
```typescript
fontFamily: {
  sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
  heading: ['var(--font-space-grotesk)', 'system-ui', 'sans-serif'],
  mono: ['var(--font-fira-code)', 'monospace']
}
```

---

## 三、字重修复计划

### 3.1 问题识别

通过代码扫描发现 **30+ 文件** 存在 `font-medium` 使用，其中大部分违反规范：

**❌ 错误使用**:
- 按钮文本: `<Button className="font-medium">`
- 普通标签: `<span className="font-medium text-slate-700">`
- 菜单项: `<DropdownMenuItem className="font-medium">`
- 表格单元格: `<td className="font-medium">`

**✅ 正确使用**:
- 代码块: `<code className="font-mono font-medium">`
- 枚举名: `<span className="font-mono font-medium">UserRole</span>`
- API标识: `<Badge className="font-mono font-medium">GET /api/users</Badge>`

### 3.2 修复优先级

#### 高优先级 (立即修复)
1. **UI组件** (`src/components/ui/`)
   - `label.tsx`, `button.tsx`, `input.tsx`, `table.tsx`
   - 这些是基础组件，影响全局

2. **布局组件** (`src/components/layout/`)
   - `Header.tsx`, `Sidebar.tsx`, `TopBar.tsx`
   - 高频使用组件

3. **shadcn/ui组件覆盖**
   - `badge.tsx`, `card.tsx`, `dialog.tsx`

#### 中优先级
4. **业务组件** (`src/components/`)
   - `ProjectCard.tsx`, `UserMenu.tsx`

5. **页面文件** (`src/app/`)
   - 按模块逐步修复

#### 低优先级
6. **备份文件** (`.backup` 后缀)
   - 可忽略或删除

### 3.3 修复策略

**批量替换规则**:

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 普通文本 | `font-medium` | `font-normal` |
| 按钮 | `font-medium` | `font-normal` |
| 表单标签 | `font-medium` | `font-semibold` |
| 卡片标题 | `font-medium` | `font-semibold` |
| 代码/标识符 | `font-medium` | `font-mono font-medium` (保持) |

**手动审查**:
- 技术内容（代码、枚举名、API路径）→ 保持 `font-medium`
- 其他所有内容 → 改为 `font-normal` 或 `font-semibold`

---

## 四、实施步骤

### 阶段一：基础设施 ✅ (已完成)

- [x] 创建 `src/lib/typography.ts` 常量库
- [x] 更新 `.claude/skills/design-palette/SKILL.md` 添加字体系统章节
- [x] 创建本实施文档 `docs/FONT_UNIFICATION.md`

### 阶段二：核心组件修复 (进行中)

**任务列表**:

- [ ] 修复 `src/components/ui/` 中的基础组件
  - [ ] `label.tsx` - 改为 `font-semibold`
  - [ ] `button.tsx` - 改为 `font-normal`
  - [ ] `input.tsx` - 改为 `font-normal`
  - [ ] `table.tsx` - 表头 `font-semibold`，单元格 `font-normal`
  - [ ] `badge.tsx` - 保持 `font-semibold`
  - [ ] `card.tsx` - 标题 `font-semibold`

- [ ] 修复 `src/components/layout/` 布局组件
  - [ ] `Header.tsx`
  - [ ] `Sidebar.tsx`
  - [ ] `TopBar.tsx`
  - [ ] `UnifiedSidebar.tsx`
  - [ ] `UserMenu.tsx`

- [ ] 修复 `src/components/` 业务组件
  - [ ] `ProjectCard.tsx`
  - [ ] 其他业务组件

### 阶段三：页面文件修复

- [ ] `src/app/org/[orgName]/project/[projectName]/` 项目相关页面
- [ ] `src/app/org/[orgName]/settings/` 设置页面
- [ ] `src/app/login/`, `src/app/org-selector/` 登录相关页面

### 阶段四：验证与测试

- [ ] 视觉回归测试 - 确保UI外观正确
- [ ] 无障碍测试 - 检查文本对比度
- [ ] 跨浏览器测试 - Chrome, Firefox, Safari
- [ ] 响应式测试 - 移动端、平板、桌面

### 阶段五：文档与维护

- [ ] 更新 `CLAUDE.md` 排版系统章节
- [ ] 创建字体使用决策流程图
- [ ] 添加 ESLint 规则（可选，防止未来违规）

---

## 五、使用指南

### 5.1 开发者快速参考

**决策树**:

```
需要添加文本？
│
├─ 是否为代码/枚举名/API标识？
│  ├─ 是 → 使用 TYPOGRAPHY.code 或 font-mono font-medium
│  └─ 否 ↓
│
├─ 是否为标题？
│  ├─ 是 → 使用 font-heading font-semibold/bold
│  │       (TYPOGRAPHY.pageTitle, sectionTitle, cardTitle)
│  └─ 否 ↓
│
└─ 普通文本/UI元素 → 使用 font-sans font-normal
                    (TYPOGRAPHY.body, button, input)
```

### 5.2 shadcn/ui 组件推荐

```tsx
import { TYPOGRAPHY } from '@/lib/typography';

// Button
<Button className={TYPOGRAPHY.button}>Create</Button>

// Card
<Card>
  <CardTitle className={TYPOGRAPHY.cardTitle}>Title</CardTitle>
  <CardContent>
    <p className={TYPOGRAPHY.body}>Content</p>
  </CardContent>
</Card>

// Form
<Label className={TYPOGRAPHY.label}>Name</Label>
<Input className={TYPOGRAPHY.input} />

// Table
<TableHead className={TYPOGRAPHY.tableHeader}>Header</TableHead>
<TableCell className={TYPOGRAPHY.tableCell}>Data</TableCell>

// Code
<code className={TYPOGRAPHY.code}>UserRole</code>
```

### 5.3 常见错误与修复

**错误 1: 按钮使用 font-medium**
```tsx
// ❌ 错误
<Button className="font-medium">Submit</Button>

// ✅ 正确
<Button className={TYPOGRAPHY.button}>Submit</Button>
// 或
<Button className="font-normal">Submit</Button>
```

**错误 2: 普通文本使用 font-medium**
```tsx
// ❌ 错误
<span className="font-medium text-slate-700">Status: Active</span>

// ✅ 正确
<span className={TYPOGRAPHY.body}>Status: Active</span>
```

**错误 3: 标题使用 font-medium**
```tsx
// ❌ 错误
<h2 className="font-medium text-xl">Section Title</h2>

// ✅ 正确
<h2 className={TYPOGRAPHY.sectionTitle}>Section Title</h2>
// 或
<h2 className="font-heading font-semibold text-xl">Section Title</h2>
```

**错误 4: 代码未使用 font-mono**
```tsx
// ❌ 错误
<code className="font-medium">UserRole</code>

// ✅ 正确
<code className={TYPOGRAPHY.code}>UserRole</code>
// 或
<code className="font-mono font-medium">UserRole</code>
```

---

## 六、质量指标

### 成功标准

- ✅ 100% 组件使用 `font-sans` / `font-heading` / `font-mono` 之一
- ✅ `font-medium` 仅出现在技术标识符（代码、枚举名等）
- ✅ 所有标题使用 `font-heading` + `font-semibold/font-bold`
- ✅ 正文默认 `font-normal` (400)
- ✅ 零硬编码 `font-family` 字符串
- ✅ 零 linter 错误
- ✅ 视觉回归测试通过

### 监控方式

**定期检查**:
```bash
# 查找 font-medium 使用（应仅在技术内容中出现）
cd modelcraft-react
grep -r "font-medium" src/ | grep -v "font-mono font-medium"

# 查找硬编码 font-family
grep -r "font-family" src/ --include="*.tsx" --include="*.ts"

# 运行 linter
npm run lint
```

---

## 七、维护与扩展

### 7.1 添加新的语义化组合

在 `src/lib/typography.ts` 中添加：

```typescript
export const TYPOGRAPHY = {
  // ... existing
  
  // 新增
  alert: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm}`,
  notification: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.base}`,
} as const;
```

### 7.2 调整字重

如需全局调整字重，修改 `layout.tsx` 中的字体加载配置：

```typescript
const inter = Inter({
  weight: ['400', '600', '700'], // 移除 500，强制规范化
  // ... other config
})
```

### 7.3 添加 ESLint 规则（可选）

创建 `.eslintrc.custom.js`：

```javascript
module.exports = {
  rules: {
    // 禁止单独使用 font-medium（必须配合 font-mono）
    'no-restricted-syntax': [
      'error',
      {
        selector: 'JSXAttribute[name.name="className"] > Literal[value=/font-medium(?!.*font-mono)/]',
        message: 'font-medium should only be used with font-mono for technical identifiers'
      }
    ]
  }
}
```

---

## 八、参考资料

### 内部文档
- `src/lib/typography.ts` - 字体常量库（源代码）
- `.claude/skills/design-palette/SKILL.md` - 设计系统完整文档
- `CLAUDE.md` - 前端开发规范
- `openspec/specs/design-system/spec.md` - 设计系统规范

### 外部资源
- [Inter Font](https://rsms.me/inter/) - UI字体官方网站
- [Space Grotesk](https://fonts.google.com/specimen/Space+Grotesk) - 标题字体
- [Fira Code](https://github.com/tonsky/FiraCode) - 代码字体
- [shadcn/ui](https://ui.shadcn.com/) - 组件库文档
- [Tailwind Typography](https://tailwindcss.com/docs/font-family) - Tailwind字体文档

---

## 九、变更日志

### 2026-02-22 - 初始版本
- 创建字体统一实施计划
- 建立 `src/lib/typography.ts` 常量库
- 更新设计系统文档
- 定义字重使用规范
- 规划修复路径

---

## 附录：快速命令

```bash
# 检查 font-medium 使用
grep -rn "font-medium" src/ | wc -l

# 查找需要修复的文件
grep -rl "font-medium" src/components/ui/

# 运行 linter
npm run lint

# 构建测试
npm run build

# 类型检查
npm run type-check
```

---

**最后更新**: 2026-02-22  
**维护人**: AI Agent (CodeBuddy)  
**状态**: 阶段一完成，阶段二进行中
