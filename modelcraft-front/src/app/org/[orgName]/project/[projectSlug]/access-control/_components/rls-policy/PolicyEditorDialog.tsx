'use client'

import * as React from 'react'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@web/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import type { RlsAction, RlsExprType } from '@/generated/graphql'
import { RlsExpressionEditor } from './RlsExpressionEditor'
import {
  getRlsExpressionType,
  validateRlsExpressionSyntax,
} from './rls-expression-utils'

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
  onDryRun: (input: {
    expression: string
    exprType: RlsExprType
  }) => Promise<{ success: boolean; message?: string }>
  saving: boolean
}

export function PolicyEditorDialog({
  open,
  onOpenChange,
  onSave,
  onDryRun,
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
    const usingSyntax = validateRlsExpressionSyntax(usingExpr)
    if (!usingSyntax.valid) { toast.error(`Using Expr ${usingSyntax.message}`); return }
    const checkSyntax = validateRlsExpressionSyntax(withCheckExpr)
    if (!checkSyntax.valid) { toast.error(`Check Expr ${checkSyntax.message}`); return }

    await onSave({
      policyName: policyName.trim(),
      action,
      role: role.trim(),
      usingExpr: usingExpr.trim() || undefined,
      withCheckExpr: withCheckExpr.trim() || undefined,
    })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex size-full flex-col overflow-hidden p-0 sm:w-[760px] sm:max-w-[760px]">
        <SheetHeader className="border-b border-border px-6 py-4">
          <SheetTitle className="text-base">添加策略</SheetTitle>
        </SheetHeader>
        <div className="flex-1 space-y-5 overflow-y-auto px-6 py-5">
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

          <RlsExpressionEditor
            label="Using Expr"
            value={usingExpr}
            onChange={setUsingExpr}
            exprType={getRlsExpressionType(action, 'using')}
            onDryRun={onDryRun}
          />

          <RlsExpressionEditor
            label="Check Expr"
            value={withCheckExpr}
            onChange={setWithCheckExpr}
            exprType={getRlsExpressionType(action, 'check')}
            onDryRun={onDryRun}
          />
        </div>
        <SheetFooter className="border-t border-border px-6 py-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
          <Button onClick={handleSave} disabled={saving} className="bg-primary text-primary-foreground hover:bg-primary/90">
            {saving ? <><Loader2 className="mr-2 size-4 animate-spin" />保存中...</> : '保存'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
