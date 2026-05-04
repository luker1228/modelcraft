## Why

当前前端把开发者数据调试与终端用户运行时数据操作压在同一个 `ModelRecordWorkspace` 中，通过 `workspaceMode`、不同 token、不同 Apollo client 和不同 GraphQL 文档做分支复用。这个抽象已经开始反噬实现：场景语义、身份语义、数据访问语义耦合在同一层，导致 runtime 后续独立演进和“开发者扮演终端用户”这类能力都难以落地。

项目仍处于 MVP 阶段，允许破坏性重构。此时直接拆掉错误抽象，比继续在 `workspaceMode` 上叠加分支和兼容层更低成本，也更能为 runtime 后续发展保留空间。

## What Changes

- **BREAKING** 删除单一 `ModelRecordWorkspace` 的双模式复用方式，改为独立的 `DevelopRecordWorkspace` 与 `RuntimeRecordWorkspace`。
- **BREAKING** 删除通过 `workspaceMode` 在业务组件内部切换 client、context、GraphQL 文档和权限展示的实现方式。
- 将字段插入、字段生命周期维护、关系维护、模型级调试能力收敛到 develop workspace，不再暴露给 runtime workspace。
- 保留 record query/create/edit/delete 的共有能力，但通过注入式 `record access adapter` 让两套 workspace 各自决定身份与端点，不再由共享组件直接创建 client。
- 收缩共享层，只保留 schema 过滤、表单模板、纯字段组件、纯显示组件等与身份无关的 primitives。
- 本次变更明确为 runtime workspace 预留独立 actor/access boundary，作为后续“开发者扮演终端用户”能力的承接点，但 **不在本次实现 impersonation**。

## Capabilities

### New Capabilities
- `record-workspace-separation`: 为 develop 与 runtime 提供独立的数据工作区实现，并通过共享 primitives + access adapter 保持有限复用。

### Modified Capabilities

## Impact

- 受影响前端入口：
  - `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx`
  - `src/app/end-user/[orgName]/[projectSlug]/data/page.tsx`
- 受影响前端组件：
  - `src/web/components/features/model-editor/model-record-form/*`
  - 相关 widgets、relation manager、insert menu、table/workspace 壳组件
- 受影响数据访问层：
  - `src/api-client/apollo/*`
  - record form 内部直接构造 runtime/end-user client 的组件
- 受影响验证：
  - shared widget / workspace 级别测试
  - develop/runtime 页面回归验证
  - 托管模型只读、record create/edit/delete 的回归检查
