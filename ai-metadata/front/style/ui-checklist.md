# UI Checklist

> 每次提交前、CR 时、AI 生成代码后必查。规则来自真实改错记录。

---

## 颜色

- [ ] **禁止 `bg-blue-100 text-blue-600` 作为选中/激活态** — nav item、list item 统一用 `bg-accent text-foreground`
- [ ] **蓝色只出现在**：创建/提交等主操作按钮（`bg-primary`）、当前 org 指示符（`text-blue-600` Check icon）、数据库选择器选中态（`bg-primary/5 border-primary/40`）
- [ ] **禁止 `text-gray-*` 具体值** — 主文本用 `text-foreground`，次要文本用 `text-muted-foreground`
- [ ] **禁止 `font-bold`** — 最高用 `font-semibold`（600），正文用 `font-medium`（500）
- [ ] **`font-semibold` 不用于列表条目名称** — 列表里的 name/label 用 `text-foreground`（默认 400/500），不加 semibold

---

## 选中 / 激活态

- [ ] **Nav item 激活态**：`bg-accent text-foreground`，hover 态：`hover:bg-accent/60 hover:text-foreground`
- [ ] **List item（侧边栏模型列表等）**：只有 hover 态（`hover:bg-accent/60`），**不加持久选中背景**
- [ ] **Tab 激活态**：下划线 `border-b-2 border-foreground`（黑色），不用蓝色
- [ ] **选择类 UI（单选卡片、bundle 选择）**：允许 `border-primary bg-primary/5`，这是主动选择操作的反馈

---

## 页面宽度规范

- [ ] **内容页面统一用 `maxWidth="7xl"`（1280px）** — 这是项目标准宽度；`6xl` 仅在极特殊场景使用
- [ ] **全宽页面用 `maxWidth="full"`** — 仅适用于需要横向充分展开的复杂视图（如 rls-settings）
- [ ] **设置/详情页面** — 可用 `maxWidth="5xl"` 体现其聚焦性质；一般内容页不得低于 `7xl`
- [ ] **`PageLayout` 默认值为 `7xl`** — 新增页面无特殊需求不传 maxWidth 即可

---

## 布局 / 分割

- [ ] **不同功能区之间加 `border-t border-border` 分割线** — 例如侧边栏 DB 区域与模型列表区域之间
- [ ] **侧边栏操作按钮**：`variant="outline"`，不用 `bg-primary`；侧边栏里 primary 蓝色按钮会遮挡列表条目视觉焦点
- [ ] **并排按钮 vs 纵向按钮**：操作数量 ≤ 2 且语义平行时可横排；语义有主次或数量 ≥ 2 时纵向排列，`w-full justify-start`

---

## 图标 / 按钮

- [ ] **`⋮` 更多按钮**：默认 `opacity-0`，hover row 时显现（`group-hover:opacity-60`），自身 hover 时 `opacity-100`；不用持久可见
- [ ] **侧边栏左侧竖条选中指示器（`w-0.5 bg-foreground absolute left-0`）**：与 `rounded-md` 同用时会被裁切，不用此模式
- [ ] **导入按钮 `strokeWidth={1.5}`**：Download icon 用细线权重，与其他 icon 区分层级

---

## 空状态

- [ ] **未选择数据库时的空态图标**：用 `Database` icon，不用 `Table2`——语义对应

---

## Lint

- [ ] **lucide-react 导入的 icon 必须实际用到** — 删除未使用的 import（如 `Filter`）
- [ ] **`border-slate-200` 硬码** — 改为 `border-border`，跟随主题
- [ ] **`shadow-lg` + `border-slate-200`** — dropdown/popover 统一用 `border border-border shadow-lg`
