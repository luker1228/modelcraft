import 'dotenv/config'
import { GraphQLClient as GqlClient } from 'graphql-request'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

export class GraphQLClient {
  private client: GqlClient

  constructor(orgName: string, projectSlug: string) {
    const url = `${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/`
    this.client = new GqlClient(url)
  }

  setAuth(token: string): void {
    this.client.setHeader('Authorization', `Bearer ${token}`)
  }

  async query<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    return this.client.request<T>(document, variables)
  }

  async mutate<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    return this.client.request<T>(document, variables)
  }
}
