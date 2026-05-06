'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { ChevronDown, Database, Loader2, Search, Table2 } from 'lucide-react'
import { toast } from 'sonner'
import { Input } from '@web/components/ui/input'
import { Button } from '@web/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@web/components/ui/alert-dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { getEndUserToken } from '@api-client/end-user/public'
import { createEndUserScopedClient } from '@api-client/apollo/clients'
import {
  MODEL_CATALOG_END_USER,
  MODEL_DATABASE_CATALOG_END_USER,
} from '@/api-client/model/graphql-docs.end-user.catalog'
import { useEndUser } from '@web/hooks/end-user-auth/useRequireEndUserAuth'
import { EndUserRecordWorkspace } from '@web/components/features/end-user-data'
import { DataWorkspacePanel } from '@web/components/shared/data-workspace/DataWorkspacePanel'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

type DataModel = {
  id: string
  name: string
  title?: string | null
  databaseName: string
}

type CatalogError = {
  __typename: string
  message?: string
}

type DatabaseCatalogQueryResult = {
  modelDatabaseCatalog: {
    data?: {
      databases: Array<{ name: string }>
      totalCount: number
      page: number
      pageSize: number
    } | null
    error?: CatalogError | null
  }
}

type ModelCatalogQueryResult = {
  models: {
    edges: Array<{ node: DataModel }>
    totalCount: number
  }
}

const MAX_MODEL_TABS = 8

export default function EndUserDataPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const router = useRouter()
  const orgName = params.orgName
  const projectSlug = params.projectSlug
  const { user, logout } = useEndUser()

  const [selectedDatabase, setSelectedDatabase] = useState('')
  const [modelFilter, setModelFilter] = useState('')
  const [openedTabs, setOpenedTabs] = useState<DataModel[]>([])
  const [activeModelId, setActiveModelId] = useState('')
  const [databases, setDatabases] = useState<string[]>([])
  const [databasesLoading, setDatabasesLoading] = useState(false)
  const [models, setModels] = useState<DataModel[]>([])
  const [modelsLoading, setModelsLoading] = useState(false)
  const [privateDbInitDialogOpen, setPrivateDbInitDialogOpen] = useState(false)
  const [initPrivateDbLoading, setInitPrivateDbLoading] = useState(false)

  // 从 sessionStorage 读取可访问项目列表（登录时已缓存）
  const [accessibleProjects, setAccessibleProjects] = useState<EndUserAccessibleProject[]>([])

  useEffect(() => {
    if (!orgName) return
    const raw = sessionStorage.getItem(`eu_accessible_projects_${orgName}`)
    if (raw) {
      try {
        setAccessibleProjects(JSON.parse(raw) as EndUserAccessibleProject[])
      } catch {
        setAccessibleProjects([])
      }
    }
  }, [orgName])

  useEffect(() => {
    if (!orgName || !projectSlug) return

    const selectedProject = sessionStorage.getItem(`eu_selected_project_${orgName}`)
    if (selectedProject && selectedProject !== projectSlug) {
      router.replace(`/end-user/${orgName}/projects/${selectedProject}/data`)
      return
    }

    sessionStorage.setItem(`eu_selected_project_${orgName}`, projectSlug)
  }, [orgName, projectSlug, router])

  const loadDatabaseCatalog = async (): Promise<void> => {
    const accessToken = getEndUserToken()
    if (!accessToken || !orgName || !projectSlug) return

    setDatabasesLoading(true)
    try {
      const client = createEndUserScopedClient(orgName, projectSlug, accessToken)
      const { data } = await client.query<DatabaseCatalogQueryResult>({
        query: MODEL_DATABASE_CATALOG_END_USER,
        variables: { input: { page: 1, pageSize: 100 } },
        fetchPolicy: 'network-only',
      })

      const payload = data?.modelDatabaseCatalog
      if (payload?.error) {
        const { __typename: errType, message: errMsg } = payload.error
        if (errType === 'PRIVATE_DB_NOT_INITIALIZED') {
          setDatabases([])
          setSelectedDatabase('')
          setModels([])
          setPrivateDbInitDialogOpen(true)
          return
        }
        throw new Error(errMsg ?? '加载数据库目录失败')
      }

      setDatabases(
        (payload?.data?.databases ?? [])
          .map((item) => item.name)
          .filter((name): name is string => name.length > 0)
      )
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载数据库目录失败'
      toast.error(message)
      setDatabases([])
    } finally {
      setDatabasesLoading(false)
    }
  }

  useEffect(() => {
    void loadDatabaseCatalog()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orgName, projectSlug, user?.id])

  useEffect(() => {
    if (!selectedDatabase && databases.length > 0) {
      setSelectedDatabase(databases[0])
    }
  }, [databases, selectedDatabase])

  useEffect(() => {
    let cancelled = false

    const loadModelCatalog = async () => {
      if (!selectedDatabase) {
        setModels([])
        return
      }

      const accessToken = getEndUserToken()
      if (!accessToken || !orgName || !projectSlug) return

      setModelsLoading(true)
      try {
        const client = createEndUserScopedClient(orgName, projectSlug, accessToken)
        const { data } = await client.query<ModelCatalogQueryResult>({
          query: MODEL_CATALOG_END_USER,
          variables: { input: { databaseName: selectedDatabase, limit: 200 } },
          fetchPolicy: 'network-only',
        })

        const payload = data?.models
        if (cancelled) return

        setModels(
          (payload?.edges ?? [])
            .map((edge) => edge.node)
            .filter(
              (model) => Boolean(model?.id && model?.name && model?.databaseName)
            )
        )
      } catch (err) {
        if (cancelled) return
        const message = err instanceof Error ? err.message : '加载模型目录失败'
        toast.error(message)
        setModels([])
      } finally {
        if (!cancelled) setModelsLoading(false)
      }
    }

    void loadModelCatalog()
    return () => {
      cancelled = true
    }
  }, [selectedDatabase, orgName, projectSlug, user?.id])

  const filteredModels = useMemo(() => {
    const keyword = modelFilter.trim().toLowerCase()
    if (!keyword) return models

    return models.filter((model) =>
      (model.title ?? '').toLowerCase().includes(keyword) || model.name.toLowerCase().includes(keyword)
    )
  }, [models, modelFilter])

  const handleInitPrivateDB = async () => {
    const accessToken = getEndUserToken()
    if (!accessToken || !orgName || !projectSlug) return

    setInitPrivateDbLoading(true)
    try {
      const res = await fetch(`/internal/end-user/data/init-private-db`, {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      })

      type InitResp = { success?: boolean; error?: { code?: string; message?: string } }
      const respData = (await res.json()) as InitResp

      if (!res.ok || !respData.success) {
        throw new Error(respData.error?.message ?? '初始化私有库失败')
      }

      setPrivateDbInitDialogOpen(false)
      toast.success('私有库初始化成功')
      await loadDatabaseCatalog()
    } catch (err) {
      const message = err instanceof Error ? err.message : '初始化私有库失败'
      toast.error(message)
    } finally {
      setInitPrivateDbLoading(false)
    }
  }

  const handleOpenModelTab = (model: DataModel) => {
    setOpenedTabs((prev) => {
      if (prev.some((item) => item.id === model.id)) return prev
      if (prev.length >= MAX_MODEL_TABS) {
        toast.warning(`最多可同时打开 ${MAX_MODEL_TABS} 个模型标签`)
        return prev
      }
      return [...prev, model]
    })
    setActiveModelId(model.id)
  }

  const handleCloseTab = (modelId: string) => {
    setOpenedTabs((prev) => {
      const next = prev.filter((item) => item.id !== modelId)
      if (activeModelId === modelId) {
        setActiveModelId(next[next.length - 1]?.id ?? '')
      }
      return next
    })
  }

  return (
    <main className="min-h-screen bg-[#f6f7fb]">
      <div className="border-b border-border bg-background px-6 py-4">
        <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-foreground">数据管理</h1>
            {accessibleProjects.length > 1 ? (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="mt-1 flex items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground">
                    <span>{orgName} / {projectSlug}</span>
                    <ChevronDown className="size-3.5" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="min-w-[200px]">
                  <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
                    切换项目
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {accessibleProjects.map((p) => (
                    <DropdownMenuItem
                      key={p.slug}
                      onSelect={() => {
                        if (p.slug !== projectSlug) {
                          sessionStorage.setItem(`eu_selected_project_${orgName}`, p.slug)
                          router.push(`/end-user/${orgName}/projects/${p.slug}/data`)
                        }
                      }}
                      className={p.slug === projectSlug ? 'bg-muted font-medium' : ''}
                    >
                      <span className="flex-1">{p.title || p.slug}</span>
                      {p.slug === projectSlug && (
                        <span className="ml-2 text-xs text-muted-foreground">当前</span>
                      )}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            ) : (
              <p className="mt-1 text-sm text-muted-foreground">
                {orgName} / {projectSlug}
              </p>
            )}
          </div>
          <div className="flex items-center gap-3">
            <span className="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
              当前用户：{user?.username || user?.id || '已登录'}
            </span>
            <Button variant="outline" size="sm" onClick={() => void logout()}>
              退出登录
            </Button>
          </div>
        </div>
      </div>

      <div className="mx-auto flex w-full max-w-[1440px] gap-4 p-4">
        <aside className="w-[280px] rounded-xl border border-border bg-background p-3 shadow-sm">
          <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-foreground">
            <Database className="size-4" />
            数据库
          </div>
          <Select value={selectedDatabase} onValueChange={setSelectedDatabase}>
            <SelectTrigger className="h-9 text-sm">
              <SelectValue placeholder={databasesLoading ? '加载数据库中...' : '选择数据库'} />
            </SelectTrigger>
            <SelectContent>
              {databases.map((db) => (
                <SelectItem key={db} value={db} className="font-mono text-xs">
                  {db}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="my-4 border-t border-border" />

          <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-foreground">
            <Table2 className="size-4" />
            模型
          </div>
          <div className="relative mb-2">
            <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
            <Input
              value={modelFilter}
              onChange={(e) => setModelFilter(e.target.value)}
              placeholder="筛选模型"
              className="h-9 pl-8"
            />
          </div>

          {modelsLoading ? (
            <div className="flex items-center gap-2 rounded-md border border-dashed border-border px-2.5 py-3 text-xs text-muted-foreground">
              <Loader2 className="size-3.5 animate-spin" />
              加载模型中...
            </div>
          ) : (
            <div className="space-y-1">
              {filteredModels.map((model) => {
                const isActive = activeModelId === model.id
                return (
                  <button
                    key={model.id}
                    onClick={() => handleOpenModelTab(model)}
                    className={`w-full rounded-md border px-2.5 py-2 text-left transition-colors ${
                      isActive
                        ? 'border-blue-200 bg-blue-50'
                        : 'border-border bg-background hover:bg-muted/50'
                    }`}
                  >
                    <div className="text-sm font-medium text-foreground">{model.title || model.name}</div>
                    <div className="text-xs text-muted-foreground">{model.name}</div>
                  </button>
                )
              })}
              {filteredModels.length === 0 && (
                <div className="rounded-md border border-dashed border-border px-2.5 py-3 text-xs text-muted-foreground">
                  没有匹配的模型
                </div>
              )}
            </div>
          )}
        </aside>

        <DataWorkspacePanel
          tabs={openedTabs}
          activeTabId={activeModelId}
          onTabChange={setActiveModelId}
          onTabClose={handleCloseTab}
          renderContent={(activeTab) => (
            <EndUserRecordWorkspace
              key={activeTab.id}
              modelId={activeTab.id}
              projectSlug={projectSlug}
              orgName={orgName}
            />
          )}
        />
      </div>

      <AlertDialog open={privateDbInitDialogOpen} onOpenChange={setPrivateDbInitDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>私有库未初始化</AlertDialogTitle>
            <AlertDialogDescription>
              检测到私有库 mc_private_{projectSlug} 不存在。请确认是否立即初始化后继续访问数据。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={initPrivateDbLoading}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void handleInitPrivateDB()} disabled={initPrivateDbLoading}>
              {initPrivateDbLoading ? '初始化中...' : '确认初始化'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </main>
  )
}
