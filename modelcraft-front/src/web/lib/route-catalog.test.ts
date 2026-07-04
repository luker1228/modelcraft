import { describe, expect, it } from 'vitest'
import { ROUTE_CATALOG } from './route-catalog'

describe('routeCatalog', () => {
  it('keeps API Token as a top-level org route', () => {
    const apiTokenRoute = ROUTE_CATALOG.find((route) => route.title === 'API Token')

    expect(apiTokenRoute?.routeTemplate).toBe('/org/:orgName/api-tokens')
  })
})
