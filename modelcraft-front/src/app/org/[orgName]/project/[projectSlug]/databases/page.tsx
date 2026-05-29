'use client'

import { useEffect, useRef, useState } from 'react'
import { useParams } from 'next/navigation'
import { Plus, Pencil, MoreHorizontal, RefreshCw, Trash2 } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
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
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@web/components/ui/dialog'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import {
  useModelDatabases,
  useUnregisterModelDatabase,
  useStartModelDatabaseSync,
  useFetchModelDatabaseSyncJob,
  type ModelDatabase,
  type ModelDatabaseSyncJob,
} from '@web/hooks/model-database/use-model-databases'
import { RegisterDatabaseDialog } from './_components/RegisterDatabaseDialog'
import { EditDatabaseSheet } from './_components/EditDatabaseSheet'
import { toast } from 'sonner'

function isTerminalJobStatus(status: ModelDatabaseSyncJob['status']) {
  return status === 'SUCCEEDED' || status === 'PARTIAL_SUCCESS' || status === 'FAILED'
}

export default function DatabasesPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { databases, loading } = useModelDatabases(params.projectSlug)
  const { unregister } = useUnregisterModelDatabase(params.projectSlug)
  const { startSync, loading: syncStarting } = useStartModelDatabaseSync(params.projectSlug)
  const { fetchJob } = useFetchModelDatabaseSyncJob(params.projectSlug)

  const [registerOpen, setRegisterOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<ModelDatabase | null>(null)
  const [unregisterTarget, setUnregisterTarget] = useState<ModelDatabase | null>(null)
  const [syncTarget, setSyncTarget] = useState<ModelDatabase | null>(null)
  const [resultTarget, setResultTarget] = useState<ModelDatabase | null>(null)
  const [jobsByDatabaseId, setJobsByDatabaseId] = useState<Record<string, ModelDatabaseSyncJob>>({})
  const pollersRef = useRef<Record<string, ReturnType<typeof setInterval>>>({})

  useEffect(() => {
    return () => {
      Object.values(pollersRef.current).forEach(clearInterval)
      pollersRef.current = {}
    }
  }, [])

  const upsertJob = (job: ModelDatabaseSyncJob) => {
    setJobsByDatabaseId((prev) => ({ ...prev, [job.databaseId]: job }))
  }

  const stopPolling = (databaseId: string) => {
    const timer = pollersRef.current[databaseId]
    if (timer) {
      clearInterval(timer)
      delete pollersRef.current[databaseId]
    }
  }

  const startPolling = (databaseId: string, jobId: string) => {
    stopPolling(databaseId)
    pollersRef.current[databaseId] = setInterval(() => {
      void fetchJob(jobId).then((job) => {
        if (!job) return
        upsertJob(job)
        if (isTerminalJobStatus(job.status)) {
          stopPolling(databaseId)
        }
      }).catch(() => {
        stopPolling(databaseId)
      })
    }, 2000)
  }

  const handleStartSync = async () => {
    if (!syncTarget) return

    try {
      const job = await startSync(syncTarget.id)
      if (!job) {
        toast.error('创建同步任务失败')
        return
      }
      upsertJob(job)
      startPolling(syncTarget.id, job.id)
      setResultTarget(syncTarget)
      setSyncTarget(null)
      toast.success('同步任务已创建')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : '同步任务创建失败')
    }
  }

  const renderJobStatus = (job?: ModelDatabaseSyncJob) => {
    if (!job) return null

    switch (job.status) {
      case 'PENDING':
        return <span className="text-xs text-muted-foreground">排队中</span>
      case 'RUNNING':
        return (
          <span className="text-xs text-muted-foreground">
            同步中 {job.processedTables}/{job.totalTables || '?'}
          </span>
        )
      case 'SUCCEEDED':
        return <span className="text-xs text-green-700 dark:text-green-400">同步完成</span>
      case 'PARTIAL_SUCCESS':
        return <span className="text-xs text-amber-700 dark:text-amber-400">部分成功</span>
      case 'FAILED':
        return <span className="text-xs text-destructive">同步失败</span>
      default:
        return null
    }
  }

  const selectedJob = resultTarget ? jobsByDatabaseId[resultTarget.id] : null

  return (
    <PageLayout maxWidth="7xl" padding="default">
      <PageHeader
        title="数据库管理"
        spacing="compact"
        actions={
          <Button size="sm" onClick={() => setRegisterOpen(true)} className="gap-1.5">
            <Plus className="size-4" strokeWidth={1.5} />
            接管数据库
          </Button>
        }
      />

      <p className="mb-5 text-sm text-muted-foreground">
        接管此项目使用的 MySQL 数据库，设置访问模式
      </p>

      <div className="overflow-hidden rounded-lg border border-border bg-card">
        {loading ? (
          <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
            加载中…
          </div>
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead>模式</TableHead>
                  <TableHead className="w-16" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {databases.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      className="py-12 text-center text-sm text-muted-foreground"
                    >
                      暂无已接管的数据库，点击右上角"接管数据库"开始
                    </TableCell>
                  </TableRow>
                ) : (
                  databases.map((db) => (
                    <TableRow key={db.id}>
                      <TableCell>
                        <div className="flex flex-col">
                          <span className="font-medium">{db.title}</span>
                          {db.title !== db.name && (
                            <span className="text-xs text-muted-foreground">{db.name}</span>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
                        {db.description || '—'}
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-col gap-1">
                          {db.mode === 'SELF_HOSTED' ? (
                            <Badge
                              variant="outline"
                              className="w-fit border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400"
                            >
                              自建
                            </Badge>
                          ) : (
                            <Badge
                              variant="outline"
                              className="w-fit border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
                            >
                              托管
                            </Badge>
                          )}
                          {renderJobStatus(jobsByDatabaseId[db.id])}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7"
                            onClick={() => setEditTarget(db)}
                          >
                            <Pencil className="size-3.5" />
                          </Button>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="size-7">
                                <MoreHorizontal className="size-3.5" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => setSyncTarget(db)}>
                                <RefreshCw className="mr-2 size-3.5" />
                                同步数据库
                              </DropdownMenuItem>
                              {jobsByDatabaseId[db.id] && (
                                <DropdownMenuItem onClick={() => setResultTarget(db)}>
                                  查看同步结果
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuItem
                                className="text-destructive focus:text-destructive"
                                onClick={() => setUnregisterTarget(db)}
                              >
                                <Trash2 className="mr-2 size-3.5" />
                                取消接管
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </>
        )}
      </div>

      <RegisterDatabaseDialog open={registerOpen} onOpenChange={setRegisterOpen} />
      <EditDatabaseSheet database={editTarget} onClose={() => setEditTarget(null)} />

      <AlertDialog open={syncTarget !== null} onOpenChange={(open) => { if (!open) setSyncTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>同步数据库为模型</AlertDialogTitle>
            <AlertDialogDescription>
              将扫描 {syncTarget?.title} 中的全部表。新表会导入为模型，已有模型会同步 schema，新建模型将统一放入“数据库导入”分组。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={syncStarting}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void handleStartSync()} disabled={syncStarting}>
              {syncStarting ? '创建中…' : '开始同步'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={unregisterTarget !== null}
        onOpenChange={(open) => { if (!open) setUnregisterTarget(null) }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确定取消接管？</AlertDialogTitle>
            <AlertDialogDescription>
              确定取消接管 {unregisterTarget?.title} 吗？已关联的模型将无法通过此入口访问数据库。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => {
                if (unregisterTarget) {
                  void unregister(unregisterTarget.id)
                  setUnregisterTarget(null)
                }
              }}
            >
              确认取消接管
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Dialog open={!!resultTarget} onOpenChange={(open) => { if (!open) setResultTarget(null) }}>
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>同步结果</DialogTitle>
            <DialogDescription>
              {resultTarget?.title} 的最近一次页面内同步任务结果
            </DialogDescription>
          </DialogHeader>
          {selectedJob ? (
            <div className="space-y-4 text-sm">
              <div className="grid grid-cols-2 gap-3 rounded-md border border-border p-4 sm:grid-cols-3">
                <div>
                  <p className="text-xs text-muted-foreground">状态</p>
                  <p className="font-medium">{selectedJob.status}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">扫描表数</p>
                  <p className="font-medium">{selectedJob.totalTables}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">已处理</p>
                  <p className="font-medium">{selectedJob.processedTables}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">新建模型</p>
                  <p className="font-medium">{selectedJob.createdModels}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">同步模型</p>
                  <p className="font-medium">{selectedJob.syncedModels}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">失败表数</p>
                  <p className="font-medium">{selectedJob.failedCount}</p>
                </div>
              </div>

              <div className="rounded-md border border-border">
                <div className="border-b border-border px-4 py-3 text-sm font-medium">
                  失败明细
                </div>
                {selectedJob.failedTables.length === 0 ? (
                  <div className="px-4 py-6 text-sm text-muted-foreground">无失败表</div>
                ) : (
                  <div className="divide-y divide-border">
                    {selectedJob.failedTables.map((item) => (
                      <div key={`${item.tableName}-${item.message}`} className="px-4 py-3">
                        <p className="font-medium">{item.tableName}</p>
                        <p className="mt-1 text-xs text-muted-foreground">{item.message}</p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ) : (
            <div className="py-8 text-sm text-muted-foreground">暂无同步结果</div>
          )}
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}
