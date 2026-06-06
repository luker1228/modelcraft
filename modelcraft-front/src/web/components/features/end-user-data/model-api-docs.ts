export const API_DOC_SERVER_URL = 'http://lukemxjia.devcloud.woa.com:9080'

export interface ModelApiDocContext {
  orgName: string
  projectSlug: string
  databaseName: string
  modelName: string
}

function buildModelRuntimePath(context: ModelApiDocContext): string {
  return `/end-user/graphql/org/${context.orgName}/project/${context.projectSlug}/db/${context.databaseName}/model/${context.modelName}`
}

export function buildModelRuntimeEndpoint(context: ModelApiDocContext): string {
  return `${API_DOC_SERVER_URL}${buildModelRuntimePath(context)}`
}

export function buildFindManyCurlSnippet(context: ModelApiDocContext): string {
  const runtimePath = buildModelRuntimePath(context)

  return [
    `SERVER_URL="${API_DOC_SERVER_URL}"`,
    'TOKEN="replace-with-your-api-token"',
    `curl -X POST "\${SERVER_URL}${runtimePath}" \\`,
    '  -H "Content-Type: application/json" \\',
    '  -H "Authorization: Bearer ${TOKEN}" \\',
    `  -d '{"query":"query { findMany(take: 5, skip: 0) { items { id } } }"}'`,
  ].join('\n')
}

export function buildModelApiAiPrompt(context: ModelApiDocContext): string {
  const endpoint = buildModelRuntimeEndpoint(context)

  return [
    'You are writing model-scoped runtime API docs.',
    `Server URL: ${API_DOC_SERVER_URL}`,
    `Endpoint: ${endpoint}`,
    'Use this authorization header pattern: Authorization: Bearer <API_TOKEN>',
    'Working example: query { findMany(take: 5, skip: 0) { items { id } } }',
    'Business goal: [把你想查询或写入的业务目标写在这里]',
    'Output constraints:',
    '1. GraphQL 查询或变更',
    '2. 对应 curl 命令',
    '3. 如有字段需要替换，请明确标出',
  ].join('\n')
}
