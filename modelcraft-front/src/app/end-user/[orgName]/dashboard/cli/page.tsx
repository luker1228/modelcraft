'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { EndUserAppLayout } from '@web/components/features/layout'
import { cn } from '@/shared/utils'

interface CliGuidePageProps {
  params: { orgName: string }
}

interface RefreshResponse {
  accessToken?: string
  expiresAt?: string
}

const CLI_STEPS = [
  {
    n: '01',
    id: 'install',
    title: '下载并安装 CLI',
    description: '下载对应平台的预编译二进制，安装到 PATH 中的任意目录。',
    tabs: [
      {
        label: 'macOS (Apple Silicon)',
        command: [
          'curl -fL "https://github.com/patientCat/modelcraft/releases/latest/download/mc-darwin-arm64" -o mc',
          'chmod +x mc',
          'sudo mv mc /usr/local/bin/mc',
          'mc version',
        ].join('\n'),
      },
      {
        label: 'Linux (x86_64)',
        command: [
          'curl -fL "https://github.com/patientCat/modelcraft/releases/latest/download/mc-linux-amd64" -o mc',
          'chmod +x mc',
          'sudo mv mc /usr/local/bin/mc',
          'mc version',
        ].join('\n'),
      },
    ],
  },
  {
    n: '02',
    id: 'login',
    title: '创建 PAT 并登录',
    description: '先在控制台创建 Personal Access Token，再用 --token 完成登录。凭证写入 ~/.config/modelcraft/credentials.json。',
    command: [
      '# 使用 PAT 登录（默认服务器）',
      "mc auth login --token '<your-pat-token>'",
      '',
      '# 自托管部署时额外指定 --server',
      '# mc auth login \\',
      "#   --server 'https://<gateway-host>' \\",
      "#   --token '<your-pat-token>'",
    ].join('\n'),
  },
  {
    n: '03',
    id: 'context',
    title: '选择项目上下文',
    description: '查看当前登录状态，切换默认项目后续命令可省略 --project 参数。',
    command: 'mc auth status\nmc auth switch-project <project-slug>',
  },
  {
    n: '04',
    id: 'discover',
    title: '发现可用资源',
    description: '列出项目、数据库、模型，确认目标路径。',
    command:
      'mc catalog projects\nmc catalog databases --project <project-slug>\nmc catalog models --project <project-slug> --database <database-name>',
  },
  {
    n: '05',
    id: 'query',
    title: '查询模型数据',
    description: '用 describe 查看字段结构，再用 run 发送 GraphQL 查询。资源路径格式：project.database.model。',
    command:
      "mc describe <project>.<database>.<model>\nmc run <project>.<database>.<model> '{ findMany(take: 5) { id } }'",
  },
]

const FAQ_ITEMS = [
  {
    q: '404 Not Found',
    a: (
      <>
        访问{' '}
        <a
          href="https://github.com/patientCat/modelcraft/releases"
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary hover:underline"
        >
          Releases 页面
        </a>{' '}
        确认 latest 版本下存在对应平台资产（mc-darwin-arm64、mc-linux-amd64）。
      </>
    ),
  },
  {
    q: 'Permission denied',
    a: (
      <>
        无 sudo 权限时可安装到用户目录：
        <code className="mx-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
          mkdir -p ~/.local/bin {'&&'} mv mc ~/.local/bin/mc
        </code>
        并将 ~/.local/bin 加入 PATH。
      </>
    ),
  },
  {
    q: 'command not found: mc',
    a: (
      <>
        确认 /usr/local/bin 在 PATH 中。运行{' '}
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">echo $PATH</code> 检查，
        或将二进制移到已在 PATH 中的目录。
      </>
    ),
  },
  {
    q: 'exec format error',
    a: (
      <>
        平台不匹配。macOS Apple Silicon 使用{' '}
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">mc-darwin-arm64</code>，
        Linux x86_64 使用{' '}
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">mc-linux-amd64</code>。
      </>
    ),
  },
  {
    q: 'Connection refused / timeout',
    a: (
      <>
        公司内网可能需要配置代理。尝试设置{' '}
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">https_proxy</code>{' '}
        环境变量后重试。
      </>
    ),
  },
]

const TOC_SECTIONS = [
  { id: 'install', label: '安装' },
  { id: 'login', label: '登录' },
  { id: 'context', label: '选择上下文' },
  { id: 'discover', label: '发现资源' },
  { id: 'query', label: '查询数据' },
  { id: 'faq', label: '常见问题' },
]

function DarkCode({ code }: { code: string }) {
  return (
    <div className="mt-3 overflow-hidden rounded-lg border border-[#2a3050]">
      <div className="border-b border-[#2a3050] bg-[#1a1f36] px-4 py-2">
        <span className="text-[10px] font-medium uppercase tracking-wider text-[#697386]">BASH</span>
      </div>
      <pre className="overflow-x-auto bg-[#1e2330] p-4 font-mono text-[12.5px] leading-6 text-[#cdd6f4]">
        {code}
      </pre>
    </div>
  )
}

function StepCodeTabs({ tabs }: { tabs: { label: string; command: string }[] }) {
  const [active, setActive] = useState(0)
  return (
    <div className="mt-3 overflow-hidden rounded-lg border border-[#2a3050]">
      {/* Tab bar */}
      <div className="flex border-b border-[#2a3050] bg-[#1a1f36]">
        {tabs.map((t, i) => (
          <button
            key={t.label}
            type="button"
            onClick={() => setActive(i)}
            className={cn(
              '-mb-px px-4 py-2 text-xs transition-colors',
              active === i
                ? 'border-b border-[#6366f1] font-medium text-white'
                : 'text-[#697386] hover:text-[#9ca3af]'
            )}
          >
            {t.label}
          </button>
        ))}
      </div>
      <pre className="overflow-x-auto bg-[#1e2330] p-4 font-mono text-[12.5px] leading-6 text-[#cdd6f4]">
        {tabs[active]?.command}
      </pre>
    </div>
  )
}

function CliToc() {
  const [active, setActive] = useState('install')

  const scrollTo = (id: string) => {
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    setActive(id)
  }

  return (
    <nav className="hidden w-52 shrink-0 border-r border-border lg:block">
      <div className="sticky top-0 p-5">
        <p className="mb-3 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          本页内容
        </p>
        <ul className="space-y-0.5">
          {TOC_SECTIONS.map((s) => (
            <li key={s.id}>
              <button
                type="button"
                onClick={() => scrollTo(s.id)}
                className={cn(
                  'w-full rounded px-2.5 py-1.5 text-left text-sm transition-colors',
                  active === s.id
                    ? 'bg-primary/[0.06] font-medium text-primary'
                    : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
                )}
              >
                {s.label}
              </button>
            </li>
          ))}
        </ul>
      </div>
    </nav>
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
        if (!res.ok) { router.replace(`/end-user/${orgName}/login`); return }
        const data = (await res.json()) as RefreshResponse
        if (!data.accessToken) { router.replace(`/end-user/${orgName}/login`); return }
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
  const ready = useEndUserTokenReady(orgName)

  if (!ready) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="size-5 animate-spin rounded-full border-2 border-border border-t-foreground" />
      </div>
    )
  }

  return (
    <EndUserAppLayout orgName={orgName} activePage="cli">
      <div className="flex h-full overflow-hidden">

        {/* ── Sticky TOC ──────────────────────────────────────────────────── */}
        <CliToc />

        {/* ── Scrollable body ─────────────────────────────────────────────── */}
        <div className="flex-1 overflow-y-auto">
        <div className="px-10 py-8">

          {/* Page header */}
          <div className="mb-10">
            <p className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              CLI 快速上手
            </p>
            <h1 className="text-xl font-semibold text-foreground">ModelCraft CLI</h1>
            <p className="mt-2 max-w-lg text-sm leading-6 text-muted-foreground">
              从安装到查询数据的最小闭环，按顺序执行完成首次配置。
            </p>
          </div>

          {/* Timeline steps */}
          <section className="mb-14">
            {CLI_STEPS.map((step, i) => (
              <div key={step.n} id={step.id} className="flex scroll-mt-6 gap-5">
                {/* Left: step indicator + connector line */}
                <div className="flex shrink-0 flex-col items-center">
                  <div className="flex size-8 items-center justify-center rounded-full border-2 border-border bg-card text-[11px] font-semibold tabular-nums text-muted-foreground">
                    {step.n}
                  </div>
                  {i < CLI_STEPS.length - 1 && (
                    <div className="mt-1 w-px flex-1 bg-border" />
                  )}
                </div>

                {/* Right: content */}
                <div className={cn('min-w-0 flex-1', i < CLI_STEPS.length - 1 ? 'pb-10' : 'pb-0')}>
                  <h3 className="mt-1 text-sm font-semibold text-foreground">{step.title}</h3>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">{step.description}</p>
                  {'tabs' in step ? (
                    <StepCodeTabs tabs={step.tabs!} />
                  ) : (
                    <DarkCode code={step.command} />
                  )}
                </div>
              </div>
            ))}
          </section>

          {/* FAQ */}
          <section id="faq" className="scroll-mt-6 pb-16">
            <h2 className="mb-5 text-base font-semibold text-foreground">常见问题</h2>
            <dl className="space-y-0 divide-y divide-border rounded-lg border">
              {FAQ_ITEMS.map((item) => (
                <div key={item.q} className="grid grid-cols-[180px_1fr] gap-4 px-5 py-4">
                  <dt className="text-sm font-medium text-foreground">{item.q}</dt>
                  <dd className="text-sm leading-6 text-muted-foreground">{item.a}</dd>
                </div>
              ))}
            </dl>
          </section>

        </div>
        </div>
      </div>
    </EndUserAppLayout>
  )
}
