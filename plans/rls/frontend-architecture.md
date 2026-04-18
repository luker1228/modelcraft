# RLS 前端实现计划

> 文档日期：2026-04-17  
> 依赖文档：`prd/rls/prd.md`、`prd/rls/01~04-*.md`、`plans/rls/api-contract.md`  
> 前提：后端 API Contract（`plans/rls/api-contract.md`）已冻结，前端按此契约实现。

---

## 概览

### 影响范围

| 层 | 受影响文件/目录 | 变更性质 |
|----|--------------|---------|
| **设计态 UI** | `model-editor/_components/ModelDetailPanel.tsx` | 新增 `END_USER_REF` 格式徽章、`isRLSEnabled` 状态展示 |
| **设计态 UI** | `model-editor/_components/ModelSidebar.tsx` | Model 列表条目中展示 RLS 状态图标 |
| **设计态 UI** | `model-editor/_hooks/use-field-operations.ts` | `handleRemoveField` 增加前置二次确认弹窗逻辑 |
| **GraphQL 查询/变更** | `web/graphql/queries/model.ts` | 在 Model fragment 补充 `isRLSEnabled` 字段 |
| **GraphQL 查询/变更** | `web/graphql/mutations/model.ts` | `REMOVE_FIELD` 补充 `warning` 返回字段；`ADD_FIELDS` 补充 `EndUserRefAlreadyExists` 错误分支 |
| **BFF Runtime** | `bff/cms/runtime-query-builder.ts` | 新增 `isEndUserRefField()` 判断，从 CreateInput/UpdateInput 的字段中排除 `owner` |
| **BFF Auth** | `bff/auth/jwt-utils.ts` | Developer JWT 的 `iss` 从 `"modelcraft"` 迁移到 `"mc-developer"` |
| **BFF EndUser** | `bff/end-user/end-user-jwt-utils.ts` | EndUser JWT 的 `iss` 从 `"modelcraft-end-user"` 迁移到 `"mc-enduser"` |
| **MSW Mock** | `mocks/handlers/` | 新增/修改 GraphQL handler 覆盖新返回结构 |
| **GraphQL Codegen** | `generated/graphql.ts` | 重新生成，同步新增类型 |

### 不需要改的部分（明确排除）

- **Runtime 认证中间件**（`middleware.ts`）：Runtime endpoint 的 JWT 校验由后端完成，BFF 仅负责 EndUser JWT 的签发与 `iss` 值对齐，不需要在 Next.js middleware 中加新逻辑。
- **EndUser 账号管理 UI**：管理后台已有创建 EndUser 的能力，PRD 明确此为现有功能，本期不新增 UI。
- **RLS 开关 Toggle**：按 PRD，字段即开关，不新增独立 UI toggle。
- **多 EndUserRef 字段支持**：第一期仅支持单字段，UI 无需变更此约束展示。
- **Runtime GraphQL 查询构建的 WHERE 逻辑**：WHERE 注入在后端执行，前端 `runtime-query-builder.ts` 只需处理 `owner` 字段的输入类型屏蔽，不需要在前端构造 WHERE 条件。
- **End-user 登录/注册页面 UI**（`app/org/[orgName]/project/[projectSlug]/user/login/`）：表单 UI 无需改动，仅需确认底层 token 携带方式正确。
- **`model-record-form` 的表单渲染逻辑**（`filter-json-schema-for-form.ts`、`build-ui-schema.ts`）：Runtime Schema 对 EndUser 已在后端屏蔽 `owner` 字段，前端 JSON Schema 渲染链路无需额外处理。

---

## 变更模块

### 模块 A：管理后台 — 字段系统（Design-time UI）

#### A1. `END_USER_REF` 格式在字段列表中的标识

**影响组件：**
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelDetailPanel.tsx`

**需要新增/修改的 UI 行为：**

1. **格式徽章颜色区分**  
   在 `ModelDetailPanel` 的字段表格中，`format` 列渲染格式徽章时，对 `END_USER_REF` 单独使用区别于其他格式（如 `STRING`、`RELATION`）的颜色样式（建议使用紫色系，与"归属/身份"语义匹配）。当前代码使用 `emerald` 色，`END_USER_REF` 应视为特殊类型用独立色系。

2. **锁定图标 + "归属字段"标签**  
   对 `format === 'END_USER_REF'` 的字段，在字段名旁展示一个专属标签（如"归属字段"或 Shield 图标），让开发者直觉上理解这是 RLS 关键字段。参考现有"系统字段"标签的展示方式（同一 `<div className="flex items-center gap-2">` 内追加 `<span>` 标签）。

3. **字段操作菜单的限制**  
   `END_USER_REF` 字段（`owner`）的三点菜单中：
   - **编辑** 项：禁用（`owner` 字段名固定、格式不可变）
   - **废弃** 项：禁用（不支持废弃 `owner` 字段）
   - **删除** 项：**保留可点击**，但跳过当前"必须先废弃"的前置检查（因为 `owner` 字段不允许废弃，但允许直接删除）。点击后触发 **A3 的二次确认弹窗**，而不是直接执行删除。
   
   > 依赖：需要在 `use-field-operations.ts` 中为 `handleRemoveField` 增加 `END_USER_REF` 字段的特殊处理分支。

4. **`nonNull` 标记展示**  
   `owner` 字段天然是 `nonNull: true`，在字段类型列或 tooltip 中无需额外处理，但若 UI 层有 nullable 展示逻辑，需确保 `END_USER_REF` 字段不会被误标为可空。

---

#### A2. Model 列表与详情中的 `isRLSEnabled` 状态展示

**影响组件：**
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx`
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelDetailPanel.tsx`
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/types.ts`（需在 `EditorModel` 类型中补充 `isRLSEnabled` 字段）

**需要新增/修改的 UI 行为：**

1. **Model 列表条目（ModelSidebar）**  
   在 Model 列表每个条目右侧，若 `model.isRLSEnabled === true`，展示一个小图标（Shield 或 Lock，尺寸 `size-3`），颜色用 `text-purple-500` 或 `text-emerald-500`，鼠标 hover 时 tooltip 文案为"数据隔离已启用"。这让开发者在侧边栏一眼识别哪些 Model 启用了 RLS。

2. **Model 详情面板（ModelDetailPanel）元信息区块**  
   在元信息 `grid` 区域（当前有"标识名称"、"显示标题"、"数据库"、"描述"、"展示字段"共 5 个格），新增一个只读状态项：
   - 标签："数据隔离"
   - 值：若 `isRLSEnabled === true` 显示绿色"已启用"徽章；若 `false` 显示灰色"未启用"徽章
   - 该字段永远只读，无编辑态（对应 PRD 中"无单独 RLS 开关"）

3. **`EditorModel` / `EditorModelField` 类型补充**  
   `_hooks/types.ts` 中的 `EditorModel` 接口需新增 `isRLSEnabled: boolean` 字段，以便组件使用。

---

#### A3. 删除 `END_USER_REF` 字段的二次确认弹窗

**影响文件：**
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/use-field-operations.ts`
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelDetailPanel.tsx`（或新建确认弹窗组件）

**需要新增/修改的行为：**

> ⚠️ **关键交互顺序（来自 api-contract.md）**：后端在删除执行后才通过 `RemoveFieldPayload.warning` 返回 `RLS_WILL_DISABLE`。这意味着前端**不依赖后端 warning 来决定是否弹窗**，而是在调用 mutation 之前，由前端判断 `field.format === 'END_USER_REF'` 自主弹出确认弹窗，用户确认后才发请求。

1. **`handleRemoveField` 增加前置逻辑**  
   在 `use-field-operations.ts` 的 `handleRemoveField` 中，检查 `field.format === 'END_USER_REF'`：
   - 若是，设置一个新状态（如 `pendingRemoveOwnerField`）并打开确认弹窗，暂不执行 mutation
   - 若否，走现有流程（当前要求先废弃才能删除，`END_USER_REF` 字段绕过此检查）

2. **确认弹窗内容**  
   使用现有 `AlertDialog` 组件（`src/web/components/ui/alert-dialog.tsx`），内容：
   - **标题**："确认关闭数据隔离？"
   - **正文**："删除 `owner` 字段后，该 Model 的数据隔离将关闭，所有终端用户将可访问全量数据。此操作不可撤销。"
   - **取消按钮**：关闭弹窗，不执行删除
   - **确认按钮**（Destructive 样式）："删除并关闭隔离"

3. **后端 `warning` 的处理**  
   `REMOVE_FIELD` mutation 的返回值中新增 `warning` 字段（见模块 D）。删除成功后：
   - 若 `warning === 'RLS_WILL_DISABLE'`，在删除成功的 toast 中附加提示文案："数据隔离已关闭"
   - 这仅是信息性展示（操作已完成），无需再次弹窗

---

#### A4. 新建 Model 时 `owner` 字段的预置展示

**影响组件：**
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/CreateModelDialog.tsx`（若有字段预览）

**需要新增/修改的行为：**

1. **CreateModelDialog 中的提示文案**  
   在新建 Model 的表单弹窗中，添加一段说明文字（或 Info Alert）：
   > "新建模型将自动包含一个 `owner` 字段（归属字段），用于数据隔离。您可以在创建后删除该字段以关闭隔离。"
   
   这是纯 UI 文案，不影响 API 调用逻辑（后端自动生成 `owner` 字段）。

2. **创建后的字段列表刷新**  
   无需特殊处理——现有 `refetchQueries: ['GetModel']` 逻辑会在创建后刷新 Model 详情，`owner` 字段会自然出现在列表中，附带 A1 中定义的"归属字段"标签。

3. **导入 Model 场景**  
   导入 Model（`ImportModel` mutation）后不会有 `owner` 字段，这是后端行为，前端无需额外处理。Model 列表中该 Model 的 `isRLSEnabled` 为 `false`，按 A2 展示"未启用"状态即可。

---

### 模块 B：BFF 层 — Runtime Query Builder

**影响文件：**
- `src/bff/cms/runtime-query-builder.ts`

#### B1. `FieldDefinition` 扩展

当前 `FieldDefinition` 接口（第 9-16 行）只有 `name`、`type`、`format`、`schemaType`、`storageHint` 字段，不含 `isPrimary`。需要在此接口（或通过现有 `ModelField` 接口的映射路径）中能够识别 `format === 'END_USER_REF'` 的字段。

**需要新增的判断逻辑（声明式描述）：**

1. **`isEndUserRefField(field: FieldDefinition): boolean`**  
   判断一个字段是否是 `END_USER_REF` 类型（即 `field.format === 'END_USER_REF'`）。作为内部工具函数供下方函数调用。

2. **`extractWritableFieldNamesFromSchema` 的行为不变**  
   该函数基于 JSON Schema 的 `readOnly` 属性过滤字段，**如果后端已在 EndUser 视角下将 `owner` 从 Schema properties 中排除（或标记为 `readOnly: true`），则此函数天然正确，无需修改**。  
   但需确认后端 JSON Schema 的实际行为：若后端在 EndUser 视角的 `jsonSchema` 中已不包含 `owner` 字段，则前端零改动；若包含且不带 `readOnly`，则需在此函数中增加对 `END_USER_REF` 格式字段的过滤。

   > 📌 **待确认**：后端 Runtime Schema 对 EndUser 视角的 `jsonSchema` 是否已排除 `owner` 字段？如果是，B1.2 无需改动。

3. **`buildCreateMutation` / `buildUpdateMutation` 调用侧**  
   当 `sanitizeMutationInputData` 被调用时，`allowedFieldNames` 来自 `extractWritableFieldNamesFromSchema`。如后端 Schema 已排除 `owner`，则 `sanitizeMutationInputData` 自动不会提交 `owner` 字段，无需额外改动。

   **只有当后端 EndUser 视角的 Schema 仍包含 `owner` 字段时**，才需要在 `runtime-query-builder.ts` 中新增一个 `filterOwnerFieldFromMutationInput(fields: FieldDefinition[]): FieldDefinition[]` 函数，在构建 CreateInput/UpdateInput 时将 `owner` 字段从 `allowedFieldNames` 中过滤掉。

#### B2. `buildFieldSelections` 的 owner 处理

对于 Query（FindMany/FindFirst）返回结果，`owner` 字段对 EndUser **可见**（按 PRD）。因此：
- **Query 侧（`buildFindManyQuery`、`buildFindFirstQuery`）**：无需任何改动，`owner` 字段正常出现在 `fieldSelection` 中即可。
- **Mutation 侧（`buildCreateMutation`、`buildUpdateMutation`）**：仅排除 `owner` 字段不出现在 input 变量中（见 B1.3）。

#### B3. `model-field-mapping.ts` 的相关函数

**影响文件：**
- `src/web/components/features/model-editor/model-record-form/runtime/model-field-mapping.ts`

`buildEditFormData` 函数（第 43-55 行）在构建编辑表单初始值时跳过了 `isPrimary` 字段，但未跳过 `END_USER_REF` 字段。若后端 Schema 已排除 `owner`，则此函数无需改动；若未排除，则需在此处补充对 `format === 'END_USER_REF'` 的跳过逻辑（类似对 `isPrimary` 的处理）。

---

### 模块 C：BFF 层 — EndUser 认证

#### C1. Developer JWT `iss` 迁移

**影响文件：**
- `src/bff/auth/jwt-utils.ts`

**当前状态：**  
第 10 行：`const ISSUER = 'modelcraft'`

**需要变更的行为：**

1. 将 `ISSUER` 常量从 `'modelcraft'` 改为 `'mc-developer'`（对应 api-contract.md §3.1）。

2. **迁移期兼容性**：后端在迁移期内同时接受 `'modelcraft'` 和 `'mc-developer'`（api-contract.md 第 204 行明确说明）。因此前端可直接切换，无需在前端同时支持旧值。

3. **受影响的 BFF Route**：`src/app/api/bff/auth/login/route.ts` 和 `refresh/route.ts` 通过调用 `signAccessToken`（来自 `jwt-utils.ts`）颁发 Developer JWT，无需修改这些 route 文件本身，只需修改 `jwt-utils.ts` 中的 `ISSUER` 常量。

4. **验证函数同步**：`verifyAccessToken` 也使用 `ISSUER` 做验证，同步修改后，BFF 内部（如 middleware.ts 调用 `verifyAccessToken`）也会正确验证新 iss 值。

---

#### C2. EndUser JWT `iss` 迁移

**影响文件：**
- `src/bff/end-user/end-user-jwt-utils.ts`

**当前状态：**  
第 14 行：`const ISSUER = 'modelcraft-end-user'`（注释说明与开发者 `'modelcraft'` 区分）

**需要变更的行为：**

1. 将 `ISSUER` 常量从 `'modelcraft-end-user'` 改为 `'mc-enduser'`（对应 api-contract.md §3.3）。

2. `signEndUserAccessToken`（第 29 行）颁发的 JWT payload 已包含 `sub`（userId）、`org_name`、`project_slug`、`role: 'end_user'`，无需新增字段（api-contract.md §3.3 仅要求 `iss`、`sub`、`exp`、`projectSlug`）。

3. `verifyEndUserAccessToken`（第 50 行）同步验证新 iss 值，保持逻辑一致。

4. **Runtime 请求中携带 EndUser JWT 的方式**：  
   当前 `bff/end-user/end-user-auth-client.ts` 通过 `getEndUserToken()` 从 in-memory store 获取 access token。Runtime 调用时，该 token 应以 `Authorization: Bearer <token>` 方式携带。需确认（或在 `ModelRecordWorkspace.tsx` 等调用 Runtime 的地方验证）Apollo Client 实例对 Runtime endpoint 的请求头是否已携带此 token。若未携带，需在 Runtime 专用 Apollo Client 的 authLink 中补充从 `useEndUserAuthStore` 读取 EndUser token 的逻辑。

   > 📌 **待确认**：当前 Runtime Apollo Client（`BFF 层`）是否已在请求头中携带 EndUser access token？若已有，C2.4 无需改动；若无，需在对应的 Apollo link 配置中补充。

---

#### C3. BFF EndUser Route 无需改动

`src/app/api/bff/end-user/auth/login/route.ts` 等 BFF Route 不直接操作 `iss` 值，均通过调用 `signEndUserAccessToken`（来自 `end-user-jwt-utils.ts`）颁发 token。因此 C2 修改 `end-user-jwt-utils.ts` 中的 `ISSUER` 后，这些 route 文件无需任何改动。

---

### 模块 D：GraphQL Codegen 同步

#### D1. 需要修改的 GraphQL 查询/变更文件

**文件：`src/web/graphql/queries/model.ts`**

所有查询了 `Model` 类型的 query（`GET_MODELS`、`GET_MODEL`、`GET_MODEL_BY_NAME`、`GET_MODEL_GROUPS`）的 Model 节点中，需新增：
```
isRLSEnabled
rlsPolicy {
  selectPredicate
  insertCheck
  updatePredicate
  updateCheck
  deletePredicate
  preset
}
```

具体受影响的 query：

- `GET_MODELS`（第 4 行）：在 `node` 片段中补充
- `GET_MODEL`（第 92 行）：在 `model` 片段中补充
- `GET_MODEL_BY_NAME`（第 217 行）：在 `model` 片段中补充
- `GET_MODEL_RECORD_WORKSPACE`（第 187 行）：按需添加
- `GET_MODEL_GROUPS`（第 291 行）：groups 中的 model 列表如需展示 RLS 状态，也需补充

**文件：`src/web/graphql/mutations/model.ts`**

1. **`REMOVE_FIELD`（第 563 行）**：在返回结构中，与 `model`、`error` 同级新增 `warning` 字段

2. **`ADD_FIELDS`（第 416 行）**：在 `error` 的 union 片段中，新增 `EndUserRefAlreadyExists` 分支：
   ```graphql
   ... on EndUserRefAlreadyExists {
     message
     code
   }
   ```

3. **新增 `SET_MODEL_RLS_POLICY` mutation**：
   ```graphql
   mutation SetModelRLSPolicy($input: SetModelRLSPolicyInput!) {
     setModelRLSPolicy(input: $input) {
       model {
         id
         isRLSEnabled
         rlsPolicy {
           selectPredicate
           insertCheck
           updatePredicate
           updateCheck
           deletePredicate
           preset
         }
       }
       error {
         ... on PolicyNotFound {
           message
           code
         }
         ... on InvalidInput {
           message
           code
         }
         ... on RLSInvalidExpr {
           message
           code
           details
         }
       }
     }
   }
   ```

4. **新增 `VALIDATE_RLS_EXPR` mutation**（供可视化构建器实时校验）：
   ```graphql
   mutation ValidateRLSExpr($input: ValidateRLSExprInput!) {
     validateRLSExpr(input: $input) {
       valid
       errors
     }
   }
   ```

5. **新增 `GET_PROJECT_AUTH_SCHEMA` query 和 `SET_PROJECT_AUTH_SCHEMA` mutation**（模块 G 使用）：
   ```graphql
   query GetProjectAuthSchema($projectId: ID!) {
     getProjectAuthSchema(projectId: $projectId) {
       projectId
       variables {
         name
         source
         type
       }
     }
   }
   mutation SetProjectAuthSchema($input: SetProjectAuthSchemaInput!) {
     setProjectAuthSchema(input: $input) {
       projectId
       variables { name source type }
     }
   }
   ```

6. **各 mutation 中 Model 片段**：`CREATE_MODEL`、`UPDATE_MODEL`、`ADD_FIELDS`、`UPDATE_FIELD`、`REMOVE_FIELD`、`SYNC_MODEL_SCHEMA`、`REPAIR_MODEL` 的返回 model 片段均需补充 `isRLSEnabled` + `rlsPolicy { selectPredicate insertCheck updatePredicate updateCheck deletePredicate preset }` 字段。

#### D2. 需要重新生成的类型

执行 `npm run codegen`（或等效命令）后，以下类型将自动更新：

- **`FormatType` 枚举**：新增 `END_USER_REF` 值
- **`RemoveFieldPayload` 类型**：新增 `warning?: RemoveFieldWarning | null`
- **`RemoveFieldWarning` 枚举**：新增值 `RLS_WILL_DISABLE`
- **`AddFieldsError` union**：新增 `EndUserRefAlreadyExists` 类型
- **`EndUserRefAlreadyExists` 类型**：新增 `{ message: string; code: string; __typename: 'EndUserRefAlreadyExists' }`
- **`Model` 类型**：新增 `isRLSEnabled: boolean`，新增 `rlsPolicy?: ModelRLSPolicy | null`
- **`ModelRLSPolicy` 类型**：新增 `{ selectPredicate: string; insertCheck: string; updatePredicate: string; updateCheck: string; deletePredicate: string; preset?: RLSPreset | null }`（五件套字段类型为 `string`，存储 JSON 表达式）
- **`RLSPreset` 枚举**：`READ_WRITE_OWNER | READ_ALL_WRITE_OWNER | READ_ALL | READ_WRITE_ALL | NO_ACCESS`
- **`SetModelRLSPolicyInput`**：`{ modelId: string; selectPredicate: string; insertCheck: string; updatePredicate: string; updateCheck: string; deletePredicate: string }`（五件套类型为 `string`，传 JSON）
- **`SetModelRLSPolicyPayload`**：包含 `model` 和 `error` 联合类型（含新增 `RLSInvalidExpr`）
- **`ValidateRLSExprInput`**：`{ projectId: string; modelId: string; operation: RLSOperation; expression: string }`
- **`ValidateRLSExprPayload`**：`{ valid: boolean; errors: string[] }`
- **`RLSOperation` 枚举**：`SELECT_PREDICATE | INSERT_CHECK | UPDATE_PREDICATE | UPDATE_CHECK | DELETE_PREDICATE`
- **`AuthVariable`**：`{ name: string; source: string; type: string }`
- **`ProjectAuthSchema`**：`{ projectId: string; variables: AuthVariable[] }`
- **`SetProjectAuthSchemaInput`**：`{ projectId: string; variables: AuthVariableInput[] }`
- **（删除）`ExprType` 枚举**：已被 `string` JSON 表达式替代

> `generated/graphql.ts`（75.8K）为 codegen 自动生成文件，禁止手动修改，通过重新运行 codegen 同步所有变更。

---

### 模块 E：Policy 配置 UI（访问控制 Tab）

#### E1. 访问控制 Tab 入口

**影响组件：**
- `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelDetailPanel.tsx`
- 或新建 `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelAccessControlTab.tsx`

**展示条件**：`model.rlsPolicy != null`（即存在 owner 字段），否则不渲染此 Tab

在 `ModelDetailPanel` 的 Tab 列表中新增"访问控制"Tab，与"字段"、"设置"等 Tab 并列。

---

#### E2. Preset 选择器

**UI 结构**：5 张卡片，展示 5 种 Preset，当前激活的高亮（选中态）。

| Preset 卡片 | 标题 | 描述 | 标记 |
|-------------|------|------|------|
| `READ_WRITE_OWNER` | 读写自己 | 每个用户只能访问自己的数据（任务、订单） | |
| `READ_ALL_WRITE_OWNER` | 读全部写自己 | 所有人可读，但只能修改自己的数据（评论、帖子） | |
| `READ_ALL` | 只读全部 | 所有人可读，任何人不可写（公告、商品目录） | |
| `READ_WRITE_ALL` | 读写全部 | 所有用户可读写任意数据（无隔离） | ⚠️ 高危 |
| `NO_ACCESS` | 完全封锁 | 终端用户无法访问（系统内部表） | |

**交互逻辑**：
1. 读取 `model.rlsPolicy.preset` 确定当前选中态（null 时不高亮任何卡片，显示"自定义"）
2. 点击非高危 Preset 卡片 → 映射到五件套 JSON 表达式 → 调用 `SET_MODEL_RLS_POLICY` mutation
3. 点击 `READ_WRITE_ALL`（⚠️高危）→ 先弹**二次确认弹窗**，用户确认后再调用 mutation
4. mutation 成功 → toast 提示"访问控制已更新"，Apollo Cache 更新

**高危二次确认弹窗**：
- **标题**："确认开启高危策略？"
- **正文**："此策略允许所有终端用户读写任意数据，包括其他用户的数据，请确认你了解风险。"
- **取消按钮**：关闭弹窗，不执行操作
- **确认按钮**（Destructive 样式）："确认，我了解风险"

**Preset → 五件套 JSON 映射表**（前端负责映射，不新增独立 mutation）：

```typescript
const OWNER_EXPR = '{"owner":{"_eq":{"_auth":"uid"}}}'

const PRESET_SCOPE_MAP: Record<RLSPreset, {
  selectPredicate: string;
  insertCheck: string;
  updatePredicate: string;
  updateCheck: string;
  deletePredicate: string;
}> = {
  READ_WRITE_OWNER: {
    selectPredicate: OWNER_EXPR, insertCheck: OWNER_EXPR,
    updatePredicate: OWNER_EXPR, updateCheck: OWNER_EXPR,
    deletePredicate: OWNER_EXPR,
  },
  READ_ALL_WRITE_OWNER: {
    selectPredicate: 'true', insertCheck: OWNER_EXPR,
    updatePredicate: OWNER_EXPR, updateCheck: OWNER_EXPR,
    deletePredicate: OWNER_EXPR,
  },
  READ_ALL: {
    selectPredicate: 'true', insertCheck: 'false',
    updatePredicate: 'false', updateCheck: 'false',
    deletePredicate: 'false',
  },
  READ_WRITE_ALL: {
    selectPredicate: 'true', insertCheck: 'true',
    updatePredicate: 'true', updateCheck: 'true',
    deletePredicate: 'true',
  },
  NO_ACCESS: {
    selectPredicate: 'false', insertCheck: 'false',
    updatePredicate: 'false', updateCheck: 'false',
    deletePredicate: 'false',
  },
}
```

---

#### E3. 可视化条件构建器

**目的**：不需要手写 JSON，提供图形化条件编辑界面，类似 Notion filter。

**UI 结构示例**：
```
┌─────────────────────────────────────────────┐
│ SELECT 条件                                  │
│  ┌──────────┐ ┌─────┐ ┌───────────────┐    │
│  │ owner_id │ │ = ▾ │ │ auth.uid()  ▾ │    │
│  └──────────┘ └─────┘ └───────────────┘    │
│  + 添加条件   + 添加 EXISTS                 │
│  关系：● AND  ○ OR                         │
└─────────────────────────────────────────────┘
```

**条件行组件**：字段名下拉（Model 字段列表）+ 操作符下拉（`=`、`!=`、`>`、`in` 等）+ 值输入（支持字面量或 auth 变量选择）

**值来源下拉**：文字输入 | auth.uid()（内置）| 已声明的 auth 变量（来自 auth_schema）

**JSON 预览区**：实时展示当前构建器状态对应的 JSON 字符串，便于高级用户理解

**实时校验**：构建器状态变化时（或切换到自定义 JSON 输入时），自动调用 `VALIDATE_RLS_EXPR` mutation：
- 请求：当前操作类型（`SELECT_PREDICATE` 等）+ 当前 JSON 表达式
- 响应：`valid=true` 时正常；`valid=false` 时在构建器区域展示红色错误提示

**PREDICATE vs CHECK 差异**：
- `selectPredicate`、`updatePredicate`、`deletePredicate`（PREDICATE 类）：显示 `+ 添加 EXISTS` 按钮
- `insertCheck`、`updateCheck`（CHECK 类）：隐藏 `+ 添加 EXISTS` 和 `_ref` 引用，操作符选项中禁用相关项

---

#### E4. 无 owner 字段时的状态处理

- `model.rlsPolicy == null`："访问控制"Tab 不显示（条件渲染）
- Model 列表（ModelSidebar）中该 Model 不展示 Shield 图标
- 删除 owner 字段后，Apollo Cache 中 `model.rlsPolicy` 变为 null，Tab 自动消失

---

### 模块 F：Runtime 错误处理（新增）

#### F1. `RLS_CHECK_VIOLATION` 前端捕获

**触发场景**：EndUser 调用 Runtime INSERT（insertCheck 不通过）或 UPDATE（updateCheck 不通过）时，后端返回 `RLS_CHECK_VIOLATION` 错误。

**前端处理策略**：

在调用 Runtime `createOne` / `updateOne` 的 Apollo Client 错误处理中，识别 `extensions.code === 'RLS_CHECK_VIOLATION'`：

```typescript
// 示例：在 model-record-form 的 submit handler 中
if (error.graphQLErrors?.some(e => e.extensions?.code === 'RLS_CHECK_VIOLATION')) {
  toast.error('操作失败：写入数据违反访问策略约束')
  return
}
```

**文案建议**：
- INSERT 失败："创建失败：当前策略不允许写入此数据"
- UPDATE 失败："更新失败：修改后的数据违反访问策略约束"

**注意**：USING 失败（SELECT/UPDATE/DELETE 的谓词过滤）**不产生** `RLS_CHECK_VIOLATION`，为静默行为（0行/空集），前端无需特殊处理。

---

### 模块 G：auth_schema 配置 UI（新增，Project 设置页）

**入口**：Project 设置页 → "认证变量"配置区（新 Tab 或新 Section）

**目的**：开发者声明额外的 JWT 变量，供 Policy 构建器中 `_auth` 引用。`uid` 内置不可编辑。

#### G1. 变量列表展示

- 查询 `GET_PROJECT_AUTH_SCHEMA` 获取当前已声明变量
- 以表格形式展示：名称 | JWT 来源 | 类型 | 操作（删除）
- `uid` 行固定置顶，展示"内置"标签，不可删除

#### G2. 添加变量

- 点击"添加变量"弹出表单：
  - 变量名（文字输入，英文字母/数字/下划线，不可为 `uid`）
  - JWT 来源（文字输入，如 `jwt.tenant_id`）
  - 类型（下拉：`uuid` | `string` | `integer`）
- 提交时调用 `SET_PROJECT_AUTH_SCHEMA`（全量覆盖变量列表）

#### G3. 删除变量

- 点击变量行的删除按钮 → 弹出二次确认（"删除后，引用该变量的 Policy 将校验失败"）
- 确认后调用 `SET_PROJECT_AUTH_SCHEMA`（移除该变量）

#### G4. 与 Policy 构建器联动

- Policy 构建器中"值来源"下拉，除 `auth.uid()` 外，自动列出当前 `auth_schema` 已声明的变量
- 若 Project 无 auth_schema 变量，"值来源"下拉仅显示字面量输入和 `auth.uid()`

---

## Mock 契约（前端先行）

### MSW Handler 变更清单

当后端接口未就绪时，前端通过 MSW mock 先行开发和联调。

#### M1. GraphQL Model query handler

**文件：`src/mocks/handlers/project/generated.ts`（codegen 自动生成）或手工 handler**

需要覆盖（或在手工 handler 中 override）以下 query/mutation 的返回值，补充 RLS 相关字段：

- **`GetModels` / `GetModel`**：在 model 节点中注入 `isRLSEnabled: true/false` 和 `rlsPolicy`（场景驱动：RLS 启用场景的 mock model 返回 `isRLSEnabled: true` + `rlsPolicy: { readScope: 'OWNER', writeScope: 'OWNER', preset: 'READ_WRITE_OWNER' }`）
- **`RemoveField`**：当被删除的字段是 `END_USER_REF` 时，返回 `warning: 'RLS_WILL_DISABLE'`；其他字段返回 `warning: null`
- **`AddFields`**：当尝试重复添加 `END_USER_REF` 字段时，返回 `EndUserRefAlreadyExists` 错误（通过 `x-mock-scenario` header 触发）
- **`SetModelRLSPolicy`**：按 input 参数返回更新后的 Policy

#### M2. 场景 header 约定

遵循现有约定（`x-mock-profile-scenario` / `x-mock-end-user-scenario`），为 Model 相关操作新增 mock 场景头：

| Header | 场景值 | 用途 |
|--------|--------|------|
| `x-mock-model-scenario` | `rlsEnabled` | `GetModel` 返回 `isRLSEnabled: true` + `rlsPolicy: {五件套均为{"owner":{"_eq":{"_auth":"uid"}}}, preset: READ_WRITE_OWNER}` |
| `x-mock-model-scenario` | `rlsReadAll` | `GetModel` 返回 `rlsPolicy: {selectPredicate:'true', 其余=owner_expr, preset: READ_ALL_WRITE_OWNER}` |
| `x-mock-model-scenario` | `rlsReadWriteAll` | `GetModel` 返回 `rlsPolicy: {五件套均为'true', preset: READ_WRITE_ALL}`（高危场景测试） |
| `x-mock-model-scenario` | `rlsDisabled` | `GetModel` 返回 `isRLSEnabled: false`，`rlsPolicy: null` |
| `x-mock-field-scenario` | `removeOwnerField` | `RemoveField` 返回 `warning: RLS_WILL_DISABLE` |
| `x-mock-field-scenario` | `endUserRefAlreadyExists` | `AddFields` 返回 `EndUserRefAlreadyExists` 错误 |
| `x-mock-policy-scenario` | `policyNotFound` | `SetModelRLSPolicy` 返回 `PolicyNotFound` 错误 |
| `x-mock-policy-scenario` | `rlsInvalidExpr` | `SetModelRLSPolicy` / `ValidateRLSExpr` 返回 `RLSInvalidExpr` 错误（含 details） |
| `x-mock-policy-scenario` | `validateValid` | `ValidateRLSExpr` 返回 `valid: true, errors: []` |
| `x-mock-auth-schema-scenario` | `withTenantId` | `GetProjectAuthSchema` 返回含 `tenant_id`（source: jwt.tenant_id, type: uuid）的变量列表 |

#### M3. Mock 数据工厂补充

**新增/修改文件：**
- `src/mocks/data/project/`（或类似路径）中的 model factory，补充 `isRLSEnabled` 字段
- 新增包含 `END_USER_REF` 格式字段（`owner`）的 model mock fixture

#### M4. EndUser auth mock 无需变更

`src/mocks/handlers/end-user/auth-handlers.ts` 无需改动。JWT iss 变更在 BFF 服务端处理，MSW 只 mock REST 响应体（不含 JWT payload），且前端 JWT 工具函数的测试应通过单元测试而非 MSW 覆盖。

---

## 实现顺序

### 依赖关系图

```
[D: GraphQL 类型同步（codegen）]
        │
        ├──→ [A: 管理后台 UI（字段系统）]
        │         ├── A1 字段格式徽章
        │         ├── A2 isRLSEnabled 状态展示
        │         ├── A3 删除确认弹窗 (依赖 A1 + D)
        │         └── A4 新建 Model 提示文案 (独立)
        │
        ├──→ [E: Policy 配置 UI]
        │         ├── E1 访问控制 Tab（依赖 D 中 rlsPolicy 类型）
        │         ├── E2 Preset 选择器（依赖 E1 + SET_MODEL_RLS_POLICY mutation）
        │         ├── E3 可视化条件构建器（依赖 E1 + VALIDATE_RLS_EXPR + G 中 auth_schema）
        │         └── E4 无 owner 字段时状态处理（依赖 E1）
        │
        ├──→ [G: auth_schema 配置 UI（Project 设置页）]
        │         ├── G1 变量列表（依赖 D 中 GET_PROJECT_AUTH_SCHEMA）
        │         ├── G2 添加变量（依赖 SET_PROJECT_AUTH_SCHEMA mutation）
        │         ├── G3 删除变量
        │         └── G4 与 Policy 构建器联动（依赖 E3）
        │
        └──→ [M: MSW Mock handler 更新] (可与 A/E/G 并行)

[C: BFF Auth JWT iss 迁移]  ←── 独立，可最先执行
        ├── C1 Developer JWT iss
        └── C2 EndUser JWT iss

[B: BFF Runtime Query Builder]  ←── 依赖后端确认 Schema 行为
        └── (依赖 D2 中 FormatType 枚举生成)
```

### 阶段划分与并行建议

#### Phase 0：前置（单人，约 0.5 天）

| 任务 | 文件 | 说明 |
|------|------|------|
| **D 先行：更新 GraphQL 查询/变更文件** | `web/graphql/queries/model.ts`、`web/graphql/mutations/model.ts` | 补充 `isRLSEnabled`、`rlsPolicy`、`warning`、`EndUserRefAlreadyExists`、`SET_MODEL_RLS_POLICY` |
| **D 重跑 codegen** | `generated/graphql.ts` | 生成新类型，供后续所有模块使用 |

> Phase 0 是所有模块的 unblock 前置，必须最先完成。

---

#### Phase 1：可 100% 并行（Phase 0 完成后，2 人以上同时进行）

| 任务编号 | 负责方向 | 文件 | 预估 |
|----------|---------|------|------|
| **P1-A** | UI Worker | `ModelDetailPanel.tsx`：A1 字段格式徽章 + A3 删除确认弹窗 | 1 天 |
| **P1-B** | UI Worker | `ModelSidebar.tsx` + `types.ts`：A2 isRLSEnabled 状态展示 | 0.5 天 |
| **P1-C** | BFF Worker | `jwt-utils.ts` + `end-user-jwt-utils.ts`：C1 + C2 JWT iss 迁移 | 0.5 天 |
| **P1-D** | BFF Worker | `use-field-operations.ts`：A3 删除弹窗的 Hook 侧逻辑 | 0.5 天 |
| **P1-E** | UI Worker | `ModelAccessControlTab.tsx`：E1 + E2 Preset 选择器 | 1 天 |
| **P1-G** | UI Worker | Project 设置页：G1 + G2 + G3 auth_schema 配置 UI | 1 天 |

> **P1-A 和 P1-D 需要协调**：UI 弹窗（P1-A）和弹窗触发逻辑（P1-D）由 Hook 驱动，建议同一 Worker 做；或者先定义好弹窗的 state interface，再拆分。

---

#### Phase 2：串行补充（Phase 1 完成后）

| 任务编号 | 文件 | 说明 |
|----------|------|------|
| **P2-E** | `ModelAccessControlTab.tsx` | E3 可视化条件构建器 + VALIDATE_RLS_EXPR 实时校验 + G4 auth_schema 联动（需 G 先行） |
| **P2-A** | `runtime-query-builder.ts` | B1-B3 EndUser 视角的 owner 字段排除（需先确认后端 Schema 行为，见"待确认"事项） |
| **P2-B** | `model-field-mapping.ts` | B3 `buildEditFormData` 对 `END_USER_REF` 字段的跳过处理（依赖 B1 确认结论） |

---

#### Phase 3：MSW Mock 同步（可与 Phase 1 并行）

| 任务编号 | 文件 | 说明 |
|----------|------|------|
| **P3-A** | `mocks/handlers/project/` | 新增 Model RLS 场景 mock handler（含 Policy 相关场景） |
| **P3-B** | `mocks/data/project/` | 新增含 `isRLSEnabled`、`END_USER_REF` 字段、`rlsPolicy` 的 mock fixture |

---

#### Phase 4：CreateModelDialog 提示文案（最低优先级，独立）

| 任务编号 | 文件 | 说明 |
|----------|------|------|
| **P4-A** | `CreateModelDialog.tsx` | A4 新建 Model 时的 Info 提示文案（纯 UI，无逻辑依赖） |

---

### 待确认事项（阻塞 Phase 2）

在开始 Phase 2（模块 B）之前，需要后端明确以下问题：

1. **后端 EndUser 视角的 `jsonSchema` 是否已排除 `owner` 字段？**  
   - 若是（`properties` 中不含 `owner`）：`runtime-query-builder.ts` **无需修改**，现有 `extractWritableFieldNamesFromSchema` 天然正确。
   - 若否（`properties` 中包含 `owner`，不带 `readOnly: true`）：需在 `extractWritableFieldNamesFromSchema` 或调用侧补充过滤逻辑。

2. **Runtime Apollo Client 是否已在请求头携带 EndUser token？**  
   - 若是：C2.4 无需改动。
   - 若否：需找到 Runtime endpoint 使用的 Apollo Client 实例（可能在 `ModelRecordWorkspace.tsx` 或 BFF CMS 层），在其 authLink 中补充 `Authorization: Bearer ${getEndUserToken()}` 逻辑。

