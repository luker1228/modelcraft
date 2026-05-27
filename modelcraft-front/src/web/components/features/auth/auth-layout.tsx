'use client'

import Image from 'next/image'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Separator } from '@web/components/ui/separator'

interface AuthLayoutProps {
  children: React.ReactNode
  title: string
  subtitle: string
}

const features = [
  {
    iconName: 'link2',
    iconVariant: 'on-primary',
    iconAlt: '数据模型',
    title: '模型定义',
    desc: '结构化配置字段、关系和枚举',
    bgColor: 'bg-primary',
  },
  {
    iconName: 'shield-check',
    iconVariant: 'on-primary',
    iconAlt: '权限管理',
    title: '访问策略',
    desc: '基于角色和组织范围的授权控制',
    bgColor: 'bg-emerald-600',
  },
  {
    iconName: 'globe',
    iconVariant: 'on-primary',
    iconAlt: 'GraphQL API',
    title: '运行时接口',
    desc: '统一输出可治理的 GraphQL 能力',
    bgColor: 'bg-amber-600',
  },
]

export function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  return (
    <div className="relative flex min-h-screen overflow-hidden bg-muted/40">
      {/* Left brand panel — hidden on mobile */}
      <div className="relative z-10 hidden lg:flex lg:w-1/2">
        {/* Texture layers */}
        <div
          className="pointer-events-none absolute inset-0 opacity-60"
          style={{
            backgroundImage: `
              radial-gradient(circle at 18% 22%, rgba(79,70,229,0.12), transparent 38%),
              radial-gradient(circle at 78% 72%, rgba(8,145,178,0.1), transparent 42%),
              linear-gradient(to bottom right, rgba(15,23,42,0.03), transparent 48%)
            `
          }}
        />
        <div
          className="pointer-events-none absolute inset-0 opacity-25"
          style={{
            backgroundImage: `
              linear-gradient(rgba(15,23,42,0.06) 1px, transparent 1px),
              linear-gradient(90deg, rgba(15,23,42,0.06) 1px, transparent 1px)
            `,
            backgroundSize: '24px 24px'
          }}
        />
        <div
          className="pointer-events-none absolute inset-0 opacity-20 mix-blend-multiply"
          style={{
            backgroundImage: `
              repeating-linear-gradient(
                -35deg,
                rgba(15,23,42,0.045) 0 2px,
                transparent 2px 10px
              )
            `
          }}
        />
        <div className="mx-auto flex w-full max-w-2xl flex-col justify-between p-12">
          {/* Logo */}
          <div className="flex items-center gap-3">
            <div className="flex size-12 items-center justify-center rounded-lg bg-primary shadow-sm">
              <Image src="/icons/icon-model-graphql.svg" alt="ModelCraft" width={24} height={24} />
            </div>
            <span className="text-2xl font-semibold text-foreground">
              ModelCraft
            </span>
          </div>

          {/* Middle content */}
          <div className="space-y-8">
            <div className="space-y-5">
              <span className="inline-flex select-none rounded-md bg-primary/10 px-3 py-1.5 text-xs font-semibold tracking-wide text-primary">
                MODELCRAFT PLATFORM
              </span>

              <h1 className="max-w-xl text-4xl font-semibold leading-tight text-foreground">
                让数据模型与权限策略
                <br />
                在同一工作台协同演进
              </h1>

              <p className="max-w-xl text-[15px] leading-relaxed text-muted-foreground">
                面向工程团队的配置化数据平台，统一管理模型结构、访问控制与接口发布，降低上线风险并提升协作效率。
              </p>
            </div>

            {/* Cover visual */}
            <div className="relative overflow-hidden rounded-xl border border-border bg-card p-5 shadow-sm">
              <div className="pointer-events-none absolute -left-14 -top-14 size-36 rounded-full bg-indigo-200/30 blur-2xl" />
              <div className="pointer-events-none absolute -bottom-16 right-8 size-44 rounded-full bg-cyan-200/25 blur-2xl" />
              <div className="mb-4 flex items-center justify-between">
                <p className="text-sm font-semibold text-foreground">运行链路总览</p>
                <span className="rounded bg-primary/10 px-2 py-1 text-[11px] font-medium text-primary">
                  Production Ready
                </span>
              </div>

              <div className="grid grid-cols-[1fr_auto_1fr_auto_1fr] items-center gap-2">
                {features.map((feature, i) => (
                  <div key={i} className="contents">
                    <div className="rounded-lg border border-border bg-background p-3">
                      <div
                        className={`mb-2 flex size-8 items-center justify-center rounded-md text-white ${feature.bgColor}`}
                      >
                        <Image
                          src={`/icons/icon-${feature.iconName}${feature.iconVariant === 'on-primary' ? '-on-primary' : ''}.svg`}
                          alt={feature.iconAlt}
                          width={16}
                          height={16}
                        />
                      </div>
                      <h3 className="text-sm font-semibold text-foreground">{feature.title}</h3>
                      <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{feature.desc}</p>
                    </div>
                    {i < features.length - 1 && (
                      <div className="flex justify-center">
                        <div className="h-px w-6 bg-border/90" />
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Capability list */}
            <div className="grid grid-cols-1 gap-4">
              {features.map((feature, i) => (
                <Card
                  key={i}
                  className="group cursor-default rounded-lg border border-border bg-background transition-colors duration-200 hover:border-primary/30"
                >
                  <CardContent className="flex items-start gap-4 p-4">
                    <div
                      className={`flex size-10 shrink-0 items-center justify-center rounded-md text-white ${feature.bgColor}`}
                    >
                      <Image
                        src={`/icons/icon-${feature.iconName}${feature.iconVariant === 'on-primary' ? '-on-primary' : ''}.svg`}
                        alt={feature.iconAlt}
                        width={20}
                        height={20}
                      />
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
        <div
          className="pointer-events-none absolute inset-0 opacity-40"
          style={{
            backgroundImage: `
              radial-gradient(circle at 82% 18%, rgba(16,185,129,0.08), transparent 34%),
              radial-gradient(circle at 24% 78%, rgba(245,158,11,0.08), transparent 36%),
              linear-gradient(to top left, rgba(15,23,42,0.02), transparent 52%)
            `
          }}
        />
        <div
          className="pointer-events-none absolute inset-0 opacity-20"
          style={{
            backgroundImage: `
              linear-gradient(rgba(15,23,42,0.06) 1px, transparent 1px),
              linear-gradient(90deg, rgba(15,23,42,0.06) 1px, transparent 1px)
            `,
            backgroundSize: '24px 24px'
          }}
        />
        <div
          className="pointer-events-none absolute inset-0 opacity-10 mix-blend-multiply"
          style={{
            backgroundImage: `
              repeating-linear-gradient(
                -35deg,
                rgba(15,23,42,0.04) 0 2px,
                transparent 2px 12px
              )
            `
          }}
        />
        <Card className="relative z-10 w-full max-w-md rounded-lg border border-border bg-background shadow-sm">
          {/* Mobile logo */}
          <div className="flex items-center justify-center gap-3 px-8 pt-8 lg:hidden">
            <div className="flex size-10 items-center justify-center rounded-lg bg-primary">
              <Image src="/icons/icon-model-graphql.svg" alt="ModelCraft" width={20} height={20} />
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
