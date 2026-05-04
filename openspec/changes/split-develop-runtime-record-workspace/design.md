## Context

当前 develop 入口与 end-user 入口都挂载同一个 workspace：

- develop 路由在 `ModelEditorView` 中渲染 `ModelRecordWorkspace workspaceMode="develop"`
- runtime 路由在 `/end-user/.../data` 中渲染 `ModelRecordWorkspace workspaceMode="end_user"`

这一层共享进一步把以下逻辑扩散到了多个业务组件：

- 选择 project query 还是 end-user query
- 选择 developer token 还是 end-user token
- 选择 project runtime client 还是 end-user scoped client
- 根据 mode 决定是否展示字段生命周期、关系维护、插列等控制

结果是：

- workspace 壳组件承担了两套产品语义
- shared widgets 也开始直接创建 client，无法视为真正的共享原语
- runtime 若要继续演化为独立产品面，需要持续背负 develop 语义

这个 change 解决的是架构边界问题，而不是单个 UI 分支问题。

## Goals / Non-Goals

**Goals:**

- 将 develop 与 runtime 的 record workspace 从页面壳层面彻底分开。
- 删除 `workspaceMode` 这条跨组件的共享分支链路。
- 建立 `record access adapter` 边界，让共享 primitives 不再直接依赖某一种身份或端点。
- 保持当前 runtime 数据查询/新增/编辑/删除能力可用。
- 保持当前 develop 的结构维护与关系维护能力可用。
- 为后续 runtime-only 演进和 impersonation 预留独立落点。

**Non-Goals:**

- 本次不实现“开发者扮演终端用户”功能本身。
- 本次不调整 EndUser 账号、Role、Project Access 的权限模型。
- 本次不重写 JSON Schema 协议、RJSF 表单协议或运行时 GraphQL 协议。
- 本次不追求所有 widgets 都继续共享；必要时允许将场景相关组件分叉。

## Execution Ownership

- `front-architect`：负责 workspace 拆分边界、目录规划、shared primitives 判定、access adapter 接口骨架。
- `front-worker`：负责路由切换、workspace 拆分实现、widgets/access adapter 改造、旧结构清理。
- `front-reviewer`：负责架构一致性复审、Lint/类型/回归检查、确认 shared 层未重新引入身份耦合。
- `backend-api` / `backend-worker`：本次不参与实现；待后续新增 impersonation 或 runtime debug session 协议时接手后端接口与实现。

## Decisions

### 1. 用两个明确的 workspace 组件替换一个 mode-driven workspace

决策：

- 新建 `DevelopRecordWorkspace`
- 新建 `RuntimeRecordWorkspace`
- 入口路由直接引用对应 workspace，不再传 `workspaceMode`

原因：

- 当前问题的根源是“用 mode 在一个组件内承载两套产品语义”
- 将差异上提到入口层，可以立即切断 mode 分支在下游蔓延

备选方案：

- 保留单 workspace，仅抽取 hooks：不能改变错误抽象，只是把分支藏起来
- 继续在现有组件内整理条件分支：短期更省事，但会继续阻碍 runtime 独立演进

### 2. 共享层只保留 identity-agnostic primitives

决策：

- 共享层只保留 schema/filter/template/display/widget 等纯原语
- 共享原语不得直接 import `createModelRuntimeClient`、`createEndUserScopedClient`、auth store 或基于路由推导身份

原因：

- 一旦共享层自己决定身份与端点，它就不再是共享层，而是隐式场景层
- 当前 `RelationSelector`、`RelationPicker` 就属于这种“看起来共享，实际绑死 runtime client”的伪共享

备选方案：

- 保持 widgets 自己建 client：会把 access model 继续埋进底层组件
- 彻底复制全部 widgets：实现快，但会放大后续维护成本

### 3. 通过注入式 access adapter 连接共享层和场景层

决策：

- develop/runtime workspace 各自拥有自己的 access adapter
- shared primitives 若需要远程数据，必须通过 adapter/context 注入访问能力

建议边界：

```ts
interface RecordAccessAdapter {
  queryModel(...): Promise<...>
  findMany(...): Promise<...>
  findUnique(...): Promise<...>
  createRecord(...): Promise<...>
  updateRecord(...): Promise<...>
  deleteRecord(...): Promise<...>
}
```

原因：

- adapter 让“谁来访问、用什么 token、打哪个 endpoint”只由 workspace 决定
- runtime 后续加入 impersonation 时，只需要替换 runtime adapter，不需要回流污染 shared 层和 develop workspace

备选方案：

- 继续把 client 作为 props 逐层透传：可以工作，但容易退化成大量 plumbing
- 用全局 store 让 widgets 自己取身份：隐式依赖更强，测试和演进更差

### 4. develop-only 能力留在 develop workspace，不追求 runtime 兼容

决策：

- 以下能力明确只属于 develop workspace：
  - 插入字段
  - 字段废弃/删除
  - 关系维护对话框
  - 模型调试型控制

原因：

- 这些能力的产品语义属于模型设计与结构调试，不属于终端用户运行时数据操作
- 把它们留在 runtime 只会让 runtime 继续背负 develop 复杂度

备选方案：

- 继续在 runtime 中隐藏按钮但复用同一实现：代码仍然耦合，复杂度并未消失

### 5. 采用破坏性迁移，不保留兼容 shim

决策：

- 直接删除 `workspaceMode`
- 路由入口一次切换到新 workspace
- 旧共享壳组件允许被拆散、重命名或删除

原因：

- 项目仍处于 MVP 阶段，兼容层只会延长错误结构的寿命
- 本次变更的主要收益就是把边界一次性拉正

备选方案：

- 先保留兼容层再逐步切换：对用户风险更低，但会显著增加重构时间和认知负担

## Risks / Trade-offs

- [Risk] 运行时数据表单行为在拆分后出现回归 → Mitigation：将 query/create/edit/delete 与托管模型只读列为强制验收项，先对 runtime 做等价迁移，再加差异化能力。
- [Risk] 共享 primitives 识别不准，导致复制过度或共享过度 → Mitigation：以“是否直接依赖身份/端点”为硬标准，宁可短期分叉，也不保留伪共享。
- [Risk] 路径和命名调整较大，影响现有导入关系 → Mitigation：先完成入口切换和 compile pass，再逐步清理旧文件。
- [Risk] relation 相关 widgets 改成 adapter 注入后改动面较广 → Mitigation：先收敛 develop/runtime workspace，再只对仍需共享的远程 widgets 做 adapter 改造。

## Migration Plan

1. 新建 develop/runtime 两套 workspace 入口组件。
2. 将路由入口从 `ModelRecordWorkspace + workspaceMode` 切换为显式 workspace。
3. 把 develop-only 功能从共享壳中抽离到 develop workspace。
4. 为 runtime workspace 建立独立 access adapter，并完成 record query/create/edit/delete 等价迁移。
5. 为仍保留共享的远程 widgets 引入 adapter/context，去除其内部 client 构造逻辑。
6. 删除 `workspaceMode`、相关分支代码和不再使用的共享壳组件。

回滚策略：

- 若迁移中断，可在本地分支回退到旧 workspace 结构；本次不设计线上双栈兼容。

## Open Questions

- runtime workspace 首版是否保留当前 table chrome 与 tab 交互，还是只保证行为等价即可。
- relation picker / selector 是否值得继续共享；若 adapter 改造成本过高，允许先做 runtime/develop 双实现。
- impersonation 的 token/source-of-truth 最终落在 runtime adapter 还是更上层 provider；本次只预留边界，不做最终决策。
