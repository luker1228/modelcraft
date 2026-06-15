// RLS Policy V2 Step Definitions
// 对应 Feature: Runtime RLS 注入和行级过滤（policy v2 场景）

import { Given, Then, DataTable } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'

const UPSERT_RLS_POLICY = `
  mutation UpsertRlsPolicy($modelId: ID!, $input: RlsPolicyInput!) {
    upsertRlsPolicy(modelId: $modelId, input: $input) {
      policy {
        id
        policyName
        action
        role
        usingExpr
        withCheckExpr
      }
      error {
        __typename
        ... on InvalidInput { message }
        ... on ResourceNotFound { message resourceType }
      }
    }
  }
`

const GET_RLS_POLICIES = `
  query GetRlsPolicies($modelId: ID!) {
    rlsPolicies(modelId: $modelId) {
      id
      policyName
      action
      role
      usingExpr
      withCheckExpr
    }
  }
`

function getCurrentModelId(world: ModelCraftWorld): string {
  const id = world.createdModelIds[world.createdModelIds.length - 1]
  if (!id) throw new Error('没有可用的模型 ID')
  return id
}

Given('我为该模型配置以下 RLS v2 policies:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getCurrentModelId(this)
  const rows = table.hashes()
  const results: Array<{ policyName: string; action: string; role: string }> = []

  for (const row of rows) {
    const input: Record<string, string> = {
      policyName: row.policyName,
      action: row.action,
      role: row.role ?? '',
    }
    if (row.usingExpr) {
      input.usingExpr = row.usingExpr
    }
    if (row.withCheckExpr) {
      input.withCheckExpr = row.withCheckExpr
    }

    const res = await this.projectClient.mutate<{
      upsertRlsPolicy: {
        policy: { id: string; policyName: string; action: string; role: string } | null
        error: { __typename: string; message?: string } | null
      }
    }>(UPSERT_RLS_POLICY, {
      modelId,
      input,
    })

    this.lastResponse = { upsertRlsPolicy: res.upsertRlsPolicy }

    if (res.upsertRlsPolicy.error) {
      throw new Error(
        `RLS v2 policy ${row.policyName} 配置失败: ${res.upsertRlsPolicy.error.__typename} ${res.upsertRlsPolicy.error.message ?? ''}`,
      )
    }
    if (!res.upsertRlsPolicy.policy) {
      throw new Error(`RLS v2 policy ${row.policyName} 配置失败：未返回 policy`)
    }
    results.push({
      policyName: res.upsertRlsPolicy.policy.policyName,
      action: res.upsertRlsPolicy.policy.action,
      role: res.upsertRlsPolicy.policy.role,
    })
  }

  this.modelMap['rlsV2Policies'] = results
})

Then('该模型应该包含 {int} 条 RLS v2 policy', async function (
  this: ModelCraftWorld, expectedCount: number,
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    rlsPolicies: Array<{ id: string }>
  }>(GET_RLS_POLICIES, { modelId })

  expect(res.rlsPolicies.length).toBe(expectedCount)
})
