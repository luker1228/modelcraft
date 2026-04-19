'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import { Database, Loader2, Search, Table2 } from 'lucide-react'
import { toast } from 'sonner'
import { Input } from '@web/components/ui/input'
import { Button } from '@web/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { getEndUserToken } from '@bff/end-user/public'
import { useEndUser } from '@web/hooks/end-user-auth/useRequireEndUserAuth'
import ModelRecordWorkspace from '@web/components/features/model-editor/ModelRecordWorkspace'
import { DataWorkspacePanel } from '@web/components/features/model-editor/DataWorkspacePanel'

type DataModel = {
  id: string
  name: string
  title?: string | null
  databaseName: string
}

const MAX_MODEL_TABS = 8

export default function EndUserDataPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
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

  useEffect(() => {
    let cancelled = false

    const loadDatabaseCatalog = async () => {
      const accessToken = getEndUserToken()
      if (!accessToken || !orgName || !projectSlug) return

      setDatabasesLoading(true)
      try {
        const res = await fetch('/api/bff/end-user/data/database-catalog?page=1&pageSize=100', {
          method: 'GET',
          credentials: 'same-origin',
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        })

        const data = (await res.json()) as {
          databases?: Array<{ name?: string }>
          error?: { message?: string }
        }

        if (!res.ok) {
          throw new Error(data.error?.message || '加载数据库目录失败')
        }

        if (cancelled) return

        setDatabases(
          (data.databases ?? [])
            .map((item) => item?.name ?? '')
            .filter((name): name is string => name.length > 0)
        )
      } catch (err) {
        if (cancelled) return
        const message = err instanceof Error ? err.message : '加载数据库目录失败'
        toast.error(message)
        setDatabases([])
      } finally {
        if (!cancelled) setDatabasesLoading(false)
      }
    }

    void loadDatabaseCatalog()
    return () => {
      cancelled = true
    }
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
        const searchParams = new URLSearchParams({
          databaseName: selectedDatabase,
          page: '1',
          pageSize: '200',
        })
        const res = await fetch(`/api/bff/end-user/data/model-catalog?${searchParams.toString()}`, {
          method: 'GET',
          credentials: 'same-origin',
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        })

        const data = (await res.json()) as {
          models?: Array<DataModel>
          error?: { message?: string }
        }

        if (!res.ok) {
          throw new Error(data.error?.message || '加载模型目录失败')
        }
        if (cancelled) return

        setModels(
          (data.models ?? []).filter(
            (model): model is DataModel => Boolean(model?.id && model?.name && model?.databaseName)
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
            <p className="mt-1 text-sm text-muted-foreground">
              {orgName} / {projectSlug}
            </p>
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
            <ModelRecordWorkspace
              key={activeTab.id}
              modelId={activeTab.id}
              projectSlug={projectSlug}
              orgName={orgName}
              workspaceMode="end_user"
            />
          )}
        />
      </div>
    </main>
  )
}
