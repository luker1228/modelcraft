'use client'

import { useState } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { GET_ROLES, CREATE_ROLE, DELETE_ROLE } from '@/api-client/user'
import { RoleTable } from '@web/components/features/settings/RoleTable'
import { CreateRoleDialog } from '@web/components/features/settings/CreateRoleDialog'
import { Button } from '@web/components/ui/button'
import { usePermission } from '@web/hooks/auth/use-permission'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { Role } from '@/types'

interface RolesQueryData {
  roles: Role[]
}

interface RolePayloadError {
  message: string
}

interface CreateRoleResult {
  createRole: {
    error?: RolePayloadError
    role?: Role
  }
}

interface DeleteRoleResult {
  deleteRole: {
    error?: RolePayloadError
    success?: boolean
  }
}

export default function DevelopersRolesPage() {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const canManageRoles = usePermission('*')

  const { data, loading, error, refetch } = useQuery<RolesQueryData>(GET_ROLES)

  const [createRole, { loading: creating }] = useMutation<CreateRoleResult>(CREATE_ROLE, {
    onCompleted: (result) => {
      if (result.createRole.error) {
        toast.error(result.createRole.error.message)
        return
      }
      toast.success('角色已创建')
      setShowCreateDialog(false)
      refetch()
    },
    onError: (err) => {
      toast.error(`创建角色失败：${err.message}`)
    },
  })

  const [deleteRole] = useMutation<DeleteRoleResult>(DELETE_ROLE, {
    onCompleted: (result) => {
      if (result.deleteRole.error) {
        toast.error(result.deleteRole.error.message)
        return
      }
      toast.success('角色已删除')
      refetch()
    },
    onError: (err) => {
      toast.error(`删除角色失败：${err.message}`)
    },
  })

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="size-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-md border border-border bg-destructive/5 px-4 py-3">
        <p className="text-sm text-destructive">加载角色失败：{error.message}</p>
      </div>
    )
  }

  const roles = data?.roles ?? []

  const handleCreateRole = (input: {
    name: string
    description?: string
    permissions: string[]
  }) => {
    createRole({ variables: { input } })
  }

  const handleDeleteRole = (roleId: string) => {
    deleteRole({ variables: { id: roleId } })
  }

  return (
    <div>
      {canManageRoles && (
        <div className="mb-4 flex justify-end">
          <Button size="sm" onClick={() => setShowCreateDialog(true)}>
            <Plus className="mr-1.5 size-4" />
            新建角色
          </Button>
        </div>
      )}

      <RoleTable
        roles={roles}
        onDelete={canManageRoles ? handleDeleteRole : undefined}
      />

      <CreateRoleDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSubmit={handleCreateRole}
        loading={creating}
      />
    </div>
  )
}
