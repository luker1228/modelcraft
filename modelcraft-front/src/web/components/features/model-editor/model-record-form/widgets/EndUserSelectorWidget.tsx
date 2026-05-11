'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { createEndUserOrgScopedClient, getOrgScopedClient } from '@api-client/apollo/clients'
import { getEndUserToken } from '@api-client/end-user/public'
import { FIND_USERS } from '@api-client/end-user/graphql-docs'
import { useWidgetRouteContext } from '../_hooks/useWidgetRouteContext'
import { resolveWidgetFormContext } from './resolveWidgetFormContext'
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
 *
 * NOTE: In RJSF v6, formContext lives in props.registry.formContext — NOT as a direct
 * top-level prop. We read from registry first and fall back to URL pathname parsing
 * via useWidgetRouteContext so the widget is resilient across render paths.
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean

  // RJSF v6: formContext lives in props.registry.formContext.
  // Fall back to props.formContext for forward-compatibility.
  // See resolveWidgetFormContext.test.ts for the locked contract.
  const rawFormContext = resolveWidgetFormContext(
    props as unknown as Record<string, unknown>
  ) as FormContext | undefined
  const { workspaceMode = 'design' } = rawFormContext ?? {}

  // orgName: prefer formContext, fall back to URL pathname (same pattern as RelationSelector)
  const { orgName } = useWidgetRouteContext(rawFormContext)

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  const client = useMemo(() => {
    // orgName is required — without it we cannot build any Apollo client.
    // Return null and skip the query rather than throwing, so the form stays usable.
    if (!orgName) {
      console.warn(
        '[EndUserSelectorWidget] orgName is unavailable — skipping user query. ' +
        `workspaceMode=${workspaceMode}, formContext=${JSON.stringify(rawFormContext)}`
      )
      return null
    }

    if (workspaceMode === 'end_user') {
      const token = getEndUserToken()
      if (!token) {
        console.warn('[EndUserSelectorWidget] client=null: no end-user token in localStorage')
        return null
      }
      return createEndUserOrgScopedClient(orgName, token)
    }

    // Design path
    return getOrgScopedClient()
  }, [orgName, workspaceMode, rawFormContext])

  useEffect(() => {
    if (!client) {
      return
    }

    let cancelled = false
    setLoading(true)

    client
      .query<FindUsersData>({
        query: FIND_USERS,
        variables: { first: 50 },
        // network-only: 每次表单打开（widget mount）都必须拿最新用户列表。
        // 不能用 cache-first：Sheet 关闭再打开时 Apollo 会命中空缓存（第一次
        // 打开时后端可能确实没有用户），导致新建用户后再开表单仍然看不到用户。
        fetchPolicy: 'network-only',
      })
      .then((result) => {
        if (!cancelled) {
          const items = result.data?.findUsers?.items ?? []
          setUsers(items)

          // Default new records to the builtin admin's ID
          if (!value) {
            const builtin = items.find((u) => u.isBuiltin)
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

