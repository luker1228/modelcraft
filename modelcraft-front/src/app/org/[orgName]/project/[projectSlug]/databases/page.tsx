'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { Plus, Pencil, MoreHorizontal, Trash2 } from 'lucide-react'
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
import { PageLayout, PageHeader } from '@web/components/features/layout'
import {
  useModelDatabases,
  useUnregisterModelDatabase,
  type ModelDatabase,
} from '@web/hooks/model-database/use-model-databases'
import { RegisterDatabaseDialog } from './_components/RegisterDatabaseDialog'
import { EditDatabaseSheet } from './_components/EditDatabaseSheet'

export default function DatabasesPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { databases, loading } = useModelDatabases(params.projectSlug)
  const { unregister } = useUnregisterModelDatabase(params.projectSlug)

  const [registerOpen, setRegisterOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<ModelDatabase | null>(null)

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
                        {db.mode === 'SELF_HOSTED' ? (
                          <Badge
                            variant="outline"
                            className="border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400"
                          >
                            自建
                          </Badge>
                        ) : (
                          <Badge
                            variant="outline"
                            className="border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
                          >
                            托管
                          </Badge>
                        )}
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
                              <DropdownMenuItem
                                className="text-destructive focus:text-destructive"
                                onClick={() => unregister(db.id)}
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
    </PageLayout>
  )
}
