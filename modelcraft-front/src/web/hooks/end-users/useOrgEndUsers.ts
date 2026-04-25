'use client'

// src/web/hooks/end-users/useOrgEndUsers.ts
// Org 级终端用户管理 hook（EndUser v2）

import { useState, useEffect, useCallback } from 'react'

export interface OrgEndUser {
  id: string
  username: string
  displayName?: string
  status: 'ACTIVE' | 'DISABLED'
  createdAt: string
}

export interface CreateEndUserPayload {
  username: string
  password: string
  displayName?: string
}

interface UseOrgEndUsersReturn {
  users: OrgEndUser[]
  isLoading: boolean
  error: string | null
  reload: () => void
  createUser: (payload: CreateEndUserPayload) => Promise<void>
  toggleUserStatus: (userId: string, status: 'ACTIVE' | 'DISABLED') => Promise<void>
  deleteUser: (userId: string) => Promise<void>
}

interface GraphQLErrorItem {
  message?: string
}

interface GraphQLPayload<TData> {
  data?: TData
  errors?: GraphQLErrorItem[]
}

interface ListEndUsersData {
  listEndUsers?: {
    connection?: {
      nodes?: Array<{
        id: string
        username: string
        isForbidden: boolean
        createdAt: string
      }>
    }
    error?: {
      __typename?: string
      message?: string
    } | null
  }
}

interface MutationError {
  __typename?: string
  message?: string
}

interface CreateEndUserData {
  createEndUser?: {
    endUser?: {
      id: string
      username: string
      isForbidden: boolean
      createdAt: string
    } | null
    error?: MutationError | null
  }
}

interface UpdateEndUserStatusData {
  updateEndUserStatus?: {
    endUser?: {
      id: string
    } | null
    error?: MutationError | null
  }
}

interface DeleteEndUserData {
  deleteEndUser?: {
    success?: boolean
    error?: MutationError | null
  }
}

const LIST_END_USERS_QUERY = `
  query ListOrgEndUsers($input: ListEndUsersInput) {
    listEndUsers(input: $input) {
      connection {
        nodes {
          id
          username
          isForbidden
          createdAt
        }
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

const CREATE_END_USER_MUTATION = `
  mutation CreateOrgEndUser($input: CreateEndUserInput!) {
    createEndUser(input: $input) {
      endUser {
        id
        username
        isForbidden
        createdAt
      }
      error {
        __typename
        ... on EndUserAlreadyExists {
          message
        }
        ... on EndUserPasswordTooWeak {
          message
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

const UPDATE_END_USER_STATUS_MUTATION = `
  mutation UpdateOrgEndUserStatus($input: UpdateEndUserStatusInput!) {
    updateEndUserStatus(input: $input) {
      endUser {
        id
      }
      error {
        __typename
        ... on EndUserNotFound {
          message
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

const DELETE_END_USER_MUTATION = `
  mutation DeleteOrgEndUser($input: DeleteEndUserInput!) {
    deleteEndUser(input: $input) {
      success
      error {
        __typename
        ... on EndUserNotFound {
          message
        }
      }
    }
  }
`

async function postOrgGraphQL<TData>(
  orgName: string,
  query: string,
  variables?: Record<string, unknown>
): Promise<TData> {
  const response = await fetch(`/graphql/org/${orgName}/`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, variables }),
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(`GraphQL request failed: ${response.status}`)
  }

  const payload = (await response.json()) as GraphQLPayload<TData>

  if (payload.errors && payload.errors.length > 0) {
    throw new Error(payload.errors[0]?.message || 'GraphQL error')
  }

  if (!payload.data) {
    throw new Error('GraphQL response missing data')
  }

  return payload.data
}

function pickMessage(error?: MutationError | null, fallback = '操作失败'): string {
  return error?.message || fallback
}

export function useOrgEndUsers(orgName: string): UseOrgEndUsersReturn {
  const [users, setUsers] = useState<OrgEndUser[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  const reload = useCallback(() => setVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName) return

    setIsLoading(true)
    setError(null)

    postOrgGraphQL<ListEndUsersData>(orgName, LIST_END_USERS_QUERY, { input: { first: 500 } })
      .then((data) => {
        const payloadError = data.listEndUsers?.error
        if (payloadError?.message) {
          throw new Error(payloadError.message)
        }

        const nodes = data.listEndUsers?.connection?.nodes ?? []
        setUsers(
          nodes.map((u) => ({
            id: u.id,
            username: u.username,
            displayName: undefined,
            status: u.isForbidden ? 'DISABLED' : 'ACTIVE',
            createdAt: u.createdAt,
          }))
        )
      })
      .catch((e: unknown) => setError(e instanceof Error ? e.message : '加载终端用户失败'))
      .finally(() => setIsLoading(false))
  }, [orgName, version])

  const createUser = useCallback(
    async (payload: CreateEndUserPayload) => {
      const data = await postOrgGraphQL<CreateEndUserData>(orgName, CREATE_END_USER_MUTATION, {
        input: {
          username: payload.username,
          password: payload.password,
        },
      })

      const result = data.createEndUser
      if (!result?.endUser) {
        throw new Error(pickMessage(result?.error, '创建用户失败'))
      }

      reload()
    },
    [orgName, reload]
  )

  const toggleUserStatus = useCallback(
    async (userId: string, status: 'ACTIVE' | 'DISABLED') => {
      const data = await postOrgGraphQL<UpdateEndUserStatusData>(orgName, UPDATE_END_USER_STATUS_MUTATION, {
        input: {
          userId,
          isForbidden: status === 'DISABLED',
        },
      })

      const result = data.updateEndUserStatus
      if (!result?.endUser) {
        throw new Error(pickMessage(result?.error, '更新用户状态失败'))
      }

      reload()
    },
    [orgName, reload]
  )

  const deleteUser = useCallback(
    async (userId: string) => {
      const data = await postOrgGraphQL<DeleteEndUserData>(orgName, DELETE_END_USER_MUTATION, {
        input: { userId },
      })

      const result = data.deleteEndUser
      if (!result?.success) {
        throw new Error(pickMessage(result?.error, '删除用户失败'))
      }

      reload()
    },
    [orgName, reload]
  )

  return { users, isLoading, error, reload, createUser, toggleUserStatus, deleteUser }
}
