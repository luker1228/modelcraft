import { redirect } from 'next/navigation'

interface RBACPageProps {
  params: { orgName: string; projectSlug: string }
}

/**
 * /rbac 根路由重定向到 /rbac/bundles
 */
export default function RBACPage({ params }: RBACPageProps) {
  redirect(`/org/${params.orgName}/project/${params.projectSlug}/rbac/bundles`)
}
