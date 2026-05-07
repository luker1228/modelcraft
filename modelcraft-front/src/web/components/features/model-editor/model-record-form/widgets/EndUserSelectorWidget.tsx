'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { getOrgScopedClient, createEndUserOrgScopedClient } from '@api-client/apollo/public'
import { getEndUserToken } from '@api-client/end-user/public'
import { FIND_USERS } from '@api-client/end-user/graphql-docs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'

interface UserNode {
  id: string
  username: string
  createdAt: string
}

interface FindUsersData {
  findUsers?: {
    items?: UserNode[]
    nextCursor?: string
    hasMore?: boolean
    reqId: string
  }
}

interface FormContext {
  orgName?: string
  projectSlug?: string
  workspaceMode?: 'design' | 'end_user'
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Org-scoped query: findUsers is served at /graphql/org/{orgName}/.
 * - design mode: uses getOrgScopedClient() (admin token via Cookie)
 * - end_user mode: uses createEndUserOrgScopedClient() (end-user Bearer token)
 *
 * Default option "— 自己 —" submits empty string; backend fills current user from JWT.
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean
  const { orgName, workspaceMode = 'design' } = (props.formContext ?? {}) as FormContext

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  const client = useMemo(() => {
    if (workspaceMode === 'end_user' && orgName) {
      const token = getEndUserToken()
      if (token) {
        return createEndUserOrgScopedClient(orgName, token)
      }
    }
    return getOrgScopedClient()
  }, [workspaceMode, orgName])

  useEffect(() => {
    let cancelled = false
    setLoading(true)

    client
      .query<FindUsersData>({
        query: FIND_USERS,
        variables: { first: 50 },
        fetchPolicy: 'cache-first',
      })
      .then((result) => {
        if (!cancelled) {
          setUsers(result.data?.findUsers?.items ?? [])
          setLoading(false)
        }
      })
      .catch(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [client])

  const handleChange = (val: string) => {
    onChange(val === '__none__' ? undefined : val)
  }

  return (
    <Select
      value={value ?? '__none__'}
      onValueChange={handleChange}
      disabled={disabled === true || readonly === true || loading}
    >
      <SelectTrigger>
        <SelectValue placeholder={loading ? '加载中...' : '选择用户'} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__">— 自己 —</SelectItem>
        {users.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            {user.username}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
