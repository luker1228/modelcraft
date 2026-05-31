'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import {
  ChevronsUpDown,
  Database,
  Loader2,
  Search,
  Table2,
  X,
} from 'lucide-react'
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
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import { cn } from '@/shared/utils'
import { getEndUserToken } from '@api-client/end-user/public'
import { createEndUserScopedClient } from '@api-client/apollo/clients'
import {
  MODEL_CATALOG_END_USER,
  MODEL_DATABASE_CATALOG_END_USER,
} from '@/api-client/model/graphql-docs.end-user.catalog'
import { useEndUser } from '@web/hooks/end-user-auth/useRequireEndUserAuth'
import { EndUserRecordWorkspace } from '@web/components/features/end-user-data'
import { DataWorkspacePanel } from '@web/components/shared/data-workspace/DataWorkspacePanel'
import { EndUserAppLayout } from '@web/components/features/layout/EndUserAppLayout'

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
    items: Array<DataModel>
    hasNextPage: boolean
  }
}

const MAX_MODEL_TABS = 8

export default function EndUserDataPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const orgName = params.orgName
  const projectSlug = params.projectSlug
  const { user } = useEndUser()

  const [selectedDatabase, setSelectedDatabase] = useState('')
  const [databaseOpen, setDatabaseOpen] = useState(false)
  const [modelFilter, setModelFilter] = useState('')
  const [openedTabs, setOpenedTabs] = useState<DataModel[]>([])
  const [activeModelId, setActiveModelId] = useState('')
  const [databases, setDatabases] = useState<string[]>([])
  const [databasesLoading, setDatabasesLoading] = useState(false)
  const [models, setModels] = useState<DataModel[]>([])
  const [modelsLoading, setModelsLoading] = useState(false)
  const [privateDbInitDialogOpen, setPrivateDbInitDialogOpen] = useState(false)
  const [initPrivateDbLoading, setInitPrivateDbLoading] = useState(false)

  const loadDatabaseCatalog = async (): Promise<void> => {
    if (!orgName || !projectSlug) return

    setDatabasesLoading(true)
    try {
      const client = createEndUserScopedClient(orgName, projectSlug)
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
      if (!orgName || !projectSlug) return

      setModelsLoading(true)
      try {
        const client = createEndUserScopedClient(orgName, projectSlug)
        const { data } = await client.query<ModelCatalogQueryResult>({
          query: MODEL_CATALOG_END_USER,
          variables: { input: { databaseName: selectedDatabase, pageSize: 200 } },
          fetchPolicy: 'network-only',
        })

        const payload = data?.models
        if (cancelled) return
        setModels(
          (payload?.items ?? []).filter(
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
    return () => { cancelled = true }
  }, [selectedDatabase, orgName, projectSlug, user?.id])

  const filteredModels = useMemo(() => {
    const keyword = modelFilter.trim().toLowerCase()
    if (!keyword) return models
    return models.filter((model) =>
      (model.title ?? '').toLowerCase().includes(keyword) ||
      model.name.toLowerCase().includes(keyword)
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
        headers: { Authorization: `Bearer ${accessToken}` },
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
    <EndUserAppLayout orgName={orgName} activePage="projects">
      <div className="flex h-full overflow-hidden bg-background">

        {/* Model sidebar — database selector + model list */}
        <aside className="flex w-[240px] flex-shrink-0 flex-col overflow-hidden border-r border-border bg-sidebar">

          {/* Zone 1: Database selector */}
          <div className="p-3">
            <Popover open={databaseOpen} onOpenChange={setDatabaseOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className={cn(
                    'h-9 w-full justify-between px-3 text-sm font-medium transition-colors',
                    selectedDatabase
                      ? 'border-primary/40 bg-primary/5 text-foreground hover:border-primary/60 hover:bg-primary/10'
                      : 'border-dashed border-muted-foreground/40 bg-background text-muted-foreground hover:border-primary/40 hover:bg-primary/5 hover:text-foreground'
                  )}
                  disabled={databasesLoading}
                >
                  <span className="flex min-w-0 items-center gap-2">
                    <Database className={cn('size-3.5 shrink-0', selectedDatabase ? 'text-primary' : 'text-muted-foreground')} />
                    {databasesLoading ? (
                      <>
                        <Loader2 className="size-3 shrink-0 animate-spin" />
                        <span className="text-muted-foreground">加载中...</span>
                      </>
                    ) : selectedDatabase ? (
                      <span className="truncate font-medium text-foreground">{selectedDatabase}</span>
                    ) : (
                      <span>选择数据库</span>
                    )}
                  </span>
                  <ChevronsUpDown className="ml-2 size-3.5 shrink-0 text-muted-foreground" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[208px] border border-border p-1 shadow-lg" align="start">
                {databases.length === 0 ? (
                  <div className="px-2.5 py-3 text-center text-sm text-muted-foreground">
                    暂无数据库
                  </div>
                ) : (
                  databases.map((db) => (
                    <button
                      key={db}
                      type="button"
                      className={cn(
                        'w-full cursor-pointer rounded-sm px-2.5 py-1.5 text-left text-sm transition-colors',
                        selectedDatabase === db
                          ? 'bg-accent text-foreground'
                          : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                      )}
                      onClick={() => {
                        setSelectedDatabase(db)
                        setDatabaseOpen(false)
                        setOpenedTabs([])
                        setActiveModelId('')
                      }}
                    >
                      {db}
                    </button>
                  ))
                )}
              </PopoverContent>
            </Popover>
          </div>

          {/* Divider */}
          <div className="border-t border-border" />

          {/* Zone 2: Model list */}
          <div className="flex flex-1 flex-col overflow-hidden">
            {/* Search */}
            <div className="px-2 pb-2 pt-2.5">
              <div className="relative">
                <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  type="text"
                  placeholder="查询模型..."
                  value={modelFilter}
                  onChange={(e) => setModelFilter(e.target.value)}
                  className="h-7 bg-foreground/[.026] px-8 text-xs"
                />
                {modelFilter && (
                  <button
                    type="button"
                    className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                    onClick={() => setModelFilter('')}
                  >
                    <X className="size-3.5" />
                  </button>
                )}
              </div>
            </div>

            {/* Model items */}
            <nav className="min-h-0 flex-1 overflow-y-auto px-2 pb-4">
              <div className="space-y-0.5">
                {modelsLoading && (
                  <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                    <Loader2 className="mb-3 size-6 animate-spin" />
                    <p className="text-sm">加载模型中...</p>
                  </div>
                )}

                {!modelsLoading && filteredModels.map((model) => (
                  <div
                    key={model.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => handleOpenModelTab(model)}
                    onKeyDown={(e) => e.key === 'Enter' && handleOpenModelTab(model)}
                    className={cn(
                      'group flex h-8 cursor-pointer select-none items-center gap-1.5 rounded-md border-l-[3px] pl-2 pr-1 transition-colors',
                      activeModelId === model.id
                        ? 'border-l-primary bg-primary/[0.08] text-primary'
                        : 'border-l-transparent text-muted-foreground hover:bg-accent/60 hover:text-foreground'
                    )}
                  >
                    <Table2 className={cn(
                      'size-[15px] shrink-0 transition-colors',
                      activeModelId === model.id ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'
                    )} />
                    <span className="min-w-0 flex-1 truncate text-xs">{model.name}</span>
                    {model.title && model.title !== model.name && (
                      <span className="max-w-[56px] shrink-0 truncate text-xs text-muted-foreground/60" title={model.title}>
                        {model.title}
                      </span>
                    )}
                  </div>
                ))}

                {!modelsLoading && filteredModels.length === 0 && selectedDatabase && (
                  <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                    <Table2 className="mb-3 size-10 opacity-20" />
                    <p className="text-sm">暂无模型</p>
                  </div>
                )}

                {!selectedDatabase && !databasesLoading && (
                  <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                    <Database className="mb-3 size-8 opacity-20" />
                    <p className="text-sm">请先选择数据库</p>
                  </div>
                )}
              </div>
            </nav>
          </div>
        </aside>

        {/* Workspace panel */}
        <div className="flex min-w-0 flex-1 flex-col bg-background p-4">
          <DataWorkspacePanel
            tabs={openedTabs}
            activeTabId={activeModelId}
            onTabChange={setActiveModelId}
            onTabClose={handleCloseTab}
            emptyText="从左侧选择模型以打开数据表"
            className="h-full min-h-0"
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
    </EndUserAppLayout>
  )
}
