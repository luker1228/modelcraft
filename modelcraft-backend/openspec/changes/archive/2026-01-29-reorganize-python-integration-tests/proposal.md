# Proposal: Reorganize Python Integration Tests

## Change ID
`reorganize-python-integration-tests`

## Problem Statement

当前 `tests/` 目录的 Python 集成测试存在以下组织问题：

1. **目录结构不清晰**：
   - `design/` 和 `runtime/` 目录存在但使用不一致（design 有 2 个文件，runtime 只有 1 个）
   - `automated/` 目录与 `design/runtime` 平级，但所有测试都是自动化的，存在概念混淆
   - 根目录有多个工具脚本（`cleanup_test_data.py`, `health_check.py`, `config.py`）分散
   - `automated/` 中的 6 个测试文件应该归类到 `design/` 或 `runtime/`

2. **依赖关系不明确**：
   - Runtime 测试依赖 Design 测试的工具类和数据初始化
   - 但当前结构没有体现这种依赖关系
   - 缺少明确的共享代码位置

3. **共享代码缺失**：
   - 每个测试文件独立创建 GraphQL client
   - 缺少统一的 fixtures 和 utilities
   - 没有 `conftest.py` 来共享测试配置
   - Design 和 Runtime 之间共享工具的位置不明确

4. **可扩展性问题**：
   - 作为中小型项目，当前结构在测试增长到 20-30 个文件时会难以维护
   - `design/` 目录下只有 2 个测试，但应该有更多（如 `automated/` 下的项目、集群测试）
   - 新测试的添加位置不明确（应该放 design 还是 runtime？）

## Proposed Solution

采用 **"Design-Runtime 分层 + 领域组织"** 策略，明确测试阶段依赖关系：

### 核心理念

1. **Design-Time 测试**: 测试模型设计、集群配置、项目管理等"设计时"操作
2. **Runtime 测试**: 测试运行时数据查询、GraphQL 执行等，**依赖** Design 阶段的初始化
3. **明确依赖**: Runtime 可以使用 Design 的工具类和 fixtures

### 目录结构设计

```
tests/
├── design/                             # Design-Time 测试（设计阶段）
│   ├── __init__.py
│   ├── conftest.py                     # Design 测试共享 fixtures
│   ├── README.md                       # Design 测试说明
│   │
│   ├── common/                         # Design 共享工具（可被 runtime 复用）
│   │   ├── __init__.py
│   │   ├── graphql_client.py          # GraphQL 客户端封装
│   │   ├── fixtures.py                # 共享 fixtures
│   │   ├── test_data.py               # 测试数据生成器
│   │   └── assertions.py              # 自定义断言助手
│   │
│   ├── project/                        # 项目管理测试
│   │   ├── __init__.py
│   │   ├── test_project_crud.py       # 项目 CRUD（从 automated/ 移入）
│   │   └── test_project_isolation.py  # 项目隔离（从 automated/ 移入）
│   │
│   ├── cluster/                        # 数据库集群测试
│   │   ├── __init__.py
│   │   └── test_cluster_graphql.py    # 集群操作（已存在）
│   │
│   ├── model/                          # 模型设计测试
│   │   ├── __init__.py
│   │   ├── test_model_graphql.py      # 模型 GraphQL 操作（已存在）
│   │   ├── test_jsonschema_export.py  # JSON Schema 导出（从 automated/ 移入）
│   │   └── test_schema_operations.py  # Schema 操作（从 automated/ 移入）
│   │
│   └── compatibility/                  # 向后兼容性测试
│       ├── __init__.py
│       └── test_backward_compatibility.py  # 从 automated/ 移入
│
├── runtime/                            # Runtime 测试（运行时阶段，依赖 design）
│   ├── __init__.py
│   ├── conftest.py                     # Runtime 测试共享 fixtures
│   ├── README.md                       # Runtime 测试说明
│   │
│   ├── query/                          # 数据查询测试
│   │   ├── __init__.py
│   │   └── test_user_graphql.py       # 用户数据查询（已存在）
│   │
│   └── integration/                    # 完整流程集成测试
│       ├── __init__.py
│       └── test_modelcraft_client.py  # 客户端测试（从 automated/ 移入）
│
├── common/                             # 跨阶段共享工具
│   ├── __init__.py
│   ├── config.py                      # 配置管理（从根目录移入）
│   ├── health_check.py                # 健康检查工具（从根目录移入）
│   └── cleanup_test_data.py           # 数据清理脚本（从根目录移入）
│
├── manual/                             # 手动测试资源（保持不变）
│   ├── README.md
│   ├── curl_commands.md
│   └── ...
│
├── conftest.py                         # 根级别 pytest 配置
├── pytest.ini                          # Pytest 配置
├── README.md                           # 测试套件总览
├── requirements.txt                    # Python 依赖
├── Taskfile.yml                        # Task 命令定义
├── Makefile                            # Make 命令（向后兼容）
├── reports/                            # 测试报告输出目录
└── envs/                               # 环境配置文件
    ├── local.env
    └── docker.env
```

### 关键设计决策

1. **Design-Runtime 分层**：
   - **明确测试阶段**：`design/` 测试设计时操作，`runtime/` 测试运行时查询
   - **依赖方向清晰**：runtime 可以导入使用 `design/common/` 的工具类
   - **独立运行**：`pytest design/` 或 `pytest runtime/` 可独立执行

2. **领域驱动组织**：
   - Design 阶段按业务领域（project, cluster, model）组织
   - Runtime 阶段按功能（query, integration）组织
   - 与主项目的 DDD 架构保持一致

3. **三层共享代码结构**：
   - `tests/common/` - 跨阶段共享（config, health_check, cleanup）
   - `design/common/` - Design 阶段工具（GraphQL client, fixtures），**可被 runtime 复用**
   - `runtime/conftest.py` - Runtime 专属 fixtures

4. **去除 automated 目录**：
   - 所有测试都是自动化的，不需要 `automated/` 层级
   - 减少一层不必要的嵌套
   - 测试分类更清晰：design 还是 runtime？

5. **保持简洁**：
   - 最多 2-3 层嵌套：`design/domain/test_*.py` 或 `runtime/category/test_*.py`
   - 每个子目录包含 `__init__.py` 使其成为 Python 包

6. **未来扩展路径**：
   - Design 领域测试增长：在 `design/{domain}/` 内再分类
   - Runtime 功能增加：在 `runtime/` 下添加新类别（如 `mutation/`, `subscription/`）

## Benefits

1. **清晰的测试阶段划分**：
   - Design 和 Runtime 分离，测试目标明确
   - 依赖关系清晰：Runtime 依赖 Design 的工具和初始化
   - 新开发者能快速理解测试结构

2. **减少代码重复**：
   - `design/common/` 提供共享的 fixtures 和工具
   - Runtime 可以复用 Design 的工具类
   - `tests/common/` 提供跨阶段的配置和清理工具

3. **更好的可维护性**：
   - 每个阶段和领域的测试独立演化
   - 清晰的依赖关系（通过 Python 包导入）
   - 去除不必要的 `automated/` 层级

4. **适应项目增长**：
   - 结构支持从当前 9 个测试文件增长到 30+ 个
   - Design 阶段可按领域扩展
   - Runtime 阶段可按功能扩展

5. **与主项目一致**：
   - Design/Runtime 划分镜像主项目架构
   - 领域组织与 DDD 结构对应
   - 降低认知负担

## Risks and Mitigations

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 重构过程中测试失效 | 高 | 分阶段迁移，每步后运行测试验证 |
| CI/CD 脚本需要更新 | 中 | 保持 pytest 命令不变，只更新导入路径 |
| 开发者习惯改变 | 低 | 提供迁移指南和新结构文档 |
| `utils/` 目录引入额外复杂度 | 低 | 保持 `utils/` 简洁，只放非测试代码 |

## Out of Scope

以下内容不在本次重构范围内：

- 测试用例的内容修改（只移动，不改写）
- 测试框架或依赖的升级
- `manual/` 目录的重组（保持现状）
- 性能测试或其他新测试类型的添加

## Success Criteria

- [ ] 所有现有测试文件按新结构重组完成
- [ ] 所有测试在新结构下通过（`pytest automated/ -v`）
- [ ] 创建 `conftest.py` 和 `common/` 共享代码
- [ ] CI/CD 流程无中断（`task auto-test` 正常工作）
- [ ] 更新 `tests/README.md` 和 `tests/automated/README.md` 文档
- [ ] 提供迁移指南（如何添加新测试）

## Implementation Notes

- 使用 Git 的 `git mv` 命令保留文件历史
- 一次迁移一个领域，逐步验证
- 在 `conftest.py` 中提取现有的重复 fixture 代码
- 确保所有 `__init__.py` 存在以支持 Python 包导入
