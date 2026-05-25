'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'

interface CliGuidePageProps {
  params: { orgName: string }
}

interface RefreshResponse {
  accessToken?: string
  expiresAt?: string
}

interface StepCardProps {
  step: string
  title: string
  description: string
  command: string
}

const CLI_STEPS: StepCardProps[] = [
  {
    step: '01',
    title: '下载并安装 CLI',
    description:
      '从 GitHub Release 下载 macOS arm64 预编译二进制并安装到 /usr/local/bin。当前仅支持 macOS Apple Silicon。',
    command: [
      'curl -fL "https://github.com/patientCat/modelcraft/releases/latest/download/mc-darwin-arm64" -o mc',
      'chmod +x mc',
      'sudo mv mc /usr/local/bin/mc',
      'mc version',
    ].join('\n'),
  },
  {
    step: '02',
    title: '登录获取本地凭证',
    description: '登录后会写入凭证文件（默认 ~/.config/modelcraft/credentials.json）。',
    command:
      "mc auth login \\\n  --server 'https://<gateway-host>' \\\n  --org '<org-slug>' \\\n  --username '<username>' \\\n  --password '<password>'",
  },
  {
    step: '03',
    title: '选择项目上下文',
    description: '先查看登录状态，再切换默认项目，后续 catalog/run 可省略 --project。',
    command: 'mc auth status\nmc auth switch-project <project-slug>',
  },
  {
    step: '04',
    title: '发现可用资源',
    description: '先看项目，再看数据库和模型，确认目标路径。',
    command:
      'mc catalog projects\nmc catalog databases --project <project-slug>\nmc catalog models --project <project-slug> --database <database-name>',
  },
  {
    step: '05',
    title: '查询模型数据',
    description: '使用 describe 查看字段，再用 run 发送 GraphQL 查询。资源路径格式：project.database.model。',
    command:
      "mc describe <project>.<database>.<model>\nmc run <project>.<database>.<model> '{ findMany(take: 5) { id } }'",
  },
]

function StepCard({ step, title, description, command }: StepCardProps) {
  return (
    <section className="rounded-lg border bg-background p-5 sm:p-6">
      <div className="flex items-start gap-4">
        <span className="inline-flex h-7 min-w-7 items-center justify-center rounded-md border bg-muted px-2 text-xs font-semibold text-foreground">
          {step}
        </span>
        <div className="min-w-0 flex-1">
          <h3 className="text-base font-semibold text-foreground">{title}</h3>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{description}</p>
          <pre className="mt-4 overflow-x-auto rounded-md border bg-muted p-4 font-mono text-xs leading-5 text-foreground">
            {command}
          </pre>
        </div>
      </div>
    </section>
  )
}

function useEndUserTokenReady(orgName: string): boolean {
  const setAccessToken = useEndUserAuthStore((s) => s.setAccessToken)
  const router = useRouter()

  const [ready, setReady] = useState(() => {
    const storeState = useEndUserAuthStore.getState()
    if (storeState.accessToken && !storeState.isTokenExpired()) return true

    if (typeof window !== 'undefined') {
      const savedToken = sessionStorage.getItem(`eu_token_${orgName}`)
      const savedExpiresAt = Number(sessionStorage.getItem(`eu_token_expires_at_${orgName}`) ?? '0')
      if (savedToken && Date.now() < savedExpiresAt - 5 * 60 * 1000) {
        const expiresIn = Math.floor((savedExpiresAt - Date.now()) / 1000)
        useEndUserAuthStore.getState().setAccessToken(savedToken, expiresIn)
        return true
      }
    }

    return false
  })

  useEffect(() => {
    const storeState = useEndUserAuthStore.getState()
    if (storeState.accessToken && !storeState.isTokenExpired()) {
      setReady(true)
      return
    }

    void (async () => {
      try {
        const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/refresh`, {
          method: 'POST',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ orgName }),
        })
        if (!res.ok) {
          router.replace(`/end-user/${orgName}/login`)
          return
        }

        const data = (await res.json()) as RefreshResponse
        if (!data.accessToken) {
          router.replace(`/end-user/${orgName}/login`)
          return
        }

        let expiresIn = 3600
        if (data.expiresAt) {
          const ms = new Date(data.expiresAt).getTime() - Date.now()
          if (ms > 0) expiresIn = Math.floor(ms / 1000)
        }

        setAccessToken(data.accessToken, expiresIn)
        setReady(true)
      } catch {
        router.replace(`/end-user/${orgName}/login`)
      }
    })()
  }, [orgName, router, setAccessToken])

  return ready
}

export default function CliGuidePage({ params }: CliGuidePageProps) {
  const { orgName } = params
  const router = useRouter()
  const ready = useEndUserTokenReady(orgName)

  const handleLogout = async () => {
    if (!orgName) return
    await fetch(`/api/bff/org/${orgName}/end-user/auth/logout`, {
      method: 'POST',
      credentials: 'same-origin',
    })
    useEndUserAuthStore.getState().clearSession()
    router.push(`/end-user/${orgName}/login`)
  }

  if (!ready) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-muted-foreground">加载中...</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b bg-background px-6">
        <span className="text-base font-semibold text-foreground">{orgName}</span>
        <button onClick={() => void handleLogout()} className="text-sm text-destructive hover:underline">
          登出
        </button>
      </header>

      <nav className="flex border-b bg-background px-6">
        <button
          onClick={() => router.push(`/end-user/${orgName}/workspace`)}
          className="border-b-2 border-transparent px-4 py-3 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          Projects
        </button>
        <button
          onClick={() => router.push(`/end-user/${orgName}/workspace/cli`)}
          className="border-b-2 border-primary px-4 py-3 text-sm font-medium text-primary"
        >
          CLI 使用
        </button>
      </nav>

      <main className="mx-auto w-full max-w-5xl flex-1 px-6 py-8">
        <div className="space-y-6">
          <section className="rounded-lg border bg-background p-6">
            <p className="text-xs font-medium tracking-[0.08em] text-muted-foreground">CLI 快速上手</p>
            <h2 className="mt-2 text-xl font-semibold text-foreground">ModelCraft CLI 从下载到使用</h2>
            <p className="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">
              下面这份流程覆盖从安装、登录到查询数据的最小闭环。按顺序执行可快速完成首次可用配置。
            </p>
            <div className="mt-4 flex flex-wrap gap-2 text-xs text-muted-foreground">
              <span className="rounded-md border bg-muted px-2 py-1">当前支持 macOS arm64</span>
              <span className="rounded-md border bg-muted px-2 py-1">默认凭证路径 ~/.config/modelcraft/credentials.json</span>
            </div>
          </section>

          <section aria-labelledby="cli-steps" className="space-y-3">
            <h3 id="cli-steps" className="px-1 text-sm font-medium text-muted-foreground">
              操作步骤
            </h3>
            <div className="space-y-4">
              {CLI_STEPS.map((step) => (
                <StepCard
                  key={step.step}
                  step={step.step}
                  title={step.title}
                  description={step.description}
                  command={step.command}
                />
              ))}
            </div>
          </section>

          <section className="rounded-lg border bg-background p-6">
            <h3 className="text-base font-semibold text-foreground">常见问题排查</h3>
            <p className="mt-2 text-sm text-muted-foreground">遇到安装或执行异常时，可优先按以下顺序检查。</p>
            <dl className="mt-4 divide-y rounded-md border bg-muted/20">
              <div className="space-y-2 p-4">
                <dt className="text-sm font-semibold text-foreground">404 Not Found</dt>
                <dd className="text-sm leading-6 text-muted-foreground">
                  访问{' '}
                  <a
                    href="https://github.com/patientCat/modelcraft/releases"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary underline"
                  >
                    Releases 页面
                  </a>{' '}
                  确认 latest 版本下存在 `mc-darwin-arm64` 资产。
                </dd>
              </div>
              <div className="space-y-2 p-4">
                <dt className="text-sm font-semibold text-foreground">Permission denied</dt>
                <dd className="text-sm leading-6 text-muted-foreground">
                  无 sudo 权限时可安装到用户目录：
                  <code className="mt-2 block rounded border bg-muted px-3 py-2 font-mono text-xs text-foreground">
                    mkdir -p ~/.local/bin &amp;&amp; mv mc ~/.local/bin/mc
                  </code>
                  并将 <code className="rounded bg-muted px-1 font-mono text-xs">~/.local/bin</code> 加入 PATH。
                </dd>
              </div>
              <div className="space-y-2 p-4">
                <dt className="text-sm font-semibold text-foreground">command not found: mc</dt>
                <dd className="text-sm leading-6 text-muted-foreground">
                  确认 <code className="rounded bg-muted px-1 font-mono text-xs">/usr/local/bin</code> 在 PATH 中。运行{' '}
                  <code className="rounded bg-muted px-1 font-mono text-xs">echo $PATH</code> 检查，或将二进制移到其他已在
                  PATH 中的目录。
                </dd>
              </div>
              <div className="space-y-2 p-4">
                <dt className="text-sm font-semibold text-foreground">架构不匹配 (exec format error)</dt>
                <dd className="text-sm leading-6 text-muted-foreground">
                  当前仅提供 macOS arm64 版本（Apple Silicon）。请使用 M 系列芯片设备安装。
                </dd>
              </div>
              <div className="space-y-2 p-4">
                <dt className="text-sm font-semibold text-foreground">网络问题 (Connection refused / timeout)</dt>
                <dd className="text-sm leading-6 text-muted-foreground">
                  如在公司内网，可能需要配置代理。尝试设置{' '}
                  <code className="rounded bg-muted px-1 font-mono text-xs">https_proxy</code> 环境变量后重试。
                </dd>
              </div>
            </dl>
          </section>
        </div>
      </main>
    </div>
  )
}
