## ADDED Requirements

### Requirement: 数据面访问 MUST 使用三层身份模型
系统 MUST 在 runtime/data-plane 能力中使用三层身份模型：组织级身份、项目级身份、用户级身份。组织级与项目级身份表示可进入数据面的管理主体；用户级身份表示实际的数据使用主体。

#### Scenario: 用户级身份访问数据面
- **WHEN** 一个用户级主体访问 database catalog、model catalog、`meta/user` 或 runtime model data
- **THEN** 系统将该请求解释为用户级数据访问请求

#### Scenario: 上层身份访问数据面
- **WHEN** 一个组织级或项目级主体访问 database catalog、model catalog、`meta/user` 或 runtime model data
- **THEN** 系统将该请求解释为管理主体进入数据面的请求

### Requirement: 数据面 principal MUST 同时表达主体、范围、凭证与访问模式
系统 MUST 将数据面 principal 解析为统一结构，至少包含身份层级、作用范围、凭证类型与访问模式；系统 MUST NOT 仅根据某一种 token 类型直接决定全部数据访问语义。

#### Scenario: 不同凭证映射到同一访问模式
- **WHEN** 两个请求分别携带不同的 credential type，但声明相同的主体层级、作用范围与访问模式
- **THEN** 系统按同一数据面授权语义处理这两个请求

#### Scenario: token 类型不足以直接判定访问语义
- **WHEN** 某请求仅提供 credential type 而未提供可解析的主体层级或访问模式
- **THEN** 系统拒绝将其视为合法的数据面 principal

### Requirement: 数据面 MUST 统一支持三种访问模式
系统 MUST 为数据面能力统一支持三种访问模式：用户本人访问、管理员扮演某个用户访问、管理员全权限访问。

#### Scenario: 用户本人访问
- **WHEN** 用户级主体以自己的身份访问数据面
- **THEN** 系统按用户本人访问模式执行权限校验与数据过滤

#### Scenario: 管理员扮演用户访问
- **WHEN** 组织级或项目级主体携带某个具体用户的有效上下文访问数据面
- **THEN** 系统按管理员扮演用户访问模式执行权限校验与数据过滤

#### Scenario: 管理员全权限访问
- **WHEN** 组织级或项目级主体以全权限模式访问数据面
- **THEN** 系统按管理员全权限访问模式执行数据面授权判断

### Requirement: 任何用户级数据面能力 MUST 可被上层身份兼容访问
系统 MUST 保证：任何可被用户级主体访问的数据面能力，在满足上下文条件时也可被组织级或项目级主体访问；其中依赖具体用户主体的能力 MUST 仅允许用户本人访问或管理员扮演用户访问，不依赖具体用户主体的能力 MAY 允许管理员全权限访问。

#### Scenario: 依赖具体用户主体的能力
- **WHEN** 某数据面能力依赖当前具体用户主体，例如 `me` 或基于 owner/RLS 的主体绑定能力
- **THEN** 系统仅允许用户本人访问或管理员扮演用户访问

#### Scenario: 不依赖具体用户主体的能力
- **WHEN** 某数据面能力不依赖当前具体用户主体，例如 catalog 查询或按条件查询某个用户记录
- **THEN** 系统允许管理员全权限访问

### Requirement: 用户级主体在查看数据前 MUST 先具备 catalog 可见性
系统 MUST 将 database catalog、model catalog 与 model schema subset 视为用户级数据访问的基础能力。若用户级主体无对应 catalog 可见性，系统 MUST NOT 允许其直接进入具体 runtime 数据查询或写入。

#### Scenario: 先查看 catalog 再查看数据
- **WHEN** 一个用户级主体被授权访问某个 project 下的运行态数据
- **THEN** 该主体先能够查看其授权范围内的 database catalog 与 model catalog
- **THEN** 在此基础上才允许其进入具体 model 的数据查询或写入

#### Scenario: 无 catalog 可见性时拒绝数据访问
- **WHEN** 一个用户级主体对某 project 不具备 catalog 可见性
- **THEN** 系统不得允许其直接访问该 project 下的 runtime model data
