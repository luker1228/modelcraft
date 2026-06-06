import { describe, expect, it } from 'vitest'
import {
  API_DOC_SERVER_URL,
  buildFindManyCurlSnippet,
  buildModelApiAiPrompt,
  buildModelRuntimeEndpoint,
} from './model-api-docs'

const context = {
  orgName: 'acme',
  projectSlug: 'alpha-project',
  databaseName: 'users_db',
  modelName: 'User',
}

describe('model-api-docs', () => {
  it('builds the full runtime endpoint with real values', () => {
    expect(buildModelRuntimeEndpoint(context)).toBe(
      `${API_DOC_SERVER_URL}/end-user/graphql/org/acme/project/alpha-project/db/users_db/model/User`
    )
  })

  it('builds a curl snippet with the real endpoint and query example', () => {
    const snippet = buildFindManyCurlSnippet(context)
    const runtimePath = '/end-user/graphql/org/acme/project/alpha-project/db/users_db/model/User'

    expect(snippet).toContain(`SERVER_URL="${API_DOC_SERVER_URL}"`)
    expect(snippet).toContain('TOKEN="replace-with-your-api-token"')
    expect(snippet).toContain(`curl -X POST "\${SERVER_URL}${runtimePath}" \\`)
    expect(snippet).toContain('-H "Authorization: Bearer ${TOKEN}"')
    expect(snippet).toContain('-H "Content-Type: application/json"')
    expect(snippet).toContain(`-d '{"query":"query { findMany(take: 5, skip: 0) { items { id } } }"}'`)
    expect(snippet).not.toContain(`curl -X POST "${API_DOC_SERVER_URL}/end-user/graphql/`)
    expect(snippet).not.toContain('/api/bff/')
    expect(snippet).not.toContain('<projectSlug>')
  })

  it('builds an AI prompt with endpoint guidance and a placeholder business goal', () => {
    const prompt = buildModelApiAiPrompt(context)

    expect(prompt).toContain(API_DOC_SERVER_URL)
    expect(prompt).toContain(buildModelRuntimeEndpoint(context))
    expect(prompt).toContain('Authorization: Bearer <API_TOKEN>')
    expect(prompt).toContain('query { findMany(take: 5, skip: 0) { items { id } } }')
    expect(prompt).toContain('Business goal: [把你想查询或写入的业务目标写在这里]')
    expect(prompt).toContain('1. GraphQL 查询或变更')
    expect(prompt).toContain('2. 对应 curl 命令')
    expect(prompt).toContain('3. 如有字段需要替换，请明确标出')
  })
})
