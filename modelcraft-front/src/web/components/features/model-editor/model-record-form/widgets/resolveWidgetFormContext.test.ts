import { describe, expect, it } from 'vitest'
import { resolveWidgetFormContext } from './resolveWidgetFormContext'

// ─────────────────────────────────────────────────────────────────────────────
// resolveWidgetFormContext
//
// 为什么存在这些测试：
//   RJSF v6 将 formContext 移入 props.registry.formContext，不再作为顶层 prop
//   直接传给 widget。如果读取逻辑被"简化"为 props.formContext，widget 会在
//   design 模式下拿到 undefined 并崩溃（见 EndUserSelectorWidget 历史 bug）。
//
// 这些测试锁定了优先级契约：
//   registry.formContext > props.formContext > undefined（不 throw）
// ─────────────────────────────────────────────────────────────────────────────

describe('resolveWidgetFormContext', () => {
  // ── RJSF v6 正常路径 ──────────────────────────────────────────────────────
  it('从 registry.formContext 读取 orgName（RJSF v6 标准路径）', () => {
    const props = {
      registry: {
        formContext: { orgName: 'acme', workspaceMode: 'design' },
      },
    }

    const ctx = resolveWidgetFormContext(props)

    expect(ctx?.orgName).toBe('acme')
    expect(ctx?.workspaceMode).toBe('design')
  })

  // ── 降级路径：registry 不存在时用直接 prop ────────────────────────────────
  it('当 registry 为 undefined 时降级到 props.formContext', () => {
    const props = {
      registry: undefined,
      formContext: { orgName: 'fallback-org', workspaceMode: 'end_user' },
    }

    const ctx = resolveWidgetFormContext(props)

    expect(ctx?.orgName).toBe('fallback-org')
    expect(ctx?.workspaceMode).toBe('end_user')
  })

  // ── registry 存在但没有 formContext 时的降级 ───────────────────────────────
  it('当 registry 存在但无 formContext 时降级到 props.formContext', () => {
    const props = {
      registry: { widgets: {}, fields: {} }, // 没有 formContext 字段
      formContext: { orgName: 'partial-org' },
    }

    const ctx = resolveWidgetFormContext(props)

    expect(ctx?.orgName).toBe('partial-org')
  })

  // ── registry.formContext 优先级高于 props.formContext ─────────────────────
  it('registry.formContext 优先于 props.formContext（不合并）', () => {
    const props = {
      registry: {
        formContext: { orgName: 'registry-org' },
      },
      formContext: { orgName: 'direct-org' },
    }

    const ctx = resolveWidgetFormContext(props)

    expect(ctx?.orgName).toBe('registry-org')
  })

  // ── 两者都空时不 throw ────────────────────────────────────────────────────
  it('registry 和 formContext 均为 undefined 时返回 undefined，不抛出异常', () => {
    const props = {
      registry: undefined,
      formContext: undefined,
    }

    expect(() => resolveWidgetFormContext(props)).not.toThrow()
    expect(resolveWidgetFormContext(props)).toBeUndefined()
  })
})
