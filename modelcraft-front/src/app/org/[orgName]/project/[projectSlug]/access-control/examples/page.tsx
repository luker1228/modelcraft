'use client'

import { useParams } from 'next/navigation'
import { PageHeader, PageLayout } from '@web/components/features/layout'
import { Badge } from '@web/components/ui/badge'

const EXAMPLES = [
  {
    title: '只允许查看自己的数据',
    context: 'Using Filter',
    expression: 'row.owner_id == auth.userid',
    description: '常见于 read / update / delete 的行级过滤。',
  },
  {
    title: '只允许写入自己的 owner_id',
    context: 'Input Check',
    expression: 'input.owner_id == auth.userid',
    description: '常见于 create / update 的输入校验。',
  },
  {
    title: '限制状态范围',
    context: 'Input Check',
    expression: 'input.status in ["draft", "pending"]',
    description: '校验输入字段是否落在允许集合中。',
  },
  {
    title: '公开可读',
    context: 'Using Filter',
    expression: 'true',
    description: '表示对当前 action 不额外加过滤条件。',
  },
]

export default function AccessControlExamplesPage() {
  const params = useParams()
  const projectSlug = params?.projectSlug as string

  return (
    <PageLayout maxWidth="5xl">
      <PageHeader
        title="RLS 表达式示例"
        description={`项目 ${projectSlug} 的常用 CEL / RLS 写法速查`}
      />

      <div className="mt-6 space-y-6">
        <section className="rounded-lg border border-border bg-card p-5">
          <h2 className="text-sm font-semibold text-foreground">上下文对象</h2>
          <div className="mt-3 flex flex-wrap gap-2">
            <Badge variant="outline" className="font-mono text-xs">row.</Badge>
            <Badge variant="outline" className="font-mono text-xs">input.</Badge>
            <Badge variant="outline" className="font-mono text-xs">auth.</Badge>
          </div>
          <div className="mt-3 space-y-2 text-sm text-muted-foreground">
            <p><code>row</code> 表示当前记录行，主要用于 Using Filter。</p>
            <p><code>input</code> 表示本次写入内容，主要用于 Input Check。</p>
            <p><code>auth</code> 表示当前认证身份变量，例如 <code>auth.userid</code>。</p>
          </div>
        </section>

        <section className="space-y-3">
          {EXAMPLES.map((example) => (
            <div key={example.title} className="rounded-lg border border-border bg-card p-5">
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-foreground">{example.title}</h3>
                <Badge variant="secondary" className="text-xs">{example.context}</Badge>
              </div>
              <p className="mt-2 text-sm text-muted-foreground">{example.description}</p>
              <pre className="mt-3 overflow-x-auto rounded-md bg-muted/40 p-3 font-mono text-xs leading-6 text-foreground">
                {example.expression}
              </pre>
            </div>
          ))}
        </section>
      </div>
    </PageLayout>
  )
}
