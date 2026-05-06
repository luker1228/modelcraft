'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { createProjectScopedClient } from '@api-client/apollo/public'
import { FIND_USERS } from '@api-client/end-user/graphql-docs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'

interface FormContext {
  orgName?: string
  projectSlug?: string
}

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
 * Renders a <Select> dropdown listing EndUsers via the project-scoped
 * `findUsers` query. Value is the EndUser UUID string.
 * Used only in the Tenant (design) workspace.
 *
 * formContext must provide: orgName, projectSlug
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean
  const ctx = (props.formContext as FormContext | undefined) ?? {}
  const orgName = ctx.orgName ?? ''
  const projectSlug = ctx.projectSlug ?? ''

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  // Stable client per (orgName, projectSlug) pair
  const client = useMemo(
    () => (orgName && projectSlug ? createProjectScopedClient(orgName, projectSlug) : null),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [orgName, projectSlug],
  )

  useEffect(() => {
    if (!client) return

    let cancelled = false
    setLoading(true)

    client
      .query<FindUsersData>({
        query: FIND_USERS,
        variables: { take: 200 },
        fetchPolicy: 'cache-first',
      })
      .then((result) => {
        if (!cancelled) {
          setUsers(result.data?.findUsers?.items ?? [])
          setLoading(false)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setLoading(false)
        }
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
