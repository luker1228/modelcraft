'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { createEndUserOrgScopedClient } from '@api-client/apollo/public'
import { getEndUserToken } from '@api-client/end-user/public'
import { useAuthStore } from '@shared/stores/auth-store'
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
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Always goes through /api/bff/graphql/end-user/org/{orgName}/.
 * Token is picked by availability: end-user token first, then admin token.
 * Gateway determines X-User-Type from JWT audience.
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
  const { orgName } = (props.formContext ?? {}) as FormContext

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  const client = useMemo(() => {
    if (!orgName) return null
    const token = getEndUserToken() || useAuthStore.getState().accessToken
    if (!token) return null
    return createEndUserOrgScopedClient(orgName, token)
  }, [orgName])

  useEffect(() => {
    if (!client) return

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
      .catch(() => {
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

