import * as React from "react"
import { Copy, RefreshCw, Loader2, Check } from "lucide-react"
import { cn } from "@/shared/utils"
import { Label } from "@web/components/ui/label"
import { Input } from "@web/components/ui/input"
import { Button } from "@web/components/ui/button"
import { Badge } from "@web/components/ui/badge"
import { toast } from "sonner"
import { Controller, type Control, type FieldValues } from "react-hook-form"

// ── Types ──────────────────────────────────────────────────────────────────────

export interface ReadOnlyField {
  label: string
  value: string | React.ReactNode
  description?: string
  copyable?: boolean
}

/**
 * Runtime-only control holder.
 *
 * react-hook-form's Control<TFieldValues> has strict variance constraints,
 * which makes different form schemas incompatible at compile time.
 * We keep this field as unknown and cast at the Controller boundary.
 */
type AnyControl = unknown

export interface IdentityFormSectionProps extends React.HTMLAttributes<HTMLDivElement> {
  /** 区块标题 */
  title?: string

  /** 显示模式：edit（默认）或 view */
  mode?: 'edit' | 'view'

  /** 显示名称字段 */
  displayNameField: {
    name: string
    label?: string
    placeholder?: string
    description?: string
    /** edit mode: react-hook-form control */
    control?: AnyControl
    /** view mode: display value */
    value?: string
  }

  /** 描述信息字段（可选） */
  descriptionField?: {
    name: string
    label?: string
    placeholder?: string
    /** edit mode: react-hook-form control */
    control?: AnyControl
    /** view mode: display value */
    value?: string
  }

  /** 标识符字段（只读展示，可选） */
  identifierField?: {
    name: string
    label?: string
    value: string
    description?: string
    /** URL 前缀，如 "org.example.com/" */
    prefix?: string
    copyable?: boolean
    regeneratable?: boolean
    onRegenerate?: () => void
  }

  /** 额外只读字段列表 */
  readOnlyFields?: ReadOnlyField[]

  /** 标识符生成状态 */
  identifierStatus?: {
    type: 'auto' | 'manual' | 'loading' | 'error'
    message?: string
  }

  /** 自定义操作按钮（覆盖默认取消/保存） */
  actions?: React.ReactNode

  /** 是否渲染操作按钮区域，默认 true */
  showActions?: boolean

  cancelText?: string
  saveText?: string
  onCancel?: () => void
  onSave?: () => void
  saveDisabled?: boolean
  saveLoading?: boolean
}

// ── Internal helpers ───────────────────────────────────────────────────────────

function FieldRow({
  children,
  withDivider = true,
}: {
  children: React.ReactNode
  withDivider?: boolean
}) {
  return (
    <div className={`flex items-start justify-between gap-4 px-5 py-4 ${withDivider ? 'border-b border-gray-100' : ''}`}>
      {children}
    </div>
  )
}

function FieldLayout({
  label,
  required,
  description,
  badge,
  children,
}: {
  label: string
  required?: boolean
  description?: string
  badge?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <>
      {/* left: label group - 38% fixed width */}
      <div className="w-[38%] flex-none">
        <div className="flex flex-wrap items-center gap-2">
          <Label className="text-sm font-medium leading-none text-foreground">
            {label}
            {required && <span className="ml-0.5 text-xs text-red-500">*</span>}
          </Label>
          {badge}
        </div>
        {description && (
          <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{description}</p>
        )}
      </div>
      {/* right: control - flex:1 */}
      <div className="flex-1">{children}</div>
    </>
  )
}

function IdStatusBadge({
  type,
  message,
}: {
  type: 'auto' | 'manual' | 'loading' | 'error'
  message?: string
}) {
  if (type === 'loading') {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
        <Loader2 className="size-3 animate-spin" />
        生成中…
      </span>
    )
  }
  if (type === 'error') {
    return (
      <Badge variant="destructive" className="px-1.5 py-0 text-xs font-normal">
        {message ?? '生成失败'}
      </Badge>
    )
  }
  if (type === 'auto') {
    return (
      <Badge variant="secondary" className="px-1.5 py-0 text-xs font-normal">
        自动生成
      </Badge>
    )
  }
  if (type === 'manual') {
    return (
      <Badge variant="outline" className="px-1.5 py-0 text-xs font-normal">
        手动设置
      </Badge>
    )
  }
  return null
}

function CopyBtn({ value, label }: { value: string; label: string }) {
  const [done, setDone] = React.useState(false)

  const copy = () => {
    navigator.clipboard.writeText(value)
    toast.success(`${label} 已复制到剪贴板`)
    setDone(true)
    setTimeout(() => setDone(false), 1500)
  }

  return (
    <button
      type="button"
      onClick={copy}
      title={`复制 ${label}`}
      className={cn(
        "h-9 w-9 flex items-center justify-center",
        "border border-l-0 border-input rounded-r-md",
        "bg-muted hover:bg-accent transition-colors"
      )}
    >
      {done ? (
        <Check className="size-3.5 text-[#059669]" />
      ) : (
        <Copy className="size-3.5 text-muted-foreground" />
      )}
    </button>
  )
}

// ── Main component ─────────────────────────────────────────────────────────────

export const IdentityFormSection = React.forwardRef<
  HTMLDivElement,
  IdentityFormSectionProps
>(
  (
    {
      title,
      mode = 'edit',
      displayNameField,
      descriptionField,
      identifierField,
      readOnlyFields = [],
      identifierStatus = { type: 'auto' },
      actions,
      showActions = true,
      cancelText = '取消',
      saveText = '保存',
      onCancel,
      onSave,
      saveDisabled = false,
      saveLoading = false,
      className,
      ...props
    },
    ref
  ) => {
    // ── Display name ──────────────────────────────────────────────────────────

    const renderDisplayName = (isLast: boolean) => {
      if (mode === 'view') {
        return (
          <FieldRow withDivider={!isLast}>
            <FieldLayout
              label={displayNameField.label ?? '显示名称'}
              description={displayNameField.description}
            >
              <div className="flex h-9 items-center rounded-md border border-input bg-muted/50 px-3 text-sm text-foreground">
                {displayNameField.value ?? (
                  <span className="italic text-muted-foreground">未设置</span>
                )}
              </div>
            </FieldLayout>
          </FieldRow>
        )
      }

      if (!displayNameField.control) {
        console.error('[IdentityFormSection] control is required in edit mode')
        return null
      }

      return (
        <FieldRow withDivider={!isLast}>
          <FieldLayout
            label={displayNameField.label ?? '显示名称'}
            required
            description={displayNameField.description}
          >
            <Controller
              control={displayNameField.control as Control}
              name={displayNameField.name}
              render={({ field }) => (
                <Input
                  {...field}
                  placeholder={displayNameField.placeholder ?? '输入显示名称'}
                  className="h-9 px-3 text-sm"
                />
              )}
            />
          </FieldLayout>
        </FieldRow>
      )
    }

    // ── Description ───────────────────────────────────────────────────────────

    const renderDescription = (isLast: boolean) => {
      if (!descriptionField) return null

      if (mode === 'view') {
        return (
          <FieldRow withDivider={!isLast}>
            <FieldLayout label={descriptionField.label ?? '描述'}>
              <div className="flex h-9 items-center rounded-md border border-input bg-muted/50 px-3 text-sm text-foreground">
                {descriptionField.value ?? (
                  <span className="italic text-muted-foreground">暂无描述</span>
                )}
              </div>
            </FieldLayout>
          </FieldRow>
        )
      }

      if (!descriptionField.control) {
        console.error('[IdentityFormSection] control is required for descriptionField in edit mode')
        return null
      }

      return (
        <FieldRow withDivider={!isLast}>
          <FieldLayout label={descriptionField.label ?? '描述'}>
            <Controller
              control={descriptionField.control as Control}
              name={descriptionField.name}
              render={({ field }) => (
                <Input
                  {...field}
                  placeholder={descriptionField.placeholder ?? ''}
                  className="h-9 px-3 text-sm"
                />
              )}
            />
          </FieldLayout>
        </FieldRow>
      )
    }

    // ── Identifier ────────────────────────────────────────────────────────────

    const renderIdentifier = (isLast: boolean) => {
      if (!identifierField?.value) return null

      const hasSuffix = identifierField.copyable || identifierField.regeneratable

      return (
        <FieldRow withDivider={!isLast}>
          <FieldLayout
            label={identifierField.label ?? '标识符'}
            description={identifierField.description}
            badge={
              <IdStatusBadge
                type={identifierStatus.type}
                message={identifierStatus.message}
              />
            }
          >
            <div className="flex w-full">
              {identifierField.prefix && (
                <div className="flex h-9 items-center whitespace-nowrap rounded-l-md border border-r-0 border-input bg-muted px-3 font-mono text-xs text-muted-foreground">
                  {identifierField.prefix}
                </div>
              )}
              <Input
                disabled
                value={identifierField.value}
                className={cn(
                  "h-9 font-mono text-sm bg-muted/50",
                  identifierField.prefix && "rounded-l-none",
                  hasSuffix && "rounded-r-none"
                )}
              />
              {identifierField.regeneratable && identifierField.onRegenerate && (
                <button
                  type="button"
                  onClick={identifierField.onRegenerate}
                  title="重新生成"
                  className={cn(
                    "h-9 w-9 flex items-center justify-center",
                    "border border-l-0 border-input",
                    identifierField.copyable ? "" : "rounded-r-md",
                    "bg-muted hover:bg-accent transition-colors"
                  )}
                >
                  <RefreshCw className="size-3.5 text-muted-foreground" />
                </button>
              )}
              {identifierField.copyable && (
                <CopyBtn
                  value={identifierField.value}
                  label={identifierField.label ?? '标识符'}
                />
              )}
            </div>
          </FieldLayout>
        </FieldRow>
      )
    }

    // ── Read-only extra fields ─────────────────────────────────────────────────

    const renderReadOnly = (isLast: boolean) =>
      readOnlyFields.map((f, i) => (
        <FieldRow key={i} withDivider={!(isLast && i === readOnlyFields.length - 1)}>
          <FieldLayout label={f.label} description={f.description}>
            <div className="flex w-full">
              <div
                className={cn(
                  "flex h-9 flex-1 items-center rounded-md border border-input bg-muted/50 px-3 text-sm",
                  f.copyable && typeof f.value === 'string' && "rounded-r-none"
                )}
              >
                {typeof f.value === 'string' ? (
                  <span className="truncate">{f.value}</span>
                ) : (
                  f.value
                )}
              </div>
              {f.copyable && typeof f.value === 'string' && (
                <CopyBtn value={f.value} label={f.label} />
              )}
            </div>
          </FieldLayout>
        </FieldRow>
      ))

    // ── Action bar ────────────────────────────────────────────────────────────

    const defaultActions = (
      <div className="flex items-center gap-2">
        {onCancel && (
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onCancel}
          disabled={saveLoading}
          className="h-9 px-4 text-sm font-medium"
        >
          {cancelText}
        </Button>
        )}
        <Button
          type="submit"
          size="sm"
          onClick={onSave}
          disabled={saveDisabled || saveLoading}
          className={cn(
            "h-9 px-4 text-sm font-medium border-0 text-white transition-colors duration-200",
            (saveDisabled || saveLoading)
              ? "cursor-not-allowed bg-blue-200"
              : "cursor-pointer bg-[#2563eb] hover:bg-[#1d4ed8]",
          )}
        >
          {saveLoading ? (
            <>
              <Loader2 className="mr-1.5 size-4 animate-spin" />
              保存中…
            </>
          ) : (
            saveText
          )}
        </Button>
      </div>
    )

    // Compute which field is the last visible one
    const hasDescription = descriptionField ? 1 : 0
    const hasIdentifier = identifierField?.value ? 1 : 0
    const readOnlyCount = readOnlyFields.length

    return (
      <div ref={ref} className={cn("overflow-hidden rounded-lg border border-gray-200 bg-white", className)} {...props}>
        {title && (
          <div className="flex items-center gap-2 border-b border-gray-200 px-5 py-4">
            <h3 className="text-sm font-semibold text-foreground">{title}</h3>
          </div>
        )}

        <div>
          {renderDisplayName(!hasIdentifier && !hasDescription && readOnlyCount === 0)}
          {renderIdentifier(!hasDescription && readOnlyCount === 0)}
          {renderDescription(readOnlyCount === 0)}
          {renderReadOnly(true)}
        </div>

        {showActions && (
          <div className="flex items-center justify-end border-t border-gray-200 bg-gray-50 px-5 py-3">
            {actions ?? defaultActions}
          </div>
        )}
      </div>
    )
  }
)

IdentityFormSection.displayName = 'IdentityFormSection'
