# TypeScript 开发指南

本文档说明项目的 TypeScript 配置和最佳实践。

## TypeScript 配置

配置文件: `tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["dom", "dom.iterable", "esnext"],
    "strict": true,
    "jsx": "preserve",
    "moduleResolution": "bundler",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

## 核心配置说明

### 1. 严格模式

```json
{
  "strict": true
}
```

启用所有严格类型检查选项:

- `noImplicitAny`: 禁止隐式 any
- `strictNullChecks`: 严格空值检查
- `strictFunctionTypes`: 严格函数类型检查
- `strictBindCallApply`: 严格 bind/call/apply 检查
- `strictPropertyInitialization`: 严格属性初始化
- `noImplicitThis`: 禁止隐式 this
- `alwaysStrict`: 始终使用严格模式

### 2. 路径别名

```json
{
  "paths": {
    "@/*": ["./src/*"]
  }
}
```

允许使用 `@/` 作为 `src/` 的快捷方式:

```typescript
// ✅ 推荐
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// ❌ 避免:相对路径过长
import { Button } from '../../../components/ui/button'
```

### 3. 模块解析

```json
{
  "moduleResolution": "bundler"
}
```

使用现代打包工具(Next.js)的模块解析策略。

## TypeScript 最佳实践

### 1. 类型定义

#### 优先使用 interface

```typescript
// ✅ 推荐:组件 Props
interface ButtonProps {
  variant?: 'default' | 'destructive' | 'outline'
  size?: 'sm' | 'md' | 'lg'
  onClick?: () => void
  children: React.ReactNode
}

// ✅ 也可以:需要联合类型或交叉类型时使用 type
type CardVariant = 'default' | 'outlined' | 'filled'

type ButtonWithIcon = ButtonProps & {
  icon: React.ReactNode
}
```

#### 避免 any

```typescript
// ❌ 错误:使用 any
function processData(data: any) {
  return data.value
}

// ✅ 正确:使用具体类型
interface DataItem {
  value: string
  id: number
}

function processData(data: DataItem) {
  return data.value
}

// ✅ 正确:使用泛型
function processData<T extends { value: string }>(data: T) {
  return data.value
}
```

### 2. React 组件类型

#### 函数组件

```typescript
// ✅ 推荐:使用 interface + 函数声明
interface CardProps {
  title: string
  description?: string
  children: React.ReactNode
}

export function Card({ title, description, children }: CardProps) {
  return (
    <div className="rounded-lg border p-4">
      <h3 className="font-semibold">{title}</h3>
      {description && <p className="text-muted-foreground">{description}</p>}
      {children}
    </div>
  )
}

// ⚠️ 可以但不推荐:React.FC(已不是最佳实践)
export const Card: React.FC<CardProps> = ({ title, description, children }) => {
  // ...
}
```

#### 事件处理器

```typescript
// ✅ 推荐:明确的事件类型
interface FormProps {
  onSubmit: (data: FormData) => void
}

export function Form({ onSubmit }: FormProps) {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    onSubmit(formData)
  }

  return <form onSubmit={handleSubmit}>...</form>
}

// 常见事件类型:
// - React.MouseEvent<HTMLButtonElement>
// - React.ChangeEvent<HTMLInputElement>
// - React.KeyboardEvent<HTMLInputElement>
// - React.FormEvent<HTMLFormElement>
```

### 3. Hooks 类型

#### useState

```typescript
// ✅ 简单类型自动推断
const [count, setCount] = useState(0) // number
const [name, setName] = useState('') // string

// ✅ 复杂类型需要显式指定
interface User {
  id: number
  name: string
  email: string
}

const [user, setUser] = useState<User | null>(null)
```

#### useRef

```typescript
// ✅ DOM 引用
const inputRef = useRef<HTMLInputElement>(null)

// ✅ 可变值
const countRef = useRef<number>(0)

// 使用时检查 null
if (inputRef.current) {
  inputRef.current.focus()
}
```

#### 自定义 Hook

```typescript
// ✅ 明确返回类型
interface UseToggleReturn {
  isOpen: boolean
  open: () => void
  close: () => void
  toggle: () => void
}

export function useToggle(initialValue = false): UseToggleReturn {
  const [isOpen, setIsOpen] = useState(initialValue)

  return {
    isOpen,
    open: () => setIsOpen(true),
    close: () => setIsOpen(false),
    toggle: () => setIsOpen(prev => !prev),
  }
}
```

### 4. GraphQL 类型

项目使用 Apollo Client,GraphQL 查询应有明确类型:

```typescript
import { gql, useQuery } from '@apollo/client'

// ✅ 定义查询结果类型
interface GetModelsData {
  models: Array<{
    id: string
    name: string
    description: string
  }>
}

interface GetModelsVariables {
  clusterId: string
}

const GET_MODELS = gql`
  query GetModels($clusterId: ID!) {
    models(clusterId: $clusterId) {
      id
      name
      description
    }
  }
`

// ✅ 使用类型参数
export function useModels(clusterId: string) {
  const { data, loading, error } = useQuery<GetModelsData, GetModelsVariables>(
    GET_MODELS,
    { variables: { clusterId } }
  )

  return { models: data?.models ?? [], loading, error }
}
```

### 5. 类型断言和类型守卫

```typescript
// ⚠️ 类型断言:仅在确定时使用
const value = getValue() as string

// ✅ 更好:类型守卫
function isString(value: unknown): value is string {
  return typeof value === 'string'
}

const value = getValue()
if (isString(value)) {
  // 这里 value 是 string 类型
  console.log(value.toUpperCase())
}

// ✅ 非空断言:仅在确定不为 null 时使用
const element = document.getElementById('root')!
```

### 6. 泛型使用

```typescript
// ✅ 泛型函数
function identity<T>(value: T): T {
  return value
}

// ✅ 泛型组件
interface SelectProps<T> {
  options: T[]
  value: T
  onChange: (value: T) => void
  getLabel: (option: T) => string
}

export function Select<T>({ options, value, onChange, getLabel }: SelectProps<T>) {
  return (
    <select onChange={(e) => onChange(options[Number(e.target.value)])}>
      {options.map((option, index) => (
        <option key={index} value={index}>
          {getLabel(option)}
        </option>
      ))}
    </select>
  )
}
```

## 常见问题

### 1. 类型错误:"Object is possibly 'null'"

```typescript
// ❌ 错误
const element = document.getElementById('root')
element.innerHTML = 'Hello' // Error: Object is possibly 'null'

// ✅ 方案 1:可选链
element?.setAttribute('class', 'container')

// ✅ 方案 2:非空断言(确定元素存在时)
const element = document.getElementById('root')!
element.innerHTML = 'Hello'

// ✅ 方案 3:类型守卫
const element = document.getElementById('root')
if (element) {
  element.innerHTML = 'Hello'
}
```

### 2. 类型错误:"Parameter implicitly has an 'any' type"

```typescript
// ❌ 错误
const numbers = [1, 2, 3]
numbers.map(n => n * 2) // n 隐式为 any(配置了 noImplicitAny)

// ✅ 方案 1:TypeScript 通常能自动推断
const numbers = [1, 2, 3]
numbers.map((n: number) => n * 2)

// ✅ 方案 2:明确数组类型
const numbers: number[] = [1, 2, 3]
numbers.map(n => n * 2) // n 自动推断为 number
```

### 3. 第三方库缺少类型定义

```bash
# 安装类型定义包
npm install --save-dev @types/[library-name]

# 如果不存在类型定义包,创建声明文件
# 创建 src/types/[library-name].d.ts
declare module 'library-name' {
  export function someFunction(): void
}
```

## 编辑器配置

### VS Code 推荐设置

`.vscode/settings.json`:

```json
{
  "typescript.tsdk": "node_modules/typescript/lib",
  "typescript.enablePromptUseWorkspaceTsdk": true,
  "typescript.preferences.importModuleSpecifier": "non-relative",
  "typescript.preferences.quoteStyle": "single",
  "editor.formatOnSave": true
}
```

### 推荐扩展

- **TypeScript and JavaScript Language Features** (内置)
- **Pretty TypeScript Errors** - 更友好的错误提示

## 性能优化

### 类型检查加速

```bash
# 使用项目引用(如果项目分模块)
# tsconfig.json
{
  "references": [
    { "path": "./packages/shared" }
  ]
}

# 增量编译
{
  "compilerOptions": {
    "incremental": true
  }
}
```

## 相关文档

- [TypeScript 官方文档](https://www.typescriptlang.org/)
- [React TypeScript Cheatsheet](https://react-typescript-cheatsheet.netlify.app/)
- [代码规范](./code-conventions.md)
- [React 最佳实践](./react-best-practices.md)
