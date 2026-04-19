# Workspace 模式边界（Design vs End User）

## 背景

`ModelRecordWorkspace` 被设计态页面和终端用户页面复用。  
如果没有显式模式边界，设计态能力（如「插入列」「字段废弃/删除」）可能暴露到 end-user 页面。

## 目标

通过显式 `workspaceMode` 定义能力边界，避免通过“临时隐藏按钮”做权限控制。

## 模式定义

- `design`: 开发者设计态（模型结构可变）
- `end_user`: 终端用户运行态（仅数据读写，不允许结构变更）

## 当前能力矩阵

- `canInsertField`
  - design: `true`
  - end_user: `false`
- `canManageFieldLifecycle`（字段废弃/删除）
  - design: `true`
  - end_user: `false`
- `canManageRelations`（关联管理）
  - design: `true`
  - end_user: `true`
  - 说明：关联管理属于记录级数据操作，不属于模型结构变更。
- `canCopyRuntimeEndpoint`（复制 Runtime GraphQL Endpoint）
  - design: `true`
  - end_user: `false`
  - 说明：该能力偏开发调试用途，终端用户页面不展示。

## 落地约定

1. 所有 `ModelRecordWorkspace` 调用方必须显式传入 `workspaceMode`。
2. 新增设计态能力时，必须先进入能力矩阵，再渲染 UI。
3. 不允许在 end-user 页面通过 CSS/样式隐藏来替代能力开关。
4. 前端能力开关只负责“展示与交互边界”，后端鉴权仍是最终防线。

## 已接入文件

- `modelcraft-front/src/web/components/features/model-editor/ModelRecordWorkspace.tsx`
- `modelcraft-front/src/web/components/features/model-editor/ModelRecordInsertMenu.tsx`
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/ModelRecordTable.tsx`
- `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx`
- `modelcraft-front/src/app/u/[orgName]/[projectSlug]/data/page.tsx`

## 后续建议

后续可继续扩展能力矩阵：

- `canUseDesignGraphQLEndpoint`

每新增一项都必须同时更新本文档与调用方。
