## 1. Workspace 拆分

- [x] 1.1 `front-architect` 规划 `DevelopRecordWorkspace` 边界、目录位置与对外接口。
- [x] 1.2 `front-architect` 规划 `RuntimeRecordWorkspace` 边界、目录位置与对外接口。
- [x] 1.3 `front-worker` 新建 `DevelopRecordWorkspace`，承接当前 develop 入口所需的数据表、表单、结构维护与关系维护能力。
- [x] 1.4 `front-worker` 新建 `RuntimeRecordWorkspace`，承接当前 end-user 入口所需的 record 查询、新增、编辑、删除能力。
- [x] 1.5 `front-worker` 更新 develop 与 end-user 路由入口，删除 `workspaceMode` 对外接口及其入口传参。

## 2. Develop / Runtime 职责收敛

- [x] 2.1 `front-worker` 将字段插入、字段废弃/删除、关系维护等能力迁移到 develop workspace 自有实现。
- [x] 2.2 `front-worker` 从 runtime workspace 中移除 develop-only 控制与对应请求路径。
- [x] 2.3 `front-worker` 确保 runtime workspace 保持当前 record query/create/edit/delete 行为等价。
- [x] 2.4 `front-worker` 确保 develop workspace 保持当前模型调试与结构维护行为等价。

## 3. Shared Primitives 与 Access Adapter 重构

- [x] 3.1 `front-architect` 识别并标注当前 `model-record-form` 目录下的 shared primitives、develop-only 组件、runtime-only 组件。
- [x] 3.2 `front-architect` 为 develop/runtime workspace 设计 `record access adapter` 或等价注入边界。
- [x] 3.3 `front-worker` 改造仍需共享的远程 widgets，使其通过注入边界访问数据，而不是在组件内部直接创建 client。
- [x] 3.4 `front-worker` 清理 shared primitives 中对 `createModelRuntimeClient`、`createEndUserScopedClient`、auth store、路由推导身份的直接依赖。

## 4. 旧结构清理

- [x] 4.1 `front-worker` 删除或下线旧的 mode-driven 共享壳组件与无用分支逻辑。
- [x] 4.2 `front-worker` 清理与 `workspaceMode` 相关的类型、props、条件渲染与 helper。
- [x] 4.3 `front-architect` 与 `front-worker` 按新边界整理目录结构与导出入口，避免 develop/runtime 再次通过同一壳组件耦合。

## 5. 验收关键点

- [x] 5.1 `front-reviewer` 验证 develop 与 runtime 路由分别挂载独立 workspace，代码中不再存在 `ModelRecordWorkspace + workspaceMode` 的入口复用。
- [x] 5.2 `front-reviewer` 验证 runtime workspace 不显示字段插入、字段生命周期维护、关系维护等 develop-only 控制。
- [x] 5.3 `front-reviewer` 验证 develop workspace 仍可执行结构维护与关系维护，runtime workspace 仍可执行 record query/create/edit/delete。
- [x] 5.4 `front-reviewer` 验证托管模型只读限制未回归，develop/runtime 两侧都不会绕过既有只读约束。
- [x] 5.5 `front-reviewer` 验证被定义为 shared 的 primitives 不再直接 import runtime/end-user client factory 或 auth store。
- [x] 5.6 `front-reviewer` 补充至少一组针对 workspace 拆分后的回归验证，覆盖 develop 入口、runtime 入口与一个共享远程 widget 的访问路径。
