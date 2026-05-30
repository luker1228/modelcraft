---
name: frontend-debug
description: 排查前端报错（HTTP 5xx、请求失败、BFF 异常）。当用户描述前端出现报错、请求失败、500、页面功能不正常时触发。两步走：先用 curl 确认是否是后端问题，再加 console 日志缩小范围。
---

# Frontend Debug — 两步排查法

## 核心思路

前端报错通常来自两个地方：**后端接口** 或 **前端自身逻辑（BFF / client 代码）**。先排除后端，再看前端。

---

## Step 1：用 curl 确认是否是后端问题

直接用 curl 请求出问题的接口，**绕过浏览器和 BFF**，看后端裸响应。

```bash
# 基本用法
curl -sv -X POST http://<host>/<path> \
  -H "Content-Type: application/json" \
  -d '{"key":"value"}'

# 带 cookie（从浏览器 DevTools → Application → Cookies 里复制）
curl -sv -X POST http://<host>/<path> \
  -H "Cookie: mc_refresh_token=<token>"

# 只看状态码
curl -s -o /dev/null -w "%{http_code}" -X POST http://<host>/<path>
```

**判断标准：**
- 后端返回 `4xx/5xx` → 后端问题，去查后端日志
- 后端返回 `2xx` → 前端/BFF 问题，进入 Step 2

> 本项目后端地址见 `.env.development` 里的 `BACKEND_URL`。BFF 地址是 `http://localhost:3001/api/bff/...`。

---

## Step 2：加 console 日志，观察错误

### 2a. BFF Route Handler（Next.js）

在出问题的 route handler 加 try/catch，把错误打到响应里，这样 curl 就能直接看到错误信息：

```typescript
export async function POST(req: NextRequest) {
  try {
    return await someHandler(req)
  } catch (err) {
    console.error('[BFF] error:', err)
    // 开发期临时：把错误信息打到响应里，方便 curl 直接看
    return NextResponse.json({ error: String(err) }, { status: 500 })
  }
}
```

然后用 curl 触发请求，响应 body 里直接显示报错原因：

```bash
curl -s -X POST http://localhost:3001/api/bff/org/<orgName>/...
# 输出：{"error":"TypeError: Response constructor: Invalid response status code 204"}
```

### 2b. 客户端代码

在可疑位置加 `console.log` / `console.error`，然后用 Playwright 或浏览器重现操作，读取 console 日志。

---

## 常见根因速查（本项目）

| 现象 | 根因 | 修复 |
|------|------|------|
| BFF 500，response body 为空，响应头含 `vary: RSC` | route handler 有编译错误，或被 Next.js 当成 Page 路由 | 检查 import 路径，确保 route.ts 只用 Node.js 兼容模块 |
| `TypeError: Response constructor: Invalid response status code 204` | `new NextResponse(body, { status: 204 })` — 204 不能带 body | 判断状态码：`const hasBody = status !== 204 && status !== 304` |
| BFF 500，但后端 curl 正常 | 从 `middleware.ts` 导入常量（Edge Runtime 文件不能被 Node.js route import） | 把常量移到 `shared/constants/` 普通文件里 |
| curl 后端也 500 | 后端问题，查 Go 服务日志 | 见 `backend-debug` skill |

---

## 调试完成后清理

临时加的 try/catch 和 console.error **必须在修复后删除**，保持代码整洁。
