'use client'

import React, { useMemo } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { useQuery } from '@apollo/client'
import { getOrgScopedClient } from '@api-client/apollo/public'
import { LIST_END_USERS } from '@api-client/end-user/graphql-docs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'

interface FormContext {
  orgName?: string
}

interface OrgEndUserNode {
  id: string
  username: string
  isForbidden: boolean
}

interface ListEndUsersData {
  listEndUsers?: {
    connection?: {
      nodes?: OrgEndUserNode[]
    }
    error?: { message?: string }
  }
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Renders a <Select> dropdown listing all active EndUsers in the Org.
 * Value is the EndUser UUID string. Used only in Tenant (design) workspace.
 *
 * formContext must provide: orgName
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean
  const ctx = (props.formContext as FormContext | undefined) ?? {}
  const orgName = ctx.orgName ?? ''

  const client = useMemo(() => getOrgScopedClient(), [])

  const { data, loading } = useQuery<ListEndUsersData>(LIST_END_USERS, {
    client,
    variables: { input: { first: 200 } },
    skip: !orgName,
    fetchPolicy: 'cache-first',
  })

  const users = data?.listEndUsers?.connection?.nodes ?? []
  const activeUsers = users.filter((u) => !u.isForbidden)

  const handleChange = (val: string) => {
    onChange(val === '__none__' ? undefined : val)
  }

  return (
    <Select
      value={(value as string | undefined) ?? '__none__'}
      onValueChange={handleChange}
      disabled={disabled === true || readonly === true || loading}
    >
      <SelectTrigger>
        <SelectValue placeholder={loading ? '加载中...' : '选择用户'} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__">— 不指定 —</SelectItem>
        {activeUsers.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            {user.username}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
