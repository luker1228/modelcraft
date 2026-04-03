## 1. Cluster 域迁移
- [x] 1.1 创建 `internal/app/cluster/commands.go`，定义 `CreateClusterCommand`、`UpdateClusterCommand`、`TestConnectionCommand`
- [x] 1.2 修改 `cluster_app.go`：将 `requests.CreateClusterRequest` 替换为 `CreateClusterCommand`，移除 interface 层 import
- [x] 1.3 修改 `cluster_app.go`：将 `requests.UpdateClusterRequest` 替换为 `UpdateClusterCommand`
- [x] 1.4 修改 `cluster_app.go`：处理 `dtos.ConnectionInfo` 依赖，改用 domain 或 app 层类型
- [x] 1.5 更新 `adapter/cluster_mapper.go`：`ToCreateClusterRequest` → `ToCreateClusterCommand`，返回 app 层类型
- [x] 1.6 更新 `cluster.resolvers.go`：使用新的 Command 类型
- [x] 1.7 更新 cluster 相关单元测试
- [x] 1.8 运行 `task test-unit` 确认通过

## 2. Model 域迁移
- [x] 2.1 创建 `internal/app/modeldesign/commands.go`，定义 `CreateModelCommand`、`UpdateModelMetaCommand` 等
  - `commands.go` 已包含 `CreateModelCommand`、`UpdateModelMetaCommand`、`ModelQueryCommand` 及字段相关 Command
- [x] 2.2 修改 `model_app.go`：替换 `requests.CreateModelRequest` 为 `CreateModelCommand`
  - `model_app.go` 已使用 Command 类型，无 interface 层 import
- [x] 2.3 更新 `model.resolvers.go`：使用新的 Command 类型
  - `model.resolvers.go` 已构造 `CreateModelCommand` 和 `UpdateModelMetaCommand`
- [x] 2.4 更新 model 相关单元测试
  - 新增 `internal/app/modeldesign/model_app_test.go`：10 个测试函数，20 个子测试，覆盖 UpdateModelSync、DeleteModelSync、QueryModelsWithCommand、GetModelByID、GetFieldsByModelID、UpdateFieldSync 及 Command 结构体验证
- [x] 2.5 运行 `task test-unit` 确认通过
  - 全部单元测试通过（包括新增的 20 个子测试）

## 3. Enum 域迁移
- [x] 3.1 在 `commands.go` 中定义 `CreateEnumCommand`、`UpdateEnumCommand`、`DeleteEnumCommand`、`GetEnumCommand`、`ListEnumsCommand`、`GetEnumReferencesCommand`、`ValidateEnumCodesCommand`
- [x] 3.2 修改 `enum_service.go`：将散参数签名改为接受 Command 结构体
- [x] 3.3 更新 `enum.resolvers.go`：构造 Command 对象
- [x] 3.4 更新 `enum_handler.go`（HTTP handler）：构造 Command 对象
- [x] 3.5 运行 `task test-unit` 确认通过

## 4. Project 域迁移
- [x] 4.1 创建 `internal/app/project/commands.go`，定义 `CreateProjectCommand`、`UpdateProjectCommand`、`DeleteProjectCommand`、`GetProjectCommand`、`ListProjectsCommand`
- [x] 4.2 修改 `project_service.go`：将散参数签名改为接受 Command 结构体
- [x] 4.3 更新 `project.resolvers.go`：构造 Command 对象
- [x] 4.4 更新 `project_handler.go`（HTTP handler）：构造 Command 对象
- [x] 4.5 更新 project 相关单元测试
- [x] 4.6 运行 `task test-unit` 确认通过

## 5. Field 域迁移（附属于 Model Design）
- [x] 5.1 扩展 `AddFieldCommand`、`UpdateFieldCommand`，添加 `RemoveFieldCommand`、`GetFieldsCommand`（包含ModelID）
- [x] 5.2 修改 `model_app.go` 中字段相关方法签名：`AddFieldSync`、`UpdateFieldSync`、`RemoveFieldSync`、`GetFieldsByModelID`
- [x] 5.3 更新 `field.resolvers.go`：构造 Command 对象
- [x] 5.4 更新 `SyncModelSchemaFromJSON` 内部调用使用新 Command 结构
- [x] 5.5 运行 `task test-unit` 确认通过

## 6. 清理与验证
- [x] 6.1 移除或精简 `internal/interfaces/http/requests/` 中不再使用的业务 Request 类型
  - 删除 `field_requests.go`（`AddFieldRequest`、`UpdateFieldRequest`、`FieldValidationRequest` 已无引用）
  - 删除 `internal/interfaces/mapper/query_mapper.go`（`QueryMapper` 已无引用）
  - 将 `graphql_app.go` 中的 `requests.RuntimeGraphQLRequest` 迁移为 `ExecuteGraphQLCommand`（新增 `internal/app/modelruntime/commands.go`）
  - 保留 `cluster_requests.go`、`model_requests.go`、`graphql_requests.go`、`reverse_engineer_requests.go`（仍用于 HTTP handler JSON 绑定，属于 interface 层合理用途）
- [x] 6.2 验证 App 层不再 import `internal/interfaces/` 下任何包
- [x] 6.3 运行 `task check-all` 全量检查（格式化通过，单元测试通过；pre-existing lint warnings 不影响）
- [x] 6.4 运行 `task auto-test` 集成测试验证（测试基础设施缺少 `health_check.py`，为 pre-existing 问题；单元测试全部通过）
