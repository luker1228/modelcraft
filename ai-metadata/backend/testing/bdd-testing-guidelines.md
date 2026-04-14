# BDD 验收测试注意要点

> 适用范围：`tests-bdd/` 下的 Cucumber.js 验收测试。

## 核心原则

1. **默认不耦合注册流程**
   - 对于 `model` / `field` / `enum` / `lfk` / `smoke` 这类业务能力测试，前置条件应直接使用可登录的稳定测试用户与固定项目。
   - 不应把“先注册一个新用户”作为默认启动路径。

2. **只有在明确测试注册能力时才走注册流程**
   - 例如 `features/auth/register.feature` 以及用户明确要求验证注册接口时。
   - 其它功能测试若依赖注册，容易引入 `CONFLICT.USER`、组织不匹配、角色缺失等与目标无关的噪音失败。
   - 既然注册流程已保证 init-org，`org/init` 不应再作为常规回归场景。

3. **认证优先级：TOKEN > 登录 > 注册**
   - 优先使用 `TEST_ACCESS_TOKEN`。
   - 未提供 token 时，使用已有测试账号登录后再签发测试 JWT。
   - 不把自动注册作为默认兜底。

## 推荐环境约定

`tests-bdd/.env.test` 建议至少包含：

```env
API_BASE_URL=http://localhost:8080
TEST_ORG_NAME=<固定组织>
TEST_PROJECT_SLUG=<固定项目>
```

可选但推荐：

```env
TEST_ACCESS_TOKEN=<稳定测试 token>
# 或者
TEST_LOGIN_PHONE=<稳定测试手机号>
TEST_LOGIN_PASSWORD=<测试密码>
```

## 常见反模式

- 在所有领域测试中复用“注册新用户”作为 Given 前置。
- 用注册副作用隐式创建组织/角色，再拿这些副作用去驱动 model/enum 等非 auth 场景。
- 因注册冲突导致场景失败（`CONFLICT.USER`）后误判为业务功能失败。

## 失败归因建议

- `Permission denied: requires 'xxx' permission`：优先检查测试用户是否在目标 `TEST_ORG_NAME` 下具备角色绑定。
- `CONFLICT.USER`：通常属于测试数据冲突或前置设计不当，优先去耦合注册前置。

## 落地建议

- 把注册相关 Given/When/Then 限定在 Auth-Register 场景中。
- 其它场景统一采用“稳定用户登录 + 固定 org/project”模式，降低测试波动并提升可复现性。