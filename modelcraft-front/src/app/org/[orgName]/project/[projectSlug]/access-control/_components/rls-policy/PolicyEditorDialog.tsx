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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@web/components/ui/tooltip'
import type { RlsAction, RlsExprType } from '@/generated/graphql'
import { RlsExpressionEditor } from './RlsExpressionEditor'
import {
  buildRlsExpressionHelp,
  getRlsExpressionType,
  shouldShowRlsCheckExpression,
  shouldShowRlsUsingExpression,
  validateRlsExpressionSyntax,
} from './rls-expression-utils'

const ACTIONS: RlsAction[] = ['read', 'create', 'update', 'delete']

interface EditingPolicy {
  policyName: string
  action: RlsAction
  role: string
  usingExpr?: string | null
  withCheckExpr?: string | null
}

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
  modelFields?: Array<{ name: string; title?: string | null }>
  authVariables?: Array<{ name: string; source?: string | null; type?: string | null }>
  docsHref?: string
  editingPolicy?: EditingPolicy | null
}

export function PolicyEditorDialog({
  open,
  onOpenChange,
  onSave,
  onDryRun,
  saving,
  modelFields = [],
  authVariables = [],
  docsHref,
  editingPolicy = null,
}: PolicyEditorDialogProps) {
  const isEditing = !!editingPolicy
  const [policyName, setPolicyName] = React.useState('')
  const [action, setAction] = React.useState<RlsAction>('read')
  const [role, setRole] = React.useState('')
  const [usingExpr, setUsingExpr] = React.useState('')
  const [withCheckExpr, setWithCheckExpr] = React.useState('')
  // null = not yet validated, true = passed, false = failed
  const [usingExprValid, setUsingExprValid] = React.useState<boolean | null>(null)
  const [checkExprValid, setCheckExprValid] = React.useState<boolean | null>(null)
  const showUsingExpr = shouldShowRlsUsingExpression(action)
  const showCheckExpr = shouldShowRlsCheckExpression(action)
  const usingHelp = React.useMemo(() => buildRlsExpressionHelp('using', modelFields), [modelFields])
  const checkHelp = React.useMemo(() => buildRlsExpressionHelp('check', modelFields), [modelFields])

  React.useEffect(() => {
    if (open) {
      if (editingPolicy) {
        setPolicyName(editingPolicy.policyName)
        setAction(editingPolicy.action)
        setRole(editingPolicy.role)
        setUsingExpr(editingPolicy.usingExpr ?? '')
        setWithCheckExpr(editingPolicy.withCheckExpr ?? '')
      } else {
        setPolicyName('')
        setAction('read')
        setRole('')
        setUsingExpr('')
        setWithCheckExpr('')
      }
      setUsingExprValid(null)
      setCheckExprValid(null)
    }
  }, [open, editingPolicy])

  const usingExprBlocked = showUsingExpr && usingExpr.trim().length > 0 && usingExprValid === false
  const checkExprBlocked = showCheckExpr && withCheckExpr.trim().length > 0 && checkExprValid === false
  const saveDisabled = saving || usingExprBlocked || checkExprBlocked

  const saveTooltip = (() => {
    if (usingExprBlocked && checkExprBlocked) return 'Using Filter 和 Input Check 校验未通过，请修改后重试'
    if (usingExprBlocked) return 'Using Filter 校验未通过，请修改后重试'
    if (checkExprBlocked) return 'Input Check 校验未通过，请修改后重试'
    return null
  })()

  const handleSave = async () => {
    if (!policyName.trim()) { toast.error('请输入策略名称'); return }
    if (!role.trim()) { toast.error('请输入角色'); return }
    if (showUsingExpr) {
      const usingSyntax = validateRlsExpressionSyntax(usingExpr)
      if (!usingSyntax.valid) { toast.error(`Using Filter ${usingSyntax.message}`); return }
    }
    if (showCheckExpr) {
      const checkSyntax = validateRlsExpressionSyntax(withCheckExpr)
      if (!checkSyntax.valid) { toast.error(`Input Check ${checkSyntax.message}`); return }
    }

    await onSave({
      policyName: policyName.trim(),
      action,
      role: role.trim(),
      usingExpr: showUsingExpr ? usingExpr.trim() || undefined : undefined,
      withCheckExpr: showCheckExpr ? withCheckExpr.trim() || undefined : undefined,
    })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex size-full flex-col overflow-hidden p-0 sm:w-[760px] sm:max-w-[760px]">
        <SheetHeader className="border-b border-border px-6 py-4">
          <SheetTitle className="text-base">{isEditing ? '编辑策略' : '添加策略'}</SheetTitle>
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
              disabled={isEditing}
              readOnly={isEditing}
            />
            {isEditing && (
              <p className="text-xs text-muted-foreground">策略名称为唯一键，编辑时不可修改</p>
            )}
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

          {showUsingExpr && (
            <RlsExpressionEditor
              label="Using Filter"
              placeholder={usingHelp.placeholder}
              example={usingHelp.example}
              availableFields={usingHelp.availableFields}
              modelFields={modelFields}
              authVariables={authVariables}
              rootLabel={usingHelp.rootLabel}
              docsHref={docsHref}
              value={usingExpr}
              onChange={(v) => { setUsingExpr(v); setUsingExprValid(null) }}
              exprType={getRlsExpressionType(action, 'using')}
              onDryRun={onDryRun}
              onValidationChange={setUsingExprValid}
            />
          )}

          {showCheckExpr && (
            <RlsExpressionEditor
              label="Input Check"
              placeholder={checkHelp.placeholder}
              example={checkHelp.example}
              availableFields={checkHelp.availableFields}
              modelFields={modelFields}
              authVariables={authVariables}
              rootLabel={checkHelp.rootLabel}
              docsHref={docsHref}
              value={withCheckExpr}
              onChange={(v) => { setWithCheckExpr(v); setCheckExprValid(null) }}
              exprType={getRlsExpressionType(action, 'check')}
              onDryRun={onDryRun}
              onValidationChange={setCheckExprValid}
            />
          )}
        </div>
        <SheetFooter className="border-t border-border px-6 py-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <span tabIndex={saveDisabled ? 0 : undefined}>
                  <Button
                    onClick={handleSave}
                    disabled={saveDisabled}
                    className="bg-primary text-primary-foreground hover:bg-primary/90"
                  >
                    {saving ? <><Loader2 className="mr-2 size-4 animate-spin" />保存中...</> : '保存'}
                  </Button>
                </span>
              </TooltipTrigger>
              {saveTooltip && (
                <TooltipContent side="top">
                  <p>{saveTooltip}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
