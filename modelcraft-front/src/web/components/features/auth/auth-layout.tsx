'use client'

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Separator } from '@web/components/ui/separator'
import { Sparkles, Globe, ShieldCheck, Link as LinkIcon } from 'lucide-react'

interface AuthLayoutProps {
  children: React.ReactNode
  title: string
  subtitle: string
}

const features = [
  {
    icon: <Globe className="size-5" strokeWidth={1.5} />,
    title: '数据库 HTTP 接口',
    desc: '自动将数据库表暴露为 GraphQL API 接口',
    bgColor: 'bg-primary',
  },
  {
    icon: <ShieldCheck className="size-5" strokeWidth={1.5} />,
    title: '权限管理',
    desc: '基于角色的细粒度数据访问权限控制',
    bgColor: 'bg-emerald-600',
  },
  {
    icon: <LinkIcon className="size-5" strokeWidth={1.5} />,
    title: '逻辑外键与枚举',
    desc: '无需数据库约束，灵活定义关联关系与枚举类型',
    bgColor: 'bg-amber-600',
  },
]

export function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  return (
    <div className="relative flex min-h-screen overflow-hidden bg-muted/40">
      {/* Left brand panel — hidden on mobile */}
      <div className="relative z-10 hidden lg:flex lg:w-1/2">
        <div className="mx-auto flex w-full max-w-2xl flex-col justify-between p-12">
          {/* Logo */}
          <div className="flex items-center gap-3">
            <div className="flex size-12 items-center justify-center rounded-lg bg-primary shadow-sm">
              <Sparkles className="size-6 text-primary-foreground" strokeWidth={1.5} />
            </div>
            <span className="text-2xl font-semibold text-foreground">
              ModelCraft
            </span>
          </div>

          {/* Middle content */}
          <div className="space-y-8">
            <div className="space-y-6">
              <span className="inline-flex select-none rounded bg-primary/10 px-4 py-2 text-sm font-semibold text-primary">
                ✨ 数据库能力即服务平台
              </span>

              <h1 className="text-5xl font-semibold leading-tight text-foreground">
                让数据库能力触手可及
              </h1>

              <p className="max-w-lg text-lg leading-relaxed text-muted-foreground">
                为开发者打造的数据库接口平台。HTTP 接口、行级权限、逻辑关联 ——
                无需编写后端代码。
              </p>
            </div>

            {/* Feature cards */}
            <div className="grid grid-cols-1 gap-4">
              {features.map((feature, i) => (
                <Card
                  key={i}
                  className="group cursor-default rounded-lg border border-border bg-background transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
                >
                  <CardContent className="flex items-start gap-4 p-4">
                    <div
                      className={`flex size-10 shrink-0 items-center justify-center rounded-md text-white ${feature.bgColor}`}
                    >
                      {feature.icon}
                    </div>
                    <div className="min-w-0 flex-1">
                      <h3 className="mb-1 font-semibold text-foreground">
                        {feature.title}
                      </h3>
                      <p className="text-sm text-muted-foreground">
                        {feature.desc}
                      </p>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>&copy; {new Date().getFullYear()} ModelCraft. 保留所有权利</span>
            <div className="flex gap-4">
              <a href="#" className="transition-colors hover:text-foreground">
                隐私政策
              </a>
              <a href="#" className="transition-colors hover:text-foreground">
                服务条款
              </a>
            </div>
          </div>
        </div>
      </div>

      {/* Right form panel */}
      <div className="relative z-10 flex flex-1 items-center justify-center p-6">
        <Card className="w-full max-w-md rounded-lg border border-border bg-background shadow-sm">
          {/* Mobile logo */}
          <div className="flex items-center justify-center gap-3 px-8 pt-8 lg:hidden">
            <div className="flex size-10 items-center justify-center rounded-lg bg-primary">
              <Sparkles className="size-5 text-primary-foreground" strokeWidth={1.5} />
            </div>
            <span className="text-xl font-semibold text-foreground">
              ModelCraft
            </span>
          </div>

          <CardHeader className="px-8 pt-8 lg:pt-10">
            <CardTitle className="text-center text-3xl lg:text-left">
              {title}
            </CardTitle>
            <CardDescription className="text-center lg:text-left">
              {subtitle}
            </CardDescription>
          </CardHeader>

          <CardContent className="px-8 pb-8">
            {children}

            {/* Footer disclaimer */}
            <div className="mt-8 space-y-4">
              <Separator />
              <p className="pt-2 text-center text-xs leading-relaxed text-muted-foreground">
                登录即表示您同意我们的{' '}
                <a href="#" className="text-primary hover:underline">
                  服务条款
                </a>{' '}
                和{' '}
                <a href="#" className="text-primary hover:underline">
                  隐私政策
                </a>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
