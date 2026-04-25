import { describe, expect, it } from 'vitest'
import { buildAppLayoutBreadcrumbs } from './app-layout-breadcrumbs'

describe('buildAppLayoutBreadcrumbs', () => {
  it('returns org -> project breadcrumbs in project context', () => {
    const breadcrumbs = buildAppLayoutBreadcrumbs({
      showProjectNav: true,
      orgName: 'acme',
      orgDisplayName: 'ACME Org',
      projectSlug: 'order-center',
      projectDisplayName: 'Order Center',
    })

    expect(breadcrumbs).toEqual([
      {
        label: 'ACME Org',
        href: '/org/acme/workspace',
        isCurrent: false,
      },
      {
        label: 'Order Center',
        isCurrent: true,
      },
    ])
  })

  it('falls back to route params when display names are unavailable', () => {
    const breadcrumbs = buildAppLayoutBreadcrumbs({
      showProjectNav: true,
      orgName: 'acme',
      projectSlug: 'order-center',
    })

    expect(breadcrumbs).toEqual([
      {
        label: 'acme',
        href: '/org/acme/workspace',
        isCurrent: false,
      },
      {
        label: 'order-center',
        isCurrent: true,
      },
    ])
  })

  it('returns empty breadcrumbs outside project context', () => {
    const breadcrumbs = buildAppLayoutBreadcrumbs({
      showProjectNav: false,
      orgName: 'acme',
      projectSlug: 'order-center',
    })

    expect(breadcrumbs).toEqual([
      {
        label: 'acme',
        isCurrent: true,
      },
    ])
  })

  it('uses org display name in workspace context', () => {
    const breadcrumbs = buildAppLayoutBreadcrumbs({
      showProjectNav: false,
      orgName: 'acme',
      orgDisplayName: 'ACME Org',
    })

    expect(breadcrumbs).toEqual([
      {
        label: 'ACME Org',
        isCurrent: true,
      },
    ])
  })
})
