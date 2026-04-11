# Field LabelField 设计（`__label` 展示协议）

## 是什么问题
当前运行时查询在关系字段上默认拼接 `id + name`，当被关联模型不存在 `name` 字段时会直接触发 GraphQL 校验错误，导致列表/表单页面无法渲染。

典型问题链路：
- 前端关系字段默认选择写死在 `buildFieldSelections`
- 被关联模型真实可展示字段并不一定叫 `name`
- 前端缺少稳定、通用的“展示文本”协议

## 目标
1. 定义一个稳定的关系展示协议，避免前端猜字段名。
2. 统一多对一 / 一对多场景下的展示读取方式。
3. 保持现有关系结构不变，最小成本落地。

## 范围
### 本次要做
- 采用**模型级默认展示字段（A 方案）**。
- 在设计时模型协议中新增 `displayField`（由**被关联模型 B**配置）。
- 在运行时动态模型对象统一提供 `__label` 字段（`String!`）。
- 前端关系查询统一为 `id + __label`。
- 前端展示规则统一为：`__label(id)`；当 `__label` 为空字符串时展示 `空(id)`。

### 本次不做
- 不新增关系级覆盖字段（如 `relationDisplayField`）。
- 不新增 `__labels` 字段。
- 不改变关系字段的返回结构（对象 / 对象数组保持不变）。

## 接口协议

### 1) 设计时协议（Project GraphQL）
在 Model 上新增模型级展示字段配置：

```graphql
type Model implements Node {
  # ...existing fields
  displayField: String
}

input CreateModelInput {
  # ...existing fields
  displayField: String
}

input UpdateModelMetaInput {
  # ...existing fields
  displayField: String
}
```

语义：
- `displayField` 表示该模型默认用于展示的字段名（例如 `title`）。
- 在 A 模型关联 B 模型场景下，关系返回对象的 `__label` 由 **B 模型的 `displayField`** 决定。
- 本方案不定义回退链路，`__label` 仅由 `displayField` 解析。

### 2) 运行时协议（Runtime GraphQL）
所有运行时动态对象统一暴露：

```graphql
__label: String!
```

语义：
- `__label` 是**记录级展示文本**。
- 多对一返回单对象时，读取 `relation.__label`。
- 一对多返回对象数组时，读取 `relation[i].__label`。
- `__label` 本身不是数组类型；数组来自关系返回类型。

### 3) `__label` 解析规则
解析来源（唯一）：
1. 读取当前记录所属模型的 `displayField` 配置。
2. 使用该字段在当前记录中的值作为 `__label`。

值规范：
- `__label` 类型为 `String!`。
- 源值允许 string / number / boolean / date-time，统一转字符串返回。
- 当源值为 `null`、空串、对象、数组，统一返回空字符串 `""`。
- 不做 `title/name/id` 回退。

### 4) 查询示例
多对一：

```graphql
query {
  findMany {
    items {
      id
      tl {
        id
        __label
      }
    }
  }
}
```

一对多：

```graphql
query {
  findMany {
    items {
      id
      children {
        id
        __label
      }
    }
  }
}
```

## 切换策略（直接切换）
- 不做分阶段迁移，直接将前端关系字段查询从 `['id', 'name']` 切换为 `['id', '__label']`。
- 同步上线 `displayField` + `__label` 协议后，以新协议为唯一标准。
- 新功能与存量页面均禁止继续依赖 `name` 作为关系展示字段。

## 验收标准
1. 前端关系查询仅包含 `id` 与 `__label`，不再查询 `name`。
2. 多对一、一对多页面均可渲染展示文本，且前端只依赖 `__label`。
3. A 关联 B 时，B 模型 `displayField` 配置生效：修改后 `__label` 输出同步变化。
4. 当 `displayField` 对应值为空或不可用时，`__label` 返回空字符串，不做回退。
5. 前端展示遵循 `__label(id)`：当 `__label` 为空时展示 `空(id)`。

## 待确认
- [ ] `displayField` 的管理入口放在模型基础设置还是字段管理页。
- [ ] 若模型未配置 `displayField`，创建/发布流程是阻断还是仅告警。
- [ ] 是否需要审计日志记录 `displayField` 变更历史。
