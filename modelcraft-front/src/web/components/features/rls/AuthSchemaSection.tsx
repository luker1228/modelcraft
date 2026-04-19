'use client'

import * as React from 'react'
import { Loader2, Shield, Info } from 'lucide-react'

import { Button } from '@/web/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/web/components/ui/card'
import { Alert, AlertDescription } from '@/web/components/ui/alert'
import { useAuthSchema } from '@web/hooks/rls/use-auth-schema'
import type { AuthVariableInput } from '@/types/rls'
import type { AuthSchemaSectionProps } from './types'
import { AuthVariableEditor } from './AuthVariableEditor'
import { LoadingSpinner } from '@/web/components/ui/loading-spinner'

/**
 * 认证变量配置区组件
 *
 * 集成 AuthVariableEditor，展示说明文字
 * 用于 Project 设置页
 */
export function AuthSchemaSection({
  orgName,
  projectSlug,
}: AuthSchemaSectionProps) {
  const { authSchema, loading, error, updateAuthSchema, updating } = useAuthSchema(
    orgName,
    projectSlug
  )

  // 本地编辑状态（排除内置 uid）
  const [editingVariables, setEditingVariables] = React.useState<AuthVariableInput[]>([])
  const [hasChanges, setHasChanges] = React.useState(false)

  // 初始化编辑状态
  React.useEffect(() => {
    if (authSchema) {
      // 过滤掉内置变量，只保留自定义变量
      const customVars = authSchema
        .filter((v) => !v.isBuiltin)
        .map((v) => ({
          name: v.name,
          source: v.source,
          type: v.type,
        }))
      setEditingVariables(customVars)
      setHasChanges(false)
    }
  }, [authSchema])

  const handleVariablesChange = React.useCallback((variables: AuthVariableInput[]) => {
    setEditingVariables(variables)
    setHasChanges(true)
  }, [])

  const handleSave = React.useCallback(async () => {
    // 过滤掉空名称的变量
    const validVariables = editingVariables.filter((v) => v.name.trim() !== '')
    const success = await updateAuthSchema(validVariables)
    if (success) {
      setHasChanges(false)
    }
  }, [editingVariables, updateAuthSchema])

  const handleCancel = React.useCallback(() => {
    if (authSchema) {
      const customVars = authSchema
        .filter((v) => !v.isBuiltin)
        .map((v) => ({
          name: v.name,
          source: v.source,
          type: v.type,
        }))
      setEditingVariables(customVars)
      setHasChanges(false)
    }
  }, [authSchema])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="size-5" />
            认证变量配置
          </CardTitle>
          <CardDescription>配置 JWT Token 中可用于 RLS 的变量</CardDescription>
        </CardHeader>
        <CardContent>
          <LoadingSpinner />
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
            认证变量配置
          </CardTitle>
          <CardDescription>配置 JWT Token 中可用于 RLS 的变量</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border border-destructive/20 bg-destructive/5 p-4 text-center">
            <p className="text-sm text-destructive">加载失败: {error.message}</p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => window.location.reload()}
              className="mt-2"
            >
              重试
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="size-5" />
          认证变量配置
        </CardTitle>
        <CardDescription>配置 JWT Token 中可用于 RLS 的变量</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* 说明文字 */}
        <Alert variant="default" className="bg-muted/50">
          <Info className="size-4" />
          <AlertDescription className="text-sm">
            <p className="mb-2">
              认证变量用于在 RLS 策略中引用终端用户的 JWT Token 中的字段。
            </p>
            <ul className="ml-4 list-disc space-y-1 text-muted-foreground">
              <li>
                <code className="rounded bg-muted px-1 py-0.5 text-xs">uid</code>{' '}
                是内置变量，对应 JWT 中的 user_id
              </li>
              <li>可以添加自定义变量，如 tenant_id、role 等</li>
              <li>Source 路径使用 jwt.xxx 格式，如 jwt.tenant_id</li>
            </ul>
          </AlertDescription>
        </Alert>

        {/* 变量编辑器 */}
        <AuthVariableEditor
          variables={[
            // 内置 uid
            ...(authSchema?.filter((v) => v.isBuiltin) || []),
            // 自定义变量（使用编辑状态）
            ...editingVariables.map((v) => ({
              ...v,
              isBuiltin: false as const,
            })),
          ]}
          onChange={(vars) => {
            // 只处理非内置变量的变更
            const customVars = vars.filter((v) => !v.isBuiltin)
            handleVariablesChange(
              customVars.map((v) => ({
                name: v.name,
                source: v.source,
                type: v.type,
              }))
            )
          }}
          disabled={updating}
        />

        {/* 操作按钮 */}
        {hasChanges && (
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleCancel}
              disabled={updating}
            >
              取消
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={handleSave}
              disabled={updating}
            >
              {updating && <Loader2 className="mr-2 size-4 animate-spin" />}
              保存
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
