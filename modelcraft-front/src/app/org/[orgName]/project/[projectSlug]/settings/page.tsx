'use client'

import { useState, useMemo } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { useParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { Button } from '@web/components/ui/button'
import { Form } from '@web/components/ui/form'
import { GET_CLUSTER } from '@/api-client/cluster'
import { TEST_CLUSTER_CONNECTION } from '@/api-client/cluster'
import { UPDATE_PROJECT_CLUSTER } from '@/api-client/project'
import { cn } from '@/shared/utils'
import { IdentityFormSection } from '@web/components/ui/identity-form-section'
import { DatabaseConfigFields } from '@web/components/features/database/DatabaseConfigFields'
import { PageHeader } from '@web/components/features/layout'
import {
  Database,
  CheckCircle,
  XCircle,
  Loader2,
  Trash2,
} from 'lucide-react'

// Sentinel value returned by the server to represent an unchanged password.
// Send it back as-is to tell the server to keep the existing password.
const ENCRYPTED_BY_SERVER = '<encrypted_by_server>'

interface ConnectionInfo {
  host: string
  port: number
  username: string
  password: string
}

interface Cluster {
  id: string
  title: string
  description: string
  host: string
  port: number
  username: string
  password: string
  status: 'ACTIVE' | 'DISABLED'
  createdAt: string
  updatedAt: string
}

interface FormValues {
  title: string
  description: string
  host: string
  port: number
  username: string
  password: string
}

interface SettingsNavItem {
  id: string
  label: string
  icon: React.ReactNode
  anchor: string
}

interface ClusterQueryData {
  databaseCluster?: {
    cluster?: {
      id: string
      title: string
      description: string
      connectionInfo: ConnectionInfo
      status: 'ACTIVE' | 'DISABLED'
      createdAt: string
      updatedAt: string
    }
  }
}

interface UpdateProjectClusterResult {
  updateProjectCluster?: {
    cluster?: {
      id: string
      title: string
      description: string
      connectionInfo: ConnectionInfo
      status: 'ACTIVE' | 'DISABLED'
      createdAt: string
      updatedAt: string
    }
    error?: {
      message: string
    }
  }
}

interface TestConnectionResult {
  testDatabaseConnection?: {
    success: boolean
    connectionTime?: number
    error?: {
      message: string
    }
  }
}

const settingsNavItems: SettingsNavItem[] = [
  {
    id: 'cluster',
    label: '集群配置',
    icon: <Database className="size-4 shrink-0" strokeWidth={1.5} />,
    anchor: '#cluster',
  },
]

const statusConfig = {
  ACTIVE: {
    label: '已连接',
    dot: 'bg-emerald-600 animate-pulse',
    badge: 'bg-emerald-50 text-emerald-600',
  },
  DISABLED: {
    label: '未连接',
    dot: 'bg-muted-foreground/50',
    badge: 'bg-muted text-muted-foreground',
  },
}

export default function SettingsPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const [activeNav, setActiveNav] = useState('cluster')
  const [cluster, setCluster] = useState<Cluster | null>(null)
  const [testingConnection, setTestingConnection] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)

  const form = useForm<FormValues>({
    defaultValues: {
      title: '',
      description: '',
      host: '',
      port: 3306,
      username: '',
      password: ENCRYPTED_BY_SERVER,
    },
  })

  const { control, reset, watch, formState } = form
  const formData = watch()

  const orgScopedContext = useMemo(() => {
    if (!orgName) return undefined
    return { uri: `/api/bff/graphql/org/${orgName}/` }
  }, [orgName])

  const { loading } = useQuery<ClusterQueryData>(GET_CLUSTER, {
    skip: !orgName || !projectSlug,
    variables: { projectSlug },
    context: orgScopedContext,
    onCompleted: (queryData) => {
      const c = queryData?.databaseCluster?.cluster
      if (c) {
        const clusterData: Cluster = {
          id: c.id,
          title: c.title,
          description: c.description,
          host: c.connectionInfo.host,
          port: c.connectionInfo.port,
          username: c.connectionInfo.username,
          password: c.connectionInfo.password,
          status: c.status,
          createdAt: c.createdAt,
          updatedAt: c.updatedAt,
        }
        setCluster(clusterData)
        reset({
          title: c.title,
          description: c.description,
          host: c.connectionInfo.host,
          port: c.connectionInfo.port,
          username: c.connectionInfo.username,
          password: c.connectionInfo.password,
        })
      } else {
        setCluster(null)
      }
    },
    onError: () => setCluster(null),
  })

  const [testConnection] = useMutation<TestConnectionResult>(TEST_CLUSTER_CONNECTION)
  const [updateProjectCluster] = useMutation<UpdateProjectClusterResult>(UPDATE_PROJECT_CLUSTER, { context: orgScopedContext })

  const handleSaveIdentity = async () => {
    if (!cluster) return
    try {
      const input: Record<string, unknown> = {
        title: formData.title,
        description: formData.description,
        connectionInfo: {
          host: cluster.host,
          port: cluster.port,
          username: cluster.username,
          password: cluster.password,
        },
      }
      const result = await updateProjectCluster({
        variables: { projectSlug, input },
      })
      const updated = result.data?.updateProjectCluster?.cluster
      if (updated) {
        setCluster((prev) => prev ? {
          ...prev,
          title: updated.title,
          description: updated.description,
        } : prev)
        reset({
          ...formData,
          title: updated.title,
          description: updated.description,
        }, { keepDirty: false })
      }
    } catch {
      // keep current state on error
    }
  }

  const handleSaveConnection = async () => {
    if (!cluster || !formData.host) return
    try {
      const input: Record<string, unknown> = {
        title: cluster.title,
        description: cluster.description,
        connectionInfo: {
          host: formData.host,
          port: formData.port,
          username: formData.username,
          password: formData.password,
        },
      }
      const result = await updateProjectCluster({
        variables: { projectSlug, input },
      })
      const updated = result.data?.updateProjectCluster?.cluster
      if (updated) {
        setCluster({
          id: updated.id,
          title: updated.title,
          description: updated.description,
          host: updated.connectionInfo.host,
          port: updated.connectionInfo.port,
          username: updated.connectionInfo.username,
          password: updated.connectionInfo.password,
          status: updated.status,
          createdAt: updated.createdAt,
          updatedAt: updated.updatedAt,
        })
        reset({
          title: cluster.title,
          description: cluster.description,
          host: updated.connectionInfo.host,
          port: updated.connectionInfo.port,
          username: updated.connectionInfo.username,
          password: updated.connectionInfo.password,
        })
      }
    } catch {
      // keep current state on error
    }
    setTestResult(null)
  }

  const handleTestFormConnection = async () => {
    if (!formData.host || !formData.username || !formData.password) return
    setTestingConnection(true)
    setTestResult(null)
    try {
      const result = await testConnection({
        variables: {
          input: {
            projectSlug,
            connectionInfo: {
              host: formData.host,
              port: formData.port,
              username: formData.username,
              password: formData.password,
            },
          },
        },
        context: orgScopedContext,
      })
      const payload = result.data?.testDatabaseConnection
      if (payload?.success) {
        setTestResult({ success: true, message: `连接成功！耗时 ${Number(payload.connectionTime).toFixed(2) ?? 0}ms` })
      } else {
        setTestResult({ success: false, message: payload?.error?.message ?? '连接失败' })
      }
    } catch (e) {
      setTestResult({ success: false, message: e instanceof Error ? e.message : '未知错误' })
    } finally {
      setTestingConnection(false)
    }
  }

  const identityDirty = !!(formState.dirtyFields.title || formState.dirtyFields.description)

  return (
    <div className="flex h-full overflow-hidden">

      {/* ===== Settings Sidebar ===== */}
      <aside className="w-[200px] flex-shrink-0 overflow-y-auto border-r border-border bg-card">
        <div className="p-3">
          <p className="mb-1 px-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            配置
          </p>
          <nav className="flex flex-col gap-0.5">
            {settingsNavItems.map((item) => (
              <button
                key={item.id}
                onClick={() => setActiveNav(item.id)}
                className={cn(
                  'flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm font-medium transition-colors duration-150',
                  activeNav === item.id
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                )}
              >
                {item.icon}
                {item.label}
              </button>
            ))}
          </nav>
        </div>
      </aside>

      {/* ===== Main Content ===== */}
      <div className="flex-1 overflow-y-auto bg-muted">
        <div className="mx-auto max-w-3xl p-8">

          <PageHeader title="项目设置" description="管理您的项目配置和数据库连接" />

          {/* Loading */}
          {loading && (
            <div className="flex items-center justify-center py-20 text-muted-foreground">
              <Loader2 className="mr-2 size-5 animate-spin" />
              <span className="text-sm">加载中...</span>
            </div>
          )}

          {/* No Cluster */}
          {!loading && !cluster && (
            <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-border bg-card py-20">
              <div className="flex size-12 items-center justify-center rounded-lg bg-muted">
                <Database className="size-5 text-muted-foreground" strokeWidth={1.5} />
              </div>
              <div className="space-y-1 text-center">
                <p className="text-sm font-semibold text-foreground">暂无数据库集群</p>
                <p className="text-sm text-muted-foreground">在创建项目时已关联集群，若未显示请刷新页面</p>
              </div>
            </div>
          )}

          {/* Cluster Edit Mode */}
          {!loading && cluster && (
            <div className="space-y-4">

              {/* ── Identity Section ── */}
              <IdentityFormSection
                title="基本信息"
                displayNameField={{
                  name: 'title',
                  label: '显示名称',
                  placeholder: '主数据库集群',
                  control,
                }}
                descriptionField={{
                  name: 'description',
                  label: '描述',
                  control,
                }}
                showActions={true}
                saveDisabled={!identityDirty}
                onSave={handleSaveIdentity}
                onCancel={() => reset({ ...formData, title: cluster.title, description: cluster.description }, { keepDirty: false })}
              />

              {/* ── Connection Config Section ── */}
              <Form {...form}>
                <div className="overflow-hidden rounded-lg border border-border bg-card">
                  {/* Section header */}
                  <div className="flex items-center gap-2 border-b border-border px-5 py-4">
                    <Database className="size-4 text-muted-foreground" strokeWidth={1.5} />
                    <h2 className="text-sm font-semibold text-foreground">连接配置</h2>
                    <span
                      className={cn(
                        'ml-auto inline-flex items-center gap-1.5 rounded px-3 py-1 text-xs font-medium',
                        statusConfig[cluster.status].badge
                      )}
                    >
                      <span className={cn('size-1.5 rounded-full', statusConfig[cluster.status].dot)} />
                      {statusConfig[cluster.status].label}
                    </span>
                  </div>

                  {/* Section body */}
                  <div className="p-5">
                    <DatabaseConfigFields
                      form={form}
                      showPasswordToggle={true}
                      showIcons={false}
                      showRequired={false}
                      encryptedByServerPlaceholder={ENCRYPTED_BY_SERVER}
                    />
                  </div>

                  {/* Section footer */}
                  <div className="flex items-center justify-between border-t border-border bg-muted px-5 py-3">
                    {/* Test result feedback */}
                    <div className="min-h-[20px]">
                      {testResult && (
                        <span
                          className={cn(
                            'flex items-center gap-1.5 text-sm',
                            testResult.success ? 'text-emerald-600' : 'text-destructive'
                          )}
                        >
                          {testResult.success ? (
                            <CheckCircle className="size-4 shrink-0" strokeWidth={1.5} />
                          ) : (
                            <XCircle className="size-4 shrink-0" strokeWidth={1.5} />
                          )}
                          {testResult.message}
                        </span>
                      )}
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleTestFormConnection}
                        disabled={testingConnection || !formData.host || !formData.username}
                        className="h-9 px-4 text-sm font-medium"
                      >
                        {testingConnection && (
                          <Loader2 className="mr-1.5 size-4 animate-spin" strokeWidth={1.5} />
                        )}
                        测试连接
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          if (cluster) {
                            reset({
                              title: formData.title,
                              description: formData.description,
                              host: cluster.host,
                              port: cluster.port,
                              username: cluster.username,
                              password: cluster.password,
                            })
                            setTestResult(null)
                          }
                        }}
                        className="h-9 px-4 text-sm font-medium"
                      >
                        取消
                      </Button>
                      <Button
                        size="sm"
                        onClick={handleSaveConnection}
                        disabled={!formData.host || !testResult?.success}
                        className="h-9 border-0 bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors duration-200 hover:bg-primary/90"
                      >
                        保存
                      </Button>
                    </div>
                  </div>
                </div>
              </Form>

              {/* ── Danger Zone Section ── */}
              <div className="overflow-hidden rounded-lg border border-destructive/30 bg-card">
                <div className="border-b border-red-200 bg-red-50 px-5 py-4">
                  <h2 className="text-sm font-semibold text-red-600">危险区域</h2>
                  <p className="mt-0.5 text-xs text-red-400">以下操作不可撤销，请谨慎操作</p>
                </div>
                <div className="flex items-center justify-between px-5 py-4">
                  <div>
                    <p className="text-sm font-medium text-foreground">删除项目</p>
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      永久删除此项目及所有相关数据，此操作无法撤销
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled
                    className="h-9 shrink-0 cursor-not-allowed border-red-200 px-4 text-sm font-medium text-red-500 hover:border-red-300 hover:bg-red-50 hover:text-red-600"
                  >
                    <Trash2 className="mr-1.5 size-4" strokeWidth={1.5} />
                    删除项目
                  </Button>
                </div>
              </div>

            </div>
          )}

        </div>
      </div>

    </div>
  )
}
