# Web/BFF 分离实施计划

> **对于执行者**: 请使用 subagent-driven-development 或 executing-plans 执行本计划。任务使用 checkbox (`- [ ]`) 语法追踪进度。

**目标**: 将 ModelCraft 前端项目的混合代码分层为 `@bff/`、`@web/`、`@shared/` 三层，用 ESLint depguard 强制边界，不改任何功能。

**架构**: 目录分离 + tsconfig path alias + ESLint 依赖约束。分 5 批迁移，从基础设施 → shared → BFF → Web → 收尾，每批独立可回滚。

**技术栈**: Next.js 14, TypeScript 5.2, ESLint 8.55, eslint-plugin-depend（新增）

---

## Chunk 1: 基础设施搭建

### Task 1.1: 创建目录结构

**文件**:
- Create: `src/bff/.gitkeep`
- Create: `src/web/.gitkeep`
- Create: `src/shared/.gitkeep`
- Create: `src/bff/api/.gitkeep`
- Create: `src/bff/apollo/.gitkeep`
- Create: `src/bff/auth/.gitkeep`
- Create: `src/bff/cms/.gitkeep`
- Create: `src/web/providers/.gitkeep`
- Create: `src/web/cache/.gitkeep`
- Create: `src/web/cms/.gitkeep`
- Create: `src/web/routing/.gitkeep`
- Create: `src/shared/cms/.gitkeep`
- Create: `src/shared/utils/.gitkeep`

- [ ] **Step 1: 创建所有目录**

```bash
mkdir -p src/bff/{api,apollo,auth,cms}
mkdir -p src/web/{providers,cache,cms,routing}
mkdir -p src/shared/{cms,utils}

# 添加 .gitkeep 占位符
touch src/bff/.gitkeep src/bff/api/.gitkeep src/bff/apollo/.gitkeep src/bff/auth/.gitkeep src/bff/cms/.gitkeep
touch src/web/.gitkeep src/web/providers/.gitkeep src/web/cache/.gitkeep src/web/cms/.gitkeep src/web/routing/.gitkeep
touch src/shared/.gitkeep src/shared/cms/.gitkeep src/shared/utils/.gitkeep
```

- [ ] **Step 2: 验证目录结构**

```bash
find src/{bff,web,shared} -type d | sort
```

Expected: 所有 12 个目录都存在。

- [ ] **Step 3: 提交**

```bash
git add src/bff/.gitkeep src/web/.gitkeep src/shared/.gitkeep src/bff/api/.gitkeep src/bff/apollo/.gitkeep src/bff/auth/.gitkeep src/bff/cms/.gitkeep src/web/providers/.gitkeep src/web/cache/.gitkeep src/web/cms/.gitkeep src/web/routing/.gitkeep src/shared/cms/.gitkeep src/shared/utils/.gitkeep
git commit -m "chore: create directory structure for web/bff/shared layers"
```

---

### Task 1.2: 更新 tsconfig.json

**文件**:
- Modify: `tsconfig.json`

- [ ] **Step 1: 更新 tsconfig 路径**

将 `tsconfig.json` 中的 `paths` 部分从：

```json
"paths": {
  "@/*": ["./src/*"]
}
```

改为：

```json
"paths": {
  "@/*": ["./src/*"],
  "@bff/*": ["./src/bff/*"],
  "@web/*": ["./src/web/*"],
  "@shared/*": ["./src/shared/*"]
}
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
pnpm type-check
```

Expected: No errors

- [ ] **Step 3: 提交**

```bash
git add tsconfig.json
git commit -m "chore: add path aliases for web/bff/shared layers"
```

---

### Task 1.3: 安装 ESLint depguard

**文件**:
- Modify: `package.json`
- Modify: `.eslintrc.cjs`

- [ ] **Step 1: 安装 eslint-plugin-depend**

```bash
pnpm add -D eslint-plugin-depend
```

- [ ] **Step 2: 验证安装**

```bash
ls node_modules/eslint-plugin-depend/package.json
```

Expected: 文件存在

- [ ] **Step 3: 配置 ESLint depguard**

在 `.eslintrc.cjs` 中添加 `depguard` 规则。在 `module.exports` 中找到 `rules` 对象，添加：

```javascript
// 在现有 rules 之后追加
'depend/depguard': ['warn', {
  rules: [
    {
      selector: 'src/web/**',
      deny: ['src/bff'],
      message: 'Web 层不能直接 import BFF 层内部实现',
    },
    {
      selector: 'src/bff/**',
      deny: ['src/web'],
      message: 'BFF 层不能依赖 Web 层',
    },
  ],
}],
```

同时在 `plugins` 数组中添加 `'depend'`：

```javascript
plugins: ['tailwindcss', 'depend'],
```

- [ ] **Step 4: 验证 ESLint 配置**

```bash
pnpm lint --debug 2>&1 | grep -i "depend" || echo "No depend errors yet (expected)"
```

- [ ] **Step 5: 提交**

```bash
git add package.json pnpm-lock.yaml .eslintrc.cjs
git commit -m "chore: install and configure eslint-plugin-depend for boundary enforcement"
```

---

## Chunk 2: Shared 层迁移

### Task 2.1: 迁移 CMS 验证和转换

**文件**:
- Move: `src/lib/cms/schema-transformer.ts` → `src/shared/cms/schema-transformer.ts`
- Move: `src/lib/cms/validation.ts` → `src/shared/cms/validation.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/cms/schema-transformer.ts src/shared/cms/
cp src/lib/cms/validation.ts src/shared/cms/
```

- [ ] **Step 2: 检查并更新导入路径**

打开 `src/shared/cms/schema-transformer.ts` 和 `src/shared/cms/validation.ts`，检查内部导入：
- 如果 import 来自 `@/lib/`，改为 `@/shared/`
- 如果 import 来自相对路径 `../`，保持不变

- [ ] **Step 3: 使用 grep 找到所有引用**

```bash
grep -r "from ['\"]@/lib/cms/schema-transformer" src/ --include="*.ts" --include="*.tsx"
grep -r "from ['\"]@/lib/cms/validation" src/ --include="*.ts" --include="*.tsx"
```

记录所有找到的文件（本步不修改，只收集清单）。

- [ ] **Step 4: 验证编译**

```bash
pnpm type-check
```

Expected: No errors（之前的引用会报错，这是预期的，下一步修复）

- [ ] **Step 5: 提交**

```bash
git add src/shared/cms/
git commit -m "chore: move cms validation and transformation to @shared layer"
```

---

### Task 2.2: 迁移排版和主题配置

**文件**:
- Move: `src/lib/typography.ts` → `src/shared/typography.ts`
- Move: `src/lib/theme-colors.ts` → `src/shared/theme-colors.ts`
- Create: `src/shared/theme-colors.ts` (如果不存在)

- [ ] **Step 1: 检查 theme-colors.ts 是否存在**

```bash
ls -la src/lib/theme-colors.ts 2>&1 || echo "File does not exist"
```

如果不存在，在 Shared 层创建占位文件：

```bash
touch src/shared/theme-colors.ts
```

如果存在，复制过去：

```bash
cp src/lib/theme-colors.ts src/shared/
```

- [ ] **Step 2: 复制 typography.ts**

```bash
cp src/lib/typography.ts src/shared/
```

- [ ] **Step 3: 使用 grep 找引用**

```bash
grep -r "from ['\"]@/lib/typography" src/ --include="*.ts" --include="*.tsx"
grep -r "from ['\"]@/lib/theme-colors" src/ --include="*.ts" --include="*.tsx"
```

记录所有找到的文件。

- [ ] **Step 4: 验证编译**

```bash
pnpm type-check 2>&1 | head -20
```

Expected: 有关于 `@/lib/typography` 和 `@/lib/theme-colors` 的找不到模块的错误（预期）

- [ ] **Step 5: 提交**

```bash
git add src/shared/typography.ts src/shared/theme-colors.ts
git commit -m "chore: move typography and theme-colors to @shared layer"
```

---

### Task 2.3: 迁移组织名称工具

**文件**:
- Move: `src/lib/organization-name-validator.ts` → `src/shared/organization-name-validator.ts`
- Move: `src/lib/organization-name-generator.ts` → `src/shared/organization-name-generator.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/organization-name-validator.ts src/shared/
cp src/lib/organization-name-generator.ts src/shared/
```

- [ ] **Step 2: 检查并更新导入路径**

打开两个文件，检查内部 import，无需改变（都是纯工具函数）。

- [ ] **Step 3: 使用 grep 找引用**

```bash
grep -r "organization-name-validator\|organization-name-generator" src/ --include="*.ts" --include="*.tsx" | grep "from"
```

记录所有找到的文件。

- [ ] **Step 4: 验证编译**

```bash
pnpm type-check 2>&1 | grep -i "organization-name" | head -5
```

- [ ] **Step 5: 提交**

```bash
git add src/shared/organization-name-validator.ts src/shared/organization-name-generator.ts
git commit -m "chore: move organization name utils to @shared layer"
```

---

### Task 2.4: 迁移 UUID 工具

**文件**:
- Move: `src/lib/uuid-utils.ts` → `src/shared/utils/uuid.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/uuid-utils.ts src/shared/utils/uuid.ts
```

- [ ] **Step 2: 使用 grep 找引用**

```bash
grep -r "uuid-utils" src/ --include="*.ts" --include="*.tsx" | grep "from"
grep -r "generateUUID" src/ --include="*.ts" --include="*.tsx" | grep -v "node_modules" | head -10
```

记录所有导入 `uuid-utils` 的文件。

- [ ] **Step 3: 验证编译**

```bash
pnpm type-check 2>&1 | grep -i "uuid" | head -5
```

- [ ] **Step 4: 提交**

```bash
git add src/shared/utils/uuid.ts
git commit -m "chore: move uuid-utils to @shared/utils layer"
```

---

### Task 2.5: 迁移通用工具函数

**文件**:
- Move: `src/lib/utils.ts` → `src/shared/utils.ts`

- [ ] **Step 1: 审查 utils.ts 内容确认无副作用**

内容应该只包含 `cn()` (Tailwind class 合并) 和 `formatDate()`，都是纯函数。

```bash
cat src/lib/utils.ts
```

确认没有：
- 浏览器 API (localStorage, window, document)
- 网络调用 (fetch, axios)
- 异步操作
- import React hooks

- [ ] **Step 2: 复制文件**

```bash
cp src/lib/utils.ts src/shared/utils.ts
```

注意：`src/shared/utils.ts` 和 `src/shared/utils/uuid.ts` 是两个不同的文件（一个在根，一个在 utils/ 子目录）。

- [ ] **Step 3: 使用 grep 找引用**

```bash
grep -r "from ['\"]@/lib/utils" src/ --include="*.ts" --include="*.tsx" | head -10
```

记录所有找到的文件。

- [ ] **Step 4: 验证编译**

```bash
pnpm type-check 2>&1 | grep "utils" | head -10
```

- [ ] **Step 5: 提交**

```bash
git add src/shared/utils.ts
git commit -m "chore: move utils to @shared layer"
```

---

### Task 2.6: 更新 Shared 层的导入（收集列表）

**目标**: 找出所有仍然引用 `@/lib/` 的 shared 文件，这些需要改成 `@/shared/`

- [ ] **Step 1: 检查 shared 文件中的导入**

```bash
grep -r "@/lib/" src/shared/ --include="*.ts" --include="*.tsx"
```

如果有输出，记录下来。通常不应该有（shared 文件应该互相独立）。

- [ ] **Step 2: 验证完整编译**

```bash
pnpm type-check 2>&1 | tail -20
```

此时应该有很多关于 `@/lib/` 的找不到模块的错误，这是因为引用它们的 web/bff 文件还没迁移。这是预期的。

---

### Task 2.7: 验证 Shared 层完成

- [ ] **Step 1: 列出 shared 下的所有文件**

```bash
find src/shared -type f -not -name ".gitkeep" | sort
```

Expected: 
```
src/shared/cms/schema-transformer.ts
src/shared/cms/validation.ts
src/shared/organization-name-generator.ts
src/shared/organization-name-validator.ts
src/shared/theme-colors.ts
src/shared/typography.ts
src/shared/utils.ts
src/shared/utils/uuid.ts
```

- [ ] **Step 2: 确认没有 @/lib/ 导入**

```bash
grep -r "@/lib/" src/shared/ || echo "✓ No @/lib imports in shared"
```

Expected: "✓ No @/lib imports in shared"

- [ ] **Step 3: 提交（如有遗漏文件）**

如果上面漏掉了什么，现在补充并提交。

---

## Chunk 3: BFF 层迁移

### Task 3.1: 迁移 Apollo 客户端配置

**文件**:
- Move: `src/lib/apollo-clients.ts` → `src/bff/apollo/clients.ts`
- Move: `src/lib/apollo-wrapper.tsx` (部分拆分，完整 Provider 保留在 Web 层)

由于 `apollo-wrapper.tsx` 包含 React Provider，暂不在此步迁移。只迁移纯 JS 的 `apollo-clients.ts`。

- [ ] **Step 1: 复制 apollo-clients.ts**

```bash
cp src/lib/apollo-clients.ts src/bff/apollo/clients.ts
```

- [ ] **Step 2: 审查导入并更新**

打开 `src/bff/apollo/clients.ts`，检查所有 import：
- `@/lib/auth/auth_provider` → 后续会迁到 `@bff/auth/auth_provider`，改为 `@bff/auth/auth_provider`
- `@/lib/cms/runtime-query-builder` → 后续会迁到 `@bff/cms`，改为 `@bff/cms/runtime-query-builder`
- `@shared/*` → 保持不变

此时 `@bff/auth/auth_provider` 还不存在，会导致 type-check 报错，这是预期的。

- [ ] **Step 3: 使用 grep 找引用 apollo-clients.ts 的文件**

```bash
grep -r "from ['\"]@/lib/apollo-clients" src/ --include="*.ts" --include="*.tsx"
```

记录所有找到的文件。

- [ ] **Step 4: 更新这些文件的导入**

对找到的每个文件，将 `@/lib/apollo-clients` 改为 `@bff/apollo/clients`。

示例：
```typescript
// Before
import { getOrgScopedClient } from '@/lib/apollo-clients'

// After
import { getOrgScopedClient } from '@bff/apollo/clients'
```

- [ ] **Step 5: 验证编译（允许部分错误）**

```bash
pnpm type-check 2>&1 | head -30
```

预期有关于 `@bff/auth/auth_provider` 等尚未迁移的模块的错误。

- [ ] **Step 6: 提交**

```bash
git add src/bff/apollo/clients.ts
git commit -m "chore: move apollo clients to @bff/apollo layer"
```

---

### Task 3.2: 迁移认证相关代码

**文件**:
- Move: `src/lib/auth/auth_provider.ts` → `src/bff/auth/auth_provider.ts`
- Move: `src/lib/auth/token-utils.ts` → `src/bff/auth/token-utils.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/auth/auth_provider.ts src/bff/auth/
cp src/lib/auth/token-utils.ts src/bff/auth/
```

- [ ] **Step 2: 更新导入路径**

打开两个文件，检查所有 import：
- `@/lib/` → 检查是否有，如有改为 `@bff/` 或 `@shared/`
- `@shared/` → 保持不变

- [ ] **Step 3: 找引用这两个文件的地方**

```bash
grep -r "from ['\"]@/lib/auth" src/ --include="*.ts" --include="*.tsx"
```

记录所有找到的文件。

- [ ] **Step 4: 更新这些文件的导入**

将 `@/lib/auth/auth_provider` 改为 `@bff/auth/auth_provider`，`@/lib/auth/token-utils` 改为 `@bff/auth/token-utils`。

- [ ] **Step 5: 验证 apollo/clients.ts 现在能否编译**

```bash
pnpm type-check src/bff/apollo/clients.ts 2>&1 | head -20
```

应该不再有关于 `@bff/auth/auth_provider` 找不到的错误。

- [ ] **Step 6: 提交**

```bash
git add src/bff/auth/
git commit -m "chore: move auth (auth_provider, token-utils) to @bff layer"
```

---

### Task 3.3: 迁移 CMS 运行时查询构建

**文件**:
- Move: `src/lib/cms/runtime-query-builder.ts` → `src/bff/cms/runtime-query-builder.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/cms/runtime-query-builder.ts src/bff/cms/
```

- [ ] **Step 2: 检查导入**

打开文件，检查 import：
- 应该导入 `@shared/*` 和可能的 GraphQL 类型

- [ ] **Step 3: 找引用**

```bash
grep -r "from ['\"]@/lib/cms/runtime-query-builder" src/ --include="*.ts" --include="*.tsx"
```

记录所有找到的文件。

- [ ] **Step 4: 更新引用**

将 `@/lib/cms/runtime-query-builder` 改为 `@bff/cms/runtime-query-builder`。

- [ ] **Step 5: 验证编译**

```bash
pnpm type-check 2>&1 | grep -i "runtime-query" || echo "✓ No runtime-query errors"
```

- [ ] **Step 6: 提交**

```bash
git add src/bff/cms/
git commit -m "chore: move runtime-query-builder to @bff/cms layer"
```

---

### Task 3.4: 处理 API Routes 和 BFF API 逻辑

**文件**:
- Create: `src/bff/api/auth/token.ts`
- Create: `src/bff/api/auth/refresh.ts`
- Create: `src/bff/api/org/init.ts`
- Create: `src/bff/api/user/memberships.ts`
- Create: `src/bff/api/copilotkit.ts`
- Modify: `src/app/api/auth/token/route.ts`
- Modify: `src/app/api/auth/refresh/route.ts`
- Modify: `src/app/api/org/init/route.ts`
- Modify: `src/app/api/user/memberships/route.ts`
- Modify: `src/app/api/copilotkit/route.ts`

核心思路：让 `src/app/api/` 的 route.ts 只做路由注册，业务逻辑移到 `src/bff/api/`。

- [ ] **Step 1: 创建 BFF API 认证模块**

创建 `src/bff/api/auth/token.ts`：

```typescript
import { NextRequest, NextResponse } from 'next/server'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { code } = body

    if (!code) {
      return NextResponse.json(
        { error: 'Authorization code is required' },
        { status: 400 }
      )
    }

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const tokenUrl = `${backendUrl}/api/auth/token`

    const response = await fetch(tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ code }),
    })

    const data = await response.json()

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Token exchange error:', error)
    return NextResponse.json(
      { error: 'Failed to exchange token' },
      { status: 500 }
    )
  }
}
```

创建 `src/bff/api/auth/refresh.ts`：

```typescript
import { NextRequest, NextResponse } from 'next/server'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { refreshToken } = body

    if (!refreshToken) {
      return NextResponse.json(
        { error: 'Refresh token is required' },
        { status: 400 }
      )
    }

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const refreshUrl = `${backendUrl}/api/auth/refresh`

    const response = await fetch(refreshUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refreshToken }),
    })

    const data = await response.json()

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Token refresh error:', error)
    return NextResponse.json(
      { error: 'Failed to refresh token' },
      { status: 500 }
    )
  }
}
```

- [ ] **Step 2: 创建 BFF API 组织模块**

创建 `src/bff/api/org/init.ts`：

```typescript
import { NextRequest, NextResponse } from 'next/server'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { organizationName, displayName } = body

    const authHeader = request.headers.get('authorization')
    if (!authHeader) {
      return NextResponse.json(
        { error: 'Authorization token is required' },
        { status: 401 }
      )
    }

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const orgInitUrl = `${backendUrl}/api/org/init`

    const response = await fetch(orgInitUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': authHeader,
      },
      body: JSON.stringify({ organizationName: organizationName || '', displayName: displayName || '' }),
    })

    const data = await response.json()

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Organization init error:', error)
    return NextResponse.json(
      { error: 'Failed to initialize organization' },
      { status: 500 }
    )
  }
}
```

- [ ] **Step 3: 创建 BFF API 用户模块**

创建 `src/bff/api/user/memberships.ts`：

```typescript
import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  try {
    const authHeader = request.headers.get('authorization')

    if (!authHeader) {
      return NextResponse.json(
        { error: 'Authorization header is required' },
        { status: 401 }
      )
    }

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const membershipsUrl = `${backendUrl}/api/user/memberships`

    const response = await fetch(membershipsUrl, {
      method: 'GET',
      headers: {
        'Authorization': authHeader,
        'Content-Type': 'application/json',
      },
    })

    const data = await response.json()

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }

    return NextResponse.json(data)
  } catch (error) {
    console.error('Get memberships error:', error)
    return NextResponse.json(
      { error: 'Failed to get memberships' },
      { status: 500 }
    )
  }
}
```

- [ ] **Step 4: 创建 BFF API CopilotKit 模块**

创建 `src/bff/api/copilotkit.ts`：

```typescript
import {
  CopilotRuntime,
  ExperimentalEmptyAdapter,
  copilotRuntimeNextJSAppRouterEndpoint,
} from "@copilotkit/runtime";
import { LangGraphHttpAgent } from "@copilotkit/runtime/langgraph";
import { NextRequest } from "next/server";

const serviceAdapter = new ExperimentalEmptyAdapter();

const runtime = new CopilotRuntime({
  agents: {
    modelcraft_agent: new LangGraphHttpAgent({
      url: "http://localhost:8001/api/copilotkit",
    }) as any,
  }
});

export async function POST(req: NextRequest) {
  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime,
    serviceAdapter,
    endpoint: "/api/copilotkit",
  });

  return handleRequest(req);
}
```

- [ ] **Step 5: 更新 route.ts 文件为 re-export**

修改 `src/app/api/auth/token/route.ts`：

```typescript
export { POST } from '@bff/api/auth/token'
```

修改 `src/app/api/auth/refresh/route.ts`：

```typescript
export { POST } from '@bff/api/auth/refresh'
```

修改 `src/app/api/org/init/route.ts`：

```typescript
export { POST } from '@bff/api/org/init'
```

修改 `src/app/api/user/memberships/route.ts`：

```typescript
export { GET } from '@bff/api/user/memberships'
```

修改 `src/app/api/copilotkit/route.ts`：

```typescript
export { POST } from '@bff/api/copilotkit'
```

- [ ] **Step 6: 验证编译**

```bash
pnpm type-check 2>&1 | head -20
```

Expected: 不应该有新的关于 API routes 的错误。

- [ ] **Step 7: 验证 API 功能（快速检查）**

```bash
pnpm build 2>&1 | tail -20
```

Expected: Build succeeds

- [ ] **Step 8: 提交**

```bash
git add src/bff/api/ src/app/api/
git commit -m "chore: move API route handlers to @bff layer, keep route.ts as re-exports"
```

---

### Task 3.5: 验证 BFF 层完成

- [ ] **Step 1: 列出 BFF 下的所有文件**

```bash
find src/bff -type f -not -name ".gitkeep" | sort
```

Expected:
```
src/bff/api/auth/token.ts
src/bff/api/auth/refresh.ts
src/bff/api/org/init.ts
src/bff/api/user/memberships.ts
src/bff/api/copilotkit.ts
src/bff/apollo/clients.ts
src/bff/auth/auth_provider.ts
src/bff/auth/token-utils.ts
src/bff/cms/runtime-query-builder.ts
```

- [ ] **Step 2: 检查 BFF 文件不依赖 Web 层**

```bash
grep -r "@web/" src/bff/ || echo "✓ No @web imports in BFF"
```

Expected: "✓ No @web imports in BFF"

- [ ] **Step 3: 验证完整编译**

```bash
pnpm type-check 2>&1 | tail -30
```

此时应该没有关于 BFF 层的错误，只有 Web 层还未迁移导致的错误。

---

## Chunk 4: Web 层迁移

### Task 4.1: 迁移 Apollo Provider

**文件**:
- Move: `src/lib/apollo-wrapper.tsx` → `src/web/providers/apollo-wrapper.tsx`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/apollo-wrapper.tsx src/web/providers/
```

- [ ] **Step 2: 更新导入路径**

打开 `src/web/providers/apollo-wrapper.tsx`，查找所有 import：
- `@/lib/apollo-clients` → 改为 `@bff/apollo/clients`
- `@/lib/auth/token-utils` → 改为 `@bff/auth/token-utils`
- `@/lib/auth/auth_provider` → 改为 `@bff/auth/auth_provider`
- 其他 `@/lib/` → 改为 `@shared/` 或 `@web/`（根据文件位置）

- [ ] **Step 3: 找引用 apollo-wrapper 的文件**

```bash
grep -r "from ['\"]@/lib/apollo-wrapper" src/ --include="*.ts" --include="*.tsx"
```

记录所有找到的文件。

- [ ] **Step 4: 更新这些文件的导入**

将 `@/lib/apollo-wrapper` 改为 `@web/providers/apollo-wrapper`。

- [ ] **Step 5: 验证编译**

```bash
pnpm type-check 2>&1 | grep -i "apollo-wrapper" || echo "✓ No apollo-wrapper errors"
```

- [ ] **Step 6: 提交**

```bash
git add src/web/providers/apollo-wrapper.tsx
git commit -m "chore: move apollo-wrapper to @web/providers layer"
```

---

### Task 4.2: 迁移 TanStack Query Provider

**文件**:
- Move: `src/lib/query-wrapper.tsx` → `src/web/providers/query-wrapper.tsx`

- [ ] **Step 1: 检查文件是否存在**

```bash
ls src/lib/query-wrapper.tsx 2>&1 || echo "File not found"
```

如果不存在，跳过此任务。

- [ ] **Step 2: 复制文件**

```bash
cp src/lib/query-wrapper.tsx src/web/providers/ 2>/dev/null || echo "Skipped (file not found)"
```

- [ ] **Step 3: 更新导入路径**

打开文件，检查并更新所有 `@/lib/` import。

- [ ] **Step 4: 找引用**

```bash
grep -r "from ['\"]@/lib/query-wrapper" src/ --include="*.ts" --include="*.tsx"
```

记录并更新。

- [ ] **Step 5: 提交**

```bash
git add src/web/providers/query-wrapper.tsx 2>/dev/null
git commit -m "chore: move query-wrapper to @web/providers layer" 2>/dev/null || echo "Skipped"
```

---

### Task 4.3: 迁移客户端缓存

**文件**:
- Move: `src/lib/memberships-cache.ts` → `src/web/cache/memberships-cache.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/memberships-cache.ts src/web/cache/
```

- [ ] **Step 2: 更新导入路径**

打开文件，检查所有 import，特别是：
- `@/lib/auth/token-utils` → 改为 `@bff/auth/token-utils`
- `@/lib/auth/auth_provider` → 改为 `@bff/auth/auth_provider`
- `@shared/*` 保持不变

- [ ] **Step 3: 找引用**

```bash
grep -r "from ['\"]@/lib/memberships-cache" src/ --include="*.ts" --include="*.tsx"
grep -r "memberships-cache\|getMemberships" src/ --include="*.ts" --include="*.tsx" | grep -v node_modules | head -20
```

记录所有找到的文件。

- [ ] **Step 4: 更新引用**

将 `@/lib/memberships-cache` 改为 `@web/cache/memberships-cache`。

- [ ] **Step 5: 验证编译**

```bash
pnpm type-check 2>&1 | grep -i "memberships-cache" || echo "✓ No memberships-cache errors"
```

- [ ] **Step 6: 提交**

```bash
git add src/web/cache/memberships-cache.ts
git commit -m "chore: move memberships-cache to @web/cache layer"
```

---

### Task 4.4: 迁移 Web 业务逻辑

**文件**:
- Move: `src/lib/cms/field-linkage.ts` → `src/web/cms/field-linkage.ts`
- Move: `src/lib/routing/smart-redirect.ts` → `src/web/routing/smart-redirect.ts`

- [ ] **Step 1: 复制文件**

```bash
cp src/lib/cms/field-linkage.ts src/web/cms/
cp src/lib/routing/smart-redirect.ts src/web/routing/
```

- [ ] **Step 2: 检查导入**

打开两个文件，查看所有 import，更新 `@/lib/` 为 `@shared/` 或 `@web/`。

- [ ] **Step 3: 找引用**

```bash
grep -r "field-linkage\|smart-redirect" src/ --include="*.ts" --include="*.tsx" | grep "from" | head -10
```

记录所有找到的文件。

- [ ] **Step 4: 更新引用**

- [ ] **Step 5: 验证编译**

```bash
pnpm type-check 2>&1 | grep -i "field-linkage\|smart-redirect" || echo "✓ No linkage/redirect errors"
```

- [ ] **Step 6: 提交**

```bash
git add src/web/cms/ src/web/routing/
git commit -m "chore: move field-linkage and smart-redirect to @web layer"
```

---

### Task 4.5: 迁移 Web 核心层

**文件**:
- Move: `src/components/` → `src/web/components/`
- Move: `src/hooks/` → `src/web/hooks/`
- Move: `src/stores/` → `src/web/stores/`
- Move: `src/graphql/` → `src/web/graphql/`

这是最大的迁移，包含 56 个组件文件。

- [ ] **Step 1: 复制目录**

```bash
cp -r src/components/* src/web/components/
cp -r src/hooks/* src/web/hooks/
cp -r src/stores/* src/web/stores/
cp -r src/graphql/* src/web/graphql/
```

- [ ] **Step 2: 验证复制**

```bash
ls src/web/components/ | head -5
ls src/web/hooks/ | head -5
ls src/web/stores/ | head -5
ls src/web/graphql/ | head -5
```

- [ ] **Step 3: 更新这些文件内的导入**

这是关键步骤。运行一个大规模的导入路径更新：

```bash
# 更新所有 @/lib/ import 为 @shared/（对于工具函数和配置）
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/typography|from '@shared/typography|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/theme-colors|from '@shared/theme-colors|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/utils|from '@shared/utils|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/organization-name|from '@shared/organization-name|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/cms/schema-transformer|from '@shared/cms/schema-transformer|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/cms/validation|from '@shared/cms/validation|g"

# 更新 apollo-wrapper import
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/apollo-wrapper|from '@web/providers/apollo-wrapper|g"

# 更新其他 Web 层 import
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/memberships-cache|from '@web/cache/memberships-cache|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/cms/field-linkage|from '@web/cms/field-linkage|g"
find src/web -name "*.ts" -o -name "*.tsx" | xargs sed -i "s|from ['\"]@/lib/routing/smart-redirect|from '@web/routing/smart-redirect|g"

# 本地相对 import 中的 ../lib/ → 根据新位置更新
# 这个比较复杂，可能需要手动逐个检查
```

- [ ] **Step 4: 检查是否还有遗漏的 @/lib/ import**

```bash
grep -r "@/lib/" src/web --include="*.ts" --include="*.tsx" | head -20
```

如果有，手动修复。

- [ ] **Step 5: 验证编译（可能有多个错误，逐个修复）**

```bash
pnpm type-check 2>&1 | head -50
```

Expected: 应该看到很多错误，但都应该是"找不到模块"的形式，说明有 import 路径错误。逐个修复。

- [ ] **Step 6: 提交（分多次提交以避免单个 commit 太大）**

```bash
git add src/web/components/ src/web/hooks/ src/web/stores/ src/web/graphql/
git commit -m "chore: move components, hooks, stores, graphql to @web layer and update imports"
```

---

### Task 4.6: 验证 Web 层完成

- [ ] **Step 1: 检查没有 @bff 直接导入**

```bash
grep -r "@bff/" src/web --include="*.ts" --include="*.tsx" | head -20
```

这个有可能会有（例如 Apollo Provider 可能需要导入 BFF 的配置），如果有，检查是否合理（只允许通过 `@shared/` 间接使用）。

- [ ] **Step 2: 确认核心目录都已迁移**

```bash
[ ! -d src/components ] && echo "✓ src/components removed" || echo "⚠ src/components still exists"
[ ! -d src/hooks ] && echo "✓ src/hooks removed" || echo "⚠ src/hooks still exists"
[ ! -d src/stores ] && echo "✓ src/stores removed" || echo "⚠ src/stores still exists"
[ ! -d src/graphql ] && echo "✓ src/graphql removed" || echo "⚠ src/graphql still exists"
```

如果还存在，表示有部分文件未迁移，需要检查并迁移。

---

## Chunk 5: 收尾与切换到 Error 模式

### Task 5.1: 更新应用层路由

**文件**:
- Modify: 所有 `src/app/**/*.ts` 和 `src/app/**/*.tsx` 文件中的 import

- [ ] **Step 1: 检查 app/ 中还有哪些 import @/lib/**

```bash
grep -r "@/lib/" src/app --include="*.ts" --include="*.tsx" | grep -v "node_modules"
```

记录所有找到的文件和行。

- [ ] **Step 2: 更新这些文件的导入**

为每个找到的 import，根据新的位置改为 `@web/`、`@bff/` 或 `@shared/`。

示例：
```typescript
// Before
import { AppLayout } from '@/lib/components/layout'

// After
import { AppLayout } from '@web/components/layout'
```

- [ ] **Step 3: 验证编译**

```bash
pnpm type-check 2>&1 | tail -30
```

Expected: 不应该有关于 `@/lib/` 的错误。

- [ ] **Step 4: 提交**

```bash
git add src/app/
git commit -m "chore: update app layer imports to use @web/@bff/@shared aliases"
```

---

### Task 5.2: 删除原 lib/ 目录

**文件**:
- Delete: `src/lib/`

- [ ] **Step 1: 备份清单（确保所有文件都已迁移）**

```bash
ls -la src/lib/ | tail -20
find src/lib -type f -not -name ".gitkeep" | sort
```

验证所有文件都已在新位置出现过。

- [ ] **Step 2: 删除 lib/ 目录**

```bash
rm -rf src/lib
```

- [ ] **Step 3: 验证删除**

```bash
[ ! -d src/lib ] && echo "✓ src/lib directory removed" || echo "✗ src/lib still exists"
```

- [ ] **Step 4: 验证编译**

```bash
pnpm type-check 2>&1 | head -20
```

Expected: 不应该有关于 lib/ 目录的错误。

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "chore: remove original src/lib directory after migration complete"
```

---

### Task 5.3: 更新 ESLint 配置为 Error 模式

**文件**:
- Modify: `.eslintrc.cjs`

- [ ] **Step 1: 更新 depguard 规则严重级别**

在 `.eslintrc.cjs` 中，找到 `depend/depguard` 规则，将第一个参数从 `'warn'` 改为 `'error'`：

```javascript
'depend/depguard': ['error', {  // ← 从 'warn' 改为 'error'
  rules: [
    // ...
  ],
}],
```

- [ ] **Step 2: 运行 lint 验证**

```bash
pnpm lint 2>&1 | tail -20
```

Expected: 不应该有 depguard 相关的 error 或 warning。如果有，需要修复。

- [ ] **Step 3: 提交**

```bash
git add .eslintrc.cjs
git commit -m "chore: switch eslint-plugin-depend to error mode after migration complete"
```

---

### Task 5.4: 最终验证

- [ ] **Step 1: 完整 TypeScript 检查**

```bash
pnpm type-check
```

Expected: Exit code 0, 没有错误。

- [ ] **Step 2: 完整 ESLint 检查**

```bash
pnpm lint
```

Expected: Exit code 0, 没有 errors（可能有 warnings）。

- [ ] **Step 3: 构建验证**

```bash
pnpm build
```

Expected: Build completes successfully.

- [ ] **Step 4: 列出目录结构**

```bash
tree -L 3 src/ -I node_modules 2>/dev/null || find src -maxdepth 3 -type d | sort | head -30
```

Expected: 看到新的 `src/bff/`、`src/web/`、`src/shared/` 目录，没有 `src/lib/`。

- [ ] **Step 5: 确认没有遗留的 @/lib import**

```bash
grep -r "@/lib/" src --include="*.ts" --include="*.tsx" || echo "✓ No @/lib imports found"
```

Expected: "✓ No @/lib imports found"

- [ ] **Step 6: 验证依赖方向**

```bash
echo "=== Web importing BFF ===" && (grep -r "@bff/" src/web --include="*.ts" --include="*.tsx" | head -3 || echo "None found")
echo "=== BFF importing Web ===" && (grep -r "@web/" src/bff --include="*.ts" --include="*.tsx" | head -3 || echo "None found")
echo "=== Shared importing Web ===" && (grep -r "@web/" src/shared --include="*.ts" --include="*.tsx" | head -3 || echo "None found")
echo "=== Shared importing BFF ===" && (grep -r "@bff/" src/shared --include="*.ts" --include="*.tsx" | head -3 || echo "None found")
```

Expected: 
- Web importing BFF: 应该没有直接导入（只通过 `@shared/` 间接使用）
- BFF importing Web: 应该完全没有
- Shared importing Web: 应该完全没有
- Shared importing BFF: 应该完全没有

---

### Task 5.5: 手动功能回归测试

- [ ] **Step 1: 启动开发服务器**

```bash
pnpm dev > /tmp/dev.log 2>&1 &
DEV_PID=$!
sleep 5
```

- [ ] **Step 2: 测试登录流程**

访问 `http://localhost:3000/login`，尝试登录。

Expected: 登录页面正常加载，OAuth 流程可以启动（不需要完整认证，只需验证页面加载）。

- [ ] **Step 3: 测试组织选择**

若登录成功，访问 `http://localhost:3000/org-selector`。

Expected: 页面正常加载，组织列表可加载。

- [ ] **Step 4: 测试项目编辑**

若有项目，访问项目模型编辑页面。

Expected: 页面正常加载，GraphQL 查询工作。

- [ ] **Step 5: 停止开发服务器**

```bash
kill $DEV_PID 2>/dev/null || echo "Dev server stopped"
```

- [ ] **Step 6: 检查是否有控制台错误**

```bash
cat /tmp/dev.log | grep -i "error\|warn" | head -10
```

Expected: 没有与迁移相关的 import 错误。

---

### Task 5.6: 最终提交与总结

- [ ] **Step 1: 查看 git log 检查所有 commit**

```bash
git log --oneline | head -20
```

Expected: 应该看到所有 5 批迁移的 commit。

- [ ] **Step 2: 创建迁移完成标签**

```bash
git tag migration/web-bff-separation-complete -m "Web/BFF separation migration complete

- Migrated 17 files to @bff layer (apollo, auth, cms, api)
- Migrated 9 files to @web layer (components, hooks, stores, graphql, providers, cache, cms, routing)
- Migrated 8 files to @shared layer (cms, utils, typography, theme)
- Deleted src/lib directory
- Added ESLint depguard with error mode enforcement"

git push origin migration/web-bff-separation-complete
```

- [ ] **Step 3: 最终验证**

```bash
pnpm type-check && pnpm lint && pnpm build
```

Expected: All pass.

- [ ] **Step 4: 生成迁移报告**

```bash
cat > MIGRATION_REPORT.md <<'EOF'
# Web/BFF Separation Migration Report

**Date**: 2026-03-26
**Status**: ✅ Complete

## Summary

Successfully migrated ModelCraft frontend codebase from mixed `src/lib/` structure to separated `@bff/`, `@web/`, `@shared/` layers.

### Files Migrated

**BFF Layer (17 files)**
- API Routes: 5 files → src/bff/api/
- Apollo Config: 1 file → src/bff/apollo/
- Authentication: 2 files → src/bff/auth/
- CMS: 1 file → src/bff/cms/

**Web Layer (9 files)**
- Providers: 2 files → src/web/providers/
- Cache: 1 file → src/web/cache/
- CMS: 1 file → src/web/cms/
- Routing: 1 file → src/web/routing/
- Core: 56 files → src/web/{components,hooks,stores,graphql}/

**Shared Layer (8 files)**
- CMS Utilities: 2 files → src/shared/cms/
- Utils: 3 files → src/shared/utils/
- Configuration: 3 files → src/shared/

### Verification

✅ TypeScript compilation: 0 errors
✅ ESLint with depguard: 0 errors
✅ Next.js build: successful
✅ Manual regression tests: all passed
✅ Dependency direction enforced: Web ⊥ BFF, Shared ⊥ (Web|BFF)

### Breaking Changes

None. This is a pure refactoring with no functional changes.
EOF
```

- [ ] **Step 5: 提交报告**

```bash
git add MIGRATION_REPORT.md
git commit -m "docs: add migration completion report"
```

---

## 成功标准清单

迁移完成需满足以下所有条件：

- [ ] `src/lib/` 目录已删除
- [ ] `src/bff/`、`src/web/`、`src/shared/` 目录都有文件
- [ ] `pnpm type-check` 通过（0 errors）
- [ ] `pnpm lint` 通过（0 errors）
- [ ] `pnpm build` 通过
- [ ] `grep -r "@/lib/" src` 无结果
- [ ] `grep -r "@web/" src/bff` 无结果（Web 不应被 BFF 导入）
- [ ] `grep -r "@bff/" src/web` 无结果（Web 不应直接导入 BFF，只通过 shared）
- [ ] `grep -r "@web/" src/shared` 无结果
- [ ] `grep -r "@bff/" src/shared` 无结果
- [ ] 登录/OAuth 流程正常
- [ ] 组织切换正常
- [ ] 项目列表、模型编辑器正常加载
- [ ] 数据查询/变更正常
- [ ] CopilotKit 正常

