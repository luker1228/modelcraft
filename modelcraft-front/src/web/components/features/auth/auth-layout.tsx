'use client'

import { ArrowLeft } from 'lucide-react'
import Image from 'next/image'
import NextLink from 'next/link'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { NeuralCanvas } from './neural-canvas'

interface AuthLayoutProps {
  children: React.ReactNode
  title: string
  subtitle?: string
  showCliPromo?: boolean
  backLink?: {
    href: string
    label: string
    onClick?: React.MouseEventHandler<HTMLAnchorElement>
  }
}

export function AuthLayout({
  children,
  title,
  subtitle,
  showCliPromo = false,
  backLink,
}: AuthLayoutProps) {
  return (
    <div className="flex min-h-dvh bg-background">
      {/* Left brand panel — hidden on mobile */}
      <div
        className="relative hidden min-h-dvh overflow-hidden lg:flex lg:w-5/12 xl:w-1/2"
        style={{ background: '#0E0E11' }}
      >
        {/* Animated audit network background */}
        <NeuralCanvas />

        {/* Radial glow over network */}
        <div
          className="pointer-events-none absolute inset-0"
          style={{
            background: 'radial-gradient(ellipse 60% 50% at 30% 55%, rgba(79,70,229,0.12) 0%, transparent 70%)',
          }}
        />

        {/* Logo — pinned to top */}
        <div className="absolute inset-x-0 top-0 flex items-center gap-3 px-12 py-8">
          <div
            className="flex size-9 items-center justify-center rounded-lg"
            style={{ background: 'linear-gradient(135deg, #4F46E5 0%, #7C3AED 100%)' }}
          >
            <Image src="/icons/icon-model-graphql.svg" alt="ModelCraft" width={18} height={18} />
          </div>
          <span className="text-lg font-semibold text-white">ModelCraft</span>
        </div>

        {/* Content */}
        <div className="relative mx-auto flex w-full max-w-lg flex-col justify-center gap-10 px-12 py-16">
          <div className="space-y-5">
            <p className="text-xs font-semibold uppercase tracking-widest" style={{ color: '#4F46E5' }}>
              AI Data Infrastructure
            </p>
            <h1 className="text-3xl font-semibold leading-snug text-white">
              让 AI 安全、可控地使用数据库
            </h1>
            <p className="text-sm leading-relaxed" style={{ color: '#697386' }}>
              面向 AI 的数据访问底座。ModelCraft 把数据库能力封装成 AI 可调用的 GraphQL 和 CLI 接口，
              让自然语言驱动的查询与操作都保持安全、可控、可审计。
            </p>
            <div className="flex flex-col gap-2 pt-1">
              <div className="group/item cursor-default">
                <div className="flex items-center gap-3 py-2">
                  <span
                    className="flex size-5 shrink-0 items-center justify-center rounded text-[10px] font-semibold"
                    style={{ background: 'rgba(139,130,255,0.25)', color: '#EDE7FF' }}
                  >GQL</span>
                  <span className="text-sm font-medium text-white">GraphQL 统一接口</span>
                </div>
                <div className="grid grid-rows-[0fr] transition-all duration-300 ease-out group-hover/item:grid-rows-[1fr]">
                  <p className="overflow-hidden pl-8 text-xs leading-relaxed" style={{ color: '#697386' }}>
                    将数据库能力标准化为可调用接口，方便 AI 和应用统一接入
                  </p>
                </div>
              </div>
              <div className="group/item cursor-default">
                <div className="flex items-center gap-3 py-2">
                  <span
                    className="flex size-5 shrink-0 items-center justify-center rounded text-[10px] font-semibold"
                    style={{ background: 'rgba(139,130,255,0.25)', color: '#EDE7FF' }}
                  >ACL</span>
                  <span className="text-sm font-medium text-white">RBAC 细粒度授权</span>
                </div>
                <div className="grid grid-rows-[0fr] transition-all duration-300 ease-out group-hover/item:grid-rows-[1fr]">
                  <p className="overflow-hidden pl-8 text-xs leading-relaxed" style={{ color: '#697386' }}>
                    角色、权限包与字段级策略，让数据访问边界清晰可控
                  </p>
                </div>
              </div>
              {showCliPromo && (
                <div className="group/item cursor-default">
                  <div className="flex items-center gap-3 py-2">
                    <span
                      className="flex size-5 shrink-0 items-center justify-center rounded text-[10px] font-semibold"
                      style={{ background: 'rgba(139,130,255,0.25)', color: '#EDE7FF' }}
                    >CLI</span>
                    <span className="text-sm font-medium text-white">CLI 可编排调用</span>
                  </div>
                  <div className="grid grid-rows-[0fr] transition-all duration-300 ease-out group-hover/item:grid-rows-[1fr]">
                    <p className="overflow-hidden pl-8 text-xs leading-relaxed" style={{ color: '#697386' }}>
                      适合 AI Agent 和自动化任务，通过命令行安全地查询和操作数据。
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Copyright — pinned to bottom */}
        <p className="absolute inset-x-0 bottom-6 text-center text-xs" style={{ color: '#697386' }}>
          &copy; {new Date().getFullYear()} ModelCraft. 保留所有权利
        </p>
      </div>

      {/* Right form panel */}
      <div className="flex flex-1 items-center justify-center px-6 py-8">
        <div className="flex w-full max-w-[420px] flex-col items-start">
          {backLink && (
            <NextLink
              href={backLink.href}
              onClick={backLink.onClick}
              className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <ArrowLeft className="size-3.5" />
              <span>{backLink.label}</span>
            </NextLink>
          )}

          <Card className="w-full rounded-xl border border-border bg-background shadow-sm">
            {/* Mobile logo */}
            <div className="flex items-center justify-center gap-2.5 px-8 pt-8 lg:hidden">
              <div className="flex size-8 items-center justify-center rounded-lg bg-primary">
                <Image src="/icons/icon-model-graphql.svg" alt="ModelCraft" width={16} height={16} />
              </div>
              <span className="text-base font-semibold text-foreground">ModelCraft</span>
            </div>

            <CardHeader className="px-8 pb-2 pt-8">
              <CardTitle className="text-2xl">{title}</CardTitle>
              {subtitle && <CardDescription>{subtitle}</CardDescription>}
            </CardHeader>

            <CardContent className="px-8 pb-8 pt-4">
              {children}

              <p className="mt-6 text-center text-xs text-muted-foreground">
                登录即表示您同意我们的{' '}
                <a href="#" className="text-primary hover:underline">服务条款</a>
                {' '}和{' '}
                <a href="#" className="text-primary hover:underline">隐私政策</a>
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
