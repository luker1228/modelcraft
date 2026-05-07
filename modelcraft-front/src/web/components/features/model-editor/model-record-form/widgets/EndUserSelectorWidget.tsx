'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { getOrgScopedClient } from '@api-client/apollo/public'
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
    totalCount?: number
    reqId: string
  }
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Fetches EndUsers via the org-scoped `findUsers` query.
 * EndUser is Org-scoped (not Project-scoped), so uses getOrgScopedClient().
 * Used only in the Tenant (design) workspace.
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  const client = useMemo(() => getOrgScopedClient(), [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)

    client
      .query<FindUsersData>({
        query: FIND_USERS,
        variables: { take: 50 },
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
        <SelectItem value="__none__">— 不指定 —</SelectItem>
        {users.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            {user.username}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
