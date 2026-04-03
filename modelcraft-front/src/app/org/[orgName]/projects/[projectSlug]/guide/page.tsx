'use client'

import { Card, CardContent, CardHeader } from "@web/components/ui/card"
import { Button } from "@web/components/ui/button"
import {
  Rocket,
  Database,
  Network,
  FileCode,
  Sparkles,
  ArrowRight,
  CheckCircle2,
  Play
} from "lucide-react"

const guideSteps = [
  {
    id: 1,
    icon: Database,
    title: "连接数据源",
    description: "配置并连接您的第一个数据库集群,开启数据管理之旅",
    status: "completed",
    action: "前往集群管理",
    link: "/cluster",
    color: "emerald"
  },
  {
    id: 2,
    icon: FileCode,
    title: "定义数据模型",
    description: "创建数据模型和枚举类型,构建强类型的数据架构",
    status: "active",
    action: "开始定义模型",
    link: "/models",
    color: "blue"
  },
  {
    id: 3,
    icon: Network,
    title: "生成 API 接口",
    description: "自动生成 GraphQL API 和 RESTful 接口,快速构建后端服务",
    status: "pending",
    action: "配置 API",
    link: "/api",
    color: "amber"
  },
  {
    id: 4,
    icon: Sparkles,
    title: "部署与集成",
    description: "部署您的 ModelCraft 项目,与前端应用无缝集成",
    status: "pending",
    action: "查看部署指南",
    link: "/deploy",
    color: "purple"
  }
]

const quickLinks = [
  { title: "文档中心", description: "深入了解 ModelCraft 功能特性" },
  { title: "示例项目", description: "查看完整的项目示例代码" },
  { title: "API 参考", description: "GraphQL Schema 和 API 文档" },
  { title: "社区支持", description: "加入开发者社区获取帮助" }
]

export default function GuidePage() {
  const getStatusStyles = (status: string) => {
    switch (status) {
      case 'completed':
        return {
          bg: 'bg-emerald-500/10',
          border: 'border-emerald-500/30',
          text: 'text-emerald-700',
          icon: 'text-emerald-600'
        }
      case 'active':
        return {
          bg: 'bg-blue-500/10',
          border: 'border-blue-500/30',
          text: 'text-blue-700',
          icon: 'text-blue-600'
        }
      default:
        return {
          bg: 'bg-slate-100',
          border: 'border-slate-200',
          text: 'text-muted-foreground',
          icon: 'text-muted-foreground'
        }
    }
  }

  const getColorStyles = (color: string) => {
    const colors = {
      emerald: 'from-emerald-500 to-emerald-600 shadow-emerald-500/20',
      blue: 'from-primary to-primary shadow-primary/20',
      amber: 'from-amber-500 to-amber-600 shadow-amber-500/20',
      purple: 'from-purple-500 to-purple-600 shadow-purple-500/20'
    }
    return colors[color as keyof typeof colors] || colors.blue
  }

  return (
    <div className="max-w-7xl space-y-8">
      {/* Hero Section */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary via-primary/80 to-primary p-8 shadow-2xl shadow-primary/25">
        {/* Decorative Elements */}
        <div className="absolute right-0 top-0 size-64 rounded-full bg-cyan-400/20 blur-3xl" />
        <div className="absolute bottom-0 left-0 size-96 rounded-full bg-blue-500/20 blur-3xl" />

        <div className="relative z-10">
          <div className="mb-3 flex items-center space-x-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-white/25 shadow-lg backdrop-blur-sm">
              <Rocket className="size-5 text-white drop-shadow" />
            </div>
            <span className="rounded-full bg-white/25 px-3 py-1 text-xs font-semibold text-white shadow-sm backdrop-blur-sm">
              快速开始
            </span>
          </div>

          <h1 className="mb-3 font-heading text-3xl font-semibold leading-tight text-white md:text-4xl">
            欢迎使用 ModelCraft
          </h1>
          <p className="max-w-2xl text-sm leading-relaxed text-white/95 drop-shadow md:text-base">
            跟随下方引导,快速搭建您的数据驱动应用。从连接数据库到生成 API,只需四个简单步骤。
          </p>

          <div className="mt-6 flex flex-wrap gap-3">
            <Button className="cursor-pointer bg-white font-semibold text-blue-600 shadow-lg transition-all duration-200 hover:bg-white/95 hover:text-blue-700 hover:shadow-xl">
              <Play className="mr-2 size-4" />
              观看视频教程
            </Button>
            <Button className="cursor-pointer border-2 border-white/50 bg-white/20 font-semibold text-white shadow-lg backdrop-blur-sm transition-all duration-200 hover:border-white/70 hover:bg-white/30 hover:shadow-xl">
              查看文档
              <ArrowRight className="ml-2 size-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Progress Overview */}
      <div>
        <h2 className="mb-4 flex items-center gap-2 font-heading text-xl font-semibold text-foreground">
          <Sparkles className="size-5 text-blue-600" />
          学习进度
        </h2>
        <div className="grid grid-cols-1 gap-5 md:grid-cols-4">
          {[
            { label: "已完成步骤", value: "1/4", color: "emerald" },
            { label: "预计完成时间", value: "15 分钟", color: "blue" },
            { label: "当前进度", value: "25%", color: "amber" },
            { label: "解锁功能", value: "3 项", color: "purple" }
          ].map((stat, index) => (
            <Card
              key={index}
              className="animate-fade-in border-blue-100/50 bg-sidebar shadow-lg backdrop-blur-sm transition-all duration-300 hover:shadow-xl"
              style={{ animationDelay: `${index * 50}ms` }}
            >
              <CardContent className="p-6">
                <div className="mb-1 text-sm text-muted-foreground">{stat.label}</div>
                <div className={`font-heading text-3xl font-semibold ${getColorStyles(stat.color)}`}>
                  {stat.value}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Guide Steps */}
      <div>
        <h2 className="mb-6 font-heading text-2xl font-semibold text-foreground">开始指南</h2>
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {guideSteps.map((step, index) => {
            const styles = getStatusStyles(step.status)
            const Icon = step.icon

            return (
              <Card
                key={step.id}
                className={`group animate-fade-in overflow-hidden border-primary/10 bg-sidebar shadow-lg backdrop-blur-sm transition-all duration-300 hover:shadow-xl ${
                  step.status === 'active' ? 'ring-2 ring-primary/50' : ''
                }`}
                style={{ animationDelay: `${index * 100}ms` }}
              >
                {/* Top Bar */}
                <div className={`h-1.5 bg-gradient-to-r ${getColorStyles(step.color)}`} />

                <CardHeader className="pb-4">
                  <div className="mb-4 flex items-start justify-between">
                    <div className="flex items-center space-x-4">
                      <div className={`size-12 rounded-xl ${styles.bg} ${styles.border} flex items-center justify-center border-2 transition-transform duration-200 group-hover:scale-110`}>
                        <Icon className={`size-6 ${styles.icon}`} />
                      </div>
                      <div>
                        <div className="mb-1 flex items-center space-x-2">
                          <span className="font-mono text-xs text-muted-foreground">STEP {step.id}</span>
                          {step.status === 'completed' && (
                            <CheckCircle2 className="size-4 text-emerald-500" />
                          )}
                        </div>
                        <h3 className="font-heading text-xl font-semibold text-foreground">{step.title}</h3>
                      </div>
                    </div>
                  </div>
                  <p className="leading-relaxed text-muted-foreground">{step.description}</p>
                </CardHeader>

                <CardContent className="pt-0">
                  <Button
                    className={`w-full bg-gradient-to-r ${getColorStyles(step.color)} cursor-pointer text-white transition-all duration-200 hover:shadow-lg`}
                    disabled={step.status === 'pending'}
                  >
                    {step.action}
                    <ArrowRight className="ml-2 size-4" />
                  </Button>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>

      {/* Quick Links */}
      <div>
        <h2 className="mb-6 font-heading text-2xl font-semibold text-foreground">快速链接</h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          {quickLinks.map((link, index) => (
            <Card
              key={index}
              className="group animate-fade-in cursor-pointer border-blue-100/50 bg-sidebar shadow-lg backdrop-blur-sm transition-all duration-300 hover:border-blue-300/50 hover:shadow-xl"
              style={{ animationDelay: `${index * 50}ms` }}
            >
              <CardContent className="p-6">
                <h3 className="mb-2 font-heading font-semibold text-foreground transition-colors group-hover:text-blue-600">
                  {link.title}
                </h3>
                <p className="text-sm leading-relaxed text-muted-foreground">{link.description}</p>
                <ArrowRight className="mt-3 size-4 text-blue-500 transition-transform group-hover:translate-x-1" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  )
}
