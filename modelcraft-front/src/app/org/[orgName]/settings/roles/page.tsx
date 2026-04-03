'use client'

import { useState } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { GET_ROLES, CREATE_ROLE, DELETE_ROLE } from '@web/graphql'
import { RoleTable } from '@web/components/settings/RoleTable'
import { CreateRoleDialog } from '@web/components/settings/CreateRoleDialog'
import { Button } from '@web/components/ui/button'
import { usePermission } from '@web/hooks/usePermission'
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

export default function RolesPage() {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const canManageRoles = usePermission('*') // Only owners can manage roles

  const { data, loading, error, refetch } = useQuery<RolesQueryData>(GET_ROLES)

  const [createRole, { loading: creating }] = useMutation<CreateRoleResult>(CREATE_ROLE, {
    onCompleted: (result) => {
      if (result.createRole.error) {
        const err = result.createRole.error
        toast.error(err.message)
        return
      }
      toast.success('Role created successfully')
      setShowCreateDialog(false)
      refetch()
    },
    onError: (err) => {
      toast.error(`Failed to create role: ${err.message}`)
    },
  })

  const [deleteRole] = useMutation<DeleteRoleResult>(DELETE_ROLE, {
    onCompleted: (result) => {
      if (result.deleteRole.error) {
        const err = result.deleteRole.error
        toast.error(err.message)
        return
      }
      toast.success('Role deleted successfully')
      refetch()
    },
    onError: (err) => {
      toast.error(`Failed to delete role: ${err.message}`)
    },
  })

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="size-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
        <p className="text-sm text-destructive">
          Failed to load roles: {error.message}
        </p>
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
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Roles</h2>
          <p className="text-sm text-muted-foreground">
            Manage roles and permissions for your organization.
          </p>
        </div>
        {canManageRoles && (
          <Button onClick={() => setShowCreateDialog(true)} size="sm">
            <Plus className="mr-1 size-4" />
            Create Role
          </Button>
        )}
      </div>

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
