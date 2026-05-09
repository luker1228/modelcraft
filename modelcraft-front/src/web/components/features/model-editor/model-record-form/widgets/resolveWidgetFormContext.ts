/**
 * resolveWidgetFormContext
 *
 * RJSF v6 将 formContext 存入 props.registry.formContext，
 * 不再作为顶层 prop 直接传给 widget（props.formContext 始终 undefined）。
 *
 * 优先级：props.registry.formContext > props.formContext > undefined
 *
 * 历史背景：
 *   EndUserSelectorWidget 曾直接读 props.formContext，RJSF v6 下拿到 undefined，
 *   导致 orgName 为 undefined 并抛出 Runtime Error。本函数封装读取逻辑，
 *   配合 resolveWidgetFormContext.test.ts 防止同类回归。
 */
export function resolveWidgetFormContext(
  props: Record<string, unknown>
): Record<string, unknown> | undefined {
  const registry = props.registry as Record<string, unknown> | undefined
  if (registry?.formContext !== undefined) {
    return registry.formContext as Record<string, unknown>
  }
  return props.formContext as Record<string, unknown> | undefined
}
