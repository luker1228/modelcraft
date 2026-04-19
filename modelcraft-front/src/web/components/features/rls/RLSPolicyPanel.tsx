'use client'

import * as React from 'react'
import { Loader2, Shield, AlertTriangle } from 'lucide-react'

import { Button } from '@/web/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/web/components/ui/card'
import { Alert, AlertDescription } from '@/web/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/web/components/ui/tabs'
import { useRLSPolicy } from '@web/hooks/rls/use-rls-policy'
import type {
  RLSPreset,
  SetModelRLSPolicyInput,
  JsonExpr,
} from '@/types/rls'
import type { RLSPolicyPanelProps } from './types'
import { RLSPresetSelector } from './RLSPresetSelector'
import { DangerConfirmDialog } from './DangerConfirmDialog'
import { PRESET_CONFIGS } from '@/mocks/data/project/rls-factory'

/**
 * 预设类型到显示名称的映射
 */
const PRESET_LABELS: Record<RLSPreset, string> = {
  READ_WRITE_OWNER: '仅所有者读写',
  READ_ALL_WRITE_OWNER: '全员可读，仅所有者可写',
  READ_ALL: '仅只读',
  READ_WRITE_ALL: '完全开放',
  NO_ACCESS: '禁止访问',
}

/**
 * 表达式类型标签
 */
const EXPR_TYPE_LABELS = {
  selectPredicate: '查询条件 (SELECT)',
  insertCheck: '插入检查 (INSERT)',
  updatePredicate: '更新条件 (UPDATE)',
  updateCheck: '更新检查 (UPDATE CHECK)',
  deletePredicate: '删除条件 (DELETE)',
}

/**
 * JSON 预览组件（简化版）
 */
function PolicyJSONPreview({ value }: { value: JsonExpr | null | undefined }) {
  const jsonString = React.useMemo(() => {
    if (!value || Object.keys(value).length === 0) {
      return 'true'
    }
    return JSON.stringify(value, null, 2)
  }, [value])

  return (
    <pre className="max-h-[300px] overflow-auto rounded-md bg-muted p-4 font-mono text-xs">
      {jsonString}
    </pre>
  )
}

/**
 * 简单的条件编辑器占位组件
 */
function SimpleConditionEditor({
  value,
  onChange,
  placeholder,
}: {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}) {
  return (
    <textarea
      className="min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 font-mono text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder || '输入条件表达式，如 {"owner": {"_eq": {"_auth": "uid"}}}'}
    />
  )
}

/**
 * RLS 策略面板组件
 *
 * 集成 RLSPresetSelector、条件编辑器、JSON 预览
 * 支持保存策略和高危策略二次确认
 */
export function RLSPolicyPanel({
  modelId,
  modelName,
  policy,
  fields,
  onPolicyUpdated,
}: RLSPolicyPanelProps) {
  const {
    policy: currentPolicy,
    loading,
    error,
    updatePolicy,
    updating,
    refetch,
  } = useRLSPolicy(modelId)

  // 使用传入的 policy 或从 hook 获取的 policy
  const effectivePolicy = policy ?? currentPolicy

  // 本地编辑状态
  const [selectedPreset, setSelectedPreset] = React.useState<RLSPreset | null>(
    effectivePolicy?.preset || null
  )
  const [editingExpressions, setEditingExpressions] = React.useState({
    selectPredicate: effectivePolicy?.selectPredicate || 'true',
    insertCheck: effectivePolicy?.insertCheck || 'false',
    updatePredicate: effectivePolicy?.updatePredicate || 'false',
    updateCheck: effectivePolicy?.updateCheck || 'false',
    deletePredicate: effectivePolicy?.deletePredicate || 'false',
  })
  const [hasChanges, setHasChanges] = React.useState(false)

  // 高危策略确认弹窗
  const [showDangerConfirm, setShowDangerConfirm] = React.useState(false)
  const [pendingPreset, setPendingPreset] = React.useState<RLSPreset | null>(null)

  // 同步外部 policy 变化
  React.useEffect(() => {
    if (effectivePolicy) {
      setSelectedPreset(effectivePolicy.preset)
      setEditingExpressions({
        selectPredicate: effectivePolicy.selectPredicate || 'true',
        insertCheck: effectivePolicy.insertCheck || 'false',
        updatePredicate: effectivePolicy.updatePredicate || 'false',
        updateCheck: effectivePolicy.updateCheck || 'false',
        deletePredicate: effectivePolicy.deletePredicate || 'false',
      })
      setHasChanges(false)
    }
  }, [effectivePolicy])

  // 检查是否为高危策略
  const isDangerousPreset = React.useCallback((preset: RLSPreset | null): boolean => {
    if (!preset) return false
    const config = PRESET_CONFIGS.find((c) => c.value === preset)
    return Boolean(config?.isDangerous)
  }, [])

  const handlePresetChange = React.useCallback(
    (preset: RLSPreset, confirmed: boolean) => {
      setSelectedPreset(preset)
      setHasChanges(true)

      // 应用预设的表达式
      // 实际项目中应该从 rls-factory 获取预设表达式
      const presetExpressions = getPresetExpressions(preset)
      setEditingExpressions(presetExpressions)
    },
    []
  )

  const handleExpressionChange = React.useCallback(
    (type: keyof typeof editingExpressions, value: string) => {
      setEditingExpressions((prev) => ({
        ...prev,
        [type]: value,
      }))
      setHasChanges(true)

      // 如果手动修改表达式，清除 preset（变为自定义）
      if (selectedPreset) {
        setSelectedPreset(null)
      }
    },
    [selectedPreset]
  )

  const handleSave = React.useCallback(async () => {
    // 如果是高危策略且未确认，显示确认弹窗
    if (isDangerousPreset(selectedPreset) && !showDangerConfirm) {
      setPendingPreset(selectedPreset)
      setShowDangerConfirm(true)
      return
    }

    const input: SetModelRLSPolicyInput = {
      modelId,
      ...editingExpressions,
    }

    const result = await updatePolicy(input)
    if (result) {
      setHasChanges(false)
      onPolicyUpdated?.()
    }
    setShowDangerConfirm(false)
    setPendingPreset(null)
  }, [
    modelId,
    editingExpressions,
    selectedPreset,
    isDangerousPreset,
    showDangerConfirm,
    updatePolicy,
    onPolicyUpdated,
  ])

  const handleCancel = React.useCallback(() => {
    if (effectivePolicy) {
      setSelectedPreset(effectivePolicy.preset)
      setEditingExpressions({
        selectPredicate: effectivePolicy.selectPredicate || 'true',
        insertCheck: effectivePolicy.insertCheck || 'false',
        updatePredicate: effectivePolicy.updatePredicate || 'false',
        updateCheck: effectivePolicy.updateCheck || 'false',
        deletePredicate: effectivePolicy.deletePredicate || 'false',
      })
    }
    setHasChanges(false)
  }, [effectivePolicy])

  const handleConfirmDangerous = React.useCallback(() => {
    setShowDangerConfirm(false)
    handleSave()
  }, [handleSave])

  // 解析 JSON 用于预览
  const parsedExpressions = React.useMemo(() => {
    const parse = (str: string): JsonExpr | null => {
      try {
        if (str === 'true' || str === 'false') {
          return { _literal: str === 'true' }
        }
        return JSON.parse(str) as JsonExpr
      } catch {
        return null
      }
    }

    return {
      selectPredicate: parse(editingExpressions.selectPredicate),
      insertCheck: parse(editingExpressions.insertCheck),
      updatePredicate: parse(editingExpressions.updatePredicate),
      updateCheck: parse(editingExpressions.updateCheck),
      deletePredicate: parse(editingExpressions.deletePredicate),
    }
  }, [editingExpressions])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="size-5" />
            访问控制 (RLS)
          </CardTitle>
          <CardDescription>为模型 {modelName} 配置行级安全策略</CardDescription>
        </CardHeader>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="size-8 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="size-5" />
            访问控制 (RLS)
          </CardTitle>
          <CardDescription>为模型 {modelName} 配置行级安全策略</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border border-destructive/20 bg-destructive/5 p-4 text-center">
            <p className="text-sm text-destructive">加载失败: {error.message}</p>
            <Button variant="outline" size="sm" onClick={refetch} className="mt-2">
              重试
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  // 如果没有 policy，说明模型没有 owner 字段
  if (!effectivePolicy) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="size-5" />
            访问控制 (RLS)
          </CardTitle>
          <CardDescription>为模型 {modelName} 配置行级安全策略</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="default" className="bg-muted/50">
            <Info className="size-4" />
            <AlertDescription>
              该模型缺少 owner 字段，无法启用 RLS。请先添加一个 owner 类型的字段。
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* 预设选择器 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base font-semibold">选择预设策略</CardTitle>
          <CardDescription>快速应用常见的 RLS 策略组合</CardDescription>
        </CardHeader>
        <CardContent>
          <RLSPresetSelector
            value={selectedPreset}
            onChange={handlePresetChange}
            disabled={updating}
          />
        </CardContent>
      </Card>

      {/* 高级编辑 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base font-semibold">高级编辑</CardTitle>
          <CardDescription>手动编辑各个操作的访问条件</CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="selectPredicate" className="w-full">
            <TabsList className="grid w-full grid-cols-5">
              <TabsTrigger value="selectPredicate">查询</TabsTrigger>
              <TabsTrigger value="insertCheck">插入</TabsTrigger>
              <TabsTrigger value="updatePredicate">更新条件</TabsTrigger>
              <TabsTrigger value="updateCheck">更新检查</TabsTrigger>
              <TabsTrigger value="deletePredicate">删除</TabsTrigger>
            </TabsList>

            {(Object.keys(EXPR_TYPE_LABELS) as Array<keyof typeof EXPR_TYPE_LABELS>).map(
              (exprType) => (
                <TabsContent key={exprType} value={exprType} className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">
                      {EXPR_TYPE_LABELS[exprType]}
                    </label>
                    <SimpleConditionEditor
                      value={editingExpressions[exprType]}
                      onChange={(value) => handleExpressionChange(exprType, value)}
                    />
                  </div>

                  {/* JSON 预览 */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">预览</label>
                    <PolicyJSONPreview value={parsedExpressions[exprType]} />
                  </div>
                </TabsContent>
              )
            )}
          </Tabs>
        </CardContent>
      </Card>

      {/* 操作按钮 */}
      {hasChanges && (
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={handleCancel} disabled={updating}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={updating}>
            {updating && <Loader2 className="mr-2 size-4 animate-spin" />}
            保存策略
          </Button>
        </div>
      )}

      {/* 高危策略确认弹窗 */}
      <DangerConfirmDialog
        open={showDangerConfirm}
        onOpenChange={setShowDangerConfirm}
        title="确认选择高危策略"
        description={`"${pendingPreset ? PRESET_LABELS[pendingPreset] : ''}" 策略将允许所有用户读写所有数据，存在数据安全风险，请确认是否继续？`}
        confirmText="确认选择"
        cancelText="取消"
        onConfirm={handleConfirmDangerous}
      />
    </div>
  )
}

// 辅助函数：获取预设表达式
function getPresetExpressions(preset: RLSPreset) {
  const ownerExpr = JSON.stringify({ owner: { _eq: { _auth: 'uid' } } })

  switch (preset) {
    case 'READ_WRITE_OWNER':
      return {
        selectPredicate: ownerExpr,
        insertCheck: ownerExpr,
        updatePredicate: ownerExpr,
        updateCheck: ownerExpr,
        deletePredicate: ownerExpr,
      }
    case 'READ_ALL_WRITE_OWNER':
      return {
        selectPredicate: 'true',
        insertCheck: ownerExpr,
        updatePredicate: ownerExpr,
        updateCheck: ownerExpr,
        deletePredicate: ownerExpr,
      }
    case 'READ_ALL':
      return {
        selectPredicate: 'true',
        insertCheck: 'false',
        updatePredicate: 'false',
        updateCheck: 'false',
        deletePredicate: 'false',
      }
    case 'READ_WRITE_ALL':
      return {
        selectPredicate: 'true',
        insertCheck: 'true',
        updatePredicate: 'true',
        updateCheck: 'true',
        deletePredicate: 'true',
      }
    case 'NO_ACCESS':
      return {
        selectPredicate: 'false',
        insertCheck: 'false',
        updatePredicate: 'false',
        updateCheck: 'false',
        deletePredicate: 'false',
      }
    default:
      return {
        selectPredicate: 'true',
        insertCheck: 'false',
        updatePredicate: 'false',
        updateCheck: 'false',
        deletePredicate: 'false',
      }
  }
}

// Info 图标组件
function Info({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M12 16v-4" />
      <path d="M12 8h.01" />
    </svg>
  )
}
