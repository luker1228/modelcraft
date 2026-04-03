'use client'

import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { Badge } from '@web/components/ui/badge'
import { Button } from '@web/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@web/components/ui/alert-dialog'
import type { Role } from '@/types'
import { Shield, Trash2 } from 'lucide-react'

// Re-export AlertDialog components from ui/dialog since we may not have alert-dialog
// This component uses a simple confirm pattern

interface RoleTableProps {
  roles: Role[]
  onDelete?: (roleId: string) => void
}

export function RoleTable({ roles, onDelete }: RoleTableProps) {
  const [deleteTarget, setDeleteTarget] = useState<Role | null>(null)

  if (roles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Shield className="mb-3 size-10 text-muted-foreground/50" />
        <p className="text-sm text-muted-foreground">No roles found.</p>
      </div>
    )
  }

  const handleConfirmDelete = () => {
    if (deleteTarget && onDelete) {
      onDelete(deleteTarget.id)
    }
    setDeleteTarget(null)
  }

  return (
    <>
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Permissions</TableHead>
              <TableHead>Type</TableHead>
              {onDelete && <TableHead className="w-[80px]">Actions</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {roles.map((role) => (
              <TableRow key={role.id}>
                <TableCell className="font-semibold">{role.name}</TableCell>
                <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                  {role.description || '-'}
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {role.permissions.slice(0, 4).map((perm) => (
                      <Badge
                        key={perm}
                        variant="secondary"
                        className="text-xs"
                      >
                        {perm}
                      </Badge>
                    ))}
                    {role.permissions.length > 4 && (
                      <Badge variant="outline" className="text-xs">
                        +{role.permissions.length - 4}
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant={role.isSystem ? 'default' : 'outline'}>
                    {role.isSystem ? 'System' : 'Custom'}
                  </Badge>
                </TableCell>
                {onDelete && (
                  <TableCell>
                    {!role.isSystem && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="size-8 p-0 text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteTarget(role)}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    )}
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Delete Confirmation Dialog */}
      {deleteTarget && (
        <DeleteConfirmDialog
          roleName={deleteTarget.name}
          open={!!deleteTarget}
          onOpenChange={(open) => !open && setDeleteTarget(null)}
          onConfirm={handleConfirmDelete}
        />
      )}
    </>
  )
}

function DeleteConfirmDialog({
  roleName,
  open,
  onOpenChange,
  onConfirm,
}: {
  roleName: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Role</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the role &quot;{roleName}&quot;? This
            action cannot be undone. Users with this role will lose their
            assigned permissions.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
