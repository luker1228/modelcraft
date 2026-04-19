// RLS JWT Authentication Step Definitions
// 对应 Feature: Runtime JWT 认证

import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'
import { EndUserRestClient } from '../../support/end-user-rest-client'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

// Runtime 调用结果存储
const runtimeCallResults = new Map<string, {
  status: number
  data: unknown
  error: { message: string; code?: string } | null
}>()

// 辅助函数：调用 Runtime endpoint
async function callRuntime(
  orgName: string,
  projectSlug: string,
  modelName: string,
  token?: string
): Promise<{ status: number; data: unknown; error: { message: string; code?: string } | null }> {
  const query = `
    query FindMany${modelName} {
      findMany${modelName} {
        id
        name
      }
    }
  `

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(
    `${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/runtime`,
    {
      method: 'POST',
      headers,
      body: JSON.stringify({ query }),
    }
  )

  const body = await res.json()

  return {
    status: res.status,
    data: body.data,
    error: body.errors?.[0] || null,
  }
}

// Given 步骤

Given('我已获取开发者 JWT', function (this: ModelCraftWorld) {
  // 开发者 JWT 应该已经在 World 中设置
  if (!this.token) {
    throw new Error('开发者 JWT 未设置，请确保环境变量 TEST_ACCESS_TOKEN 已配置')
  }
})

// When 步骤

When('终端用户调用 Runtime 查询 {string}', async function (
  this: ModelCraftWorld, modelName: string
) {
  const token = this.currentEndUserToken
  if (!token) {
    throw new Error('终端用户未登录，请先执行登录步骤')
  }

  const orgName = this.endUserOrgName || this.orgName
  const projectSlug = this.endUserProjectSlug || this.projectSlug

  const result = await callRuntime(orgName, projectSlug, modelName, token)
  runtimeCallResults.set('lastCall', result)
  this.lastResponse = { runtimeCall: result }
})

When('开发者使用 JWT 调用 Runtime 查询 {string}', async function (
  this: ModelCraftWorld, modelName: string
) {
  const token = this.token
  if (!token) {
    throw new Error('开发者 JWT 未设置')
  }

  const orgName = this.endUserOrgName || this.orgName
  const projectSlug = this.endUserProjectSlug || this.projectSlug

  const result = await callRuntime(orgName, projectSlug, modelName, token)
  runtimeCallResults.set('lastCall', result)
  this.lastResponse = { runtimeCall: result }
})

When('使用无效的 JWT 调用 Runtime 查询 {string}', async function (
  this: ModelCraftWorld, modelName: string
) {
  const orgName = this.endUserOrgName || this.orgName
  const projectSlug = this.endUserProjectSlug || this.projectSlug

  const result = await callRuntime(
    orgName,
    projectSlug,
    modelName,
    'invalid-token-format'
  )
  runtimeCallResults.set('lastCall', result)
  this.lastResponse = { runtimeCall: result }
})

When('不带 JWT 调用 Runtime 查询 {string}', async function (
  this: ModelCraftWorld, modelName: string
) {
  const orgName = this.endUserOrgName || this.orgName
  const projectSlug = this.endUserProjectSlug || this.projectSlug

  const result = await callRuntime(orgName, projectSlug, modelName, undefined)
  runtimeCallResults.set('lastCall', result)
  this.lastResponse = { runtimeCall: result }
})

// Then 步骤

Then('Runtime 应该返回成功响应', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeCall: { status: number; error: unknown }
  }).runtimeCall

  expect(result.status).toBe(200)
  expect(result.error).toBeNull()
})

Then('Runtime 应该返回 401 未授权错误', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeCall: { status: number; error: { message: string } | null }
  }).runtimeCall

  // Runtime 应该返回 401 或 GraphQL error 包含 UNAUTHORIZED
  expect(result.status === 401 || result.error !== null).toBe(true)
  if (result.error) {
    const errorMsg = result.error.message || ''
    expect(
      errorMsg.includes('Unauthorized') ||
      errorMsg.includes('UNAUTHORIZED') ||
      errorMsg.includes('401')
    ).toBe(true)
  }
})
