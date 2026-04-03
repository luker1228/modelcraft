# Spec: Python Integration Test Organization

## Context

ModelCraft 项目使用 Python + pytest 编写集成测试，位于 `tests/` 目录。当前测试规模为 6-10 个文件，作为中小型项目需要一个可扩展的组织结构以支持未来增长到 20-30 个测试文件。

## ADDED Requirements

### Requirement: 领域驱动的测试目录结构

**ID**: `TEST-ORG-001`

**Description**:
测试文件 **SHALL** 按业务领域组织，镜像主代码库的 DDD 架构，确保测试结构与应用结构保持一致性。每个业务领域（project, model, cluster 等）**MUST** 有独立的子目录包含相关测试。

#### Scenario: 开发者查找项目管理相关测试

**Given**:
- 开发者正在开发项目管理功能
- 测试位于 `tests/automated/` 目录

**When**:
- 开发者查看测试目录结构

**Then**:
- 应存在 `tests/automated/project/` 目录
- 该目录包含所有项目管理领域的测试文件
- 文件命名遵循 `test_*.py` 约定

**Example**:
```
tests/automated/project/
├── __init__.py
├── test_project_crud.py      # 项目 CRUD 操作测试
└── test_project_isolation.py # 项目隔离验证测试
```

#### Scenario: 按业务领域组织模型测试

**Given**:
- 模型领域包含设计时和运行时两类操作
- 相关测试文件分散在不同位置

**When**:
- 测试文件按新结构组织

**Then**:
- 存在 `tests/automated/model/` 目录
- 包含所有模型相关测试：设计、运行时、导出、schema 操作
- 测试文件名清晰表达其测试范围

**Example**:
```
tests/automated/model/
├── __init__.py
├── test_model_design.py      # 模型设计操作
├── test_model_runtime.py     # 运行时查询
├── test_jsonschema_export.py # JSON Schema 导出
└── test_schema_operations.py # 基于 schema 的操作
```

### Requirement: 共享测试基础设施

**ID**: `TEST-ORG-002`

**Description**:
项目 **MUST** 提供集中的共享测试工具和 fixtures，减少代码重复。所有重复的 GraphQL 客户端创建代码 **SHALL** 提取到 `common/graphql_client.py`，共享 fixtures **SHALL** 定义在 `conftest.py` 中。

#### Scenario: 多个测试共享 GraphQL 客户端

**Given**:
- 多个测试文件需要 GraphQL 客户端
- 客户端初始化逻辑相同

**When**:
- 使用共享的 GraphQL 客户端工具

**Then**:
- 存在 `tests/automated/common/graphql_client.py` 提供客户端创建函数
- 存在 `tests/automated/conftest.py` 提供 pytest fixtures
- 测试文件通过 fixture 获取客户端，无需重复初始化代码

**Example**:
```python
# tests/automated/conftest.py
import pytest
from tests.automated.common.graphql_client import create_graphql_client
from tests.utils.config import config

@pytest.fixture(scope="session")
def graphql_client():
    """提供会话级别的 GraphQL 客户端"""
    return create_graphql_client(config.get_design_graphql_url())

# tests/automated/project/test_project_crud.py
def test_create_project(graphql_client):
    # 直接使用 fixture 提供的客户端
    result = graphql_client.execute(CREATE_PROJECT_MUTATION, ...)
```

#### Scenario: 测试数据清理 fixture

**Given**:
- 多个测试创建项目资源
- 测试完成后需要清理资源

**When**:
- 使用共享的清理 fixture

**Then**:
- `conftest.py` 提供资源追踪 fixture
- Fixture 使用 `yield` 模式在测试后自动清理
- 即使测试失败，清理逻辑也会执行

**Example**:
```python
# tests/automated/project/conftest.py
import pytest

@pytest.fixture(scope="module")
def created_projects(graphql_client):
    """追踪模块中创建的项目，测试后自动清理"""
    projects = []
    yield projects
    # 清理阶段
    for project_id in projects:
        try:
            delete_project(graphql_client, project_id)
        except Exception as e:
            print(f"清理项目 {project_id} 失败: {e}")
```

### Requirement: 工具脚本与测试用例分离

**ID**: `TEST-ORG-003`

**Description**:
测试支持工具（配置加载、健康检查、数据清理脚本）**MUST** 与测试用例分开存放在 `tests/utils/` 目录，避免被 pytest 错误识别为测试。工具脚本文件 **SHALL NOT** 使用 `test_` 前缀命名。

#### Scenario: 配置模块不被 pytest 收集

**Given**:
- `config.py` 是配置加载工具，不是测试用例
- Pytest 自动发现 `tests/` 目录下的所有 `.py` 文件

**When**:
- 运行 `pytest automated/ --collect-only`

**Then**:
- `config.py` 不出现在收集的测试列表中
- 配置模块位于 `tests/utils/config.py`
- 测试可以通过 `from tests.utils.config import config` 导入

**Example**:
```
tests/
├── automated/           # Pytest 搜索此目录
│   └── project/
│       └── test_*.py
└── utils/               # 工具脚本（pytest 不搜索）
    ├── __init__.py
    ├── config.py
    ├── health_check.py
    └── cleanup_test_data.py
```

#### Scenario: 手动运行数据清理脚本

**Given**:
- 测试数据残留在数据库中
- 需要手动清理

**When**:
- 开发者运行 `python tests/utils/cleanup_test_data.py`

**Then**:
- 脚本成功执行，清理测试数据
- 脚本不被 pytest 当作测试用例运行

### Requirement: 浅层目录结构（最多 2 层）

**ID**: `TEST-ORG-004`

**Description**:
测试目录嵌套 **MUST NOT** 超过 2 层（`automated/{domain}/test_*.py`），保持结构简单易导航。测试文件路径 **SHALL** 遵循 `tests/automated/{domain}/test_*.py` 格式，除非某个领域包含超过 5 个测试文件需要进一步分类。

#### Scenario: 限制目录嵌套深度

**Given**:
- 项目规模为中小型（10-30 个测试文件）

**When**:
- 组织测试目录结构

**Then**:
- 测试文件路径为 `tests/automated/{domain}/test_*.py` 格式
- 不存在三层或更深的嵌套（如 `automated/domain/subdomain/test_*.py`）
- 导入路径简洁：`from tests.automated.project import ...`

**Example**:
```
✅ 符合要求:
tests/automated/project/test_project_crud.py
tests/automated/model/test_model_design.py

❌ 不符合要求（过度嵌套）:
tests/automated/project/crud/create/test_create_project.py
tests/automated/model/design/fields/test_field_validation.py
```

#### Scenario: 领域测试增长时的扩展路径

**Given**:
- 模型领域测试增长到 6+ 个文件
- 需要进一步组织

**When**:
- 在领域目录内添加子分类

**Then**:
- 可在 `automated/model/` 下创建 `design/` 和 `runtime/` 子目录
- 仍保持最多 2 层嵌套
- 文档说明何时应该添加子分类（>5 个文件）

**Example**:
```
# 未来扩展（当模型测试超过 5 个文件时）
tests/automated/model/
├── design/
│   ├── test_model_crud.py
│   └── test_field_operations.py
└── runtime/
    ├── test_graphql_queries.py
    └── test_aggregations.py
```

### Requirement: Pytest 约定优于配置

**ID**: `TEST-ORG-005`

**Description**:
项目 **MUST** 遵循 pytest 标准约定进行文件命名、fixture 定义和测试发现。测试文件 **SHALL** 使用 `test_*.py` 命名，测试函数 **SHALL** 使用 `test_*` 前缀，共享 fixtures **SHALL** 定义在 `conftest.py` 中以利用 pytest 自动发现机制。

#### Scenario: 自动发现测试文件

**Given**:
- 测试文件遵循 `test_*.py` 命名约定
- 测试函数遵循 `test_*` 命名约定

**When**:
- 运行 `pytest automated/`

**Then**:
- Pytest 自动发现所有符合约定的测试
- 无需在 `pytest.ini` 中显式列出测试文件
- 测试函数自动被识别和执行

**Example**:
```python
# tests/automated/project/test_project_crud.py
def test_create_project_success():  # ✅ 自动发现
    pass

def validate_project():             # ❌ 不会被发现（不符合命名约定）
    pass
```

#### Scenario: Conftest.py 自动加载 fixtures

**Given**:
- `conftest.py` 定义了共享 fixtures
- 测试文件位于同一目录或子目录

**When**:
- 测试函数声明需要某个 fixture

**Then**:
- Pytest 自动从 `conftest.py` 加载 fixture
- 无需显式导入 fixture
- Fixture 作用域（session/module/function）自动生效

**Example**:
```python
# tests/automated/conftest.py
@pytest.fixture(scope="session")
def graphql_client():
    return create_client()

# tests/automated/project/test_project_crud.py
def test_create_project(graphql_client):  # 自动注入 fixture
    result = graphql_client.execute(...)
```

### Requirement: 迁移保持测试通过

**ID**: `TEST-ORG-006`

**Description**:
从旧结构迁移到新结构的过程中，**MUST** 确保所有现有测试持续通过，不引入测试失败。每个迁移阶段完成后 **SHALL** 运行测试验证。文件移动 **MUST** 使用 `git mv` 命令以保留版本历史。

#### Scenario: 分阶段迁移验证

**Given**:
- 迁移分为 5 个阶段
- 每个阶段迁移不同领域的测试

**When**:
- 完成一个领域的迁移（如项目领域）

**Then**:
- 运行 `pytest automated/project/ -v` 验证该领域测试全部通过
- 测试计数与迁移前一致
- 无新增失败或错误

#### Scenario: 导入路径更新验证

**Given**:
- 测试文件从 `tests/automated/` 移动到 `tests/automated/project/`
- 配置模块从 `tests/config.py` 移动到 `tests/utils/config.py`

**When**:
- 更新测试文件中的导入语句

**Then**:
- 旧导入 `from config import config` 更新为 `from tests.utils.config import config`
- 运行测试无 ImportError
- 配置正确加载

**Example**:
```python
# 迁移前
from config import config

# 迁移后
from tests.utils.config import config
```

#### Scenario: Git 历史保留

**Given**:
- 使用 Git 进行版本控制

**When**:
- 移动测试文件到新位置

**Then**:
- 使用 `git mv` 命令移动文件
- 文件历史保留（`git log --follow` 能追溯）
- 代码审查工具能识别文件移动而非删除+创建

**Example**:
```bash
# ✅ 正确方式（保留历史）
git mv tests/automated/test_project_crud.py tests/automated/project/test_project_crud.py

# ❌ 错误方式（丢失历史）
mv tests/automated/test_project_crud.py tests/automated/project/test_project_crud.py
git add tests/automated/project/test_project_crud.py
```

## MODIFIED Requirements

None. This is a new organizational structure, not modifying existing requirements.

## REMOVED Requirements

None. Existing test functionality remains unchanged.

## Related Capabilities

- **Test Execution**: CI/CD 流程需要适配新的测试目录结构
- **Documentation**: `tests/README.md` 和 `tests/automated/README.md` 需要更新
- **Developer Onboarding**: 新开发者指南需要说明测试组织原则

## Implementation Notes

### File Movement Mapping

```
旧路径 → 新路径

tests/automated/test_project_crud.py
  → tests/automated/project/test_project_crud.py

tests/automated/test_project_isolation.py
  → tests/automated/project/test_project_isolation.py

tests/automated/test_model_jsonschema_export.py
  → tests/automated/model/test_jsonschema_export.py

tests/automated/test_schema_based_operations.py
  → tests/automated/model/test_schema_operations.py

tests/automated/test_backward_compatibility.py
  → tests/automated/compatibility/test_backward_compatibility.py

tests/automated/test_modelcraft_client_test_graphql.py
  → tests/automated/integration/test_modelcraft_client.py

tests/design/test_model_graphql.py
  → tests/automated/model/test_model_design.py

tests/design/test_cluster_graphql.py
  → tests/automated/cluster/test_cluster_graphql.py

tests/runtime/test_user_graphql.py
  → tests/automated/model/test_model_runtime.py

tests/config.py
  → tests/utils/config.py

tests/health_check.py
  → tests/utils/health_check.py

tests/cleanup_test_data.py
  → tests/utils/cleanup_test_data.py
```

### Pytest Configuration Updates

```ini
# tests/automated/pytest.ini
[tool:pytest]
testpaths = .
python_files = test_*.py
python_functions = test_*
addopts =
    --html=../reports/test_report.html
    --self-contained-html
    -v
markers =
    integration: Integration tests requiring running server
    slow: Tests that take more than 5 seconds
```

### Import Path Pattern

所有测试文件统一使用以下导入模式：

```python
# 配置导入
from tests.utils.config import config

# 共享工具导入
from tests.automated.common.graphql_client import create_graphql_client

# Pytest fixtures（通过 conftest.py 自动注入，无需显式导入）
def test_example(graphql_client):  # graphql_client 来自 conftest.py
    pass
```

## Validation

### 迁移完成验证清单

- [ ] 所有测试文件已移动到目标位置
- [ ] 每个领域目录包含 `__init__.py`
- [ ] `tests/utils/` 目录创建并包含工具脚本
- [ ] 所有导入路径已更新
- [ ] `pytest automated/ -v` 运行成功，所有测试通过
- [ ] `pytest --collect-only` 显示正确的测试计数
- [ ] CI/CD 流程（`task auto-test`）运行正常
- [ ] 旧目录 `design/` 和 `runtime/` 已删除（在确认迁移成功后）
- [ ] 文档已更新以反映新结构

### 性能验证

- 测试发现时间：应与迁移前相当（< 1 秒）
- 测试执行时间：无显著变化（±5% 范围内）
- CI/CD 构建时间：无增加

## Open Questions

1. **是否需要为每个领域创建独立的 `conftest.py`？**
   - 建议：如果领域有专属 fixtures，创建；否则使用根 `conftest.py`

2. **未来如何处理跨领域的集成测试？**
   - 建议：放在 `automated/integration/` 目录

3. **性能测试和 E2E 测试应该放在哪里？**
   - 建议：创建 `tests/performance/` 和 `tests/e2e/` 作为 `automated/` 的同级目录
