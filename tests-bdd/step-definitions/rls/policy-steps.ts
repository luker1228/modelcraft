// RLS Policy Step Definitions
// 对应 Feature: Policy 配置

import { Given, When, Then, DataTable } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'

const GET_MODEL_POLICY = `
  query GetModelPolicy($id: ID!) {
    model(id: $id) {
      model {
        id
        name
        rlsPolicy {
          modelId
          selectPredicate
          insertCheck
          updatePredicate
          updateCheck
          deletePredicate
          preset
        }
      }
      error {
        __typename
      }
    }
  }
`

const SET_MODEL_RLS_POLICY = `
  mutation SetModelRLSPolicy($input: SetModelRLSPolicyInput!) {
    setModelRLSPolicy(input: $input) {
      policy {
        modelId
        selectPredicate
        insertCheck
        updatePredicate
        updateCheck
        deletePredicate
        preset
      }
      error {
        __typename
        ... on ModelHasNoOwnerField { message }
        ... on InvalidRLSExpression { message path }
        ... on InvalidAuthVariable { message variable }
      }
    }
  }
`

const VALIDATE_RLS_EXPR = `
  mutation ValidateRLSExpr($input: ValidateRLSExprInput!) {
    validateRLSExpr(input: $input) {
      result {
        valid
        errors {
          path
          message
          code
        }
      }
      error {
        __typename
      }
    }
  }
`

const SET_PROJECT_AUTH_SCHEMA = `
  mutation SetProjectAuthSchema($input: SetProjectAuthSchemaInput!) {
    setProjectAuthSchema(input: $input) {
      authSchema {
        variables {
          name
          source
          type
        }
      }
      error {
        __typename
      }
    }
  }
`

const GET_PROJECT_AUTH_SCHEMA = `
  query GetProject($orgName: String!, $slug: String!) {
    project(orgName: $orgName, slug: $slug) {
      project {
        authSchema {
          variables {
            name
            source
            type
          }
        }
      }
      error {
        __typename
      }
    }
  }
`

function getCurrentModelId(world: ModelCraftWorld): string {
  const id = world.createdModelIds[world.createdModelIds.length - 1]
  if (!id) throw new Error('没有可用的模型 ID')
  return id
}

// Given 步骤

Given('该模型不存在 RLS 策略', async function (this: ModelCraftWorld) {
  // 确保模型存在但无 owner 字段（即无 Policy）
  // 通过删除 owner 字段来实现
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    model: { model: { fields: Array<{ name: string }> } | null }
  }>(GET_MODEL_POLICY, { id: modelId })

  const hasOwner = res.model.model?.fields.some(f => f.name === 'owner')
  if (hasOwner) {
    // 删除 owner 字段
    await this.projectClient.mutate(`
      mutation RemoveField($modelID: ID!, $fieldName: String!) {
        removeField(modelID: $modelID, fieldName: $fieldName) {
          error { __typename }
        }
      }
    `, { modelID: modelId, fieldName: 'owner' })
  }
})

Given('已为项目设置 auth_schema，包含变量 {string}', async function (
  this: ModelCraftWorld, varName: string
) {
  await this.orgClient.mutate<{ setProjectAuthSchema: unknown }>(SET_PROJECT_AUTH_SCHEMA, {
    input: {
      projectSlug: this.projectSlug,
      variables: [
        { name: varName, source: `jwt.${varName}`, type: 'UUID' },
      ],
    },
  })
})

// When 步骤

When('我验证表达式 {string} 用于 {word}', async function (
  this: ModelCraftWorld, expression: string, exprType: string
) {
  const modelId = getCurrentModelId(this)

  const res = await this.projectClient.mutate<{ validateRLSExpr: unknown }>(
    VALIDATE_RLS_EXPR,
    {
      input: {
        modelId,
        exprType,
        expression,
      },
    }
  )
  this.lastResponse = { validateRLSExpr: res.validateRLSExpr }
})

When('我为项目设置 auth_schema，添加变量 {string}，source 为 {string}，type 为 {string}',
  async function (this: ModelCraftWorld, varName: string, source: string, type: string) {
    const res = await this.orgClient.mutate<{ setProjectAuthSchema: unknown }>(
      SET_PROJECT_AUTH_SCHEMA,
      {
        input: {
          projectSlug: this.projectSlug,
          variables: [
            { name: varName, source, type },
          ],
        },
      }
    )
    this.lastResponse = { setProjectAuthSchema: res.setProjectAuthSchema }
  }
)

When('我设置该模型的 RLS 策略为以下五件套:', async function (this: ModelCraftWorld, table: DataTable) {
  const modelId = getCurrentModelId(this)
  const rows = table.hashes()

  const input: Record<string, string> = { modelId }
  for (const row of rows) {
    input[row.predicateType] = row.expression
  }

  const res = await this.projectClient.mutate<{ setModelRLSPolicy: unknown }>(
    SET_MODEL_RLS_POLICY,
    { input }
  )
  this.lastResponse = { setModelRLSPolicy: res.setModelRLSPolicy }
})

// Then 步骤

Then('Policy 设置成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    setModelRLSPolicy: { error: unknown; policy: unknown }
  }).setModelRLSPolicy
  expect(payload.error).toBeNull()
  expect(payload.policy).not.toBeNull()
})

Then('Policy 的 {word} 应该为 {string}', function (
  this: ModelCraftWorld, predicateType: string, expectedValue: string
) {
  const payload = (this.lastResponse as {
    setModelRLSPolicy: { policy: Record<string, string> | null }
  }).setModelRLSPolicy

  expect(payload.policy).not.toBeNull()
  expect(payload.policy?.[predicateType]).toBe(expectedValue)
})

Then('验证应该通过', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    validateRLSExpr: { result: { valid: boolean }; error: unknown }
  }).validateRLSExpr
  expect(payload.error).toBeNull()
  expect(payload.result.valid).toBe(true)
})

Then('验证应该失败', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    validateRLSExpr: { result: { valid: boolean }; error: unknown }
  }).validateRLSExpr
  expect(payload.result.valid).toBe(false)
})

Then('验证应该失败并返回错误类型 {string}', function (
  this: ModelCraftWorld, errorType: string
) {
  const payload = (this.lastResponse as {
    validateRLSExpr: {
      result: { valid: boolean; errors: Array<{ code: string }> }
      error: { __typename: string } | null
    }
  }).validateRLSExpr

  if (payload.error) {
    expect(payload.error.__typename).toBe(errorType)
  } else {
    expect(payload.result.valid).toBe(false)
    expect(payload.result.errors.length).toBeGreaterThan(0)
  }
})

Then('auth_schema 设置成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    setProjectAuthSchema: { error: unknown }
  }).setProjectAuthSchema
  expect(payload.error).toBeNull()
})

Then('项目应该包含 auth 变量 {string}', async function (
  this: ModelCraftWorld, varName: string
) {
  const res = await this.orgClient.query<{
    project: {
      project: {
        authSchema: {
          variables: Array<{ name: string }>
        }
      } | null
    }
  }>(GET_PROJECT_AUTH_SCHEMA, { orgName: this.orgName, slug: this.projectSlug })

  const variables = res.project.project?.authSchema.variables ?? []
  const found = variables.some(v => v.name === varName)
  expect(found).toBe(true)
})
