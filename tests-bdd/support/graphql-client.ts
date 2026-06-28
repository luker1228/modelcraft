import 'dotenv/config'
import { GraphQLClient as GqlClient } from 'graphql-request'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

class BaseGraphQLClient {
  protected client: GqlClient
  protected token: string | null = null

  constructor(url: string) {
    this.client = new GqlClient(url)
  }

  setAuth(token: string): void {
    this.token = token
    this.client.setHeader('Authorization', `Bearer ${token}`)
  }

  protected setURL(url: string): void {
    this.client = new GqlClient(url)
    if (this.token) {
      this.client.setHeader('Authorization', `Bearer ${this.token}`)
    }
  }

  async query<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    const action = extractAction(document)
    return this.client.request<T>({ document, variables, requestHeaders: action ? { 'X-Action': action } : {} })
  }

  async mutate<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    const action = extractAction(document)
    return this.client.request<T>({ document, variables, requestHeaders: action ? { 'X-Action': action } : {} })
  }
}

/**
 * 从 GraphQL document 字符串中提取 X-Action 值（格式："{type}:{operationName}"）。
 * 例如：`query ListModelDatabases { ... }` → `"query:ListModelDatabases"`
 */
function extractAction(document: string): string | null {
  const match = /^\s*(query|mutation|subscription)\s+(\w+)/m.exec(document)
  if (!match) return null
  return `${match[1]}:${match[2]}`
}

export class GraphQLClient extends BaseGraphQLClient {
  constructor(orgName: string, projectSlug: string) {
    super(`${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/`)
  }
}

export class OrgGraphQLClient extends BaseGraphQLClient {
  private orgName: string

  constructor(orgName: string) {
    super(`${API_BASE_URL}/graphql/org/${orgName}/`)
    this.orgName = orgName
  }

  setOrgName(orgName: string): void {
    if (this.orgName === orgName) {
      return
    }
    this.orgName = orgName
    this.setURL(`${API_BASE_URL}/graphql/org/${orgName}/`)
  }
}

// End-user scoped project GraphQL client
// Hits /graphql/org/{orgName}/project/{projectSlug}/ but with end-user auth headers
export class EndUserProjectGraphQLClient extends BaseGraphQLClient {
  constructor(orgName: string, projectSlug: string) {
    super(`${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/`)
  }

  setEndUserAuth(userId: string): void {
    this.token = userId
    this.client.setHeader('X-User-ID', userId)
    this.client.setHeader('X-User-Type', 'end_user')
  }
}
