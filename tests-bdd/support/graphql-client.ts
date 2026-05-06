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
    return this.client.request<T>(document, variables)
  }

  async mutate<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    return this.client.request<T>(document, variables)
  }
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

  /** Use X-Internal-Token header instead of JWT Bearer token.
   * Optionally pass a userId to set X-User-ID header for permission directive checks.
   */
  setInternalTokenAuth(internalToken: string, userId?: string): void {
    this.client.setHeader('X-Internal-Token', internalToken)
    if (userId) {
      this.client.setHeader('X-User-ID', userId)
    }
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
