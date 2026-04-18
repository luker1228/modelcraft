'use client'

import { useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import { ShieldCheck, Save, Info } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Switch } from '@web/components/ui/switch'

export default function ProjectLoginSettingsPage() {
  const params = useParams()
  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string

  const [enabled, setEnabled] = useState(true)
  const [allowSignup, setAllowSignup] = useState(false)
  const [sessionTTL, setSessionTTL] = useState('60')
  const [issuer, setIssuer] = useState('modelcraft')
  const [saved, setSaved] = useState(false)

  const endUserLoginPath = useMemo(() => {
    return `/u/${orgName}/${projectSlug}/login`
  }, [orgName, projectSlug])

  const handleSave = () => {
    // TODO: connect project login settings API
    setSaved(true)
    setTimeout(() => setSaved(false), 1800)
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <section className="rounded-lg border border-border bg-background p-6 shadow-sm">
        <div className="mb-4 flex items-center gap-3">
          <div className="rounded-md bg-blue-100 p-2 text-blue-600">
            <ShieldCheck className="size-5" strokeWidth={1.5} />
          </div>
          <div>
            <h1 className="text-2xl font-semibold text-foreground">登录配置</h1>
            <p className="text-sm text-muted-foreground">
              配置项目级登录能力与会话策略
            </p>
          </div>
        </div>

        <div className="mb-6 rounded-md border border-blue-200 bg-blue-50 px-3 py-2 text-sm text-blue-700">
          <div className="flex items-start gap-2">
            <Info className="mt-0.5 size-4 flex-shrink-0" strokeWidth={1.5} />
            <p>
              当前页面已接入路由与导航，保存动作暂为前端占位。后续可直接对接项目登录配置 API。
            </p>
          </div>
        </div>

        <div className="space-y-5">
          <div className="flex items-center justify-between rounded-md border border-border p-4">
            <div>
              <Label className="text-base">启用项目登录</Label>
              <p className="mt-1 text-sm text-muted-foreground">
                关闭后，终端用户将无法通过当前项目进行登录
              </p>
            </div>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>

          <div className="flex items-center justify-between rounded-md border border-border p-4">
            <div>
              <Label className="text-base">允许用户自助注册</Label>
              <p className="mt-1 text-sm text-muted-foreground">
                开启后，用户可在登录页进行注册并自动登录
              </p>
            </div>
            <Switch checked={allowSignup} onCheckedChange={setAllowSignup} />
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="session-ttl">会话有效期（分钟）</Label>
              <Input
                id="session-ttl"
                value={sessionTTL}
                onChange={(e) => setSessionTTL(e.target.value)}
                placeholder="60"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="token-issuer">Token Issuer</Label>
              <Input
                id="token-issuer"
                value={issuer}
                onChange={(e) => setIssuer(e.target.value)}
                placeholder="modelcraft"
              />
            </div>
          </div>
        </div>

        <div className="mt-6 flex items-center gap-3">
          <Button onClick={handleSave}>
            <Save className="mr-2 size-4" strokeWidth={1.5} />
            保存配置
          </Button>
          {saved && (
            <span className="text-sm text-green-600">已保存（前端占位）</span>
          )}
        </div>
      </section>

      <section className="rounded-lg border border-border bg-background p-4 text-sm text-muted-foreground">
        End-User 登录地址：<span className="font-mono text-foreground">{endUserLoginPath}</span>
      </section>
    </div>
  )
}
