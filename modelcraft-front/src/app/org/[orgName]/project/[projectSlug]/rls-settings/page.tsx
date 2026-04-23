'use client'

import * as React from 'react'
import { useQuery } from '@apollo/client'
import { useSearchParams, useParams } from 'next/navigation'
import { Check, ChevronsUpDown, Shield, UserRoundCog } from 'lucide-react'
import { toast } from 'sonner'

import { useProjectScopedClient } from '@bff/apollo/public'
import { useAppStore } from '@web/stores/app'
import { useRLSPolicy } from '@web/hooks/rls/use-rls-policy'
import { useAuthSchema } from '@web/hooks/rls/use-auth-schema'
import { DATABASE_CATALOG } from '@web/graphql/queries/cluster'
import { GET_MODELS_FOR_RELATION } from '@web/graphql/queries/model'
import { getPresetExpressions } from '@/mocks/data/project/rls-factory'
import type { AuthVariable, AuthVariableInput, RLSPreset } from '@/types/rls'

import { cn } from '@/shared/utils'
import { Badge } from '@/web/components/ui/badge'
import { Button } from '@/web/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/web/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/web/components/ui/dropdown-menu'
import {
  Command,
  CommandEmpty,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/web/components/ui/command'
import { Input } from '@/web/components/ui/input'
import { Label } from '@/web/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/web/components/ui/popover'
import { Separator } from '@/web/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/web/components/ui/sheet'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/web/components/ui/tabs'
import {
  AuthVariableEditor,
  PolicyJSONPreview,
} from '@/web/components/features/rls'

const DEFAULT_DATABASE_NAME = 'test_db'

const PRESET_LABELS: Record<RLSPreset, string> = {
  READ_WRITE_OWNER: '仅所有者可读写',
  READ_ALL_WRITE_OWNER: '全员可读，仅所有者可写',
  READ_ALL: '全员只读',
  READ_WRITE_ALL: '全员可读写',
  NO_ACCESS: '禁止访问',
}

const POLICY_MATCH_MODE_LABELS = {
  ALL: '全量匹配',
  ANY: '任一匹配',
} as const

const POLICY_PRESET_OPTIONS: Array<{ label: string; preset: RLSPreset | null }> = [
  { label: PRESET_LABELS.READ_WRITE_OWNER, preset: 'READ_WRITE_OWNER' },
  { label: PRESET_LABELS.READ_ALL_WRITE_OWNER, preset: 'READ_ALL_WRITE_OWNER' },
  { label: PRESET_LABELS.READ_ALL, preset: 'READ_ALL' },
  { label: PRESET_LABELS.READ_WRITE_ALL, preset: 'READ_WRITE_ALL' },
  { label: PRESET_LABELS.NO_ACCESS, preset: 'NO_ACCESS' },
  { label: '自定义策略', preset: null },
]

type ExprState = {
  selectPredicate: string
  insertCheck: string
  updatePredicate: string
  updateCheck: string
  deletePredicate: string
}

type PolicyCommand = '查询' | '插入' | '更新' | '删除'

type PolicyRow = {
  name: string
  command: PolicyCommand
  appliedRoles: string
}

interface ModelConnectionData {
  id: string
  name: string
}

interface ModelsForRelationData {
  models: {
    edges: Array<{
      node: {
        id: string
        name: string
      }
    }>
  }
}

interface ModelDatabaseCatalogData {
  modelDatabaseCatalog: {
    data: {
      databases: Array<{
        name: string
      }>
    } | null
  }
}

const POLICY_ACTIONS: Array<PolicyCommand> = ['查询', '插入', '更新', '删除']

const COMMAND_TO_TAB_KEY: Record<PolicyCommand, keyof ExprState> = {
  查询: 'selectPredicate',
  插入: 'insertCheck',
  更新: 'updatePredicate',
  删除: 'deletePredicate',
}

function parseExpr(input: string): Record<string, unknown> | null {
  try {
    if (input === 'true') return { _const: true }
    if (input === 'false') return { _const: false }
    return JSON.parse(input) as Record<string, unknown>
  } catch {
    return null
  }
}

export default function RLSSettingsPage() {
  const searchParams = useSearchParams()
  const params = useParams()

  const orgName = String(params.orgName || '')
  const projectSlug = String(params.projectSlug || '')
  const projectClient = useProjectScopedClient(projectSlug, orgName)

  const selectedProject = useAppStore((state) => state.selectedProject)

  const queryModelId = searchParams.get('modelId')?.trim() || ''
  const queryModelName = searchParams.get('modelName')?.trim() || ''
  const initialDatabaseName =
    searchParams.get('database')?.trim()
    || selectedProject?.databaseName
    || DEFAULT_DATABASE_NAME

  const [policyMatchMode, setPolicyMatchMode] = React.useState<'ALL' | 'ANY'>('ALL')
  const [databaseSelectorOpen, setDatabaseSelectorOpen] = React.useState(false)
  const [modelSelectorOpen, setModelSelectorOpen] = React.useState(false)
  const [selectedDatabase, setSelectedDatabase] = React.useState(initialDatabaseName)
  const [selectedModelId, setSelectedModelId] = React.useState(queryModelId)

  const { data: databaseData, loading: databasesLoading } = useQuery<ModelDatabaseCatalogData>(
    DATABASE_CATALOG,
    {
      variables: {
        input: {
          page: 1,
          pageSize: 100,
        },
      },
      skip: !orgName || !projectSlug,
      client: projectClient,
    }
  )

  const databaseOptions = React.useMemo(() => {
    const options = (databaseData?.modelDatabaseCatalog?.data?.databases ?? [])
      .map((item) => item.name)
      .filter((name) => name.trim() !== '')

    if (initialDatabaseName && !options.includes(initialDatabaseName)) {
      options.unshift(initialDatabaseName)
    }

    return options
  }, [databaseData, initialDatabaseName])

  React.useEffect(() => {
    if (databaseOptions.length === 0) return
    if (databaseOptions.includes(selectedDatabase)) return
    if (initialDatabaseName && databaseOptions.includes(initialDatabaseName)) {
      setSelectedDatabase(initialDatabaseName)
      return
    }
    setSelectedDatabase(databaseOptions[0])
  }, [databaseOptions, initialDatabaseName, selectedDatabase])

  const { data: modelsData, loading: modelsLoading } = useQuery<ModelsForRelationData>(
    GET_MODELS_FOR_RELATION,
    {
      variables: {
        input: {
          databaseName: selectedDatabase,
          limit: 100,
        },
      },
      skip: !selectedDatabase || !orgName || !projectSlug,
      client: projectClient,
    }
  )

  const modelOptions = React.useMemo<ModelConnectionData[]>(() => {
    const options =
      modelsData?.models?.edges
        ?.map((edge) => ({
          id: edge.node.id,
          name: edge.node.name,
        }))
        .filter((item) => item.id && item.name) ?? []

    const shouldIncludeQueryModel = selectedDatabase === initialDatabaseName
    if (shouldIncludeQueryModel && queryModelId && !options.some((item) => item.id === queryModelId)) {
      options.unshift({
        id: queryModelId,
        name: queryModelName || queryModelId,
      })
    }

    return options
  }, [modelsData, queryModelId, queryModelName, selectedDatabase, initialDatabaseName])

  React.useEffect(() => {
    if (modelOptions.length === 0) return

    const hasCurrent = modelOptions.some((m) => m.id === selectedModelId)
    if (hasCurrent) return

    if (queryModelId) {
      const byQueryId = modelOptions.find((m) => m.id === queryModelId)
      if (byQueryId) {
        setSelectedModelId(byQueryId.id)
        return
      }
    }

    if (queryModelName) {
      const byName = modelOptions.find((m) => m.name === queryModelName)
      if (byName) {
        setSelectedModelId(byName.id)
        return
      }
    }

    setSelectedModelId(modelOptions[0].id)
  }, [modelOptions, queryModelId, queryModelName, selectedModelId])

  const selectedModel = React.useMemo(
    () => modelOptions.find((m) => m.id === selectedModelId) || null,
    [modelOptions, selectedModelId]
  )

  const {
    policy,
    loading: policyLoading,
    error: policyError,
    updating: savingPolicy,
    updatePolicy,
  } = useRLSPolicy(selectedModelId, projectSlug)

  const {
    authSchema,
    loading: authLoading,
    error: authError,
    updating: savingSchema,
    updateAuthSchema,
  } = useAuthSchema(orgName, projectSlug)

  const [selectedPreset, setSelectedPreset] = React.useState<RLSPreset | null>(null)
  const [expressions, setExpressions] = React.useState<ExprState>({
    selectPredicate: 'true',
    insertCheck: 'false',
    updatePredicate: 'false',
    updateCheck: 'false',
    deletePredicate: 'false',
  })

  const [activePolicyCommand, setActivePolicyCommand] = React.useState<PolicyCommand | null>(null)
  const [editorTab, setEditorTab] = React.useState<keyof ExprState>('selectPredicate')
  const [policyDetailOpen, setPolicyDetailOpen] = React.useState(false)

  const [authVariables, setAuthVariables] = React.useState<AuthVariable[]>([])

  React.useEffect(() => {
    if (!policy) return
    setSelectedPreset(policy.preset)
    setExpressions({
      selectPredicate: policy.selectPredicate || 'true',
      insertCheck: policy.insertCheck || 'false',
      updatePredicate: policy.updatePredicate || 'false',
      updateCheck: policy.updateCheck || 'false',
      deletePredicate: policy.deletePredicate || 'false',
    })
  }, [policy])

  React.useEffect(() => {
    setActivePolicyCommand(null)
    setPolicyDetailOpen(false)
  }, [selectedModelId])

  const isRLSEnabled = Boolean(policy)

  React.useEffect(() => {
    setAuthVariables(authSchema)
  }, [authSchema])

  const isCustomMode = selectedPreset === null

  const parsed = React.useMemo(
    () => ({
      selectPredicate: parseExpr(expressions.selectPredicate),
      insertCheck: parseExpr(expressions.insertCheck),
      updatePredicate: parseExpr(expressions.updatePredicate),
      updateCheck: parseExpr(expressions.updateCheck),
      deletePredicate: parseExpr(expressions.deletePredicate),
    }),
    [expressions]
  )

  const policyRows = React.useMemo<PolicyRow[]>(() => {
    const appliedRoles = 'authenticated'

    if (selectedPreset) {
      const presetLabel = PRESET_LABELS[selectedPreset]
      return POLICY_ACTIONS.map((command) => ({
        name: `当前预设：${presetLabel}`,
        command,
        appliedRoles,
      }))
    }

    const customRowNameMap: Record<PolicyCommand, string> = {
      查询: '用户可查看自己的数据',
      插入: '用户可新增自己的数据',
      更新: '用户可更新自己的数据',
      删除: '用户可删除自己的数据',
    }

    return POLICY_ACTIONS.map((command) => ({
      name: customRowNameMap[command],
      command,
      appliedRoles,
    }))
  }, [selectedPreset])

  const handleCreatePolicy = React.useCallback((preset: RLSPreset | null) => {
    if (preset) {
      setSelectedPreset(preset)
      setExpressions({ ...getPresetExpressions(preset) })
      toast.success(`已应用预设策略：${PRESET_LABELS[preset]}`)
      return
    }

    setSelectedPreset(null)
    toast.success('已切换为自定义策略')
  }, [])

  const handleSavePolicy = React.useCallback(async () => {
    if (!selectedModelId) {
      toast.error('请选择模型')
      return
    }

    if (!isCustomMode) {
      toast.info('预设策略不可直接编辑，请先切换到“自定义策略”')
      return
    }

    const result = await updatePolicy({
      modelId: selectedModelId,
      ...expressions,
    })

    if (result) {
      setSelectedPreset(result.preset)
      setExpressions({
        selectPredicate: result.selectPredicate,
        insertCheck: result.insertCheck,
        updatePredicate: result.updatePredicate,
        updateCheck: result.updateCheck,
        deletePredicate: result.deletePredicate,
      })
    }
  }, [expressions, isCustomMode, selectedModelId, updatePolicy])

  const handleEnableRLS = React.useCallback(async () => {
    if (!selectedModelId) {
      toast.error('请选择模型')
      return
    }

    const preset: RLSPreset = 'READ_WRITE_OWNER'
    const initialExpressions = getPresetExpressions(preset)

    const result = await updatePolicy({
      modelId: selectedModelId,
      ...initialExpressions,
    })

    if (result) {
      setSelectedPreset(result.preset)
      setExpressions({
        selectPredicate: result.selectPredicate,
        insertCheck: result.insertCheck,
        updatePredicate: result.updatePredicate,
        updateCheck: result.updateCheck,
        deletePredicate: result.deletePredicate,
      })
      toast.success('已启用 RLS（默认策略：仅所有者可读写）')
    }
  }, [selectedModelId, updatePolicy])

  const handleSaveSchema = React.useCallback(async () => {
    const customVariables: AuthVariableInput[] = authVariables
      .filter((item) => !item.isBuiltin)
      .filter((item) => item.name.trim() !== '')
      .map((item) => ({
        name: item.name.trim(),
        source: item.source.trim(),
        type: item.type,
      }))

    await updateAuthSchema(customVariables)
  }, [authVariables, updateAuthSchema])

  return (
    <main className="size-full overflow-y-auto overflow-x-hidden bg-background">
      <div className="mx-auto w-full max-w-[1600px] px-6 pb-12 pt-10 xl:px-10">
        <section className="mb-10 flex flex-col gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-semibold tracking-tight">策略</h1>
            <p className="text-sm text-muted-foreground">管理各模型的行级安全（RLS）策略</p>
          </div>

          <div className="flex flex-col gap-2 lg:flex-row lg:items-center">
            <Popover open={databaseSelectorOpen} onOpenChange={setDatabaseSelectorOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={databaseSelectorOpen}
                  className="h-[26px] w-full justify-between px-2 text-xs lg:w-[180px]"
                  disabled={databaseOptions.length === 0}
                >
                  <span className="truncate text-muted-foreground">
                    {databasesLoading
                      ? '数据库：加载中...'
                      : databaseOptions.length === 0
                        ? '没有数据库'
                        : `数据库：${selectedDatabase || '未选择'}`}
                  </span>
                  <ChevronsUpDown className="size-3.5 opacity-60" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[180px] p-0" align="start">
                <Command>
                  <CommandInput placeholder="搜索数据库..." className="h-9 text-xs" />
                  <CommandList>
                    <CommandEmpty>没有匹配的数据库</CommandEmpty>
                    {databaseOptions.map((name) => (
                      <CommandItem
                        key={name}
                        value={name}
                        onSelect={() => {
                          if (name === selectedDatabase) {
                            setDatabaseSelectorOpen(false)
                            return
                          }
                          setSelectedDatabase(name)
                          setSelectedModelId('')
                          setDatabaseSelectorOpen(false)
                        }}
                        className="text-xs"
                      >
                        <Check
                          className={cn(
                            'mr-2 size-3.5',
                            selectedDatabase === name ? 'opacity-100' : 'opacity-0'
                          )}
                        />
                        <span className="font-mono">{name}</span>
                      </CommandItem>
                    ))}
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>

            <Popover open={modelSelectorOpen} onOpenChange={setModelSelectorOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={modelSelectorOpen}
                  className="h-[26px] w-full justify-between px-2 text-xs lg:w-[320px]"
                  disabled={!selectedDatabase || modelOptions.length === 0}
                >
                  <span className="truncate text-muted-foreground">
                    {!selectedDatabase
                      ? databaseOptions.length === 0
                        ? '没有模型先创建模型'
                        : '请先选择数据库'
                      : modelOptions.length === 0 && !modelsLoading
                        ? '没有模型先创建模型'
                      : modelsLoading
                        ? '选择模型策略：加载中...'
                        : `选择模型策略：${selectedModel?.name || queryModelName || '未选择'}`}
                  </span>
                  <ChevronsUpDown className="size-3.5 opacity-60" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[320px] p-0" align="start">
                <Command>
                  <CommandInput placeholder="搜索模型..." className="h-9 text-xs" />
                  <CommandList>
                    <CommandEmpty>没有匹配的模型</CommandEmpty>
                    {modelOptions.map((modelOption) => (
                      <CommandItem
                        key={modelOption.id}
                        value={`${modelOption.name}-${modelOption.id}`}
                        onSelect={() => {
                          setSelectedModelId(modelOption.id)
                          setModelSelectorOpen(false)
                        }}
                        className="text-xs"
                      >
                        <Check
                          className={cn(
                            'mr-2 size-3.5',
                            selectedModelId === modelOption.id ? 'opacity-100' : 'opacity-0'
                          )}
                        />
                        <span className="font-mono">{modelOption.name}</span>
                      </CommandItem>
                    ))}
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>
        </section>

        <section className="space-y-4 pb-8">
          <Card>
            <CardHeader className="space-y-3">
              <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <div className="flex items-center gap-3">
                  <Shield className="size-4 text-muted-foreground" />
                  <h3 className="font-mono text-sm font-semibold">{selectedModel?.name || queryModelName || '-'}</h3>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-[26px] text-xs"
                    disabled={!selectedModelId || isRLSEnabled || savingPolicy}
                    onClick={handleEnableRLS}
                  >
                    {savingPolicy ? '处理中...' : '启用 RLS'}
                  </Button>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="outline" size="sm" className="h-[26px] text-xs" disabled={!selectedModelId}>
                        创建策略
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-56">
                      {POLICY_PRESET_OPTIONS.map((option) => (
                        <DropdownMenuItem
                          key={option.label}
                          className="text-xs"
                          onClick={() => handleCreatePolicy(option.preset)}
                        >
                          {option.label}
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>

                  <Tabs
                    value={policyMatchMode}
                    onValueChange={(value) => setPolicyMatchMode(value as 'ALL' | 'ANY')}
                    className="w-[180px]"
                  >
                    <TabsList className="grid h-[26px] w-full grid-cols-2">
                      <TabsTrigger value="ALL" className="text-[11px]">全量匹配</TabsTrigger>
                      <TabsTrigger value="ANY" className="text-[11px]">任一匹配</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>
              </div>
            </CardHeader>

            <CardContent className="pt-0">
              {policyError && (
                <div className="mb-3 rounded-md border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs text-destructive">
                  策略加载失败：{policyError.message}
                </div>
              )}

              {!selectedModelId ? (
                <p className="py-4 text-sm text-muted-foreground">请先选择模型</p>
              ) : policyLoading ? (
                <p className="py-4 text-sm text-muted-foreground">策略加载中...</p>
              ) : !isRLSEnabled ? (
                <div className="rounded-md border border-dashed px-4 py-6 text-sm">
                  <p className="text-foreground">当前模型尚未启用 RLS。</p>
                  <p className="mt-1 text-muted-foreground">启用后将创建默认策略：仅所有者可读写。</p>
                  <div className="mt-4">
                    <Button size="sm" onClick={handleEnableRLS} disabled={savingPolicy || !selectedModelId}>
                      {savingPolicy ? '启用中...' : '启用 RLS'}
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="overflow-x-auto rounded-md border">
                  <table className="w-full table-fixed text-sm">
                    <thead className="bg-muted/40 text-xs text-muted-foreground">
                      <tr>
                        <th className="h-9 px-4 text-left font-medium">策略名称</th>
                        <th className="h-9 px-4 text-left font-medium">命令</th>
                        <th className="h-9 px-4 text-left font-medium">应用角色</th>
                        <th className="h-9 px-4 text-right font-medium">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {policyRows.map((row) => (
                        <tr key={`${selectedModelId}-${row.command}`} className="border-t">
                          <td className="px-4 py-3 text-sm">{row.name}</td>
                          <td className="px-4 py-3">
                            <code className="text-xs text-muted-foreground">{row.command}</code>
                          </td>
                          <td className="px-4 py-3">
                            <code className="text-xs text-muted-foreground">{row.appliedRoles}</code>
                          </td>
                          <td className="px-4 py-3 text-right">
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-7 px-2 text-xs"
                              onClick={() => {
                                setActivePolicyCommand(row.command)
                                setEditorTab(COMMAND_TO_TAB_KEY[row.command])
                                setPolicyDetailOpen(true)
                              }}
                            >
                              {selectedPreset ? '查看' : '编辑'}
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </section>

        <section className="pb-8">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <UserRoundCog className="size-4" />
                认证变量配置
              </CardTitle>
              <CardDescription>配置 JWT 变量并用于 RLS 表达式</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {authError && (
                <div className="rounded-md border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs text-destructive">
                  认证变量加载失败：{authError.message}
                </div>
              )}

              {authLoading ? (
                <p className="py-2 text-sm text-muted-foreground">认证变量加载中...</p>
              ) : (
                <AuthVariableEditor
                  variables={authVariables}
                  onChange={(items: AuthVariable[]) => setAuthVariables(items)}
                  disabled={savingSchema}
                />
              )}

              <div className="rounded-md border px-3 py-2 text-xs text-muted-foreground">
                当前匹配模式：{POLICY_MATCH_MODE_LABELS[policyMatchMode]}
              </div>
              <div className="rounded-md border px-3 py-2 text-xs text-muted-foreground">
                最近更新时间：{policy?.updatedAt || '-'}
              </div>
              <div className="flex justify-end">
                <Button variant="outline" onClick={handleSaveSchema} disabled={savingSchema || authLoading}>
                  {savingSchema ? '保存中...' : '保存变量'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>

      <Sheet open={policyDetailOpen} onOpenChange={setPolicyDetailOpen}>
        <SheetContent side="right" className="w-[760px] overflow-y-auto p-0 sm:max-w-[860px]">
          <div className="p-6">
            <SheetHeader>
              <SheetTitle>
                {activePolicyCommand ? `策略详情 · ${activePolicyCommand}` : '策略详情'}
              </SheetTitle>
              <SheetDescription>
                数据库: <span className="font-mono">{selectedDatabase}</span> / 模型:{' '}
                <span className="font-mono">{selectedModel?.name || '-'}</span>
              </SheetDescription>
            </SheetHeader>

            <div className="mt-6 space-y-6">
              <div className="flex items-center gap-2">
                <Badge variant="outline">实时数据</Badge>
                <Badge variant="secondary">RLS 预览</Badge>
              </div>

              {!isCustomMode && selectedPreset && (
                <div className="rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-xs text-amber-900">
                  当前为预设策略（{PRESET_LABELS[selectedPreset]}），仅可查看。请通过“创建策略”切换到“自定义策略”后再编辑。
                </div>
              )}

              <Separator />

              <Tabs value={editorTab} onValueChange={(value) => setEditorTab(value as keyof ExprState)} className="w-full">
                <TabsList className="grid w-full grid-cols-5">
                  <TabsTrigger value="selectPredicate">查询</TabsTrigger>
                  <TabsTrigger value="insertCheck">插入</TabsTrigger>
                  <TabsTrigger value="updatePredicate">更新</TabsTrigger>
                  <TabsTrigger value="updateCheck">校验</TabsTrigger>
                  <TabsTrigger value="deletePredicate">删除</TabsTrigger>
                </TabsList>

                {(
                  ['selectPredicate', 'insertCheck', 'updatePredicate', 'updateCheck', 'deletePredicate'] as const
                ).map((key) => (
                  <TabsContent key={key} value={key} className="space-y-3">
                    <Label className="font-mono text-xs text-muted-foreground">{key}</Label>
                    <Input
                      value={expressions[key]}
                      disabled={!isCustomMode || savingPolicy}
                      onChange={(event) => {
                        setSelectedPreset(null)
                        setExpressions((prev) => ({
                          ...prev,
                          [key]: event.target.value,
                        }))
                      }}
                    />
                    <PolicyJSONPreview value={parsed[key]} />
                  </TabsContent>
                ))}
              </Tabs>

              <div className="flex justify-end">
                <Button onClick={handleSavePolicy} disabled={savingPolicy || !selectedModelId || !isCustomMode}>
                  {savingPolicy ? '保存中...' : '保存策略'}
                </Button>
              </div>
            </div>
          </div>
        </SheetContent>
      </Sheet>
    </main>
  )
}
