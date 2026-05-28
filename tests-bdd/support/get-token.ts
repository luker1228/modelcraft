#!/usr/bin/env npx tsx
/**
 * get-token.ts — BDD 测试 Token 获取工具
 *
 * 用途：通过账号密码登录获取 access_token，并自动写入 .env.test
 *
 * 用法：
 *   npx tsx support/get-token.ts --phone 13800138000 --password yourpassword
 *   npx tsx support/get-token.ts --username myuser --password yourpassword
 *   npx tsx support/get-token.ts --phone 13800138000 --password yourpassword --write
 *
 * 选项：
 *   --phone      手机号（与 --username 二选一）
 *   --username   用户名（与 --phone 二选一）
 *   --password   密码（必填）
 *   --write      自动将 TEST_ACCESS_TOKEN 写入 .env.test（默认：仅打印）
 *   --url        API base URL（默认读取 .env.test 的 API_BASE_URL，或 http://localhost:8080）
 */

import 'dotenv/config'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const ENV_FILE = path.resolve(__dirname, '../.env.test')

// ── 解析命令行参数 ────────────────────────────────────────────────────────────

function parseArgs(): {
  phone?: string
  username?: string
  password: string
  write: boolean
  url: string
} {
  const args = process.argv.slice(2)
  const get = (flag: string): string | undefined => {
    const idx = args.indexOf(flag)
    return idx !== -1 ? args[idx + 1] : undefined
  }

  const phone = get('--phone')
  const username = get('--username')
  const password = get('--password')
  const write = args.includes('--write')
  const urlArg = get('--url')

  // 读取 .env.test 中的 API_BASE_URL
  const envUrl = process.env.API_BASE_URL

  if (!phone && !username) {
    console.error('Error: --phone 或 --username 必须提供其中一个')
    process.exit(1)
  }
  if (!password) {
    console.error('Error: --password 是必填项')
    process.exit(1)
  }

  return {
    phone,
    username,
    password,
    write,
    url: urlArg ?? envUrl ?? 'http://localhost:8080',
  }
}

// ── 登录并获取 token ──────────────────────────────────────────────────────────

async function login(opts: {
  phone?: string
  username?: string
  password: string
  url: string
}): Promise<{ accessToken: string; userId: string; orgName: string }> {
  const identifier = opts.phone ?? opts.username!
  const identifierType = opts.phone ? 'PHONE' : 'USERNAME'

  const res = await fetch(`${opts.url}/api/tenant/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identifier, identifierType, password: opts.password }),
  })

  const body = (await res.json()) as Record<string, unknown>

  if (!res.ok) {
    // 兼容旧版 phone 字段
    if (opts.phone && res.status === 400) {
      const legacyRes = await fetch(`${opts.url}/api/tenant/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: opts.phone, password: opts.password }),
      })
      const legacyBody = (await legacyRes.json()) as Record<string, unknown>
      if (legacyRes.ok) {
        return {
          accessToken: legacyBody.accessToken as string,
          userId: legacyBody.userId as string,
          orgName: legacyBody.orgName as string,
        }
      }
    }
    const err = body as { error?: { code?: string; message?: string } }
    throw new Error(`Login failed [${res.status}]: ${err.error?.code} — ${err.error?.message}`)
  }

  return {
    accessToken: body.accessToken as string,
    userId: body.userId as string,
    orgName: body.orgName as string,
  }
}

// ── 写入 .env.test ────────────────────────────────────────────────────────────

function updateEnvFile(token: string): void {
  if (!fs.existsSync(ENV_FILE)) {
    console.error(`找不到 .env.test 文件：${ENV_FILE}`)
    process.exit(1)
  }

  let content = fs.readFileSync(ENV_FILE, 'utf-8')

  if (content.includes('TEST_ACCESS_TOKEN=')) {
    // 替换已有行
    content = content.replace(/^TEST_ACCESS_TOKEN=.*$/m, `TEST_ACCESS_TOKEN=${token}`)
  } else {
    // 追加新行
    content = content.trimEnd() + `\nTEST_ACCESS_TOKEN=${token}\n`
  }

  fs.writeFileSync(ENV_FILE, content, 'utf-8')
  console.log(`✅ .env.test 已更新：TEST_ACCESS_TOKEN`)
}

// ── 解码 JWT payload（仅用于展示过期时间）────────────────────────────────────

function decodeJwtExpiry(token: string): string {
  try {
    const payload = token.split('.')[1]
    const decoded = JSON.parse(Buffer.from(payload, 'base64url').toString('utf-8')) as {
      exp?: number
    }
    if (decoded.exp) {
      return new Date(decoded.exp * 1000).toLocaleString()
    }
  } catch {
    // ignore
  }
  return '(unknown)'
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main(): Promise<void> {
  const opts = parseArgs()

  console.log(`🔐 正在登录 ${opts.url} ...`)
  const identifier = opts.phone
    ? `手机号 ${opts.phone}`
    : `用户名 ${opts.username}`
  console.log(`   ${identifier}`)

  const result = await login(opts)

  console.log(`\n✅ 登录成功`)
  console.log(`   user_id  : ${result.userId}`)
  console.log(`   org_name : ${result.orgName}`)
  console.log(`   过期时间 : ${decodeJwtExpiry(result.accessToken)}`)
  console.log(`\nTEST_ACCESS_TOKEN=${result.accessToken}\n`)

  if (opts.write) {
    updateEnvFile(result.accessToken)
    console.log(`\n💡 现在可以运行：`)
    console.log(`   npm test -- features/database/`)
  } else {
    console.log(`💡 提示：加 --write 参数可自动写入 .env.test`)
    console.log(`   npx tsx support/get-token.ts ... --write`)
  }
}

main().catch((err: unknown) => {
  console.error('❌', err instanceof Error ? err.message : err)
  process.exit(1)
})
