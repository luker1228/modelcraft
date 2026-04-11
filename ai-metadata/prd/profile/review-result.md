# Profile 分表后端验收测试报告（**二次验收**）

## 执行范围
- BDD：`tests-bdd/features/profile/manage-profile.feature`
- Go：`internal/domain/profile`、`internal/app/profile` 相关测试
- Runtime Integration（可行性验证）：`modelcraft-backend/tests/runtime/integration/test_modelcraft_client.py`

---

## 执行命令
1. `npm --prefix "./tests-bdd" run test -- features/profile/manage-profile.feature`
2. `npm --prefix "./tests-bdd" run test -- --name "phone\+userName\+password 注册成功后自动创建 profile 并应用默认昵称"`
3. `npm --prefix "./tests-bdd" run test -- --name "updateMyProfile 按 PATCH 语义更新"`
4. `npm --prefix "./tests-bdd" run test -- --name "myUserProfile 在 profile 缺失时返回 ProfileNotFound"`
5. `npm --prefix "./tests-bdd" run test -- --name "me 查询兼容性（行为不变）"`
6. `go -C "./modelcraft-backend" test -v ./internal/domain/profile ./internal/app/profile`
7. `python3 -m pytest -v "./modelcraft-backend/tests/runtime/integration/test_modelcraft_client.py" -c "./modelcraft-backend/tests/pytest.ini" --html="./modelcraft-backend/tests/reports/profile-integration-report.html"`

---

## 通过/失败统计

### 1) BDD（Profile）
- 场景总数：4
- ✅ 通过：0
- ❌ 失败：4

### 2) Go（profile 相关包）
- 包总数：2（`internal/domain/profile`、`internal/app/profile`）
- ✅ 通过：2
- ❌ 失败：0

### 3) Runtime Integration（目标用例）
- 用例总数：2
- ✅ 通过：0
- ❌ 错误：2（setup 阶段）

---

## 失败清单（含 requestId）

🔍 [requestId: BR-20260411-0001]  
📁 文件: `tests-bdd/features/profile/manage-profile.feature`  
📍 位置: Scenario `phone+userName+password 注册成功后自动创建 profile 并应用默认昵称`  
🏷️ 严重程度: HIGH  
🏷️ 分类: logic

**问题**: 预期注册成功（201），实际返回 409，场景在注册断言处失败。  
**说明**: 固定用户名 `profile_user_a` 已存在，阻断后续 profile 快照断言。  
**后端 requestId**: 未在该断言日志中暴露（仅看到 HTTP 409 结果）。

---

🔍 [requestId: BR-20260411-0002]  
📁 文件: `tests-bdd/features/profile/manage-profile.feature`  
📍 位置: Scenario `updateMyProfile 按 PATCH 语义更新`（Given 阶段）  
🏷️ 严重程度: HIGH  
🏷️ 分类: logic

**问题**: 前置注册失败，返回 `CONFLICT.USER`，导致未进入 `updateMyProfile` 调用。  
**说明**: 固定用户名 `profile_user_b` 已存在。  
**后端 requestId**: `459a2994-ffeb-455f-b39b-3fe6172d800c`

---

🔍 [requestId: BR-20260411-0003]  
📁 文件: `tests-bdd/features/profile/manage-profile.feature`  
📍 位置: Scenario `myUserProfile 在 profile 缺失时返回 ProfileNotFound`（Given 阶段）  
🏷️ 严重程度: HIGH  
🏷️ 分类: error-handling

**问题**: 构造“仅 user 无 profile”测试前置失败（`Failed to save user`）。  
**说明**: 未进入 `myUserProfile` 查询与错误类型断言。  
**后端 requestId**: `2828445b-1e6e-4c09-ac50-51a58fcb5bff`

---

🔍 [requestId: BR-20260411-0004]  
📁 文件: `tests-bdd/features/profile/manage-profile.feature`  
📍 位置: Scenario `me 查询兼容性（行为不变）`（Given 阶段）  
🏷️ 严重程度: MEDIUM  
🏷️ 分类: logic

**问题**: 前置注册失败，返回 `CONFLICT.USER`。  
**说明**: 固定用户名 `profile_user_c` 已存在，未进入 `me` 查询断言。  
**后端 requestId**: `b42f8b5a-dda8-4626-ac29-de2ab047fe6c`

---

🔍 [requestId: BR-20260411-0005]  
📁 文件: `modelcraft-backend/tests/runtime/integration/test_modelcraft_client.py`  
📍 位置: `test_complete_model_lifecycle` / `test_project_cluster_model_integration` setup  
🏷️ 严重程度: HIGH  
🏷️ 分类: test

**问题**: runtime integration 缺失 `graphql_client` fixture，2 个用例在 setup 阶段直接报错。  
**说明**: 当前可执行到 pytest 收集，但被测试环境/fixture 装配阻塞。  
**错误摘要**: `fixture 'graphql_client' not found`

---

## 与上次相比是否有改善（重点：Unknown type/field）

- **上次**：出现 `Unknown type ...` / `Cannot query field ...`（`updateMyProfile`、`myUserProfile`）类 GraphQL 合同错误。  
- **本次**：在本轮执行日志中，**未再出现 Unknown type/field 报错（0 次）**。

### 结论判断
- **现象层面有改善**：Unknown type/field 未复现。  
- **但证据仍不充分**：本次 4 个 profile 场景均在 Given/注册阶段被数据冲突或建数失败拦截，未完整走到关键 GraphQL 断言路径，因此暂不能判定该问题已被彻底修复。

---

## 环境阻塞记录（runtime integration）
- 阻塞点：`graphql_client` fixture 缺失（测试装配问题）
- 影响：目标 runtime integration 用例无法进入业务断言阶段
- 报告：`modelcraft-backend/tests/reports/profile-integration-report.html`

---

## 附：BDD 报告位置
- `tests-bdd/reports/test-report.html`

> 说明：本次严格执行“只测试，不修复源码”，未修改任何后端生产代码。
