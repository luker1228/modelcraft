'use client'

import * as React from 'react'
import { useParams, useRouter } from 'next/navigation'
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
import { RowScopeSelector, ColumnPolicyEditor } from '@web/components/features/rbac'

import {
  useCreatePermissionWizard,
  type WizardStep,
} from './_hooks/useCreatePermissionWizard'
import type { EndUserPermissionAction } from '@/types'
import { cn } from '@/shared/utils'

// ---------------------------------------------------------------------------
// Mocks (Wave 1)
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

interface ActionCardProps {
  meta: ActionMeta
  checked: boolean
  onChange: () => void
}

function ActionCard({ meta, checked, onChange }: ActionCardProps) {
  return (
    <label
      className={cn(
        'flex cursor-pointer flex-col gap-1 rounded-md border border-border px-4 py-3 transition-colors',
        checked
          ? 'border-primary bg-primary/5'
          : 'hover:bg-muted/40',
      )}
    >
      <div className="flex items-center gap-2">
        <input
          type="radio"
          name="permission-action"
          value={meta.value}
          checked={checked}
          onChange={onChange}
          className="size-4 accent-primary"
        />
        <span className="text-sm font-semibold text-foreground">{meta.label}</span>
        <span className="ml-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
          {meta.value}
        </span>
      </div>
      <p className="pl-6 text-xs text-muted-foreground">{meta.description}</p>
    </label>
  )
}

// ---------------------------------------------------------------------------
// Progress indicator
// ---------------------------------------------------------------------------

const STEPS: { key: WizardStep; label: string }[] = [
  { key: 'model-action', label: '模型与动作' },
  { key: 'row-scope', label: '行策略' },
  { key: 'column-policy', label: '列策略' },
]

interface StepIndicatorProps {
  currentStep: WizardStep
}

function StepIndicator({ currentStep }: StepIndicatorProps) {
  const currentIndex = STEPS.findIndex((s) => s.key === currentStep)

  return (
    <div className="flex items-center gap-0">
      {STEPS.map((step, index) => {
        const isActive = index === currentIndex
        const isDone = index < currentIndex

        return (
          <React.Fragment key={step.key}>
            {/* Circle */}
            <div className="flex flex-col items-center gap-1.5">
              <div
                className={cn(
                  'flex size-7 items-center justify-center rounded-full text-xs font-semibold transition-colors',
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
                  'w-20 text-center text-xs',
                  isActive ? 'font-semibold text-foreground' : 'text-muted-foreground',
                )}
              >
                {step.label}
              </span>
            </div>

            {/* Connector line (not after last item) */}
            {index < STEPS.length - 1 && (
              <div
                className={cn(
                  'mb-5 h-px w-12 transition-colors',
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
// Step content panels
// ---------------------------------------------------------------------------

// ── Step 1 ──────────────────────────────────────────────────────────────────

interface StepModelActionProps {
  modelId: string
  onModelChange: (id: string, displayName: string) => void
  action: EndUserPermissionAction | null
  onActionChange: (action: EndUserPermissionAction) => void
  displayName: string
  onDisplayNameChange: (v: string) => void
  description: string
  onDescriptionChange: (v: string) => void
}

function StepModelAction({
  modelId,
  onModelChange,
  action,
  onActionChange,
  displayName,
  onDisplayNameChange,
  description,
  onDescriptionChange,
}: StepModelActionProps) {
  return (
    <div className="flex flex-col gap-6">
      {/* Model select */}
      <div className="flex flex-col gap-2">
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
          <SelectTrigger className="w-full max-w-sm">
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
      <div className="flex flex-col gap-2">
        <Label className="text-sm font-semibold text-foreground">
          操作动作 <span className="text-destructive">*</span>
        </Label>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
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
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="flex flex-col gap-2">
          <Label htmlFor="perm-display-name" className="text-sm font-semibold text-foreground">
            显示名称
            <span className="ml-1 text-xs font-normal text-muted-foreground">（选填）</span>
          </Label>
          <Input
            id="perm-display-name"
            placeholder="如：订单查询权限"
            value={displayName}
            onChange={(e) => onDisplayNameChange(e.target.value)}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="perm-description" className="text-sm font-semibold text-foreground">
            描述
            <span className="ml-1 text-xs font-normal text-muted-foreground">（选填）</span>
          </Label>
          <Textarea
            id="perm-description"
            placeholder="描述该权限点的用途…"
            className="resize-none"
            rows={3}
            value={description}
            onChange={(e) => onDescriptionChange(e.target.value)}
          />
        </div>
      </div>
    </div>
  )
}

// ── Step 2 ──────────────────────────────────────────────────────────────────

interface StepRowScopeProps {
  value: import('@/types').EndUserRowScope
  onChange: (scope: import('@/types').EndUserRowScope) => void
  hasOwnerField: boolean
  hasDeptIdField: boolean
}

function StepRowScope({ value, onChange, hasOwnerField, hasDeptIdField }: StepRowScopeProps) {
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

// ── Step 3 ──────────────────────────────────────────────────────────────────

interface StepColumnPolicyProps {
  action: EndUserPermissionAction | null
  value: import('@/types').ColumnPolicy
  onChange: (policy: import('@/types').ColumnPolicy) => void
}

function StepColumnPolicy({ action, value, onChange }: StepColumnPolicyProps) {
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
// Page
// ---------------------------------------------------------------------------

export default function CreatePermissionPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const router = useRouter()
  const orgName = params.orgName
  const projectSlug = params.projectSlug

  const { state, updateField, goNext, goBack, submit, submitting, submitError } =
    useCreatePermissionWizard(orgName, projectSlug)

  const currentStepIndex = STEPS.findIndex((s) => s.key === state.step)
  const isLastStep = currentStepIndex === STEPS.length - 1
  const isFirstStep = currentStepIndex === 0

  // ── Step 1 validation: model + action are required to proceed
  const canProceedFromStep1 = state.modelId !== '' && state.action !== null

  const canGoNext = state.step === 'model-action' ? canProceedFromStep1 : true

  const handleNext = async () => {
    if (isLastStep) {
      try {
        await submit()
        toast.success('权限点创建成功')
      } catch (err) {
        toast.error(err instanceof Error ? err.message : '创建权限点失败，请重试')
      }
    } else {
      goNext()
    }
  }

  return (
    <main className="size-full overflow-y-auto bg-background">
      <div className="mx-auto w-full max-w-[760px] px-6 pb-16 pt-10 xl:px-8">
        {/* Page header */}
        <div className="mb-8 space-y-1">
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">创建权限点</h1>
          <p className="text-sm text-muted-foreground">
            配置一个 Model 上的操作能力（动作 × 行策略 × 列策略）
          </p>
        </div>

        {/* Progress indicator */}
        <div className="mb-10 flex justify-center">
          <StepIndicator currentStep={state.step} />
        </div>

        {/* Step content card */}
        <div className="rounded-lg border border-border bg-card px-6 py-6 shadow-sm">
          {/* Step title */}
          <h2 className="mb-5 text-base font-semibold text-foreground">
            {state.step === 'model-action' && '选择模型与动作'}
            {state.step === 'row-scope' && '配置行策略'}
            {state.step === 'column-policy' && '配置列策略'}
          </h2>

          {/* Step panels */}
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
        </div>

        {/* Submit error */}
        {submitError && (
          <div className="mt-4 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            {submitError}
          </div>
        )}

        {/* Bottom action row */}
        <div className="mt-6 flex items-center justify-between gap-3">
          {/* Left: back button */}
          <div>
            {!isFirstStep && (
              <Button
                variant="outline"
                onClick={goBack}
                disabled={submitting}
              >
                <ChevronLeft className="mr-1 size-4" />
                上一步
              </Button>
            )}
          </div>

          {/* Right: cancel + next/submit */}
          <div className="flex items-center gap-3">
            <Button
              variant="ghost"
              onClick={() => router.back()}
              disabled={submitting}
            >
              取消
            </Button>

            <Button
              onClick={() => { void handleNext() }}
              disabled={!canGoNext || submitting}
            >
              {submitting && <Loader2 className="mr-1.5 size-4 animate-spin" />}
              {isLastStep ? (
                submitting ? '创建中…' : '创建权限点'
              ) : (
                <>
                  下一步
                  <ChevronRight className="ml-1 size-4" />
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </main>
  )
}
