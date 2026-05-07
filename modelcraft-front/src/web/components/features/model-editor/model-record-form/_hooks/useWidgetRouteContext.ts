import { useMemo } from 'react'

interface WidgetFormContext {
  orgName?: string
  projectSlug?: string
}

function getRouteContextFromPathname(): { orgName?: string; projectSlug?: string } {
  if (typeof window === 'undefined') return {}
  const match = window.location.pathname.match(/\/org\/([^/]+)\/project\/([^/]+)/)
  if (!match) return {}
  return {
    orgName: decodeURIComponent(match[1]),
    projectSlug: decodeURIComponent(match[2]),
  }
}

/**
 * useWidgetRouteContext — 为 RJSF custom widget 解析 orgName / projectSlug。
 *
 * 优先级：formContext.orgName/projectSlug > URL pathname 解析
 *
 * 背景：RJSF 的 props.formContext 在 SSR 首次渲染或某些版本下可能为 undefined，
 * 而 widget 所在页面的 URL 始终包含 /org/{orgName}/project/{projectSlug}，
 * 因此用 URL 作为兜底保证 client 能正常创建。
 */
export function useWidgetRouteContext(formContext: unknown): {
  orgName: string
  projectSlug: string
} {
  const ctx = (formContext as WidgetFormContext | undefined) ?? {}
  const routeCtx = useMemo(() => getRouteContextFromPathname(), [])

  return {
    orgName: ctx.orgName ?? routeCtx.orgName ?? '',
    projectSlug: ctx.projectSlug ?? routeCtx.projectSlug ?? '',
  }
}
