'use client'

import { useState, useEffect, useMemo } from 'react'
import { useQuery } from '@apollo/client'
import { ChevronLeft, ChevronRight, Loader2, Search, Table2, X } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from '@web/components/ui/sheet'
import { Input } from '@web/components/ui/input'
import { Button } from '@web/components/ui/button'
import { toast } from 'sonner'
import { LIST_TABLES } from '@/api-client/project'
import { useModelSyncJob } from '@web/hooks/model/use-sync-models-from-db'
import { useStartModelSync } from '@web/hooks/model/use-model-sync'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'

const PAGE_SIZE = 20

interface ImportModelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectSlug: string
  databaseName: string
  databaseId: string | null
  onSuccess: () => void
}

interface TableInfo {
  name: string
}

interface ListTablesQueryData {
  listTables?: {
    items: TableInfo[]
    totalCount: number
  } | null
}

export function ImportModelDialog({
  open,
  onOpenChange,
  projectSlug,
  databaseName,
  databaseId,
  onSuccess,
}: ImportModelDialogProps) {
  // listTables lives on the project endpoint: /graphql/org/{orgName}/project/{projectSlug}/
  const projectClient = useProjectScopedClient(projectSlug)

  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [jobId, setJobId] = useState<string | null>(null)

  const offset = (currentPage - 1) * PAGE_SIZE

  const { data, loading: tablesLoading } = useQuery<ListTablesQueryData>(LIST_TABLES, {
    client: projectClient,
    variables: {
      input: {
        databaseName,
        excludeExisting: true,
        limit: PAGE_SIZE,
        offset,
      },
    },
    skip: !open || !projectSlug || !databaseName,
    fetchPolicy: 'network-only',
  })

  const { startSync, loading: syncing } = useStartModelSync(projectSlug)
  const { data: jobData, loading: jobLoading } = useModelSyncJob(jobId, projectSlug)
  const job = jobData?.modelSyncJobs?.[0]

  useEffect(() => {
    if (!open) {
      setSearchQuery('')
      setSelectedTable(null)
      setCurrentPage(1)
      setJobId(null)
    }
  }, [open])

  // Watch async job status
  useEffect(() => {
    if (!job) return
    const status = job.status
    if (status === 'SUCCEEDED') {
      toast.success('模型导入成功')
      setJobId(null)
      onSuccess()
      onOpenChange(false)
    } else if (status === 'FAILED') {
      toast.error('导入失败，请重试')
      setJobId(null)
    }
  }, [job, onSuccess, onOpenChange])

  const isJobRunning = job?.status === 'PENDING' || job?.status === 'RUNNING'
  const isImporting = syncing || (!!jobId && (jobLoading || isJobRunning))

  useEffect(() => {
    setCurrentPage(1)
  }, [searchQuery])

  const tables: TableInfo[] = useMemo(() => {
    if (!data?.listTables?.items) return []
    return data.listTables.items as TableInfo[]
  }, [data])

  const totalCount: number = data?.listTables?.totalCount ?? 0
  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE))

  const filteredTables = useMemo(() => {
    if (!searchQuery) return tables
    const q = searchQuery.toLowerCase()
    return tables.filter((t) => t.name.toLowerCase().includes(q))
  }, [tables, searchQuery])

  const handleImport = async () => {
    if (!selectedTable || !databaseId) return

    try {
      const result = await startSync([{
        databaseId,
        tableNames: [selectedTable],
      }])
      if (result && result.jobs.length > 0) {
        setJobId(result.jobs[0].jobId)
      } else {
        toast.error('导入失败，请重试')
      }
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : '导入失败，请重试'
      toast.error(message)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-[400px] flex-col sm:max-w-[400px]">
        <SheetHeader>
          <SheetTitle className="text-base">导入模型</SheetTitle>
          <SheetDescription className="text-sm">
            从数据库 <span className="font-mono text-blue-600">{databaseName}</span> 选择一张表导入为模型
          </SheetDescription>
        </SheetHeader>

        <div className="flex flex-1 flex-col gap-3 overflow-hidden py-4">
          {/* Search input */}
          <div className="relative">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
            <Input
              type="text"
              placeholder="搜索表名..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-8 px-8 text-sm"
            />
            {searchQuery && (
              <button
                type="button"
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                onClick={() => setSearchQuery('')}
              >
                <X className="size-3.5" strokeWidth={1.5} />
              </button>
            )}
          </div>

          {/* Table list */}
          <div className="flex-1 overflow-y-auto rounded-md border border-border bg-muted/10">
            {tablesLoading ? (
              <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-2 text-muted-foreground">
                <Loader2 className="size-5 animate-spin" strokeWidth={1.5} />
                <span className="text-sm">加载表列表...</span>
              </div>
            ) : filteredTables.length === 0 ? (
              <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-2 text-muted-foreground">
                <Table2 className="size-8 opacity-30" strokeWidth={1.5} />
                <p className="text-sm">
                  {totalCount === 0 ? '所有表已导入' : '未找到匹配的表'}
                </p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {filteredTables.map((table) => (
                  <button
                    key={table.name}
                    type="button"
                    className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-[#dadee5]"
                    style={
                      selectedTable === table.name
                        ? { backgroundColor: '#dadee5' }
                        : undefined
                    }
                    onClick={() => setSelectedTable(table.name)}
                  >
                    <Table2 className="size-3.5 shrink-0 text-muted-foreground" strokeWidth={1.5} />
                    <span className="font-mono text-sm text-foreground">{table.name}</span>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Pagination */}
          {!searchQuery && totalPages > 1 && (
            <div className="flex items-center justify-between px-1">
              <span className="text-xs text-muted-foreground">
                第 {currentPage} / {totalPages} 页，共 {totalCount} 张表
              </span>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  className="size-7 p-0"
                  onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                  disabled={currentPage <= 1 || tablesLoading}
                  aria-label="上一页"
                >
                  <ChevronLeft className="size-3.5" strokeWidth={1.5} />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="size-7 p-0"
                  onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                  disabled={currentPage >= totalPages || tablesLoading}
                  aria-label="下一页"
                >
                  <ChevronRight className="size-3.5" strokeWidth={1.5} />
                </Button>
              </div>
            </div>
          )}
        </div>

        <SheetFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onOpenChange(false)}
            disabled={isImporting}
          >
            取消
          </Button>
          <Button
            size="sm"
            className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
            onClick={handleImport}
            disabled={!selectedTable || isImporting || !databaseId}
          >
            {isImporting && <Loader2 className="mr-1.5 size-3.5 animate-spin" strokeWidth={1.5} />}
            {isImporting ? '导入中...' : '导入'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
