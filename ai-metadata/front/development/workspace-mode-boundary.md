# Workspace 模式边界（Design vs End User）

> ⚠️ 终端用户前端页面已移除，`workspaceMode: 'end_user'` 当前无调用方。本文档保留作为设计参考，未来若恢复终端用户界面可重新启用。

## 背景

`ModelRecordWorkspace` 被设计态页面复用。通过 `workspaceMode` 定义能力边界。

## 模式定义

- `design`: 开发者设计态（模型结构可变）
- `end_user`: 终端用户运行态（仅数据读写，不允许结构变更）— 当前未使用

## 当前能力矩阵（全为 design 态）

- `canInsertField`: `true`
- `canManageFieldLifecycle`（字段废弃/删除）: `true`
- `canManageRelations`（关联管理）: `true`
- `canCopyRuntimeEndpoint`（复制 Runtime GraphQL Endpoint）: `true`

## 落地约定

1. 所有 `ModelRecordWorkspace` 调用方必须显式传入 `workspaceMode`。
2. 新增设计态能力时，必须先进入能力矩阵，再渲染 UI。
3. 前端能力开关只负责”展示与交互边界”，后端鉴权仍是最终防线。

## 已接入文件

- `modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx`
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/ModelRecordInsertMenu.tsx`
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/ModelRecordTable.tsx`
- `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx`

## 后续建议

后续可继续扩展能力矩阵：

- `canUseDesignGraphQLEndpoint`

每新增一项都必须同时更新本文档与调用方。
