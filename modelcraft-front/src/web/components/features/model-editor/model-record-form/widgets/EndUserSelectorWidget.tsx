'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { createEndUserOrgScopedClient, getOrgScopedClient } from '@api-client/apollo/clients'
import { getEndUserToken } from '@api-client/end-user/public'
import { FIND_USERS } from '@api-client/end-user/graphql-docs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Badge } from '@web/components/ui/badge'

interface UserNode {
  id: string
  username: string
  isBuiltin: boolean
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
  workspaceMode?: 'design' | 'end_user'
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Client selection is driven by workspaceMode from formContext:
 * - 'design'   → Tenant client via /graphql/org/ (allowEndUser=true on backend, RBAC for tenant)
 * - 'end_user' → End-user client via /graphql/end-user/org/ (end-user token required)
 *
 * Fetches EndUsers via the org-scoped `findUsers` query.
 * - Builtin admin is pinned at the top with a 「系统」chip.
 * - No 「不指定」option: every record must have an owner.
 * - New records default to the builtin admin's ID.
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean
  const { orgName, workspaceMode = 'design' } = (props.formContext ?? {}) as FormContext

  console.log('[EndUserSelectorWidget] render', {
    orgName,
    workspaceMode,
    value,
    disabled,
    readonly,
    // 打印原始 formContext，确认 RJSF 是否真的传进来了
    rawFormContext: props.formContext,
    // 打印全量 props key，看有没有意外字段
    propsKeys: Object.keys(props),
  })

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  const client = useMemo(() => {
    console.log('[EndUserSelectorWidget] computing client', { orgName, workspaceMode })

    if (!orgName) {
      console.warn('[EndUserSelectorWidget] client=null: orgName is empty')
      return null
    }

    if (workspaceMode === 'end_user') {
      const token = getEndUserToken()
      console.log('[EndUserSelectorWidget] end_user mode, token present:', !!token)
      if (!token) {
        console.warn('[EndUserSelectorWidget] client=null: no end-user token in localStorage')
        return null
      }
      const c = createEndUserOrgScopedClient(orgName, token)
      console.log('[EndUserSelectorWidget] created end-user org client →', `/api/bff/graphql/end-user/org/${orgName}/`)
      return c
    }

    // Design path
    const c = getOrgScopedClient()
    console.log('[EndUserSelectorWidget] using tenant org-scoped client →', `/api/bff/graphql/org/${orgName}/`)
    return c
  }, [orgName, workspaceMode])

  useEffect(() => {
    console.log('[EndUserSelectorWidget] useEffect fired', { client: !!client, orgName, workspaceMode })

    if (!client) {
      console.warn('[EndUserSelectorWidget] skipping query: client is null')
      return
    }

    let cancelled = false
    setLoading(true)
    console.log('[EndUserSelectorWidget] firing findUsers query', { first: 50 })

    client
      .query<FindUsersData>({
        query: FIND_USERS,
        variables: { first: 50 },
        fetchPolicy: 'cache-first',
      })
      .then((result) => {
        console.log('[EndUserSelectorWidget] findUsers response', {
          errors: result.errors,
          itemCount: result.data?.findUsers?.items?.length ?? 0,
          items: result.data?.findUsers?.items,
          hasMore: result.data?.findUsers?.hasMore,
          reqId: result.data?.findUsers?.reqId,
        })
        if (!cancelled) {
          const items = result.data?.findUsers?.items ?? []
          setUsers(items)

          // Default new records to the builtin admin's ID
          if (!value) {
            const builtin = items.find((u) => u.isBuiltin)
            console.log('[EndUserSelectorWidget] auto-select builtin admin:', builtin ?? 'not found')
            if (builtin) {
              onChange(builtin.id)
            }
          }
          setLoading(false)
        }
      })
      .catch((err) => {
        console.error('[EndUserSelectorWidget] findUsers query failed', err)
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [client]) // eslint-disable-line react-hooks/exhaustive-deps

  // Builtin admin pinned at top, then the rest sorted by username
  const sortedUsers = useMemo(() => {
    const builtin = users.filter((u) => u.isBuiltin)
    const normal = users.filter((u) => !u.isBuiltin)
    return [...builtin, ...normal]
  }, [users])

  return (
    <Select
      value={value ?? ''}
      onValueChange={onChange}
      disabled={disabled === true || readonly === true || loading}
    >
      <SelectTrigger>
        <SelectValue placeholder={loading ? '加载中...' : '选择用户'} />
      </SelectTrigger>
      <SelectContent>
        {sortedUsers.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            <span className="flex items-center gap-2">
              {user.username}
              {user.isBuiltin && (
                <Badge variant="secondary" className="text-xs">
                  系统
                </Badge>
              )}
            </span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

