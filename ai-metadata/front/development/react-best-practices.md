# React 最佳实践

本文档提供 React 开发的最佳实践和常见模式。

## UI 组件规范

### 强制使用 shadcn/ui 组件

**重要**: 项目使用 shadcn/ui 作为基础 UI 组件库,所有基础 UI 元素必须使用 `@/components/ui` 中的组件。

```typescript
// ✅ 必须这样做
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog } from '@/components/ui/dialog'

// ❌ 禁止使用原生元素实现 UI
<button>...</button>
<input />
<div className="dialog">...</div>
```

**理由**:
- 保证 UI 一致性
- 自动适配设计系统(颜色、字体、间距)
- 支持主题切换
- 可访问性(Accessibility)已内置
- 减少重复代码

详见: [代码规范 - UI 组件规范](./code-conventions.md#ui-组件规范)

## 组件设计原则

### 1. 单一职责原则

每个组件应该只做一件事。

```typescript
// ❌ 避免:一个组件做太多事
function UserDashboard() {
  const [user, setUser] = useState(null)
  const [orders, setOrders] = useState([])
  const [notifications, setNotifications] = useState([])
  
  // 获取用户数据
  useEffect(() => { /* ... */ }, [])
  // 获取订单数据
  useEffect(() => { /* ... */ }, [])
  // 获取通知数据
  useEffect(() => { /* ... */ }, [])
  
  // 渲染用户信息
  // 渲染订单列表
  // 渲染通知
  // 处理各种交互
  
  return (
    // 几百行 JSX...
  )
}

// ✅ 推荐:拆分成小组件
function UserDashboard() {
  return (
    <div className="grid gap-6">
      <UserProfile />
      <RecentOrders />
      <NotificationPanel />
    </div>
  )
}

function UserProfile() {
  const { user, loading } = useUser()
  
  if (loading) return <Skeleton />
  
  return (
    <Card>
      <CardHeader>
        <Avatar src={user.avatar} />
        <h2>{user.name}</h2>
      </CardHeader>
    </Card>
  )
}
```

### 2. 组件组合优于继承

使用组合模式而不是继承来复用代码。

```typescript
// ✅ 推荐:使用 children 组合
function Dialog({ title, children }: DialogProps) {
  return (
    <div className="dialog">
      <div className="dialog-header">{title}</div>
      <div className="dialog-content">{children}</div>
    </div>
  )
}

// 使用
<Dialog title="确认删除">
  <p>确定要删除这条记录吗?</p>
  <Button>确认</Button>
</Dialog>

// ✅ 推荐:使用命名插槽
interface CardProps {
  header: React.ReactNode
  footer?: React.ReactNode
  children: React.ReactNode
}

function Card({ header, footer, children }: CardProps) {
  return (
    <div className="card">
      <div className="card-header">{header}</div>
      <div className="card-body">{children}</div>
      {footer && <div className="card-footer">{footer}</div>}
    </div>
  )
}
```

### 3. 保持组件纯净

相同的 props 应该产生相同的输出。

```typescript
// ❌ 避免:依赖外部可变状态
let globalCounter = 0

function BadComponent({ name }: Props) {
  globalCounter++ // 副作用!
  return <div>{name} - {globalCounter}</div>
}

// ✅ 推荐:纯组件
function GoodComponent({ name, count }: Props) {
  return <div>{name} - {count}</div>
}
```

## State 管理

### 1. 状态放置原则

将状态放在最合适的位置。

```typescript
// ✅ 本地状态:仅在一个组件中使用
function SearchInput() {
  const [query, setQuery] = useState('')
  
  return (
    <input 
      value={query} 
      onChange={(e) => setQuery(e.target.value)} 
    />
  )
}

// ✅ 提升状态:多个组件需要共享
function SearchPage() {
  const [query, setQuery] = useState('')
  
  return (
    <>
      <SearchInput value={query} onChange={setQuery} />
      <SearchResults query={query} />
    </>
  )
}

// ✅ 全局状态:跨页面/深层组件需要
// stores/use-user-store.ts
export const useUserStore = create<UserStore>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
}))
```

### 2. 避免过度使用 State

不是所有数据都需要 state。

```typescript
// ❌ 避免:派生状态应该计算而不是存储
function UserList({ users }: Props) {
  const [userCount, setUserCount] = useState(users.length)
  
  useEffect(() => {
    setUserCount(users.length)
  }, [users])
  
  return <div>用户数:{userCount}</div>
}

// ✅ 推荐:直接计算
function UserList({ users }: Props) {
  const userCount = users.length
  
  return <div>用户数:{userCount}</div>
}

// ❌ 避免:可以从 props 派生的状态
function SearchResults({ query }: Props) {
  const [lowercaseQuery, setLowercaseQuery] = useState(query.toLowerCase())
  
  useEffect(() => {
    setLowercaseQuery(query.toLowerCase())
  }, [query])
  
  // ...
}

// ✅ 推荐:直接计算
function SearchResults({ query }: Props) {
  const lowercaseQuery = query.toLowerCase()
  // ...
}
```

### 3. 状态更新最佳实践

```typescript
// ✅ 使用函数式更新(依赖前一个状态)
setCount(prev => prev + 1)

// ✅ 批量更新会自动合并
function handleClick() {
  setCount(count + 1)
  setName('John')
  setIsActive(true)
  // React 会批量处理这些更新
}

// ✅ 对象状态使用扩展运算符
setUser(prev => ({
  ...prev,
  name: 'New Name'
}))

// ❌ 避免:直接修改状态
user.name = 'New Name' // 错误!
setUser(user)

// ✅ 复杂状态考虑使用 useReducer
const [state, dispatch] = useReducer(reducer, initialState)
```

## Hooks 使用

### 1. Hooks 调用规则

```typescript
// ✅ 必须在组件顶层调用
function Component() {
  const [count, setCount] = useState(0)
  const user = useUser()
  
  // ...
}

// ❌ 不能在条件语句中调用
function BadComponent() {
  if (condition) {
    const [count, setCount] = useState(0) // 错误!
  }
}

// ✅ 正确的条件逻辑
function GoodComponent() {
  const [count, setCount] = useState(0)
  
  if (condition) {
    // 使用 state,而不是声明 state
    setCount(1)
  }
}
```

### 2. useEffect 最佳实践

```typescript
// ✅ 明确依赖项
useEffect(() => {
  fetchUser(userId)
}, [userId])

// ✅ 清理副作用
useEffect(() => {
  const subscription = subscribe()
  
  return () => {
    subscription.unsubscribe()
  }
}, [])

// ✅ 避免在 useEffect 中定义函数
// ❌ 不推荐
useEffect(() => {
  function fetchData() {
    // ...
  }
  fetchData()
}, [dep1, dep2])

// ✅ 推荐:使用 useCallback
const fetchData = useCallback(() => {
  // ...
}, [dep1, dep2])

useEffect(() => {
  fetchData()
}, [fetchData])

// ✅ 或者:在 effect 外定义
function fetchData() {
  // ...
}

useEffect(() => {
  fetchData()
}, [dep1, dep2])
```

### 3. 自定义 Hooks

```typescript
// ✅ 提取可复用逻辑
function useLocalStorage<T>(key: string, initialValue: T) {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key)
      return item ? JSON.parse(item) : initialValue
    } catch (error) {
      console.error(error)
      return initialValue
    }
  })

  const setValue = (value: T | ((val: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value
      setStoredValue(valueToStore)
      window.localStorage.setItem(key, JSON.stringify(valueToStore))
    } catch (error) {
      console.error(error)
    }
  }

  return [storedValue, setValue] as const
}

// 使用
function Component() {
  const [name, setName] = useLocalStorage('name', '')
  
  return (
    <input 
      value={name} 
      onChange={(e) => setName(e.target.value)} 
    />
  )
}
```

## 性能优化

### 1. React.memo

```typescript
// ✅ 对昂贵的组件使用 memo
export const ExpensiveList = React.memo(function ExpensiveList({ items }: Props) {
  // 复杂的渲染逻辑
  return (
    <ul>
      {items.map(item => (
        <ExpensiveItem key={item.id} item={item} />
      ))}
    </ul>
  )
})

// ✅ 自定义比较函数
export const UserCard = React.memo(
  function UserCard({ user }: Props) {
    return <div>{user.name}</div>
  },
  (prevProps, nextProps) => {
    // 只在 user.id 改变时重新渲染
    return prevProps.user.id === nextProps.user.id
  }
)
```

### 2. useMemo 和 useCallback

```typescript
// ✅ useMemo:缓存计算结果
function TodoList({ todos, filter }: Props) {
  const filteredTodos = useMemo(() => {
    return todos.filter(todo => {
      if (filter === 'active') return !todo.completed
      if (filter === 'completed') return todo.completed
      return true
    })
  }, [todos, filter])
  
  return (
    <ul>
      {filteredTodos.map(todo => (
        <TodoItem key={todo.id} todo={todo} />
      ))}
    </ul>
  )
}

// ✅ useCallback:缓存函数引用
function TodoList({ todos }: Props) {
  const [filter, setFilter] = useState('all')
  
  // 避免子组件不必要的重新渲染
  const handleToggle = useCallback((id: string) => {
    toggleTodo(id)
  }, [])
  
  return (
    <ul>
      {todos.map(todo => (
        <TodoItem 
          key={todo.id} 
          todo={todo} 
          onToggle={handleToggle} 
        />
      ))}
    </ul>
  )
}
```

### 3. 虚拟化长列表

```typescript
// ✅ 使用虚拟化库处理大量数据
import { useVirtualizer } from '@tanstack/react-virtual'

function VirtualList({ items }: Props) {
  const parentRef = useRef<HTMLDivElement>(null)
  
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,
  })
  
  return (
    <div ref={parentRef} style={{ height: '500px', overflow: 'auto' }}>
      <div style={{ height: `${virtualizer.getTotalSize()}px` }}>
        {virtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.index}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualRow.size}px`,
              transform: `translateY(${virtualRow.start}px)`,
            }}
          >
            {items[virtualRow.index].name}
          </div>
        ))}
      </div>
    </div>
  )
}
```

## 错误处理

### 1. Error Boundary

```typescript
// 创建错误边界组件
'use client'

import { Component, ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: any) {
    console.error('Error caught by boundary:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="error-container">
          <h2>出错了</h2>
          <p>{this.state.error?.message}</p>
        </div>
      )
    }

    return this.props.children
  }
}

// 使用
<ErrorBoundary fallback={<ErrorPage />}>
  <App />
</ErrorBoundary>
```

### 2. 异步错误处理

```typescript
// ✅ 使用 try-catch
function UserProfile() {
  const [user, setUser] = useState(null)
  const [error, setError] = useState<Error | null>(null)
  const [loading, setLoading] = useState(true)
  
  useEffect(() => {
    async function fetchUser() {
      try {
        setLoading(true)
        const data = await fetch('/api/user').then(r => r.json())
        setUser(data)
        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'))
      } finally {
        setLoading(false)
      }
    }
    
    fetchUser()
  }, [])
  
  if (loading) return <Spinner />
  if (error) return <ErrorMessage error={error} />
  if (!user) return <NotFound />
  
  return <UserCard user={user} />
}
```

## 代码组织

### 1. 按功能组织

```
components/
├── model-editor/           # 模型编辑器功能
│   ├── ModelCanvas.tsx
│   ├── FieldList.tsx
│   ├── RelationshipPanel.tsx
│   └── use-model-editor.ts
├── auth/                   # 认证功能
│   ├── LoginForm.tsx
│   ├── RegisterForm.tsx
│   └── use-auth.ts
└── ui/                     # 通用 UI 组件
    ├── button.tsx
    ├── input.tsx
    └── dialog.tsx
```

### 2. 组件文件结构

```typescript
// ModelEditor.tsx

// 1. 导入
import React from 'react'
import { useModelEditor } from './use-model-editor'
import { FieldList } from './FieldList'
import { Button } from '@/components/ui/button'

// 2. 类型定义
interface ModelEditorProps {
  modelId: string
  onSave?: (data: ModelData) => void
}

interface ModelData {
  // ...
}

// 3. 常量
const DEFAULT_FIELD_TYPE = 'string'

// 4. 辅助函数(如果不能复用就放在文件内)
function validateFieldName(name: string): boolean {
  return /^[a-zA-Z_][a-zA-Z0-9_]*$/.test(name)
}

// 5. 主组件
export function ModelEditor({ modelId, onSave }: ModelEditorProps) {
  // Hooks
  const { model, updateField, addField } = useModelEditor(modelId)
  const [isEditing, setIsEditing] = React.useState(false)
  
  // 事件处理
  const handleSave = () => {
    onSave?.(model)
  }
  
  // 渲染
  return (
    <div className="model-editor">
      <FieldList fields={model.fields} onUpdate={updateField} />
      <Button onClick={handleSave}>保存</Button>
    </div>
  )
}

// 6. 子组件(如果只在这里使用)
function EditorHeader({ title }: { title: string }) {
  return <h2 className="font-semibold">{title}</h2>
}
```

## 相关文档

- [代码规范](./code-conventions.md)
- [TypeScript 指南](./typescript-guide.md)
- [ESLint 规则](./eslint-rules.md)
- [性能优化指南](./performance.md)
