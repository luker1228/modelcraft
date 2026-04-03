# ModelCraft 测试套件

ModelCraft 项目的集成测试套件，采用 Design-Runtime 分层架构组织测试。

## 🚀 快速开始

```bash
# 1. 部署环境
task deploy-local       # 本地开发环境
# Or: task deploy-docker  # Docker 环境

# 2. 运行所有测试
pytest -v

# 3. 运行特定阶段测试
pytest design/ -v       # Design-time tests only
pytest runtime/ -v      # Runtime tests only

# 4. 停止环境（可选）
task deploy-stop
```

## 📁 目录结构（新架构）

```
tests/
├── design/                    # Design-Time 测试
│   ├── common/               # 共享工具（可被 runtime 复用）
│   │   ├── graphql_client.py # GraphQL 客户端
│   │   ├── test_data.py      # 测试数据生成器
│   │   ├── assertions.py     # 自定义断言
│   │   └── fixtures.py       # 共享 fixtures
│   │
│   ├── project/              # 项目管理测试
│   │   ├── test_project_crud.py
│   │   └── test_project_isolation.py
│   │
│   ├── cluster/              # 数据库集群测试
│   │   └── test_cluster_graphql.py
│   │
│   ├── model/                # 模型设计测试
│   │   ├── test_model_graphql.py
│   │   ├── test_jsonschema_export.py
│   │   └── test_schema_operations.py
│   │
│   ├── compatibility/        # 向后兼容性测试
│   │   └── test_backward_compatibility.py
│   │
│   ├── conftest.py           # Design fixtures
│   └── README.md
│
├── runtime/                   # Runtime 测试（依赖 Design）
│   ├── query/                # 数据查询测试
│   │   └── test_user_graphql.py
│   │
│   ├── integration/          # 端到端集成测试
│   │   └── test_modelcraft_client.py
│   │
│   ├── conftest.py           # Runtime fixtures
│   └── README.md
│
├── common/                    # 跨阶段共享工具
│   ├── config.py             # 配置管理
│   ├── health_check.py       # 健康检查
│   └── cleanup_test_data.py  # 数据清理
│
├── manual/                    # 手动测试资源
│   ├── curl_commands.md
│   └── test_scenarios.sh
│
├── conftest.py               # 根级 pytest 配置
├── pytest.ini                # Pytest 设置
├── requirements.txt          # Python 依赖
└── README.md                 # 本文档
```

## 🎯 测试架构

### Design-Time Tests (`design/`)

测试模型设计、集群配置、项目管理等"设计时"操作：
- 项目 CRUD
- 数据库集群管理
- 模型和字段设计
- JSON Schema 导入/导出
- 向后兼容性

**运行**:
```bash
pytest design/ -v
pytest design/project/ -v       # 只运行项目管理测试
pytest design/model/ -v         # 只运行模型设计测试
```

### Runtime Tests (`runtime/`)

测试运行时数据查询、GraphQL 执行等，**依赖 Design 阶段的初始化**：
- 数据查询和过滤
- GraphQL 查询执行
- 端到端集成流程

**运行**:
```bash
pytest runtime/ -v
pytest runtime/query/ -v        # 只运行查询测试
pytest runtime/integration/ -v  # 只运行集成测试
```

### 依赖关系

```
Runtime Tests
     ↓
  (can use)
     ↓
Design Common Utilities
     ↓
  (shared)
     ↓
Common Utilities
```

Runtime 测试可以导入 Design 工具类：
```python
from tests.design.common.graphql_client import create_design_graphql_client
from tests.design.common.test_data import build_model_input
```

## 🔧 配置

测试配置通过环境变量和 `.env` 文件加载：

```bash
# 关键配置变量
MODELCRAFT_BASE_URL=http://localhost:8080
DB_HOST=localhost
DB_PORT=3306
DB_PASSWORD=your_password
CLEANUP_ENABLED=true
```

配置文件位置：
- `tests/common/config.py` - 配置加载模块
- `.env` - 本地开发配置
- `.env.autotest` - 自动化测试配置
- `.env.docker` - Docker 环境配置

## 👤 Test User Setup

The test suite automatically manages a test user for integration testing. This eliminates the need for manual SQL script execution and ensures tests run consistently across all environments.

### Automatic Setup

The `test_user_with_owner_role` fixture in `tests/conftest.py` automatically:
- Creates a test user (test-integration) with owner role
- Assigns the user to the modelcraft organization
- Cleans up after the test session completes

This happens automatically when running integration tests - no manual setup required.

### Manual Control

You can manually manage the test user using Task commands:

```bash
# Create test user
task test-user-setup

# Clean up test user
task test-user-cleanup

# Clean up specific user
task test-user-cleanup USER_ID=487101d6-92bb-459e-b4f1-426255126d27
```

### Environment Variables

Control automatic test user management with environment variables:

```bash
# Skip automatic test user setup (use existing user)
SKIP_TEST_USER_SETUP=true pytest

# Keep test user after tests (for debugging)
KEEP_TEST_USER=true pytest
```

### Test User Details

The automated test user has the following fixed configuration:
- **User ID**: `487101d6-92bb-459e-b4f1-426255126d27`
- **External ID**: `test-integration`
- **Name**: Test Integration User
- **Organization**: modelcraft
- **Role**: Owner

### Troubleshooting

If test user setup fails:

1. **Database not running**
   - Deploy: `task deploy-local` or `task deploy-docker`
   - Verify: `curl http://localhost:8080/health`

2. **Database migrations not applied**
   - Run: `task db:migrate-up`

3. **Missing organization or role**
   - Ensure 'modelcraft' organization and 'Owner' role exist
   - Run migrations: `task db:migrate-up`

4. **Connection errors**
   - Check database credentials in `.env` file
   - Verify DB_HOST, DB_PORT, DB_USER, DB_PASSWORD

### Direct Utility Script

You can also use the Python utility directly:

```bash
# Setup
cd tests && python common/test_user_setup.py

# Cleanup
cd tests && python common/test_user_setup.py --cleanup

# With custom user ID
cd tests && python common/test_user_setup.py --cleanup --user-id <uuid>
```

## 🧪 常用测试命令


## 📝 编写新测试

### 添加 Design 测试

1. 选择合适的领域目录：`design/project/`, `design/cluster/`, `design/model/`
2. 创建 `test_*.py` 文件
3. 使用共享工具：
   ```python
   from tests.design.common.test_data import build_project_input
   from tests.design.common.assertions import assert_graphql_success

   def test_create_project(graphql_client, created_projects):
       input_data = build_project_input(title="Test")
       result = graphql_client.execute(CREATE_PROJECT, ...)
       assert_graphql_success(result, "createProject")
       created_projects.append(result["createProject"]["projectId"])
   ```

### 添加 Runtime 测试

1. 选择合适的分类：`runtime/query/` 或 `runtime/integration/`
2. 创建 `test_*.py` 文件
3. 可以使用 Design 工具进行设置：
   ```python
   from tests.design.common.test_data import build_model_input

   def test_query_data(runtime_graphql_client, sample_model_setup):
       model_id = sample_model_setup["id"]  # 由 Design fixture 创建
       # Runtime 查询逻辑...
   ```

## 🧪 常用测试命令

```bash
# 运行所有测试
pytest

# 运行特定阶段
pytest design/
pytest runtime/

# 运行特定标记
pytest -m design        # 只运行 Design 测试
pytest -m runtime       # 只运行 Runtime 测试

# 详细输出
pytest -v -s

# 生成 HTML 报告
pytest --html=reports/test_report.html

# 并发运行（需要 pytest-xdist）
pytest -n auto

# 只运行失败的测试
pytest --lf

# 在第一个失败时停止
pytest -x
```

## 🧹 数据清理

测试使用 fixtures 自动清理创建的资源：
- `created_projects` - 追踪并清理项目
- `created_clusters` - 追踪并清理集群
- `created_models` - 追踪并清理模型

手动清理：
```bash
cd tests
python common/cleanup_test_data.py
```

## 📊 CI/CD 集成

```bash
# 使用 task 命令
task auto-test                 # 自动化测试（完整流程）
task test-design              # 只运行 Design 测试
task test-runtime             # 只运行 Runtime 测试

# 使用 make 命令（向后兼容）
make auto-test
make deploy-local && make auto-test
```

## 📖 更多文档

- **[Design Tests](design/README.md)** - Design 测试详细说明
- **[Runtime Tests](runtime/README.md)** - Runtime 测试详细说明
- **[Manual Testing](manual/README.md)** - 手动测试指南

## 🐛 故障排除

### 导入错误

如果遇到导入错误，确保 `tests/` 目录在 Python 路径中：
```python
import sys
sys.path.insert(0, '/path/to/tests')
```

### 连接失败

检查服务是否运行：
```bash
curl http://localhost:8080/health
```

### 测试数据冲突

清理测试数据库：
```bash
python common/cleanup_test_data.py
```

## ⚡ 性能提示

- 使用 `pytest-xdist` 并发运行：`pytest -n auto`
- 使用 `pytest-repeat` 重复测试：`pytest --count=10`
- 使用 `pytest-benchmark` 性能测试

## 📞 获取帮助

- 查看测试日志：`pytest -v -s`
- 查看完整错误：`pytest --tb=long`
- 检查 fixture 状态：`pytest --setup-show`
