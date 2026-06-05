'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Check, Copy } from 'lucide-react'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { EndUserAppLayout } from '@web/components/features/layout'
import { cn } from '@/shared/utils'

interface ApiDocsPageProps {
  params: { orgName: string }
}

interface RefreshResponse {
  accessToken?: string
  expiresAt?: string
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

// ── Snippet builders ───────────────────────────────────────────────────────

function buildPythonSnippet(orgName: string): string {
  return `import os
import requests

TOKEN = os.environ["MC_API_TOKEN"]

ORG_NAME     = "${orgName}"
PROJECT_SLUG = "your-project"
DB_NAME      = "your-db"
MODEL_NAME   = "your-model"

ENDPOINT = (
    f"http://localhost:8080/end-user/graphql"
    f"/org/{'{ORG_NAME}'}/project/{'{PROJECT_SLUG}'}"
    f"/db/{'{DB_NAME}'}/model/{'{MODEL_NAME}'}"
)

query = """
query {
  list(limit: 10) {
    id
  }
}
"""

resp = requests.post(
    ENDPOINT,
    json={"query": query},
    headers={"Authorization": f"Bearer {'{TOKEN}'}"},
)
resp.raise_for_status()
print(resp.json())`
}

function buildCurlSnippet(orgName: string): string {
  const host = 'http://localhost:8080'
  // 端点路径为 Runtime API 合法路径，非 BFF 代理路径
  const path = `/end-user/graphql/org/${orgName}/project/$PROJECT_SLUG/db/$DB_NAME/model/$MODEL_NAME`
  return `ORG_NAME="${orgName}"
PROJECT_SLUG="your-project"
DB_NAME="your-db"
MODEL_NAME="your-model"
TOKEN="your-token"

ENDPOINT="${host}${path}"

curl -s -X POST "$ENDPOINT" \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"query":"{ list(limit: 10) { id } }"}' | jq .`
}

// ── Constants ──────────────────────────────────────────────────────────────

const TOC = [
  { id: 'quickstart', label: '快速开始' },
  { id: 'endpoint', label: '端点' },
  { id: 'auth', label: '认证' },
  { id: 'examples', label: '代码示例' },
  { id: 'reference', label: '查询参考' },
]

type TabId = 'python' | 'curl'
const TABS: { id: TabId; label: string; lang: string }[] = [
  { id: 'python', label: 'Python', lang: 'PYTHON' },
  { id: 'curl', label: 'curl', lang: 'BASH' },
]

const OPERATIONS = [
  {
    type: 'QUERY' as const,
    name: '列表查询',
    code: `query {
  list(limit: 20, offset: 0) {
    id
    # 其他字段...
  }
}`,
  },
  {
    type: 'QUERY' as const,
    name: '按 ID 查询',
    code: `query {
  get(id: "record-id") {
    id
    # 其他字段...
  }
}`,
  },
  {
    type: 'MUTATION' as const,
    name: '创建记录',
    code: `mutation {
  create(input: {
    field: "value"
  }) {
    id
  }
}`,
  },
  {
    type: 'MUTATION' as const,
    name: '更新记录',
    code: `mutation {
  update(
    id: "record-id",
    input: { field: "new-value" }
  ) {
    id
  }
}`,
  },
]

// ── Sub-components ─────────────────────────────────────────────────────────

function DarkCodeBlock({ code, lang }: { code: string; lang: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = () => {
    void navigator.clipboard.writeText(code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }
  return (
    <div className="overflow-hidden rounded-lg border border-[#2a3050]">
      <div className="flex items-center justify-between border-b border-[#2a3050] bg-[#1a1f36] px-4 py-2">
        <span className="text-[10px] font-medium uppercase tracking-wider text-[#697386]">{lang}</span>
        <button
          type="button"
          onClick={handleCopy}
          className="flex items-center gap-1.5 text-[11px] text-[#697386] transition-colors hover:text-white"
        >
          {copied ? <Check className="size-3 text-emerald-400" /> : <Copy className="size-3" />}
          {copied ? '已复制' : '复制'}
        </button>
      </div>
      <pre className="max-h-80 overflow-auto bg-[#1e2330] p-4 font-mono text-[12.5px] leading-6 text-[#cdd6f4]">
        {code}
      </pre>
    </div>
  )
}

function Divider() {
  return <div className="my-10 border-t border-border" />
}

// ── Main content ───────────────────────────────────────────────────────────

function ApiDocsContent({ orgName }: { orgName: string }) {
  const [activeTab, setActiveTab] = useState<TabId>('python')
  const [activeSection, setActiveSection] = useState('quickstart')

  const snippet = activeTab === 'python' ? buildPythonSnippet(orgName) : buildCurlSnippet(orgName)
  const activeLang = TABS.find((t) => t.id === activeTab)?.lang ?? 'CODE'

  const scrollTo = (id: string) => {
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    setActiveSection(id)
  }

  // 端点路径展示文本，此处为文档字符串，非 API 调用
  const endpointDisplay = '/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}'

  return (
    <div className="flex h-full overflow-hidden">

      {/* ── Sticky TOC ──────────────────────────────────────────────────── */}
      <nav className="hidden w-52 shrink-0 border-r border-border lg:block">
        <div className="sticky top-0 p-5">
          <p className="mb-3 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            本页内容
          </p>
          <ul className="space-y-0.5">
            {TOC.map((s) => (
              <li key={s.id}>
                <button
                  type="button"
                  onClick={() => scrollTo(s.id)}
                  className={cn(
                    'w-full rounded px-2.5 py-1.5 text-left text-sm transition-colors',
                    activeSection === s.id
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

      {/* ── Scrollable body ─────────────────────────────────────────────── */}
      <div className="flex-1 overflow-y-auto">
        <div className="px-10 py-8">

          {/* Page header */}
          <div className="mb-10">
            <p className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              API 文档
            </p>
            <h1 className="text-xl font-semibold text-foreground">Runtime GraphQL API</h1>
            <p className="mt-2 max-w-lg text-sm leading-6 text-muted-foreground">
              通过 API Token 直接调用 Runtime GraphQL 端点，对任意模型执行数据查询与变更。
              每个模型对应独立端点。
            </p>
          </div>

          {/* ── 快速开始 ─────────────────────────────────────────────────── */}
          <section id="quickstart" className="scroll-mt-6">
            <h2 className="text-base font-semibold text-foreground">快速开始</h2>
            <div className="mt-5 flex items-start gap-0">
              {[
                {
                  n: 1,
                  title: '创建 Token',
                  desc: '前往 API Token 管理页面，生成 Personal Access Token 并复制明文。',
                },
                {
                  n: 2,
                  title: '定位端点',
                  desc: '按 org / project / db / model 路径拼出专属端点 URL。',
                },
                {
                  n: 3,
                  title: '发起请求',
                  desc: '在请求头携带 Bearer Token，POST GraphQL 查询语句。',
                },
              ].map((item, i, arr) => (
                <div key={item.n} className="flex flex-1 items-start gap-3 pr-4 last:pr-0">
                  <div className="flex shrink-0 flex-col items-center pt-0.5">
                    <div className="flex size-6 items-center justify-center rounded-full bg-primary/[0.08] text-[11px] font-semibold text-primary">
                      {item.n}
                    </div>
                    {i < arr.length - 1 && (
                      <div className="mt-1.5 h-4 w-px bg-border" />
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-foreground">{item.title}</p>
                    <p className="mt-0.5 text-xs leading-5 text-muted-foreground">{item.desc}</p>
                  </div>
                  {i < arr.length - 1 && (
                    <div className="mt-3 flex-1 border-t border-dashed border-border" />
                  )}
                </div>
              ))}
            </div>
          </section>

          <Divider />

          {/* ── 端点 ──────────────────────────────────────────────────────── */}
          <section id="endpoint" className="scroll-mt-6">
            <h2 className="text-base font-semibold text-foreground">端点</h2>
            <p className="mt-1.5 text-sm text-muted-foreground">
              每个模型拥有独立的 GraphQL 端点，路径由四级参数组成。
            </p>

            {/* Endpoint display */}
            <div className="mt-4 rounded-lg border border-border bg-[#F6F8FA] px-4 py-3">
              <div className="flex flex-wrap items-baseline gap-2.5">
                <span className="shrink-0 rounded bg-primary/[0.08] px-2 py-0.5 font-mono text-[11px] font-semibold text-primary">
                  POST
                </span>
                <code className="break-all font-mono text-xs leading-6 text-foreground">
                  {'/end-user/graphql/org/'}
                  <span className="text-amber-600">{'{orgName}'}</span>
                  {'/project/'}
                  <span className="text-amber-600">{'{projectSlug}'}</span>
                  {'/db/'}
                  <span className="text-amber-600">{'{db}'}</span>
                  {'/model/'}
                  <span className="text-amber-600">{'{model}'}</span>
                </code>
              </div>
            </div>

            {/* Path params */}
            <div className="mt-3 overflow-hidden rounded-lg border border-border">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-[#F6F8FA]">
                    <th className="px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                      参数
                    </th>
                    <th className="px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                      说明
                    </th>
                    <th className="px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                      示例
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border bg-card">
                  {[
                    { param: 'orgName', desc: '组织名称', example: orgName },
                    { param: 'projectSlug', desc: '项目标识符', example: 'my-project' },
                    { param: 'db', desc: '数据库名称', example: 'prod_db' },
                    { param: 'model', desc: '模型名称', example: 'users' },
                  ].map((row) => (
                    <tr key={row.param}>
                      <td className="px-3 py-2.5">
                        <code className="rounded bg-amber-50 px-1.5 py-0.5 font-mono text-xs text-amber-700">
                          {row.param}
                        </code>
                      </td>
                      <td className="px-3 py-2.5 text-xs text-muted-foreground">{row.desc}</td>
                      <td className="px-3 py-2.5">
                        <code className="font-mono text-xs text-foreground">{row.example}</code>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>

          <Divider />

          {/* ── 认证 ──────────────────────────────────────────────────────── */}
          <section id="auth" className="scroll-mt-6">
            <h2 className="text-base font-semibold text-foreground">认证</h2>
            <p className="mt-1.5 text-sm text-muted-foreground">
              所有请求需在 HTTP 头中携带 API Token。
            </p>
            <div className="mt-4 overflow-hidden rounded-lg border border-[#2a3050]">
              <div className="border-b border-[#2a3050] bg-[#1a1f36] px-4 py-2">
                <span className="text-[10px] font-medium uppercase tracking-wider text-[#697386]">
                  Request Header
                </span>
              </div>
              <div className="bg-[#1e2330] px-4 py-3 font-mono text-sm">
                <span className="text-[#a78bfa]">Authorization</span>
                <span className="text-[#cdd6f4]">: </span>
                <span className="text-[#89dceb]">Bearer</span>
                <span className="text-amber-300"> {'<your-token>'}</span>
              </div>
            </div>
            <p className="mt-3 text-xs text-muted-foreground">
              还没有 Token？前往{' '}
              <a
                href={`/end-user/${orgName}/dashboard/token`}
                className="text-primary hover:underline"
              >
                API Token 管理
              </a>{' '}
              页面创建。
            </p>
          </section>

          <Divider />

          {/* ── 代码示例 ──────────────────────────────────────────────────── */}
          <section id="examples" className="scroll-mt-6">
            <h2 className="text-base font-semibold text-foreground">代码示例</h2>
            <p className="mt-1.5 text-sm text-muted-foreground">
              选择语言，复制并替换参数后直接使用。
            </p>
            <div className="mt-4">
              {/* Horizontal language tabs */}
              <div className="flex border-b border-border">
                {TABS.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    onClick={() => setActiveTab(t.id)}
                    className={cn(
                      '-mb-px px-4 py-2 text-sm transition-colors',
                      activeTab === t.id
                        ? 'border-b-2 border-primary font-medium text-primary'
                        : 'text-muted-foreground hover:text-foreground'
                    )}
                  >
                    {t.label}
                  </button>
                ))}
              </div>
              <div className="mt-4">
                <DarkCodeBlock code={snippet} lang={activeLang} />
              </div>
            </div>
          </section>

          <Divider />

          {/* ── 查询参考 ──────────────────────────────────────────────────── */}
          <section id="reference" className="scroll-mt-6 pb-16">
            <h2 className="text-base font-semibold text-foreground">查询参考</h2>
            <p className="mt-1.5 text-sm text-muted-foreground">
              常用操作模板，替换字段名后直接使用。
            </p>
            <div className="mt-4 grid grid-cols-2 gap-4">
              {OPERATIONS.map((op) => (
                <div
                  key={op.name}
                  className="overflow-hidden rounded-lg border border-[#2a3050]"
                >
                  <div className="flex items-center gap-2 border-b border-[#2a3050] bg-[#1a1f36] px-3 py-2">
                    <span
                      className={cn(
                        'rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider',
                        op.type === 'QUERY'
                          ? 'bg-primary/[0.15] text-[#818cf8]'
                          : 'bg-emerald-900/40 text-emerald-400'
                      )}
                    >
                      {op.type}
                    </span>
                    <span className="text-xs font-medium text-[#cdd6f4]">{op.name}</span>
                  </div>
                  <pre className="bg-[#1e2330] p-3 font-mono text-[12px] leading-5 text-[#cdd6f4]">
                    {op.code}
                  </pre>
                </div>
              ))}
            </div>
          </section>

        </div>
      </div>
    </div>
  )
}

export default function ApiDocsPage({ params }: ApiDocsPageProps) {
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
    <EndUserAppLayout orgName={orgName} activePage="api-docs">
      <ApiDocsContent orgName={orgName} />
    </EndUserAppLayout>
  )
}
