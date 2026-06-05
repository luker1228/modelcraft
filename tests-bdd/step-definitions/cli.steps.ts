// tests-bdd/step-definitions/cli.steps.ts
//
// CLI (mc) BDD 步骤定义
// 通过 execSync 调用 mc 二进制，解析 JSON 输出进行断言。
// mc 始终输出 JSON（ok: true/false），适合机器断言。

import { Given, When, Then, Before, After } from '@cucumber/cucumber'
import { execSync, ExecSyncOptions } from 'child_process'
import { expect } from 'expect'
import * as fs from 'fs'
import * as os from 'os'
import * as path from 'path'

// ─── 类型 ──────────────────────────────────────────────────────────

interface CLIResult {
  ok: boolean
  data?: Record<string, unknown>
  error?: {
    code: string
    message: string
    retryable?: boolean
    suggestion?: string
    details?: Record<string, unknown>
  }
  meta?: Record<string, unknown>
}

// ─── World 扩展（Cucumber 允许 this 扩展） ─────────────────────────

interface CLIWorld {
  cliResult: CLIResult | null
  cliExitCode: number
  cliCredentialsFile: string
  // 测试用的隔离 credentials 文件，避免影响系统全局 mc 登录状态
}

// ─── 辅助函数 ──────────────────────────────────────────────────────

const MC = process.env.MC_BIN ?? 'mc'
const PAT = process.env.CLI_PAT_TOKEN ?? ''
const SERVER = process.env.CLI_SERVER ?? 'http://lukemxjia.devcloud.woa.com:9080'
const ORG_NAME = process.env.CLI_ORG_NAME ?? ''
const PROJECT_SLUG = process.env.CLI_PROJECT_SLUG ?? ''
const DATABASE = process.env.CLI_DATABASE ?? ''
const MODEL = process.env.CLI_MODEL ?? ''

/**
 * Commands that do NOT accept --credentials flag.
 * These are stateless or global commands.
 */
const NO_CREDENTIALS_CMDS = new Set(['version', 'schema', 'completion', 'help'])

function commandSupportsCredentials(args: string): boolean {
  const firstToken = args.trim().split(/\s+/)[0] ?? ''
  return !NO_CREDENTIALS_CMDS.has(firstToken)
}

/**
 * 执行 mc 命令，返回解析后的 JSON 输出和退出码。
 * mc 始终向 stdout 写入 JSON（成功或失败），不抛异常。
 */
function runMC(args: string, opts: { credFile?: string } = {}): { result: CLIResult; exitCode: number } {
  // Only append --credentials for commands that support it
  const credFlag = (opts.credFile && commandSupportsCredentials(args))
    ? `--credentials "${opts.credFile}"`
    : ''
  const cmd = `${MC} ${args} ${credFlag}`.trim()

  let stdout = ''
  let exitCode = 0

  try {
    const execOpts: ExecSyncOptions = { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] }
    stdout = execSync(cmd, execOpts) as string
  } catch (e: unknown) {
    const err = e as { stdout?: Buffer | string; status?: number }
    stdout = err.stdout?.toString() ?? ''
    exitCode = err.status ?? 1
  }

  const parsed = JSON.parse(stdout.trim()) as CLIResult
  if (exitCode === 0 && !parsed.ok) {
    // mc exits non-zero on errors, but parse the JSON to extract exit code
    // Some mc errors still produce exit 0 on the shell layer
    exitCode = exitCode || 1
  }

  return { result: parsed, exitCode }
}

// ─── Before/After 钩子 ─────────────────────────────────────────────

Before({ tags: '@cli' }, function (this: CLIWorld) {
  // 每个 CLI 场景使用独立的临时 credentials 文件，场景间互不干扰
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'mc-bdd-'))
  this.cliCredentialsFile = path.join(tmpDir, 'credentials.json')
  this.cliResult = null
  this.cliExitCode = 0
})

After({ tags: '@cli' }, function (this: CLIWorld) {
  // 清理临时 credentials 文件
  if (this.cliCredentialsFile && fs.existsSync(path.dirname(this.cliCredentialsFile))) {
    fs.rmSync(path.dirname(this.cliCredentialsFile), { recursive: true, force: true })
  }
})

// ─── Given 步骤 ────────────────────────────────────────────────────

Given('用户已通过 PAT token 登录 CLI', function (this: CLIWorld) {
  const { result, exitCode } = runMC(
    `auth login --token "${PAT}" --server "${SERVER}"`,
    { credFile: this.cliCredentialsFile }
  )
  expect(result.ok).toBe(true)
  expect(exitCode).toBe(0)
  expect(result.data?.orgName).toBeTruthy()
})

Given('用户已切换到项目 {string}', function (this: CLIWorld, projectSlug: string) {
  const { result, exitCode } = runMC(
    `auth switch-project ${projectSlug}`,
    { credFile: this.cliCredentialsFile }
  )
  expect(result.ok).toBe(true)
  expect(exitCode).toBe(0)
})

Given('用户未登录 CLI', function (this: CLIWorld) {
  // Remove the credentials file if it exists so subsequent commands are unauthenticated
  if (this.cliCredentialsFile && fs.existsSync(this.cliCredentialsFile)) {
    fs.unlinkSync(this.cliCredentialsFile)
  }
})

// ─── When 步骤 ─────────────────────────────────────────────────────

When('用户执行 {string}', function (this: CLIWorld, cmd: string) {
  // 替换占位符变量
  const resolved = cmd
    .replace('{PAT}', PAT)
    .replace('{SERVER}', SERVER)
    .replace('{ORG}', ORG_NAME)
    .replace('{PROJECT}', PROJECT_SLUG)
    .replace('{DATABASE}', DATABASE)
    .replace('{MODEL}', MODEL)

  const { result, exitCode } = runMC(resolved, { credFile: this.cliCredentialsFile })
  this.cliResult = result
  this.cliExitCode = exitCode
})

When('用户执行 mc run {string} 查询 {string}', function (
  this: CLIWorld,
  modelPath: string,
  query: string
) {
  const resolved = modelPath
    .replace('{PROJECT}', PROJECT_SLUG)
    .replace('{DATABASE}', DATABASE)
    .replace('{MODEL}', MODEL)

  const { result, exitCode } = runMC(
    `run "${resolved}" '${query}'`,
    { credFile: this.cliCredentialsFile }
  )
  this.cliResult = result
  this.cliExitCode = exitCode
})

// ─── Then 步骤 ─────────────────────────────────────────────────────

/**
 * mc run returns data as { <operation>: { items: [...], totalCount: N } }
 * e.g. { findMany: { items: [...], totalCount: 15 } }
 * These helpers look inside the first operation key for items/totalCount.
 */
function getOperationResult(data: Record<string, unknown> | undefined): Record<string, unknown> | null {
  if (!data) return null
  // Direct items (catalog commands): { items: [...] }
  if (Array.isArray(data.items)) return data
  // Nested operation result (mc run): { findMany: { items: [...] } }
  for (const val of Object.values(data)) {
    if (val && typeof val === 'object' && !Array.isArray(val)) {
      const nested = val as Record<string, unknown>
      if (Array.isArray(nested.items) || typeof nested.totalCount === 'number') {
        return nested
      }
    }
  }
  return data
}

function getNestedItems(data: Record<string, unknown> | undefined): unknown {
  return getOperationResult(data)?.items
}

function getNestedTotalCount(data: Record<string, unknown> | undefined): number | null {
  const op = getOperationResult(data)
  if (!op) return null
  const tc = op.totalCount
  if (typeof tc === 'number') return tc
  return null
}

Then('CLI 命令应该成功', function (this: CLIWorld) {
  expect(this.cliResult).not.toBeNull()
  expect(this.cliResult!.ok).toBe(true)
})

Then('CLI 命令应该失败', function (this: CLIWorld) {
  expect(this.cliResult).not.toBeNull()
  expect(this.cliResult!.ok).toBe(false)
})

Then('CLI 退出码应该为 {int}', function (this: CLIWorld, expectedCode: number) {
  expect(this.cliExitCode).toBe(expectedCode)
})

Then('CLI 错误码应该为 {string}', function (this: CLIWorld, expectedCode: string) {
  expect(this.cliResult?.error?.code).toBe(expectedCode)
})

Then('CLI 响应应该包含字段 {string}', function (this: CLIWorld, fieldPath: string) {
  // Check in data first, then top-level result (for meta, types, etc.)
  const parts = fieldPath.split('.')
  // Try data first
  let found = false
  if (this.cliResult?.data) {
    let current: unknown = this.cliResult.data
    let ok = true
    for (const part of parts) {
      if (current == null || typeof current !== 'object' || !(part in (current as Record<string, unknown>))) {
        ok = false; break
      }
      current = (current as Record<string, unknown>)[part]
    }
    if (ok) found = true
  }
  // Try top-level result
  if (!found) {
    let current: unknown = this.cliResult as unknown
    let ok = true
    for (const part of parts) {
      if (current == null || typeof current !== 'object' || !(part in (current as Record<string, unknown>))) {
        ok = false; break
      }
      current = (current as Record<string, unknown>)[part]
    }
    if (ok) found = true
  }
  expect(found).toBe(true)
})

Then('CLI data.{string} 应该等于 {string}', function (
  this: CLIWorld, field: string, expected: string
) {
  const actual = this.cliResult?.data?.[field]
  expect(String(actual)).toBe(expected)
})

Then('CLI data.{string} 应该不为空', function (this: CLIWorld, field: string) {
  const actual = this.cliResult?.data?.[field]
  expect(actual).toBeTruthy()
})

Then('CLI data.items 数量应该大于 {int}', function (this: CLIWorld, minCount: number) {
  // Support both direct data.items and data.<operation>.items (e.g. data.findMany.items)
  const items = getNestedItems(this.cliResult?.data)
  expect(Array.isArray(items)).toBe(true)
  expect((items as unknown[]).length).toBeGreaterThan(minCount)
})

Then('CLI data.items 数量应该等于 {int}', function (this: CLIWorld, count: number) {
  const items = getNestedItems(this.cliResult?.data)
  expect(Array.isArray(items)).toBe(true)
  expect((items as unknown[]).length).toBe(count)
})

Then('CLI data.totalCount 应该大于 {int}', function (this: CLIWorld, minCount: number) {
  const tc = getNestedTotalCount(this.cliResult?.data)
  expect(tc).not.toBeNull()
  expect(tc!).toBeGreaterThan(minCount)
})

Then('CLI data.totalCount 应该等于 {int}', function (this: CLIWorld, count: number) {
  const tc = getNestedTotalCount(this.cliResult?.data)
  expect(tc).toBe(count)
})

Then('CLI meta.{string} 应该等于 {string}', function (
  this: CLIWorld, field: string, expected: string
) {
  const actual = this.cliResult?.meta?.[field]
  expect(String(actual)).toBe(expected)
})

Then('CLI 错误信息应该包含 {string}', function (this: CLIWorld, substring: string) {
  const msg = this.cliResult?.error?.message ?? ''
  expect(msg).toContain(substring)
})

Then('CLI data 中的 {string} 字段应该包含项目 {string}', function (
  this: CLIWorld, field: string, projectSlug: string
) {
  const items = this.cliResult?.data?.[field] as Array<{ slug: string }> | null
  expect(Array.isArray(items)).toBe(true)
  const found = items!.some((p) => p.slug === projectSlug)
  expect(found).toBe(true)
})

Then('CLI findMany 第一条记录应该包含字段 {string}', function (
  this: CLIWorld, fieldName: string
) {
  const fm = this.cliResult?.data?.findMany as Record<string, unknown> | null
  expect(fm).not.toBeNull()
  const items = fm!.items as Array<Record<string, unknown>>
  expect(items.length).toBeGreaterThan(0)
  expect(items[0]).toHaveProperty(fieldName)
})
