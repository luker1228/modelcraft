import { useQuery } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { GET_RLS_POLICIES } from '@/api-client/rls-policy'
import type { RlsPolicy, RlsPoliciesOrderBy } from '@/generated/graphql'

interface UseRlsPolicyListProps {
  projectSlug: string
  modelId: string | null
  orderBy?: RlsPoliciesOrderBy | null
}

interface UseRlsPolicyListReturn {
  policies: RlsPolicy[]
  loading: boolean
  error: Error | undefined
}

export function useRlsPolicyList({ projectSlug, modelId, orderBy }: UseRlsPolicyListProps): UseRlsPolicyListReturn {
  const client = useProjectScopedClient(projectSlug)

  const { data, loading, error } = useQuery(GET_RLS_POLICIES, {
    client,
    variables: { modelId, orderBy: orderBy ?? null },
    skip: !modelId,
  })

  const policies: RlsPolicy[] = data?.rlsPolicies ?? []

  return { policies, loading, error }
}
