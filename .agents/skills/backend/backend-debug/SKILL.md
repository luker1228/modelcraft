---
name: backend-debug
description: >
  排查和修复 ModelCraft 后端错误。当用户提供 GraphQL 响应中的错误（包含 errors 数组、requestId、message 字段），
  或者描述后端报错、接口异常、服务崩溃时，使用此 skill。
  触发场景包括：
  (1) 用户粘贴了带 errors/requestId 的 JSON 响应，
  (2) 说"后端报错了"、"接口返回错误"、"帮我看看这个错误"、"定位问题"、
  (3) 说"使用 just log"、"查看日志"、"找到错误原因"后想修复，
  (4) 任何需要通过日志定位再修复代码的后端问题。
---

# 后端问题排查与修复

当用户遇到后端错误时，目标是：**先用日志精准定位根因，再最小化修复代码**。

## 第一步：从错误中提取关键信息

从用户提供的错误响应中找到：
- `requestId`（在 `extensions.requestId` 字段）
- `message`（错误描述，往往包含错误链）
- `path`（出错的 GraphQL 操作名）

示例错误：
```json
{
  "errors": [{"message": "failed to introspect table: ...", "path": ["importModel"]}],
  "extensions": {"requestId": "aa65e02c-f1fc-4de6-a7cb-5169d92e0cdf"}
}
```

## 第二步：用 requestId 查日志

使用 `just log-cat` 按 requestId 过滤完整的请求日志链：

```bash
just log-cat <requestId>
```

这会展示该请求从进入到结束的所有日志行，包括 panic、error、warn 级别的详细堆栈。

如果需要查看最近的日志流：
```bash
just logs
```

**读日志的重点**：
- 找最早出现的 `error` 或 `panic` 行——这通常是根本原因，而不是传播后的错误
- 注意错误链（`wrapped error: ... caused by: ...`），从最内层开始理解
- 关注 SQL 错误、nil pointer、类型断言失败等具体技术原因

## 第三步：定位源码

根据日志中的错误信息和堆栈，定位到具体文件和行号：
- 错误消息通常包含函数名和包路径（如 `failed to query table comment`）
- 用 Grep 搜索错误字符串找到对应代码
- 用 Read 读取相关文件理解上下文

**常见错误模式**：

| 错误特征 | 可能原因 | 定位方向 |
|---------|---------|---------|
| `sql: no rows in result set` | 查询返回空，代码未处理 | 找对应的 DB 查询，检查是否应用了 `errors.Is(err, sql.ErrNoRows)` |
| `unsupported Scan, storing []uint8` | sqlc 扫描类型不匹配 | 查目标结构体字段类型，改为扫描到原生类型 |
| `[REPO_NOT_FOUND]` | 仓库层未找到记录 | 找 repository 实现，检查查询条件 |
| `nil pointer dereference` | 未检查空指针 | 找返回指针的地方，补 nil 检查 |
| `context deadline exceeded` | 超时 | 检查是否有慢查询或死锁 |

## 第四步：制定修复方案

理解根因后，选择最小侵入的修复：
- **不要过度重构**——只修改导致当前错误的代码
- **不要添加不必要的错误处理**——只在真正需要的边界添加
- 先确认修复逻辑正确，再动手写代码

## 第五步：修复并验证

修改代码后，可用以下命令验证：
```bash
just build          # 确认编译通过
just test-unit      # 跑单元测试（如果相关包有测试）
just run --force    # 强制重启服务（自动杀掉占用端口的旧进程）后手动复现
```

## 注意事项

- `just log-cat` 依赖日志文件存在，如果服务未运行过或日志被清除，则无法查到
- 某些错误（如启动时的配置错误）不会有 requestId，直接看 `just logs` 即可
- 修复后建议用原来的请求再试一次，确认问题消失
- 重启服务时如果遇到端口占用，使用 `just run --force` 而非手动 `fuser -k`
