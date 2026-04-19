'use client'

import * as React from 'react'
import { Check, AlertTriangle } from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/web/components/ui/card'
import { cn } from '@/shared/utils'
import type { RLSPreset } from '@/types/rls'
import type { RLSPresetSelectorProps } from './types'

/**
 * Preset 配置定义
 */
const PRESET_CONFIGS: Array<{
  value: RLSPreset
  label: string
  description: string
  scenario: string
  isDangerous?: boolean
}> = [
  {
    value: 'READ_WRITE_OWNER',
    label: '仅所有者读写',
    description: '用户只能读写自己创建的数据',
    scenario: '适用于用户私有数据，如个人资料、私有文档',
  },
  {
    value: 'READ_ALL_WRITE_OWNER',
    label: '全员可读，仅所有者可写',
    description: '所有用户可读，但只有所有者可以修改',
    scenario: '适用于公开信息但受控修改的场景，如公开文章、系统配置',
  },
  {
    value: 'READ_ALL',
    label: '仅只读',
    description: '所有用户只读，无人可以修改',
    scenario: '适用于完全公开的数据，如分类目录、字典表',
  },
  {
    value: 'READ_WRITE_ALL',
    label: '完全开放',
    description: '所有用户可读写所有数据',
    scenario: '适用于开放协作场景，如公开 Wiki',
    isDangerous: true,
  },
  {
    value: 'NO_ACCESS',
    label: '禁止访问',
    description: '禁止所有用户访问（管理员除外）',
    scenario: '适用于敏感数据或维护中的模型',
  },
]

/**
 * 单个 Preset 卡片组件
 */
interface PresetCardProps {
  config: (typeof PRESET_CONFIGS)[number]
  isSelected: boolean
  disabled?: boolean
  onSelect: () => void
}

function PresetCard({ config, isSelected, disabled, onSelect }: PresetCardProps) {
  return (
    <Card
      className={cn(
        'relative cursor-pointer transition-all hover:border-primary/50',
        isSelected && 'border-primary ring-1 ring-primary',
        disabled && 'cursor-not-allowed opacity-50',
        config.isDangerous && 'border-destructive/30 hover:border-destructive/50'
      )}
      onClick={() => !disabled && onSelect()}
    >
      {isSelected && (
        <div className="absolute right-3 top-3 flex size-5 items-center justify-center rounded-full bg-primary text-primary-foreground">
          <Check className="size-3" />
        </div>
      )}
      {config.isDangerous && (
        <div className="absolute right-3 top-3 flex size-5 items-center justify-center text-destructive">
          <AlertTriangle className="size-4" />
        </div>
      )}
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">{config.label}</CardTitle>
        <CardDescription className="text-sm">{config.description}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-xs text-muted-foreground">
          <span className="font-medium">适用场景：</span>
          {config.scenario}
        </p>
      </CardContent>
    </Card>
  )
}

/**
 * RLS Preset 选择器
 *
 * 展示 5 种预设策略卡片，供用户选择
 */
export function RLSPresetSelector({ value, onChange, disabled }: RLSPresetSelectorProps) {
  const [pendingPreset, setPendingPreset] = React.useState<RLSPreset | null>(null)

  const handleSelect = React.useCallback(
    (preset: RLSPreset) => {
      if (preset === value) return

      const config = PRESET_CONFIGS.find((c) => c.value === preset)
      if (config?.isDangerous) {
        // 高危策略需要二次确认
        setPendingPreset(preset)
      } else {
        onChange(preset, true)
      }
    },
    [value, onChange]
  )

  const handleConfirmDangerous = React.useCallback(() => {
    if (pendingPreset) {
      onChange(pendingPreset, true)
      setPendingPreset(null)
    }
  }, [pendingPreset, onChange])

  const handleCancelDangerous = React.useCallback(() => {
    setPendingPreset(null)
  }, [])

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {PRESET_CONFIGS.map((config) => (
          <PresetCard
            key={config.value}
            config={config}
            isSelected={value === config.value}
            disabled={disabled}
            onSelect={() => handleSelect(config.value)}
          />
        ))}
      </div>

      {/* 高危策略确认弹窗 */}
      {pendingPreset && (
        <DangerConfirmDialogInternal
          open={!!pendingPreset}
          onOpenChange={(open) => !open && handleCancelDangerous()}
          title="确认选择高危策略"
          description={`"${PRESET_CONFIGS.find((c) => c.value === pendingPreset)?.label}" 策略将允许所有用户读写所有数据，存在数据安全风险，请确认是否继续？`}
          confirmText="确认选择"
          cancelText="取消"
          onConfirm={handleConfirmDangerous}
        />
      )}
    </div>
  )
}

/**
 * 内部使用的确认弹窗（避免循环依赖）
 */
interface DangerConfirmDialogInternalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  confirmText?: string
  cancelText?: string
  onConfirm: () => void
}

function DangerConfirmDialogInternal({
  open,
  onOpenChange,
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  onConfirm,
}: DangerConfirmDialogInternalProps) {
  const handleConfirm = React.useCallback(() => {
    onConfirm()
    onOpenChange(false)
  }, [onConfirm, onOpenChange])

  return (
    <div
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center bg-black/50',
        !open && 'pointer-events-none opacity-0'
      )}
      onClick={() => onOpenChange(false)}
    >
      <div
        className="w-full max-w-md rounded-lg border bg-background p-6 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="mb-2 text-lg font-semibold text-destructive">{title}</h3>
        <p className="mb-6 text-sm text-muted-foreground">{description}</p>
        <div className="flex justify-end gap-2">
          <button
            type="button"
            className="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-normal shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            onClick={() => onOpenChange(false)}
          >
            {cancelText}
          </button>
          <button
            type="button"
            className="inline-flex h-9 items-center justify-center rounded-md bg-destructive px-4 py-2 text-sm font-normal text-destructive-foreground shadow-sm transition-colors hover:bg-destructive/90"
            onClick={handleConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}
