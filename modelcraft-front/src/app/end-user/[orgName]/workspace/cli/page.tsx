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

function StepCard({
  title,
  description,
  command,
}: {
  title: string
  description: string
  command: string
}) {
  return (
    <section className="rounded-lg border bg-background p-5">
      <h3 className="text-base font-semibold text-foreground">{title}</h3>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
      <pre className="mt-4 overflow-x-auto rounded-md border bg-muted p-3 font-mono text-xs text-foreground">
        {command}
      </pre>
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

      <main className="mx-auto w-full max-w-5xl flex-1 space-y-4 p-6">
        <section className="rounded-lg border bg-background p-5">
          <h2 className="text-xl font-semibold text-foreground">ModelCraft CLI 从下载到使用</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            下面这份流程覆盖从安装、登录到查询数据的最小闭环。当前仅支持 macOS Apple Silicon (arm64)。
          </p>
        </section>

        <StepCard
          title="1) 下载并安装 CLI"
          description="从 GitHub Release 下载 macOS arm64 预编译二进制并安装到 /usr/local/bin。当前仅支持 macOS Apple Silicon。"
          command={[
            'export MC_VERSION=cli-v0.1.0',
            'curl -fL "https://github.com/patientCat/modelcraft/releases/download/${MC_VERSION}/mc-darwin-arm64" -o mc',
            'chmod +x mc',
            'sudo mv mc /usr/local/bin/mc',
            'mc version',
          ].join('\n')}
        />

        <StepCard
          title="2) 登录获取本地凭证"
          description="登录后会写入凭证文件（默认 ~/.config/modelcraft/credentials.json）。"
          command={
            "mc auth login \\\n  --server 'https://<gateway-host>' \\\n  --org '<org-slug>' \\\n  --username '<username>' \\\n  --password '<password>'"
          }
        />

        <StepCard
          title="3) 选择项目上下文"
          description="先查看登录状态，再切换默认项目，后续 catalog/run 可省略 --project。"
          command={'mc auth status\nmc auth switch-project <project-slug>'}
        />

        <StepCard
          title="4) 发现可用资源"
          description="先看项目，再看数据库和模型，确认目标路径。"
          command={
            'mc catalog projects\nmc catalog databases --project <project-slug>\nmc catalog models --project <project-slug> --database <database-name>'
          }
        />

        <StepCard
          title="5) 查询模型数据"
          description="使用 describe 查看字段，再用 run 发送 GraphQL 查询。资源路径格式：project.database.model。"
          command={
            "mc describe <project>.<database>.<model>\nmc run <project>.<database>.<model> '{ findMany(take: 5) { id } }'"
          }
        />

        <section className="rounded-lg border bg-background p-5">
          <h3 className="text-base font-semibold text-foreground">常见问题排查</h3>
          <dl className="mt-3 space-y-3 text-sm">
            <div>
              <dt className="font-medium text-foreground">404 Not Found</dt>
              <dd className="mt-1 text-muted-foreground">
                检查版本号是否正确。访问{' '}
                <a
                  href="https://github.com/patientCat/modelcraft/releases"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary underline"
                >
                  Releases 页面
                </a>{' '}
                确认 cli-vX.Y.Z 标签已发布。
              </dd>
            </div>
            <div>
              <dt className="font-medium text-foreground">Permission denied</dt>
              <dd className="mt-1 text-muted-foreground">
                无 sudo 权限时可安装到用户目录：
                <code className="mt-1 block rounded bg-muted px-2 py-1 font-mono text-xs">
                  mkdir -p ~/.local/bin &amp;&amp; mv mc ~/.local/bin/mc
                </code>
                并将 <code className="rounded bg-muted px-1">~/.local/bin</code> 加入 PATH。
              </dd>
            </div>
            <div>
              <dt className="font-medium text-foreground">command not found: mc</dt>
              <dd className="mt-1 text-muted-foreground">
                确认 <code className="rounded bg-muted px-1">/usr/local/bin</code> 在 PATH 中。运行{' '}
                <code className="rounded bg-muted px-1">echo $PATH</code>{' '}
                检查，或将二进制移到其他已在 PATH 中的目录。
              </dd>
            </div>
            <div>
              <dt className="font-medium text-foreground">架构不匹配 (exec format error)</dt>
              <dd className="mt-1 text-muted-foreground">
                当前仅提供 macOS arm64 版本（Apple Silicon）。请使用 M 系列芯片设备安装。
              </dd>
            </div>
            <div>
              <dt className="font-medium text-foreground">网络问题 (Connection refused / timeout)</dt>
              <dd className="mt-1 text-muted-foreground">
                如在公司内网，可能需要配置代理。尝试设置{' '}
                <code className="rounded bg-muted px-1">https_proxy</code> 环境变量后重试。
              </dd>
            </div>
          </dl>
        </section>
      </main>
    </div>
  )
}
