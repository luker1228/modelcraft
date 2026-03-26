# 代码规范

本文档定义了 ModelCraft 前端项目的代码风格和约定。

## 目录结构规范

### 1. 项目结构

```
modelcraft-front/
├── src/
│   ├── app/                    # Next.js App Router 页面
│   │   ├── (authenticated)/    # 需要认证的路由组
│   │   ├── api/               # API 路由
│   │   └── layout.tsx         # 根布局
│   ├── components/            # React 组件
│   │   ├── ui/               # 基础 UI 组件(shadcn/ui)
│   │   ├── model-editor/     # 功能组件(模型编辑器)
│   │   └── layout/           # 布局组件
│   ├── lib/                  # 工具函数和配置
│   │   ├── apollo/          # Apollo Client 配置
│   │   ├── utils.ts         # 通用工具函数
│   │   └── typography.ts    # 字体工具
│   ├── hooks/               # 自定义 Hooks
│   ├── stores/              # Zustand 状态管理
│   ├── types/               # TypeScript 类型定义
│   └── styles/              # 全局样式
├── public/                  # 静态资源
├── ai-metadata/            # AI 辅助开发文档
│   ├── style/             # 设计系统文档
│   └── development/       # 开发规范文档
└── prototypes/            # 原型和设计稿
```

### 2. 文件命名

#### 组件文件

```
PascalCase.tsx       # React 组件
kebab-case.ts        # 工具函数、hooks
kebab-case.css       # 样式文件
```

**示例**:

```
✅ 正确
components/ui/Button.tsx
components/model-editor/FieldList.tsx
hooks/use-toggle.ts
lib/string-utils.ts

❌ 错误
components/ui/button.tsx          # 组件应使用 PascalCase
components/model-editor/fieldList.tsx
hooks/useToggle.ts               # hooks 文件名用 kebab-case
lib/stringUtils.ts               # 工具文件用 kebab-case
```

#### 特殊文件

```
layout.tsx           # Next.js 布局
page.tsx            # Next.js 页面
loading.tsx         # Next.js 加载状态
error.tsx           # Next.js 错误处理
not-found.tsx       # 404 页面
route.ts            # API 路由
```

## 命名规范

### 1. 变量和函数

```typescript
// ✅ 使用 camelCase
const userName = 'John'
const isActive = true
const itemCount = 10

function fetchUserData() { }
function calculateTotal() { }

// ❌ 避免
const user_name = 'John'        // 不要用 snake_case
const UserName = 'John'         // 变量不要用 PascalCase
const ITEM_COUNT = 10          // 不要用 UPPER_CASE(除非是常量)
```

### 2. 常量

```typescript
// ✅ 使用 UPPER_SNAKE_CASE
const API_BASE_URL = 'https://api.example.com'
const MAX_RETRY_COUNT = 3
const DEFAULT_PAGE_SIZE = 20

// ✅ 配置对象可以用 camelCase
const config = {
  apiBaseUrl: 'https://api.example.com',
  maxRetryCount: 3,
} as const
```

### 3. 组件

```typescript
// ✅ 使用 PascalCase
export function Button() { }
export function UserProfile() { }
export function ModelEditorPanel() { }

// ❌ 避免
export function button() { }
export function userProfile() { }
```

### 4. Interface 和 Type

```typescript
// ✅ 使用 PascalCase,Props 后缀
interface ButtonProps {
  variant?: 'default' | 'primary'
  children: React.ReactNode
}

interface User {
  id: string
  name: string
}

type CardVariant = 'default' | 'outlined' | 'filled'

// ❌ 避免
interface IButton { }           // 不要用 I 前缀
interface button_props { }      // 不要用 snake_case
interface propsButton { }       // 不要倒序
```

### 5. Hooks

```typescript
// ✅ 使用 use 前缀 + camelCase
function useUser() { }
function useToggle() { }
function useModelEditor() { }

// 文件名:kebab-case
// hooks/use-user.ts
// hooks/use-toggle.ts
```

### 6. 事件处理器

```typescript
// ✅ 使用 handle 前缀
function handleClick() { }
function handleSubmit() { }
function handleInputChange() { }

// ✅ Props 中使用 on 前缀
interface ButtonProps {
  onClick?: () => void
  onSubmit?: (data: FormData) => void
}
```

## 代码风格

### 1. 导入顺序

```typescript
// 1. React 和 Next.js
import React from 'react'
import { useRouter } from 'next/navigation'

// 2. 第三方库
import { gql, useQuery } from '@apollo/client'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

// 3. 内部模块(使用 @ 别名)
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useUser } from '@/hooks/use-user'

// 4. 相对导入
import { Header } from './Header'
import { Footer } from './Footer'

// 5. 样式和资源
import './styles.css'
```

### 2. 组件结构

```typescript
// ✅ 推荐的组件结构顺序
import React from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// 1. 类型定义
interface CardProps {
  title: string
  description?: string
  children: React.ReactNode
  className?: string
}

// 2. 常量(如果有)
const DEFAULT_TITLE = '默认标题'

// 3. 主组件
export function Card({ title, description, children, className }: CardProps) {
  // 3.1 Hooks
  const router = useRouter()
  const [isOpen, setIsOpen] = React.useState(false)

  // 3.2 派生状态和计算
  const hasDescription = Boolean(description)

  // 3.3 事件处理器
  const handleClick = () => {
    setIsOpen(!isOpen)
  }

  // 3.4 副作用
  React.useEffect(() => {
    // ...
  }, [])

  // 3.5 渲染
  return (
    <div className={cn('rounded-lg border p-4', className)}>
      <h3 className="font-semibold">{title}</h3>
      {hasDescription && (
        <p className="text-muted-foreground">{description}</p>
      )}
      {children}
    </div>
  )
}

// 4. 子组件(如果有)
function CardHeader({ children }: { children: React.ReactNode }) {
  return <div className="mb-2">{children}</div>
}
```

### 3. 条件渲染

```typescript
// ✅ 推荐:使用 && 运算符
{hasError && <ErrorMessage />}
{isLoading && <Spinner />}

// ✅ 推荐:三元运算符(两个分支)
{isLoggedIn ? <Dashboard /> : <Login />}

// ✅ 复杂条件:提取到变量
const showWarning = isLowStock && !isBackordered && hasActiveOrders

{showWarning && <WarningBanner />}

// ❌ 避免:复杂的内联条件
{isLowStock && !isBackordered && hasActiveOrders && (
  <WarningBanner />
)}
```

### 4. 列表渲染

```typescript
// ✅ 推荐:使用有意义的 key
{users.map((user) => (
  <UserCard key={user.id} user={user} />
))}

// ✅ 数组索引作为 key(仅在列表静态且不重排时)
{items.map((item, index) => (
  <li key={index}>{item}</li>
))}

// ❌ 避免:随机生成 key
{items.map((item) => (
  <li key={Math.random()}>{item}</li>
))}
```

### 5. 注释规范

```typescript
/**
 * 用户信息卡片组件
 * 
 * @param user - 用户对象
 * @param onEdit - 编辑回调函数
 */
export function UserCard({ user, onEdit }: UserCardProps) {
  // TODO: 添加头像上传功能
  
  // FIXME: 修复在移动端的显示问题
  
  // NOTE: 这个逻辑需要与后端保持一致
  const isVerified = user.emailVerified && user.phoneVerified
  
  return (
    // ...
  )
}

// ✅ 复杂逻辑需要注释说明
// 计算用户等级:基础分 + 活跃度分 + 贡献度分
const userLevel = Math.floor(
  (user.baseScore + user.activityScore * 2 + user.contributionScore * 3) / 100
)

// ❌ 避免:无意义的注释
// 设置 count 为 0
const count = 0
```

## UI 组件规范

### 必须使用 shadcn/ui 基础组件

**强制要求**: 所有基础 UI 组件必须使用 `@/components/ui` 中的 shadcn/ui 组件,不要自己实现或使用其他 UI 库。

#### 可用的 shadcn/ui 组件

```typescript
// ✅ 必须使用这些组件
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Toast, useToast } from '@/components/ui/toast'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import { Dropdown, DropdownMenu, DropdownMenuItem } from '@/components/ui/dropdown-menu'

// ❌ 不要自己实现这些组件
// ❌ 不要使用原生 HTML 元素(如 <button>, <input>)来实现 UI
```

#### 使用示例

```typescript
// ✅ 正确:使用 shadcn/ui Button
import { Button } from '@/components/ui/button'

function MyComponent() {
  return (
    <Button variant="default" size="md" onClick={handleClick}>
      点击我
    </Button>
  )
}

// ❌ 错误:自己实现按钮
function MyComponent() {
  return (
    <button className="px-4 py-2 bg-blue-500 rounded" onClick={handleClick}>
      点击我
    </button>
  )
}

// ✅ 正确:使用 shadcn/ui Dialog
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

function MyDialog({ open, onOpenChange }: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>对话框标题</DialogTitle>
        </DialogHeader>
        <div>对话框内容</div>
      </DialogContent>
    </Dialog>
  )
}

// ❌ 错误:自己实现对话框
function MyDialog({ open }: Props) {
  return (
    <div className="fixed inset-0 bg-black/50">
      <div className="bg-white rounded p-4">
        {/* ... */}
      </div>
    </div>
  )
}
```

#### 扩展 shadcn/ui 组件

如果需要定制功能,应该基于 shadcn/ui 组件扩展:

```typescript
// ✅ 正确:扩展 shadcn/ui 组件
import { Button } from '@/components/ui/button'
import { Loader2 } from 'lucide-react'

interface LoadingButtonProps extends React.ComponentProps<typeof Button> {
  loading?: boolean
}

export function LoadingButton({ loading, children, ...props }: LoadingButtonProps) {
  return (
    <Button {...props} disabled={loading || props.disabled}>
      {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
      {children}
    </Button>
  )
}

// 使用
<LoadingButton loading={isSubmitting} onClick={handleSubmit}>
  提交
</LoadingButton>
```

#### 表单组件

```typescript
// ✅ 正确:使用 shadcn/ui 表单组件
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

function LoginForm() {
  return (
    <div className="space-y-4">
      <div>
        <Label htmlFor="email">邮箱</Label>
        <Input id="email" type="email" placeholder="your@email.com" />
      </div>
      <div>
        <Label htmlFor="password">密码</Label>
        <Input id="password" type="password" />
      </div>
      <Button type="submit">登录</Button>
    </div>
  )
}

// ❌ 错误:使用原生 input
function LoginForm() {
  return (
    <form>
      <input type="email" className="border rounded px-2 py-1" />
      <input type="password" className="border rounded px-2 py-1" />
      <button type="submit" className="bg-blue-500 text-white px-4 py-2">
        登录
      </button>
    </form>
  )
}
```

#### 何时可以不使用 shadcn/ui

只有以下情况可以不使用 shadcn/ui:

1. **特殊业务组件** - 如模型编辑器、图表等领域特定组件
2. **布局容器** - 如 `<div>`, `<section>`, `<main>` 等纯布局元素
3. **shadcn/ui 不提供的组件** - 确认 shadcn/ui 没有对应组件时

```typescript
// ✅ 允许:布局元素
<div className="flex gap-4">
  <main className="flex-1">
    {/* 使用 shadcn/ui 组件 */}
  </main>
</div>

// ✅ 允许:特殊业务组件
function ModelEditor() {
  return (
    <div className="model-canvas">
      {/* 模型编辑器特有的 UI */}
    </div>
  )
}
```

## 最佳实践

### 1. 组件职责单一

```typescript
// ❌ 避免:组件功能过多
function UserDashboard() {
  // 处理用户信息
  // 处理订单列表
  // 处理支付逻辑
  // 处理通知
  // ...几百行代码
}

// ✅ 推荐:拆分成小组件
function UserDashboard() {
  return (
    <div>
      <UserProfile />
      <OrderList />
      <NotificationPanel />
    </div>
  )
}
```

### 2. 提取可复用逻辑

```typescript
// ❌ 避免:重复的逻辑
function ComponentA() {
  const [isOpen, setIsOpen] = useState(false)
  const toggle = () => setIsOpen(!isOpen)
  // ...
}

function ComponentB() {
  const [isOpen, setIsOpen] = useState(false)
  const toggle = () => setIsOpen(!isOpen)
  // ...
}

// ✅ 推荐:提取为自定义 Hook
function useToggle(initialValue = false) {
  const [isOpen, setIsOpen] = useState(initialValue)
  const toggle = () => setIsOpen(!isOpen)
  return { isOpen, toggle, setIsOpen }
}

function ComponentA() {
  const { isOpen, toggle } = useToggle()
  // ...
}
```

### 3. Props 解构

```typescript
// ✅ 推荐:解构 props
function Button({ variant, size, children, onClick }: ButtonProps) {
  return (
    <button onClick={onClick} className={getButtonClass(variant, size)}>
      {children}
    </button>
  )
}

// ❌ 避免:使用 props 对象
function Button(props: ButtonProps) {
  return (
    <button onClick={props.onClick}>
      {props.children}
    </button>
  )
}
```

### 4. 避免魔法数字/字符串

```typescript
// ❌ 避免
if (user.role === 'admin') { }
setTimeout(() => { }, 3000)

// ✅ 推荐
const USER_ROLE = {
  ADMIN: 'admin',
  USER: 'user',
  GUEST: 'guest',
} as const

const DEBOUNCE_DELAY = 3000

if (user.role === USER_ROLE.ADMIN) { }
setTimeout(() => { }, DEBOUNCE_DELAY)
```

### 5. 错误处理

```typescript
// ✅ 推荐:明确的错误处理
async function fetchUser(id: string) {
  try {
    const response = await fetch(`/api/users/${id}`)
    
    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.status}`)
    }
    
    return await response.json()
  } catch (error) {
    console.error('Error fetching user:', error)
    throw error // 或者返回默认值
  }
}

// ❌ 避免:忽略错误
async function fetchUser(id: string) {
  const response = await fetch(`/api/users/${id}`)
  return await response.json()
}
```

## 性能优化

### 1. 避免不必要的渲染

```typescript
// ✅ 使用 React.memo
export const ExpensiveComponent = React.memo(function ExpensiveComponent({ data }: Props) {
  // ...
})

// ✅ 使用 useMemo 缓存计算结果
const sortedItems = useMemo(() => {
  return items.sort((a, b) => a.name.localeCompare(b.name))
}, [items])

// ✅ 使用 useCallback 缓存函数
const handleClick = useCallback(() => {
  doSomething(id)
}, [id])
```

### 2. 懒加载

```typescript
// ✅ 动态导入大型组件
const HeavyComponent = dynamic(() => import('./HeavyComponent'), {
  loading: () => <Spinner />,
})

// ✅ 条件加载
function Dashboard() {
  const [showChart, setShowChart] = useState(false)
  
  return (
    <div>
      <button onClick={() => setShowChart(true)}>显示图表</button>
      {showChart && <HeavyComponent />}
    </div>
  )
}
```

## 相关文档

- [ESLint 规则](./eslint-rules.md)
- [TypeScript 指南](./typescript-guide.md)
- [React 最佳实践](./react-best-practices.md)
- [设计系统规范](../style/STYLE.md)
