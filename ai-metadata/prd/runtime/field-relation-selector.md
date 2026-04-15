# 外键字段关联选择器 PRD

> 依赖领域文档：@../../backend/design/domain-model/8-runtime/jsonschema-contract.md
> `x-relation` 元数据结构和 `x-mc.widget: "relation-selector"` 规范见上述文档。

## 是什么问题

用户在运行时向多对一关系表单中写入外键值时，必须先打开另一张表查到目标记录、手动复制 id，再回来粘贴到 string 输入框里。这个流程在日常使用中摩擦极高，尤其在没有对外键字段任何视觉提示的情况下，用户根本看不出这个 string 框背后是一条关联记录。

根本原因：运行时 GraphQL schema 目前对"普通 string 字段"和"被外键引用的 string 字段"一视同仁，前端没有任何元信息来区分两者，无法升级渲染为选择控件。

---

## 目标用户

**数据录入者**（运营、产品、测试等非开发角色）：在 ModelCraft 管理界面直接填表，需要建立记录之间的关联关系。当前他们要么依赖开发人员告知 id，要么自己翻表查找，效率极低。

**开发者/建模者**：在调试阶段需要快速插入测试数据，手动 copy id 是重复且低价值的操作。

---

## 目标与成功标准

- **目标**：外键字段在运行时表单中自动升级为关联选择器，用户通过搜索 + 下拉即可完成关联，不再需要手动 copy id。
- **指标**：
  - 新记录创建时外键字段完成率（不留空 / 不填错格式）提升
  - 用户反馈中"需要手动复制 id"的抱怨消失
  - 关联选择的操作路径从"打开另一页 → 复制 → 返回 → 粘贴"缩短为"点击下拉 → 搜索 → 选中"

---

## 用户故事

- 作为数据录入者，我希望在外键字段上看到一个下拉搜索框，以便直接选择关联记录，而不需要知道目标记录的 id 是什么。
- 作为建模者，我希望设计时定义的关联关系能在运行时表单中自动生效，不需要额外配置。
- 作为数据录入者，我希望下拉选项能显示目标记录的可读名称（`__label`），以便我能认出要选的是哪条。

---

## 功能范围（Must-have）

### 1. 设计态：运行时 schema 携带外键元信息

当字段被 `LogicalForeignKey` 定义为多对一关系的外键端（`Direction = normal`）时，在运行时动态 JSON Schema 中为该字段打上 `x-relation` 标记。

**最小必要属性**（仅保留前端构建 GraphQL 查询所必须的信息）：

```json
"x-relation": {
  "databaseName": "users_db",
  "modelName": "User"
}
```

| 属性 | 来源（LFK 字段） | 作用 |
|------|----------------|------|
| `databaseName` | `RuntimeModel.DatabaseName`（通过 `RefModelID` 查到目标模型后取） | 构建运行时 GraphQL endpoint：`/graphql/.../db/{databaseName}/model/{modelName}` |
| `modelName` | `LogicalForeignKey.RefModelName` | 同上，endpoint 的 model 段 + 查询操作名前缀 |

**为什么只需要这两个字段：**
运行时 GraphQL endpoint 格式为：
```
POST /graphql/org/{orgName}/project/{projectSlug}/db/{databaseName}/model/{modelName}
```
前端已知当前的 `orgName` 和 `projectSlug`（来自页面上下文），只缺 `databaseName` 和 `modelName`。有这两个值，前端即可直接发起 `findMany` 查询拉取候选列表，无需任何其他信息。

目标字段（通常是 `id`）无需显式标记——前端选中后写入的始终是目标记录的 `id` 字段值，这是关系字段的固定约定。

**为什么用元信息而非改字段类型：** 外键字段在数据库和 schema 层面就是 string，改类型会破坏现有 schema 兼容性。通过扩展元信息是成本最低、影响最小的方式，前端按需读取后决定渲染策略。

### 2. 运行时：关联选择器控件

前端在读取到字段带有 `x-relation` 标记后，将该字段渲染为**关联选择器**，而非普通文本输入框：

- 点击展示下拉面板
- 支持关键词搜索（调用目标模型的 `findMany` 接口，以 `__label` 字段做 contains 过滤）
- 下拉列表每项展示：`__label`（主要）+ id（次要，辅助识别）
- 选中后将目标记录的 id 写入字段值（实际存储的仍是 id）
- 已有值时展示 `__label(id)` 而非裸 id，与 `__label` 展示协议保持一致

### 3. 空数据与边界处理

- 目标模型无数据时，下拉提示"暂无可选记录"，不报错
- 目标模型 `displayField` 未配置时，`__label` 为空，降级展示为 id
- 关联选择器支持清空（nullable 外键字段）

---

## 不做什么（Out of scope）

- **一对多（反向）字段**不做选择器升级，仅处理多对一的外键端（SourceField）
- **批量关联**（多选）本期不做
- 运行时搜索的防抖、分页、无限滚动等高级交互优化
- 目标模型权限拦截（当前运行时不做细粒度权限，选择器直接读取，与当前行为一致）
- 关联选择器的"快速新建"功能（先选再跳转新建目标记录）

---

## 验收标准

- [ ] 存在 `LogicalForeignKey` 关联的外键字段，在运行时表单中渲染为关联选择器（而非普通 input）
- [ ] 普通 string 字段不受影响，仍渲染为文本输入框
- [ ] 关联选择器下拉展示目标模型记录的 `__label`（及 id 辅助）
- [ ] 选中后字段实际写入目标记录 id，运行时 insert/update 操作正常
- [ ] 目标模型 `displayField` 未配置时，选择器仍可使用，降级展示 id
- [ ] 目标模型无记录时，下拉给出友好提示，不报错
- [ ] 关系元信息（`x-relation`）在 schema 中正确携带且结构稳定，不影响非关联字段

---

## 关键设计决策

### 为什么在 jsonSchema/schema 层标记，而不是前端硬编码？

硬编码方案需要前端单独查询 `LogicalForeignKey` 接口并对每个字段做匹配，增加了额外的网络请求和耦合。将元信息附在 schema 上，是 "schema 即契约" 的思路——前端只需消费 schema 就能知道如何渲染，无需感知设计态的实现细节。这也为后续更多字段类型的 UI 升级（如日期选择器、枚举选择器）建立统一的扩展模式。

### `x-relation` 的附加位置

优先附加在运行时 GraphQL schema 的字段描述（description 扩展）或通过独立的 schema introspection 字段暴露。具体形式待技术评审确认，PRD 不约束实现手段。

---

## 依赖

- **前置依赖**：`__label` 展示协议（`ai-metadata/prd/field/00-field-label-field.md`）需已实现，因为选择器下拉展示依赖目标模型的 `__label`。若 `__label` 未就绪，选择器降级为展示 id，仍可先行开发。
- `LogicalForeignKey` 设计态数据（已有）：运行时 schema 生成逻辑需要读取 LFK 关系确定哪些字段是外键端。

---

## 待确认

- [ ] `x-relation` 元信息的具体附加形式（GraphQL schema description JSON 扩展 vs 单独的 introspection 字段 vs REST 附属 schema endpoint）——需技术评审拍板
- [ ] 当目标模型未同步到运行时（`DeploymentStatus` 非 ready）时，选择器如何处理：报错、禁用还是展示 warning？
- [ ] 关联选择器的搜索触发方式：输入即搜（防抖 300ms）还是点击后加载全量再本地过滤？（与目标模型数据量有关）
