'use client'

import * as React from 'react'
import { ChevronLeft, ChevronRight, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@web/components/ui/sheet'
import { ScrollArea } from '@web/components/ui/scroll-area'
import { ColumnPolicyEditor } from './ColumnPolicyEditor'
import { RowScopeSelector } from './RowScopeSelector'

import {
  useCreatePermissionWizard,
  type WizardStep,
} from '@/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/permissions/useCreatePermissionWizard'
import type { EndUserPermissionAction } from '@/types'
import { cn } from '@/shared/utils'

// ---------------------------------------------------------------------------
// Mocks (Wave 1) — mirrors new/page.tsx, replaced by real data in Wave 2
// ---------------------------------------------------------------------------

const MOCK_MODELS = [
  { id: 'model-orders', displayName: 'Orders（订单）' },
  { id: 'model-customers', displayName: 'Customers（客户）' },
  { id: 'model-products', displayName: 'Products（产品）' },
]

const MOCK_FIELDS = [
  { name: 'id', title: 'ID', format: 'TEXT' },
  { name: 'name', title: '名称', format: 'TEXT' },
  { name: 'status', title: '状态', format: 'TEXT' },
  { name: 'owner', title: '负责人', format: 'EndUserRef' },
  { name: 'dept_id', title: '部门', format: 'TEXT' },
  { name: 'amount', title: '金额', format: 'NUMBER' },
]

// ---------------------------------------------------------------------------
// Action radio cards
// ---------------------------------------------------------------------------

interface ActionMeta {
  value: EndUserPermissionAction
  label: string
  description: string
}

const ACTIONS: ActionMeta[] = [
  { value: 'SELECT', label: '查询', description: '读取记录数据' },
  { value: 'INSERT', label: '新增', description: '创建新记录' },
  { value: 'UPDATE', label: '修改', description: '编辑现有记录' },
  { value: 'DELETE', label: '删除', description: '删除记录' },
  { value: 'EXPORT', label: '导出', description: '导出数据为文件' },
]

function ActionCard({
  meta,
  checked,
  onChange,
}: {
  meta: ActionMeta
  checked: boolean
  onChange: () => void
}) {
  return (
    <label
      className={cn(
        'flex cursor-pointer flex-col gap-1 rounded-md border px-3 py-2.5 transition-colors',
        checked
          ? 'border-primary bg-primary/5'
          : 'border-border hover:bg-muted/40',
      )}
    >
      <div className="flex items-center gap-2">
        <input
          type="radio"
          name="permission-action-sheet"
          value={meta.value}
          checked={checked}
          onChange={onChange}
          className="size-3.5 accent-primary"
        />
        <span className="text-sm font-semibold text-foreground">{meta.label}</span>
        <span className="ml-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
          {meta.value}
        </span>
      </div>
      <p className="pl-5 text-xs text-muted-foreground">{meta.description}</p>
    </label>
  )
}

// ---------------------------------------------------------------------------
// Step progress indicator (horizontal, compact)
// ---------------------------------------------------------------------------

const STEPS: { key: WizardStep; label: string }[] = [
  { key: 'model-action', label: '模型与动作' },
  { key: 'row-scope', label: '行策略' },
  { key: 'column-policy', label: '列策略' },
]

function StepIndicator({ currentStep }: { currentStep: WizardStep }) {
  const currentIndex = STEPS.findIndex((s) => s.key === currentStep)

  return (
    <div className="flex items-center gap-0">
      {STEPS.map((step, index) => {
        const isActive = index === currentIndex
        const isDone = index < currentIndex

        return (
          <React.Fragment key={step.key}>
            <div className="flex flex-col items-center gap-1">
              <div
                className={cn(
                  'flex size-6 items-center justify-center rounded-full text-xs font-semibold transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : isDone
                      ? 'bg-primary/20 text-primary'
                      : 'bg-muted text-muted-foreground',
                )}
              >
                {isDone ? '✓' : index + 1}
              </div>
              <span
                className={cn(
                  'w-16 text-center text-[11px]',
                  isActive ? 'font-semibold text-foreground' : 'text-muted-foreground',
                )}
              >
                {step.label}
              </span>
            </div>
            {index < STEPS.length - 1 && (
              <div
                className={cn(
                  'mb-4 h-px w-8 transition-colors',
                  isDone ? 'bg-primary/40' : 'bg-border',
                )}
              />
            )}
          </React.Fragment>
        )
      })}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Step panels
// ---------------------------------------------------------------------------

function StepModelAction({
  modelId,
  onModelChange,
  action,
  onActionChange,
  displayName,
  onDisplayNameChange,
  description,
  onDescriptionChange,
}: {
  modelId: string
  onModelChange: (id: string, displayName: string) => void
  action: EndUserPermissionAction | null
  onActionChange: (action: EndUserPermissionAction) => void
  displayName: string
  onDisplayNameChange: (v: string) => void
  description: string
  onDescriptionChange: (v: string) => void
}) {
  return (
    <div className="flex flex-col gap-5">
      {/* Model select */}
      <div className="flex flex-col gap-1.5">
        <Label className="text-sm font-semibold text-foreground">
          目标模型 <span className="text-destructive">*</span>
        </Label>
        <Select
          value={modelId}
          onValueChange={(val) => {
            const model = MOCK_MODELS.find((m) => m.id === val)
            onModelChange(val, model?.displayName ?? val)
          }}
        >
          <SelectTrigger className="w-full">
            <SelectValue placeholder="选择模型…" />
          </SelectTrigger>
          <SelectContent>
            {MOCK_MODELS.map((m) => (
              <SelectItem key={m.id} value={m.id}>
                {m.displayName}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Action cards */}
      <div className="flex flex-col gap-1.5">
        <Label className="text-sm font-semibold text-foreground">
          操作动作 <span className="text-destructive">*</span>
        </Label>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          {ACTIONS.map((meta) => (
            <ActionCard
              key={meta.value}
              meta={meta}
              checked={action === meta.value}
              onChange={() => onActionChange(meta.value)}
            />
          ))}
        </div>
      </div>

      {/* Optional metadata */}
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="sheet-perm-display-name" className="text-sm font-semibold text-foreground">
          显示名称
          <span className="ml-1 text-xs font-normal text-muted-foreground">（选填）</span>
        </Label>
        <Input
          id="sheet-perm-display-name"
          placeholder="如：订单查询权限"
          value={displayName}
          onChange={(e) => onDisplayNameChange(e.target.value)}
        />
      </div>
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="sheet-perm-description" className="text-sm font-semibold text-foreground">
          描述
          <span className="ml-1 text-xs font-normal text-muted-foreground">（选填）</span>
        </Label>
        <Textarea
          id="sheet-perm-description"
          placeholder="描述该权限点的用途…"
          className="resize-none"
          rows={3}
          value={description}
          onChange={(e) => onDescriptionChange(e.target.value)}
        />
      </div>
    </div>
  )
}

function StepRowScope({
  value,
  onChange,
  hasOwnerField,
  hasDeptIdField,
}: {
  value: import('@/types').EndUserRowScope
  onChange: (scope: import('@/types').EndUserRowScope) => void
  hasOwnerField: boolean
  hasDeptIdField: boolean
}) {
  return (
    <div className="flex flex-col gap-4">
      <div className="rounded-md border border-border bg-muted/20 px-4 py-3">
        <p className="text-sm text-muted-foreground">
          选择该权限点能访问哪些行。行策略在运行时根据当前用户的身份自动过滤数据。
        </p>
      </div>
      <RowScopeSelector
        value={value}
        onChange={onChange}
        hasOwnerField={hasOwnerField}
        hasDeptIdField={hasDeptIdField}
      />
    </div>
  )
}

function StepColumnPolicy({
  action,
  value,
  onChange,
}: {
  action: EndUserPermissionAction | null
  value: import('@/types').ColumnPolicy
  onChange: (policy: import('@/types').ColumnPolicy) => void
}) {
  return (
    <div className="flex flex-col gap-4">
      <div className="rounded-md border border-border bg-muted/20 px-4 py-3">
        <p className="text-sm text-muted-foreground">
          配置该权限点能访问哪些列，以及每列的访问模式。未单独配置的列使用默认模式。
        </p>
      </div>
      <ColumnPolicyEditor
        fields={MOCK_FIELDS}
        value={value}
        onChange={onChange}
        action={action ?? 'SELECT'}
      />
    </div>
  )
}

// ---------------------------------------------------------------------------
// CreatePermissionSheet
// ---------------------------------------------------------------------------

export interface CreatePermissionSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgName: string
  projectSlug: string
}

export function CreatePermissionSheet({
  open,
  onOpenChange,
  orgName,
  projectSlug,
}: CreatePermissionSheetProps) {
  const { state, updateField, goNext, goBack, reset, submit, submitting, submitError } =
    useCreatePermissionWizard(orgName, projectSlug, () => {
      toast.success('权限点创建成功')
      onOpenChange(false)
    })

  // Reset wizard state when sheet opens
  const prevOpen = React.useRef(false)
  React.useEffect(() => {
    if (open && !prevOpen.current) {
      reset()
    }
    prevOpen.current = open
  }, [open, reset])

  const currentStepIndex = STEPS.findIndex((s) => s.key === state.step)
  const isLastStep = currentStepIndex === STEPS.length - 1
  const isFirstStep = currentStepIndex === 0

  const canProceedFromStep1 = state.modelId !== '' && state.action !== null
  const canGoNext = state.step === 'model-action' ? canProceedFromStep1 : true

  const handleNext = async () => {
    if (isLastStep) {
      try {
        await submit()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : '创建权限点失败，请重试')
      }
    } else {
      goNext()
    }
  }

  const stepTitle =
    state.step === 'model-action'
      ? '选择模型与动作'
      : state.step === 'row-scope'
        ? '配置行策略'
        : '配置列策略'

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className="flex w-full flex-col gap-0 p-0 sm:max-w-lg"
        onInteractOutside={(e) => {
          // Prevent accidental close while submitting
          if (submitting) e.preventDefault()
        }}
      >
        {/* Header */}
        <SheetHeader className="border-b border-border px-6 py-5">
          <SheetTitle>创建权限点</SheetTitle>
          <SheetDescription>
            配置一个 Model 上的操作能力（动作 × 行策略 × 列策略）
          </SheetDescription>
        </SheetHeader>

        {/* Step indicator */}
        <div className="flex justify-center border-b border-border px-6 py-4">
          <StepIndicator currentStep={state.step} />
        </div>

        {/* Scrollable step content */}
        <ScrollArea className="flex-1">
          <div className="px-6 py-5">
            {/* Step subtitle */}
            <p className="mb-4 text-sm font-semibold text-foreground">{stepTitle}</p>

            {state.step === 'model-action' && (
              <StepModelAction
                modelId={state.modelId}
                onModelChange={(id, displayName) => {
                  updateField('modelId', id)
                  updateField('modelDisplayName', displayName)
                }}
                action={state.action}
                onActionChange={(action) => updateField('action', action)}
                displayName={state.displayName}
                onDisplayNameChange={(v) => updateField('displayName', v)}
                description={state.description}
                onDescriptionChange={(v) => updateField('description', v)}
              />
            )}

            {state.step === 'row-scope' && (
              <StepRowScope
                value={state.rowScope}
                onChange={(scope) => updateField('rowScope', scope)}
                hasOwnerField={state.hasOwnerField}
                hasDeptIdField={state.hasDeptIdField}
              />
            )}

            {state.step === 'column-policy' && (
              <StepColumnPolicy
                action={state.action}
                value={state.columnPolicy}
                onChange={(policy) => updateField('columnPolicy', policy)}
              />
            )}

            {/* Submit error */}
            {submitError && (
              <div className="mt-4 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
                {submitError}
              </div>
            )}
          </div>
        </ScrollArea>

        {/* Footer action bar */}
        <div className="flex items-center justify-between border-t border-border px-6 py-4">
          <div>
            {!isFirstStep && (
              <Button variant="outline" size="sm" onClick={goBack} disabled={submitting}>
                <ChevronLeft className="mr-1 size-3.5" />
                上一步
              </Button>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
            >
              取消
            </Button>
            <Button
              size="sm"
              onClick={() => { void handleNext() }}
              disabled={!canGoNext || submitting}
            >
              {submitting && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
              {isLastStep ? (
                submitting ? '创建中…' : '创建权限点'
              ) : (
                <>
                  下一步
                  <ChevronRight className="ml-1 size-3.5" />
                </>
              )}
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
