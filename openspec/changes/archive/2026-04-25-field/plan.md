# Field LabelField 后端实现计划（plan）

## 目标
围绕已确认协议完成后端实现：
1. A 模型关联 B 模型时，关系对象 `__label` 由 **B 模型 `displayField`** 解析。
2. 前端关系查询只使用 `id + __label`。
3. 不做兼容迁移与回退链路（不再依赖 `name`，不做 `title/name/id` fallback）。

---

## 已确认协议约束（冻结）
- 直接切换：不做 phased migration，不保留旧展示协议。
- `__label` 无 fallback：仅按 `displayField` 取值。
- `__label` 类型固定 `String!`。
- `displayField` 配置在**被关联模型**（B 模型）上。
- 当取值不可用（null/空串/对象/数组）时，后端返回空字符串 `""`。

---

## 后端改动范围

### 1) 设计时协议：Model 增加 displayField
**文件**
- `modelcraft-backend/api/graph/project/schema/model.graphql`

**改动**
- `type Model` 新增 `displayField: String`
- `CreateModelInput` 新增 `displayField: String`（可选）
- `UpdateModelMetaInput` 新增 `displayField: String`（可选）

**实现点**
- Resolver/adapter 透传 `displayField`（create/update/query）
- 校验规则：字段名必须存在于该模型字段集合中；不合法返回业务错误

**强校验规则（仅后端实现）**
1. **写入校验（create/update model）**
   - `displayField` 必须非空且必须命中当前模型字段集合。
   - 命中字段类型必须可字符串化（string/number/boolean/date-time）；对象/数组类型禁止设置为 `displayField`。
2. **字段变更校验（remove/rename field）**
   - 若目标字段被当前模型用作 `displayField`，删除或重命名必须阻断。
   - 提示调用方先更新模型 `displayField`，再执行字段变更。
3. **发布/可用性校验**
   - 模型进入可用状态前，`displayField` 必须已配置且仍然有效。
   - 若未配置或失效，阻断发布（返回明确业务错误码）。
4. **并发一致性**
   - `displayField` 更新与字段变更操作需走同一事务边界下的二次校验，避免并发产生悬挂引用。

---

### 2) 持久化：models 表增加 display_field
**文件**
- `modelcraft-backend/db/schema/mysql/03_model_domain.sql`
- `modelcraft-backend/db/queries/model.sql`

**改动**
- `models` 表新增 `display_field` 可空列（varchar）
- create / update / get 查询补充该列
- 执行 `just generate-sqlc` 更新 dbgen

**实现点**
- domain model / repository conversion 增加 `displayField` 字段映射
- 读写一致性：create、update、get/list 都返回同一值

---

### 3) runtime domain：RuntimeModel 增加 displayField
**文件**
- `modelcraft-backend/internal/domain/modelruntime/runtimemodel.go`
- `modelcraft-backend/internal/infrastructure/repository/sql_modelruntime_repository.go`

**改动**
- `RuntimeModel` 新增 `DisplayField`
- `DbgenModelToRuntimeModel` 映射 `display_field -> DisplayField`

---

### 4) runtime GraphQL：注入 __label 字段
**文件**
- `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`
- （可选）`modelcraft-backend/internal/domain/modelruntime/graphql_constants.go`

**改动**
- 在动态对象构建（`generateModelType`）中统一注入 `__label: String!`
- `__label` resolver 逻辑：
  1. 读取当前对象所属模型的 `DisplayField`
  2. 从当前记录 map 中取该字段值
  3. 可字符串化则返回字符串
  4. 否则返回 `""`

**注意**
- `__label` 是记录级字段，不区分多对一/一对多
- 一对多仅是对象数组，每个元素独立解析 `__label`

---

## 行为定义（后端）

### 成功路径
- B 模型 `displayField=title`，记录 `title="foo"` -> `__label="foo"`
- B 模型 `displayField=code`，记录 `code=123` -> `__label="123"`

### 空值路径
- `displayField` 对应值为 `null`/`""`/对象/数组 -> `__label=""`

### 非法配置路径
- `displayField` 指向不存在字段：
  - 在配置时（create/update model）拦截并返回业务错误
  - 运行时不做 fallback
- `displayField` 目标字段被删除/重命名：
  - 在字段变更时拦截并返回业务错误
  - 要求先修改 `displayField` 再执行字段变更
- `displayField` 未配置或配置失效：
  - 在模型发布/可用化流程拦截并返回业务错误

---

## 测试计划

### 单元测试
- model domain：`displayField` 校验与更新逻辑
- runtime resolver：`__label` 字符串化、空值处理、对象/数组处理

### 回归点
- 多对一关系返回对象包含 `__label`
- 一对多关系返回数组对象，每个元素包含 `__label`

---

## 实施顺序
1. 改 DB schema + SQL query，生成 sqlc。
2. 改 domain/infra 映射（Model / RuntimeModel）。
3. 改 project GraphQL schema，生成 gql。
4. 实现 create/update/query 的 `displayField` 读写与校验。
5. 实现 runtime `__label` 注入与 resolver。
6. 跑单测 + 回归验证。

---

## 风险与决策
- 本方案不提供旧协议兼容能力，发布后前端必须同步使用 `id + __label`。
- 本方案不提供 fallback，可确保规则单一，但要求模型必须正确配置 `displayField`。
- 若未配置 `displayField`，建议在模型发布流程阻断（由产品最终拍板）。
