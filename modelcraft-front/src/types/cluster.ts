export type ClusterStatus = 'ACTIVE' | 'DISABLED'

export interface DatabaseConnectionInfo {
  host: string
  port: number
  username: string
  password: string
}

export interface DatabaseCluster {
  id: string
  projectSlug: string
  title: string
  description: string
  connectionInfo: DatabaseConnectionInfo
  connectionTimeout: number
  status: ClusterStatus
  version: number
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface Database {
  name: string
}

export interface ConnectionInfoInput {
  host: string
  port: number
  username: string
  password: string
  connectionTimeout?: number
}

export interface ClusterConnectionInput {
  title: string
  description?: string
  connectionInfo: ConnectionInfoInput
}

export interface UpdateClusterConnectionInput {
  title?: string
  description?: string
  connectionInfo?: ConnectionInfoInput
  skipConnectionTest?: boolean
}

export interface ListDatabasesInput {
  projectSlug: string
  offset?: number
  limit?: number
  search?: string
}

export interface TestDatabaseConnectionInput {
  projectSlug?: string
  connectionInfo?: ConnectionInfoInput
}
