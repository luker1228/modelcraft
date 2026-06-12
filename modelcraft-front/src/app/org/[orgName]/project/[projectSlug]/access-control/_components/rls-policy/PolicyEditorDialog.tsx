'use client'

import * as React from 'react'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Textarea } from '@web/components/ui/textarea'
import type { RlsAction } from '@/generated/graphql'

const ACTIONS: RlsAction[] = ['read', 'create', 'update', 'delete']

interface PolicyEditorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (data: {
    policyName: string
    action: RlsAction
    role: string
    usingExpr?: string
    withCheckExpr?: string
  }) => Promise<void>
  saving: boolean
}

export function PolicyEditorDialog({
  open,
  onOpenChange,
  onSave,
  saving,
}: PolicyEditorDialogProps) {
  const [policyName, setPolicyName] = React.useState('')
  const [action, setAction] = React.useState<RlsAction>('read')
  const [role, setRole] = React.useState('')
  const [usingExpr, setUsingExpr] = React.useState('')
  const [withCheckExpr, setWithCheckExpr] = React.useState('')

  React.useEffect(() => {
    if (open) {
      setPolicyName('')
      setAction('read')
      setRole('')
      setUsingExpr('')
      setWithCheckExpr('')
    }
  }, [open])

  const handleSave = async () => {
    if (!policyName.trim()) { toast.error('请输入策略名称'); return }
    if (!role.trim()) { toast.error('请输入角色'); return }
    await onSave({
      policyName: policyName.trim(),
      action,
      role: role.trim(),
      usingExpr: usingExpr.trim() || undefined,
      withCheckExpr: withCheckExpr.trim() || undefined,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>添加策略</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-1.5">
            <Label>
              策略名称 <span className="text-destructive">*</span>
            </Label>
            <Input
              value={policyName}
              onChange={(e) => setPolicyName(e.target.value)}
              placeholder="例如：admin_full_access"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>
                Action <span className="text-destructive">*</span>
              </Label>
              <Select value={action} onValueChange={(v) => setAction(v as RlsAction)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ACTIONS.map((a) => (
                    <SelectItem key={a} value={a}>{a}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>
                Role <span className="text-destructive">*</span>
              </Label>
              <Input
                value={role}
                onChange={(e) => setRole(e.target.value)}
                placeholder="例如：admin"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>Using Expr</Label>
            <Textarea
              value={usingExpr}
              onChange={(e) => setUsingExpr(e.target.value)}
              placeholder='例如：{"owner_id": {"equals": "{{user_id}}"}}'
              rows={3}
              className="font-mono text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <Label>Check Expr</Label>
            <Textarea
              value={withCheckExpr}
              onChange={(e) => setWithCheckExpr(e.target.value)}
              placeholder='例如：{"owner_id": {"equals": "{{user_id}}"}}'
              rows={3}
              className="font-mono text-xs"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
          <Button onClick={handleSave} disabled={saving} className="bg-primary text-primary-foreground hover:bg-primary/90">
            {saving ? <><Loader2 className="mr-2 size-4 animate-spin" />保存中...</> : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
