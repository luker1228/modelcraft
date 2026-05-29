'use client'

import { useState } from 'react'
import { Save } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Switch } from '@web/components/ui/switch'
import { toast } from 'sonner'

export default function OrgLoginSettingsPage() {
  const [enabled, setEnabled] = useState(true)
  const [allowSignup, setAllowSignup] = useState(false)
  const [sessionTTL, setSessionTTL] = useState('60')
  const [issuer, setIssuer] = useState('modelcraft')
  const [saving, setSaving] = useState(false)

  const endUserLoginPath = '/end-user/login'

  const handleSave = () => {
    // TODO: connect org login settings API
    setSaving(true)
    setTimeout(() => {
      setSaving(false)
      toast.success('配置已保存')
    }, 800)
  }

  return (
    <div className="space-y-6">
      {/* Toggle rows */}
      <div className="divide-y divide-border rounded-md border border-border">
        <div className="flex items-center justify-between p-4">
          <div className="space-y-0.5">
            <p className="text-sm font-medium text-foreground">启用终端用户登录</p>
            <p className="text-sm text-muted-foreground">
              关闭后，终端用户将无法通过本 Org 进行登录
            </p>
          </div>
          <Switch checked={enabled} onCheckedChange={setEnabled} />
        </div>

        <div className="flex items-center justify-between p-4">
          <div className="space-y-0.5">
            <p className="text-sm font-medium text-foreground">允许用户自助注册</p>
            <p className="text-sm text-muted-foreground">
              开启后，用户可在登录页进行注册并自动登录
            </p>
          </div>
          <Switch checked={allowSignup} onCheckedChange={setAllowSignup} />
        </div>
      </div>

      {/* Input fields */}
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1.5">
          <Label htmlFor="session-ttl">会话有效期（分钟）</Label>
          <Input
            id="session-ttl"
            type="number"
            min={1}
            value={sessionTTL}
            onChange={(e) => setSessionTTL(e.target.value)}
            placeholder="60"
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="token-issuer">令牌颁发者（Token Issuer）</Label>
          <Input
            id="token-issuer"
            value={issuer}
            onChange={(e) => setIssuer(e.target.value)}
            placeholder="modelcraft"
          />
        </div>
      </div>

      {/* Login path info row */}
      <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <span>终端用户登录地址</span>
        <span className="text-border">·</span>
        <span className="font-mono text-foreground">{endUserLoginPath}</span>
      </div>

      {/* Actions */}
      <div>
        <Button onClick={handleSave} disabled={saving}>
          <Save className="mr-2 size-4" strokeWidth={1.5} />
          {saving ? '保存中…' : '保存配置'}
        </Button>
      </div>
    </div>
  )
}
