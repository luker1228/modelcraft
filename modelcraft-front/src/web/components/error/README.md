# GraphQL 错误处理系统

这个系统提供了完整的 GraphQL 错误处理和调试功能，包括错误弹窗、上下文信息展示和错误历史记录。

## 功能特性

### 1. 自动错误弹窗
- 自动捕获所有 GraphQL 错误（查询、变更、订阅）
- 显示详细的错误信息和上下文
- 支持网络错误处理
- 开发环境自动弹窗，生产环境可配置

### 2. 丰富的上下文信息
- 操作名称和类型
- 请求变量
- GraphQL 查询内容
- 错误位置（行号、列号）
- 错误路径
- 时间戳和用户代理
- 堆栈跟踪（如果可用）

### 3. 错误历史记录
- 保存最近 50 条错误记录
- 支持查看、复制和重新显示历史错误
- 错误分类和标签

### 4. 开发者工具
- 测试错误弹窗功能
- 模拟不同类型的错误
- 查看错误历史
- 仅在开发环境显示

## 使用方法

### 1. 基本使用（自动处理）

系统已经集成到 Apollo Client 中，会自动处理所有 GraphQL 错误：

```tsx
// 无需额外配置，错误会自动弹窗显示
const { data, loading, error } = useQuery(GET_PROJECTS)
```

### 2. 手动错误处理

在组件中使用 `useGraphQLErrorHandler` hook：

```tsx
import { useGraphQLErrorHandler } from '@/hooks/useGraphQLErrorHandler'

function MyComponent() {
  const { handleError, handleCustomError } = useGraphQLErrorHandler()

  const [createProject] = useMutation(CREATE_PROJECT, {
    onError: (error, { variables }) => {
      // 手动处理错误，显示详细信息
      handleError(error, 'CreateProject', 'mutation', variables)
    }
  })

  const handleCustomErrorExample = () => {
    // 显示自定义错误
    handleCustomError(
      "这是一个自定义错误",
      "CUSTOM_ERROR",
      { additionalInfo: "额外信息" }
    )
  }
}
```

### 3. 使用增强的 hooks

使用带错误处理的 hooks：

```tsx
import { useProjectsWithErrorHandling } from '@/hooks/useProjectsWithErrorHandling'

function ProjectsPage() {
  const {
    projects,
    loading,
    createProject,
    updateProject,
    deleteProject
  } = useProjectsWithErrorHandling()

  // 错误会自动处理并弹窗显示
  const handleCreate = async (input) => {
    try {
      await createProject(input)
    } catch (error) {
      // 错误已经通过弹窗显示了
      console.log('操作失败')
    }
  }
}
```

### 4. 错误弹窗组件

直接使用错误弹窗组件：

```tsx
import { GraphQLErrorDialog } from '@/components/error/GraphQLErrorDialog'
import { useErrorStore } from '@/stores/error'

function MyComponent() {
  const { isErrorDialogOpen, currentErrors, currentContext, hideErrorDialog } = useErrorStore()

  return (
    <GraphQLErrorDialog
      open={isErrorDialogOpen}
      onOpenChange={hideErrorDialog}
      errors={currentErrors}
      context={currentContext}
    />
  )
}
```

### 5. 错误历史查看

```tsx
import { ErrorHistoryDialog } from '@/components/error/ErrorHistoryDialog'
import { useErrorStore } from '@/stores/error'

function MyComponent() {
  const [showHistory, setShowHistory] = useState(false)
  const { errorHistory } = useErrorStore()

  return (
    <>
      <Button onClick={() => setShowHistory(true)}>
        查看错误历史 ({errorHistory.length})
      </Button>
      <ErrorHistoryDialog
        open={showHistory}
        onOpenChange={setShowHistory}
      />
    </>
  )
}
```

## 错误类型

系统支持以下错误类型的特殊处理：

- `GRAPHQL_VALIDATION_FAILED`: GraphQL 验证失败
- `BAD_USER_INPUT`: 用户输入错误
- `UNAUTHENTICATED`: 未认证
- `FORBIDDEN`: 权限不足
- `NETWORK_ERROR`: 网络错误
- `CUSTOM_ERROR`: 自定义错误

## 配置选项

### Apollo Client 配置

```tsx
// 在 apollo-wrapper.tsx 中已经配置了错误处理
const client = new ApolloClient({
  link: from([
    errorLink,  // 错误处理链接
    authLink,
    httpLink,
  ]),
  defaultOptions: {
    watchQuery: { errorPolicy: 'all' },
    query: { errorPolicy: 'all' },
    mutate: { errorPolicy: 'all' },
  },
})
```

### 环境配置

- 开发环境：自动显示所有错误弹窗
- 生产环境：可以通过配置控制是否显示错误弹窗

## 开发者工具

在开发环境中，页面右下角会显示开发者工具面板，提供：

- 测试不同类型的错误
- 查看错误历史记录
- 手动触发错误弹窗

## 最佳实践

1. **保持错误策略一致**: 使用 `errorPolicy: 'all'` 确保即使有错误也能获取部分数据
2. **提供有意义的操作名称**: 在 GraphQL 操作中使用清晰的操作名称
3. **合理使用自定义错误**: 对于业务逻辑错误，使用 `handleCustomError`
4. **定期清理错误历史**: 在生产环境中定期清理错误历史记录
5. **错误分类**: 根据错误类型提供不同的用户体验

## 故障排除

### 错误弹窗不显示
- 检查是否正确导入了 `ErrorProvider`
- 确认 Apollo Client 配置了错误链接
- 检查控制台是否有 JavaScript 错误

### 上下文信息缺失
- 确保 GraphQL 操作有正确的操作名称
- 检查变量是否正确传递给 mutation/query

### 性能问题
- 定期清理错误历史记录
- 在生产环境中禁用详细的错误信息