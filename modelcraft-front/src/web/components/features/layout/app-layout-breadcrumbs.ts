export interface AppLayoutBreadcrumbItem {
  label: string
  href?: string
  isCurrent: boolean
}

interface BuildAppLayoutBreadcrumbsInput {
  showProjectNav: boolean
  orgName?: string
  orgDisplayName?: string
  projectSlug?: string
  projectDisplayName?: string
}

export function buildAppLayoutBreadcrumbs({
  showProjectNav,
  orgName,
  orgDisplayName,
  projectSlug,
  projectDisplayName,
}: BuildAppLayoutBreadcrumbsInput): AppLayoutBreadcrumbItem[] {
  if (!orgName) {
    return []
  }

  if (!showProjectNav) {
    return [
      {
        label: orgDisplayName || orgName,
        isCurrent: true,
      },
    ]
  }

  if (!projectSlug) {
    return []
  }

  return [
    {
      label: orgDisplayName || orgName,
      href: `/org/${orgName}/workspace`,
      isCurrent: false,
    },
    {
      label: projectDisplayName || projectSlug,
      isCurrent: true,
    },
  ]
}
